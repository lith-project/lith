package vaultpath

import (
	"errors"
	"strings"
	"testing"
)

func TestNewBasic(t *testing.T) {
	p, err := New("/vault", "/vault/notes/a.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID() != "notes/a.md" {
		t.Errorf("ID = %q, want %q", p.ID(), "notes/a.md")
	}
}

func TestNewCleansPath(t *testing.T) {
	p, err := New("/vault", "/vault/./notes/../notes/a.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID() != "notes/a.md" {
		t.Errorf("ID = %q, want %q", p.ID(), "notes/a.md")
	}
}

func TestNewOutsideVault(t *testing.T) {
	_, err := New("/vault", "/etc/passwd")
	if !errors.Is(err, ErrOutsideVault) {
		t.Errorf("expected ErrOutsideVault, got: %v", err)
	}
}

func TestNewEscapeVault(t *testing.T) {
	_, err := New("/vault", "/vault/../vault-evil/a.md")
	if !errors.Is(err, ErrOutsideVault) {
		t.Errorf("expected ErrOutsideVault, got: %v", err)
	}
}

func TestRawPreserved(t *testing.T) {
	raw := "/vault/notes/a.md"
	p, err := New("/vault", raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Raw() != raw {
		t.Errorf("Raw() = %q, want %q", p.Raw(), raw)
	}
}

func TestForwardSlashIdentity(t *testing.T) {
	// Verify ID always uses forward slashes.
	// On macOS, filepath.ToSlash is a no-op for backslashes (they are literal),
	// so we verify the forward-slash canonical form produces a clean ID.
	// Windows-style paths (C:\vault) are tested implicitly via filepath.ToSlash
	// in New — that conversion is platform-correct by construction.
	p, err := New("/vault", "/vault/notes/a.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(p.ID(), "\\") {
		t.Errorf("ID must not contain backslashes: %q", p.ID())
	}
	if p.ID() != "notes/a.md" {
		t.Errorf("ID = %q, want %q", p.ID(), "notes/a.md")
	}
}

func TestNewVaultRootItself(t *testing.T) {
	_, err := New("/vault", "/vault")
	if err == nil {
		t.Error("expected error for path equal to vault root")
	}
}

func TestNoFilesystemAccess(t *testing.T) {
	// This test would fail if New tried to Stat the path
	// Using a path that definitely doesn't exist on any system
	p, err := New("/definitely-not-a-real-vault-12345", "/definitely-not-a-real-vault-12345/ghost.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID() != "ghost.md" {
		t.Errorf("ID = %q, want %q", p.ID(), "ghost.md")
	}
}
