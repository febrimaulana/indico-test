// Package clock abstracts the current time so expiry and sweeper logic can be
// unit tested without waiting for real time to pass.
package clock

import "time"

// Clock provides the current time in UTC.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

// Now returns the current UTC time.
func (realClock) Now() time.Time { return time.Now().UTC() }

// New returns the production clock.
func New() Clock { return realClock{} }
