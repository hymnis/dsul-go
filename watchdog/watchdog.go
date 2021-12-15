// DSUL - Disturb State USB Light : Watchdog module.
package watchdog

import (
	"time"
)

type Watchdog struct {
	interval time.Duration
	timer    *time.Timer
}

func NewCallbackTimer(interval time.Duration, callback func()) *Watchdog {
	w := Watchdog{
		interval: interval,
		timer:    time.AfterFunc(interval, callback),
	}
	return &w
}

func NewChannelTimer(interval time.Duration) *Watchdog {
	w := Watchdog{
		interval: interval,
		timer:    time.NewTimer(interval),
	}
	return &w
}

func (w *Watchdog) Stop() {
	w.timer.Stop()
}

func (w *Watchdog) Kick() {
	w.timer.Stop()
	w.timer.Reset(w.interval)
}

func (w *Watchdog) Channel() <-chan time.Time {
	return w.timer.C
}
