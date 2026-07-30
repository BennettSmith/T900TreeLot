package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/troop900/treelot/internal/web/middleware"
)

func TestBrowserHeaders(t *testing.T) {
	t.Parallel()

	handler := middleware.BrowserHeaders(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := response.Header().Get("Content-Security-Policy"); got == "" {
		t.Fatal("missing CSP")
	}
	if got := response.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q", got)
	}
}
