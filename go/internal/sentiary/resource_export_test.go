package sentiary

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
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
			Capability string   `json:"capability"`
			Format     string   `json:"format"`
			Languages  []string `json:"languages"`
			Keys       []string `json:"keys"`
			Required   []string `json:"requiredKeys"`
			Plurals    []string `json:"pluralBases"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body.Capability != download.ResourceExportCapability || body.Format != "android" || !reflect.DeepEqual(body.Keys, []string{"count"}) || !reflect.DeepEqual(body.Required, body.Keys) || !reflect.DeepEqual(body.Plurals, body.Keys) {
			t.Errorf("wrong payload: %+v", body)
		}
		w.Header().Set("X-Sentiary-Export-Capability", download.ResourceExportCapability)
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

func TestResourceExportRequiresCapabilityAcknowledgment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("full catalog")) }))
	defer server.Close()
	repo := NewDownloadRepository(Config{BaseURL: server.URL}, server.Client())
	if _, err := repo.Export(context.Background(), download.Project{ID: "p", APIKey: "key"}, download.File{Format: "json", Layout: download.PerLanguage, Languages: []string{"en"}, Options: download.ExportOptions{Contexts: []string{"screen"}}}); err == nil {
		t.Fatal("accepted unconfirmed response")
	}
}
