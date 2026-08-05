package corpustest

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/text/unicode/norm"
)

func TestGenerateCaseNFD(t *testing.T) {
	dir := t.TempDir()
	paths, err := GenerateCase(dir, "EC-FS-X-006")
	if errors.Is(err, ErrUnsupportedPlatform) {
		t.Skip("filesystem normalizes NFD filenames; case not representable")
	}
	if err != nil {
		t.Fatalf("GenerateCase: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(paths))
	}

	absPath := paths[0]
	if !filepath.IsAbs(absPath) {
		t.Errorf("path is not absolute: %s", absPath)
	}
	if !strings.HasPrefix(absPath, dir) {
		t.Errorf("path %s is not inside temp dir %s", absPath, dir)
	}

	// Verify the filename is the NFD form, not NFC.
	nfcName := norm.NFC.String("café.md")
	nfdName := norm.NFD.String("café.md")
	gotName := filepath.Base(absPath)
	if gotName == nfcName {
		t.Errorf("filename is NFC form %q; expected NFD form %q", gotName, nfdName)
	}
	if gotName != nfdName {
		t.Errorf("filename = %q, want NFD form %q", gotName, nfdName)
	}

	// Verify the file exists on disk with non-empty content.
	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("file is empty")
	}
}

func TestGenerateCaseUnknown(t *testing.T) {
	dir := t.TempDir()
	_, err := GenerateCase(dir, "UNKNOWN-ID")
	if err == nil {
		t.Fatal("expected error for unknown identifier, got nil")
	}
	if errors.Is(err, ErrUnsupportedPlatform) {
		t.Error("error should not be ErrUnsupportedPlatform")
	}
	if !strings.Contains(err.Error(), "unknown edge case identifier") {
		t.Errorf("error %q does not contain 'unknown edge case identifier'", err.Error())
	}
}

func TestGenerateCaseProbe(t *testing.T) {
	dir := t.TempDir()
	_, err := GenerateCase(dir, "EC-FS-X-006")
	// On any host this should either succeed or skip cleanly — never fail
	// for platform reasons.
	if err != nil && !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("GenerateCase returned unexpected error: %v", err)
	}
	// If ErrUnsupportedPlatform, the probe mechanism is working correctly.
}
