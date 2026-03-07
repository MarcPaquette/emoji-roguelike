package assets

import "emoji-roguelike/internal/generate"

// ChronolithsFurnitureByFloor maps floor number to furniture tables for the Temporal Ruins.
var ChronolithsFurnitureByFloor = [11]FloorFurnitureDef{
	{}, // floor 0 — Anchorpoint (city uses repeatable furniture)
	{ // Floor 1 — Amber Antechamber
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "⏰", Name: "Broken Clock", Description: "A clock whose hands spin in both directions. It is correct twice per eternity."},
			{Glyph: "🪙", Name: "Temporal Coin", Description: "A coin frozen mid-flip. Both faces show the same result. The result changes when you look away."},
			{Glyph: "🛡️", Name: "Rusted Shield", Description: "A shield from a battle that hasn't happened. The dents are from weapons that don't exist."},
			{Glyph: "🗂️", Name: "Deployment Roster", Description: "A list of Construct units assigned to this sector. The deployment date is listed as 'pending'."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "🏺", Name: "Amber Vial", Description: "A vial of crystallised temporal energy. Drinking it makes you more stubbornly present.", BonusMaxHP: 8},
			{Glyph: "⚔️", Name: "Temporal Shard", Description: "A blade fragment that exists in two moments. Its edge cuts where you were and where you will be.", BonusATK: 1},
		},
	},
	{ // Floor 2 — Frozen Barracks
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "🪖", Name: "Frozen Helmet", Description: "A helmet mid-dent from an impact that occurred 300 years ago. The impact has not concluded."},
			{Glyph: "📜", Name: "War Dispatch", Description: "Orders from Timeline Seven's high command. The ink changes language depending on when you read it."},
			{Glyph: "🛋️", Name: "Barracks Bunk", Description: "A cot with an indentation from a soldier who hasn't arrived yet. The sheets are warm."},
			{Glyph: "🗃️", Name: "Requisition Forms", Description: "Supply forms requesting ammunition for weapons not yet invented. Stamped APPROVED."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "🧮", Name: "Chrono-Abacus", Description: "An abacus that counts in temporal units. Sliding a bead aligns your internal clock. You feel sturdier.", BonusMaxHP: 8},
			{Glyph: "🗺️", Name: "Battle Map", Description: "A tactical map showing positions from three timelines simultaneously. Studying it reveals weak points.", BonusATK: 1},
		},
	},
	{ // Floor 3 — The Repeating Hall
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "🔁", Name: "Loop Marker", Description: "A stone engraved with a symbol that is also the stone. The recursion is architectural."},
			{Glyph: "🪞", Name: "Echo Mirror", Description: "A mirror reflecting the room as it was three seconds ago. You see yourself arriving."},
			{Glyph: "📋", Name: "Maintenance Log", Description: "LOOP DETECTED IN CORRIDOR B. FIX SCHEDULED FOR: [DATE FIELD LOOPS TO THIS ENTRY]."},
			{Glyph: "🕯️", Name: "Eternal Candle", Description: "A candle that has been burning for exactly as long as it will burn. The wax replenishes."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "📐", Name: "Temporal Compass", Description: "A compass calibrated to repeat beneficial moments. Holding it reinforces your defenses.", BonusDEF: 1},
			{Glyph: "💡", Name: "Recursion Lamp", Description: "A lamp that generates light from its own future brightness. Absorbing its glow extends your vitality.", BonusMaxHP: 9},
		},
	},
	{ // Floor 4 — Temporal Breach
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "🌡️", Name: "Chrono-Thermometer", Description: "Records the temperature at three different points in time simultaneously. All disagree."},
			{Glyph: "📊", Name: "Breach Data", Description: "Readings from the temporal breach. The numbers increase, decrease, and hold steady — all at once."},
			{Glyph: "🔩", Name: "Displaced Bolt", Description: "A bolt from a machine that hasn't been assembled yet. It hums with anticipatory tension."},
			{Glyph: "🖥️", Name: "Temporal Monitor", Description: "A screen displaying timestamps from three eras. The cursor blinks at a frequency that causes mild headaches."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "⚡", Name: "Breach Capacitor", Description: "A capacitor charged by the temporal breach itself. Its discharge sharpens your reflexes across timelines.", BonusATK: 1},
			{Glyph: "🫙", Name: "Bottled Moment", Description: "A jar containing a preserved moment of perfect health. Opening it floods you with its vitality.", BonusMaxHP: 10},
		},
	},
	{ // Floor 5 — Clockwork Sanctuary
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "⚙️", Name: "Sanctuary Gear", Description: "A gear from the sanctuary's mechanism. It turns without an axle. The mechanism it serves is conceptual."},
			{Glyph: "🔑", Name: "Temporal Key", Description: "A key to a lock that hasn't been installed. The keyhole is scheduled for next century."},
			{Glyph: "🪝", Name: "Pendulum Hook", Description: "A hook from which a pendulum once swung. The pendulum is still swinging. In another time."},
			{Glyph: "📌", Name: "Pinned Moment", Description: "A moment pinned to the wall like a butterfly specimen. The moment contains a clock striking noon."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "🔧", Name: "Calibration Tool", Description: "A precision tool attuned to clockwork rhythms. Handling it synchronises your body's timing.", BonusDEF: 1},
			{Glyph: "⏳", Name: "Amber Hourglass", Description: "An hourglass containing frozen time. Breaking it releases vitality that should have expired centuries ago.", BonusMaxHP: 10},
		},
	},
	{ // Floor 6 — The Paradox Wing
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "🎭", Name: "Paradox Mask", Description: "A mask depicting a face that is smiling and frowning. The expression depends on when you look."},
			{Glyph: "🖼️", Name: "Temporal Portrait", Description: "A portrait of someone who ages and de-ages as you watch. They seem to be having a perfectly normal day."},
			{Glyph: "📡", Name: "Paradox Antenna", Description: "An antenna receiving transmissions from a broadcast that hasn't been made. The content is a weather report."},
			{Glyph: "🪣", Name: "Displaced Bucket", Description: "A bucket of water from a rainstorm that occurred forty years from now. The water is fresh."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "🎲", Name: "Paradox Die", Description: "A die that always shows the result before it is rolled. Studying its patterns sharpens your strikes.", BonusATK: 1},
			{Glyph: "🌻", Name: "Temporal Bloom", Description: "A flower that blooms, wilts, and blooms again in a three-second loop. Its pollen invigorates.", BonusMaxHP: 10},
		},
	},
	{ // Floor 7 — Timeline Scar
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "⚰️", Name: "Empty Casket", Description: "A casket for a soldier who hasn't died yet. The nameplate is filled in. The date is blank."},
			{Glyph: "🧳", Name: "War Trunk", Description: "A trunk of personal effects from Timeline Seven. The photographs show places that didn't happen."},
			{Glyph: "🧯", Name: "Depleted Extinguisher", Description: "Used to put out a fire that will start in twelve hours. The fire will be surprised."},
			{Glyph: "🗂️", Name: "Casualty Report", Description: "CASUALTIES FOR TIMELINE SEVEN: [NUMBER VARIES BY OBSERVER]. We have standardised on 'many'."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "🛡️", Name: "Scar-Steel Plate", Description: "Armor forged from timeline scar material. It remembers not being hit. You are harder to damage.", BonusDEF: 1},
			{Glyph: "🌾", Name: "Temporal Root", Description: "A root that grows backward through time. Consuming it grounds you more firmly in the present.", BonusMaxHP: 11},
		},
	},
	{ // Floor 8 — War Room Seven
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "📊", Name: "Strategy Board", Description: "A tactical display showing troop movements from a war that ended and didn't. Arrows point in all directions."},
			{Glyph: "🎨", Name: "War Mural", Description: "A mural depicting the final battle. The paint is fresh. The battle ended 300 years ago."},
			{Glyph: "📋", Name: "After-Action Report", Description: "OUTCOME: VICTORY. CASUALTIES: ACCEPTABLE. NOTE: Definition of 'acceptable' has been revised seven times."},
			{Glyph: "🛠️", Name: "Construct Toolkit", Description: "Tools for maintaining war Constructs. The wrench is set to a tolerance that hasn't been standardised."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "🏆", Name: "Victory Medallion", Description: "Awarded for valor in a battle that may not have occurred. Wearing it sharpens your resolve.", BonusATK: 1},
			{Glyph: "⚖️", Name: "Temporal Anchor", Description: "A heavy device that pins your existence to the current moment. You feel more solidly real.", BonusMaxHP: 11},
		},
	},
	{ // Floor 9 — The Convergence
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "🔭", Name: "Convergence Scope", Description: "A scope aimed at the point where timelines meet. Through it, everything is happening at once."},
			{Glyph: "🪟", Name: "Timeline Window", Description: "A window into an adjacent timeline. In it, you are standing in the same room, reading a different window."},
			{Glyph: "🧴", Name: "Bottled Convergence", Description: "A bottle containing the exact moment where three timelines overlap. It is very heavy for its size."},
			{Glyph: "📌", Name: "Fixed Point", Description: "A nail driven into a moment that refuses to move. Time flows around it like water around a stone."},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "🎯", Name: "Convergence Lens", Description: "A lens that focuses multiple timelines into a single strike. Your attacks arrive from several directions at once.", BonusATK: 1},
			{Glyph: "🧸", Name: "Temporal Anchor Bear", Description: "A child's toy from Anchorpoint. It radiates temporal stability. Holding it makes you feel impossibly safe.", BonusMaxHP: 12},
		},
	},
	{ // Floor 10 — The Eternal Moment
		Common: []generate.FurnitureSpawnEntry{
			{Glyph: "♾️", Name: "Infinity Loop", Description: "A strip of metal twisted into a shape with one surface. Walking along it takes you back to where you started. Always."},
			{Glyph: "🪞", Name: "Recursion Mirror", Description: "A mirror reflecting a mirror reflecting a mirror. Somewhere in the infinite regression, something waves."},
			{Glyph: "🕯️", Name: "Final Candle", Description: "The last candle that will ever be lit. It has been burning since before the concept of 'first'."},
			{Glyph: "📜", Name: "Last Entry", Description: "A researcher's final note: 'I have been here before. I will be here again. I am here now. I think that's all of us.'"},
		},
		Rare: []generate.FurnitureSpawnEntry{
			{Glyph: "⏳", Name: "Eternal Hourglass", Description: "An hourglass in which no sand falls because all the sand has already fallen and hasn't yet begun. You feel more real.", BonusMaxHP: 12},
			{Glyph: "⚔️", Name: "Temporal Blade", Description: "A sword that exists in two moments simultaneously. Its edge cuts where you were and where you will be.", BonusATK: 1},
		},
	},
}
