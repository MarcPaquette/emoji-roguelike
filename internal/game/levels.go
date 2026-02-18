package game

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/generate"
	"math"
	"math/rand"
)

const MaxFloors = 5

// levelConfig builds a generate.Config for the given floor number.
func levelConfig(floor int, rng *rand.Rand) *generate.Config {
	t := 0.0
	if MaxFloors > 1 {
		t = float64(floor-1) / float64(MaxFloors-1)
	}

	return &generate.Config{
		MapWidth:    lerpi(40, 80, t),
		MapHeight:   lerpi(20, 45, t),
		MinLeafSize: 8,
		MaxLeafSize: lerpi(20, 12, t),
		SplitRatio:  0.5,
		MinRoomSize: 4,
		RoomPadding: 1,
		CorridorStyle: generate.CorridorLShaped,
		FloorNumber: floor,
		EnemyBudget: lerpi(5, 40, t),
		ItemCount:   lerpi(3, 6, t),
		EnemyTable:  assets.EnemyTables[floor],
		ItemTable:   assets.ItemTables[floor],
		Rand:        rng,
	}
}

func lerpi(a, b int, t float64) int {
	return int(math.Round(float64(a) + t*float64(b-a)))
}
