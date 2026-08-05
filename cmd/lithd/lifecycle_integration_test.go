// Package main_test exercises the real lithd binary end to end: config
// loading, vault lock lifecycle, file observation, signal-driven shutdown,
// and the zero-plugin default. Process behavior is tested by exec'ing the
// compiled binary, never by calling run() directly.
package main_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lith-project/lith/internal/core/daemon"
	"github.com/lith-project/lith/internal/core/logging"
)

// lithdBinary is the path of the daemon binary, built once by TestMain.
var lithdBinary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "lithd-test-bin-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "lifecycle test: create build dir: %v\n", err)
		os.Exit(1)
	}
	binary := filepath.Join(dir, "lithd")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	out, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "lifecycle test: go build lithd: %v\n%s\n", err, out)
		os.Exit(1)
	}
	lithdBinary = binary
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// syncBuffer is a concurrency-safe bytes.Buffer: the child process's log
// stream is copied into it by os/exec while the test polls it.
type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (w *syncBuffer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.b.Write(p)
}

func (w *syncBuffer) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.b.String()
}

// daemonProc is a running lithd process under test.
type daemonProc struct {
	cmd      *exec.Cmd
	stderr   *syncBuffer
	exited   chan struct{}
	exitCode int
	mu       sync.Mutex
}

// startDaemon launches the real lithd binary with an isolated XDG state dir.
func startDaemon(t *testing.T, configPath, stateHome string) *daemonProc {
	t.Helper()
	cmd := exec.Command(lithdBinary, "--config", configPath)
	cmd.Env = append(os.Environ(), "XDG_STATE_HOME="+stateHome)
	d := &daemonProc{
		cmd:    cmd,
		stderr: &syncBuffer{},
		exited: make(chan struct{}),
	}
	cmd.Stderr = d.stderr
	cmd.Stdout = &bytes.Buffer{}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start lithd: %v", err)
	}
	go func() {
		code := 0
		if err := cmd.Wait(); err != nil {
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				code = ee.ExitCode()
			} else {
				code = -1
			}
		}
		d.mu.Lock()
		d.exitCode = code
		d.mu.Unlock()
		close(d.exited)
	}()
	t.Cleanup(func() {
		select {
		case <-d.exited:
		default:
			_ = d.cmd.Process.Kill()
			<-d.exited
		}
	})
	return d
}

func (d *daemonProc) code() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.exitCode
}

// waitForExit blocks until the process exits or the timeout elapses.
func (d *daemonProc) waitForExit(t *testing.T, timeout time.Duration) int {
	t.Helper()
	select {
	case <-d.exited:
		return d.code()
	case <-time.After(timeout):
		t.Fatalf("lithd did not exit within %v; stderr:\n%s", timeout, d.stderr.String())
		return -1
	}
}

// records parses all JSON log lines captured from stderr.
func (d *daemonProc) records(t *testing.T) []map[string]any {
	t.Helper()
	var recs []map[string]any
	for _, line := range strings.Split(d.stderr.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("parse JSON log line: %v\n%s", err, line)
		}
		recs = append(recs, rec)
	}
	return recs
}

// pollEvent checks the current log stream once for a record whose msg equals
// the given event and satisfies match.
func (d *daemonProc) pollEvent(t *testing.T, msg string, match func(map[string]any) bool) map[string]any {
	t.Helper()
	for _, rec := range d.records(t) {
		if m, _ := rec["msg"].(string); m == msg && (match == nil || match(rec)) {
			return rec
		}
	}
	return nil
}

// waitForEvent polls the log stream until a matching record appears, the
// process exits, or the timeout elapses.
func (d *daemonProc) waitForEvent(t *testing.T, msg string, timeout time.Duration, match func(map[string]any) bool) map[string]any {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if rec := d.pollEvent(t, msg, match); rec != nil {
			return rec
		}
		select {
		case <-d.exited:
			t.Fatalf("lithd exited (code %d) before emitting %q; stderr:\n%s", d.code(), msg, d.stderr.String())
		default:
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out after %v waiting for %q; stderr:\n%s", timeout, msg, d.stderr.String())
	return nil
}

// daemonEnv holds the isolated directories and process for one scenario.
type daemonEnv struct {
	vault     string
	stateHome string
	config    string
	lock      string
	proc      *daemonProc
}

// newDaemonEnv writes a config, starts the real binary with an isolated
// XDG_STATE_HOME, and waits for the daemon.started record so every component
// is known to be up before the test acts.
func newDaemonEnv(t *testing.T, watcherEnabled bool) *daemonEnv {
	t.Helper()
	vault := t.TempDir()
	stateHome := t.TempDir()
	cfg := writeConfig(t, vault, watcherEnabled)
	d := startDaemon(t, cfg, stateHome)
	d.waitForEvent(t, logging.EventDaemonStarted, 10*time.Second, nil)
	return &daemonEnv{
		vault:     vault,
		stateHome: stateHome,
		config:    cfg,
		lock:      lockPath(vault, stateHome),
		proc:      d,
	}
}

// writeConfig writes a JSON-logging config for the given vault.
func writeConfig(t *testing.T, vault string, watcherEnabled bool) string {
	t.Helper()
	cfg := fmt.Sprintf(
		"vault:\n  path: %s\nlog:\n  format: json\n  level: debug\nwatcher:\n  enabled: %t\n",
		strconv.Quote(vault), watcherEnabled,
	)
	path := filepath.Join(t.TempDir(), "lithd.yaml")
	if err := os.WriteFile(path, []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

// lockPath mirrors daemon.resolveLockPath so tests can assert lock presence
// and removal without reaching into production code.
func lockPath(vaultRoot, stateHome string) string {
	h := sha256.Sum256([]byte(filepath.Clean(vaultRoot)))
	return filepath.Join(stateHome, "lith", fmt.Sprintf("%x", h[:16])+".lock")
}

// assertCleanSignalShutdown delivers sig and verifies the full clean-shutdown
// contract: shutdown.begin carries cause=signal and the signal name, the
// daemon emits its deterministic completion record, the process exits 0
// within ShutdownBound, and the lock file is removed.
func assertCleanSignalShutdown(t *testing.T, env *daemonEnv, sig os.Signal, wantSignal string) {
	t.Helper()
	start := time.Now()
	if err := env.proc.cmd.Process.Signal(sig); err != nil {
		t.Fatalf("signal %v: %v", sig, err)
	}
	env.proc.waitForEvent(t, logging.EventShutdownBegin, daemon.ShutdownBound+time.Second, func(rec map[string]any) bool {
		cause, _ := rec[logging.AttrCause].(string)
		name, _ := rec[logging.AttrSignal].(string)
		return cause == "signal" && name == wantSignal
	})
	// daemon.Run's generic shutdown-start record.
	env.proc.waitForEvent(t, "shutting down", daemon.ShutdownBound+time.Second, func(rec map[string]any) bool {
		cause, _ := rec[logging.AttrCause].(string)
		return cause == "context canceled"
	})
	// daemon.Run's deterministic completion record, emitted before it returns.
	env.proc.waitForEvent(t, "shutdown complete", daemon.ShutdownBound+time.Second, func(rec map[string]any) bool {
		_, ok := rec[logging.AttrDuration].(float64)
		return ok
	})
	code := env.proc.waitForExit(t, daemon.ShutdownBound+2*time.Second)
	elapsed := time.Since(start)
	if code != 0 {
		t.Errorf("lithd exit code = %d, want 0", code)
	}
	if elapsed >= daemon.ShutdownBound+time.Second {
		t.Errorf("shutdown took %v, want under %v", elapsed, daemon.ShutdownBound+time.Second)
	}
	if _, err := os.Stat(env.lock); !os.IsNotExist(err) {
		t.Errorf("lock file %s still present after clean shutdown", env.lock)
	}
}

// stopClean stops the daemon and asserts a clean shutdown. On Windows, where
// a child process cannot be delivered SIGINT/SIGTERM, the process is killed
// and only its exit is asserted.
func stopClean(t *testing.T, env *daemonEnv) {
	t.Helper()
	if runtime.GOOS == "windows" {
		if err := env.proc.cmd.Process.Kill(); err != nil {
			t.Fatalf("kill lithd: %v", err)
		}
		env.proc.waitForExit(t, daemon.ShutdownBound+2*time.Second)
		return
	}
	assertCleanSignalShutdown(t, env, os.Interrupt, "interrupt")
}

// TestZeroPluginStart verifies the M1-A portion of RFC-0001/C-6: with
// watcher.enabled=false and no plugin package, configuration, or registry,
// the daemon starts and stops cleanly.
func TestZeroPluginStart(t *testing.T) {
	env := newDaemonEnv(t, false)
	if _, err := os.Stat(env.lock); err != nil {
		t.Fatalf("lock file %s missing while zero-plugin daemon runs: %v", env.lock, err)
	}
	stopClean(t, env)
}
