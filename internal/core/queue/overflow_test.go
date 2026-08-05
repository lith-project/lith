package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/lith-project/lith/internal/core/logging"
	"github.com/lith-project/lith/internal/core/vaultpath"
	"github.com/lith-project/lith/internal/core/watch"
)

// makeEvent creates a watch.Event with the given path name and operation.
func makeEvent(name string, op watch.Op) watch.Event {
	p, err := vaultpath.New("/vault", "/vault/"+name)
	if err != nil {
		panic("test: creating path: " + err.Error())
	}
	return watch.Event{Path: p, Op: op}
}

// --- OverflowPolicy String ---

func TestOverflowPolicy_String(t *testing.T) {
	tests := []struct {
		p    OverflowPolicy
		want string
	}{
		{CoalesceByPath, "coalesce_by_path"},
		{ShedOldest, "shed_oldest"},
		{OverflowPolicy(255), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.p.String(); got != tt.want {
			t.Errorf("OverflowPolicy(%d).String() = %q, want %q", tt.p, got, tt.want)
		}
	}
}

// --- NewWithPolicy ---

func TestNewWithPolicy_PositiveCapacity(t *testing.T) {
	for _, p := range []OverflowPolicy{CoalesceByPath, ShedOldest} {
		q, err := NewWithPolicy(10, p, testLogger)
		if err != nil {
			t.Fatalf("policy %v: unexpected error: %v", p, err)
		}
		if q == nil {
			t.Fatalf("policy %v: expected non-nil queue", p)
		}
		if q.Depth() != 0 {
			t.Fatalf("policy %v: expected depth 0, got %d", p, q.Depth())
		}
		if q.Policy() != p {
			t.Fatalf("policy %v: Policy() = %v", p, q.Policy())
		}
	}
}

func TestNewWithPolicy_ZeroCapacity(t *testing.T) {
	_, err := NewWithPolicy(0, CoalesceByPath, testLogger)
	if err == nil {
		t.Fatal("expected error for zero capacity")
	}
}

func TestNewWithPolicy_NegativeCapacity(t *testing.T) {
	_, err := NewWithPolicy(-5, ShedOldest, testLogger)
	if err == nil {
		t.Fatal("expected error for negative capacity")
	}
}

func TestNew_DelegatesToNewWithPolicy(t *testing.T) {
	q, err := New(5, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Policy() != CoalesceByPath {
		t.Fatalf("New() policy = %v, want CoalesceByPath", q.Policy())
	}
}

// --- CoalesceByPath ---

func TestCoalesceByPath_DepthConstant(t *testing.T) {
	q, err := NewWithPolicy(3, CoalesceByPath, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Fill the queue.
	if !q.Push(makeEvent("a.md", watch.OpWrite)) {
		t.Fatal("first push should succeed")
	}
	if !q.Push(makeEvent("b.md", watch.OpWrite)) {
		t.Fatal("second push should succeed")
	}
	if !q.Push(makeEvent("c.md", watch.OpWrite)) {
		t.Fatal("third push should succeed")
	}

	if q.Depth() != 3 {
		t.Fatalf("expected depth 3, got %d", q.Depth())
	}

	// Push duplicates of the same path — should coalesce, not grow.
	for i := range 10 {
		e := makeEvent("a.md", watch.OpWrite)
		if !q.Push(e) {
			t.Fatalf("coalesce push %d should succeed", i)
		}
		if q.Depth() != 3 {
			t.Fatalf("after coalesce %d: expected depth 3, got %d", i, q.Depth())
		}
	}
}

func TestCoalesceByPath_RetainsNewestEvent(t *testing.T) {
	q, err := NewWithPolicy(3, CoalesceByPath, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q.Push(makeEvent("a.md", watch.OpCreate))
	q.Push(makeEvent("b.md", watch.OpWrite))
	q.Push(makeEvent("c.md", watch.OpRemove))

	// Replace a.md with a Write op.
	q.Push(makeEvent("a.md", watch.OpWrite))

	// Pop a.md — should be the newest version (Write, not Create).
	got, err := q.Pop(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Op != watch.OpWrite {
		t.Fatalf("expected Op %v (newest), got %v", watch.OpWrite, got.Op)
	}
}

func TestCoalesceByPath_PreservesQueuePosition(t *testing.T) {
	q, err := NewWithPolicy(5, CoalesceByPath, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Enqueue: a, b, c, d
	q.Push(makeEvent("a.md", watch.OpCreate))
	q.Push(makeEvent("b.md", watch.OpCreate))
	q.Push(makeEvent("c.md", watch.OpCreate))
	q.Push(makeEvent("d.md", watch.OpCreate))

	// Coalesce b.md (at index 1) — it should stay at index 1.
	q.Push(makeEvent("b.md", watch.OpWrite))

	// Pop all and verify order: a, b(newest), c, d
	want := []struct {
		name string
		op   watch.Op
	}{
		{"a.md", watch.OpCreate},
		{"b.md", watch.OpWrite},
		{"c.md", watch.OpCreate},
		{"d.md", watch.OpCreate},
	}

	for i, w := range want {
		got, err := q.Pop(context.Background())
		if err != nil {
			t.Fatalf("pop %d: unexpected error: %v", i, err)
		}
		if got.Path.ID() != w.name {
			t.Fatalf("pop %d: expected path %q, got %q", i, w.name, got.Path.ID())
		}
		if got.Op != w.op {
			t.Fatalf("pop %d: expected op %v, got %v", i, w.op, got.Op)
		}
	}
}

func TestCoalesceByPath_NewPathAtCapacity_Sheds(t *testing.T) {
	q, err := NewWithPolicy(2, CoalesceByPath, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q.Push(makeEvent("a.md", watch.OpWrite))
	q.Push(makeEvent("b.md", watch.OpWrite))

	// c.md is not in the queue and queue is full — should be shed.
	if q.Push(makeEvent("c.md", watch.OpWrite)) {
		t.Fatal("expected push to fail for new path at capacity")
	}
	if q.Depth() != 2 {
		t.Fatalf("expected depth 2, got %d", q.Depth())
	}
}

// --- ShedOldest ---

func TestShedOldest_DropsOldest(t *testing.T) {
	q, err := NewWithPolicy(3, ShedOldest, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q.Push(makeEvent("a.md", watch.OpCreate))
	q.Push(makeEvent("b.md", watch.OpCreate))
	q.Push(makeEvent("c.md", watch.OpCreate))

	// Push d.md — oldest (a.md) should be shed.
	if !q.Push(makeEvent("d.md", watch.OpCreate)) {
		t.Fatal("ShedOldest push should succeed")
	}

	if q.Depth() != 3 {
		t.Fatalf("expected depth 3, got %d", q.Depth())
	}

	// Pop should return b, c, d (a was shed).
	want := []string{"b.md", "c.md", "d.md"}
	for i, w := range want {
		got, err := q.Pop(context.Background())
		if err != nil {
			t.Fatalf("pop %d: unexpected error: %v", i, err)
		}
		if got.Path.ID() != w {
			t.Fatalf("pop %d: expected %q, got %q", i, w, got.Path.ID())
		}
	}
}

func TestShedOldest_FullQueue_EmptyPath(t *testing.T) {
	q, err := NewWithPolicy(2, ShedOldest, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q.Push(makeEvent("a.md", watch.OpWrite))
	q.Push(makeEvent("b.md", watch.OpWrite))

	// Push c.md — a.md should be shed.
	if !q.Push(makeEvent("c.md", watch.OpWrite)) {
		t.Fatal("ShedOldest push should succeed")
	}

	got, err := q.Pop(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Path.ID() != "b.md" {
		t.Fatalf("expected b.md, got %q", got.Path.ID())
	}
}

// --- Depth never exceeds capacity ---

func TestDepth_NeverExceedsCapacity(t *testing.T) {
	cap := 3
	for _, p := range []OverflowPolicy{CoalesceByPath, ShedOldest} {
		q, err := NewWithPolicy(cap, p, testLogger)
		if err != nil {
			t.Fatalf("policy %v: unexpected error: %v", p, err)
		}

		// Push far more events than capacity.
		for i := range 100 {
			q.Push(makeEvent("test.md", watch.OpWrite))
			if q.Depth() > cap {
				t.Fatalf("policy %v: push %d: depth %d exceeds capacity %d", p, i, q.Depth(), cap)
			}
		}

		// Push different paths.
		for i := range 100 {
			name := "other" + string(rune('a'+i%26)) + ".md"
			q.Push(makeEvent(name, watch.OpWrite))
			if q.Depth() > cap {
				t.Fatalf("policy %v: push %d: depth %d exceeds capacity %d", p, i, q.Depth(), cap)
			}
		}
	}
}

// --- Overflow logging ---

func TestShedOldest_EmitsLog(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	q, err := NewWithPolicy(2, ShedOldest, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q.Push(makeEvent("a.md", watch.OpWrite))
	q.Push(makeEvent("b.md", watch.OpWrite))
	q.Push(makeEvent("c.md", watch.OpWrite)) // triggers shed

	// Wait for the rate limiter to flush.
	q.limiter.waitFlush()

	output := buf.String()
	if output == "" {
		t.Fatal("expected shed log output, got none")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}

	if parsed["msg"] != logging.EventQueueOverflow {
		t.Errorf("msg = %q, want %q", parsed["msg"], logging.EventQueueOverflow)
	}
	if parsed[logging.AttrPolicy] != "shed_oldest" {
		t.Errorf("policy = %q, want %q", parsed[logging.AttrPolicy], "shed_oldest")
	}
	if parsed[logging.AttrCount] != float64(1) {
		t.Errorf("count = %v, want 1", parsed[logging.AttrCount])
	}
}

func TestCoalesceByPath_ShedNewPath_EmitsLog(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	q, err := NewWithPolicy(1, CoalesceByPath, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q.Push(makeEvent("a.md", watch.OpWrite))     // fills queue
	q.Push(makeEvent("b.md", watch.OpWrite))     // different path at capacity — shed

	// Wait for flush.
	q.limiter.waitFlush()

	output := buf.String()
	if output == "" {
		t.Fatal("expected shed log output, got none")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}

	if parsed["msg"] != logging.EventQueueOverflow {
		t.Errorf("msg = %q, want %q", parsed["msg"], logging.EventQueueOverflow)
	}
	if parsed[logging.AttrPolicy] != "coalesce_by_path" {
		t.Errorf("policy = %q, want %q", parsed[logging.AttrPolicy], "coalesce_by_path")
	}
}

func TestShedLog_NeverUsesWatcherGap(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	q, err := NewWithPolicy(1, ShedOldest, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q.Push(makeEvent("a.md", watch.OpWrite))
	q.Push(makeEvent("b.md", watch.OpWrite)) // shed

	q.limiter.waitFlush()

	output := buf.String()
	if output == "" {
		t.Fatal("expected shed log output, got none")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}

	if parsed["msg"] == logging.EventWatcherGap {
		t.Error("shed log must not use watcher.gap event name")
	}
}

// --- Rate-limited logging ---

func TestShedLog_RateLimited(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	q, err := NewWithPolicy(1, ShedOldest, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Fill the queue.
	q.Push(makeEvent("a.md", watch.OpWrite))

	// Shed many events in rapid succession — use unique paths to avoid coalescing.
	for i := range 20 {
		q.Push(makeEvent("other"+string(rune('A'+i%26))+".md", watch.OpWrite))
	}

	// Wait for flush.
	q.limiter.waitFlush()

	output := buf.String()
	if output == "" {
		t.Fatal("expected shed log output, got none")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}

	// The count should be 20 (all sheds), not 20 separate log lines.
	count, ok := parsed[logging.AttrCount].(float64)
	if !ok {
		t.Fatalf("count is not a number: %v", parsed[logging.AttrCount])
	}
	if count != 20 {
		t.Errorf("count = %v, want 20", count)
	}

	// Verify only one log line was emitted (one JSON object).
	dec := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	numLines := 0
	for dec.More() {
		numLines++
		dec.Decode(&struct{}{})
	}
	if numLines != 1 {
		t.Errorf("expected 1 log line, got %d", numLines)
	}
}

// --- Concurrent safety ---

func TestConcurrent_Push_Pop(t *testing.T) {
	q, err := NewWithPolicy(100, CoalesceByPath, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const (
		numProducers = 10
		numConsumers = 10
		eventsPer    = 500
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	var pushedCount int64
	var mu sync.Mutex

	// Producers — each uses a unique path to avoid coalescing.
	for i := range numProducers {
		wg.Add(1)
		go func(prodID int) {
			defer wg.Done()
			local := 0
			for j := range eventsPer {
			// Unique path: producer ID * eventsPer + sequence number.
			seq := prodID*eventsPer + j
			name := fmt.Sprintf("evt_%d.md", seq)
				if q.Push(makeEvent(name, watch.OpWrite)) {
					local++
				}
			}
			mu.Lock()
			pushedCount += int64(local)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Consume everything.
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

func TestConcurrent_ShedOldest(t *testing.T) {
	q, err := NewWithPolicy(10, ShedOldest, testLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const (
		numProducers = 5
		eventsPer    = 200
	)

	var wg sync.WaitGroup
	var pushedCount int64
	var mu sync.Mutex

	for i := range numProducers {
		wg.Add(1)
		go func(prodID int) {
			defer wg.Done()
			local := 0
			for j := range eventsPer {
				name := "prod" + string(rune('a'+prodID%26)) + "_event" + string(rune('0'+j%10)) + ".md"
				if q.Push(makeEvent(name, watch.OpWrite)) {
					local++
				}
			}
			mu.Lock()
			pushedCount += int64(local)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	if q.Depth() > 10 {
		t.Fatalf("depth %d exceeds capacity 10", q.Depth())
	}
	if q.Depth() != 10 {
		t.Fatalf("expected depth 10 after concurrent pushes, got %d", q.Depth())
	}

	// Drain.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var consumedCount int64
	var consumerWg sync.WaitGroup

	for range 5 {
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

	time.Sleep(100 * time.Millisecond)
	cancel()
	consumerWg.Wait()

	mu.Lock()
	finalConsumed := consumedCount
	queueDepth := q.Depth()
	mu.Unlock()

	if queueDepth != 0 {
		t.Fatalf("expected all events consumed, depth=%d (consumed=%d)", queueDepth, finalConsumed)
	}
	// Some events were shed, so consumed < pushed * eventsPer.
	// We just verify that exactly 10 were consumed (capacity).
	if finalConsumed != 10 {
		t.Fatalf("expected exactly 10 consumed (capacity), got %d", finalConsumed)
	}
}

// --- Existing tests regression (channel-based tests still pass) ---

func TestPush_Pop_FIFO_WithPolicy(t *testing.T) {
	for _, p := range []OverflowPolicy{CoalesceByPath, ShedOldest} {
		q, err := NewWithPolicy(3, p, testLogger)
		if err != nil {
			t.Fatalf("policy %v: unexpected error: %v", p, err)
		}

		events := []watch.Event{
			makeEvent("a.md", watch.OpWrite),
			makeEvent("b.md", watch.OpWrite),
			makeEvent("c.md", watch.OpWrite),
		}

		for _, e := range events {
			if !q.Push(e) {
				t.Fatalf("policy %v: expected push to succeed for %s", p, e.Path)
			}
		}

		for _, want := range events {
			got, err := q.Pop(context.Background())
			if err != nil {
				t.Fatalf("policy %v: unexpected pop error: %v", p, err)
			}
			if got != want {
				t.Fatalf("policy %v: expected %v, got %v", p, want, got)
			}
		}
	}
}

func TestPop_ContextCancelled_WithPolicy(t *testing.T) {
	for _, p := range []OverflowPolicy{CoalesceByPath, ShedOldest} {
		q, err := NewWithPolicy(5, p, testLogger)
		if err != nil {
			t.Fatalf("policy %v: unexpected error: %v", p, err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		got, err := q.Pop(ctx)
		if err == nil {
			t.Fatalf("policy %v: expected error on cancelled context", p)
		}
		if got != (watch.Event{}) {
			t.Fatalf("policy %v: expected zero event, got %v", p, got)
		}
	}
}

func TestPop_Blocks_WithPolicy(t *testing.T) {
	for _, p := range []OverflowPolicy{CoalesceByPath, ShedOldest} {
		q, err := NewWithPolicy(5, p, testLogger)
		if err != nil {
			t.Fatalf("policy %v: unexpected error: %v", p, err)
		}

		done := make(chan struct{})
		var got watch.Event

		go func() {
			defer close(done)
			got, _ = q.Pop(context.Background())
		}()

		time.Sleep(50 * time.Millisecond)

		want := makeEvent("delayed.md", watch.OpWrite)
		q.Push(want)

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatalf("policy %v: Pop did not return after push", p)
		}

		if got != want {
			t.Fatalf("policy %v: expected %v, got %v", p, want, got)
		}
	}
}
