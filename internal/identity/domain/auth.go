package domain

import "errors"

var ErrInvalidRole = errors.New("invalid identity role")

const (
	RoleCommittee       Role = "committee"
	RoleFamilyManager   Role = "family_manager"
	RoleYoungAdultScout Role = "young_adult_scout"
)

// ParseRole accepts only roles persisted by Identity and Access.
func ParseRole(value string) (Role, error) {
	role := Role(value)
	switch role {
	case RoleAdmin, RoleCommittee, RoleFamilyManager, RoleYoungAdultScout:
		return role, nil
	default:
		return "", ErrInvalidRole
	}
}
