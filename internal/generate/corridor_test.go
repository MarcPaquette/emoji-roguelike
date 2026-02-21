package generate

import (
	"emoji-roguelike/internal/gamemap"
	"math/rand"
	"testing"
)

// allFloorRow checks that every tile at y between x1 and x2 (inclusive) is walkable.
func allFloorRow(gmap *gamemap.GameMap, x1, x2, y int) bool {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		if !gmap.IsWalkable(x, y) {
			return false
		}
	}
	return true
}

// allFloorCol checks that every tile at x between y1 and y2 (inclusive) is walkable.
func allFloorCol(gmap *gamemap.GameMap, y1, y2, x int) bool {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		if !gmap.IsWalkable(x, y) {
			return false
		}
	}
	return true
}

func TestCarveH(t *testing.T) {
	gmap := gamemap.New(20, 20)
	carveH(gmap, 3, 8, 5)

	if !allFloorRow(gmap, 3, 8, 5) {
		t.Error("carveH(3,8,5) should carve floor tiles from x=3 to x=8 at y=5")
	}
	// Tiles just outside the segment must remain walls.
	if gmap.IsWalkable(2, 5) {
		t.Error("tile at x=2 should remain wall (not part of segment)")
	}
	if gmap.IsWalkable(9, 5) {
		t.Error("tile at x=9 should remain wall (not part of segment)")
	}
}

func TestCarveHReversedArgs(t *testing.T) {
	// carveH must swap x1/x2 when x1 > x2.
	gmap := gamemap.New(20, 20)
	carveH(gmap, 8, 3, 5) // reversed
	if !allFloorRow(gmap, 3, 8, 5) {
		t.Error("carveH with reversed x args should still carve x=3..8")
	}
}

func TestCarveV(t *testing.T) {
	gmap := gamemap.New(20, 20)
	carveV(gmap, 2, 7, 4)

	if !allFloorCol(gmap, 2, 7, 4) {
		t.Error("carveV(2,7,4) should carve floor tiles from y=2 to y=7 at x=4")
	}
	if gmap.IsWalkable(4, 1) {
		t.Error("tile at y=1 should remain wall")
	}
	if gmap.IsWalkable(4, 8) {
		t.Error("tile at y=8 should remain wall")
	}
}

func TestCarveVReversedArgs(t *testing.T) {
	gmap := gamemap.New(20, 20)
	carveV(gmap, 7, 2, 4) // reversed
	if !allFloorCol(gmap, 2, 7, 4) {
		t.Error("carveV with reversed y args should still carve y=2..7")
	}
}

func TestCarveZShaped(t *testing.T) {
	// Z-shaped: vertical at x1 from y1→midY, horizontal at midY, vertical at x2 from midY→y2.
	gmap := gamemap.New(20, 20)
	carveZShaped(gmap, 2, 2, 8, 10)
	midY := (2 + 10) / 2 // = 6

	if !allFloorCol(gmap, 2, midY, 2) {
		t.Errorf("Z-shaped: first vertical segment (x=2, y=2..%d) should be floor", midY)
	}
	if !allFloorRow(gmap, 2, 8, midY) {
		t.Errorf("Z-shaped: horizontal segment (y=%d, x=2..8) should be floor", midY)
	}
	if !allFloorCol(gmap, midY, 10, 8) {
		t.Errorf("Z-shaped: last vertical segment (x=8, y=%d..10) should be floor", midY)
	}
}

func TestCorridorStyleStraight(t *testing.T) {
	// Straight: carveH at y1 then carveV at x2.
	gmap := gamemap.New(20, 20)
	cfg := &Config{
		CorridorStyle: CorridorStraight,
		Rand:          rand.New(rand.NewSource(0)),
	}
	carveCorridor(gmap, 2, 2, 8, 8, cfg)

	if !allFloorRow(gmap, 2, 8, 2) {
		t.Error("straight corridor: horizontal segment at y=2 should be floor")
	}
	if !allFloorCol(gmap, 2, 8, 8) {
		t.Error("straight corridor: vertical segment at x=8 should be floor")
	}
}

func TestCorridorStyleZShaped(t *testing.T) {
	gmap := gamemap.New(20, 20)
	cfg := &Config{
		CorridorStyle: CorridorZShaped,
		Rand:          rand.New(rand.NewSource(0)),
	}
	carveCorridor(gmap, 2, 2, 10, 8, cfg)
	midY := (2 + 8) / 2 // = 5

	if !allFloorCol(gmap, 2, midY, 2) {
		t.Errorf("Z-shaped corridor: first vertical (x=2, y=2..%d) should be floor", midY)
	}
	if !allFloorRow(gmap, 2, 10, midY) {
		t.Errorf("Z-shaped corridor: horizontal (y=%d, x=2..10) should be floor", midY)
	}
	if !allFloorCol(gmap, midY, 8, 10) {
		t.Errorf("Z-shaped corridor: last vertical (x=10, y=%d..8) should be floor", midY)
	}
}

func TestCorridorStyleLShaped(t *testing.T) {
	// LShaped: both variants (H then V, or V then H) must connect the endpoints.
	// Test multiple seeds so both random branches are exercised.
	for seed := range 10 {
		gmap := gamemap.New(20, 20)
		cfg := &Config{
			CorridorStyle: CorridorLShaped,
			Rand:          rand.New(rand.NewSource(int64(seed))),
		}
		carveCorridor(gmap, 2, 2, 10, 8, cfg)

		// Both endpoints must be floor tiles after carving.
		if !gmap.IsWalkable(2, 2) {
			t.Errorf("seed %d: start tile (2,2) should be floor after L-shaped corridor", seed)
		}
		if !gmap.IsWalkable(10, 8) {
			t.Errorf("seed %d: end tile (10,8) should be floor after L-shaped corridor", seed)
		}
	}
}
