package handlers_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/troop900/treelot/internal/web/handlers"
	"github.com/troop900/treelot/internal/web/views"
)

func TestDevelopmentRoutesAreUnavailableInProduction(t *testing.T) {
	t.Parallel()

	server := newServer(t, false)
	for _, test := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/_dev/components"},
		{method: http.MethodPost, path: "/_dev/parity"},
	} {
		request := httptest.NewRequest(test.method, test.path, nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		if response.Code != http.StatusNotFound {
			t.Errorf("%s %s status = %d, want %d", test.method, test.path, response.Code, http.StatusNotFound)
		}
	}
}

func TestDevelopmentGalleryAndParityResponses(t *testing.T) {
	t.Parallel()

	server := newServer(t, true)

	t.Run("gallery", func(t *testing.T) {
		response := request(t, server, http.MethodGet, "/_dev/components", "", nil)
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
		if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/html") {
			t.Errorf("Content-Type = %q, want text/html", contentType)
		}
	})

	t.Run("normal form receives full page", func(t *testing.T) {
		response := request(t, server, http.MethodPost, "/_dev/parity", "message=Signal+parity+confirmed", map[string]string{"Content-Type": "application/x-www-form-urlencoded"})
		if body := response.Body.String(); !strings.Contains(body, "<!doctype html>") || !strings.Contains(body, "Signal parity confirmed") {
			t.Errorf("full-page response missing expected content: %s", body)
		}
	})

	t.Run("HTMX receives equivalent fragment", func(t *testing.T) {
		response := request(t, server, http.MethodPost, "/_dev/parity", "message=Signal+parity+confirmed", map[string]string{"Content-Type": "application/x-www-form-urlencoded", "HX-Request": "true"})
		body := response.Body.String()
		if strings.Contains(body, "<!doctype html>") || !strings.Contains(body, `id="parity-result"`) || !strings.Contains(body, "Signal parity confirmed") {
			t.Errorf("fragment response framing or content is wrong: %s", body)
		}
	})
}

func TestInfrastructureRoutesAndBrowserHeaders(t *testing.T) {
	t.Parallel()

	server := newServer(t, false)

	for _, test := range []struct {
		path        string
		contentType string
		body        string
	}{
		{path: "/health/live", contentType: "text/plain", body: "ok\n"},
		{path: "/static/app.css", contentType: "text/css", body: "--color-canvas"},
	} {
		response := request(t, server, http.MethodGet, test.path, "", nil)
		if response.Code != http.StatusOK {
			t.Errorf("%s status = %d, want %d", test.path, response.Code, http.StatusOK)
		}
		if got := response.Header().Get("Content-Type"); !strings.HasPrefix(got, test.contentType) {
			t.Errorf("%s Content-Type = %q, want prefix %q", test.path, got, test.contentType)
		}
		if !strings.Contains(response.Body.String(), test.body) {
			t.Errorf("%s response missing %q", test.path, test.body)
		}
		if got := response.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("%s X-Content-Type-Options = %q, want nosniff", test.path, got)
		}
	}
}

func TestParityRejectsMalformedFormAndUsesDefaultMessage(t *testing.T) {
	t.Parallel()

	server := newServer(t, true)
	malformed := request(t, server, http.MethodPost, "/_dev/parity", "%", map[string]string{"Content-Type": "application/x-www-form-urlencoded"})
	if malformed.Code != http.StatusBadRequest {
		t.Errorf("malformed form status = %d, want %d", malformed.Code, http.StatusBadRequest)
	}

	defaulted := request(t, server, http.MethodPost, "/_dev/parity", "", map[string]string{"Content-Type": "application/x-www-form-urlencoded"})
	if !strings.Contains(defaulted.Body.String(), "Signal parity confirmed") {
		t.Error("empty form did not receive the representative default message")
	}
}

func TestRenderFailureReturnsGenericServerError(t *testing.T) {
	t.Parallel()

	server := handlers.New(failingRenderer{}, handlers.Options{Development: true})
	response := request(t, server, http.MethodGet, "/_dev/components", "", nil)

	if response.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if body := response.Body.String(); body != "Unable to render this page.\n" {
		t.Errorf("body = %q, want generic error", body)
	}
}

type failingRenderer struct{}

func (failingRenderer) ComponentGallery(context.Context, io.Writer, views.Gallery) error {
	return errors.New("representative render failure")
}

func (failingRenderer) ParityResult(context.Context, io.Writer, views.Gallery) error {
	return errors.New("representative render failure")
}

func newServer(t *testing.T, development bool) http.Handler {
	t.Helper()
	renderer, err := views.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	return handlers.New(renderer, handlers.Options{Development: development})
}

func request(t *testing.T, handler http.Handler, method, target, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	_, _ = io.Copy(io.Discard, response.Result().Body)
	return response
}
