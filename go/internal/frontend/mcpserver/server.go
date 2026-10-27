package mcpserver

import (
	"context"
	"net/http"

	"github.com/MateeDevs/sentiary-tools/go/internal/sentiary"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func New(config sentiary.Config, httpClient *http.Client, version string) *mcp.Server {
	return NewServer(sentiary.NewClient(config, httpClient), version)
}

func NewServer(api sentiary.API, version string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "sentiary",
		Title:   "Sentiary Strings",
		Version: version,
	}, nil)

	registerTools(server, api)
	return server
}

func registerTools(server *mcp.Server, api sentiary.API) {
	registerManagementTools(server, api)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_project_strings",
		Title:       "List project strings",
		Description: "Lists project strings with pagination and optional translation languages. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.ListStringsInput) (*mcp.CallToolResult, sentiary.Paging, error) {
		output, err := api.ListStrings(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_project_strings",
		Title:       "Search project strings",
		Description: "Searches string keys, translations, descriptions, and context. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.SearchStringsInput) (*mcp.CallToolResult, sentiary.Paging, error) {
		output, err := api.SearchStrings(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_project_string",
		Title:       "Get project string",
		Description: "Gets one string by its ID, including translations. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.GetStringInput) (*mcp.CallToolResult, sentiary.Term, error) {
		output, err := api.GetString(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_project_string_by_key",
		Title:       "Get project string by key",
		Description: "Gets one string by its exact key, including the requested translations. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.GetStringByKeyInput) (*mcp.CallToolResult, sentiary.Term, error) {
		output, err := api.GetStringByKey(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_project_string",
		Title:       "Add project string",
		Description: "Creates one string and can set its first translation. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.AddStringInput) (*mcp.CallToolResult, sentiary.Term, error) {
		output, err := api.AddString(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "edit_project_string",
		Title:       "Edit project string",
		Description: "Edits a string name, description, and context. Clear nullable fields with their clear options. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.EditStringInput) (*mcp.CallToolResult, sentiary.Term, error) {
		output, err := api.EditString(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "remove_project_string",
		Title:       "Remove project string",
		Description: "Removes one string and its translations. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.RemoveStringInput) (*mcp.CallToolResult, sentiary.DeleteOutput, error) {
		output, err := api.RemoveString(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_project_string_translation",
		Title:       "Set project string translation",
		Description: "Creates or updates one string translation. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.SetTranslationInput) (*mcp.CallToolResult, sentiary.Term, error) {
		output, err := api.SetStringTranslation(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "remove_project_string_translation",
		Title:       "Remove project string translation",
		Description: "Removes one translation from a string. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.RemoveTranslationInput) (*mcp.CallToolResult, sentiary.Term, error) {
		output, err := api.RemoveStringTranslation(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_project_strings_info",
		Title:       "Get project strings info",
		Description: "Gets project string metadata, languages, and the last modification date. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.ProjectInput) (*mcp.CallToolResult, sentiary.ProjectBatchInfo, error) {
		output, err := api.GetProjectBatchInfo(ctx, input.ProjectID)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "export_project_strings",
		Title:       "Export project strings",
		Description: "Exports one language in JSON, Android, Apple, or Compose format. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.ExportInput) (*mcp.CallToolResult, sentiary.ExportOutput, error) {
		output, err := api.ExportStrings(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "import_project_strings",
		Title:       "Import project strings",
		Description: "Imports one language in JSON, Android, Apple, or Compose format. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.ImportInput) (*mcp.CallToolResult, sentiary.ImportResult, error) {
		output, err := api.ImportStrings(ctx, input)
		return nil, output, err
	})
}
