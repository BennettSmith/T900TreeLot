//go:build acceptance

package dsl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	clockdriver "github.com/troop900/treelot/acceptance/drivers/clock"
	outboxdriver "github.com/troop900/treelot/acceptance/drivers/outbox"
	"github.com/troop900/treelot/acceptance/drivers/providers"
	"github.com/troop900/treelot/acceptance/drivers/web"
	webauthndriver "github.com/troop900/treelot/acceptance/drivers/webauthn"
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

// Bootstrap exposes first-Admin enrollment checks.
func (l *Lot) Bootstrap() *Bootstrap { return &Bootstrap{lot: l} }

// SignIn exposes authenticated passkey sign-in checks.
func (l *Lot) SignIn() *SignIn { return &SignIn{lot: l} }

// AccountSecurity exposes self-service credential and email management checks.
func (l *Lot) AccountSecurity() *AccountSecurity { return &AccountSecurity{lot: l} }

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

// Bootstrap captures the browser-visible first Admin enrollment flow.
type Bootstrap struct{ lot *Lot }

// CreatesFirstAdminAndClosesBootstrap proves the full HTTP/WebAuthn bootstrap journey.
func (b *Bootstrap) CreatesFirstAdminAndClosesBootstrap(token, email string) {
	b.lot.t.Helper()
	b.reset()
	b.completeEnrollment(b.lot.web, token, email, "First Admin")
	b.RejectsBootstrapBecauseItIsClosed(token)
}

// RejectsBootstrapBecauseItIsClosed proves permanent post-success closure.
func (b *Bootstrap) RejectsBootstrapBecauseItIsClosed(token string) {
	b.lot.t.Helper()
	status, body, err := b.lot.web.Get("/bootstrap")
	if err != nil || status != http.StatusOK {
		b.lot.t.Fatalf("closed bootstrap page status=%d err=%v", status, err)
	}
	csrf, err := web.CSRFToken(body)
	if err != nil {
		b.lot.t.Fatalf("closed csrf: %v", err)
	}
	closed := url.Values{"csrf_token": {csrf}, "bootstrap_token": {token}}
	status, body, _, err = b.lot.web.PostForm("/bootstrap/start", closed)
	if err != nil || status != http.StatusOK || !strings.Contains(body, "Bootstrap enrollment is unavailable") {
		b.lot.t.Fatalf("closed bootstrap status=%d err=%v body=%q", status, err, body)
	}
	if strings.Contains(strings.ToLower(body), "@example.org") {
		b.lot.t.Fatalf("closed bootstrap response leaked account email details")
	}
}

// RejectsInvalidTokenWithoutRevealingAccounts proves AC3 privacy for bad tokens.
func (b *Bootstrap) RejectsInvalidTokenWithoutRevealingAccounts(invalidToken, existingEmail string) {
	b.lot.t.Helper()
	b.reset()
	status, body, err := b.lot.web.Get("/bootstrap")
	if err != nil || status != http.StatusOK {
		b.lot.t.Fatalf("bootstrap entry status=%d err=%v", status, err)
	}
	csrf, err := web.CSRFToken(body)
	if err != nil {
		b.lot.t.Fatalf("bootstrap csrf: %v", err)
	}
	values := url.Values{"csrf_token": {csrf}, "bootstrap_token": {invalidToken}}
	status, body, _, err = b.lot.web.PostForm("/bootstrap/start", values)
	if err != nil || status != http.StatusOK {
		b.lot.t.Fatalf("invalid token status=%d err=%v", status, err)
	}
	if !strings.Contains(body, "Bootstrap enrollment is unavailable") {
		b.lot.t.Fatalf("invalid token body missing generic unavailable message: %q", body)
	}
	if strings.Contains(body, existingEmail) || strings.Contains(body, "already") {
		b.lot.t.Fatalf("invalid token response revealed unrelated account details: %q", body)
	}
}

// RateLimitsRepeatedInvalidBootstrapAttempts proves AC3 rate limiting.
func (b *Bootstrap) RateLimitsRepeatedInvalidBootstrapAttempts(invalidToken string, maxAttempts int) {
	b.lot.t.Helper()
	b.reset()
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		status, body, err := b.lot.web.Get("/bootstrap")
		if err != nil || status != http.StatusOK {
			b.lot.t.Fatalf("attempt %d entry status=%d err=%v", attempt, status, err)
		}
		csrf, err := web.CSRFToken(body)
		if err != nil {
			b.lot.t.Fatalf("attempt %d csrf: %v", attempt, err)
		}
		values := url.Values{"csrf_token": {csrf}, "bootstrap_token": {invalidToken}}
		status, body, _, err = b.lot.web.PostForm("/bootstrap/start", values)
		if err != nil || status != http.StatusOK {
			b.lot.t.Fatalf("attempt %d status=%d err=%v", attempt, status, err)
		}
		if !strings.Contains(body, "Bootstrap enrollment is unavailable") {
			b.lot.t.Fatalf("attempt %d missing unavailable message: %q", attempt, body)
		}
	}
	status, body, err := b.lot.web.Get("/bootstrap")
	if err != nil || status != http.StatusOK {
		b.lot.t.Fatalf("rate-limit probe entry status=%d err=%v", status, err)
	}
	csrf, err := web.CSRFToken(body)
	if err != nil {
		b.lot.t.Fatalf("rate-limit csrf: %v", err)
	}
	values := url.Values{"csrf_token": {csrf}, "bootstrap_token": {invalidToken}}
	status, body, _, err = b.lot.web.PostForm("/bootstrap/start", values)
	if err != nil || status != http.StatusOK {
		b.lot.t.Fatalf("rate-limited status=%d err=%v", status, err)
	}
	if !strings.Contains(body, "Too many attempts") {
		b.lot.t.Fatalf("rate-limited body=%q", body)
	}
}

// FailedPasskeyCeremonyLeavesBootstrapOpen proves cancelled/invalid WebAuthn creates no Admin.
func (b *Bootstrap) FailedPasskeyCeremonyLeavesBootstrapOpen(token, email string) {
	b.lot.t.Helper()
	b.reset()
	csrf := b.startAndClaim(b.lot.web, token, email, "First Admin")
	beginBody, err := json.Marshal(map[string]string{
		"token":                  token,
		"email":                  email,
		"first_name":             "First",
		"last_name":              "Admin",
		"preferred_display_name": "First Admin",
	})
	if err != nil {
		b.lot.t.Fatal(err)
	}
	status, body, _, err := b.lot.web.PostJSON("/bootstrap/passkey/begin", string(beginBody), map[string]string{"X-CSRF-Token": csrf})
	if err != nil || status != http.StatusOK {
		b.lot.t.Fatalf("passkey begin status=%d err=%v body=%q", status, err, body)
	}
	begin, err := webauthndriver.ParseBeginPayload(body)
	if err != nil {
		b.lot.t.Fatal(err)
	}
	finishBody := fmt.Sprintf(`{"token":%q,"email":%q,"first_name":"First","last_name":"Admin","preferred_display_name":"First Admin","ceremonyId":%q,"credential":{"id":"bad","rawId":"bad","type":"public-key","response":{"clientDataJSON":"e30","attestationObject":"e30"}}}`, token, email, begin.CeremonyID)
	status, body, _, err = b.lot.web.PostJSON("/bootstrap/passkey/finish", finishBody, map[string]string{"X-CSRF-Token": csrf})
	if err != nil || status == http.StatusOK {
		b.lot.t.Fatalf("failed ceremony unexpectedly succeeded: status=%d err=%v body=%q", status, err, body)
	}
	if !strings.Contains(body, "Passkey registration could not be completed") {
		b.lot.t.Fatalf("failed ceremony body=%q", body)
	}
	status, accountBody, err := b.lot.web.Get("/account")
	if err != nil {
		b.lot.t.Fatalf("account after failure: %v", err)
	}
	if status != http.StatusSeeOther {
		b.lot.t.Fatalf("account after failed ceremony status=%d body=%q", status, accountBody)
	}
	b.completeEnrollment(b.lot.web, token, email, "First Admin")
}

// RejectsChangedProfileAfterPasskeyRegistrationBegins proves the attested claim cannot be replaced at finish.
func (b *Bootstrap) RejectsChangedProfileAfterPasskeyRegistrationBegins(token, email string) {
	b.lot.t.Helper()
	b.reset()
	csrf := b.startAndClaim(b.lot.web, token, email, "First Admin")
	beginBody, err := json.Marshal(map[string]string{
		"token":                  token,
		"email":                  email,
		"first_name":             "First",
		"last_name":              "Admin",
		"preferred_display_name": "First Admin",
	})
	if err != nil {
		b.lot.t.Fatal(err)
	}
	status, body, _, err := b.lot.web.PostJSON("/bootstrap/passkey/begin", string(beginBody), map[string]string{"X-CSRF-Token": csrf})
	if err != nil || status != http.StatusOK {
		b.lot.t.Fatalf("passkey begin status=%d err=%v body=%q", status, err, body)
	}
	begin, err := webauthndriver.ParseBeginPayload(body)
	if err != nil {
		b.lot.t.Fatal(err)
	}
	attestation, err := webauthndriver.CreateAttestationResponse(webauthndriver.RelyingParty{
		Name:   "Troop 900 Tree Lot",
		ID:     "treelot.test",
		Origin: "https://treelot.test",
	}, begin.PublicKey)
	if err != nil {
		b.lot.t.Fatalf("create attestation: %v", err)
	}

	changedFinish := fmt.Sprintf(
		`{"token":%q,"email":"changed@example.org","first_name":"Changed","last_name":"Person","preferred_display_name":"Changed Person","ceremonyId":%q,"credential":%s}`,
		token,
		begin.CeremonyID,
		attestation,
	)
	status, body, _, err = b.lot.web.PostJSON("/bootstrap/passkey/finish", changedFinish, map[string]string{"X-CSRF-Token": csrf})
	if err != nil || status != http.StatusBadRequest || !strings.Contains(body, "Passkey registration could not be completed") {
		b.lot.t.Fatalf("changed profile finish status=%d err=%v body=%q", status, err, body)
	}
	if strings.Contains(body, email) || strings.Contains(body, "changed@example.org") {
		b.lot.t.Fatalf("changed profile failure exposed identity fields: %q", body)
	}
	status, accountBody, err := b.lot.web.Get("/account")
	if err != nil || status != http.StatusSeeOther {
		b.lot.t.Fatalf("account after changed profile status=%d err=%v body=%q", status, err, accountBody)
	}

	originalFinish := fmt.Sprintf(
		`{"token":%q,"email":%q,"first_name":"First","last_name":"Admin","preferred_display_name":"First Admin","ceremonyId":%q,"credential":%s}`,
		token,
		email,
		begin.CeremonyID,
		attestation,
	)
	status, body, _, err = b.lot.web.PostJSON("/bootstrap/passkey/finish", originalFinish, map[string]string{"X-CSRF-Token": csrf})
	if err != nil || status != http.StatusOK {
		b.lot.t.Fatalf("original profile retry status=%d err=%v body=%q", status, err, body)
	}
	status, accountBody, err = b.lot.web.Get("/account")
	if err != nil || status != http.StatusOK || !strings.Contains(accountBody, "Welcome, First Admin") {
		b.lot.t.Fatalf("account after original retry status=%d err=%v body=%q", status, err, accountBody)
	}
}

// OnlyOneConcurrentBootstrapSucceeds proves transactional one-Admin closure under concurrency.
func (b *Bootstrap) OnlyOneConcurrentBootstrapSucceeds(token string) {
	b.lot.t.Helper()
	b.reset()

	type clientState struct {
		client *web.Client
		csrf   string
		email  string
		begin  webauthndriver.BeginPayload
	}
	states := make([]clientState, 2)
	for i := range states {
		client, err := web.NewClient(b.lot.config.BaseURL)
		if err != nil {
			b.lot.t.Fatalf("client %d: %v", i, err)
		}
		email := fmt.Sprintf("concurrent-admin-%d-%d@example.org", i, time.Now().UTC().UnixNano())
		csrf := b.startAndClaim(client, token, email, fmt.Sprintf("Concurrent %d", i))
		beginBody, err := json.Marshal(map[string]string{
			"token":                  token,
			"email":                  email,
			"first_name":             "Concurrent",
			"last_name":              fmt.Sprintf("Admin%d", i),
			"preferred_display_name": fmt.Sprintf("Concurrent %d", i),
		})
		if err != nil {
			b.lot.t.Fatal(err)
		}
		status, body, _, err := client.PostJSON("/bootstrap/passkey/begin", string(beginBody), map[string]string{"X-CSRF-Token": csrf})
		if err != nil || status != http.StatusOK {
			b.lot.t.Fatalf("concurrent begin %d status=%d err=%v body=%q", i, status, err, body)
		}
		begin, err := webauthndriver.ParseBeginPayload(body)
		if err != nil {
			b.lot.t.Fatal(err)
		}
		states[i] = clientState{client: client, csrf: csrf, email: email, begin: begin}
	}

	var (
		wg      sync.WaitGroup
		results = make([]int, len(states))
	)
	wg.Add(len(states))
	for i, state := range states {
		go func(i int, state clientState) {
			defer wg.Done()
			attestation, err := webauthndriver.CreateAttestationResponse(webauthndriver.RelyingParty{
				Name:   "Troop 900 Tree Lot",
				ID:     "treelot.test",
				Origin: "https://treelot.test",
			}, state.begin.PublicKey)
			if err != nil {
				results[i] = -1
				return
			}
			finishBody := fmt.Sprintf(
				`{"token":%q,"email":%q,"first_name":"Concurrent","last_name":%q,"preferred_display_name":%q,"ceremonyId":%q,"credential":%s}`,
				token,
				state.email,
				fmt.Sprintf("Admin%d", i),
				fmt.Sprintf("Concurrent %d", i),
				state.begin.CeremonyID,
				attestation,
			)
			status, _, _, err := state.client.PostJSON("/bootstrap/passkey/finish", finishBody, map[string]string{"X-CSRF-Token": state.csrf})
			if err != nil {
				results[i] = -1
				return
			}
			results[i] = status
		}(i, state)
	}
	wg.Wait()

	successes := 0
	for _, status := range results {
		if status == http.StatusOK {
			successes++
		}
	}
	if successes != 1 {
		b.lot.t.Fatalf("concurrent bootstrap successes=%d results=%v, want exactly 1", successes, results)
	}
	b.RejectsBootstrapBecauseItIsClosed(token)
}

func (b *Bootstrap) completeEnrollment(client *web.Client, token, email, displayName string) {
	passkey := webauthndriver.NewPasskey(webauthndriver.RelyingParty{
		Name:   "Troop 900 Tree Lot",
		ID:     "treelot.test",
		Origin: "https://treelot.test",
	})
	b.completeEnrollmentWithPasskey(client, token, email, displayName, passkey)
}

func (b *Bootstrap) completeEnrollmentWithPasskey(client *web.Client, token, email, displayName string, passkey *webauthndriver.Passkey) {
	b.lot.t.Helper()
	csrf := b.startAndClaim(client, token, email, displayName)
	beginBody, err := json.Marshal(map[string]string{
		"token":                  token,
		"email":                  email,
		"first_name":             "First",
		"last_name":              "Admin",
		"preferred_display_name": displayName,
	})
	if err != nil {
		b.lot.t.Fatal(err)
	}
	status, body, _, err := client.PostJSON("/bootstrap/passkey/begin", string(beginBody), map[string]string{"X-CSRF-Token": csrf})
	if err != nil || status != http.StatusOK {
		b.lot.t.Fatalf("passkey begin status=%d err=%v body=%q", status, err, body)
	}
	begin, err := webauthndriver.ParseBeginPayload(body)
	if err != nil {
		b.lot.t.Fatal(err)
	}
	attestationResponse, err := passkey.CreateAttestationResponse(begin.PublicKey)
	if err != nil {
		b.lot.t.Fatalf("create attestation: %v", err)
	}
	finishBody := fmt.Sprintf(`{"token":%q,"email":%q,"first_name":"First","last_name":"Admin","preferred_display_name":%q,"ceremonyId":%q,"credential":%s}`, token, email, displayName, begin.CeremonyID, attestationResponse)
	status, body, _, err = client.PostJSON("/bootstrap/passkey/finish", finishBody, map[string]string{"X-CSRF-Token": csrf})
	if err != nil || status != http.StatusOK {
		b.lot.t.Fatalf("passkey finish status=%d err=%v body=%q", status, err, body)
	}
	var finish struct {
		RedirectTo string `json:"redirectTo"`
	}
	if err := json.Unmarshal([]byte(body), &finish); err != nil {
		b.lot.t.Fatalf("decode finish body: %v body=%q", err, body)
	}
	if finish.RedirectTo != "/account" {
		b.lot.t.Fatalf("finish redirectTo=%q body=%q", finish.RedirectTo, body)
	}
	status, body, err = client.Get(finish.RedirectTo)
	if err != nil || status != http.StatusOK || !strings.Contains(body, "Welcome, "+displayName) {
		b.lot.t.Fatalf("account status=%d err=%v body=%q", status, err, body)
	}
}

func (b *Bootstrap) startAndClaim(client *web.Client, token, email, displayName string) string {
	b.lot.t.Helper()
	status, body, err := client.Get("/bootstrap")
	if err != nil || status != http.StatusOK {
		b.lot.t.Fatalf("bootstrap entry status=%d err=%v", status, err)
	}
	csrf, err := web.CSRFToken(body)
	if err != nil {
		b.lot.t.Fatalf("bootstrap csrf: %v", err)
	}
	if !strings.Contains(body, `data-bootstrap-entry`) || strings.Contains(body, "token="+token) {
		b.lot.t.Fatalf("bootstrap entry missing form or leaked token")
	}
	values := url.Values{
		"csrf_token":             {csrf},
		"bootstrap_token":        {token},
		"email":                  {email},
		"first_name":             {"First"},
		"last_name":              {"Admin"},
		"preferred_display_name": {displayName},
	}
	status, body, _, err = client.PostForm("/bootstrap/start", values)
	if err != nil || status != http.StatusOK || !strings.Contains(body, `action="/bootstrap/claim"`) {
		b.lot.t.Fatalf("bootstrap start status=%d err=%v body=%q", status, err, body)
	}
	status, body, _, err = client.PostForm("/bootstrap/claim", values)
	if err != nil || status != http.StatusOK || !strings.Contains(body, `data-bootstrap-passkey`) {
		b.lot.t.Fatalf("bootstrap claim status=%d err=%v body=%q", status, err, body)
	}
	return csrf
}

func (b *Bootstrap) reset() {
	b.lot.t.Helper()
	status, body, _, err := b.lot.web.PostJSON("/_test/bootstrap/reset", `{}`, map[string]string{"X-Test-Control-Key": b.lot.config.TestControlKey})
	if err != nil || status != http.StatusOK || !strings.Contains(body, "reset") {
		b.lot.t.Fatalf("bootstrap reset status=%d err=%v body=%q", status, err, body)
	}
}

// SignIn captures the browser-visible passkey assertion flow.
type SignIn struct{ lot *Lot }

// FamilyManagerUsesDiscoverablePasskey proves usernameless role-aware sign-in.
func (s *SignIn) FamilyManagerUsesDiscoverablePasskey(token, email string) {
	s.signInWithRole(token, email, "Family Manager", "family_manager", "/family", "Family dashboard", "")
}

// YoungAdultScoutUsesPersonalSchedule proves self-only scout landing.
func (s *SignIn) YoungAdultScoutUsesPersonalSchedule(token, email string) {
	s.signInWithRole(token, email, "Young Adult Scout", "young_adult_scout", "/scout/schedule", "Personal schedule", "household")
}

func (s *SignIn) signInWithRole(token, email, displayName, role, redirect, heading, forbidden string) *web.Client {
	s.lot.t.Helper()
	bootstrap := s.lot.Bootstrap()
	bootstrap.reset()
	passkey := webauthndriver.NewPasskey(webauthndriver.RelyingParty{
		Name:   "Troop 900 Tree Lot",
		ID:     "treelot.test",
		Origin: "https://treelot.test",
	})
	bootstrap.completeEnrollmentWithPasskey(s.lot.web, token, email, displayName, passkey)
	s.setRole(email, role)

	client, err := web.NewClient(s.lot.config.BaseURL)
	if err != nil {
		s.lot.t.Fatalf("anonymous sign-in client: %v", err)
	}
	status, body, err := client.Get("/sign-in")
	if err != nil || status != http.StatusOK {
		s.lot.t.Fatalf("sign-in entry status=%d err=%v body=%q", status, err, body)
	}
	csrf, err := web.CSRFToken(body)
	if err != nil {
		s.lot.t.Fatalf("sign-in csrf: %v", err)
	}
	status, body, _, err = client.PostJSON("/sign-in/passkey/begin", `{}`, map[string]string{"X-CSRF-Token": csrf})
	if err != nil || status != http.StatusOK {
		s.lot.t.Fatalf("sign-in begin status=%d err=%v body=%q", status, err, body)
	}
	begin, err := webauthndriver.ParseBeginPayload(body)
	if err != nil {
		s.lot.t.Fatal(err)
	}
	assertion, err := passkey.CreateAssertionResponse(begin.PublicKey)
	if err != nil {
		s.lot.t.Fatalf("create assertion: %v", err)
	}
	finishBody := fmt.Sprintf(`{"ceremonyId":%q,"credential":%s}`, begin.CeremonyID, assertion)
	status, body, _, err = client.PostJSON("/sign-in/passkey/finish", finishBody, map[string]string{"X-CSRF-Token": csrf})
	if err != nil || status != http.StatusOK {
		s.lot.t.Fatalf("sign-in finish status=%d err=%v body=%q", status, err, body)
	}
	var finish struct {
		RedirectTo string `json:"redirectTo"`
	}
	if err := json.Unmarshal([]byte(body), &finish); err != nil {
		s.lot.t.Fatalf("decode sign-in finish: %v body=%q", err, body)
	}
	if finish.RedirectTo != redirect {
		s.lot.t.Fatalf("redirectTo=%q, want %q", finish.RedirectTo, redirect)
	}
	status, body, err = client.Get(finish.RedirectTo)
	if err != nil || status != http.StatusOK || !strings.Contains(body, heading) {
		s.lot.t.Fatalf("role landing status=%d err=%v body=%q", status, err, body)
	}
	if forbidden != "" && strings.Contains(strings.ToLower(body), forbidden) {
		s.lot.t.Fatalf("role landing exposed forbidden %q context: %q", forbidden, body)
	}
	return client
}

// EmailHintDoesNotRevealAccount compares observable known and unknown hint outcomes.
func (s *SignIn) EmailHintDoesNotRevealAccount(token, knownEmail, unknownEmail string) {
	s.lot.t.Helper()
	bootstrap := s.lot.Bootstrap()
	bootstrap.reset()
	bootstrap.completeEnrollment(s.lot.web, token, knownEmail, "Hinted Manager")
	s.setRole(knownEmail, "family_manager")

	for label, email := range map[string]string{"known": knownEmail, "unknown": unknownEmail} {
		client, err := web.NewClient(s.lot.config.BaseURL)
		if err != nil {
			s.lot.t.Fatal(err)
		}
		status, body, err := client.Get("/sign-in")
		if err != nil || status != http.StatusOK {
			s.lot.t.Fatalf("%s hint entry status=%d err=%v", label, status, err)
		}
		csrf, err := web.CSRFToken(body)
		if err != nil {
			s.lot.t.Fatal(err)
		}
		payload, err := json.Marshal(map[string]string{"email": email})
		if err != nil {
			s.lot.t.Fatal(err)
		}
		status, body, _, err = client.PostJSON("/sign-in/passkey/begin", string(payload), map[string]string{"X-CSRF-Token": csrf})
		if err != nil || status != http.StatusOK {
			s.lot.t.Fatalf("%s hint status=%d err=%v body=%q", label, status, err, body)
		}
		if _, err := webauthndriver.ParseBeginPayload(body); err != nil {
			s.lot.t.Fatalf("%s hint payload: %v", label, err)
		}
		lowerBody := strings.ToLower(body)
		if strings.Contains(lowerBody, strings.ToLower(knownEmail)) ||
			strings.Contains(lowerBody, strings.ToLower(unknownEmail)) ||
			strings.Contains(lowerBody, "registered") ||
			strings.Contains(lowerBody, "unknown account") {
			s.lot.t.Fatalf("%s hint disclosed account existence: %q", label, body)
		}
	}
}

// RateLimitsEmailHints proves repeated account-hint requests are bounded.
func (s *SignIn) RateLimitsEmailHints(email string, maxAttempts int) {
	s.lot.t.Helper()
	s.lot.Bootstrap().reset()
	client, err := web.NewClient(s.lot.config.BaseURL)
	if err != nil {
		s.lot.t.Fatal(err)
	}
	status, body, err := client.Get("/sign-in")
	if err != nil || status != http.StatusOK {
		s.lot.t.Fatalf("rate limit entry status=%d err=%v", status, err)
	}
	csrf, err := web.CSRFToken(body)
	if err != nil {
		s.lot.t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]string{"email": email})
	if err != nil {
		s.lot.t.Fatal(err)
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		status, body, _, err = client.PostJSON("/sign-in/passkey/begin", string(payload), map[string]string{"X-CSRF-Token": csrf})
		if err != nil || status != http.StatusOK {
			s.lot.t.Fatalf("hint attempt %d status=%d err=%v body=%q", attempt, status, err, body)
		}
	}
	status, body, _, err = client.PostJSON("/sign-in/passkey/begin", string(payload), map[string]string{"X-CSRF-Token": csrf})
	if err != nil || status != http.StatusTooManyRequests || !strings.Contains(body, "Too many attempts") {
		s.lot.t.Fatalf("rate-limited hint status=%d err=%v body=%q", status, err, body)
	}
}

// RevokedSessionIsRejected proves every protected request observes revocation.
func (s *SignIn) RevokedSessionIsRejected(token, email string) {
	s.lot.t.Helper()
	client := s.signInWithRole(token, email, "Revoked Manager", "family_manager", "/family", "Family dashboard", "")
	payload, err := json.Marshal(map[string]string{"email": email})
	if err != nil {
		s.lot.t.Fatal(err)
	}
	status, body, _, err := s.lot.web.PostJSON("/_test/identity/revoke", string(payload), map[string]string{"X-Test-Control-Key": s.lot.config.TestControlKey})
	if err != nil || status != http.StatusOK {
		s.lot.t.Fatalf("revoke fixture status=%d err=%v body=%q", status, err, body)
	}
	status, body, err = client.Get("/family")
	if err != nil || status != http.StatusSeeOther {
		s.lot.t.Fatalf("revoked family access status=%d err=%v body=%q", status, err, body)
	}
}

// RejectsFailedAndReplayedAssertion proves failed and consumed ceremonies grant no access.
func (s *SignIn) RejectsFailedAndReplayedAssertion(token, email string) {
	s.lot.t.Helper()
	bootstrap := s.lot.Bootstrap()
	bootstrap.reset()
	passkey := webauthndriver.NewPasskey(webauthndriver.RelyingParty{
		Name: "Troop 900 Tree Lot", ID: "treelot.test", Origin: "https://treelot.test",
	})
	bootstrap.completeEnrollmentWithPasskey(s.lot.web, token, email, "Failed Sign-in Manager", passkey)
	s.setRole(email, "family_manager")
	client, err := web.NewClient(s.lot.config.BaseURL)
	if err != nil {
		s.lot.t.Fatal(err)
	}
	status, body, err := client.Get("/sign-in")
	if err != nil || status != http.StatusOK {
		s.lot.t.Fatalf("failed assertion entry status=%d err=%v", status, err)
	}
	csrf, err := web.CSRFToken(body)
	if err != nil {
		s.lot.t.Fatal(err)
	}
	status, body, _, err = client.PostJSON("/sign-in/passkey/begin", `{}`, map[string]string{"X-CSRF-Token": csrf})
	if err != nil || status != http.StatusOK {
		s.lot.t.Fatalf("failed assertion begin status=%d err=%v body=%q", status, err, body)
	}
	begin, err := webauthndriver.ParseBeginPayload(body)
	if err != nil {
		s.lot.t.Fatal(err)
	}
	badFinish := fmt.Sprintf(`{"ceremonyId":%q,"credential":{"id":"bad","rawId":"bad","type":"public-key","response":{"clientDataJSON":"e30","authenticatorData":"e30","signature":"e30","userHandle":""}}}`, begin.CeremonyID)
	status, body, _, err = client.PostJSON("/sign-in/passkey/finish", badFinish, map[string]string{"X-CSRF-Token": csrf})
	if err != nil || status != http.StatusBadRequest || !strings.Contains(body, "Sign-in could not be completed") {
		s.lot.t.Fatalf("bad assertion status=%d err=%v body=%q", status, err, body)
	}
	status, _, err = client.Get("/family")
	if err != nil || status != http.StatusSeeOther {
		s.lot.t.Fatalf("failed assertion granted family access status=%d err=%v", status, err)
	}

	assertion, err := passkey.CreateAssertionResponse(begin.PublicKey)
	if err != nil {
		s.lot.t.Fatal(err)
	}
	validFinish := fmt.Sprintf(`{"ceremonyId":%q,"credential":%s}`, begin.CeremonyID, assertion)
	status, body, _, err = client.PostJSON("/sign-in/passkey/finish", validFinish, map[string]string{"X-CSRF-Token": csrf})
	if err != nil || status != http.StatusOK {
		s.lot.t.Fatalf("valid retry status=%d err=%v body=%q", status, err, body)
	}
	status, signInBody, err := client.Get("/sign-in")
	if err != nil || status != http.StatusOK {
		s.lot.t.Fatalf("replay csrf page status=%d err=%v", status, err)
	}
	replayCSRF, err := web.CSRFToken(signInBody)
	if err != nil {
		s.lot.t.Fatal(err)
	}
	status, body, _, err = client.PostJSON("/sign-in/passkey/finish", validFinish, map[string]string{"X-CSRF-Token": replayCSRF})
	if err != nil || status != http.StatusBadRequest || !strings.Contains(body, "Sign-in could not be completed") {
		s.lot.t.Fatalf("replayed assertion status=%d err=%v body=%q", status, err, body)
	}
}

// SignedInPersonCanSignOut proves self-revocation through the browser boundary.
func (s *SignIn) SignedInPersonCanSignOut(token, email string) {
	s.lot.t.Helper()
	bootstrap := s.lot.Bootstrap()
	bootstrap.reset()
	bootstrap.completeEnrollment(s.lot.web, token, email, "Signing Out Admin")

	status, body, err := s.lot.web.Get("/account")
	if err != nil || status != http.StatusOK || !strings.Contains(body, `action="/sign-out"`) {
		s.lot.t.Fatalf("account sign-out control status=%d err=%v body=%q", status, err, body)
	}
	csrf, err := web.CSRFToken(body)
	if err != nil {
		s.lot.t.Fatal(err)
	}
	status, _, headers, err := s.lot.web.PostForm("/sign-out", url.Values{"csrf_token": {csrf}})
	if err != nil || status != http.StatusSeeOther || headers.Get("Location") != "/" {
		s.lot.t.Fatalf("sign-out status=%d err=%v location=%q", status, err, headers.Get("Location"))
	}
	if cookie := headers.Get("Set-Cookie"); !strings.Contains(cookie, "treelot_session=") || !strings.Contains(cookie, "Max-Age=0") {
		s.lot.t.Fatalf("sign-out did not clear session cookie: %q", cookie)
	}
	status, body, err = s.lot.web.Get("/account")
	if err != nil || status != http.StatusSeeOther {
		s.lot.t.Fatalf("account after sign-out status=%d err=%v body=%q", status, err, body)
	}
}

func (s *SignIn) setRole(email, role string) {
	s.lot.t.Helper()
	body, err := json.Marshal(map[string]string{"email": email, "role": role})
	if err != nil {
		s.lot.t.Fatal(err)
	}
	status, response, _, err := s.lot.web.PostJSON("/_test/identity/role", string(body), map[string]string{"X-Test-Control-Key": s.lot.config.TestControlKey})
	if err != nil || status != http.StatusOK {
		s.lot.t.Fatalf("set fixture role status=%d err=%v body=%q", status, err, response)
	}
}

// AccountSecurity captures self-service passkey and claimed-email management.
type AccountSecurity struct{ lot *Lot }

// RequiresStepUpBeforeCredentialChanges proves account security gates credential changes.
func (a *AccountSecurity) RequiresStepUpBeforeCredentialChanges(token, email string) {
	a.lot.t.Helper()
	bootstrap := a.lot.Bootstrap()
	bootstrap.reset()
	passkey := webauthndriver.NewPasskey(webauthndriver.RelyingParty{
		Name: "Troop 900 Tree Lot", ID: "treelot.test", Origin: "https://treelot.test",
	})
	bootstrap.completeEnrollmentWithPasskey(a.lot.web, token, email, "Security Admin", passkey)

	status, body, err := a.lot.web.Get("/account/security")
	if err != nil || status != http.StatusOK {
		a.lot.t.Fatalf("account security status=%d err=%v body=%q", status, err, body)
	}
	if !strings.Contains(body, `data-account-step-up`) {
		a.lot.t.Fatalf("account security missing passkey step-up requirement")
	}
	if strings.Contains(body, `data-account-passkeys`) || strings.Contains(body, `data-account-change-email`) {
		a.lot.t.Fatalf("credential controls visible before step-up")
	}

	csrf, err := web.CSRFToken(body)
	if err != nil {
		a.lot.t.Fatal(err)
	}
	newEmail := fmt.Sprintf("changed-%d@example.org", time.Now().UTC().UnixNano())
	status, body, _, err = a.lot.web.PostForm("/account/email", url.Values{
		"csrf_token": {csrf},
		"email":      {newEmail},
	})
	if err != nil || (status != http.StatusForbidden && status != http.StatusBadRequest && status != http.StatusSeeOther) {
		a.lot.t.Fatalf("email change without step-up status=%d err=%v body=%q", status, err, body)
	}
	if status == http.StatusSeeOther {
		status, body, err = a.lot.web.Get("/account/security")
		if err != nil || status != http.StatusOK {
			a.lot.t.Fatalf("reload after denied email change status=%d err=%v", status, err)
		}
	}
	if strings.Contains(body, newEmail) {
		a.lot.t.Fatalf("email changed without step-up: body=%q", body)
	}
	status, accountBody, err := a.lot.web.Get("/account")
	if err != nil || status != http.StatusOK || !strings.Contains(accountBody, email) {
		a.lot.t.Fatalf("original email not preserved status=%d err=%v body=%q", status, err, accountBody)
	}
}
