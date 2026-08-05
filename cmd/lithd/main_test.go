package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// testVault is the vault path hardcoded in the testdata configs.
const testVault = "/tmp/test-vault"

// lockedBuffer is a concurrency-safe bytes.Buffer. A live daemon writes log
// records from several goroutines (components plus the shutdown observer), so
// tests must not hand run() a bare bytes.Buffer under -race.
type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (w *lockedBuffer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.b.Write(p)
}

func (w *lockedBuffer) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.b.String()
}

// newTestEnv isolates lock-file side effects per test (every test gets its own
// XDG state dir) and ensures the shared test vault directory exists, which the
// fsnotify watcher requires.
func newTestEnv(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	if err := os.MkdirAll(testVault, 0o755); err != nil {
		t.Fatalf("create test vault: %v", err)
	}
}

func testdataPath(t *testing.T, name string) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "testdata", name)
}

// startRun invokes run in a goroutine; the returned channel receives the exit code.
func startRun(t *testing.T, ctx context.Context, args []string, stdout, stderr io.Writer) <-chan int {
	t.Helper()
	result := make(chan int, 1)
	go func() { result <- run(ctx, args, stdout, stderr) }()
	return result
}

// waitForOutput polls buf until it contains want or times out.
func waitForOutput(t *testing.T, buf *lockedBuffer, want string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(buf.String(), want) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("output did not contain %q within timeout:\n%s", want, buf.String())
}

// waitForExit asserts that run returned the expected exit code.
func waitForExit(t *testing.T, result <-chan int, want int) {
	t.Helper()
	select {
	case got := <-result:
		if got != want {
			t.Errorf("run() = %d, want %d", got, want)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("run did not return within timeout")
	}
}

// parseRecords splits stderr into JSON log records.
func parseRecords(t *testing.T, buf *lockedBuffer) []map[string]interface{} {
	t.Helper()
	var records []map[string]interface{}
	for _, line := range strings.Split(buf.String(), "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var rec map[string]interface{}
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("failed to parse JSON log line: %v\nline: %s", err, line)
		}
		records = append(records, rec)
	}
	return records
}

// findRecord returns the first record whose msg field equals want.
func findRecord(t *testing.T, buf *lockedBuffer, msg string) map[string]interface{} {
	t.Helper()
	for _, rec := range parseRecords(t, buf) {
		if m, ok := rec["msg"]; ok && m == msg {
			return rec
		}
	}
	t.Fatalf("no %q log record found", msg)
	return nil
}

func TestRun(t *testing.T) {
	tests := []struct {
		name       string
		config     string // empty means no --config flag
		wantCode   int
		wantStderr string
	}{
		{
			name:       "no args",
			wantCode:   2,
			wantStderr: "Usage:",
		},
		{
			name:       "missing --config flag",
			wantCode:   2,
			wantStderr: "Usage:",
		},
		{
			name:       "missing file",
			config:     "/nonexistent.yaml",
			wantCode:   2,
			wantStderr: "file not found",
		},
		{
			name:       "malformed yaml",
			config:     testdataPath(t, "bad.yaml"),
			wantCode:   2,
			wantStderr: "line ",
		},
		{
			name:       "invalid log format",
			config:     testdataPath(t, "bad-log.yaml"),
			wantCode:   2,
			wantStderr: "log.format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var args []string
			if tt.config != "" {
				args = []string{"--config", tt.config}
			}
			var stdout, stderr lockedBuffer
			got := run(context.Background(), args, &stdout, &stderr)

			if got != tt.wantCode {
				t.Errorf("run() = %d, want %d", got, tt.wantCode)
			}
			if !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr = %q, want to contain %q", stderr.String(), tt.wantStderr)
			}
			if stdout.String() != "" {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
		})
	}
}

// TestRunNoopWatcherCancellation verifies the M1-A9 DoD: a cancelled context
// cleanly stops the full watcher → debouncer → queue pipeline and returns 0.
func TestRunNoopWatcherCancellation(t *testing.T) {
	newTestEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stdout, stderr lockedBuffer
	result := startRun(t, ctx, []string{"--config", testdataPath(t, "noop-watcher.yaml")}, &stdout, &stderr)

	waitForOutput(t, &stderr, "daemon.started")
	cancel()

	waitForExit(t, result, 0)

	if stdout.String() != "" {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
}

// TestRunWatcherEnabled verifies clean cancellation with the real fsnotify
// watcher running over the test vault directory.
func TestRunWatcherEnabled(t *testing.T) {
	newTestEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stdout, stderr lockedBuffer
	result := startRun(t, ctx, []string{"--config", testdataPath(t, "valid-log.yaml")}, &stdout, &stderr)

	waitForOutput(t, &stderr, "daemon.started")
	cancel()

	waitForExit(t, result, 0)

	if stdout.String() != "" {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunJSONLogRecords(t *testing.T) {
	newTestEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stdout, stderr lockedBuffer
	result := startRun(t, ctx, []string{"--config", testdataPath(t, "valid-log.yaml")}, &stdout, &stderr)

	waitForOutput(t, &stderr, "daemon.started")
	cancel()
	waitForExit(t, result, 0)
	// shutdown.begin is written by a separate goroutine and may land after
	// run() returns; poll until it appears.
	waitForOutput(t, &stderr, "shutdown.begin")

	if stdout.String() != "" {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}

	records := parseRecords(t, &stderr)
	if len(records) < 6 {
		t.Fatalf("got %d log records, want at least 6", len(records))
	}

	// Startup records 1-6 are emitted in order before daemon.Run starts.
	if msg, ok := records[0]["msg"]; !ok || msg != "daemon.starting" {
		t.Errorf("record 1 msg = %v, want \"daemon.starting\"", records[0]["msg"])
	}

	if msg, ok := records[1]["msg"]; !ok || msg != "config.loaded" {
		t.Errorf("record 2 msg = %v, want \"config.loaded\"", records[1]["msg"])
	}
	vp, ok := records[1]["vault_path"]
	if !ok {
		t.Errorf("record 2 missing vault_path field")
	} else {
		vpStr, ok := vp.(string)
		if !ok || vpStr == "" {
			t.Errorf("record 2 vault_path = %v, want non-empty string", vp)
		} else if !strings.HasSuffix(vpStr, "test-vault") {
			t.Errorf("record 2 vault_path = %v, want path ending with \"test-vault\"", vp)
		} else if !filepath.IsAbs(vpStr) {
			t.Errorf("record 2 vault_path = %v, want absolute path", vp)
		}
	}

	if msg, ok := records[2]["msg"]; !ok || msg != "vault.watching" {
		t.Errorf("record 3 msg = %v, want \"vault.watching\"", records[2]["msg"])
	}

	if msg, ok := records[3]["msg"]; !ok || msg != "debounce.bounds" {
		t.Errorf("record 4 msg = %v, want \"debounce.bounds\"", records[3]["msg"])
	}
	if quiet, ok := records[3]["quiet"]; !ok || quiet != "200ms" {
		t.Errorf("record 4 quiet = %v, want \"200ms\"", records[3]["quiet"])
	}
	if maxDelay, ok := records[3]["max_delay"]; !ok || maxDelay != "5s" {
		t.Errorf("record 4 max_delay = %v, want \"5s\"", records[3]["max_delay"])
	}

	if msg, ok := records[4]["msg"]; !ok || msg != "queue.capacity" {
		t.Errorf("record 5 msg = %v, want \"queue.capacity\"", records[4]["msg"])
	}
	if cap, ok := records[4]["capacity"]; !ok || cap != float64(4096) {
		t.Errorf("record 5 capacity = %v, want 4096", records[4]["capacity"])
	}

	if msg, ok := records[5]["msg"]; !ok || msg != "daemon.started" {
		t.Errorf("record 6 msg = %v, want \"daemon.started\"", records[5]["msg"])
	}

	// Shutdown event: shutdown.begin with cause=signal and the fallback name
	// (the test context carries no signal-name channel).
	begin := findRecord(t, &stderr, "shutdown.begin")
	if cause, ok := begin["cause"]; !ok || cause != "signal" {
		t.Errorf("shutdown.begin cause = %v, want \"signal\"", begin["cause"])
	}
	if sig, ok := begin["signal"]; !ok || sig != "signal" {
		t.Errorf("shutdown.begin signal = %v, want \"signal\"", begin["signal"])
	}
}

// TestWatchSignals verifies the main() signal wiring: the first signal's name
// is published and the context cancelled, and a second signal during shutdown
// forces an immediate non-zero exit after a distinct warn log.
func TestWatchSignals(t *testing.T) {
	sigCh := make(chan os.Signal, 2)
	nameCh := make(chan string, 1)
	ctx, cancel := context.WithCancel(context.Background())
	var warn bytes.Buffer
	exitCodes := make(chan int, 1)
	done := make(chan struct{})

	go func() {
		defer close(done)
		watchSignals(sigCh, nameCh, cancel, &warn, func(code int) { exitCodes <- code })
	}()

	// First signal: name is published before the context is cancelled.
	sigCh <- os.Interrupt
	select {
	case n := <-nameCh:
		if n != "interrupt" {
			t.Errorf("signal name = %q, want \"interrupt\"", n)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("signal name not published within timeout")
	}
	select {
	case <-ctx.Done():
	case <-time.After(10 * time.Second):
		t.Fatal("context not cancelled within timeout")
	}

	// Second signal during shutdown: force exit 1 with a warn log.
	sigCh <- syscall.SIGTERM
	select {
	case code := <-exitCodes:
		if code != 1 {
			t.Errorf("force exit code = %d, want 1", code)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("force exit not triggered within timeout")
	}
	if !strings.Contains(warn.String(), "second signal received, forcing exit") {
		t.Errorf("warn = %q, want force-exit message", warn.String())
	}
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("watchSignals did not return within timeout")
	}
}

// TestWatchSignalsClosedChannel verifies that a closed signal channel stops
// the watcher without publishing a name or cancelling.
func TestWatchSignalsClosedChannel(t *testing.T) {
	sigCh := make(chan os.Signal)
	close(sigCh)
	nameCh := make(chan string, 1)
	ctx, cancel := context.WithCancel(context.Background())

	watchSignals(sigCh, nameCh, cancel, io.Discard, func(code int) {
		t.Errorf("force exit called with code %d on closed channel", code)
	})

	select {
	case n := <-nameCh:
		t.Errorf("signal name published on closed channel: %q", n)
	default:
	}
	select {
	case <-ctx.Done():
		t.Error("context cancelled on closed channel")
	default:
	}
}

// TestRunShutdownBeginSignalName verifies that shutdown.begin carries the
// published signal name, mirroring how main() wires the signal context.
func TestRunShutdownBeginSignalName(t *testing.T) {
	newTestEnv(t)

	sigNameCh := make(chan string, 1)
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, signalNameKey{}, sigNameCh)
	defer cancel()

	var stdout, stderr lockedBuffer
	result := startRun(t, ctx, []string{"--config", testdataPath(t, "noop-watcher.yaml")}, &stdout, &stderr)

	waitForOutput(t, &stderr, "daemon.started")

	// main() publishes the signal name before cancelling the context.
	sigNameCh <- "interrupt"
	cancel()

	waitForExit(t, result, 0)
	// noop-watcher.yaml uses the default text log format, so the shutdown
	// record is asserted with key=value pairs.
	waitForOutput(t, &stderr, "msg=shutdown.begin")

	out := stderr.String()
	if !strings.Contains(out, "cause=signal") {
		t.Errorf("stderr missing cause=signal:\n%s", out)
	}
	if !strings.Contains(out, "signal=interrupt") {
		t.Errorf("stderr missing signal=interrupt:\n%s", out)
	}
}

// TestRunDebounceOverride verifies that overriding debounce bounds in config
// changes the effective bounds logged at startup. This proves the composition
// root reads config and passes bounds explicitly.
func TestRunDebounceOverride(t *testing.T) {
	newTestEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stdout, stderr lockedBuffer
	result := startRun(t, ctx, []string{"--config", testdataPath(t, "override-debounce.yaml")}, &stdout, &stderr)

	waitForOutput(t, &stderr, "daemon.started")
	cancel()
	waitForExit(t, result, 0)

	boundsRec := findRecord(t, &stderr, "debounce.bounds")
	if quiet, ok := boundsRec["quiet"]; !ok || quiet != "500ms" {
		t.Errorf("debounce.bounds quiet = %v, want \"500ms\"", boundsRec["quiet"])
	}
	if maxDelay, ok := boundsRec["max_delay"]; !ok || maxDelay != "10s" {
		t.Errorf("debounce.bounds max_delay = %v, want \"10s\"", boundsRec["max_delay"])
	}
}
