package cli

import (
	"bytes"
	"context"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/MateeDevs/sentiary-tools/go/internal/sentiary"
)

const credentialConfig = `config_version = 1
[projects.app]
id = "mobile"
api_key = { env = "MOBILE_KEY" }
languages = ["en"]
[projects.web]
id = "backoffice"
api_key = { env = "WEB_KEY" }
languages = ["en"]
[outputs.mobile]
project = "app"
format = "compose"
path = "generated/mobile.xml"
[outputs.web]
project = "web"
format = "key_value_json"
path = "generated/web.json"
`

func TestPullProjectEnvironmentCredentialsAndRotation(t *testing.T) {
	for _, mode := range []string{"references", "single project"} {
		t.Run(mode, func(t *testing.T) {
			config := credentialConfig
			expected := map[string]string{"mobile": "mobile-secret", "backoffice": "web-secret"}
			if mode == "single project" {
				config = strings.Replace(catalogConfig, `api_key = "private-project-key"`, `api_key = " "`, 1)
				expected = map[string]string{"mobile-project": "project-secret"}
			}
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				parts := strings.Split(r.URL.Path, "/")
				project := parts[4]
				if r.Header.Get("Authorization") != "Bearer "+expected[project] {
					t.Error("incorrect project authorization")
				}
				w.Header().Set("Content-Type", "application/json")
				if strings.HasSuffix(r.URL.Path, "/state") {
					_, _ = w.Write([]byte(`{"projectGeneration":"generation","dataRevision":"1","exporterVersion":"v1","languages":["en","cs"]}`))
					return
				}
				w.Header().Set("ETag", `W/"stable"`)
				w.Header().Set("X-Sentiary-Project-Generation", "generation")
				w.Header().Set("X-Sentiary-Data-Revision", "1")
				w.Header().Set("X-Sentiary-Exporter-Version", "v1")
				_, _ = w.Write([]byte("exact server bytes"))
			}))
			defer server.Close()
			app, path, stdout := pullApp(t, config, func(*http.Request) (*http.Response, error) { t.Fatal("unexpected fixture transport"); return nil, nil })
			app.Downloads = sentiary.NewDownloadRepository(sentiary.Config{BaseURL: server.URL}, server.Client())
			stderr := &bytes.Buffer{}
			app.Stderr = stderr
			if mode == "references" {
				t.Setenv("MOBILE_KEY", "mobile-secret")
				t.Setenv("WEB_KEY", "web-secret")
				t.Setenv("SENTIARY_USER_API_KEY", "unused-user-secret")
			} else {
				t.Setenv("SENTIARY_PROJECT_API_KEY", "project-secret")
			}
			args := []string{"pull", "--config", path}
			if err := app.Run(context.Background(), append(args, "--dry-run")); err != nil {
				t.Fatal(err)
			}
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatal(err)
			}
			cache := os.Getenv("SENTIARY_CACHE_DIR")
			before := readCredentialArtifacts(t, cache)
			if mode == "references" {
				expected["mobile"] = "rotated-mobile-secret"
				expected["backoffice"] = "rotated-web-secret"
				t.Setenv("MOBILE_KEY", expected["mobile"])
				t.Setenv("WEB_KEY", expected["backoffice"])
			} else {
				expected["mobile-project"] = "rotated-project-secret"
				t.Setenv("SENTIARY_PROJECT_API_KEY", expected["mobile-project"])
			}
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatal(err)
			}
			after := readCredentialArtifacts(t, cache)
			if len(before) != len(after) {
				t.Fatal("rotation changed cache identity or created export objects")
			}
			for path := range before {
				if _, ok := after[path]; !ok {
					t.Fatal("rotation changed cache identity")
				}
			}
			withoutReferences := strings.ReplaceAll(strings.ReplaceAll(config, `api_key = { env = "MOBILE_KEY" }`, ""), `api_key = { env = "WEB_KEY" }`, "")
			if err := os.WriteFile(path, []byte(withoutReferences), 0o600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("SENTIARY_PROJECT_API_KEY", "")
			t.Setenv("SENTIARY_USER_API_KEY", "switched-user-secret")
			for project := range expected {
				expected[project] = "switched-user-secret"
			}
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatal(err)
			}
			switched := readCredentialArtifacts(t, cache)
			if len(before) != len(switched) {
				t.Fatal("source switch changed cache artifacts")
			}
			for filename := range before {
				if _, ok := switched[filename]; !ok {
					t.Fatal("source switch changed cache identity")
				}
			}
			// Restore the original declarations; offline identity depends on neither source.
			if err := os.WriteFile(path, []byte(config), 0o600); err != nil {
				t.Fatal(err)
			}
			requests := calls.Load()
			// Offline ignores even conflicting globals, missing references, and malformed local secrets.
			t.Setenv("MOBILE_KEY", "")
			t.Setenv("WEB_KEY", "")
			t.Setenv("SENTIARY_USER_API_KEY", "conflicting-user-secret")
			t.Setenv("SENTIARY_PROJECT_API_KEY", "conflicting-project-secret")
			if err := os.WriteFile(filepath.Join(filepath.Dir(path), "sentiary.local.toml"), []byte("invalid local TOML [secret"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := app.Run(context.Background(), append(args, "--offline")); err != nil {
				t.Fatal(err)
			}
			if calls.Load() != requests {
				t.Fatal("offline contacted server")
			}
			t.Setenv("SENTIARY_USER_API_KEY", "")
			t.Setenv("SENTIARY_PROJECT_API_KEY", "")
			if err := app.Run(context.Background(), append(args, "--offline")); err != nil {
				t.Fatal(err)
			}
			if calls.Load() != requests {
				t.Fatal("credential-free offline contacted server")
			}
			if strings.Contains(stdout.String(), "secret") || strings.Contains(stderr.String(), "secret") {
				t.Fatal("credential leaked in output")
			}
			for _, content := range readCredentialArtifacts(t, cache) {
				if bytes.Contains(content, []byte("secret")) {
					t.Fatal("credential leaked in cache")
				}
			}
			for _, content := range readCredentialArtifacts(t, filepath.Join(filepath.Dir(path), ".sentiary")) {
				if bytes.Contains(content, []byte("secret")) {
					t.Fatal("credential leaked in ownership state")
				}
			}
		})
	}
}

func readCredentialArtifacts(t *testing.T, root string) map[string][]byte {
	t.Helper()
	files := map[string][]byte{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files[relative] = data
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestMissingExplicitReferenceFailsEvenWithGlobalProjectCredential(t *testing.T) {
	var calls int
	app, path, _ := pullApp(t, credentialConfig, func(*http.Request) (*http.Response, error) { calls++; return nil, nil })
	t.Setenv("SENTIARY_PROJECT_API_KEY", "project-secret")
	destination := filepath.Join(filepath.Dir(path), "generated", "mobile.xml")
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := app.Run(context.Background(), []string{"pull", "--config", path, "--output", "mobile"})
	if err == nil || !strings.Contains(err.Error(), "missing or empty") || strings.Contains(err.Error(), "secret") {
		t.Fatalf("wrong conflict: %v", err)
	}
	if calls != 0 {
		t.Fatal("conflict reached export transport")
	}
	content, readErr := os.ReadFile(destination)
	if readErr != nil || string(content) != "original" {
		t.Fatal("conflict changed destination")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), ".sentiary")); !os.IsNotExist(err) {
		t.Fatal("conflict wrote ownership state")
	}
}

func TestRejectedReferencedCredentialDoesNotFallbackAndRedactsError(t *testing.T) {
	config := strings.Replace(catalogConfig, `api_key = "private-project-key"`, `api_key = { env = "MOBILE_KEY" }`, 1)
	calls := 0
	app, path, _ := pullApp(t, config, func(request *http.Request) (*http.Response, error) {
		calls++
		if request.Header.Get("Authorization") != "Bearer rejected-project-secret" {
			t.Fatal("explicit reference was overridden")
		}
		response := textResponse(request, "rejected rejected-project-secret")
		response.StatusCode = http.StatusUnauthorized
		return response, nil
	})
	t.Setenv("MOBILE_KEY", "rejected-project-secret")
	t.Setenv("SENTIARY_USER_API_KEY", "fallback-user-secret")
	err := app.Run(context.Background(), []string{"pull", "--config", path})
	if err == nil || strings.Contains(err.Error(), "secret") || !strings.Contains(err.Error(), "[redacted]") {
		t.Fatalf("unsafe error: %v", err)
	}
	if calls != 1 {
		t.Fatal("rejected credential caused a fallback request")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), "generated")); !os.IsNotExist(err) {
		t.Fatal("authentication failure changed destinations")
	}
}
