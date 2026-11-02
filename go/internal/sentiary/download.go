package sentiary

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/MateeDevs/sentiary-tools/go/internal/download"
)

// DownloadRepository implements config-based exports without changing the legacy API interface.
type DownloadRepository struct {
	config Config
	http   *http.Client
}

func NewDownloadRepository(config Config, httpClient *http.Client) *DownloadRepository {
	return &DownloadRepository{config: config, http: httpClient}
}

func (repository *DownloadRepository) client(project download.Project) *Client {
	config := repository.config
	config.APIKey = project.APIKey
	return NewClient(config, repository.http)
}

func (repository *DownloadRepository) Languages(ctx context.Context, project download.Project) (languages []string, err error) {
	defer func() { err = redactCredential(err, project.APIKey) }()
	info, err := repository.client(project).GetProjectBatchInfo(ctx, project.ID)
	return info.Languages, err
}

func (repository *DownloadRepository) Server() string {
	return repository.client(download.Project{}).config.BaseURL
}

func (repository *DownloadRepository) State(ctx context.Context, project download.Project) (state download.State, err error) {
	defer func() { err = redactCredential(err, project.APIKey) }()
	client := repository.client(project)
	endpoint, err := client.endpoint("batch", project.ID, "export", "state", nil)
	if err != nil {
		return state, err
	}
	err = client.doJSON(ctx, http.MethodGet, endpoint, nil, authAPIKey, &state)
	return state, err
}

func (repository *DownloadRepository) Export(ctx context.Context, project download.Project, file download.File) ([]byte, error) {
	response, err := repository.ConditionalExport(ctx, project, file, "")
	return response.Content, err
}

func (repository *DownloadRepository) ConditionalExport(ctx context.Context, project download.Project, file download.File, etag string) (result download.ExportResponse, err error) {
	defer func() { err = redactCredential(err, project.APIKey) }()
	if err := file.Options.Validate(file.Format); err != nil {
		return result, err
	}
	format, ok := download.LookupFormat(file.Format)
	if !ok {
		return result, fmt.Errorf("unsupported export format")
	}
	if err := download.ValidateLanguages(file.Languages); err != nil {
		return result, err
	}
	query := url.Values{}
	query.Set("format", file.Format)
	query.Set("includeEmptyTranslations", fmt.Sprint(file.IncludeEmpty))
	switch file.Layout {
	case download.Combined:
		if !format.Combined || !slices.Contains(file.Languages, file.SourceLanguage) {
			return result, fmt.Errorf("invalid combined export selection")
		}
		for _, language := range file.Languages {
			query.Add("languageIds", language)
		}
		query.Set("sourceLanguageId", file.SourceLanguage)
	case download.PerLanguage:
		if len(file.Languages) != 1 {
			return result, fmt.Errorf("per_language export requires one language")
		}
		query.Set("languageId", file.Languages[0])
	default:
		return result, fmt.Errorf("unsupported export layout")
	}
	client := repository.client(project)
	// Pass path segments separately: endpoint escapes each segment.
	var endpoint string
	if file.Layout == download.Combined {
		endpoint, err = client.endpoint("batch", project.ID, "export", "combined", query)
	} else {
		endpoint, err = client.endpoint("batch", project.ID, "export", query)
	}
	if err != nil {
		return result, err
	}
	method := http.MethodGet
	var body io.Reader
	if !file.Options.IsDefault() {
		endpoint, err = client.endpoint("batch", project.ID, "export", "resources", nil)
		if err != nil {
			return result, err
		}
		payload := struct {
			Capability     string   `json:"capability"`
			Format         string   `json:"format"`
			Languages      []string `json:"languages"`
			SourceLanguage *string  `json:"sourceLanguage,omitempty"`
			IncludeEmpty   bool     `json:"includeEmptyTranslations"`
			*download.ExportOptions
		}{Capability: download.ResourceExportCapability, Format: file.Format, Languages: file.Languages,
			IncludeEmpty: file.IncludeEmpty, ExportOptions: file.Options.Canonical()}
		if file.Layout == download.Combined {
			payload.SourceLanguage = &file.SourceLanguage
		}
		encoded, encodeErr := json.Marshal(payload)
		if encodeErr != nil {
			return result, encodeErr
		}
		body = bytes.NewReader(encoded)
		method = http.MethodPost
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return result, err
	}
	if err = client.authorize(request, authAPIKey); err != nil {
		return result, err
	}
	if method == http.MethodPost {
		request.Header.Set("Content-Type", "application/json")
	}
	if etag != "" {
		request.Header.Set("If-None-Match", etag)
	}
	response, err := client.http.Do(request)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNotModified {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return result, fmt.Errorf("export failed: HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(message)))
	}
	if !file.Options.IsDefault() && response.Header.Get("X-Sentiary-Export-Capability") != download.ResourceExportCapability {
		return result, fmt.Errorf("server did not confirm %s", download.ResourceExportCapability)
	}
	result.ETag = response.Header.Get("ETag")
	result.State = download.State{ProtocolVersion: 1, ProjectGeneration: response.Header.Get("X-Sentiary-Project-Generation"),
		DataRevision: response.Header.Get("X-Sentiary-Data-Revision"), ExporterVersion: response.Header.Get("X-Sentiary-Exporter-Version")}
	result.NotModified = response.StatusCode == http.StatusNotModified
	result.Content, err = io.ReadAll(response.Body)
	if err != nil {
		return result, fmt.Errorf("read export response: %w", err)
	}
	return result, nil
}

type credentialError struct {
	cause error
	key   string
}

func (err credentialError) Error() string {
	return strings.ReplaceAll(err.cause.Error(), err.key, "[redacted]")
}
func (err credentialError) Unwrap() error { return err.cause }

func redactCredential(err error, key string) error {
	if err == nil || key == "" {
		return err
	}
	return credentialError{cause: err, key: key}
}
