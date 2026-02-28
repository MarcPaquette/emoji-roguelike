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
// The map is 80×40, statically laid out with buildings, streets, NPCs,
// inscriptions, and a stairs-down tile that leads to dungeon Floor 1.
// SafeZone is set true: no combat, no enemies, no AI ticks.
func newCityFloor(rng *rand.Rand) *Floor {
	gmap := gamemap.New(80, 40)
	w := ecs.NewWorld()

	// ── Carve building interiors ───────────────────────────────────────────

	// Tavern "The Sunken Flagon" (outer: 2,2→17,13)
	carveRect(gmap, 3, 3, 16, 12)

	// Church "Eternal Flame" (outer: 61,2→77,13)
	carveRect(gmap, 62, 3, 76, 12)

	// Home NW (outer: 20,2→32,10)
	carveRect(gmap, 21, 3, 31, 9)

	// Home NE (outer: 47,2→59,10)
	carveRect(gmap, 48, 3, 58, 9)

	// Tower of Emberveil: carve interior then restore outer shell walls
	carveRect(gmap, 36, 8, 43, 14) // tower interior floor

	// Town square (outer: 28,14→51,21) — carved after tower so we can patch
	carveRect(gmap, 28, 14, 51, 21)

	// Restore tower outer shell walls that the square carve set to floor.
	// West wall at y=14: x=35 (tower left boundary at y=14)
	gmap.Set(35, 14, gamemap.MakeWall())
	// East wall at y=14: x=44
	gmap.Set(44, 14, gamemap.MakeWall())
	// South wall y=15 — keep entrance open at x=38..41
	for x := 35; x <= 44; x++ {
		if x < 38 || x > 41 {
			gmap.Set(x, 15, gamemap.MakeWall())
		}
	}

	// General Store "Yeva's Provisions" (outer: 61,22→77,33)
	carveRect(gmap, 62, 23, 76, 32)

	// Market District (outer: 2,22→17,33)
	carveRect(gmap, 3, 23, 16, 32)

	// Home SW (outer: 20,25→32,33)
	carveRect(gmap, 21, 26, 31, 32)

	// Home SE (outer: 47,25→59,33)
	carveRect(gmap, 48, 26, 58, 32)

	// ── Carve streets and alleys ───────────────────────────────────────────

	// Main E-W street at y=21 (full width)
	carveRect(gmap, 0, 21, 79, 21)

	// N-S tower alley: x=40, y=15..21 (tower entrance → main street)
	carveRect(gmap, 40, 15, 40, 21)

	// NW alley: connect tavern east exit to square west (y=14, x=17..28)
	carveRect(gmap, 17, 14, 28, 14)
	// Tavern south exit at x=9
	carveRect(gmap, 9, 13, 9, 14)

	// NE alley: connect church south exit to square east (y=14, x=51..61)
	carveRect(gmap, 51, 14, 61, 14)
	// Church south exit at x=69
	carveRect(gmap, 69, 13, 69, 14)

	// SW alley: connect market north to square south (y=22, x=2..28)
	carveRect(gmap, 2, 22, 28, 22)
	// Market north exit at x=9
	carveRect(gmap, 9, 22, 9, 23)

	// SE alley: connect store north to square south (y=22, x=51..77)
	carveRect(gmap, 51, 22, 77, 22)
	// Store north exit at x=69
	carveRect(gmap, 69, 22, 69, 23)

	// Home NW → main street: x=25, y=9..21
	carveRect(gmap, 25, 10, 25, 21)

	// Home NE → main street: x=53, y=9..21
	carveRect(gmap, 53, 10, 53, 21)

	// Home SW → main street: x=25, y=22..26
	carveRect(gmap, 25, 22, 25, 26)

	// Home SE → main street: x=53, y=22..26
	carveRect(gmap, 53, 22, 53, 26)

	// ── Stairs down ──────────────────────────────────────────────────────
	stairsDownX, stairsDownY := 40, 16
	gmap.Set(stairsDownX, stairsDownY, gamemap.MakeStairsDown())

	// ── Populate gmap.Rooms so findFreeSpawn works ────────────────────────
	gmap.Rooms = []gamemap.Rect{
		{X1: 3, Y1: 3, X2: 16, Y2: 12},   // Tavern
		{X1: 62, Y1: 3, X2: 76, Y2: 12},  // Church
		{X1: 21, Y1: 3, X2: 31, Y2: 9},   // Home NW
		{X1: 48, Y1: 3, X2: 58, Y2: 9},   // Home NE
		{X1: 36, Y1: 8, X2: 43, Y2: 14},  // Tower interior
		{X1: 28, Y1: 14, X2: 51, Y2: 21}, // Town Square
		{X1: 62, Y1: 23, X2: 76, Y2: 32}, // Store
		{X1: 3, Y1: 23, X2: 16, Y2: 32},  // Market
		{X1: 21, Y1: 26, X2: 31, Y2: 32}, // Home SW
		{X1: 48, Y1: 26, X2: 58, Y2: 32}, // Home SE
	}

	// ── Place NPCs ────────────────────────────────────────────────────────
	placeNPC := func(def assets.NPCDef, x, y int) {
		factory.NewNPC(w, def.Name, def.Glyph, component.NPCKind(def.Kind), def.Lines, x, y)
	}

	// Named NPCs (indices match CityNPCs order)
	placeNPC(assets.CityNPCs[0], 68, 7)  // Sister Maris — church altar
	placeNPC(assets.CityNPCs[1], 65, 5)  // Father Brennan — church left
	placeNPC(assets.CityNPCs[2], 9, 7)   // Ol' Rudwig — tavern center
	placeNPC(assets.CityNPCs[3], 40, 20) // Soldier Greta — tower square south
	placeNPC(assets.CityNPCs[4], 68, 28) // Merchant Yeva — store center
	placeNPC(assets.CityNPCs[5], 38, 17) // Scholar Alaric — town square
	placeNPC(assets.CityNPCs[6], 9, 28)  // Street Urchin Pip — market
	placeNPC(assets.CityNPCs[7], 25, 29) // Townsfolk Maren — home SW
	placeNPC(assets.CityNPCs[8], 9, 30)  // Old Fisher Bram — market interior
	placeNPC(assets.CityNPCs[9], 64, 5)  // Sister Lena — church interior

	// Animals
	pigeon := assets.CityAnimals[2] // Pigeon def used 3 times
	placeNPC(assets.CityAnimals[0], 20, 21)  // Stray Dog — main street
	placeNPC(assets.CityAnimals[1], 63, 3)   // Town Cat — church entrance
	placeNPC(pigeon, 35, 17)                 // Pigeon A — town square west
	placeNPC(pigeon, 41, 18)                 // Pigeon B — town square
	placeNPC(pigeon, 45, 17)                 // Pigeon C — town square east
	placeNPC(assets.CityAnimals[3], 16, 22)  // Market Hen — market entrance

	// ── Place inscriptions ────────────────────────────────────────────────
	// Place inscriptions on walls (non-walkable adjacent tiles serve as the
	// inscription's "location"; we put them on walkable tiles near walls).
	insc := assets.CityInscriptions
	factory.NewInscription(w, insc[0], 40, 8)  // Tower north interior
	factory.NewInscription(w, insc[1], 40, 19) // Town square notice board
	factory.NewInscription(w, insc[2], 4, 3)   // Tavern entrance
	factory.NewInscription(w, insc[3], 63, 23) // Store entrance
	factory.NewInscription(w, insc[4], 4, 23)  // Market entrance
	factory.NewInscription(w, insc[5], 30, 17) // Town square memorial

	// ── Atmospheric furniture (IsRepeatable = true) ───────────────────────
	placeFurniture := func(glyph, name, desc string, x, y int) {
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

	// Tavern furniture
	placeFurniture("🛋️", "Worn Bench", "A well-worn bench near the fire. Many an adventurer has slept here.", 5, 10)
	placeFurniture("🕯️", "Tavern Candle", "A beeswax candle flickering on a sticky table. The wax has names carved in it.", 13, 5)
	placeFurniture("🫗", "Ale Tap", "A large barrel tap dripping something amber. It smells remarkable.", 14, 9)
	placeFurniture("🎭", "Dartboard", "A battered dartboard. Someone has drawn a face on it. It looks familiar.", 15, 4)

	// Church furniture
	placeFurniture("⚰️", "Stone Pew", "A cold stone pew. The kneeling impressions are worn deep.", 63, 10)
	placeFurniture("🕯️", "Church Candle", "A tall ceremonial candle. The Eternal Flame, in miniature.", 74, 10)
	placeFurniture("🌺", "Flower Offering", "Fresh wildflowers left at the altar. Someone still cares.", 68, 11)
	placeFurniture("🔑", "Altar Key", "A ornate key hanging on the altar. The lock it fits hasn't been seen in years.", 70, 6)

	// Store furniture
	placeFurniture("🗃️", "Supply Shelf", "Shelves crammed with provisions. Yeva stocks more than she sells.", 63, 25)
	placeFurniture("📦", "Supply Crate", "A crate stamped 'HANDLE WITH CARE — FRAGILE'. It rattles faintly.", 75, 30)
	placeFurniture("📊", "Account Ledger", "Yeva's ledger. The numbers are large. The margins have strong opinions.", 74, 24)
	placeFurniture("🧯", "Fire Bucket", "A bucket of sand by the entrance. Reassuring, somehow.", 63, 32)

	// Town square
	placeFurniture("🌱", "Flower Bed", "A neat flower bed surrounding the square's central fountain. The flowers are defiant.", 32, 17)
	placeFurniture("🪞", "Town Fountain", "A circular fountain with a worn stone basin. The water is clear and cold.", 40, 17)
	placeFurniture("🌱", "Square Planter", "A large planter filled with hardy city shrubs. They've outlasted several mayors.", 47, 17)

	// Homes
	placeFurniture("🛋️", "Sitting Chair", "A worn armchair facing a cold hearth. Homely, in a weary way.", 23, 6)
	placeFurniture("📚", "Bookshelf", "A tall shelf of well-thumbed books. They look well-loved.", 56, 6)
	placeFurniture("🌱", "Window Plant", "A small plant on the windowsill. It's doing its best.", 22, 28)
	placeFurniture("🖼️", "Family Portrait", "A painted portrait of a family. Everyone is smiling except the dog.", 29, 27)

	return &Floor{
		Num:             0,
		World:           w,
		GMap:            gmap,
		Rng:             rng,
		SpawnX:          40,
		SpawnY:          21,
		StairsDownX:     stairsDownX,
		StairsDownY:     stairsDownY,
		StairsUpX:       -1, // no stairs up in the city (top of the world)
		StairsUpY:       -1,
		RespawnCooldown: -1,
		SafeZone:        true,
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
