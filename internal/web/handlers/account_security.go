package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/identity/domain"
	"github.com/troop900/treelot/internal/web/middleware"
	"github.com/troop900/treelot/internal/web/views"
)

const (
	accountSecurityFailureMessage = "That security change could not be completed. Try again."
	accountSecurityStepUpMessage  = "Confirm with your passkey before changing credentials."
	accountSecurityEmailTakenMsg  = "That email cannot be used. Choose a different address."
	accountSecurityLastPasskeyMsg = "Keep at least one passkey registered on this account."
)

type AccountSecurityService interface {
	BeginStepUp(context.Context, application.BeginStepUpCommand) (application.BeginSignInResult, error)
	FinishStepUp(context.Context, application.FinishStepUpCommand) error
	BeginAddPasskey(context.Context, application.BeginAddPasskeyCommand) (application.RegistrationOptions, error)
	FinishAddPasskey(context.Context, application.FinishAddPasskeyCommand) error
	RemovePasskey(context.Context, application.RemovePasskeyCommand) error
	ChangeEmail(context.Context, application.ChangeEmailCommand) error
}

type AccountSecurityReader interface {
	FindAccountProfile(context.Context, string) (application.AccountProfile, error)
	ListPasskeys(context.Context, string) ([]application.PasskeyCredential, error)
}

func (s *Server) accountSecurityPage(response http.ResponseWriter, request *http.Request) {
	current := middleware.FromContext(request.Context())
	if current == nil || current.IdentityID == "" {
		http.Redirect(response, request, "/sign-in", http.StatusSeeOther)
		return
	}
	if s.securityReader == nil {
		http.Error(response, "Account security is unavailable.", http.StatusServiceUnavailable)
		return
	}
	profile, err := s.securityReader.FindAccountProfile(request.Context(), current.IdentityID)
	if errors.Is(err, application.ErrAccountNotFound) {
		http.Redirect(response, request, "/", http.StatusSeeOther)
		return
	}
	if err != nil {
		s.renderError(response, request, err)
		return
	}
	passkeys, err := s.securityReader.ListPasskeys(request.Context(), current.IdentityID)
	if err != nil {
		s.renderError(response, request, err)
		return
	}
	stepUpRequired := domain.RequireRecentStepUp(current.StepUpAt, s.clock.Now().UTC(), s.stepUpTTL) != nil
	passkeyViews := make([]views.AccountPasskeyView, 0, len(passkeys))
	for _, passkey := range passkeys {
		passkeyViews = append(passkeyViews, views.AccountPasskeyView{
			ID:        passkey.ID,
			CreatedAt: passkey.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	data := views.AccountSecurityPage{
		PageTitle:      "Account security",
		Brand:          "Troop 900 Tree Lot",
		DisplayName:    profile.DisplayName,
		PrimaryEmail:   profile.PrimaryEmail,
		CSRFToken:      current.CSRFToken,
		StepUpRequired: stepUpRequired,
		Passkeys:       passkeyViews,
		Message:        request.URL.Query().Get("message"),
		ErrorMessage:   request.URL.Query().Get("error"),
		Navigation: []views.Link{
			{Label: "Home", Href: "/"},
			{Label: "Account", Href: "/account"},
			{Label: "Security", Href: "/account/security", Current: true},
		},
	}
	s.renderHTML(response, request, func(output io.Writer) error {
		return s.renderer.AccountSecurity(request.Context(), output, data)
	})
}

func (s *Server) accountStepUpBegin(response http.ResponseWriter, request *http.Request) {
	current, ok := s.requireAuthenticatedJSON(response, request)
	if !ok {
		return
	}
	if s.securityService == nil {
		writeAccountSecurityJSONError(response, http.StatusServiceUnavailable, accountSecurityFailureMessage)
		return
	}
	result, err := s.securityService.BeginStepUp(request.Context(), application.BeginStepUpCommand{
		IdentityID: current.IdentityID,
		SessionID:  current.ID,
	})
	if err != nil {
		writeAccountSecurityJSONError(response, http.StatusBadRequest, accountSecurityFailureMessage)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(map[string]any{
		"ceremonyId": result.CeremonyID,
		"publicKey":  result.PublicKey,
	})
}

func (s *Server) accountStepUpFinish(response http.ResponseWriter, request *http.Request) {
	current, ok := s.requireAuthenticatedJSON(response, request)
	if !ok {
		return
	}
	if s.securityService == nil {
		writeAccountSecurityJSONError(response, http.StatusServiceUnavailable, accountSecurityFailureMessage)
		return
	}
	var payload struct {
		CeremonyID string          `json:"ceremonyId"`
		Credential json.RawMessage `json:"credential"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil || payload.CeremonyID == "" || len(payload.Credential) == 0 {
		writeAccountSecurityJSONError(response, http.StatusBadRequest, accountSecurityFailureMessage)
		return
	}
	if err := s.securityService.FinishStepUp(request.Context(), application.FinishStepUpCommand{
		IdentityID:        current.IdentityID,
		SessionID:         current.ID,
		PasskeyCeremonyID: payload.CeremonyID,
		PasskeyResponse:   payload.Credential,
	}); err != nil {
		writeAccountSecurityJSONError(response, http.StatusBadRequest, accountSecurityFailureMessage)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(map[string]string{"redirectTo": "/account/security"})
}

func (s *Server) accountPasskeyBegin(response http.ResponseWriter, request *http.Request) {
	current, ok := s.requireAuthenticatedJSON(response, request)
	if !ok {
		return
	}
	if s.securityService == nil {
		writeAccountSecurityJSONError(response, http.StatusServiceUnavailable, accountSecurityFailureMessage)
		return
	}
	result, err := s.securityService.BeginAddPasskey(request.Context(), application.BeginAddPasskeyCommand{
		IdentityID: current.IdentityID,
		SessionID:  current.ID,
	})
	if err != nil {
		status, message := accountSecurityError(err)
		writeAccountSecurityJSONError(response, status, message)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(map[string]any{
		"ceremonyId": result.CeremonyID,
		"publicKey":  result.PublicKey,
	})
}

func (s *Server) accountPasskeyFinish(response http.ResponseWriter, request *http.Request) {
	current, ok := s.requireAuthenticatedJSON(response, request)
	if !ok {
		return
	}
	if s.securityService == nil {
		writeAccountSecurityJSONError(response, http.StatusServiceUnavailable, accountSecurityFailureMessage)
		return
	}
	var payload struct {
		CeremonyID string          `json:"ceremonyId"`
		Credential json.RawMessage `json:"credential"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil || payload.CeremonyID == "" || len(payload.Credential) == 0 {
		writeAccountSecurityJSONError(response, http.StatusBadRequest, accountSecurityFailureMessage)
		return
	}
	if err := s.securityService.FinishAddPasskey(request.Context(), application.FinishAddPasskeyCommand{
		IdentityID:        current.IdentityID,
		SessionID:         current.ID,
		PasskeyCeremonyID: payload.CeremonyID,
		PasskeyResponse:   payload.Credential,
	}); err != nil {
		status, message := accountSecurityError(err)
		writeAccountSecurityJSONError(response, status, message)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(map[string]string{"redirectTo": "/account/security?message=" + url.QueryEscape("Passkey added.")})
}

func (s *Server) accountPasskeyRemove(response http.ResponseWriter, request *http.Request) {
	current := middleware.FromContext(request.Context())
	if current == nil || current.IdentityID == "" {
		http.Redirect(response, request, "/sign-in", http.StatusSeeOther)
		return
	}
	if s.securityService == nil {
		http.Error(response, "Account security is unavailable.", http.StatusServiceUnavailable)
		return
	}
	if err := request.ParseForm(); err != nil {
		http.Redirect(response, request, "/account/security?error="+url.QueryEscape(accountSecurityFailureMessage), http.StatusSeeOther)
		return
	}
	err := s.securityService.RemovePasskey(request.Context(), application.RemovePasskeyCommand{
		IdentityID:   current.IdentityID,
		SessionID:    current.ID,
		CredentialID: strings.TrimSpace(request.FormValue("credential_id")),
	})
	if err != nil {
		_, message := accountSecurityError(err)
		http.Redirect(response, request, "/account/security?error="+url.QueryEscape(message), http.StatusSeeOther)
		return
	}
	http.Redirect(response, request, "/account/security?message="+url.QueryEscape("Passkey removed."), http.StatusSeeOther)
}

func (s *Server) accountChangeEmail(response http.ResponseWriter, request *http.Request) {
	current := middleware.FromContext(request.Context())
	if current == nil || current.IdentityID == "" {
		http.Redirect(response, request, "/sign-in", http.StatusSeeOther)
		return
	}
	if s.securityService == nil {
		http.Error(response, "Account security is unavailable.", http.StatusServiceUnavailable)
		return
	}
	if err := request.ParseForm(); err != nil {
		http.Redirect(response, request, "/account/security?error="+url.QueryEscape(accountSecurityFailureMessage), http.StatusSeeOther)
		return
	}
	err := s.securityService.ChangeEmail(request.Context(), application.ChangeEmailCommand{
		IdentityID: current.IdentityID,
		SessionID:  current.ID,
		NewEmail:   request.FormValue("email"),
	})
	if err != nil {
		status, message := accountSecurityError(err)
		if status == http.StatusForbidden {
			http.Error(response, message, http.StatusForbidden)
			return
		}
		http.Redirect(response, request, "/account/security?error="+url.QueryEscape(message), http.StatusSeeOther)
		return
	}
	middleware.ClearSessionCookie(response, s.secureCookies)
	http.Redirect(response, request, "/sign-in?message="+url.QueryEscape("Email updated. Sign in again with your passkey."), http.StatusSeeOther)
}

func (s *Server) requireAuthenticatedJSON(response http.ResponseWriter, request *http.Request) (*sessionFromRequest, bool) {
	current := middleware.FromContext(request.Context())
	if current == nil || current.IdentityID == "" {
		writeAccountSecurityJSONError(response, http.StatusUnauthorized, accountSecurityFailureMessage)
		return nil, false
	}
	return &sessionFromRequest{ID: current.ID, IdentityID: current.IdentityID}, true
}

type sessionFromRequest struct {
	ID         int64
	IdentityID string
}

func accountSecurityError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrStepUpRequired):
		return http.StatusForbidden, accountSecurityStepUpMessage
	case errors.Is(err, domain.ErrEmailTaken), errors.Is(err, domain.ErrInvalidToken):
		return http.StatusBadRequest, accountSecurityEmailTakenMsg
	case errors.Is(err, domain.ErrLastPasskey):
		return http.StatusBadRequest, accountSecurityLastPasskeyMsg
	default:
		return http.StatusBadRequest, accountSecurityFailureMessage
	}
}

func writeAccountSecurityJSONError(response http.ResponseWriter, status int, message string) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(map[string]string{"error": message})
}
