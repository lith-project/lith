package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// testRepoRoot returns the repo root directory (where this module lives).
func testRepoRoot(t *testing.T) string {
	t.Helper()
	// Walk up from this file to find go.mod
	_, thisFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(thisFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (go.mod not found)")
		}
		dir = parent
	}
}

func TestLoadGlossary(t *testing.T) {
	root := testRepoRoot(t)
	glossary, err := loadGlossary(filepath.Join(root, "docs", "glossary.md"))
	if err != nil {
		t.Fatalf("loadGlossary: %v", err)
	}

	// Check some known terms
	for _, term := range []string{
		"Vault", "Derived State", "Capability Registry", "Transaction Coordinator",
		"Knowledge Source", "Workspace", "Tenant", "Note", "Section", "Block",
		"Frontmatter", "Link", "Tag", "Task", "Capability", "Job",
		"Capability Catalog", "Conformance Assertion", "Plugin Host",
		"Composition Root", "Core Package", "Adapter Package", "Plugin Package",
		"Asset", "Diagnostic", "Opaque Block", "Canonical Dump",
		"Durable Column", "Volatile Column", "Vault Fingerprint",
		"Dirty Set", "Reconciliation Scan", "Resolution Key", "Watcher Gap",
		"Change Proposal", "Intent Journal", "Identity Key",
	} {
		if !glossary[term] {
			t.Errorf("glossary missing expected term %q", term)
		}
	}
}

func TestLoadGlossaryFromFixture(t *testing.T) {
	glossary, err := loadGlossary("testdata/glossary-sample.md")
	if err != nil {
		t.Fatalf("loadGlossary: %v", err)
	}

	expected := map[string]bool{
		"Vault":               true,
		"Derived State":       true,
		"Capability Registry": true,
	}

	if len(glossary) != len(expected) {
		t.Fatalf("glossary has %d terms, expected %d", len(glossary), len(expected))
	}
	for term := range expected {
		if !glossary[term] {
			t.Errorf("glossary missing expected term %q", term)
		}
	}
}

func TestLoadAllowlist(t *testing.T) {
	allowlist, err := loadAllowlist("testdata/allowlist-sample.txt")
	if err != nil {
		t.Fatalf("loadAllowlist: %v", err)
	}
	if !allowlist["GitHub"] {
		t.Error("allowlist missing 'GitHub'")
	}
	if len(allowlist) != 1 {
		t.Errorf("allowlist has %d terms, expected 1", len(allowlist))
	}
}

// scanTestFile is a helper that loads the test glossary and allowlist, then scans a file.
func scanTestFile(t *testing.T, glossaryPath, allowlistPath, filePath string) []Violation {
	t.Helper()
	glossary, err := loadGlossary(glossaryPath)
	if err != nil {
		t.Fatalf("loadGlossary: %v", err)
	}
	allowlist, err := loadAllowlist(allowlistPath)
	if err != nil {
		t.Fatalf("loadAllowlist: %v", err)
	}
	violations, err := scanFile(filePath, glossary, allowlist)
	if err != nil {
		t.Fatalf("scanFile: %v", err)
	}
	return violations
}

func TestScanClean(t *testing.T) {
	violations := scanTestFile(t, "testdata/glossary-sample.md", "testdata/allowlist-sample.txt", "testdata/clean.md")
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d:", len(violations))
		for _, v := range violations {
			t.Errorf("  %s:%d: %q", v.File, v.Line, v.Term)
		}
	}
}

func TestScanUndefined(t *testing.T) {
	violations := scanTestFile(t, "testdata/glossary-sample.md", "testdata/allowlist-sample.txt", "testdata/undefined.md")
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Term != "Frobulator" {
		t.Errorf("expected violation term 'Frobulator', got %q", violations[0].Term)
	}
}

func TestScanCodeIgnored(t *testing.T) {
	violations := scanTestFile(t, "testdata/glossary-sample.md", "testdata/allowlist-sample.txt", "testdata/code-ignored.md")
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d:", len(violations))
		for _, v := range violations {
			t.Errorf("  %s:%d: %q", v.File, v.Line, v.Term)
		}
	}
}

func TestScanMultipleViolations(t *testing.T) {
	violations := scanTestFile(t, "testdata/glossary-sample.md", "testdata/allowlist-sample.txt", "testdata/multi-violations.md")
	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}
	terms := make(map[string]bool)
	for _, v := range violations {
		terms[v.Term] = true
	}
	if !terms["Frobulator"] {
		t.Error("expected violation for 'Frobulator'")
	}
	if !terms["Gizmologist"] {
		t.Error("expected violation for 'Gizmologist'")
	}
}

func TestDeterminerStripping(t *testing.T) {
	// "A Derived State" → strip "A" → "Derived State" → found in glossary
	if got := stripDeterminers("A Derived State"); got != "Derived State" {
		t.Errorf("stripDeterminers('A Derived State') = %q, want 'Derived State'", got)
	}

	// "An Unknown Thing" → strip "An" → "Unknown Thing" → not in glossary
	if got := stripDeterminers("An Unknown Thing"); got != "Unknown Thing" {
		t.Errorf("stripDeterminers('An Unknown Thing') = %q, want 'Unknown Thing'", got)
	}

	// "The Vault" → strip "The" → "Vault" → found in glossary
	if got := stripDeterminers("The Vault"); got != "Vault" {
		t.Errorf("stripDeterminers('The Vault') = %q, want 'Vault'", got)
	}

	// No determiner → unchanged
	if got := stripDeterminers("Derived State"); got != "Derived State" {
		t.Errorf("stripDeterminers('Derived State') = %q, want 'Derived State'", got)
	}
}

func TestSentenceBoundarySkipped(t *testing.T) {
	// Create a test file with a word at sentence boundary
	content := `---
title: Test
---

# Heading

A Frobulator is a thing.
`

	tmpFile := filepath.Join(t.TempDir(), "test.md")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	glossary := map[string]bool{"Vault": true}
	allowlist := map[string]bool{}

	violations, err := scanFile(tmpFile, glossary, allowlist)
	if err != nil {
		t.Fatalf("scanFile: %v", err)
	}

	// "A" at sentence boundary should be skipped.
	// "Frobulator" mid-sentence should be flagged.
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Term != "Frobulator" {
		t.Errorf("expected violation term 'Frobulator', got %q", violations[0].Term)
	}
}

func TestFrontmatterExcluded(t *testing.T) {
	content := `---
title: Frobulator
status: Draft
---

# Heading

This describes a Vault.
`

	tmpFile := filepath.Join(t.TempDir(), "test.md")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	glossary := map[string]bool{"Vault": true}
	allowlist := map[string]bool{}

	violations, err := scanFile(tmpFile, glossary, allowlist)
	if err != nil {
		t.Fatalf("scanFile: %v", err)
	}

	// "Frobulator" in frontmatter should NOT be flagged.
	// "Vault" in body should pass (in glossary).
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d:", len(violations))
		for _, v := range violations {
			t.Errorf("  %s:%d: %q", v.File, v.Line, v.Term)
		}
	}
}

func TestReferencesExcluded(t *testing.T) {
	content := `# Heading

This describes a Vault.

## References

### External
- Frobulator Specification
`

	tmpFile := filepath.Join(t.TempDir(), "test.md")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	glossary := map[string]bool{"Vault": true}
	allowlist := map[string]bool{}

	violations, err := scanFile(tmpFile, glossary, allowlist)
	if err != nil {
		t.Fatalf("scanFile: %v", err)
	}

	// "Frobulator" in References section should NOT be flagged.
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d:", len(violations))
		for _, v := range violations {
			t.Errorf("  %s:%d: %q", v.File, v.Line, v.Term)
		}
	}
}

func TestTableRowsExcluded(t *testing.T) {
	content := `# Heading

| Column | Value |
| ------ | ----- |
| Frobulator | test |

This describes a Vault.
`

	tmpFile := filepath.Join(t.TempDir(), "test.md")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	glossary := map[string]bool{"Vault": true}
	allowlist := map[string]bool{}

	violations, err := scanFile(tmpFile, glossary, allowlist)
	if err != nil {
		t.Fatalf("scanFile: %v", err)
	}

	// "Frobulator" in table row should NOT be flagged.
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d:", len(violations))
		for _, v := range violations {
			t.Errorf("  %s:%d: %q", v.File, v.Line, v.Term)
		}
	}
}

func TestMarkdownBoldStripped(t *testing.T) {
	content := `# Heading

The **Frobulator** is a thing.
`

	tmpFile := filepath.Join(t.TempDir(), "test.md")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	glossary := map[string]bool{}
	allowlist := map[string]bool{}

	violations, err := scanFile(tmpFile, glossary, allowlist)
	if err != nil {
		t.Fatalf("scanFile: %v", err)
	}

	// "Frobulator" inside bold should still be flagged (bold is stripped, word remains).
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Term != "Frobulator" {
		t.Errorf("expected violation term 'Frobulator', got %q", violations[0].Term)
	}
}
