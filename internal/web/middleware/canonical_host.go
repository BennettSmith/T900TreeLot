package middleware

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

// CanonicalHost redirects noncanonical public Host values to the configured
// PUBLIC_BASE_URL while leaving platform health checks reachable on any host.
func CanonicalHost(canonical *url.URL, next http.Handler) http.Handler {
	if canonical == nil || canonical.Host == "" {
		return next
	}
	canonicalHost := strings.ToLower(canonical.Hostname())
	if canonicalHost == "" {
		return next
	}
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if isHealthPath(request.URL.Path) {
			next.ServeHTTP(response, request)
			return
		}
		requestHost := requestHost(request)
		if requestHost == "" || strings.EqualFold(requestHost, canonicalHost) {
			next.ServeHTTP(response, request)
			return
		}

		target := *canonical
		target.Path = request.URL.Path
		target.RawPath = request.URL.RawPath
		target.RawQuery = request.URL.RawQuery
		target.Fragment = ""
		http.Redirect(response, request, target.String(), http.StatusMovedPermanently)
	})
}

func isHealthPath(path string) bool {
	return path == "/health/live" || path == "/health/ready"
}

func requestHost(request *http.Request) string {
	host := request.Host
	if host == "" {
		return ""
	}
	if hostname, _, err := net.SplitHostPort(host); err == nil {
		return strings.ToLower(hostname)
	}
	return strings.ToLower(host)
}
