package middleware_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/session"
	"github.com/troop900/treelot/internal/web/middleware"
)

type failingSessionStore struct {
	getErr error
}

func (s failingSessionStore) Create(ctx context.Context) (session.Session, string, error) {
	_ = ctx
	return session.Session{}, "", errors.New("create should not be called")
}

func (s failingSessionStore) Get(ctx context.Context, rawToken string) (session.Session, error) {
	_ = ctx
	_ = rawToken
	return session.Session{}, s.getErr
}

func TestSessionCSRFProtectsStateChangingRequests(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	var sawSession bool
	handler := middleware.SessionCSRF(store, false, http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		current := middleware.FromContext(request.Context())
		if current == nil || current.CSRFToken == "" {
			t.Fatal("missing session in context")
		}
		sawSession = true
		body, _ := io.ReadAll(request.Body)
		_, _ = io.WriteString(response, current.CSRFToken+string(body))
	}))

	getResponse := httptest.NewRecorder()
	getRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(getResponse, getRequest)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("GET status = %d", getResponse.Code)
	}
	cookie := firstCookie(getResponse, middleware.SessionCookieName)
	if cookie == nil {
		t.Fatal("missing session cookie")
	}
	csrfToken := getResponse.Body.String()
	if csrfToken == "" {
		t.Fatal("missing csrf token body")
	}
	if !sawSession {
		t.Fatal("handler not invoked")
	}

	badPost := httptest.NewRecorder()
	badRequest := httptest.NewRequest(http.MethodPost, "/smoke", strings.NewReader("message=hi"))
	badRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	badRequest.AddCookie(cookie)
	handler.ServeHTTP(badPost, badRequest)
	if badPost.Code != http.StatusForbidden {
		t.Fatalf("POST without CSRF status = %d, want 403", badPost.Code)
	}

	form := url.Values{}
	form.Set("message", "hi")
	form.Set(middleware.CSRFFormField, csrfToken)
	goodPost := httptest.NewRecorder()
	goodRequest := httptest.NewRequest(http.MethodPost, "/smoke", strings.NewReader(form.Encode()))
	goodRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	goodRequest.AddCookie(cookie)
	handler.ServeHTTP(goodPost, goodRequest)
	if goodPost.Code != http.StatusOK {
		t.Fatalf("POST with CSRF status = %d, want 200", goodPost.Code)
	}

	jsonPost := httptest.NewRecorder()
	jsonRequest := httptest.NewRequest(http.MethodPost, "/json", strings.NewReader(`{"csrf_token":"`+csrfToken+`","message":"hi"}`))
	jsonRequest.Header.Set("Content-Type", "application/json")
	jsonRequest.AddCookie(cookie)
	handler.ServeHTTP(jsonPost, jsonRequest)
	if jsonPost.Code != http.StatusOK {
		t.Fatalf("POST with JSON CSRF status = %d, want 200", jsonPost.Code)
	}
	if !strings.Contains(jsonPost.Body.String(), csrfToken) {
		t.Fatal("handler could not read body after JSON CSRF validation")
	}
}

func TestSessionCSRFDoesNotReplaceCookieOnStoreFailure(t *testing.T) {
	store := failingSessionStore{getErr: errors.New("database unavailable")}
	handler := middleware.SessionCSRF(store, false, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler should not run when session lookup fails")
	}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: "existing-session-token"})
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.Code)
	}
	if cookie := firstCookie(response, middleware.SessionCookieName); cookie != nil {
		t.Fatalf("unexpected session cookie replacement: %#v", cookie)
	}
}

func TestSessionCSRFCreatesSessionWhenMissing(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	handler := middleware.SessionCSRF(store, false, http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: "stale-unknown-token"})
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", response.Code)
	}
	if cookie := firstCookie(response, middleware.SessionCookieName); cookie == nil || cookie.Value == "" {
		t.Fatal("expected a new session cookie for unknown token")
	}
}

func firstCookie(response *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}
