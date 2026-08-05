package vaultpath

import "golang.org/x/text/unicode/norm"

// normalizeID applies NFC normalization to the vault-relative identity.
// NFC is the canonical composed form; two paths differing only by
// normal form resolve to one identity. The raw on-disk bytes are
// never normalised — the filesystem is the authority on how to open
// the file.
//
// Normalize is applied after the relative-path computation but the
// order is irrelevant to the result; this ordering is fixed and
// documented so two implementations cannot disagree.
func normalizeID(id string) string {
	return norm.NFC.String(id)
}
