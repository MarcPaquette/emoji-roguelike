package assets

import "emoji-roguelike/internal/generate"

// ── Floor numbering scheme ─────────────────────────────────────────────
//
//   0       Emberveil (Spire city)
//   1–10    Prismatic Spire dungeon
//   100     Anchorpoint (Chronoliths city)
//   101–110 Temporal Ruins dungeon
//
// IsChronoliths returns true if the absolute floor number belongs to
// The Chronoliths dungeon set.
func IsChronoliths(floor int) bool { return floor >= 100 }

// DungeonFloor converts an absolute floor number to a dungeon-local
// index (0 = city, 1–10 = dungeon floors).
func DungeonFloor(floor int) int {
	if floor >= 100 {
		return floor - 100
	}
	return floor
}

// DungeonMaxFloor returns the highest valid floor in the dungeon
// containing the given floor.
func DungeonMaxFloor(floor int) int {
	if floor >= 100 {
		return 110
	}
	return 10
}

// DungeonCityFloor returns the city floor number for the dungeon
// containing the given floor.
func DungeonCityFloor(floor int) int {
	if floor >= 100 {
		return 100
	}
	return 0
}

// ── Data accessors ─────────────────────────────────────────────────────

// FloorName returns the lore name for any absolute floor.
func FloorName(floor int) string {
	df := DungeonFloor(floor)
	if IsChronoliths(floor) {
		if df >= 0 && df < len(ChronolithsFloorNames) {
			return ChronolithsFloorNames[df]
		}
		return "Unknown"
	}
	if df >= 0 && df < len(FloorNames) {
		return FloorNames[df]
	}
	return "Unknown"
}

// BossGlyph returns the boss glyph for a floor (empty if no boss).
func BossGlyph(floor int) string {
	df := DungeonFloor(floor)
	if IsChronoliths(floor) {
		if df >= 0 && df < len(ChronolithsBossGlyphs) {
			return ChronolithsBossGlyphs[df]
		}
		return ""
	}
	if df >= 0 && df < len(BossGlyphs) {
		return BossGlyphs[df]
	}
	return ""
}

// EnemyTable returns the enemy spawn table for a floor.
func EnemyTable(floor int) []generate.EnemySpawnEntry {
	df := DungeonFloor(floor)
	if IsChronoliths(floor) {
		if df >= 0 && df < len(ChronolithsEnemyTables) {
			return ChronolithsEnemyTables[df]
		}
		return nil
	}
	if df >= 0 && df < len(EnemyTables) {
		return EnemyTables[df]
	}
	return nil
}

// ItemTable returns the consumable item table for a floor.
func ItemTable(floor int) []generate.ItemSpawnEntry {
	df := DungeonFloor(floor)
	if IsChronoliths(floor) {
		if df >= 0 && df < len(ChronolithsItemTables) {
			return ChronolithsItemTables[df]
		}
		return nil
	}
	if df >= 0 && df < len(ItemTables) {
		return ItemTables[df]
	}
	return nil
}

// EliteEnemy returns the floor elite for any absolute floor.
func EliteEnemy(floor int) *generate.EnemySpawnEntry {
	if IsChronoliths(floor) {
		return ChronolithsFloorElite(DungeonFloor(floor))
	}
	return FloorElite(DungeonFloor(floor))
}

// FloorLoreSnippets returns the atmospheric lore snippets for a floor.
func FloorLoreSnippets(floor int) []string {
	df := DungeonFloor(floor)
	if IsChronoliths(floor) {
		if df >= 0 && df < len(ChronolithsFloorLore) {
			return ChronolithsFloorLore[df]
		}
		return nil
	}
	if df >= 0 && df < len(FloorLore) {
		return FloorLore[df]
	}
	return nil
}

// WallWritingsFor returns the inscription texts for a floor.
func WallWritingsFor(floor int) []string {
	df := DungeonFloor(floor)
	if IsChronoliths(floor) {
		if df >= 0 && df < len(ChronolithsWallWritings) {
			return ChronolithsWallWritings[df]
		}
		return nil
	}
	if df >= 0 && df < len(WallWritings) {
		return WallWritings[df]
	}
	return nil
}

// FurnitureFor returns the furniture tables for a floor.
func FurnitureFor(floor int) FloorFurnitureDef {
	df := DungeonFloor(floor)
	if IsChronoliths(floor) {
		if df >= 0 && df < len(ChronolithsFurnitureByFloor) {
			return ChronolithsFurnitureByFloor[df]
		}
		return FloorFurnitureDef{}
	}
	if df >= 0 && df < len(FurnitureByFloor) {
		return FurnitureByFloor[df]
	}
	return FloorFurnitureDef{}
}
