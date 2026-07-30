package middleware_test

import (
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

func TestSessionCSRFProtectsStateChangingRequests(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour)
	var sawSession bool
	handler := middleware.SessionCSRF(store, false, http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		current := middleware.FromContext(request.Context())
		if current == nil || current.CSRFToken == "" {
			t.Fatal("missing session in context")
		}
		sawSession = true
		_, _ = io.WriteString(response, current.CSRFToken)
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
}

func firstCookie(response *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}
