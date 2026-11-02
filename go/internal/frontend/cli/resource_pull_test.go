package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestResourceValidationFailurePreservesEveryDestination(t *testing.T) {
	config := `config_version = 1
[projects.mobile]
id = "mobile-project"
api_key = "private-project-key"
languages = ["cs", "en"]
[outputs.mobile]
project = "mobile"
format = "compose"
path = "generated/{language}.xml"
keys_file = "keys.txt"
required_keys_file = "required.txt"
plural_bases_file = "plurals.txt"
`
	fail := false
	app, path, _ := pullApp(t, config, func(request *http.Request) (*http.Response, error) {
		var body struct {
			Languages []string `json:"languages"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if fail && body.Languages[0] == "en" {
			response := textResponse(request, "Required plural other is empty")
			response.StatusCode = http.StatusBadRequest
			return response, nil
		}
		content := `<resources><plurals name="count"><item quantity="other">%1$d trucks</item></plurals></resources>`
		if fail {
			content = "changed bytes that must never reach the destination"
		}
		response := textResponse(request, content)
		response.Header.Set("X-Sentiary-Export-Capability", "resource-export-v1")
		return response, nil
	})
	for _, name := range []string{"keys.txt", "required.txt", "plurals.txt"} {
		if err := os.WriteFile(filepath.Join(filepath.Dir(path), name), []byte("count\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	ctx := context.Background()
	args := []string{"pull", "--config", path}
	if err := app.Run(ctx, args); err != nil {
		t.Fatal(err)
	}
	original := map[string]string{}
	for _, language := range []string{"cs", "en"} {
		filename := filepath.Join(filepath.Dir(path), "generated", language+".xml")
		content, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		original[filename] = string(content)
	}
	fail = true
	if err := app.Run(ctx, args); err == nil {
		t.Fatal("accepted required translation failure")
	}
	for filename, expected := range original {
		content, err := os.ReadFile(filename)
		if err != nil || string(content) != expected {
			t.Fatalf("destination changed after failed pull: %s (%v)", filename, err)
		}
	}
	if err := app.Run(ctx, append(args, "--offline")); err != nil {
		t.Fatal(err)
	}
	for filename, expected := range original {
		content, err := os.ReadFile(filename)
		if err != nil || string(content) != expected {
			t.Fatalf("offline bytes changed: %s (%v)", filename, err)
		}
	}
}
