package application

import (
	"errors"

	"github.com/troop900/treelot/internal/identity/domain"
)

// ErrAccountNotFound indicates that an authenticated session references no active account.
var ErrAccountNotFound = errors.New("account not found")

// AccountProfile is the browser-facing identity summary for authenticated account pages.
type AccountProfile struct {
	IdentityID   string
	DisplayName  string
	PrimaryEmail string
}

type LandingProfile struct {
	IdentityID  string
	DisplayName string
	Roles       []domain.Role
}

func (p LandingProfile) HasRole(role domain.Role) bool {
	for _, assigned := range p.Roles {
		if assigned == role {
			return true
		}
	}
	return false
}
