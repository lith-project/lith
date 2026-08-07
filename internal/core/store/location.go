package store

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrDerivedStateInsideVault = errors.New("store: derived state root is inside vault root")

type Location struct {
	Directory        string
	Database         string
	WriterLock       string
	VaultFingerprint string
}

func resolveLocation(derivedStateRoot, vaultRoot string) (Location, error) {
	vault, err := canonicalExistingPath(vaultRoot)
	if err != nil {
		return Location{}, fmt.Errorf("store: canonicalize vault root: %w", err)
	}
	derived, err := canonicalPath(derivedStateRoot)
	if err != nil {
		return Location{}, fmt.Errorf("store: canonicalize derived state root: %w", err)
	}
	if target, ok, err := danglingSymlinkTarget(derivedStateRoot); err != nil {
		return Location{}, fmt.Errorf("store: inspect derived state root: %w", err)
	} else if ok && isWithinRoot(target, vault) {
		return Location{}, ErrDerivedStateInsideVault
	}
	if isWithinRoot(derived, vault) {
		return Location{}, ErrDerivedStateInsideVault
	}
	vaultHash := sha256.Sum256([]byte(vault))
	directory := filepath.Join(derived, "lith", hex.EncodeToString(vaultHash[:]))
	hasSymlink, err := containsSymlinkComponent(directory, derived)
	if err != nil {
		return Location{}, fmt.Errorf("store: inspect store directory: %w", err)
	}
	if hasSymlink {
		return Location{}, ErrDerivedStateInsideVault
	}
	canonicalDirectory, err := canonicalPath(directory)
	if err != nil {
		return Location{}, fmt.Errorf("store: canonicalize store directory: %w", err)
	}
	if isWithinRoot(canonicalDirectory, vault) {
		return Location{}, ErrDerivedStateInsideVault
	}
	return Location{
		Directory:        canonicalDirectory,
		Database:         filepath.Join(canonicalDirectory, "store.sqlite"),
		WriterLock:       filepath.Join(canonicalDirectory, "writer.lock"),
		VaultFingerprint: "root=" + vault + "\nnormalization=NFC",
	}, nil
}

func canonicalExistingPath(raw string) (string, error) {
	if raw == "" {
		return "", errors.New("path is empty")
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	return filepath.Clean(resolved), nil
}

func canonicalPath(raw string) (string, error) {
	if raw == "" {
		return "", errors.New("path is empty")
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", err
	}
	for probe := abs; ; probe = filepath.Dir(probe) {
		resolved, evalErr := filepath.EvalSymlinks(probe)
		if evalErr == nil {
			relative, relErr := filepath.Rel(probe, abs)
			if relErr != nil {
				return "", relErr
			}
			return filepath.Clean(filepath.Join(resolved, relative)), nil
		}
		if !errors.Is(evalErr, os.ErrNotExist) {
			return "", evalErr
		}
		if filepath.Dir(probe) == probe {
			return "", evalErr
		}
	}
}

func isWithinRoot(path, root string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !filepath.IsAbs(relative) && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
}

func isOutsideRoot(path, root string) bool {
	return !isWithinRoot(path, root)
}

func containsSymlinkComponent(path, root string) (bool, error) {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false, err
	}
	if relative == "." {
		return false, nil
	}
	current := root
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return true, nil
		}
	}
	return false, nil
}

func danglingSymlinkTarget(raw string) (string, bool, error) {
	if raw == "" {
		return "", false, errors.New("path is empty")
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", false, err
	}
	info, err := os.Lstat(abs)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return "", false, nil
	}
	target, err := os.Readlink(abs)
	if err != nil {
		return "", false, err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(abs), target)
	}
	canonicalTarget, err := canonicalPath(target)
	if err != nil {
		return "", false, err
	}
	return canonicalTarget, true, nil
}
