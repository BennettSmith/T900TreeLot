//go:build acceptance

package dsl

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	clockdriver "github.com/troop900/treelot/acceptance/drivers/clock"
	"github.com/troop900/treelot/acceptance/drivers/web"
	"github.com/troop900/treelot/acceptance/environment"
)

// Lot is the acceptance DSL for foundation smoke journeys.
type Lot struct {
	t      *testing.T
	config environment.Config
	web    *web.Client
	clock  *clockdriver.Driver
}

// NewLot constructs a scenario helper.
func NewLot(t *testing.T) *Lot {
	t.Helper()
	config := environment.Load()
	client, err := web.NewClient(config.BaseURL)
	if err != nil {
		t.Fatalf("web client: %v", err)
	}
	lot := &Lot{
		t:      t,
		config: config,
		web:    client,
		clock:  clockdriver.New(config.BaseURL, config.TestControlKey),
	}
	lot.WaitUntilReady()
	return lot
}

// WaitUntilReady polls liveness and readiness until the deployment is usable.
func (l *Lot) WaitUntilReady() {
	l.t.Helper()
	err := environment.Eventually(l.config.ReadyTimeout, 200*time.Millisecond, func() error {
		status, body, err := l.web.Get("/health/live")
		if err != nil {
			return err
		}
		if status != http.StatusOK || !strings.Contains(body, "ok") {
			return fmt.Errorf("live status=%d body=%q", status, body)
		}
		status, body, err = l.web.Get("/health/ready")
		if err != nil {
			return err
		}
		if status != http.StatusOK || !strings.Contains(body, "ready") {
			return fmt.Errorf("ready status=%d body=%q", status, body)
		}
		return nil
	})
	if err != nil {
		l.t.Fatalf("deployment not ready: %v", err)
	}
}

// Presence observes anonymous visitor-facing pages.
func (l *Lot) Presence() *Presence {
	return &Presence{lot: l}
}

// Clock exposes test-time control.
func (l *Lot) Clock() *Clock {
	return &Clock{lot: l}
}

// Presence captures anonymous visitor checks.
type Presence struct {
	lot *Lot
}

// SeesHealthyFoundation asserts live/ready/static/home navigation.
func (p *Presence) SeesHealthyFoundation() {
	p.lot.t.Helper()
	status, body, err := p.lot.web.Get("/health/live")
	if err != nil || status != http.StatusOK || !strings.Contains(body, "ok") {
		p.lot.t.Fatalf("live check failed: status=%d err=%v body=%q", status, err, body)
	}
	status, body, err = p.lot.web.Get("/health/ready")
	if err != nil || status != http.StatusOK || !strings.Contains(body, "ready") {
		p.lot.t.Fatalf("ready check failed: status=%d err=%v body=%q", status, err, body)
	}
	status, body, err = p.lot.web.Get("/static/app.css")
	if err != nil || status != http.StatusOK || !strings.Contains(body, "--color-canvas") {
		p.lot.t.Fatalf("static asset missing: status=%d err=%v", status, err)
	}
	status, body, err = p.lot.web.Get("/")
	if err != nil || status != http.StatusOK {
		p.lot.t.Fatalf("home status=%d err=%v", status, err)
	}
	if !strings.Contains(body, "Troop 900 Tree Lot") || !strings.Contains(body, `aria-label="Primary"`) {
		p.lot.t.Fatalf("home missing brand or navigation")
	}
}

// SubmitsSmokeCheck posts the CSRF-protected foundation form.
func (p *Presence) SubmitsSmokeCheck(message string) {
	p.lot.t.Helper()
	status, body, err := p.lot.web.Get("/")
	if err != nil || status != http.StatusOK {
		p.lot.t.Fatalf("home for csrf: status=%d err=%v", status, err)
	}
	token, err := web.CSRFToken(body)
	if err != nil {
		p.lot.t.Fatalf("csrf: %v", err)
	}
	values := url.Values{}
	values.Set("message", message)
	values.Set("csrf_token", token)
	status, _, headers, err := p.lot.web.PostForm("/smoke", values)
	if err != nil {
		p.lot.t.Fatalf("smoke post: %v", err)
	}
	if status != http.StatusSeeOther {
		p.lot.t.Fatalf("smoke status=%d, want 303", status)
	}
	if location := headers.Get("Location"); !strings.Contains(location, "message=") {
		p.lot.t.Fatalf("smoke location=%q", location)
	}

	status, _, _, err = p.lot.web.PostForm("/smoke", url.Values{"message": {message}})
	if err != nil {
		p.lot.t.Fatalf("smoke without csrf: %v", err)
	}
	if status != http.StatusForbidden {
		p.lot.t.Fatalf("smoke without csrf status=%d, want 403", status)
	}
}

// Clock wraps test-control time operations.
type Clock struct {
	lot *Lot
}

// CanBeAdvanced asserts the acceptance clock endpoint works.
func (c *Clock) CanBeAdvanced(duration time.Duration) {
	c.lot.t.Helper()
	before, err := c.lot.clock.Now()
	if err != nil {
		c.lot.t.Fatalf("clock now: %v", err)
	}
	after, err := c.lot.clock.Advance(duration)
	if err != nil {
		c.lot.t.Fatalf("clock advance: %v", err)
	}
	if !after.After(before) {
		c.lot.t.Fatalf("clock did not advance: before=%v after=%v", before, after)
	}
}

// IsUnavailableWithoutKey asserts missing credentials are rejected.
func (c *Clock) IsUnavailableWithoutKey() {
	c.lot.t.Helper()
	unauthorized := clockdriver.New(c.lot.config.BaseURL, "wrong-key")
	if _, err := unauthorized.Now(); err == nil {
		c.lot.t.Fatal("clock allowed unauthorized access")
	}
}
