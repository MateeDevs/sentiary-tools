// Package configfile adapts TOML configuration to download policies.
package configfile

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/MateeDevs/sentiary-tools/internal/download"
)

type defaults struct {
	Languages      any     `toml:"languages"`
	SourceLanguage *string `toml:"source_language"`
	IncludeEmpty   *bool   `toml:"include_empty"`
}

type project struct {
	ID     string `toml:"id"`
	APIKey string `toml:"api_key"`
	defaults
}

type output struct {
	Project         string             `toml:"project"`
	Format          string             `toml:"format"`
	Layout          download.Layout    `toml:"layout"`
	Path            string             `toml:"path"`
	LocaleMap       *map[string]string `toml:"locale_map"`
	LanguageAliases map[string]string  `toml:"language_aliases"`
	defaults
}

type document struct {
	Version  int                `toml:"config_version"`
	Projects map[string]project `toml:"projects"`
	Outputs  map[string]output  `toml:"outputs"`
}

// Load discovers a shared config and reads the credentials-only local supplement.
// Decode errors deliberately omit source snippets, which may contain API keys.
func Load(filename, directory, userAPIKey string) (download.Config, error) {
	return load(filename, directory, userAPIKey, false)
}

// LoadOffline ignores local credentials and permits projects without an API key.
func LoadOffline(filename, directory string) (download.Config, error) {
	return load(filename, directory, "", true)
}

func load(filename, directory, userAPIKey string, offline bool) (download.Config, error) {
	path, err := discover(filename, directory)
	if err != nil {
		return download.Config{}, err
	}
	var doc document
	if err := decode(path, &doc); err != nil {
		return download.Config{}, err
	}
	if doc.Version != 1 {
		return download.Config{}, fmt.Errorf("unsupported config_version; expected 1")
	}
	root, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		return download.Config{}, err
	}
	var local struct {
		UserAPIKey string `toml:"user_api_key"`
	}
	localPath := filepath.Join(root, "sentiary.local.toml")
	if err := func() error {
		if offline {
			return nil
		}
		return decode(localPath, &local)
	}(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return download.Config{}, err
	}
	key := strings.TrimSpace(userAPIKey)
	if key == "" {
		key = strings.TrimSpace(local.UserAPIKey)
	}
	config := download.Config{Name: filepath.Base(path), Directory: root, Projects: map[string]download.Project{}, Outputs: map[string]download.Output{}}
	for alias, project := range doc.Projects {
		if strings.TrimSpace(alias) == "" || strings.TrimSpace(project.ID) == "" {
			return download.Config{}, fmt.Errorf("projects need a nonempty alias and id")
		}
		apiKey := key
		if apiKey == "" {
			apiKey = strings.TrimSpace(project.APIKey)
		}
		if apiKey == "" && !offline {
			return download.Config{}, fmt.Errorf("project %q: provide api_key or SENTIARY_USER_API_KEY", alias)
		}
		config.Projects[alias] = download.Project{ID: project.ID, APIKey: apiKey}
	}
	if len(doc.Outputs) == 0 {
		return download.Config{}, fmt.Errorf("config must define at least one output")
	}
	for name, raw := range doc.Outputs {
		if strings.TrimSpace(name) == "" {
			return download.Config{}, fmt.Errorf("outputs need nonempty names")
		}
		parent, ok := doc.Projects[raw.Project]
		if !ok {
			return download.Config{}, fmt.Errorf("output %q: unknown project", name)
		}
		format, ok := download.LookupFormat(raw.Format)
		if !ok {
			return download.Config{}, fmt.Errorf("output %q: unsupported format", name)
		}
		layout := raw.Layout
		if layout == "" {
			layout = download.PerLanguage
		}
		if layout != download.PerLanguage && layout != download.Combined {
			return download.Config{}, fmt.Errorf("output %q: unsupported layout", name)
		}
		if strings.TrimSpace(raw.Path) == "" || strings.ContainsRune(raw.Path, 0) {
			return download.Config{}, fmt.Errorf("output %q: path is required", name)
		}
		remaining := strings.NewReplacer("{language}", "", "{locale}", "").Replace(raw.Path)
		if strings.ContainsAny(remaining, "{}") {
			return download.Config{}, fmt.Errorf("output %q: unknown path expression", name)
		}
		if layout == download.Combined && (!format.Combined || raw.LocaleMap != nil || raw.LanguageAliases != nil || strings.ContainsAny(raw.Path, "{}")) {
			return download.Config{}, fmt.Errorf("output %q: combined requires a supported format, fixed path, and no locale_map", name)
		}
		languages := raw.Languages
		if languages == nil {
			languages = parent.Languages
		}
		ids, all, err := languageSelection(languages)
		if err != nil {
			return download.Config{}, fmt.Errorf("output %q: %w", name, err)
		}
		includeEmpty := parent.IncludeEmpty
		if raw.IncludeEmpty != nil {
			includeEmpty = raw.IncludeEmpty
		}
		source := raw.SourceLanguage
		if source == nil {
			source = parent.SourceLanguage
		}
		sourceLanguage := ""
		if source != nil {
			sourceLanguage = *source
		}
		if layout == download.Combined && strings.TrimSpace(sourceLanguage) == "" {
			return download.Config{}, fmt.Errorf("output %q: source_language is required", name)
		}
		var localeMap map[string]string
		if raw.LocaleMap != nil {
			localeMap = *raw.LocaleMap
		}
		config.Outputs[name] = download.Output{Project: raw.Project, Format: raw.Format, Layout: layout,
			Languages: ids, AllLanguages: all, SourceLanguage: sourceLanguage, IncludeEmpty: includeEmpty != nil && *includeEmpty,
			Path: raw.Path, LocaleMap: localeMap, LanguageAliases: raw.LanguageAliases}
		if err := download.ValidateOutput(config.Outputs[name]); err != nil {
			return download.Config{}, fmt.Errorf("output %q: %w", name, err)
		}
	}
	return config, nil
}

func languageSelection(value any) ([]string, bool, error) {
	if text, ok := value.(string); ok && text == "*" {
		return nil, true, nil
	}
	values, ok := value.([]any)
	if !ok {
		return nil, false, fmt.Errorf("languages must be a nonempty list or \"*\"")
	}
	ids := make([]string, len(values))
	for index, value := range values {
		id, ok := value.(string)
		if !ok {
			return nil, false, fmt.Errorf("language IDs must be strings")
		}
		ids[index] = id
	}
	return ids, false, download.ValidateLanguages(ids)
}

func decode(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := toml.NewDecoder(bytes.NewReader(data)).DisallowUnknownFields().Decode(target); err != nil {
		return fmt.Errorf("invalid TOML or unsupported fields in %s", filepath.Base(path))
	}
	return nil
}

func discover(filename, directory string) (string, error) {
	if filename != "" {
		if !filepath.IsAbs(filename) {
			filename = filepath.Join(directory, filename)
		}
		return filepath.Abs(filename)
	}
	directory, err := filepath.Abs(directory)
	if err != nil {
		return "", err
	}
	for {
		path := filepath.Join(directory, "sentiary.toml")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("no sentiary.toml found; use --config")
		}
		directory = parent
	}
}
