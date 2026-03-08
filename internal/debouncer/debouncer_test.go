package debouncer

import (
	"sync/atomic"
	"testing"
	"time"
)

// TestSingleTrigger verifies that a single Trigger call fires the callback exactly once.
func TestSingleTrigger(t *testing.T) {
	var count atomic.Int32

	d := NewDebouncer(50, func() { count.Add(1) })
	defer d.Stop()

	d.Trigger()
	time.Sleep(150 * time.Millisecond)

	if got := count.Load(); got != 1 {
		t.Errorf("expected callback to fire 1 time, got %d", got)
	}
}

// TestRapidTriggersDebounce verifies that many rapid triggers result in only one callback.
func TestRapidTriggersDebounce(t *testing.T) {
	var count atomic.Int32

	d := NewDebouncer(100, func() { count.Add(1) })
	defer d.Stop()

	for i := 0; i < 20; i++ {
		d.Trigger()
		time.Sleep(10 * time.Millisecond)
	}

	// Wait well past the debounce window.
	time.Sleep(300 * time.Millisecond)

	if got := count.Load(); got != 1 {
		t.Errorf("expected exactly 1 callback after rapid triggers, got %d", got)
	}
}

// TestStopPreventsCallback verifies that Stop cancels a pending callback.
func TestStopPreventsCallback(t *testing.T) {
	var count atomic.Int32

	d := NewDebouncer(200, func() { count.Add(1) })

	d.Trigger()
	d.Stop() // cancel before the 200ms window elapses

	time.Sleep(400 * time.Millisecond)

	if got := count.Load(); got != 0 {
		t.Errorf("expected 0 callbacks after Stop, got %d", got)
	}
}

// TestMultipleWindowsFire verifies that triggers separated by more than the
// debounce window each fire their own callback.
func TestMultipleWindowsFire(t *testing.T) {
	var count atomic.Int32

	d := NewDebouncer(50, func() { count.Add(1) })
	defer d.Stop()

	d.Trigger()
	time.Sleep(150 * time.Millisecond) // let first window expire

	d.Trigger()
	time.Sleep(150 * time.Millisecond) // let second window expire

	if got := count.Load(); got != 2 {
		t.Errorf("expected 2 callbacks for 2 separated triggers, got %d", got)
	}
}
