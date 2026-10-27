package sentiary

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const apiPrefix = "/api/v1"

type authMode int

const authAPIKey authMode = iota

type Client struct {
	config Config
	http   *http.Client
}

func NewClient(config Config, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if strings.TrimSpace(config.BaseURL) == "" {
		config.BaseURL = defaultBaseURL
	}
	return &Client{config: config, http: httpClient}
}

func (c *Client) GetProjectBatchInfo(ctx context.Context, projectID string) (ProjectBatchInfo, error) {
	resolvedProjectID, err := c.projectID(projectID)
	if err != nil {
		return ProjectBatchInfo{}, err
	}

	endpoint, err := c.endpoint("batch", resolvedProjectID, "info", nil)
	if err != nil {
		return ProjectBatchInfo{}, err
	}

	var output ProjectBatchInfo
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, authAPIKey, &output); err != nil {
		return ProjectBatchInfo{}, err
	}
	return output, nil
}

func (c *Client) ExportStrings(ctx context.Context, input ExportInput) (ExportOutput, error) {
	resolvedProjectID, err := c.projectID(input.ProjectID)
	if err != nil {
		return ExportOutput{}, err
	}
	format, err := normalizeFormat(input.Format)
	if err != nil {
		return ExportOutput{}, err
	}
	languageID := strings.TrimSpace(input.LanguageID)
	if languageID == "" {
		return ExportOutput{}, errors.New("languageId is required")
	}

	query := url.Values{}
	query.Set("languageId", languageID)
	query.Set("format", format)
	query.Set("includeEmptyTranslations", fmt.Sprintf("%t", input.IncludeEmptyTranslations))

	endpoint, err := c.endpoint("batch", resolvedProjectID, "export", query)
	if err != nil {
		return ExportOutput{}, err
	}

	content, err := c.doText(ctx, http.MethodGet, endpoint, nil, authAPIKey)
	if err != nil {
		return ExportOutput{}, err
	}

	return ExportOutput{
		ProjectID:                resolvedProjectID,
		LanguageID:               languageID,
		Format:                   format,
		IncludeEmptyTranslations: input.IncludeEmptyTranslations,
		Content:                  content,
	}, nil
}

func (c *Client) ImportStrings(ctx context.Context, input ImportInput) (ImportResult, error) {
	resolvedProjectID, err := c.projectID(input.ProjectID)
	if err != nil {
		return ImportResult{}, err
	}
	format, err := normalizeFormat(input.Format)
	if err != nil {
		return ImportResult{}, err
	}
	languageID := strings.TrimSpace(input.LanguageID)
	if languageID == "" {
		return ImportResult{}, errors.New("languageId is required")
	}
	if input.Content == "" {
		return ImportResult{}, errors.New("content is required")
	}

	query := url.Values{}
	query.Set("languageId", languageID)
	query.Set("format", format)
	query.Set("overrideExisting", fmt.Sprintf("%t", input.OverrideExisting))

	endpoint, err := c.endpoint("batch", resolvedProjectID, "import", query)
	if err != nil {
		return ImportResult{}, err
	}

	var output ImportResult
	if err := c.doJSON(ctx, http.MethodPost, endpoint, textBody(input.Content), authAPIKey, &output); err != nil {
		return ImportResult{}, err
	}
	return output, nil
}

func (c *Client) ListStrings(ctx context.Context, input ListStringsInput) (Paging, error) {
	projectID, err := c.projectID(input.ProjectID)
	if err != nil {
		return Paging{}, err
	}

	query := url.Values{}
	if input.Size > 0 {
		query.Set("size", fmt.Sprintf("%d", input.Size))
	}
	if strings.TrimSpace(input.Cursor) != "" {
		query.Set("cursor", strings.TrimSpace(input.Cursor))
	}
	for _, languageID := range input.IncludeLanguageIDs {
		if trimmed := strings.TrimSpace(languageID); trimmed != "" {
			query.Add("includeLanguage", trimmed)
		}
	}
	setOptionalQuery(query, "filter", input.Filter)
	setOptionalQuery(query, "filterLanguage", input.FilterLanguage)
	setOptionalQuery(query, "order", input.Order)

	endpoint, err := c.endpoint("project", projectID, "term", query)
	if err != nil {
		return Paging{}, err
	}

	var output Paging
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, authAPIKey, &output); err != nil {
		return Paging{}, err
	}
	return output, nil
}

func (c *Client) SearchStrings(ctx context.Context, input SearchStringsInput) (Paging, error) {
	searchQuery := strings.TrimSpace(input.Query)
	if searchQuery == "" {
		return Paging{}, errors.New("query is required")
	}

	listInput := ListStringsInput{
		ProjectID:          input.ProjectID,
		Size:               input.Size,
		Cursor:             input.Cursor,
		IncludeLanguageIDs: input.IncludeLanguageIDs,
		Filter:             input.Filter,
		FilterLanguage:     input.FilterLanguage,
		Order:              input.Order,
	}
	if listInput.Size <= 0 {
		listInput.Size = 25
	}

	projectID, err := c.projectID(listInput.ProjectID)
	if err != nil {
		return Paging{}, err
	}

	query := url.Values{}
	query.Set("query", searchQuery)
	query.Set("size", fmt.Sprintf("%d", listInput.Size))
	setOptionalQuery(query, "cursor", listInput.Cursor)
	for _, languageID := range listInput.IncludeLanguageIDs {
		if trimmed := strings.TrimSpace(languageID); trimmed != "" {
			query.Add("includeLanguage", trimmed)
		}
	}
	setOptionalQuery(query, "filter", listInput.Filter)
	setOptionalQuery(query, "filterLanguage", listInput.FilterLanguage)
	setOptionalQuery(query, "order", listInput.Order)

	endpoint, err := c.endpoint("project", projectID, "term", query)
	if err != nil {
		return Paging{}, err
	}

	var output Paging
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, authAPIKey, &output); err != nil {
		return Paging{}, err
	}

	return output, nil
}

func (c *Client) GetString(ctx context.Context, input GetStringInput) (Term, error) {
	projectID, termID, err := c.projectAndTermID(input.ProjectID, input.StringID)
	if err != nil {
		return Term{}, err
	}

	endpoint, err := c.endpoint("project", projectID, "term", termID, nil)
	if err != nil {
		return Term{}, err
	}

	var output Term
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, authAPIKey, &output); err != nil {
		return Term{}, err
	}
	return output, nil
}

func (c *Client) GetStringByKey(ctx context.Context, input GetStringByKeyInput) (Term, error) {
	projectID, err := c.projectID(input.ProjectID)
	if err != nil {
		return Term{}, err
	}
	key := strings.TrimSpace(input.Key)
	if key == "" {
		return Term{}, errors.New("key is required")
	}

	query := url.Values{}
	for _, languageID := range input.IncludeLanguageIDs {
		if trimmed := strings.TrimSpace(languageID); trimmed != "" {
			query.Add("includeLanguage", trimmed)
		}
	}

	endpoint, err := c.endpoint("project", projectID, "term", "key", key, query)
	if err != nil {
		return Term{}, err
	}

	var output Term
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, authAPIKey, &output); err != nil {
		return Term{}, err
	}
	return output, nil
}

func (c *Client) AddString(ctx context.Context, input AddStringInput) (Term, error) {
	projectID, err := c.projectID(input.ProjectID)
	if err != nil {
		return Term{}, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Term{}, errors.New("name is required")
	}

	body := map[string]any{
		"name":        name,
		"description": optionalString(input.Description),
		"context":     optionalString(input.Context),
	}
	endpoint, err := c.endpoint("project", projectID, "term", nil)
	if err != nil {
		return Term{}, err
	}

	var output Term
	if err := c.doJSON(ctx, http.MethodPost, endpoint, jsonBody(body), authAPIKey, &output); err != nil {
		return Term{}, err
	}

	if strings.TrimSpace(input.LanguageID) != "" {
		return c.SetStringTranslation(ctx, SetTranslationInput{
			ProjectID:  projectID,
			StringID:   output.ID,
			LanguageID: input.LanguageID,
			Text:       input.Translation,
		})
	}

	return output, nil
}

func (c *Client) EditString(ctx context.Context, input EditStringInput) (Term, error) {
	projectID, termID, err := c.projectAndTermID(input.ProjectID, input.StringID)
	if err != nil {
		return Term{}, err
	}

	current, err := c.GetString(ctx, GetStringInput{ProjectID: projectID, StringID: termID})
	if err != nil {
		return Term{}, err
	}

	name := current.Name
	if strings.TrimSpace(input.Name) != "" {
		name = strings.TrimSpace(input.Name)
	}

	var description any = current.Description
	if input.ClearDescription {
		description = nil
	} else if input.Description != "" {
		description = input.Description
	}

	var contextValue any = current.Context
	if input.ClearContext {
		contextValue = nil
	} else if input.Context != "" {
		contextValue = input.Context
	}

	body := map[string]any{
		"name":        name,
		"description": description,
		"context":     contextValue,
	}
	endpoint, err := c.endpoint("project", projectID, "term", termID, nil)
	if err != nil {
		return Term{}, err
	}

	var output Term
	if err := c.doJSON(ctx, http.MethodPut, endpoint, jsonBody(body), authAPIKey, &output); err != nil {
		return Term{}, err
	}
	return output, nil
}

func (c *Client) RemoveString(ctx context.Context, input RemoveStringInput) (DeleteOutput, error) {
	projectID, termID, err := c.projectAndTermID(input.ProjectID, input.StringID)
	if err != nil {
		return DeleteOutput{}, err
	}

	endpoint, err := c.endpoint("project", projectID, "term", termID, nil)
	if err != nil {
		return DeleteOutput{}, err
	}
	if err := c.doNoContent(ctx, http.MethodDelete, endpoint, nil, authAPIKey); err != nil {
		return DeleteOutput{}, err
	}
	return DeleteOutput{Deleted: true}, nil
}

func (c *Client) SetStringTranslation(ctx context.Context, input SetTranslationInput) (Term, error) {
	projectID, termID, err := c.projectAndTermID(input.ProjectID, input.StringID)
	if err != nil {
		return Term{}, err
	}
	languageID := strings.TrimSpace(input.LanguageID)
	if languageID == "" {
		return Term{}, errors.New("languageId is required")
	}

	body := map[string]any{"languageId": languageID, "text": input.Text}
	endpoint, err := c.endpoint("project", projectID, "term", termID, "translation", nil)
	if err != nil {
		return Term{}, err
	}

	var output Term
	if err := c.doJSON(ctx, http.MethodPost, endpoint, jsonBody(body), authAPIKey, &output); err != nil {
		return Term{}, err
	}
	return output, nil
}

func (c *Client) RemoveStringTranslation(ctx context.Context, input RemoveTranslationInput) (Term, error) {
	projectID, termID, err := c.projectAndTermID(input.ProjectID, input.StringID)
	if err != nil {
		return Term{}, err
	}
	languageID := strings.TrimSpace(input.LanguageID)
	if languageID == "" {
		return Term{}, errors.New("languageId is required")
	}

	endpoint, err := c.endpoint("project", projectID, "term", termID, "translation", languageID, nil)
	if err != nil {
		return Term{}, err
	}

	var output Term
	if err := c.doJSON(ctx, http.MethodDelete, endpoint, nil, authAPIKey, &output); err != nil {
		return Term{}, err
	}
	return output, nil
}

func (c *Client) doJSON(ctx context.Context, method string, endpoint string, body io.Reader, mode authMode, output any) error {
	responseBody, err := c.do(ctx, method, endpoint, body, mode)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	decoder := json.NewDecoder(responseBody)
	if err := decoder.Decode(output); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) doNoContent(ctx context.Context, method string, endpoint string, body io.Reader, mode authMode) error {
	responseBody, err := c.do(ctx, method, endpoint, body, mode)
	if err != nil {
		return err
	}
	return responseBody.Close()
}

func (c *Client) doText(ctx context.Context, method string, endpoint string, body io.Reader, mode authMode) (string, error) {
	responseBody, err := c.do(ctx, method, endpoint, body, mode)
	if err != nil {
		return "", err
	}
	defer responseBody.Close()

	data, err := io.ReadAll(responseBody)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	return string(data), nil
}

func (c *Client) do(ctx context.Context, method string, endpoint string, body io.Reader, mode authMode) (io.ReadCloser, error) {
	request, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.authorize(request, mode); err != nil {
		return nil, err
	}
	if body != nil {
		contentType := "application/json; charset=utf-8"
		if typedBody, ok := body.(contentTypedReader); ok {
			contentType = typedBody.contentType
		}
		request.Header.Set("Content-Type", contentType)
	}
	request.Header.Set("Accept", "application/json, text/plain, */*")

	response, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("call Sentiary API: %w", err)
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return response.Body, nil
	}

	defer response.Body.Close()
	message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	return nil, fmt.Errorf("Sentiary API returned %s: %s", response.Status, strings.TrimSpace(string(message)))
}

func (c *Client) authorize(request *http.Request, _ authMode) error {
	apiKey := strings.TrimSpace(c.config.APIKey)
	if apiKey == "" {
		return errors.New("missing API key: set SENTIARY_USER_API_KEY")
	}
	request.Header.Set("Authorization", "Bearer "+apiKey)
	return nil
}

func (c *Client) endpoint(parts ...any) (string, error) {
	base, err := url.Parse(strings.TrimRight(c.config.BaseURL, "/"))
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	pathParts := []string{apiPrefix}
	var query url.Values
	for _, part := range parts {
		switch value := part.(type) {
		case string:
			trimmed := strings.Trim(value, "/")
			if trimmed != "" {
				pathParts = append(pathParts, url.PathEscape(trimmed))
			}
		case url.Values:
			query = value
		case nil:
		default:
			return "", fmt.Errorf("unsupported endpoint part %T", value)
		}
	}

	base.Path = strings.Join(pathParts, "/")
	base.RawQuery = query.Encode()
	return base.String(), nil
}

func (c *Client) projectID(projectID string) (string, error) {
	resolved := strings.TrimSpace(projectID)
	if resolved == "" {
		resolved = strings.TrimSpace(c.config.DefaultProjectID)
	}
	if resolved == "" {
		return "", errors.New("projectId is required or set SENTIARY_PROJECT_ID")
	}
	return resolved, nil
}

func (c *Client) projectAndTermID(projectID string, termID string) (string, string, error) {
	resolvedProjectID, err := c.projectID(projectID)
	if err != nil {
		return "", "", err
	}
	resolvedTermID := strings.TrimSpace(termID)
	if resolvedTermID == "" {
		return "", "", errors.New("stringId is required")
	}
	return resolvedProjectID, resolvedTermID, nil
}

func normalizeFormat(format string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(format))
	switch normalized {
	case "json", "android", "apple", "compose":
		return normalized, nil
	default:
		return "", fmt.Errorf("unsupported format %q; use json, android, apple, or compose", format)
	}
}

func setOptionalQuery(query url.Values, name string, value string) {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		query.Set(name, trimmed)
	}
}

func optionalString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

type contentTypedReader struct {
	io.Reader
	contentType string
}

func jsonBody(value any) io.Reader {
	data, _ := json.Marshal(value)
	return contentTypedReader{
		Reader:      bytes.NewReader(data),
		contentType: "application/json; charset=utf-8",
	}
}

func textBody(value string) io.Reader {
	return contentTypedReader{
		Reader:      strings.NewReader(value),
		contentType: "text/plain; charset=utf-8",
	}
}
