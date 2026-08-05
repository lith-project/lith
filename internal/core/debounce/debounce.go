// Package debounce provides a per-path event coalescer that settles
// filesystem change events within configurable time bounds.
package debounce

import (
	"context"
	"fmt"
	"time"

	"github.com/lith-project/lith/internal/core/watch"
)

// timerKind distinguishes quiet timers from max-delay timers in the
// timer-event channel.
type timerKind int

const (
	kindQuiet timerKind = iota
	kindMax
)

// timerEvent is sent from AfterFunc callbacks to the main select loop.
type timerEvent struct {
	pathID string
	kind   timerKind
}

// pending tracks the coalescence state for a single path.
type pending struct {
	last   watch.Event
	quiet  *time.Timer // reset on each new event for this path
	maxOut *time.Timer // fires once at maxDelay from first event — never reset
}

// Debouncer coalesces events per path in time. It is safe to call Run only
// once; the Debouncer is not reusable after Run returns.
type Debouncer struct {
	quiet    time.Duration
	maxDelay time.Duration
	out      chan<- watch.Event
}

// New builds a Debouncer with the given bounds. Both must be positive and
// quiet must be strictly less than maxDelay.
func New(quiet, maxDelay time.Duration, out chan<- watch.Event) (*Debouncer, error) {
	if quiet <= 0 {
		return nil, fmt.Errorf("debounce: quiet must be positive, got %v", quiet)
	}
	if maxDelay <= 0 {
		return nil, fmt.Errorf("debounce: maxDelay must be positive, got %v", maxDelay)
	}
	if quiet >= maxDelay {
		return nil, fmt.Errorf("debounce: quiet (%v) must be strictly less than maxDelay (%v)", quiet, maxDelay)
	}
	if out == nil {
		return nil, fmt.Errorf("debounce: output channel must not be nil")
	}
	return &Debouncer{
		quiet:    quiet,
		maxDelay: maxDelay,
		out:      out,
	}, nil
}

// Run consumes in until it closes or ctx is cancelled, emitting settled
// events on the channel given to New. It closes that channel on return.
func (d *Debouncer) Run(ctx context.Context, in <-chan watch.Event) error {
	entries := make(map[string]*pending)
	timerCh := make(chan timerEvent, 16)

	defer func() {
		// Drain any remaining buffered input events so nothing is lost
		// when ctx fires between two buffered sends.
	drain:
		for {
			select {
			case ev, ok := <-in:
				if !ok {
					break drain
				}
				id := ev.Path.ID()
				if p, exists := entries[id]; exists {
					p.last = ev
				} else {
					entries[id] = &pending{last: ev}
				}
			default:
				break drain
			}
		}
		// Flush all pending events.
		for id, p := range entries {
			stopTimers(p)
			d.out <- p.last
			delete(entries, id)
		}
		close(d.out)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case ev, ok := <-in:
			if !ok {
				// Input channel closed — upstream is done.
				return nil
			}
			id := ev.Path.ID()
			if p, exists := entries[id]; exists {
				// Existing path: update last event (last-op-wins),
				// reset quiet timer, do NOT reset maxOut timer.
				p.last = ev
				if !p.quiet.Stop() {
					select {
					case <-p.quiet.C:
					default:
					}
				}
				p.quiet.Reset(d.quiet)
			} else {
				// New path: create pending entry, start both timers.
				p := &pending{last: ev}
				pathID := id // capture for closure
				p.quiet = time.AfterFunc(d.quiet, func() {
					timerCh <- timerEvent{pathID: pathID, kind: kindQuiet}
				})
				p.maxOut = time.AfterFunc(d.maxDelay, func() {
					timerCh <- timerEvent{pathID: pathID, kind: kindMax}
				})
				entries[id] = p
			}

		case te := <-timerCh:
			p, exists := entries[te.pathID]
			if !exists {
				// Timer fired for a path that was already cleaned up.
				continue
			}
			stopTimers(p)
			d.out <- p.last
			delete(entries, te.pathID)
		}
	}
}

// stopTimers stops both timers of a pending entry without blocking.
// It is a no-op for entries created without timers (e.g. during shutdown drain).
func stopTimers(p *pending) {
	if p.quiet != nil {
		if !p.quiet.Stop() {
			select {
			case <-p.quiet.C:
			default:
			}
		}
	}
	if p.maxOut != nil {
		if !p.maxOut.Stop() {
			select {
			case <-p.maxOut.C:
			default:
			}
		}
	}
}
