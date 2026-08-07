package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	// Register the SQLite driver for database/sql.
	_ "modernc.org/sqlite"

	"github.com/lith-project/lith/internal/core/daemon"
)

const sqliteDriverName = "sqlite"

var ErrWriterLocked = errors.New("store: writer is already claimed")

type RebuildReason string

const (
	RebuildReasonSchemaVersion   RebuildReason = "schema_version"
	RebuildReasonFingerprint     RebuildReason = "vault_fingerprint"
	RebuildReasonTokenizer       RebuildReason = "tokenizer"
	RebuildReasonSchemaIntegrity RebuildReason = "schema_integrity"
)

type RebuildRequiredError struct {
	Reason   RebuildReason
	Expected string
	Actual   string
}

func (e *RebuildRequiredError) Error() string {
	return fmt.Sprintf("store: rebuild required (%s: expected %q, got %q)", e.Reason, e.Expected, e.Actual)
}

type OpenOptions struct {
	VaultRoot        string
	DerivedStateRoot string
	BuiltByVersion   string
	ReadOnly         bool
}

type Store struct {
	db       *sql.DB
	lock     *daemon.Lock
	location Location
}

func Open(ctx context.Context, options OpenOptions) (_ *Store, err error) {
	location, err := resolveLocation(options.DerivedStateRoot, options.VaultRoot)
	if err != nil {
		return nil, err
	}
	if !options.ReadOnly {
		if err := os.MkdirAll(location.Directory, 0o755); err != nil {
			return nil, fmt.Errorf("store: create derived state directory: %w", err)
		}
	}

	var lock *daemon.Lock
	if !options.ReadOnly {
		lock, err = daemon.Acquire(options.VaultRoot, location.WriterLock, slog.Default())
		if errors.Is(err, daemon.ErrLocked) {
			return nil, fmt.Errorf("%w: %w", ErrWriterLocked, err)
		}
		if err != nil {
			return nil, fmt.Errorf("store: claim writer: %w", err)
		}
	}

	dsn := location.Database
	if options.ReadOnly {
		dsn = "file:" + dsn + "?mode=ro"
	}
	db, err := sql.Open(sqliteDriverName, dsn)
	if err != nil {
		if lock != nil {
			if releaseErr := lock.Release(); releaseErr != nil {
				return nil, errors.Join(fmt.Errorf("store: open sqlite: %w", err), fmt.Errorf("store: release writer: %w", releaseErr))
			}
		}
		return nil, fmt.Errorf("store: open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	store := &Store{db: db, lock: lock, location: location}
	defer func() {
		if err != nil {
			store.Close()
		}
	}()
	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("store: ping sqlite: %w", err)
	}
	if err = configureConnection(ctx, db, options.ReadOnly); err != nil {
		return nil, err
	}
	if options.ReadOnly {
		if err = verifyMetadata(ctx, db, location); err != nil {
			return nil, err
		}
	} else if err = initialize(ctx, db, location, options.BuiltByVersion); err != nil {
		return nil, err
	}
	return store, nil
}

func configureConnection(ctx context.Context, db *sql.DB, readOnly bool) error {
	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout = 5000"); err != nil {
		return fmt.Errorf("store: configure busy timeout: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("store: enable foreign keys: %w", err)
	}
	if readOnly {
		return nil
	}
	var mode string
	if err := db.QueryRowContext(ctx, "PRAGMA journal_mode = WAL").Scan(&mode); err != nil {
		return fmt.Errorf("store: enable WAL: %w", err)
	}
	if mode != "wal" {
		return fmt.Errorf("store: SQLite selected journal mode %q, want wal", mode)
	}
	return nil
}

func initialize(ctx context.Context, db *sql.DB, location Location, builtByVersion string) error {
	var metaTableExists int
	if err := db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = 'meta')").Scan(&metaTableExists); err != nil {
		return fmt.Errorf("store: inspect metadata table: %w", err)
	}
	if metaTableExists == 1 {
		var version, fingerprint, tokenizer string
		err := db.QueryRowContext(ctx, "SELECT schema_version, vault_fingerprint, tokenizer FROM meta WHERE singleton = 1").Scan(&version, &fingerprint, &tokenizer)
		if err == nil {
			if err := metadataError(version, fingerprint, tokenizer, location); err != nil {
				return err
			}
			return requirePhysicalSchema(ctx, db)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("store: read metadata: %w", err)
		}
	}
	if err := createSchema(ctx, db); err != nil {
		return err
	}
	var version, fingerprint, tokenizer string
	err := db.QueryRowContext(ctx, "SELECT schema_version, vault_fingerprint, tokenizer FROM meta WHERE singleton = 1").Scan(&version, &fingerprint, &tokenizer)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		if builtByVersion == "" {
			builtByVersion = "unknown"
		}
		_, err = db.ExecContext(ctx, "INSERT INTO meta (singleton, schema_version, vault_fingerprint, tokenizer, built_by_version, built_at_unix) VALUES (1, ?, ?, ?, ?, ?)", SchemaVersion, location.VaultFingerprint, Tokenizer, builtByVersion, time.Now().Unix())
		if err != nil {
			return fmt.Errorf("store: initialize metadata: %w", err)
		}
		return nil
	case err != nil:
		return fmt.Errorf("store: read metadata: %w", err)
	default:
		return metadataError(version, fingerprint, tokenizer, location)
	}
}

func metadataError(version, fingerprint, tokenizer string, location Location) error {
	if version != SchemaVersion {
		return &RebuildRequiredError{Reason: RebuildReasonSchemaVersion, Expected: SchemaVersion, Actual: version}
	}
	if fingerprint != location.VaultFingerprint {
		return &RebuildRequiredError{Reason: RebuildReasonFingerprint, Expected: location.VaultFingerprint, Actual: fingerprint}
	}
	if tokenizer != Tokenizer {
		return &RebuildRequiredError{Reason: RebuildReasonTokenizer, Expected: Tokenizer, Actual: tokenizer}
	}
	return nil
}

func verifyMetadata(ctx context.Context, db *sql.DB, location Location) error {
	var version, fingerprint, tokenizer string
	if err := db.QueryRowContext(ctx, "SELECT schema_version, vault_fingerprint, tokenizer FROM meta WHERE singleton = 1").Scan(&version, &fingerprint, &tokenizer); err != nil {
		return fmt.Errorf("store: read metadata: %w", err)
	}
	if err := metadataError(version, fingerprint, tokenizer, location); err != nil {
		return err
	}
	return requirePhysicalSchema(ctx, db)
}

func (s *Store) Close() error {
	var errs []error
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("store: close sqlite: %w", err))
		}
		s.db = nil
	}
	if s.lock != nil {
		if err := s.lock.Release(); err != nil {
			errs = append(errs, fmt.Errorf("store: release writer: %w", err))
		}
		s.lock = nil
	}
	return errors.Join(errs...)
}
