package watch

import (
	"context"
	"testing"
	"time"
)

func TestNewNoopEventsClosed(t *testing.T) {
	w := NewNoop()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- w.Start(ctx) }()

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after context cancellation")
	}

	_, ok := <-w.Events()
	if ok {
		t.Fatal("Events channel not closed after stop")
	}
}

func TestNewNoopGapsClosed(t *testing.T) {
	w := NewNoop()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- w.Start(ctx) }()

	cancel()
	<-done

	_, ok := <-w.Gaps()
	if ok {
		t.Fatal("Gaps channel not closed after stop")
	}
}

func TestNewNoopNonNilChannels(t *testing.T) {
	w := NewNoop()
	if w.Events() == nil {
		t.Fatal("Events() returned nil channel")
	}
	if w.Gaps() == nil {
		t.Fatal("Gaps() returned nil channel")
	}
}
