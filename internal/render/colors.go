package render

// FloorTiles holds the emoji glyphs used to draw one floor's terrain.
// Emoji are rendered by the terminal with their own colors, so we use
// distinct glyphs for visible vs explored-but-dark states instead of
// trying to tint them with terminal FG color.
type FloorTiles struct {
	Wall     string // fully-visible wall tile
	Floor    string // fully-visible floor tile
	DimWall  string // explored but not currently visible wall
	DimFloor string // explored but not currently visible floor
}

// TileThemes maps floor number (0-indexed) to its tile set.
// Index 0 is Emberveil, the starting city.
var TileThemes = [11]FloorTiles{
	{
		// Floor 0 — Emberveil: cobblestone city
		Wall:     "🏠",
		Floor:    "🟫",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{
		// Floor 1 — Crystalline Labs: ice and frost
		Wall:     "🧊",
		Floor:    "❄️",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{
		// Floor 2 — Bioluminescent Warrens: fungal growth, living walls
		Wall:     "🍄",
		Floor:    "🌿",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{
		// Floor 3 — Resonance Engine: brass gears, golden sparks
		Wall:     "⚙️",
		Floor:    "✨",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{
		// Floor 4 — Fractured Observatory: stone and crystal lenses
		Wall:     "🪨",
		Floor:    "💠",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{
		// Floor 5 — Apex Nexus: skulls and void energy
		Wall:     "💀",
		Floor:    "🔴",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{
		// Floor 6 — Membrane of Echoes: bubble membrane walls, tidal fluid floor
		Wall:     "🫧",
		Floor:    "🌊",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{
		// Floor 7 — The Calcified Archive: ossified knowledge
		Wall:     "📚",
		Floor:    "📄",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{
		// Floor 8 — Abyssal Foundry: volcanic walls, alembic forge-vessel floor
		Wall:     "🌋",
		Floor:    "⚗️",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{
		// Floor 9 — The Dreaming Cortex: thought-bubble walls, shooting-star psionic floor
		Wall:     "💭",
		Floor:    "🌠",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
	{
		// Floor 10 — The Prismatic Heart: rainbow prismatic walls, glowing crystalline floor
		Wall:     "🌈",
		Floor:    "🌟",
		DimWall:  "🌑",
		DimFloor: "🔲",
	},
}
