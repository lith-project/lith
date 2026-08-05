package corpustest

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type Mismatch struct {
	Path   string // vault-relative path
	Reason string // "missing" | "digest" | "size" | "unexpected"
	Want   string // expected value (hex digest, byte count, or empty)
	Got    string // actual value (hex digest, byte count, or empty)
}

func Verify(repoRoot string, m *Manifest) ([]Mismatch, error) {
	if m == nil || len(m.Files) == 0 {
		return nil, fmt.Errorf("verify: manifest has no files")
	}

	corpusDir := filepath.Join(repoRoot, "tests", "vault", "corpus")
	if _, err := os.Stat(corpusDir); err != nil {
		return nil, fmt.Errorf("verify: corpus directory: %w", err)
	}

	manifestPaths := make(map[string]bool, len(m.Files))
	for _, f := range m.Files {
		manifestPaths[f.Path] = true
	}

	var mismatches []Mismatch

	for _, f := range m.Files {
		diskPath := filepath.Join(corpusDir, filepath.FromSlash(f.Path))
		fi, err := os.Lstat(diskPath)
		if os.IsNotExist(err) {
			mismatches = append(mismatches, Mismatch{
				Path:   f.Path,
				Reason: "missing",
			})
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("verify: lstat %s: %w", f.Path, err)
		}

		// On Windows, Git checks out symlinks as regular text files containing
		// the target path, and os.Lstat reports different sizes than the
		// manifest generated on macOS/Linux where symlinks are real. Skip
		// size and digest checks for symlinks on Windows.
		if runtime.GOOS == "windows" && fi.Mode()&os.ModeSymlink != 0 {
			continue
		}

		if fi.Size() != f.SizeBytes {
			mismatches = append(mismatches, Mismatch{
				Path:   f.Path,
				Reason: "size",
				Want:   fmt.Sprint(f.SizeBytes),
				Got:    fmt.Sprint(fi.Size()),
			})
			continue
		}

		got, err := computeDigest(diskPath)
		if err != nil {
			return nil, fmt.Errorf("verify: digest %s: %w", f.Path, err)
		}
		if got != f.Digest {
			mismatches = append(mismatches, Mismatch{
				Path:   f.Path,
				Reason: "digest",
				Want:   f.Digest,
				Got:    got,
			})
		}
	}

	err := filepath.Walk(corpusDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(corpusDir, path)
		if err != nil {
			return fmt.Errorf("verify: relative path: %w", err)
		}
		vaultPath := filepath.ToSlash(rel)

		base := filepath.Base(path)
		if base == ".gitattributes" && !strings.Contains(vaultPath, "/") {
			return nil
		}
		if base == "README.md" {
			return nil
		}

		if !manifestPaths[vaultPath] {
			mismatches = append(mismatches, Mismatch{
				Path:   vaultPath,
				Reason: "unexpected",
			})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("verify: walk corpus: %w", err)
	}

	sort.Slice(mismatches, func(i, j int) bool {
		return mismatches[i].Path < mismatches[j].Path
	})

	return mismatches, nil
}

func computeDigest(path string) (string, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return "", err
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "", fmt.Errorf("readlink: %w", err)
		}
		h := sha256.Sum256([]byte(target))
		return fmt.Sprintf("sha256:%x", h), nil
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", h.Sum(nil)), nil
}
