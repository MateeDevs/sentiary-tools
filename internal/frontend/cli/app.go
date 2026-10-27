package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MateeDevs/sentiary-tools/internal/download"
	"github.com/MateeDevs/sentiary-tools/internal/sentiary"
)

var errHelp = errors.New("help shown")

type UsageError struct {
	Message string
}

func (err *UsageError) Error() string {
	return err.Message
}

func IsUsageError(err error) bool {
	var usageError *UsageError
	return errors.As(err, &usageError)
}

type App struct {
	DownloadServer string
	Downloads      download.Repository
	API            sentiary.API
	Stdin          io.Reader
	Stdout         io.Writer
	Stderr         io.Writer
	ReadFile       func(string) ([]byte, error)
	WriteFile      func(string, []byte, os.FileMode) error
}

func New(api sentiary.API) *App {
	return &App{
		API:            api,
		DownloadServer: sentiary.ConfigFromEnv().BaseURL,
		Stdin:          os.Stdin,
		Stdout:         os.Stdout,
		Stderr:         os.Stderr,
		ReadFile:       os.ReadFile,
		WriteFile:      os.WriteFile,
	}
}

func (app *App) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		app.writeUsage()
		return &UsageError{Message: "a command is required"}
	}

	var err error
	switch args[0] {
	case "help", "--help", "-h":
		app.writeUsage()
		return nil
	case "list-projects":
		err = app.runListProjects(ctx, args[1:])
	case "get-project":
		err = app.runGetProject(ctx, args[1:])
	case "create-project":
		err = app.runCreateProject(ctx, args[1:])
	case "edit-project":
		err = app.runEditProject(ctx, args[1:])
	case "remove-project":
		err = app.runRemoveProject(ctx, args[1:])
	case "list-languages":
		err = app.runListLanguages(ctx, args[1:])
	case "list-project-languages":
		err = app.runListProjectLanguages(ctx, args[1:])
	case "add-project-language":
		err = app.runAddProjectLanguage(ctx, args[1:])
	case "remove-project-language":
		err = app.runRemoveProjectLanguage(ctx, args[1:])
	case "list-project-members":
		err = app.runListProjectMembers(ctx, args[1:])
	case "set-project-member-role":
		err = app.runSetProjectMemberRole(ctx, args[1:])
	case "remove-project-member":
		err = app.runRemoveProjectMember(ctx, args[1:])
	case "list-invitations":
		err = app.runListInvitations(ctx, args[1:])
	case "create-invitation":
		err = app.runCreateInvitation(ctx, args[1:])
	case "accept-invitation":
		err = app.runAcceptInvitation(ctx, args[1:])
	case "decline-invitation":
		err = app.runDeclineInvitation(ctx, args[1:])
	case "list-project-invitations":
		err = app.runListProjectInvitations(ctx, args[1:])
	case "remove-project-invitation":
		err = app.runRemoveProjectInvitation(ctx, args[1:])
	case "list":
		err = app.runList(ctx, args[1:])
	case "search":
		err = app.runSearch(ctx, args[1:])
	case "get":
		err = app.runGet(ctx, args[1:])
	case "add":
		err = app.runAdd(ctx, args[1:])
	case "edit":
		err = app.runEdit(ctx, args[1:])
	case "remove":
		err = app.runRemove(ctx, args[1:])
	case "set-translation":
		err = app.runSetTranslation(ctx, args[1:])
	case "remove-translation":
		err = app.runRemoveTranslation(ctx, args[1:])
	case "info":
		err = app.runInfo(ctx, args[1:])
	case "export":
		err = app.runExport(ctx, args[1:])
	case "init":
		err = app.runInit(ctx, args[1:])
	case "pull":
		err = app.runPull(ctx, args[1:])
	case "import":
		err = app.runImport(ctx, args[1:])
	default:
		app.writeUsage()
		return &UsageError{Message: fmt.Sprintf("unknown command %q", args[0])}
	}

	if errors.Is(err, errHelp) {
		return nil
	}
	return err
}

func (app *App) runList(ctx context.Context, args []string) error {
	flags := app.newFlagSet("list")
	projectID := flags.String("project-id", "", "Project ID. Uses SENTIARY_PROJECT_ID when empty.")
	size := flags.Int("size", 0, "Maximum number of strings.")
	cursor := flags.String("cursor", "", "Pagination cursor.")
	languages := flags.String("include-language", "", "Comma-separated language IDs to include.")
	filter := flags.String("filter", "", "Translation filter.")
	filterLanguage := flags.String("filter-language", "", "Language ID for the filter.")
	order := flags.String("order", "", "Result order.")
	if err := app.parse(flags, args); err != nil {
		return err
	}

	output, err := app.API.ListStrings(ctx, sentiary.ListStringsInput{
		ProjectID:          *projectID,
		Size:               *size,
		Cursor:             *cursor,
		IncludeLanguageIDs: splitList(*languages),
		Filter:             *filter,
		FilterLanguage:     *filterLanguage,
		Order:              *order,
	})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runSearch(ctx context.Context, args []string) error {
	flags := app.newFlagSet("search")
	projectID := flags.String("project-id", "", "Project ID. Uses SENTIARY_PROJECT_ID when empty.")
	query := flags.String("query", "", "Text to find in keys, translations, descriptions, or context.")
	size := flags.Int("size", 25, "Maximum number of strings.")
	cursor := flags.String("cursor", "", "Pagination cursor.")
	languages := flags.String("include-language", "", "Comma-separated language IDs to include.")
	filter := flags.String("filter", "", "Translation filter.")
	filterLanguage := flags.String("filter-language", "", "Language ID for the filter.")
	order := flags.String("order", "", "Result order.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if err := require("query", *query); err != nil {
		return err
	}

	output, err := app.API.SearchStrings(ctx, sentiary.SearchStringsInput{
		ProjectID:          *projectID,
		Query:              *query,
		Size:               *size,
		Cursor:             *cursor,
		IncludeLanguageIDs: splitList(*languages),
		Filter:             *filter,
		FilterLanguage:     *filterLanguage,
		Order:              *order,
	})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runGet(ctx context.Context, args []string) error {
	flags := app.newFlagSet("get")
	projectID := flags.String("project-id", "", "Project ID. Uses SENTIARY_PROJECT_ID when empty.")
	stringID := flags.String("string-id", "", "String ID.")
	key := flags.String("key", "", "Exact string key.")
	languages := flags.String("include-language", "", "Comma-separated language IDs to include with a key lookup.")
	if err := app.parse(flags, args); err != nil {
		return err
	}

	if strings.TrimSpace(*stringID) != "" && strings.TrimSpace(*key) != "" {
		return &UsageError{Message: "use string-id or key, not both"}
	}
	if strings.TrimSpace(*key) != "" {
		output, err := app.API.GetStringByKey(ctx, sentiary.GetStringByKeyInput{
			ProjectID:          *projectID,
			Key:                *key,
			IncludeLanguageIDs: splitList(*languages),
		})
		if err != nil {
			return err
		}
		return app.writeJSON(output)
	}
	if err := require("string-id", *stringID); err != nil {
		return err
	}

	output, err := app.API.GetString(ctx, sentiary.GetStringInput{ProjectID: *projectID, StringID: *stringID})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runAdd(ctx context.Context, args []string) error {
	flags := app.newFlagSet("add")
	projectID := flags.String("project-id", "", "Project ID. Uses SENTIARY_PROJECT_ID when empty.")
	name := flags.String("name", "", "String name.")
	description := flags.String("description", "", "String description.")
	contextValue := flags.String("context", "", "String context.")
	language := flags.String("language", "", "Language ID for the first translation.")
	translation := flags.String("translation", "", "First translation text.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if err := require("name", *name); err != nil {
		return err
	}

	output, err := app.API.AddString(ctx, sentiary.AddStringInput{
		ProjectID:   *projectID,
		Name:        *name,
		Description: *description,
		Context:     *contextValue,
		LanguageID:  *language,
		Translation: *translation,
	})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runEdit(ctx context.Context, args []string) error {
	flags := app.newFlagSet("edit")
	projectID := flags.String("project-id", "", "Project ID. Uses SENTIARY_PROJECT_ID when empty.")
	stringID := flags.String("string-id", "", "String ID.")
	name := flags.String("name", "", "New string name.")
	description := flags.String("description", "", "New string description.")
	contextValue := flags.String("context", "", "New string context.")
	clearDescription := flags.Bool("clear-description", false, "Remove the description.")
	clearContext := flags.Bool("clear-context", false, "Remove the context.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if err := require("string-id", *stringID); err != nil {
		return err
	}

	output, err := app.API.EditString(ctx, sentiary.EditStringInput{
		ProjectID:        *projectID,
		StringID:         *stringID,
		Name:             *name,
		Description:      *description,
		Context:          *contextValue,
		ClearDescription: *clearDescription,
		ClearContext:     *clearContext,
	})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runRemove(ctx context.Context, args []string) error {
	flags := app.newFlagSet("remove")
	projectID := flags.String("project-id", "", "Project ID. Uses SENTIARY_PROJECT_ID when empty.")
	stringID := flags.String("string-id", "", "String ID.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if err := require("string-id", *stringID); err != nil {
		return err
	}

	output, err := app.API.RemoveString(ctx, sentiary.RemoveStringInput{ProjectID: *projectID, StringID: *stringID})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runSetTranslation(ctx context.Context, args []string) error {
	flags := app.newFlagSet("set-translation")
	projectID := flags.String("project-id", "", "Project ID. Uses SENTIARY_PROJECT_ID when empty.")
	stringID := flags.String("string-id", "", "String ID.")
	language := flags.String("language", "", "Translation language ID.")
	text := flags.String("text", "", "Translation text.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if err := require("string-id", *stringID); err != nil {
		return err
	}
	if err := require("language", *language); err != nil {
		return err
	}

	output, err := app.API.SetStringTranslation(ctx, sentiary.SetTranslationInput{
		ProjectID:  *projectID,
		StringID:   *stringID,
		LanguageID: *language,
		Text:       *text,
	})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runRemoveTranslation(ctx context.Context, args []string) error {
	flags := app.newFlagSet("remove-translation")
	projectID := flags.String("project-id", "", "Project ID. Uses SENTIARY_PROJECT_ID when empty.")
	stringID := flags.String("string-id", "", "String ID.")
	language := flags.String("language", "", "Translation language ID.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if err := require("string-id", *stringID); err != nil {
		return err
	}
	if err := require("language", *language); err != nil {
		return err
	}

	output, err := app.API.RemoveStringTranslation(ctx, sentiary.RemoveTranslationInput{
		ProjectID:  *projectID,
		StringID:   *stringID,
		LanguageID: *language,
	})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runInfo(ctx context.Context, args []string) error {
	flags := app.newFlagSet("info")
	projectID := flags.String("project-id", "", "Project ID. Uses SENTIARY_PROJECT_ID when empty.")
	if err := app.parse(flags, args); err != nil {
		return err
	}

	output, err := app.API.GetProjectBatchInfo(ctx, *projectID)
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) runExport(ctx context.Context, args []string) error {
	flags := app.newFlagSet("export")
	projectID := flags.String("project-id", "", "Project ID. Uses SENTIARY_PROJECT_ID when empty.")
	format := flags.String("format", "", "Export format: json, android, apple, or compose.")
	language := flags.String("language", "", "Language ID to export.")
	includeEmpty := flags.Bool("include-empty", false, "Include empty translations.")
	outputPath := flags.String("output", "-", "Output file. Use - for standard output.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if err := require("format", *format); err != nil {
		return err
	}
	if err := require("language", *language); err != nil {
		return err
	}

	output, err := app.API.ExportStrings(ctx, sentiary.ExportInput{
		ProjectID:                *projectID,
		LanguageID:               *language,
		Format:                   *format,
		IncludeEmptyTranslations: *includeEmpty,
	})
	if err != nil {
		return err
	}

	if *outputPath == "" || *outputPath == "-" {
		_, err = io.WriteString(app.Stdout, output.Content)
		return err
	}
	if err := app.WriteFile(*outputPath, []byte(output.Content), 0o644); err != nil {
		return fmt.Errorf("write export file: %w", err)
	}
	return nil
}

func (app *App) runImport(ctx context.Context, args []string) error {
	flags := app.newFlagSet("import")
	projectID := flags.String("project-id", "", "Project ID. Uses SENTIARY_PROJECT_ID when empty.")
	format := flags.String("format", "", "Import format: json, android, apple, or compose.")
	language := flags.String("language", "", "Language ID to import.")
	inputPath := flags.String("input", "-", "Input file. Use - for standard input.")
	overrideExisting := flags.Bool("override-existing", false, "Replace existing translations.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if err := require("format", *format); err != nil {
		return err
	}
	if err := require("language", *language); err != nil {
		return err
	}

	content, err := app.readInput(*inputPath)
	if err != nil {
		return err
	}
	output, err := app.API.ImportStrings(ctx, sentiary.ImportInput{
		ProjectID:        *projectID,
		LanguageID:       *language,
		Format:           *format,
		Content:          string(content),
		OverrideExisting: *overrideExisting,
	})
	if err != nil {
		return err
	}
	return app.writeJSON(output)
}

func (app *App) readInput(path string) ([]byte, error) {
	if path == "" || path == "-" {
		content, err := io.ReadAll(app.Stdin)
		if err != nil {
			return nil, fmt.Errorf("read standard input: %w", err)
		}
		return content, nil
	}

	content, err := app.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read import file: %w", err)
	}
	return content, nil
}

func (app *App) newFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(app.Stderr)
	return flags
}

func (app *App) parse(flags *flag.FlagSet, args []string) error {
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return errHelp
		}
		return &UsageError{Message: err.Error()}
	}
	if flags.NArg() != 0 {
		return &UsageError{Message: fmt.Sprintf("unexpected argument %q", flags.Arg(0))}
	}
	return nil
}

func (app *App) writeJSON(value any) error {
	encoder := json.NewEncoder(app.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func (app *App) writeUsage() {
	_, _ = fmt.Fprintln(app.Stderr, `Usage: sentiary-cli <command> [options]

Commands:
  list-projects           List projects.
  get-project             Get a project and membership.
  create-project          Create a project.
  edit-project            Rename a project.
  remove-project          Remove a project and its data.
  list-languages          List all supported languages.
  list-project-languages  List project languages.
  add-project-language    Add a language to a project.
  remove-project-language Remove a language and its translations.
  list-project-members    List project members.
  set-project-member-role Set a project member role.
  remove-project-member   Remove a project member.
  list-invitations        List invitations for the current user.
  create-invitation       Invite a user to a project.
  accept-invitation       Accept an invitation.
  decline-invitation      Decline an invitation.
  list-project-invitations List project invitations.
  remove-project-invitation Remove a project invitation.
  list                    List project strings.
  search                  Search project strings.
  get                     Get a string by ID or key.
  add                     Add a string.
  edit                    Edit a string.
  remove                  Remove a string.
  set-translation         Set a translation.
  remove-translation      Remove a translation.
  info                    Get project string information.
  export                  Export project strings.
  init                    Create a config for existing projects.
  pull                    Download configured translation outputs.
  import                  Import project strings.

Run sentiary-cli <command> --help for command options.`)
}

func require(name string, value string) error {
	if strings.TrimSpace(value) == "" {
		return &UsageError{Message: fmt.Sprintf("%s is required", name)}
	}
	return nil
}

func splitList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}
