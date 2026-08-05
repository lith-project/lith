package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()

	// Pin default values - silent drift fails tests
	if cfg.Log.Level != "info" {
		t.Errorf("default log level: got %q, want %q", cfg.Log.Level, "info")
	}
	if cfg.Log.Format != "text" {
		t.Errorf("default log format: got %q, want %q", cfg.Log.Format, "text")
	}
	if cfg.Watcher.Enabled != true {
		t.Errorf("default watcher enabled: got %v, want true", cfg.Watcher.Enabled)
	}
	if cfg.Watcher.Debounce.Quiet != 200*time.Millisecond {
		t.Errorf("default debounce quiet: got %v, want 200ms", cfg.Watcher.Debounce.Quiet)
	}
	if cfg.Watcher.Debounce.MaxDelay != 5*time.Second {
		t.Errorf("default debounce max_delay: got %v, want 5s", cfg.Watcher.Debounce.MaxDelay)
	}
	if cfg.Queue.Capacity != 4096 {
		t.Errorf("default queue capacity: got %d, want 4096", cfg.Queue.Capacity)
	}
	// vault.path should be empty (no default)
	if cfg.Vault.Path != "" {
		t.Errorf("default vault.path: got %q, want empty", cfg.Vault.Path)
	}
	// daemon.lock_file should be empty (resolved later)
	if cfg.Daemon.LockFile != "" {
		t.Errorf("default daemon.lock_file: got %q, want empty", cfg.Daemon.LockFile)
	}
}

func TestValidateZeroConfig(t *testing.T) {
	var cfg Config
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for zero Config, got nil")
	}

	// Should report missing vault.path
	if !errors.Is(err, err) {
		// Just check that we got an error
	}
	errStr := err.Error()
	if !contains(errStr, "vault.path") {
		t.Errorf("expected error about vault.path, got: %s", errStr)
	}
}

func TestValidateMultipleErrors(t *testing.T) {
	cfg := Config{
		Vault: Vault{Path: ""},       // empty path
		Log:   Log{Level: "invalid"}, // invalid level
		Queue: Queue{Capacity: 0},    // invalid capacity
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected errors, got nil")
	}

	errStr := err.Error()

	// Should contain all three problems
	if !contains(errStr, "vault.path") {
		t.Errorf("expected error about vault.path in: %s", errStr)
	}
	if !contains(errStr, "log.level") {
		t.Errorf("expected error about log.level in: %s", errStr)
	}
	if !contains(errStr, "queue.capacity") {
		t.Errorf("expected error about queue.capacity in: %s", errStr)
	}
}

func TestValidateVaultPathExpansion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantAbs  bool
		wantHome bool
	}{
		{
			name:     "tilde expansion",
			input:    "~/vault",
			wantAbs:  true,
			wantHome: true,
		},
		{
			name:    "relative path",
			input:   "./vault",
			wantAbs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Vault: Vault{Path: tt.input},
				Log:   Log{Level: "info", Format: "text"},
				Watcher: Watcher{
					Debounce: Debounce{
						Quiet:    200 * time.Millisecond,
						MaxDelay: 5 * time.Second,
					},
				},
				Queue: Queue{Capacity: 4096},
			}

			err := cfg.Validate()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Should be absolute
			if !filepath.IsAbs(cfg.Vault.Path) {
				t.Errorf("expected absolute path, got: %s", cfg.Vault.Path)
			}

			// If ~ input, should contain home directory
			if tt.wantHome {
				home, _ := os.UserHomeDir()
				if !contains(cfg.Vault.Path, home) {
					t.Errorf("expected path to contain home dir %q, got: %s", home, cfg.Vault.Path)
				}
			}
		})
	}
}

func TestValidateDebounceBounds(t *testing.T) {
	cfg := Config{
		Vault: Vault{Path: "/tmp/vault"},
		Log:   Log{Level: "info", Format: "text"},
		Watcher: Watcher{
			Debounce: Debounce{
				Quiet:    5 * time.Second,
				MaxDelay: 2 * time.Second, // quiet > max_delay
			},
		},
		Queue: Queue{Capacity: 4096},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for quiet >= max_delay, got nil")
	}

	errStr := err.Error()
	if !contains(errStr, "watcher.debounce") {
		t.Errorf("expected error about watcher.debounce, got: %s", errStr)
	}
}

func TestValidateQueueCapacity(t *testing.T) {
	cfg := Config{
		Vault: Vault{Path: "/tmp/vault"},
		Log:   Log{Level: "info", Format: "text"},
		Watcher: Watcher{
			Debounce: Debounce{
				Quiet:    200 * time.Millisecond,
				MaxDelay: 5 * time.Second,
			},
		},
		Queue: Queue{Capacity: 0}, // invalid
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for capacity < 1, got nil")
	}

	errStr := err.Error()
	if !contains(errStr, "queue.capacity") {
		t.Errorf("expected error about queue.capacity, got: %s", errStr)
	}
}

func TestValidateValidConfig(t *testing.T) {
	cfg := Config{
		Vault: Vault{Path: "/tmp/vault"},
		Log:   Log{Level: "info", Format: "text"},
		Watcher: Watcher{
			Debounce: Debounce{
				Quiet:    200 * time.Millisecond,
				MaxDelay: 5 * time.Second,
			},
		},
		Queue: Queue{Capacity: 4096},
	}

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("expected nil error for valid config, got: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
