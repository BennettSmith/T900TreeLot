package handlers_test

import (
	"context"
	"errors"
	"net/http"
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

func TestSignInRoutesBeginAndFinishPasskeyAssertion(t *testing.T) {
	service := &fakeSignInService{
		begin: application.BeginSignInResult{
			CeremonyID: "ceremony-1",
			PublicKey:  map[string]any{"challenge": "challenge"},
		},
		finish: application.SignInResult{
			IdentityID: "identity-1",
			RedirectTo: "/family",
			Session: application.IssuedSession{
				RawToken:  "signed-in-token",
				ExpiresAt: time.Now().Add(time.Hour),
			},
		},
	}
	server := newServer(t, handlers.Options{
		SignIn:   service,
		Sessions: session.NewMemoryStore(clock.System(), time.Hour),
	})

	entry := request(t, server, http.MethodGet, "/sign-in", "", nil, nil)
	if entry.Code != http.StatusOK || !strings.Contains(entry.Body.String(), "Sign in with a passkey") {
		t.Fatalf("entry status=%d body=%q", entry.Code, entry.Body.String())
	}
	cookie := firstCookie(entry, middleware.SessionCookieName)
	csrf := csrfFromHome(entry.Body.String())
	begin := request(t, server, http.MethodPost, "/sign-in/passkey/begin", `{}`, map[string]string{
		"Content-Type": "application/json",
		"X-CSRF-Token": csrf,
	}, cookie)
	if begin.Code != http.StatusOK || !strings.Contains(begin.Body.String(), `"ceremonyId":"ceremony-1"`) {
		t.Fatalf("begin status=%d body=%q", begin.Code, begin.Body.String())
	}
	finish := request(t, server, http.MethodPost, "/sign-in/passkey/finish", `{"ceremonyId":"ceremony-1","credential":{"id":"credential"}}`, map[string]string{
		"Content-Type": "application/json",
		"X-CSRF-Token": csrf,
	}, cookie)
	if finish.Code != http.StatusOK || !strings.Contains(finish.Body.String(), `"redirectTo":"/family"`) {
		t.Fatalf("finish status=%d body=%q", finish.Code, finish.Body.String())
	}
	if service.beginCommand.SessionID == 0 || service.finishCommand.SessionID != service.beginCommand.SessionID {
		t.Fatalf("session commands = %#v / %#v", service.beginCommand, service.finishCommand)
	}
}

func TestSignInErrorsRemainGenericAndRateLimitsAreExplicit(t *testing.T) {
	service := &fakeSignInService{beginErr: domain.ErrRateLimited}
	server := newServer(t, handlers.Options{SignIn: service})
	cookie, csrf := establishSession(t, server)

	rateLimited := request(t, server, http.MethodPost, "/sign-in/passkey/begin", `{}`, map[string]string{
		"Content-Type": "application/json", "X-CSRF-Token": csrf,
	}, cookie)
	if rateLimited.Code != http.StatusTooManyRequests || !strings.Contains(rateLimited.Body.String(), "Too many attempts") {
		t.Fatalf("rate limited status=%d body=%q", rateLimited.Code, rateLimited.Body.String())
	}
	service.beginErr = errors.New("database included secret@example.org")
	failed := request(t, server, http.MethodPost, "/sign-in/passkey/begin", `{"email":"secret@example.org"}`, map[string]string{
		"Content-Type": "application/json", "X-CSRF-Token": csrf,
	}, cookie)
	if failed.Code != http.StatusBadRequest || strings.Contains(failed.Body.String(), "secret@example.org") {
		t.Fatalf("generic begin status=%d body=%q", failed.Code, failed.Body.String())
	}
	service.finishErr = errors.New("credential details")
	finish := request(t, server, http.MethodPost, "/sign-in/passkey/finish", `{"ceremonyId":"one","credential":{"id":"bad"}}`, map[string]string{
		"Content-Type": "application/json", "X-CSRF-Token": csrf,
	}, cookie)
	if finish.Code != http.StatusBadRequest || !strings.Contains(finish.Body.String(), "Sign-in could not be completed") {
		t.Fatalf("generic finish status=%d body=%q", finish.Code, finish.Body.String())
	}
}

func TestSignInEndpointsRejectUnavailableMissingSessionAndMalformedRequests(t *testing.T) {
	unavailable := newServer(t, handlers.Options{})
	cookie, csrf := establishSession(t, unavailable)
	for _, target := range []string{"/sign-in/passkey/begin", "/sign-in/passkey/finish"} {
		response := request(t, unavailable, http.MethodPost, target, `{}`, map[string]string{
			"Content-Type": "application/json", "X-CSRF-Token": csrf,
		}, cookie)
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s unavailable status=%d", target, response.Code)
		}
	}

	renderer, err := views.NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	noSession := handlers.New(renderer, handlers.Options{SignIn: &fakeSignInService{}})
	for _, target := range []string{"/sign-in/passkey/begin", "/sign-in/passkey/finish"} {
		response := request(t, noSession, http.MethodPost, target, `{}`, map[string]string{"Content-Type": "application/json"}, nil)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s missing session status=%d", target, response.Code)
		}
	}

	service := &fakeSignInService{}
	malformedServer := newServer(t, handlers.Options{SignIn: service})
	cookie, csrf = establishSession(t, malformedServer)
	for _, target := range []string{"/sign-in/passkey/begin", "/sign-in/passkey/finish"} {
		response := request(t, malformedServer, http.MethodPost, target, `{`, map[string]string{
			"Content-Type": "application/json", "X-CSRF-Token": csrf,
		}, cookie)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("%s malformed status=%d", target, response.Code)
		}
	}
}

func TestRoleLandingsRequireAuthenticatedMatchingRole(t *testing.T) {
	store := session.NewMemoryStore(clock.System(), time.Hour)
	_, rawToken, err := store.CreateForIdentity(context.Background(), "identity-1")
	if err != nil {
		t.Fatal(err)
	}
	landings := &fakeLandingReader{profile: application.LandingProfile{
		IdentityID: "identity-1", DisplayName: "Family Manager", Roles: []domain.Role{domain.RoleFamilyManager},
	}}
	server := newServer(t, handlers.Options{Sessions: store, Landings: landings})
	cookie := &http.Cookie{Name: middleware.SessionCookieName, Value: rawToken}

	family := request(t, server, http.MethodGet, "/family", "", nil, cookie)
	if family.Code != http.StatusOK || !strings.Contains(family.Body.String(), "Family dashboard") {
		t.Fatalf("family status=%d body=%q", family.Code, family.Body.String())
	}
	scoutDenied := request(t, server, http.MethodGet, "/scout/schedule", "", nil, cookie)
	if scoutDenied.Code != http.StatusForbidden {
		t.Fatalf("scout denied status=%d", scoutDenied.Code)
	}
	anonymous := request(t, server, http.MethodGet, "/family", "", nil, nil)
	if anonymous.Code != http.StatusSeeOther || anonymous.Header().Get("Location") != "/sign-in" {
		t.Fatalf("anonymous status=%d location=%q", anonymous.Code, anonymous.Header().Get("Location"))
	}

	landings.profile = application.LandingProfile{
		IdentityID: "identity-1", DisplayName: "Young Adult Scout", Roles: []domain.Role{domain.RoleYoungAdultScout},
	}
	scout := request(t, server, http.MethodGet, "/scout/schedule", "", nil, cookie)
	if scout.Code != http.StatusOK || !strings.Contains(scout.Body.String(), "Personal schedule") || strings.Contains(strings.ToLower(scout.Body.String()), "household") {
		t.Fatalf("scout status=%d body=%q", scout.Code, scout.Body.String())
	}

	landings.err = application.ErrAccountNotFound
	notFound := request(t, server, http.MethodGet, "/family", "", nil, cookie)
	if notFound.Code != http.StatusSeeOther || notFound.Header().Get("Location") != "/sign-in" {
		t.Fatalf("missing profile status=%d location=%q", notFound.Code, notFound.Header().Get("Location"))
	}
	landings.err = errors.New("lookup unavailable")
	failed := request(t, server, http.MethodGet, "/family", "", nil, cookie)
	if failed.Code != http.StatusInternalServerError {
		t.Fatalf("failed landing status=%d", failed.Code)
	}
	noReader := newServer(t, handlers.Options{Sessions: store})
	unavailable := request(t, noReader, http.MethodGet, "/family", "", nil, cookie)
	if unavailable.Code != http.StatusServiceUnavailable {
		t.Fatalf("unavailable landing status=%d", unavailable.Code)
	}
}

type fakeSignInService struct {
	begin         application.BeginSignInResult
	finish        application.SignInResult
	beginCommand  application.BeginSignInCommand
	finishCommand application.FinishSignInCommand
	beginErr      error
	finishErr     error
}

func (s *fakeSignInService) BeginSignIn(_ context.Context, command application.BeginSignInCommand) (application.BeginSignInResult, error) {
	s.beginCommand = command
	return s.begin, s.beginErr
}

func (s *fakeSignInService) FinishSignIn(_ context.Context, command application.FinishSignInCommand) (application.SignInResult, error) {
	s.finishCommand = command
	return s.finish, s.finishErr
}

type fakeLandingReader struct {
	profile application.LandingProfile
	err     error
}

func (r *fakeLandingReader) FindLandingProfile(context.Context, string) (application.LandingProfile, error) {
	return r.profile, r.err
}
