//go:build acceptance

package providers

// Stub is a placeholder for Groups.io stub controls used by later increments.
type Stub struct {
	BaseURL string
}

// New constructs a provider stub driver.
func New(baseURL string) *Stub {
	return &Stub{BaseURL: baseURL}
}

// Available reports whether the stub endpoint is configured.
func (s *Stub) Available() bool {
	return s.BaseURL != ""
}
