package queue

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func TestLoadTest(t *testing.T) {
	const (
		capacity      = 4096
		totalPushes   = 50_000
		numProducers  = 10
		eventsPerProd = totalPushes / numProducers
	)

	q, err := New(capacity, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var violations atomic.Int64

	// Monitor goroutine: periodically check Depth() ≤ capacity.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var monitorWg sync.WaitGroup
	monitorWg.Add(1)
	go func() {
		defer monitorWg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				depth := q.Depth()
				if depth > capacity {
					violations.Add(1)
				}
			}
		}
	}()

	// Consumer goroutine: pop events in background to make room.
	var consumed atomic.Int64
	var consumerWg sync.WaitGroup
	consumerWg.Add(1)
	go func() {
		defer consumerWg.Done()
		for {
			_, err := q.Pop(ctx)
			if err != nil {
				return
			}
			consumed.Add(1)
		}
	}()

	// Producers: each pushes unique events.
	var producerWg sync.WaitGroup
	for i := range numProducers {
		producerWg.Add(1)
		go func(prodID int) {
			defer producerWg.Done()
			for j := range eventsPerProd {
				seq := prodID*eventsPerProd + j
				q.Push(testEvent(fmt.Sprintf("load_%d.md", seq)))
			}
		}(i)
	}

	// Wait for all producers to finish.
	producerWg.Wait()

	// Cancel consumers and monitor.
	cancel()
	consumerWg.Wait()
	monitorWg.Wait()

	// Assert: Depth() never exceeded capacity at any observation point.
	if v := violations.Load(); v > 0 {
		t.Fatalf("Depth() exceeded capacity %d times (observed depth > %d)", v, capacity)
	}

	t.Logf("load test complete: pushed=%d, consumed=%d, capacity=%d, violations=%d",
		totalPushes, consumed.Load(), capacity, violations.Load())
}

func TestQueueStateIsIntentionallyLost(t *testing.T) {
	q, err := New(10, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Fill queue to capacity with unique events.
	for i := range 15 {
		q.Push(testEvent(fmt.Sprintf("fill_%d.md", i)))
	}

	if q.Depth() != 10 {
		t.Fatalf("expected depth 10, got %d", q.Depth())
	}

	// Queue state is intentionally not durable. On process restart, the
	// watcher re-emits all events. This test documents that design choice —
	// discarding the queue does not lose information that cannot be
	// reconstructed.
	q = nil // drop reference
	_ = q   // suppress unused variable

	// If Go GC collects q, no panic or error occurs — the queue is gone.
	// This is by design per RFC-0005/C-1.
}
