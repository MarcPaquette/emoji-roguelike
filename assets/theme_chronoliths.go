package assets

import "emoji-roguelike/internal/generate"

// ── Chronoliths enemy glyph constants ──────────────────────────────────
const (
	GlyphTimeBeetle      = "⏱️"
	GlyphChronoSentry    = "🪖" // cross-region reuse OK (separate dungeon set)
	GlyphTemporalSpark   = "⚡"
	GlyphWarFragment     = "🦿"
	GlyphRustedWarden    = "🔧"
	GlyphParadoxEngine   = "⚙️"
	GlyphTimelineSoldier = "🤖"
	GlyphTheRecursion    = "⌛"

	// Chronoliths floor elites
	GlyphAmberKeeper     = "🏺"
	GlyphFrozenCaptain   = "🛡️"
	GlyphLoopGuardian    = "🔁"
	GlyphBreachWatcher   = "⏲️"
	GlyphClockPriest     = "🏛️"
	GlyphParadoxKnot     = "🪢"
	GlyphScarPhantom     = "🫥" // cross-region reuse OK
	GlyphGeneralSeven    = "🎖️"
	GlyphConvergenceNode = "🌀" // cross-region reuse OK
	GlyphEpochGuardian   = "⏳"
)

// ChronolithsFloorNames maps floor number to its lore name.
var ChronolithsFloorNames = [11]string{
	"Anchorpoint",
	"Amber Antechamber",
	"Frozen Barracks",
	"The Repeating Hall",
	"Temporal Breach",
	"Clockwork Sanctuary",
	"The Paradox Wing",
	"Timeline Scar",
	"War Room Seven",
	"The Convergence",
	"The Eternal Moment",
}

// ChronolithsBossGlyphs maps floor number to the boss glyph (empty if no boss).
var ChronolithsBossGlyphs = [11]string{
	"", "", "", "", "", "",
	"", "", "", "",
	GlyphTheRecursion, // floor 10
}

// ChronolithsEnemyTables holds enemies per floor for the Temporal Ruins.
var ChronolithsEnemyTables = [11][]generate.EnemySpawnEntry{
	{}, // floor 0 — Anchorpoint (city, no enemies)
	{ // Floor 1 — Amber Antechamber
		{Glyph: GlyphTimeBeetle, Name: "Time Beetle", ThreatCost: 2, Attack: 3, Defense: 2, MaxHP: 8, SightRange: 5},
		{Glyph: GlyphChronoSentry, Name: "Chrono Sentry", ThreatCost: 3, Attack: 5, Defense: 2, MaxHP: 10, SightRange: 6},
	},
	{ // Floor 2 — Frozen Barracks
		{Glyph: GlyphTimeBeetle, Name: "Time Beetle", ThreatCost: 2, Attack: 3, Defense: 2, MaxHP: 8, SightRange: 5},
		{Glyph: GlyphChronoSentry, Name: "Chrono Sentry", ThreatCost: 3, Attack: 5, Defense: 2, MaxHP: 10, SightRange: 6},
		{Glyph: GlyphTemporalSpark, Name: "Temporal Spark", ThreatCost: 4, Attack: 7, Defense: 0, MaxHP: 11, SightRange: 8,
			SpecialKind: 4, SpecialChance: 30, SpecialMag: 0, SpecialDur: 2},
	},
	{ // Floor 3 — The Repeating Hall
		{Glyph: GlyphTimeBeetle, Name: "Time Beetle", ThreatCost: 2, Attack: 3, Defense: 2, MaxHP: 8, SightRange: 5},
		{Glyph: GlyphChronoSentry, Name: "Chrono Sentry", ThreatCost: 3, Attack: 5, Defense: 2, MaxHP: 10, SightRange: 6},
		{Glyph: GlyphTemporalSpark, Name: "Temporal Spark", ThreatCost: 4, Attack: 7, Defense: 0, MaxHP: 11, SightRange: 8,
			SpecialKind: 4, SpecialChance: 30, SpecialMag: 0, SpecialDur: 2},
		{Glyph: GlyphWarFragment, Name: "War Fragment", ThreatCost: 5, Attack: 6, Defense: 4, MaxHP: 18, SightRange: 5,
			SpecialKind: 5, SpecialChance: 35, SpecialMag: 2, SpecialDur: 3},
	},
	{ // Floor 4 — Temporal Breach
		{Glyph: GlyphChronoSentry, Name: "Chrono Sentry", ThreatCost: 3, Attack: 5, Defense: 2, MaxHP: 10, SightRange: 6},
		{Glyph: GlyphTemporalSpark, Name: "Temporal Spark", ThreatCost: 4, Attack: 7, Defense: 0, MaxHP: 11, SightRange: 8,
			SpecialKind: 4, SpecialChance: 30, SpecialMag: 0, SpecialDur: 2},
		{Glyph: GlyphWarFragment, Name: "War Fragment", ThreatCost: 5, Attack: 6, Defense: 4, MaxHP: 18, SightRange: 5,
			SpecialKind: 5, SpecialChance: 35, SpecialMag: 2, SpecialDur: 3},
	},
	{ // Floor 5 — Clockwork Sanctuary
		{Glyph: GlyphTemporalSpark, Name: "Temporal Spark", ThreatCost: 4, Attack: 7, Defense: 0, MaxHP: 11, SightRange: 8,
			SpecialKind: 4, SpecialChance: 30, SpecialMag: 0, SpecialDur: 2},
		{Glyph: GlyphWarFragment, Name: "War Fragment", ThreatCost: 5, Attack: 6, Defense: 4, MaxHP: 18, SightRange: 5,
			SpecialKind: 5, SpecialChance: 35, SpecialMag: 2, SpecialDur: 3},
		{Glyph: GlyphRustedWarden, Name: "Rusted Warden", ThreatCost: 6, Attack: 5, Defense: 7, MaxHP: 24, SightRange: 4},
	},
	{ // Floor 6 — The Paradox Wing
		{Glyph: GlyphWarFragment, Name: "War Fragment", ThreatCost: 5, Attack: 6, Defense: 4, MaxHP: 18, SightRange: 5,
			SpecialKind: 5, SpecialChance: 35, SpecialMag: 2, SpecialDur: 3},
		{Glyph: GlyphRustedWarden, Name: "Rusted Warden", ThreatCost: 6, Attack: 5, Defense: 7, MaxHP: 24, SightRange: 4},
		{Glyph: GlyphParadoxEngine, Name: "Paradox Engine", ThreatCost: 7, Attack: 9, Defense: 3, MaxHP: 20, SightRange: 7,
			SpecialKind: 2, SpecialChance: 40, SpecialMag: 3, SpecialDur: 4},
	},
	{ // Floor 7 — Timeline Scar
		{Glyph: GlyphWarFragment, Name: "War Fragment", ThreatCost: 5, Attack: 6, Defense: 4, MaxHP: 18, SightRange: 5,
			SpecialKind: 5, SpecialChance: 35, SpecialMag: 2, SpecialDur: 3},
		{Glyph: GlyphRustedWarden, Name: "Rusted Warden", ThreatCost: 6, Attack: 5, Defense: 7, MaxHP: 24, SightRange: 4},
		{Glyph: GlyphParadoxEngine, Name: "Paradox Engine", ThreatCost: 7, Attack: 9, Defense: 3, MaxHP: 20, SightRange: 7,
			SpecialKind: 2, SpecialChance: 40, SpecialMag: 3, SpecialDur: 4},
	},
	{ // Floor 8 — War Room Seven
		{Glyph: GlyphRustedWarden, Name: "Rusted Warden", ThreatCost: 6, Attack: 5, Defense: 7, MaxHP: 24, SightRange: 4},
		{Glyph: GlyphParadoxEngine, Name: "Paradox Engine", ThreatCost: 7, Attack: 9, Defense: 3, MaxHP: 20, SightRange: 7,
			SpecialKind: 2, SpecialChance: 40, SpecialMag: 3, SpecialDur: 4},
		{Glyph: GlyphTimelineSoldier, Name: "Timeline Soldier", ThreatCost: 8, Attack: 11, Defense: 5, MaxHP: 28, SightRange: 8,
			SpecialKind: 1, SpecialChance: 45, SpecialMag: 3, SpecialDur: 3},
	},
	{ // Floor 9 — The Convergence
		{Glyph: GlyphParadoxEngine, Name: "Paradox Engine", ThreatCost: 7, Attack: 9, Defense: 3, MaxHP: 20, SightRange: 7,
			SpecialKind: 2, SpecialChance: 40, SpecialMag: 3, SpecialDur: 4},
		{Glyph: GlyphTimelineSoldier, Name: "Timeline Soldier", ThreatCost: 8, Attack: 11, Defense: 5, MaxHP: 28, SightRange: 8,
			SpecialKind: 1, SpecialChance: 45, SpecialMag: 3, SpecialDur: 3},
	},
	{ // Floor 10 — The Eternal Moment
		{Glyph: GlyphTimelineSoldier, Name: "Timeline Soldier", ThreatCost: 8, Attack: 11, Defense: 5, MaxHP: 28, SightRange: 8,
			SpecialKind: 1, SpecialChance: 45, SpecialMag: 3, SpecialDur: 3},
		{Glyph: GlyphTheRecursion, Name: "The Recursion", ThreatCost: 22, Attack: 17, Defense: 8, MaxHP: 85, SightRange: 12,
			SpecialKind: 4, SpecialChance: 50, SpecialMag: 0, SpecialDur: 3},
	},
}

// chronolithsFloorElites defines the unique elite enemy for each Chronoliths floor.
var chronolithsFloorElites = [11]*generate.EnemySpawnEntry{
	nil, // floor 0 — Anchorpoint
	{ // Floor 1 — Amber Keeper
		Glyph: GlyphAmberKeeper, Name: "Amber Keeper",
		ThreatCost: 0, MaxHP: 18, Attack: 5, Defense: 4, SightRange: 6,
		SpecialKind: 4, SpecialChance: 25, SpecialMag: 0, SpecialDur: 2,
		Drops: []generate.DropEntry{{Glyph: GlyphHyperflask, Chance: 60}},
	},
	{ // Floor 2 — Frozen Captain
		Glyph: GlyphFrozenCaptain, Name: "Frozen Captain",
		ThreatCost: 0, MaxHP: 22, Attack: 6, Defense: 4, SightRange: 7,
		SpecialKind: 5, SpecialChance: 30, SpecialMag: 2, SpecialDur: 3,
		Drops: []generate.DropEntry{{Glyph: GlyphResonanceCoil, Chance: 65}},
	},
	{ // Floor 3 — Loop Guardian
		Glyph: GlyphLoopGuardian, Name: "Loop Guardian",
		ThreatCost: 0, MaxHP: 24, Attack: 7, Defense: 3, SightRange: 8,
		SpecialKind: 2, SpecialChance: 35, SpecialMag: 2, SpecialDur: 4,
		Drops: []generate.DropEntry{{Glyph: GlyphMemoryScroll, Chance: 65}},
	},
	{ // Floor 4 — Breach Watcher
		Glyph: GlyphBreachWatcher, Name: "Breach Watcher",
		ThreatCost: 0, MaxHP: 28, Attack: 8, Defense: 3, SightRange: 9,
		SpecialKind: 4, SpecialChance: 35, SpecialMag: 0, SpecialDur: 3,
		Drops: []generate.DropEntry{{Glyph: GlyphPrismShard, Chance: 60}},
	},
	{ // Floor 5 — Clock Priest
		Glyph: GlyphClockPriest, Name: "Clock Priest",
		ThreatCost: 0, MaxHP: 32, Attack: 9, Defense: 5, SightRange: 7,
		SpecialKind: 1, SpecialChance: 40, SpecialMag: 3, SpecialDur: 3,
		Drops: []generate.DropEntry{{Glyph: GlyphPrismaticWard, Chance: 65}},
	},
	{ // Floor 6 — Paradox Knot
		Glyph: GlyphParadoxKnot, Name: "Paradox Knot",
		ThreatCost: 0, MaxHP: 36, Attack: 10, Defense: 3, SightRange: 8,
		SpecialKind: 2, SpecialChance: 40, SpecialMag: 3, SpecialDur: 4,
		Drops: []generate.DropEntry{{Glyph: GlyphVoidEssence, Chance: 65}},
	},
	{ // Floor 7 — Scar Phantom
		Glyph: GlyphScarPhantom, Name: "Scar Phantom",
		ThreatCost: 0, MaxHP: 40, Attack: 10, Defense: 5, SightRange: 8,
		SpecialKind: 5, SpecialChance: 35, SpecialMag: 3, SpecialDur: 0,
		Drops: []generate.DropEntry{{Glyph: GlyphNanoSyringe, Chance: 65}},
	},
	{ // Floor 8 — General Seven
		Glyph: GlyphGeneralSeven, Name: "General Seven",
		ThreatCost: 0, MaxHP: 44, Attack: 12, Defense: 6, SightRange: 8,
		SpecialKind: 3, SpecialChance: 45, SpecialMag: 4, SpecialDur: 0,
		Drops: []generate.DropEntry{{Glyph: GlyphPhaseRod, Chance: 65}},
	},
	{ // Floor 9 — Convergence Node
		Glyph: GlyphConvergenceNode, Name: "Convergence Node",
		ThreatCost: 0, MaxHP: 48, Attack: 13, Defense: 4, SightRange: 10,
		SpecialKind: 4, SpecialChance: 40, SpecialMag: 0, SpecialDur: 3,
		Drops: []generate.DropEntry{{Glyph: GlyphResonanceBurst, Chance: 65}},
	},
	{ // Floor 10 — Epoch Guardian
		Glyph: GlyphEpochGuardian, Name: "Epoch Guardian",
		ThreatCost: 0, MaxHP: 55, Attack: 15, Defense: 7, SightRange: 10,
		SpecialKind: 2, SpecialChance: 50, SpecialMag: 5, SpecialDur: 5,
		Drops: []generate.DropEntry{{Glyph: GlyphApexCore, Chance: 80}},
	},
}

// ChronolithsFloorElite returns the elite enemy for the given Chronoliths floor.
func ChronolithsFloorElite(floor int) *generate.EnemySpawnEntry {
	if floor < 1 || floor > 10 {
		return nil
	}
	return chronolithsFloorElites[floor]
}

// ChronolithsItemTables holds consumable items per Chronoliths floor.
var ChronolithsItemTables = [11][]generate.ItemSpawnEntry{
	{}, // floor 0 — Anchorpoint
	{ // Floor 1 — Amber Antechamber
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
	},
	{ // Floor 2 — Frozen Barracks
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
	},
	{ // Floor 3 — The Repeating Hall
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
	},
	{ // Floor 4 — Temporal Breach
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
	},
	{ // Floor 5 — Clockwork Sanctuary
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphPrismaticWard, Name: "Prismatic Ward"},
	},
	{ // Floor 6 — The Paradox Wing
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphPrismaticWard, Name: "Prismatic Ward"},
		{Glyph: GlyphVoidEssence, Name: "Void Essence"},
	},
	{ // Floor 7 — Timeline Scar
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphPrismaticWard, Name: "Prismatic Ward"},
		{Glyph: GlyphVoidEssence, Name: "Void Essence"},
		{Glyph: GlyphTesseract, Name: "Tesseract Cube"},
	},
	{ // Floor 8 — War Room Seven
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphPrismaticWard, Name: "Prismatic Ward"},
		{Glyph: GlyphVoidEssence, Name: "Void Essence"},
		{Glyph: GlyphTesseract, Name: "Tesseract Cube"},
		{Glyph: GlyphSporeDraught, Name: "Spore Draught"},
	},
	{ // Floor 9 — The Convergence
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphPrismaticWard, Name: "Prismatic Ward"},
		{Glyph: GlyphVoidEssence, Name: "Void Essence"},
		{Glyph: GlyphTesseract, Name: "Tesseract Cube"},
		{Glyph: GlyphSporeDraught, Name: "Spore Draught"},
	},
	{ // Floor 10 — The Eternal Moment
		{Glyph: GlyphHyperflask, Name: "Hyperflask"},
		{Glyph: GlyphMemoryScroll, Name: "Memory Scroll"},
		{Glyph: GlyphResonanceCoil, Name: "Resonance Coil"},
		{Glyph: GlyphPrismShard, Name: "Prism Shard"},
		{Glyph: GlyphNullCloak, Name: "Null Cloak"},
		{Glyph: GlyphPrismaticWard, Name: "Prismatic Ward"},
		{Glyph: GlyphVoidEssence, Name: "Void Essence"},
		{Glyph: GlyphTesseract, Name: "Tesseract Cube"},
		{Glyph: GlyphSporeDraught, Name: "Spore Draught"},
	},
}
