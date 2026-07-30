//go:build acceptance

package outbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Driver enqueues and inspects outbox rows through acceptance test-control.
type Driver struct {
	baseURL string
	key     string
	client  *http.Client
}

// New constructs an outbox test-control driver.
func New(baseURL, key string) *Driver {
	return &Driver{
		baseURL: baseURL,
		key:     key,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Enqueue creates a pending outbox message.
func (d *Driver) Enqueue(idempotencyKey, channel string) error {
	body, err := json.Marshal(map[string]string{
		"idempotency_key": idempotencyKey,
		"channel":         channel,
	})
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPost, d.baseURL+"/_test/outbox", bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Test-Control-Key", d.key)
	response, err := d.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		payload, _ := io.ReadAll(response.Body)
		return fmt.Errorf("enqueue status %d: %s", response.StatusCode, payload)
	}
	return nil
}

// Status returns the current outbox status for a key.
func (d *Driver) Status(idempotencyKey string) (string, error) {
	endpoint := d.baseURL + "/_test/outbox?idempotency_key=" + url.QueryEscape(idempotencyKey)
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("X-Test-Control-Key", d.key)
	response, err := d.client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(response.Body)
		return "", fmt.Errorf("status %d: %s", response.StatusCode, payload)
	}
	var payload struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	return payload.Status, nil
}
