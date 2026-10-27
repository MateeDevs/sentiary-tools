package hosted

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MateeDevs/sentiary-tools/internal/sentiary"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestAPIKeyFromRequest(t *testing.T) {
	testCases := []struct {
		name    string
		headers map[string]string
		want    string
	}{
		{name: "no headers", want: ""},
		{name: "dedicated header", headers: map[string]string{
			"X-Sentiary-User-Api-Key": " dedicated ",
			"Authorization":           "Bearer other",
		}, want: "dedicated"},
		{name: "bearer", headers: map[string]string{"Authorization": "Bearer abc123"}, want: "abc123"},
		{name: "lowercase bearer", headers: map[string]string{"Authorization": "bearer abc123"}, want: "abc123"},
		{name: "ribbon", headers: map[string]string{"Authorization": "Ribbon abc123"}, want: "abc123"},
		{name: "extra spaces", headers: map[string]string{"Authorization": "Bearer   abc123"}, want: "abc123"},
		{name: "bare token", headers: map[string]string{"Authorization": "abc123"}, want: "abc123"},
		{name: "empty bearer", headers: map[string]string{"Authorization": "Bearer   "}, want: ""},
		{name: "empty ribbon", headers: map[string]string{"Authorization": "ribbon"}, want: ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", nil)
			for name, value := range testCase.headers {
				request.Header.Set(name, value)
			}

			if got := apiKeyFromRequest(request); got != testCase.want {
				t.Fatalf("apiKeyFromRequest() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestAPIKeyFromNilRequest(t *testing.T) {
	if got := apiKeyFromRequest(nil); got != "" {
		t.Fatalf("apiKeyFromRequest(nil) = %q, want an empty key", got)
	}
	if got := apiKeyFromRequest(&http.Request{}); got != "" {
		t.Fatalf("apiKeyFromRequest(request) = %q, want an empty key", got)
	}
}

func TestHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	NewHandler(sentiary.Config{}, http.DefaultClient, "test").ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Health status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Body.String() != "ok" {
		t.Fatalf("Health body = %q, want %q", response.Body.String(), "ok")
	}
}

func TestMCPRequiresAPIKey(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	response := httptest.NewRecorder()

	NewHandler(sentiary.Config{}, http.DefaultClient, "test").ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("MCP status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestRequestAPIKeyReachesSentiaryAPI(t *testing.T) {
	apiKey := make(chan string, 1)
	apiClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		apiKey <- request.Header.Get("Authorization")
		body, err := json.Marshal(sentiary.ProjectBatchInfo{
			ID:        "project-id",
			Name:      "Project",
			Languages: []string{"en"},
		})
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(string(body))),
			Request:    request,
		}, nil
	})}

	config := sentiary.Config{BaseURL: "https://api.test"}
	mcpHandler := NewHandler(config, apiClient, "test")

	currentAPIKey := "session-key"
	authenticatedClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		request = request.Clone(request.Context())
		request.Header.Set("Authorization", "Bearer "+currentAPIKey)
		response := httptest.NewRecorder()
		mcpHandler.ServeHTTP(response, request)
		return response.Result(), nil
	})}

	transport := &mcp.StreamableClientTransport{
		Endpoint:             "https://mcp.test/",
		HTTPClient:           authenticatedClient,
		DisableStandaloneSSE: true,
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "test"}, nil)
	contextWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session, err := client.Connect(contextWithTimeout, transport, nil)
	if err != nil {
		t.Fatalf("Connect hosted MCP: %v", err)
	}
	defer session.Close()

	result, err := session.CallTool(contextWithTimeout, &mcp.CallToolParams{
		Name:      "get_project_strings_info",
		Arguments: sentiary.ProjectInput{ProjectID: "project-id"},
	})
	if err != nil {
		t.Fatalf("Call tool: %v", err)
	}
	if result.IsError {
		t.Fatalf("Tool returned an error: %v", result.Content)
	}

	select {
	case got := <-apiKey:
		if got != "Bearer session-key" {
			t.Fatalf("API authorization = %q, want %q", got, "Bearer session-key")
		}
	case <-contextWithTimeout.Done():
		t.Fatal("The Sentiary API did not receive a request")
	}

	currentAPIKey = "another-key"
	if _, err := session.ListTools(contextWithTimeout, nil); err == nil {
		t.Fatal("The hosted MCP accepted a different key for an existing session")
	}
}
