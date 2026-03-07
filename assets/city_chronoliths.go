package assets

// ChronolithsNPCs lists the named NPCs of Anchorpoint.
var ChronolithsNPCs = []NPCDef{
	{
		Glyph: "🦾",
		Name:  "Construct 4471",
		Kind:  0, // NPCKindDialogue
		Lines: []string{
			"I was built for a war I cannot remember. Retirement suits me. Mostly.",
			"Time here is reliable. I find this suspicious.",
			"My warranty expired 300 years ago. I am still functioning. The warranty was pessimistic.",
		},
	},
	{
		Glyph: "🙏",
		Name:  "Chrono Keeper Essa",
		Kind:  1, // NPCKindHealer
		Lines: []string{
			"The Chronolith keeps us steady. Let me keep you steady too.",
			"Hold still — I need to synchronise your biology to local time.",
			"You look like someone who's been in two places at once. That's bad for the joints.",
		},
	},
	{
		Glyph: "🛍️",
		Name:  "Salvager Korr",
		Kind:  2, // NPCKindShop
		Lines: []string{
			"Genuine temporal salvage. Everything here was, is, or will be useful.",
			"This sword is from a future that didn't happen. Still sharp, though.",
			"No returns. Literally — the timeline doesn't allow it.",
		},
	},
	{
		Glyph: "👴",
		Name:  "Watcher Brin",
		Kind:  0, // NPCKindDialogue
		Lines: []string{
			"I've watched the badlands for 40 years. Same frozen moments. Same looping valleys. It never gets less unsettling.",
			"The Constructs wander in sometimes. Some remember the war. Some remember peace. Some remember both.",
			"If you hear marching, don't investigate. It's either 300 years ago or 300 years from now, and neither is your problem.",
		},
	},
	{
		Glyph: "👩",
		Name:  "Settler Mira",
		Kind:  0, // NPCKindDialogue
		Lines: []string{
			"We came here because time is reliable. That's worth more than gold in the Chronoliths.",
			"The children play near the amber zones. They think frozen soldiers are funny. Children are terrifying.",
			"My grandmother was born after me. We don't talk about it at family dinners.",
		},
	},
	{
		Glyph: "🐕",
		Name:  "Temporal Hound",
		Kind:  3, // NPCKindAnimal
		Lines: []string{
			"The dog fetches a stick that won't be thrown for another hour. It is very patient.",
			"The hound sits at your feet, then sits at your feet again. You count two dogs. Then one. Time is flexible.",
			"It barks at something that isn't there yet. It will be there soon.",
		},
	},
}

// ChronolithsShopCatalogue defines the items Salvager Korr sells.
var ChronolithsShopCatalogue = []ShopEntry{
	// Consumables
	{Glyph: "🧪", Name: "Hyperflask", Price: 20, IsConsumable: true},
	{Glyph: "🧲", Name: "Resonance Coil", Price: 22, IsConsumable: true},
	{Glyph: "📜", Name: "Memory Scroll", Price: 18, IsConsumable: true},
	{Glyph: "💫", Name: "Prismatic Ward", Price: 28, IsConsumable: true},
	// Equipment
	{Glyph: "🪖", Name: "Chrono Helm", Price: 55, IsConsumable: false, BonusATK: 1, Slot: "head"},
	{Glyph: "🧥", Name: "Temporal Plate", Price: 60, IsConsumable: false, BonusDEF: 2, Slot: "body"},
	{Glyph: "⚔️", Name: "Timeline Edge", Price: 70, IsConsumable: false, BonusATK: 2, Slot: "onehand"},
	{Glyph: "🔋", Name: "Epoch Cell", Price: 50, IsConsumable: false, BonusDEF: 1, Slot: "offhand"},
}

// ChronolithsCityInscriptions is the pool of inscription texts placed in Anchorpoint.
var ChronolithsCityInscriptions = []string{
	"Welcome to Anchorpoint. Current time: now. This is not guaranteed outside city limits.",
	"CHRONO-HAZARD ADVISORY: Do not leave the Chronolith's radius without a temporal compass. We will not retrieve you from last Thursday.",
	"Construct Memorial: To those who fought in Timeline Seven. Or will fight. Tense pending.",
	"The Chronolith has stood since before the concept of 'standing' was temporally fixed.",
	"Salvager's Guild: If we don't have it, it hasn't existed yet. Check back yesterday.",
}

// ChronolithsNPCSchedules maps NPC name to daily schedule for Anchorpoint NPCs.
var ChronolithsNPCSchedules = map[string]NPCScheduleDef{
	// Commuter: sleeps in workshop area (west), wanders central area during day
	"Construct 4471": {
		MoveInterval: 18,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 12, StandY: 8},                                       // Night: workshop home
			{StartTick: 1500, Behavior: 2, Waypoints: [][2]int{{18, 8}, {24, 12}, {35, 20}}},          // Morning: path to central
			{StartTick: 2500, Behavior: 1, BoundsX1: 30, BoundsY1: 22, BoundsX2: 65, BoundsY2: 33},   // Day: wander market plaza
			{StartTick: 5000, Behavior: 2, Waypoints: [][2]int{{35, 20}, {24, 12}, {18, 8}, {12, 8}}}, // Evening: path home
		},
	},
	// Homebody: stays near Chronolith area, wanders there
	"Chrono Keeper Essa": {
		MoveInterval: 16,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 44, StandY: 14},                                    // Night: near Chronolith
			{StartTick: 1500, Behavior: 1, BoundsX1: 40, BoundsY1: 12, BoundsX2: 56, BoundsY2: 18}, // Morning: wander Chronolith area
			{StartTick: 2500, Behavior: 1, BoundsX1: 40, BoundsY1: 12, BoundsX2: 56, BoundsY2: 18}, // Day: wander Chronolith area
			{StartTick: 5000, Behavior: 1, BoundsX1: 40, BoundsY1: 12, BoundsX2: 56, BoundsY2: 18}, // Evening: wander Chronolith area
		},
	},
	// Commuter: home south, shop during day
	"Salvager Korr": {
		MoveInterval: 15,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 42, StandY: 43},                                    // Night: home south
			{StartTick: 1500, Behavior: 2, Waypoints: [][2]int{{42, 38}, {48, 38}, {52, 42}}},       // Morning: path to shop
			{StartTick: 2500, Behavior: 1, BoundsX1: 48, BoundsY1: 38, BoundsX2: 56, BoundsY2: 44}, // Day: wander shop area
			{StartTick: 5000, Behavior: 2, Waypoints: [][2]int{{48, 38}, {42, 38}, {42, 43}}},       // Evening: path home
		},
	},
	// Guard: observation post east, patrols eastern edge
	"Watcher Brin": {
		MoveInterval: 20,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 85, StandY: 8},                                    // Night: observation post
			{StartTick: 1500, Behavior: 2, Waypoints: [][2]int{{85, 14}, {85, 20}}},                // Morning: path to patrol start
			{StartTick: 2500, Behavior: 1, BoundsX1: 78, BoundsY1: 6, BoundsX2: 90, BoundsY2: 20}, // Day: patrol east quarter
			{StartTick: 5000, Behavior: 2, Waypoints: [][2]int{{85, 14}, {85, 8}}},                 // Evening: return to post
		},
	},
	// Wanderer: home east, paths to market, wanders
	"Settler Mira": {
		MoveInterval: 16,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 72, StandY: 10},                                        // Night: home east
			{StartTick: 1500, Behavior: 2, Waypoints: [][2]int{{72, 15}, {65, 20}, {55, 25}}},           // Morning: path to market
			{StartTick: 2500, Behavior: 1, BoundsX1: 30, BoundsY1: 22, BoundsX2: 65, BoundsY2: 33},     // Day: wander market plaza
			{StartTick: 5000, Behavior: 2, Waypoints: [][2]int{{55, 25}, {65, 20}, {72, 15}, {72, 10}}}, // Evening: path home
		},
	},
	// Wanderer (fast): roams large area
	"Temporal Hound": {
		MoveInterval: 10,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 1, BoundsX1: 30, BoundsY1: 20, BoundsX2: 65, BoundsY2: 33}, // Always wander plaza
		},
	},
}
