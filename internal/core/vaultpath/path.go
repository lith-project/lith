package vaultpath

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// Path is a file's identity within a vault, plus the bytes needed to open it.
type Path struct {
	id  string // vault-relative, forward slashes, cleaned
	raw string // absolute on-disk path, exactly as given
}

// ErrOutsideVault is returned, wrapped, when absPath is not within vaultRoot.
var ErrOutsideVault = errors.New("vaultpath: path is outside the vault root")

// New builds a Path for an absolute on-disk path within an absolute vault root.
// It performs no filesystem I/O.
func New(vaultRoot, absPath string) (Path, error) {
	rootFwd := filepath.ToSlash(vaultRoot)
	pathFwd := filepath.ToSlash(absPath)

	rootClean := path.Clean(rootFwd)
	pathClean := path.Clean(pathFwd)

	rel, err := filepath.Rel(rootClean, pathClean)
	if err != nil {
		return Path{}, fmt.Errorf("vaultpath: %w", ErrOutsideVault)
	}

	if rel == ".." || strings.HasPrefix(rel, "../") || rel == "." {
		return Path{}, fmt.Errorf("vaultpath: %w", ErrOutsideVault)
	}

	return Path{
		id:  rel,
		raw: absPath,
	}, nil
}

// ID returns the vault-relative identity.
func (p Path) ID() string { return p.id }

// Raw returns the absolute on-disk path to open.
func (p Path) Raw() string { return p.raw }
