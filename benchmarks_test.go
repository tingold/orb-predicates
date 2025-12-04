package predicates

import (
	"math"
	"testing"

	"github.com/paulmach/orb"
)

// ==================== Geometry Generators ====================

// generateCircularPolygon creates a polygon approximating a circle with n vertices
func generateCircularPolygon(centerX, centerY, radius float64, n int) orb.Polygon {
	ring := make(orb.Ring, n+1)
	for i := 0; i < n; i++ {
		angle := 2 * math.Pi * float64(i) / float64(n)
		ring[i] = orb.Point{
			centerX + radius*math.Cos(angle),
			centerY + radius*math.Sin(angle),
		}
	}
	ring[n] = ring[0] // Close the ring
	return orb.Polygon{ring}
}

// generateSquarePolygon creates a simple square polygon
func generateSquarePolygon(centerX, centerY, size float64) orb.Polygon {
	half := size / 2
	return orb.Polygon{
		orb.Ring{
			orb.Point{centerX - half, centerY - half},
			orb.Point{centerX + half, centerY - half},
			orb.Point{centerX + half, centerY + half},
			orb.Point{centerX - half, centerY + half},
			orb.Point{centerX - half, centerY - half},
		},
	}
}

// generateLineString creates a linestring with n points
func generateLineString(startX, startY, endX, endY float64, n int) orb.LineString {
	ls := make(orb.LineString, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1)
		ls[i] = orb.Point{
			startX + t*(endX-startX),
			startY + t*(endY-startY),
		}
	}
	return ls
}

// generateMultiPoint creates a multipoint with n points in a grid
func generateMultiPoint(centerX, centerY, spread float64, n int) orb.MultiPoint {
	mp := make(orb.MultiPoint, n)
	side := int(math.Ceil(math.Sqrt(float64(n))))
	step := spread / float64(side)
	startX := centerX - spread/2
	startY := centerY - spread/2

	for i := 0; i < n; i++ {
		row := i / side
		col := i % side
		mp[i] = orb.Point{
			startX + float64(col)*step,
			startY + float64(row)*step,
		}
	}
	return mp
}

// ==================== Test Geometries ====================

var (
	// Small polygon (5 vertices)
	benchSmallPoly = generateSquarePolygon(50, 50, 100)

	// Medium polygon (50 vertices)
	benchMediumPoly = generateCircularPolygon(50, 50, 50, 50)

	// Large polygon (500 vertices)
	benchLargePoly = generateCircularPolygon(50, 50, 50, 500)

	// Very large polygon (2000 vertices)
	benchVeryLargePoly = generateCircularPolygon(50, 50, 50, 2000)

	// Points at various positions
	benchPointInside   = orb.Point{50, 50}
	benchPointOutside  = orb.Point{200, 200}
	benchPointOnEdge   = orb.Point{0, 50}

	// LineStrings
	benchLineInside   = generateLineString(30, 30, 70, 70, 10)
	benchLineCrossing = generateLineString(-50, 50, 150, 50, 10)
	benchLineOutside  = generateLineString(200, 200, 300, 300, 10)

	// Polygons for polygon-polygon tests
	benchPolyContained    = generateSquarePolygon(50, 50, 20)
	benchPolyOverlapping  = generateSquarePolygon(75, 75, 50)
	benchPolyDisjoint     = generateSquarePolygon(300, 300, 50)
	benchPolyTouching     = generateSquarePolygon(150, 50, 100) // Touches at edge
	benchPolyLargeOverlap = generateCircularPolygon(60, 60, 40, 100)

	// MultiPoints
	benchMultiPointSmall  = generateMultiPoint(50, 50, 20, 10)
	benchMultiPointMedium = generateMultiPoint(50, 50, 80, 100)
	benchMultiPointLarge  = generateMultiPoint(50, 50, 80, 500)

	// MultiPolygon
	benchMultiPoly = orb.MultiPolygon{
		generateSquarePolygon(25, 25, 30),
		generateSquarePolygon(75, 75, 30),
	}
)

// ==================== Point vs Polygon Benchmarks ====================

func BenchmarkWithin_PointInSmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Within(benchPointInside, benchSmallPoly)
	}
}

func BenchmarkWithin_PointInMediumPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Within(benchPointInside, benchMediumPoly)
	}
}

func BenchmarkWithin_PointInLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Within(benchPointInside, benchLargePoly)
	}
}

func BenchmarkWithin_PointInVeryLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Within(benchPointInside, benchVeryLargePoly)
	}
}

func BenchmarkWithin_PointOutsideLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Within(benchPointOutside, benchLargePoly)
	}
}

func BenchmarkContains_SmallPolyContainsPoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Contains(benchSmallPoly, benchPointInside)
	}
}

func BenchmarkContains_LargePolyContainsPoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Contains(benchLargePoly, benchPointInside)
	}
}

func BenchmarkIntersects_PointSmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Intersects(benchPointInside, benchSmallPoly)
	}
}

func BenchmarkIntersects_PointLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Intersects(benchPointInside, benchLargePoly)
	}
}

func BenchmarkIntersects_PointOutsideLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Intersects(benchPointOutside, benchLargePoly)
	}
}

func BenchmarkDisjoint_PointSmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Disjoint(benchPointOutside, benchSmallPoly)
	}
}

func BenchmarkDisjoint_PointLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Disjoint(benchPointOutside, benchLargePoly)
	}
}

func BenchmarkCovers_SmallPolyPoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Covers(benchSmallPoly, benchPointInside)
	}
}

func BenchmarkCovers_LargePolyPoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Covers(benchLargePoly, benchPointInside)
	}
}

func BenchmarkCoveredBy_PointSmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CoveredBy(benchPointInside, benchSmallPoly)
	}
}

func BenchmarkCoveredBy_PointLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CoveredBy(benchPointInside, benchLargePoly)
	}
}

func BenchmarkTouches_PointOnEdgeSmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Touches(benchPointOnEdge, benchSmallPoly)
	}
}

func BenchmarkTouches_PointOnEdgeLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Touches(benchPointOnEdge, benchLargePoly)
	}
}

// ==================== LineString vs Polygon Benchmarks ====================

func BenchmarkWithin_LineInSmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Within(benchLineInside, benchSmallPoly)
	}
}

func BenchmarkWithin_LineInLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Within(benchLineInside, benchLargePoly)
	}
}

func BenchmarkIntersects_LineCrossingSmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Intersects(benchLineCrossing, benchSmallPoly)
	}
}

func BenchmarkIntersects_LineCrossingLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Intersects(benchLineCrossing, benchLargePoly)
	}
}

func BenchmarkIntersects_LineOutsideLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Intersects(benchLineOutside, benchLargePoly)
	}
}

func BenchmarkCrosses_LineCrossingSmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Crosses(benchLineCrossing, benchSmallPoly)
	}
}

func BenchmarkCrosses_LineCrossingLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Crosses(benchLineCrossing, benchLargePoly)
	}
}

func BenchmarkCovers_SmallPolyLine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Covers(benchSmallPoly, benchLineInside)
	}
}

func BenchmarkCovers_LargePolyLine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Covers(benchLargePoly, benchLineInside)
	}
}

func BenchmarkTouches_LineSmallPoly(b *testing.B) {
	touchLine := generateLineString(-50, 0, 0, 0, 5) // Touches at corner
	for i := 0; i < b.N; i++ {
		Touches(touchLine, benchSmallPoly)
	}
}

// ==================== Polygon vs Polygon Benchmarks ====================

func BenchmarkWithin_SmallPolyInSmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Within(benchPolyContained, benchSmallPoly)
	}
}

func BenchmarkWithin_SmallPolyInLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Within(benchPolyContained, benchLargePoly)
	}
}

func BenchmarkContains_SmallPolyContainsSmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Contains(benchSmallPoly, benchPolyContained)
	}
}

func BenchmarkContains_LargePolyContainsSmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Contains(benchLargePoly, benchPolyContained)
	}
}

func BenchmarkIntersects_SmallPolySmallPoly_Overlapping(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Intersects(benchSmallPoly, benchPolyOverlapping)
	}
}

func BenchmarkIntersects_LargePolyLargePoly_Overlapping(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Intersects(benchLargePoly, benchPolyLargeOverlap)
	}
}

func BenchmarkIntersects_SmallPolySmallPoly_Disjoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Intersects(benchSmallPoly, benchPolyDisjoint)
	}
}

func BenchmarkIntersects_LargePolyLargePoly_Disjoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Intersects(benchLargePoly, benchPolyDisjoint)
	}
}

func BenchmarkDisjoint_SmallPolySmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Disjoint(benchSmallPoly, benchPolyDisjoint)
	}
}

func BenchmarkDisjoint_LargePolyLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Disjoint(benchLargePoly, benchPolyDisjoint)
	}
}

func BenchmarkOverlaps_SmallPolySmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Overlaps(benchSmallPoly, benchPolyOverlapping)
	}
}

func BenchmarkOverlaps_LargePolyLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Overlaps(benchLargePoly, benchPolyLargeOverlap)
	}
}

func BenchmarkTouches_SmallPolySmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Touches(benchSmallPoly, benchPolyTouching)
	}
}

func BenchmarkTouches_LargePolyLargePoly(b *testing.B) {
	largeTouching := generateCircularPolygon(150, 50, 50, 500)
	for i := 0; i < b.N; i++ {
		Touches(benchLargePoly, largeTouching)
	}
}

func BenchmarkCovers_SmallPolySmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Covers(benchSmallPoly, benchPolyContained)
	}
}

func BenchmarkCovers_LargePolySmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Covers(benchLargePoly, benchPolyContained)
	}
}

// ==================== MultiPoint Benchmarks ====================

func BenchmarkWithin_MultiPointSmallInSmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Within(benchMultiPointSmall, benchSmallPoly)
	}
}

func BenchmarkWithin_MultiPointLargeInLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Within(benchMultiPointLarge, benchLargePoly)
	}
}

func BenchmarkIntersects_MultiPointSmallPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Intersects(benchMultiPointMedium, benchSmallPoly)
	}
}

func BenchmarkIntersects_MultiPointLargePoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Intersects(benchMultiPointLarge, benchLargePoly)
	}
}

func BenchmarkCrosses_MultiPointSmallPoly(b *testing.B) {
	// MultiPoint that crosses (some in, some out)
	mpCrossing := orb.MultiPoint{
		orb.Point{50, 50},   // inside
		orb.Point{200, 200}, // outside
	}
	for i := 0; i < b.N; i++ {
		Crosses(mpCrossing, benchSmallPoly)
	}
}

func BenchmarkCrosses_MultiPointLargePoly(b *testing.B) {
	mpCrossing := orb.MultiPoint{
		orb.Point{50, 50},
		orb.Point{200, 200},
	}
	for i := 0; i < b.N; i++ {
		Crosses(mpCrossing, benchLargePoly)
	}
}

// ==================== MultiPolygon Benchmarks ====================

func BenchmarkWithin_PointInMultiPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Within(orb.Point{25, 25}, benchMultiPoly)
	}
}

func BenchmarkContains_MultiPolyContainsPoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Contains(benchMultiPoly, orb.Point{25, 25})
	}
}

func BenchmarkIntersects_MultiPolyPoly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Intersects(benchMultiPoly, benchSmallPoly)
	}
}

// ==================== Bound Benchmarks ====================

func BenchmarkWithin_PointInBound(b *testing.B) {
	bound := orb.Bound{Min: orb.Point{0, 0}, Max: orb.Point{100, 100}}
	for i := 0; i < b.N; i++ {
		Within(benchPointInside, bound)
	}
}

func BenchmarkIntersects_PolyBound(b *testing.B) {
	bound := orb.Bound{Min: orb.Point{0, 0}, Max: orb.Point{100, 100}}
	for i := 0; i < b.N; i++ {
		Intersects(benchSmallPoly, bound)
	}
}

func BenchmarkCovers_BoundPoly(b *testing.B) {
	bound := orb.Bound{Min: orb.Point{-50, -50}, Max: orb.Point{150, 150}}
	for i := 0; i < b.N; i++ {
		Covers(bound, benchSmallPoly)
	}
}

// ==================== LineString vs LineString Benchmarks ====================

func BenchmarkIntersects_LineStringLineString_Small(b *testing.B) {
	ls1 := generateLineString(0, 0, 100, 100, 10)
	ls2 := generateLineString(0, 100, 100, 0, 10)
	for i := 0; i < b.N; i++ {
		Intersects(ls1, ls2)
	}
}

func BenchmarkIntersects_LineStringLineString_Large(b *testing.B) {
	ls1 := generateLineString(0, 0, 100, 100, 100)
	ls2 := generateLineString(0, 100, 100, 0, 100)
	for i := 0; i < b.N; i++ {
		Intersects(ls1, ls2)
	}
}

func BenchmarkIntersects_LineStringLineString_VeryLarge(b *testing.B) {
	ls1 := generateLineString(0, 0, 100, 100, 500)
	ls2 := generateLineString(0, 100, 100, 0, 500)
	for i := 0; i < b.N; i++ {
		Intersects(ls1, ls2)
	}
}

func BenchmarkIntersects_LineStringLineString_Disjoint(b *testing.B) {
	ls1 := generateLineString(0, 0, 100, 100, 100)
	ls2 := generateLineString(200, 200, 300, 300, 100)
	for i := 0; i < b.N; i++ {
		Intersects(ls1, ls2)
	}
}

func BenchmarkCrosses_LineStringLineString_Small(b *testing.B) {
	ls1 := generateLineString(0, 0, 100, 100, 10)
	ls2 := generateLineString(0, 100, 100, 0, 10)
	for i := 0; i < b.N; i++ {
		Crosses(ls1, ls2)
	}
}

func BenchmarkCrosses_LineStringLineString_Large(b *testing.B) {
	ls1 := generateLineString(0, 0, 100, 100, 100)
	ls2 := generateLineString(0, 100, 100, 0, 100)
	for i := 0; i < b.N; i++ {
		Crosses(ls1, ls2)
	}
}

// ==================== Ring Benchmarks ====================

func BenchmarkIntersects_RingRing_Small(b *testing.B) {
	r1 := benchSmallPoly[0]
	r2 := benchPolyOverlapping[0]
	for i := 0; i < b.N; i++ {
		Intersects(r1, r2)
	}
}

func BenchmarkIntersects_RingRing_Large(b *testing.B) {
	r1 := benchLargePoly[0]
	r2 := benchPolyLargeOverlap[0]
	for i := 0; i < b.N; i++ {
		Intersects(r1, r2)
	}
}

func BenchmarkWithin_RingInRing(b *testing.B) {
	outer := benchLargePoly[0]
	inner := benchPolyContained[0]
	for i := 0; i < b.N; i++ {
		Within(inner, outer)
	}
}

// ==================== Collection Benchmarks ====================

func BenchmarkWithin_CollectionInPoly(b *testing.B) {
	coll := orb.Collection{
		orb.Point{50, 50},
		generateLineString(40, 40, 60, 60, 5),
		generateSquarePolygon(50, 50, 10),
	}
	for i := 0; i < b.N; i++ {
		Within(coll, benchSmallPoly)
	}
}

func BenchmarkIntersects_CollectionPoly(b *testing.B) {
	coll := orb.Collection{
		orb.Point{50, 50},
		generateLineString(40, 40, 60, 60, 5),
		generateSquarePolygon(50, 50, 10),
	}
	for i := 0; i < b.N; i++ {
		Intersects(coll, benchSmallPoly)
	}
}

// ==================== Worst Case Benchmarks ====================

// These test potentially expensive edge cases

func BenchmarkWorstCase_ManyPointsOnBoundary(b *testing.B) {
	// Test with many points exactly on the polygon boundary
	poly := generateCircularPolygon(50, 50, 50, 100)
	pointOnBoundary := poly[0][0] // First vertex is on boundary
	for i := 0; i < b.N; i++ {
		Within(pointOnBoundary, poly)
	}
}

func BenchmarkWorstCase_NearlyCollinearSegments(b *testing.B) {
	// Test with nearly collinear line segments
	ls1 := orb.LineString{
		orb.Point{0, 0},
		orb.Point{100, 0.0001}, // Nearly horizontal
	}
	ls2 := orb.LineString{
		orb.Point{50, -1},
		orb.Point{50, 1},
	}
	for i := 0; i < b.N; i++ {
		Intersects(ls1, ls2)
	}
}

func BenchmarkWorstCase_DegeneratePolygon(b *testing.B) {
	// Test with a very thin polygon
	thin := orb.Polygon{
		orb.Ring{
			orb.Point{0, 0},
			orb.Point{100, 0},
			orb.Point{100, 0.001},
			orb.Point{0, 0.001},
			orb.Point{0, 0},
		},
	}
	for i := 0; i < b.N; i++ {
		Contains(thin, orb.Point{50, 0.0005})
	}
}

// ==================== Helper Function Benchmarks ====================

func BenchmarkHelper_PointOnSegment(b *testing.B) {
	p := orb.Point{50, 50}
	a := orb.Point{0, 0}
	seg_b := orb.Point{100, 100}
	for i := 0; i < b.N; i++ {
		pointOnSegment(p, a, seg_b)
	}
}

func BenchmarkHelper_SegmentsIntersect(b *testing.B) {
	p1, p2 := orb.Point{0, 0}, orb.Point{100, 100}
	p3, p4 := orb.Point{0, 100}, orb.Point{100, 0}
	for i := 0; i < b.N; i++ {
		segmentsIntersect(p1, p2, p3, p4)
	}
}

func BenchmarkHelper_PointsEqual(b *testing.B) {
	p1 := orb.Point{50.0000001, 50.0000001}
	p2 := orb.Point{50, 50}
	for i := 0; i < b.N; i++ {
		pointsEqual(p1, p2)
	}
}

func BenchmarkHelper_Cross2D(b *testing.B) {
	p1 := orb.Point{0, 0}
	p2 := orb.Point{100, 100}
	p3 := orb.Point{50, 50}
	for i := 0; i < b.N; i++ {
		cross2D(p1, p2, p3)
	}
}

func BenchmarkHelper_BoundingBoxOverlap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		boundingBoxOverlap(benchSmallPoly, benchPolyOverlapping)
	}
}

func BenchmarkHelper_PointOnRingBoundary_Small(b *testing.B) {
	ring := benchSmallPoly[0]
	for i := 0; i < b.N; i++ {
		pointOnRingBoundary(benchPointOnEdge, ring)
	}
}

func BenchmarkHelper_PointOnRingBoundary_Large(b *testing.B) {
	ring := benchLargePoly[0]
	for i := 0; i < b.N; i++ {
		pointOnRingBoundary(benchPointOnEdge, ring)
	}
}

func BenchmarkHelper_PointInRingInterior_Small(b *testing.B) {
	ring := benchSmallPoly[0]
	for i := 0; i < b.N; i++ {
		pointInRingInterior(benchPointInside, ring)
	}
}

func BenchmarkHelper_PointInRingInterior_Large(b *testing.B) {
	ring := benchLargePoly[0]
	for i := 0; i < b.N; i++ {
		pointInRingInterior(benchPointInside, ring)
	}
}
