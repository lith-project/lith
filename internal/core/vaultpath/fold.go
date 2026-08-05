package vaultpath

import "golang.org/x/text/cases"

// foldCaser performs Unicode case folding — language-neutral,
// correct for lookup keys.
var foldCaser = cases.Fold()

// FoldKey returns a case-folded form of ID, for lookup and collision
// detection. It is never the identity of a Path.
func (p Path) FoldKey() string {
	return foldCaser.String(p.id)
}
