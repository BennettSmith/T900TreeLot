// Package application contains Families use-case ports and commands.
package application

import (
	"context"
	"time"
)

type PersonalProfile struct {
	ID                   string
	FirstName            string
	LastName             string
	PreferredDisplayName string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type ProfileCreator interface {
	CreatePersonalProfile(context.Context, PersonalProfile) error
}
