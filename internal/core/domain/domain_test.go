package domain_test

import (
	"errors"
	"testing"

	"github.com/lith-project/lith/internal/core/domain"
	"github.com/lith-project/lith/internal/core/vaultpath"
)

func noteID(t *testing.T) domain.NoteID {
	t.Helper()
	path, err := vaultpath.New("/vault", "/vault/notes/design.md")
	if err != nil {
		t.Fatal(err)
	}
	return domain.NewNoteID(path)
}

func TestRangeRejectsReversedAndUsesUTF8Bytes(t *testing.T) {
	if _, err := domain.NewRange(8, 7); !errors.Is(err, domain.ErrInvalidRange) {
		t.Fatalf("error = %v, want ErrInvalidRange", err)
	}
	source := []byte("aé")
	value, err := domain.NewRange(1, 3)
	if err != nil {
		t.Fatal(err)
	}
	if string(source[value.Start():value.End()]) != "é" {
		t.Fatal("range did not address original UTF-8 bytes")
	}
	if value.Contains(value.End()) {
		t.Error("end must be exclusive")
	}
}

func TestDiagnosticIdentityIsStable(t *testing.T) {
	code, err := domain.NewDiagnosticCode("LITH-P-0042")
	if err != nil {
		t.Fatal(err)
	}
	value, err := domain.NewRange(4, 9)
	if err != nil {
		t.Fatal(err)
	}
	diagnostic, err := domain.NewDiagnostic(code, domain.Warning, value)
	if err != nil {
		t.Fatal(err)
	}
	if diagnostic.Code().String() != "LITH-P-0042" || diagnostic.Range() != value {
		t.Fatal("diagnostic identity changed")
	}
	if _, err := domain.NewDiagnosticCode("LITH-P-42"); !errors.Is(err, domain.ErrInvalidDiagnosticCode) {
		t.Fatalf("error = %v, want ErrInvalidDiagnosticCode", err)
	}
}

func TestIdentityPreservesCaseAndSectionOccurrence(t *testing.T) {
	path, err := vaultpath.New("/vault", "/vault/Notes/Design.md")
	if err != nil {
		t.Fatal(err)
	}
	id := domain.NewNoteID(path)
	if id.String() != "Notes/Design.md" {
		t.Fatal(id.String())
	}
	section, err := domain.NewSectionID(id, []string{"Design"}, 2)
	if err != nil {
		t.Fatal(err)
	}
	headings := section.HeadingPath()
	headings[0] = "changed"
	if section.HeadingPath()[0] != "Design" || section.Occurrence() != 2 {
		t.Fatal("section identity was mutable")
	}
	if _, err := domain.NewSectionID(id, []string{"Design"}, 0); !errors.Is(err, domain.ErrInvalidSectionID) {
		t.Fatalf("error = %v, want ErrInvalidSectionID", err)
	}
}

func TestLinkOptionalFieldsAndTaskStates(t *testing.T) {
	rangeValue, err := domain.NewRange(0, 8)
	if err != nil {
		t.Fatal(err)
	}
	link, err := domain.NewLink(domain.LinkInput{Kind: domain.WikiLink, Target: "Target", Origin: domain.BodyOrigin, Range: rangeValue})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := link.Subpath(); ok {
		t.Error("subpath should be absent")
	}
	if _, ok := link.Display(); ok {
		t.Error("display should be absent")
	}
	if _, ok := domain.NewTask('-', rangeValue).State().(domain.TaskOther); !ok {
		t.Error("other marker lost typed state")
	}
	if _, ok := domain.NewTask('X', rangeValue).State().(domain.TaskDone); !ok {
		t.Error("done marker lost typed state")
	}
}

func TestNoteCopiesChildrenAndDurableTargetRejectsOffset(t *testing.T) {
	id := noteID(t)
	rangeValue, err := domain.NewRange(0, 3)
	if err != nil {
		t.Fatal(err)
	}
	offset, err := domain.NewOffsetBlockID(id, 0)
	if err != nil {
		t.Fatal(err)
	}
	block, err := domain.NewBlock(offset, domain.ParagraphBlock, rangeValue)
	if err != nil {
		t.Fatal(err)
	}
	children := []domain.Block{block}
	note, err := domain.NewNote(id, rangeValue, domain.NoteParts{Blocks: children})
	if err != nil {
		t.Fatal(err)
	}
	children[0] = domain.Block{}
	if len(note.Blocks()) != 1 {
		t.Fatal("Note did not copy children")
	}
	var outcome domain.ParseOutcome = note
	if _, ok := outcome.(domain.Note); !ok {
		t.Fatal("Note is not a ParseOutcome")
	}
	anchor, err := domain.NewAnchor("block")
	if err != nil {
		t.Fatal(err)
	}
	anchored, err := domain.NewAnchoredBlockID(id, anchor)
	if err != nil {
		t.Fatal(err)
	}
	target, err := domain.NewDurableBlockTarget(id, anchored)
	if err != nil || target.BlockID().Anchor().String() != "^block" {
		t.Fatalf("target = %#v, error = %v", target, err)
	}
	if _, err := domain.NewDurableBlockTarget(id, domain.AnchoredBlockID{}); !errors.Is(err, domain.ErrInvalidAnchoredBlockID) {
		t.Fatalf("error = %v, want ErrInvalidAnchoredBlockID", err)
	}
}

func TestAmbiguousCandidatesAreSorted(t *testing.T) {
	aPath, err := vaultpath.New("/vault", "/vault/notes/alpha.md")
	if err != nil {
		t.Fatal(err)
	}
	bPath, err := vaultpath.New("/vault", "/vault/notes/beta.md")
	if err != nil {
		t.Fatal(err)
	}
	result, err := domain.NewAmbiguous([]domain.NoteID{domain.NewNoteID(bPath), domain.NewNoteID(aPath)})
	if err != nil {
		t.Fatal(err)
	}
	candidates := result.Candidates()
	if candidates[0].String() != "notes/alpha.md" {
		t.Fatal("ambiguity candidates are not sorted")
	}
}
