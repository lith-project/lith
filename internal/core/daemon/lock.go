package daemon

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ErrLocked is returned when another live process holds the lock.
var ErrLocked = errors.New("daemon: vault is locked by another process")

// Lock is an exclusive, process-scoped claim on a vault.
type Lock struct {
	path     string
	pid      int
	released bool
}

// Acquire claims a lock for vaultRoot, writing the current PID. overridePath
// may be empty to use the XDG state directory. It returns ErrLocked when
// another live process holds the lock.
func Acquire(vaultRoot, overridePath string, log *slog.Logger) (*Lock, error) {
	lockPath, err := resolveLockPath(vaultRoot, overridePath)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return nil, fmt.Errorf("daemon: create lock dir: %w", err)
	}

	pid := os.Getpid()
	pidStr := strconv.Itoa(pid)

	err = writeLockFile(lockPath, pidStr)
	if errors.Is(err, os.ErrExist) {
		return reclaimOrReject(lockPath, pid, log)
	}
	if err != nil {
		return nil, fmt.Errorf("daemon: create lock: %w", err)
	}

	return &Lock{path: lockPath, pid: pid}, nil
}

// writeLockFile creates a lock file atomically with O_CREATE|O_EXCL.
func writeLockFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, writeErr := f.WriteString(content)
	closeErr := f.Close()
	if writeErr != nil {
		os.Remove(path)
		return writeErr
	}
	return closeErr
}

// reclaimOrReject attempts to reclaim a stale lock or returns ErrLocked.
func reclaimOrReject(lockPath string, myPid int, log *slog.Logger) (*Lock, error) {
	data, err := os.ReadFile(lockPath)
	if err != nil {
		log.Warn("daemon: lock file unreadable, reclaiming", "path", lockPath)
		return forceAcquire(lockPath, myPid)
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		log.Warn("daemon: empty lock file, reclaiming", "path", lockPath)
		return forceAcquire(lockPath, myPid)
	}

	oldPid, err := strconv.Atoi(content)
	if err != nil {
		log.Warn("daemon: corrupt lock file, reclaiming", "path", lockPath, "content", content)
		return forceAcquire(lockPath, myPid)
	}

	if isProcessAlive(oldPid) {
		return nil, fmt.Errorf("%w: PID %d", ErrLocked, oldPid)
	}

	log.Warn("daemon: stale lock reclaimed", "path", lockPath, "dead_pid", oldPid)
	return forceAcquire(lockPath, myPid)
}

// forceAcquire removes the existing lock file and creates a new one.
func forceAcquire(lockPath string, myPid int) (*Lock, error) {
	os.Remove(lockPath)
	pidStr := strconv.Itoa(myPid)
	if err := writeLockFile(lockPath, pidStr); err != nil {
		return nil, fmt.Errorf("daemon: create lock: %w", err)
	}
	return &Lock{path: lockPath, pid: myPid}, nil
}

// Release removes the lock. Safe to call more than once.
func (l *Lock) Release() error {
	if l.released {
		return nil
	}

	data, err := os.ReadFile(l.path)
	if err != nil {
		if os.IsNotExist(err) {
			l.released = true
			return nil
		}
		return fmt.Errorf("daemon: read lock for release: %w", err)
	}

	content := strings.TrimSpace(string(data))
	if content != strconv.Itoa(l.pid) {
		l.released = true
		return nil
	}

	if err := os.Remove(l.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("daemon: remove lock: %w", err)
	}

	l.released = true
	return nil
}

// Path returns the absolute lock-file path.
func (l *Lock) Path() string {
	return l.path
}

// resolveLockPath determines the lock file path based on override or XDG state directory.
func resolveLockPath(vaultRoot, overridePath string) (string, error) {
	if overridePath != "" {
		abs, err := filepath.Abs(overridePath)
		if err != nil {
			return "", fmt.Errorf("daemon: resolve lock path: %w", err)
		}
		vaultAbs := filepath.Clean(vaultRoot)
		if strings.HasPrefix(abs, vaultAbs+string(filepath.Separator)) || abs == vaultAbs {
			return "", fmt.Errorf("daemon: lock path must not be inside vault root")
		}
		return abs, nil
	}

	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("daemon: cannot determine home dir: %w", err)
		}
		stateHome = filepath.Join(home, ".local", "state")
	}

	vaultHash := hashPath(vaultRoot)
	return filepath.Join(stateHome, "lith", vaultHash+".lock"), nil
}

// hashPath returns a deterministic hash of the cleaned path.
func hashPath(p string) string {
	h := sha256.Sum256([]byte(filepath.Clean(p)))
	return fmt.Sprintf("%x", h[:16])
}
