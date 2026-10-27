package sentiary

type ProjectInput struct {
	ProjectID string `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
}

type CreateProjectInput struct {
	Name string `json:"name" jsonschema:"Project name."`
}

type EditProjectInput struct {
	ProjectID string `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	Name      string `json:"name" jsonschema:"New project name."`
}

type ListLanguagesInput struct {
	InLanguage string `json:"inLanguage,omitempty" jsonschema:"Language ID for translated language names. The API uses the authenticated user's language when empty."`
}

type ListProjectLanguagesInput struct {
	ProjectID  string `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	InLanguage string `json:"inLanguage,omitempty" jsonschema:"Language ID for translated language names. The API uses the authenticated user's language when empty."`
}

type ProjectLanguageInput struct {
	ProjectID  string `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	LanguageID string `json:"languageId" jsonschema:"IETF BCP 47 language ID to add or remove."`
	InLanguage string `json:"inLanguage,omitempty" jsonschema:"Language ID for translated language names. The API uses the authenticated user's language when empty."`
}

type ProjectMemberInput struct {
	ProjectID string `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	UserID    string `json:"userId" jsonschema:"Sentiary user ID."`
}

type SetProjectMemberRoleInput struct {
	ProjectID string `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	UserID    string `json:"userId" jsonschema:"Sentiary user ID."`
	Role      string `json:"role" jsonschema:"Member role: administrator, developer, or translator."`
}

type CreateInvitationInput struct {
	ProjectID string `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	UserEmail string `json:"userEmail" jsonschema:"Email address of the user to invite."`
	Role      string `json:"role" jsonschema:"Invited member role: administrator, developer, or translator."`
}

type InvitationInput struct {
	InvitationID string `json:"invitationId" jsonschema:"Invitation ID."`
}

type ProjectInvitationInput struct {
	ProjectID    string `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	InvitationID string `json:"invitationId" jsonschema:"Invitation ID."`
}

type ListStringsInput struct {
	ProjectID          string   `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	Size               int      `json:"size,omitempty" jsonschema:"Maximum number of strings to return. Defaults to the API default."`
	Cursor             string   `json:"cursor,omitempty" jsonschema:"Pagination cursor returned by a previous response."`
	IncludeLanguageIDs []string `json:"includeLanguageIds,omitempty" jsonschema:"Translation language IDs to include in each string."`
	Filter             string   `json:"filter,omitempty" jsonschema:"Optional API filter: ALL, TRANSLATED, or NOT_TRANSLATED."`
	FilterLanguage     string   `json:"filterLanguage,omitempty" jsonschema:"Language ID used by filter."`
	Order              string   `json:"order,omitempty" jsonschema:"Optional API order, for example UPDATED_ASC or UPDATED_DESC."`
}

type SearchStringsInput struct {
	ProjectID          string   `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	Query              string   `json:"query" jsonschema:"Text to find in string keys, translations, descriptions, or context."`
	Size               int      `json:"size,omitempty" jsonschema:"Maximum API page size to search. Defaults to 25."`
	Cursor             string   `json:"cursor,omitempty" jsonschema:"Pagination cursor returned by a previous search response."`
	IncludeLanguageIDs []string `json:"includeLanguageIds,omitempty" jsonschema:"Translation language IDs to include in each string."`
	Filter             string   `json:"filter,omitempty" jsonschema:"Optional API filter: ALL, TRANSLATED, or NOT_TRANSLATED."`
	FilterLanguage     string   `json:"filterLanguage,omitempty" jsonschema:"Language ID used by filter."`
	Order              string   `json:"order,omitempty" jsonschema:"Optional API order, for example UPDATED_ASC or UPDATED_DESC."`
}

type GetStringInput struct {
	ProjectID string `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	StringID  string `json:"stringId" jsonschema:"String/term ID."`
}

type GetStringByKeyInput struct {
	ProjectID          string   `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	Key                string   `json:"key" jsonschema:"String key/name to fetch exactly."`
	IncludeLanguageIDs []string `json:"includeLanguageIds,omitempty" jsonschema:"Translation language IDs to include in the string."`
}

type AddStringInput struct {
	ProjectID   string `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	Name        string `json:"name" jsonschema:"String name/key."`
	Description string `json:"description,omitempty" jsonschema:"Optional string description."`
	Context     string `json:"context,omitempty" jsonschema:"Optional string context."`
	LanguageID  string `json:"languageId,omitempty" jsonschema:"Optional language ID for an initial translation."`
	Translation string `json:"translation,omitempty" jsonschema:"Optional initial translation text. Requires languageId."`
}

type EditStringInput struct {
	ProjectID        string `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	StringID         string `json:"stringId" jsonschema:"String/term ID."`
	Name             string `json:"name,omitempty" jsonschema:"New string name/key. Omit to keep unchanged."`
	Description      string `json:"description,omitempty" jsonschema:"New string description. Omit to keep unchanged unless clearDescription is true."`
	Context          string `json:"context,omitempty" jsonschema:"New string context. Omit to keep unchanged unless clearContext is true."`
	ClearDescription bool   `json:"clearDescription,omitempty" jsonschema:"Set description to null."`
	ClearContext     bool   `json:"clearContext,omitempty" jsonschema:"Set context to null."`
}

type RemoveStringInput struct {
	ProjectID string `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	StringID  string `json:"stringId" jsonschema:"String/term ID to remove."`
}

type SetTranslationInput struct {
	ProjectID  string `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	StringID   string `json:"stringId" jsonschema:"String/term ID."`
	LanguageID string `json:"languageId" jsonschema:"Translation language ID."`
	Text       string `json:"text" jsonschema:"Translation text."`
}

type RemoveTranslationInput struct {
	ProjectID  string `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	StringID   string `json:"stringId" jsonschema:"String/term ID."`
	LanguageID string `json:"languageId" jsonschema:"Translation language ID to remove."`
}

type ExportInput struct {
	ProjectID                string `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	LanguageID               string `json:"languageId" jsonschema:"Language ID to export, for example en or cs."`
	Format                   string `json:"format" jsonschema:"Export format: json, android, apple, or compose."`
	IncludeEmptyTranslations bool   `json:"includeEmptyTranslations,omitempty" jsonschema:"When true, includes terms with empty translations."`
}

type ExportOutput struct {
	ProjectID                string `json:"projectId"`
	LanguageID               string `json:"languageId"`
	Format                   string `json:"format"`
	IncludeEmptyTranslations bool   `json:"includeEmptyTranslations"`
	Content                  string `json:"content"`
}

type ImportInput struct {
	ProjectID        string `json:"projectId,omitempty" jsonschema:"Project ID. Optional when SENTIARY_PROJECT_ID is configured."`
	LanguageID       string `json:"languageId" jsonschema:"Language ID to import into, for example en or cs."`
	Format           string `json:"format" jsonschema:"Import format: json, android, apple, or compose."`
	Content          string `json:"content" jsonschema:"String file content to import."`
	OverrideExisting bool   `json:"overrideExisting,omitempty" jsonschema:"When true, existing translations can be overwritten."`
}

type ProjectBatchInfo struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	TermsLastModified *string  `json:"termsLastModified"`
	Languages         []string `json:"languages"`
}

type ImportResult struct {
	TotalImportedCount int      `json:"totalImportedCount"`
	Skipped            []string `json:"skipped"`
	ImportLanguage     string   `json:"importLanguage"`
}

type Paging struct {
	Items      []Term     `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Size     int     `json:"size"`
	Cursor   *string `json:"cursor"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
}

type Term struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  *string       `json:"description"`
	Translations []Translation `json:"translations"`
	Updated      string        `json:"updated"`
	AuthorEmail  *string       `json:"authorEmail"`
	AuthorID     *string       `json:"authorId"`
	Statistics   any           `json:"statistics" jsonschema:"Arbitrary per-string statistics object returned by the Sentiary API. The shape is not modeled."`
	Context      *string       `json:"context"`
}

type Translation struct {
	LanguageID  string  `json:"languageId"`
	Text        string  `json:"text"`
	Updated     string  `json:"updated"`
	AuthorEmail *string `json:"authorEmail"`
	AuthorID    *string `json:"authorId"`
}

type DeleteOutput struct {
	Deleted bool `json:"deleted"`
}

type OperationOutput struct {
	Completed bool `json:"completed"`
}

type ProjectsOutput struct {
	Projects []Project `json:"projects"`
}

type LanguagesOutput struct {
	Languages []Language `json:"languages"`
}

type ProjectMembersOutput struct {
	Members []ProjectMember `json:"members"`
}

type InvitationsOutput struct {
	Invitations []Invitation `json:"invitations"`
}

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Membership struct {
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

type ProjectWithMembership struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Membership Membership `json:"membership"`
}

type Language struct {
	LanguageID        string  `json:"languageId"`
	LanguageName      string  `json:"languageName"`
	LocalLanguageName string  `json:"localLanguageName"`
	CountryEmoji      *string `json:"countryEmoji"`
}

type User struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	EmailAddress string `json:"emailAddress"`
	IsAdmin      bool   `json:"isAdmin"`
}

type ProjectMember struct {
	User User   `json:"user"`
	Role string `json:"role"`
}

type Invitation struct {
	InvitationID     string  `json:"inviteId"`
	Project          Project `json:"project"`
	Role             string  `json:"role"`
	InvitedUserEmail string  `json:"invitedUserEmail"`
	InvitedBy        *User   `json:"invitedBy"`
	Accepted         bool    `json:"accepted"`
	Declined         bool    `json:"declined"`
}
