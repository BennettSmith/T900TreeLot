// Package middleware provides HTTP middleware for browser security foundations.
package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/troop900/treelot/internal/platform/session"
)

const (
	// SessionCookieName is the host-only session cookie.
	SessionCookieName = "treelot_session"
	// CSRFFormField is the form field carrying the synchronizer token.
	CSRFFormField = "csrf_token"
	// CSRFHeader is an alternate header for HTMX/AJAX posts.
	CSRFHeader = "X-CSRF-Token"
)

type contextKey struct{}

// FromContext returns the request session when present.
func FromContext(ctx context.Context) *session.Session {
	value, _ := ctx.Value(contextKey{}).(*session.Session)
	return value
}

// Sessions creates and loads browser sessions.
type Sessions interface {
	Create(ctx context.Context) (session.Session, string, error)
	Get(ctx context.Context, rawToken string) (session.Session, error)
}

// SessionCSRF loads or creates a browser session and enforces CSRF on unsafe methods.
func SessionCSRF(store Sessions, secureCookies bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		current, rawToken, err := loadOrCreate(request.Context(), store, request)
		if err != nil {
			http.Error(response, "Unable to establish a secure session.", http.StatusInternalServerError)
			return
		}
		if rawToken != "" {
			http.SetCookie(response, &http.Cookie{
				Name:     SessionCookieName,
				Value:    rawToken,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   secureCookies,
				Expires:  current.ExpiresAt,
			})
		}

		if isUnsafe(request.Method) && !csrfValid(request, current) {
			http.Error(response, "CSRF token missing or invalid.", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(request.Context(), contextKey{}, current)
		next.ServeHTTP(response, request.WithContext(ctx))
	})
}

func loadOrCreate(ctx context.Context, store Sessions, request *http.Request) (*session.Session, string, error) {
	cookie, err := request.Cookie(SessionCookieName)
	if err == nil {
		loaded, getErr := store.Get(ctx, cookie.Value)
		if getErr == nil {
			return &loaded, "", nil
		}
		if !errors.Is(getErr, session.ErrNotFound) {
			return nil, "", getErr
		}
	}
	created, rawToken, createErr := store.Create(ctx)
	if createErr != nil {
		return nil, "", createErr
	}
	return &created, rawToken, nil
}

func isUnsafe(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func csrfValid(request *http.Request, current *session.Session) bool {
	token := request.Header.Get(CSRFHeader)
	if token == "" {
		_ = request.ParseForm()
		token = request.FormValue(CSRFFormField)
	}
	return token != "" && token == current.CSRFToken
}

// BrowserHeaders applies baseline browser security headers.
func BrowserHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'")
		response.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(response, request)
	})
}

// RequestTimeout is retained for future use by entry points coordinating middleware.
var RequestTimeout = 30 * time.Second
