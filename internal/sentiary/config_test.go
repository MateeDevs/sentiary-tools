package sentiary

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestUserAPIKeyFromEnvUsesBearerHeader(t *testing.T) {
	t.Setenv("SENTIARY_USER_API_KEY", "  fixture-user-key  ")
	t.Setenv("SENTIARY_PROJECT_ID", " fixture-project ")
	config := ConfigFromEnv()
	if config.DefaultProjectID != "fixture-project" {
		t.Fatal("Default project does not match the environment")
	}
	calls := 0
	client := NewClient(config, &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		calls++
		if request.Header.Get("Authorization") != "Bearer fixture-user-key" {
			t.Error("Request does not contain the user key in the Bearer header")
		}
		if strings.Contains(request.URL.String(), "fixture-user-key") {
			t.Error("Request contains the key in its URL")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`[]`)),
			Header:     make(http.Header),
		}, nil
	})})
	if _, err := client.ListProjects(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListLanguages(context.Background(), ListLanguagesInput{}); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("Request count = %d, want 2", calls)
	}
}
