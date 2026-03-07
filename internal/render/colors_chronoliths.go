package render

// TileTheme returns the tile set for any absolute floor number.
// Floors 0-10 use the Prismatic Spire themes; floors 100-110 use Chronoliths.
func TileTheme(floor int) FloorTiles {
	if floor >= 100 {
		df := floor - 100
		if df >= 0 && df < len(ChronolithsTileThemes) {
			return ChronolithsTileThemes[df]
		}
		return ChronolithsTileThemes[1]
	}
	if floor >= 0 && floor < len(TileThemes) {
		return TileThemes[floor]
	}
	return TileThemes[1]
}

// ChronolithsTileThemes maps floor number to tile set for the Temporal Ruins.
var ChronolithsTileThemes = [11]FloorTiles{
	{ // Floor 0 — Anchorpoint: similar to Emberveil (brick/cobblestone)
		Wall:     "🧱",
		Floor:    "🟫",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{ // Floor 1 — Amber Antechamber: hourglasses, amber stone
		Wall:     "⏳",
		Floor:    "🟧",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{ // Floor 2 — Frozen Barracks: hourglasses, amber stone
		Wall:     "⏳",
		Floor:    "🟧",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{ // Floor 3 — The Repeating Hall: hourglasses, amber stone
		Wall:     "⏳",
		Floor:    "🟧",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{ // Floor 4 — Temporal Breach: clocks, golden energy
		Wall:     "⏰",
		Floor:    "🟨",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{ // Floor 5 — Clockwork Sanctuary: clocks, golden energy
		Wall:     "⏰",
		Floor:    "🟨",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{ // Floor 6 — The Paradox Wing: clocks, golden energy
		Wall:     "⏰",
		Floor:    "🟨",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{ // Floor 7 — Timeline Scar: crystallized time, deep amber
		Wall:     "🔶",
		Floor:    "🟠",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{ // Floor 8 — War Room Seven: crystallized time, deep amber
		Wall:     "🔶",
		Floor:    "🟠",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{ // Floor 9 — The Convergence: crystallized time, deep amber
		Wall:     "🔶",
		Floor:    "🟠",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{ // Floor 10 — The Eternal Moment: infinity, pure temporal energy
		Wall:     "♾️",
		Floor:    "🟡",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
}
