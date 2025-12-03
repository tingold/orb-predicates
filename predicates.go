// Package predicates provides spatial relationship predicates for orb geometry types.
//
// This package implements the standard OGC/DE-9IM spatial predicates:
//   - Within: geometry A is completely inside geometry B
//   - Contains: geometry A completely contains geometry B
//   - Covers: no point in B is outside of A
//   - CoveredBy: no point in A is outside of B
//   - Crosses: geometries have some but not all interior points in common
//   - Disjoint: geometries have no points in common
//   - Intersects: geometries have at least one point in common
//   - Overlaps: geometries share some but not all points, same dimension
//   - Touches: geometries touch at boundaries only
//
// Supported geometry types:
//   - Point
//   - MultiPoint
//   - LineString
//   - MultiLineString
//   - Ring
//   - Polygon
//   - MultiPolygon
//   - Collection
//   - Bound
//
// All predicates handle all valid combinations of geometry types.
//
// Example usage:
//
//	poly := orb.Polygon{
//	    orb.Ring{
//	        orb.Point{0, 0},
//	        orb.Point{10, 0},
//	        orb.Point{10, 10},
//	        orb.Point{0, 10},
//	        orb.Point{0, 0},
//	    },
//	}
//	point := orb.Point{5, 5}
//
//	if predicates.Within(point, poly) {
//	    fmt.Println("Point is within polygon")
//	}
//
//	if predicates.Contains(poly, point) {
//	    fmt.Println("Polygon contains point")
//	}
package predicates

// The main predicate functions are implemented in separate files:
// - within.go: Within, Contains
// - covers.go: Covers, CoveredBy
// - intersects.go: Intersects
// - disjoint.go: Disjoint
// - crosses.go: Crosses
// - overlaps.go: Overlaps
// - touches.go: Touches
//
// Helper functions are in helpers.go
