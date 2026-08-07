package domain

import (
	"errors"

	"golang.org/x/text/cases"
)

var ErrInvalidTag = errors.New("domain: invalid tag")

// Tag preserves source case and carries a folded lookup key.
type Tag struct {
	name, folded string
	origin       Origin
	rangeVal     Range
}

func NewTag(name string, origin Origin, rangeVal Range) (Tag, error) {
	if name == "" || !origin.valid() {
		return Tag{}, ErrInvalidTag
	}
	return Tag{name: name, folded: cases.Fold().String(name), origin: origin, rangeVal: rangeVal}, nil
}
func (t Tag) Name() string   { return t.name }
func (t Tag) Folded() string { return t.folded }
func (t Tag) Origin() Origin { return t.origin }
func (t Tag) Range() Range   { return t.rangeVal }
