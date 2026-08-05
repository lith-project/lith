package queue

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/lith-project/lith/internal/core/vaultpath"
	"github.com/lith-project/lith/internal/core/watch"
)

var testLogger = slog.Default()

// makePath creates a vaultpath.Path for testing. It panics on error because
// test paths are always well-formed.
func makePath(id string) vaultpath.Path {
	p, err := vaultpath.New("/vault", "/vault/"+id)
	if err != nil {
		panic("test: creating path: " + err.Error())
	}
	return p
}

func testEvent(name string) watch.Event {
	return watch.Event{
		Path: makePath(name),
		Op:   watch.OpWrite,
	}
}

func TestNew_PositiveCapacity(t *testing.T) {
	q, err := New(10, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q == nil {
		t.Fatal("expected non-nil queue")
	}
	if q.Depth() != 0 {
		t.Fatalf("expected depth 0, got %d", q.Depth())
	}
}

func TestNew_ZeroCapacity(t *testing.T) {
	_, err := New(0, testLogger)
	if err == nil {
		t.Fatal("expected error for zero capacity")
	}
}

func TestNew_NegativeCapacity(t *testing.T) {
	_, err := New(-5, testLogger)
	if err == nil {
		t.Fatal("expected error for negative capacity")
	}
}

func TestPush_Pop_FIFO(t *testing.T) {
	q, err := New(3, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := []watch.Event{
		testEvent("a.md"),
		testEvent("b.md"),
		testEvent("c.md"),
	}

	for _, e := range events {
		if !q.Push(e) {
			t.Fatalf("expected push to succeed for %s", e.Path)
		}
	}

	for _, want := range events {
		got, err := q.Pop(context.Background())
		if err != nil {
			t.Fatalf("unexpected pop error: %v", err)
		}
		if got != want {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

func TestPush_AtCapacity_ReturnsFalse(t *testing.T) {
	q, err := New(2, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !q.Push(testEvent("a.md")) {
		t.Fatal("first push should succeed")
	}
	if !q.Push(testEvent("b.md")) {
		t.Fatal("second push should succeed")
	}
	if q.Push(testEvent("c.md")) {
		t.Fatal("third push should fail (queue at capacity)")
	}
	if q.Depth() != 2 {
		t.Fatalf("expected depth 2, got %d", q.Depth())
	}
}

func TestPop_Blocks(t *testing.T) {
	q, err := New(5, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	done := make(chan struct{})
	var got watch.Event

	go func() {
		defer close(done)
		got, err = q.Pop(context.Background())
	}()

	// Give the goroutine time to block.
	time.Sleep(50 * time.Millisecond)

	want := testEvent("delayed.md")
	if !q.Push(want) {
		t.Fatal("push should succeed")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Pop did not return after push")
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestPop_ContextCancelled(t *testing.T) {
	q, err := New(5, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	got, err := q.Pop(ctx)
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
	if got != (watch.Event{}) {
		t.Fatalf("expected zero event, got %v", got)
	}
}

func TestPop_ContextCancelledAfterDelay(t *testing.T) {
	q, err := New(5, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		_, err := q.Pop(ctx)
		done <- err
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error on cancelled context")
		}
	case <-time.After(time.Second):
		t.Fatal("Pop did not return after context cancellation")
	}
}

func TestDepth(t *testing.T) {
	q, err := New(5, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if q.Depth() != 0 {
		t.Fatalf("expected depth 0, got %d", q.Depth())
	}

	q.Push(testEvent("a.md"))
	q.Push(testEvent("b.md"))
	if q.Depth() != 2 {
		t.Fatalf("expected depth 2, got %d", q.Depth())
	}

	_, _ = q.Pop(context.Background())
	if q.Depth() != 1 {
		t.Fatalf("expected depth 1, got %d", q.Depth())
	}
}

func TestConcurrent_ProducersAndConsumers(t *testing.T) {
	q, err := New(100, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const (
		numProducers  = 10
		numConsumers  = 10
		eventsPerProd = 500
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Track how many events were successfully pushed.
	var pushedCount int64
	var mu sync.Mutex

	// Producers — each uses a unique path to avoid coalescing.
	for i := range numProducers {
		wg.Add(1)
		go func(prodID int) {
			defer wg.Done()
			localCount := 0
			for j := range eventsPerProd {
				// Unique path: producer ID * eventsPerProd + sequence number.
				seq := prodID*eventsPerProd + j
				e := watch.Event{
					Path: makePath(fmt.Sprintf("evt_%d.md", seq)),
					Op:   watch.OpWrite,
				}
				if q.Push(e) {
					localCount++
				}
			}
			mu.Lock()
			pushedCount += int64(localCount)
			mu.Unlock()
		}(i)
	}

	// Wait for all producers to finish pushing.
	wg.Wait()

	// Now consume everything that was pushed.
	var consumedCount int64
	var consumerWg sync.WaitGroup

	for range numConsumers {
		consumerWg.Add(1)
		go func() {
			defer consumerWg.Done()
			for {
				_, err := q.Pop(ctx)
				if err != nil {
					return
				}
				mu.Lock()
				consumedCount++
				mu.Unlock()
			}
		}()
	}

	// Give consumers a moment to drain, then cancel.
	time.Sleep(100 * time.Millisecond)
	cancel()

	consumerWg.Wait()

	mu.Lock()
	finalPushed := pushedCount
	finalConsumed := consumedCount
	queueDepth := q.Depth()
	mu.Unlock()

	if queueDepth != 0 {
		t.Fatalf("expected all events consumed, depth=%d (pushed=%d, consumed=%d)",
			queueDepth, finalPushed, finalConsumed)
	}
	if finalConsumed != finalPushed {
		t.Fatalf("consumed count mismatch: pushed=%d, consumed=%d", finalPushed, finalConsumed)
	}
}
