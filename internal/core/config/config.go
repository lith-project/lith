package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config is the root configuration for lith.
type Config struct {
	Vault   Vault   `yaml:"vault"`
	Log     Log     `yaml:"log"`
	Daemon  Daemon  `yaml:"daemon"`
	Watcher Watcher `yaml:"watcher"`
	Queue   Queue   `yaml:"queue"`
}

type Vault struct {
	Path string `yaml:"path"` // required, absolute after Validate
}

type Log struct {
	Level  string `yaml:"level"`  // debug | info | warn | error
	Format string `yaml:"format"` // text | json
}

type Daemon struct {
	LockFile string `yaml:"lock_file"`
}

type Watcher struct {
	Enabled  bool     `yaml:"enabled"`
	Debounce Debounce `yaml:"debounce"`
}

type Debounce struct {
	Quiet    time.Duration `yaml:"quiet"`     // default 200ms
	MaxDelay time.Duration `yaml:"max_delay"` // default 5s
}

type Queue struct {
	Capacity int `yaml:"capacity"`
}

// Defaults returns a Config with sensible default values.
// vault.path has no default and is required.
// daemon.lock_file is resolved in a later task.
func Defaults() Config {
	return Config{
		Log: Log{
			Level:  "info",
			Format: "text",
		},
		Watcher: Watcher{
			Enabled: true,
			Debounce: Debounce{
				Quiet:    200 * time.Millisecond,
				MaxDelay: 5 * time.Second,
			},
		},
		Queue: Queue{
			Capacity: 4096,
		},
	}
}

// Validate checks all configuration fields and returns ALL problems.
// It uses errors.Join to return multiple errors as a single error.
// Each error names the field by YAML path.
// Validate does NOT touch the filesystem beyond path resolution.
func (c *Config) Validate() error {
	var errs []error

	// vault.path: must not be empty; expand ~ and make absolute
	if c.Vault.Path == "" {
		errs = append(errs, fmt.Errorf("vault.path: must not be empty"))
	} else {
		expanded, err := expandAndAbsPath(c.Vault.Path)
		if err != nil {
			errs = append(errs, fmt.Errorf("vault.path: %w", err))
		} else {
			c.Vault.Path = expanded
		}
	}

	// log.level: must be one of "debug", "info", "warn", "error"
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Log.Level] {
		errs = append(errs, fmt.Errorf("log.level: must be one of debug, info, warn, error (got %q)", c.Log.Level))
	}

	// log.format: must be one of "text", "json"
	validFormats := map[string]bool{"text": true, "json": true}
	if !validFormats[c.Log.Format] {
		errs = append(errs, fmt.Errorf("log.format: must be one of text, json (got %q)", c.Log.Format))
	}

	// debounce: reject quiet >= max_delay
	if c.Watcher.Debounce.Quiet >= c.Watcher.Debounce.MaxDelay {
		errs = append(errs, fmt.Errorf("watcher.debounce: quiet (%v) must be less than max_delay (%v)", c.Watcher.Debounce.Quiet, c.Watcher.Debounce.MaxDelay))
	}

	// queue.capacity: reject capacity < 1
	if c.Queue.Capacity < 1 {
		errs = append(errs, fmt.Errorf("queue.capacity: must be at least 1 (got %d)", c.Queue.Capacity))
	}

	return errors.Join(errs...)
}

// expandAndAbsPath expands ~ to the user's home directory and makes the path absolute.
func expandAndAbsPath(path string) (string, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot resolve home directory: %w", err)
		}
		path = filepath.Join(homeDir, path[1:])
	}

	// Make absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("cannot make path absolute: %w", err)
	}

	return absPath, nil
}
