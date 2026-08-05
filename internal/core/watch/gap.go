package watch

// GapReason describes why watcher guarantees were invalidated.
type GapReason string

const (
	// GapQueueOverflow indicates the internal event queue overflowed.
	GapQueueOverflow GapReason = "queue_overflow"
	// GapWatchError indicates a non-platform watcher error occurred.
	GapWatchError GapReason = "watch_error"
	// GapPlatformLimit indicates a platform-specific limit was exhausted (e.g. inotify watches).
	GapPlatformLimit GapReason = "platform_limit"
)
