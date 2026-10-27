package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/MateeDevs/sentiary-tools/internal/download"
	"github.com/pelletier/go-toml/v2"
)

// init only assembles local configuration for existing project IDs. No API is called.
func (app *App) runInit(ctx context.Context, args []string) error {
	flags := app.newFlagSet("init")
	directory := flags.String("directory", ".", "Directory for a new sentiary.toml.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	configPath := filepath.Join(*directory, "sentiary.toml")
	if _, err := os.Lstat(configPath); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("config already exists or cannot be inspected")
	}
	reader := bufio.NewReader(app.Stdin)
	ask := func(prompt string) (string, error) {
		fmt.Fprint(app.Stderr, prompt+": ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("init cancelled: %w", err)
		}
		if err := ctx.Err(); err != nil {
			return "", err
		}
		value := strings.TrimSpace(line)
		if value == "cancel" {
			return "", errors.New("init cancelled")
		}
		return value, nil
	}
	projects := map[string]any{}
	outputs := map[string]any{}
	for {
		alias, err := ask("Project alias (enter an existing project; 'cancel' aborts)")
		if err != nil {
			return err
		}
		if alias == "" || projects[alias] != nil {
			return errors.New("project alias must be nonempty and unique")
		}
		id, err := ask("Existing project ID")
		if err != nil {
			return err
		}
		if id == "" {
			return errors.New("project ID is required")
		}
		languages, err := ask("Language IDs, comma separated (or *)")
		if err != nil {
			return err
		}
		var selection any = splitList(languages)
		if languages == "*" {
			selection = "*"
		} else if err := download.ValidateLanguages(selection.([]string)); err != nil {
			return err
		}
		project := map[string]any{"id": id, "languages": selection}
		saveKey, err := ask("Save SENTIARY_PROJECT_API_KEY in shared TOML? [y/N]")
		if err != nil {
			return err
		}
		if strings.EqualFold(saveKey, "y") {
			key := os.Getenv("SENTIARY_PROJECT_API_KEY")
			if key == "" {
				return errors.New("set SENTIARY_PROJECT_API_KEY before choosing to store it")
			}
			project["api_key"] = key
		}
		projects[alias] = project
		for {
			name, err := ask("Output name")
			if err != nil {
				return err
			}
			if name == "" || outputs[name] != nil {
				return errors.New("output name must be nonempty and unique")
			}
			format, err := ask("Format (android, compose, xcstrings for combined Xcode, or any export format)")
			if err != nil {
				return err
			}
			capability, ok := download.LookupFormat(format)
			if !ok {
				return errors.New("unsupported format")
			}
			layout := string(download.PerLanguage)
			path := ""
			switch format {
			case "android":
				path = "build/sentiary/" + name + "/{locale}/strings.xml"
			case "compose":
				path = "build/sentiary/" + name + "/{locale}/strings.xml"
			case "xcstrings":
				layout = string(download.Combined)
				path = "build/sentiary/" + name + "/Localizable.xcstrings"
			}
			entered, err := ask("Output path [" + path + "]")
			if err != nil {
				return err
			}
			if entered != "" {
				path = entered
			}
			if path == "" {
				return errors.New("output path is required")
			}
			output := map[string]any{"project": alias, "format": format, "layout": layout, "path": path}
			if layout == string(download.Combined) && capability.Combined {
				source, err := ask("Source language ID (must be selected)")
				if err != nil {
					return err
				}
				if source == "" {
					return errors.New("source language is required")
				}
				if ids, ok := selection.([]string); ok {
					found := false
					for _, id := range ids {
						if id == source {
							found = true
						}
					}
					if !found {
						return errors.New("source language must be selected")
					}
				}
				output["source_language"] = source
			}
			if strings.Contains(path, "{locale}") {
				mapping, err := ask("Locale mappings, e.g. en=values,cs=values-cs (include every target)")
				if err != nil {
					return err
				}
				locales := map[string]string{}
				for _, item := range splitList(mapping) {
					key, value, ok := strings.Cut(item, "=")
					if !ok || key == "" || value == "" {
						return errors.New("invalid locale mapping")
					}
					locales[key] = value
				}
				if ids, ok := selection.([]string); ok {
					for _, id := range ids {
						if locales[id] == "" {
							return errors.New("missing locale mapping")
						}
					}
				}
				output["locale_map"] = locales
			}
			policy := download.Output{Format: format, Layout: download.Layout(layout), Path: path}
			if ids, ok := selection.([]string); ok {
				policy.Languages = ids
			} else {
				policy.AllLanguages = true
			}
			if source, ok := output["source_language"].(string); ok {
				policy.SourceLanguage = source
			}
			if locales, ok := output["locale_map"].(map[string]string); ok {
				policy.LocaleMap = locales
			}
			if err := download.ValidateOutput(policy); err != nil {
				return err
			}
			outputs[name] = output
			more, err := ask("Add another output for this project? [y/N]")
			if err != nil {
				return err
			}
			if !strings.EqualFold(more, "y") {
				break
			}
		}
		more, err := ask("Add another existing project? [y/N]")
		if err != nil {
			return err
		}
		if !strings.EqualFold(more, "y") {
			break
		}
	}
	local, err := ask("Save SENTIARY_USER_API_KEY to sentiary.local.toml? Default uses environment [y/N]")
	if err != nil {
		return err
	}
	var localData []byte
	if strings.EqualFold(local, "y") {
		key := os.Getenv("SENTIARY_USER_API_KEY")
		if key == "" {
			return errors.New("set SENTIARY_USER_API_KEY before choosing to store it")
		}
		localData, err = toml.Marshal(map[string]any{"user_api_key": key})
		if err != nil {
			return err
		}
		if _, err := os.Lstat(filepath.Join(*directory, "sentiary.local.toml")); !errors.Is(err, os.ErrNotExist) {
			return errors.New("local config already exists or cannot be inspected")
		}
	}
	ignore, err := ask("Add sentiary.local.toml and .sentiary/ to .gitignore? [Y/n]")
	if err != nil {
		return err
	}
	data, err := toml.Marshal(map[string]any{"config_version": 1, "projects": projects, "outputs": outputs})
	if err != nil {
		return err
	}
	// Marshal a separate preview so credentials can never appear in it.
	for _, raw := range projects {
		p := raw.(map[string]any)
		if _, ok := p["api_key"]; ok {
			p["api_key"] = "[redacted]"
		}
	}
	preview, _ := toml.Marshal(map[string]any{"config_version": 1, "projects": projects, "outputs": outputs})
	fmt.Fprintf(app.Stderr, "\nPreview (%s):\n%s\nLocal credentials: %t\nUpdate .gitignore: %t\n", configPath, preview, len(localData) > 0, !strings.EqualFold(ignore, "n"))
	confirm, err := ask("Write these files? [y/N]")
	if err != nil {
		return err
	}
	if !strings.EqualFold(confirm, "y") {
		return errors.New("init cancelled")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	created := []string{}
	committed := false
	defer func() {
		if !committed {
			for _, name := range created {
				_ = os.Remove(name)
			}
		}
	}()
	create := func(path string, data []byte) error {
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			return err
		}
		created = append(created, path)
		_, writeErr := file.Write(data)
		return errors.Join(writeErr, file.Close())
	}
	if err := create(configPath, data); err != nil {
		return err
	}
	if len(localData) > 0 {
		if err := create(filepath.Join(*directory, "sentiary.local.toml"), localData); err != nil {
			return err
		}
	}
	if !strings.EqualFold(ignore, "n") {
		path := filepath.Join(*directory, ".gitignore")
		previous, err := os.ReadFile(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		additions := ""
		for _, entry := range []string{"sentiary.local.toml", ".sentiary/"} {
			found := false
			for _, line := range strings.Split(string(previous), "\n") {
				if strings.TrimSpace(line) == entry {
					found = true
				}
			}
			if !found {
				additions += entry + "\n"
			}
		}
		if additions != "" {
			if len(previous) > 0 && previous[len(previous)-1] != '\n' {
				additions = "\n" + additions
			}
			file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				return err
			}
			_, writeErr := io.WriteString(file, additions)
			if err := errors.Join(writeErr, file.Close()); err != nil {
				return err
			}
		}
	}
	committed = true
	fmt.Fprintln(app.Stdout, "Created sentiary.toml.")
	return nil
}
