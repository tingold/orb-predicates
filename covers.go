package predicates

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
)

// Covers returns true if no point in geometry b is outside of geometry a.
// This is similar to Contains but allows b to be entirely on the boundary of a.
func Covers(a, b orb.Geometry) bool {
	// Empty geometries
	if isEmpty(a) || isEmpty(b) {
		return false
	}

	// Quick bounding box check
	ba := a.Bound()
	bb := b.Bound()
	if bb.Min[0] < ba.Min[0]-epsilon || bb.Max[0] > ba.Max[0]+epsilon ||
		bb.Min[1] < ba.Min[1]-epsilon || bb.Max[1] > ba.Max[1]+epsilon {
		return false
	}

	switch gA := a.(type) {
	case orb.Point:
		return coversPoint(gA, b)
	case orb.MultiPoint:
		return coversMultiPoint(gA, b)
	case orb.LineString:
		return coversLineString(gA, b)
	case orb.MultiLineString:
		return coversMultiLineString(gA, b)
	case orb.Ring:
		return coversRing(gA, b)
	case orb.Polygon:
		return coversPolygon(gA, b)
	case orb.MultiPolygon:
		return coversMultiPolygon(gA, b)
	case orb.Collection:
		return coversCollection(gA, b)
	case orb.Bound:
		return coversBound(gA, b)
	}

	return false
}

// CoveredBy returns true if no point in geometry a is outside of geometry b.
func CoveredBy(a, b orb.Geometry) bool {
	return Covers(b, a)
}

// coversPoint handles Point covers all geometry types
func coversPoint(p orb.Point, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		return pointsEqual(p, gB)
	case orb.MultiPoint:
		// Point can only cover MultiPoint if all points are equal to it
		for _, pt := range gB {
			if !pointsEqual(p, pt) {
				return false
			}
		}
		return true
	default:
		// Point cannot cover higher dimensional geometries
		return false
	}
}

// coversMultiPoint handles MultiPoint covers all geometry types
func coversMultiPoint(mp orb.MultiPoint, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		for _, pt := range mp {
			if pointsEqual(pt, gB) {
				return true
			}
		}
		return false
	case orb.MultiPoint:
		// All points in gB must be in mp
		for _, p := range gB {
			found := false
			for _, pt := range mp {
				if pointsEqual(p, pt) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// coversLineString handles LineString covers all geometry types
func coversLineString(ls orb.LineString, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		return pointIntersectsLineString(gB, ls)
	case orb.MultiPoint:
		for _, p := range gB {
			if !pointIntersectsLineString(p, ls) {
				return false
			}
		}
		return true
	case orb.LineString:
		return lineStringCoversLineString(ls, gB)
	case orb.MultiLineString:
		for _, ls2 := range gB {
			if !lineStringCoversLineString(ls, ls2) {
				return false
			}
		}
		return true
	default:
		// LineString cannot cover 2D geometries
		return false
	}
}

// lineStringCoversLineString checks if ls1 covers ls2
func lineStringCoversLineString(ls1, ls2 orb.LineString) bool {
	// All points of ls2 must be on ls1
	for _, p := range ls2 {
		if !pointIntersectsLineString(p, ls1) {
			return false
		}
	}

	// All segment midpoints must also be on ls1
	for i := 0; i < len(ls2)-1; i++ {
		mid := orb.Point{(ls2[i][0] + ls2[i+1][0]) / 2, (ls2[i][1] + ls2[i+1][1]) / 2}
		if !pointIntersectsLineString(mid, ls1) {
			return false
		}
	}

	return true
}

// coversMultiLineString handles MultiLineString covers all geometry types
func coversMultiLineString(mls orb.MultiLineString, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		for _, ls := range mls {
			if pointIntersectsLineString(gB, ls) {
				return true
			}
		}
		return false
	case orb.MultiPoint:
		for _, p := range gB {
			covered := false
			for _, ls := range mls {
				if pointIntersectsLineString(p, ls) {
					covered = true
					break
				}
			}
			if !covered {
				return false
			}
		}
		return true
	case orb.LineString:
		return multiLineStringCoversLineString(mls, gB)
	case orb.MultiLineString:
		for _, ls := range gB {
			if !multiLineStringCoversLineString(mls, ls) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// multiLineStringCoversLineString checks if mls covers ls
func multiLineStringCoversLineString(mls orb.MultiLineString, ls orb.LineString) bool {
	// All points of ls must be on some component of mls
	for _, p := range ls {
		covered := false
		for _, ls2 := range mls {
			if pointIntersectsLineString(p, ls2) {
				covered = true
				break
			}
		}
		if !covered {
			return false
		}
	}

	// Check midpoints
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		covered := false
		for _, ls2 := range mls {
			if pointIntersectsLineString(mid, ls2) {
				covered = true
				break
			}
		}
		if !covered {
			return false
		}
	}

	return true
}

// coversRing handles Ring covers all geometry types
func coversRing(r orb.Ring, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		return planar.RingContains(r, gB) || pointOnRingBoundary(gB, r)
	case orb.MultiPoint:
		for _, p := range gB {
			if !planar.RingContains(r, p) && !pointOnRingBoundary(p, r) {
				return false
			}
		}
		return true
	case orb.LineString:
		return ringCoversLineString(r, gB)
	case orb.MultiLineString:
		for _, ls := range gB {
			if !ringCoversLineString(r, ls) {
				return false
			}
		}
		return true
	case orb.Ring:
		return ringCoversRing(r, gB)
	case orb.Polygon:
		return ringCoversPolygon(r, gB)
	case orb.MultiPolygon:
		for _, poly := range gB {
			if !ringCoversPolygon(r, poly) {
				return false
			}
		}
		return true
	case orb.Collection:
		for _, geom := range gB {
			if !coversRing(r, geom) {
				return false
			}
		}
		return true
	case orb.Bound:
		return ringCoversBound(r, gB)
	}
	return false
}

// ringCoversLineString checks if ring covers linestring
func ringCoversLineString(r orb.Ring, ls orb.LineString) bool {
	for _, p := range ls {
		if !planar.RingContains(r, p) && !pointOnRingBoundary(p, r) {
			return false
		}
	}

	// Check midpoints
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		if !planar.RingContains(r, mid) && !pointOnRingBoundary(mid, r) {
			return false
		}
	}

	return true
}

// ringCoversRing checks if r1 covers r2
func ringCoversRing(r1, r2 orb.Ring) bool {
	// All points of r2 must be inside or on boundary of r1
	for _, p := range r2 {
		if !planar.RingContains(r1, p) && !pointOnRingBoundary(p, r1) {
			return false
		}
	}

	// No interior edge crossings that would place parts outside
	for i := 0; i < len(r2)-1; i++ {
		mid := orb.Point{(r2[i][0] + r2[i+1][0]) / 2, (r2[i][1] + r2[i+1][1]) / 2}
		if !planar.RingContains(r1, mid) && !pointOnRingBoundary(mid, r1) {
			return false
		}
	}

	return true
}

// ringCoversPolygon checks if ring covers polygon
func ringCoversPolygon(r orb.Ring, poly orb.Polygon) bool {
	if len(poly) == 0 {
		return true
	}

	// All points of exterior ring must be covered
	for _, p := range poly[0] {
		if !planar.RingContains(r, p) && !pointOnRingBoundary(p, r) {
			return false
		}
	}

	return true
}

// ringCoversBound checks if ring covers bound
func ringCoversBound(r orb.Ring, b orb.Bound) bool {
	corners := []orb.Point{
		{b.Min[0], b.Min[1]},
		{b.Max[0], b.Min[1]},
		{b.Max[0], b.Max[1]},
		{b.Min[0], b.Max[1]},
	}

	for _, c := range corners {
		if !planar.RingContains(r, c) && !pointOnRingBoundary(c, r) {
			return false
		}
	}

	return true
}

// coversPolygon handles Polygon covers all geometry types
func coversPolygon(poly orb.Polygon, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		return planar.PolygonContains(poly, gB) || pointOnPolygonBoundary(gB, poly)
	case orb.MultiPoint:
		for _, p := range gB {
			if !planar.PolygonContains(poly, p) && !pointOnPolygonBoundary(p, poly) {
				return false
			}
		}
		return true
	case orb.LineString:
		return polygonCoversLineString(poly, gB)
	case orb.MultiLineString:
		for _, ls := range gB {
			if !polygonCoversLineString(poly, ls) {
				return false
			}
		}
		return true
	case orb.Ring:
		return polygonCoversRing(poly, gB)
	case orb.Polygon:
		return polygonCoversPolygon(poly, gB)
	case orb.MultiPolygon:
		for _, poly2 := range gB {
			if !polygonCoversPolygon(poly, poly2) {
				return false
			}
		}
		return true
	case orb.Collection:
		for _, geom := range gB {
			if !coversPolygon(poly, geom) {
				return false
			}
		}
		return true
	case orb.Bound:
		return polygonCoversBound(poly, gB)
	}
	return false
}

// polygonCoversLineString checks if polygon covers linestring
func polygonCoversLineString(poly orb.Polygon, ls orb.LineString) bool {
	for _, p := range ls {
		if !planar.PolygonContains(poly, p) && !pointOnPolygonBoundary(p, poly) {
			return false
		}
	}

	// Check midpoints
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		if !planar.PolygonContains(poly, mid) && !pointOnPolygonBoundary(mid, poly) {
			return false
		}
	}

	return true
}

// polygonCoversRing checks if polygon covers ring
func polygonCoversRing(poly orb.Polygon, r orb.Ring) bool {
	for _, p := range r {
		if !planar.PolygonContains(poly, p) && !pointOnPolygonBoundary(p, poly) {
			return false
		}
	}

	// Check edge midpoints
	for i := 0; i < len(r)-1; i++ {
		mid := orb.Point{(r[i][0] + r[i+1][0]) / 2, (r[i][1] + r[i+1][1]) / 2}
		if !planar.PolygonContains(poly, mid) && !pointOnPolygonBoundary(mid, poly) {
			return false
		}
	}

	return true
}

// polygonCoversPolygon checks if poly1 covers poly2
func polygonCoversPolygon(poly1, poly2 orb.Polygon) bool {
	if len(poly2) == 0 {
		return true
	}

	// All points of poly2's exterior must be covered by poly1
	for _, p := range poly2[0] {
		if !planar.PolygonContains(poly1, p) && !pointOnPolygonBoundary(p, poly1) {
			return false
		}
	}

	// Check edge midpoints
	for i := 0; i < len(poly2[0])-1; i++ {
		mid := orb.Point{(poly2[0][i][0] + poly2[0][i+1][0]) / 2, (poly2[0][i][1] + poly2[0][i+1][1]) / 2}
		if !planar.PolygonContains(poly1, mid) && !pointOnPolygonBoundary(mid, poly1) {
			return false
		}
	}

	return true
}

// polygonCoversBound checks if polygon covers bound
func polygonCoversBound(poly orb.Polygon, b orb.Bound) bool {
	corners := []orb.Point{
		{b.Min[0], b.Min[1]},
		{b.Max[0], b.Min[1]},
		{b.Max[0], b.Max[1]},
		{b.Min[0], b.Max[1]},
	}

	for _, c := range corners {
		if !planar.PolygonContains(poly, c) && !pointOnPolygonBoundary(c, poly) {
			return false
		}
	}

	// Check edge midpoints
	edges := []orb.Point{
		{(b.Min[0] + b.Max[0]) / 2, b.Min[1]},
		{b.Max[0], (b.Min[1] + b.Max[1]) / 2},
		{(b.Min[0] + b.Max[0]) / 2, b.Max[1]},
		{b.Min[0], (b.Min[1] + b.Max[1]) / 2},
	}

	for _, e := range edges {
		if !planar.PolygonContains(poly, e) && !pointOnPolygonBoundary(e, poly) {
			return false
		}
	}

	return true
}

// coversMultiPolygon handles MultiPolygon covers all geometry types
func coversMultiPolygon(mp orb.MultiPolygon, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		for _, poly := range mp {
			if planar.PolygonContains(poly, gB) || pointOnPolygonBoundary(gB, poly) {
				return true
			}
		}
		return false
	case orb.MultiPoint:
		for _, p := range gB {
			covered := false
			for _, poly := range mp {
				if planar.PolygonContains(poly, p) || pointOnPolygonBoundary(p, poly) {
					covered = true
					break
				}
			}
			if !covered {
				return false
			}
		}
		return true
	case orb.LineString:
		return multiPolygonCoversLineString(mp, gB)
	case orb.MultiLineString:
		for _, ls := range gB {
			if !multiPolygonCoversLineString(mp, ls) {
				return false
			}
		}
		return true
	case orb.Ring:
		return multiPolygonCoversRing(mp, gB)
	case orb.Polygon:
		return multiPolygonCoversPolygon(mp, gB)
	case orb.MultiPolygon:
		for _, poly := range gB {
			if !multiPolygonCoversPolygon(mp, poly) {
				return false
			}
		}
		return true
	case orb.Collection:
		for _, geom := range gB {
			if !coversMultiPolygon(mp, geom) {
				return false
			}
		}
		return true
	case orb.Bound:
		return multiPolygonCoversBound(mp, gB)
	}
	return false
}

// multiPolygonCoversLineString checks if multipolygon covers linestring
func multiPolygonCoversLineString(mp orb.MultiPolygon, ls orb.LineString) bool {
	// Each point must be covered by some polygon
	for _, p := range ls {
		covered := false
		for _, poly := range mp {
			if planar.PolygonContains(poly, p) || pointOnPolygonBoundary(p, poly) {
				covered = true
				break
			}
		}
		if !covered {
			return false
		}
	}

	// Check midpoints
	for i := 0; i < len(ls)-1; i++ {
		mid := orb.Point{(ls[i][0] + ls[i+1][0]) / 2, (ls[i][1] + ls[i+1][1]) / 2}
		covered := false
		for _, poly := range mp {
			if planar.PolygonContains(poly, mid) || pointOnPolygonBoundary(mid, poly) {
				covered = true
				break
			}
		}
		if !covered {
			return false
		}
	}

	return true
}

// multiPolygonCoversRing checks if multipolygon covers ring
func multiPolygonCoversRing(mp orb.MultiPolygon, r orb.Ring) bool {
	for _, p := range r {
		covered := false
		for _, poly := range mp {
			if planar.PolygonContains(poly, p) || pointOnPolygonBoundary(p, poly) {
				covered = true
				break
			}
		}
		if !covered {
			return false
		}
	}

	// Check edge midpoints
	for i := 0; i < len(r)-1; i++ {
		mid := orb.Point{(r[i][0] + r[i+1][0]) / 2, (r[i][1] + r[i+1][1]) / 2}
		covered := false
		for _, poly := range mp {
			if planar.PolygonContains(poly, mid) || pointOnPolygonBoundary(mid, poly) {
				covered = true
				break
			}
		}
		if !covered {
			return false
		}
	}

	return true
}

// multiPolygonCoversPolygon checks if multipolygon covers polygon
func multiPolygonCoversPolygon(mp orb.MultiPolygon, poly orb.Polygon) bool {
	if len(poly) == 0 {
		return true
	}

	// Check if any single polygon covers it
	for _, p := range mp {
		if polygonCoversPolygon(p, poly) {
			return true
		}
	}

	// Otherwise, check point-by-point coverage
	for _, pt := range poly[0] {
		covered := false
		for _, p := range mp {
			if planar.PolygonContains(p, pt) || pointOnPolygonBoundary(pt, p) {
				covered = true
				break
			}
		}
		if !covered {
			return false
		}
	}

	// Check edge midpoints
	for i := 0; i < len(poly[0])-1; i++ {
		mid := orb.Point{(poly[0][i][0] + poly[0][i+1][0]) / 2, (poly[0][i][1] + poly[0][i+1][1]) / 2}
		covered := false
		for _, p := range mp {
			if planar.PolygonContains(p, mid) || pointOnPolygonBoundary(mid, p) {
				covered = true
				break
			}
		}
		if !covered {
			return false
		}
	}

	return true
}

// multiPolygonCoversBound checks if multipolygon covers bound
func multiPolygonCoversBound(mp orb.MultiPolygon, b orb.Bound) bool {
	poly := boundToPolygon(b)
	return multiPolygonCoversPolygon(mp, poly)
}

// coversCollection handles Collection covers all geometry types
func coversCollection(c orb.Collection, b orb.Geometry) bool {
	// For each point in b, check if it's covered by some geometry in c
	switch gB := b.(type) {
	case orb.Point:
		for _, geom := range c {
			if Covers(geom, gB) {
				return true
			}
		}
		return false
	case orb.MultiPoint:
		for _, p := range gB {
			covered := false
			for _, geom := range c {
				if Covers(geom, p) {
					covered = true
					break
				}
			}
			if !covered {
				return false
			}
		}
		return true
	default:
		// For complex geometries, check if any single geometry covers it
		for _, geom := range c {
			if Covers(geom, b) {
				return true
			}
		}
		return false
	}
}

// coversBound handles Bound covers all geometry types
func coversBound(bound orb.Bound, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		return boundContainsPoint(bound, gB)
	case orb.MultiPoint:
		for _, p := range gB {
			if !boundContainsPoint(bound, p) {
				return false
			}
		}
		return true
	case orb.LineString:
		for _, p := range gB {
			if !boundContainsPoint(bound, p) {
				return false
			}
		}
		return true
	case orb.MultiLineString:
		for _, ls := range gB {
			for _, p := range ls {
				if !boundContainsPoint(bound, p) {
					return false
				}
			}
		}
		return true
	case orb.Ring:
		for _, p := range gB {
			if !boundContainsPoint(bound, p) {
				return false
			}
		}
		return true
	case orb.Polygon:
		for _, ring := range gB {
			for _, p := range ring {
				if !boundContainsPoint(bound, p) {
					return false
				}
			}
		}
		return true
	case orb.MultiPolygon:
		for _, poly := range gB {
			for _, ring := range poly {
				for _, p := range ring {
					if !boundContainsPoint(bound, p) {
						return false
					}
				}
			}
		}
		return true
	case orb.Collection:
		for _, geom := range gB {
			if !coversBound(bound, geom) {
				return false
			}
		}
		return true
	case orb.Bound:
		return bound.Min[0] <= gB.Min[0]+epsilon &&
			bound.Min[1] <= gB.Min[1]+epsilon &&
			bound.Max[0] >= gB.Max[0]-epsilon &&
			bound.Max[1] >= gB.Max[1]-epsilon
	}
	return false
}
