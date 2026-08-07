package domain

import "errors"

var (
	ErrInvalidBlock           = errors.New("domain: invalid block")
	ErrInvalidAnchoredBlockID = errors.New("domain: invalid anchored block identity")
	ErrInvalidAsset           = errors.New("domain: invalid asset")
)

// Section is a heading-derived structural subdivision.
type Section struct {
	id       SectionID
	level    uint8
	heading  string
	rangeVal Range
}
type SectionInput struct {
	Level   uint8
	Heading string
	Range   Range
}

func NewSection(id SectionID, input SectionInput) (Section, error) {
	if id.noteID.IsZero() || input.Level < 1 || input.Level > 6 || input.Heading == "" {
		return Section{}, ErrInvalidSectionID
	}
	return Section{id: id, level: input.Level, heading: input.Heading, rangeVal: input.Range}, nil
}
func (s Section) ID() SectionID   { return s.id }
func (s Section) Level() uint8    { return s.level }
func (s Section) Heading() string { return s.heading }
func (s Section) Range() Range    { return s.rangeVal }

// BlockKind identifies addressable Markdown structure.
type BlockKind string

const (
	ParagraphBlock     BlockKind = "paragraph"
	ListItemBlock      BlockKind = "list_item"
	BlockQuoteBlock    BlockKind = "blockquote"
	CodeFenceBlock     BlockKind = "code_fence"
	CalloutBlock       BlockKind = "callout"
	TableBlock         BlockKind = "table"
	ThematicBreakBlock BlockKind = "thematic_break"
	HTMLBlock          BlockKind = "html"
	HeadingBlock       BlockKind = "heading"
	OpaqueBlock        BlockKind = "opaque"
)

func (k BlockKind) valid() bool {
	switch k {
	case ParagraphBlock, ListItemBlock, BlockQuoteBlock, CodeFenceBlock, CalloutBlock, TableBlock, ThematicBreakBlock, HTMLBlock, HeadingBlock, OpaqueBlock:
		return true
	default:
		return false
	}
}

// BlockIdentity is sealed so implicit offsets cannot become durable targets.
type BlockIdentity interface {
	blockIdentity()
	NoteID() NoteID
}

// AnchoredBlockID is the durable identity of an explicit block anchor.
type AnchoredBlockID struct {
	noteID NoteID
	anchor Anchor
}

func NewAnchoredBlockID(noteID NoteID, anchor Anchor) (AnchoredBlockID, error) {
	if noteID.IsZero() || anchor.value == "" {
		return AnchoredBlockID{}, ErrInvalidAnchoredBlockID
	}
	return AnchoredBlockID{noteID: noteID, anchor: anchor}, nil
}
func (AnchoredBlockID) blockIdentity()    {}
func (id AnchoredBlockID) NoteID() NoteID { return id.noteID }
func (id AnchoredBlockID) Anchor() Anchor { return id.anchor }

// OffsetBlockID is a local, non-durable byte-offset identity.
type OffsetBlockID struct {
	noteID NoteID
	offset ByteOffset
}

func NewOffsetBlockID(noteID NoteID, offset ByteOffset) (OffsetBlockID, error) {
	if noteID.IsZero() {
		return OffsetBlockID{}, ErrInvalidBlock
	}
	return OffsetBlockID{noteID: noteID, offset: offset}, nil
}
func (OffsetBlockID) blockIdentity()        {}
func (id OffsetBlockID) NoteID() NoteID     { return id.noteID }
func (id OffsetBlockID) Offset() ByteOffset { return id.offset }

// Block is the smallest addressable structural unit.
type Block struct {
	id       BlockIdentity
	kind     BlockKind
	rangeVal Range
}

func NewBlock(id BlockIdentity, kind BlockKind, rangeVal Range) (Block, error) {
	if id == nil || id.NoteID().IsZero() || !kind.valid() {
		return Block{}, ErrInvalidBlock
	}
	return Block{id: id, kind: kind, rangeVal: rangeVal}, nil
}
func (b Block) ID() BlockIdentity { return b.id }
func (b Block) Kind() BlockKind   { return b.kind }
func (b Block) Range() Range      { return b.rangeVal }

// AssetKind identifies an uninterpreted non-Markdown asset.
type AssetKind string

const (
	ImageAsset  AssetKind = "image"
	PDFAsset    AssetKind = "pdf"
	CanvasAsset AssetKind = "canvas"
	BaseAsset   AssetKind = "base"
	OtherAsset  AssetKind = "other"
)

func (k AssetKind) valid() bool {
	switch k {
	case ImageAsset, PDFAsset, CanvasAsset, BaseAsset, OtherAsset:
		return true
	default:
		return false
	}
}

type Asset struct {
	id   AssetID
	kind AssetKind
}

func NewAsset(id AssetID, kind AssetKind) (Asset, error) {
	if id.IsZero() || !kind.valid() {
		return Asset{}, ErrInvalidAsset
	}
	return Asset{id: id, kind: kind}, nil
}
func (a Asset) ID() AssetID     { return a.id }
func (a Asset) Kind() AssetKind { return a.kind }
