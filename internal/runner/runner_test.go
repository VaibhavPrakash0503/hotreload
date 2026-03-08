package runner

import (
	"context"
	"testing"
	"time"
)

// TestStartSetsRunning verifies that Start transitions running to true.
func TestStartSetsRunning(t *testing.T) {
	r := NewRunner("sleep 10")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := r.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer r.Stop() //nolint:errcheck

	r.mu.Lock()
	running := r.running
	r.mu.Unlock()

	if !running {
		t.Error("expected running=true after Start")
	}
}

// TestDoubleStartIsNoop verifies that calling Start twice is safe and idempotent.
func TestDoubleStartIsNoop(t *testing.T) {
	r := NewRunner("sleep 10")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := r.Start(ctx); err != nil {
		t.Fatalf("first Start failed: %v", err)
	}
	defer r.Stop() //nolint:errcheck

	// Second Start should be a no-op, not an error.
	if err := r.Start(ctx); err != nil {
		t.Fatalf("second Start returned unexpected error: %v", err)
	}
}

// TestStopSetsNotRunning verifies that Stop transitions running to false.
func TestStopSetsNotRunning(t *testing.T) {
	r := NewRunner("sleep 10")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := r.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if err := r.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	r.mu.Lock()
	running := r.running
	r.mu.Unlock()

	if running {
		t.Error("expected running=false after Stop")
	}
}

// TestStopOnIdleRunnerIsNoop verifies that Stop on a never-started runner is safe.
func TestStopOnIdleRunnerIsNoop(t *testing.T) {
	r := NewRunner("sleep 10")
	if err := r.Stop(); err != nil {
		t.Fatalf("Stop on idle runner returned error: %v", err)
	}
}

// TestCrashCountIncrementsOnFastCrash verifies that a process that exits
// immediately increments the crash counter.
func TestCrashCountIncrementsOnFastCrash(t *testing.T) {
	// 'false' exits immediately with code 1.
	r := NewRunner("false")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := r.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give monitor() time to call cmd.Wait() and update crashCount.
	time.Sleep(300 * time.Millisecond)

	r.mu.Lock()
	crashes := r.crashCount
	r.mu.Unlock()

	if crashes != 1 {
		t.Errorf("expected crashCount=1 after instant crash, got %d", crashes)
	}
}

// TestCrashCountResetsOnHealthyExit verifies that a process which lives longer
// than 2 seconds causes crashCount to reset to 0 when it exits.
func TestCrashCountResetsOnHealthyExit(t *testing.T) {
	// Start with a non-zero crash count to confirm it gets cleared.
	r := NewRunner("sleep 3")
	r.crashCount = 2

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := r.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait past the 2s fast-crash threshold while the process is still alive.
	time.Sleep(2100 * time.Millisecond)

	// Stop the process — monitor() will see uptime > 2s and reset crashCount.
	if err := r.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Give monitor() a moment to finish the lock-and-reset.
	time.Sleep(100 * time.Millisecond)

	r.mu.Lock()
	crashes := r.crashCount
	r.mu.Unlock()

	if crashes != 0 {
		t.Errorf("expected crashCount=0 after healthy exit, got %d", crashes)
	}
}

// TestBackoffRespectsContextCancellation verifies that a pending backoff
// is aborted when the context is cancelled.
func TestBackoffRespectsContextCancellation(t *testing.T) {
	r := NewRunner("echo ok")
	r.crashCount = 1 // will trigger 2s backoff

	ctx, cancel := context.WithCancel(context.Background())

	start := time.Now()
	cancel() // cancel immediately so backoff select hits ctx.Done()

	err := r.Start(ctx)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected error due to context cancellation, got nil")
	}
	if elapsed > time.Second {
		t.Errorf("backoff did not respect context cancellation (took %v)", elapsed)
	}
}
