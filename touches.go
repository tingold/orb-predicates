package predicates

import (
	"math"

	"github.com/paulmach/orb"
)

// Touches returns true if the geometries have at least one point in common,
// but their interiors do not intersect.
// The geometries must touch only at their boundaries.
func Touches(a, b orb.Geometry) bool {
	// Empty geometries
	if isEmpty(a) || isEmpty(b) {
		return false
	}

	// Quick bounding box check
	if !boundingBoxOverlap(a, b) {
		return false
	}

	// Must intersect but not have overlapping interiors
	if !Intersects(a, b) {
		return false
	}

	// Check that interiors don't intersect
	return !interiorsIntersect(a, b)
}

// interiorsIntersect checks if the interiors of two geometries intersect
func interiorsIntersect(a, b orb.Geometry) bool {
	switch gA := a.(type) {
	case orb.Point:
		return pointInteriorIntersects(gA, b)
	case orb.MultiPoint:
		return multiPointInteriorIntersects(gA, b)
	case orb.LineString:
		return lineStringInteriorIntersects(gA, b)
	case orb.MultiLineString:
		return multiLineStringInteriorIntersects(gA, b)
	case orb.Ring:
		return ringInteriorIntersects(gA, b)
	case orb.Polygon:
		return polygonInteriorIntersects(gA, b)
	case orb.MultiPolygon:
		return multiPolygonInteriorIntersects(gA, b)
	case orb.Collection:
		return collectionInteriorIntersects(gA, b)
	case orb.Bound:
		return boundInteriorIntersects(gA, b)
	}
	return false
}

// pointInteriorIntersects checks if a point's interior intersects geometry b
// For a point, the entire point is its interior (no boundary)
func pointInteriorIntersects(p orb.Point, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		// Points touching: they must be equal, and point interiors = the points themselves
		return pointsEqual(p, gB)
	case orb.MultiPoint:
		for _, pt := range gB {
			if pointsEqual(p, pt) {
				return true
			}
		}
		return false
	case orb.LineString:
		// Point interior intersects linestring interior
		return pointInLineStringInterior(p, gB)
	case orb.MultiLineString:
		for _, ls := range gB {
			if pointInLineStringInterior(p, ls) {
				return true
			}
		}
		return false
	case orb.Ring:
		// Point interior intersects ring interior (inside, not on boundary)
		return pointInRingInterior(p, gB)
	case orb.Polygon:
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
			if pointInteriorIntersects(p, geom) {
				return true
			}
		}
		return false
	case orb.Bound:
		return boundContainsPointInterior(gB, p)
	}
	return false
}

// multiPointInteriorIntersects checks if any point's interior intersects b
func multiPointInteriorIntersects(mp orb.MultiPoint, b orb.Geometry) bool {
	for _, p := range mp {
		if pointInteriorIntersects(p, b) {
			return true
		}
	}
	return false
}

// lineStringInteriorIntersects checks if linestring interior intersects b
func lineStringInteriorIntersects(ls orb.LineString, b orb.Geometry) bool {
	if len(ls) < 2 {
		return false
	}

	switch gB := b.(type) {
	case orb.Point:
		return pointInLineStringInterior(gB, ls)
	case orb.MultiPoint:
		for _, p := range gB {
			if pointInLineStringInterior(p, ls) {
				return true
			}
		}
		return false
	case orb.LineString:
		return lineStringInteriorsIntersect(ls, gB)
	case orb.MultiLineString:
		for _, ls2 := range gB {
			if lineStringInteriorsIntersect(ls, ls2) {
				return true
			}
		}
		return false
	case orb.Ring:
		return lineStringInteriorIntersectsRingInterior(ls, gB)
	case orb.Polygon:
		return lineStringInteriorIntersectsPolygonInterior(ls, gB)
	case orb.MultiPolygon:
		for _, poly := range gB {
			if lineStringInteriorIntersectsPolygonInterior(ls, poly) {
				return true
			}
		}
		return false
	case orb.Collection:
		for _, geom := range gB {
			if lineStringInteriorIntersects(ls, geom) {
				return true
			}
		}
		return false
	case orb.Bound:
		return lineStringInteriorIntersectsBoundInterior(ls, gB)
	}
	return false
}

// lineStringInteriorsIntersect checks if interiors of two linestrings intersect
func lineStringInteriorsIntersect(ls1, ls2 orb.LineString) bool {
	// Check for proper interior crossing
	for i := 0; i < len(ls1)-1; i++ {
		for j := 0; j < len(ls2)-1; j++ {
			if segmentsIntersectInterior(ls1[i], ls1[i+1], ls2[j], ls2[j+1]) {
				return true
			}
		}
	}

	// Check if interior points of ls1 are on interior of ls2 and vice versa
	for i := 1; i < len(ls1)-1; i++ {
		if pointInLineStringInterior(ls1[i], ls2) {
			return true
		}
	}

	for i := 1; i < len(ls2)-1; i++ {
		if pointInLineStringInterior(ls2[i], ls1) {
			return true
		}
	}

	// Check segment midpoints
	for i := 0; i < len(ls1)-1; i++ {
		mid := orb.Point{(ls1[i][0] + ls1[i+1][0]) / 2, (ls1[i][1] + ls1[i+1][1]) / 2}
		if pointInLineStringInterior(mid, ls2) {
			return true
		}
	}

	return false
}

// lineStringInteriorIntersectsRingInterior checks if linestring interior intersects ring interior
func lineStringInteriorIntersectsRingInterior(ls orb.LineString, r orb.Ring) bool {
	// Check if any interior point of ls is inside ring
	for i := 1; i < len(ls)-1; i++ {
		if pointInRingInterior(ls[i], r) {
			return true
		}
	}

	// Check segment midpoints
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		if pointInRingInterior(mid, r) {
			return true
		}
	}

	return false
}

// lineStringInteriorIntersectsPolygonInterior checks if linestring interior intersects polygon interior
func lineStringInteriorIntersectsPolygonInterior(ls orb.LineString, poly orb.Polygon) bool {
	if len(ls) < 2 {
		return false
	}

	// If endpoints are strictly inside, then the interior connected to them is inside
	if pointInPolygonInterior(ls[0], poly) || pointInPolygonInterior(ls[len(ls)-1], poly) {
		return true
	}

	// Check if any interior point of ls is inside polygon
	for i := 1; i < len(ls)-1; i++ {
		if pointInPolygonInterior(ls[i], poly) {
			return true
		}
	}

	// Check segment midpoints
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		if pointInPolygonInterior(mid, poly) {
			return true
		}
	}

	return false
}

// lineStringInteriorIntersectsBoundInterior checks if linestring interior intersects bound interior
func lineStringInteriorIntersectsBoundInterior(ls orb.LineString, b orb.Bound) bool {
	if len(ls) < 2 {
		return false
	}

	if boundContainsPointInterior(b, ls[0]) || boundContainsPointInterior(b, ls[len(ls)-1]) {
		return true
	}

	// Check interior points
	for i := 1; i < len(ls)-1; i++ {
		if boundContainsPointInterior(b, ls[i]) {
			return true
		}
	}

	// Check segment midpoints
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		if boundContainsPointInterior(b, mid) {
			return true
		}
	}

	return false
}

// multiLineStringInteriorIntersects checks if multilinestring interior intersects b
func multiLineStringInteriorIntersects(mls orb.MultiLineString, b orb.Geometry) bool {
	for _, ls := range mls {
		if lineStringInteriorIntersects(ls, b) {
			return true
		}
	}
	return false
}

// ringInteriorIntersects checks if ring interior intersects b
func ringInteriorIntersects(r orb.Ring, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		return pointInRingInterior(gB, r)
	case orb.MultiPoint:
		for _, p := range gB {
			if pointInRingInterior(p, r) {
				return true
			}
		}
		return false
	case orb.LineString:
		return lineStringInteriorIntersectsRingInterior(gB, r)
	case orb.MultiLineString:
		for _, ls := range gB {
			if lineStringInteriorIntersectsRingInterior(ls, r) {
				return true
			}
		}
		return false
	case orb.Ring:
		return ringInteriorsIntersect(r, gB)
	case orb.Polygon:
		return ringInteriorIntersectsPolygonInterior(r, gB)
	case orb.MultiPolygon:
		for _, poly := range gB {
			if ringInteriorIntersectsPolygonInterior(r, poly) {
				return true
			}
		}
		return false
	case orb.Collection:
		for _, geom := range gB {
			if ringInteriorIntersects(r, geom) {
				return true
			}
		}
		return false
	case orb.Bound:
		return ringInteriorIntersectsBoundInterior(r, gB)
	}
	return false
}

// ringInteriorsIntersect checks if the interiors of two rings intersect
func ringInteriorsIntersect(r1, r2 orb.Ring) bool {
	// Use polygon implementation which is robust
	return polygonInteriorsIntersect(orb.Polygon{r1}, orb.Polygon{r2})
}

// ringInteriorIntersectsPolygonInterior checks if ring interior intersects polygon interior
func ringInteriorIntersectsPolygonInterior(r orb.Ring, poly orb.Polygon) bool {
	if len(poly) == 0 {
		return false
	}

	centroid := ringCentroid(r)
	if pointInRingInterior(centroid, r) && pointInPolygonInterior(centroid, poly) {
		return true
	}

	// Check polygon centroid in ring
	polyCentroid := ringCentroid(poly[0])
	if pointInPolygonInterior(polyCentroid, poly) && pointInRingInterior(polyCentroid, r) {
		return true
	}

	return false
}

// ringInteriorIntersectsBoundInterior checks if ring interior intersects bound interior
func ringInteriorIntersectsBoundInterior(r orb.Ring, b orb.Bound) bool {
	centroid := ringCentroid(r)
	if pointInRingInterior(centroid, r) && boundContainsPointInterior(b, centroid) {
		return true
	}

	// Check bound center
	center := orb.Point{(b.Min[0] + b.Max[0]) / 2, (b.Min[1] + b.Max[1]) / 2}
	return pointInRingInterior(center, r)
}

// polygonInteriorIntersects checks if polygon interior intersects b
func polygonInteriorIntersects(poly orb.Polygon, b orb.Geometry) bool {
	if len(poly) == 0 {
		return false
	}

	switch gB := b.(type) {
	case orb.Point:
		return pointInPolygonInterior(gB, poly)
	case orb.MultiPoint:
		for _, p := range gB {
			if pointInPolygonInterior(p, poly) {
				return true
			}
		}
		return false
	case orb.LineString:
		return lineStringInteriorIntersectsPolygonInterior(gB, poly)
	case orb.MultiLineString:
		for _, ls := range gB {
			if lineStringInteriorIntersectsPolygonInterior(ls, poly) {
				return true
			}
		}
		return false
	case orb.Ring:
		return ringInteriorIntersectsPolygonInterior(gB, poly)
	case orb.Polygon:
		return polygonInteriorsIntersect(poly, gB)
	case orb.MultiPolygon:
		for _, poly2 := range gB {
			if polygonInteriorsIntersect(poly, poly2) {
				return true
			}
		}
		return false
	case orb.Collection:
		for _, geom := range gB {
			if polygonInteriorIntersects(poly, geom) {
				return true
			}
		}
		return false
	case orb.Bound:
		return polygonInteriorIntersectsBoundInterior(poly, gB)
	}
	return false
}

// polygonInteriorsIntersect checks if the interiors of two polygons intersect
func polygonInteriorsIntersect(p1, p2 orb.Polygon) bool {
	if len(p1) == 0 || len(p2) == 0 {
		return false
	}

	// 1. Check for proper edge crossings (implies interior intersection)
	for _, r1 := range p1 {
		for _, r2 := range p2 {
			for i := 0; i < len(r1)-1; i++ {
				for j := 0; j < len(r2)-1; j++ {
					if segmentsCrossProper(r1[i], r1[i+1], r2[j], r2[j+1]) {
						return true
					}
				}
			}
		}
	}

	// 2. Check if any vertex of p1 is strictly inside p2
	for _, ring := range p1 {
		for _, p := range ring {
			if pointInPolygonInterior(p, p2) {
				return true
			}
		}
	}

	// 3. Check if any vertex of p2 is strictly inside p1
	for _, ring := range p2 {
		for _, p := range ring {
			if pointInPolygonInterior(p, p1) {
				return true
			}
		}
	}

	// 4. Check for overlapping edges where interiors might merge
	for _, r1 := range p1 {
		for _, r2 := range p2 {
			for i := 0; i < len(r1)-1; i++ {
				for j := 0; j < len(r2)-1; j++ {
					p1a, p1b := r1[i], r1[i+1]
					p2a, p2b := r2[j], r2[j+1]

					if segmentsAreCollinear(p1a, p1b, p2a, p2b) &&
						segmentsOverlapInterior(p1a, p1b, p2a, p2b) {

						// Find midpoint of the overlapping section
						mid := getOverlapMidpoint(p1a, p1b, p2a, p2b)

						// Create probe points perpendicular to the segment
						dx := p1b[0] - p1a[0]
						dy := p1b[1] - p1a[1]
						len := math.Sqrt(dx*dx + dy*dy)
						if len == 0 {
							continue
						}

						// Normalize and rotate 90 degrees
						nx, ny := -dy/len, dx/len

						// Probe distance (epsilon)
						eps := 1e-5

						probe1 := orb.Point{mid[0] + nx*eps, mid[1] + ny*eps}
						probe2 := orb.Point{mid[0] - nx*eps, mid[1] - ny*eps}

						// Check if probe points are inside both polygons
						// One of them should be inside P1 (if valid geometry and not degenerate)
						if pointInPolygonInterior(probe1, p1) && pointInPolygonInterior(probe1, p2) {
							return true
						}
						if pointInPolygonInterior(probe2, p1) && pointInPolygonInterior(probe2, p2) {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

func getOverlapMidpoint(p1, p2, p3, p4 orb.Point) orb.Point {
	// Project to 1D to find overlap range
	horizontal := math.Abs(p2[0]-p1[0]) > math.Abs(p2[1]-p1[1])

	getVal := func(p orb.Point) float64 {
		if horizontal {
			return p[0]
		}
		return p[1]
	}

	v1, v2 := getVal(p1), getVal(p2)
	v3, v4 := getVal(p3), getVal(p4)

	// Sort endpoints of each segment for 1D range logic
	if v1 > v2 {
		v1, v2 = v2, v1
	}
	if v3 > v4 {
		v3, v4 = v4, v3
	}

	// Intersection of [v1, v2] and [v3, v4]
	start := math.Max(v1, v3)
	end := math.Min(v2, v4)
	midVal := (start + end) / 2

	// Map back to point on p1-p2 line
	dx := p2[0] - p1[0]
	dy := p2[1] - p1[1]

	var t float64
	if horizontal {
		if dx == 0 {
			return p1 // Should not happen if horizontal
		}
		t = (midVal - p1[0]) / dx
	} else {
		if dy == 0 {
			return p1
		}
		t = (midVal - p1[1]) / dy
	}

	return orb.Point{p1[0] + t*dx, p1[1] + t*dy}
}

// polygonInteriorIntersectsBoundInterior checks if polygon interior intersects bound interior
func polygonInteriorIntersectsBoundInterior(poly orb.Polygon, b orb.Bound) bool {
	if len(poly) == 0 {
		return false
	}

	centroid := ringCentroid(poly[0])
	if pointInPolygonInterior(centroid, poly) && boundContainsPointInterior(b, centroid) {
		return true
	}

	center := orb.Point{(b.Min[0] + b.Max[0]) / 2, (b.Min[1] + b.Max[1]) / 2}
	return pointInPolygonInterior(center, poly)
}

// multiPolygonInteriorIntersects checks if multipolygon interior intersects b
func multiPolygonInteriorIntersects(mp orb.MultiPolygon, b orb.Geometry) bool {
	for _, poly := range mp {
		if polygonInteriorIntersects(poly, b) {
			return true
		}
	}
	return false
}

// collectionInteriorIntersects checks if collection interior intersects b
func collectionInteriorIntersects(c orb.Collection, b orb.Geometry) bool {
	for _, geom := range c {
		if interiorsIntersect(geom, b) {
			return true
		}
	}
	return false
}

// boundInteriorIntersects checks if bound interior intersects b
func boundInteriorIntersects(bound orb.Bound, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		return boundContainsPointInterior(bound, gB)
	case orb.MultiPoint:
		for _, p := range gB {
			if boundContainsPointInterior(bound, p) {
				return true
			}
		}
		return false
	case orb.LineString:
		return lineStringInteriorIntersectsBoundInterior(gB, bound)
	case orb.MultiLineString:
		for _, ls := range gB {
			if lineStringInteriorIntersectsBoundInterior(ls, bound) {
				return true
			}
		}
		return false
	case orb.Ring:
		return ringInteriorIntersectsBoundInterior(gB, bound)
	case orb.Polygon:
		return polygonInteriorIntersectsBoundInterior(gB, bound)
	case orb.MultiPolygon:
		for _, poly := range gB {
			if polygonInteriorIntersectsBoundInterior(poly, bound) {
				return true
			}
		}
		return false
	case orb.Collection:
		for _, geom := range gB {
			if boundInteriorIntersects(bound, geom) {
				return true
			}
		}
		return false
	case orb.Bound:
		return boundsInteriorsIntersect(bound, gB)
	}
	return false
}

// boundsInteriorsIntersect checks if two bounds' interiors intersect
func boundsInteriorsIntersect(b1, b2 orb.Bound) bool {
	// Interiors intersect if there's overlap beyond just touching edges
	overlapMinX := max(b1.Min[0], b2.Min[0])
	overlapMaxX := min(b1.Max[0], b2.Max[0])
	overlapMinY := max(b1.Min[1], b2.Min[1])
	overlapMaxY := min(b1.Max[1], b2.Max[1])

	// Check if there's actual interior overlap (not just edge touching)
	return overlapMaxX-overlapMinX > epsilon && overlapMaxY-overlapMinY > epsilon
}

// Helper functions for min/max
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// touchesPoint handles Point touches geometry (for reference/completeness)
func touchesPoint(p orb.Point, b orb.Geometry) bool {
	// Point can only touch at boundaries
	switch gB := b.(type) {
	case orb.Point:
		// Two points either intersect completely or not at all - they can't "touch"
		return false
	case orb.MultiPoint:
		return false
	case orb.LineString:
		// Point touches linestring if it's at an endpoint (boundary)
		if len(gB) < 2 {
			return false
		}
		return pointsEqual(p, gB[0]) || pointsEqual(p, gB[len(gB)-1])
	case orb.MultiLineString:
		for _, ls := range gB {
			if len(ls) >= 2 && (pointsEqual(p, ls[0]) || pointsEqual(p, ls[len(ls)-1])) {
				return true
			}
		}
		return false
	case orb.Ring:
		return pointOnRingBoundary(p, gB)
	case orb.Polygon:
		return pointOnPolygonBoundary(p, gB)
	case orb.MultiPolygon:
		for _, poly := range gB {
			if pointOnPolygonBoundary(p, poly) {
				return true
			}
		}
		return false
	case orb.Bound:
		return pointOnBoundBoundary(p, gB)
	default:
		return false
	}
}

// touchesMultiPoint handles MultiPoint touches geometry
func touchesMultiPoint(mp orb.MultiPoint, b orb.Geometry) bool {
	// At least one point must touch, none can be in interior
	hasBoundaryContact := false

	for _, p := range mp {
		if pointInteriorIntersects(p, b) {
			return false
		}
		if intersectsPoint(p, b) {
			hasBoundaryContact = true
		}
	}

	return hasBoundaryContact
}

// touchesLineString handles LineString touches geometry
func touchesLineString(ls orb.LineString, b orb.Geometry) bool {
	// Linestring must intersect b but interiors must not intersect
	if !intersectsLineString(ls, b) {
		return false
	}
	return !lineStringInteriorIntersects(ls, b)
}

// touchesRing handles Ring touches geometry
func touchesRing(r orb.Ring, b orb.Geometry) bool {
	if !intersectsRing(r, b) {
		return false
	}
	return !ringInteriorIntersects(r, b)
}

// touchesPolygon handles Polygon touches geometry
func touchesPolygon(poly orb.Polygon, b orb.Geometry) bool {
	if !intersectsPolygon(poly, b) {
		return false
	}
	return !polygonInteriorIntersects(poly, b)
}

// touchesMultiPolygon handles MultiPolygon touches geometry
func touchesMultiPolygon(mp orb.MultiPolygon, b orb.Geometry) bool {
	hasTouch := false
	for _, poly := range mp {
		if intersectsPolygon(poly, b) {
			if polygonInteriorIntersects(poly, b) {
				return false
			}
			hasTouch = true
		}
	}
	return hasTouch
}

// touchesCollection handles Collection touches geometry
func touchesCollection(c orb.Collection, b orb.Geometry) bool {
	hasTouch := false
	for _, geom := range c {
		if Intersects(geom, b) {
			if interiorsIntersect(geom, b) {
				return false
			}
			hasTouch = true
		}
	}
	return hasTouch
}

// touchesBound handles Bound touches geometry
func touchesBound(bound orb.Bound, b orb.Geometry) bool {
	if !intersectsBound(bound, b) {
		return false
	}
	return !boundInteriorIntersects(bound, b)
}
