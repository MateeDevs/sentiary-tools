package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MateeDevs/sentiary-tools/go/internal/buildinfo"
	"github.com/MateeDevs/sentiary-tools/go/internal/frontend/hosted"
	"github.com/MateeDevs/sentiary-tools/go/internal/sentiary"
)

func main() {
	if buildinfo.IsVersionCommand(os.Args[1:]) {
		fmt.Println(buildinfo.String())
		return
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	handler := hosted.NewHandler(sentiary.ConfigFromEnv(), httpClient, buildinfo.String())
	address := ":" + port
	log.Printf("Sentiary MCP listens on %s", address)
	if err := http.ListenAndServe(address, handler); err != nil {
		log.Printf("Sentiary MCP stopped: %v", err)
		os.Exit(1)
	}
}
