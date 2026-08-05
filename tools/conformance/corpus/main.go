package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lith-project/lith/internal/core/corpustest"
)

func main() {
	repoRoot, err := findRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	m, err := corpustest.Load(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	mismatches, err := corpustest.Verify(repoRoot, m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if len(mismatches) > 0 {
		for _, mm := range mismatches {
			switch mm.Reason {
			case "missing":
				fmt.Fprintf(os.Stderr, "missing: %s\n", mm.Path)
			case "digest":
				fmt.Fprintf(
					os.Stderr,
					"digest mismatch: %s (want %s, got %s)\n",
					mm.Path, mm.Want, mm.Got,
				)
			case "size":
				fmt.Fprintf(
					os.Stderr,
					"size mismatch: %s (want %s, got %s)\n",
					mm.Path, mm.Want, mm.Got,
				)
			case "unexpected":
				fmt.Fprintf(os.Stderr, "unexpected: %s\n", mm.Path)
			}
		}
		os.Exit(1)
	}
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("finding repo root: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("finding repo root: go.mod not found")
		}
		dir = parent
	}
}
