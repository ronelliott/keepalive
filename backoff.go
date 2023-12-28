package keepalive

import "time"

// NewExponentialBackoff returns a function that returns an exponentially
// increasing duration. The first call to the function will return the start
// duration. Each subsequent call will double the duration until the max
// duration is reached. The max duration will be returned for all subsequent
// calls to the function after the max duration is reached.
func NewExponentialBackoff(start time.Duration, max time.Duration) func() time.Duration {
	last := start / 2
	return func() time.Duration {
		last *= 2
		if last > max {
			last = max
		}
		return last
	}
}
