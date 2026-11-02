package sentiary

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/MateeDevs/sentiary-tools/go/internal/download"
)

func TestResourceExportPostsValuesAndPreservesBytes(t *testing.T) {
	content := []byte{0, 255, 10}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/batch/project/export/resources" || r.URL.RawQuery != "" || r.Header.Get("Content-Type") != "application/json" || r.Header.Get("If-None-Match") != `W/"old"` {
			t.Errorf("incorrect request: %+v", r)
		}
		var body struct {
			IncludeEmpty bool     `json:"includeEmptyTranslations"`
			Format       string   `json:"format"`
			Languages    []string `json:"languages"`
			Keys         []string `json:"keys"`
			Required     []string `json:"requiredKeys"`
			Plurals      []string `json:"pluralBases"`
		}
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			t.Error(err)
		}
		if body.Format != "android" || !reflect.DeepEqual(body.Keys, []string{"count"}) || !reflect.DeepEqual(body.Required, body.Keys) || !reflect.DeepEqual(body.Plurals, body.Keys) {
			t.Errorf("wrong payload: %+v", body)
		}
		_, _ = w.Write(content)
	}))
	defer server.Close()
	keys := []string{"count"}
	repo := NewDownloadRepository(Config{BaseURL: server.URL}, server.Client())
	response, err := repo.ConditionalExport(context.Background(), download.Project{ID: "project", APIKey: "key"}, download.File{
		Format: "android", Layout: download.PerLanguage, Languages: []string{"en"},
		Options: download.ExportOptions{Keys: &keys, RequiredKeys: keys, PluralBases: keys},
	}, `W/"old"`)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(content, response.Content) {
		t.Fatal("server bytes changed")
	}
}

func TestResourceExportFailureDoesNotFallback(t *testing.T) {
	for _, test := range []struct {
		name string
		file download.File
	}{
		{"default", download.File{Format: "json", Layout: download.PerLanguage, Languages: []string{"en"}}},
		{"extended", download.File{Format: "json", Layout: download.PerLanguage, Languages: []string{"en"}, Options: download.ExportOptions{Contexts: []string{"screen"}}}},
		{"combined", download.File{Format: "xcstrings", Layout: download.Combined, Languages: []string{"en", "cs"}, SourceLanguage: "en"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != http.MethodPost || r.URL.Path != "/api/v1/batch/p/export/resources" {
					t.Error("unexpected fallback")
				}
				http.Error(w, "resource export endpoint required", http.StatusNotFound)
			}))
			defer server.Close()
			repo := NewDownloadRepository(Config{BaseURL: server.URL}, server.Client())
			_, err := repo.Export(context.Background(), download.Project{ID: "p", APIKey: "key"}, test.file)
			if err == nil || !strings.Contains(err.Error(), "requires backend POST resource export") {
				t.Fatalf("unclear unsupported endpoint error: %v", err)
			}
			if calls != 1 {
				t.Fatal("retried failed export")
			}
		})
	}
}

func TestDefaultConditionalExportUsesPOSTAndRealRevisionHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/batch/p/export/resources" || r.URL.RawQuery != "" || r.Header.Get("If-None-Match") != `W/"stable"` {
			t.Error("wrong conditional export transport")
		}
		var body struct {
			Format       string   `json:"format"`
			Languages    []string `json:"languages"`
			IncludeEmpty bool     `json:"includeEmptyTranslations"`
		}
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			t.Error(err)
		}
		if body.Format != "xlsx" || !reflect.DeepEqual(body.Languages, []string{"cs"}) || !body.IncludeEmpty {
			t.Errorf("wrong default payload: %+v", body)
		}
		w.Header().Set("ETag", `W/"stable"`)
		w.Header().Set("X-Sentiary-Project-Generation", "real-generation")
		w.Header().Set("X-Sentiary-Data-Revision", "9007199254740993")
		w.Header().Set("X-Sentiary-Exporter-Version", "real-exporter")
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()
	repo := NewDownloadRepository(Config{BaseURL: server.URL}, server.Client())
	response, err := repo.ConditionalExport(context.Background(), download.Project{ID: "p", APIKey: "secret"}, download.File{
		Format: "xlsx", Layout: download.PerLanguage, Languages: []string{"cs"}, IncludeEmpty: true,
	}, `W/"stable"`)
	if err != nil {
		t.Fatal(err)
	}
	if !response.NotModified || response.ETag != `W/"stable"` || len(response.Content) != 0 {
		t.Fatal("incorrect 304 response")
	}
	if !response.State.SameRevision(download.State{ProjectGeneration: "real-generation", DataRevision: "9007199254740993", ExporterVersion: "real-exporter"}) {
		t.Fatal("lost real export state")
	}
}
