package watch

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"

	"github.com/lith-project/lith/internal/core/logging"
	"github.com/lith-project/lith/internal/core/vaultpath"
)

type fsnotifyWatcher struct {
	vaultRoot string
	log       *slog.Logger
	events    chan Event
	gaps      chan GapReason
}

// NewFSNotify builds a Watcher over the vault rooted at vaultRoot.
func NewFSNotify(vaultRoot string, log *slog.Logger) (Watcher, error) {
	if !filepath.IsAbs(vaultRoot) {
		return nil, fmt.Errorf("watch: vault root must be absolute, got %q", vaultRoot)
	}
	log.Info(logging.EventVaultWatching, logging.AttrVaultPath, vaultRoot)
	return &fsnotifyWatcher{
		vaultRoot: vaultRoot,
		log:       log,
		events:    make(chan Event, 64),
		gaps:      make(chan GapReason, 8),
	}, nil
}

func (w *fsnotifyWatcher) Events() <-chan Event   { return w.events }
func (w *fsnotifyWatcher) Gaps() <-chan GapReason { return w.gaps }

func (w *fsnotifyWatcher) Start(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("watch: creating watcher: %w", err)
	}
	defer watcher.Close()
	defer close(w.events)
	defer close(w.gaps)

	if err := watcher.Add(w.vaultRoot); err != nil {
		return fmt.Errorf("watch: adding vault root %q: %w", w.vaultRoot, err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			op := mapOp(ev.Op)
			vp, err := vaultpath.New(w.vaultRoot, ev.Name)
			if err != nil {
				w.log.Warn("watch: skipping event outside vault",
					logging.AttrPath, ev.Name,
					logging.AttrCause, err.Error(),
				)
				continue
			}
			event := Event{Path: vp, Op: op}
			w.log.Info(logging.EventFileChanged,
				logging.AttrPath, vp.ID(),
				logging.AttrOp, op.String(),
			)
			select {
			case w.events <- event:
			case <-ctx.Done():
				return nil
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			reason := classifyGap(err)
			w.log.Warn(logging.EventWatcherGap,
				logging.AttrCode, string(reason),
				logging.AttrCause, err.Error(),
			)
			select {
			case w.gaps <- reason:
			default:
			}
		}
	}
}

func classifyGap(err error) GapReason {
	msg := err.Error()
	if strings.Contains(msg, "inotify") && strings.Contains(msg, "watch limit") {
		return GapPlatformLimit
	}
	return GapWatchError
}

func mapOp(o fsnotify.Op) Op {
	switch {
	case o&fsnotify.Create != 0:
		return OpCreate
	case o&fsnotify.Write != 0:
		return OpWrite
	case o&fsnotify.Remove != 0:
		return OpRemove
	case o&fsnotify.Rename != 0:
		return OpRename
	default:
		return OpWrite
	}
}
