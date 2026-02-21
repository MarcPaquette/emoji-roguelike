package assets

import "emoji-roguelike/internal/generate"

// equipTemplates defines all 16 equipment item templates.
// Slot values match component.ItemSlot: 1=Head 2=Body 3=Feet 4=OneHand 5=TwoHand 6=OffHand
var equipTemplates = []generate.EquipSpawnEntry{
	// Head
	{Glyph: GlyphCrystalHelm, Name: "Crystal Helm", Slot: 1, BaseATK: 0, BaseDEF: 2, BaseMaxHP: 5, ATKScale: 0, DEFScale: 3, HPScale: 8, MinFloor: 1},
	{Glyph: GlyphVoidCrown, Name: "Void Crown", Slot: 1, BaseATK: 1, BaseDEF: 1, BaseMaxHP: 3, ATKScale: 2, DEFScale: 2, HPScale: 5, MinFloor: 4},
	{Glyph: GlyphResonanceCowl, Name: "Resonance Cowl", Slot: 1, BaseATK: 0, BaseDEF: 0, BaseMaxHP: 0, ATKScale: 1, DEFScale: 0, HPScale: 10, MinFloor: 7},
	// Body
	{Glyph: GlyphFrostWeave, Name: "Frost Weave", Slot: 2, BaseATK: 0, BaseDEF: 4, BaseMaxHP: 0, ATKScale: 0, DEFScale: 5, HPScale: 0, MinFloor: 1},
	{Glyph: GlyphPrismaticPlate, Name: "Prismatic Plate", Slot: 2, BaseATK: 0, BaseDEF: 3, BaseMaxHP: 5, ATKScale: 0, DEFScale: 4, HPScale: 8, MinFloor: 5},
	{Glyph: GlyphCalcifiedCarapace, Name: "Calcified Carapace", Slot: 2, BaseATK: 0, BaseDEF: 2, BaseMaxHP: 8, ATKScale: 0, DEFScale: 3, HPScale: 10, MinFloor: 7},
	// Feet
	{Glyph: GlyphFluxTreads, Name: "Flux Treads", Slot: 3, BaseATK: 0, BaseDEF: 1, BaseMaxHP: 3, ATKScale: 0, DEFScale: 2, HPScale: 4, MinFloor: 1},
	{Glyph: GlyphForgeBoots, Name: "Forge Boots", Slot: 3, BaseATK: 0, BaseDEF: 3, BaseMaxHP: 0, ATKScale: 0, DEFScale: 4, HPScale: 0, MinFloor: 6},
	{Glyph: GlyphMembraneWalkers, Name: "Membrane Walkers", Slot: 3, BaseATK: 1, BaseDEF: 0, BaseMaxHP: 5, ATKScale: 2, DEFScale: 0, HPScale: 6, MinFloor: 8},
	// One-hand weapons
	{Glyph: GlyphShardBlade, Name: "Shard Blade", Slot: 4, BaseATK: 4, BaseDEF: 0, BaseMaxHP: 0, ATKScale: 6, DEFScale: 0, HPScale: 0, MinFloor: 1},
	{Glyph: GlyphTendrilWhip, Name: "Tendril Whip", Slot: 4, BaseATK: 3, BaseDEF: 1, BaseMaxHP: 0, ATKScale: 5, DEFScale: 1, HPScale: 0, MinFloor: 3},
	{Glyph: GlyphEchoCutter, Name: "Echo Cutter", Slot: 4, BaseATK: 2, BaseDEF: 2, BaseMaxHP: 0, ATKScale: 4, DEFScale: 2, HPScale: 0, MinFloor: 5},
	// Two-hand weapons
	{Glyph: GlyphResonanceMaul, Name: "Resonance Maul", Slot: 5, BaseATK: 7, BaseDEF: 0, BaseMaxHP: 0, ATKScale: 9, DEFScale: 0, HPScale: 0, MinFloor: 4},
	{Glyph: GlyphAbyssalCleaver, Name: "Abyssal Cleaver", Slot: 5, BaseATK: 6, BaseDEF: 0, BaseMaxHP: -5, ATKScale: 9, DEFScale: 0, HPScale: 0, MinFloor: 7},
	// Off-hand
	{Glyph: GlyphPhaseMirror, Name: "Phase Mirror", Slot: 6, BaseATK: 0, BaseDEF: 3, BaseMaxHP: 0, ATKScale: 0, DEFScale: 5, HPScale: 0, MinFloor: 2},
	{Glyph: GlyphPowerCell, Name: "Power Cell", Slot: 6, BaseATK: 2, BaseDEF: 2, BaseMaxHP: 0, ATKScale: 3, DEFScale: 3, HPScale: 0, MinFloor: 5},
}

// EquipTablesForFloor returns all equipment templates available on the given floor.
func EquipTablesForFloor(floor int) []generate.EquipSpawnEntry {
	var out []generate.EquipSpawnEntry
	for _, e := range equipTemplates {
		if e.MinFloor <= floor {
			out = append(out, e)
		}
	}
	return out
}

// consumableNames maps glyph to human-readable name for all consumable items.
var consumableNames = map[string]string{
	GlyphHyperflask:     "Hyperflask",
	GlyphPrismShard:     "Prism Shard",
	GlyphNullCloak:      "Null Cloak",
	GlyphTesseract:      "Tesseract Cube",
	GlyphMemoryScroll:   "Memory Scroll",
	GlyphSporeDraught:   "Spore Draught",
	GlyphResonanceCoil:  "Resonance Coil",
	GlyphPrismaticWard:  "Prismatic Ward",
	GlyphVoidEssence:    "Void Essence",
	GlyphNanoSyringe:    "Nano-Syringe",
	GlyphResonanceBurst: "Resonance Burst",
	GlyphPhaseRod:       "Phase Rod",
	GlyphApexCore:       "Apex Core",
}

// ConsumableName returns the human-readable name for a consumable glyph.
func ConsumableName(glyph string) string {
	if name, ok := consumableNames[glyph]; ok {
		return name
	}
	return glyph // fallback
}
