package watch

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestGapWatchError(t *testing.T) {
	err := fmt.Errorf("some error")
	reason := classifyGap(err)
	if reason != GapWatchError {
		t.Errorf("expected GapWatchError, got %q", reason)
	}
}

func TestGapPlatformLimit(t *testing.T) {
	err := fmt.Errorf("inotify: watch limit reached")
	reason := classifyGap(err)
	if reason != GapPlatformLimit {
		t.Errorf("expected GapPlatformLimit, got %q", reason)
	}
}

func TestGapReasonsUnique(t *testing.T) {
	cases := []struct {
		reason GapReason
		want   string
	}{
		{GapQueueOverflow, "queue_overflow"},
		{GapWatchError, "watch_error"},
		{GapPlatformLimit, "platform_limit"},
	}
	for _, c := range cases {
		if got := string(c.reason); got != c.want {
			t.Errorf("GapReason(%q).String() = %q, want %q", c.reason, got, c.want)
		}
	}
}

func TestGapNonBlocking(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, nil))
	w, err := NewFSNotify(dir, log)
	if err != nil {
		t.Fatalf("NewFSNotify: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = w.Start(ctx) }()

	time.Sleep(50 * time.Millisecond)

	// Write a file — event should still flow even though no one reads Gaps()
	path := dir + "/test.md"
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	select {
	case e := <-w.Events():
		if e.Op != OpCreate {
			t.Errorf("expected OpCreate, got %v", e.Op)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestGapClassifyVariants(t *testing.T) {
	cases := []struct {
		err    string
		reason GapReason
	}{
		{"inotify: watch limit", GapPlatformLimit},
		{"inotify: watch limit reached", GapPlatformLimit},
		{"something inotify watch limit something", GapPlatformLimit},
		{"generic error", GapWatchError},
		{"inotify but no limit", GapWatchError},
		{"something: watch limit exceeded", GapWatchError},
	}
	for _, c := range cases {
		reason := classifyGap(fmt.Errorf("%s", c.err))
		if reason != c.reason {
			t.Errorf("classifyGap(%q) = %q, want %q", c.err, reason, c.reason)
		}
	}
}
