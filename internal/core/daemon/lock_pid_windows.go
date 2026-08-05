//go:build windows

package daemon

// maxWindowsPID is an upper bound on Windows PID values. PIDs above this
// cannot be live processes on Windows, so a lock file carrying one is stale.
// The bound is deliberately generous; real Windows PIDs are well below it.
const maxWindowsPID = 4194304

// isProcessAlive on Windows conservatively reports a PID as alive unless it
// is provably impossible (above maxWindowsPID). There is no portable
// signal-0 liveness probe here, and a Windows PID can be recycled; wrongly
// reclaiming a live process's lock is worse than never reclaiming a stale
// one. A more precise probe can be implemented when the daemon runs on
// Windows (see issue #61 Non-Goals).
func isProcessAlive(pid int) bool {
	return pid <= maxWindowsPID
}
