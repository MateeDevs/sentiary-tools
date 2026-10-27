package cli

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MateeDevs/sentiary-tools/go/internal/sentiary"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestExportCommandWritesContent(t *testing.T) {
	var requestURL *url.URL
	var authorization string
	api := newTestAPI(func(request *http.Request) (*http.Response, error) {
		requestURL = request.URL
		authorization = request.Header.Get("Authorization")
		return textResponse(request, `{"hello":"Ahoj"}`), nil
	})

	stdout := &bytes.Buffer{}
	app := New(api)
	app.Stdout = stdout
	app.Stderr = io.Discard

	err := app.Run(context.Background(), []string{
		"export",
		"--project-id=asdf",
		"--format=json",
		"--language=cs-CZ",
	})
	if err != nil {
		t.Fatalf("Run export: %v", err)
	}
	if stdout.String() != `{"hello":"Ahoj"}` {
		t.Fatalf("Export output = %q", stdout.String())
	}
	if requestURL.Path != "/api/v1/batch/asdf/export" {
		t.Fatalf("Export path = %q", requestURL.Path)
	}
	if got := requestURL.Query().Get("format"); got != "json" {
		t.Fatalf("Export format = %q", got)
	}
	if got := requestURL.Query().Get("languageId"); got != "cs-CZ" {
		t.Fatalf("Export language = %q", got)
	}
	if authorization != "Bearer test-key" {
		t.Fatalf("Authorization = %q", authorization)
	}
}

func TestExportCommandWritesFile(t *testing.T) {
	api := newTestAPI(func(request *http.Request) (*http.Response, error) {
		return textResponse(request, "exported content"), nil
	})
	app := New(api)
	app.Stdout = io.Discard
	app.Stderr = io.Discard
	outputPath := filepath.Join(t.TempDir(), "strings.json")

	err := app.Run(context.Background(), []string{
		"export",
		"--project-id=asdf",
		"--format=json",
		"--language=cs-CZ",
		"--output=" + outputPath,
	})
	if err != nil {
		t.Fatalf("Run export: %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Read export: %v", err)
	}
	if string(content) != "exported content" {
		t.Fatalf("Export file = %q", content)
	}
}

func TestImportCommandReadsStandardInput(t *testing.T) {
	var importedContent string
	var requestURL *url.URL
	api := newTestAPI(func(request *http.Request) (*http.Response, error) {
		requestURL = request.URL
		content, err := io.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}
		importedContent = string(content)
		return jsonResponse(request, `{"totalImportedCount":1,"skipped":[],"importLanguage":"cs-CZ"}`), nil
	})

	stdout := &bytes.Buffer{}
	app := New(api)
	app.Stdin = strings.NewReader(`{"hello":"Ahoj"}`)
	app.Stdout = stdout
	app.Stderr = io.Discard

	err := app.Run(context.Background(), []string{
		"import",
		"--project-id=asdf",
		"--format=json",
		"--language=cs-CZ",
		"--override-existing",
	})
	if err != nil {
		t.Fatalf("Run import: %v", err)
	}
	if importedContent != `{"hello":"Ahoj"}` {
		t.Fatalf("Import content = %q", importedContent)
	}
	if got := requestURL.Query().Get("overrideExisting"); got != "true" {
		t.Fatalf("Override existing = %q", got)
	}
	if !strings.Contains(stdout.String(), `"totalImportedCount": 1`) {
		t.Fatalf("Import output = %q", stdout.String())
	}
}

func TestSearchCommandUsesGeneralQuery(t *testing.T) {
	var requestQuery string
	api := newTestAPI(func(request *http.Request) (*http.Response, error) {
		requestQuery = request.URL.Query().Get("query")
		return jsonResponse(request, `{
			"items":[{
				"id":"term-id",
				"name":"welcome_message",
				"translations":[{"languageId":"cs-CZ","text":"Ahoj","updated":""}],
				"updated":""
			}],
			"pagination":{"size":25}
		}`), nil
	})
	stdout := &bytes.Buffer{}
	app := New(api)
	app.Stdout = stdout
	app.Stderr = io.Discard

	err := app.Run(context.Background(), []string{
		"search",
		"--project-id=asdf",
		"--query=Ahoj",
	})
	if err != nil {
		t.Fatalf("Run search: %v", err)
	}
	if requestQuery != "Ahoj" {
		t.Fatalf("Search query = %q", requestQuery)
	}
	if !strings.Contains(stdout.String(), `"text": "Ahoj"`) {
		t.Fatalf("Search output = %q", stdout.String())
	}
}

func TestManagementCommandsUseExpectedRoutes(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		method   string
		path     string
		query    string
		status   int
		response string
	}{
		{name: "list projects", args: []string{"list-projects"}, method: http.MethodGet, path: "/api/v1/project", response: `[]`},
		{name: "get project", args: []string{"get-project", "--project-id=p1"}, method: http.MethodGet, path: "/api/v1/project/p1", response: cliProjectJSON},
		{name: "create project", args: []string{"create-project", "--name=App"}, method: http.MethodPost, path: "/api/v1/project", response: `{"id":"p1","name":"App"}`},
		{name: "edit project", args: []string{"edit-project", "--project-id=p1", "--name=New"}, method: http.MethodPost, path: "/api/v1/project/p1", response: cliProjectJSON},
		{name: "remove project", args: []string{"remove-project", "--project-id=p1"}, method: http.MethodDelete, path: "/api/v1/project/p1", status: http.StatusNoContent},
		{name: "list languages", args: []string{"list-languages", "--in-language=cs-CZ"}, method: http.MethodGet, path: "/api/v1/language", query: "inLanguage=cs-CZ", response: `[]`},
		{name: "list project languages", args: []string{"list-project-languages", "--project-id=p1"}, method: http.MethodGet, path: "/api/v1/project/p1/language", response: `[]`},
		{name: "add project language", args: []string{"add-project-language", "--project-id=p1", "--language=cs-CZ"}, method: http.MethodPost, path: "/api/v1/project/p1/language/cs-CZ", response: `[]`},
		{name: "remove project language", args: []string{"remove-project-language", "--project-id=p1", "--language=cs-CZ"}, method: http.MethodDelete, path: "/api/v1/project/p1/language/cs-CZ", response: `[]`},
		{name: "list project members", args: []string{"list-project-members", "--project-id=p1"}, method: http.MethodGet, path: "/api/v1/project/p1/member", response: `[]`},
		{name: "set project member role", args: []string{"set-project-member-role", "--project-id=p1", "--user-id=u1", "--role=translator"}, method: http.MethodPost, path: "/api/v1/project/p1/member/u1", response: `[]`},
		{name: "remove project member", args: []string{"remove-project-member", "--project-id=p1", "--user-id=u1"}, method: http.MethodDelete, path: "/api/v1/project/p1/member/u1", response: `[]`},
		{name: "list invitations", args: []string{"list-invitations"}, method: http.MethodGet, path: "/api/v1/invite", response: `[]`},
		{name: "create invitation", args: []string{"create-invitation", "--project-id=p1", "--user-email=user@example.com", "--role=developer"}, method: http.MethodPost, path: "/api/v1/invite", status: http.StatusNoContent},
		{name: "accept invitation", args: []string{"accept-invitation", "--invitation-id=i1"}, method: http.MethodPost, path: "/api/v1/invite/accept", query: "inviteId=i1", status: http.StatusNoContent},
		{name: "decline invitation", args: []string{"decline-invitation", "--invitation-id=i1"}, method: http.MethodPost, path: "/api/v1/invite/decline", query: "inviteId=i1", status: http.StatusNoContent},
		{name: "list project invitations", args: []string{"list-project-invitations", "--project-id=p1"}, method: http.MethodGet, path: "/api/v1/project/p1/invite", response: `[]`},
		{name: "remove project invitation", args: []string{"remove-project-invitation", "--project-id=p1", "--invitation-id=i1"}, method: http.MethodDelete, path: "/api/v1/project/p1/invite/i1", status: http.StatusNoContent},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			api := newTestAPI(func(request *http.Request) (*http.Response, error) {
				if request.Method != test.method {
					t.Errorf("Method = %q, want %q", request.Method, test.method)
				}
				if request.URL.Path != test.path {
					t.Errorf("Path = %q, want %q", request.URL.Path, test.path)
				}
				if request.URL.RawQuery != test.query {
					t.Errorf("Query = %q, want %q", request.URL.RawQuery, test.query)
				}
				status := test.status
				if status == 0 {
					status = http.StatusOK
				}
				return &http.Response{
					StatusCode: status,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(test.response)),
					Request:    request,
				}, nil
			})
			app := New(api)
			app.Stdout = io.Discard
			app.Stderr = io.Discard
			if err := app.Run(context.Background(), test.args); err != nil {
				t.Fatalf("Run command: %v", err)
			}
		})
	}
}

func TestCommandHelpDoesNotCallAPI(t *testing.T) {
	stderr := &bytes.Buffer{}
	app := New(nil)
	app.Stdout = io.Discard
	app.Stderr = stderr

	if err := app.Run(context.Background(), []string{"export", "--help"}); err != nil {
		t.Fatalf("Run help: %v", err)
	}
	if !strings.Contains(stderr.String(), "-project-id") {
		t.Fatalf("Help output = %q", stderr.String())
	}
}

func TestExportRequiresFormatAndLanguage(t *testing.T) {
	app := New(nil)
	app.Stdout = io.Discard
	app.Stderr = io.Discard

	err := app.Run(context.Background(), []string{"export"})
	if !IsUsageError(err) {
		t.Fatalf("Run export error = %v, want a usage error", err)
	}
}

func newTestAPI(roundTrip roundTripFunc) *sentiary.Client {
	return sentiary.NewClient(sentiary.Config{
		BaseURL: "https://api.test",
		APIKey:  "test-key",
	}, &http.Client{Transport: roundTrip})
}

func textResponse(request *http.Request, content string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       io.NopCloser(strings.NewReader(content)),
		Request:    request,
	}
}

func jsonResponse(request *http.Request, content string) *http.Response {
	response := textResponse(request, content)
	response.Header.Set("Content-Type", "application/json")
	return response
}

const cliProjectJSON = `{
	"id":"p1",
	"name":"App",
	"membership":{"role":"administrator","permissions":[]}
}`
