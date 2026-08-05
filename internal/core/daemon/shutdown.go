package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/lith-project/lith/internal/core/logging"
)

// ShutdownBound is the maximum time shutdown may take before the process
// exits regardless, per RFC-0005 §3.
const ShutdownBound = 5 * time.Second

// Component is one cancellable daemon loop. Run must return when ctx is
// cancelled; the daemon still enforces ShutdownBound if it does not.
type Component struct {
	Name string
	Run  func(context.Context) error
}

// componentResult holds the outcome of a single component's Run call.
type componentResult struct {
	Name string
	Err  error
}

// Run starts every component, cancels siblings on the first component error,
// and waits no longer than ShutdownBound after cancellation.
func Run(ctx context.Context, components []Component, log *slog.Logger) error {
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Launch every component in its own goroutine.
	results := make(chan componentResult, len(components))
	var wg sync.WaitGroup
	for _, comp := range components {
		wg.Add(1)
		go func(c Component) {
			defer wg.Done()
			err := c.Run(childCtx)
			results <- componentResult{Name: c.Name, Err: err}
		}(comp)
	}

	// Drain results until one of two things happens:
	//   1. A component returns a non-cancellation error → cancel siblings, return.
	//   2. The parent context is cancelled (signal / external) → begin bounded shutdown.
	var firstErr error
	var firstErrName string
	shutdownStarted := false

	for !shutdownStarted {
		select {
		case res, ok := <-results:
			if !ok {
				// Channel closed – all components finished.
				if firstErr != nil {
					return fmt.Errorf("daemon: %s: %w", firstErrName, firstErr)
				}
				return nil
			}
			if res.Err == nil {
				continue
			}
			// Context cancellation errors are expected during shutdown; ignore.
			if res.Err == context.Canceled || res.Err == context.DeadlineExceeded {
				continue
			}
			// First real error: cancel siblings and record it.
			firstErr = res.Err
			firstErrName = res.Name
			cancel()
			shutdownStarted = true

		case <-ctx.Done():
			// Parent cancelled – begin bounded shutdown.
			shutdownStarted = true
		}
	}

	// ---- bounded shutdown path ----
	shutdownStart := time.Now()
	log.Warn("shutting down", slog.String(logging.AttrCause, "context canceled"))

	timer := time.NewTimer(ShutdownBound)
	defer timer.Stop()

	// Wait for all goroutines to finish or the timer to fire.
	finished := make(chan struct{})
	go func() {
		wg.Wait()
		close(finished)
	}()

	select {
	case <-timer.C:
		// Timed out – log remaining components as unresponsive.
		log.Warn("shutdown bound reached, some components unresponsive")
		if firstErr != nil {
			return fmt.Errorf("daemon: %s: %w", firstErrName, firstErr)
		}
		return nil

	case <-finished:
		// All components finished within the bound.
		log.Warn("shutdown complete", slog.Int64(logging.AttrDuration, time.Since(shutdownStart).Milliseconds()))
		if firstErr != nil {
			return fmt.Errorf("daemon: %s: %w", firstErrName, firstErr)
		}
		return nil
	}
}
