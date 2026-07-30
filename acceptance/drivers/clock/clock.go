//go:build acceptance

package clock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Driver controls the acceptance test clock.
type Driver struct {
	baseURL string
	key     string
	client  *http.Client
}

// New constructs a clock driver.
func New(baseURL, key string) *Driver {
	return &Driver{
		baseURL: baseURL,
		key:     key,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Now returns the application clock instant.
func (d *Driver) Now() (time.Time, error) {
	request, err := http.NewRequest(http.MethodGet, d.baseURL+"/_test/clock", nil)
	if err != nil {
		return time.Time{}, err
	}
	request.Header.Set("X-Test-Control-Key", d.key)
	response, err := d.client.Do(request)
	if err != nil {
		return time.Time{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return time.Time{}, fmt.Errorf("get clock status %d: %s", response.StatusCode, body)
	}
	var payload struct {
		Now string `json:"now"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, payload.Now)
}

// Advance moves the application clock forward.
func (d *Driver) Advance(duration time.Duration) (time.Time, error) {
	body, err := json.Marshal(map[string]string{"duration": duration.String()})
	if err != nil {
		return time.Time{}, err
	}
	request, err := http.NewRequest(http.MethodPost, d.baseURL+"/_test/clock/advance", bytes.NewReader(body))
	if err != nil {
		return time.Time{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Test-Control-Key", d.key)
	response, err := d.client.Do(request)
	if err != nil {
		return time.Time{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(response.Body)
		return time.Time{}, fmt.Errorf("advance clock status %d: %s", response.StatusCode, payload)
	}
	var payload struct {
		Now string `json:"now"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, payload.Now)
}
