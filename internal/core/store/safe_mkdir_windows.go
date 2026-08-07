//go:build windows

package store

import (
	"errors"
	"os"
	"path/filepath"
)

func safeMkdirAll(path string, mode uint32) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	probe := abs
	for {
		info, statErr := os.Stat(probe)
		if statErr == nil {
			if !info.IsDir() {
				return errors.New("store: derived state path is not a directory")
			}
			break
		}
		if !os.IsNotExist(statErr) {
			return statErr
		}
		parent := filepath.Dir(probe)
		if parent == probe {
			return statErr
		}
		probe = parent
	}
	root, err := os.OpenRoot(probe)
	if err != nil {
		return err
	}
	defer root.Close()
	relative, err := filepath.Rel(probe, abs)
	if err != nil {
		return err
	}
	return root.MkdirAll(relative, os.FileMode(mode))
}
