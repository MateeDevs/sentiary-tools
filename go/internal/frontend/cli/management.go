package cli

import (
	"context"

	"github.com/MateeDevs/sentiary-tools/go/internal/sentiary"
)

func (app *App) runListProjects(ctx context.Context, args []string) error {
	flags := app.newFlagSet("list-projects")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	output, err := app.API.ListProjects(ctx)
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runGetProject(ctx context.Context, args []string) error {
	flags := app.newFlagSet("get-project")
	projectID := projectIDFlag(flags)
	if err := app.parse(flags, args); err != nil {
		return err
	}
	output, err := app.API.GetProject(ctx, sentiary.ProjectInput{ProjectID: *projectID})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runCreateProject(ctx context.Context, args []string) error {
	flags := app.newFlagSet("create-project")
	name := flags.String("name", "", "Project name.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if err := require("name", *name); err != nil {
		return err
	}
	output, err := app.API.CreateProject(ctx, sentiary.CreateProjectInput{Name: *name})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runEditProject(ctx context.Context, args []string) error {
	flags := app.newFlagSet("edit-project")
	projectID := projectIDFlag(flags)
	name := flags.String("name", "", "New project name.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if err := require("name", *name); err != nil {
		return err
	}
	output, err := app.API.EditProject(ctx, sentiary.EditProjectInput{ProjectID: *projectID, Name: *name})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runRemoveProject(ctx context.Context, args []string) error {
	flags := app.newFlagSet("remove-project")
	projectID := projectIDFlag(flags)
	if err := app.parse(flags, args); err != nil {
		return err
	}
	output, err := app.API.RemoveProject(ctx, sentiary.ProjectInput{ProjectID: *projectID})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runListLanguages(ctx context.Context, args []string) error {
	flags := app.newFlagSet("list-languages")
	inLanguage := flags.String("in-language", "", "Language ID for translated language names.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	output, err := app.API.ListLanguages(ctx, sentiary.ListLanguagesInput{InLanguage: *inLanguage})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runListProjectLanguages(ctx context.Context, args []string) error {
	flags := app.newFlagSet("list-project-languages")
	projectID := projectIDFlag(flags)
	inLanguage := flags.String("in-language", "", "Language ID for translated language names.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	output, err := app.API.ListProjectLanguages(ctx, sentiary.ListProjectLanguagesInput{
		ProjectID: *projectID, InLanguage: *inLanguage,
	})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runAddProjectLanguage(ctx context.Context, args []string) error {
	return app.runChangeProjectLanguage(ctx, "add-project-language", args, app.API.AddProjectLanguage)
}

func (app *App) runRemoveProjectLanguage(ctx context.Context, args []string) error {
	return app.runChangeProjectLanguage(ctx, "remove-project-language", args, app.API.RemoveProjectLanguage)
}

func (app *App) runChangeProjectLanguage(
	ctx context.Context,
	command string,
	args []string,
	operation func(context.Context, sentiary.ProjectLanguageInput) ([]sentiary.Language, error),
) error {
	flags := app.newFlagSet(command)
	projectID := projectIDFlag(flags)
	languageID := flags.String("language", "", "IETF BCP 47 language ID.")
	inLanguage := flags.String("in-language", "", "Language ID for translated language names.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if err := require("language", *languageID); err != nil {
		return err
	}
	output, err := operation(ctx, sentiary.ProjectLanguageInput{
		ProjectID: *projectID, LanguageID: *languageID, InLanguage: *inLanguage,
	})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runListProjectMembers(ctx context.Context, args []string) error {
	flags := app.newFlagSet("list-project-members")
	projectID := projectIDFlag(flags)
	if err := app.parse(flags, args); err != nil {
		return err
	}
	output, err := app.API.ListProjectMembers(ctx, sentiary.ProjectInput{ProjectID: *projectID})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runSetProjectMemberRole(ctx context.Context, args []string) error {
	flags := app.newFlagSet("set-project-member-role")
	projectID := projectIDFlag(flags)
	userID := flags.String("user-id", "", "Sentiary user ID.")
	role := flags.String("role", "", "Role: administrator, developer, or translator.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if err := require("user-id", *userID); err != nil {
		return err
	}
	if err := require("role", *role); err != nil {
		return err
	}
	output, err := app.API.SetProjectMemberRole(ctx, sentiary.SetProjectMemberRoleInput{
		ProjectID: *projectID, UserID: *userID, Role: *role,
	})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runRemoveProjectMember(ctx context.Context, args []string) error {
	flags := app.newFlagSet("remove-project-member")
	projectID := projectIDFlag(flags)
	userID := flags.String("user-id", "", "Sentiary user ID.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if err := require("user-id", *userID); err != nil {
		return err
	}
	output, err := app.API.RemoveProjectMember(ctx, sentiary.ProjectMemberInput{ProjectID: *projectID, UserID: *userID})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runListInvitations(ctx context.Context, args []string) error {
	flags := app.newFlagSet("list-invitations")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	output, err := app.API.ListInvitations(ctx)
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runCreateInvitation(ctx context.Context, args []string) error {
	flags := app.newFlagSet("create-invitation")
	projectID := projectIDFlag(flags)
	userEmail := flags.String("user-email", "", "Email address of the user to invite.")
	role := flags.String("role", "", "Role: administrator, developer, or translator.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if err := require("user-email", *userEmail); err != nil {
		return err
	}
	if err := require("role", *role); err != nil {
		return err
	}
	output, err := app.API.CreateInvitation(ctx, sentiary.CreateInvitationInput{
		ProjectID: *projectID, UserEmail: *userEmail, Role: *role,
	})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runAcceptInvitation(ctx context.Context, args []string) error {
	return app.runInvitationDecision(ctx, "accept-invitation", args, app.API.AcceptInvitation)
}

func (app *App) runDeclineInvitation(ctx context.Context, args []string) error {
	return app.runInvitationDecision(ctx, "decline-invitation", args, app.API.DeclineInvitation)
}

func (app *App) runInvitationDecision(
	ctx context.Context,
	command string,
	args []string,
	operation func(context.Context, sentiary.InvitationInput) (sentiary.OperationOutput, error),
) error {
	flags := app.newFlagSet(command)
	invitationID := flags.String("invitation-id", "", "Invitation ID.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if err := require("invitation-id", *invitationID); err != nil {
		return err
	}
	output, err := operation(ctx, sentiary.InvitationInput{InvitationID: *invitationID})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runListProjectInvitations(ctx context.Context, args []string) error {
	flags := app.newFlagSet("list-project-invitations")
	projectID := projectIDFlag(flags)
	if err := app.parse(flags, args); err != nil {
		return err
	}
	output, err := app.API.ListProjectInvitations(ctx, sentiary.ProjectInput{ProjectID: *projectID})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runRemoveProjectInvitation(ctx context.Context, args []string) error {
	flags := app.newFlagSet("remove-project-invitation")
	projectID := projectIDFlag(flags)
	invitationID := flags.String("invitation-id", "", "Invitation ID.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if err := require("invitation-id", *invitationID); err != nil {
		return err
	}
	output, err := app.API.RemoveProjectInvitation(ctx, sentiary.ProjectInvitationInput{
		ProjectID: *projectID, InvitationID: *invitationID,
	})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func projectIDFlag(flags interface {
	String(string, string, string) *string
}) *string {
	return flags.String("project-id", "", "Project ID. Uses SENTIARY_PROJECT_ID when empty.")
}
