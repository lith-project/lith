package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ErrNotFound is returned when the configuration file does not exist.
var ErrNotFound = errors.New("config: file not found")

// Load reads a YAML configuration file, applies defaults, and validates.
// The resulting Config is fully resolved and ready for use.
func Load(path string) (*Config, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("config: %w: %w", ErrNotFound, err)
		}
		return nil, fmt.Errorf("config: %w", err)
	}

	// Start with defaults
	cfg := Defaults()

	// Decode YAML onto defaults with strict mode
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(&cfg); err != nil {
		// yaml.v3 errors include line numbers
		return nil, fmt.Errorf("config: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	return &cfg, nil
}

// containsLine checks if an error message contains a line number reference.
func containsLine(s string) bool {
	// yaml.v3 errors typically include "line X:"
	return strings.Contains(s, "line ")
}
