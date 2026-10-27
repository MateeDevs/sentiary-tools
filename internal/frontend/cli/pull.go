package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/MateeDevs/sentiary-tools/internal/configfile"
	"github.com/MateeDevs/sentiary-tools/internal/download"
	"github.com/MateeDevs/sentiary-tools/internal/exportcache"
	"github.com/MateeDevs/sentiary-tools/internal/fileoutput"
)

type outputSelection []string

func (selection *outputSelection) String() string { return fmt.Sprint([]string(*selection)) }
func (selection *outputSelection) Set(name string) error {
	*selection = append(*selection, name)
	return nil
}

func (app *App) runPull(ctx context.Context, args []string) (err error) {
	flags := app.newFlagSet("pull")
	configPath := flags.String("config", "", "Config path. Otherwise searches the working directory and parents.")
	offline := flags.Bool("offline", false, "Restore from one completed cached download set without network access.")
	cacheDir := flags.String("cache-dir", "", "Export cache directory; overrides SENTIARY_CACHE_DIR.")
	dryRun := flags.Bool("dry-run", false, "Validate and print the resolved download plan as JSON without writing files.")
	jsonOutput := flags.Bool("json", false, "Emit integration protocol v1 JSON on stdout.")
	integrationPath := flags.String("integration-request", "", "Protocol v1 build integration request JSON file.")
	force := flags.Bool("force", false, "Take ownership of existing destinations and discard edits to managed files.")
	var selected outputSelection
	flags.Var(&selected, "output", "Named output to download. Repeat to select multiple outputs; defaults to all outputs.")
	if err := app.parse(flags, args); err != nil {
		return err
	}
	if app.Downloads == nil && !*offline {
		return errors.New("download repository is not configured")
	}
	directory, err := os.Getwd()
	if err != nil {
		return err
	}
	var config download.Config
	if *offline {
		config, err = configfile.LoadOffline(*configPath, directory)
	} else {
		config, err = configfile.Load(*configPath, directory, os.Getenv("SENTIARY_USER_API_KEY"))
	}
	if err != nil {
		return err
	}
	request, outputs, err := readIntegration(*integrationPath, config)
	if err != nil {
		return err
	}
	if request != nil {
		if len(selected) != 0 {
			return errors.New("--output and --integration-request cannot be combined")
		}
		selected = outputs
	}
	// Hold a kernel-backed lock across fetching, cache publication, and target writes.
	// Preview has no output state and does not acquire the lock.
	var store *fileoutput.Store
	if !*dryRun {
		store, err = fileoutput.OpenContext(ctx, config, selected, *force)
		if err != nil {
			return err
		}
		defer func() { err = errors.Join(err, store.Close()) }()
	}
	server := app.DownloadServer
	repository, ok := app.Downloads.(download.SnapshotRepository)
	if ok {
		server = repository.Server()
	}
	cache, err := exportcache.Open(*cacheDir, server)
	if err != nil {
		return err
	}
	var downloaded []download.DownloadedFile
	var completed time.Time
	if *offline {
		downloaded, completed, err = download.RestoreSnapshot(ctx, config, selected, cache)
	} else if *dryRun {
		// Dry runs resolve languages without downloading or publishing a set.
		files, err := download.Plan(ctx, config, selected, app.Downloads)
		if err != nil {
			return err
		}
		if err := request.validate(files); err != nil {
			return err
		}
		return app.writeJSON(downloadPreview(files))
	} else {
		if !ok {
			return errors.New("download repository does not support export/state protocol v1")
		}
		downloaded, err = download.FetchSnapshot(ctx, config, selected, repository, cache)
	}
	if err != nil {
		return err
	}
	files := make([]download.File, len(downloaded))
	for i, file := range downloaded {
		files[i] = file.File
	}
	if *dryRun {
		if err := request.validate(files); err != nil {
			return err
		}
		return app.writeJSON(downloadPreview(files))
	}
	if *offline {
		fmt.Fprintf(app.Stderr, "Restoring download set completed %s (%s old).\n", completed.Format(time.RFC3339), time.Since(completed).Round(time.Second))
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := request.validate(files); err != nil {
		return err
	}
	if err := store.Validate(files); err != nil {
		return err
	}
	if err := store.Commit(downloaded); err != nil {
		return err
	}
	if *jsonOutput {
		result := integrationResult{ProtocolVersion: integrationProtocol, Files: []integrationFile{}}
		for _, file := range downloaded {
			result.Files = append(result.Files, integrationFile{file.Output, file.Path, download.Digest(file.Content)})
		}
		return app.writeJSON(result)
	}
	_, err = fmt.Fprintf(app.Stdout, "Downloaded %d file(s).\n", len(files))
	return err
}

type plannedDownload struct {
	Output         string          `json:"output"`
	Project        string          `json:"project"`
	Format         string          `json:"format"`
	Layout         download.Layout `json:"layout"`
	LanguageIDs    []string        `json:"languageIds"`
	SourceLanguage string          `json:"sourceLanguageId,omitempty"`
	IncludeEmpty   bool            `json:"includeEmptyTranslations"`
	Path           string          `json:"path"`
}

func downloadPreview(files []download.File) []plannedDownload {
	result := make([]plannedDownload, 0, len(files))
	for _, file := range files {
		result = append(result, plannedDownload{file.Output, file.Project, file.Format, file.Layout, file.Languages,
			file.SourceLanguage, file.IncludeEmpty, file.Path})
	}
	return result
}
