package watch

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/lith-project/lith/internal/core/logging"
)

// addTree registers dir and every subdirectory beneath it, refusing to
// traverse outside vaultRoot.
// (unexported — used by the fsnotify watcher)
func addTree(watcher *fsnotify.Watcher, vaultRoot string, log *slog.Logger) error {
	// Resolve the vault root to its real path. On macOS /var → /private/var,
	// and every containment check must use the resolved form so that
	// EvalSymlinks-produced targets match.
	resolvedRoot, err := filepath.EvalSymlinks(vaultRoot)
	if err != nil {
		return err
	}

	// Track visited directories by resolved real path to detect cycles
	visited := make(map[string]bool)

	// Use a queue for BFS traversal. We store the original (user-facing) path
	// so that watcher.Add and readDir receive the path the caller expects,
	// while containment and cycle checks use the resolved form.
	type entry struct {
		orig string // original path passed to the caller / watcher
		res  string // resolved real path
	}

	queue := []entry{{orig: vaultRoot, res: resolvedRoot}}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		// Check if we've already visited this resolved path (cycle detection)
		if visited[cur.res] {
			log.Debug("skipping already visited directory",
				slog.String(logging.AttrPath, cur.orig),
				slog.String("resolved", cur.res))
			continue
		}
		visited[cur.res] = true

		// Verify containment: resolved path must be within resolved vaultRoot
		if !containsPath(resolvedRoot, cur.res) {
			log.Debug("skipping path outside vault",
				slog.String(logging.AttrPath, cur.orig),
				slog.String("resolved", cur.res))
			continue
		}

		// Add this directory to the watcher
		if err := watcher.Add(cur.orig); err != nil {
			log.Debug("failed to add directory to watcher",
				slog.String(logging.AttrPath, cur.orig),
				slog.String(logging.AttrCause, err.Error()))
			continue
		}

		// Read directory entries
		entries, err := readDir(cur.orig)
		if err != nil {
			// Permission error or other issue — skip but continue walk (Rule 5)
			log.Debug("failed to read directory",
				slog.String(logging.AttrPath, cur.orig),
				slog.String(logging.AttrCause, err.Error()))
			continue
		}

		// Process each entry
		for _, de := range entries {
			name := de.Name()
			childOrig := filepath.Join(cur.orig, name)

			// Skip .git and .obsidian entirely (Rule 6)
			if name == ".git" || name == ".obsidian" {
				continue
			}

			// Get info about the entry (Lstat — does not follow symlinks)
			info, err := de.Info()
			if err != nil {
				log.Debug("failed to get entry info",
					slog.String(logging.AttrPath, childOrig),
					slog.String(logging.AttrCause, err.Error()))
				continue
			}

			// Handle symlinks (detected via Lstat mode bits)
			if info.Mode()&fs.ModeSymlink != 0 {
				target, err := filepath.EvalSymlinks(childOrig)
				if err != nil {
					// Broken symlink — skip with debug log (Rule 4)
					log.Debug("skipping broken symlink",
						slog.String(logging.AttrPath, childOrig),
						slog.String(logging.AttrCause, err.Error()))
					continue
				}

				// Check if target is within vault (Rule 2)
				if !containsPath(resolvedRoot, target) {
					log.Debug("skipping escaping symlink",
						slog.String(logging.AttrPath, childOrig),
						slog.String("target", target))
					continue
				}

				// Check if target is a directory
				targetInfo, err := os.Stat(target)
				if err != nil {
					log.Debug("failed to stat symlink target",
						slog.String(logging.AttrPath, childOrig),
						slog.String("target", target),
						slog.String(logging.AttrCause, err.Error()))
					continue
				}

				if targetInfo.IsDir() {
					// Resolve the childOrig so we have a real path for the next
					// iteration's containment / cycle check, but keep childOrig
					// as the watcher path so that events use the vault-relative
					// path the user expects.
					childRes, err := filepath.EvalSymlinks(childOrig)
					if err != nil {
						log.Debug("failed to resolve symlink for traversal",
							slog.String(logging.AttrPath, childOrig),
							slog.String(logging.AttrCause, err.Error()))
						continue
					}
					queue = append(queue, entry{orig: childOrig, res: childRes})
				}
			} else if info.IsDir() {
				// Regular directory — resolve for the next iteration
				childRes, err := filepath.EvalSymlinks(childOrig)
				if err != nil {
					log.Debug("failed to resolve directory",
						slog.String(logging.AttrPath, childOrig),
						slog.String(logging.AttrCause, err.Error()))
					continue
				}
				queue = append(queue, entry{orig: childOrig, res: childRes})
			}
		}
	}

	return nil
}

// containsPath checks if target is within vaultRoot using cleaned prefix
// comparison. Both arguments must already be resolved real paths. A plain
// string prefix check is defeated by /vault-evil matching /vault, so we
// require root + "/".
func containsPath(vaultRoot, target string) bool {
	cleanRoot := filepath.Clean(vaultRoot)
	cleanTarget := filepath.Clean(target)

	if cleanTarget == cleanRoot {
		return true
	}

	// Check if target starts with root + separator to avoid /vault-evil
	// matching /vault.
	return filepath.HasPrefix(cleanTarget, cleanRoot+string(filepath.Separator))
}

// readDir reads directory entries, returning an error if the directory
// cannot be read.
func readDir(dir string) ([]fs.DirEntry, error) {
	return os.ReadDir(dir)
}
