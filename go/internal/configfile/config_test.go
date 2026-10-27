package configfile

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/MateeDevs/sentiary-tools/go/internal/download"
)

const validConfig = `config_version = 1
[projects.app]
id = "project"
api_key = "project-secret"
languages = ["en", "cs"]
source_language = "en"
include_empty = true
[outputs.ios]
project = "app"
format = "xcstrings"
layout = "combined"
path = "generated/Localizable.xcstrings"
`

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sentiary.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestOutputInheritsProjectSettings(t *testing.T) {
	config, err := Load(writeConfig(t, validConfig), "", "")
	if err != nil {
		t.Fatal(err)
	}
	output := config.Outputs["ios"]
	if !reflect.DeepEqual(output.Languages, []string{"en", "cs"}) || !output.IncludeEmpty || output.SourceLanguage != "en" {
		t.Fatalf("incorrect inherited output: %+v", output)
	}
}

func TestDocumentedExamplePlansAndroidComposeAndOneCatalog(t *testing.T) {
	config, err := Load("../../examples/sentiary.toml", ".", "")
	if err != nil {
		t.Fatal(err)
	}
	files, err := download.Plan(context.Background(), config, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 5 {
		t.Fatalf("expected two Android files, two Compose files, and one catalog; got %d", len(files))
	}
	catalog := files[len(files)-1]
	if catalog.Output != "ios" || catalog.Layout != download.Combined || !reflect.DeepEqual(catalog.Languages, []string{"cs", "en"}) {
		t.Fatalf("incorrect catalog plan: %+v", catalog)
	}
}

func TestExplicitOutputSettingsReplaceDefaults(t *testing.T) {
	config, err := Load(writeConfig(t, validConfig+`languages = ["cs"]
source_language = "cs"
include_empty = false
`), "", "")
	if err != nil {
		t.Fatal(err)
	}
	output := config.Outputs["ios"]
	if !reflect.DeepEqual(output.Languages, []string{"cs"}) || output.IncludeEmpty || output.SourceLanguage != "cs" {
		t.Fatalf("incorrect overridden output: %+v", output)
	}
}

func TestCredentialsPrecedence(t *testing.T) {
	for _, test := range []struct{ name, environment, local, expected string }{
		{"project", "", "", "project-secret"},
		{"local", "", "personal-secret", "personal-secret"},
		{"environment", "environment-secret", "personal-secret", "environment-secret"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := writeConfig(t, validConfig)
			if test.local != "" {
				if err := os.WriteFile(filepath.Join(filepath.Dir(path), "sentiary.local.toml"), []byte(`user_api_key = "`+test.local+`"`), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			config, err := Load(path, "", test.environment)
			if err != nil {
				t.Fatal(err)
			}
			if config.Projects["app"].APIKey != test.expected {
				t.Fatal("incorrect credential precedence")
			}
		})
	}
}

func TestLocalSupplementCannotChangeOutputs(t *testing.T) {
	for _, test := range []struct{ name, content string }{
		{"populated table", "[outputs.other]\npath = 'other'"},
		{"empty table", "[outputs.other]\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := writeConfig(t, validConfig)
			if err := os.WriteFile(filepath.Join(filepath.Dir(path), "sentiary.local.toml"), []byte(test.content), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(path, "", ""); err == nil {
				t.Fatal("accepted local output overrides")
			}
		})
	}
}

func TestInvalidConfigurations(t *testing.T) {
	for _, test := range []struct{ name, content string }{
		{"empty languages", strings.Replace(validConfig, `["en", "cs"]`, `[]`, 1)},
		{"missing languages", strings.Replace(validConfig, `languages = ["en", "cs"]`, "", 1)},
		{"duplicates", strings.Replace(validConfig, `["en", "cs"]`, `["en", "en"]`, 1)},
		{"missing source", strings.Replace(validConfig, `source_language = "en"`, "", 1)},
		{"unsupported combined format", strings.Replace(validConfig, `format = "xcstrings"`, `format = "android"`, 1)},
		{"combined path expression", strings.Replace(validConfig, `generated/Localizable`, `generated/{language}/Localizable`, 1)},
		{"combined map", validConfig + "[outputs.ios.locale_map]\nen = 'en'\n"},
		{"combined empty map", validConfig + "[outputs.ios.locale_map]\n"},
		{"empty source override", validConfig + "source_language = ''\n"},
		{"unknown property", validConfig + "unknown = true\n"},
		{"unknown version", strings.Replace(validConfig, "config_version = 1", "config_version = 2", 1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Load(writeConfig(t, test.content), "", ""); err == nil {
				t.Fatal("invalid config accepted")
			}
		})
	}
}

func TestDecodeErrorsDoNotRevealSecrets(t *testing.T) {
	_, err := Load(writeConfig(t, strings.Replace(validConfig, `api_key = "project-secret"`, `api_key = ["private-secret"]`, 1)), "", "")
	if err == nil || strings.Contains(err.Error(), "private-secret") {
		t.Fatalf("unsafe decode error: %v", err)
	}
}

func TestDiscoveryUsesNearestParentAndExplicitOverride(t *testing.T) {
	path := writeConfig(t, validConfig)
	child := filepath.Join(filepath.Dir(path), "child", "nested")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	config, err := Load("", child, "")
	if err != nil {
		t.Fatal(err)
	}
	expected, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if config.Directory != expected {
		t.Fatalf("wrong config directory: %q", config.Directory)
	}
	if err := os.WriteFile(filepath.Join(child, "sentiary.toml"), []byte("invalid"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load("", child, ""); err == nil {
		t.Fatal("ignored nearest config")
	}
	if _, err := Load(path, child, ""); err != nil {
		t.Fatal(err)
	}
}
