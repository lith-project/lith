package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/lith-project/lith/internal/core/config"
	"github.com/lith-project/lith/internal/core/debounce"
	"github.com/lith-project/lith/internal/core/logging"
	"github.com/lith-project/lith/internal/core/watch"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
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

	var w watch.Watcher
	if cfg.Watcher.Enabled {
		w, err = watch.NewFSNotify(cfg.Vault.Path, logger)
		if err != nil {
			fmt.Fprintf(stderr, "lithd: %v\n", err)
			return 2
		}
	} else {
		w = watch.NewNoop()
	}

	// Create debouncer with config-derived bounds.
	settled := make(chan watch.Event, 64)
	db, err := debounce.New(
		cfg.Watcher.Debounce.Quiet,
		cfg.Watcher.Debounce.MaxDelay,
		settled,
	)
	if err != nil {
		fmt.Fprintf(stderr, "lithd: %v\n", err)
		return 2
	}

	logger.Info("debounce.bounds", "quiet", cfg.Watcher.Debounce.Quiet.String(), "max_delay", cfg.Watcher.Debounce.MaxDelay.String())

	// Drain settled events and log each one. The watcher and debouncer
	// are not started here — the daemon loop will wire them in a later
	// task. The goroutine blocks on <-settled until the channel is closed.
	_ = w  // will be started in the daemon loop task
	_ = db // will be started in the daemon loop task
	go func() {
		for ev := range settled {
			logger.Info(logging.EventFileChanged, logging.AttrPath, ev.Path.Raw())
		}
	}()

	return 0
}
