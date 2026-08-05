package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/lith-project/lith/internal/core/config"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

func run(args []string, stderr io.Writer) int {
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

	if _, err := config.Load(*configPath); err != nil {
		fmt.Fprintf(stderr, "lithd: %v\n", err)
		return 2
	}

	return 0
}
