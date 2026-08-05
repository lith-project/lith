package daemon

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lith-project/lith/internal/core/config"
)

// TestShutdownBoundMatchesDebounceMaxDelay asserts that ShutdownBound equals
// the configured default debounce maximum. The equal bounds are deliberate and
// must drift together (Rule 7).
func TestShutdownBoundMatchesDebounceMaxDelay(t *testing.T) {
	cfg := config.Defaults()
	if ShutdownBound != cfg.Watcher.Debounce.MaxDelay {
		t.Errorf("ShutdownBound = %v, want %v (config.Defaults().Watcher.Debounce.MaxDelay)",
			ShutdownBound, cfg.Watcher.Debounce.MaxDelay)
	}
}

// TestRunContextCancellation verifies that cancelling the context triggers
// clean shutdown of every responsive component.
func TestRunContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	log := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))

	var mu sync.Mutex
	cleaned := make(map[string]bool)
	components := []Component{
		{
			Name: "alpha",
			Run: func(ctx context.Context) error {
				<-ctx.Done()
				mu.Lock()
				cleaned["alpha"] = true
				mu.Unlock()
				return nil
			},
		},
		{
			Name: "beta",
			Run: func(ctx context.Context) error {
				<-ctx.Done()
				mu.Lock()
				cleaned["beta"] = true
				mu.Unlock()
				return nil
			},
		},
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, components, log)
	}()

	// Allow goroutines to start, then cancel.
	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run returned unexpected error: %v", err)
		}
	case <-time.After(ShutdownBound + time.Second):
		t.Fatal("Run did not return within ShutdownBound + margin")
	}

	mu.Lock()
	defer mu.Unlock()
	if !cleaned["alpha"] || !cleaned["beta"] {
		t.Errorf("components not both cleaned up: %v", cleaned)
	}
}

// TestRunFirstErrorCancelsSiblings verifies that the first non-cancellation
// error cancels siblings and is returned with the component name.
func TestRunFirstErrorCancelsSiblings(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))

	failingErr := &testError{msg: "something broke"}

	components := []Component{
		{
			Name: "good",
			Run: func(ctx context.Context) error {
				// Wait for context cancellation.
				<-ctx.Done()
				return nil
			},
		},
		{
			Name: "bad",
			Run: func(ctx context.Context) error {
				// Return a non-cancellation error immediately.
				return failingErr
			},
		},
	}

	err := Run(ctx, components, log)
	if err == nil {
		t.Fatal("expected non-nil error from Run")
	}

	if !strings.Contains(err.Error(), "bad") {
		t.Errorf("error should contain component name 'bad', got: %v", err)
	}
	if !strings.Contains(err.Error(), "something broke") {
		t.Errorf("error should wrap original message, got: %v", err)
	}
}

// TestRunUnresponsiveComponentLogged verifies that a component that ignores
// cancellation is logged at warn, and Run still returns within ShutdownBound.
func TestRunUnresponsiveComponentLogged(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	components := []Component{
		{
			Name: "stuck",
			Run: func(ctx context.Context) error {
				// Block forever, ignoring cancellation.
				select {}
			},
		},
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, components, log)
	}()

	// Allow goroutine to start, then cancel.
	time.Sleep(10 * time.Millisecond)
	cancel()

	start := time.Now()
	select {
	case <-errCh:
	case <-time.After(ShutdownBound + 2*time.Second):
		t.Fatal("Run did not return within ShutdownBound + margin")
	}
	elapsed := time.Since(start)

	if elapsed > ShutdownBound+time.Second {
		t.Errorf("Run took %v, expected <= ShutdownBound + margin", elapsed)
	}

	// Check that the warn log was emitted.
	logOutput := buf.String()
	if !strings.Contains(logOutput, "shutting down") {
		t.Errorf("expected 'shutting down' in log output, got: %s", logOutput)
	}
}

// TestRunShutdownDurationLogged verifies that EventShutdownDone is logged
// with AttrDuration as a number (milliseconds).
func TestRunShutdownDurationLogged(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	ctx, cancel := context.WithCancel(context.Background())

	components := []Component{
		{
			Name: "fast",
			Run: func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			},
		},
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, components, log)
	}()

	// Allow goroutine to start, then cancel.
	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case <-errCh:
	case <-time.After(ShutdownBound + time.Second):
		t.Fatal("Run did not return within ShutdownBound + margin")
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "shutdown complete") {
		t.Errorf("expected 'shutdown complete' in log output, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "duration_ms") {
		t.Errorf("expected 'duration_ms' in log output, got: %s", logOutput)
	}
}

// testError is a simple error type for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
