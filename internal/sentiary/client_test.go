package sentiary

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestSearchStringsKeepsTranslationMatches(t *testing.T) {
	var searchRequest *http.Request
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		searchRequest = request
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
				"items": [{
					"id": "term-id",
					"name": "welcome_message",
					"translations": [{"languageId": "cs-CZ", "text": "Ahoj", "updated": ""}],
					"updated": ""
				}],
				"pagination": {"size": 10, "cursor": "2", "next": "3", "previous": "1"}
			}`)),
			Request: request,
		}, nil
	})}
	client := NewClient(Config{BaseURL: "https://api.test", APIKey: "test-key"}, httpClient)

	output, err := client.SearchStrings(context.Background(), SearchStringsInput{
		ProjectID:          "project-id",
		Query:              "Ahoj",
		Size:               10,
		Cursor:             "2",
		IncludeLanguageIDs: []string{"cs-CZ"},
		Filter:             "TRANSLATED",
		FilterLanguage:     "cs-CZ",
		Order:              "UPDATED_DESC",
	})
	if err != nil {
		t.Fatalf("Search strings: %v", err)
	}
	if len(output.Items) != 1 {
		t.Fatalf("Search result count = %d, want 1", len(output.Items))
	}
	if output.Items[0].Name != "welcome_message" {
		t.Fatalf("Search result name = %q", output.Items[0].Name)
	}

	query := searchRequest.URL.Query()
	checks := map[string]string{
		"query":           "Ahoj",
		"size":            "10",
		"cursor":          "2",
		"includeLanguage": "cs-CZ",
		"filter":          "TRANSLATED",
		"filterLanguage":  "cs-CZ",
		"order":           "UPDATED_DESC",
	}
	for name, want := range checks {
		if got := query.Get(name); got != want {
			t.Errorf("Search parameter %s = %q, want %q", name, got, want)
		}
	}
	if got := searchRequest.Header.Get("Authorization"); got != "Bearer test-key" {
		t.Fatalf("Authorization = %q", got)
	}
}

func TestSearchStringsRequiresQuery(t *testing.T) {
	client := NewClient(Config{}, nil)
	if _, err := client.SearchStrings(context.Background(), SearchStringsInput{}); err == nil {
		t.Fatal("Search strings accepted an empty query")
	}
}
