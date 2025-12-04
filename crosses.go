package predicates

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
)

// Crosses returns true if the geometries have some but not all interior points in common.
// The definition varies by geometry dimension:
// - Point/Line: The point is in the interior of the line
// - Line/Line: Lines cross if they intersect in a point (not a line segment)
// - Line/Area: Line passes through the area (enters and exits, or has endpoints outside)
// - MultiPoint/Line: Some points inside line, some outside
// - MultiPoint/Area: Some points inside area, some outside
func Crosses(a, b orb.Geometry) bool {
	// Empty geometries
	if isEmpty(a) || isEmpty(b) {
		return false
	}

	// Quick bounding box check
	if !boundingBoxOverlap(a, b) {
		return false
	}

	dimA := getGeometryDimension(a)
	dimB := getGeometryDimension(b)

	// Crosses is only defined for certain dimension combinations
	// Point(0)/Line(1), Line(1)/Line(1), Line(1)/Area(2), MultiPoint(0)/Line(1), MultiPoint(0)/Area(2)
	if dimA == dimB && dimA != 1 {
		// Same dimensions (except line/line) cannot cross
		return false
	}

	switch gA := a.(type) {
	case orb.Point:
		return crossesPoint(gA, b)
	case orb.MultiPoint:
		return crossesMultiPoint(gA, b)
	case orb.LineString:
		return crossesLineString(gA, b)
	case orb.MultiLineString:
		return crossesMultiLineString(gA, b)
	case orb.Ring:
		return crossesRing(gA, b)
	case orb.Polygon:
		return crossesPolygon(gA, b)
	case orb.MultiPolygon:
		return crossesMultiPolygon(gA, b)
	case orb.Collection:
		return crossesCollection(gA, b)
	case orb.Bound:
		return crossesBound(gA, b)
	}

	return false
}

// crossesPoint handles Point crosses geometry
// Points can't really "cross" anything in the standard sense
func crossesPoint(p orb.Point, b orb.Geometry) bool {
	// A single point cannot cross - it either intersects or doesn't
	// Crosses requires "some but not all" which a single point can't satisfy
	return false
}

// crossesMultiPoint handles MultiPoint crosses geometry
func crossesMultiPoint(mp orb.MultiPoint, b orb.Geometry) bool {
	if len(mp) < 2 {
		return false
	}

	switch gB := b.(type) {
	case orb.Point, orb.MultiPoint:
		// Point/Point cannot cross
		return false
	case orb.LineString:
		return multiPointCrossesLineString(mp, gB)
	case orb.MultiLineString:
		// Check if multipoint crosses any linestring
		hasInside := false
		hasOutside := false
		for _, p := range mp {
			inside := false
			for _, ls := range gB {
				if pointIntersectsLineString(p, ls) {
					inside = true
					break
				}
			}
			if inside {
				hasInside = true
			} else {
				hasOutside = true
			}
			if hasInside && hasOutside {
				return true
			}
		}
		return false
	case orb.Ring:
		return multiPointCrossesRing(mp, gB)
	case orb.Polygon:
		return multiPointCrossesPolygon(mp, gB)
	case orb.MultiPolygon:
		return multiPointCrossesMultiPolygon(mp, gB)
	case orb.Collection:
		// Check against the combined coverage
		hasInside := false
		hasOutside := false
		for _, p := range mp {
			inside := false
			for _, geom := range gB {
				if Intersects(p, geom) {
					inside = true
					break
				}
			}
			if inside {
				hasInside = true
			} else {
				hasOutside = true
			}
			if hasInside && hasOutside {
				return true
			}
		}
		return false
	case orb.Bound:
		return multiPointCrossesBound(mp, gB)
	}
	return false
}

// multiPointCrossesLineString checks if multipoint has some points on linestring and some off
func multiPointCrossesLineString(mp orb.MultiPoint, ls orb.LineString) bool {
	hasInside := false
	hasOutside := false

	for _, p := range mp {
		if pointIntersectsLineString(p, ls) {
			hasInside = true
		} else {
			hasOutside = true
		}
		if hasInside && hasOutside {
			return true
		}
	}
	return false
}

// multiPointCrossesRing checks if multipoint has some points inside ring and some outside
func multiPointCrossesRing(mp orb.MultiPoint, r orb.Ring) bool {
	hasInside := false
	hasOutside := false

	for _, p := range mp {
		if planar.RingContains(r, p) || pointOnRingBoundary(p, r) {
			hasInside = true
		} else {
			hasOutside = true
		}
		if hasInside && hasOutside {
			return true
		}
	}
	return false
}

// multiPointCrossesPolygon checks if multipoint has some points inside polygon and some outside
// Note: Points on the boundary do not count as "inside" for crosses - only interior points count
func multiPointCrossesPolygon(mp orb.MultiPoint, poly orb.Polygon) bool {
	hasInside := false
	hasOutside := false

	for _, p := range mp {
		if pointOnPolygonBoundary(p, poly) {
			// On boundary - doesn't count for crosses
			continue
		}
		if planar.PolygonContains(poly, p) {
			hasInside = true
		} else {
			hasOutside = true
		}
		if hasInside && hasOutside {
			return true
		}
	}
	return false
}

// multiPointCrossesMultiPolygon checks if multipoint crosses multipolygon
func multiPointCrossesMultiPolygon(mp orb.MultiPoint, mpoly orb.MultiPolygon) bool {
	hasInside := false
	hasOutside := false

	for _, p := range mp {
		inside := false
		for _, poly := range mpoly {
			if planar.PolygonContains(poly, p) || pointOnPolygonBoundary(p, poly) {
				inside = true
				break
			}
		}
		if inside {
			hasInside = true
		} else {
			hasOutside = true
		}
		if hasInside && hasOutside {
			return true
		}
	}
	return false
}

// multiPointCrossesBound checks if multipoint crosses bound
func multiPointCrossesBound(mp orb.MultiPoint, b orb.Bound) bool {
	hasInside := false
	hasOutside := false

	for _, p := range mp {
		if boundContainsPoint(b, p) {
			hasInside = true
		} else {
			hasOutside = true
		}
		if hasInside && hasOutside {
			return true
		}
	}
	return false
}

// crossesLineString handles LineString crosses geometry
func crossesLineString(ls orb.LineString, b orb.Geometry) bool {
	if len(ls) < 2 {
		return false
	}

	switch gB := b.(type) {
	case orb.Point:
		return false // Point can't be crossed by a line
	case orb.MultiPoint:
		return crossesMultiPoint(gB, ls) // Symmetric
	case orb.LineString:
		return lineStringCrossesLineString(ls, gB)
	case orb.MultiLineString:
		// If any component of the MultiLineString overlaps with ls,
		// then they don't cross (they share a segment instead)
		for _, ls2 := range gB {
			if linesHaveSegmentOverlap(ls, ls2) {
				return false
			}
		}
		// Now check for actual crossing
		for _, ls2 := range gB {
			if lineStringCrossesLineString(ls, ls2) {
				return true
			}
		}
		return false
	case orb.Ring:
		return lineStringCrossesRing(ls, gB)
	case orb.Polygon:
		return lineStringCrossesPolygonArea(ls, gB)
	case orb.MultiPolygon:
		for _, poly := range gB {
			if lineStringCrossesPolygonArea(ls, poly) {
				return true
			}
		}
		return false
	case orb.Collection:
		for _, geom := range gB {
			if crossesLineString(ls, geom) {
				return true
			}
		}
		return false
	case orb.Bound:
		return lineStringCrossesBound(ls, gB)
	}
	return false
}

// lineStringCrossesLineString checks if two linestrings cross (intersect at a point)
func lineStringCrossesLineString(ls1, ls2 orb.LineString) bool {
	// Lines cross if they intersect at a point (not overlap along a segment)
	// and the intersection is in the interior of both
	//
	// Important: If one line is contained within the other, or if they share
	// a segment, they do NOT cross - crosses requires the intersection to be
	// a point, not a line segment.

	// First check if there's any segment overlap (collinear overlap)
	// If so, the lines do not "cross" - they overlap
	if linesHaveSegmentOverlap(ls1, ls2) {
		return false
	}

	for i := 0; i < len(ls1)-1; i++ {
		for j := 0; j < len(ls2)-1; j++ {
			// Check for proper crossing (interior intersection)
			if segmentsCross(ls1[i], ls1[i+1], ls2[j], ls2[j+1]) {
				return true
			}
		}
	}
	return false
}

// linesHaveSegmentOverlap checks if two linestrings share a common segment (overlap)
func linesHaveSegmentOverlap(ls1, ls2 orb.LineString) bool {
	for i := 0; i < len(ls1)-1; i++ {
		for j := 0; j < len(ls2)-1; j++ {
			if segmentsOverlap(ls1[i], ls1[i+1], ls2[j], ls2[j+1]) {
				return true
			}
		}
	}
	return false
}

// segmentsOverlap checks if two segments are collinear and overlap
func segmentsOverlap(p1, p2, p3, p4 orb.Point) bool {
	// Check if segments are collinear
	d1 := sign(cross2D(p3, p4, p1))
	d2 := sign(cross2D(p3, p4, p2))
	d3 := sign(cross2D(p1, p2, p3))
	d4 := sign(cross2D(p1, p2, p4))

	// All points must be collinear
	if d1 != 0 || d2 != 0 || d3 != 0 || d4 != 0 {
		return false
	}

	// Check for overlap (more than just touching at endpoints)
	return segmentsOverlapInterior(p1, p2, p3, p4)
}

// segmentsCross checks if two segments cross (intersect in their interiors)
func segmentsCross(p1, p2, p3, p4 orb.Point) bool {
	d1 := sign(cross2D(p3, p4, p1))
	d2 := sign(cross2D(p3, p4, p2))
	d3 := sign(cross2D(p1, p2, p3))
	d4 := sign(cross2D(p1, p2, p4))

	// Proper crossing: segments straddle each other
	if ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) &&
		((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0)) {
		return true
	}

	return false
}

// lineStringCrossesRing checks if linestring crosses ring boundary
func lineStringCrossesRing(ls orb.LineString, r orb.Ring) bool {
	// Line crosses ring if it intersects the boundary at isolated points
	// (passes through from inside to outside or vice versa)

	hasInside := false
	hasOutside := false

	for _, p := range ls {
		if pointOnRingBoundary(p, r) {
			continue // On boundary, don't count
		}
		if planar.RingContains(r, p) {
			hasInside = true
		} else {
			hasOutside = true
		}
	}

	// Also check segment midpoints
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		if pointOnRingBoundary(mid, r) {
			continue
		}
		if planar.RingContains(r, mid) {
			hasInside = true
		} else {
			hasOutside = true
		}
	}

	return hasInside && hasOutside
}

// lineStringCrossesPolygonArea checks if linestring crosses polygon area
func lineStringCrossesPolygonArea(ls orb.LineString, poly orb.Polygon) bool {
	if len(poly) == 0 {
		return false
	}

	hasInside := false
	hasOutside := false

	for _, p := range ls {
		if pointOnPolygonBoundary(p, poly) {
			continue
		}
		if planar.PolygonContains(poly, p) {
			hasInside = true
		} else {
			hasOutside = true
		}
	}

	// Check segment midpoints
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		if pointOnPolygonBoundary(mid, poly) {
			continue
		}
		if planar.PolygonContains(poly, mid) {
			hasInside = true
		} else {
			hasOutside = true
		}
	}

	return hasInside && hasOutside
}

// lineStringCrossesBound checks if linestring crosses bound
func lineStringCrossesBound(ls orb.LineString, b orb.Bound) bool {
	hasInside := false
	hasOutside := false

	for _, p := range ls {
		if pointOnBoundBoundary(p, b) {
			continue
		}
		if boundContainsPointInterior(b, p) {
			hasInside = true
		} else if !boundContainsPoint(b, p) {
			hasOutside = true
		}
	}

	// Check segment midpoints
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		if pointOnBoundBoundary(mid, b) {
			continue
		}
		if boundContainsPointInterior(b, mid) {
			hasInside = true
		} else if !boundContainsPoint(b, mid) {
			hasOutside = true
		}
	}

	return hasInside && hasOutside
}

// crossesMultiLineString handles MultiLineString crosses geometry
func crossesMultiLineString(mls orb.MultiLineString, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		return false
	case orb.MultiPoint:
		return crossesMultiPoint(gB, mls)
	case orb.LineString:
		// If any component of mls overlaps with gB, they don't cross
		for _, ls := range mls {
			if linesHaveSegmentOverlap(ls, gB) {
				return false
			}
		}
		// Now check for actual crossing
		for _, ls := range mls {
			if lineStringCrossesLineString(ls, gB) {
				return true
			}
		}
		return false
	case orb.MultiLineString:
		// If any components overlap, they don't cross
		for _, ls1 := range mls {
			for _, ls2 := range gB {
				if linesHaveSegmentOverlap(ls1, ls2) {
					return false
				}
			}
		}
		// Now check for actual crossing
		for _, ls1 := range mls {
			for _, ls2 := range gB {
				if lineStringCrossesLineString(ls1, ls2) {
					return true
				}
			}
		}
		return false
	case orb.Ring:
		for _, ls := range mls {
			if lineStringCrossesRing(ls, gB) {
				return true
			}
		}
		return false
	case orb.Polygon:
		for _, ls := range mls {
			if lineStringCrossesPolygonArea(ls, gB) {
				return true
			}
		}
		return false
	case orb.MultiPolygon:
		for _, ls := range mls {
			for _, poly := range gB {
				if lineStringCrossesPolygonArea(ls, poly) {
					return true
				}
			}
		}
		return false
	case orb.Collection:
		for _, geom := range gB {
			if crossesMultiLineString(mls, geom) {
				return true
			}
		}
		return false
	case orb.Bound:
		for _, ls := range mls {
			if lineStringCrossesBound(ls, gB) {
				return true
			}
		}
		return false
	}
	return false
}

// crossesRing handles Ring crosses geometry
// Ring as a 2D area cannot "cross" another geometry in the standard sense
func crossesRing(r orb.Ring, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		return false
	case orb.MultiPoint:
		return crossesMultiPoint(gB, r)
	case orb.LineString:
		return lineStringCrossesRing(gB, r)
	case orb.MultiLineString:
		for _, ls := range gB {
			if lineStringCrossesRing(ls, r) {
				return true
			}
		}
		return false
	default:
		// 2D/2D cannot cross
		return false
	}
}

// crossesPolygon handles Polygon crosses geometry
func crossesPolygon(poly orb.Polygon, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		return false
	case orb.MultiPoint:
		return crossesMultiPoint(gB, poly)
	case orb.LineString:
		return lineStringCrossesPolygonArea(gB, poly)
	case orb.MultiLineString:
		for _, ls := range gB {
			if lineStringCrossesPolygonArea(ls, poly) {
				return true
			}
		}
		return false
	default:
		// Polygon/Polygon, Polygon/Ring etc. cannot cross (same dimension)
		return false
	}
}

// crossesMultiPolygon handles MultiPolygon crosses geometry
func crossesMultiPolygon(mp orb.MultiPolygon, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		return false
	case orb.MultiPoint:
		return crossesMultiPoint(gB, mp)
	case orb.LineString:
		for _, poly := range mp {
			if lineStringCrossesPolygonArea(gB, poly) {
				return true
			}
		}
		return false
	case orb.MultiLineString:
		for _, ls := range gB {
			for _, poly := range mp {
				if lineStringCrossesPolygonArea(ls, poly) {
					return true
				}
			}
		}
		return false
	default:
		return false
	}
}

// crossesCollection handles Collection crosses geometry
func crossesCollection(c orb.Collection, b orb.Geometry) bool {
	for _, geom := range c {
		if Crosses(geom, b) {
			return true
		}
	}
	return false
}

// crossesBound handles Bound crosses geometry
func crossesBound(bound orb.Bound, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		return false
	case orb.MultiPoint:
		return crossesMultiPoint(gB, bound)
	case orb.LineString:
		return lineStringCrossesBound(gB, bound)
	case orb.MultiLineString:
		for _, ls := range gB {
			if lineStringCrossesBound(ls, bound) {
				return true
			}
		}
		return false
	default:
		return false
	}
}
