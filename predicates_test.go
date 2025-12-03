package predicates

import (
	"testing"

	"github.com/paulmach/orb"
)

// Test geometries
var (
	// Basic shapes
	unitSquare = orb.Polygon{
		orb.Ring{
			orb.Point{0, 0},
			orb.Point{10, 0},
			orb.Point{10, 10},
			orb.Point{0, 10},
			orb.Point{0, 0},
		},
	}

	smallSquare = orb.Polygon{
		orb.Ring{
			orb.Point{2, 2},
			orb.Point{4, 2},
			orb.Point{4, 4},
			orb.Point{2, 4},
			orb.Point{2, 2},
		},
	}

	overlappingSquare = orb.Polygon{
		orb.Ring{
			orb.Point{5, 5},
			orb.Point{15, 5},
			orb.Point{15, 15},
			orb.Point{5, 15},
			orb.Point{5, 5},
		},
	}

	disjointSquare = orb.Polygon{
		orb.Ring{
			orb.Point{20, 20},
			orb.Point{30, 20},
			orb.Point{30, 30},
			orb.Point{20, 30},
			orb.Point{20, 20},
		},
	}

	touchingSquare = orb.Polygon{
		orb.Ring{
			orb.Point{10, 0},
			orb.Point{20, 0},
			orb.Point{20, 10},
			orb.Point{10, 10},
			orb.Point{10, 0},
		},
	}

	// Points
	pointInside    = orb.Point{5, 5}
	pointOutside   = orb.Point{15, 15}
	pointOnEdge    = orb.Point{5, 0}
	pointOnCorner  = orb.Point{0, 0}
	pointInSmall   = orb.Point{3, 3}
	pointInOverlap = orb.Point{7, 7} // In both unitSquare and overlappingSquare

	// LineStrings
	lineInside = orb.LineString{
		orb.Point{2, 2},
		orb.Point{8, 8},
	}

	lineCrossing = orb.LineString{
		orb.Point{-5, 5},
		orb.Point{15, 5},
	}

	lineOutside = orb.LineString{
		orb.Point{15, 15},
		orb.Point{20, 20},
	}

	lineTouching = orb.LineString{
		orb.Point{-5, 0},
		orb.Point{0, 0},
	}

	lineOnEdge = orb.LineString{
		orb.Point{0, 0},
		orb.Point{10, 0},
	}

	// MultiPoints
	multiPointAllInside = orb.MultiPoint{
		orb.Point{2, 2},
		orb.Point{5, 5},
		orb.Point{8, 8},
	}

	multiPointSomeInside = orb.MultiPoint{
		orb.Point{5, 5},
		orb.Point{15, 15},
	}

	multiPointAllOutside = orb.MultiPoint{
		orb.Point{15, 15},
		orb.Point{20, 20},
	}

	// Rings
	ringInside = orb.Ring{
		orb.Point{2, 2},
		orb.Point{4, 2},
		orb.Point{4, 4},
		orb.Point{2, 4},
		orb.Point{2, 2},
	}

	ringOverlapping = orb.Ring{
		orb.Point{5, 5},
		orb.Point{15, 5},
		orb.Point{15, 15},
		orb.Point{5, 15},
		orb.Point{5, 5},
	}

	// MultiLineString
	multiLineString = orb.MultiLineString{
		orb.LineString{orb.Point{1, 1}, orb.Point{2, 2}},
		orb.LineString{orb.Point{3, 3}, orb.Point{4, 4}},
	}

	// MultiPolygon
	multiPolygon = orb.MultiPolygon{
		orb.Polygon{
			orb.Ring{
				orb.Point{0, 0},
				orb.Point{5, 0},
				orb.Point{5, 5},
				orb.Point{0, 5},
				orb.Point{0, 0},
			},
		},
		orb.Polygon{
			orb.Ring{
				orb.Point{10, 10},
				orb.Point{15, 10},
				orb.Point{15, 15},
				orb.Point{10, 15},
				orb.Point{10, 10},
			},
		},
	}

	// Bound
	testBound = orb.Bound{
		Min: orb.Point{0, 0},
		Max: orb.Point{10, 10},
	}

	// Collection
	testCollection = orb.Collection{
		orb.Point{5, 5},
		orb.LineString{orb.Point{1, 1}, orb.Point{2, 2}},
	}
)

// ==================== Within Tests ====================

func TestWithin(t *testing.T) {
	tests := []struct {
		name     string
		a, b     orb.Geometry
		expected bool
	}{
		// Point in Polygon
		{"point inside polygon", pointInside, unitSquare, true},
		{"point outside polygon", pointOutside, unitSquare, false},
		{"point on edge (not within, just on boundary)", pointOnEdge, unitSquare, false},
		{"point on corner (not within, just on boundary)", pointOnCorner, unitSquare, false},

		// Point in Point
		{"point in same point", pointInside, pointInside, true},
		{"point in different point", pointInside, pointOutside, false},

		// LineString in Polygon
		{"line inside polygon", lineInside, unitSquare, true},
		{"line crossing polygon", lineCrossing, unitSquare, false},
		{"line outside polygon", lineOutside, unitSquare, false},

		// Polygon in Polygon
		{"small polygon inside large", smallSquare, unitSquare, true},
		{"overlapping polygons", overlappingSquare, unitSquare, false},
		{"disjoint polygons", disjointSquare, unitSquare, false},

		// MultiPoint in Polygon
		{"multipoint all inside", multiPointAllInside, unitSquare, true},
		{"multipoint some inside", multiPointSomeInside, unitSquare, false},
		{"multipoint all outside", multiPointAllOutside, unitSquare, false},

		// Ring in Polygon
		{"ring inside polygon", ringInside, unitSquare, true},
		{"ring overlapping polygon", ringOverlapping, unitSquare, false},

		// Point in Bound
		{"point inside bound", pointInside, testBound, true},
		{"point outside bound", pointOutside, testBound, false},

		// Point in MultiPolygon
		{"point in multipolygon", orb.Point{2, 2}, multiPolygon, true},
		{"point outside multipolygon", orb.Point{7, 7}, multiPolygon, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Within(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Within(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ==================== Contains Tests ====================

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		a, b     orb.Geometry
		expected bool
	}{
		{"polygon contains point inside", unitSquare, pointInside, true},
		{"polygon not contains point outside", unitSquare, pointOutside, false},
		{"polygon contains smaller polygon", unitSquare, smallSquare, true},
		{"polygon not contains overlapping polygon", unitSquare, overlappingSquare, false},
		{"bound contains point", testBound, pointInside, true},
		{"multipolygon contains point", multiPolygon, orb.Point{2, 2}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Contains(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Contains(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ==================== Intersects Tests ====================

func TestIntersects(t *testing.T) {
	tests := []struct {
		name     string
		a, b     orb.Geometry
		expected bool
	}{
		// Point-Point
		{"same points intersect", pointInside, pointInside, true},
		{"different points don't intersect", pointInside, pointOutside, false},

		// Point-Polygon
		{"point inside polygon intersects", pointInside, unitSquare, true},
		{"point outside polygon no intersect", pointOutside, unitSquare, false},
		{"point on edge intersects", pointOnEdge, unitSquare, true},
		{"point on corner intersects", pointOnCorner, unitSquare, true},

		// LineString-Polygon
		{"line inside polygon intersects", lineInside, unitSquare, true},
		{"line crossing polygon intersects", lineCrossing, unitSquare, true},
		{"line outside polygon no intersect", lineOutside, unitSquare, false},

		// Polygon-Polygon
		{"overlapping polygons intersect", unitSquare, overlappingSquare, true},
		{"disjoint polygons no intersect", unitSquare, disjointSquare, false},
		{"touching polygons intersect", unitSquare, touchingSquare, true},

		// LineString-LineString
		{"crossing lines intersect", lineCrossing, lineInside, true},
		{"parallel lines no intersect", lineOutside,
			orb.LineString{orb.Point{25, 25}, orb.Point{30, 30}}, false},

		// MultiPoint
		{"multipoint intersects polygon", multiPointAllInside, unitSquare, true},
		{"multipoint some intersects polygon", multiPointSomeInside, unitSquare, true},
		{"multipoint outside no intersect", multiPointAllOutside, unitSquare, false},

		// Bound
		{"bound intersects polygon", testBound, unitSquare, true},
		{"point intersects bound", pointInside, testBound, true},

		// Collection
		{"collection intersects polygon", testCollection, unitSquare, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Intersects(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Intersects(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ==================== Disjoint Tests ====================

func TestDisjoint(t *testing.T) {
	tests := []struct {
		name     string
		a, b     orb.Geometry
		expected bool
	}{
		{"same points not disjoint", pointInside, pointInside, false},
		{"different points disjoint", pointInside, pointOutside, true},
		{"point inside polygon not disjoint", pointInside, unitSquare, false},
		{"point outside polygon disjoint", pointOutside, unitSquare, true},
		{"overlapping polygons not disjoint", unitSquare, overlappingSquare, false},
		{"disjoint polygons disjoint", unitSquare, disjointSquare, true},
		{"touching polygons not disjoint", unitSquare, touchingSquare, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Disjoint(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Disjoint(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ==================== Covers Tests ====================

func TestCovers(t *testing.T) {
	tests := []struct {
		name     string
		a, b     orb.Geometry
		expected bool
	}{
		// Covers allows boundary contact, Contains does not
		{"polygon covers point inside", unitSquare, pointInside, true},
		{"polygon covers point on edge", unitSquare, pointOnEdge, true},
		{"polygon covers point on corner", unitSquare, pointOnCorner, true},
		{"polygon not covers point outside", unitSquare, pointOutside, false},
		{"polygon covers smaller polygon", unitSquare, smallSquare, true},
		{"polygon covers line inside", unitSquare, lineInside, true},
		{"polygon covers line on edge", unitSquare, lineOnEdge, true},
		{"polygon not covers crossing line", unitSquare, lineCrossing, false},

		// Point covers
		{"point covers itself", pointInside, pointInside, true},
		{"point not covers different point", pointInside, pointOutside, false},

		// LineString covers
		{"line covers point on it", lineInside, orb.Point{5, 5}, true},
		{"line not covers point off it", lineInside, orb.Point{1, 5}, false},

		// Bound covers
		{"bound covers point inside", testBound, pointInside, true},
		{"bound covers point on edge", testBound, pointOnEdge, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Covers(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Covers(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ==================== CoveredBy Tests ====================

func TestCoveredBy(t *testing.T) {
	tests := []struct {
		name     string
		a, b     orb.Geometry
		expected bool
	}{
		{"point inside covered by polygon", pointInside, unitSquare, true},
		{"point on edge covered by polygon", pointOnEdge, unitSquare, true},
		{"point outside not covered by polygon", pointOutside, unitSquare, false},
		{"small polygon covered by large", smallSquare, unitSquare, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CoveredBy(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("CoveredBy(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ==================== Crosses Tests ====================

func TestCrosses(t *testing.T) {
	tests := []struct {
		name     string
		a, b     orb.Geometry
		expected bool
	}{
		// Line crosses polygon (enters and exits)
		{"line crosses polygon", lineCrossing, unitSquare, true},
		{"line inside polygon no cross", lineInside, unitSquare, false},
		{"line outside polygon no cross", lineOutside, unitSquare, false},

		// Line crosses line
		{"lines cross", orb.LineString{orb.Point{0, 5}, orb.Point{10, 5}},
			orb.LineString{orb.Point{5, 0}, orb.Point{5, 10}}, true},
		{"parallel lines no cross", orb.LineString{orb.Point{0, 0}, orb.Point{10, 0}},
			orb.LineString{orb.Point{0, 5}, orb.Point{10, 5}}, false},

		// MultiPoint crosses polygon (some inside, some outside)
		{"multipoint crosses polygon", multiPointSomeInside, unitSquare, true},
		{"multipoint all inside no cross", multiPointAllInside, unitSquare, false},

		// Points cannot cross
		{"point cannot cross polygon", pointInside, unitSquare, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Crosses(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Crosses(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ==================== Overlaps Tests ====================

func TestOverlaps(t *testing.T) {
	tests := []struct {
		name     string
		a, b     orb.Geometry
		expected bool
	}{
		// Polygon overlaps polygon
		{"overlapping polygons", unitSquare, overlappingSquare, true},
		{"contained polygon no overlap", unitSquare, smallSquare, false},
		{"disjoint polygons no overlap", unitSquare, disjointSquare, false},

		// MultiPoint overlaps multipoint
		{"multipoints overlap", orb.MultiPoint{orb.Point{1, 1}, orb.Point{2, 2}, orb.Point{3, 3}},
			orb.MultiPoint{orb.Point{2, 2}, orb.Point{4, 4}, orb.Point{5, 5}}, true},
		{"multipoints same no overlap", multiPointAllInside, multiPointAllInside, false},

		// Different dimensions cannot overlap
		{"point and polygon cannot overlap", pointInside, unitSquare, false},
		{"line and polygon cannot overlap", lineInside, unitSquare, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Overlaps(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Overlaps(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ==================== Touches Tests ====================

func TestTouches(t *testing.T) {
	tests := []struct {
		name     string
		a, b     orb.Geometry
		expected bool
	}{
		// Polygons touch at edge
		{"touching polygons", unitSquare, touchingSquare, true},
		{"overlapping polygons don't touch", unitSquare, overlappingSquare, false},
		{"disjoint polygons don't touch", unitSquare, disjointSquare, false},

		// Point touches polygon edge
		{"point on edge touches polygon", pointOnEdge, unitSquare, true},
		{"point on corner touches polygon", pointOnCorner, unitSquare, true},
		{"point inside doesn't touch polygon", pointInside, unitSquare, false},

		// Line touches polygon
		{"line touching polygon at endpoint", lineTouching, unitSquare, true},
		{"line crossing doesn't touch", lineCrossing, unitSquare, false},

		// Line touches line at endpoint
		{"lines touch at endpoints",
			orb.LineString{orb.Point{0, 0}, orb.Point{5, 5}},
			orb.LineString{orb.Point{5, 5}, orb.Point{10, 0}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Touches(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Touches(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ==================== Edge Cases ====================

func TestEmptyGeometries(t *testing.T) {
	emptyPolygon := orb.Polygon{}
	emptyLineString := orb.LineString{}
	emptyMultiPoint := orb.MultiPoint{}

	tests := []struct {
		name      string
		predicate func(a, b orb.Geometry) bool
		a, b      orb.Geometry
		expected  bool
	}{
		{"within empty polygon", Within, pointInside, emptyPolygon, false},
		{"contains empty polygon", Contains, unitSquare, emptyPolygon, false},
		{"intersects empty linestring", Intersects, pointInside, emptyLineString, false},
		{"disjoint empty multipoint", Disjoint, pointInside, emptyMultiPoint, true},
		{"covers empty polygon", Covers, unitSquare, emptyPolygon, false},
		{"crosses empty linestring", Crosses, lineCrossing, emptyLineString, false},
		{"overlaps empty polygon", Overlaps, unitSquare, emptyPolygon, false},
		{"touches empty polygon", Touches, unitSquare, emptyPolygon, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.predicate(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestSymmetry(t *testing.T) {
	// Test symmetric predicates
	t.Run("Intersects is symmetric", func(t *testing.T) {
		if Intersects(pointInside, unitSquare) != Intersects(unitSquare, pointInside) {
			t.Error("Intersects should be symmetric")
		}
		if Intersects(lineInside, unitSquare) != Intersects(unitSquare, lineInside) {
			t.Error("Intersects should be symmetric")
		}
	})

	t.Run("Disjoint is symmetric", func(t *testing.T) {
		if Disjoint(pointOutside, unitSquare) != Disjoint(unitSquare, pointOutside) {
			t.Error("Disjoint should be symmetric")
		}
	})

	t.Run("Touches is symmetric", func(t *testing.T) {
		if Touches(unitSquare, touchingSquare) != Touches(touchingSquare, unitSquare) {
			t.Error("Touches should be symmetric")
		}
	})

	t.Run("Overlaps is symmetric", func(t *testing.T) {
		if Overlaps(unitSquare, overlappingSquare) != Overlaps(overlappingSquare, unitSquare) {
			t.Error("Overlaps should be symmetric")
		}
	})
}

func TestInverseRelationships(t *testing.T) {
	// Within/Contains are inverses
	t.Run("Within and Contains are inverses", func(t *testing.T) {
		if Within(pointInside, unitSquare) != Contains(unitSquare, pointInside) {
			t.Error("Within(a,b) should equal Contains(b,a)")
		}
		if Within(smallSquare, unitSquare) != Contains(unitSquare, smallSquare) {
			t.Error("Within(a,b) should equal Contains(b,a)")
		}
	})

	// Covers/CoveredBy are inverses
	t.Run("Covers and CoveredBy are inverses", func(t *testing.T) {
		if Covers(unitSquare, pointInside) != CoveredBy(pointInside, unitSquare) {
			t.Error("Covers(a,b) should equal CoveredBy(b,a)")
		}
		if Covers(unitSquare, pointOnEdge) != CoveredBy(pointOnEdge, unitSquare) {
			t.Error("Covers(a,b) should equal CoveredBy(b,a)")
		}
	})

	// Intersects/Disjoint are complements
	t.Run("Intersects and Disjoint are complements", func(t *testing.T) {
		if Intersects(pointInside, unitSquare) == Disjoint(pointInside, unitSquare) {
			t.Error("Intersects and Disjoint should be complements")
		}
		if Intersects(pointOutside, unitSquare) == Disjoint(pointOutside, unitSquare) {
			t.Error("Intersects and Disjoint should be complements")
		}
	})
}

// ==================== Bound Tests ====================

func TestBoundPredicates(t *testing.T) {
	bound := orb.Bound{Min: orb.Point{0, 0}, Max: orb.Point{10, 10}}
	innerBound := orb.Bound{Min: orb.Point{2, 2}, Max: orb.Point{8, 8}}
	overlappingBound := orb.Bound{Min: orb.Point{5, 5}, Max: orb.Point{15, 15}}
	disjointBound := orb.Bound{Min: orb.Point{20, 20}, Max: orb.Point{30, 30}}

	tests := []struct {
		name      string
		predicate func(a, b orb.Geometry) bool
		a, b      orb.Geometry
		expected  bool
	}{
		{"bound contains inner bound", Contains, bound, innerBound, true},
		{"bound intersects overlapping bound", Intersects, bound, overlappingBound, true},
		{"bound disjoint from disjoint bound", Disjoint, bound, disjointBound, true},
		{"bound covers point", Covers, bound, orb.Point{5, 5}, true},
		{"point within bound", Within, orb.Point{5, 5}, bound, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.predicate(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

// ==================== Collection Tests ====================

func TestCollectionPredicates(t *testing.T) {
	collection := orb.Collection{
		orb.Point{5, 5},
		orb.LineString{orb.Point{1, 1}, orb.Point{9, 9}},
		smallSquare,
	}

	tests := []struct {
		name      string
		predicate func(a, b orb.Geometry) bool
		a, b      orb.Geometry
		expected  bool
	}{
		{"collection within polygon", Within, collection, unitSquare, true},
		{"polygon contains collection", Contains, unitSquare, collection, true},
		{"collection intersects polygon", Intersects, collection, unitSquare, true},
		{"collection disjoint from distant polygon", Disjoint, collection, disjointSquare, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.predicate(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

// ==================== MultiPolygon Tests ====================

func TestMultiPolygonPredicates(t *testing.T) {
	mp := orb.MultiPolygon{
		orb.Polygon{orb.Ring{
			orb.Point{0, 0}, orb.Point{5, 0}, orb.Point{5, 5}, orb.Point{0, 5}, orb.Point{0, 0},
		}},
		orb.Polygon{orb.Ring{
			orb.Point{10, 10}, orb.Point{15, 10}, orb.Point{15, 15}, orb.Point{10, 15}, orb.Point{10, 10},
		}},
	}

	tests := []struct {
		name      string
		predicate func(a, b orb.Geometry) bool
		a, b      orb.Geometry
		expected  bool
	}{
		{"multipolygon contains point in first poly", Contains, mp, orb.Point{2, 2}, true},
		{"multipolygon contains point in second poly", Contains, mp, orb.Point{12, 12}, true},
		{"multipolygon not contains point between", Contains, mp, orb.Point{7, 7}, false},
		{"point within multipolygon", Within, orb.Point{2, 2}, mp, true},
		{"multipolygon intersects polygon", Intersects, mp, unitSquare, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.predicate(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

// ==================== Ring Tests ====================

func TestRingPredicates(t *testing.T) {
	ring := orb.Ring{
		orb.Point{0, 0}, orb.Point{10, 0}, orb.Point{10, 10}, orb.Point{0, 10}, orb.Point{0, 0},
	}

	smallRing := orb.Ring{
		orb.Point{2, 2}, orb.Point{4, 2}, orb.Point{4, 4}, orb.Point{2, 4}, orb.Point{2, 2},
	}

	tests := []struct {
		name      string
		predicate func(a, b orb.Geometry) bool
		a, b      orb.Geometry
		expected  bool
	}{
		{"ring contains point", Contains, ring, orb.Point{5, 5}, true},
		{"ring contains smaller ring", Contains, ring, smallRing, true},
		{"small ring within larger ring", Within, smallRing, ring, true},
		{"ring intersects polygon", Intersects, ring, unitSquare, true},
		{"point on ring boundary", Covers, ring, orb.Point{5, 0}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.predicate(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

// ==================== MultiLineString Tests ====================

func TestMultiLineStringPredicates(t *testing.T) {
	mls := orb.MultiLineString{
		orb.LineString{orb.Point{1, 1}, orb.Point{4, 4}},
		orb.LineString{orb.Point{6, 6}, orb.Point{9, 9}},
	}

	tests := []struct {
		name      string
		predicate func(a, b orb.Geometry) bool
		a, b      orb.Geometry
		expected  bool
	}{
		{"multilinestring within polygon", Within, mls, unitSquare, true},
		{"polygon contains multilinestring", Contains, unitSquare, mls, true},
		{"multilinestring intersects polygon", Intersects, mls, unitSquare, true},
		{"point on multilinestring", Intersects, orb.Point{2, 2}, mls, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.predicate(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}
