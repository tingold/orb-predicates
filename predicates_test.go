package predicates

import (
	"testing"

	"github.com/paulmach/orb"
)

func TestWithin(t *testing.T) {
	// Point in Polygon
	poly := orb.Polygon{
		orb.Ring{
			orb.Point{0, 0},
			orb.Point{10, 0},
			orb.Point{10, 10},
			orb.Point{0, 10},
			orb.Point{0, 0},
		},
	}
	p1 := orb.Point{5, 5}
	if !Within(p1, poly) {
		t.Errorf("Point should be within polygon")
	}

	p2 := orb.Point{15, 15}
	if Within(p2, poly) {
		t.Errorf("Point should not be within polygon")
	}

	// LineString in Polygon
	ls1 := orb.LineString{
		orb.Point{1, 1},
		orb.Point{2, 2},
	}
	if !Within(ls1, poly) {
		t.Errorf("LineString should be within polygon")
	}

	ls2 := orb.LineString{
		orb.Point{1, 1},
		orb.Point{12, 12},
	}
	if Within(ls2, poly) {
		t.Errorf("LineString should not be within polygon")
	}

	// Polygon in Polygon
	poly2 := orb.Polygon{
		orb.Ring{
			orb.Point{1, 1},
			orb.Point{2, 1},
			orb.Point{2, 2},
			orb.Point{1, 2},
			orb.Point{1, 1},
		},
	}
	if !Within(poly2, poly) {
		t.Errorf("Polygon should be within polygon")
	}

	poly3 := orb.Polygon{
		orb.Ring{
			orb.Point{1, 1},
			orb.Point{12, 1},
			orb.Point{12, 12},
			orb.Point{1, 12},
			orb.Point{1, 1},
		},
	}
	if Within(poly3, poly) {
		t.Errorf("Polygon should not be within polygon")
	}
}

func TestContains(t *testing.T) {
	poly := orb.Polygon{
		orb.Ring{
			orb.Point{0, 0},
			orb.Point{10, 0},
			orb.Point{10, 10},
			orb.Point{0, 10},
			orb.Point{0, 0},
		},
	}
	p1 := orb.Point{5, 5}
	if !Contains(poly, p1) {
		t.Errorf("Polygon should contain point")
	}

	p2 := orb.Point{15, 15}
	if Contains(poly, p2) {
		t.Errorf("Polygon should not contain point")
	}
}
