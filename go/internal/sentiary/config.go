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
	return Config{
		BaseURL:          BaseURLFromEnv(),
		APIKey:           firstEnv("SENTIARY_USER_API_KEY", ""),
		DefaultProjectID: firstEnv("SENTIARY_PROJECT_ID", ""),
	}
}

// BaseURLFromEnv reads operational configuration without evaluating credentials.
func BaseURLFromEnv() string {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("SENTIARY_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return baseURL
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
