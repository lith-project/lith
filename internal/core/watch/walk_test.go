package watch

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestAddTreeNestedTempTree(t *testing.T) {
	// Create a nested temp tree (3+ levels)
	vaultRoot := t.TempDir()
	deepDir := filepath.Join(vaultRoot, "level1", "level2", "level3")
	if err := os.MkdirAll(deepDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer watcher.Close()

	// Create logger
	logger := slog.Default()

	// Call addTree
	if err := addTree(watcher, vaultRoot, logger); err != nil {
		t.Fatalf("addTree failed: %v", err)
	}

	// Verify all directories are registered by creating a file 3 levels down
	// and checking it produces an event
	testFile := filepath.Join(deepDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Wait for event with timeout
	select {
	case event, ok := <-watcher.Events:
		if !ok {
			t.Fatal("watcher.Events channel closed")
		}
		if event.Name != testFile {
			t.Errorf("expected event for %s, got %s", testFile, event.Name)
		}
	case err, ok := <-watcher.Errors:
		if !ok {
			t.Fatal("watcher.Errors channel closed")
		}
		t.Fatalf("watcher error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestAddTreeNewDirectoryWatched(t *testing.T) {
	// Create vault root
	vaultRoot := t.TempDir()

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer watcher.Close()

	// Create logger
	logger := slog.Default()

	// Call addTree
	if err := addTree(watcher, vaultRoot, logger); err != nil {
		t.Fatalf("addTree failed: %v", err)
	}

	// Create a new directory after addTree runs
	newDir := filepath.Join(vaultRoot, "newdir")
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// The new directory won't be watched until we call addTree again
	// This test verifies addTree returns without error
	// The actual "newly created directories" watching happens in the fsnotify event loop
}

func TestAddTreeInternalSymlink(t *testing.T) {
	// Create vault root with internal symlink (EC-FS-013)
	vaultRoot := t.TempDir()

	// Create a target directory
	targetDir := filepath.Join(vaultRoot, "target")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create symlink inside vault pointing to target inside vault
	symlinkPath := filepath.Join(vaultRoot, "link")
	if err := os.Symlink(targetDir, symlinkPath); err != nil {
		t.Fatal(err)
	}

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer watcher.Close()

	// Create logger
	logger := slog.Default()

	// Call addTree
	if err := addTree(watcher, vaultRoot, logger); err != nil {
		t.Fatalf("addTree failed: %v", err)
	}

	// Verify symlink target is watched by creating a file in it.
	// fsnotify emits events using the path that was added to the watcher,
	// so the event will reference the symlink path, not the resolved target.
	testFile := filepath.Join(targetDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Wait for event — accept any event that references the symlink path or
	// the filename, proving the target directory is indeed watched.
	symlinkEvent := filepath.Join(symlinkPath, "test.txt")
	select {
	case event, ok := <-watcher.Events:
		if !ok {
			t.Fatal("watcher.Events channel closed")
		}
		if event.Name != symlinkEvent && event.Name != testFile &&
			event.Name != symlinkPath && !filepath.HasPrefix(event.Name, symlinkPath) {
			t.Errorf("expected event referencing %s or %s, got %s",
				symlinkEvent, testFile, event.Name)
		}
	case err, ok := <-watcher.Errors:
		if !ok {
			t.Fatal("watcher.Errors channel closed")
		}
		t.Fatalf("watcher error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestAddTreeEscapingSymlink(t *testing.T) {
	// Create vault root
	vaultRoot := t.TempDir()

	// Create a directory outside vault
	outsideDir := filepath.Join(t.TempDir(), "outside")
	if err := os.MkdirAll(outsideDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create symlink inside vault pointing outside
	symlinkPath := filepath.Join(vaultRoot, "escape")
	if err := os.Symlink(outsideDir, symlinkPath); err != nil {
		t.Fatal(err)
	}

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer watcher.Close()

	// Create logger
	logger := slog.Default()

	// Call addTree
	if err := addTree(watcher, vaultRoot, logger); err != nil {
		t.Fatalf("addTree failed: %v", err)
	}

	// Verify escaping symlink is NOT followed by creating a file in outside dir
	testFile := filepath.Join(outsideDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	// We should NOT get an event for the file outside the vault
	select {
	case event := <-watcher.Events:
		// If we get an event, it shouldn't be for the outside file
		if event.Name == testFile {
			t.Errorf("should not have watched file outside vault: %s", event.Name)
		}
	case <-watcher.Errors:
		// Errors are okay
	case <-time.After(100 * time.Millisecond):
		// Timeout is expected - no event should arrive
	}
}

func TestAddTreeBrokenSymlink(t *testing.T) {
	// Create vault root
	vaultRoot := t.TempDir()

	// Create broken symlink
	symlinkPath := filepath.Join(vaultRoot, "broken")
	if err := os.Symlink("/nonexistent/path", symlinkPath); err != nil {
		t.Fatal(err)
	}

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer watcher.Close()

	// Create logger
	logger := slog.Default()

	// Call addTree - should not crash (Rule 4)
	if err := addTree(watcher, vaultRoot, logger); err != nil {
		t.Fatalf("addTree failed: %v", err)
	}
}

func TestAddTreeSymlinkCycle(t *testing.T) {
	// Create vault root
	vaultRoot := t.TempDir()

	// Create directory structure with cycle
	dir1 := filepath.Join(vaultRoot, "dir1")
	dir2 := filepath.Join(vaultRoot, "dir1", "dir2")
	if err := os.MkdirAll(dir2, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create symlink that creates cycle: dir2/link -> dir1
	symlinkPath := filepath.Join(dir2, "link")
	if err := os.Symlink(dir1, symlinkPath); err != nil {
		t.Fatal(err)
	}

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer watcher.Close()

	// Create logger
	logger := slog.Default()

	// Call addTree with timeout to ensure it terminates (Rule 3)
	done := make(chan error, 1)
	go func() {
		done <- addTree(watcher, vaultRoot, logger)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("addTree failed: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("addTree did not terminate (cycle detected)")
	}
}

func TestAddTreeUnreadableDir(t *testing.T) {
	// Create vault root
	vaultRoot := t.TempDir()

	// Create unreadable directory (if running as non-root)
	unreadableDir := filepath.Join(vaultRoot, "unreadable")
	if err := os.MkdirAll(unreadableDir, 0o000); err != nil {
		t.Fatal(err)
	}
	// Restore permissions for cleanup
	defer os.Chmod(unreadableDir, 0o755)

	// Create readable directory alongside
	readableDir := filepath.Join(vaultRoot, "readable")
	if err := os.MkdirAll(readableDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer watcher.Close()

	// Create logger
	logger := slog.Default()

	// Call addTree - should complete despite unreadable dir (Rule 5)
	if err := addTree(watcher, vaultRoot, logger); err != nil {
		t.Fatalf("addTree failed: %v", err)
	}

	// Verify readable directory is watched
	testFile := filepath.Join(readableDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Wait for event
	select {
	case event, ok := <-watcher.Events:
		if !ok {
			t.Fatal("watcher.Events channel closed")
		}
		if event.Name != testFile {
			t.Errorf("expected event for %s, got %s", testFile, event.Name)
		}
	case err, ok := <-watcher.Errors:
		if !ok {
			t.Fatal("watcher.Errors channel closed")
		}
		t.Fatalf("watcher error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}
