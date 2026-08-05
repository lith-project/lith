package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestLoadValidFile(t *testing.T) {
	path := filepath.Join("testdata", "valid.yaml")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// vault.path should be loaded
	absPath, _ := filepath.Abs("/tmp/test-vault")
	if cfg.Vault.Path != absPath {
		t.Errorf("vault.path: got %q, want %q", cfg.Vault.Path, absPath)
	}

	// All other fields should be defaults
	if cfg.Log.Level != "info" {
		t.Errorf("log.level: got %q, want %q", cfg.Log.Level, "info")
	}
	if cfg.Log.Format != "text" {
		t.Errorf("log.format: got %q, want %q", cfg.Log.Format, "text")
	}
	if cfg.Watcher.Enabled != true {
		t.Errorf("watcher.enabled: got %v, want true", cfg.Watcher.Enabled)
	}
	if cfg.Watcher.Debounce.Quiet != 200*time.Millisecond {
		t.Errorf("watcher.debounce.quiet: got %v, want 200ms", cfg.Watcher.Debounce.Quiet)
	}
	if cfg.Watcher.Debounce.MaxDelay != 5*time.Second {
		t.Errorf("watcher.debounce.max_delay: got %v, want 5s", cfg.Watcher.Debounce.MaxDelay)
	}
	if cfg.Queue.Capacity != 4096 {
		t.Errorf("queue.capacity: got %d, want 4096", cfg.Queue.Capacity)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected error to wrap ErrNotFound, got: %v", err)
	}
}

func TestLoadMalformedYAML(t *testing.T) {
	path := filepath.Join("testdata", "malformed.yaml")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}

	errStr := err.Error()
	if !containsLine(errStr) {
		t.Errorf("expected error to contain line number, got: %s", errStr)
	}
}

func TestLoadUnknownKey(t *testing.T) {
	path := filepath.Join("testdata", "unknown-key.yaml")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unknown key, got nil")
	}

	errStr := err.Error()
	if !contains(errStr, "unknown_key") {
		t.Errorf("expected error to name unknown key 'unknown_key', got: %s", errStr)
	}
}

func TestLoadMissingVaultPath(t *testing.T) {
	path := filepath.Join("testdata", "missing-vault-path.yaml")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing vault.path, got nil")
	}

	errStr := err.Error()
	if !contains(errStr, "vault.path") {
		t.Errorf("expected error about vault.path, got: %s", errStr)
	}
}

func TestLoadOptionalFieldsOmitted(t *testing.T) {
	path := filepath.Join("testdata", "optional-fields.yaml")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// vault.path should be set
	absPath, _ := filepath.Abs("/tmp/optional-fields-vault")
	if cfg.Vault.Path != absPath {
		t.Errorf("vault.path: got %q, want %q", cfg.Vault.Path, absPath)
	}

	// All optional fields should match Defaults()
	defaults := Defaults()
	if cfg.Log.Level != defaults.Log.Level {
		t.Errorf("log.level: got %q, want %q", cfg.Log.Level, defaults.Log.Level)
	}
	if cfg.Log.Format != defaults.Log.Format {
		t.Errorf("log.format: got %q, want %q", cfg.Log.Format, defaults.Log.Format)
	}
	if cfg.Watcher.Enabled != defaults.Watcher.Enabled {
		t.Errorf("watcher.enabled: got %v, want %v", cfg.Watcher.Enabled, defaults.Watcher.Enabled)
	}
	if cfg.Watcher.Debounce.Quiet != defaults.Watcher.Debounce.Quiet {
		t.Errorf("watcher.debounce.quiet: got %v, want %v", cfg.Watcher.Debounce.Quiet, defaults.Watcher.Debounce.Quiet)
	}
	if cfg.Watcher.Debounce.MaxDelay != defaults.Watcher.Debounce.MaxDelay {
		t.Errorf("watcher.debounce.max_delay: got %v, want %v", cfg.Watcher.Debounce.MaxDelay, defaults.Watcher.Debounce.MaxDelay)
	}
	if cfg.Queue.Capacity != defaults.Queue.Capacity {
		t.Errorf("queue.capacity: got %d, want %d", cfg.Queue.Capacity, defaults.Queue.Capacity)
	}
}

func TestLoadDurationStrings(t *testing.T) {
	path := filepath.Join("testdata", "full.yaml")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify duration strings decoded correctly
	if cfg.Watcher.Debounce.Quiet != 100*time.Millisecond {
		t.Errorf("watcher.debounce.quiet: got %v, want 100ms", cfg.Watcher.Debounce.Quiet)
	}
	if cfg.Watcher.Debounce.MaxDelay != 2*time.Second {
		t.Errorf("watcher.debounce.max_delay: got %v, want 2s", cfg.Watcher.Debounce.MaxDelay)
	}
}

func TestLoadFullConfig(t *testing.T) {
	path := filepath.Join("testdata", "full.yaml")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all fields are set
	absPath, _ := filepath.Abs("/tmp/full-vault")
	if cfg.Vault.Path != absPath {
		t.Errorf("vault.path: got %q, want %q", cfg.Vault.Path, absPath)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("log.level: got %q, want %q", cfg.Log.Level, "debug")
	}
	if cfg.Log.Format != "json" {
		t.Errorf("log.format: got %q, want %q", cfg.Log.Format, "json")
	}
	if cfg.Watcher.Enabled != false {
		t.Errorf("watcher.enabled: got %v, want false", cfg.Watcher.Enabled)
	}
	if cfg.Watcher.Debounce.Quiet != 100*time.Millisecond {
		t.Errorf("watcher.debounce.quiet: got %v, want 100ms", cfg.Watcher.Debounce.Quiet)
	}
	if cfg.Watcher.Debounce.MaxDelay != 2*time.Second {
		t.Errorf("watcher.debounce.max_delay: got %v, want 2s", cfg.Watcher.Debounce.MaxDelay)
	}
	if cfg.Queue.Capacity != 8192 {
		t.Errorf("queue.capacity: got %d, want 8192", cfg.Queue.Capacity)
	}
}

func TestLoadFilePermissionError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows: file permissions not enforced for file owner")
	}
	if os.Geteuid() == 0 {
		t.Skip("root: mode bits do not deny read, so the error cannot be provoked")
	}

	// Create a temp file with no read permissions
	tmpDir := t.TempDir()
	unreadable := filepath.Join(tmpDir, "unreadable.yaml")
	err := os.WriteFile(unreadable, []byte("vault:\n  path: /tmp/test\n"), 0000)
	if err != nil {
		t.Fatalf("failed to create unreadable file: %v", err)
	}

	_, err = Load(unreadable)
	if err == nil {
		t.Fatal("expected error for unreadable file, got nil")
	}

	// Should not be ErrNotFound (file exists)
	if errors.Is(err, ErrNotFound) {
		t.Error("expected error to NOT wrap ErrNotFound for permission error")
	}
}
