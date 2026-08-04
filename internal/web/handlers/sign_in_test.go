package handlers_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/session"
	"github.com/troop900/treelot/internal/web/handlers"
	"github.com/troop900/treelot/internal/web/middleware"
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

type fakeSignInService struct {
	begin         application.BeginSignInResult
	finish        application.SignInResult
	beginCommand  application.BeginSignInCommand
	finishCommand application.FinishSignInCommand
}

func (s *fakeSignInService) BeginSignIn(_ context.Context, command application.BeginSignInCommand) (application.BeginSignInResult, error) {
	s.beginCommand = command
	return s.begin, nil
}

func (s *fakeSignInService) FinishSignIn(_ context.Context, command application.FinishSignInCommand) (application.SignInResult, error) {
	s.finishCommand = command
	return s.finish, nil
}
