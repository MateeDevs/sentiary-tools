package hosted

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MateeDevs/sentiary-tools/internal/frontend/mcpserver"
	"github.com/MateeDevs/sentiary-tools/internal/sentiary"
	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const sessionIdleTimeout = 30 * time.Minute

func NewHandler(baseConfig sentiary.Config, httpClient *http.Client, version string) http.Handler {
	mcpHandler := mcp.NewStreamableHTTPHandler(func(request *http.Request) *mcp.Server {
		config := baseConfig
		if apiKey := apiKeyFromRequest(request); apiKey != "" {
			config.APIKey = apiKey
		}
		return mcpserver.New(config, httpClient, version)
	}, &mcp.StreamableHTTPOptions{
		SessionTimeout: sessionIdleTimeout,
	})
	protectedMCPHandler := withAPIKey(baseConfig.APIKey, mcpHandler)

	mux := http.NewServeMux()
	mux.Handle("/", protectedMCPHandler)
	mux.HandleFunc("GET /healthz", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("ok"))
	})
	return mux
}

func withAPIKey(fallbackAPIKey string, handler http.Handler) http.Handler {
	verifyToken := func(_ context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
		identifier := sha256.Sum256([]byte(token))
		return &auth.TokenInfo{
			Expiration: time.Now().Add(time.Hour),
			UserID:     fmt.Sprintf("%x", identifier),
		}, nil
	}
	requireToken := auth.RequireBearerToken(verifyToken, nil)(handler)

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		apiKey := apiKeyFromRequest(request)
		if apiKey == "" {
			apiKey = strings.TrimSpace(fallbackAPIKey)
		}

		request = request.Clone(request.Context())
		request.Header.Set("Authorization", "Bearer "+apiKey)
		requireToken.ServeHTTP(writer, request)
	})
}

func apiKeyFromRequest(request *http.Request) string {
	if request == nil || request.Header == nil {
		return ""
	}

	if apiKey := strings.TrimSpace(request.Header.Get("X-Sentiary-User-Api-Key")); apiKey != "" {
		return apiKey
	}

	authorization := strings.TrimSpace(request.Header.Get("Authorization"))
	if authorization == "" {
		return ""
	}

	scheme, token, hasToken := strings.Cut(authorization, " ")
	if hasToken {
		switch strings.ToLower(scheme) {
		case "bearer", "ribbon":
			return strings.TrimSpace(token)
		}
	}

	switch strings.ToLower(authorization) {
	case "bearer", "ribbon":
		return ""
	default:
		return authorization
	}
}
