package domain

import "errors"

var ErrInvalidLink = errors.New("domain: invalid link")

// Origin identifies body versus frontmatter syntax.
type Origin string

const (
	BodyOrigin        Origin = "body"
	FrontmatterOrigin Origin = "frontmatter"
)

func (o Origin) valid() bool { return o == BodyOrigin || o == FrontmatterOrigin }

// LinkKind identifies unresolved source syntax.
type LinkKind string

const (
	WikiLink      LinkKind = "wiki_link"
	WikiEmbed     LinkKind = "wiki_embed"
	MarkdownLink  LinkKind = "markdown_link"
	MarkdownImage LinkKind = "markdown_image"
	ExternalLink  LinkKind = "external_link"
)

func (k LinkKind) valid() bool {
	switch k {
	case WikiLink, WikiEmbed, MarkdownLink, MarkdownImage, ExternalLink:
		return true
	default:
		return false
	}
}

// LinkInput contains the source facts for one Link.
type LinkInput struct {
	Kind    LinkKind
	Target  string
	Origin  Origin
	Range   Range
	Subpath LinkSubpath
	Display *string
}

// Link is an unresolved original-byte-addressed reference.
type Link struct {
	kind     LinkKind
	target   string
	subpath  LinkSubpath
	display  *string
	origin   Origin
	rangeVal Range
}

func NewLink(input LinkInput) (Link, error) {
	if !input.Kind.valid() || input.Target == "" || !input.Origin.valid() {
		return Link{}, ErrInvalidLink
	}
	if input.Subpath != nil {
		switch input.Subpath.(type) {
		case HeadingSubpath, AnchorSubpath:
		default:
			return Link{}, ErrInvalidLink
		}
	}
	if input.Display != nil && *input.Display == "" {
		return Link{}, ErrInvalidLink
	}
	var display *string
	if input.Display != nil {
		value := *input.Display
		display = &value
	}
	return Link{kind: input.Kind, target: input.Target, subpath: input.Subpath, display: display, origin: input.Origin, rangeVal: input.Range}, nil
}
func (l Link) Kind() LinkKind               { return l.kind }
func (l Link) Target() string               { return l.target }
func (l Link) Subpath() (LinkSubpath, bool) { return l.subpath, l.subpath != nil }
func (l Link) Display() (string, bool) {
	if l.display == nil {
		return "", false
	}
	return *l.display, true
}
func (l Link) Origin() Origin { return l.origin }
func (l Link) Range() Range   { return l.rangeVal }
