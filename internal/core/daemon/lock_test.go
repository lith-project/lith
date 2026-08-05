package daemon

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// testLogger returns a slog.Logger writing to a buffer for log inspection.
func testLogger(t *testing.T) (*slog.Logger, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	return log, &buf
}

// TestLockPathOutsideVault asserts that the lock path is outside the vault root.
func TestLockPathOutsideVault(t *testing.T) {
	vaultRoot := t.TempDir()
	log, _ := testLogger(t)

	lock, err := Acquire(vaultRoot, "", log)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}
	defer func() { _ = lock.Release() }()

	lockPath := lock.Path()
	vaultAbs := filepath.Clean(vaultRoot)

	if lockPath == vaultAbs || strings.HasPrefix(lockPath, vaultAbs+string(filepath.Separator)) {
		t.Errorf("lock path %q is inside vault root %q", lockPath, vaultRoot)
	}
}

// TestAcquireTwiceReturnsErrLocked verifies that Acquire twice in one process
// (same override path) returns ErrLocked the second time, naming the PID.
func TestAcquireTwiceReturnsErrLocked(t *testing.T) {
	vaultRoot := t.TempDir()
	lockPath := filepath.Join(t.TempDir(), "test.lock")
	log, _ := testLogger(t)

	// First acquire succeeds — writes our own (live) PID.
	lock1, err := Acquire(vaultRoot, lockPath, log)
	if err != nil {
		t.Fatalf("first Acquire failed: %v", err)
	}
	defer func() { _ = lock1.Release() }()

	// Second acquire on the same path should see our live PID and reject.
	_, err = Acquire(vaultRoot, lockPath, log)
	if err == nil {
		t.Fatal("expected ErrLocked, got nil")
	}
	if !errors.Is(err, ErrLocked) {
		t.Errorf("expected ErrLocked, got: %v", err)
	}
	// Verify the error names the PID.
	if !strings.Contains(err.Error(), strconv.Itoa(os.Getpid())) {
		t.Errorf("error should name our PID, got: %v", err)
	}
}

// TestDeadPIDReclaimed verifies that a lock file with a dead PID is reclaimed.
func TestDeadPIDReclaimed(t *testing.T) {
	vaultRoot := t.TempDir()
	log, buf := testLogger(t)

	// Write a lock file with a definitely dead PID
	lockPath := filepath.Join(t.TempDir(), "test.lock")
	deadPid := 999999
	if err := os.WriteFile(lockPath, []byte(strconv.Itoa(deadPid)), 0o644); err != nil {
		t.Fatalf("write dead lock: %v", err)
	}

	// Acquire should reclaim
	lock, err := Acquire(vaultRoot, lockPath, log)
	if err != nil {
		t.Fatalf("Acquire should reclaim dead PID, got: %v", err)
	}
	defer func() { _ = lock.Release() }()

	// Verify warn log
	logOutput := buf.String()
	if !strings.Contains(logOutput, "stale lock reclaimed") {
		t.Errorf("expected 'stale lock reclaimed' in log, got: %s", logOutput)
	}
}

// TestEmptyLockReclaimed verifies that an empty lock file is reclaimed.
func TestEmptyLockReclaimed(t *testing.T) {
	vaultRoot := t.TempDir()
	log, buf := testLogger(t)

	lockPath := filepath.Join(t.TempDir(), "empty.lock")
	if err := os.WriteFile(lockPath, []byte{}, 0o644); err != nil {
		t.Fatalf("write empty lock: %v", err)
	}

	lock, err := Acquire(vaultRoot, lockPath, log)
	if err != nil {
		t.Fatalf("Acquire should reclaim empty lock, got: %v", err)
	}
	defer func() { _ = lock.Release() }()

	logOutput := buf.String()
	if !strings.Contains(logOutput, "empty lock file, reclaiming") {
		t.Errorf("expected 'empty lock file, reclaiming' in log, got: %s", logOutput)
	}
}

// TestCorruptLockReclaimed verifies that a corrupt lock file is reclaimed.
func TestCorruptLockReclaimed(t *testing.T) {
	vaultRoot := t.TempDir()
	log, buf := testLogger(t)

	lockPath := filepath.Join(t.TempDir(), "corrupt.lock")
	if err := os.WriteFile(lockPath, []byte("not-a-number"), 0o644); err != nil {
		t.Fatalf("write corrupt lock: %v", err)
	}

	lock, err := Acquire(vaultRoot, lockPath, log)
	if err != nil {
		t.Fatalf("Acquire should reclaim corrupt lock, got: %v", err)
	}
	defer func() { _ = lock.Release() }()

	logOutput := buf.String()
	if !strings.Contains(logOutput, "corrupt lock file, reclaiming") {
		t.Errorf("expected 'corrupt lock file, reclaiming' in log, got: %s", logOutput)
	}
}

// TestReleaseDoesNotRemoveOtherLock verifies that Release does not remove
// a lock file that was taken over by another process.
func TestReleaseDoesNotRemoveOtherLock(t *testing.T) {
	vaultRoot := t.TempDir()
	log, _ := testLogger(t)

	lock, err := Acquire(vaultRoot, "", log)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	// Simulate another process taking over the lock file
	// by writing a different PID
	otherPid := os.Getpid() + 10000 // definitely different from our PID
	if err := os.WriteFile(lock.Path(), []byte(strconv.Itoa(otherPid)), 0o644); err != nil {
		t.Fatalf("write other PID: %v", err)
	}

	// Release should not remove the file
	if err := lock.Release(); err != nil {
		t.Fatalf("Release failed: %v", err)
	}

	// Verify file still exists
	if _, err := os.Stat(lock.Path()); os.IsNotExist(err) {
		t.Error("Release removed another process's lock file")
	}
}

// TestReleaseIdempotent verifies that calling Release twice is safe.
func TestReleaseIdempotent(t *testing.T) {
	vaultRoot := t.TempDir()
	log, _ := testLogger(t)

	lock, err := Acquire(vaultRoot, "", log)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	if err := lock.Release(); err != nil {
		t.Fatalf("first Release failed: %v", err)
	}

	if err := lock.Release(); err != nil {
		t.Fatalf("second Release failed: %v", err)
	}
}

// TestOverrideInsideVaultRejected verifies that a lock path inside the vault is rejected.
func TestOverrideInsideVaultRejected(t *testing.T) {
	vaultRoot := t.TempDir()
	log, _ := testLogger(t)

	insidePath := filepath.Join(vaultRoot, "locks", "test.lock")
	_, err := Acquire(vaultRoot, insidePath, log)
	if err == nil {
		t.Fatal("expected error for lock path inside vault, got nil")
	}
}
