package game

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/generate"
	"math"
	"math/rand"
)

const MaxFloors = 10

// levelConfig builds a generate.Config for the given floor number.
func levelConfig(floor int, rng *rand.Rand) *generate.Config {
	t := 0.0
	if MaxFloors > 1 {
		t = float64(floor-1) / float64(MaxFloors-1)
	}

	return &generate.Config{
		MapWidth:      lerpi(40, 90, t),
		MapHeight:     lerpi(20, 50, t),
		MinLeafSize:   8,
		MaxLeafSize:   lerpi(20, 10, t),
		SplitRatio:    0.5,
		MinRoomSize:   4,
		RoomPadding:   1,
		CorridorStyle: generate.CorridorLShaped,
		FloorNumber:   floor,
		EnemyBudget:   lerpi(5, 55, t),
		ItemCount:     lerpi(3, 8, t),
		EquipCount:    lerpi(1, 3, t),
		EnemyTable:       assets.EnemyTables[floor],
		ItemTable:        itemTableForFloor(floor),
		EquipTable:       assets.EquipTablesForFloor(floor),
		InscriptionTexts: assets.WallWritings[floor],
		InscriptionCount: 2 + rng.Intn(4), // 2–5 per floor
		Rand:             rng,
	}
}

// itemTableForFloor returns the consumable item table for a given floor,
// including any new consumables unlocked at that floor.
func itemTableForFloor(floor int) []generate.ItemSpawnEntry {
	base := assets.ItemTables[floor]
	var extra []generate.ItemSpawnEntry

	if floor >= 3 {
		extra = append(extra, generate.ItemSpawnEntry{Glyph: assets.GlyphResonanceBurst, Name: "Resonance Burst"})
	}
	if floor >= 5 {
		extra = append(extra, generate.ItemSpawnEntry{Glyph: assets.GlyphNanoSyringe, Name: "Nano-Syringe"})
	}
	if floor >= 6 {
		extra = append(extra, generate.ItemSpawnEntry{Glyph: assets.GlyphPhaseRod, Name: "Phase Rod"})
	}
	if floor >= 8 {
		extra = append(extra, generate.ItemSpawnEntry{Glyph: assets.GlyphApexCore, Name: "Apex Core"})
	}

	if len(extra) == 0 {
		return base
	}
	combined := make([]generate.ItemSpawnEntry, len(base)+len(extra))
	copy(combined, base)
	copy(combined[len(base):], extra)
	return combined
}

func lerpi(a, b int, t float64) int {
	return int(math.Round(float64(a) + t*float64(b-a)))
}
