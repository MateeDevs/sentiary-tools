package sentiary

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type expectedManagementRequest struct {
	method       string
	path         string
	query        string
	bodyContains string
	status       int
	response     string
}

func TestManagementOperationsUseAPIContracts(t *testing.T) {
	requests := []expectedManagementRequest{
		{method: http.MethodGet, path: "/api/v1/project", response: `[{"id":"p1","name":"App"}]`},
		{method: http.MethodPost, path: "/api/v1/project", bodyContains: `"name":"New App"`, response: `{"id":"p2","name":"New App"}`},
		{method: http.MethodGet, path: "/api/v1/project/p1", response: projectWithMembershipJSON},
		{method: http.MethodPost, path: "/api/v1/project/p1", bodyContains: `"name":"Renamed"`, response: projectWithMembershipJSON},
		{method: http.MethodDelete, path: "/api/v1/project/p1", status: http.StatusNoContent},
		{method: http.MethodGet, path: "/api/v1/language", query: "inLanguage=cs-CZ", response: languagesJSON},
		{method: http.MethodGet, path: "/api/v1/project/p1/language", query: "inLanguage=cs-CZ", response: languagesJSON},
		{method: http.MethodPost, path: "/api/v1/project/p1/language/de-DE", query: "inLanguage=cs-CZ", response: languagesJSON},
		{method: http.MethodDelete, path: "/api/v1/project/p1/language/de-DE", query: "inLanguage=cs-CZ", response: languagesJSON},
		{method: http.MethodGet, path: "/api/v1/project/p1/member", response: membersJSON},
		{method: http.MethodPost, path: "/api/v1/project/p1/member/u1", bodyContains: `"role":"administrator"`, response: membersJSON},
		{method: http.MethodDelete, path: "/api/v1/project/p1/member/u1", response: `[]`},
		{method: http.MethodGet, path: "/api/v1/invite", response: invitationsJSON},
		{method: http.MethodPost, path: "/api/v1/invite", bodyContains: `"role":"developer"`, status: http.StatusNoContent},
		{method: http.MethodPost, path: "/api/v1/invite/accept", query: "inviteId=i1", status: http.StatusNoContent},
		{method: http.MethodPost, path: "/api/v1/invite/decline", query: "inviteId=i1", status: http.StatusNoContent},
		{method: http.MethodGet, path: "/api/v1/project/p1/invite", response: invitationsJSON},
		{method: http.MethodDelete, path: "/api/v1/project/p1/invite/i1", status: http.StatusNoContent},
	}

	requestIndex := 0
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if requestIndex >= len(requests) {
			t.Fatalf("Unexpected request: %s %s", request.Method, request.URL)
		}
		if request.Header.Get("Authorization") != "Ribbon test-key" {
			t.Error("Request does not contain the user key in the Ribbon header")
		}
		if request.URL.Query().Has("apiKey") {
			t.Error("Request contains an API key in the URL")
		}
		expected := requests[requestIndex]
		requestIndex++
		if request.Method != expected.method {
			t.Errorf("Request %d method = %q, want %q", requestIndex, request.Method, expected.method)
		}
		if request.URL.Path != expected.path {
			t.Errorf("Request %d path = %q, want %q", requestIndex, request.URL.Path, expected.path)
		}
		if request.URL.RawQuery != expected.query {
			t.Errorf("Request %d query = %q, want %q", requestIndex, request.URL.RawQuery, expected.query)
		}
		if expected.bodyContains != "" {
			body, err := io.ReadAll(request.Body)
			if err != nil {
				t.Fatalf("Read request body: %v", err)
			}
			if !strings.Contains(string(body), expected.bodyContains) {
				t.Errorf("Request %d body = %q, want content %q", requestIndex, body, expected.bodyContains)
			}
		}
		status := expected.status
		if status == 0 {
			status = http.StatusOK
		}
		return &http.Response{
			StatusCode: status,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(expected.response)),
			Request:    request,
		}, nil
	})}
	client := NewClient(Config{BaseURL: "https://api.test", APIKey: "test-key"}, httpClient)
	ctx := context.Background()

	_, err := client.ListProjects(ctx)
	checkCall(t, err)
	_, err = client.CreateProject(ctx, CreateProjectInput{Name: "New App"})
	checkCall(t, err)
	_, err = client.GetProject(ctx, ProjectInput{ProjectID: "p1"})
	checkCall(t, err)
	_, err = client.EditProject(ctx, EditProjectInput{ProjectID: "p1", Name: "Renamed"})
	checkCall(t, err)
	_, err = client.RemoveProject(ctx, ProjectInput{ProjectID: "p1"})
	checkCall(t, err)
	_, err = client.ListLanguages(ctx, ListLanguagesInput{InLanguage: "cs-CZ"})
	checkCall(t, err)
	_, err = client.ListProjectLanguages(ctx, ListProjectLanguagesInput{ProjectID: "p1", InLanguage: "cs-CZ"})
	checkCall(t, err)
	_, err = client.AddProjectLanguage(ctx, ProjectLanguageInput{ProjectID: "p1", LanguageID: "de-DE", InLanguage: "cs-CZ"})
	checkCall(t, err)
	_, err = client.RemoveProjectLanguage(ctx, ProjectLanguageInput{ProjectID: "p1", LanguageID: "de-DE", InLanguage: "cs-CZ"})
	checkCall(t, err)
	_, err = client.ListProjectMembers(ctx, ProjectInput{ProjectID: "p1"})
	checkCall(t, err)
	_, err = client.SetProjectMemberRole(ctx, SetProjectMemberRoleInput{ProjectID: "p1", UserID: "u1", Role: "ADMINISTRATOR"})
	checkCall(t, err)
	_, err = client.RemoveProjectMember(ctx, ProjectMemberInput{ProjectID: "p1", UserID: "u1"})
	checkCall(t, err)
	_, err = client.ListInvitations(ctx)
	checkCall(t, err)
	_, err = client.CreateInvitation(ctx, CreateInvitationInput{ProjectID: "p1", UserEmail: "user@example.com", Role: "DEVELOPER"})
	checkCall(t, err)
	_, err = client.AcceptInvitation(ctx, InvitationInput{InvitationID: "i1"})
	checkCall(t, err)
	_, err = client.DeclineInvitation(ctx, InvitationInput{InvitationID: "i1"})
	checkCall(t, err)
	_, err = client.ListProjectInvitations(ctx, ProjectInput{ProjectID: "p1"})
	checkCall(t, err)
	_, err = client.RemoveProjectInvitation(ctx, ProjectInvitationInput{ProjectID: "p1", InvitationID: "i1"})
	checkCall(t, err)

	if requestIndex != len(requests) {
		t.Fatalf("Request count = %d, want %d", requestIndex, len(requests))
	}
}

func TestManagementOperationsValidateRoles(t *testing.T) {
	client := NewClient(Config{DefaultProjectID: "p1"}, nil)
	if _, err := client.SetProjectMemberRole(context.Background(), SetProjectMemberRoleInput{UserID: "u1", Role: "owner"}); err == nil {
		t.Fatal("Set member role accepted an unknown role")
	}
	if _, err := client.CreateInvitation(context.Background(), CreateInvitationInput{UserEmail: "user@example.com", Role: "owner"}); err == nil {
		t.Fatal("Create invitation accepted an unknown role")
	}
}

func checkCall(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

const projectWithMembershipJSON = `{
	"id":"p1",
	"name":"App",
	"membership":{"role":"administrator","permissions":["MODIFY_PROJECT"]}
}`

const languagesJSON = `[{"languageId":"cs-CZ","languageName":"Czech","localLanguageName":"Čeština","countryEmoji":"🇨🇿"}]`

const membersJSON = `[{
	"user":{"id":"u1","name":"User","emailAddress":"user@example.com","isAdmin":false},
	"role":"administrator"
}]`

const invitationsJSON = `[{
	"inviteId":"i1",
	"project":{"id":"p1","name":"App"},
	"role":"developer",
	"invitedUserEmail":"user@example.com",
	"invitedBy":null,
	"accepted":false,
	"declined":false
}]`
