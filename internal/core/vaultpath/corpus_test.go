package vaultpath

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/text/unicode/norm"

	"github.com/lith-project/lith/internal/core/corpustest"
)

// repoRoot returns the repository root by walking up from the test file's
// location to find go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	// Get the absolute path of the current working directory (the package dir).
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repository root (no go.mod found)")
		}
		dir = parent
	}
}

func TestCorpus(t *testing.T) {
	root := repoRoot(t)
	vaultRoot := filepath.Join(root, "tests/vault", "corpus")
	m, err := corpustest.Load(root)
	if err != nil {
		t.Fatalf("failed to load corpus manifest: %v", err)
	}

	tests := []struct {
		id       string
		assertID func(t *testing.T, id string)
	}{
		{
			id: "EC-FS-001",
			assertID: func(t *testing.T, id string) {
				t.Helper()
				if !strings.Contains(id, " ") {
					t.Errorf("ID must contain spaces, got %q", id)
				}
			},
		},
		{
			id: "EC-FS-002",
			assertID: func(t *testing.T, id string) {
				t.Helper()
				specials := []string{"#", "[", "]", "%", "&"}
				for _, s := range specials {
					if !strings.Contains(id, s) {
						t.Errorf("ID must contain %q, got %q", s, id)
					}
				}
			},
		},
		{
			id: "EC-FS-003",
			assertID: func(t *testing.T, id string) {
				t.Helper()
				if id != strings.TrimSpace(id) {
					t.Errorf("ID must have no leading/trailing spaces, got %q", id)
				}
				if strings.HasPrefix(id, " ") || strings.HasSuffix(id, " ") {
					t.Errorf("ID has leading or trailing space: %q", id)
				}
			},
		},
		{
			id: "EC-FS-004",
			assertID: func(t *testing.T, id string) {
				t.Helper()
				// The NFC form of "café.md" uses the precomposed é (U+00E9).
				nfc := norm.NFC.String("caf\u00e9.md")
				// The ID is the full vault-relative path; check the filename component.
				if !strings.HasSuffix(id, nfc) {
					t.Errorf("ID must end with NFC form %q, got %q", nfc, id)
				}
			},
		},
		{
			id: "EC-FS-006",
			assertID: func(t *testing.T, id string) {
				t.Helper()
				if !strings.Contains(id, "\xf0\x9f\x8e\x89") {
					t.Errorf("ID must contain emoji bytes, got %q", id)
				}
			},
		},
		{
			id: "EC-FS-008",
			assertID: func(t *testing.T, id string) {
				t.Helper()
				// Deep path should contain nested directory separators.
				if !strings.Contains(id, "/") {
					t.Errorf("ID must contain forward slashes for deep path, got %q", id)
				}
				// Should have many directory levels.
				parts := strings.Split(id, "/")
				if len(parts) < 5 {
					t.Errorf("ID must have many nested levels, got %d parts in %q", len(parts), id)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			files := m.ByEdgeCase(tt.id)
			if len(files) == 0 {
				t.Fatalf("no corpus files tagged with %s", tt.id)
			}
			for _, f := range files {
				absPath := filepath.Join(root, "tests/vault/corpus", f.Path)
				p, err := New(vaultRoot, absPath)
				if err != nil {
					t.Errorf("New(%q, %q) failed: %v", vaultRoot, absPath, err)
					continue
				}
				tt.assertID(t, p.ID())
			}
		})
	}
}

func TestCorpusNFD(t *testing.T) {
	paths, err := corpustest.GenerateCase(t.TempDir(), "EC-FS-X-006")
	if errors.Is(err, corpustest.ErrUnsupportedPlatform) {
		t.Skipf("filesystem normalizes NFD filenames: %v", err)
	}
	if err != nil {
		t.Fatalf("GenerateCase: %v", err)
	}

	nfc := norm.NFC.String("caf\u00e9.md")
	for _, p := range paths {
		vaultRoot := filepath.Dir(p)
		vp, err := New(vaultRoot, p)
		if err != nil {
			t.Errorf("New(%q, %q) failed: %v", vaultRoot, p, err)
			continue
		}
		if vp.ID() != nfc {
			t.Errorf("NFD ID = %q, want NFC form %q", vp.ID(), nfc)
		}
	}
}
