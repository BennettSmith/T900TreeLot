// Package ids generates opaque application identifiers.
package ids

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

type Generator struct{}

func NewGenerator() Generator {
	return Generator{}
}

func (Generator) NewID() (string, error) {
	buffer := make([]byte, 18)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
