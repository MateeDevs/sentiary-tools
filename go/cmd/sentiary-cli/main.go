package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/MateeDevs/sentiary-tools/go/internal/buildinfo"
	"github.com/MateeDevs/sentiary-tools/go/internal/frontend/cli"
	"github.com/MateeDevs/sentiary-tools/go/internal/sentiary"
	"github.com/MateeDevs/sentiary-tools/go/internal/updates"
	"golang.org/x/term"
)

func main() {
	args, environment, err := cli.ReadEnvironment(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if cli.IsUsageError(err) {
			os.Exit(2)
		}
		os.Exit(1)
	}
	if buildinfo.IsVersionCommand(args) {
		if len(args) == 2 && args[1] == "--json" {
			_ = json.NewEncoder(os.Stdout).Encode(struct {
				Version   string `json:"version"`
				Protocols []int  `json:"integrationProtocols"`
			}{buildinfo.String(), []int{1}})
		} else if len(args) == 1 {
			fmt.Println(buildinfo.String())
		} else {
			fmt.Fprintln(os.Stderr, "Usage: sentiary-cli version [--json]")
			os.Exit(2)
		}
		return
	}
	updater := newUpdater()
	if len(args) > 0 && args[0] == updates.BackgroundArgument {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		updater.Background(ctx)
		return
	}
	if len(args) > 0 && args[0] == "update" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if err := updater.Manual(ctx, args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	var config sentiary.Config
	if len(args) > 0 && args[0] == "pull" {
		// Configured pulls resolve their own credentials, or skip them entirely offline.
		config = sentiary.Config{BaseURL: sentiary.BaseURLFromEnv()}
	} else {
		config = sentiary.ConfigFromEnv()
		config.APIKey = environment.Get("SENTIARY_USER_API_KEY")
	}
	api := sentiary.NewClient(config, httpClient)
	app := cli.New(api)
	app.CredentialEnvironment = environment
	app.Downloads = sentiary.NewDownloadRepository(config, httpClient)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := app.Run(ctx, args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if cli.IsUsageError(err) {
			os.Exit(2)
		}
		os.Exit(1)
	}
	updater.AfterSuccess(ctx, args)
}

func newUpdater() *updates.Manager {
	executable, _ := os.Executable()
	directory, err := os.UserCacheDir()
	disabled := os.Getenv("SENTIARY_NO_UPDATE") == "1" || os.Getenv("CI") != "" || err != nil
	releases := updates.Releases{Client: &http.Client{Timeout: 90 * time.Second}}
	return &updates.Manager{
		Version: buildinfo.String(), Executable: executable,
		Directory:   filepath.Join(directory, "sentiary", "updates"),
		Interactive: term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) && term.IsTerminal(int(os.Stderr.Fd())),
		Disabled:    disabled, Input: os.Stdin, Output: os.Stderr,
		Now: time.Now, Latest: releases.Latest, Install: releases.Install,
		Owned: updates.Owned, Spawn: updates.SpawnBackground,
	}
}
