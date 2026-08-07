package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

const (
	SchemaVersion = "1"
	Tokenizer     = "unicode61"
)

type tableDefinition struct {
	Name       string
	Columns    []string
	NaturalKey []string
}

var schemaTables = []tableDefinition{
	{Name: "asset", Columns: []string{"asset_id", "content_hash", "kind", "raw_path"}, NaturalKey: []string{"asset_id"}},
	{Name: "block", Columns: []string{"anchor", "kind", "note_id", "range_end", "range_start"}, NaturalKey: []string{"note_id", "range_start"}},
	{Name: "diagnostic", Columns: []string{"code", "note_id", "range_end", "range_start", "severity"}, NaturalKey: []string{"note_id", "range_start", "code"}},
	{Name: "fm_entry", Columns: []string{"key", "note_id", "ordinal", "raw_value", "typed_kind"}, NaturalKey: []string{"note_id", "ordinal"}},
	{Name: "fts_row", Columns: []string{"body_text", "note_id"}, NaturalKey: []string{"note_id"}},
	{Name: "link", Columns: []string{"display", "kind", "note_id", "origin", "range_start", "subpath", "target_raw"}, NaturalKey: []string{"note_id", "range_start"}},
	{Name: "meta", Columns: []string{"built_at_unix", "built_by_version", "schema_version", "singleton", "tokenizer", "vault_fingerprint"}, NaturalKey: []string{"singleton"}},
	{Name: "note", Columns: []string{"content_hash", "encoding", "mtime_unix", "note_id", "raw_path", "size_bytes", "skipped_reason"}, NaturalKey: []string{"note_id"}},
	{Name: "resolution", Columns: []string{"candidates", "note_id", "outcome", "range_start", "target_id", "target_kind"}, NaturalKey: []string{"note_id", "range_start"}},
	{Name: "section", Columns: []string{"level", "note_id", "range_end", "range_start", "section_id"}, NaturalKey: []string{"note_id", "section_id"}},
	{Name: "tag", Columns: []string{"name", "name_folded", "note_id", "origin", "range_start"}, NaturalKey: []string{"note_id", "range_start"}},
	{Name: "task", Columns: []string{"note_id", "range_start", "state_kind", "state_raw"}, NaturalKey: []string{"note_id", "range_start"}},
}

const schemaDDL = `
CREATE TABLE IF NOT EXISTS meta (
    singleton INTEGER PRIMARY KEY CHECK (singleton = 1),
    schema_version TEXT NOT NULL,
    vault_fingerprint TEXT NOT NULL,
    tokenizer TEXT NOT NULL,
    built_by_version TEXT NOT NULL,
    built_at_unix INTEGER NOT NULL
) WITHOUT ROWID;
CREATE TABLE IF NOT EXISTS note (
    note_id TEXT PRIMARY KEY,
    raw_path TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    size_bytes INTEGER NOT NULL CHECK (size_bytes >= 0),
    mtime_unix INTEGER NOT NULL,
    encoding TEXT NOT NULL,
    skipped_reason TEXT
) WITHOUT ROWID;
CREATE TABLE IF NOT EXISTS asset (
    asset_id TEXT PRIMARY KEY,
    raw_path TEXT NOT NULL,
    kind TEXT NOT NULL,
    content_hash TEXT NOT NULL
) WITHOUT ROWID;
CREATE TABLE IF NOT EXISTS fm_entry (
    note_id TEXT NOT NULL,
    ordinal INTEGER NOT NULL CHECK (ordinal >= 0),
    key TEXT NOT NULL,
    raw_value TEXT NOT NULL,
    typed_kind TEXT NOT NULL,
    PRIMARY KEY (note_id, ordinal),
    FOREIGN KEY (note_id) REFERENCES note(note_id) ON DELETE CASCADE
) WITHOUT ROWID;
CREATE TABLE IF NOT EXISTS section (
    note_id TEXT NOT NULL,
    section_id TEXT NOT NULL,
    level INTEGER NOT NULL CHECK (level BETWEEN 1 AND 6),
    range_start INTEGER NOT NULL CHECK (range_start >= 0),
    range_end INTEGER NOT NULL CHECK (range_end >= range_start),
    PRIMARY KEY (note_id, section_id),
    FOREIGN KEY (note_id) REFERENCES note(note_id) ON DELETE CASCADE
) WITHOUT ROWID;
CREATE TABLE IF NOT EXISTS block (
    note_id TEXT NOT NULL,
    range_start INTEGER NOT NULL CHECK (range_start >= 0),
    range_end INTEGER NOT NULL CHECK (range_end >= range_start),
    kind TEXT NOT NULL,
    anchor TEXT,
    PRIMARY KEY (note_id, range_start),
    FOREIGN KEY (note_id) REFERENCES note(note_id) ON DELETE CASCADE
) WITHOUT ROWID;
CREATE TABLE IF NOT EXISTS link (
    note_id TEXT NOT NULL,
    range_start INTEGER NOT NULL CHECK (range_start >= 0),
    kind TEXT NOT NULL,
    target_raw TEXT NOT NULL,
    subpath TEXT,
    display TEXT,
    origin TEXT NOT NULL,
    PRIMARY KEY (note_id, range_start),
    FOREIGN KEY (note_id) REFERENCES note(note_id) ON DELETE CASCADE
) WITHOUT ROWID;
CREATE TABLE IF NOT EXISTS resolution (
    note_id TEXT NOT NULL,
    range_start INTEGER NOT NULL CHECK (range_start >= 0),
    outcome TEXT NOT NULL CHECK (outcome IN ('resolved', 'ambiguous', 'broken', 'external')),
    target_kind TEXT CHECK (target_kind IS NULL OR target_kind IN ('note', 'asset')),
    target_id TEXT,
    candidates TEXT,
    PRIMARY KEY (note_id, range_start),
    FOREIGN KEY (note_id, range_start) REFERENCES link(note_id, range_start) ON DELETE CASCADE,
    CHECK ((outcome = 'resolved' AND target_kind IS NOT NULL AND target_id IS NOT NULL) OR
           (outcome <> 'resolved' AND target_kind IS NULL AND target_id IS NULL))
) WITHOUT ROWID;
CREATE TABLE IF NOT EXISTS tag (
    note_id TEXT NOT NULL,
    range_start INTEGER NOT NULL CHECK (range_start >= 0),
    name TEXT NOT NULL,
    name_folded TEXT NOT NULL,
    origin TEXT NOT NULL,
    PRIMARY KEY (note_id, range_start),
    FOREIGN KEY (note_id) REFERENCES note(note_id) ON DELETE CASCADE
) WITHOUT ROWID;
CREATE TABLE IF NOT EXISTS task (
    note_id TEXT NOT NULL,
    range_start INTEGER NOT NULL CHECK (range_start >= 0),
    state_raw TEXT NOT NULL,
    state_kind TEXT NOT NULL,
    PRIMARY KEY (note_id, range_start),
    FOREIGN KEY (note_id) REFERENCES note(note_id) ON DELETE CASCADE
) WITHOUT ROWID;
CREATE TABLE IF NOT EXISTS diagnostic (
    note_id TEXT NOT NULL,
    range_start INTEGER NOT NULL CHECK (range_start >= 0),
    range_end INTEGER NOT NULL CHECK (range_end >= range_start),
    code TEXT NOT NULL,
    severity TEXT NOT NULL,
    PRIMARY KEY (note_id, range_start, code),
    FOREIGN KEY (note_id) REFERENCES note(note_id) ON DELETE CASCADE
) WITHOUT ROWID;
CREATE TABLE IF NOT EXISTS fts_row (
    note_id TEXT PRIMARY KEY,
    body_text TEXT NOT NULL,
    FOREIGN KEY (note_id) REFERENCES note(note_id) ON DELETE CASCADE
);
CREATE VIRTUAL TABLE IF NOT EXISTS fts_index USING fts5(note_id UNINDEXED, body_text, tokenize = 'unicode61');
`

func createSchema(ctx context.Context, db *sql.DB) error {
	if err := validateClassification(schemaTables, schemaClassification); err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store: begin schema transaction: %w", err)
	}
	rollback := func(operationErr error) error {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Join(operationErr, fmt.Errorf("store: rollback schema transaction: %w", rollbackErr))
		}
		return operationErr
	}
	if _, err := tx.ExecContext(ctx, schemaDDL); err != nil {
		return rollback(fmt.Errorf("store: create schema: %w", err))
	}
	if err := validatePhysicalSchema(ctx, tx); err != nil {
		return rollback(err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("store: commit schema transaction: %w", err)
	}
	return nil
}

type schemaQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func validatePhysicalSchema(ctx context.Context, db schemaQueryer) error {
	for _, table := range schemaTables {
		rows, err := db.QueryContext(ctx, "PRAGMA table_info("+table.Name+")")
		if err != nil {
			return fmt.Errorf("store: inspect schema table %s: %w", table.Name, err)
		}
		actual := make(map[string]struct{}, len(table.Columns))
		actualPrimaryKey := make(map[int]string, len(table.NaturalKey))
		for rows.Next() {
			var cid int
			var name, typ string
			var notNull, primaryKey int
			var defaultValue sql.NullString
			if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &primaryKey); err != nil {
				rows.Close()
				return fmt.Errorf("store: read schema table %s: %w", table.Name, err)
			}
			actual[name] = struct{}{}
			if primaryKey > 0 {
				actualPrimaryKey[primaryKey] = name
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return fmt.Errorf("store: iterate schema table %s: %w", table.Name, err)
		}
		if err := rows.Close(); err != nil {
			return fmt.Errorf("store: close schema table %s: %w", table.Name, err)
		}
		if len(actual) != len(table.Columns) {
			return fmt.Errorf("store: schema table %s has %d columns, want %d", table.Name, len(actual), len(table.Columns))
		}
		for _, column := range table.Columns {
			if _, ok := actual[column]; !ok {
				return fmt.Errorf("store: schema table %s is missing column %s", table.Name, column)
			}
		}
		if len(actualPrimaryKey) != len(table.NaturalKey) {
			return fmt.Errorf("store: schema table %s has %d primary-key columns, want %d", table.Name, len(actualPrimaryKey), len(table.NaturalKey))
		}
		for position, expected := range table.NaturalKey {
			if actual := actualPrimaryKey[position+1]; actual != expected {
				return fmt.Errorf("store: schema table %s primary-key column %d = %q, want %q", table.Name, position+1, actual, expected)
			}
		}
	}
	return nil
}

func requirePhysicalSchema(ctx context.Context, db schemaQueryer) error {
	if err := validatePhysicalSchema(ctx, db); err != nil {
		return &RebuildRequiredError{Reason: RebuildReasonSchemaIntegrity, Expected: "complete physical schema", Actual: err.Error()}
	}
	return nil
}
