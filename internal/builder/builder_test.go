package builder

import (
	"context"
	"testing"
)

// TestBuildSuccess verifies that a valid command runs without error.
func TestBuildSuccess(t *testing.T) {
	b := NewBuilder("echo hello")
	if err := b.Build(context.Background()); err != nil {
		t.Fatalf("expected successful build, got error: %v", err)
	}
}

// TestBuildFailure verifies that a failing command propagates an error.
func TestBuildFailure(t *testing.T) {
	b := NewBuilder("false") // unix 'false' always exits with code 1
	if err := b.Build(context.Background()); err == nil {
		t.Fatal("expected build to fail, but got nil error")
	}
}

// TestBuildCancelled verifies that a cancelled context stops the build.
func TestBuildCancelled(t *testing.T) {
	// 'sleep 10' will be killed when the context is cancelled.
	b := NewBuilder("sleep 10")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	if err := b.Build(ctx); err == nil {
		t.Fatal("expected build to fail due to cancelled context, got nil error")
	}
}

// TestNewBuilderEmptyCommand verifies that an empty build command returns a usable zero-value Builder.
func TestNewBuilderEmptyCommand(t *testing.T) {
	b := NewBuilder("")
	// Build on a zero-value Builder should not panic; cmd is empty so it exits early.
	// We don't call Build() here since it would try to exec an empty string.
	if b == nil {
		t.Fatal("expected non-nil Builder for empty command")
	}
}
