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
	if isWithinRoot(derived, vault) {
		return Location{}, ErrDerivedStateInsideVault
	}
	vaultHash := sha256.Sum256([]byte(vault))
	directory := filepath.Join(derived, "lith", hex.EncodeToString(vaultHash[:]))
	if isWithinRoot(directory, vault) {
		return Location{}, ErrDerivedStateInsideVault
	}
	return Location{
		Directory:        directory,
		Database:         filepath.Join(directory, "store.sqlite"),
		WriterLock:       filepath.Join(directory, "writer.lock"),
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
