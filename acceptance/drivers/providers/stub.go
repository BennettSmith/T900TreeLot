//go:build acceptance

package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Stub drives the Groups.io provider stub.
type Stub struct {
	BaseURL string
	client  *http.Client
}

// New constructs a provider stub driver.
func New(baseURL string) *Stub {
	return &Stub{
		BaseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Available reports whether the stub health endpoint responds.
func (s *Stub) Available() error {
	response, err := s.client.Get(s.BaseURL + "/health/live")
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("stub health status %d", response.StatusCode)
	}
	return nil
}

// PostGroupMessage sends a Groups.io-shaped authorized message.
func (s *Stub) PostGroupMessage(group, subject string) error {
	body, err := json.Marshal(map[string]string{"subject": subject, "body": subject})
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPost, s.BaseURL+"/api/v1/groups/"+group+"/messages", bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer acceptance-stub-token")
	request.Header.Set("Content-Type", "application/json")
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		payload, _ := io.ReadAll(response.Body)
		return fmt.Errorf("post message status %d: %s", response.StatusCode, payload)
	}
	return nil
}

// MessageCount returns how many stubbed messages were recorded.
func (s *Stub) MessageCount() (int, error) {
	response, err := s.client.Get(s.BaseURL + "/_stub/groupsio/messages")
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("list status %d", response.StatusCode)
	}
	var messages []map[string]any
	if err := json.NewDecoder(response.Body).Decode(&messages); err != nil {
		return 0, err
	}
	return len(messages), nil
}
