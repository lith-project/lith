package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/lith-project/lith/internal/core/config"
	"github.com/lith-project/lith/internal/core/daemon"
	"github.com/lith-project/lith/internal/core/debounce"
	"github.com/lith-project/lith/internal/core/logging"
	"github.com/lith-project/lith/internal/core/queue"
	"github.com/lith-project/lith/internal/core/watch"
)

// signalNameKey is the context key carrying the channel that receives the
// name of the first terminating signal, if one fires. run() reads it when the
// context is cancelled so shutdown.begin can carry the signal name; the
// daemon core stays signal-agnostic.
type signalNameKey struct{}

func main() {
	// The first SIGINT or SIGTERM cancels the daemon context; a second one
	// during shutdown forces an immediate non-zero exit.
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	sigNameCh := make(chan string, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go watchSignals(sigCh, sigNameCh, cancel, os.Stderr, os.Exit)

	ctx = context.WithValue(ctx, signalNameKey{}, sigNameCh)
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

// watchSignals forwards the first terminating signal's name, cancels the
// daemon context, and forces an immediate non-zero exit when a second signal
// arrives during shutdown. The name is published before cancel so run()
// observes it on shutdown. It returns when sigCh is closed with no signal.
func watchSignals(sigCh <-chan os.Signal, nameCh chan<- string, cancel context.CancelFunc, warn io.Writer, forceExit func(int)) {
	s, ok := <-sigCh
	if !ok {
		return
	}
	nameCh <- s.String()
	cancel()
	if _, ok := <-sigCh; ok {
		fmt.Fprintln(warn, "lithd: second signal received, forcing exit")
		forceExit(1)
	}
}

// run executes the daemon until ctx is cancelled. It returns an exit code:
// 0 for clean shutdown, 1 for lock/watcher/runtime failure, 2 for
// configuration failure.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("lithd", flag.ContinueOnError)
	fs.SetOutput(stderr)

	configPath := fs.String("config", "", "path to YAML configuration file (required)")

	if err := fs.Parse(args); err != nil {
		// flag.Parse already printed usage
		return 2
	}

	if *configPath == "" {
		fmt.Fprintln(stderr, "Usage: lithd --config <path>")
		return 2
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(stderr, "lithd: %v\n", err)
		return 2
	}

	logger, err := logging.New(stderr, cfg.Log)
	if err != nil {
		fmt.Fprintf(stderr, "lithd: %v\n", err)
		return 2
	}

	logger.Info(logging.EventDaemonStarting)
	logger.Info(logging.EventConfigLoaded, logging.AttrVaultPath, cfg.Vault.Path)

	// Acquire the exclusive vault lock before starting any component. Every
	// return path after this point releases it via defer.
	lock, err := daemon.Acquire(cfg.Vault.Path, cfg.Daemon.LockFile, logger)
	if err != nil {
		fmt.Fprintf(stderr, "lithd: %v\n", err)
		return 1
	}
	defer func() { _ = lock.Release() }()

	// Watcher: FSNotify or Noop.
	var w watch.Watcher
	if cfg.Watcher.Enabled {
		w, err = watch.NewFSNotify(cfg.Vault.Path, logger)
		if err != nil {
			fmt.Fprintf(stderr, "lithd: %v\n", err)
			return 1
		}
	} else {
		w = watch.NewNoop()
	}

	// Debouncer.
	settled := make(chan watch.Event, 64)
	db, err := debounce.New(cfg.Watcher.Debounce.Quiet, cfg.Watcher.Debounce.MaxDelay, settled)
	if err != nil {
		fmt.Fprintf(stderr, "lithd: %v\n", err)
		return 1
	}
	logger.Info("debounce.bounds", "quiet", cfg.Watcher.Debounce.Quiet.String(), "max_delay", cfg.Watcher.Debounce.MaxDelay.String())

	// Queue.
	q, err := queue.New(cfg.Queue.Capacity, logger)
	if err != nil {
		fmt.Fprintf(stderr, "lithd: %v\n", err)
		return 1
	}
	logger.Info("queue.capacity", "capacity", cfg.Queue.Capacity)

	// Log signal-driven shutdown with the signal name. daemon.Run is
	// signal-agnostic; the name is published through the context by main().
	// shutdownLogged is closed once the record has been written so run() can
	// wait for it before returning: main() ends in os.Exit, which does not
	// wait for goroutines, and a fast shutdown otherwise drops the record.
	shutdownLogged := make(chan struct{})
	go func() {
		defer close(shutdownLogged)
		<-ctx.Done()
		sigName := "signal"
		if ch, ok := ctx.Value(signalNameKey{}).(chan string); ok {
			select {
			case n := <-ch:
				sigName = n
			default:
			}
		}
		logger.Warn(logging.EventShutdownBegin,
			logging.AttrCause, "signal",
			logging.AttrSignal, sigName,
		)
	}()

	components := []daemon.Component{
		{Name: "watcher", Run: func(ctx context.Context) error {
			return w.Start(ctx)
		}},
		{Name: "debouncer", Run: func(ctx context.Context) error {
			return db.Run(ctx, w.Events())
		}},
		{Name: "settle-forwarder", Run: func(ctx context.Context) error {
			for ev := range settled {
				q.Push(ev)
			}
			return nil
		}},
		{Name: "consumer", Run: func(ctx context.Context) error {
			for {
				ev, err := q.Pop(ctx)
				if err != nil {
					return err
				}
				logger.Info(logging.EventFileChanged, logging.AttrPath, ev.Path.Raw())
			}
		}},
	}

	logger.Info(logging.EventDaemonStarted)

	runErr := daemon.Run(ctx, components, logger)

	// Wait for the shutdown.begin record to reach the log before returning
	// into os.Exit. Only when the context was actually cancelled: on the
	// component-error path the goroutine is still parked on ctx.Done() and
	// would never close the channel.
	if ctx.Err() != nil {
		<-shutdownLogged
	}

	if runErr != nil {
		fmt.Fprintf(stderr, "lithd: %v\n", runErr)
		return 1
	}
	return 0
}
