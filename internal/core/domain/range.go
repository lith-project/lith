// Package domain contains the immutable RFC-0002 model shared by core packages.
package domain

import "errors"

// ByteOffset is an absolute offset in the original source bytes.
type ByteOffset uint64

// Range is a valid half-open original-byte range, [start,end).
type Range struct{ start, end ByteOffset }

// ErrInvalidRange reports an end before a start.
var ErrInvalidRange = errors.New("domain: range end precedes start")

// NewRange constructs a valid source-byte range.
func NewRange(start, end ByteOffset) (Range, error) {
	if end < start {
		return Range{}, ErrInvalidRange
	}
	return Range{start: start, end: end}, nil
}
func (r Range) Start() ByteOffset               { return r.start }
func (r Range) End() ByteOffset                 { return r.end }
func (r Range) Len() ByteOffset                 { return r.end - r.start }
func (r Range) IsEmpty() bool                   { return r.start == r.end }
func (r Range) Contains(offset ByteOffset) bool { return r.start <= offset && offset < r.end }
func (r Range) Encloses(other Range) bool       { return r.start <= other.start && other.end <= r.end }
