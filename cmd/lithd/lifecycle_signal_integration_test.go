//go:build !windows

// Signal and crash lifecycle tests for the real lithd binary. These use
// POSIX signal delivery and are excluded from Windows builds (rule 7).
package main_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/lith-project/lith/internal/core/daemon"
	"github.com/lith-project/lith/internal/core/logging"
)

// TestLifecycle runs the real lithd binary through the M1-A end-to-end
// scenarios: file observation, SIGINT, SIGTERM, and SIGKILL with stale-lock
// reclaim.
func TestLifecycle(t *testing.T) {
	t.Run("file change is logged with a vault-relative path", func(t *testing.T) {
		env := newDaemonEnv(t, true)
		// Rewrite the file until the daemon observes it. The first write can
		// land in the brief window between daemon.started and the watcher's
		// root registration, so a bounded rewrite loop makes the assertion
		// deterministic rather than timing-dependent.
		writeUntilObserved(t, env, "note.md")
		assertCleanSignalShutdown(t, env, syscall.SIGTERM, "terminated")
	})

	t.Run("SIGINT clean shutdown", func(t *testing.T) {
		env := newDaemonEnv(t, true)
		assertCleanSignalShutdown(t, env, os.Interrupt, "interrupt")
	})

	t.Run("SIGTERM clean shutdown", func(t *testing.T) {
		env := newDaemonEnv(t, true)
		assertCleanSignalShutdown(t, env, syscall.SIGTERM, "terminated")
	})

	t.Run("SIGKILL leaves a stale lock that restart reclaims", func(t *testing.T) {
		env := newDaemonEnv(t, true)

		if _, err := os.Stat(env.lock); err != nil {
			t.Fatalf("lock file %s missing while daemon runs: %v", env.lock, err)
		}

		if err := env.proc.cmd.Process.Kill(); err != nil {
			t.Fatalf("kill lithd: %v", err)
		}
		if code := env.proc.waitForExit(t, 10*time.Second); code != -1 {
			t.Errorf("lithd exit code after SIGKILL = %d, want -1", code)
		}

		if _, err := os.Stat(env.lock); err != nil {
			t.Fatalf("lock file %s missing after SIGKILL (crash must leave it): %v", env.lock, err)
		}

		// Restart against the same vault and state dir: the stale lock is
		// reclaimed and startup completes.
		restarted := startDaemon(t, env.config, env.stateHome)
		restarted.waitForEvent(t, logging.EventDaemonStarted, 10*time.Second, nil)

		// The reclaimed daemon shuts down cleanly and removes the lock.
		start := time.Now()
		if err := restarted.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			t.Fatalf("signal restarted daemon: %v", err)
		}
		restarted.waitForEvent(t, logging.EventShutdownBegin, daemon.ShutdownBound+time.Second, func(rec map[string]any) bool {
			cause, _ := rec[logging.AttrCause].(string)
			return cause == "signal"
		})
		code := restarted.waitForExit(t, daemon.ShutdownBound+2*time.Second)
		if code != 0 {
			t.Errorf("restarted lithd exit code = %d, want 0", code)
		}
		if elapsed := time.Since(start); elapsed >= daemon.ShutdownBound+time.Second {
			t.Errorf("restart shutdown took %v, want under %v", elapsed, daemon.ShutdownBound+time.Second)
		}
		if _, err := os.Stat(env.lock); !os.IsNotExist(err) {
			t.Errorf("lock file %s still present after restarted daemon shutdown", env.lock)
		}
	})
}

// writeUntilObserved writes and rewrites a file in the vault root until the
// daemon logs a file.changed record carrying the file's vault-relative path.
func writeUntilObserved(t *testing.T, env *daemonEnv, name string) {
	t.Helper()
	path := filepath.Join(env.vault, name)
	payload := []byte("lith lifecycle integration\n")
	deadline := time.Now().Add(10 * time.Second)
	nextWrite := time.Now()
	for time.Now().Before(deadline) {
		if time.Now().After(nextWrite) {
			if err := os.WriteFile(path, payload, 0o644); err != nil {
				t.Fatalf("write %s: %v", path, err)
			}
			nextWrite = time.Now().Add(500 * time.Millisecond)
		}
		if rec := env.proc.pollEvent(t, logging.EventFileChanged, func(rec map[string]any) bool {
			p, _ := rec[logging.AttrPath].(string)
			return p == name
		}); rec != nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("no file.changed record for %q within timeout; stderr:\n%s", name, env.proc.stderr.String())
}
