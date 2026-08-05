// Package corpustest provides helpers for tests that consume the committed
// Lith corpus. It reads the corpus manifest and exposes lookups by edge-case
// identifier, so tests never hardcode corpus paths.
package corpustest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// manifestRelPath is the canonical location of the corpus manifest relative
// to the repository root.
const manifestRelPath = "tests/vault/expected/manifest.json"

// File describes a single corpus entry from the manifest.
type File struct {
	Path      string   `json:"path"`       // vault-relative, forward slashes, e.g. "assets/board.canvas"
	SizeBytes int64    `json:"size_bytes"` // file size in bytes
	Digest    string   `json:"digest"`     // "sha256:<hex>"
	EdgeCases []string `json:"edge_cases"` // e.g. ["EC-FS-009"]
}

// Manifest is the parsed and validated corpus manifest.
type Manifest struct {
	CorpusID     string    `json:"corpus_id"`
	SpecVersion  int       `json:"spec_version"`
	FileCount    int       `json:"file_count"`
	GeneratedAt  time.Time `json:"generated_at"`
	CorpusDigest string    `json:"corpus_digest"`
	Files        []File    `json:"files"`
}

// Load reads and validates the manifest at tests/vault/expected/manifest.json
// relative to repoRoot. It performs structural validation only; it does not
// open, read, or hash the corpus files themselves.
func Load(repoRoot string) (*Manifest, error) {
	path := filepath.Join(repoRoot, manifestRelPath)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("manifest: reading %s: %w", path, err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("manifest: parsing %s: %w", path, err)
	}

	if err := m.validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

// ByEdgeCase returns every file tagged with the given edge-case identifier.
// It returns an empty slice, never nil, when nothing matches.
func (m *Manifest) ByEdgeCase(id string) []File {
	out := []File{}
	for _, f := range m.Files {
		for _, ec := range f.EdgeCases {
			if ec == id {
				out = append(out, f)
				break
			}
		}
	}
	return out
}

// validate enforces the structural rules from the manifest specification.
// Each returned error names the offending field so callers can pinpoint the
// problem.
func (m *Manifest) validate() error {
	if m.SpecVersion != 1 {
		return fmt.Errorf("manifest: spec_version: must be 1, got %d", m.SpecVersion)
	}
	if len(m.Files) != m.FileCount {
		return fmt.Errorf("manifest: file_count: declared %d but files array has %d entries", m.FileCount, len(m.Files))
	}
	if !strings.HasPrefix(m.CorpusDigest, "sha256:") {
		return fmt.Errorf("manifest: corpus_digest: missing sha256: prefix, got %q", m.CorpusDigest)
	}
	for i, f := range m.Files {
		if !strings.HasPrefix(f.Digest, "sha256:") {
			return fmt.Errorf("manifest: files[%d].digest: missing sha256: prefix, got %q", i, f.Digest)
		}
		if f.Path == "" {
			return fmt.Errorf("manifest: files[%d].path: empty path", i)
		}
		if strings.HasPrefix(f.Path, "/") {
			return fmt.Errorf("manifest: files[%d].path: must be relative, got absolute path %q", i, f.Path)
		}
		if strings.Contains(f.Path, `\`) {
			return fmt.Errorf("manifest: files[%d].path: must use forward slashes, got backslash in %q", i, f.Path)
		}
	}
	return nil
}
