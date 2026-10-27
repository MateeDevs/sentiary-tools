package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MateeDevs/sentiary-tools/internal/sentiary"
)

const catalogConfig = `config_version = 1
[projects.mobile]
id = "mobile-project"
api_key = "private-project-key"
languages = ["en", "cs"]
source_language = "en"
[outputs.ios]
project = "mobile"
format = "xcstrings"
layout = "combined"
path = "generated/Localizable.xcstrings"
`

func pullApp(t *testing.T, config string, transport roundTripFunc) (*App, string, *bytes.Buffer) {
	t.Helper()
	t.Setenv("SENTIARY_USER_API_KEY", "")
	t.Setenv("SENTIARY_CACHE_DIR", t.TempDir())
	directory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "sentiary.toml")
	if err := os.WriteFile(path, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	app := New(newTestAPI(transport))
	app.Downloads = sentiary.NewDownloadRepository(sentiary.Config{BaseURL: "https://example.test"}, &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if strings.HasSuffix(request.URL.Path, "/export/state") {
			return jsonResponse(request, `{"protocolVersion":1,"projectGeneration":"g","dataRevision":"1","exporterVersion":"v","languages":["en","cs"]}`), nil
		}
		response, err := transport(request)
		if err == nil && response.StatusCode == 200 {
			response.Header.Set("ETag", `W/"test"`)
			response.Header.Set("X-Sentiary-Project-Generation", "g")
			response.Header.Set("X-Sentiary-Data-Revision", "1")
			response.Header.Set("X-Sentiary-Exporter-Version", "v")
		}
		return response, err
	})})
	stdout := &bytes.Buffer{}
	app.Stdout, app.Stderr = stdout, io.Discard
	return app, path, stdout
}

func TestPullCombinedUsesOneExportRequestAndSavesExactBytes(t *testing.T) {
	calls := 0
	content := "{\"sourceLanguage\":\"en\",\"strings\":{\"hello\":\"Ahoj 👋\"}}"
	app, path, _ := pullApp(t, catalogConfig, func(request *http.Request) (*http.Response, error) {
		calls++
		if request.URL.Path != "/api/v1/batch/mobile-project/export/combined" {
			t.Fatalf("wrong endpoint: %s", request.URL.Path)
		}
		if !reflect.DeepEqual(request.URL.Query()["languageIds"], []string{"cs", "en"}) {
			t.Fatalf("wrong languages: %v", request.URL.Query())
		}
		if request.URL.Query().Get("sourceLanguageId") != "en" || request.URL.Query().Get("format") != "xcstrings" {
			t.Fatal("wrong combined options")
		}
		if request.Header.Get("Authorization") != "Bearer private-project-key" {
			t.Fatal("wrong credentials")
		}
		return textResponse(request, content), nil
	})
	if err := app.Run(context.Background(), []string{"pull", "--config", path}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(path), "generated", "Localizable.xcstrings"))
	if err != nil || string(data) != content || calls != 1 {
		t.Fatalf("incorrect saved catalog: %q, calls=%d, error=%v", data, calls, err)
	}
}

func TestInvalidCredentialFailsWithoutFallbackAndRedactsErrorBody(t *testing.T) {
	calls := 0
	app, path, _ := pullApp(t, catalogConfig, func(request *http.Request) (*http.Response, error) {
		calls++
		response := textResponse(request, "rejected private-personal-key")
		response.StatusCode = http.StatusUnauthorized
		response.Status = "401 Unauthorized"
		return response, nil
	})
	t.Setenv("SENTIARY_USER_API_KEY", "private-personal-key")
	err := app.Run(context.Background(), []string{"pull", "--config", path})
	if err == nil || strings.Contains(err.Error(), "private-personal-key") || !strings.Contains(err.Error(), "[redacted]") {
		t.Fatalf("unsafe error: %v", err)
	}
	if calls != 1 {
		t.Fatal("invalid credential triggered fallback")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), "generated")); !os.IsNotExist(err) {
		t.Fatal("failed download changed outputs")
	}
}

func TestPullDryRunContainsNoCredentialsAndWritesNoState(t *testing.T) {
	app, path, stdout := pullApp(t, catalogConfig, func(*http.Request) (*http.Response, error) { t.Fatal("dry run exported content"); return nil, nil })
	if err := app.Run(context.Background(), []string{"pull", "--config", path, "--dry-run"}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stdout.String(), "private-project-key") || !strings.Contains(stdout.String(), `"layout": "combined"`) {
		t.Fatalf("incorrect preview: %s", stdout)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), ".sentiary")); !os.IsNotExist(err) {
		t.Fatal("dry run wrote state")
	}
}

func TestPullExpandsAllLanguagesThroughExportState(t *testing.T) {
	calls := 0
	app, path, _ := pullApp(t, strings.Replace(catalogConfig, `["en", "cs"]`, `"*"`, 1), func(request *http.Request) (*http.Response, error) {
		calls++
		if !reflect.DeepEqual(request.URL.Query()["languageIds"], []string{"cs", "en"}) {
			t.Fatal("all languages not expanded")
		}
		return textResponse(request, "catalog"), nil
	})
	if err := app.Run(context.Background(), []string{"pull", "--config", path}); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("requests: %d", calls)
	}
}

func TestPullPreservesBinaryExportContent(t *testing.T) {
	content := []byte{0, 0xff, 0xd0, 0xcf, 0x50, 0x4b, 10}
	config := strings.Replace(catalogConfig, `format = "xcstrings"`, `format = "xlsx"`, 1)
	config = strings.Replace(config, `layout = "combined"`, `layout = "per_language"`, 1)
	config = strings.Replace(config, "generated/Localizable.xcstrings", "generated/{language}.xlsx", 1)
	app, path, _ := pullApp(t, config, func(request *http.Request) (*http.Response, error) {
		return textResponse(request, string(content)), nil
	})
	if err := app.Run(context.Background(), []string{"pull", "--config", path}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(path), "generated", "cs.xlsx"))
	if err != nil || !bytes.Equal(data, content) {
		t.Fatalf("binary export changed: %v", err)
	}
}

func TestInvalidCombinedConfigDoesNotCallAPIOrWriteOutputs(t *testing.T) {
	app, path, _ := pullApp(t, strings.Replace(catalogConfig, "generated/Localizable", "generated/{locale}/Localizable", 1), func(*http.Request) (*http.Response, error) {
		t.Fatal("invalid config reached API")
		return nil, nil
	})
	if err := app.Run(context.Background(), []string{"pull", "--config", path}); err == nil {
		t.Fatal("invalid config accepted")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), "generated")); !os.IsNotExist(err) {
		t.Fatal("invalid config wrote output directory")
	}
}

func TestOfflinePullRestoresAfterBuildDeletionWithoutCredentialsOrNetwork(t *testing.T) {
	app, path, _ := pullApp(t, strings.ReplaceAll(catalogConfig, "generated/", "build/"), func(request *http.Request) (*http.Response, error) {
		return textResponse(request, string([]byte{0, 255, 10})), nil
	})
	if err := app.Run(context.Background(), []string{"pull", "--config", path}); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(filepath.Dir(path), "build")); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content = bytes.ReplaceAll(content, []byte(`api_key = "private-project-key"`), nil)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	app.DownloadServer = "https://example.test"
	app.Downloads = nil
	if err := app.Run(context.Background(), []string{"pull", "--config", path, "--offline"}); err != nil {
		t.Fatal(err)
	}
	restored, err := os.ReadFile(filepath.Join(filepath.Dir(path), "build", "Localizable.xcstrings"))
	if err != nil || !bytes.Equal(restored, []byte{0, 255, 10}) {
		t.Fatalf("offline restore failed: %v", err)
	}
}

func TestDestinationFailureLeavesPublishedSetAvailableForOfflineRestore(t *testing.T) {
	app, path, _ := pullApp(t, catalogConfig, func(request *http.Request) (*http.Response, error) { return textResponse(request, "cached"), nil })
	target := filepath.Join(filepath.Dir(path), "generated", "Localizable.xcstrings")
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("manual"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := app.Run(context.Background(), []string{"pull", "--config", path}); err == nil {
		t.Fatal("overwrote unowned target")
	}
	if err := os.RemoveAll(filepath.Dir(target)); err != nil {
		t.Fatal(err)
	}
	app.DownloadServer = "https://example.test"
	app.Downloads = nil
	if err := app.Run(context.Background(), []string{"pull", "--config", path, "--offline"}); err != nil {
		t.Fatal(err)
	}
	restored, err := os.ReadFile(target)
	if err != nil || string(restored) != "cached" {
		t.Fatal("completed set was not published before target write")
	}
}

func TestPullAliasesRestoreExactBinaryAfterPathAndAliasChange(t *testing.T) {
	content := string([]byte{0, 255, 0, 13, 10})
	cfg := `config_version = 1
[projects.app]
id = "p"
api_key = "key"
languages = ["sk"]
[outputs.mobile]
project = "app"
format = "mo"
layout = "per_language"
path = "old/{language}.mo"
language_aliases = { sk = "cs" }
`
	app, path, stdout := pullApp(t, cfg, func(request *http.Request) (*http.Response, error) {
		if request.URL.Query().Get("languageId") != "cs" {
			t.Fatalf("wrong API language: %s", request.URL)
		}
		return textResponse(request, content), nil
	})
	if err := app.Run(context.Background(), []string{"pull", "--config", path, "--json"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), `"sha256"`) || strings.Contains(stdout.String(), `Downloaded`) {
		t.Fatal(stdout.String())
	}
	cfg = strings.ReplaceAll(cfg, "old/{language}", "new/{language}")
	cfg = strings.ReplaceAll(cfg, `["sk"]`, `["de"]`)
	cfg = strings.ReplaceAll(cfg, `{ sk = "cs" }`, `{ de = "cs" }`)
	if err := os.WriteFile(path, []byte(cfg), 0600); err != nil {
		t.Fatal(err)
	}
	app.Downloads = nil
	app.DownloadServer = "https://example.test"
	if err := app.Run(context.Background(), []string{"pull", "--config", path, "--offline"}); err != nil {
		t.Fatal(err)
	}
	actual, err := os.ReadFile(filepath.Join(filepath.Dir(path), "new/de.mo"))
	if err != nil || string(actual) != content {
		t.Fatalf("binary alias changed: %v %v", actual, err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), "old/sk.mo")); !os.IsNotExist(err) {
		t.Fatal("obsolete target remains")
	}
}

func TestIntegrationRequestRejectsBeforeDestinationWrites(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"protocol", `{"protocolVersion":2,"bindings":[]}`},
		{"format", `{"protocolVersion":1,"bindings":[{"output":"ios","format":"android","resourceRoot":"/tmp"}]}`},
		{"unknown field", `{"protocolVersion":1,"bindings":[],"unknown":true}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app, path, _ := pullApp(t, catalogConfig, func(*http.Request) (*http.Response, error) {
				t.Fatal("invalid integration reached export")
				return nil, nil
			})
			request := filepath.Join(filepath.Dir(path), "request.json")
			if err := os.WriteFile(request, []byte(tc.body), 0600); err != nil {
				t.Fatal(err)
			}
			if err := app.Run(context.Background(), []string{"pull", "--config", path, "--integration-request", request}); err == nil {
				t.Fatal("accepted invalid integration")
			}
			if _, err := os.Stat(filepath.Join(filepath.Dir(path), "generated")); !os.IsNotExist(err) {
				t.Fatal("wrote destinations")
			}
		})
	}
}

func TestPullProcessHelper(t *testing.T) {
	if os.Getenv("SENTIARY_PULL_HELPER") != "1" {
		return
	}
	client := sentiary.NewClient(sentiary.Config{BaseURL: os.Getenv("SENTIARY_TEST_SERVER")}, http.DefaultClient)
	app := New(client)
	app.Downloads = sentiary.NewDownloadRepository(sentiary.Config{BaseURL: os.Getenv("SENTIARY_TEST_SERVER")}, http.DefaultClient)
	if err := app.Run(context.Background(), []string{"pull", "--config", os.Getenv("SENTIARY_TEST_CONFIG")}); err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentPullProcessesSerializeFetchingAndWriting(t *testing.T) {
	var states atomic.Int32
	firstStarted := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/state") {
			if states.Add(1) == 1 {
				close(firstStarted)
				<-release
			}
			_, _ = io.WriteString(w, `{"protocolVersion":1,"projectGeneration":"g","dataRevision":"1","exporterVersion":"v","languages":["en","cs"]}`)
			return
		}
		w.Header().Set("ETag", `W/"test"`)
		w.Header().Set("X-Sentiary-Project-Generation", "g")
		w.Header().Set("X-Sentiary-Data-Revision", "1")
		w.Header().Set("X-Sentiary-Exporter-Version", "v")
		_, _ = io.WriteString(w, "complete catalog")
	}))
	defer server.Close()
	root := t.TempDir()
	path := filepath.Join(root, "sentiary.toml")
	if err := os.WriteFile(path, []byte(catalogConfig), 0600); err != nil {
		t.Fatal(err)
	}
	command := func() *exec.Cmd {
		child := exec.Command(os.Args[0], "-test.run=^TestPullProcessHelper$")
		child.Env = append(os.Environ(), "SENTIARY_PULL_HELPER=1", "SENTIARY_TEST_CONFIG="+path, "SENTIARY_TEST_SERVER="+server.URL, "SENTIARY_CACHE_DIR="+filepath.Join(root, "cache"), "SENTIARY_USER_API_KEY=")
		child.Stdout, child.Stderr = io.Discard, os.Stderr
		return child
	}
	first := command()
	if err := first.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = first.Process.Kill() }()
	select {
	case <-firstStarted:
	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("first pull never started")
	}
	second := command()
	if err := second.Start(); err != nil {
		close(release)
		t.Fatal(err)
	}
	defer func() { _ = second.Process.Kill() }()
	time.Sleep(150 * time.Millisecond)
	if states.Load() != 1 {
		close(release)
		t.Fatal("second pull fetched before first completed")
	}
	close(release)
	if err := first.Wait(); err != nil {
		t.Fatal(err)
	}
	if err := second.Wait(); err != nil {
		t.Fatal(err)
	}
	if states.Load() != 2 {
		t.Fatalf("state checks: %d", states.Load())
	}
	data, err := os.ReadFile(filepath.Join(root, "generated", "Localizable.xcstrings"))
	if err != nil || string(data) != "complete catalog" {
		t.Fatalf("incomplete destination: %v", err)
	}
}

func TestIntegrationResourceRootRejectsBeforeTargetWrites(t *testing.T) {
	cfg := strings.Replace(catalogConfig, `format = "xcstrings"`, `format = "compose"`, 1)
	cfg = strings.Replace(cfg, `layout = "combined"`, `layout = "per_language"`, 1)
	cfg = strings.Replace(cfg, `generated/Localizable.xcstrings`, `generated/{language}/strings.xml`, 1)
	app, path, _ := pullApp(t, cfg, func(request *http.Request) (*http.Response, error) { return textResponse(request, "bytes"), nil })
	requestPath := filepath.Join(filepath.Dir(path), "request.json")
	body, _ := json.Marshal(integrationRequest{ProtocolVersion: 1, Bindings: []integrationBinding{{Output: "ios", Format: "compose", ResourceRoot: filepath.Join(filepath.Dir(path), "different")}}})
	if err := os.WriteFile(requestPath, body, 0600); err != nil {
		t.Fatal(err)
	}
	err := app.Run(context.Background(), []string{"pull", "--config", path, "--integration-request", requestPath, "--json"})
	if err == nil || !strings.Contains(err.Error(), "outside integration resourceRoot") {
		t.Fatalf("invalid root accepted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), "generated")); !os.IsNotExist(err) {
		t.Fatal("wrote rejected targets")
	}
}
