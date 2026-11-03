package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MateeDevs/sentiary-tools/go/internal/configfile"
)

func TestInitMultipleProjectsAndOutputs(t *testing.T) {
	directory := t.TempDir()
	app := New(nil)
	app.Stdin = strings.NewReader("mobile\np1\nen,cs\nn\nandroid\nandroid\n\nen=values,cs=values-cs\ny\ncompose\ncompose\n\nen=values,cs=values-cs\nn\ny\napple\np2\nen,cs\nn\nios\nxcstrings\n\nen\nn\nn\nn\ny\ny\n")
	app.Stdout, app.Stderr = io.Discard, io.Discard
	if err := app.Run(context.Background(), []string{"init", "--directory", directory}); err != nil {
		t.Fatal(err)
	}
	cfg, err := configfile.LoadOffline("", directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Projects) != 2 || len(cfg.Outputs) != 3 {
		t.Fatalf("wrong config: %+v", cfg)
	}
	if cfg.Outputs["ios"].SourceLanguage != "en" {
		t.Fatal("missing combined source")
	}
}

func TestInitCancellationWritesNothing(t *testing.T) {
	directory := t.TempDir()
	app := New(nil)
	app.Stdin = strings.NewReader("app\np1\nen\nn\nweb\njson\nweb/{language}.json\nn\nn\nn\ny\nn\n")
	app.Stdout, app.Stderr = io.Discard, io.Discard
	if err := app.Run(context.Background(), []string{"init", "--directory", directory}); err == nil {
		t.Fatal("expected cancellation")
	}
	entries, _ := os.ReadDir(directory)
	if len(entries) != 0 {
		t.Fatal("cancelled wizard wrote files")
	}
}

func TestInitCredentialsRedactedAndIgnorePreserved(t *testing.T) {
	directory := t.TempDir()
	app := New(nil)
	t.Setenv("SENTIARY_PROJECT_API_KEY", "secret-project")
	t.Setenv("SENTIARY_USER_API_KEY", "secret-personal")
	if err := os.WriteFile(filepath.Join(directory, ".gitignore"), []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}
	app.Stdin = strings.NewReader("app\np1\nen\ny\nweb\njson\nweb/{language}.json\nn\nn\ny\ny\ny\n")
	var out bytes.Buffer
	app.Stdout, app.Stderr = &out, &out
	if err := app.Run(context.Background(), []string{"init", "--directory", directory}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "secret-") {
		t.Fatal("leaked credentials")
	}
	ignore, _ := os.ReadFile(filepath.Join(directory, ".gitignore"))
	if string(ignore) != "existing\nsentiary.local.toml\n.sentiary/\n" {
		t.Fatalf("ignore changed: %s", ignore)
	}
	// Pull honors the project literal written by init before the local user credential.
	cfg, err := configfile.Load("", directory, func(string) string { return "" })
	if err != nil || cfg.Projects["app"].APIKey != "secret-project" {
		t.Fatalf("local credential: %v", err)
	}
	app.Stdin = strings.NewReader("")
	if err := app.Run(context.Background(), []string{"init", "--directory", directory}); err == nil {
		t.Fatal("overwrote config")
	}
}

func TestInitDotenvProjectKeyWritesOnlyEnvironmentReference(t *testing.T) {
	t.Setenv("SENTIARY_PROJECT_API_KEY", "")
	if err := os.Unsetenv("SENTIARY_PROJECT_API_KEY"); err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	app := New(nil)
	app.CredentialEnvironment = CredentialEnvironment{values: map[string]string{"SENTIARY_PROJECT_API_KEY": "dotenv-secret-project"}}
	app.Stdin = strings.NewReader("app\np1\nen\ny\nweb\njson\nweb/{language}.json\nn\nn\nn\ny\ny\n")
	var output bytes.Buffer
	app.Stdout, app.Stderr = &output, &output
	if err := app.Run(context.Background(), []string{"init", "--directory", directory}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(directory, "sentiary.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "dotenv-secret") || strings.Contains(output.String(), "dotenv-secret") {
		t.Fatal("dotenv credential escaped into config or output")
	}
	config, err := configfile.Load("", directory, app.CredentialEnvironment.Get)
	if err != nil {
		t.Fatal(err)
	}
	if config.Projects["app"].APIKey != "dotenv-secret-project" || !strings.Contains(string(data), "env = 'SENTIARY_PROJECT_API_KEY'") {
		t.Fatalf("expected functional environment reference, config: %s", data)
	}
	if !strings.Contains(output.String(), "--env-file again") {
		t.Fatal("missing future-invocation instructions")
	}
}
