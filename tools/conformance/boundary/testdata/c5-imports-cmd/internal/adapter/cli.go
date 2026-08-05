package cli

import "github.com/lith-project/lith/cmd/lithd/config"

// Load reads the lithd config to demonstrate the adapter→cmd dependency.
func Load() *config.Settings { return nil }
