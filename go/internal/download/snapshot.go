package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"time"
)

// State identifies one project's export semantics. Revision is decimal, never a JSON number.
type State struct {
	ProtocolVersion   int      `json:"protocolVersion"`
	ProjectGeneration string   `json:"projectGeneration"`
	DataRevision      string   `json:"dataRevision"`
	ExporterVersion   string   `json:"exporterVersion"`
	Languages         []string `json:"languages"`
}

func (s State) Validate() error {
	if s.ProtocolVersion != 1 || s.ProjectGeneration == "" || s.ExporterVersion == "" || s.DataRevision == "" {
		return fmt.Errorf("server requires export/state protocol version 1")
	}
	for _, digit := range s.DataRevision {
		if digit < '0' || digit > '9' {
			return fmt.Errorf("invalid export revision")
		}
	}
	return nil
}
func (s State) SameRevision(other State) bool {
	return s.ProjectGeneration == other.ProjectGeneration && s.DataRevision == other.DataRevision && s.ExporterVersion == other.ExporterVersion
}

type ExportResponse struct {
	Content     []byte
	ETag        string
	State       State
	NotModified bool
}

// SnapshotRepository exposes conditional exports without changing legacy export/import commands.
type SnapshotRepository interface {
	Repository
	Server() string
	State(context.Context, Project) (State, error)
	ConditionalExport(context.Context, Project, File, string) (ExportResponse, error)
}

// Request excludes credentials, aliases, and output paths.
type Request struct {
	ProjectID      string
	Format         string
	Layout         Layout
	Languages      []string
	SourceLanguage string
	IncludeEmpty   bool
}

func RequestFor(config Config, file File) Request {
	return Request{config.Projects[file.Project].ID, file.Format, file.Layout, slices.Sorted(slices.Values(file.Languages)), file.SourceLanguage, file.IncludeEmpty}
}
func Digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func (r Request) ID() string    { data, _ := json.Marshal(r); return Digest(data) }

type CachedExport struct {
	Request  Request
	Checksum string
	ETag     string
	Content  []byte `json:"-"`
	Valid    bool   `json:"-"`
}

type Snapshot struct {
	Version     int
	Selection   string
	CompletedAt time.Time
	Projects    map[string]State
	Exports     map[string]CachedExport
}

// SnapshotCache reads and publishes complete sets under short process locks.
type SnapshotCache interface {
	Read() ([]Snapshot, error)
	Publish(Snapshot) error
}

type statePlanner struct{ states map[string]State }

func (p statePlanner) Languages(_ context.Context, project Project) ([]string, error) {
	return p.states[project.ID].Languages, nil
}
func (statePlanner) Export(context.Context, Project, File) ([]byte, error) {
	return nil, fmt.Errorf("planner cannot export")
}

func selectedConfig(config Config, selected []string) (Config, error) {
	copy := config
	copy.Outputs = map[string]Output{}
	if len(selected) == 0 {
		return config, nil
	}
	for _, name := range selected {
		output, ok := config.Outputs[name]
		if !ok {
			return Config{}, fmt.Errorf("unknown output %q", name)
		}
		if _, duplicate := copy.Outputs[name]; duplicate {
			return Config{}, fmt.Errorf("duplicate output selection")
		}
		copy.Outputs[name] = output
	}
	return copy, nil
}

func selectionID(config Config) string {
	var requests []string
	for _, output := range config.Outputs {
		languages := output.Languages
		if output.AllLanguages {
			languages = []string{"*"}
		}
		source := output.SourceLanguage
		if output.Layout == PerLanguage {
			source = ""
		}
		requests = append(requests, (Request{config.Projects[output.Project].ID, output.Format, output.Layout,
			slices.Sorted(slices.Values(languages)), source, output.IncludeEmpty}).ID())
	}
	slices.Sort(requests)
	data, _ := json.Marshal(slices.Compact(requests))
	return Digest(data)
}

// FetchSnapshot validates every project against one observed state, retrying only that project.
// It never falls back to offline data on errors. Cache publication precedes destination writes.
func FetchSnapshot(ctx context.Context, config Config, selected []string, repository SnapshotRepository, cache SnapshotCache) ([]DownloadedFile, error) {
	config, err := selectedConfig(config, selected)
	if err != nil {
		return nil, err
	}
	previous, err := cache.Read()
	if err != nil {
		return nil, err
	}
	snapshot := Snapshot{Version: 1, Selection: selectionID(config), Projects: map[string]State{}, Exports: map[string]CachedExport{}}
	projectOutputs := map[string]map[string]Output{}
	for name, output := range config.Outputs {
		id := config.Projects[output.Project].ID
		if projectOutputs[id] == nil {
			projectOutputs[id] = map[string]Output{}
		}
		projectOutputs[id][name] = output
	}
	for id, outputs := range projectOutputs {
		var project Project
		for _, output := range outputs {
			project = config.Projects[output.Project]
			break
		}
		settled := false
		for attempt := 0; attempt < 3; attempt++ {
			state, err := repository.State(ctx, project)
			if err != nil {
				return nil, fmt.Errorf("export/state protocol v1 required: %w", err)
			}
			if err := state.Validate(); err != nil {
				return nil, err
			}
			part := config
			part.Outputs = outputs
			files, err := Plan(ctx, part, nil, statePlanner{map[string]State{id: state}})
			if err != nil {
				return nil, err
			}
			exports := map[string]CachedExport{}
			changed := false
			for _, file := range files {
				request := RequestFor(config, file)
				key := request.ID()
				if _, duplicate := exports[key]; duplicate {
					continue
				}
				cached, found := findExport(previous, key)
				etag := ""
				if found {
					etag = cached.ETag
				}
				response, err := repository.ConditionalExport(ctx, config.Projects[file.Project], file, etag)
				if err != nil {
					return nil, err
				}
				if !response.State.SameRevision(state) {
					changed = true
					break
				}
				if response.ETag == "" {
					return nil, fmt.Errorf("export response is missing ETag")
				}
				content := response.Content
				if response.NotModified {
					if !found || etag == "" || response.ETag != etag {
						return nil, fmt.Errorf("unexpected conditional export response")
					}
					content = cached.Content
				}
				exports[key] = CachedExport{request, Digest(content), response.ETag, content, true}
			}
			if changed {
				continue
			}
			snapshot.Projects[id] = state
			for key, export := range exports {
				snapshot.Exports[key] = export
			}
			settled = true
			break
		}
		if !settled {
			return nil, fmt.Errorf("project changed during three download attempts")
		}
	}
	files, err := Plan(ctx, config, nil, statePlanner{snapshot.Projects})
	if err != nil {
		return nil, err
	}
	downloaded, err := restoreFiles(config, files, snapshot)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	snapshot.CompletedAt = time.Now().UTC()
	if err := cache.Publish(snapshot); err != nil {
		return nil, fmt.Errorf("publish download set: %w", err)
	}
	return downloaded, nil
}

func findExport(snapshots []Snapshot, key string) (CachedExport, bool) {
	for _, snapshot := range snapshots {
		export, ok := snapshot.Exports[key]
		if ok && export.Valid && export.Checksum == Digest(export.Content) {
			return export, true
		}
	}
	return CachedExport{}, false
}

// RestoreSnapshot uses exactly one completed set, including its language expansion.
func RestoreSnapshot(ctx context.Context, config Config, selected []string, cache SnapshotCache) ([]DownloadedFile, time.Time, error) {
	config, err := selectedConfig(config, selected)
	if err != nil {
		return nil, time.Time{}, err
	}
	snapshots, err := cache.Read()
	if err != nil {
		return nil, time.Time{}, err
	}
	for _, snapshot := range snapshots {
		covered := true
		for _, output := range config.Outputs {
			if _, ok := snapshot.Projects[config.Projects[output.Project].ID]; !ok {
				covered = false
			}
		}
		if !covered {
			continue
		}
		files, err := Plan(ctx, config, nil, statePlanner{snapshot.Projects})
		if err != nil {
			return nil, time.Time{}, err
		}
		for _, file := range files {
			if _, ok := snapshot.Exports[RequestFor(config, file).ID()]; !ok {
				covered = false
			}
		}
		if !covered {
			continue
		}
		restored, err := restoreFiles(config, files, snapshot)
		return restored, snapshot.CompletedAt, err
	}
	return nil, time.Time{}, fmt.Errorf("no completed download set covers the requested exports")
}
func restoreFiles(config Config, files []File, snapshot Snapshot) ([]DownloadedFile, error) {
	result := make([]DownloadedFile, 0, len(files))
	for _, file := range files {
		export, ok := snapshot.Exports[RequestFor(config, file).ID()]
		if !ok || !export.Valid || export.Checksum != Digest(export.Content) {
			return nil, fmt.Errorf("cached export is missing or corrupt; run an online pull")
		}
		result = append(result, DownloadedFile{file, export.Content})
	}
	return result, nil
}
