package hosted

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/MateeDevs/sentiary-tools/internal/sentiary"
)

func TestIdleSessionExpires(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		handler := NewHandler(sentiary.Config{}, http.DefaultClient, "test")
		sessionID := initializeSession(t, handler)
		defer sessionRequest(handler, http.MethodDelete, sessionID, "")

		time.Sleep(31 * time.Minute)
		synctest.Wait()

		response := sessionRequest(handler, http.MethodPost, sessionID, `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`)
		if response.Code != http.StatusNotFound {
			t.Fatalf("Expired session status = %d, want %d", response.Code, http.StatusNotFound)
		}
	})
}

func TestSessionActivityResetsIdleTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		handler := NewHandler(sentiary.Config{}, http.DefaultClient, "test")
		sessionID := initializeSession(t, handler)
		defer sessionRequest(handler, http.MethodDelete, sessionID, "")

		for range 2 {
			time.Sleep(20 * time.Minute)
			response := sessionRequest(handler, http.MethodPost, sessionID, `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`)
			if response.Code != http.StatusOK {
				t.Fatalf("Active session status = %d, want %d", response.Code, http.StatusOK)
			}
		}

		time.Sleep(31 * time.Minute)
		synctest.Wait()

		response := sessionRequest(handler, http.MethodPost, sessionID, `{"jsonrpc":"2.0","id":3,"method":"tools/list","params":{}}`)
		if response.Code != http.StatusNotFound {
			t.Fatalf("Expired session status = %d, want %d", response.Code, http.StatusNotFound)
		}
	})
}

func initializeSession(t *testing.T, handler http.Handler) string {
	t.Helper()
	response := sessionRequest(handler, http.MethodPost, "", `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"test"}}}`)
	if response.Code != http.StatusOK {
		t.Fatalf("Initialize status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	sessionID := response.Header().Get("Mcp-Session-Id")
	if sessionID == "" {
		t.Fatal("Initialize did not return a session ID")
	}
	response = sessionRequest(handler, http.MethodPost, sessionID, `{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	if response.Code != http.StatusAccepted {
		sessionRequest(handler, http.MethodDelete, sessionID, "")
		t.Fatalf("Initialized status = %d, want %d", response.Code, http.StatusAccepted)
	}
	return sessionID
}

func sessionRequest(handler http.Handler, method, sessionID, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, "https://mcp.test/", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer test-key")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	request.Header.Set("Mcp-Session-Id", sessionID)
	request.Header.Set("Mcp-Protocol-Version", "2025-06-18")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
