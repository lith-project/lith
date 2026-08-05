//go:build !windows

package daemon

import "syscall"

// isProcessAlive reports whether the process with the given PID is running.
// syscall.Kill(pid, 0) is the canonical liveness probe on POSIX systems.
func isProcessAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil
}
