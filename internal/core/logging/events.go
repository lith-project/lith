package logging

// Event names, used as the slog message.
const (
	EventDaemonStarting = "daemon.starting"
	EventDaemonStarted  = "daemon.started"
	EventConfigLoaded   = "config.loaded"
	EventVaultWatching  = "vault.watching"
	EventFileChanged    = "file.changed"
	EventWatcherGap     = "watcher.gap"
	EventShutdownBegin  = "shutdown.begin"
	EventShutdownDone   = "shutdown.done"
	EventError          = "error"
	EventQueueOverflow  = "queue.overflow"
)

// Attribute keys. Every event uses these names and no synonyms.
const (
	AttrVaultPath = "vault_path"
	AttrPath      = "path"
	AttrOp        = "op"
	AttrCode      = "code"  // LITH-* diagnostic code, when one applies
	AttrCause     = "cause" // shutdown cause: "signal", "error"
	AttrSignal    = "signal"
	AttrDuration  = "duration_ms"
	AttrCount     = "count"
	AttrPolicy    = "policy"
)
