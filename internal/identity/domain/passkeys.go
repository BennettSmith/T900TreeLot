package domain

import "errors"

// ErrLastPasskey indicates an identity must retain at least one registered passkey.
var ErrLastPasskey = errors.New("identity must retain at least one passkey")

// CanRemovePasskey allows removal only when another passkey would remain.
func CanRemovePasskey(currentCount int) error {
	if currentCount <= 1 {
		return ErrLastPasskey
	}
	return nil
}
