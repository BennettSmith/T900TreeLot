package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/troop900/treelot/internal/web/middleware"
)

func TestCanonicalHostRedirectsNoncanonicalHosts(t *testing.T) {
	t.Parallel()

	canonical, err := url.Parse("https://treelot.troop900livermore.org")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	called := false
	handler := middleware.CanonicalHost(canonical, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	request := httptest.NewRequest(http.MethodGet, "/sign-in?next=%2Ffamily", nil)
	request.Host = "treelot-web.onrender.com"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if called {
		t.Fatal("next handler should not run for noncanonical host")
	}
	if response.Code != http.StatusMovedPermanently {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMovedPermanently)
	}
	location := response.Header().Get("Location")
	want := "https://treelot.troop900livermore.org/sign-in?next=%2Ffamily"
	if location != want {
		t.Fatalf("Location = %q, want %q", location, want)
	}
}

func TestCanonicalHostAllowsCanonicalHostAndExemptsHealth(t *testing.T) {
	t.Parallel()

	canonical, err := url.Parse("https://treelot.troop900livermore.org")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	t.Run("canonical host passes through", func(t *testing.T) {
		called := false
		handler := middleware.CanonicalHost(canonical, http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			called = true
			response.WriteHeader(http.StatusNoContent)
		}))
		request := httptest.NewRequest(http.MethodGet, "/account", nil)
		request.Host = "treelot.troop900livermore.org"
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if !called || response.Code != http.StatusNoContent {
			t.Fatalf("called=%v status=%d", called, response.Code)
		}
	})

	t.Run("health live on noncanonical host is not redirected", func(t *testing.T) {
		called := false
		handler := middleware.CanonicalHost(canonical, http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			called = true
			response.WriteHeader(http.StatusOK)
			_, _ = response.Write([]byte("ok\n"))
		}))
		request := httptest.NewRequest(http.MethodGet, "/health/live", nil)
		request.Host = "treelot-web.onrender.com"
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if !called || response.Code != http.StatusOK {
			t.Fatalf("called=%v status=%d", called, response.Code)
		}
		if location := response.Header().Get("Location"); location != "" {
			t.Fatalf("unexpected redirect Location=%q", location)
		}
	})

	t.Run("health ready on noncanonical host is not redirected", func(t *testing.T) {
		called := false
		handler := middleware.CanonicalHost(canonical, http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			called = true
			response.WriteHeader(http.StatusOK)
		}))
		request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		request.Host = "treelot-web.onrender.com"
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if !called || response.Code != http.StatusOK {
			t.Fatalf("called=%v status=%d", called, response.Code)
		}
	})
}

func TestCanonicalHostRejectsOpenRedirectTricks(t *testing.T) {
	t.Parallel()

	canonical, err := url.Parse("https://treelot.troop900livermore.org")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	handler := middleware.CanonicalHost(canonical, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler must not run")
	}))

	request := httptest.NewRequest(http.MethodGet, "https://evil.example/steal", nil)
	request.Host = "treelot-web.onrender.com"
	request.URL.Scheme = "https"
	request.URL.Host = "evil.example"
	request.URL.Path = "/steal"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusMovedPermanently {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMovedPermanently)
	}
	location := response.Header().Get("Location")
	want := "https://treelot.troop900livermore.org/steal"
	if location != want {
		t.Fatalf("Location = %q, want fixed canonical target %q", location, want)
	}
}

func TestCanonicalHostNilOrEmptySkipsRedirect(t *testing.T) {
	t.Parallel()

	called := false
	handler := middleware.CanonicalHost(nil, http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		called = true
		response.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Host = "anything.example"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if !called || response.Code != http.StatusNoContent {
		t.Fatalf("called=%v status=%d", called, response.Code)
	}
}
