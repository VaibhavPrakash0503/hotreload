//go:build unix

package runner

import (
	"log/slog"
	"os/exec"
	"syscall"
)

// configureProcAttr sets up process group for Unix systems
func configureProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create new process group
	}
}

// forceKill kills the process and all its children
func forceKill(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}

	pgid := cmd.Process.Pid

	// Kill entire process group (negative PID)
	if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
		slog.Warn("Failed to kill process group, trying process only",
			"pgid", pgid,
			"error", err)
		// Fallback: kill just the process
		return syscall.Kill(pgid, syscall.SIGKILL)
	}

	slog.Debug("Killed process group", "pgid", pgid)
	return nil
}
