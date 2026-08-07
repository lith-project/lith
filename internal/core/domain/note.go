package domain

import (
	"errors"
	"slices"
)

var (
	ErrInvalidFrontmatter = errors.New("domain: invalid frontmatter")
	ErrInvalidNote        = errors.New("domain: invalid note")
	ErrInvalidSkipped     = errors.New("domain: invalid skipped outcome")
)

// FrontmatterEntry retains key/value text and both source ranges.
type FrontmatterEntry struct {
	key, rawValue        string
	rangeVal, valueRange Range
}
type FrontmatterEntryInput struct {
	Key        string
	RawValue   string
	Range      Range
	ValueRange Range
}

func NewFrontmatterEntry(input FrontmatterEntryInput) (FrontmatterEntry, error) {
	if input.Key == "" || !input.Range.Encloses(input.ValueRange) {
		return FrontmatterEntry{}, ErrInvalidFrontmatter
	}
	return FrontmatterEntry{key: input.Key, rawValue: input.RawValue, rangeVal: input.Range, valueRange: input.ValueRange}, nil
}
func (e FrontmatterEntry) Key() string       { return e.key }
func (e FrontmatterEntry) RawValue() string  { return e.rawValue }
func (e FrontmatterEntry) Range() Range      { return e.rangeVal }
func (e FrontmatterEntry) ValueRange() Range { return e.valueRange }

// Frontmatter is immutable schema-free metadata.
type Frontmatter struct {
	rangeVal      Range
	entries       []FrontmatterEntry
	tags, aliases []string
}
type FrontmatterInput struct {
	Range   Range
	Entries []FrontmatterEntry
	Tags    []string
	Aliases []string
}

func NewFrontmatter(input FrontmatterInput) Frontmatter {
	return Frontmatter{rangeVal: input.Range, entries: slices.Clone(input.Entries), tags: slices.Clone(input.Tags), aliases: slices.Clone(input.Aliases)}
}
func (f Frontmatter) Range() Range                { return f.rangeVal }
func (f Frontmatter) Entries() []FrontmatterEntry { return slices.Clone(f.entries) }
func (f Frontmatter) Tags() []string              { return slices.Clone(f.tags) }
func (f Frontmatter) Aliases() []string           { return slices.Clone(f.aliases) }
func (f Frontmatter) clone() Frontmatter {
	return NewFrontmatter(FrontmatterInput{Range: f.rangeVal, Entries: f.entries, Tags: f.tags, Aliases: f.aliases})
}

// NoteParts contains parse children copied into a Note.
type NoteParts struct {
	Frontmatter *Frontmatter
	Sections    []Section
	Blocks      []Block
	Links       []Link
	Tags        []Tag
	Tasks       []Task
	Diagnostics []Diagnostic
}

// Note is an immutable parse result with no serializer or writer API.
type Note struct {
	id          NoteID
	rangeVal    Range
	frontmatter *Frontmatter
	sections    []Section
	blocks      []Block
	links       []Link
	tags        []Tag
	tasks       []Task
	diagnostics []Diagnostic
}

func NewNote(id NoteID, rangeVal Range, parts NoteParts) (Note, error) {
	if id.IsZero() || rangeVal.Start() != 0 {
		return Note{}, ErrInvalidNote
	}
	var frontmatter *Frontmatter
	if parts.Frontmatter != nil {
		value := parts.Frontmatter.clone()
		frontmatter = &value
	}
	return Note{id: id, rangeVal: rangeVal, frontmatter: frontmatter, sections: slices.Clone(parts.Sections), blocks: slices.Clone(parts.Blocks), links: slices.Clone(parts.Links), tags: slices.Clone(parts.Tags), tasks: slices.Clone(parts.Tasks), diagnostics: slices.Clone(parts.Diagnostics)}, nil
}
func (Note) parseOutcome()  {}
func (n Note) ID() NoteID   { return n.id }
func (n Note) Range() Range { return n.rangeVal }
func (n Note) Frontmatter() (Frontmatter, bool) {
	if n.frontmatter == nil {
		return Frontmatter{}, false
	}
	return n.frontmatter.clone(), true
}
func (n Note) Sections() []Section       { return slices.Clone(n.sections) }
func (n Note) Blocks() []Block           { return slices.Clone(n.blocks) }
func (n Note) Links() []Link             { return slices.Clone(n.links) }
func (n Note) Tags() []Tag               { return slices.Clone(n.tags) }
func (n Note) Tasks() []Task             { return slices.Clone(n.tasks) }
func (n Note) Diagnostics() []Diagnostic { return slices.Clone(n.diagnostics) }

// SkipReason identifies a total-parser skip outcome.
type SkipReason string

const BinarySource SkipReason = "binary"

type Skipped struct {
	id          NoteID
	reason      SkipReason
	rangeVal    Range
	diagnostics []Diagnostic
}
type SkippedInput struct {
	Reason      SkipReason
	Range       Range
	Diagnostics []Diagnostic
}

func NewSkipped(id NoteID, input SkippedInput) (Skipped, error) {
	if id.IsZero() || input.Reason != BinarySource || input.Range.Start() != 0 {
		return Skipped{}, ErrInvalidSkipped
	}
	return Skipped{id: id, reason: input.Reason, rangeVal: input.Range, diagnostics: slices.Clone(input.Diagnostics)}, nil
}
func (Skipped) parseOutcome()               {}
func (s Skipped) ID() NoteID                { return s.id }
func (s Skipped) Reason() SkipReason        { return s.reason }
func (s Skipped) Range() Range              { return s.rangeVal }
func (s Skipped) Diagnostics() []Diagnostic { return slices.Clone(s.diagnostics) }

// ParseOutcome is the sealed total result: Note or Skipped.
type ParseOutcome interface{ parseOutcome() }
