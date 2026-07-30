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
	outboxdriver "github.com/troop900/treelot/acceptance/drivers/outbox"
	"github.com/troop900/treelot/acceptance/drivers/providers"
	"github.com/troop900/treelot/acceptance/drivers/web"
	"github.com/troop900/treelot/acceptance/environment"
)

// Lot is the acceptance DSL for foundation smoke journeys.
type Lot struct {
	t          *testing.T
	config     environment.Config
	web        *web.Client
	production *web.Client
	clock      *clockdriver.Driver
	outbox     *outboxdriver.Driver
	stub       *providers.Stub
	processes  *environment.ProcessDriver
}

// NewLot constructs a scenario helper.
func NewLot(t *testing.T) *Lot {
	t.Helper()
	config := environment.Load()
	client, err := web.NewClient(config.BaseURL)
	if err != nil {
		t.Fatalf("web client: %v", err)
	}
	production, err := web.NewClient(config.ProductionBaseURL)
	if err != nil {
		t.Fatalf("production web client: %v", err)
	}
	lot := &Lot{
		t:          t,
		config:     config,
		web:        client,
		production: production,
		clock:      clockdriver.New(config.BaseURL, config.TestControlKey),
		outbox:     outboxdriver.New(config.BaseURL, config.TestControlKey),
		stub:       providers.New(config.StubBaseURL),
		processes:  environment.NewProcessDriver(config),
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
func (l *Lot) Presence() *Presence { return &Presence{lot: l} }

// Clock exposes test-time control.
func (l *Lot) Clock() *Clock { return &Clock{lot: l} }

// Platform exposes deployable-foundation operator checks.
func (l *Lot) Platform() *Platform { return &Platform{lot: l} }

// Worker exposes asynchronous outbox observations.
func (l *Lot) Worker() *Worker { return &Worker{lot: l} }

// Providers exposes external stub observations.
func (l *Lot) Providers() *Providers { return &Providers{lot: l} }

// Presence captures anonymous visitor checks.
type Presence struct{ lot *Lot }

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
type Clock struct{ lot *Lot }

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

// Platform captures operator/deploy invariants.
type Platform struct{ lot *Lot }

// RejectsUnmigratedDatabaseWithoutSchemaChange proves migrate-only schema mutation.
func (p *Platform) RejectsUnmigratedDatabaseWithoutSchemaChange() {
	p.lot.t.Helper()
	if err := p.lot.processes.RejectsUnmigratedDatabaseWithoutSchemaChange(); err != nil {
		p.lot.t.Fatal(err)
	}
}

// TestControlIsAbsentOutsideAcceptance proves production-shaped hosts hide test control.
func (p *Platform) TestControlIsAbsentOutsideAcceptance() {
	p.lot.t.Helper()
	status, _, err := p.lot.production.Get("/_test/clock")
	if err != nil {
		p.lot.t.Fatalf("production test-control probe: %v", err)
	}
	if status != http.StatusNotFound {
		p.lot.t.Fatalf("production test-control status=%d, want 404", status)
	}
}

// IssuesSecureSessionCookiesInProduction proves Secure cookie attribute in production.
func (p *Platform) IssuesSecureSessionCookiesInProduction() {
	p.lot.t.Helper()
	status, _, headers, err := p.lot.production.GetWithHeaders("/")
	if err != nil || status != http.StatusOK {
		p.lot.t.Fatalf("production home status=%d err=%v", status, err)
	}
	secure, err := web.SessionCookieSecure(headers)
	if err != nil {
		p.lot.t.Fatal(err)
	}
	if !secure {
		p.lot.t.Fatal("production session cookie missing Secure attribute")
	}
}

// Worker captures asynchronous queue observations.
type Worker struct{ lot *Lot }

// DeliversEnqueuedOutboxMessage proves the deployed worker processes outbox rows.
func (w *Worker) DeliversEnqueuedOutboxMessage(idempotencyKey string) {
	w.lot.t.Helper()
	if err := w.lot.outbox.Enqueue(idempotencyKey, "groupsio"); err != nil {
		w.lot.t.Fatalf("enqueue outbox: %v", err)
	}
	err := environment.Eventually(w.lot.config.ReadyTimeout, 200*time.Millisecond, func() error {
		status, err := w.lot.outbox.Status(idempotencyKey)
		if err != nil {
			return err
		}
		if status != "delivered" {
			return fmt.Errorf("outbox status=%q, want delivered", status)
		}
		return nil
	})
	if err != nil {
		w.lot.t.Fatalf("worker did not deliver outbox message: %v", err)
	}
}

// Providers captures external stub observations.
type Providers struct{ lot *Lot }

// GroupsIOStubAcceptsAuthorizedTraffic proves the stub is protocol-shaped and live.
func (p *Providers) GroupsIOStubAcceptsAuthorizedTraffic() {
	p.lot.t.Helper()
	if err := p.lot.stub.Available(); err != nil {
		p.lot.t.Fatalf("groups.io stub unavailable: %v", err)
	}
	if err := p.lot.stub.PostGroupMessage("troop900", "Foundation stub probe"); err != nil {
		p.lot.t.Fatalf("groups.io stub post: %v", err)
	}
	count, err := p.lot.stub.MessageCount()
	if err != nil || count < 1 {
		p.lot.t.Fatalf("groups.io stub message count=%d err=%v", count, err)
	}
}
