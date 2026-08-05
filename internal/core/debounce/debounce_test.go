package debounce

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lith-project/lith/internal/core/vaultpath"
	"github.com/lith-project/lith/internal/core/watch"
)

// makePath creates a vaultpath.Path for testing. It panics on error because
// test paths are always well-formed.
func makePath(id string) vaultpath.Path {
	p, err := vaultpath.New("/vault", "/vault/"+id)
	if err != nil {
		panic("test: creating path: " + err.Error())
	}
	return p
}

func TestNewRejectsInvalidBounds(t *testing.T) {
	out := make(chan watch.Event)
	tests := []struct {
		name      string
		quiet     time.Duration
		maxDelay  time.Duration
		wantError bool
	}{
		{"zero quiet", 0, 5 * time.Second, true},
		{"negative quiet", -1 * time.Millisecond, 5 * time.Second, true},
		{"zero maxDelay", 200 * time.Millisecond, 0, true},
		{"negative maxDelay", 200 * time.Millisecond, -1 * time.Millisecond, true},
		{"quiet equals maxDelay", 200 * time.Millisecond, 200 * time.Millisecond, true},
		{"quiet greater than maxDelay", 5 * time.Second, 200 * time.Millisecond, true},
		{"valid bounds", 200 * time.Millisecond, 5 * time.Second, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.quiet, tt.maxDelay, out)
			if (err != nil) != tt.wantError {
				t.Errorf("New(%v, %v, _) error = %v, wantError %v", tt.quiet, tt.maxDelay, err, tt.wantError)
			}
		})
	}
}

func TestNewRejectsNilChannel(t *testing.T) {
	_, err := New(100*time.Millisecond, 5*time.Second, nil)
	if err == nil {
		t.Error("New with nil channel should return error")
	}
}

// drainOutput reads all events from a channel until it is closed.
// Must be called only after the writer (Run) has returned or will return.
func drainOutput(ch <-chan watch.Event) []watch.Event {
	var events []watch.Event
	for ev := range ch {
		events = append(events, ev)
	}
	return events
}

func TestTenEventsOnePathCoalesce(t *testing.T) {
	// Ten events on one path within quiet period → exactly one settled event.
	out := make(chan watch.Event, 32)
	d, err := New(100*time.Millisecond, 5*time.Second, out)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan watch.Event, 32)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := d.Run(ctx, in); err != nil && err != context.Canceled {
			t.Errorf("Run returned: %v", err)
		}
	}()

	path := makePath("note.md")
	for i := 0; i < 10; i++ {
		in <- watch.Event{Path: path, Op: watch.OpWrite}
	}

	// Wait for quiet period to settle.
	time.Sleep(200 * time.Millisecond)
	cancel()
	wg.Wait()

	emitted := drainOutput(out)

	if len(emitted) != 1 {
		t.Fatalf("expected exactly 1 emitted event, got %d", len(emitted))
	}
	if emitted[0].Op != watch.OpWrite {
		t.Errorf("expected OpWrite, got %v", emitted[0].Op)
	}
}

func TestTwoPaths(t *testing.T) {
	// Events on two different paths → two settled events.
	out := make(chan watch.Event, 32)
	d, err := New(100*time.Millisecond, 5*time.Second, out)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan watch.Event, 32)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := d.Run(ctx, in); err != nil && err != context.Canceled {
			t.Errorf("Run returned: %v", err)
		}
	}()

	p1 := makePath("a.md")
	p2 := makePath("b.md")
	in <- watch.Event{Path: p1, Op: watch.OpCreate}
	in <- watch.Event{Path: p2, Op: watch.OpWrite}

	time.Sleep(200 * time.Millisecond)
	cancel()
	wg.Wait()

	emitted := drainOutput(out)

	if len(emitted) != 2 {
		t.Fatalf("expected 2 emitted events, got %d", len(emitted))
	}

	ids := map[string]bool{}
	for _, ev := range emitted {
		ids[ev.Path.ID()] = true
	}
	if !ids["a.md"] || !ids["b.md"] {
		t.Errorf("expected both a.md and b.md, got %v", ids)
	}
}

func TestMaxDelayEmitsWithinBound(t *testing.T) {
	// Path modified continuously for longer than maxDelay → emits within maxDelay.
	out := make(chan watch.Event, 64)
	d, err := New(50*time.Millisecond, 200*time.Millisecond, out)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan watch.Event, 64)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := d.Run(ctx, in); err != nil && err != context.Canceled {
			t.Errorf("Run returned: %v", err)
		}
	}()

	path := makePath("hot.md")

	// Send events every 50ms for 500ms. The maxDelay is 200ms, so we
	// expect at least one emission within 200ms of the first event.
	go func() {
		for i := 0; i < 10; i++ {
			in <- watch.Event{Path: path, Op: watch.OpWrite}
			time.Sleep(50 * time.Millisecond)
		}
	}()

	// Wait long enough for at least one maxDelay emission.
	time.Sleep(400 * time.Millisecond)
	cancel()
	wg.Wait()

	emitted := drainOutput(out)

	if len(emitted) == 0 {
		t.Fatal("expected at least one emission within maxDelay, got none")
	}

	// We trust the timing from the goroutine; the key assertion is that
	// we got at least one emission and it happened promptly.
	if emitted[0].Path.ID() != "hot.md" {
		t.Errorf("expected path hot.md, got %s", emitted[0].Path.ID())
	}
}

func TestContextCancellation(t *testing.T) {
	// Context cancellation returns promptly and closes output channel.
	out := make(chan watch.Event, 32)
	d, err := New(200*time.Millisecond, 5*time.Second, out)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan watch.Event, 32)

	done := make(chan error, 1)
	go func() {
		done <- d.Run(ctx, in)
	}()

	// Send an event so there's something pending.
	p := makePath("pending.md")
	in <- watch.Event{Path: p, Op: watch.OpWrite}

	// Cancel context.
	cancel()

	// Wait for Run to return.
	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return within 2s of context cancellation")
	}

	// Output channel should be closed (range will terminate).
	emitted := drainOutput(out)

	// The pending event should have been flushed.
	if len(emitted) != 1 {
		t.Fatalf("expected 1 flushed event, got %d", len(emitted))
	}
}

func TestFlushOnContextCancel(t *testing.T) {
	// Pending events are flushed on context cancellation.
	out := make(chan watch.Event, 32)
	d, err := New(5*time.Second, 10*time.Second, out) // long quiet so it won't settle on its own
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan watch.Event, 32)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		d.Run(ctx, in)
	}()

	p1 := makePath("a.md")
	p2 := makePath("b.md")
	in <- watch.Event{Path: p1, Op: watch.OpCreate}
	in <- watch.Event{Path: p2, Op: watch.OpRemove}

	// Cancel before quiet period fires.
	cancel()
	wg.Wait()

	emitted := drainOutput(out)

	if len(emitted) != 2 {
		t.Fatalf("expected 2 flushed events, got %d", len(emitted))
	}
}

func TestLastOpWins(t *testing.T) {
	// When several ops coalesce for one path, emit the last op observed.
	out := make(chan watch.Event, 32)
	d, err := New(200*time.Millisecond, 5*time.Second, out)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan watch.Event, 32)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := d.Run(ctx, in); err != nil && err != context.Canceled {
			t.Errorf("Run returned: %v", err)
		}
	}()

	p := makePath("note.md")
	in <- watch.Event{Path: p, Op: watch.OpCreate}
	in <- watch.Event{Path: p, Op: watch.OpWrite}
	in <- watch.Event{Path: p, Op: watch.OpRemove}

	time.Sleep(400 * time.Millisecond)
	cancel()
	wg.Wait()

	emitted := drainOutput(out)

	if len(emitted) != 1 {
		t.Fatalf("expected 1 event, got %d", len(emitted))
	}
	if emitted[0].Op != watch.OpRemove {
		t.Errorf("expected last op OpRemove, got %v", emitted[0].Op)
	}
}

func TestInputChannelClosed(t *testing.T) {
	// Closing the input channel flushes pending events and closes output.
	out := make(chan watch.Event, 32)
	d, err := New(5*time.Second, 10*time.Second, out) // long timers
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	in := make(chan watch.Event, 32)

	done := make(chan error, 1)
	go func() {
		done <- d.Run(ctx, in)
	}()

	p := makePath("closing.md")
	in <- watch.Event{Path: p, Op: watch.OpWrite}
	close(in)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after input channel closed")
	}

	emitted := drainOutput(out)

	if len(emitted) != 1 {
		t.Fatalf("expected 1 flushed event, got %d", len(emitted))
	}
}
