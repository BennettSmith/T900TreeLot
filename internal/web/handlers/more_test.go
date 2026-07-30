package handlers_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/web/handlers"
	"github.com/troop900/treelot/internal/web/middleware"
)

func TestGetClockAndAdvanceValidation(t *testing.T) {
	t.Parallel()

	controllable := clock.NewControllable(time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC))
	server := newServer(t, handlers.Options{
		TestControlEnabled: true,
		TestControlKey:     "secret",
		Clock:              controllable,
		ControllableClock:  controllable,
	})

	unauthorized := request(t, server, http.MethodGet, "/_test/clock", "", nil, nil)
	if unauthorized.Code != http.StatusForbidden {
		t.Fatalf("unauthorized get clock = %d", unauthorized.Code)
	}
	get := request(t, server, http.MethodGet, "/_test/clock", "", map[string]string{"X-Test-Control-Key": "secret"}, nil)
	if get.Code != http.StatusOK || !strings.Contains(get.Body.String(), "2026-07-30T12:00:00") {
		t.Fatalf("get clock = %d %s", get.Code, get.Body.String())
	}

	badBody := request(t, server, http.MethodPost, "/_test/clock/advance", `{`, map[string]string{
		"Content-Type":       "application/json",
		"X-Test-Control-Key": "secret",
	}, nil)
	if badBody.Code != http.StatusBadRequest {
		t.Fatalf("bad body status = %d", badBody.Code)
	}

	badDuration := request(t, server, http.MethodPost, "/_test/clock/advance", `{"duration":"nope"}`, map[string]string{
		"Content-Type":       "application/json",
		"X-Test-Control-Key": "secret",
	}, nil)
	if badDuration.Code != http.StatusBadRequest {
		t.Fatalf("bad duration status = %d", badDuration.Code)
	}
}

func TestSmokeDefaultsMessage(t *testing.T) {
	t.Parallel()

	server := newServer(t, handlers.Options{})
	cookie, csrf := establishSession(t, server)
	form := url.Values{middleware.CSRFFormField: {csrf}}
	response := request(t, server, http.MethodPost, "/smoke", form.Encode(), map[string]string{"Content-Type": "application/x-www-form-urlencoded"}, cookie)
	if response.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", response.Code)
	}
	if location := response.Header().Get("Location"); !strings.Contains(location, "Foundation+smoke+confirmed") && !strings.Contains(location, "Foundation%20smoke%20confirmed") {
		t.Fatalf("Location = %q", location)
	}
}
