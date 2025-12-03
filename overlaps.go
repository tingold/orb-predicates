package predicates

import (
	"github.com/paulmach/orb"
)

// Overlaps returns true if the geometries have some but not all points in common,
// have the same dimension, and the intersection of the interiors of the two
// geometries has the same dimension as the geometries themselves.
//
// For points: some points match, some don't (both ways)
// For lines: lines share a line segment but neither covers the other
// For areas: areas share some area but neither covers the other
func Overlaps(a, b orb.Geometry) bool {
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

	// Overlaps only applies to geometries of the same dimension
	if dimA != dimB {
		return false
	}

	switch gA := a.(type) {
	case orb.Point:
		// Single point cannot overlap (needs "some but not all")
		return false
	case orb.MultiPoint:
		return overlapsMultiPoint(gA, b)
	case orb.LineString:
		return overlapsLineString(gA, b)
	case orb.MultiLineString:
		return overlapsMultiLineString(gA, b)
	case orb.Ring:
		return overlapsRing(gA, b)
	case orb.Polygon:
		return overlapsPolygon(gA, b)
	case orb.MultiPolygon:
		return overlapsMultiPolygon(gA, b)
	case orb.Collection:
		return overlapsCollection(gA, b)
	case orb.Bound:
		return overlapsBound(gA, b)
	}

	return false
}

// overlapsMultiPoint handles MultiPoint overlaps geometry
func overlapsMultiPoint(mp orb.MultiPoint, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Point:
		// Single point cannot overlap with MultiPoint
		return false
	case orb.MultiPoint:
		return multiPointsOverlap(mp, gB)
	default:
		// Different dimensions
		return false
	}
}

// multiPointsOverlap checks if two multipoints have overlapping points (some shared, some unique each)
func multiPointsOverlap(mp1, mp2 orb.MultiPoint) bool {
	// Need: at least one shared, at least one unique in mp1, at least one unique in mp2
	hasShared := false
	hasUnique1 := false
	hasUnique2 := false

	// Check for shared and unique in mp1
	for _, p1 := range mp1 {
		found := false
		for _, p2 := range mp2 {
			if pointsEqual(p1, p2) {
				found = true
				break
			}
		}
		if found {
			hasShared = true
		} else {
			hasUnique1 = true
		}
	}

	// Check for unique in mp2
	for _, p2 := range mp2 {
		found := false
		for _, p1 := range mp1 {
			if pointsEqual(p1, p2) {
				found = true
				break
			}
		}
		if !found {
			hasUnique2 = true
			break
		}
	}

	return hasShared && hasUnique1 && hasUnique2
}

// overlapsLineString handles LineString overlaps geometry
func overlapsLineString(ls orb.LineString, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.LineString:
		return lineStringsOverlap(ls, gB)
	case orb.MultiLineString:
		// Check if ls overlaps the combined multilinestring
		return lineStringOverlapsMultiLineString(ls, gB)
	default:
		return false
	}
}

// lineStringsOverlap checks if two linestrings overlap (share a segment but neither covers the other)
func lineStringsOverlap(ls1, ls2 orb.LineString) bool {
	if len(ls1) < 2 || len(ls2) < 2 {
		return false
	}

	// Check for collinear overlap (shared line segment)
	hasSharedSegment := false
	for i := 0; i < len(ls1)-1; i++ {
		for j := 0; j < len(ls2)-1; j++ {
			if segmentsShareLine(ls1[i], ls1[i+1], ls2[j], ls2[j+1]) {
				hasSharedSegment = true
				break
			}
		}
		if hasSharedSegment {
			break
		}
	}

	if !hasSharedSegment {
		return false
	}

	// Check that neither covers the other
	if lineStringCoversLineString(ls1, ls2) || lineStringCoversLineString(ls2, ls1) {
		return false
	}

	return true
}

// segmentsShareLine checks if two collinear segments share a portion
func segmentsShareLine(p1, p2, p3, p4 orb.Point) bool {
	// Check if all four points are collinear
	if sign(cross2D(p1, p2, p3)) != 0 || sign(cross2D(p1, p2, p4)) != 0 {
		return false
	}

	// Check for overlap
	return segmentsOverlapInterior(p1, p2, p3, p4)
}

// lineStringOverlapsMultiLineString checks if linestring overlaps multilinestring
func lineStringOverlapsMultiLineString(ls orb.LineString, mls orb.MultiLineString) bool {
	// Check for overlapping segments with any component
	hasOverlap := false
	for _, ls2 := range mls {
		if lineStringsOverlap(ls, ls2) {
			hasOverlap = true
			break
		}
	}

	if !hasOverlap {
		return false
	}

	// Ensure neither covers the other
	if multiLineStringCoversLineString(mls, ls) {
		return false
	}

	// Check if ls covers all of mls
	allCovered := true
	for _, ls2 := range mls {
		if !lineStringCoversLineString(ls, ls2) {
			allCovered = false
			break
		}
	}
	if allCovered {
		return false
	}

	return true
}

// overlapsMultiLineString handles MultiLineString overlaps geometry
func overlapsMultiLineString(mls orb.MultiLineString, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.LineString:
		return lineStringOverlapsMultiLineString(gB, mls)
	case orb.MultiLineString:
		return multiLineStringsOverlap(mls, gB)
	default:
		return false
	}
}

// multiLineStringsOverlap checks if two multilinestrings overlap
func multiLineStringsOverlap(mls1, mls2 orb.MultiLineString) bool {
	// Check for any overlapping pair of linestrings
	hasOverlap := false
	for _, ls1 := range mls1 {
		for _, ls2 := range mls2 {
			if lineStringsOverlap(ls1, ls2) {
				hasOverlap = true
				break
			}
		}
		if hasOverlap {
			break
		}
	}

	if !hasOverlap {
		return false
	}

	// Check that neither fully covers the other
	covered1 := true
	for _, ls := range mls1 {
		if !multiLineStringCoversLineString(mls2, ls) {
			covered1 = false
			break
		}
	}

	covered2 := true
	for _, ls := range mls2 {
		if !multiLineStringCoversLineString(mls1, ls) {
			covered2 = false
			break
		}
	}

	return !covered1 && !covered2
}

// overlapsRing handles Ring overlaps geometry
func overlapsRing(r orb.Ring, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Ring:
		return ringsOverlap(r, gB)
	case orb.Polygon:
		return ringOverlapsPolygon(r, gB)
	case orb.MultiPolygon:
		return ringOverlapsMultiPolygon(r, gB)
	case orb.Bound:
		return ringOverlapsBound(r, gB)
	default:
		return false
	}
}

// ringsOverlap checks if two rings overlap (share area but neither covers the other)
func ringsOverlap(r1, r2 orb.Ring) bool {
	// First check if they intersect at all
	if !ringsIntersect(r1, r2) {
		return false
	}

	// Check for shared interior area
	// Some point of r1's interior must be in r2's interior and vice versa
	r1InR2 := false
	r2InR1 := false

	// Sample some interior points of r1 and check if any are inside r2
	centroid1 := ringCentroid(r1)
	if pointInRingInterior(centroid1, r2) {
		r1InR2 = true
	}

	centroid2 := ringCentroid(r2)
	if pointInRingInterior(centroid2, r1) {
		r2InR1 = true
	}

	// Also check other vertices
	for _, p := range r1 {
		if pointInRingInterior(p, r2) {
			r1InR2 = true
			break
		}
	}

	for _, p := range r2 {
		if pointInRingInterior(p, r1) {
			r2InR1 = true
			break
		}
	}

	// Both must have some interior area in the other
	if !r1InR2 || !r2InR1 {
		return false
	}

	// Neither should cover the other
	if ringCoversRing(r1, r2) || ringCoversRing(r2, r1) {
		return false
	}

	return true
}

// ringOverlapsPolygon checks if ring overlaps polygon
func ringOverlapsPolygon(r orb.Ring, poly orb.Polygon) bool {
	if len(poly) == 0 {
		return false
	}

	// Check if they intersect
	if !ringIntersectsPolygon(r, poly) {
		return false
	}

	// Check for shared interior
	centroidR := ringCentroid(r)
	centroidP := ringCentroid(poly[0])

	rInPoly := pointInPolygonInterior(centroidR, poly)
	polyInR := pointInRingInterior(centroidP, r)

	// Also check vertices
	for _, p := range r {
		if pointInPolygonInterior(p, poly) {
			rInPoly = true
			break
		}
	}

	for _, p := range poly[0] {
		if pointInRingInterior(p, r) {
			polyInR = true
			break
		}
	}

	if !rInPoly || !polyInR {
		return false
	}

	// Neither covers the other
	if ringCoversRing(r, poly[0]) || polygonCoversRing(poly, r) {
		return false
	}

	return true
}

// ringOverlapsMultiPolygon checks if ring overlaps multipolygon
func ringOverlapsMultiPolygon(r orb.Ring, mp orb.MultiPolygon) bool {
	for _, poly := range mp {
		if ringOverlapsPolygon(r, poly) {
			return true
		}
	}
	return false
}

// ringOverlapsBound checks if ring overlaps bound
func ringOverlapsBound(r orb.Ring, b orb.Bound) bool {
	poly := boundToPolygon(b)
	return ringOverlapsPolygon(r, poly)
}

// overlapsPolygon handles Polygon overlaps geometry
func overlapsPolygon(poly orb.Polygon, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Ring:
		return ringOverlapsPolygon(gB, poly)
	case orb.Polygon:
		return polygonsOverlap(poly, gB)
	case orb.MultiPolygon:
		return polygonOverlapsMultiPolygon(poly, gB)
	case orb.Bound:
		return polygonsOverlap(poly, boundToPolygon(gB))
	default:
		return false
	}
}

// polygonsOverlap checks if two polygons overlap
func polygonsOverlap(p1, p2 orb.Polygon) bool {
	if len(p1) == 0 || len(p2) == 0 {
		return false
	}

	// Check if they intersect
	if !polygonsIntersect(p1, p2) {
		return false
	}

	// Check for shared interior area
	centroid1 := ringCentroid(p1[0])
	centroid2 := ringCentroid(p2[0])

	p1InP2 := pointInPolygonInterior(centroid1, p2)
	p2InP1 := pointInPolygonInterior(centroid2, p1)

	// Check vertices too
	for _, p := range p1[0] {
		if pointInPolygonInterior(p, p2) {
			p1InP2 = true
			break
		}
	}

	for _, p := range p2[0] {
		if pointInPolygonInterior(p, p1) {
			p2InP1 = true
			break
		}
	}

	if !p1InP2 || !p2InP1 {
		return false
	}

	// Neither covers the other
	if polygonCoversPolygon(p1, p2) || polygonCoversPolygon(p2, p1) {
		return false
	}

	return true
}

// polygonOverlapsMultiPolygon checks if polygon overlaps multipolygon
func polygonOverlapsMultiPolygon(poly orb.Polygon, mp orb.MultiPolygon) bool {
	for _, poly2 := range mp {
		if polygonsOverlap(poly, poly2) {
			return true
		}
	}
	return false
}

// overlapsMultiPolygon handles MultiPolygon overlaps geometry
func overlapsMultiPolygon(mp orb.MultiPolygon, b orb.Geometry) bool {
	switch gB := b.(type) {
	case orb.Ring:
		return ringOverlapsMultiPolygon(gB, mp)
	case orb.Polygon:
		return polygonOverlapsMultiPolygon(gB, mp)
	case orb.MultiPolygon:
		return multiPolygonsOverlap(mp, gB)
	case orb.Bound:
		return polygonOverlapsMultiPolygon(boundToPolygon(gB), mp)
	default:
		return false
	}
}

// multiPolygonsOverlap checks if two multipolygons overlap
func multiPolygonsOverlap(mp1, mp2 orb.MultiPolygon) bool {
	for _, poly1 := range mp1 {
		for _, poly2 := range mp2 {
			if polygonsOverlap(poly1, poly2) {
				return true
			}
		}
	}
	return false
}

// overlapsCollection handles Collection overlaps geometry
func overlapsCollection(c orb.Collection, b orb.Geometry) bool {
	for _, geom := range c {
		if Overlaps(geom, b) {
			return true
		}
	}
	return false
}

// overlapsBound handles Bound overlaps geometry
func overlapsBound(bound orb.Bound, b orb.Geometry) bool {
	poly := boundToPolygon(bound)
	return overlapsPolygon(poly, b)
}
