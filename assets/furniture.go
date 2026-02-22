package assets

import "emoji-roguelike/internal/generate"

// FloorFurnitureDef holds the common and rare furniture tables for one floor.
type FloorFurnitureDef struct {
	Common []generate.FurnitureSpawnEntry
	Rare   []generate.FurnitureSpawnEntry
}

// FurnitureByFloor maps floor number (1-indexed, index 0 unused) to its furniture tables.
var FurnitureByFloor = [11]FloorFurnitureDef{
	{}, // floor 0 unused
	{ // Floor 1 — Crystalline Labs
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "🔬", Name: "Microscope", Description: "A lens-array for studying crystalline microstructures. Dust coats the eyepiece."},
			{Glyph: "🌡️", Name: "Thermometer", Description: "Records sub-zero temperatures at the lab's core. The mercury is frozen solid."},
			{Glyph: "🧫", Name: "Petri Dish", Description: "A culture of crystalline spores. They pulse faintly in the cold."},
			{Glyph: "🖥️", Name: "Terminal", Description: "A cracked workstation. The screen flickers: EXPERIMENT 7 — CONTAINED."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "🫙", Name: "Sample Jar", Description: "A sealed jar of crystalline growth serum. Warmth floods through you as it dissolves.", BonusMaxHP: 8},
			{Glyph: "⚡", Name: "Energy Cell", Description: "A charged capacitor from a prototype weapon. Absorbing its discharge sharpens your reflexes.", BonusATK: 1},
		},
	},
	{ // Floor 2 — Bioluminescent Warrens
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "🪸", Name: "Coral Formation", Description: "A branching coral that pulses with cold blue light. It hums at a frequency you feel in your teeth."},
			{Glyph: "🌺", Name: "Luminous Bloom", Description: "A flower whose petals emit soft violet bioluminescence. It smells of deep ocean."},
			{Glyph: "🌱", Name: "Spore Cluster", Description: "Dormant spores clinging to the tunnel wall. They sense your warmth and quiver."},
			{Glyph: "🪴", Name: "Glowing Plant", Description: "A fibrous plant whose roots glow green through the soil. Mycorrhizal tendrils reach everywhere."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "🌻", Name: "Sun-Bloom", Description: "A bioluminescent flower that converts ambient radiation into vitality. You feel it knit your cells together.", BonusMaxHP: 10},
			{Glyph: "🌾", Name: "Spirit Grass", Description: "A stand of resonant grass whose vibrations harden the body. Your skin feels denser.", BonusDEF: 1},
		},
	},
	{ // Floor 3 — Resonance Engine
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "🔩", Name: "Resonance Bolt", Description: "A bolt humming at 440Hz. It vibrates loose in your grip — no socket fits it."},
			{Glyph: "🪝", Name: "Frequency Hook", Description: "A hook mounted at a node in the resonance field. Things left here oscillate forever."},
			{Glyph: "🛠️", Name: "Harmonic Tool", Description: "A wrench-and-tuning-fork hybrid. The engravings read: CALIBRATE BEFORE USE."},
			{Glyph: "📡", Name: "Signal Dish", Description: "A small dish pointed at the engine core. It receives only static — and something beneath the static."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "🔧", Name: "Tuning Wrench", Description: "A precision wrench attuned to your body's natural frequency. Your joints lock into perfect alignment.", BonusDEF: 1},
			{Glyph: "🗺️", Name: "Resonance Map", Description: "A map of the engine's harmonic weak-points. Studying it reveals exactly where to strike.", BonusATK: 1},
		},
	},
	{ // Floor 4 — Fractured Observatory
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "🔭", Name: "Fractured Telescope", Description: "A telescope with a cracked lens. It still focuses — on something that isn't there."},
			{Glyph: "🪟", Name: "Observatory Window", Description: "A porthole looking out onto fractured space. The stars outside are wrong."},
			{Glyph: "🖼️", Name: "Star Chart", Description: "A chart of constellations that no longer exist. The coordinates lead somewhere unreachable."},
			{Glyph: "🗂️", Name: "Index Files", Description: "Files cataloguing every observable anomaly. The last entry reads: IT LOOKED BACK."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "💡", Name: "Insight Bulb", Description: "A globe of crystallised insight from a shattered mind. Absorbing it clarifies your strike patterns.", BonusATK: 1},
			{Glyph: "📐", Name: "Stellar Compass", Description: "A compass calibrated to fractured spacetime. Holding it aligns your body to a stronger configuration.", BonusMaxHP: 10},
		},
	},
	{ // Floor 5 — Apex Nexus
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "⚰️", Name: "Stasis Pod", Description: "A pod for preserving specimens indefinitely. The occupant was removed — recently."},
			{Glyph: "🕯️", Name: "Nexus Flame", Description: "A flame that burns without fuel. It casts shadows in directions that don't exist."},
			{Glyph: "🎭", Name: "Power Mask", Description: "A ceremonial mask worn by Nexus overseers. The expression is serene. The eyes are not."},
			{Glyph: "☠️", Name: "Trophy Skull", Description: "A skull mounted on a pedestal. Engraved beneath: THE LAST CHALLENGER."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "🛡️", Name: "Nexus Shield", Description: "A shield fragment from a decommissioned Apex Warden. Integrating its alloy thickens your defenses.", BonusDEF: 1},
			{Glyph: "🎲", Name: "Fate Die", Description: "A die that always lands on the result you need most. Today it grants you resilience.", BonusMaxHP: 8},
		},
	},
	{ // Floor 6 — Membrane of Echoes
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "🫗", Name: "Echo Vessel", Description: "A container that holds sound. Tilt it and you hear voices from floors you haven't reached."},
			{Glyph: "🧴", Name: "Membrane Serum", Description: "A vial of liquefied membrane tissue. It ripples like a living thing."},
			{Glyph: "🧳", Name: "Resonant Case", Description: "A case that hums when touched. Whatever it held has long since echoed away."},
			{Glyph: "🪣", Name: "Echo Bucket", Description: "A bucket filled with echo-fluid. Sounds made here repeat three seconds later — exactly."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "🏺", Name: "Ancient Urn", Description: "An urn sealed with membrane wax. Breaking the seal releases stored vital energy.", BonusMaxHP: 12},
			{Glyph: "🎪", Name: "Carnival Node", Description: "A resonance node shaped like a carnival tent. Its output frequency sharpens aggression.", BonusATK: 1},
		},
	},
	{ // Floor 7 — The Calcified Archive
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "📊", Name: "Data Chart", Description: "A chart tracking dimensional breach events over centuries. The trend is unmistakably upward."},
			{Glyph: "📋", Name: "Archive Clipboard", Description: "A clipboard with petrified parchment. The final entry is mid-sentence."},
			{Glyph: "📌", Name: "Index Pin", Description: "A pin marking a location on a petrified map. The label reads: DO NOT ENTER."},
			{Glyph: "🗃️", Name: "File Cabinet", Description: "A cabinet calcified shut. The label says: CASES NEVER CLOSED."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "🧮", Name: "Calcified Abacus", Description: "An abacus whose beads are made of compressed combat data. Running the numbers improves your precision.", BonusATK: 1},
			{Glyph: "⚖️", Name: "Archive Scales", Description: "Scales calibrated to measure dimensional weight. Standing in their field stabilises your biology.", BonusMaxHP: 10},
		},
	},
	{ // Floor 8 — Abyssal Foundry
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "🫕", Name: "Molten Crucible", Description: "A crucible of abyssal slag still glowing orange. Whatever was forged here didn't survive the process."},
			{Glyph: "🔑", Name: "Master Key", Description: "A key that opens nothing you've found. You pocket it anyway. Force of habit."},
			{Glyph: "🪙", Name: "Forge Coin", Description: "A coin minted from foundry slag. One face is a skull, the other is also a skull."},
			{Glyph: "🧯", Name: "Extinguisher", Description: "An emergency extinguisher. The gauge reads EMPTY. The foundry did not go well."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "🎯", Name: "Precision Target", Description: "A target used to calibrate forge-weapons. Studying the impact patterns hones your aim.", BonusATK: 1},
			{Glyph: "🏆", Name: "Forge Trophy", Description: "A trophy awarded to the foundry's most resilient construct. Absorbing its tempered alloy hardens you.", BonusDEF: 1},
		},
	},
	{ // Floor 9 — The Dreaming Cortex
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "🪞", Name: "Dream Mirror", Description: "A mirror showing a version of you that made different choices. It looks more tired."},
			{Glyph: "🛋️", Name: "Mind Couch", Description: "A couch where the Cortex's dreamers once lay. It's still warm."},
			{Glyph: "🎨", Name: "Dream Palette", Description: "A palette of colours that don't exist in waking reality. The brush moves on its own."},
			{Glyph: "🖌️", Name: "Memory Brush", Description: "A brush that paints recalled images onto reality. The canvas shows only darkness."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "🧸", Name: "Comfort Bear", Description: "A stuffed bear soaked in the Cortex's restorative dream-fluid. Holding it heals something deep.", BonusMaxHP: 12},
			{Glyph: "🪁", Name: "Focus Toy", Description: "A toy used to train the Cortex's dreamers in precision. Its rhythm sharpens your strikes.", BonusATK: 1},
		},
	},
	{ // Floor 10 — The Prismatic Heart
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "🪆", Name: "Nested Reality", Description: "A set of dolls, each containing a smaller version of the Prismatic Heart. They nest infinitely inward."},
			{Glyph: "🎠", Name: "Prismatic Carousel", Description: "A carousel of crystallised light-horses. They rotate without a motor — and without stopping."},
			{Glyph: "🎡", Name: "Reality Wheel", Description: "A ferris wheel where each gondola contains a different version of now. Most versions end badly."},
			{Glyph: "🎢", Name: "Dimension Coaster", Description: "A roller coaster that loops through adjacent dimensions. The track ends in a wall you can't see through."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "🧩", Name: "Missing Piece", Description: "The last fragment needed to complete the Heart's design. Fitting it into your body clarifies your purpose.", BonusATK: 1},
			{Glyph: "🎰", Name: "Probability Engine", Description: "A machine that samples all possible futures and delivers the best outcome. It delivers vitality.", BonusMaxHP: 15},
		},
	},
}
