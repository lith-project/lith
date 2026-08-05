package watch

import "context"

// noopWatcher is a Watcher that emits nothing and stops on context cancellation.
type noopWatcher struct {
	events chan Event
	gaps   chan GapReason
}

// NewNoop returns a Watcher that emits nothing and stops on context
// cancellation. Used when watcher.enabled is false.
func NewNoop() Watcher {
	return &noopWatcher{
		events: make(chan Event),
		gaps:   make(chan GapReason),
	}
}

func (w *noopWatcher) Events() <-chan Event   { return w.events }
func (w *noopWatcher) Gaps() <-chan GapReason { return w.gaps }

// Start blocks until ctx is cancelled, then closes both channels.
func (w *noopWatcher) Start(ctx context.Context) error {
	defer close(w.events)
	defer close(w.gaps)
	<-ctx.Done()
	return nil
}
