package store

import (
	"fmt"
	"sort"
)

type ColumnClass string

const (
	Durable  ColumnClass = "durable"
	Volatile ColumnClass = "volatile"
)

var schemaClassification = map[string]map[string]ColumnClass{
	"asset":      {"asset_id": Durable, "content_hash": Durable, "kind": Durable, "raw_path": Volatile},
	"block":      {"anchor": Durable, "kind": Durable, "note_id": Durable, "range_end": Durable, "range_start": Durable},
	"diagnostic": {"code": Durable, "note_id": Durable, "range_end": Durable, "range_start": Durable, "severity": Durable},
	"fm_entry":   {"key": Durable, "note_id": Durable, "ordinal": Durable, "raw_value": Durable, "typed_kind": Durable},
	"fts_row":    {"body_text": Durable, "note_id": Durable},
	"link":       {"display": Durable, "kind": Durable, "note_id": Durable, "origin": Durable, "range_start": Durable, "subpath": Durable, "target_raw": Durable},
	"meta":       {"built_at_unix": Volatile, "built_by_version": Volatile, "schema_version": Durable, "singleton": Durable, "tokenizer": Durable, "vault_fingerprint": Durable},
	"note":       {"content_hash": Durable, "encoding": Durable, "mtime_unix": Volatile, "note_id": Durable, "raw_path": Volatile, "size_bytes": Durable, "skipped_reason": Durable},
	"resolution": {"candidates": Durable, "note_id": Durable, "outcome": Durable, "range_start": Durable, "target_id": Durable, "target_kind": Durable},
	"section":    {"level": Durable, "note_id": Durable, "range_end": Durable, "range_start": Durable, "section_id": Durable},
	"tag":        {"name": Durable, "name_folded": Durable, "note_id": Durable, "origin": Durable, "range_start": Durable},
	"task":       {"note_id": Durable, "range_start": Durable, "state_kind": Durable, "state_raw": Durable},
}

func validateClassification(tables []tableDefinition, manifest map[string]map[string]ColumnClass) error {
	var problems []string
	for _, table := range tables {
		columns, ok := manifest[table.Name]
		if !ok {
			problems = append(problems, "missing table "+table.Name)
			continue
		}
		for _, column := range table.Columns {
			class, ok := columns[column]
			if !ok {
				problems = append(problems, "missing column "+table.Name+"."+column)
				continue
			}
			if class != Durable && class != Volatile {
				problems = append(problems, "invalid class "+table.Name+"."+column)
			}
		}
		for column := range columns {
			if !contains(table.Columns, column) {
				problems = append(problems, "extra column "+table.Name+"."+column)
			}
		}
	}
	for table := range manifest {
		if !containsTable(tables, table) {
			problems = append(problems, "extra table "+table)
		}
	}
	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("store: schema classification mismatch: %v", problems)
}

func cloneClassificationManifest(manifest map[string]map[string]ColumnClass) map[string]map[string]ColumnClass {
	clone := make(map[string]map[string]ColumnClass, len(manifest))
	for table, columns := range manifest {
		clone[table] = make(map[string]ColumnClass, len(columns))
		for column, class := range columns {
			clone[table][column] = class
		}
	}
	return clone
}

func contains(columns []string, candidate string) bool {
	for _, column := range columns {
		if column == candidate {
			return true
		}
	}
	return false
}

func containsTable(tables []tableDefinition, candidate string) bool {
	for _, table := range tables {
		if table.Name == candidate {
			return true
		}
	}
	return false
}
