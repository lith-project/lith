package store

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOpen_initializesVaultExternalStoreWithSQLiteInvariants(t *testing.T) {
	// Given
	vaultRoot := t.TempDir()
	derivedRoot := t.TempDir()
	options := OpenOptions{VaultRoot: vaultRoot, DerivedStateRoot: derivedRoot, BuiltByVersion: "test"}

	// When
	store, err := Open(context.Background(), options)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := store.Close(); closeErr != nil {
			t.Errorf("close store: %v", closeErr)
		}
	})

	// Then
	if !isOutsideRoot(store.location.Directory, vaultRoot) {
		t.Fatalf("store directory %q is inside vault %q", store.location.Directory, vaultRoot)
	}
	if _, err := os.Stat(store.location.Database); err != nil {
		t.Fatalf("stat database: %v", err)
	}
	var journalMode string
	if err := store.db.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatalf("read journal mode: %v", err)
	}
	if journalMode != "wal" {
		t.Fatalf("journal mode = %q, want wal", journalMode)
	}
	var foreignKeys int
	if err := store.db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatalf("read foreign-key configuration: %v", err)
	}
	if foreignKeys != 1 {
		t.Fatalf("foreign keys = %d, want 1", foreignKeys)
	}
	for _, table := range schemaTables {
		var found int
		err := store.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", table.Name).Scan(&found)
		if err != nil {
			t.Fatalf("check table %s: %v", table.Name, err)
		}
		if found != 1 {
			t.Fatalf("table %s was not created", table.Name)
		}
	}
	if _, err := store.db.Exec("INSERT INTO meta (singleton, schema_version, vault_fingerprint, tokenizer, built_by_version, built_at_unix) VALUES (2, 'x', 'x', 'x', 'x', 0)"); err == nil {
		t.Fatal("insert a second meta row: expected singleton constraint failure")
	}
	if _, err := store.db.Exec("INSERT INTO section (note_id, section_id, level, range_start, range_end) VALUES ('missing.md', 'section', 1, 0, 1)"); err == nil {
		t.Fatal("insert child without note: expected foreign-key failure")
	}
}

func TestOpen_rejectsDerivedStateRootInsideVault(t *testing.T) {
	// Given
	vaultRoot := t.TempDir()
	options := OpenOptions{VaultRoot: vaultRoot, DerivedStateRoot: filepath.Join(vaultRoot, "derived"), BuiltByVersion: "test"}

	// When
	_, err := Open(context.Background(), options)

	// Then
	if !errors.Is(err, ErrDerivedStateInsideVault) {
		t.Fatalf("open error = %v, want ErrDerivedStateInsideVault", err)
	}
}

func TestOpen_serializesWritersAndReleasesOnClose(t *testing.T) {
	// Given
	options := OpenOptions{VaultRoot: t.TempDir(), DerivedStateRoot: t.TempDir(), BuiltByVersion: "test"}
	first, err := Open(context.Background(), options)
	if err != nil {
		t.Fatalf("open first writer: %v", err)
	}

	// When
	_, err = Open(context.Background(), options)

	// Then
	if !errors.Is(err, ErrWriterLocked) {
		t.Fatalf("open concurrent writer error = %v, want ErrWriterLocked", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("close first writer: %v", err)
	}
	second, err := Open(context.Background(), options)
	if err != nil {
		t.Fatalf("open writer after close: %v", err)
	}
	if err := second.Close(); err != nil {
		t.Fatalf("close second writer: %v", err)
	}
}

func TestOpen_readOnlyAccessDoesNotClaimWriterLock(t *testing.T) {
	// Given
	options := OpenOptions{VaultRoot: t.TempDir(), DerivedStateRoot: t.TempDir(), BuiltByVersion: "test"}
	writer, err := Open(context.Background(), options)
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := writer.Close(); closeErr != nil {
			t.Errorf("close writer: %v", closeErr)
		}
	})

	// When
	readerOptions := options
	readerOptions.ReadOnly = true
	reader, err := Open(context.Background(), readerOptions)

	// Then
	if err != nil {
		t.Fatalf("open read-only store while writer is active: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close read-only store: %v", err)
	}
}

func TestOpen_reportsSchemaVersionMismatchWithoutMutation(t *testing.T) {
	// Given
	options := OpenOptions{VaultRoot: t.TempDir(), DerivedStateRoot: t.TempDir(), BuiltByVersion: "test"}
	store, err := Open(context.Background(), options)
	if err != nil {
		t.Fatalf("open initial store: %v", err)
	}
	location := store.location
	if err := store.Close(); err != nil {
		t.Fatalf("close initial store: %v", err)
	}
	db, err := sql.Open(sqliteDriverName, location.Database)
	if err != nil {
		t.Fatalf("open database for test setup: %v", err)
	}
	if _, err := db.Exec("UPDATE meta SET schema_version = ?", "future"); err != nil {
		t.Fatalf("stamp future schema version: %v", err)
	}
	if _, err := db.Exec("DROP TABLE section"); err != nil {
		t.Fatalf("remove a table from the old schema fixture: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close test setup database: %v", err)
	}

	// When
	_, err = Open(context.Background(), options)

	// Then
	var rebuildErr *RebuildRequiredError
	if !errors.As(err, &rebuildErr) {
		t.Fatalf("open error = %v, want RebuildRequiredError", err)
	}
	if rebuildErr.Reason != RebuildReasonSchemaVersion {
		t.Fatalf("rebuild reason = %q, want %q", rebuildErr.Reason, RebuildReasonSchemaVersion)
	}
	db, err = sql.Open(sqliteDriverName, location.Database)
	if err != nil {
		t.Fatalf("re-open database to verify no migration: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("close verification database: %v", closeErr)
		}
	})
	var version string
	if err := db.QueryRow("SELECT schema_version FROM meta WHERE singleton = 1").Scan(&version); err != nil {
		t.Fatalf("read schema version: %v", err)
	}
	if version != "future" {
		t.Fatalf("schema version = %q, want future; store must not migrate in place", version)
	}
	var sectionCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'section'").Scan(&sectionCount); err != nil {
		t.Fatalf("check section table after mismatch: %v", err)
	}
	if sectionCount != 0 {
		t.Fatal("schema mismatch recreated a missing table")
	}
}

func TestValidateClassification_rejectsMissingAndExtraColumns(t *testing.T) {
	// Given
	manifest := cloneClassificationManifest(schemaClassification)
	delete(manifest["note"], "content_hash")

	// When
	err := validateClassification(schemaTables, manifest)

	// Then
	if err == nil {
		t.Fatal("validate classification with missing column: expected error")
	}
	manifest = cloneClassificationManifest(schemaClassification)
	manifest["note"]["not_a_column"] = Durable

	// When
	err = validateClassification(schemaTables, manifest)

	// Then
	if err == nil {
		t.Fatal("validate classification with extra column: expected error")
	}
}
