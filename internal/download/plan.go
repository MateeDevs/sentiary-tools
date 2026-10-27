// Package download plans translation downloads independently of config syntax and transport.
package download

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

type Layout string

const (
	PerLanguage Layout = "per_language"
	Combined    Layout = "combined"
)

// Format describes the export capabilities shared by planning and API access.
type Format struct {
	Combined bool
}

var formats = map[string]Format{
	"json": {}, "android": {}, "apple": {}, "compose": {}, "arb": {}, "csv": {},
	"ini": {}, "key_value_json": {}, "i18next": {}, "poeditor_json": {}, "po": {},
	"pot": {}, "mo": {}, "properties": {}, "resw": {}, "resx": {}, "ts": {},
	"xliff": {}, "xcstrings": {Combined: true}, "stringsdict": {}, "xlf": {},
	"xmb": {}, "xtb": {}, "xliff_1_2": {}, "rise_360_xliff": {}, "xls": {}, "xlsx": {}, "yml": {},
}

func LookupFormat(name string) (Format, bool) {
	format, ok := formats[name]
	return format, ok
}

type Project struct {
	ID     string
	APIKey string
}

type Output struct {
	Project         string
	Format          string
	Layout          Layout
	Languages       []string
	AllLanguages    bool
	SourceLanguage  string
	IncludeEmpty    bool
	Path            string
	LocaleMap       map[string]string
	LanguageAliases map[string]string
}

type Config struct {
	Name      string
	Directory string
	Projects  map[string]Project
	Outputs   map[string]Output
}

type DownloadedFile struct {
	File
	Content []byte
}

type Store interface {
	Validate([]File) error
	Commit([]DownloadedFile) error
}

// Pull downloads the complete selection before changing any destination file.
func Pull(ctx context.Context, config Config, files []File, repository Repository, store Store) error {
	if err := store.Validate(files); err != nil {
		return err
	}
	downloaded := make([]DownloadedFile, 0, len(files))
	for _, file := range files {
		content, err := repository.Export(ctx, config.Projects[file.Project], file)
		if err != nil {
			return fmt.Errorf("output %q: download failed: %w", file.Output, err)
		}
		downloaded = append(downloaded, DownloadedFile{File: file, Content: content})
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// Detect edits made while requests were in flight as well.
	if err := store.Validate(files); err != nil {
		return err
	}
	return store.Commit(downloaded)
}

// File contains no credentials and can safely be shown in a download preview.
type File struct {
	Output         string
	Project        string
	Format         string
	Layout         Layout
	Languages      []string
	SourceLanguage string
	IncludeEmpty   bool
	Path           string
}

// Repository provides project languages and complete exported files.
type Repository interface {
	Languages(context.Context, Project) ([]string, error)
	Export(context.Context, Project, File) ([]byte, error)
}

// Plan resolves all selected destinations before any export or file write.
func Plan(ctx context.Context, config Config, selected []string, repository Repository) ([]File, error) {
	names := slices.Clone(selected)
	if len(names) == 0 {
		for name := range config.Outputs {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	if len(slices.Compact(slices.Clone(names))) != len(names) {
		return nil, fmt.Errorf("duplicate output selection")
	}
	for _, name := range names {
		if _, ok := config.Outputs[name]; !ok {
			return nil, fmt.Errorf("unknown output %q", name)
		}
	}
	var files []File
	destinations := map[string]string{}
	projectLanguages := map[string][]string{}
	for _, name := range names {
		output := config.Outputs[name]
		languages := slices.Clone(output.Languages)
		var actual []string
		if output.AllLanguages || len(output.LanguageAliases) > 0 {
			var ok bool
			actual, ok = projectLanguages[output.Project]
			if !ok {
				var err error
				actual, err = repository.Languages(ctx, config.Projects[output.Project])
				if err != nil {
					return nil, fmt.Errorf("output %q: get project languages: %w", name, err)
				}
				projectLanguages[output.Project] = actual
			}
		}
		if len(output.LanguageAliases) > 0 {
			if output.Layout != PerLanguage {
				return nil, fmt.Errorf("output %q: language_aliases require per_language", name)
			}
			for target := range output.LanguageAliases {
				if _, err := resolveAlias(target, output.LanguageAliases, actual); err != nil {
					return nil, fmt.Errorf("output %q: %w", name, err)
				}
			}
		}
		if output.AllLanguages {
			languages = slices.Clone(actual)
			for target := range output.LanguageAliases {
				if !slices.Contains(languages, target) {
					languages = append(languages, target)
				}
			}
		}
		if err := ValidateLanguages(languages); err != nil {
			return nil, fmt.Errorf("output %q: %w", name, err)
		}
		languages = slices.Sorted(slices.Values(languages))
		base := File{Output: name, Project: output.Project, Format: output.Format, Layout: output.Layout,
			Languages: languages, SourceLanguage: output.SourceLanguage, IncludeEmpty: output.IncludeEmpty}
		if output.Layout == Combined {
			if !slices.Contains(languages, output.SourceLanguage) {
				return nil, fmt.Errorf("output %q: source_language must be selected", name)
			}
			base.Path = resolvePath(config.Directory, output.Path)
			files = append(files, base)
		} else {
			for _, language := range languages {
				path := strings.ReplaceAll(output.Path, "{language}", language)
				if strings.Contains(path, "{locale}") {
					locale, ok := output.LocaleMap[language]
					if !ok || locale == "" {
						return nil, fmt.Errorf("output %q: missing locale_map entry for %q", name, language)
					}
					path = strings.ReplaceAll(path, "{locale}", locale)
				}
				file := base
				source := language
				if len(output.LanguageAliases) > 0 {
					var err error
					source, err = resolveAlias(language, output.LanguageAliases, actual)
					if err != nil {
						return nil, fmt.Errorf("output %q: %w", name, err)
					}
				}
				file.Languages = []string{source}
				file.SourceLanguage = ""
				file.Path = resolvePath(config.Directory, path)
				files = append(files, file)
			}
		}
	}
	for _, file := range files {
		identity := strings.ToLower(file.Path)
		if previous, ok := destinations[identity]; ok {
			return nil, fmt.Errorf("outputs %q and %q have the same destination", previous, file.Output)
		}
		destinations[identity] = file.Output
	}
	for path := range destinations {
		for parent := filepath.Dir(path); ; parent = filepath.Dir(parent) {
			if _, exists := destinations[parent]; exists {
				return nil, fmt.Errorf("an output destination is also used as a directory")
			}
			if parent == filepath.Dir(parent) {
				break
			}
		}
	}
	return files, nil
}

func ValidateLanguages(languages []string) error {
	if len(languages) == 0 {
		return fmt.Errorf("languages must not be empty")
	}
	seen := map[string]bool{}
	for _, language := range languages {
		if strings.TrimSpace(language) == "" || strings.ContainsAny(language, "/\\{}\x00") || language == "." || language == ".." {
			return fmt.Errorf("invalid language ID")
		}
		if seen[language] {
			return fmt.Errorf("duplicate language ID %q", language)
		}
		seen[language] = true
	}
	return nil
}

func resolvePath(directory, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(directory, filepath.FromSlash(path))
}

// resolveAlias keeps target paths separate from the final API/cache language.
func resolveAlias(target string, aliases map[string]string, actual []string) (string, error) {
	seen := map[string]bool{}
	current := target
	for {
		if err := ValidateLanguages([]string{current}); err != nil {
			return "", err
		}
		if seen[current] {
			return "", fmt.Errorf("language alias cycle at %q", current)
		}
		seen[current] = true
		next, ok := aliases[current]
		if !ok {
			break
		}
		current = next
	}
	if !slices.Contains(actual, current) {
		return "", fmt.Errorf("language alias %q ends at unknown project language %q", target, current)
	}
	return current, nil
}

// ValidateOutput checks representation-independent rules before writing configuration.
func ValidateOutput(output Output) error {
	format, ok := LookupFormat(output.Format)
	if !ok {
		return fmt.Errorf("unsupported format")
	}
	if output.Layout != PerLanguage && output.Layout != Combined {
		return fmt.Errorf("unsupported layout")
	}
	if strings.TrimSpace(output.Path) == "" || strings.ContainsRune(output.Path, 0) {
		return fmt.Errorf("output path is required")
	}
	remaining := strings.NewReplacer("{language}", "", "{locale}", "").Replace(output.Path)
	if strings.ContainsAny(remaining, "{}") {
		return fmt.Errorf("unknown path expression")
	}
	if !output.AllLanguages {
		if err := ValidateLanguages(output.Languages); err != nil {
			return err
		}
	}
	if output.Layout == Combined {
		if !format.Combined || output.LocaleMap != nil || output.LanguageAliases != nil || strings.ContainsAny(output.Path, "{}") {
			return fmt.Errorf("combined output requires a supported format, fixed path, and no locale map or aliases")
		}
		if output.SourceLanguage == "" || (!output.AllLanguages && !slices.Contains(output.Languages, output.SourceLanguage)) {
			return fmt.Errorf("combined source language must be selected")
		}
	}
	if strings.Contains(output.Path, "{locale}") && !output.AllLanguages {
		for _, language := range output.Languages {
			if output.LocaleMap[language] == "" {
				return fmt.Errorf("missing locale mapping for %q", language)
			}
		}
	}
	return nil
}
