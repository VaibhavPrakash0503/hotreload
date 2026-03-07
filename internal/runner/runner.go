package runner

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Runner struct {
	command string
	args    []string
	cmd     *exec.Cmd
	mu      sync.Mutex
	running bool
}

func NewRunner(execCmd string) *Runner {
	parts := strings.Fields(execCmd)

	if len(parts) == 0 {
		slog.Error("Invalid command: empty string")
		return &Runner{}
	}

	return &Runner{
		command: parts[0],
		args:    parts[1:],
	}
}

func (r *Runner) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.running {
		slog.Warn("Attempted to start process, but it's already running")
		return nil
	}

	// Create command
	r.cmd = exec.CommandContext(ctx, r.command, r.args...)

	// Platform-specific: set up process group
	configureProcAttr(r.cmd)

	// Set up stdout/stderr pipes
	stdout, err := r.cmd.StdoutPipe()
	if err != nil {
		slog.Error("Failed to create stdout pipe", "error", err)
		return err
	}

	stderr, err := r.cmd.StderrPipe()
	if err != nil {
		slog.Error("Failed to create stderr pipe", "error", err)
		return err
	}

	// Start the process (non-blocking!)
	if err := r.cmd.Start(); err != nil {
		slog.Error("Failed to start process", "error", err)
		return err
	}

	r.running = true

	// Stream logs in goroutines
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	// Monitor process in background
	go r.monitor()

	slog.Info("Process started", "pid", r.cmd.Process.Pid)
	return nil
}

func (r *Runner) monitor() {
	if r.cmd == nil || r.cmd.Process == nil {
		return
	}

	// Wait for process to exit
	err := r.cmd.Wait()

	r.mu.Lock()
	r.running = false
	r.mu.Unlock()

	if err != nil {
		slog.Warn("Process exited with error", "error", err)
	} else {
		slog.Info("Process exited cleanly")
	}
}

func (r *Runner) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running || r.cmd == nil || r.cmd.Process == nil {
		return nil // Nothing to stop
	}

	slog.Info("Stopping process", "pid", r.cmd.Process.Pid)

	// Try graceful shutdown first
	if err := r.cmd.Process.Signal(os.Interrupt); err != nil {
		// Process might have already exited
		r.running = false
		return nil
	}

	// Wait up to 5 seconds for graceful exit
	done := make(chan error, 1)
	go func() {
		done <- r.cmd.Wait()
	}()

	select {
	case <-done:
		// Process exited gracefully
		slog.Info("Process stopped gracefully")
		r.running = false
		return nil

	case <-time.After(5 * time.Second):
		// Timeout - force kill
		slog.Warn("Process did not stop gracefully, force killing")
		if err := forceKill(r.cmd); err != nil {
			slog.Error("Failed to force kill process", "error", err)
			return err
		}
		r.running = false
		return nil
	}
}
