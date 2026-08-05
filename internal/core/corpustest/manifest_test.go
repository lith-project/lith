package corpustest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// findRepoRoot walks up from the test's working directory (the package
// directory) until it finds a go.mod file. Test helpers rely on this to
// locate tests/vault/expected/manifest.json in the real checkout.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("findRepoRoot: could not find go.mod walking up from the package directory")
		}
		dir = parent
	}
}

// writeManifest writes content to tests/vault/expected/manifest.json under
// repoRoot, creating directories as needed. It is used to plant malformed
// fixtures for the validation table tests.
func writeManifest(t *testing.T, repoRoot, content string) {
	t.Helper()
	dir := filepath.Join(repoRoot, "tests", "vault", "expected")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", dir, err)
	}
	path := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

// validManifest is a minimal manifest that passes every validation rule.
// The malformed-fixture tests start from this and mutate exactly one field.
const validManifest = `{
  "spec_version": 1,
  "corpus_id": "test-corpus",
  "generated_at": "2026-08-04T00:00:00Z",
  "corpus_digest": "sha256:abc",
  "file_count": 1,
  "files": [
    {"path": "notes/a.md", "size_bytes": 10, "digest": "sha256:deadbeef", "edge_cases": ["EC-X-001"]}
  ]
}`

func TestLoadRejectsMalformedManifests(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantField string // substring expected in the error message
		wantErr   bool
	}{
		{
			name:      "bad spec_version",
			json:      `{"spec_version":2,"corpus_id":"t","generated_at":"2026-08-04T00:00:00Z","corpus_digest":"sha256:abc","file_count":1,"files":[{"path":"a.md","size_bytes":1,"digest":"sha256:d","edge_cases":[]}]}`,
			wantField: "spec_version",
			wantErr:   true,
		},
		{
			name:      "file_count mismatch",
			json:      `{"spec_version":1,"corpus_id":"t","generated_at":"2026-08-04T00:00:00Z","corpus_digest":"sha256:abc","file_count":5,"files":[{"path":"a.md","size_bytes":1,"digest":"sha256:d","edge_cases":[]}]}`,
			wantField: "file_count",
			wantErr:   true,
		},
		{
			name:      "missing sha256 prefix on file digest",
			json:      `{"spec_version":1,"corpus_id":"t","generated_at":"2026-08-04T00:00:00Z","corpus_digest":"sha256:abc","file_count":1,"files":[{"path":"a.md","size_bytes":1,"digest":"md5:deadbeef","edge_cases":[]}]}`,
			wantField: "digest",
			wantErr:   true,
		},
		{
			name:      "missing sha256 prefix on corpus_digest",
			json:      `{"spec_version":1,"corpus_id":"t","generated_at":"2026-08-04T00:00:00Z","corpus_digest":"md5:abc","file_count":1,"files":[{"path":"a.md","size_bytes":1,"digest":"sha256:d","edge_cases":[]}]}`,
			wantField: "corpus_digest",
			wantErr:   true,
		},
		{
			name:      "absolute path",
			json:      `{"spec_version":1,"corpus_id":"t","generated_at":"2026-08-04T00:00:00Z","corpus_digest":"sha256:abc","file_count":1,"files":[{"path":"/etc/passwd.md","size_bytes":1,"digest":"sha256:d","edge_cases":[]}]}`,
			wantField: "path",
			wantErr:   true,
		},
		{
			name:      "backslash in path",
			json:      `{"spec_version":1,"corpus_id":"t","generated_at":"2026-08-04T00:00:00Z","corpus_digest":"sha256:abc","file_count":1,"files":[{"path":"assets\\board.canvas","size_bytes":1,"digest":"sha256:d","edge_cases":[]}]}`,
			wantField: "path",
			wantErr:   true,
		},
		{
			name:    "valid baseline passes",
			json:    validManifest,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeManifest(t, root, tt.json)

			m, err := Load(root)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Load(%q): expected error, got nil (manifest=%+v)", root, m)
				}
				if tt.wantField != "" && !strings.Contains(err.Error(), tt.wantField) {
					t.Errorf("Load(%q): error %q does not name field %q", root, err.Error(), tt.wantField)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load(%q): unexpected error: %v", root, err)
			}
		})
	}
}

func TestUnknownFieldsAreIgnored(t *testing.T) {
	// An unknown field is NOT an error (Rule 4): the manifest may gain fields.
	root := t.TempDir()
	writeManifest(t, root, `{
  "spec_version": 1,
  "corpus_id": "t",
  "generated_at": "2026-08-04T00:00:00Z",
  "corpus_digest": "sha256:abc",
  "file_count": 1,
  "future_field": {"anything": true},
  "files": [{"path": "a.md", "size_bytes": 1, "digest": "sha256:d", "edge_cases": [], "extra": 7}]
}`)
	if _, err := Load(root); err != nil {
		t.Fatalf("Load with unknown fields should succeed, got: %v", err)
	}
}

func TestGeneratedAtDecodedAsTime(t *testing.T) {
	// Rule 2: generated_at must be decoded into time.Time, not left as a string.
	root := t.TempDir()
	writeManifest(t, root, validManifest)
	m, err := Load(root)
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}
	want := "2026-08-04T00:00:00Z"
	got := m.GeneratedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
	if got != want {
		t.Errorf("GeneratedAt = %s, want %s", got, want)
	}
}

func TestByEdgeCaseEmptyWhenNoMatch(t *testing.T) {
	// DoD: ByEdgeCase on an unknown identifier returns an empty, non-nil slice.
	root := t.TempDir()
	writeManifest(t, root, validManifest)
	m, err := Load(root)
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}
	got := m.ByEdgeCase("DOES-NOT-EXIST")
	if got == nil {
		t.Fatal("ByEdgeCase returned nil, want non-nil empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("ByEdgeCase(unknown) = %v, want empty slice", got)
	}
}

func TestLoadCommittedManifest(t *testing.T) {
	// DoD: Load succeeds against the committed manifest and returns 96 files.
	root := findRepoRoot(t)
	m, err := Load(root)
	if err != nil {
		t.Fatalf("Load(%q): unexpected error: %v", root, err)
	}
	if m.FileCount != 96 {
		t.Errorf("FileCount = %d, want 96", m.FileCount)
	}
	if len(m.Files) != 96 {
		t.Errorf("len(Files) = %d, want 96", len(m.Files))
	}
	if m.SpecVersion != 1 {
		t.Errorf("SpecVersion = %d, want 1", m.SpecVersion)
	}

	// DoD: ByEdgeCase("EC-FS-009") returns assets/board.canvas.
	matches := m.ByEdgeCase("EC-FS-009")
	found := false
	for _, f := range matches {
		if f.Path == "assets/board.canvas" {
			found = true
		}
	}
	if !found {
		t.Errorf("ByEdgeCase(\"EC-FS-009\") = %d files; want assets/board.canvas among them", len(matches))
	}
}
