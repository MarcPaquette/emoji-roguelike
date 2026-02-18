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
				t := gmap.At(nx, ny)
				if t.Walkable {
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
