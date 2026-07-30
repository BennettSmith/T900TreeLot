// Package handlers contains inbound HTTP adapters. It translates requests into
// view rendering without defining application or domain policy.
package handlers

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/troop900/treelot/internal/web/views"
	staticassets "github.com/troop900/treelot/web/static"
)

type Options struct {
	Development bool
	Logger      *slog.Logger
}

type Server struct {
	renderer renderer
	logger   *slog.Logger
}

type renderer interface {
	ComponentGallery(context.Context, io.Writer, views.Gallery) error
	ParityResult(context.Context, io.Writer, views.Gallery) error
}

func New(viewRenderer renderer, options Options) http.Handler {
	logger := options.Logger
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	server := &Server{renderer: viewRenderer, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", server.live)
	mux.Handle("GET /static/", staticassets.Handler())
	if options.Development {
		mux.HandleFunc("GET /_dev/components", server.gallery)
		mux.HandleFunc("POST /_dev/parity", server.parity)
	}
	return browserHeaders(mux)
}

func browserHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'")
		response.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(response, request)
	})
}

func (s *Server) live(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response.WriteHeader(http.StatusOK)
	_, _ = response.Write([]byte("ok\n"))
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
