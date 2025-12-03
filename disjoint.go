package predicates

import (
	"github.com/paulmach/orb"
)

// Disjoint returns true if the geometries have no points in common.
// This is the complement of Intersects.
func Disjoint(a, b orb.Geometry) bool {
	return !Intersects(a, b)
}
