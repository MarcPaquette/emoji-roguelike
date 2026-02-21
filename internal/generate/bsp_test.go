package generate

import (
	"emoji-roguelike/internal/gamemap"
	"math/rand"
	"testing"
)

func defaultTestConfig(seed int64) *Config {
	return &Config{
		MapWidth:    60,
		MapHeight:   30,
		MinLeafSize: 8,
		MaxLeafSize: 20,
		SplitRatio:  0.5,
		MinRoomSize: 4,
		RoomPadding: 1,
		CorridorStyle: CorridorLShaped,
		FloorNumber: 1,
		EnemyBudget: 10,
		ItemCount:   3,
		Rand:        rand.New(rand.NewSource(seed)),
	}
}

// TestGenerateAllRoomsConnected verifies that every floor tile is reachable
// from the first floor tile via BFS (flood-fill).
func TestGenerateAllRoomsConnected(t *testing.T) {
	for seed := int64(0); seed < 10; seed++ {
		cfg := defaultTestConfig(seed)
		gmap, _, _ := Generate(cfg)

		// Find the first floor tile.
		startX, startY := -1, -1
		for y := 0; y < gmap.Height && startY == -1; y++ {
			for x := 0; x < gmap.Width && startX == -1; x++ {
				if gmap.At(x, y).Kind == gamemap.TileFloor ||
					gmap.At(x, y).Kind == gamemap.TileStairsDown {
					startX, startY = x, y
				}
			}
		}
		if startX == -1 {
			t.Fatalf("seed=%d: no floor tiles found", seed)
		}

		// BFS from start.
		visited := make([][]bool, gmap.Height)
		for y := range visited {
			visited[y] = make([]bool, gmap.Width)
		}
		queue := [][2]int{{startX, startY}}
		visited[startY][startX] = true

		dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			for _, d := range dirs {
				nx, ny := cur[0]+d[0], cur[1]+d[1]
				if !gmap.InBounds(nx, ny) || visited[ny][nx] {
					continue
				}
				tile := gmap.At(nx, ny)
				// Traverse both walkable tiles and closed doors (rooms are
				// connected through doors even though doors are non-walkable).
				if tile.Walkable || tile.Kind == gamemap.TileDoor {
					visited[ny][nx] = true
					queue = append(queue, [2]int{nx, ny})
				}
			}
		}

		// Every walkable tile should have been visited.
		for y := 0; y < gmap.Height; y++ {
			for x := 0; x < gmap.Width; x++ {
				if gmap.At(x, y).Walkable && !visited[y][x] {
					t.Errorf("seed=%d: unreachable floor tile at (%d,%d)", seed, x, y)
				}
			}
		}
	}
}

// TestPlaceDoors verifies that PlaceDoors converts perimeter floor tiles to doors
// and leaves non-corridor perimeter tiles and room interior tiles unchanged.
func TestPlaceDoors(t *testing.T) {
	// Build a small 10×10 map with one room (3×3, at x1=3,y1=3 to x2=5,y2=5).
	// Carve a corridor tile at (4, 2) — one tile above the room's top edge.
	// All other perimeter positions remain wall.
	gmap := gamemap.New(10, 10)

	// Carve room interior.
	for y := 3; y <= 5; y++ {
		for x := 3; x <= 5; x++ {
			gmap.Set(x, y, gamemap.MakeFloor())
		}
	}
	gmap.Rooms = append(gmap.Rooms, gamemap.Rect{X1: 3, Y1: 3, X2: 5, Y2: 5})

	// Carve one corridor tile on the top perimeter.
	gmap.Set(4, 2, gamemap.MakeFloor())

	PlaceDoors(gmap)

	// The corridor-adjacent perimeter tile should become a door.
	if got := gmap.At(4, 2).Kind; got != gamemap.TileDoor {
		t.Errorf("expected TileDoor at (4,2), got kind %d", got)
	}
	// Door must be non-walkable and opaque.
	door := gmap.At(4, 2)
	if door.Walkable {
		t.Errorf("door at (4,2) should not be walkable")
	}
	if door.Transparent {
		t.Errorf("door at (4,2) should not be transparent")
	}

	// Other top-perimeter tiles without a corridor should remain wall.
	for _, x := range []int{3, 5} {
		if got := gmap.At(x, 2).Kind; got != gamemap.TileWall {
			t.Errorf("expected TileWall at (%d,2), got kind %d", x, got)
		}
	}

	// Room interior tiles must stay as floor.
	for y := 3; y <= 5; y++ {
		for x := 3; x <= 5; x++ {
			if got := gmap.At(x, y).Kind; got != gamemap.TileFloor {
				t.Errorf("expected TileFloor at (%d,%d), got kind %d", x, y, got)
			}
		}
	}
}

// TestGenerateRoomsDoNotOverlap verifies that no two rooms share interior tiles.
func TestGenerateRoomsDoNotOverlap(t *testing.T) {
	for seed := int64(0); seed < 10; seed++ {
		cfg := defaultTestConfig(seed)
		gmap, _, _ := Generate(cfg)

		rooms := gmap.Rooms
		for i := 0; i < len(rooms); i++ {
			for j := i + 1; j < len(rooms); j++ {
				if rooms[i].Intersects(rooms[j]) {
					t.Errorf("seed=%d: room %d %v overlaps room %d %v",
						seed, i, rooms[i], j, rooms[j])
				}
			}
		}
	}
}
