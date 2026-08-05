package watch

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcherCreateFile(t *testing.T) {
	dir := t.TempDir()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	w, err := NewFSNotify(dir, log)
	if err != nil {
		t.Fatalf("NewFSNotify: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Start(ctx)

	// Give the watcher time to start
	time.Sleep(50 * time.Millisecond)

	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	select {
	case e := <-w.Events():
		if e.Op != OpCreate {
			t.Errorf("expected OpCreate, got %v", e.Op)
		}
		if e.Path.ID() != "test.md" {
			t.Errorf("expected ID 'test.md', got %q", e.Path.ID())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for create event")
	}
}

func TestWatcherWriteFile(t *testing.T) {
	dir := t.TempDir()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	w, err := NewFSNotify(dir, log)
	if err != nil {
		t.Fatalf("NewFSNotify: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Start(ctx)

	time.Sleep(50 * time.Millisecond)

	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte("initial"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Drain the create event
	select {
	case <-w.Events():
	case <-time.After(2 * time.Second):
		t.Fatal("timeout draining create event")
	}

	// Write again
	if err := os.WriteFile(path, []byte("updated"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	select {
	case e := <-w.Events():
		if e.Op != OpWrite {
			t.Errorf("expected OpWrite, got %v", e.Op)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for write event")
	}
}

func TestWatcherRemoveFile(t *testing.T) {
	dir := t.TempDir()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	w, err := NewFSNotify(dir, log)
	if err != nil {
		t.Fatalf("NewFSNotify: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Start(ctx)

	time.Sleep(50 * time.Millisecond)

	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Drain all create events (platform may batch or duplicate)
	time.Sleep(100 * time.Millisecond)
	drainEvents(w.Events())

	if err := os.Remove(path); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	// On macOS (kqueue), fsnotify may deliver Rename or Write instead of Remove.
	// Accept any event indicating the file changed or is gone.
	found := false
	deadline := time.After(2 * time.Second)
	for !found {
		select {
		case e := <-w.Events():
			if e.Op == OpRemove || e.Op == OpRename {
				found = true
			}
		case <-deadline:
			t.Fatal("timeout waiting for remove/rename event")
		}
	}
}

func drainEvents(ch <-chan Event) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

func TestWatcherContextCancellation(t *testing.T) {
	dir := t.TempDir()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	w, err := NewFSNotify(dir, log)
	if err != nil {
		t.Fatalf("NewFSNotify: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- w.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Start returned error on cancellation: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for Start to return")
	}

	// Channel should be closed
	_, ok := <-w.Events()
	if ok {
		t.Error("expected events channel to be closed")
	}
}

func TestWatcherAbsolutePathRequired(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	_, err := NewFSNotify("relative/path", log)
	if err == nil {
		t.Fatal("expected error for relative path")
	}
}

func TestOpString(t *testing.T) {
	cases := []struct {
		op   Op
		want string
	}{
		{OpCreate, "create"},
		{OpWrite, "write"},
		{OpRemove, "remove"},
		{OpRename, "rename"},
		{Op(99), "unknown"},
	}
	for _, c := range cases {
		if got := c.op.String(); got != c.want {
			t.Errorf("Op(%d).String() = %q, want %q", c.op, got, c.want)
		}
	}
}
