package debounce

import (
	"testing"
	"time"

	"github.com/lith-project/lith/internal/core/config"
)

// TestDefaultDebounceBounds pins the literal default debounce bounds so any
// accidental change is caught immediately. These values are the contract
// between the config package and the debounce package.
func TestDefaultDebounceBounds(t *testing.T) {
	cfg := config.Defaults()

	if cfg.Watcher.Debounce.Quiet != 200*time.Millisecond {
		t.Errorf("default debounce quiet = %v, want 200ms", cfg.Watcher.Debounce.Quiet)
	}

	// The max delay is the RFC-0005 cancellation bound — it must not change
	// without an approved RFC.
	if cfg.Watcher.Debounce.MaxDelay != 5*time.Second {
		t.Errorf("default debounce max_delay = %v, want 5s (RFC-0005 cancellation bound)", cfg.Watcher.Debounce.MaxDelay)
	}
}
