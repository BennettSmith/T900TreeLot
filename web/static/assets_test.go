package static_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/troop900/treelot/web/static"
)

func TestHandlerServesEmbeddedGeneratedStylesheet(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/static/app.css", nil)
	response := httptest.NewRecorder()

	static.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/css") {
		t.Errorf("Content-Type = %q, want text/css", got)
	}
	for _, token := range []string{"--color-canvas", ".shift-card", "@media"} {
		if !strings.Contains(response.Body.String(), token) {
			t.Errorf("generated stylesheet does not contain %q", token)
		}
	}
}

func TestHandlerServesEmbeddedPasskeyScript(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/static/passkeys.js", nil)
	response := httptest.NewRecorder()

	static.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/javascript") && !strings.HasPrefix(got, "application/javascript") {
		t.Errorf("Content-Type = %q, want JavaScript", got)
	}
	if body := response.Body.String(); !strings.Contains(body, "navigator.credentials.create") || !strings.Contains(body, "bootstrap_token") {
		t.Errorf("passkey script missing bootstrap flow behavior: %s", body)
	}
	if body := response.Body.String(); !strings.Contains(body, "redirectTo") || strings.Contains(body, `redirect: "manual"`) {
		t.Errorf("passkey script should follow JSON redirectTo without opaque manual redirects: %s", body)
	}
	for _, token := range []string{"navigator.credentials.get", "authenticatorData", "signature", "userHandle", "data-sign-in-form"} {
		if body := response.Body.String(); !strings.Contains(body, token) {
			t.Errorf("passkey script missing assertion behavior %q", token)
		}
	}
}

func TestScoutBucksMetricsAdaptToTheirContainer(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/static/app.css", nil)
	response := httptest.NewRecorder()
	static.Handler().ServeHTTP(response, request)
	css := response.Body.String()

	gridContract := `.bucks-summary .metric-grid{grid-template-columns:repeat(auto-fit,minmax(min(100%,10rem),1fr))}`
	if !strings.Contains(css, gridContract) {
		t.Errorf("generated stylesheet is missing responsive Scout Bucks contract %q", gridContract)
	}
	valueRule := regexp.MustCompile(`\.bucks-summary \.metric>strong\{([^}]+)\}`).FindString(css)
	for _, property := range []string{"font-size:clamp(1.1rem,2vw,1.5rem)", "white-space:nowrap"} {
		if !strings.Contains(valueRule, property) {
			t.Errorf("Scout Bucks value rule %q is missing %q", valueRule, property)
		}
	}
}
