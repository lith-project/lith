package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/lith-project/lith/internal/core/config"
	"github.com/lith-project/lith/internal/core/logging"
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

	return 0
}
