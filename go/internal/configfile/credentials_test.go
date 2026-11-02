package configfile

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/MateeDevs/sentiary-tools/go/internal/download"
)

func TestLiteralAPIKeysDoNotInterpretReferences(t *testing.T) {
	for _, test := range []struct{ name, value string }{
		{"plain string", "ordinary-secret"},
		{"variable name", "MOBILE_KEY"},
		{"placeholder text", "${MOBILE_KEY}"},
		{"reference-looking text", `{ env = "MOBILE_KEY" }`},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := writeConfig(t, strings.Replace(validConfig, `api_key = "project-secret"`, "api_key = "+strconv.Quote(test.value), 1))
			config, err := Load(path, "", func(name string) string {
				if name != "SENTIARY_USER_API_KEY" && name != "SENTIARY_PROJECT_API_KEY" {
					t.Fatal("literal key caused an environment reference lookup")
				}
				return ""
			})
			if err != nil {
				t.Fatal(err)
			}
			if config.Projects["app"].APIKey != test.value {
				t.Fatal("literal API key was interpreted or changed")
			}
		})
	}
}

func TestDocumentedProjectCredentialsExample(t *testing.T) {
	config, err := Load("../../../docs/content/examples/project-credentials.toml", ".", environment(map[string]string{
		"SENTIARY_MOBILE_API_KEY":     "mobile-secret",
		"SENTIARY_BACKOFFICE_API_KEY": "backoffice-secret",
	}))
	if err != nil {
		t.Fatal(err)
	}
	files, err := download.Plan(context.Background(), config, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 4 || config.Projects["mobile"].APIKey != "mobile-secret" || config.Projects["backoffice"].APIKey != "backoffice-secret" {
		t.Fatal("incorrect example configuration")
	}
}

func TestOfflineNeverReadsCredentialEnvironmentOrLocalSupplement(t *testing.T) {
	path := writeConfig(t, strings.Replace(validConfig, `api_key = "project-secret"`, `api_key = { env = "MOBILE_KEY" }`, 1))
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "sentiary.local.toml"), []byte("invalid TOML [private-secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := load(path, "", func(string) string { t.Fatal("offline credential lookup"); return "" }, true)
	if err != nil {
		t.Fatal(err)
	}
	if config.Projects["app"].APIKey != "" {
		t.Fatal("offline config retains credentials")
	}
}

func TestOnlineValidatesLocalSupplementEvenWithExplicitProjectKey(t *testing.T) {
	path := writeConfig(t, validConfig)
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "sentiary.local.toml"), []byte("invalid TOML [private-secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path, "", environment(map[string]string{"SENTIARY_USER_API_KEY": "unused-user-secret"}))
	if err == nil || !strings.Contains(err.Error(), "invalid TOML") || strings.Contains(err.Error(), "secret") {
		t.Fatalf("unsafe or missing local validation error: %v", err)
	}
}

func TestSharedCredentialStructureValidatedBeforeEnvironmentOnlineAndOffline(t *testing.T) {
	for _, test := range []struct{ name, field string }{
		{"empty variable name", `api_key = { env = "" }`},
		{"spaced variable name", `api_key = { env = " BAD_KEY" }`},
		{"mixed sources", `api_key = { env = "MOBILE_KEY", value = "private-secret" }`},
		{"duplicate key", "api_key = 'private-secret'\napi_key = { env = 'MOBILE_KEY' }"},
		{"empty table", `api_key = {}`},
		{"non-string env", `api_key = { env = 42 }`},
		{"unknown field", `api_key = { unknown = "private-secret" }`},
		{"array", `api_key = ["private-secret"]`},
		{"boolean", `api_key = true`},
		{"number", `api_key = 42`},
	} {
		for _, offline := range []bool{false, true} {
			t.Run(test.name+map[bool]string{true: " offline", false: " online"}[offline], func(t *testing.T) {
				path := writeConfig(t, strings.Replace(validConfig, `api_key = "project-secret"`, test.field, 1))
				_, err := load(path, "", func(string) string { t.Fatal("credential lookup before structural validation"); return "" }, offline)
				if err == nil {
					t.Fatal("invalid shared structure accepted")
				}
				if strings.Contains(err.Error(), "private-secret") {
					t.Fatal("invalid credential value leaked in error")
				}
			})
		}
	}
}

func TestOutputsValidatedBeforeCredentialResolution(t *testing.T) {
	path := writeConfig(t, strings.Replace(validConfig, `layout = "combined"`, `layout = "invalid"`, 1))
	if _, err := Load(path, "", func(string) string { t.Fatal("credential lookup before output validation"); return "" }); err == nil {
		t.Fatal("invalid layout accepted")
	}
}

func TestReferencedKeyOverridesGlobalUserAndLoadsSeparatelyPerProject(t *testing.T) {
	path := writeConfig(t, strings.Replace(validConfig, `api_key = "project-secret"`, `api_key = { env = "MOBILE_KEY" }`, 1)+"\n[projects.web]\nid = 'web-project'\napi_key = { env = 'WEB_KEY' }\n")
	config, err := Load(path, "", environment(map[string]string{"MOBILE_KEY": "mobile-secret", "WEB_KEY": "web-secret", "SENTIARY_USER_API_KEY": "user-secret"}))
	if err != nil {
		t.Fatal(err)
	}
	if config.Projects["app"].APIKey != "mobile-secret" || config.Projects["web"].APIKey != "web-secret" {
		t.Fatal("incorrect project credentials")
	}
}

func TestLocalUserOverridesGlobalProjectEvenWithEnvironmentUserUnset(t *testing.T) {
	path := writeConfig(t, strings.Replace(validConfig, `api_key = "project-secret"`, `api_key = " "`, 1))
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "sentiary.local.toml"), []byte(`user_api_key = "local-secret"`), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := Load(path, "", environment(map[string]string{"SENTIARY_PROJECT_API_KEY": "project-secret"}))
	if err != nil || config.Projects["app"].APIKey != "local-secret" {
		t.Fatalf("wrong credential precedence: %v", err)
	}
}
