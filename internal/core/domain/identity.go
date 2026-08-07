package domain

import (
	"errors"
	"slices"

	"github.com/lith-project/lith/internal/core/vaultpath"
)

var ErrInvalidSectionID = errors.New("domain: invalid section identity")

// NoteID is the case-preserving canonical vault-relative identity from vaultpath.
type NoteID struct{ value string }

func NewNoteID(path vaultpath.Path) NoteID { return NoteID{value: path.ID()} }
func (id NoteID) String() string           { return id.value }
func (id NoteID) IsZero() bool             { return id.value == "" }

// AssetID is the canonical vault-relative identity of a non-Markdown file.
type AssetID struct{ value string }

func NewAssetID(path vaultpath.Path) AssetID { return AssetID{value: path.ID()} }
func (id AssetID) String() string            { return id.value }
func (id AssetID) IsZero() bool              { return id.value == "" }

// SectionID identifies a heading path and one-based duplicate occurrence.
type SectionID struct {
	noteID      NoteID
	headingPath []string
	occurrence  uint
}

func NewSectionID(noteID NoteID, headingPath []string, occurrence uint) (SectionID, error) {
	if noteID.IsZero() || len(headingPath) == 0 || occurrence == 0 {
		return SectionID{}, ErrInvalidSectionID
	}
	for _, heading := range headingPath {
		if heading == "" {
			return SectionID{}, ErrInvalidSectionID
		}
	}
	return SectionID{noteID: noteID, headingPath: slices.Clone(headingPath), occurrence: occurrence}, nil
}
func (id SectionID) NoteID() NoteID        { return id.noteID }
func (id SectionID) HeadingPath() []string { return slices.Clone(id.headingPath) }
func (id SectionID) Occurrence() uint      { return id.occurrence }
