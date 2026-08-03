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
	bootstrapUnavailableMessage = "Bootstrap enrollment is unavailable. Check the token and try again."
	bootstrapRateLimitMessage   = "Too many attempts. Wait a few minutes and try again."
	bootstrapProfileMessage     = "Enter a valid email, first name, and last name."
	bootstrapCeremonyMessage    = "Passkey registration could not be completed. Try again."
)

func (s *Server) bootstrapEntry(response http.ResponseWriter, request *http.Request) {
	s.renderBootstrap(response, request, s.bootstrapPage(views.BootstrapStageEntry))
}

func (s *Server) bootstrapStart(response http.ResponseWriter, request *http.Request) {
	if s.bootstrap == nil {
		http.Error(response, "Bootstrap enrollment is unavailable.", http.StatusServiceUnavailable)
		return
	}
	if err := request.ParseForm(); err != nil {
		http.Error(response, "Invalid form submission.", http.StatusBadRequest)
		return
	}
	token := strings.TrimSpace(request.FormValue("bootstrap_token"))
	_, err := s.bootstrap.StartBootstrap(request.Context(), application.StartBootstrapCommand{
		Token:        token,
		RateLimitKey: bootstrapRateLimitKey(request),
	})
	if err != nil {
		page := s.bootstrapPage(views.BootstrapStageEntry)
		page.Token = token
		page.Fields = bootstrapEntryFields(token)
		page.Alert = bootstrapAlert(err)
		s.renderBootstrap(response, request, page)
		return
	}
	page := s.bootstrapPage(views.BootstrapStageEnroll)
	page.Token = token
	page.Fields = bootstrapProfileFields("", "", "", "")
	s.renderBootstrap(response, request, page)
}

func (s *Server) bootstrapClaim(response http.ResponseWriter, request *http.Request) {
	if s.bootstrap == nil {
		http.Error(response, "Bootstrap enrollment is unavailable.", http.StatusServiceUnavailable)
		return
	}
	if err := request.ParseForm(); err != nil {
		http.Error(response, "Invalid form submission.", http.StatusBadRequest)
		return
	}
	payload := bootstrapPayloadFromForm(request)
	pending, err := s.bootstrap.ClaimBootstrapProfile(request.Context(), application.ClaimBootstrapProfileCommand{
		Token:                payload.Token,
		RateLimitKey:         bootstrapRateLimitKey(request),
		Email:                payload.Email,
		FirstName:            payload.FirstName,
		LastName:             payload.LastName,
		PreferredDisplayName: payload.PreferredDisplayName,
	})
	if err != nil {
		page := s.bootstrapPage(views.BootstrapStageEnroll)
		page.Token = payload.Token
		page.Email = payload.Email
		page.FirstName = payload.FirstName
		page.LastName = payload.LastName
		page.PreferredDisplayName = payload.PreferredDisplayName
		page.Fields = bootstrapProfileFields(payload.Email, payload.FirstName, payload.LastName, payload.PreferredDisplayName)
		page.Alert = bootstrapAlert(err)
		s.renderBootstrap(response, request, page)
		return
	}
	s.renderBootstrap(response, request, s.passkeyPage(pending))
}

func (s *Server) bootstrapPasskeyBegin(response http.ResponseWriter, request *http.Request) {
	if s.bootstrap == nil {
		writeBootstrapJSONError(response, http.StatusServiceUnavailable, bootstrapUnavailableMessage)
		return
	}
	payload, err := decodeBootstrapPayload(request)
	if err != nil {
		writeBootstrapJSONError(response, http.StatusBadRequest, "Invalid request.")
		return
	}
	current := middleware.FromContext(request.Context())
	if current == nil {
		writeBootstrapJSONError(response, http.StatusUnauthorized, "A browser session is required.")
		return
	}
	pending, err := s.bootstrap.ClaimBootstrapProfile(request.Context(), application.ClaimBootstrapProfileCommand{
		Token:                payload.Token,
		RateLimitKey:         bootstrapRateLimitKey(request),
		Email:                payload.Email,
		FirstName:            payload.FirstName,
		LastName:             payload.LastName,
		PreferredDisplayName: payload.PreferredDisplayName,
	})
	if err != nil {
		writeBootstrapJSONError(response, bootstrapErrorStatus(err), bootstrapMessage(err))
		return
	}
	options, err := s.bootstrap.BeginPasskeyRegistration(request.Context(), application.BeginPasskeyRegistrationCommand{
		Pending:   pending,
		SessionID: current.ID,
	})
	if err != nil {
		writeBootstrapJSONError(response, bootstrapErrorStatus(err), bootstrapMessage(err))
		return
	}
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(map[string]any{
		"ceremonyId": options.CeremonyID,
		"publicKey":  options.PublicKey,
	})
}

func (s *Server) bootstrapPasskeyFinish(response http.ResponseWriter, request *http.Request) {
	if s.bootstrap == nil {
		writeBootstrapJSONError(response, http.StatusServiceUnavailable, bootstrapUnavailableMessage)
		return
	}
	var payload bootstrapFinishPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		writeBootstrapJSONError(response, http.StatusBadRequest, "Invalid request.")
		return
	}
	current := middleware.FromContext(request.Context())
	if current == nil {
		writeBootstrapJSONError(response, http.StatusUnauthorized, "A browser session is required.")
		return
	}
	credential, err := json.Marshal(payload.Credential)
	if err != nil || len(payload.Credential) == 0 {
		writeBootstrapJSONError(response, http.StatusBadRequest, bootstrapCeremonyMessage)
		return
	}
	result, err := s.bootstrap.FinishBootstrap(request.Context(), application.FinishBootstrapCommand{
		Token:                payload.TokenValue(),
		RateLimitKey:         bootstrapRateLimitKey(request),
		SessionID:            current.ID,
		Email:                payload.Email,
		FirstName:            payload.FirstName,
		LastName:             payload.LastName,
		PreferredDisplayName: payload.PreferredDisplayName,
		PasskeyCeremonyID:    payload.CeremonyIDValue(),
		PasskeyResponse:      credential,
	})
	if err != nil {
		writeBootstrapJSONError(response, bootstrapErrorStatus(err), bootstrapMessage(err))
		return
	}
	middleware.SetSessionCookie(response, result.Session.RawToken, result.Session.ExpiresAt, s.secureCookies)
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(response).Encode(map[string]string{"redirectTo": "/account"})
}

func (s *Server) account(response http.ResponseWriter, request *http.Request) {
	current := middleware.FromContext(request.Context())
	if current == nil || current.IdentityID == "" {
		http.Redirect(response, request, "/", http.StatusSeeOther)
		return
	}
	if s.accounts == nil {
		http.Error(response, "Account lookup is unavailable.", http.StatusServiceUnavailable)
		return
	}
	profile, err := s.accounts.FindAccountProfile(request.Context(), current.IdentityID)
	if errors.Is(err, application.ErrAccountNotFound) {
		http.Redirect(response, request, "/", http.StatusSeeOther)
		return
	}
	if err != nil {
		s.renderError(response, request, err)
		return
	}
	data := views.AccountPage{
		PageTitle:    "Admin account",
		Brand:        "Troop 900 Tree Lot",
		DisplayName:  profile.DisplayName,
		PrimaryEmail: profile.PrimaryEmail,
		Navigation: []views.Link{
			{Label: "Home", Href: "/"},
			{Label: "Account", Href: "/account", Current: true},
		},
	}
	s.renderHTML(response, request, func(output io.Writer) error {
		return s.renderer.Account(request.Context(), output, data)
	})
}

func (s *Server) resetBootstrap(response http.ResponseWriter, request *http.Request) {
	if !s.authorizeTestControl(request) {
		http.Error(response, "forbidden", http.StatusForbidden)
		return
	}
	if s.bootstrapReset == nil {
		http.Error(response, "bootstrap reset unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := s.bootstrapReset.ResetBootstrap(request.Context()); err != nil {
		http.Error(response, "unable to reset bootstrap", http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(map[string]string{"status": "reset"})
}

func (s *Server) renderBootstrap(response http.ResponseWriter, request *http.Request, page views.BootstrapPage) {
	if current := middleware.FromContext(request.Context()); current != nil && page.CSRFToken == "" {
		page.CSRFToken = current.CSRFToken
	}
	s.renderHTML(response, request, func(output io.Writer) error {
		return s.renderer.Bootstrap(request.Context(), output, page)
	})
}

func (s *Server) bootstrapPage(stage views.BootstrapStage) views.BootstrapPage {
	page := views.BootstrapPage{
		PageTitle: "Bootstrap first Admin",
		Brand:     "Troop 900 Tree Lot",
		Stage:     stage,
		Navigation: []views.Link{
			{Label: "Home", Href: "/"},
			{Label: "Bootstrap", Href: "/bootstrap", Current: true},
		},
	}
	if stage == views.BootstrapStageEntry {
		page.Fields = bootstrapEntryFields("")
	}
	return page
}

func (s *Server) passkeyPage(pending application.PendingEnrollment) views.BootstrapPage {
	return views.BootstrapPage{
		PageTitle:            "Register Admin passkey",
		Brand:                "Troop 900 Tree Lot",
		Stage:                views.BootstrapStagePasskey,
		Navigation:           []views.Link{{Label: "Home", Href: "/"}, {Label: "Bootstrap", Href: "/bootstrap", Current: true}},
		Token:                pending.Token,
		Email:                pending.Email.String(),
		FirstName:            pending.Name.FirstName,
		LastName:             pending.Name.LastName,
		PreferredDisplayName: pending.Name.DisplayName(),
	}
}

func bootstrapEntryFields(token string) []views.Field {
	return []views.Field{{
		ID:           "bootstrap-token",
		Name:         "bootstrap_token",
		Label:        "Bootstrap token",
		Type:         "text",
		Value:        token,
		Hint:         "Paste the deployment bootstrap token. Links may provide it after #token=.",
		Autocomplete: "off",
		Required:     true,
	}}
}

func bootstrapProfileFields(email, first, last, preferred string) []views.Field {
	return []views.Field{
		{ID: "bootstrap-email", Name: "email", Label: "Email address", Type: "email", Value: email, Autocomplete: "email", Required: true},
		{ID: "bootstrap-first-name", Name: "first_name", Label: "First name", Type: "text", Value: first, Autocomplete: "given-name", Required: true},
		{ID: "bootstrap-last-name", Name: "last_name", Label: "Last name", Type: "text", Value: last, Autocomplete: "family-name", Required: true},
		{ID: "bootstrap-display-name", Name: "preferred_display_name", Label: "Preferred display name", Type: "text", Value: preferred, Autocomplete: "nickname"},
	}
}

func bootstrapAlert(err error) *views.Alert {
	return &views.Alert{
		Title:    bootstrapMessage(err),
		Variant:  bootstrapErrorVariant(err),
		Blocking: true,
		Icon:     views.Icon{Name: "critical"},
	}
}

func bootstrapMessage(err error) string {
	switch {
	case errors.Is(err, domain.ErrRateLimited):
		return bootstrapRateLimitMessage
	case errors.Is(err, domain.ErrInvalidProfile):
		return bootstrapProfileMessage
	case errors.Is(err, domain.ErrCeremonyFailed):
		return bootstrapCeremonyMessage
	case errors.Is(err, domain.ErrInvalidToken), errors.Is(err, domain.ErrBootstrapClosed), errors.Is(err, domain.ErrEmailTaken):
		return bootstrapUnavailableMessage
	default:
		return "Unable to complete bootstrap enrollment. Try again."
	}
}

func bootstrapErrorVariant(err error) views.Variant {
	if errors.Is(err, domain.ErrRateLimited) {
		return views.VariantWarning
	}
	return views.VariantCritical
}

func bootstrapErrorStatus(err error) int {
	if errors.Is(err, domain.ErrRateLimited) {
		return http.StatusTooManyRequests
	}
	return http.StatusBadRequest
}

func writeBootstrapJSONError(response http.ResponseWriter, status int, message string) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(map[string]string{"error": message})
}

func bootstrapRateLimitKey(request *http.Request) string {
	host := request.RemoteAddr
	if parsedHost, _, err := net.SplitHostPort(request.RemoteAddr); err == nil {
		host = parsedHost
	}
	host = strings.TrimSpace(host)
	if host == "" {
		host = "unknown"
	}
	return "bootstrap:" + host
}

type bootstrapPayload struct {
	Token                string `json:"token"`
	BootstrapToken       string `json:"bootstrap_token"`
	Email                string `json:"email"`
	FirstName            string `json:"first_name"`
	LastName             string `json:"last_name"`
	PreferredDisplayName string `json:"preferred_display_name"`
}

func (p bootstrapPayload) TokenValue() string {
	if p.Token != "" {
		return p.Token
	}
	return p.BootstrapToken
}

type bootstrapFinishPayload struct {
	bootstrapPayload
	CeremonyID  string          `json:"ceremonyId"`
	CeremonyID2 string          `json:"ceremony_id"`
	Credential  json.RawMessage `json:"credential"`
}

func (p bootstrapFinishPayload) CeremonyIDValue() string {
	if p.CeremonyID != "" {
		return p.CeremonyID
	}
	return p.CeremonyID2
}

func decodeBootstrapPayload(request *http.Request) (bootstrapPayload, error) {
	if strings.HasPrefix(strings.ToLower(request.Header.Get("Content-Type")), "application/json") {
		var payload bootstrapPayload
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			return bootstrapPayload{}, err
		}
		payload.Token = strings.TrimSpace(payload.TokenValue())
		return payload, nil
	}
	if err := request.ParseForm(); err != nil {
		return bootstrapPayload{}, err
	}
	return bootstrapPayloadFromForm(request), nil
}

func bootstrapPayloadFromForm(request *http.Request) bootstrapPayload {
	return bootstrapPayload{
		Token:                strings.TrimSpace(firstFormValue(request, "token", "bootstrap_token")),
		Email:                strings.TrimSpace(request.FormValue("email")),
		FirstName:            strings.TrimSpace(request.FormValue("first_name")),
		LastName:             strings.TrimSpace(request.FormValue("last_name")),
		PreferredDisplayName: strings.TrimSpace(request.FormValue("preferred_display_name")),
	}
}

func firstFormValue(request *http.Request, names ...string) string {
	for _, name := range names {
		if value := request.FormValue(name); value != "" {
			return value
		}
	}
	return ""
}
