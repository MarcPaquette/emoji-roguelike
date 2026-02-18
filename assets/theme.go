package assets

import "emoji-roguelike/internal/generate"

// Emoji constants used as entity glyphs.
const (
	GlyphPlayer       = "🧙"
	GlyphCrystalCrawl = "🦀"
	GlyphNeonSpecter  = "👻"
	GlyphPrismDrake   = "🐉"
	GlyphVoidTendril  = "🪱"
	GlyphThoughtLeech = "🧠"
	GlyphFractalGolem = "🗿"
	GlyphEntropyBloom = "🌀"
	GlyphApexWarden   = "🤖"
	GlyphHyperflask   = "🧪"
	GlyphPrismShard   = "💎"
	GlyphNullCloak    = "🫥"
	GlyphTesseract    = "📦"
	GlyphMemoryScroll = "📜"
	GlyphStairsDown   = "🔽"
	GlyphStairsUp     = "🔼"
	GlyphDoor         = "🚪"
)

// ClassDef defines a player class with stats and passive mechanics.
type ClassDef struct {
	ID             string
	Name           string
	Emoji          string
	Lore           string   // one-liner shown on the class selection screen
	MaxHP          int
	Attack         int
	Defense        int
	FOVRadius      int
	PassiveDesc    string   // human-readable passive description ("—" for none)
	KillRestoreHP  int      // >0: restore N HP on each kill
	StartInvisible int      // >0: apply EffectInvisible for N turns on floor 1
	StartRevealMap bool     // true: reveal entire floor on floor 1
	StartItems     []string // glyphs of items spawned near player on floor 1
}

// Classes is the ordered list of selectable player classes.
var Classes = []ClassDef{
	{
		ID:          "arcanist",
		Name:        "Wandering Arcanist",
		Emoji:       "🧙",
		Lore:        "A nomadic dimension-hopper who collected spells like others collect debt",
		MaxHP:       30,
		Attack:      5,
		Defense:     2,
		FOVRadius:   8,
		PassiveDesc: "—",
	},
	{
		ID:            "revenant",
		Name:          "Void Revenant",
		Emoji:         "💀",
		Lore:          "Death-kissed and not quite right about it. What kills others merely inconveniences you",
		MaxHP:         15,
		Attack:        12,
		Defense:       0,
		FOVRadius:     9,
		PassiveDesc:   "Each kill restores 3 HP",
		KillRestoreHP: 3,
	},
	{
		ID:          "construct",
		Name:        "Chrono Construct",
		Emoji:       "🦾",
		Lore:        "A war machine from Timeline Seven, now retired. Mostly",
		MaxHP:       60,
		Attack:      3,
		Defense:     8,
		FOVRadius:   5,
		PassiveDesc: "— (the stats are the passive)",
	},
	{
		ID:             "dancer",
		Name:           "Entropy Dancer",
		Emoji:          "🌀",
		Lore:           "You are not moving through chaos — you ARE chaos. Enemies can't locate you at first",
		MaxHP:          22,
		Attack:         9,
		Defense:        1,
		FOVRadius:      10,
		PassiveDesc:    "Invisible to enemies for 8 turns",
		StartInvisible: 8,
	},
	{
		ID:             "oracle",
		Name:           "Crystal Oracle",
		Emoji:          "🔮",
		Lore:           "You have seen the whole map. The whole map has seen you",
		MaxHP:          20,
		Attack:         3,
		Defense:        2,
		FOVRadius:      15,
		PassiveDesc:    "Entire floor revealed from the start",
		StartRevealMap: true,
	},
	{
		ID:          "symbiont",
		Name:        "Void Symbiont",
		Emoji:       "🧬",
		Lore:        "Somewhere inside you, a dimensional parasite purrs contentedly — and provides generously",
		MaxHP:       42,
		Attack:      6,
		Defense:     5,
		FOVRadius:   7,
		PassiveDesc: "Starts with Hyperflask, Prism Shard, and Null Cloak",
		StartItems:  []string{GlyphHyperflask, GlyphPrismShard, GlyphNullCloak},
	},
}

// FloorNames maps floor number (1-indexed) to its lore name.
var FloorNames = [6]string{
	"",
	"Crystalline Labs",
	"Bioluminescent Warrens",
	"Resonance Engine",
	"Fractured Observatory",
	"Apex Nexus",
}

// LoreOpening is shown when the game begins.
const LoreOpening = `The Prismatic Spire — an ancient research station piercing
the membrane between dimensions. Science and sorcery fused
into something neither discipline can fully explain.
You climb to reach the Apex Engine at the summit.
Press any key to begin...`

// Enemy tables per floor.
var EnemyTables = [6][]generate.EnemySpawnEntry{
	{}, // floor 0 unused
	{ // Floor 1: Crystalline Labs
		{Glyph: GlyphCrystalCrawl, Name: "Crystal Crawl", ThreatCost: 2, Attack: 3, Defense: 2, MaxHP: 8, SightRange: 5},
		{Glyph: GlyphNeonSpecter, Name: "Neon Specter", ThreatCost: 3, Attack: 4, Defense: 1, MaxHP: 6, SightRange: 7},
	},
	{ // Floor 2: Bioluminescent Warrens
		{Glyph: GlyphThoughtLeech, Name: "Thought Leech", ThreatCost: 3, Attack: 4, Defense: 1, MaxHP: 10, SightRange: 8},
		{Glyph: GlyphNeonSpecter, Name: "Neon Specter", ThreatCost: 3, Attack: 4, Defense: 1, MaxHP: 6, SightRange: 7},
		{Glyph: GlyphPrismDrake, Name: "Prism Drake", ThreatCost: 5, Attack: 6, Defense: 3, MaxHP: 14, SightRange: 6},
	},
	{ // Floor 3: Resonance Engine
		{Glyph: GlyphPrismDrake, Name: "Prism Drake", ThreatCost: 5, Attack: 6, Defense: 3, MaxHP: 14, SightRange: 6},
		{Glyph: GlyphVoidTendril, Name: "Void Tendril", ThreatCost: 4, Attack: 7, Defense: 0, MaxHP: 12, SightRange: 4},
		{Glyph: GlyphFractalGolem, Name: "Fractal Golem", ThreatCost: 6, Attack: 5, Defense: 5, MaxHP: 20, SightRange: 5},
	},
	{ // Floor 4: Fractured Observatory
		{Glyph: GlyphEntropyBloom, Name: "Entropy Bloom", ThreatCost: 7, Attack: 8, Defense: 2, MaxHP: 18, SightRange: 9},
		{Glyph: GlyphFractalGolem, Name: "Fractal Golem", ThreatCost: 6, Attack: 5, Defense: 5, MaxHP: 20, SightRange: 5},
		{Glyph: GlyphThoughtLeech, Name: "Thought Leech", ThreatCost: 3, Attack: 4, Defense: 1, MaxHP: 10, SightRange: 8},
	},
	{ // Floor 5: Apex Nexus
		{Glyph: GlyphEntropyBloom, Name: "Entropy Bloom", ThreatCost: 7, Attack: 8, Defense: 2, MaxHP: 18, SightRange: 9},
		{Glyph: GlyphFractalGolem, Name: "Fractal Golem", ThreatCost: 6, Attack: 5, Defense: 5, MaxHP: 20, SightRange: 5},
		{Glyph: GlyphThoughtLeech, Name: "Thought Leech", ThreatCost: 3, Attack: 4, Defense: 1, MaxHP: 10, SightRange: 8},
		{Glyph: GlyphVoidTendril, Name: "Void Tendril", ThreatCost: 4, Attack: 7, Defense: 0, MaxHP: 12, SightRange: 4},
		{Glyph: GlyphApexWarden, Name: "Apex Warden", ThreatCost: 15, Attack: 12, Defense: 6, MaxHP: 60, SightRange: 10},
	},
}

// ItemTables per floor.
var ItemTables = [6][]generate.ItemSpawnEntry{
	{},
	{
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
	},
	{
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
	},
	{
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphTesseract, Name: "Tesseract Cube"},
	},
	{
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphTesseract, Name: "Tesseract Cube"},
	},
	{
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphTesseract, Name: "Tesseract Cube"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
	},
}
