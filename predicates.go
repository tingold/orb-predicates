package predicates

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
)

// Within returns true if geometry a is completely inside geometry b.
func Within(a, b orb.Geometry) bool {
	switch gA := a.(type) {
	case orb.Point:
		switch gB := b.(type) {
		case orb.Polygon:
			return planar.PolygonContains(gB, gA)
		}
	case orb.LineString:
		switch gB := b.(type) {
		case orb.Polygon:
			for _, p := range gA {
				if !planar.PolygonContains(gB, p) {
					return false
				}
			}
			return true
		}
	case orb.Polygon:
		switch gB := b.(type) {
		case orb.Polygon:
			for _, ring := range gA {
				for _, p := range ring {
					if !planar.PolygonContains(gB, p) {
						return false
					}
				}
			}
			return true
		}
	}

	return false
}

// Contains returns true if geometry b is completely inside geometry a.
func Contains(a, b orb.Geometry) bool {
	return Within(b, a)
}

// Covers returns true if no point in geometry b is outside of geometry a.
func Covers(a, b orb.Geometry) bool {
	return false
}

// CoveredBy returns true if no point in geometry a is outside of geometry b.
func CoveredBy(a, b orb.Geometry) bool {
	return Covers(b, a)
}

// Crosses returns true if the geometries have some but not all interior points in common.
func Crosses(a, b orb.Geometry) bool {
	return false
}

// Disjoint returns true if the geometries have no points in common.
func Disjoint(a, b orb.Geometry) bool {
	return false
}

// Intersects returns true if the geometries have at least one point in common.
func Intersects(a, b orb.Geometry) bool {
	return false
}

// Overlaps returns true if the geometries have some but not all points in common,
// have the same dimension, and the intersection of the interiors of the two
// geometries has the same dimension as the geometries themselves.
func Overlaps(a, b orb.Geometry) bool {
	return false
}

// Touches returns true if the geometries have at least one point in common, but their interiors do not.
func Touches(a, b orb.Geometry) bool {
	return false
}