// Package domain contains Identity and Access business rules.
package domain

import (
	"errors"
	"strings"
)

var (
	ErrBootstrapClosed = errors.New("bootstrap is closed")
	ErrInvalidToken    = errors.New("invalid bootstrap token")
	ErrRateLimited     = errors.New("bootstrap rate limited")
	ErrEmailTaken      = errors.New("email is already in use")
	ErrCeremonyFailed  = errors.New("passkey ceremony failed")
	ErrInvalidProfile  = errors.New("invalid profile")
)

// Role identifies an application authority granted to an identity.
type Role string

const (
	RoleAdmin Role = "admin"
)

// Email preserves the claimed email while exposing a normalized identifier.
type Email struct {
	value      string
	normalized string
}

// NewEmail validates and normalizes a claimed account email.
func NewEmail(value string) (Email, error) {
	trimmed := strings.TrimSpace(value)
	normalized := NormalizeEmail(trimmed)
	if trimmed == "" || !strings.Contains(normalized, "@") {
		return Email{}, ErrInvalidToken
	}
	return Email{value: trimmed, normalized: normalized}, nil
}

func (e Email) String() string {
	return e.value
}

func (e Email) Normalized() string {
	return e.normalized
}

// NormalizeEmail returns the comparison form for active account uniqueness.
func NormalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// ProfileName is the personal profile name captured during bootstrap.
type ProfileName struct {
	FirstName            string
	LastName             string
	PreferredDisplayName string
}

// ValidateProfile trims profile fields and requires legal first and last names.
func ValidateProfile(first, last, preferred string) (ProfileName, error) {
	name := ProfileName{
		FirstName:            strings.TrimSpace(first),
		LastName:             strings.TrimSpace(last),
		PreferredDisplayName: strings.TrimSpace(preferred),
	}
	if name.FirstName == "" || name.LastName == "" {
		return ProfileName{}, ErrInvalidProfile
	}
	return name, nil
}

// DisplayName returns the preferred display name when present.
func (p ProfileName) DisplayName() string {
	if p.PreferredDisplayName != "" {
		return p.PreferredDisplayName
	}
	return strings.TrimSpace(p.FirstName + " " + p.LastName)
}

// CanBootstrap enforces one-time first-Admin bootstrap.
func CanBootstrap(adminExists, closed bool) error {
	if adminExists || closed {
		return ErrBootstrapClosed
	}
	return nil
}
