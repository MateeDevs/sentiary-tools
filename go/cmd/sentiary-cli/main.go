package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MateeDevs/sentiary-tools/go/internal/buildinfo"
	"github.com/MateeDevs/sentiary-tools/go/internal/frontend/cli"
	"github.com/MateeDevs/sentiary-tools/go/internal/sentiary"
)

func main() {
	if buildinfo.IsVersionCommand(os.Args[1:]) {
		if len(os.Args) == 3 && os.Args[2] == "--json" {
			_ = json.NewEncoder(os.Stdout).Encode(struct {
				Version   string `json:"version"`
				Protocols []int  `json:"integrationProtocols"`
			}{buildinfo.String(), []int{1}})
		} else if len(os.Args) == 2 {
			fmt.Println(buildinfo.String())
		} else {
			fmt.Fprintln(os.Stderr, "Usage: sentiary-cli version [--json]")
			os.Exit(2)
		}
		return
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	config := sentiary.ConfigFromEnv()
	api := sentiary.NewClient(config, httpClient)
	app := cli.New(api)
	app.Downloads = sentiary.NewDownloadRepository(config, httpClient)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := app.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if cli.IsUsageError(err) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
