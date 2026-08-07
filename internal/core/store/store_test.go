package store

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
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
		assertNaturalKey(t, store.db, table)
	}
	if _, err := store.db.Exec("INSERT INTO meta (singleton, schema_version, vault_fingerprint, tokenizer, built_by_version, built_at_unix) VALUES (2, 'x', 'x', 'x', 'x', 0)"); err == nil {
		t.Fatal("insert a second meta row: expected singleton constraint failure")
	}
	if _, err := store.db.Exec("INSERT INTO section (note_id, section_id, level, range_start, range_end) VALUES ('missing.md', 'section', 1, 0, 1)"); err == nil {
		t.Fatal("insert child without note: expected foreign-key failure")
	}
	if _, err := store.db.Exec("INSERT INTO note (note_id, raw_path, content_hash, size_bytes, mtime_unix, encoding) VALUES ('negative.md', 'negative.md', 'hash', -1, 0, 'utf-8')"); err == nil {
		t.Fatal("insert note with negative size: expected CHECK constraint failure")
	}
	if _, err := store.db.Exec("INSERT INTO note (note_id, raw_path, content_hash, size_bytes, mtime_unix, encoding) VALUES ('valid.md', 'valid.md', 'hash', 0, 0, 'utf-8')"); err != nil {
		t.Fatalf("insert valid note for CHECK fixtures: %v", err)
	}
	if _, err := store.db.Exec("INSERT INTO section (note_id, section_id, level, range_start, range_end) VALUES ('valid.md', 'bad-level', 7, 0, 1)"); err == nil {
		t.Fatal("insert section with invalid level: expected CHECK constraint failure")
	}
	if _, err := store.db.Exec("INSERT INTO section (note_id, section_id, level, range_start, range_end) VALUES ('valid.md', 'bad-range', 1, 2, 1)"); err == nil {
		t.Fatal("insert section with descending range: expected CHECK constraint failure")
	}
}

func assertNaturalKey(t *testing.T, db *sql.DB, table tableDefinition) {
	t.Helper()
	rows, err := db.Query("PRAGMA table_info(" + table.Name + ")")
	if err != nil {
		t.Fatalf("inspect natural key for %s: %v", table.Name, err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			t.Errorf("close natural-key inspection for %s: %v", table.Name, closeErr)
		}
	}()
	actual := make(map[int]string, len(table.NaturalKey))
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, typ string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatalf("read natural key for %s: %v", table.Name, err)
		}
		if primaryKey > 0 {
			actual[primaryKey] = name
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate natural key for %s: %v", table.Name, err)
	}
	if len(actual) != len(table.NaturalKey) {
		t.Fatalf("natural key for %s = %#v, want %v", table.Name, actual, table.NaturalKey)
	}
	for position, expected := range table.NaturalKey {
		if actual := actual[position+1]; actual != expected {
			t.Fatalf("natural key for %s position %d = %q, want %q", table.Name, position+1, actual, expected)
		}
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

func TestOpen_rejectsDerivedStateSymlinkIntoVault(t *testing.T) {
	// Given
	vaultRoot := t.TempDir()
	derivedRoot := t.TempDir()
	if err := os.Symlink(vaultRoot, filepath.Join(derivedRoot, "lith")); err != nil {
		t.Fatalf("create derived-state symlink: %v", err)
	}

	// When
	_, err := Open(context.Background(), OpenOptions{VaultRoot: vaultRoot, DerivedStateRoot: derivedRoot, BuiltByVersion: "test"})

	// Then
	if !errors.Is(err, ErrDerivedStateInsideVault) {
		t.Fatalf("open through derived-state symlink = %v, want ErrDerivedStateInsideVault", err)
	}
}

func TestOpen_rejectsDanglingDerivedStateSymlinkIntoVault(t *testing.T) {
	// Given
	vaultRoot := t.TempDir()
	derivedRoot := t.TempDir()
	if err := os.Symlink(filepath.Join(vaultRoot, "future-derived"), filepath.Join(derivedRoot, "lith")); err != nil {
		t.Fatalf("create dangling derived-state symlink: %v", err)
	}

	// When
	_, err := Open(context.Background(), OpenOptions{VaultRoot: vaultRoot, DerivedStateRoot: derivedRoot, BuiltByVersion: "test"})

	// Then
	if !errors.Is(err, ErrDerivedStateInsideVault) {
		t.Fatalf("open through dangling derived-state symlink = %v, want ErrDerivedStateInsideVault", err)
	}
}

func TestOpen_rejectsDanglingDerivedStateRootSymlinkIntoVault(t *testing.T) {
	// Given
	vaultRoot := t.TempDir()
	derivedLink := filepath.Join(t.TempDir(), "derived-link")
	if err := os.Symlink(filepath.Join(vaultRoot, "future-derived"), derivedLink); err != nil {
		t.Fatalf("create dangling derived-root symlink: %v", err)
	}

	// When
	_, err := Open(context.Background(), OpenOptions{VaultRoot: vaultRoot, DerivedStateRoot: derivedLink, BuiltByVersion: "test"})

	// Then
	if !errors.Is(err, ErrDerivedStateInsideVault) {
		t.Fatalf("open through dangling derived-root symlink = %v, want ErrDerivedStateInsideVault", err)
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

func TestOpen_reportsCorruptSchemaWithMatchingMetadata(t *testing.T) {
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
	if _, err := db.Exec("DROP TABLE section"); err != nil {
		t.Fatalf("remove section from schema fixture: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close test setup database: %v", err)
	}

	// When
	_, err = Open(context.Background(), options)

	// Then
	if err == nil {
		t.Fatal("open corrupted schema unexpectedly succeeded")
	}
	var rebuildErr *RebuildRequiredError
	if !errors.As(err, &rebuildErr) {
		t.Fatalf("corrupt schema error = %v, want RebuildRequiredError", err)
	}
	if rebuildErr.Reason != RebuildReasonSchemaIntegrity {
		t.Fatalf("corrupt schema rebuild reason = %q, want %q", rebuildErr.Reason, RebuildReasonSchemaIntegrity)
	}
	if !strings.Contains(err.Error(), "section") {
		t.Fatalf("corrupt schema error = %v, want section context", err)
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
