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
	// Town square plaza (slightly narrowed)
	carveRect(gmap, 28, 23, 72, 36)
	// South E-W street at y=43
	carveRect(gmap, 0, 43, 109, 43)

	// North N-S alleys connecting buildings to main street
	carveRect(gmap, 6, 12, 6, 22)   // south of Tavern
	carveRect(gmap, 18, 8, 18, 10)  // between stacked Apothecary/Smithy
	carveRect(gmap, 18, 16, 18, 22) // south of Smithy
	carveRect(gmap, 29, 8, 29, 10)  // between stacked Home A/Home B
	carveRect(gmap, 29, 16, 29, 22) // south of Home B
	carveRect(gmap, 47, 14, 47, 22) // south of Church

	// South N-S alleys between south buildings
	carveRect(gmap, 13, 43, 13, 54) // between Market and Inn
	carveRect(gmap, 25, 43, 25, 54) // between Inn and Bakery
	carveRect(gmap, 36, 43, 36, 54) // between Bakery and Home South
	carveRect(gmap, 44, 43, 44, 54) // between Home South and General Store

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

	// Tavern "The Sunken Flagon" (x=1..12, y=2..12) — 2 rooms
	carveBuilding(gmap, 1, 2, 12, 12)
	wallV(gmap, 7, 2, 12)                // vertical: Common Room | Private Room
	gmap.Set(7, 7, gamemap.MakeDoor())   // Common Room → Private Room
	gmap.Set(6, 12, gamemap.MakeDoor())  // south main entrance

	// Apothecary "Thornroot Remedies" (x=14..23, y=2..8) — 2 rooms
	carveBuilding(gmap, 14, 2, 23, 8)
	wallV(gmap, 19, 2, 8)               // vertical: Shop | Lab
	gmap.Set(19, 5, gamemap.MakeDoor()) // Shop → Lab
	gmap.Set(18, 8, gamemap.MakeDoor()) // south entrance

	// Smithy "The Hearth Forge" (x=14..23, y=10..16) — 1 room
	carveBuilding(gmap, 14, 10, 23, 16)
	gmap.Set(18, 16, gamemap.MakeDoor()) // south entrance

	// Home A (x=25..33, y=2..8) — 2 rooms
	carveBuilding(gmap, 25, 2, 33, 8)
	wallV(gmap, 29, 2, 8)               // vertical: Left Room | Right Room
	gmap.Set(29, 5, gamemap.MakeDoor()) // Left → Right
	gmap.Set(29, 8, gamemap.MakeDoor()) // south entrance

	// Home B (x=25..33, y=10..16) — 1 room
	carveBuilding(gmap, 25, 10, 33, 16)
	gmap.Set(29, 16, gamemap.MakeDoor()) // south entrance

	// Church "The Eternal Flame" (x=35..58, y=2..14) — 3 rooms
	// Layout: Nave (north) + Narthex (south center) + West Chapel + East Vestry
	carveBuilding(gmap, 35, 2, 58, 14)
	wallH(gmap, 35, 58, 9)              // divides Nave (north) from Narthex/Chapels (south)
	carveRect(gmap, 41, 9, 52, 9)       // wide archway from Narthex into Nave
	wallV(gmap, 40, 9, 14)              // West Chapel divider
	wallV(gmap, 53, 9, 14)              // East Vestry divider
	gmap.Set(40, 12, gamemap.MakeDoor())  // Narthex → West Chapel
	gmap.Set(53, 12, gamemap.MakeDoor())  // Narthex → East Vestry
	gmap.Set(47, 14, gamemap.MakeDoor())  // south main entrance

	// ── Tower of Emberveil (centre of town square) ────────────────────────
	carveBuilding(gmap, 48, 26, 56, 33)
	gmap.Set(52, 33, gamemap.MakeDoor()) // south entrance
	stairsDownX, stairsDownY := 52, 30
	gmap.Set(stairsDownX, stairsDownY, gamemap.MakeStairsDown())

	// ── SOUTH BUILDINGS ──────────────────────────────────────────────────

	// Market District (x=1..12, y=44..50) — 2 stalls
	carveBuilding(gmap, 1, 44, 12, 50)
	wallV(gmap, 7, 44, 50)               // divides Stall A | Stall B
	gmap.Set(7, 47, gamemap.MakeDoor())   // between stalls
	gmap.Set(6, 44, gamemap.MakeDoor())   // north entrance

	// Inn "The Weary Boot" (x=14..24, y=44..50) — 2 rooms
	carveBuilding(gmap, 14, 44, 24, 50)
	wallV(gmap, 20, 44, 50)              // Common Room | Side Room
	gmap.Set(20, 47, gamemap.MakeDoor()) // Common → Side Room
	gmap.Set(18, 44, gamemap.MakeDoor()) // north entrance

	// Bakery "Ember's Oven" (x=26..35, y=44..50) — 2 rooms
	carveBuilding(gmap, 26, 44, 35, 50)
	wallV(gmap, 31, 44, 50)              // Shop | Kitchen
	gmap.Set(31, 47, gamemap.MakeDoor()) // Shop → Kitchen
	gmap.Set(30, 44, gamemap.MakeDoor()) // north entrance

	// Home South (x=37..43, y=44..50) — 1 room
	carveBuilding(gmap, 37, 44, 43, 50)
	gmap.Set(40, 44, gamemap.MakeDoor()) // north entrance

	// General Store "Yeva's Provisions" (x=45..62, y=44..50) — 2 rooms
	carveBuilding(gmap, 45, 44, 62, 50)
	wallV(gmap, 54, 44, 50)              // Main Shop | Storeroom
	gmap.Set(54, 47, gamemap.MakeDoor()) // Shop → Storeroom
	gmap.Set(52, 44, gamemap.MakeDoor()) // north main entrance

	// ── Rooms list (for findFreeSpawn) ───────────────────────────────────
	gmap.Rooms = []gamemap.Rect{
		// Tavern
		{X1: 2, Y1: 3, X2: 6, Y2: 11},    // Common Room
		{X1: 8, Y1: 3, X2: 11, Y2: 11},   // Private Room
		// Apothecary
		{X1: 15, Y1: 3, X2: 18, Y2: 7},   // Shop
		{X1: 20, Y1: 3, X2: 22, Y2: 7},   // Lab
		// Smithy
		{X1: 15, Y1: 11, X2: 22, Y2: 15}, // Work Floor
		// Home A
		{X1: 26, Y1: 3, X2: 28, Y2: 7},   // Left Room
		{X1: 30, Y1: 3, X2: 32, Y2: 7},   // Right Room
		// Home B
		{X1: 26, Y1: 11, X2: 32, Y2: 15}, // Main Room
		// Church
		{X1: 36, Y1: 3, X2: 57, Y2: 8},   // Nave
		{X1: 36, Y1: 10, X2: 39, Y2: 13}, // West Chapel
		{X1: 41, Y1: 10, X2: 52, Y2: 13}, // Narthex
		{X1: 54, Y1: 10, X2: 57, Y2: 13}, // East Vestry
		// Tower interior
		{X1: 49, Y1: 27, X2: 55, Y2: 32}, // Tower
		// Town square
		{X1: 28, Y1: 23, X2: 72, Y2: 36}, // Town Square
		// Market
		{X1: 2, Y1: 45, X2: 6, Y2: 49},   // Stall A
		{X1: 8, Y1: 45, X2: 11, Y2: 49},  // Stall B
		// Inn
		{X1: 15, Y1: 45, X2: 19, Y2: 49}, // Common Room
		{X1: 21, Y1: 45, X2: 23, Y2: 49}, // Side Room
		// Bakery
		{X1: 27, Y1: 45, X2: 30, Y2: 49}, // Shop
		{X1: 32, Y1: 45, X2: 34, Y2: 49}, // Kitchen
		// Home South
		{X1: 38, Y1: 45, X2: 42, Y2: 49}, // Main Room
		// General Store
		{X1: 46, Y1: 45, X2: 53, Y2: 49}, // Main Shop
		{X1: 55, Y1: 45, X2: 61, Y2: 49}, // Storeroom
	}

	// ── Place NPCs ───────────────────────────────────────────────────────
	placeNPC := func(def assets.NPCDef, x, y int) {
		factory.NewNPC(w, def.Name, def.Glyph, component.NPCKind(def.Kind), def.Lines, x, y)
	}

	// Named NPCs — repositioned to match new layout
	placeNPC(assets.CityNPCs[0], 45, 5)  // Sister Maris  — church nave (healer)
	placeNPC(assets.CityNPCs[1], 50, 4)  // Father Brennan — church nave (north end)
	placeNPC(assets.CityNPCs[2], 4, 7)   // Ol' Rudwig    — tavern common room
	placeNPC(assets.CityNPCs[3], 51, 34) // Soldier Greta — town square near tower door
	placeNPC(assets.CityNPCs[4], 49, 47) // Merchant Yeva — general store main floor
	placeNPC(assets.CityNPCs[5], 60, 30) // Scholar Alaric — town square east side
	placeNPC(assets.CityNPCs[6], 4, 47)  // Street Urchin Pip — market stall A
	placeNPC(assets.CityNPCs[7], 40, 47) // Townsfolk Maren — home south
	placeNPC(assets.CityNPCs[8], 4, 49)  // Old Fisher Bram — market stall A back
	placeNPC(assets.CityNPCs[9], 56, 11) // Sister Lena   — church east vestry

	// Animals
	pigeon := assets.CityAnimals[2]
	placeNPC(assets.CityAnimals[0], 30, 22) // Stray Dog   — main street west
	placeNPC(assets.CityAnimals[1], 60, 22) // Town Cat    — main street east
	placeNPC(pigeon, 35, 28)                 // Pigeon A    — town square west
	placeNPC(pigeon, 52, 24)                 // Pigeon B    — town square north
	placeNPC(pigeon, 62, 28)                 // Pigeon C    — town square east
	placeNPC(assets.CityAnimals[3], 8, 43)  // Market Hen  — south street near market

	// ── Place inscriptions ───────────────────────────────────────────────
	insc := assets.CityInscriptions
	factory.NewInscription(w, insc[0], 52, 27) // Tower north interior
	factory.NewInscription(w, insc[1], 60, 28) // Town square notice
	factory.NewInscription(w, insc[2], 3, 3)   // Tavern entrance sign
	factory.NewInscription(w, insc[3], 48, 45) // General Store entrance
	factory.NewInscription(w, insc[4], 3, 45)  // Market entrance
	factory.NewInscription(w, insc[5], 35, 34) // Town square memorial

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
		// West park (x=0..27, y=23..36)
		{4, 25}, {11, 27}, {7, 31}, {16, 33}, {3, 36}, {20, 26}, {24, 30}, {18, 35},
		// East park (x=73..109, y=23..36)
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
	pf("📋", "Notice Board",
		"A wooden board bristling with flyers, warnings, and job postings. Most of the jobs involve the tower. None list a salary.",
		33, 25)
	pf("🛋️", "Stone Bench",
		"A carved stone bench facing the fountain. Many an adventurer has sat here, staring at the tower, deciding whether to enter.",
		36, 34)
	pf("🛋️", "Park Bench",
		"A bench under the eastern trees. Someone has left a half-eaten apple on it. The pigeons have opinions.",
		65, 34)

	// ── Tavern — Common Room (x=2..6, y=3..11) ──────────────────────────
	pf("🕯️", "Tavern Candle",
		"A beeswax candle guttering on a sticky table. The wax has names carved into it. Some of the names have been crossed out.",
		3, 5)
	pf("🫗", "Ale Tap",
		"A large barrel tap dripping something amber and fragrant. It smells of hops and poor decisions. Remarkable.",
		5, 9)

	// ── Tavern — Private Room (x=8..11, y=3..11) ────────────────────────
	pf("🎭", "Dartboard",
		"A battered dartboard mounted on the wall. Someone has drawn a face on it. The face looks oddly familiar.",
		10, 5)
	pf("🪞", "Cracked Mirror",
		"A mirror behind the bar, cracked diagonally. It shows you looking fractionally more heroic than you feel.",
		9, 9)

	// ── Apothecary — Shop (x=15..18, y=3..7) ────────────────────────────
	pf("🧴", "Herb Bottles",
		"Rows of glass bottles filled with dried herbs, tinctures, and unlabelled powders. The powders vary widely in colour.",
		16, 5)

	// ── Apothecary — Lab (x=20..22, y=3..7) ─────────────────────────────
	pf("🔬", "Brass Microscope",
		"A fine brass microscope, lens smudged. Whatever was last examined left a stain on the slide.",
		21, 5)
	pf("🧫", "Culture Dish",
		"A sealed glass dish with something growing in it. The colour changes daily. You decide not to ask.",
		21, 6)

	// ── Smithy — Work Floor (x=15..22, y=11..15) ────────────────────────
	pf("🛠️", "Tool Rack",
		"A rack of well-maintained smithing tools, each hanging in its place. The smith clearly has opinions about order.",
		17, 13)
	pf("🔩", "Parts Crate",
		"A crate of bolts, rings, clasps, and half-finished rivets. Spare parts for a hundred projects, none quite done.",
		21, 13)

	// ── Home A — Left Room (x=26..28, y=3..7) ───────────────────────────
	pf("🛋️", "Sitting Chair",
		"A worn armchair facing a cold hearth. Homely, in a weary way. The arm has been repaired with twine.",
		27, 5)

	// ── Home A — Right Room (x=30..32, y=3..7) ──────────────────────────
	pf("📚", "Bookshelf",
		"A tall shelf of well-thumbed books. They look well-loved, read many times by lamplight.",
		31, 5)

	// ── Home B — Main Room (x=26..32, y=11..15) ─────────────────────────
	pf("🪴", "House Plant",
		"A healthy-looking fern on a stand near the window. A small sign reads: PLEASE DO NOT TOUCH.",
		28, 13)
	pf("🖼️", "Family Portrait",
		"A painted portrait of a family. Everyone is smiling except the dog, who knows something.",
		31, 13)

	// ── Church — Nave (x=36..57, y=3..8) ─────────────────────────────────
	pf("⚰️", "White Marble Pew",
		"A long pew of pale white marble, cold and smooth. The kneeling impressions are worn deep. Gold filigree traces the armrests.",
		40, 7)
	pf("⚰️", "White Marble Pew",
		"Another row of white marble pews, serene and still. The Eternal Flame's light catches the gold inlay along the base.",
		52, 7)
	pf("🕯️", "Gold Altar Candle",
		"A tall white candle in a gold-chased stand. Its flame burns without waver, without wind. The Eternal Flame, in miniature.",
		42, 3)
	pf("🕯️", "Gold Altar Candle",
		"A paired candle flanking the altar, white wax in beaten gold. The light it casts has no shadow.",
		54, 3)
	pf("🌺", "Altar Offering",
		"Fresh white flowers arranged at the altar's base. Someone climbs the tower stairs daily just to leave them here.",
		48, 3)

	// ── Church — Narthex (x=41..52, y=10..13) ────────────────────────────
	pf("🕯️", "Vestibule Candle",
		"A tall white candle burning near the entrance. It has burned here, without interruption, for longer than anyone remembers.",
		44, 12)
	pf("📋", "Parish Notice",
		"A board of parish announcements and weekly prayer schedules. The handwriting is impeccably neat.",
		49, 12)

	// ── Church — West Chapel (x=36..39, y=10..13) ────────────────────────
	pf("🌺", "Chapel Flowers",
		"White and gold wildflowers in a fluted vase. The side chapel always smells faintly of incense and cold stone.",
		37, 11)

	// ── Church — East Vestry (x=54..57, y=10..13) ────────────────────────
	pf("📚", "Parish Records",
		"Leather-bound ledgers of parish records going back centuries. The older ones are written in a script you cannot identify.",
		55, 11)

	// ── Market District — Stall A (x=2..6, y=45..49) ────────────────────
	pf("🪣", "Water Barrel",
		"A barrel of fresh water beside the stall. The market is careful about fire since the incident last summer.",
		3, 47)
	pf("🌾", "Grain Sacks",
		"Bulging sacks of dried grain and milled flour, stacked against the wall. The smell of bread is everywhere.",
		5, 49)

	// ── Market District — Stall B (x=8..11, y=45..49) ───────────────────
	pf("🧴", "Spice Jars",
		"Rows of small jars filled with dried spices. The one labelled MILD is not mild.",
		9, 47)
	pf("📊", "Market Ledger",
		"A fat ledger of market accounts. The numbers are large. The margins contain strong opinions.",
		10, 49)

	// ── Inn — Common Room (x=15..19, y=45..49) ──────────────────────────
	pf("🛋️", "Common Bench",
		"A long bench worn smooth by travellers. The wood remembers a thousand weary journeys.",
		16, 47)
	pf("🕯️", "Common Room Candle",
		"A cluster of candles on the main table. The innkeeper relights them every evening without fail.",
		18, 46)

	// ── Inn — Side Room (x=21..23, y=45..49) ────────────────────────────
	pf("🗂️", "Guest Ledger",
		"The inn's guest ledger. You notice several entries from the same adventurer across different years. They kept coming back.",
		22, 47)

	// ── Bakery — Shop (x=27..30, y=45..49) ──────────────────────────────
	pf("📋", "Order Board",
		"A chalkboard of daily orders and specials. Today's special is 'ember bread'. It glows faintly and is very popular.",
		28, 46)

	// ── Bakery — Kitchen (x=32..34, y=45..49) ───────────────────────────
	pf("🫕", "Cooking Pot",
		"A deep cast-iron pot, permanently seasoned after years on the fire. Whatever it last held smells extraordinary.",
		33, 47)

	// ── Home South — Main Room (x=38..42, y=45..49) ─────────────────────
	pf("🪴", "Potted Plant",
		"A lush potted plant on a stand. It is thriving despite the proximity to the dungeon. Inspiring.",
		39, 47)
	pf("🖼️", "Landscape Painting",
		"A painting of rolling hills and open sky. Whoever lives here misses somewhere else.",
		41, 47)

	// ── General Store — Main Shop (x=46..53, y=45..49) ──────────────────
	pf("🗃️", "Supply Shelf",
		"Shelves crammed with provisions, tools, and sundries. Yeva stocks more than she ever sells and sells more than she stocks.",
		47, 47)
	pf("📊", "Account Ledger",
		"Yeva's ledger. The numbers are very large. The margins contain pointed observations about her competitors.",
		51, 46)

	// ── General Store — Storeroom (x=55..61, y=45..49) ──────────────────
	pf("📦", "Storage Crate",
		"A heavy crate stacked with raw supplies. The lid is nailed shut and the nails are new.",
		57, 47)
	pf("🔑", "Key Cabinet",
		"A locked cabinet of small keys. The keys are labelled in a code only Yeva understands. She has never needed to explain it.",
		60, 47)

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
