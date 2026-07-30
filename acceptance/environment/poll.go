//go:build acceptance

package environment

import (
	"fmt"
	"time"
)

// Eventually polls condition until it succeeds or the deadline passes.
// Fixed sleeps are not used as the sole synchronization mechanism.
func Eventually(timeout time.Duration, interval time.Duration, condition func() error) error {
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	deadline := time.Now().Add(timeout)
	var last error
	for {
		last = condition()
		if last == nil {
			return nil
		}
		if time.Now().After(deadline) {
			if last == nil {
				return fmt.Errorf("condition not met before deadline")
			}
			return fmt.Errorf("condition not met before deadline: %w", last)
		}
		time.Sleep(interval)
	}
}
