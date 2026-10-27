package sentiary

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	endpoint, err := c.endpoint("project", nil)
	if err != nil {
		return nil, err
	}

	var output []Project
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, authAPIKey, &output); err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Client) CreateProject(ctx context.Context, input CreateProjectInput) (Project, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Project{}, errors.New("name is required")
	}

	endpoint, err := c.endpoint("project", nil)
	if err != nil {
		return Project{}, err
	}

	var output Project
	if err := c.doJSON(ctx, http.MethodPost, endpoint, jsonBody(map[string]string{"name": name}), authAPIKey, &output); err != nil {
		return Project{}, err
	}
	return output, nil
}

func (c *Client) GetProject(ctx context.Context, input ProjectInput) (ProjectWithMembership, error) {
	projectID, err := c.projectID(input.ProjectID)
	if err != nil {
		return ProjectWithMembership{}, err
	}

	endpoint, err := c.endpoint("project", projectID, nil)
	if err != nil {
		return ProjectWithMembership{}, err
	}

	var output ProjectWithMembership
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, authAPIKey, &output); err != nil {
		return ProjectWithMembership{}, err
	}
	return output, nil
}

func (c *Client) EditProject(ctx context.Context, input EditProjectInput) (ProjectWithMembership, error) {
	projectID, err := c.projectID(input.ProjectID)
	if err != nil {
		return ProjectWithMembership{}, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return ProjectWithMembership{}, errors.New("name is required")
	}

	endpoint, err := c.endpoint("project", projectID, nil)
	if err != nil {
		return ProjectWithMembership{}, err
	}

	var output ProjectWithMembership
	if err := c.doJSON(ctx, http.MethodPost, endpoint, jsonBody(map[string]string{"name": name}), authAPIKey, &output); err != nil {
		return ProjectWithMembership{}, err
	}
	return output, nil
}

func (c *Client) RemoveProject(ctx context.Context, input ProjectInput) (DeleteOutput, error) {
	projectID, err := c.projectID(input.ProjectID)
	if err != nil {
		return DeleteOutput{}, err
	}

	endpoint, err := c.endpoint("project", projectID, nil)
	if err != nil {
		return DeleteOutput{}, err
	}
	if err := c.doNoContent(ctx, http.MethodDelete, endpoint, nil, authAPIKey); err != nil {
		return DeleteOutput{}, err
	}
	return DeleteOutput{Deleted: true}, nil
}

func (c *Client) ListLanguages(ctx context.Context, input ListLanguagesInput) ([]Language, error) {
	query := languageNameQuery(input.InLanguage)
	endpoint, err := c.endpoint("language", query)
	if err != nil {
		return nil, err
	}

	var output []Language
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, authAPIKey, &output); err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Client) ListProjectLanguages(ctx context.Context, input ListProjectLanguagesInput) ([]Language, error) {
	projectID, err := c.projectID(input.ProjectID)
	if err != nil {
		return nil, err
	}

	endpoint, err := c.endpoint("project", projectID, "language", languageNameQuery(input.InLanguage))
	if err != nil {
		return nil, err
	}

	var output []Language
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, authAPIKey, &output); err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Client) AddProjectLanguage(ctx context.Context, input ProjectLanguageInput) ([]Language, error) {
	return c.changeProjectLanguage(ctx, http.MethodPost, input)
}

func (c *Client) RemoveProjectLanguage(ctx context.Context, input ProjectLanguageInput) ([]Language, error) {
	return c.changeProjectLanguage(ctx, http.MethodDelete, input)
}

func (c *Client) changeProjectLanguage(ctx context.Context, method string, input ProjectLanguageInput) ([]Language, error) {
	projectID, err := c.projectID(input.ProjectID)
	if err != nil {
		return nil, err
	}
	languageID, err := requiredValue("languageId", input.LanguageID)
	if err != nil {
		return nil, err
	}

	endpoint, err := c.endpoint("project", projectID, "language", languageID, languageNameQuery(input.InLanguage))
	if err != nil {
		return nil, err
	}

	var output []Language
	if err := c.doJSON(ctx, method, endpoint, nil, authAPIKey, &output); err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Client) ListProjectMembers(ctx context.Context, input ProjectInput) ([]ProjectMember, error) {
	projectID, err := c.projectID(input.ProjectID)
	if err != nil {
		return nil, err
	}

	endpoint, err := c.endpoint("project", projectID, "member", nil)
	if err != nil {
		return nil, err
	}

	var output []ProjectMember
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, authAPIKey, &output); err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Client) SetProjectMemberRole(ctx context.Context, input SetProjectMemberRoleInput) ([]ProjectMember, error) {
	projectID, err := c.projectID(input.ProjectID)
	if err != nil {
		return nil, err
	}
	userID, err := requiredValue("userId", input.UserID)
	if err != nil {
		return nil, err
	}
	role, err := normalizeMemberRole(input.Role)
	if err != nil {
		return nil, err
	}

	endpoint, err := c.endpoint("project", projectID, "member", userID, nil)
	if err != nil {
		return nil, err
	}

	var output []ProjectMember
	if err := c.doJSON(ctx, http.MethodPost, endpoint, jsonBody(map[string]string{"role": role}), authAPIKey, &output); err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Client) RemoveProjectMember(ctx context.Context, input ProjectMemberInput) ([]ProjectMember, error) {
	projectID, err := c.projectID(input.ProjectID)
	if err != nil {
		return nil, err
	}
	userID, err := requiredValue("userId", input.UserID)
	if err != nil {
		return nil, err
	}

	endpoint, err := c.endpoint("project", projectID, "member", userID, nil)
	if err != nil {
		return nil, err
	}

	var output []ProjectMember
	if err := c.doJSON(ctx, http.MethodDelete, endpoint, nil, authAPIKey, &output); err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Client) ListInvitations(ctx context.Context) ([]Invitation, error) {
	endpoint, err := c.endpoint("invite", nil)
	if err != nil {
		return nil, err
	}

	var output []Invitation
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, authAPIKey, &output); err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Client) CreateInvitation(ctx context.Context, input CreateInvitationInput) (OperationOutput, error) {
	projectID, err := c.projectID(input.ProjectID)
	if err != nil {
		return OperationOutput{}, err
	}
	userEmail, err := requiredValue("userEmail", input.UserEmail)
	if err != nil {
		return OperationOutput{}, err
	}
	role, err := normalizeMemberRole(input.Role)
	if err != nil {
		return OperationOutput{}, err
	}

	endpoint, err := c.endpoint("invite", nil)
	if err != nil {
		return OperationOutput{}, err
	}
	body := map[string]string{"projectId": projectID, "userEmail": userEmail, "role": role}
	if err := c.doNoContent(ctx, http.MethodPost, endpoint, jsonBody(body), authAPIKey); err != nil {
		return OperationOutput{}, err
	}
	return OperationOutput{Completed: true}, nil
}

func (c *Client) AcceptInvitation(ctx context.Context, input InvitationInput) (OperationOutput, error) {
	return c.decideInvitation(ctx, "accept", input)
}

func (c *Client) DeclineInvitation(ctx context.Context, input InvitationInput) (OperationOutput, error) {
	return c.decideInvitation(ctx, "decline", input)
}

func (c *Client) decideInvitation(ctx context.Context, decision string, input InvitationInput) (OperationOutput, error) {
	invitationID, err := requiredValue("invitationId", input.InvitationID)
	if err != nil {
		return OperationOutput{}, err
	}
	query := url.Values{}
	query.Set("inviteId", invitationID)
	endpoint, err := c.endpoint("invite", decision, query)
	if err != nil {
		return OperationOutput{}, err
	}
	if err := c.doNoContent(ctx, http.MethodPost, endpoint, nil, authAPIKey); err != nil {
		return OperationOutput{}, err
	}
	return OperationOutput{Completed: true}, nil
}

func (c *Client) ListProjectInvitations(ctx context.Context, input ProjectInput) ([]Invitation, error) {
	projectID, err := c.projectID(input.ProjectID)
	if err != nil {
		return nil, err
	}
	endpoint, err := c.endpoint("project", projectID, "invite", nil)
	if err != nil {
		return nil, err
	}

	var output []Invitation
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, authAPIKey, &output); err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Client) RemoveProjectInvitation(ctx context.Context, input ProjectInvitationInput) (DeleteOutput, error) {
	projectID, err := c.projectID(input.ProjectID)
	if err != nil {
		return DeleteOutput{}, err
	}
	invitationID, err := requiredValue("invitationId", input.InvitationID)
	if err != nil {
		return DeleteOutput{}, err
	}
	endpoint, err := c.endpoint("project", projectID, "invite", invitationID, nil)
	if err != nil {
		return DeleteOutput{}, err
	}
	if err := c.doNoContent(ctx, http.MethodDelete, endpoint, nil, authAPIKey); err != nil {
		return DeleteOutput{}, err
	}
	return DeleteOutput{Deleted: true}, nil
}

func languageNameQuery(inLanguage string) url.Values {
	query := url.Values{}
	setOptionalQuery(query, "inLanguage", inLanguage)
	return query
}

func normalizeMemberRole(role string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(role))
	switch normalized {
	case "administrator", "developer", "translator":
		return normalized, nil
	default:
		return "", errors.New("role must be administrator, developer, or translator")
	}
}

func requiredValue(name string, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", errors.New(name + " is required")
	}
	return trimmed, nil
}
