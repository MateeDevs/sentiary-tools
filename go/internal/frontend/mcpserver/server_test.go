package mcpserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/MateeDevs/sentiary-tools/go/internal/sentiary"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestServerListsAllTools(t *testing.T) {
	server := New(sentiary.Config{}, http.DefaultClient, "test")
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "test"}, nil)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(context.Background(), serverTransport, nil)
	if err != nil {
		t.Fatalf("Connect server: %v", err)
	}
	defer serverSession.Close()

	clientSession, err := client.Connect(context.Background(), clientTransport, nil)
	if err != nil {
		t.Fatalf("Connect client: %v", err)
	}
	defer clientSession.Close()

	result, err := clientSession.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("List tools: %v", err)
	}

	want := map[string]bool{
		"list_projects":                     true,
		"get_project":                       true,
		"create_project":                    true,
		"edit_project":                      true,
		"remove_project":                    true,
		"list_supported_languages":          true,
		"list_project_languages":            true,
		"add_project_language":              true,
		"remove_project_language":           true,
		"list_project_members":              true,
		"set_project_member_role":           true,
		"remove_project_member":             true,
		"list_invitations":                  true,
		"create_invitation":                 true,
		"accept_invitation":                 true,
		"decline_invitation":                true,
		"list_project_invitations":          true,
		"remove_project_invitation":         true,
		"list_project_strings":              true,
		"search_project_strings":            true,
		"get_project_string":                true,
		"get_project_string_by_key":         true,
		"add_project_string":                true,
		"edit_project_string":               true,
		"remove_project_string":             true,
		"set_project_string_translation":    true,
		"remove_project_string_translation": true,
		"get_project_strings_info":          true,
		"export_project_strings":            true,
		"import_project_strings":            true,
	}

	if len(result.Tools) != len(want) {
		t.Fatalf("Tool count = %d, want %d", len(result.Tools), len(want))
	}
	for _, tool := range result.Tools {
		if !want[tool.Name] {
			t.Errorf("Unexpected tool %q", tool.Name)
		}
		if tool.Name == "get_project_string" {
			assertObjectPropertySchema(t, tool.OutputSchema, "statistics")
		}
	}
}

func TestSearchToolReturnsTranslationMatches(t *testing.T) {
	var searchQuery string
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		searchQuery = request.URL.Query().Get("query")
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
				"items":[{
					"id":"term-id",
					"name":"welcome_message",
					"translations":[{"languageId":"cs_CZ","text":"Ahoj","updated":""}],
					"updated":""
				}],
				"pagination":{"size":25}
			}`)),
			Request: request,
		}, nil
	})}
	server := New(sentiary.Config{BaseURL: "https://api.test", APIKey: "test-key"}, httpClient, "test")
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "test"}, nil)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(context.Background(), serverTransport, nil)
	if err != nil {
		t.Fatalf("Connect server: %v", err)
	}
	defer serverSession.Close()
	clientSession, err := client.Connect(context.Background(), clientTransport, nil)
	if err != nil {
		t.Fatalf("Connect client: %v", err)
	}
	defer clientSession.Close()

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "search_project_strings",
		Arguments: map[string]any{
			"projectId": "project-id",
			"query":     "Ahoj",
		},
	})
	if err != nil {
		t.Fatalf("Call search tool: %v", err)
	}
	if result.IsError {
		t.Fatalf("Search tool returned an error: %v", result.Content)
	}
	if searchQuery != "Ahoj" {
		t.Fatalf("Search query = %q", searchQuery)
	}
	content, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("Marshal search result: %v", err)
	}
	if !strings.Contains(string(content), "Ahoj") {
		t.Fatalf("Search result does not contain the translation: %s", content)
	}
}

func TestSetProjectMemberRoleToolUsesManagementAPI(t *testing.T) {
	var requestPath string
	var requestBody string
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requestPath = request.URL.Path
		body, err := io.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}
		requestBody = string(body)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`[]`)),
			Request:    request,
		}, nil
	})}
	server := New(sentiary.Config{BaseURL: "https://api.test", APIKey: "test-key"}, httpClient, "test")
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "test"}, nil)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(context.Background(), serverTransport, nil)
	if err != nil {
		t.Fatalf("Connect server: %v", err)
	}
	defer serverSession.Close()
	clientSession, err := client.Connect(context.Background(), clientTransport, nil)
	if err != nil {
		t.Fatalf("Connect client: %v", err)
	}
	defer clientSession.Close()

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "set_project_member_role",
		Arguments: map[string]any{
			"projectId": "p1",
			"userId":    "u1",
			"role":      "ADMINISTRATOR",
		},
	})
	if err != nil {
		t.Fatalf("Call member role tool: %v", err)
	}
	if result.IsError {
		t.Fatalf("Member role tool returned an error: %v", result.Content)
	}
	if requestPath != "/api/v1/project/p1/member/u1" {
		t.Fatalf("Request path = %q", requestPath)
	}
	if !strings.Contains(requestBody, `"role":"administrator"`) {
		t.Fatalf("Request body = %q", requestBody)
	}
}

func assertObjectPropertySchema(t *testing.T, schema any, propertyName string) {
	t.Helper()

	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("Marshal output schema: %v", err)
	}

	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("Unmarshal output schema: %v", err)
	}

	properties, ok := document["properties"].(map[string]any)
	if !ok {
		t.Fatalf("Output schema has no properties: %s", data)
	}
	if _, isBoolean := properties[propertyName].(bool); isBoolean {
		t.Fatalf("Property %q has a Boolean schema: %s", propertyName, data)
	}
}
