package mud

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/factory"
	"emoji-roguelike/internal/gamemap"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

// newCityFloor builds the hand-crafted Floor 0 — the City of Emberveil.
// Map is 110×55: grass base terrain, cobblestone streets, brick buildings,
// a river with three bridges, parks with trees, a town square with fountain,
// and a multi-room church with white stone and gold accents.
// SafeZone is set true: no combat, no enemies, no AI ticks.
func newCityFloor(rng *rand.Rand) *Floor {
	gmap := gamemap.New(110, 55)
	w := ecs.NewWorld()

	// ── Base terrain: fill everything with grass ──────────────────────────
	fillTile(gmap, 0, 0, 109, 54, gamemap.MakeGrass())

	// ── Streets (cobblestone floor) ─────────────────────────────────────
	// Main E-W street at y=22
	carveRect(gmap, 0, 22, 109, 22)
	// Town square plaza
	carveRect(gmap, 28, 23, 80, 37)
	// South E-W street at y=43
	carveRect(gmap, 0, 43, 109, 43)

	// North N-S alleys connecting buildings to main street
	carveRect(gmap, 22, 0, 22, 22) // between Tavern and Apothecary/Smithy
	carveRect(gmap, 43, 0, 43, 22) // between Apothecary/Smithy and Homes
	carveRect(gmap, 61, 0, 61, 22) // between Homes and Church

	// South N-S alleys between south buildings (connect south street down)
	carveRect(gmap, 23, 43, 23, 54) // between Market and Inn
	carveRect(gmap, 45, 43, 45, 54) // between Inn and Bakery
	carveRect(gmap, 59, 43, 59, 54) // between Bakery and Home South
	carveRect(gmap, 71, 43, 71, 54) // between Home South and General Store

	// Park paths to river bridges
	carveRect(gmap, 13, 23, 13, 38) // west park path (to west bridge)
	carveRect(gmap, 96, 23, 96, 38) // east park path (to east bridge)

	// ── River ────────────────────────────────────────────────────────────
	fillTile(gmap, 0, 39, 109, 40, gamemap.MakeWater())

	// ── Bridges (floor tiles over the river) ────────────────────────────
	carveRect(gmap, 12, 39, 14, 40) // west bridge (3 wide)
	carveRect(gmap, 53, 39, 56, 40) // center bridge (4 wide, at tower axis)
	carveRect(gmap, 95, 39, 97, 40) // east bridge (3 wide)

	// ── NORTH BUILDINGS ─────────────────────────────────────────────────

	// Tavern "The Sunken Flagon" (x=1..21, y=1..21) — 3 rooms
	carveBuilding(gmap, 1, 1, 21, 21)
	wallV(gmap, 13, 1, 21)          // vertical: Common Room | Snug
	wallH(gmap, 1, 13, 14)          // horizontal: Common Room | Back Cellar
	gmap.Set(13, 8, gamemap.MakeDoor())  // Common Room → Snug
	gmap.Set(7, 14, gamemap.MakeDoor())  // Common Room → Back Cellar
	gmap.Set(10, 21, gamemap.MakeDoor()) // south main entrance

	// Apothecary "Thornroot Remedies" (x=23..42, y=1..12) — 2 rooms
	carveBuilding(gmap, 23, 1, 42, 12)
	wallV(gmap, 34, 1, 12)           // vertical: Shop | Lab
	gmap.Set(34, 7, gamemap.MakeDoor())  // Shop → Lab
	gmap.Set(31, 12, gamemap.MakeDoor()) // south entrance

	// Smithy "The Hearth Forge" (x=23..42, y=14..21) — 2 rooms
	carveBuilding(gmap, 23, 14, 42, 21)
	wallH(gmap, 23, 36, 18)          // horizontal: Work Floor | Forge Pit
	gmap.Set(32, 18, gamemap.MakeDoor())  // Work Floor → Forge Pit
	gmap.Set(32, 21, gamemap.MakeDoor())  // south entrance
	carveRect(gmap, 32, 12, 32, 13)  // connecting path between Apothecary and Smithy

	// Home A (x=45..60, y=1..12) — 2 rooms
	carveBuilding(gmap, 45, 1, 60, 12)
	wallV(gmap, 53, 1, 12)           // vertical: Left Room | Right Room
	gmap.Set(53, 7, gamemap.MakeDoor())  // Left → Right
	gmap.Set(51, 12, gamemap.MakeDoor()) // south entrance

	// Home B (x=45..60, y=14..21) — 2 rooms
	carveBuilding(gmap, 45, 14, 60, 21)
	wallH(gmap, 45, 60, 18)          // horizontal: Front Room | Back Room
	gmap.Set(51, 18, gamemap.MakeDoor())  // Front → Back
	gmap.Set(51, 21, gamemap.MakeDoor())  // south entrance
	carveRect(gmap, 51, 12, 51, 13)  // connecting path between Home A and Home B

	// Church "The Eternal Flame" (x=63..108, y=1..21) — 4 rooms
	// Layout: large Nave (north hall) + Narthex/entrance (south center) +
	//         West Side Chapel + East Vestry. White stone with gold trim.
	carveBuilding(gmap, 63, 1, 108, 21)
	wallH(gmap, 63, 108, 13)          // divides Nave (north) from Narthex/Chapels (south)
	carveRect(gmap, 73, 13, 99, 13)   // wide archway from Narthex into Nave (clear to floor)
	wallV(gmap, 72, 13, 21)           // West Side Chapel divider
	wallV(gmap, 100, 13, 21)          // East Vestry divider
	gmap.Set(72, 17, gamemap.MakeDoor())   // Narthex → West Side Chapel
	gmap.Set(100, 17, gamemap.MakeDoor())  // Narthex → East Vestry
	gmap.Set(85, 21, gamemap.MakeDoor())   // south main entrance

	// ── Tower of Emberveil (centre of town square) ────────────────────────
	carveBuilding(gmap, 48, 25, 60, 35)
	gmap.Set(54, 35, gamemap.MakeDoor()) // south entrance
	stairsDownX, stairsDownY := 54, 30
	gmap.Set(stairsDownX, stairsDownY, gamemap.MakeStairsDown())

	// ── SOUTH BUILDINGS ──────────────────────────────────────────────────

	// Market District (x=1..22, y=44..53) — 2 stalls
	carveBuilding(gmap, 1, 44, 22, 53)
	wallV(gmap, 13, 44, 53)           // divides Stall A | Stall B
	gmap.Set(13, 49, gamemap.MakeDoor())   // between stalls
	gmap.Set(11, 44, gamemap.MakeDoor())   // north entrance

	// Inn "The Weary Boot" (x=24..44, y=44..53) — 3 rooms
	carveBuilding(gmap, 24, 44, 44, 53)
	wallH(gmap, 24, 38, 50)           // Common Room | Private Rooms
	wallV(gmap, 38, 50, 53)           // divides private rooms
	gmap.Set(32, 50, gamemap.MakeDoor())   // Common → Private North
	gmap.Set(38, 48, gamemap.MakeDoor())   // Common → Side Room
	gmap.Set(32, 44, gamemap.MakeDoor())   // north entrance

	// Bakery "Ember's Oven" (x=46..58, y=44..53) — 2 rooms
	carveBuilding(gmap, 46, 44, 58, 53)
	wallV(gmap, 53, 44, 53)           // Shop | Kitchen
	gmap.Set(53, 49, gamemap.MakeDoor())   // Shop → Kitchen
	gmap.Set(51, 44, gamemap.MakeDoor())   // north entrance

	// Home South (x=60..70, y=44..53) — 2 rooms
	carveBuilding(gmap, 60, 44, 70, 53)
	wallH(gmap, 60, 70, 49)           // Front Room | Back Room
	gmap.Set(64, 49, gamemap.MakeDoor())   // Front → Back
	gmap.Set(64, 44, gamemap.MakeDoor())   // north entrance

	// General Store "Yeva's Provisions" (x=72..108, y=44..53) — 2 rooms
	carveBuilding(gmap, 72, 44, 108, 53)
	wallV(gmap, 91, 44, 53)           // Main Shop | Storeroom
	gmap.Set(91, 49, gamemap.MakeDoor())   // Shop → Storeroom
	gmap.Set(87, 44, gamemap.MakeDoor())   // north main entrance

	// ── Rooms list (for findFreeSpawn) ───────────────────────────────────
	gmap.Rooms = []gamemap.Rect{
		// Tavern
		{X1: 2, Y1: 2, X2: 12, Y2: 13},   // Common Room
		{X1: 14, Y1: 2, X2: 20, Y2: 20},   // Snug
		{X1: 2, Y1: 15, X2: 12, Y2: 20},   // Back Cellar
		// Apothecary
		{X1: 24, Y1: 2, X2: 33, Y2: 11},   // Shop
		{X1: 35, Y1: 2, X2: 41, Y2: 11},   // Lab
		// Smithy
		{X1: 24, Y1: 15, X2: 41, Y2: 17},  // Work Floor
		{X1: 24, Y1: 19, X2: 35, Y2: 20},  // Forge Pit
		// Home A
		{X1: 46, Y1: 2, X2: 52, Y2: 11},   // Left Room
		{X1: 54, Y1: 2, X2: 59, Y2: 11},   // Right Room
		// Home B
		{X1: 46, Y1: 15, X2: 59, Y2: 17},  // Front Room
		{X1: 46, Y1: 19, X2: 59, Y2: 20},  // Back Room
		// Church
		{X1: 64, Y1: 2, X2: 107, Y2: 12},  // Nave (large hall)
		{X1: 64, Y1: 14, X2: 71, Y2: 20},  // West Side Chapel
		{X1: 73, Y1: 14, X2: 99, Y2: 20},  // Narthex
		{X1: 101, Y1: 14, X2: 107, Y2: 20}, // East Vestry
		// Tower interior
		{X1: 49, Y1: 26, X2: 59, Y2: 34},  // Tower
		// Town square
		{X1: 28, Y1: 23, X2: 80, Y2: 37},  // Town Square
		// Market
		{X1: 2, Y1: 45, X2: 12, Y2: 52},   // Stall A
		{X1: 14, Y1: 45, X2: 21, Y2: 52},  // Stall B
		// Inn
		{X1: 25, Y1: 45, X2: 37, Y2: 49},  // Common Room
		{X1: 25, Y1: 51, X2: 37, Y2: 52},  // Private Room
		{X1: 39, Y1: 45, X2: 43, Y2: 52},  // Side Room
		// Bakery
		{X1: 47, Y1: 45, X2: 52, Y2: 52},  // Shop
		{X1: 54, Y1: 45, X2: 57, Y2: 52},  // Kitchen
		// Home South
		{X1: 61, Y1: 45, X2: 69, Y2: 48},  // Front Room
		{X1: 61, Y1: 50, X2: 69, Y2: 52},  // Back Room
		// General Store
		{X1: 73, Y1: 45, X2: 90, Y2: 52},  // Main Shop
		{X1: 92, Y1: 45, X2: 107, Y2: 52}, // Storeroom
	}

	// ── Place NPCs ───────────────────────────────────────────────────────
	placeNPC := func(def assets.NPCDef, x, y int) {
		factory.NewNPC(w, def.Name, def.Glyph, component.NPCKind(def.Kind), def.Lines, x, y)
	}

	// Named NPCs — repositioned to match new layout
	placeNPC(assets.CityNPCs[0], 85, 8)   // Sister Maris  — church nave (healer at altar)
	placeNPC(assets.CityNPCs[1], 78, 4)   // Father Brennan — church nave (north end)
	placeNPC(assets.CityNPCs[2], 7, 7)    // Ol' Rudwig    — tavern common room
	placeNPC(assets.CityNPCs[3], 54, 36)  // Soldier Greta — town square near tower door
	placeNPC(assets.CityNPCs[4], 82, 48)  // Merchant Yeva — general store main floor
	placeNPC(assets.CityNPCs[5], 65, 30)  // Scholar Alaric — town square east side
	placeNPC(assets.CityNPCs[6], 7, 49)   // Street Urchin Pip — market stall A
	placeNPC(assets.CityNPCs[7], 64, 46)  // Townsfolk Maren — home south front room
	placeNPC(assets.CityNPCs[8], 4, 50)   // Old Fisher Bram — market stall A back
	placeNPC(assets.CityNPCs[9], 104, 16) // Sister Lena   — church east vestry

	// Animals
	pigeon := assets.CityAnimals[2]
	placeNPC(assets.CityAnimals[0], 30, 22)  // Stray Dog   — main street west
	placeNPC(assets.CityAnimals[1], 75, 22)  // Town Cat    — main street near church
	placeNPC(pigeon, 35, 28)                  // Pigeon A    — town square west
	placeNPC(pigeon, 54, 24)                  // Pigeon B    — town square north (above tower)
	placeNPC(pigeon, 65, 28)                  // Pigeon C    — town square east
	placeNPC(assets.CityAnimals[3], 8, 43)   // Market Hen  — south street near market

	// ── Place inscriptions ───────────────────────────────────────────────
	insc := assets.CityInscriptions
	factory.NewInscription(w, insc[0], 54, 26)  // Tower north interior
	factory.NewInscription(w, insc[1], 62, 30)  // Town square notice
	factory.NewInscription(w, insc[2], 5, 2)    // Tavern entrance sign
	factory.NewInscription(w, insc[3], 78, 45)  // General Store entrance
	factory.NewInscription(w, insc[4], 3, 45)   // Market entrance
	factory.NewInscription(w, insc[5], 35, 35)  // Town square memorial

	// ── Furniture ────────────────────────────────────────────────────────
	pf := func(glyph, name, desc string, x, y int) {
		id := w.CreateEntity()
		w.Add(id, component.Position{X: x, Y: y})
		w.Add(id, component.Renderable{
			Glyph:       glyph,
			FGColor:     tcell.ColorYellow,
			BGColor:     tcell.ColorDefault,
			RenderOrder: 1,
		})
		w.Add(id, component.Furniture{
			Glyph:        glyph,
			Name:         name,
			Description:  desc,
			IsRepeatable: true,
		})
	}

	// Trees in parks and border areas
	for _, pos := range [][2]int{
		// West park (x=0..27, y=23..37)
		{4, 25}, {11, 27}, {7, 31}, {16, 33}, {3, 36}, {20, 26}, {24, 30}, {18, 35},
		// East park (x=81..109, y=23..37)
		{84, 25}, {91, 27}, {88, 31}, {100, 33}, {83, 36}, {97, 26}, {104, 30}, {87, 35},
		// North border
		{5, 0}, {20, 0}, {35, 0}, {50, 0}, {65, 0}, {80, 0}, {95, 0},
		// South border
		{5, 54}, {20, 54}, {35, 54}, {50, 54}, {65, 54}, {80, 54}, {95, 54},
		// Riverbank atmosphere
		{6, 38}, {25, 38}, {78, 38}, {101, 38},
		{6, 41}, {25, 41}, {78, 41}, {101, 41},
	} {
		pf("🌳", "Old Tree",
			"A weathered tree whose roots crack the cobblestones nearby. The bark is carved with old initials.",
			pos[0], pos[1])
	}

	// ── Town Square ──────────────────────────────────────────────────────

	pf("⛲", "Town Fountain",
		"A circular stone fountain with a wide, shallow basin. Carved fish spout arcs of water into the pool. People leave coins in it, wishing for things they won't speak aloud.",
		38, 30)
	pf("🌸", "Flower Bed",
		"A well-tended bed of mountain wildflowers in pink and white. They bloom within sight of the tower, defiant as anything.",
		36, 28)
	pf("🌸", "Flower Bed",
		"A neat square planter overflowing with blossoms. Someone refills it every season regardless of what comes out of the tower.",
		40, 32)
	pf("🌱", "Square Planter",
		"A large stone planter filled with hardy city shrubs. They have outlasted several mayors and three attempted coups.",
		70, 30)
	pf("📋", "Notice Board",
		"A wooden board bristling with flyers, warnings, and job postings. Most of the jobs involve the tower. None list a salary.",
		33, 25)
	pf("🛋️", "Stone Bench",
		"A carved stone bench facing the fountain. Many an adventurer has sat here, staring at the tower, deciding whether to enter.",
		36, 34)
	pf("🛋️", "Park Bench",
		"A bench under the eastern trees. Someone has left a half-eaten apple on it. The pigeons have opinions.",
		67, 34)

	// ── Tavern — Common Room (x=2..12, y=2..13) ─────────────────────────
	pf("🛋️", "Worn Bench",
		"A well-worn bench near a cold hearth. The wood is dark with years of use. Many an adventurer has slept here after one too many.",
		5, 10)
	pf("🕯️", "Tavern Candle",
		"A beeswax candle guttering on a sticky table. The wax has names carved into it. Some of the names have been crossed out.",
		10, 5)
	pf("🫗", "Ale Tap",
		"A large barrel tap dripping something amber and fragrant. It smells of hops and poor decisions. Remarkable.",
		4, 4)

	// ── Tavern — Snug (x=14..20, y=2..20) ───────────────────────────────
	pf("🎭", "Dartboard",
		"A battered dartboard mounted on the wall. Someone has drawn a face on it. The face looks oddly familiar.",
		19, 5)
	pf("🪞", "Cracked Mirror",
		"A mirror behind the bar, cracked diagonally. It shows you looking fractionally more heroic than you feel.",
		18, 12)

	// ── Tavern — Back Cellar (x=2..12, y=15..20) ─────────────────────────
	pf("📦", "Ale Barrel",
		"A great cask of dark ale. The label reads EMBERVEIL DARK — DO NOT OPEN. Someone has opened it.",
		7, 17)
	pf("🧯", "Sand Bucket",
		"A bucket of sand hanging by the cellar steps. Reassuring to see. Less reassuring to think about why it's needed.",
		11, 19)

	// ── Apothecary — Shop (x=24..33, y=2..11) ────────────────────────────
	pf("🧴", "Herb Bottles",
		"Rows of glass bottles filled with dried herbs, tinctures, and unlabelled powders. The powders vary widely in colour.",
		27, 6)
	pf("📋", "Customer Orders",
		"A clipboard of pending orders. Someone urgently needs twelve grams of powdered moonroot. It has been marked OVERDUE.",
		30, 3)

	// ── Apothecary — Lab (x=35..41, y=2..11) ─────────────────────────────
	pf("🔬", "Brass Microscope",
		"A fine brass microscope, lens smudged. Whatever was last examined left a stain on the slide.",
		37, 5)
	pf("🌡️", "Temperature Gauge",
		"A mercury thermometer fixed to the wall. It currently reads OPTIMAL in faded red script.",
		39, 8)
	pf("🧫", "Culture Dish",
		"A sealed glass dish with something growing in it. The colour changes daily. You decide not to ask.",
		36, 8)

	// ── Smithy — Work Floor (x=24..41, y=15..17) ─────────────────────────
	pf("🛠️", "Tool Rack",
		"A rack of well-maintained smithing tools, each hanging in its place. The smith clearly has opinions about order.",
		28, 16)
	pf("📐", "Blueprint Table",
		"A workbench spread with diagrams of blades, hinges, and mechanisms. The margins are dense with second thoughts.",
		35, 16)

	// ── Smithy — Forge Pit (x=24..35, y=19..20) ──────────────────────────
	pf("🔩", "Parts Crate",
		"A crate of bolts, rings, clasps, and half-finished rivets. Spare parts for a hundred projects, none quite done.",
		30, 19)
	pf("🔧", "Smithing Tools",
		"A heavy set of tongs, hammers, and chisels, still warm from recent use. The forge has been out for an hour at most.",
		38, 19)

	// ── Home A — Left Room (x=46..52, y=2..11) ───────────────────────────
	pf("🛋️", "Sitting Chair",
		"A worn armchair facing a cold hearth. Homely, in a weary way. The arm has been repaired with twine.",
		48, 6)
	pf("🌱", "Window Plant",
		"A small plant on a sunny windowsill. It is doing its very best under difficult circumstances.",
		47, 3)

	// ── Home A — Right Room (x=54..59, y=2..11) ──────────────────────────
	pf("📚", "Bookshelf",
		"A tall shelf of well-thumbed books. They look well-loved, read many times by lamplight.",
		57, 6)
	pf("🖼️", "Portrait",
		"A painted portrait of a couple on their wedding day. They look genuinely happy.",
		55, 3)

	// ── Home B — Front Room (x=46..59, y=15..17) ─────────────────────────
	pf("🪴", "House Plant",
		"A healthy-looking fern on a stand near the window. A small sign reads: PLEASE DO NOT TOUCH.",
		48, 16)
	pf("🗂️", "File Cabinet",
		"A cabinet crammed with papers. IMPORTANT is written in red on the outside. It has been written on and crossed out twice.",
		57, 16)

	// ── Home B — Back Room (x=46..59, y=19..20) ──────────────────────────
	pf("🖼️", "Family Portrait",
		"A painted portrait of a family. Everyone is smiling except the dog, who knows something.",
		50, 19)

	// ── Church — Nave (x=64..107, y=2..12) ───────────────────────────────
	// White marble pews with gold filigree lining the nave
	pf("⚰️", "White Marble Pew",
		"A long pew of pale white marble, cold and smooth. The kneeling impressions are worn deep. Gold filigree traces the armrests.",
		70, 10)
	pf("⚰️", "White Marble Pew",
		"Another row of white marble pews, serene and still. The Eternal Flame's light catches the gold inlay along the base.",
		98, 10)
	pf("⚰️", "Front Pew",
		"The front pew, closest to the altar. The stone here is almost luminous, shot through with veins of pale gold.",
		84, 4)
	// Altar area at north end
	pf("🕯️", "Gold Altar Candle",
		"A tall white candle in a gold-chased stand. Its flame burns without waver, without wind. The Eternal Flame, in miniature.",
		76, 3)
	pf("🕯️", "Gold Altar Candle",
		"A paired candle flanking the altar, white wax in beaten gold. The light it casts has no shadow.",
		94, 3)
	pf("🌺", "Altar Offering",
		"Fresh white flowers arranged at the altar's base. Someone climbs the tower stairs daily just to leave them here.",
		85, 3)
	pf("🏺", "Ceremonial Urn",
		"A gold-inlaid white urn on a marble pedestal. The inscription reads: Those who return, light a candle. Those who do not, are light.",
		89, 2)

	// ── Church — Narthex (x=73..99, y=14..20) ────────────────────────────
	pf("🕯️", "Vestibule Candle",
		"A tall white candle burning near the entrance. It has burned here, without interruption, for longer than anyone remembers.",
		79, 19)
	pf("📋", "Parish Notice",
		"A board of parish announcements and weekly prayer schedules. The handwriting is impeccably neat.",
		93, 19)

	// ── Church — West Side Chapel (x=64..71, y=14..20) ───────────────────
	pf("🌺", "Chapel Flowers",
		"White and gold wildflowers in a fluted vase. The side chapel always smells faintly of incense and cold stone.",
		67, 16)
	pf("🛋️", "Kneeling Bench",
		"A white cushioned kneeling bench facing a bare stone wall. The cushion is worn smooth by years of use.",
		66, 18)

	// ── Church — East Vestry (x=101..107, y=14..20) ──────────────────────
	pf("📚", "Parish Records",
		"Leather-bound ledgers of parish records going back centuries. The older ones are written in a script you cannot identify.",
		103, 16)
	pf("🗃️", "Vestment Box",
		"An ornate white and gold box of ceremonial vestments, carefully folded. The embroidery depicts the Eternal Flame.",
		105, 19)

	// ── Market District — Stall A (x=2..12, y=45..52) ────────────────────
	pf("🪣", "Water Barrel",
		"A barrel of fresh water beside the stall. The market is careful about fire since the incident last summer.",
		4, 48)
	pf("🌾", "Grain Sacks",
		"Bulging sacks of dried grain and milled flour, stacked against the wall. The smell of bread is everywhere.",
		8, 50)
	pf("📦", "Market Crate",
		"A crate labelled ASSORTED PROVISIONS. The assortment changes daily and is rarely what you expect.",
		10, 47)

	// ── Market District — Stall B (x=14..21, y=45..52) ───────────────────
	pf("🧴", "Spice Jars",
		"Rows of small jars filled with dried spices. The one labelled MILD is not mild.",
		16, 48)
	pf("📊", "Market Ledger",
		"A fat ledger of market accounts. The numbers are large. The margins contain strong opinions.",
		20, 47)

	// ── Inn — Common Room (x=25..37, y=45..49) ────────────────────────────
	pf("🛋️", "Common Bench",
		"A long bench worn smooth by travellers. The wood remembers a thousand weary journeys.",
		27, 47)
	pf("🕯️", "Common Room Candle",
		"A cluster of candles on the main table. The innkeeper relights them every evening without fail.",
		35, 46)
	pf("🫗", "Ale Tap",
		"A small tap serving the inn's house ale. It is described on the menu as 'robust'. It is robust.",
		28, 48)

	// ── Inn — Private Room (x=25..37, y=51..52) ───────────────────────────
	pf("🛋️", "Sleeping Cot",
		"A narrow cot with a thin mattress. Not comfortable, but better than the dungeon floor, which you may find out soon enough.",
		28, 51)
	pf("🪞", "Small Mirror",
		"A small mirror above the cot. It shows someone who really needs rest.",
		35, 51)

	// ── Inn — Side Room (x=39..43, y=45..52) ─────────────────────────────
	pf("🗂️", "Guest Ledger",
		"The inn's guest ledger. You notice several entries from the same adventurer across different years. They kept coming back.",
		40, 47)
	pf("🕯️", "Night Candle",
		"A candle stub in a tin holder. It will last the night. Probably.",
		42, 51)

	// ── Bakery — Shop (x=47..52, y=45..52) ───────────────────────────────
	pf("🧴", "Flour Jars",
		"Sealed jars of fine-milled flour on the counter. The baker weighs every portion by hand.",
		49, 49)
	pf("📋", "Order Board",
		"A chalkboard of daily orders and specials. Today's special is 'ember bread'. It glows faintly and is very popular.",
		51, 46)

	// ── Bakery — Kitchen (x=54..57, y=45..52) ────────────────────────────
	pf("🫕", "Cooking Pot",
		"A deep cast-iron pot, permanently seasoned after years on the fire. Whatever it last held smells extraordinary.",
		55, 48)
	pf("🌡️", "Oven Gauge",
		"A temperature gauge mounted beside the stone oven. The needle is currently pegged hard into the red. Good.",
		56, 51)

	// ── Home South — Front Room (x=61..69, y=45..48) ─────────────────────
	pf("🪴", "Potted Plant",
		"A lush potted plant on a stand. It is thriving despite the proximity to the dungeon. Inspiring.",
		63, 46)
	pf("🖼️", "Landscape Painting",
		"A painting of rolling hills and open sky. Whoever lives here misses somewhere else.",
		67, 46)

	// ── Home South — Back Room (x=61..69, y=50..52) ──────────────────────
	pf("📚", "Small Bookshelf",
		"A modest bookshelf of well-read volumes. They have been read many times, with great attention.",
		63, 51)
	pf("🌱", "Window Herb",
		"A small pot of herbs on the windowsill. The herbs are labelled in careful handwriting.",
		67, 51)

	// ── General Store — Main Shop (x=73..90, y=45..52) ───────────────────
	pf("🗃️", "Supply Shelf",
		"Shelves crammed with provisions, tools, and sundries. Yeva stocks more than she ever sells and sells more than she stocks.",
		75, 48)
	pf("📦", "Supply Crate",
		"A crate stamped HANDLE WITH CARE — FRAGILE. It rattles faintly when you lean close. Something shifts inside.",
		88, 51)
	pf("📊", "Account Ledger",
		"Yeva's ledger. The numbers are very large. The margins contain pointed observations about her competitors.",
		87, 46)

	// ── General Store — Storeroom (x=92..107, y=45..52) ──────────────────
	pf("📦", "Storage Crate",
		"A heavy crate stacked with raw supplies. The lid is nailed shut and the nails are new.",
		95, 48)
	pf("🪣", "Storage Barrel",
		"A sealed barrel with a wax-stamped lid. The stamp reads RESERVE. The reserve is apparently considerable.",
		100, 48)
	pf("🔑", "Key Cabinet",
		"A locked cabinet of small keys. The keys are labelled in a code only Yeva understands. She has never needed to explain it.",
		106, 47)

	return &Floor{
		Num:             0,
		World:           w,
		GMap:            gmap,
		Rng:             rng,
		SpawnX:          54,
		SpawnY:          22,
		StairsDownX:     stairsDownX,
		StairsDownY:     stairsDownY,
		StairsUpX:       -1, // no stairs up in the city (top of the world)
		StairsUpY:       -1,
		RespawnCooldown: -1,
		SafeZone:        true,
	}
}

// ── Map-building helpers ──────────────────────────────────────────────────────

// carveBuilding draws outer walls for the rect and carves the interior to floor.
func carveBuilding(gmap *gamemap.GameMap, x1, y1, x2, y2 int) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			if !gmap.InBounds(x, y) {
				continue
			}
			if x == x1 || x == x2 || y == y1 || y == y2 {
				gmap.Set(x, y, gamemap.MakeWall())
			} else {
				gmap.Set(x, y, gamemap.MakeFloor())
			}
		}
	}
}

// wallH fills a horizontal wall segment at y from x1 to x2.
func wallH(gmap *gamemap.GameMap, x1, x2, y int) {
	for x := x1; x <= x2; x++ {
		if gmap.InBounds(x, y) {
			gmap.Set(x, y, gamemap.MakeWall())
		}
	}
}

// wallV fills a vertical wall segment at x from y1 to y2.
func wallV(gmap *gamemap.GameMap, x, y1, y2 int) {
	for y := y1; y <= y2; y++ {
		if gmap.InBounds(x, y) {
			gmap.Set(x, y, gamemap.MakeWall())
		}
	}
}

// fillTile sets every tile in the inclusive rect [x1,y1]..[x2,y2] to t.
func fillTile(gmap *gamemap.GameMap, x1, y1, x2, y2 int, t gamemap.Tile) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			if gmap.InBounds(x, y) {
				gmap.Set(x, y, t)
			}
		}
	}
}

// carveRect sets all tiles in the inclusive rectangle [x1,y1]..[x2,y2] to floor.
func carveRect(gmap *gamemap.GameMap, x1, y1, x2, y2 int) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			if gmap.InBounds(x, y) {
				gmap.Set(x, y, gamemap.MakeFloor())
			}
		}
	}
}
