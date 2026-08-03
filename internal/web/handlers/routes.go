// Package handlers contains inbound HTTP adapters. It translates requests into
// view rendering without defining application or domain policy.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/troop900/treelot/internal/identity/application"
	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/web/middleware"
	"github.com/troop900/treelot/internal/web/views"
	staticassets "github.com/troop900/treelot/web/static"
)

// Options configures the HTTP adapter.
type Options struct {
	Development        bool
	Logger             *slog.Logger
	Ready              func(context.Context) error
	Sessions           middleware.Sessions
	SecureCookies      bool
	TestControlEnabled bool
	TestControlKey     string
	Clock              clock.Clock
	ControllableClock  *clock.Controllable
	Outbox             OutboxControl
	Bootstrap          BootstrapService
	Accounts           AccountReader
	BootstrapReset     BootstrapResetControl
}

type BootstrapService interface {
	StartBootstrap(context.Context, application.StartBootstrapCommand) (application.StartBootstrapResult, error)
	ClaimBootstrapProfile(context.Context, application.ClaimBootstrapProfileCommand) (application.PendingEnrollment, error)
	BeginPasskeyRegistration(context.Context, application.BeginPasskeyRegistrationCommand) (application.RegistrationOptions, error)
	FinishBootstrap(context.Context, application.FinishBootstrapCommand) (application.BootstrapResult, error)
}

type AccountReader interface {
	FindAccountProfile(context.Context, string) (application.AccountProfile, error)
}

type BootstrapResetControl interface {
	ResetBootstrap(context.Context) error
}

type Server struct {
	renderer          renderer
	logger            *slog.Logger
	ready             func(context.Context) error
	testControlKey    string
	controllableClock *clock.Controllable
	clock             clock.Clock
	outbox            OutboxControl
	bootstrap         BootstrapService
	accounts          AccountReader
	bootstrapReset    BootstrapResetControl
	secureCookies     bool
}

type renderer interface {
	Home(context.Context, io.Writer, views.Home) error
	Bootstrap(context.Context, io.Writer, views.BootstrapPage) error
	Account(context.Context, io.Writer, views.AccountPage) error
	ComponentGallery(context.Context, io.Writer, views.Gallery) error
	ParityResult(context.Context, io.Writer, views.Gallery) error
}

// New builds the root HTTP handler.
func New(viewRenderer renderer, options Options) http.Handler {
	logger := options.Logger
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	ready := options.Ready
	if ready == nil {
		ready = func(context.Context) error { return nil }
	}
	clk := options.Clock
	if clk == nil {
		clk = clock.System()
	}
	server := &Server{
		renderer:          viewRenderer,
		logger:            logger,
		ready:             ready,
		testControlKey:    options.TestControlKey,
		controllableClock: options.ControllableClock,
		clock:             clk,
		outbox:            options.Outbox,
		bootstrap:         options.Bootstrap,
		accounts:          options.Accounts,
		bootstrapReset:    options.BootstrapReset,
		secureCookies:     options.SecureCookies,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", server.live)
	mux.HandleFunc("GET /health/ready", server.readyHandler)
	mux.Handle("GET /static/", staticassets.Handler())

	browser := http.NewServeMux()
	browser.HandleFunc("GET /{$}", server.home)
	browser.HandleFunc("GET /bootstrap", server.bootstrapEntry)
	browser.HandleFunc("POST /bootstrap/start", server.bootstrapStart)
	browser.HandleFunc("POST /bootstrap/claim", server.bootstrapClaim)
	browser.HandleFunc("POST /bootstrap/passkey/begin", server.bootstrapPasskeyBegin)
	browser.HandleFunc("POST /bootstrap/passkey/finish", server.bootstrapPasskeyFinish)
	browser.HandleFunc("GET /account", server.account)
	browser.HandleFunc("POST /smoke", server.smoke)
	if options.Development {
		browser.HandleFunc("GET /_dev/components", server.gallery)
		browser.HandleFunc("POST /_dev/parity", server.parity)
	}
	if options.TestControlEnabled {
		mux.HandleFunc("POST /_test/clock/advance", server.advanceClock)
		mux.HandleFunc("GET /_test/clock", server.getClock)
		mux.HandleFunc("POST /_test/outbox", server.enqueueOutbox)
		mux.HandleFunc("GET /_test/outbox", server.getOutbox)
		mux.HandleFunc("POST /_test/bootstrap/reset", server.resetBootstrap)
	}

	var browserHandler http.Handler = browser
	if options.Sessions != nil {
		browserHandler = middleware.SessionCSRF(options.Sessions, options.SecureCookies, browser)
	}
	mux.Handle("/", browserHandler)

	return middleware.BrowserHeaders(mux)
}

func (s *Server) live(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response.WriteHeader(http.StatusOK)
	_, _ = response.Write([]byte("ok\n"))
}

func (s *Server) readyHandler(response http.ResponseWriter, request *http.Request) {
	if err := s.ready(request.Context()); err != nil {
		s.logger.ErrorContext(request.Context(), "readiness check failed", "error", err)
		http.Error(response, "not ready\n", http.StatusServiceUnavailable)
		return
	}
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response.WriteHeader(http.StatusOK)
	_, _ = response.Write([]byte("ready\n"))
}

func (s *Server) home(response http.ResponseWriter, request *http.Request) {
	data := views.Home{
		PageTitle:  "Troop 900 Tree Lot",
		Brand:      "Troop 900 Tree Lot",
		Headline:   "Volunteer shift scheduling",
		Supporting: "Families coordinate tree-lot shifts for the troop season from this secure site.",
		Navigation: []views.Link{
			{Label: "Home", Href: "/", Current: true},
		},
		SmokeInput: request.URL.Query().Get("message"),
	}
	if data.SmokeInput != "" {
		data.SmokeMessage = data.SmokeInput
	}
	if current := middleware.FromContext(request.Context()); current != nil {
		data.CSRFToken = current.CSRFToken
	}
	s.renderHTML(response, request, func(output io.Writer) error {
		return s.renderer.Home(request.Context(), output, data)
	})
}

func (s *Server) smoke(response http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		http.Error(response, "Invalid form submission.", http.StatusBadRequest)
		return
	}
	message := strings.TrimSpace(request.FormValue("message"))
	if message == "" {
		message = "Foundation smoke confirmed"
	}
	target := "/?message=" + url.QueryEscape(message)
	http.Redirect(response, request, target, http.StatusSeeOther)
}

func (s *Server) gallery(response http.ResponseWriter, request *http.Request) {
	s.renderHTML(response, request, func(output io.Writer) error {
		return s.renderer.ComponentGallery(request.Context(), output, views.GalleryData())
	})
}

func (s *Server) parity(response http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		http.Error(response, "Invalid form submission.", http.StatusBadRequest)
		return
	}
	data := views.GalleryData()
	data.ParityMessage = request.FormValue("message")
	if data.ParityMessage == "" {
		data.ParityMessage = "Signal parity confirmed"
	}
	data.ParityFragment = request.Header.Get("HX-Request") == "true"
	s.renderHTML(response, request, func(output io.Writer) error {
		return s.renderer.ParityResult(request.Context(), output, data)
	})
}

func (s *Server) advanceClock(response http.ResponseWriter, request *http.Request) {
	if !s.authorizeTestControl(request) {
		http.Error(response, "forbidden", http.StatusForbidden)
		return
	}
	if s.controllableClock == nil {
		http.Error(response, "clock control unavailable", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		Duration string `json:"duration"`
	}
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		http.Error(response, "invalid body", http.StatusBadRequest)
		return
	}
	duration, err := time.ParseDuration(body.Duration)
	if err != nil || duration < 0 {
		http.Error(response, "invalid duration", http.StatusBadRequest)
		return
	}
	s.controllableClock.Advance(duration)
	s.writeClock(response)
}

func (s *Server) getClock(response http.ResponseWriter, request *http.Request) {
	if !s.authorizeTestControl(request) {
		http.Error(response, "forbidden", http.StatusForbidden)
		return
	}
	s.writeClock(response)
}

func (s *Server) authorizeTestControl(request *http.Request) bool {
	return s.testControlKey != "" && request.Header.Get("X-Test-Control-Key") == s.testControlKey
}

func (s *Server) writeClock(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(map[string]string{
		"now": s.clock.Now().UTC().Format(time.RFC3339Nano),
	})
}

func (s *Server) renderHTML(response http.ResponseWriter, request *http.Request, render func(io.Writer) error) {
	var output bytes.Buffer
	if err := render(&output); err != nil {
		s.renderError(response, request, err)
		return
	}
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	response.WriteHeader(http.StatusOK)
	_, _ = output.WriteTo(response)
}

func (s *Server) renderError(response http.ResponseWriter, request *http.Request, err error) {
	s.logger.ErrorContext(request.Context(), "render response", "method", request.Method, "path", request.URL.Path, "error", err)
	http.Error(response, "Unable to render this page.", http.StatusInternalServerError)
}

// ErrNotReady is available for readiness adapters.
var ErrNotReady = errors.New("not ready")
