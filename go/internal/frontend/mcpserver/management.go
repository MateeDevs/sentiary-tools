package mcpserver

import (
	"context"

	"github.com/MateeDevs/sentiary-tools/go/internal/sentiary"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerManagementTools(server *mcp.Server, api sentiary.API) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_projects",
		Title:       "List projects",
		Description: "Lists projects for the authenticated user. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, sentiary.ProjectsOutput, error) {
		output, err := api.ListProjects(ctx)
		return nil, sentiary.ProjectsOutput{Projects: output}, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_project",
		Title:       "Get project",
		Description: "Gets a project and the authenticated user's membership. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.ProjectInput) (*mcp.CallToolResult, sentiary.ProjectWithMembership, error) {
		output, err := api.GetProject(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_project",
		Title:       "Create project",
		Description: "Creates a project. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.CreateProjectInput) (*mcp.CallToolResult, sentiary.Project, error) {
		output, err := api.CreateProject(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "edit_project",
		Title:       "Edit project",
		Description: "Changes a project name. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.EditProjectInput) (*mcp.CallToolResult, sentiary.ProjectWithMembership, error) {
		output, err := api.EditProject(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "remove_project",
		Title:       "Remove project",
		Description: "Removes a project and its data. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.ProjectInput) (*mcp.CallToolResult, sentiary.DeleteOutput, error) {
		output, err := api.RemoveProject(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_supported_languages",
		Title:       "List supported languages",
		Description: "Lists all languages that Sentiary supports. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.ListLanguagesInput) (*mcp.CallToolResult, sentiary.LanguagesOutput, error) {
		output, err := api.ListLanguages(ctx, input)
		return nil, sentiary.LanguagesOutput{Languages: output}, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_project_languages",
		Title:       "List project languages",
		Description: "Lists the languages in a project. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.ListProjectLanguagesInput) (*mcp.CallToolResult, sentiary.LanguagesOutput, error) {
		output, err := api.ListProjectLanguages(ctx, input)
		return nil, sentiary.LanguagesOutput{Languages: output}, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_project_language",
		Title:       "Add project language",
		Description: "Adds a language to a project. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.ProjectLanguageInput) (*mcp.CallToolResult, sentiary.LanguagesOutput, error) {
		output, err := api.AddProjectLanguage(ctx, input)
		return nil, sentiary.LanguagesOutput{Languages: output}, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "remove_project_language",
		Title:       "Remove project language",
		Description: "Removes a project language and its translations. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.ProjectLanguageInput) (*mcp.CallToolResult, sentiary.LanguagesOutput, error) {
		output, err := api.RemoveProjectLanguage(ctx, input)
		return nil, sentiary.LanguagesOutput{Languages: output}, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_project_members",
		Title:       "List project members",
		Description: "Lists project members and roles. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.ProjectInput) (*mcp.CallToolResult, sentiary.ProjectMembersOutput, error) {
		output, err := api.ListProjectMembers(ctx, input)
		return nil, sentiary.ProjectMembersOutput{Members: output}, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_project_member_role",
		Title:       "Set project member role",
		Description: "Sets a project member role to administrator, developer, or translator. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.SetProjectMemberRoleInput) (*mcp.CallToolResult, sentiary.ProjectMembersOutput, error) {
		output, err := api.SetProjectMemberRole(ctx, input)
		return nil, sentiary.ProjectMembersOutput{Members: output}, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "remove_project_member",
		Title:       "Remove project member",
		Description: "Removes a member from a project. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.ProjectMemberInput) (*mcp.CallToolResult, sentiary.ProjectMembersOutput, error) {
		output, err := api.RemoveProjectMember(ctx, input)
		return nil, sentiary.ProjectMembersOutput{Members: output}, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_invitations",
		Title:       "List invitations",
		Description: "Lists invitations for the authenticated user. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, sentiary.InvitationsOutput, error) {
		output, err := api.ListInvitations(ctx)
		return nil, sentiary.InvitationsOutput{Invitations: output}, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_invitation",
		Title:       "Create invitation",
		Description: "Invites a user to a project with a selected role. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.CreateInvitationInput) (*mcp.CallToolResult, sentiary.OperationOutput, error) {
		output, err := api.CreateInvitation(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "accept_invitation",
		Title:       "Accept invitation",
		Description: "Accepts an invitation for the authenticated user. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.InvitationInput) (*mcp.CallToolResult, sentiary.OperationOutput, error) {
		output, err := api.AcceptInvitation(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "decline_invitation",
		Title:       "Decline invitation",
		Description: "Declines an invitation for the authenticated user. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.InvitationInput) (*mcp.CallToolResult, sentiary.OperationOutput, error) {
		output, err := api.DeclineInvitation(ctx, input)
		return nil, output, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_project_invitations",
		Title:       "List project invitations",
		Description: "Lists outstanding invitations for a project. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.ProjectInput) (*mcp.CallToolResult, sentiary.InvitationsOutput, error) {
		output, err := api.ListProjectInvitations(ctx, input)
		return nil, sentiary.InvitationsOutput{Invitations: output}, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "remove_project_invitation",
		Title:       "Remove project invitation",
		Description: "Removes an invitation from a project. Requires a Sentiary user API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input sentiary.ProjectInvitationInput) (*mcp.CallToolResult, sentiary.DeleteOutput, error) {
		output, err := api.RemoveProjectInvitation(ctx, input)
		return nil, output, err
	})
}
