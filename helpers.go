package predicates

import (
	"math"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
)

const epsilon = 1e-10

// sign returns the sign of a float64 (-1, 0, or 1)
func sign(x float64) int {
	if x < -epsilon {
		return -1
	}
	if x > epsilon {
		return 1
	}
	return 0
}

// cross2D computes the 2D cross product of vectors (p2-p1) and (p3-p1)
func cross2D(p1, p2, p3 orb.Point) float64 {
	return (p2[0]-p1[0])*(p3[1]-p1[1]) - (p2[1]-p1[1])*(p3[0]-p1[0])
}

// pointsEqual checks if two points are equal within epsilon
func pointsEqual(p1, p2 orb.Point) bool {
	return math.Abs(p1[0]-p2[0]) < epsilon && math.Abs(p1[1]-p2[1]) < epsilon
}

// pointOnSegment checks if point p lies on segment ab (excluding endpoints by default)
func pointOnSegment(p, a, b orb.Point) bool {
	// Check collinearity using cross product
	cross := cross2D(a, b, p)
	if math.Abs(cross) > epsilon {
		return false
	}

	// Check if p is within the bounding box of ab
	minX, maxX := math.Min(a[0], b[0]), math.Max(a[0], b[0])
	minY, maxY := math.Min(a[1], b[1]), math.Max(a[1], b[1])

	return p[0] >= minX-epsilon && p[0] <= maxX+epsilon &&
		p[1] >= minY-epsilon && p[1] <= maxY+epsilon
}

// pointOnSegmentInterior checks if point p lies strictly in the interior of segment ab
func pointOnSegmentInterior(p, a, b orb.Point) bool {
	if pointsEqual(p, a) || pointsEqual(p, b) {
		return false
	}
	return pointOnSegment(p, a, b)
}

// segmentsIntersect checks if segments (p1,p2) and (p3,p4) intersect
func segmentsIntersect(p1, p2, p3, p4 orb.Point) bool {
	d1 := sign(cross2D(p3, p4, p1))
	d2 := sign(cross2D(p3, p4, p2))
	d3 := sign(cross2D(p1, p2, p3))
	d4 := sign(cross2D(p1, p2, p4))

	// Standard intersection case
	if ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) &&
		((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0)) {
		return true
	}

	// Collinear cases
	if d1 == 0 && pointOnSegment(p1, p3, p4) {
		return true
	}
	if d2 == 0 && pointOnSegment(p2, p3, p4) {
		return true
	}
	if d3 == 0 && pointOnSegment(p3, p1, p2) {
		return true
	}
	if d4 == 0 && pointOnSegment(p4, p1, p2) {
		return true
	}

	return false
}

// segmentsIntersectInterior checks if segments intersect in their interiors (not at endpoints)
func segmentsIntersectInterior(p1, p2, p3, p4 orb.Point) bool {
	d1 := sign(cross2D(p3, p4, p1))
	d2 := sign(cross2D(p3, p4, p2))
	d3 := sign(cross2D(p1, p2, p3))
	d4 := sign(cross2D(p1, p2, p4))

	// Proper intersection (not at endpoints)
	if ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) &&
		((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0)) {
		return true
	}

	// Check for collinear overlap in interior
	if d1 == 0 && d2 == 0 && d3 == 0 && d4 == 0 {
		// All collinear - check for interior overlap
		return segmentsOverlapInterior(p1, p2, p3, p4)
	}

	return false
}

// segmentsOverlapInterior checks if two collinear segments overlap in their interiors
func segmentsOverlapInterior(p1, p2, p3, p4 orb.Point) bool {
	// Project onto the axis with greater extent
	var t1, t2, t3, t4 float64
	if math.Abs(p2[0]-p1[0]) > math.Abs(p2[1]-p1[1]) {
		t1, t2 = p1[0], p2[0]
		t3, t4 = p3[0], p4[0]
	} else {
		t1, t2 = p1[1], p2[1]
		t3, t4 = p3[1], p4[1]
	}

	if t1 > t2 {
		t1, t2 = t2, t1
	}
	if t3 > t4 {
		t3, t4 = t4, t3
	}

	// Check for interior overlap (not just touching at endpoints)
	overlapStart := math.Max(t1, t3)
	overlapEnd := math.Min(t2, t4)

	return overlapEnd-overlapStart > epsilon
}

// segmentsAreCollinear checks if both segments lie on the same infinite line
func segmentsAreCollinear(p1, p2, p3, p4 orb.Point) bool {
	d1 := sign(cross2D(p3, p4, p1))
	d2 := sign(cross2D(p3, p4, p2))
	d3 := sign(cross2D(p1, p2, p3))
	d4 := sign(cross2D(p1, p2, p4))
	return d1 == 0 && d2 == 0 && d3 == 0 && d4 == 0
}

// segmentsCrossProper checks if two segments cross at a single interior point
func segmentsCrossProper(p1, p2, p3, p4 orb.Point) bool {
	d1 := sign(cross2D(p3, p4, p1))
	d2 := sign(cross2D(p3, p4, p2))
	d3 := sign(cross2D(p1, p2, p3))
	d4 := sign(cross2D(p1, p2, p4))

	return ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) &&
		((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0))
}

// pointOnRingBoundary checks if a point lies on the boundary of a ring
func pointOnRingBoundary(p orb.Point, r orb.Ring) bool {
	if len(r) < 2 {
		return false
	}
	for i := 0; i < len(r)-1; i++ {
		if pointOnSegment(p, r[i], r[i+1]) {
			return true
		}
	}
	return false
}

// pointOnPolygonBoundary checks if a point lies on the boundary of a polygon
func pointOnPolygonBoundary(p orb.Point, poly orb.Polygon) bool {
	for _, ring := range poly {
		if pointOnRingBoundary(p, ring) {
			return true
		}
	}
	return false
}

// pointInRingInterior checks if a point is strictly inside a ring (not on boundary)
func pointInRingInterior(p orb.Point, r orb.Ring) bool {
	if pointOnRingBoundary(p, r) {
		return false
	}
	return planar.RingContains(r, p)
}

// pointInPolygonInterior checks if a point is strictly inside a polygon (not on boundary)
func pointInPolygonInterior(p orb.Point, poly orb.Polygon) bool {
	if pointOnPolygonBoundary(p, poly) {
		return false
	}
	return planar.PolygonContains(poly, p)
}

// lineStringOnRingBoundary checks if all points of a linestring lie on a ring's boundary
func lineStringOnRingBoundary(ls orb.LineString, r orb.Ring) bool {
	if len(ls) < 2 {
		return false
	}
	for i := 0; i < len(ls)-1; i++ {
		if !segmentOnRingBoundary(ls[i], ls[i+1], r) {
			return false
		}
	}
	return true
}

// segmentOnRingBoundary checks if a segment lies entirely on a ring's boundary
func segmentOnRingBoundary(a, b orb.Point, r orb.Ring) bool {
	// Both endpoints must be on boundary
	if !pointOnRingBoundary(a, r) || !pointOnRingBoundary(b, r) {
		return false
	}

	// Check if the segment lies along the ring edges
	// Sample midpoint and check if it's on boundary
	mid := orb.Point{(a[0] + b[0]) / 2, (a[1] + b[1]) / 2}
	return pointOnRingBoundary(mid, r)
}

// ringsIntersect checks if two rings have any intersection (boundary or interior)
func ringsIntersect(r1, r2 orb.Ring) bool {
	// Check edge intersections
	for i := 0; i < len(r1)-1; i++ {
		for j := 0; j < len(r2)-1; j++ {
			if segmentsIntersect(r1[i], r1[i+1], r2[j], r2[j+1]) {
				return true
			}
		}
	}

	// Check if one ring is inside the other
	if len(r1) > 0 && planar.RingContains(r2, r1[0]) {
		return true
	}
	if len(r2) > 0 && planar.RingContains(r1, r2[0]) {
		return true
	}

	return false
}

// ringBoundariesIntersect checks if ring boundaries intersect
func ringBoundariesIntersect(r1, r2 orb.Ring) bool {
	for i := 0; i < len(r1)-1; i++ {
		for j := 0; j < len(r2)-1; j++ {
			if segmentsIntersect(r1[i], r1[i+1], r2[j], r2[j+1]) {
				return true
			}
		}
	}
	return false
}

// lineStringsIntersect checks if two linestrings intersect
func lineStringsIntersect(ls1, ls2 orb.LineString) bool {
	for i := 0; i < len(ls1)-1; i++ {
		for j := 0; j < len(ls2)-1; j++ {
			if segmentsIntersect(ls1[i], ls1[i+1], ls2[j], ls2[j+1]) {
				return true
			}
		}
	}
	return false
}

// lineStringIntersectsRing checks if a linestring intersects a ring
func lineStringIntersectsRing(ls orb.LineString, r orb.Ring) bool {
	for i := 0; i < len(ls)-1; i++ {
		for j := 0; j < len(r)-1; j++ {
			if segmentsIntersect(ls[i], ls[i+1], r[j], r[j+1]) {
				return true
			}
		}
	}
	return false
}

// boundingBoxOverlap checks if bounding boxes of two geometries overlap
func boundingBoxOverlap(a, b orb.Geometry) bool {
	ba := a.Bound()
	bb := b.Bound()

	return ba.Min[0] <= bb.Max[0]+epsilon &&
		ba.Max[0] >= bb.Min[0]-epsilon &&
		ba.Min[1] <= bb.Max[1]+epsilon &&
		ba.Max[1] >= bb.Min[1]-epsilon
}

// boundToPolygon converts a Bound to a Polygon
func boundToPolygon(b orb.Bound) orb.Polygon {
	return orb.Polygon{
		orb.Ring{
			orb.Point{b.Min[0], b.Min[1]},
			orb.Point{b.Max[0], b.Min[1]},
			orb.Point{b.Max[0], b.Max[1]},
			orb.Point{b.Min[0], b.Max[1]},
			orb.Point{b.Min[0], b.Min[1]},
		},
	}
}

// boundContainsPoint checks if a bound contains a point
func boundContainsPoint(b orb.Bound, p orb.Point) bool {
	return p[0] >= b.Min[0]-epsilon && p[0] <= b.Max[0]+epsilon &&
		p[1] >= b.Min[1]-epsilon && p[1] <= b.Max[1]+epsilon
}

// boundContainsPointInterior checks if a point is strictly inside a bound
func boundContainsPointInterior(b orb.Bound, p orb.Point) bool {
	return p[0] > b.Min[0]+epsilon && p[0] < b.Max[0]-epsilon &&
		p[1] > b.Min[1]+epsilon && p[1] < b.Max[1]-epsilon
}

// pointOnBoundBoundary checks if a point is on the boundary of a bound
func pointOnBoundBoundary(p orb.Point, b orb.Bound) bool {
	if !boundContainsPoint(b, p) {
		return false
	}
	return math.Abs(p[0]-b.Min[0]) < epsilon ||
		math.Abs(p[0]-b.Max[0]) < epsilon ||
		math.Abs(p[1]-b.Min[1]) < epsilon ||
		math.Abs(p[1]-b.Max[1]) < epsilon
}

// ringContainsRing checks if ring r1 completely contains ring r2
func ringContainsRing(r1, r2 orb.Ring) bool {
	// All points of r2 must be inside or on r1
	for _, p := range r2 {
		if !planar.RingContains(r1, p) && !pointOnRingBoundary(p, r1) {
			return false
		}
	}
	// No edge crossings allowed (except at boundary)
	for i := 0; i < len(r2)-1; i++ {
		for j := 0; j < len(r1)-1; j++ {
			if segmentsIntersectInterior(r2[i], r2[i+1], r1[j], r1[j+1]) {
				return false
			}
		}
	}
	return true
}

// polygonContainsRing checks if a polygon contains a ring
func polygonContainsRing(poly orb.Polygon, r orb.Ring) bool {
	if len(poly) == 0 {
		return false
	}

	// Ring must be inside the exterior ring
	if !ringContainsRing(poly[0], r) {
		return false
	}

	// Ring must be outside all holes
	for i := 1; i < len(poly); i++ {
		// Check if ring intersects the hole
		if ringsIntersect(poly[i], r) {
			// Check if any point of r is inside the hole
			for _, p := range r {
				if planar.RingContains(poly[i], p) {
					return false
				}
			}
		}
	}

	return true
}

// getGeometryDimension returns the dimension of a geometry (0=point, 1=line, 2=area)
func getGeometryDimension(g orb.Geometry) int {
	switch g.(type) {
	case orb.Point, orb.MultiPoint:
		return 0
	case orb.LineString, orb.MultiLineString:
		return 1
	case orb.Ring, orb.Polygon, orb.MultiPolygon, orb.Bound:
		return 2
	case orb.Collection:
		// Collection dimension is the max of its components
		c := g.(orb.Collection)
		maxDim := -1
		for _, geom := range c {
			d := getGeometryDimension(geom)
			if d > maxDim {
				maxDim = d
			}
		}
		return maxDim
	}
	return -1
}

// isEmpty checks if a geometry is empty
func isEmpty(g orb.Geometry) bool {
	switch geom := g.(type) {
	case orb.Point:
		return false // Points are never empty
	case orb.MultiPoint:
		return len(geom) == 0
	case orb.LineString:
		return len(geom) == 0
	case orb.MultiLineString:
		return len(geom) == 0
	case orb.Ring:
		return len(geom) == 0
	case orb.Polygon:
		return len(geom) == 0 || len(geom[0]) == 0
	case orb.MultiPolygon:
		return len(geom) == 0
	case orb.Collection:
		return len(geom) == 0
	case orb.Bound:
		return geom.IsEmpty()
	}
	return true
}

// lineStringCrossesRingInterior checks if a linestring passes through the interior of a ring
func lineStringCrossesRingInterior(ls orb.LineString, r orb.Ring) bool {
	for _, p := range ls {
		if pointInRingInterior(p, r) {
			return true
		}
	}
	// Also check segment midpoints
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		if pointInRingInterior(mid, r) {
			return true
		}
	}
	return false
}

// lineStringCrossesPolygonInterior checks if a linestring passes through the interior of a polygon
func lineStringCrossesPolygonInterior(ls orb.LineString, poly orb.Polygon) bool {
	for _, p := range ls {
		if pointInPolygonInterior(p, poly) {
			return true
		}
	}
	// Also check segment midpoints
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		if pointInPolygonInterior(mid, poly) {
			return true
		}
	}
	return false
}
