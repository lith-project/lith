package domain

import (
	"errors"
	"slices"
	"strings"
)

var (
	ErrInvalidDurableTarget = errors.New("domain: invalid durable target")
	ErrInvalidResolution    = errors.New("domain: invalid resolution")
)

// DurableTarget is sealed and cannot contain an offset block identity.
type DurableTarget interface {
	durableTarget()
	NoteID() NoteID
}
type NoteTarget struct{ noteID NoteID }

func NewNoteTarget(noteID NoteID) (NoteTarget, error) {
	if noteID.IsZero() {
		return NoteTarget{}, ErrInvalidDurableTarget
	}
	return NoteTarget{noteID: noteID}, nil
}
func (NoteTarget) durableTarget()   {}
func (t NoteTarget) NoteID() NoteID { return t.noteID }

type SectionTarget struct{ sectionID SectionID }

func NewSectionTarget(id SectionID) (SectionTarget, error) {
	if id.noteID.IsZero() {
		return SectionTarget{}, ErrInvalidDurableTarget
	}
	return SectionTarget{sectionID: id}, nil
}
func (SectionTarget) durableTarget()         {}
func (t SectionTarget) NoteID() NoteID       { return t.sectionID.NoteID() }
func (t SectionTarget) SectionID() SectionID { return t.sectionID }

// DurableBlockTarget can only be built from an explicit anchor identity.
type DurableBlockTarget struct {
	noteID NoteID
	block  AnchoredBlockID
}

func NewDurableBlockTarget(noteID NoteID, block AnchoredBlockID) (DurableBlockTarget, error) {
	if noteID.IsZero() || block.noteID.IsZero() || block.anchor.value == "" || noteID != block.noteID {
		return DurableBlockTarget{}, ErrInvalidAnchoredBlockID
	}
	return DurableBlockTarget{noteID: noteID, block: block}, nil
}
func (DurableBlockTarget) durableTarget()             {}
func (t DurableBlockTarget) NoteID() NoteID           { return t.noteID }
func (t DurableBlockTarget) BlockID() AnchoredBlockID { return t.block }

// Resolution is a sealed deterministic link-resolution result.
type Resolution interface{ resolution() }
type Resolved struct {
	target      DurableTarget
	diagnostics []Diagnostic
}

func NewResolved(target DurableTarget, diagnostics []Diagnostic) (Resolved, error) {
	if target == nil || target.NoteID().IsZero() {
		return Resolved{}, ErrInvalidResolution
	}
	return Resolved{target: target, diagnostics: slices.Clone(diagnostics)}, nil
}
func (Resolved) resolution()                 {}
func (r Resolved) Target() DurableTarget     { return r.target }
func (r Resolved) Diagnostics() []Diagnostic { return slices.Clone(r.diagnostics) }

// Ambiguous preserves all candidates sorted by NoteID and selects none.
type Ambiguous struct{ candidates []NoteID }

func NewAmbiguous(candidates []NoteID) (Ambiguous, error) {
	if len(candidates) < 2 {
		return Ambiguous{}, ErrInvalidResolution
	}
	result := slices.Clone(candidates)
	for _, id := range result {
		if id.IsZero() {
			return Ambiguous{}, ErrInvalidResolution
		}
	}
	slices.SortFunc(result, func(a, b NoteID) int { return strings.Compare(a.String(), b.String()) })
	for i := 1; i < len(result); i++ {
		if result[i-1] == result[i] {
			return Ambiguous{}, ErrInvalidResolution
		}
	}
	return Ambiguous{candidates: result}, nil
}
func (Ambiguous) resolution()            {}
func (r Ambiguous) Candidates() []NoteID { return slices.Clone(r.candidates) }

// Broken records that no deterministic local target exists.
type Broken struct{}

func NewBroken() Broken    { return Broken{} }
func (Broken) resolution() {}

// External records a target handled outside the Vault.
type External struct{ target string }

func NewExternal(target string) (External, error) {
	if target == "" {
		return External{}, ErrInvalidResolution
	}
	return External{target: target}, nil
}
func (External) resolution()      {}
func (r External) Target() string { return r.target }
