# orb-predicates

Spatial relationship predicates for [orb](https://github.com/paulmach/orb) geometry types.

This package implements the standard **OGC/DE-9IM spatial predicates** for determining topological relationships between geometries.

## Installation

```bash
go get github.com/tingold/orb-predicates
```

## Predicates

| Predicate    | Description                                                |
|--------------|-----------------------------------------------------------|
| `Within`     | Geometry A is completely inside geometry B                 |
| `Contains`   | Geometry A completely contains geometry B                  |
| `Covers`     | No point in B is outside of A (boundary contact allowed)   |
| `CoveredBy`  | No point in A is outside of B                              |
| `Intersects` | Geometries share at least one point in common              |
| `Disjoint`   | Geometries have no points in common                        |
| `Touches`    | Geometries touch at boundaries only, interiors don't intersect |
| `Crosses`    | Geometries have some but not all interior points in common |
| `Overlaps`   | Geometries share some but not all points, with same dimension |

## Supported Geometry Types

All predicates support the following `orb` geometry types:

- `orb.Point`
- `orb.MultiPoint`
- `orb.LineString`
- `orb.MultiLineString`
- `orb.Ring`
- `orb.Polygon`
- `orb.MultiPolygon`
- `orb.Collection`
- `orb.Bound`

## Usage

```go
package main

import (
    "fmt"

    "github.com/paulmach/orb"
    predicates "github.com/tingold/orb-predicates"
)

func main() {
    // Define a polygon
    poly := orb.Polygon{
        orb.Ring{
            orb.Point{0, 0},
            orb.Point{10, 0},
            orb.Point{10, 10},
            orb.Point{0, 10},
            orb.Point{0, 0},
        },
    }

    // Points to test
    pointInside := orb.Point{5, 5}
    pointOutside := orb.Point{15, 15}
    pointOnEdge := orb.Point{5, 0}

    // Within / Contains
    fmt.Println(predicates.Within(pointInside, poly))   // true
    fmt.Println(predicates.Contains(poly, pointInside)) // true

    // Intersects / Disjoint
    fmt.Println(predicates.Intersects(pointOnEdge, poly)) // true
    fmt.Println(predicates.Disjoint(pointOutside, poly))  // true

    // Touches
    fmt.Println(predicates.Touches(pointOnEdge, poly)) // true

    // Line crossing polygon
    line := orb.LineString{
        orb.Point{-5, 5},
        orb.Point{15, 5},
    }
    fmt.Println(predicates.Crosses(line, poly)) // true
}
```

### Working with Different Geometry Combinations

```go
// Polygon-Polygon relationships
smallSquare := orb.Polygon{
    orb.Ring{
        orb.Point{2, 2},
        orb.Point{4, 2},
        orb.Point{4, 4},
        orb.Point{2, 4},
        orb.Point{2, 2},
    },
}

largeSquare := orb.Polygon{
    orb.Ring{
        orb.Point{0, 0},
        orb.Point{10, 0},
        orb.Point{10, 10},
        orb.Point{0, 10},
        orb.Point{0, 0},
    },
}

fmt.Println(predicates.Within(smallSquare, largeSquare))   // true
fmt.Println(predicates.Contains(largeSquare, smallSquare)) // true
fmt.Println(predicates.Covers(largeSquare, smallSquare))   // true

// Overlapping polygons
overlapping := orb.Polygon{
    orb.Ring{
        orb.Point{5, 5},
        orb.Point{15, 5},
        orb.Point{15, 15},
        orb.Point{5, 15},
        orb.Point{5, 5},
    },
}

fmt.Println(predicates.Overlaps(largeSquare, overlapping)) // true
fmt.Println(predicates.Intersects(largeSquare, overlapping)) // true
```

### Using Bounds

```go
bound := orb.Bound{
    Min: orb.Point{0, 0},
    Max: orb.Point{10, 10},
}

point := orb.Point{5, 5}
fmt.Println(predicates.Covers(bound, point)) // true
```

## Predicate Semantics

### Within vs Covers

- **Within**: The interior of A must intersect the interior of B. Boundary-only contact is not sufficient.
- **Covers**: Every point of B must be in A (interior or boundary). A geometry entirely on the boundary of another is covered but not within.

### Intersects vs Touches

- **Intersects**: Any shared point counts.
- **Touches**: Only boundary contact; interiors must not intersect.

### Crosses vs Overlaps

- **Crosses**: For geometries of different dimensions (e.g., line crossing polygon) or lines crossing lines where the intersection is a point.
- **Overlaps**: For geometries of the same dimension that share area/length but neither contains the other.

## License

See LICENSE file.

