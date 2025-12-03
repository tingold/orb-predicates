package predicates

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
)

// Intersects returns true if the geometries have at least one point in common.
func Intersects(a, b orb.Geometry) bool {
	// Quick bounding box rejection
	if !boundingBoxOverlap(a, b) {
		return false
	}

	// Handle empty geometries
	if isEmpty(a) || isEmpty(b) {
		return false
	}

	switch gA := a.(type) {
	case orb.Point:
		return intersectsPoint(gA, b)
	case orb.MultiPoint:
		return intersectsMultiPoint(gA, b)
	case orb.LineString:
		return intersectsLineString(gA, b)
	case orb.MultiLineString:
		return intersectsMultiLineString(gA, b)
	case orb.Ring:
		return intersectsRing(gA, b)
	case orb.Polygon:
		return intersectsPolygon(gA, b)
	case orb.MultiPolygon:
		return intersectsMultiPolygon(gA, b)
	case orb.Collection:
		return intersectsCollection(gA, b)
	case orb.Bound:
		return intersectsBound(gA, b)
	}

	return false
}

// intersectsPoint handles Point vs all geometry types
func intersectsPoint(p orb.Point, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		return pointsEqual(p, gB)
	case orb.MultiPoint:
		for _, pt := range gB {
			if pointsEqual(p, pt) {
				return true
			}
		}
		return false
	case orb.LineString:
		return pointIntersectsLineString(p, gB)
	case orb.MultiLineString:
		for _, ls := range gB {
			if pointIntersectsLineString(p, ls) {
				return true
			}
		}
		return false
	case orb.Ring:
		return planar.RingContains(gB, p) || pointOnRingBoundary(p, gB)
	case orb.Polygon:
		return planar.PolygonContains(gB, p) || pointOnPolygonBoundary(p, gB)
	case orb.MultiPolygon:
		for _, poly := range gB {
			if planar.PolygonContains(poly, p) || pointOnPolygonBoundary(p, poly) {
				return true
			}
		}
		return false
	case orb.Collection:
		for _, geom := range gB {
			if intersectsPoint(p, geom) {
				return true
			}
		}
		return false
	case orb.Bound:
		return boundContainsPoint(gB, p)
	}
	return false
}

// pointIntersectsLineString checks if a point intersects a linestring
func pointIntersectsLineString(p orb.Point, ls orb.LineString) bool {
	if len(ls) == 0 {
		return false
	}
	if len(ls) == 1 {
		return pointsEqual(p, ls[0])
	}
	for i := 0; i < len(ls)-1; i++ {
		if pointOnSegment(p, ls[i], ls[i+1]) {
			return true
		}
	}
	return false
}

// intersectsMultiPoint handles MultiPoint vs all geometry types
func intersectsMultiPoint(mp orb.MultiPoint, b orb.Geometry) bool {
	for _, p := range mp {
		if intersectsPoint(p, b) {
			return true
		}
	}
	return false
}

// intersectsLineString handles LineString vs all geometry types
func intersectsLineString(ls orb.LineString, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		return pointIntersectsLineString(gB, ls)
	case orb.MultiPoint:
		for _, p := range gB {
			if pointIntersectsLineString(p, ls) {
				return true
			}
		}
		return false
	case orb.LineString:
		return lineStringsIntersect(ls, gB)
	case orb.MultiLineString:
		for _, ls2 := range gB {
			if lineStringsIntersect(ls, ls2) {
				return true
			}
		}
		return false
	case orb.Ring:
		return lineStringIntersectsRingOrInterior(ls, gB)
	case orb.Polygon:
		return lineStringIntersectsPolygon(ls, gB)
	case orb.MultiPolygon:
		for _, poly := range gB {
			if lineStringIntersectsPolygon(ls, poly) {
				return true
			}
		}
		return false
	case orb.Collection:
		for _, geom := range gB {
			if intersectsLineString(ls, geom) {
				return true
			}
		}
		return false
	case orb.Bound:
		return lineStringIntersectsBound(ls, gB)
	}
	return false
}

// lineStringIntersectsRingOrInterior checks if linestring intersects ring boundary or interior
func lineStringIntersectsRingOrInterior(ls orb.LineString, r orb.Ring) bool {
	// Check if any segment intersects the ring
	if lineStringIntersectsRing(ls, r) {
		return true
	}
	// Check if any point of the linestring is inside the ring
	for _, p := range ls {
		if planar.RingContains(r, p) {
			return true
		}
	}
	return false
}

// lineStringIntersectsPolygon checks if a linestring intersects a polygon
func lineStringIntersectsPolygon(ls orb.LineString, poly orb.Polygon) bool {
	if len(poly) == 0 {
		return false
	}

	// Check boundary intersection
	for _, ring := range poly {
		if lineStringIntersectsRing(ls, ring) {
			return true
		}
	}

	// Check if any point is inside the polygon
	for _, p := range ls {
		if planar.PolygonContains(poly, p) {
			return true
		}
	}

	return false
}

// lineStringIntersectsBound checks if a linestring intersects a bound
func lineStringIntersectsBound(ls orb.LineString, b orb.Bound) bool {
	poly := boundToPolygon(b)
	return lineStringIntersectsPolygon(ls, poly)
}

// intersectsMultiLineString handles MultiLineString vs all geometry types
func intersectsMultiLineString(mls orb.MultiLineString, b orb.Geometry) bool {
	for _, ls := range mls {
		if intersectsLineString(ls, b) {
			return true
		}
	}
	return false
}

// intersectsRing handles Ring vs all geometry types
func intersectsRing(r orb.Ring, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		return planar.RingContains(r, gB) || pointOnRingBoundary(gB, r)
	case orb.MultiPoint:
		for _, p := range gB {
			if planar.RingContains(r, p) || pointOnRingBoundary(p, r) {
				return true
			}
		}
		return false
	case orb.LineString:
		return lineStringIntersectsRingOrInterior(gB, r)
	case orb.MultiLineString:
		for _, ls := range gB {
			if lineStringIntersectsRingOrInterior(ls, r) {
				return true
			}
		}
		return false
	case orb.Ring:
		return ringsIntersect(r, gB)
	case orb.Polygon:
		return ringIntersectsPolygon(r, gB)
	case orb.MultiPolygon:
		for _, poly := range gB {
			if ringIntersectsPolygon(r, poly) {
				return true
			}
		}
		return false
	case orb.Collection:
		for _, geom := range gB {
			if intersectsRing(r, geom) {
				return true
			}
		}
		return false
	case orb.Bound:
		poly := boundToPolygon(gB)
		return ringIntersectsPolygon(r, poly)
	}
	return false
}

// ringIntersectsPolygon checks if a ring intersects a polygon
func ringIntersectsPolygon(r orb.Ring, poly orb.Polygon) bool {
	if len(poly) == 0 {
		return false
	}

	// Check boundary intersections
	for _, polyRing := range poly {
		if ringBoundariesIntersect(r, polyRing) {
			return true
		}
	}

	// Check if any point of r is inside poly
	if len(r) > 0 && planar.PolygonContains(poly, r[0]) {
		return true
	}

	// Check if any point of poly exterior is inside r
	if len(poly[0]) > 0 && planar.RingContains(r, poly[0][0]) {
		return true
	}

	return false
}

// intersectsPolygon handles Polygon vs all geometry types
func intersectsPolygon(poly orb.Polygon, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		return planar.PolygonContains(poly, gB) || pointOnPolygonBoundary(gB, poly)
	case orb.MultiPoint:
		for _, p := range gB {
			if planar.PolygonContains(poly, p) || pointOnPolygonBoundary(p, poly) {
				return true
			}
		}
		return false
	case orb.LineString:
		return lineStringIntersectsPolygon(gB, poly)
	case orb.MultiLineString:
		for _, ls := range gB {
			if lineStringIntersectsPolygon(ls, poly) {
				return true
			}
		}
		return false
	case orb.Ring:
		return ringIntersectsPolygon(gB, poly)
	case orb.Polygon:
		return polygonsIntersect(poly, gB)
	case orb.MultiPolygon:
		for _, poly2 := range gB {
			if polygonsIntersect(poly, poly2) {
				return true
			}
		}
		return false
	case orb.Collection:
		for _, geom := range gB {
			if intersectsPolygon(poly, geom) {
				return true
			}
		}
		return false
	case orb.Bound:
		return polygonsIntersect(poly, boundToPolygon(gB))
	}
	return false
}

// polygonsIntersect checks if two polygons intersect
func polygonsIntersect(p1, p2 orb.Polygon) bool {
	if len(p1) == 0 || len(p2) == 0 {
		return false
	}

	// Check boundary intersections
	for _, r1 := range p1 {
		for _, r2 := range p2 {
			if ringBoundariesIntersect(r1, r2) {
				return true
			}
		}
	}

	// Check if any point of p1 is inside p2
	if len(p1[0]) > 0 && planar.PolygonContains(p2, p1[0][0]) {
		return true
	}

	// Check if any point of p2 is inside p1
	if len(p2[0]) > 0 && planar.PolygonContains(p1, p2[0][0]) {
		return true
	}

	return false
}

// intersectsMultiPolygon handles MultiPolygon vs all geometry types
func intersectsMultiPolygon(mp orb.MultiPolygon, b orb.Geometry) bool {
	for _, poly := range mp {
		if intersectsPolygon(poly, b) {
			return true
		}
	}
	return false
}

// intersectsCollection handles Collection vs all geometry types
func intersectsCollection(c orb.Collection, b orb.Geometry) bool {
	for _, geom := range c {
		if Intersects(geom, b) {
			return true
		}
	}
	return false
}

// intersectsBound handles Bound vs all geometry types
func intersectsBound(bound orb.Bound, b orb.Geometry) bool {
	poly := boundToPolygon(bound)
	return intersectsPolygon(poly, b)
}
