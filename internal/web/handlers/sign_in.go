package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/identity/domain"
	"github.com/troop900/treelot/internal/web/middleware"
	"github.com/troop900/treelot/internal/web/views"
)

const (
	signInFailureMessage   = "Sign-in could not be completed. Try again."
	signInRateLimitMessage = "Too many attempts. Wait a few minutes and try again."
)

func (s *Server) signInEntry(response http.ResponseWriter, request *http.Request) {
	current := middleware.FromContext(request.Context())
	page := views.SignInPage{
		PageTitle:  "Sign in",
		Brand:      "Troop 900 Tree Lot",
		Navigation: []views.Link{{Label: "Home", Href: "/"}, {Label: "Sign in", Href: "/sign-in", Current: true}},
	}
	if current != nil {
		page.CSRFToken = current.CSRFToken
	}
	s.renderHTML(response, request, func(output io.Writer) error {
		return s.renderer.SignIn(request.Context(), output, page)
	})
}

func (s *Server) signInPasskeyBegin(response http.ResponseWriter, request *http.Request) {
	if s.signIn == nil {
		writeSignInJSONError(response, http.StatusServiceUnavailable, signInFailureMessage)
		return
	}
	current := middleware.FromContext(request.Context())
	if current == nil {
		writeSignInJSONError(response, http.StatusUnauthorized, signInFailureMessage)
		return
	}
	var payload struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		writeSignInJSONError(response, http.StatusBadRequest, signInFailureMessage)
		return
	}
	result, err := s.signIn.BeginSignIn(request.Context(), application.BeginSignInCommand{
		SessionID:    current.ID,
		RateLimitKey: signInRateLimitKey(request),
		EmailHint:    strings.TrimSpace(payload.Email),
	})
	if err != nil {
		status, message := signInError(err)
		writeSignInJSONError(response, status, message)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(map[string]any{
		"ceremonyId": result.CeremonyID,
		"publicKey":  result.PublicKey,
	})
}

func (s *Server) signInPasskeyFinish(response http.ResponseWriter, request *http.Request) {
	if s.signIn == nil {
		writeSignInJSONError(response, http.StatusServiceUnavailable, signInFailureMessage)
		return
	}
	current := middleware.FromContext(request.Context())
	if current == nil {
		writeSignInJSONError(response, http.StatusUnauthorized, signInFailureMessage)
		return
	}
	var payload struct {
		CeremonyID string          `json:"ceremonyId"`
		Credential json.RawMessage `json:"credential"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil || payload.CeremonyID == "" || len(payload.Credential) == 0 {
		writeSignInJSONError(response, http.StatusBadRequest, signInFailureMessage)
		return
	}
	result, err := s.signIn.FinishSignIn(request.Context(), application.FinishSignInCommand{
		SessionID:         current.ID,
		PasskeyCeremonyID: payload.CeremonyID,
		PasskeyResponse:   payload.Credential,
	})
	if err != nil {
		status, message := signInError(err)
		writeSignInJSONError(response, status, message)
		return
	}
	middleware.SetSessionCookie(response, result.Session.RawToken, result.Session.ExpiresAt, s.secureCookies)
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(map[string]string{"redirectTo": result.RedirectTo})
}

func (s *Server) signOutCurrentSession(response http.ResponseWriter, request *http.Request) {
	current := middleware.FromContext(request.Context())
	if current == nil || current.IdentityID == "" {
		middleware.ClearSessionCookie(response, s.secureCookies)
		http.Redirect(response, request, "/", http.StatusSeeOther)
		return
	}
	if s.signOut == nil {
		http.Error(response, "Sign-out is unavailable.", http.StatusServiceUnavailable)
		return
	}
	if err := s.signOut.SignOut(request.Context(), application.SignOutCommand{
		IdentityID: current.IdentityID,
		SessionID:  current.ID,
	}); err != nil {
		s.renderError(response, request, err)
		return
	}
	middleware.ClearSessionCookie(response, s.secureCookies)
	http.Redirect(response, request, "/", http.StatusSeeOther)
}

func (s *Server) familyLanding(response http.ResponseWriter, request *http.Request) {
	s.renderRoleLanding(response, request, domain.RoleFamilyManager, views.LandingPage{
		PageTitle:  "Family dashboard",
		Heading:    "Family dashboard",
		Supporting: "Your household schedule and management tools will appear here as they become available.",
		Navigation: []views.Link{{Label: "Home", Href: "/"}, {Label: "Family", Href: "/family", Current: true}},
	})
}

func (s *Server) scoutLanding(response http.ResponseWriter, request *http.Request) {
	s.renderRoleLanding(response, request, domain.RoleYoungAdultScout, views.LandingPage{
		PageTitle:  "Personal schedule",
		Heading:    "Personal schedule",
		Supporting: "Your own tree-lot shifts will appear here as they become available.",
		Navigation: []views.Link{{Label: "Home", Href: "/"}, {Label: "My schedule", Href: "/scout/schedule", Current: true}},
	})
}

func (s *Server) renderRoleLanding(response http.ResponseWriter, request *http.Request, required domain.Role, page views.LandingPage) {
	current := middleware.FromContext(request.Context())
	if current == nil || current.IdentityID == "" {
		http.Redirect(response, request, "/sign-in", http.StatusSeeOther)
		return
	}
	if s.landings == nil {
		http.Error(response, "Account lookup is unavailable.", http.StatusServiceUnavailable)
		return
	}
	profile, err := s.landings.FindLandingProfile(request.Context(), current.IdentityID)
	if errors.Is(err, application.ErrAccountNotFound) {
		http.Redirect(response, request, "/sign-in", http.StatusSeeOther)
		return
	}
	if err != nil {
		s.renderError(response, request, err)
		return
	}
	if !profile.HasRole(required) {
		http.Error(response, "You are not authorized to view this page.", http.StatusForbidden)
		return
	}
	page.Brand = "Troop 900 Tree Lot"
	page.DisplayName = profile.DisplayName
	page.CSRFToken = current.CSRFToken
	s.renderHTML(response, request, func(output io.Writer) error {
		return s.renderer.Landing(request.Context(), output, page)
	})
}

func signInError(err error) (int, string) {
	if errors.Is(err, domain.ErrRateLimited) {
		return http.StatusTooManyRequests, signInRateLimitMessage
	}
	return http.StatusBadRequest, signInFailureMessage
}

func writeSignInJSONError(response http.ResponseWriter, status int, message string) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(map[string]string{"error": message})
}

func signInRateLimitKey(request *http.Request) string {
	host := request.RemoteAddr
	if parsedHost, _, err := net.SplitHostPort(request.RemoteAddr); err == nil {
		host = parsedHost
	}
	host = strings.TrimSpace(host)
	if host == "" {
		host = "unknown"
	}
	return "sign-in:" + host
}
