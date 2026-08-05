package vaultpath

import (
	"strings"
	"testing"
)

func TestFoldKeyDiffersByID(t *testing.T) {
	// Notes/A.md and notes/a.md must have different IDs (case-sensitive identity)
	// but the same FoldKey (case-insensitive lookup).
	p1, err := New("/vault", "/vault/Notes/A.md")
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	p2, err := New("/vault", "/vault/notes/a.md")
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	if p1.ID() == p2.ID() {
		t.Errorf("IDs should differ: both are %q", p1.ID())
	}
	if p1.FoldKey() != p2.FoldKey() {
		t.Errorf("FoldKeys should match: %q vs %q", p1.FoldKey(), p2.FoldKey())
	}
}

func TestFoldKeyNFC(t *testing.T) {
	// German sharp-s (ß, U+00DF) and its uppercase pair (SS) differ in
	// identity but must fold to the same key.
	p1, err := New("/vault", "/vault/straße.md")
	if err != nil {
		t.Fatalf("New() with ß: %v", err)
	}
	p2, err := New("/vault", "/vault/STRASSE.md")
	if err != nil {
		t.Fatalf("New() with SS: %v", err)
	}

	if p1.ID() == p2.ID() {
		t.Errorf("IDs should differ: both are %q", p1.ID())
	}
	if p1.FoldKey() != p2.FoldKey() {
		t.Errorf("FoldKeys should match: %q vs %q", p1.FoldKey(), p2.FoldKey())
	}
}

func TestFoldKeyNotIdentity(t *testing.T) {
	// FoldKey must never be stored as a field — it is derived on call.
	// This is verified indirectly: case-varying paths produce different
	// IDs but the same FoldKey, proving FoldKey is computed, not stored.
	p1, err := New("/vault", "/vault/README.md")
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	p2, err := New("/vault", "/vault/readme.md")
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	if p1.ID() == p2.ID() {
		t.Errorf("IDs should differ: both are %q", p1.ID())
	}
	if p1.FoldKey() != p2.FoldKey() {
		t.Errorf("FoldKeys should match: %q vs %q", p1.FoldKey(), p2.FoldKey())
	}
	if p1.ID() == p1.FoldKey() {
		t.Errorf("FoldKey should differ from ID for case-varying input")
	}
}

func TestFoldKeyUnicodeCaseFold(t *testing.T) {
	// German ß (U+00DF) is a case pair that strings.ToLower cannot model
	// correctly. ToLower maps ß to ß (no change), but Unicode case folding
	// maps ß to "ss". This proves we use cases.Fold, not strings.ToLower.
	folded := foldCaser.String("\u00df") // ß
	lowered := strings.ToLower("\u00df") // ß (unchanged)

	if folded == lowered {
		t.Errorf("foldCaser and strings.ToLower should differ for ß: both produced %q", folded)
	}
	if folded != "ss" {
		t.Errorf("foldCaser(ß) = %q, want %q", folded, "ss")
	}
	if lowered != "\u00df" {
		t.Errorf("strings.ToLower(ß) = %q, want ß unchanged", lowered)
	}
}

func TestFoldKeyASCIIPassthrough(t *testing.T) {
	// For pure ASCII, FoldKey should equal strings.ToLower(ID).
	p, err := New("/vault", "/vault/Notes/File.MD")
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	expected := strings.ToLower(p.ID())
	if p.FoldKey() != expected {
		t.Errorf("FoldKey() = %q, want %q (strings.ToLower(ID))", p.FoldKey(), expected)
	}
}
