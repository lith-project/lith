//go:build windows

package daemon

// isProcessAlive on Windows conservatively reports every PID as alive. There
// is no portable signal-0 liveness probe on Windows, and a Windows PID can be
// recycled; wrongly reclaiming a live process's lock is worse than never
// reclaiming a stale one. Stale-lock reclaim on Windows can be implemented
// when the daemon runs there (see issue #61 Non-Goals).
func isProcessAlive(pid int) bool {
	return true
}
