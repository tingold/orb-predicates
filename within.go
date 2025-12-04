package predicates

import (
	"math"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
)

// Within returns true if geometry a is completely inside geometry b.
// The interior of a must be inside the interior or boundary of b,
// and the boundaries may touch but a cannot extend outside b.
func Within(a, b orb.Geometry) bool {
	// Empty geometries
	if isEmpty(a) || isEmpty(b) {
		return false
	}

	// Quick bounding box check - if a is not within b's bounds, it can't be within b
	ba := a.Bound()
	bb := b.Bound()
	if ba.Min[0] < bb.Min[0]-epsilon || ba.Max[0] > bb.Max[0]+epsilon ||
		ba.Min[1] < bb.Min[1]-epsilon || ba.Max[1] > bb.Max[1]+epsilon {
		return false
	}

	switch gA := a.(type) {
	case orb.Point:
		return withinPoint(gA, b)
	case orb.MultiPoint:
		return withinMultiPoint(gA, b)
	case orb.LineString:
		return withinLineString(gA, b)
	case orb.MultiLineString:
		return withinMultiLineString(gA, b)
	case orb.Ring:
		return withinRing(gA, b)
	case orb.Polygon:
		return withinPolygon(gA, b)
	case orb.MultiPolygon:
		return withinMultiPolygon(gA, b)
	case orb.Collection:
		return withinCollection(gA, b)
	case orb.Bound:
		return withinBound(gA, b)
	}

	return false
}

// Contains returns true if geometry b is completely inside geometry a.
func Contains(a, b orb.Geometry) bool {
	return Within(b, a)
}

// withinPoint handles Point within all geometry types
func withinPoint(p orb.Point, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		// Point can only be "within" another point if they're equal
		return pointsEqual(p, gB)
	case orb.MultiPoint:
		// Point is within MultiPoint if it equals one of the points
		for _, pt := range gB {
			if pointsEqual(p, pt) {
				return true
			}
		}
		return false
	case orb.LineString:
		// Point is within LineString if it's on the interior (not endpoints)
		return pointInLineStringInterior(p, gB)
	case orb.MultiLineString:
		// Point is within MultiLineString if it's in the interior of any component
		for _, ls := range gB {
			if pointInLineStringInterior(p, ls) {
				return true
			}
		}
		return false
	case orb.Ring:
		// Point is within Ring if it's inside (not on boundary)
		return pointInRingInterior(p, gB)
	case orb.Polygon:
		// Point is within Polygon if it's inside (not on boundary)
		return pointInPolygonInterior(p, gB)
	case orb.MultiPolygon:
		for _, poly := range gB {
			if pointInPolygonInterior(p, poly) {
				return true
			}
		}
		return false
	case orb.Collection:
		for _, geom := range gB {
			if withinPoint(p, geom) {
				return true
			}
		}
		return false
	case orb.Bound:
		return boundContainsPointInterior(gB, p)
	}
	return false
}

// pointInLineStringInterior checks if point is in interior of linestring (not endpoints)
func pointInLineStringInterior(p orb.Point, ls orb.LineString) bool {
	if len(ls) < 2 {
		return false
	}

	// Check if on any interior segment
	for i := 0; i < len(ls)-1; i++ {
		if pointOnSegmentInterior(p, ls[i], ls[i+1]) {
			return true
		}
	}

	// Check interior vertices (not first or last)
	for i := 1; i < len(ls)-1; i++ {
		if pointsEqual(p, ls[i]) {
			return true
		}
	}

	return false
}

// withinMultiPoint handles MultiPoint within all geometry types
func withinMultiPoint(mp orb.MultiPoint, b orb.Geometry) bool {
	if len(mp) == 0 {
		return false
	}

	switch gB := b.(type) {
	case orb.Point:
		// MultiPoint can only be within Point if all points equal that point
		for _, p := range mp {
			if !pointsEqual(p, gB) {
				return false
			}
		}
		return true
	case orb.MultiPoint:
		// All points in mp must be in gB
		for _, p := range mp {
			found := false
			for _, p2 := range gB {
				if pointsEqual(p, p2) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	case orb.LineString:
		// All points must be on the linestring (some in interior)
		hasInterior := false
		for _, p := range mp {
			if !pointIntersectsLineString(p, gB) {
				return false
			}
			if pointInLineStringInterior(p, gB) {
				hasInterior = true
			}
		}
		return hasInterior
	case orb.MultiLineString:
		hasInterior := false
		for _, p := range mp {
			onAny := false
			for _, ls := range gB {
				if pointIntersectsLineString(p, ls) {
					onAny = true
					if pointInLineStringInterior(p, ls) {
						hasInterior = true
					}
					break
				}
			}
			if !onAny {
				return false
			}
		}
		return hasInterior
	case orb.Ring:
		hasInterior := false
		for _, p := range mp {
			if !planar.RingContains(gB, p) && !pointOnRingBoundary(p, gB) {
				return false
			}
			if pointInRingInterior(p, gB) {
				hasInterior = true
			}
		}
		return hasInterior
	case orb.Polygon:
		hasInterior := false
		for _, p := range mp {
			if !planar.PolygonContains(gB, p) && !pointOnPolygonBoundary(p, gB) {
				return false
			}
			if pointInPolygonInterior(p, gB) {
				hasInterior = true
			}
		}
		return hasInterior
	case orb.MultiPolygon:
		hasInterior := false
		for _, p := range mp {
			inAny := false
			for _, poly := range gB {
				if planar.PolygonContains(poly, p) || pointOnPolygonBoundary(p, poly) {
					inAny = true
					if pointInPolygonInterior(p, poly) {
						hasInterior = true
					}
					break
				}
			}
			if !inAny {
				return false
			}
		}
		return hasInterior
	case orb.Collection:
		// All points must be within some geometry in the collection
		for _, p := range mp {
			inAny := false
			for _, geom := range gB {
				if withinPoint(p, geom) {
					inAny = true
					break
				}
			}
			if !inAny {
				return false
			}
		}
		return true
	case orb.Bound:
		hasInterior := false
		for _, p := range mp {
			if !boundContainsPoint(gB, p) {
				return false
			}
			if boundContainsPointInterior(gB, p) {
				hasInterior = true
			}
		}
		return hasInterior
	}
	return false
}

// withinLineString handles LineString within all geometry types
func withinLineString(ls orb.LineString, b orb.Geometry) bool {
	if len(ls) < 2 {
		return false
	}

	switch gB := b.(type) {
	case orb.Point:
		// LineString cannot be within a Point (lower dimension)
		return false
	case orb.MultiPoint:
		// LineString cannot be within MultiPoint
		return false
	case orb.LineString:
		return lineStringWithinLineString(ls, gB)
	case orb.MultiLineString:
		return lineStringWithinMultiLineString(ls, gB)
	case orb.Ring:
		return lineStringWithinRing(ls, gB)
	case orb.Polygon:
		return lineStringWithinPolygon(ls, gB)
	case orb.MultiPolygon:
		// First check if LineString is within a single polygon
		for _, poly := range gB {
			if lineStringWithinPolygon(ls, poly) {
				return true
			}
		}
		// Otherwise, check if the LineString spans multiple touching polygons
		// All points must be within or on boundary of some polygon
		return lineStringWithinMultiPolygon(ls, gB)
	case orb.Collection:
		// Check if within any single geometry
		for _, geom := range gB {
			if withinLineString(ls, geom) {
				return true
			}
		}
		return false
	case orb.Bound:
		return lineStringWithinBound(ls, gB)
	}
	return false
}

// lineStringWithinLineString checks if ls1 is within ls2
func lineStringWithinLineString(ls1, ls2 orb.LineString) bool {
	if len(ls2) < 2 {
		return false
	}

	// All points of ls1 must be on ls2
	for _, p := range ls1 {
		if !pointIntersectsLineString(p, ls2) {
			return false
		}
	}

	// All segments must be covered
	for i := 0; i < len(ls1)-1; i++ {
		if !segmentCoveredByLineString(ls1[i], ls1[i+1], ls2) {
			return false
		}
	}

	// Some part must be in interior
	for i := 0; i < len(ls1)-1; i++ {
		mid := orb.Point{(ls1[i][0] + ls1[i+1][0]) / 2, (ls1[i][1] + ls1[i+1][1]) / 2}
		if pointInLineStringInterior(mid, ls2) {
			return true
		}
	}

	return false
}

// segmentCoveredByLineString checks if segment (a,b) lies entirely on linestring
func segmentCoveredByLineString(a, b orb.Point, ls orb.LineString) bool {
	// Check if the segment is covered by any segment of the linestring
	for i := 0; i < len(ls)-1; i++ {
		if segmentCoversSegment(ls[i], ls[i+1], a, b) {
			return true
		}
	}

	// The segment might span multiple linestring segments
	// Check endpoints and midpoint
	if !pointIntersectsLineString(a, ls) || !pointIntersectsLineString(b, ls) {
		return false
	}

	mid := orb.Point{(a[0] + b[0]) / 2, (a[1] + b[1]) / 2}
	return pointIntersectsLineString(mid, ls)
}

// segmentCoversSegment checks if segment (c1,c2) contains segment (s1,s2)
func segmentCoversSegment(c1, c2, s1, s2 orb.Point) bool {
	return pointOnSegment(s1, c1, c2) && pointOnSegment(s2, c1, c2)
}

// lineStringWithinMultiLineString checks if ls is within mls
func lineStringWithinMultiLineString(ls orb.LineString, mls orb.MultiLineString) bool {
	// Check if within any single linestring
	for _, ls2 := range mls {
		if lineStringWithinLineString(ls, ls2) {
			return true
		}
	}

	// Otherwise, check if all parts are covered
	for _, p := range ls {
		onAny := false
		for _, ls2 := range mls {
			if pointIntersectsLineString(p, ls2) {
				onAny = true
				break
			}
		}
		if !onAny {
			return false
		}
	}

	// Check midpoints of segments
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		onAny := false
		for _, ls2 := range mls {
			if pointIntersectsLineString(mid, ls2) {
				onAny = true
				break
			}
		}
		if !onAny {
			return false
		}
	}

	return true
}

// lineStringWithinRing checks if linestring is within ring interior
func lineStringWithinRing(ls orb.LineString, r orb.Ring) bool {
	// All points must be inside or on boundary
	for _, p := range ls {
		if !planar.RingContains(r, p) && !pointOnRingBoundary(p, r) {
			return false
		}
	}

	// Check segment midpoints - at least one must be in interior
	hasInterior := false
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		if !planar.RingContains(r, mid) && !pointOnRingBoundary(mid, r) {
			return false
		}
		if pointInRingInterior(mid, r) {
			hasInterior = true
		}
	}

	return hasInterior
}

// lineStringWithinPolygon checks if linestring is within polygon interior
func lineStringWithinPolygon(ls orb.LineString, poly orb.Polygon) bool {
	if len(poly) == 0 {
		return false
	}

	// All points must be inside or on boundary
	for _, p := range ls {
		if !planar.PolygonContains(poly, p) && !pointOnPolygonBoundary(p, poly) {
			return false
		}
	}

	// Check segment midpoints
	hasInterior := false
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		if !planar.PolygonContains(poly, mid) && !pointOnPolygonBoundary(mid, poly) {
			return false
		}
		if pointInPolygonInterior(mid, poly) {
			hasInterior = true
		}
	}

	return hasInterior
}

// lineStringWithinMultiPolygon checks if linestring is within a MultiPolygon
// The linestring may span multiple polygons that touch
func lineStringWithinMultiPolygon(ls orb.LineString, mp orb.MultiPolygon) bool {
	if len(mp) == 0 || len(ls) < 2 {
		return false
	}

	// Helper to check if a point is within any polygon of the multipolygon
	pointInAnyPoly := func(p orb.Point) bool {
		for _, poly := range mp {
			if planar.PolygonContains(poly, p) || pointOnPolygonBoundary(p, poly) {
				return true
			}
		}
		return false
	}

	// All vertices must be within or on boundary of some polygon
	for _, p := range ls {
		if !pointInAnyPoly(p) {
			return false
		}
	}

	// For each segment, check sample points to catch gaps between polygons
	// Also specifically check points near polygon vertices/boundaries
	// This is much more efficient than the original 10,000 samples
	const numSamples = 50
	for i := 0; i < len(ls)-1; i++ {
		segStart, segEnd := ls[i], ls[i+1]

		// Regular sampling along the segment
		for s := 1; s < numSamples; s++ {
			t := float64(s) / float64(numSamples)
			sample := orb.Point{
				segStart[0] + t*(segEnd[0]-segStart[0]),
				segStart[1] + t*(segEnd[1]-segStart[1]),
			}
			if !pointInAnyPoly(sample) {
				return false
			}
		}

		// Additionally, check points near polygon vertex y-coordinates
		// This catches gaps at polygon junctions
		for _, poly := range mp {
			for _, ring := range poly {
				for _, vertex := range ring {
					// Find t value where line crosses this vertex's y-coordinate
					dy := segEnd[1] - segStart[1]
					if math.Abs(dy) > epsilon {
						t := (vertex[1] - segStart[1]) / dy
						if t > epsilon && t < 1-epsilon {
							// Check points slightly before and after this y-level
							for _, offset := range []float64{-0.0001, 0, 0.0001} {
								tAdj := t + offset
								if tAdj > 0 && tAdj < 1 {
									sample := orb.Point{
										segStart[0] + tAdj*(segEnd[0]-segStart[0]),
										segStart[1] + tAdj*(segEnd[1]-segStart[1]),
									}
									if !pointInAnyPoly(sample) {
										return false
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// At least one point must be in the interior of some polygon
	for _, p := range ls {
		for _, poly := range mp {
			if pointInPolygonInterior(p, poly) {
				return true
			}
		}
	}

	// Check segment midpoints for interior
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		for _, poly := range mp {
			if pointInPolygonInterior(mid, poly) {
				return true
			}
		}
	}

	return false
}

// lineStringWithinBound checks if linestring is within bound interior
func lineStringWithinBound(ls orb.LineString, b orb.Bound) bool {
	// All points must be inside or on boundary
	for _, p := range ls {
		if !boundContainsPoint(b, p) {
			return false
		}
	}

	// At least one point must be in interior
	for _, p := range ls {
		if boundContainsPointInterior(b, p) {
			return true
		}
	}

	// Check midpoints
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		if boundContainsPointInterior(b, mid) {
			return true
		}
	}

	return false
}

// withinMultiLineString handles MultiLineString within all geometry types
func withinMultiLineString(mls orb.MultiLineString, b orb.Geometry) bool {
	if len(mls) == 0 {
		return false
	}

	// All component linestrings must be within b
	for _, ls := range mls {
		if !withinLineString(ls, b) {
			return false
		}
	}
	return true
}

// withinRing handles Ring within all geometry types
func withinRing(r orb.Ring, b orb.Geometry) bool {
	if len(r) < 4 {
		return false
	}

	switch gB := b.(type) {
	case orb.Point, orb.MultiPoint, orb.LineString, orb.MultiLineString:
		// Ring (2D) cannot be within lower dimensional geometries
		return false
	case orb.Ring:
		return ringWithinRing(r, gB)
	case orb.Polygon:
		return ringWithinPolygon(r, gB)
	case orb.MultiPolygon:
		for _, poly := range gB {
			if ringWithinPolygon(r, poly) {
				return true
			}
		}
		return false
	case orb.Collection:
		for _, geom := range gB {
			if withinRing(r, geom) {
				return true
			}
		}
		return false
	case orb.Bound:
		return ringWithinBound(r, gB)
	}
	return false
}

// ringWithinRing checks if ring r1 is within ring r2
func ringWithinRing(r1, r2 orb.Ring) bool {
	// All points of r1 must be inside or on r2
	for _, p := range r1 {
		if !planar.RingContains(r2, p) && !pointOnRingBoundary(p, r2) {
			return false
		}
	}

	// No edge crossings in interior
	for i := 0; i < len(r1)-1; i++ {
		for j := 0; j < len(r2)-1; j++ {
			if segmentsIntersectInterior(r1[i], r1[i+1], r2[j], r2[j+1]) {
				return false
			}
		}
	}

	// At least one point must be in interior
	for _, p := range r1 {
		if pointInRingInterior(p, r2) {
			return true
		}
	}

	// Check centroid
	centroid := ringCentroid(r1)
	return pointInRingInterior(centroid, r2)
}

// ringCentroid computes the centroid of a ring
func ringCentroid(r orb.Ring) orb.Point {
	if len(r) == 0 {
		return orb.Point{}
	}
	var sumX, sumY float64
	for _, p := range r {
		sumX += p[0]
		sumY += p[1]
	}
	n := float64(len(r))
	return orb.Point{sumX / n, sumY / n}
}

// ringWithinPolygon checks if ring r is within polygon poly
func ringWithinPolygon(r orb.Ring, poly orb.Polygon) bool {
	if len(poly) == 0 {
		return false
	}

	// All points must be inside or on boundary of polygon
	for _, p := range r {
		if !planar.PolygonContains(poly, p) && !pointOnPolygonBoundary(p, poly) {
			return false
		}
	}

	// No interior edge crossings
	for i := 0; i < len(r)-1; i++ {
		for _, polyRing := range poly {
			for j := 0; j < len(polyRing)-1; j++ {
				if segmentsIntersectInterior(r[i], r[i+1], polyRing[j], polyRing[j+1]) {
					return false
				}
			}
		}
	}

	// At least one point must be in interior
	centroid := ringCentroid(r)
	return pointInPolygonInterior(centroid, poly)
}

// ringWithinBound checks if ring r is within bound b
func ringWithinBound(r orb.Ring, b orb.Bound) bool {
	for _, p := range r {
		if !boundContainsPoint(b, p) {
			return false
		}
	}

	// At least one point must be in interior
	centroid := ringCentroid(r)
	return boundContainsPointInterior(b, centroid)
}

// withinPolygon handles Polygon within all geometry types
func withinPolygon(poly orb.Polygon, b orb.Geometry) bool {
	if len(poly) == 0 || len(poly[0]) < 4 {
		return false
	}

	switch gB := b.(type) {
	case orb.Point, orb.MultiPoint, orb.LineString, orb.MultiLineString:
		// Polygon cannot be within lower dimensional geometries
		return false
	case orb.Ring:
		// Polygon within Ring: exterior must be within, holes are okay
		return ringWithinRing(poly[0], gB)
	case orb.Polygon:
		return polygonWithinPolygon(poly, gB)
	case orb.MultiPolygon:
		for _, poly2 := range gB {
			if polygonWithinPolygon(poly, poly2) {
				return true
			}
		}
		return false
	case orb.Collection:
		for _, geom := range gB {
			if withinPolygon(poly, geom) {
				return true
			}
		}
		return false
	case orb.Bound:
		return polygonWithinBound(poly, gB)
	}
	return false
}

// polygonWithinPolygon checks if poly1 is within poly2
func polygonWithinPolygon(poly1, poly2 orb.Polygon) bool {
	if len(poly1) == 0 || len(poly2) == 0 {
		return false
	}

	// Check if polygons are topologically equal first
	// This handles the case of equal polygons with different orientations
	if polygonsTopologicallyEqual(poly1, poly2) {
		return true
	}

	// All points of poly1's exterior must be within poly2 (inside or on boundary)
	for _, p := range poly1[0] {
		if !planar.PolygonContains(poly2, p) && !pointOnPolygonBoundary(p, poly2) {
			return false
		}
	}

	// poly1 must not overlap with poly2's holes
	// If any interior point of poly1 is inside a hole of poly2, poly1 is not within poly2
	for i := 1; i < len(poly2); i++ {
		hole := poly2[i]
		// Check the centroid of poly1's exterior
		centroid := ringCentroid(poly1[0])
		if planar.RingContains(hole, centroid) && !pointOnRingBoundary(centroid, hole) {
			return false
		}
		// Check if any point of poly1's exterior is inside poly2's hole
		for _, p := range poly1[0] {
			if planar.RingContains(hole, p) && !pointOnRingBoundary(p, hole) {
				return false
			}
		}
		// Check segment midpoints of poly1 exterior against holes
		for j := 0; j < len(poly1[0])-1; j++ {
			mid := orb.Point{(poly1[0][j][0] + poly1[0][j+1][0]) / 2, (poly1[0][j][1] + poly1[0][j+1][1]) / 2}
			if planar.RingContains(hole, mid) && !pointOnRingBoundary(mid, hole) {
				return false
			}
		}
		// Check if poly1's exterior ring intersects with the hole boundary
		// and if any part of poly1 passes through the hole
		if ringsIntersect(poly1[0], hole) {
			// If rings intersect, check if poly1 has interior in the hole
			holeCentroid := ringCentroid(hole)
			if planar.RingContains(poly1[0], holeCentroid) {
				// poly1 covers the hole area, so it intersects with the hole
				return false
			}
		}
	}

	// No interior edge crossings between poly1 and poly2 boundaries
	for _, r1 := range poly1 {
		for i := 0; i < len(r1)-1; i++ {
			for _, r2 := range poly2 {
				for j := 0; j < len(r2)-1; j++ {
					if segmentsIntersectInterior(r1[i], r1[i+1], r2[j], r2[j+1]) {
						return false
					}
				}
			}
		}
	}

	// At least one point of poly1 must be in the interior of poly2
	centroid := ringCentroid(poly1[0])
	if pointInPolygonInterior(centroid, poly2) {
		return true
	}

	// Try multiple sample points if centroid doesn't work
	for _, p := range poly1[0] {
		if pointInPolygonInterior(p, poly2) {
			return true
		}
	}

	// Check segment midpoints
	for i := 0; i < len(poly1[0])-1; i++ {
		mid := orb.Point{(poly1[0][i][0] + poly1[0][i+1][0]) / 2, (poly1[0][i][1] + poly1[0][i+1][1]) / 2}
		if pointInPolygonInterior(mid, poly2) {
			return true
		}
	}

	return false
}

// polygonsTopologicallyEqual checks if two polygons cover the same area
func polygonsTopologicallyEqual(poly1, poly2 orb.Polygon) bool {
	if len(poly1) == 0 || len(poly2) == 0 {
		return false
	}

	// All vertices of poly1's exterior must be on poly2's exterior boundary
	for _, p := range poly1[0] {
		if !pointOnRingBoundary(p, poly2[0]) {
			return false
		}
	}

	// All vertices of poly2's exterior must be on poly1's exterior boundary
	for _, p := range poly2[0] {
		if !pointOnRingBoundary(p, poly1[0]) {
			return false
		}
	}

	// Check that hole counts match and holes are equivalent
	if len(poly1) != len(poly2) {
		return false
	}

	return true
}

// polygonWithinBound checks if polygon is within bound
func polygonWithinBound(poly orb.Polygon, b orb.Bound) bool {
	if len(poly) == 0 {
		return false
	}

	for _, ring := range poly {
		for _, p := range ring {
			if !boundContainsPoint(b, p) {
				return false
			}
		}
	}

	// At least one point must be in interior
	centroid := ringCentroid(poly[0])
	return boundContainsPointInterior(b, centroid)
}

// withinMultiPolygon handles MultiPolygon within all geometry types
func withinMultiPolygon(mp orb.MultiPolygon, b orb.Geometry) bool {
	if len(mp) == 0 {
		return false
	}

	// All component polygons must be within b
	for _, poly := range mp {
		if !withinPolygon(poly, b) {
			return false
		}
	}
	return true
}

// withinCollection handles Collection within all geometry types
func withinCollection(c orb.Collection, b orb.Geometry) bool {
	if len(c) == 0 {
		return false
	}

	// All geometries in collection must be within b
	for _, geom := range c {
		if !Within(geom, b) {
			return false
		}
	}
	return true
}

// withinBound handles Bound within all geometry types
func withinBound(bound orb.Bound, b orb.Geometry) bool {
	// Treat bound as a polygon
	poly := boundToPolygon(bound)
	return withinPolygon(poly, b)
}
