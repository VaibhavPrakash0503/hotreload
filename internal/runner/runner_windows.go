//go:build windows

package runner

import (
	"log/slog"
	"os/exec"
	"syscall"
)

// configureProcAttr sets up process group for Windows
func configureProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

// forceKill kills the process
func forceKill(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}

	// On Windows, Kill() terminates the process tree
	if err := cmd.Process.Kill(); err != nil {
		slog.Warn("Failed to kill process",
			"pid", cmd.Process.Pid,
			"error", err)
		return err
	}

	slog.Debug("Killed process", "pid", cmd.Process.Pid)
	return nil
}
