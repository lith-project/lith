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
	"sync"
	"time"

	"github.com/lith-project/lith/internal/core/watch"
)

const (
	// shedLogInterval is the minimum interval between summary shed logs.
	shedLogInterval = 5 * time.Second
)

// Queue is a bounded, in-memory queue of settled events.
// It is deliberately not durable: see the package docs.
type Queue struct {
	mu       sync.Mutex
	items    []watch.Event
	pending  map[string]int // path ID → index in items
	capacity int
	policy   OverflowPolicy
	log      *slog.Logger
	notify   chan struct{} // buffered, cap 1
	limiter  *rateLimiter
}

// New builds a Queue with the given hard capacity, which must be positive.
// The default overflow policy is CoalesceByPath.
// Returns an error if capacity is not positive.
func New(capacity int, log *slog.Logger) (*Queue, error) {
	return NewWithPolicy(capacity, CoalesceByPath, log)
}

// NewWithPolicy builds a Queue with the given hard capacity and overflow policy.
// Returns an error if capacity is not positive.
func NewWithPolicy(capacity int, p OverflowPolicy, log *slog.Logger) (*Queue, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("queue: capacity must be positive, got %d", capacity)
	}
	return &Queue{
		items:    make([]watch.Event, 0, capacity),
		pending:  make(map[string]int),
		capacity: capacity,
		policy:   p,
		log:      log,
		notify:   make(chan struct{}, 1),
		limiter:  newRateLimiter(log, shedLogInterval),
	}, nil
}

// Push offers e to the queue, returning false when it was shed.
// Push never blocks; when the queue is at capacity the overflow policy applies.
func (q *Queue) Push(e watch.Event) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	pathID := e.Path.ID()

	// Check if this path is already in the queue (coalescing).
	if idx, ok := q.pending[pathID]; ok {
		q.items[idx] = e // replace in-place, preserving position
		return true
	}

	// Path not in queue — check capacity.
	if len(q.items) < q.capacity {
		q.enqueue(e)
		return true
	}

	// Queue is full and path is not present — apply overflow policy.
	switch q.policy {
	case ShedOldest:
		q.shedOldest(e)
		return true
	default:
		// CoalesceByPath (or unknown): shed the new event.
		q.drop(e)
		return false
	}
}

// Pop removes and returns the oldest event, blocking until one is available
// or ctx is cancelled. Pop returns a wrapped ctx.Err() on cancellation and
// never a zero event with a nil error.
func (q *Queue) Pop(ctx context.Context) (watch.Event, error) {
	for {
		q.mu.Lock()
		if len(q.items) > 0 {
			e := q.items[0]
			q.items = q.items[1:]
			delete(q.pending, e.Path.ID())
			q.shiftPending(-1)
			q.mu.Unlock()
			return e, nil
		}
		q.mu.Unlock()

		select {
		case <-q.notify:
			// Something might have been pushed; loop back to check.
		case <-ctx.Done():
			return watch.Event{}, fmt.Errorf("queue: pop: %w", ctx.Err())
		}
	}
}

// Depth reports the current number of queued events.
// Safe for concurrent use; intended for tests and logging only.
func (q *Queue) Depth() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// Policy returns the queue's overflow policy.
func (q *Queue) Policy() OverflowPolicy {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.policy
}

// enqueue adds an event at the end. Caller must hold q.mu.
func (q *Queue) enqueue(e watch.Event) {
	q.items = append(q.items, e)
	q.pending[e.Path.ID()] = len(q.items) - 1
	q.signal()
}

// shedOldest drops the oldest event and appends the new one.
// Caller must hold q.mu and the queue must be at capacity.
func (q *Queue) shedOldest(e watch.Event) {
	old := q.items[0]
	delete(q.pending, old.Path.ID())
	q.items = q.items[1:]
	q.shiftPending(-1)

	q.items = append(q.items, e)
	q.pending[e.Path.ID()] = len(q.items) - 1

	q.limiter.onShed(old.Path.Raw(), q.policy.String())
	q.signal()
}

// drop records a shed event for logging without adding it to the queue.
// Caller must hold q.mu.
func (q *Queue) drop(e watch.Event) {
	q.limiter.onShed(e.Path.Raw(), q.policy.String())
}

// shiftPending adjusts all pending indices by delta. Caller must hold q.mu.
func (q *Queue) shiftPending(delta int) {
	for k := range q.pending {
		q.pending[k] += delta
	}
}

// signal sends a non-blocking notification that an item is available.
func (q *Queue) signal() {
	select {
	case q.notify <- struct{}{}:
	default:
	}
}
