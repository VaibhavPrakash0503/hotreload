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
	command    string
	args       []string
	cmd        *exec.Cmd
	mu         sync.Mutex
	running    bool
	startTime  time.Time // when the process last started
	crashCount int       // consecutive fast crashes (<2s uptime)
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

	if r.running {
		slog.Warn("Attempted to start process, but it's already running")
		r.mu.Unlock()
		return nil
	}

	// Exponential backoff for consecutive fast crashes
	if r.crashCount > 0 {
		backoff := time.Duration(min(1<<r.crashCount, 30)) * time.Second
		slog.Warn("Backing off before restart", "delay", backoff, "consecutive_crashes", r.crashCount)
		r.mu.Unlock()

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return ctx.Err()
		}

		r.mu.Lock()
	}

	defer r.mu.Unlock()

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
	r.startTime = time.Now()

	// Stream logs in goroutines
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	// Monitor process in background
	cmd := r.cmd
	go r.monitor(cmd)

	slog.Info("Process started", "pid", r.cmd.Process.Pid)
	return nil
}

func (r *Runner) monitor(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	// Wait for process to exit
	err := cmd.Wait()

	r.mu.Lock()
	uptime := time.Since(r.startTime)
	if uptime < 2*time.Second {
		r.crashCount++
		slog.Warn("Process crashed too quickly",
			"uptime", uptime.Round(time.Millisecond),
			"consecutive_crashes", r.crashCount,
		)
	} else {
		r.crashCount = 0
	}
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

	if !r.running || r.cmd == nil || r.cmd.Process == nil {
		r.mu.Unlock()
		return nil // Nothing to stop
	}

	slog.Info("Stopping process", "pid", r.cmd.Process.Pid)

	// Signal graceful shutdown
	if err := r.cmd.Process.Signal(os.Interrupt); err != nil {
		r.running = false
		r.mu.Unlock()
		return nil
	}
	r.mu.Unlock()

	// Poll the running flag (set to false by monitor when Wait returns)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
		r.mu.Lock()
		stillRunning := r.running
		r.mu.Unlock()
		if !stillRunning {
			slog.Info("Process stopped gracefully")
			return nil
		}
	}

	// Timeout — force kill
	slog.Warn("Process did not stop gracefully, force killing")
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := forceKill(r.cmd); err != nil {
		slog.Error("Failed to force kill process", "error", err)
		return err
	}
	return nil
}
