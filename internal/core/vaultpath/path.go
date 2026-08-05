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

	// Compute relative path using forward-slash-normalized, cleaned strings.
	// filepath.Rel is not safe here: on Windows it uses backslash semantics and
	// mishandles forward-slash-absolute paths like "/vault".
	var rel string
	if pathClean == rootClean {
		rel = "."
	} else if strings.HasPrefix(pathClean, rootClean+"/") {
		rel = pathClean[len(rootClean)+1:]
	} else {
		rel = ".."
	}

	// filepath.Rel returns OS-native separators; the identity must always be
	// forward-slash separated, and the escape check needs that canonical form.
	rel = filepath.ToSlash(rel)
	if rel == ".." || strings.HasPrefix(rel, "../") || rel == "." {
		return Path{}, fmt.Errorf("vaultpath: %w", ErrOutsideVault)
	}

	return Path{
		id:  normalizeID(rel),
		raw: absPath,
	}, nil
}

// ID returns the vault-relative identity.
func (p Path) ID() string { return p.id }

// Raw returns the absolute on-disk path to open.
func (p Path) Raw() string { return p.raw }
