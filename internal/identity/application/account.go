package application

import "errors"

// ErrAccountNotFound indicates that an authenticated session references no active account.
var ErrAccountNotFound = errors.New("account not found")

// AccountProfile is the browser-facing identity summary for authenticated account pages.
type AccountProfile struct {
	IdentityID   string
	DisplayName  string
	PrimaryEmail string
}
