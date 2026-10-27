package sentiary

import (
	"os"
	"strings"
)

const defaultBaseURL = "https://api.sentiary.com"

type Config struct {
	BaseURL          string
	APIKey           string
	DefaultProjectID string
}

func ConfigFromEnv() Config {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("SENTIARY_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return Config{
		BaseURL:          baseURL,
		APIKey:           firstEnv("SENTIARY_USER_API_KEY", ""),
		DefaultProjectID: firstEnv("SENTIARY_PROJECT_ID", ""),
	}
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return value
		}
	}
	return ""
}
