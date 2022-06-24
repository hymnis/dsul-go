// DSUL - Disturb State USB Light : Watchdog module
package watchdog

import (
	"time"
)

// Watchdog holds a timer and an interval.
type Watchdog struct {
	interval time.Duration
	timer    *time.Timer
}

// NewCallbackTimer creates a new watchdog that calls a callback function when the timer expires.
func NewCallbackTimer(interval time.Duration, callback func()) *Watchdog {
	w := Watchdog{
		interval: interval,
		timer:    time.AfterFunc(interval, callback),
	}
	return &w
}

// NewChannelTimer creates a new watchdog.
func NewChannelTimer(interval time.Duration) *Watchdog {
	w := Watchdog{
		interval: interval,
		timer:    time.NewTimer(interval),
	}
	return &w
}

// Stop stops the watchdog timer.
func (w *Watchdog) Stop() {
	w.timer.Stop()
}

// Kick resets the watchdog timer.
func (w *Watchdog) Kick() {
	w.timer.Stop()
	w.timer.Reset(w.interval)
}

// Channel returns the channel that the watchdog timer sends on.
func (w *Watchdog) Channel() <-chan time.Time {
	return w.timer.C
}
