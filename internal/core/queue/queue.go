// Package queue provides a bounded, in-memory FIFO queue of settled events.
//
// The queue is deliberately not durable: it is not persisted to disk, cannot be
// snapshotted, and is lost on process restart. This is a conscious design
// choice — events that matter will be re-emitted by the watcher after restart.
package queue

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lith-project/lith/internal/core/watch"
)

// Queue is a bounded, in-memory queue of settled events.
// It is deliberately not durable: see the package docs.
type Queue struct {
	ch  chan watch.Event
	log *slog.Logger
}

// New builds a Queue with the given hard capacity, which must be positive.
// Returns an error if capacity is not positive.
func New(capacity int, log *slog.Logger) (*Queue, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("queue: capacity must be positive, got %d", capacity)
	}
	return &Queue{
		ch:  make(chan watch.Event, capacity),
		log: log,
	}, nil
}

// Push offers e to the queue, returning false when it was shed.
// Push never blocks; when the queue is at capacity the event is dropped.
func (q *Queue) Push(e watch.Event) bool {
	select {
	case q.ch <- e:
		return true
	default:
		return false
	}
}

// Pop removes and returns the oldest event, blocking until one is available
// or ctx is cancelled. Pop returns a wrapped ctx.Err() on cancellation and
// never a zero event with a nil error.
func (q *Queue) Pop(ctx context.Context) (watch.Event, error) {
	select {
	case e := <-q.ch:
		return e, nil
	case <-ctx.Done():
		return watch.Event{}, fmt.Errorf("queue: pop: %w", ctx.Err())
	}
}

// Depth reports the current number of queued events.
// Safe for concurrent use; intended for tests and logging only.
func (q *Queue) Depth() int {
	return len(q.ch)
}
