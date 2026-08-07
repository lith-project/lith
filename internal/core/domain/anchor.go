package domain

import (
	"errors"
	"strings"
)

var ErrInvalidAnchor = errors.New("domain: invalid anchor")

// Anchor is an explicit block identifier without its source '^' prefix.
type Anchor struct{ value string }

func NewAnchor(value string) (Anchor, error) {
	if value == "" || strings.HasPrefix(value, "^") || strings.ContainsAny(value, "\r\n") {
		return Anchor{}, ErrInvalidAnchor
	}
	return Anchor{value: value}, nil
}
func (a Anchor) Name() string   { return a.value }
func (a Anchor) String() string { return "^" + a.value }

// LinkSubpath is a sealed heading or explicit-anchor subpath.
type LinkSubpath interface{ linkSubpath() }
type HeadingSubpath struct{ value string }

func NewHeadingSubpath(value string) (HeadingSubpath, error) {
	if value == "" {
		return HeadingSubpath{}, ErrInvalidLink
	}
	return HeadingSubpath{value: value}, nil
}
func (HeadingSubpath) linkSubpath()     {}
func (s HeadingSubpath) String() string { return s.value }

type AnchorSubpath struct{ anchor Anchor }

func NewAnchorSubpath(anchor Anchor) (AnchorSubpath, error) {
	if anchor.value == "" {
		return AnchorSubpath{}, ErrInvalidAnchor
	}
	return AnchorSubpath{anchor: anchor}, nil
}
func (AnchorSubpath) linkSubpath()     {}
func (s AnchorSubpath) Anchor() Anchor { return s.anchor }
