package corpustest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/text/unicode/norm"
)

// ErrUnsupportedPlatform is returned by GenerateCase when the test filesystem
// normalizes Unicode filenames, making the edge case unrepresentable.
var ErrUnsupportedPlatform = errors.New("corpustest: case not representable on this filesystem")

// GenerateCase creates temporary file(s) in dir that exercise the edge case
// identified by id. It returns the absolute paths of the generated file(s).
//
// GenerateCase never deletes files; callers are responsible for cleanup
// (e.g. via t.Cleanup or os.Remove). Generated files are not added to the
// committed corpus or manifest.
//
// Example:
//
//	paths, err := corpustest.GenerateCase(t.TempDir(), "EC-FS-X-006")
//	if errors.Is(err, corpustest.ErrUnsupportedPlatform) {
//	    t.Skip("filesystem normalizes NFD filenames; case not representable")
//	}
func GenerateCase(dir, id string) ([]string, error) {
	switch id {
	case "EC-FS-X-006":
		return generateNFDCase(dir)
	default:
		return nil, fmt.Errorf("corpustest: unknown edge case identifier %q", id)
	}
}

// generateNFDCase generates a file with an NFD (decomposed) Unicode filename.
// The well-known base name is "café.md"; the NFD form decomposes é into
// U+0065 (e) + U+0301 (combining acute accent).
func generateNFDCase(dir string) ([]string, error) {
	base := "café.md"
	nfcName := norm.NFC.String(base)
	nfdName := norm.NFD.String(base)

	// If NFD and NFC produce the same bytes there is nothing to test.
	if nfcName == nfdName {
		return nil, fmt.Errorf("corpustest: NFD and NFC are identical for %q; cannot generate edge case", base)
	}

	// Probe the filesystem: write the NFD name, read back, and check whether
	// the OS preserved the decomposed form or normalized it to NFC.
	if err := probeNFD(dir, nfdName); err != nil {
		return nil, err
	}

	// Write the actual test file with the NFD name.
	content := []byte("EC-FS-X-006: NFD filename test\n")
	nfdPath := filepath.Join(dir, nfdName)
	if err := os.WriteFile(nfdPath, content, 0o644); err != nil {
		return nil, fmt.Errorf("corpustest: writing NFD file: %w", err)
	}

	return []string{nfdPath}, nil
}

// probeNFD writes a temporary file with the given NFD name into dir, then reads
// back the directory listing to determine whether the filesystem preserved the
// decomposed form. The probe file is left in place (the caller's temp dir
// cleans it up).
func probeNFD(dir, nfdName string) error {
	probePath := filepath.Join(dir, nfdName)
	if err := os.WriteFile(probePath, []byte("probe"), 0o644); err != nil {
		return fmt.Errorf("corpustest: writing probe file: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("corpustest: reading directory: %w", err)
	}

	for _, e := range entries {
		if e.Name() == nfdName {
			// Filesystem preserved the NFD form — proceed.
			return nil
		}
	}

	// The NFD-named file was not found; the filesystem normalized it.
	return ErrUnsupportedPlatform
}
