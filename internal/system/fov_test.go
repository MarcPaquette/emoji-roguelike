package system

import (
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/gamemap"
	"testing"
)

// openMapFOV creates a fully-open (all floor) map for FOV tests.
func openMapFOV(width, height int) *gamemap.GameMap {
	gmap := gamemap.New(width, height)
	for y := range height {
		for x := range width {
			gmap.Set(x, y, gamemap.MakeFloor())
		}
	}
	return gmap
}

// makePlayerAt creates an entity with only CPosition — suitable for FOV tests.
func makePlayerAt(w *ecs.World, px, py int) ecs.EntityID {
	player := w.CreateEntity()
	w.Add(player, component.Position{X: px, Y: py})
	return player
}

func TestFOVOriginAlwaysVisible(t *testing.T) {
	gmap := openMapFOV(20, 20)
	w := ecs.NewWorld()
	player := makePlayerAt(w, 5, 5)

	UpdateFOV(w, gmap, player, 5)

	if !gmap.At(5, 5).Visible {
		t.Error("player's own tile must always be visible")
	}
	if !gmap.At(5, 5).Explored {
		t.Error("player's own tile must be marked explored")
	}
}

func TestFOVClearsOldVisibility(t *testing.T) {
	gmap := openMapFOV(20, 20)
	w := ecs.NewWorld()
	player := makePlayerAt(w, 5, 5)

	// Pre-mark every tile as visible.
	for y := range 20 {
		for x := range 20 {
			gmap.At(x, y).Visible = true
		}
	}
	UpdateFOV(w, gmap, player, 3)

	// A tile far away should have been cleared.
	if gmap.At(19, 19).Visible {
		t.Error("UpdateFOV should clear stale visibility before recalculating")
	}
}

func TestFOVNearbyTilesVisible(t *testing.T) {
	// Tiles at cardinal distance 3 on a fully open map must be lit with radius=5.
	// The FOV radius condition is: dx²+dy² < radius² → 9 < 25 → true.
	gmap := openMapFOV(20, 20)
	w := ecs.NewWorld()
	player := makePlayerAt(w, 10, 10)

	UpdateFOV(w, gmap, player, 5)

	for _, pos := range [][2]int{{10, 7}, {10, 13}, {7, 10}, {13, 10}} {
		x, y := pos[0], pos[1]
		if !gmap.At(x, y).Visible {
			t.Errorf("tile (%d,%d) at distance 3 should be visible (radius=5)", x, y)
		}
		if !gmap.At(x, y).Explored {
			t.Errorf("tile (%d,%d) at distance 3 should be marked explored", x, y)
		}
	}
}

func TestFOVRadiusLimitsVisibility(t *testing.T) {
	// The loop bound (j <= radius) prevents processing tiles beyond radius.
	// With radius=4, tiles at distance 5 are never reached.
	gmap := openMapFOV(20, 20)
	w := ecs.NewWorld()
	player := makePlayerAt(w, 10, 10)

	UpdateFOV(w, gmap, player, 4)

	// These tiles are exactly 5 away (outside radius=4).
	for _, pos := range [][2]int{{10, 15}, {10, 5}, {15, 10}, {5, 10}} {
		x, y := pos[0], pos[1]
		if gmap.At(x, y).Visible {
			t.Errorf("tile (%d,%d) at distance 5 should not be visible with radius=4", x, y)
		}
	}
}

func TestFOVWallBlocksLight(t *testing.T) {
	// A wall at (10,8) blocks the tile at (10,7) from being lit when the
	// player is at (10,10) with radius 8.
	gmap := openMapFOV(20, 20)
	gmap.Set(10, 8, gamemap.MakeWall())
	w := ecs.NewWorld()
	player := makePlayerAt(w, 10, 10)

	UpdateFOV(w, gmap, player, 8)

	// The wall tile itself is visible (at the shadow edge).
	if !gmap.At(10, 8).Visible {
		t.Error("the wall tile at (10,8) should be visible")
	}
	// The tile directly behind the wall must be blocked.
	if gmap.At(10, 7).Visible {
		t.Error("tile (10,7) behind the wall at (10,8) should not be visible")
	}
}

func TestFOVNoPlayerPositionNoPanic(t *testing.T) {
	// UpdateFOV must not panic when the player entity has no CPosition.
	gmap := openMapFOV(10, 10)
	w := ecs.NewWorld()
	player := w.CreateEntity() // no Position added

	UpdateFOV(w, gmap, player, 5) // must not panic
}
