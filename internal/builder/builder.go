package builder

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Builder struct {
	cmd  string
	args []string
}

// NewBuilder creates a new Builder instance by parsing the provided build command string
func NewBuilder(buildCmd string) *Builder {
	// Parse command string
	parts := strings.Fields(buildCmd)

	if len(parts) == 0 {
		slog.Warn("No build command provided, skipping build")
		return &Builder{}
	}
	return &Builder{
		cmd:  parts[0],
		args: parts[1:],
	}
}

// Build the projecct using the specified command and arguments
func (b *Builder) Build(ctx context.Context) error {
	slog.Info("Building project", "command", b.cmd, "args", b.args)

	// Track time for logging
	start := time.Now()

	cmd := exec.CommandContext(ctx, b.cmd, b.args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		slog.Error("Build failed", "error", err)
		return err
	}

	duration := time.Since(start)
	slog.Info("Build completed successfully", "duration", duration)

	return nil
}
