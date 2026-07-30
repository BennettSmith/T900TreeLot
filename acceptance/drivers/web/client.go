//go:build acceptance

package web

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// Client drives browser-visible HTTP interactions.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient constructs an HTTP driver rooted at baseURL.
func NewClient(baseURL string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

// Get performs a GET and returns status and body.
func (c *Client) Get(path string) (int, string, error) {
	response, err := c.httpClient.Get(c.baseURL + path)
	if err != nil {
		return 0, "", err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return response.StatusCode, "", err
	}
	return response.StatusCode, string(body), nil
}

// GetWithHeaders performs a GET without a cookie jar so Set-Cookie attributes
// remain observable even for Secure cookies over plain HTTP.
func (c *Client) GetWithHeaders(path string) (int, string, http.Header, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	response, err := client.Get(c.baseURL + path)
	if err != nil {
		return 0, "", nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return response.StatusCode, "", response.Header, err
	}
	return response.StatusCode, string(body), response.Header.Clone(), nil
}

// PostForm posts an application/x-www-form-urlencoded body.
func (c *Client) PostForm(path string, values url.Values) (int, string, http.Header, error) {
	response, err := c.httpClient.PostForm(c.baseURL+path, values)
	if err != nil {
		return 0, "", nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return response.StatusCode, "", response.Header, err
	}
	return response.StatusCode, string(body), response.Header.Clone(), nil
}

// CSRFToken extracts the synchronizer token from a home page body.
func CSRFToken(body string) (string, error) {
	const marker = `name="csrf_token" value="`
	start := strings.Index(body, marker)
	if start < 0 {
		return "", fmt.Errorf("csrf token not found")
	}
	start += len(marker)
	end := strings.Index(body[start:], `"`)
	if end < 0 {
		return "", fmt.Errorf("csrf token not terminated")
	}
	return body[start : start+end], nil
}

// SessionCookieSecure reports whether Set-Cookie for the session includes Secure.
func SessionCookieSecure(headers http.Header) (bool, error) {
	for _, value := range headers.Values("Set-Cookie") {
		if strings.Contains(value, "treelot_session=") {
			return strings.Contains(strings.ToLower(value), "secure"), nil
		}
	}
	return false, fmt.Errorf("session cookie not set")
}
