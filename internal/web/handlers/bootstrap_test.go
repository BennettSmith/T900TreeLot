package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
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
	"github.com/troop900/treelot/internal/web/views"
)

func TestBootstrapStartRejectsBadTokenGenerically(t *testing.T) {
	t.Parallel()

	server := newServer(t, handlers.Options{
		Bootstrap: &fakeBootstrapService{startErr: domain.ErrInvalidToken},
	})
	cookie, csrf := establishSession(t, server)
	form := url.Values{
		"bootstrap_token":        {"bad-token"},
		middleware.CSRFFormField: {csrf},
	}

	response := request(t, server, http.MethodPost, "/bootstrap/start", form.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, cookie)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	body := response.Body.String()
	if !strings.Contains(body, "Bootstrap enrollment is unavailable") {
		t.Fatalf("generic bootstrap message missing: %s", body)
	}
	if strings.Contains(strings.ToLower(body), "invalid token") {
		t.Fatalf("response leaked token detail: %s", body)
	}
}

func TestBootstrapEntryStartClaimAndFinishFlow(t *testing.T) {
	t.Parallel()

	fake := &fakeBootstrapService{}
	server := newServer(t, handlers.Options{Bootstrap: fake})

	entry := request(t, server, http.MethodGet, "/bootstrap", "", nil, nil)
	if entry.Code != http.StatusOK {
		t.Fatalf("entry status = %d", entry.Code)
	}
	if !strings.Contains(entry.Body.String(), `data-bootstrap-entry`) || !strings.Contains(entry.Body.String(), `name="bootstrap_token"`) {
		t.Fatalf("entry page missing token form: %s", entry.Body.String())
	}
	cookie := firstCookie(entry, middleware.SessionCookieName)
	csrf := csrfFromHome(entry.Body.String())

	start := request(t, server, http.MethodPost, "/bootstrap/start", url.Values{
		"bootstrap_token":        {"good-token"},
		middleware.CSRFFormField: {csrf},
	}.Encode(), map[string]string{"Content-Type": "application/x-www-form-urlencoded"}, cookie)
	if start.Code != http.StatusOK || !strings.Contains(start.Body.String(), `action="/bootstrap/claim"`) {
		t.Fatalf("start status=%d body=%s", start.Code, start.Body.String())
	}

	claim := request(t, server, http.MethodPost, "/bootstrap/claim", url.Values{
		"bootstrap_token":          {"good-token"},
		"email":                    {"first@example.org"},
		"first_name":               {"First"},
		"last_name":                {"Admin"},
		"preferred_display_name":   {"First Admin"},
		middleware.CSRFFormField:   {csrf},
		"unexpected_ignored_field": {"ignored"},
	}.Encode(), map[string]string{"Content-Type": "application/x-www-form-urlencoded"}, cookie)
	if claim.Code != http.StatusOK || !strings.Contains(claim.Body.String(), `data-bootstrap-passkey`) {
		t.Fatalf("claim status=%d body=%s", claim.Code, claim.Body.String())
	}

	finish := request(t, server, http.MethodPost, "/bootstrap/passkey/finish", `{"token":"good-token","email":"first@example.org","first_name":"First","last_name":"Admin","preferred_display_name":"First Admin","ceremony_id":"ceremony-1","credential":{"id":"credential","rawId":"credential","type":"public-key","response":{"clientDataJSON":"x","attestationObject":"y"}}}`, map[string]string{
		"Content-Type":        "application/json",
		middleware.CSRFHeader: csrf,
	}, cookie)
	if finish.Code != http.StatusSeeOther {
		t.Fatalf("finish status=%d body=%s", finish.Code, finish.Body.String())
	}
	if location := finish.Header().Get("Location"); location != "/account" {
		t.Fatalf("finish location = %q", location)
	}
	rotated := firstCookie(finish, middleware.SessionCookieName)
	if rotated == nil || rotated.Value != "rotated-token" || !rotated.HttpOnly || rotated.Path != "/" {
		t.Fatalf("rotated session cookie = %#v", rotated)
	}
}

func TestBootstrapPostsRequireCSRF(t *testing.T) {
	t.Parallel()

	server := newServer(t, handlers.Options{Bootstrap: &fakeBootstrapService{}})
	response := request(t, server, http.MethodPost, "/bootstrap/start", url.Values{"bootstrap_token": {"token"}}.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}, nil)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}

func TestBootstrapErrorsMapToGenericJSONResponses(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		err    error
		status int
		want   string
	}{
		{name: "rate limit", err: domain.ErrRateLimited, status: http.StatusTooManyRequests, want: "Too many attempts"},
		{name: "ceremony", err: domain.ErrCeremonyFailed, status: http.StatusBadRequest, want: "Passkey registration could not be completed"},
	} {
		t.Run(test.name, func(t *testing.T) {
			fake := &fakeBootstrapService{claimErr: test.err}
			server := newServer(t, handlers.Options{Bootstrap: fake})
			cookie, csrf := establishSession(t, server)
			response := request(t, server, http.MethodPost, "/bootstrap/passkey/begin", `{"token":"good","email":"first@example.org","first_name":"First","last_name":"Admin"}`, map[string]string{
				"Content-Type":        "application/json",
				middleware.CSRFHeader: csrf,
			}, cookie)
			if response.Code != test.status || !strings.Contains(response.Body.String(), test.want) {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func TestBootstrapFormErrorsAndUnavailableServices(t *testing.T) {
	t.Parallel()

	server := newServer(t, handlers.Options{Bootstrap: &fakeBootstrapService{claimErr: domain.ErrInvalidProfile}})
	cookie, csrf := establishSession(t, server)
	claim := request(t, server, http.MethodPost, "/bootstrap/claim", url.Values{
		"bootstrap_token":        {"good"},
		"email":                  {"first@example.org"},
		middleware.CSRFFormField: {csrf},
	}.Encode(), map[string]string{"Content-Type": "application/x-www-form-urlencoded"}, cookie)
	if claim.Code != http.StatusOK || !strings.Contains(claim.Body.String(), "Enter a valid email") {
		t.Fatalf("claim status=%d body=%s", claim.Code, claim.Body.String())
	}

	unavailable := newServer(t, handlers.Options{})
	cookie, csrf = establishSession(t, unavailable)
	for _, target := range []string{"/bootstrap/start", "/bootstrap/claim"} {
		response := request(t, unavailable, http.MethodPost, target, url.Values{middleware.CSRFFormField: {csrf}}.Encode(), map[string]string{"Content-Type": "application/x-www-form-urlencoded"}, cookie)
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s status=%d", target, response.Code)
		}
	}
	for _, target := range []string{"/bootstrap/passkey/begin", "/bootstrap/passkey/finish"} {
		response := request(t, unavailable, http.MethodPost, target, `{"csrf_token":"`+csrf+`"}`, map[string]string{"Content-Type": "application/json"}, cookie)
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s status=%d", target, response.Code)
		}
	}
}

func TestBootstrapPasskeyBeginAcceptsJSONWithCSRFHeader(t *testing.T) {
	t.Parallel()

	fake := &fakeBootstrapService{
		beginResult: application.RegistrationOptions{
			CeremonyID: "ceremony-1",
			PublicKey:  map[string]any{"publicKey": map[string]any{"challenge": "abc"}},
			ExpiresAt:  time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	server := newServer(t, handlers.Options{Bootstrap: fake})
	cookie, csrf := establishSession(t, server)
	body := `{"token":"good","email":"first@example.org","first_name":"First","last_name":"Admin"}`

	response := request(t, server, http.MethodPost, "/bootstrap/passkey/begin", body, map[string]string{
		"Content-Type":        "application/json",
		middleware.CSRFHeader: csrf,
	}, cookie)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if payload["ceremonyId"] != "ceremony-1" || payload["publicKey"] == nil {
		t.Fatalf("payload = %#v", payload)
	}
	if fake.beginSessionID == 0 {
		t.Fatal("begin did not receive the current session id")
	}
}

func TestBootstrapPasskeyBeginAcceptsFormAndFinishRejectsBadJSON(t *testing.T) {
	t.Parallel()

	server := newServer(t, handlers.Options{Bootstrap: &fakeBootstrapService{}})
	cookie, csrf := establishSession(t, server)
	form := url.Values{
		"bootstrap_token":        {"good"},
		"email":                  {"first@example.org"},
		"first_name":             {"First"},
		"last_name":              {"Admin"},
		middleware.CSRFFormField: {csrf},
	}
	begin := request(t, server, http.MethodPost, "/bootstrap/passkey/begin", form.Encode(), map[string]string{"Content-Type": "application/x-www-form-urlencoded"}, cookie)
	if begin.Code != http.StatusOK {
		t.Fatalf("begin form status=%d body=%s", begin.Code, begin.Body.String())
	}
	finish := request(t, server, http.MethodPost, "/bootstrap/passkey/finish", `{`, map[string]string{
		"Content-Type":        "application/json",
		middleware.CSRFHeader: csrf,
	}, cookie)
	if finish.Code != http.StatusBadRequest {
		t.Fatalf("finish bad json status=%d", finish.Code)
	}
}

func TestBootstrapPasskeyBeginAndFinishMapAdditionalErrors(t *testing.T) {
	t.Parallel()

	beginServer := newServer(t, handlers.Options{Bootstrap: &fakeBootstrapService{beginErr: domain.ErrCeremonyFailed}})
	cookie, csrf := establishSession(t, beginServer)
	badBegin := request(t, beginServer, http.MethodPost, "/bootstrap/passkey/begin", `{`, map[string]string{
		"Content-Type":        "application/json",
		middleware.CSRFHeader: csrf,
	}, cookie)
	if badBegin.Code != http.StatusBadRequest {
		t.Fatalf("bad begin status=%d", badBegin.Code)
	}
	begin := request(t, beginServer, http.MethodPost, "/bootstrap/passkey/begin", `{"bootstrap_token":"good","email":"first@example.org","first_name":"First","last_name":"Admin"}`, map[string]string{
		"Content-Type":        "application/json",
		middleware.CSRFHeader: csrf,
	}, cookie)
	if begin.Code != http.StatusBadRequest || !strings.Contains(begin.Body.String(), "Passkey registration could not be completed") {
		t.Fatalf("begin ceremony status=%d body=%s", begin.Code, begin.Body.String())
	}

	finishServer := newServer(t, handlers.Options{Bootstrap: &fakeBootstrapService{finishErr: domain.ErrRateLimited}})
	cookie, csrf = establishSession(t, finishServer)
	missingCredential := request(t, finishServer, http.MethodPost, "/bootstrap/passkey/finish", `{"token":"good","ceremonyId":"ceremony-1"}`, map[string]string{
		"Content-Type":        "application/json",
		middleware.CSRFHeader: csrf,
	}, cookie)
	if missingCredential.Code != http.StatusBadRequest {
		t.Fatalf("missing credential status=%d", missingCredential.Code)
	}
	rateLimited := request(t, finishServer, http.MethodPost, "/bootstrap/passkey/finish", `{"token":"good","email":"first@example.org","first_name":"First","last_name":"Admin","ceremonyId":"ceremony-1","credential":{"id":"credential"}}`, map[string]string{
		"Content-Type":        "application/json",
		middleware.CSRFHeader: csrf,
	}, cookie)
	if rateLimited.Code != http.StatusTooManyRequests || !strings.Contains(rateLimited.Body.String(), "Too many attempts") {
		t.Fatalf("finish rate status=%d body=%s", rateLimited.Code, rateLimited.Body.String())
	}
}

func TestAccountRequiresAuthenticatedSession(t *testing.T) {
	t.Parallel()

	server := newServer(t, handlers.Options{})
	response := request(t, server, http.MethodGet, "/account", "", nil, nil)

	if response.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", response.Code)
	}
	if location := response.Header().Get("Location"); location != "/" {
		t.Fatalf("Location = %q, want /", location)
	}
}

func TestAccountHandlesMissingAndFailedProfileLookup(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		err    error
		status int
	}{
		{name: "missing", err: application.ErrAccountNotFound, status: http.StatusSeeOther},
		{name: "failed", err: errors.New("database down"), status: http.StatusInternalServerError},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := session.NewMemoryStore(clock.System(), time.Hour)
			_, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
			if err != nil {
				t.Fatalf("CreateForIdentity: %v", err)
			}
			server := newServer(t, handlers.Options{
				Sessions: store,
				Accounts: fakeAccountReader{err: test.err},
			})
			response := request(t, server, http.MethodGet, "/account", "", nil, &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken})
			if response.Code != test.status {
				t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func TestAccountWelcomesAuthenticatedAdmin(t *testing.T) {
	t.Parallel()

	store := session.NewMemoryStore(clock.System(), time.Hour)
	created, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatalf("CreateForIdentity: %v", err)
	}
	if created.IdentityID != "identity-1" {
		t.Fatalf("created session identity = %q", created.IdentityID)
	}
	server := newServer(t, handlers.Options{
		Sessions: store,
		Accounts: fakeAccountReader{
			profile: application.AccountProfile{IdentityID: "identity-1", DisplayName: "Ada Admin"},
		},
	})

	response := request(t, server, http.MethodGet, "/account", "", nil, &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken})

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "Welcome, Ada Admin") {
		t.Fatalf("account page missing welcome: %s", response.Body.String())
	}
}

type fakeBootstrapService struct {
	startErr       error
	claimErr       error
	beginErr       error
	finishErr      error
	beginResult    application.RegistrationOptions
	beginSessionID int64
}

func (f *fakeBootstrapService) StartBootstrap(context.Context, application.StartBootstrapCommand) (application.StartBootstrapResult, error) {
	if f.startErr != nil {
		return application.StartBootstrapResult{}, f.startErr
	}
	return application.StartBootstrapResult{ExpiresAt: time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)}, nil
}

func (f *fakeBootstrapService) ClaimBootstrapProfile(_ context.Context, command application.ClaimBootstrapProfileCommand) (application.PendingEnrollment, error) {
	if f.claimErr != nil {
		return application.PendingEnrollment{}, f.claimErr
	}
	email, err := domain.NewEmail(command.Email)
	if err != nil {
		return application.PendingEnrollment{}, err
	}
	name, err := domain.ValidateProfile(command.FirstName, command.LastName, command.PreferredDisplayName)
	if err != nil {
		return application.PendingEnrollment{}, err
	}
	return application.PendingEnrollment{Token: command.Token, Email: email, Name: name, ValidUntil: time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)}, nil
}

func (f *fakeBootstrapService) BeginPasskeyRegistration(_ context.Context, command application.BeginPasskeyRegistrationCommand) (application.RegistrationOptions, error) {
	f.beginSessionID = command.SessionID
	if f.beginErr != nil {
		return application.RegistrationOptions{}, f.beginErr
	}
	if f.beginResult.CeremonyID != "" {
		return f.beginResult, nil
	}
	return application.RegistrationOptions{CeremonyID: "ceremony-1", PublicKey: map[string]any{"publicKey": map[string]any{"challenge": "abc"}}}, nil
}

func (f *fakeBootstrapService) FinishBootstrap(context.Context, application.FinishBootstrapCommand) (application.BootstrapResult, error) {
	if f.finishErr != nil {
		return application.BootstrapResult{}, f.finishErr
	}
	return application.BootstrapResult{Session: application.IssuedSession{
		ID:              42,
		IdentityID:      "identity-1",
		RawToken:        "rotated-token",
		AuthenticatedAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		ExpiresAt:       time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC),
	}}, nil
}

type fakeAccountReader struct {
	profile application.AccountProfile
	err     error
}

func (f fakeAccountReader) FindAccountProfile(context.Context, string) (application.AccountProfile, error) {
	if f.err != nil {
		return application.AccountProfile{}, f.err
	}
	return f.profile, nil
}

func (failingRenderer) Bootstrap(context.Context, io.Writer, views.BootstrapPage) error {
	return errors.New("representative render failure")
}

func (failingRenderer) Account(context.Context, io.Writer, views.AccountPage) error {
	return errors.New("representative render failure")
}
