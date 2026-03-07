package debouncer

import (
	"sync"
	"time"
)

// Debouncer is a simple struct to manage debouncing of events
type Debouncer struct {
	duration time.Duration
	timer    *time.Timer
	callback func()
	mu       sync.Mutex
}

// NewDebouncer creates a new Debouncer with the specified duration and callback
func NewDebouncer(durationMs int, callback func()) *Debouncer {
	return &Debouncer{
		duration: time.Duration(durationMs) * time.Millisecond,
		callback: callback,
	}
}

// Trigger resets the timer and will call the callback after the duration has passed without another trigger
func (d *Debouncer) Trigger() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}

	d.timer = time.AfterFunc(d.duration, d.callback)
}

// Stop stops the debouncer timer if it's running
func (d *Debouncer) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
}
