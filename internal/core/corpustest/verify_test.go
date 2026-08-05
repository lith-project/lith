package corpustest

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyClean(t *testing.T) {
	repoRoot := findRepoRoot(t)
	m, err := Load(repoRoot)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	mismatches, err := Verify(repoRoot, m)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if len(mismatches) != 0 {
		t.Errorf("expected 0 mismatches, got %d", len(mismatches))
		for _, mm := range mismatches {
			t.Logf("  %s: %s", mm.Reason, mm.Path)
		}
	}
}

func TestVerifyMissingFile(t *testing.T) {
	dir := t.TempDir()
	corpusDir := filepath.Join(dir, "tests", "vault", "corpus")
	if err := os.MkdirAll(corpusDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := []byte("hello")
	if err := os.WriteFile(
		filepath.Join(corpusDir, "file.txt"),
		content,
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	m := &Manifest{
		SpecVersion: 1,
		FileCount:   1,
		Files: []File{
			{
				Path:      "file.txt",
				SizeBytes: int64(len(content)),
				Digest:    fmt.Sprintf("sha256:%x", sha256.Sum256(content)),
			},
		},
	}

	mismatches, err := Verify(dir, m)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	if len(mismatches) != 0 {
		t.Errorf("expected 0 mismatches, got %d", len(mismatches))
		for _, mm := range mismatches {
			t.Logf("  %s: %s", mm.Reason, mm.Path)
		}
	}
}

func TestVerifyDigestMismatch(t *testing.T) {
	dir := t.TempDir()
	corpusDir := filepath.Join(dir, "tests", "vault", "corpus")
	if err := os.MkdirAll(corpusDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := []byte("test content")
	if err := os.WriteFile(
		filepath.Join(corpusDir, "test.md"),
		content,
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	m := &Manifest{
		SpecVersion: 1,
		FileCount:   1,
		Files: []File{
			{
				Path:      "test.md",
				SizeBytes: int64(len(content)),
				Digest:    "sha256:0000000000000000000000000000000000000000000000000000000000000000",
			},
		},
	}

	mismatches, err := Verify(dir, m)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	found := false
	for _, mm := range mismatches {
		if mm.Path == "test.md" && mm.Reason == "digest" {
			found = true
			if mm.Want != m.Files[0].Digest {
				t.Errorf("Want = %q, want %q", mm.Want, m.Files[0].Digest)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected 'digest' mismatch for test.md")
		for _, mm := range mismatches {
			t.Logf("  %s: %s", mm.Reason, mm.Path)
		}
	}
}

func TestVerifySizeMismatch(t *testing.T) {
	dir := t.TempDir()
	corpusDir := filepath.Join(dir, "tests", "vault", "corpus")
	if err := os.MkdirAll(corpusDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := []byte("test content")
	if err := os.WriteFile(
		filepath.Join(corpusDir, "test.md"),
		content,
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	m := &Manifest{
		SpecVersion: 1,
		FileCount:   1,
		Files: []File{
			{
				Path:      "test.md",
				SizeBytes: 999,
				Digest:    fmt.Sprintf("sha256:%x", sha256.Sum256(content)),
			},
		},
	}

	mismatches, err := Verify(dir, m)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	found := false
	for _, mm := range mismatches {
		if mm.Path == "test.md" && mm.Reason == "size" {
			found = true
			if mm.Want != "999" {
				t.Errorf("Want = %q, want %q", mm.Want, "999")
			}
			break
		}
	}
	if !found {
		t.Errorf("expected 'size' mismatch for test.md")
		for _, mm := range mismatches {
			t.Logf("  %s: %s", mm.Reason, mm.Path)
		}
	}
}

func TestVerifyMissingOnDisk(t *testing.T) {
	dir := t.TempDir()
	corpusDir := filepath.Join(dir, "tests", "vault", "corpus")
	if err := os.MkdirAll(corpusDir, 0o755); err != nil {
		t.Fatal(err)
	}

	m := &Manifest{
		SpecVersion: 1,
		FileCount:   1,
		Files: []File{
			{
				Path:      "nonexistent.md",
				SizeBytes: 10,
				Digest:    "sha256:0000000000000000000000000000000000000000000000000000000000000000",
			},
		},
	}

	mismatches, err := Verify(dir, m)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	found := false
	for _, mm := range mismatches {
		if mm.Path == "nonexistent.md" && mm.Reason == "missing" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'missing' mismatch for nonexistent.md")
		for _, mm := range mismatches {
			t.Logf("  %s: %s", mm.Reason, mm.Path)
		}
	}
}

func TestVerifyUnexpectedOnDisk(t *testing.T) {
	dir := t.TempDir()
	corpusDir := filepath.Join(dir, "tests", "vault", "corpus")
	if err := os.MkdirAll(corpusDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := []byte("existing")
	if err := os.WriteFile(
		filepath.Join(corpusDir, "existing.txt"),
		content,
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(corpusDir, "extra.txt"),
		[]byte("extra"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	m := &Manifest{
		SpecVersion: 1,
		FileCount:   1,
		Files: []File{
			{
				Path:      "existing.txt",
				SizeBytes: int64(len(content)),
				Digest:    fmt.Sprintf("sha256:%x", sha256.Sum256(content)),
			},
		},
	}

	mismatches, err := Verify(dir, m)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	found := false
	for _, mm := range mismatches {
		if mm.Path == "extra.txt" && mm.Reason == "unexpected" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'unexpected' mismatch for extra.txt")
		for _, mm := range mismatches {
			t.Logf("  %s: %s", mm.Reason, mm.Path)
		}
	}
}

func TestVerifySkipGitattributes(t *testing.T) {
	dir := t.TempDir()
	corpusDir := filepath.Join(dir, "tests", "vault", "corpus")
	if err := os.MkdirAll(corpusDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := []byte("test")
	if err := os.WriteFile(
		filepath.Join(corpusDir, "test.txt"),
		content,
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(corpusDir, ".gitattributes"),
		[]byte("* text=auto"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	m := &Manifest{
		SpecVersion: 1,
		FileCount:   1,
		Files: []File{
			{
				Path:      "test.txt",
				SizeBytes: int64(len(content)),
				Digest:    fmt.Sprintf("sha256:%x", sha256.Sum256(content)),
			},
		},
	}

	mismatches, err := Verify(dir, m)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	for _, mm := range mismatches {
		if mm.Path == ".gitattributes" {
			t.Errorf("should skip .gitattributes, got %s mismatch", mm.Reason)
		}
	}
}

func TestVerifySkipReadme(t *testing.T) {
	dir := t.TempDir()
	corpusDir := filepath.Join(dir, "tests", "vault", "corpus")
	subDir := filepath.Join(corpusDir, "regression")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := []byte("test")
	if err := os.WriteFile(
		filepath.Join(subDir, "test.txt"),
		content,
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(subDir, "README.md"),
		[]byte("# Regression"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	m := &Manifest{
		SpecVersion: 1,
		FileCount:   1,
		Files: []File{
			{
				Path:      "regression/test.txt",
				SizeBytes: int64(len(content)),
				Digest:    fmt.Sprintf("sha256:%x", sha256.Sum256(content)),
			},
		},
	}

	mismatches, err := Verify(dir, m)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	for _, mm := range mismatches {
		if mm.Path == "regression/README.md" {
			t.Errorf("should skip README.md, got %s mismatch", mm.Reason)
		}
	}
}

func TestVerifyEmptyManifest(t *testing.T) {
	dir := t.TempDir()
	corpusDir := filepath.Join(dir, "tests", "vault", "corpus")
	if err := os.MkdirAll(corpusDir, 0o755); err != nil {
		t.Fatal(err)
	}

	m := &Manifest{
		SpecVersion: 1,
		FileCount:   0,
		Files:       []File{},
	}

	_, err := Verify(dir, m)
	if err == nil {
		t.Error("expected error for empty manifest")
	}
}
