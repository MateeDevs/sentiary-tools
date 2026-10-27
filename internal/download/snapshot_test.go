package download

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type memoryCache struct {
	sets []Snapshot
	fail bool
}

func (c *memoryCache) Read() ([]Snapshot, error) { return c.sets, nil }
func (c *memoryCache) Publish(s Snapshot) error {
	if c.fail {
		return errors.New("disk full")
	}
	c.sets = append([]Snapshot{s}, c.sets...)
	return nil
}

type snapshotRepo struct {
	state      State
	calls      int
	states     int
	conditions []string
	export     func(File, string) (ExportResponse, error)
}

func (r *snapshotRepo) Server() string { return "https://example.test" }
func (r *snapshotRepo) State(context.Context, Project) (State, error) {
	r.states++
	return r.state, nil
}
func (r *snapshotRepo) Languages(context.Context, Project) ([]string, error) {
	return r.state.Languages, nil
}
func (r *snapshotRepo) Export(context.Context, Project, File) ([]byte, error) { panic("legacy export") }
func (r *snapshotRepo) ConditionalExport(_ context.Context, _ Project, f File, etag string) (ExportResponse, error) {
	r.calls++
	r.conditions = append(r.conditions, etag)
	if r.export != nil {
		return r.export(f, etag)
	}
	return ExportResponse{Content: []byte{0, 255, 1}, ETag: `W/"tag"`, State: r.state, NotModified: etag != ""}, nil
}
func newSnapshotRepo() *snapshotRepo {
	return &snapshotRepo{state: State{1, "generation", "1", "version", []string{"en", "cs"}}}
}

func TestUnchangedOnlineUsesETagAndRestoresBinary(t *testing.T) {
	cache := &memoryCache{}
	repo := newSnapshotRepo()
	ctx := context.Background()
	first, err := FetchSnapshot(ctx, config(), nil, repo, cache)
	if err != nil {
		t.Fatal(err)
	}
	second, err := FetchSnapshot(ctx, config(), nil, repo, cache)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) || !reflect.DeepEqual(repo.conditions, []string{"", `W/"tag"`}) {
		t.Fatalf("incorrect conditional restoration: %+v", repo.conditions)
	}
}
func TestProjectStateChangeReexpandsWildcard(t *testing.T) {
	cache := &memoryCache{}
	repo := newSnapshotRepo()
	cfg := config()
	out := cfg.Outputs["ios"]
	out.AllLanguages = true
	out.Languages = nil
	cfg.Outputs["ios"] = out
	repo.export = func(f File, _ string) (ExportResponse, error) {
		if repo.calls == 1 {
			repo.state.DataRevision = "2"
			repo.state.Languages = []string{"en", "cs", "de"}
		}
		return ExportResponse{Content: []byte("new"), ETag: "tag", State: repo.state}, nil
	}
	files, err := FetchSnapshot(context.Background(), cfg, nil, repo, cache)
	if err != nil {
		t.Fatal(err)
	}
	if repo.states != 2 || len(files[0].Languages) != 3 || cache.sets[0].Projects["p"].DataRevision != "2" {
		t.Fatalf("did not retry selection: %+v", files)
	}
}
func TestRepeatedStateChangesDoNotPublish(t *testing.T) {
	cache := &memoryCache{}
	repo := newSnapshotRepo()
	repo.export = func(File, string) (ExportResponse, error) {
		s := repo.state
		s.DataRevision = "999"
		return ExportResponse{State: s}, nil
	}
	if _, err := FetchSnapshot(context.Background(), config(), nil, repo, cache); err == nil {
		t.Fatal("accepted moving state")
	}
	if repo.states != 3 || len(cache.sets) != 0 {
		t.Fatal("wrong retry/publication behavior")
	}
}
func TestCorruptOnlineObjectIsDownloadedWithoutValidator(t *testing.T) {
	cache := &memoryCache{}
	repo := newSnapshotRepo()
	ctx := context.Background()
	if _, err := FetchSnapshot(ctx, config(), nil, repo, cache); err != nil {
		t.Fatal(err)
	}
	for key, export := range cache.sets[0].Exports {
		export.Content = []byte("corrupt")
		cache.sets[0].Exports[key] = export
	}
	if _, err := FetchSnapshot(ctx, config(), nil, repo, cache); err != nil {
		t.Fatal(err)
	}
	if repo.conditions[1] != "" {
		t.Fatal("corrupt object sent validator")
	}
}
func TestOfflineSubsetIgnoresNamesPathsAndCredentials(t *testing.T) {
	cache := &memoryCache{}
	repo := newSnapshotRepo()
	cfg := config()
	other := cfg.Outputs["ios"]
	other.Format = "json"
	other.Layout = PerLanguage
	other.Path = "{language}.json"
	cfg.Outputs["android"] = other
	if _, err := FetchSnapshot(context.Background(), cfg, nil, repo, cache); err != nil {
		t.Fatal(err)
	}
	moved := config()
	out := moved.Outputs["ios"]
	out.Path = "/other-checkout/build/catalog"
	moved.Outputs = map[string]Output{"renamed": out}
	moved.Projects["app"] = Project{ID: "p"}
	files, completed, err := RestoreSnapshot(context.Background(), moved, nil, cache)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Path != out.Path || completed.IsZero() {
		t.Fatalf("wrong restoration: %+v", files)
	}
}
func TestOfflineDoesNotCombineSets(t *testing.T) {
	cache := &memoryCache{}
	repo := newSnapshotRepo()
	cfg := config()
	out := cfg.Outputs["ios"]
	out.Format = "json"
	out.Layout = PerLanguage
	out.Path = "{language}.json"
	cfg.Outputs["android"] = out
	for _, name := range []string{"ios", "android"} {
		if _, err := FetchSnapshot(context.Background(), cfg, []string{name}, repo, cache); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := RestoreSnapshot(context.Background(), cfg, nil, cache); err == nil {
		t.Fatal("combined independent sets")
	}
}
func TestOfflineRejectsNewestCorruptCoveringSet(t *testing.T) {
	cache := &memoryCache{}
	repo := newSnapshotRepo()
	ctx := context.Background()
	if _, err := FetchSnapshot(ctx, config(), nil, repo, cache); err != nil {
		t.Fatal(err)
	}
	for key, export := range cache.sets[0].Exports {
		export.Valid = false
		cache.sets[0].Exports[key] = export
	}
	if files, _, err := RestoreSnapshot(ctx, config(), nil, cache); err == nil || files != nil {
		t.Fatal("restored corrupt files")
	}
}
func TestCacheFailureReturnsNoFilesAndPreservesPreviousSet(t *testing.T) {
	cache := &memoryCache{}
	repo := newSnapshotRepo()
	ctx := context.Background()
	if _, err := FetchSnapshot(ctx, config(), nil, repo, cache); err != nil {
		t.Fatal(err)
	}
	previous := cache.sets[0].CompletedAt
	cache.fail = true
	files, err := FetchSnapshot(ctx, config(), nil, repo, cache)
	if err == nil || files != nil || !cache.sets[0].CompletedAt.Equal(previous) {
		t.Fatal("failed cache publication changed set")
	}
}
func TestOnlineFailureDoesNotUseCompletedSet(t *testing.T) {
	cache := &memoryCache{}
	repo := newSnapshotRepo()
	ctx := context.Background()
	if _, err := FetchSnapshot(ctx, config(), nil, repo, cache); err != nil {
		t.Fatal(err)
	}
	repo.export = func(File, string) (ExportResponse, error) { return ExportResponse{}, errors.New("unauthorized") }
	if _, err := FetchSnapshot(ctx, config(), nil, repo, cache); err == nil {
		t.Fatal("fell back offline")
	}
}
func TestOfflineWildcardUsesStoredLanguageList(t *testing.T) {
	cache := &memoryCache{}
	repo := newSnapshotRepo()
	cfg := config()
	out := cfg.Outputs["ios"]
	out.AllLanguages = true
	out.Languages = nil
	cfg.Outputs["ios"] = out
	if _, err := FetchSnapshot(context.Background(), cfg, nil, repo, cache); err != nil {
		t.Fatal(err)
	}
	cache.sets[0].CompletedAt = time.Now().Add(-365 * 24 * time.Hour)
	files, _, err := RestoreSnapshot(context.Background(), cfg, nil, cache)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(files[0].Languages, []string{"cs", "en"}) {
		t.Fatal("wrong stored languages")
	}
}

func TestPartialDownloadCannotPublishAnySet(t *testing.T) {
	cache := &memoryCache{}
	repo := newSnapshotRepo()
	cfg := config()
	out := cfg.Outputs["ios"]
	out.Format = "json"
	out.Layout = PerLanguage
	out.Path = "{language}.json"
	cfg.Outputs["ios"] = out
	repo.export = func(File, string) (ExportResponse, error) {
		if repo.calls == 2 {
			return ExportResponse{}, errors.New("connection lost")
		}
		return ExportResponse{Content: []byte("first"), State: repo.state, ETag: "tag"}, nil
	}
	files, err := FetchSnapshot(context.Background(), cfg, nil, repo, cache)
	if err == nil || files != nil || len(cache.sets) != 0 {
		t.Fatal("partial download escaped as completed set")
	}
}
