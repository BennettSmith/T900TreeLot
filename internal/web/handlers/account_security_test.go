package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/identity/domain"
	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/session"
	"github.com/troop900/treelot/internal/web/handlers"
	"github.com/troop900/treelot/internal/web/middleware"
)

func TestAccountSecurityPageRequiresStepUpBeforeCredentialControls(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	current, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	reader := &fakeAccountSecurityReader{
		profile:  application.AccountProfile{IdentityID: "identity-1", DisplayName: "Ada", PrimaryEmail: "ada@example.org"},
		passkeys: []application.PasskeyCredential{{ID: "cred-1"}},
	}
	server := newServer(t, handlers.Options{
		Sessions:              store,
		AccountSecurityReader: reader,
		StepUpTTL:             5 * time.Minute,
	})
	cookie := &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken}
	response := request(t, server, http.MethodGet, "/account/security", "", nil, cookie)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if !strings.Contains(body, `data-account-step-up`) {
		t.Fatalf("missing step-up: %q", body)
	}
	if strings.Contains(body, `data-account-passkeys`) || strings.Contains(body, `data-account-change-email`) {
		t.Fatalf("credential controls visible before step-up: %q", body)
	}
	_ = current
}

func TestAccountSecurityPageShowsControlsAfterStepUp(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	current, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.MarkStepUp(context.Background(), current.ID, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	reader := &fakeAccountSecurityReader{
		profile:  application.AccountProfile{IdentityID: "identity-1", DisplayName: "Ada", PrimaryEmail: "ada@example.org"},
		passkeys: []application.PasskeyCredential{{ID: "cred-1", CreatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)}},
	}
	server := newServer(t, handlers.Options{
		Sessions:              store,
		AccountSecurityReader: reader,
		StepUpTTL:             5 * time.Minute,
	})
	cookie := &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken}
	response := request(t, server, http.MethodGet, "/account/security", "", nil, cookie)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d", response.Code)
	}
	body := response.Body.String()
	if strings.Contains(body, `data-account-step-up`) || !strings.Contains(body, `data-account-passkeys`) || !strings.Contains(body, `data-account-change-email`) {
		t.Fatalf("unexpected security body: %q", body)
	}
}

func TestAccountChangeEmailRejectsWithoutStepUp(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	current, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	security := &fakeAccountSecurityService{changeErr: domain.ErrStepUpRequired}
	server := newServer(t, handlers.Options{Sessions: store, AccountSecurity: security})
	cookie := &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken}
	form := url.Values{middleware.CSRFFormField: {current.CSRFToken}, "email": {"new@example.org"}}
	response := request(t, server, http.MethodPost, "/account/email", form.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, cookie)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
	}
}

func TestAccountChangeEmailSuccessClearsSession(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	current, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	security := &fakeAccountSecurityService{}
	server := newServer(t, handlers.Options{Sessions: store, AccountSecurity: security})
	cookie := &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken}
	form := url.Values{middleware.CSRFFormField: {current.CSRFToken}, "email": {"new@example.org"}}
	response := request(t, server, http.MethodPost, "/account/email", form.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, cookie)
	if response.Code != http.StatusSeeOther || !strings.Contains(response.Header().Get("Location"), "/sign-in") {
		t.Fatalf("status=%d location=%q", response.Code, response.Header().Get("Location"))
	}
	if security.change.NewEmail != "new@example.org" || security.change.IdentityID != "identity-1" {
		t.Fatalf("command = %#v", security.change)
	}
	if cookieHeader := response.Header().Get("Set-Cookie"); !strings.Contains(cookieHeader, "Max-Age=0") {
		t.Fatalf("expected cleared session cookie, got %q", cookieHeader)
	}
}

func TestAccountStepUpAndPasskeyJSONEndpoints(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	current, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	security := &fakeAccountSecurityService{
		beginStepUp: application.BeginSignInResult{CeremonyID: "step-1", PublicKey: map[string]any{"challenge": "x"}},
		beginAdd:    application.RegistrationOptions{CeremonyID: "reg-1", PublicKey: map[string]any{"challenge": "y"}},
	}
	server := newServer(t, handlers.Options{Sessions: store, AccountSecurity: security})
	cookie := &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken}

	begin := request(t, server, http.MethodPost, "/account/security/step-up/begin", `{}`, map[string]string{
		"Content-Type": "application/json", "X-CSRF-Token": current.CSRFToken,
	}, cookie)
	if begin.Code != http.StatusOK || !strings.Contains(begin.Body.String(), "step-1") {
		t.Fatalf("step-up begin status=%d body=%q", begin.Code, begin.Body.String())
	}
	finish := request(t, server, http.MethodPost, "/account/security/step-up/finish", `{"ceremonyId":"step-1","credential":{"id":"c"}}`, map[string]string{
		"Content-Type": "application/json", "X-CSRF-Token": current.CSRFToken,
	}, cookie)
	if finish.Code != http.StatusOK || !strings.Contains(finish.Body.String(), "/account/security") {
		t.Fatalf("step-up finish status=%d body=%q", finish.Code, finish.Body.String())
	}

	addBegin := request(t, server, http.MethodPost, "/account/security/passkeys/begin", `{}`, map[string]string{
		"Content-Type": "application/json", "X-CSRF-Token": current.CSRFToken,
	}, cookie)
	if addBegin.Code != http.StatusOK || !strings.Contains(addBegin.Body.String(), "reg-1") {
		t.Fatalf("passkey begin status=%d body=%q", addBegin.Code, addBegin.Body.String())
	}
	addFinish := request(t, server, http.MethodPost, "/account/security/passkeys/finish", `{"ceremonyId":"reg-1","credential":{"id":"c"}}`, map[string]string{
		"Content-Type": "application/json", "X-CSRF-Token": current.CSRFToken,
	}, cookie)
	if addFinish.Code != http.StatusOK {
		t.Fatalf("passkey finish status=%d body=%q", addFinish.Code, addFinish.Body.String())
	}

	form := url.Values{middleware.CSRFFormField: {current.CSRFToken}, "credential_id": {"cred-1"}}
	remove := request(t, server, http.MethodPost, "/account/security/passkeys/remove", form.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, cookie)
	if remove.Code != http.StatusSeeOther {
		t.Fatalf("remove status=%d", remove.Code)
	}
	if security.remove.CredentialID != "cred-1" {
		t.Fatalf("remove command = %#v", security.remove)
	}
}

func TestAccountSecurityJSONUnauthorizedWithoutIdentity(t *testing.T) {
	server := newServer(t, handlers.Options{AccountSecurity: &fakeAccountSecurityService{}})
	cookie, csrf := establishSession(t, server)
	response := request(t, server, http.MethodPost, "/account/security/step-up/begin", `{}`, map[string]string{
		"Content-Type": "application/json", "X-CSRF-Token": csrf,
	}, cookie)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
	}
}

func TestAccountSecurityErrorMapping(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	current, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	security := &fakeAccountSecurityService{changeErr: domain.ErrEmailTaken}
	server := newServer(t, handlers.Options{Sessions: store, AccountSecurity: security})
	cookie := &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken}
	form := url.Values{middleware.CSRFFormField: {current.CSRFToken}, "email": {"taken@example.org"}}
	response := request(t, server, http.MethodPost, "/account/email", form.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, cookie)
	if response.Code != http.StatusSeeOther || !strings.Contains(response.Header().Get("Location"), "error=") {
		t.Fatalf("status=%d location=%q", response.Code, response.Header().Get("Location"))
	}
}

type fakeAccountSecurityReader struct {
	profile     application.AccountProfile
	passkeys    []application.PasskeyCredential
	err         error
	passkeysErr error
}

func (f *fakeAccountSecurityReader) FindAccountProfile(context.Context, string) (application.AccountProfile, error) {
	if f.err != nil {
		return application.AccountProfile{}, f.err
	}
	return f.profile, nil
}

func (f *fakeAccountSecurityReader) ListPasskeys(context.Context, string) ([]application.PasskeyCredential, error) {
	if f.passkeysErr != nil {
		return nil, f.passkeysErr
	}
	if f.err != nil {
		return nil, f.err
	}
	return f.passkeys, nil
}

type fakeAccountSecurityService struct {
	beginStepUp application.BeginSignInResult
	beginAdd    application.RegistrationOptions
	change      application.ChangeEmailCommand
	remove      application.RemovePasskeyCommand
	changeErr   error
	removeErr   error
	stepUpErr   error
	addErr      error
}

func (f *fakeAccountSecurityService) BeginStepUp(context.Context, application.BeginStepUpCommand) (application.BeginSignInResult, error) {
	if f.stepUpErr != nil {
		return application.BeginSignInResult{}, f.stepUpErr
	}
	return f.beginStepUp, nil
}

func (f *fakeAccountSecurityService) FinishStepUp(context.Context, application.FinishStepUpCommand) error {
	return f.stepUpErr
}

func (f *fakeAccountSecurityService) BeginAddPasskey(context.Context, application.BeginAddPasskeyCommand) (application.RegistrationOptions, error) {
	if f.addErr != nil {
		return application.RegistrationOptions{}, f.addErr
	}
	return f.beginAdd, nil
}

func (f *fakeAccountSecurityService) FinishAddPasskey(context.Context, application.FinishAddPasskeyCommand) error {
	return f.addErr
}

func (f *fakeAccountSecurityService) RemovePasskey(_ context.Context, command application.RemovePasskeyCommand) error {
	f.remove = command
	return f.removeErr
}

func (f *fakeAccountSecurityService) ChangeEmail(_ context.Context, command application.ChangeEmailCommand) error {
	f.change = command
	return f.changeErr
}

func TestAccountPasskeyRemoveLastPasskeyMessage(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	current, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	security := &fakeAccountSecurityService{removeErr: domain.ErrLastPasskey}
	server := newServer(t, handlers.Options{Sessions: store, AccountSecurity: security})
	cookie := &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken}
	form := url.Values{middleware.CSRFFormField: {current.CSRFToken}, "credential_id": {"cred-1"}}
	response := request(t, server, http.MethodPost, "/account/security/passkeys/remove", form.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, cookie)
	if response.Code != http.StatusSeeOther || !strings.Contains(response.Header().Get("Location"), "error=") {
		t.Fatalf("status=%d location=%q", response.Code, response.Header().Get("Location"))
	}
}

func TestAccountSecurityPageRedirectsAnonymous(t *testing.T) {
	server := newServer(t, handlers.Options{AccountSecurityReader: &fakeAccountSecurityReader{}})
	response := request(t, server, http.MethodGet, "/account/security", "", nil, nil)
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/sign-in" {
		t.Fatalf("status=%d location=%q", response.Code, response.Header().Get("Location"))
	}
}

func TestSeedConflictingIdentityFixtureEndpoint(t *testing.T) {
	fixture := &fakeIdentityFixture{}
	server := newServer(t, handlers.Options{
		TestControlEnabled: true,
		TestControlKey:     "secret",
		IdentityFixture:    fixture,
	})
	response := request(t, server, http.MethodPost, "/_test/identity/conflict-email", `{"email":"taken@example.org"}`, map[string]string{
		"Content-Type": "application/json", "X-Test-Control-Key": "secret",
	}, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
	}
}

func TestAccountSecurityUnavailableAndMalformedJSON(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	current, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	cookie := &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken}

	noService := newServer(t, handlers.Options{Sessions: store})
	for _, target := range []string{
		"/account/security/step-up/begin",
		"/account/security/step-up/finish",
		"/account/security/passkeys/begin",
		"/account/security/passkeys/finish",
	} {
		response := request(t, noService, http.MethodPost, target, `{"ceremonyId":"x","credential":{"id":"c"}}`, map[string]string{
			"Content-Type": "application/json", "X-CSRF-Token": current.CSRFToken,
		}, cookie)
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s status=%d", target, response.Code)
		}
	}

	security := &fakeAccountSecurityService{}
	server := newServer(t, handlers.Options{Sessions: store, AccountSecurity: security})
	malformed := request(t, server, http.MethodPost, "/account/security/step-up/finish", `{`, map[string]string{
		"Content-Type": "application/json", "X-CSRF-Token": current.CSRFToken,
	}, cookie)
	if malformed.Code != http.StatusBadRequest {
		t.Fatalf("malformed status=%d", malformed.Code)
	}
	security.stepUpErr = domain.ErrCeremonyFailed
	failed := request(t, server, http.MethodPost, "/account/security/step-up/begin", `{}`, map[string]string{
		"Content-Type": "application/json", "X-CSRF-Token": current.CSRFToken,
	}, cookie)
	if failed.Code != http.StatusBadRequest {
		t.Fatalf("failed begin status=%d", failed.Code)
	}
	security.addErr = domain.ErrStepUpRequired
	addDenied := request(t, server, http.MethodPost, "/account/security/passkeys/begin", `{}`, map[string]string{
		"Content-Type": "application/json", "X-CSRF-Token": current.CSRFToken,
	}, cookie)
	if addDenied.Code != http.StatusForbidden {
		t.Fatalf("add denied status=%d", addDenied.Code)
	}
}

func TestAccountSecurityPageUnavailableReader(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	_, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	cookie := &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken}
	server := newServer(t, handlers.Options{Sessions: store})
	response := request(t, server, http.MethodGet, "/account/security", "", nil, cookie)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestAccountPasskeyFinishAndRemoveUnavailable(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	current, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	cookie := &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken}
	server := newServer(t, handlers.Options{Sessions: store})
	finish := request(t, server, http.MethodPost, "/account/security/passkeys/finish", `{"ceremonyId":"x","credential":{"id":"c"}}`, map[string]string{
		"Content-Type": "application/json", "X-CSRF-Token": current.CSRFToken,
	}, cookie)
	if finish.Code != http.StatusServiceUnavailable {
		t.Fatalf("finish status=%d", finish.Code)
	}
	form := url.Values{middleware.CSRFFormField: {current.CSRFToken}, "credential_id": {"cred-1"}}
	remove := request(t, server, http.MethodPost, "/account/security/passkeys/remove", form.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, cookie)
	if remove.Code != http.StatusServiceUnavailable {
		t.Fatalf("remove status=%d", remove.Code)
	}
	change := request(t, server, http.MethodPost, "/account/email", form.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, cookie)
	if change.Code != http.StatusServiceUnavailable {
		t.Fatalf("change status=%d", change.Code)
	}
}

func TestAccountPasskeyFinishMalformedAndError(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	current, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	cookie := &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken}
	security := &fakeAccountSecurityService{addErr: domain.ErrCeremonyFailed}
	server := newServer(t, handlers.Options{Sessions: store, AccountSecurity: security})
	malformed := request(t, server, http.MethodPost, "/account/security/passkeys/finish", `{}`, map[string]string{
		"Content-Type": "application/json", "X-CSRF-Token": current.CSRFToken,
	}, cookie)
	if malformed.Code != http.StatusBadRequest {
		t.Fatalf("malformed status=%d", malformed.Code)
	}
	failed := request(t, server, http.MethodPost, "/account/security/passkeys/finish", `{"ceremonyId":"x","credential":{"id":"c"}}`, map[string]string{
		"Content-Type": "application/json", "X-CSRF-Token": current.CSRFToken,
	}, cookie)
	if failed.Code != http.StatusBadRequest {
		t.Fatalf("failed status=%d", failed.Code)
	}
}

func TestAccountSecurityPageNotFoundProfile(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	_, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	cookie := &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken}
	reader := &fakeAccountSecurityReader{err: application.ErrAccountNotFound}
	server := newServer(t, handlers.Options{Sessions: store, AccountSecurityReader: reader})
	response := request(t, server, http.MethodGet, "/account/security", "", nil, cookie)
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/" {
		t.Fatalf("status=%d location=%q", response.Code, response.Header().Get("Location"))
	}
}

func TestAccountSecurityPageReaderErrors(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	_, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	cookie := &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken}
	reader := &fakeAccountSecurityReader{err: errors.New("profile boom")}
	server := newServer(t, handlers.Options{Sessions: store, AccountSecurityReader: reader})
	response := request(t, server, http.MethodGet, "/account/security", "", nil, cookie)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("profile err status=%d", response.Code)
	}
	reader = &fakeAccountSecurityReader{
		profile:     application.AccountProfile{IdentityID: "identity-1", DisplayName: "Ada", PrimaryEmail: "ada@example.org"},
		passkeysErr: errors.New("passkeys boom"),
	}
	server = newServer(t, handlers.Options{Sessions: store, AccountSecurityReader: reader})
	response = request(t, server, http.MethodGet, "/account/security", "", nil, cookie)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("passkeys err status=%d", response.Code)
	}
}

func TestAccountSecurityAnonymousMutationsRedirect(t *testing.T) {
	server := newServer(t, handlers.Options{AccountSecurity: &fakeAccountSecurityService{}})
	cookie, csrf := establishSession(t, server)
	form := url.Values{middleware.CSRFFormField: {csrf}, "credential_id": {"cred-1"}}
	remove := request(t, server, http.MethodPost, "/account/security/passkeys/remove", form.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, cookie)
	if remove.Code != http.StatusSeeOther || remove.Header().Get("Location") != "/sign-in" {
		t.Fatalf("remove status=%d location=%q", remove.Code, remove.Header().Get("Location"))
	}
	form = url.Values{middleware.CSRFFormField: {csrf}, "email": {"new@example.org"}}
	change := request(t, server, http.MethodPost, "/account/email", form.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, cookie)
	if change.Code != http.StatusSeeOther || change.Header().Get("Location") != "/sign-in" {
		t.Fatalf("change status=%d location=%q", change.Code, change.Header().Get("Location"))
	}
}

func TestAccountStepUpFinishServiceError(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour, session.TestKey)
	current, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	cookie := &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken}
	security := &fakeAccountSecurityService{stepUpErr: domain.ErrCeremonyFailed}
	server := newServer(t, handlers.Options{Sessions: store, AccountSecurity: security})
	response := request(t, server, http.MethodPost, "/account/security/step-up/finish", `{"ceremonyId":"x","credential":{"id":"c"}}`, map[string]string{
		"Content-Type": "application/json", "X-CSRF-Token": current.CSRFToken,
	}, cookie)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
	}
}
