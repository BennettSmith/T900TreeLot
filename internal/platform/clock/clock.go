// Package clock provides an injectable clock for time-dependent behavior.
package clock

import "time"

// Clock reports the current instant in UTC.
type Clock interface {
	Now() time.Time
}

type systemClock struct{}

// System returns a clock backed by the host clock, always in UTC.
func System() Clock {
	return systemClock{}
}

func (systemClock) Now() time.Time {
	return time.Now().UTC()
}

// Controllable is a mutable clock for acceptance and focused tests.
type Controllable struct {
	now time.Time
}

// NewControllable starts a controllable clock at the given instant.
func NewControllable(now time.Time) *Controllable {
	return &Controllable{now: now.UTC()}
}

// Now returns the controlled instant in UTC.
func (c *Controllable) Now() time.Time {
	return c.now
}

// Advance moves the controlled clock forward by duration.
func (c *Controllable) Advance(duration time.Duration) {
	c.now = c.now.Add(duration)
}

// Set replaces the controlled instant.
func (c *Controllable) Set(now time.Time) {
	c.now = now.UTC()
}
