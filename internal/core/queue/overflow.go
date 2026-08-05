package queue

import (
	"log/slog"
	"sync"
	"time"

	"github.com/lith-project/lith/internal/core/logging"
)

// OverflowPolicy controls how the queue behaves when at capacity.
type OverflowPolicy uint8

const (
	// CoalesceByPath replaces an existing event for the same path,
	// keeping queue position. New paths at capacity are shed.
	CoalesceByPath OverflowPolicy = iota

	// ShedOldest drops the oldest event to make room for the new one.
	ShedOldest
)

// String returns the human-readable form of the policy.
func (p OverflowPolicy) String() string {
	switch p {
	case CoalesceByPath:
		return "coalesce_by_path"
	case ShedOldest:
		return "shed_oldest"
	default:
		return "unknown"
	}
}

// rateLimiter batches shed-event logs into a single summary per interval.
type rateLimiter struct {
	mu       sync.Mutex
	log      *slog.Logger
	count    int
	timer    *time.Timer
	interval time.Duration
	lastPath string
	policy   string
	done     chan struct{} // closed after each flush
}

func newRateLimiter(log *slog.Logger, interval time.Duration) *rateLimiter {
	return &rateLimiter{
		log:      log,
		interval: interval,
	}
}

// onShed records a shed event and starts the flush timer if not already running.
func (r *rateLimiter) onShed(path, policy string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.count++
	r.lastPath = path
	r.policy = policy

	if r.timer == nil {
		r.done = make(chan struct{})
		r.timer = time.AfterFunc(r.interval, r.flush)
	}
}

func (r *rateLimiter) flush() {
	r.mu.Lock()
	count := r.count
	r.count = 0
	r.timer = nil
	done := r.done
	lastPath := r.lastPath
	policy := r.policy
	r.mu.Unlock()

	if count > 0 {
		r.log.Warn(logging.EventQueueOverflow,
			logging.AttrPath, lastPath,
			logging.AttrCount, count,
			logging.AttrPolicy, policy,
		)
	}

	if done != nil {
		close(done)
	}
}

// waitFlush blocks until the most recent flush completes.
// Must be called after onShed; safe for tests only.
func (r *rateLimiter) waitFlush() {
	r.mu.Lock()
	done := r.done
	r.mu.Unlock()
	if done != nil {
		<-done
	}
}
