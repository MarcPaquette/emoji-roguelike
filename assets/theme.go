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
	GlyphSporeDraught  = "🍵" // floor 2+ — bioluminescent fungi brew
	GlyphResonanceCoil = "🧲" // floor 3+ — resonance engine tech
	GlyphPrismaticWard = "💫" // floor 5+ — prismatic defense field
	GlyphVoidEssence   = "🌌" // floor 6+ — dimensional void extract
	GlyphStairsDown   = "🔽"
	GlyphStairsUp     = "🔼"
	GlyphDoor         = "🚪"

	// Equipment — head
	GlyphCrystalHelm      = "🪖"
	GlyphVoidCrown        = "🎩"
	GlyphResonanceCowl    = "🪬"
	// Equipment — body
	GlyphFrostWeave       = "🧥"
	GlyphPrismaticPlate   = "🪤"
	GlyphCalcifiedCarapace = "🥼"
	// Equipment — feet
	GlyphFluxTreads       = "👟"
	GlyphForgeBoots       = "🥾"
	GlyphMembraneWalkers  = "🩴"
	// Equipment — one-hand weapons
	GlyphShardBlade       = "⚔️"
	GlyphTendrilWhip      = "🔱"
	GlyphEchoCutter       = "🪃"
	// Equipment — two-hand weapons
	GlyphResonanceMaul    = "🪚"
	GlyphAbyssalCleaver   = "🗡️"
	// Equipment — off-hand
	GlyphPhaseMirror      = "🪩"
	GlyphPowerCell        = "🔋"

	// New consumables
	GlyphNanoSyringe    = "💉" // floor 5+ — nano-medicine
	GlyphResonanceBurst = "🧨" // floor 3+ — overcharge
	GlyphPhaseRod       = "🪄" // floor 6+ — prismatic defense
	GlyphApexCore       = "🫀" // floor 8+ — permanent HP upgrade

	// Floors 6-10 enemies
	GlyphToxinSpore      = "🦠"
	GlyphTideWraith      = "🐙"
	GlyphOssifiedScholar = "🦴"
	GlyphArchiveWarden   = "🗝️"
	GlyphCinderWraith    = "🔥"
	GlyphForgeGolem      = "🧱"
	GlyphDreamStalker    = "🌙"
	GlyphPsychicEcho     = "🧿"
	GlyphCrystalRevenant = "🦂"
	GlyphUnmaker         = "☄️"

	// Floor elites — one unique boss-tier enemy per floor
	GlyphShardmind       = "💠" // Floor 1 elite
	GlyphSporeTyrant     = "🍄" // Floor 2 elite
	GlyphGearRevenant    = "⚙️" // Floor 3 elite
	GlyphPrismSpecter    = "✨" // Floor 4 elite
	GlyphTendrilOvermind = "🌿" // Floor 5 elite
	GlyphMembraneHorror  = "📄" // Floor 6 elite
	GlyphPetrifiedScholar = "📚" // Floor 7 elite
	GlyphMagmaRevenant   = "🌋" // Floor 8 elite
	GlyphSomnivore       = "💭" // Floor 9 elite
	GlyphPrismaticHorror = "🌟" // Floor 10 elite
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
var FloorNames = [11]string{
	"",
	"Crystalline Labs",
	"Bioluminescent Warrens",
	"Resonance Engine",
	"Fractured Observatory",
	"Apex Nexus",
	"Membrane of Echoes",
	"The Calcified Archive",
	"Abyssal Foundry",
	"The Dreaming Cortex",
	"The Prismatic Heart",
}

// BossGlyphs maps floor number to the glyph of the floor boss (empty if no boss).
// Victory on a boss floor requires killing this glyph.
var BossGlyphs = [11]string{
	"", "", "", "", "", "",
	"", "", "", "",
	GlyphUnmaker, // floor 10
}

// LoreOpening is shown when the game begins.
const LoreOpening = `The Prismatic Spire — an ancient research station piercing
the membrane between dimensions. Science and sorcery fused
into something neither discipline can fully explain.
You climb to reach the Prismatic Heart at the summit.
Press any key to begin...`

// Enemy tables per floor.
var EnemyTables = [11][]generate.EnemySpawnEntry{
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
	{ // Floor 6: Membrane of Echoes
		{Glyph: GlyphToxinSpore, Name: "Toxin Spore", ThreatCost: 4, Attack: 6, Defense: 1, MaxHP: 14, SightRange: 6,
			SpecialKind: 1, SpecialChance: 40, SpecialMag: 2, SpecialDur: 3},
		{Glyph: GlyphTideWraith, Name: "Tide Wraith", ThreatCost: 4, Attack: 8, Defense: 0, MaxHP: 10, SightRange: 8},
	},
	{ // Floor 7: The Calcified Archive
		{Glyph: GlyphOssifiedScholar, Name: "Ossified Scholar", ThreatCost: 5, Attack: 6, Defense: 4, MaxHP: 20, SightRange: 7,
			SpecialKind: 2, SpecialChance: 35, SpecialMag: 2, SpecialDur: 4},
		{Glyph: GlyphArchiveWarden, Name: "Archive Warden", ThreatCost: 5, Attack: 10, Defense: 3, MaxHP: 16, SightRange: 8},
	},
	{ // Floor 8: Abyssal Foundry
		{Glyph: GlyphCinderWraith, Name: "Cinder Wraith", ThreatCost: 6, Attack: 9, Defense: 1, MaxHP: 18, SightRange: 7,
			SpecialKind: 1, SpecialChance: 45, SpecialMag: 3, SpecialDur: 3},
		{Glyph: GlyphForgeGolem, Name: "Forge Golem", ThreatCost: 8, Attack: 8, Defense: 7, MaxHP: 34, SightRange: 5,
			SpecialKind: 3, SpecialChance: 50, SpecialMag: 5, SpecialDur: 0},
	},
	{ // Floor 9: The Dreaming Cortex
		{Glyph: GlyphDreamStalker, Name: "Dream Stalker", ThreatCost: 7, Attack: 11, Defense: 2, MaxHP: 24, SightRange: 9,
			SpecialKind: 2, SpecialChance: 40, SpecialMag: 3, SpecialDur: 5},
		{Glyph: GlyphPsychicEcho, Name: "Psychic Echo", ThreatCost: 6, Attack: 9, Defense: 2, MaxHP: 20, SightRange: 8,
			SpecialKind: 1, SpecialChance: 50, SpecialMag: 2, SpecialDur: 4},
	},
	{ // Floor 10: The Prismatic Heart
		{Glyph: GlyphCrystalRevenant, Name: "Crystal Revenant", ThreatCost: 8, Attack: 12, Defense: 5, MaxHP: 28, SightRange: 8,
			SpecialKind: 3, SpecialChance: 40, SpecialMag: 5, SpecialDur: 0},
		{Glyph: GlyphUnmaker, Name: "The Unmaker", ThreatCost: 22, Attack: 18, Defense: 8, MaxHP: 90, SightRange: 12,
			SpecialKind: 3, SpecialChance: 60, SpecialMag: 5, SpecialDur: 0},
	},
}

// ItemTables per floor.
// New items are introduced thematically as the player descends.
var ItemTables = [11][]generate.ItemSpawnEntry{
	{},
	{ // Floor 1 — Crystalline Labs
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
	},
	{ // Floor 2 — Bioluminescent Warrens: introduces Spore Draught
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
		{Glyph: GlyphSporeDraught, Name: "Spore Draught"},
	},
	{ // Floor 3 — Resonance Engine: introduces Resonance Coil
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphTesseract, Name: "Tesseract Cube"},
		{Glyph: GlyphSporeDraught, Name: "Spore Draught"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
	},
	{ // Floor 4 — Fractured Observatory
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphTesseract, Name: "Tesseract Cube"},
		{Glyph: GlyphSporeDraught, Name: "Spore Draught"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
	},
	{ // Floor 5 — Apex Nexus: introduces Prismatic Ward
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphTesseract, Name: "Tesseract Cube"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
		{Glyph: GlyphSporeDraught, Name: "Spore Draught"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
		{Glyph: GlyphPrismaticWard, Name: "Prismatic Ward"},
	},
	{ // Floor 6 — Membrane of Echoes: introduces Void Essence
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphTesseract, Name: "Tesseract Cube"},
		{Glyph: GlyphSporeDraught, Name: "Spore Draught"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
		{Glyph: GlyphPrismaticWard, Name: "Prismatic Ward"},
		{Glyph: GlyphVoidEssence, Name: "Void Essence"},
	},
	{ // Floor 7 — The Calcified Archive
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphTesseract, Name: "Tesseract Cube"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
		{Glyph: GlyphSporeDraught, Name: "Spore Draught"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
		{Glyph: GlyphPrismaticWard, Name: "Prismatic Ward"},
		{Glyph: GlyphVoidEssence, Name: "Void Essence"},
	},
	{ // Floor 8 — Abyssal Foundry
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphTesseract, Name: "Tesseract Cube"},
		{Glyph: GlyphSporeDraught, Name: "Spore Draught"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
		{Glyph: GlyphPrismaticWard, Name: "Prismatic Ward"},
		{Glyph: GlyphVoidEssence, Name: "Void Essence"},
	},
	{ // Floor 9 — The Dreaming Cortex
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphTesseract, Name: "Tesseract Cube"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
		{Glyph: GlyphSporeDraught, Name: "Spore Draught"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
		{Glyph: GlyphPrismaticWard, Name: "Prismatic Ward"},
		{Glyph: GlyphVoidEssence, Name: "Void Essence"},
	},
	{ // Floor 10 — The Prismatic Heart
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphTesseract, Name: "Tesseract Cube"},
		{Glyph: GlyphSporeDraught, Name: "Spore Draught"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
		{Glyph: GlyphPrismaticWard, Name: "Prismatic Ward"},
		{Glyph: GlyphVoidEssence, Name: "Void Essence"},
	},
}

// floorElites defines the unique elite enemy for each floor (index 0 unused).
// Elites are spawned once per floor outside the normal enemy budget.
var floorElites = [11]*generate.EnemySpawnEntry{
	nil, // floor 0 unused
	{ // Floor 1 — Crystalline Labs: Shardmind
		Glyph: GlyphShardmind, Name: "Shardmind",
		ThreatCost: 0, MaxHP: 20, Attack: 6, Defense: 4, SightRange: 6,
		SpecialKind: 5, SpecialChance: 30, SpecialMag: 2, SpecialDur: 4,
		Drops: []generate.DropEntry{{Glyph: GlyphHyperflask, Chance: 60}},
	},
	{ // Floor 2 — Bioluminescent Warrens: Spore Tyrant
		Glyph: GlyphSporeTyrant, Name: "Spore Tyrant",
		ThreatCost: 0, MaxHP: 22, Attack: 6, Defense: 2, SightRange: 8,
		SpecialKind: 4, SpecialChance: 25, SpecialMag: 0, SpecialDur: 2,
		Drops: []generate.DropEntry{{Glyph: GlyphSporeDraught, Chance: 70}},
	},
	{ // Floor 3 — Resonance Engine: Gear Revenant
		Glyph: GlyphGearRevenant, Name: "Gear Revenant",
		ThreatCost: 0, MaxHP: 26, Attack: 8, Defense: 4, SightRange: 7,
		SpecialKind: 2, SpecialChance: 40, SpecialMag: 3, SpecialDur: 4,
		Drops: []generate.DropEntry{{Glyph: GlyphResonanceCoil, Chance: 60}},
	},
	{ // Floor 4 — Fractured Observatory: Prism Specter
		Glyph: GlyphPrismSpecter, Name: "Prism Specter",
		ThreatCost: 0, MaxHP: 28, Attack: 9, Defense: 3, SightRange: 10,
		SpecialKind: 3, SpecialChance: 35, SpecialMag: 6, SpecialDur: 0,
		Drops: []generate.DropEntry{{Glyph: GlyphPrismShard, Chance: 65}},
	},
	{ // Floor 5 — Apex Nexus: Tendril Overmind
		Glyph: GlyphTendrilOvermind, Name: "Tendril Overmind",
		ThreatCost: 0, MaxHP: 35, Attack: 10, Defense: 3, SightRange: 8,
		SpecialKind: 1, SpecialChance: 45, SpecialMag: 3, SpecialDur: 4,
		Drops: []generate.DropEntry{{Glyph: GlyphPrismaticWard, Chance: 65}},
	},
	{ // Floor 6 — Membrane of Echoes: Membrane Horror
		Glyph: GlyphMembraneHorror, Name: "Membrane Horror",
		ThreatCost: 0, MaxHP: 32, Attack: 11, Defense: 2, SightRange: 9,
		SpecialKind: 5, SpecialChance: 40, SpecialMag: 3, SpecialDur: 3,
		Drops: []generate.DropEntry{{Glyph: GlyphVoidEssence, Chance: 60}},
	},
	{ // Floor 7 — The Calcified Archive: Petrified Scholar
		Glyph: GlyphPetrifiedScholar, Name: "Petrified Scholar",
		ThreatCost: 0, MaxHP: 38, Attack: 10, Defense: 6, SightRange: 7,
		SpecialKind: 4, SpecialChance: 30, SpecialMag: 0, SpecialDur: 3,
		Drops: []generate.DropEntry{{Glyph: GlyphNanoSyringe, Chance: 65}},
	},
	{ // Floor 8 — Abyssal Foundry: Magma Revenant
		Glyph: GlyphMagmaRevenant, Name: "Magma Revenant",
		ThreatCost: 0, MaxHP: 42, Attack: 12, Defense: 5, SightRange: 7,
		SpecialKind: 1, SpecialChance: 50, SpecialMag: 4, SpecialDur: 3,
		Drops: []generate.DropEntry{{Glyph: GlyphPhaseRod, Chance: 65}},
	},
	{ // Floor 9 — The Dreaming Cortex: Somnivore
		Glyph: GlyphSomnivore, Name: "Somnivore",
		ThreatCost: 0, MaxHP: 46, Attack: 13, Defense: 4, SightRange: 10,
		SpecialKind: 2, SpecialChance: 45, SpecialMag: 4, SpecialDur: 5,
		Drops: []generate.DropEntry{{Glyph: GlyphResonanceBurst, Chance: 65}},
	},
	{ // Floor 10 — The Prismatic Heart: Prismatic Horror
		Glyph: GlyphPrismaticHorror, Name: "Prismatic Horror",
		ThreatCost: 0, MaxHP: 55, Attack: 15, Defense: 7, SightRange: 10,
		SpecialKind: 3, SpecialChance: 50, SpecialMag: 6, SpecialDur: 0,
		Drops: []generate.DropEntry{{Glyph: GlyphApexCore, Chance: 80}},
	},
}

// FloorElite returns the elite enemy entry for the given floor, or nil if none.
func FloorElite(floor int) *generate.EnemySpawnEntry {
	if floor < 1 || floor > 10 {
		return nil
	}
	return floorElites[floor]
}
