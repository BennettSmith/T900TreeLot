package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/session"
	"github.com/troop900/treelot/internal/web/handlers"
	"github.com/troop900/treelot/internal/web/middleware"
	"github.com/troop900/treelot/internal/web/views"
)

func TestDevelopmentRoutesAreUnavailableInProduction(t *testing.T) {
	t.Parallel()

	server := newServer(t, handlers.Options{})
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

		// Unavailable development routes should not expose the gallery. Unsafe
		// methods may be rejected by CSRF before the mux returns 404.
		if response.Code != http.StatusNotFound && response.Code != http.StatusForbidden {
			t.Errorf("%s %s status = %d, want 404 or 403", test.method, test.path, response.Code)
		}
	}
}

func TestDevelopmentGalleryAndParityResponses(t *testing.T) {
	t.Parallel()

	server := newServer(t, handlers.Options{Development: true})
	cookie, csrf := establishSession(t, server)

	t.Run("gallery", func(t *testing.T) {
		response := request(t, server, http.MethodGet, "/_dev/components", "", nil, cookie)
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
		if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/html") {
			t.Errorf("Content-Type = %q, want text/html", contentType)
		}
	})

	t.Run("normal form receives full page", func(t *testing.T) {
		form := url.Values{"message": {"Signal parity confirmed"}, middleware.CSRFFormField: {csrf}}
		response := request(t, server, http.MethodPost, "/_dev/parity", form.Encode(), map[string]string{"Content-Type": "application/x-www-form-urlencoded"}, cookie)
		if body := response.Body.String(); !strings.Contains(body, "<!doctype html>") || !strings.Contains(body, "Signal parity confirmed") {
			t.Errorf("full-page response missing expected content: %s", body)
		}
	})

	t.Run("HTMX receives equivalent fragment", func(t *testing.T) {
		form := url.Values{"message": {"Signal parity confirmed"}, middleware.CSRFFormField: {csrf}}
		response := request(t, server, http.MethodPost, "/_dev/parity", form.Encode(), map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"HX-Request":   "true",
		}, cookie)
		body := response.Body.String()
		if strings.Contains(body, "<!doctype html>") || !strings.Contains(body, `id="parity-result"`) || !strings.Contains(body, "Signal parity confirmed") {
			t.Errorf("fragment response framing or content is wrong: %s", body)
		}
	})
}

func TestInfrastructureRoutesAndBrowserHeaders(t *testing.T) {
	t.Parallel()

	server := newServer(t, handlers.Options{})

	for _, test := range []struct {
		path        string
		contentType string
		body        string
	}{
		{path: "/health/live", contentType: "text/plain", body: "ok\n"},
		{path: "/health/ready", contentType: "text/plain", body: "ready\n"},
		{path: "/static/app.css", contentType: "text/css", body: "--color-canvas"},
	} {
		response := request(t, server, http.MethodGet, test.path, "", nil, nil)
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

func TestHomeAndSmokeJourney(t *testing.T) {
	t.Parallel()

	server := newServer(t, handlers.Options{})
	home := request(t, server, http.MethodGet, "/", "", nil, nil)
	if home.Code != http.StatusOK {
		t.Fatalf("home status = %d", home.Code)
	}
	body := home.Body.String()
	if !strings.Contains(body, "Troop 900 Tree Lot") || !strings.Contains(body, `aria-label="Primary"`) || !strings.Contains(body, `href="/sign-in"`) {
		t.Fatalf("home missing brand or navigation: %s", body)
	}
	cookie := firstCookie(home, middleware.SessionCookieName)
	if cookie == nil {
		t.Fatal("missing session cookie")
	}
	csrf := csrfFromHome(body)
	if csrf == "" {
		t.Fatal("missing csrf token")
	}

	form := url.Values{"message": {"lot ready"}, middleware.CSRFFormField: {csrf}}
	smoke := request(t, server, http.MethodPost, "/smoke", form.Encode(), map[string]string{"Content-Type": "application/x-www-form-urlencoded"}, cookie)
	if smoke.Code != http.StatusSeeOther {
		t.Fatalf("smoke status = %d, want 303", smoke.Code)
	}
	if location := smoke.Header().Get("Location"); location != "/?message=lot+ready" {
		t.Fatalf("Location = %q", location)
	}
}

func TestReadyFailsWhenDependencyUnready(t *testing.T) {
	t.Parallel()

	server := newServer(t, handlers.Options{
		Ready: func(context.Context) error { return errors.New("db down") },
	})
	response := request(t, server, http.MethodGet, "/health/ready", "", nil, nil)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.Code)
	}
	if strings.Contains(strings.ToLower(response.Body.String()), "postgres://") {
		t.Fatal("readiness error leaked connection details")
	}
}

func TestTestControlClockRequiresKeyAndAdvances(t *testing.T) {
	t.Parallel()

	controllable := clock.NewControllable(time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC))
	server := newServer(t, handlers.Options{
		TestControlEnabled: true,
		TestControlKey:     "secret",
		Clock:              controllable,
		ControllableClock:  controllable,
	})

	unauthorized := request(t, server, http.MethodPost, "/_test/clock/advance", `{"duration":"1h"}`, map[string]string{"Content-Type": "application/json"}, nil)
	if unauthorized.Code != http.StatusForbidden {
		t.Fatalf("unauthorized status = %d", unauthorized.Code)
	}

	advanced := request(t, server, http.MethodPost, "/_test/clock/advance", `{"duration":"1h"}`, map[string]string{
		"Content-Type":       "application/json",
		"X-Test-Control-Key": "secret",
	}, nil)
	if advanced.Code != http.StatusOK {
		t.Fatalf("advance status = %d", advanced.Code)
	}
	var payload map[string]string
	if err := json.Unmarshal(advanced.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !strings.HasPrefix(payload["now"], "2026-07-30T13:00:00") {
		t.Fatalf("now = %q", payload["now"])
	}
}

func TestTestControlAbsentInProduction(t *testing.T) {
	t.Parallel()

	server := newServer(t, handlers.Options{})
	response := request(t, server, http.MethodGet, "/_test/clock", "", nil, nil)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

func TestIdentityFixtureRoleRequiresTestControlKey(t *testing.T) {
	t.Parallel()
	fixture := &fakeIdentityFixture{}
	server := newServer(t, handlers.Options{
		TestControlEnabled: true,
		TestControlKey:     "secret",
		IdentityFixture:    fixture,
	})
	body := `{"email":"manager@example.org","role":"family_manager"}`

	unauthorized := request(t, server, http.MethodPost, "/_test/identity/role", body, map[string]string{"Content-Type": "application/json"}, nil)
	if unauthorized.Code != http.StatusForbidden {
		t.Fatalf("unauthorized status = %d", unauthorized.Code)
	}
	authorized := request(t, server, http.MethodPost, "/_test/identity/role", body, map[string]string{
		"Content-Type":       "application/json",
		"X-Test-Control-Key": "secret",
	}, nil)
	if authorized.Code != http.StatusOK {
		t.Fatalf("authorized status = %d body=%q", authorized.Code, authorized.Body.String())
	}
	if fixture.command.Email != "manager@example.org" || fixture.command.Role != "family_manager" {
		t.Fatalf("command = %#v", fixture.command)
	}
}

func TestParityRejectsMalformedFormAndUsesDefaultMessage(t *testing.T) {
	t.Parallel()

	server := newServer(t, handlers.Options{Development: true})
	cookie, csrf := establishSession(t, server)

	malformed := request(t, server, http.MethodPost, "/_dev/parity", "%", map[string]string{"Content-Type": "application/x-www-form-urlencoded"}, cookie)
	if malformed.Code != http.StatusBadRequest && malformed.Code != http.StatusForbidden {
		t.Errorf("malformed form status = %d, want 400 or 403", malformed.Code)
	}

	form := url.Values{middleware.CSRFFormField: {csrf}}
	defaulted := request(t, server, http.MethodPost, "/_dev/parity", form.Encode(), map[string]string{"Content-Type": "application/x-www-form-urlencoded"}, cookie)
	if !strings.Contains(defaulted.Body.String(), "Signal parity confirmed") {
		t.Error("empty form did not receive the representative default message")
	}
}

func TestRenderFailureReturnsGenericServerError(t *testing.T) {
	t.Parallel()

	server := handlers.New(failingRenderer{}, handlers.Options{
		Development: true,
		Sessions:    session.NewMemoryStore(clock.System(), time.Hour),
	})
	response := request(t, server, http.MethodGet, "/_dev/components", "", nil, nil)

	if response.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if body := response.Body.String(); body != "Unable to render this page.\n" {
		t.Errorf("body = %q, want generic error", body)
	}
}

type failingRenderer struct{}

func (failingRenderer) Home(context.Context, io.Writer, views.Home) error {
	return errors.New("representative render failure")
}

func (failingRenderer) ComponentGallery(context.Context, io.Writer, views.Gallery) error {
	return errors.New("representative render failure")
}

func (failingRenderer) ParityResult(context.Context, io.Writer, views.Gallery) error {
	return errors.New("representative render failure")
}

func newServer(t *testing.T, options handlers.Options) http.Handler {
	t.Helper()
	renderer, err := views.NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	if options.Sessions == nil {
		options.Sessions = session.NewMemoryStore(clock.System(), time.Hour)
	}
	return handlers.New(renderer, options)
}

func establishSession(t *testing.T, handler http.Handler) (*http.Cookie, string) {
	t.Helper()
	response := request(t, handler, http.MethodGet, "/", "", nil, nil)
	cookie := firstCookie(response, middleware.SessionCookieName)
	if cookie == nil {
		t.Fatal("missing session cookie")
	}
	csrf := csrfFromHome(response.Body.String())
	if csrf == "" {
		t.Fatal("missing csrf token")
	}
	return cookie, csrf
}

func csrfFromHome(body string) string {
	const marker = `name="csrf_token" value="`
	start := strings.Index(body, marker)
	if start < 0 {
		return ""
	}
	start += len(marker)
	end := strings.Index(body[start:], `"`)
	if end < 0 {
		return ""
	}
	return body[start : start+end]
}

func firstCookie(response *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func request(t *testing.T, handler http.Handler, method, target, body string, headers map[string]string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	_, _ = io.Copy(io.Discard, response.Result().Body)
	return response
}

type fakeIdentityFixture struct {
	command application.SetFixtureRoleCommand
}

func (f *fakeIdentityFixture) SetRole(_ context.Context, command application.SetFixtureRoleCommand) error {
	f.command = command
	return nil
}

func (*fakeIdentityFixture) RevokeSessions(context.Context, string) error {
	return nil
}
