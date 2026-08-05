package watch

import (
	"context"

	"github.com/lith-project/lith/internal/core/vaultpath"
)

// Op represents a filesystem operation observed by the watcher.
type Op uint8

const (
	OpCreate Op = iota
	OpWrite
	OpRemove
	OpRename
)

// String returns the human-readable form of the operation.
func (o Op) String() string {
	switch o {
	case OpCreate:
		return "create"
	case OpWrite:
		return "write"
	case OpRemove:
		return "remove"
	case OpRename:
		return "rename"
	default:
		return "unknown"
	}
}

// Event represents a single filesystem change observed in the vault.
type Event struct {
	Path vaultpath.Path
	Op   Op
}

// Watcher observes a vault and emits change events until ctx is cancelled.
type Watcher interface {
	// Events returns the channel events are delivered on. It is closed when
	// the watcher stops.
	Events() <-chan Event
	// Gaps returns the channel gap reasons are delivered on. It is closed when
	// the watcher stops.
	Gaps() <-chan GapReason
	// Start begins watching. It returns when ctx is cancelled or setup fails.
	Start(ctx context.Context) error
}
