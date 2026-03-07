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

// newChronolithsCityFloor builds the hand-crafted Anchorpoint city (95x50).
// Anchorpoint is a town built on a temporal stable-point anchored by a
// Chronolith. Outside the anchor's radius, time behaves differently.
// Inside: reliable, steady, almost oppressively normal.
// SafeZone is set true: no combat, no enemies, no AI ticks.
func newChronolithsCityFloor(rng *rand.Rand) *Floor {
	gmap := gamemap.New(95, 50)
	w := ecs.NewWorld()

	// ── Base terrain: fill everything with grass ──────────────────────────
	fillTile(gmap, 0, 0, 94, 49, gamemap.MakeGrass())

	// ── Amber zones at edges (non-walkable decorative strips) ────────────
	// North edge
	fillTile(gmap, 0, 0, 94, 1, gamemap.MakeWall())
	// South edge
	fillTile(gmap, 0, 48, 94, 49, gamemap.MakeWall())
	// West edge
	fillTile(gmap, 0, 0, 1, 49, gamemap.MakeWall())
	// East edge
	fillTile(gmap, 93, 0, 94, 49, gamemap.MakeWall())

	// ── Streets (cobblestone floor) ─────────────────────────────────────
	// Main N-S street from Chronolith south to market
	carveRect(gmap, 46, 10, 50, 45)
	// E-W main street
	carveRect(gmap, 2, 20, 92, 21)
	// Market plaza (south-center)
	carveRect(gmap, 30, 22, 65, 33)
	// Chronolith approach (north plaza)
	carveRect(gmap, 38, 12, 58, 18)
	// South market street
	carveRect(gmap, 20, 36, 75, 37)
	// West workshop alley
	carveRect(gmap, 8, 4, 8, 20)
	// East residence alley
	carveRect(gmap, 70, 4, 70, 20)
	// South alleys connecting to market
	carveRect(gmap, 30, 33, 30, 45)
	carveRect(gmap, 55, 33, 55, 45)

	// ── Chronolith (massive pillar, center-north ~47,6..49,8) ───────────
	// 3x3 wall structure — immovable temporal anchor
	for y := 6; y <= 8; y++ {
		for x := 47; x <= 49; x++ {
			gmap.Set(x, y, gamemap.MakeWall())
		}
	}
	// Approach path to Chronolith
	carveRect(gmap, 46, 9, 50, 11)

	// ── WESTERN QUARTER — Construct Workshops ───────────────────────────

	// Workshop A "Repair Bay" (x=3..13, y=3..9) — 2 rooms
	carveBuilding(gmap, 3, 3, 13, 9)
	wallV(gmap, 8, 3, 9)               // vertical: Main Bay | Parts Store
	gmap.Set(8, 6, gamemap.MakeDoor()) // Main Bay → Parts Store
	gmap.Set(8, 9, gamemap.MakeDoor()) // south entrance

	// Workshop B "Chrono Lab" (x=3..13, y=11..17) — 1 room
	carveBuilding(gmap, 3, 11, 13, 17)
	gmap.Set(8, 17, gamemap.MakeDoor()) // south entrance

	// Workshop C "Temporal Forge" (x=15..25, y=3..9) — 2 rooms
	carveBuilding(gmap, 15, 3, 25, 9)
	wallV(gmap, 20, 3, 9)               // vertical: Forge | Storage
	gmap.Set(20, 6, gamemap.MakeDoor()) // Forge → Storage
	gmap.Set(20, 9, gamemap.MakeDoor()) // south entrance

	// ── EASTERN QUARTER — Residences & Inn ──────────────────────────────

	// Residence A (x=72..82, y=3..9) — 2 rooms
	carveBuilding(gmap, 72, 3, 82, 9)
	wallV(gmap, 77, 3, 9)               // vertical: Living | Bedroom
	gmap.Set(77, 6, gamemap.MakeDoor()) // Living → Bedroom
	gmap.Set(77, 9, gamemap.MakeDoor()) // south entrance

	// Residence B (x=84..92, y=3..9) — 1 room
	carveBuilding(gmap, 84, 3, 92, 9)
	gmap.Set(88, 9, gamemap.MakeDoor()) // south entrance

	// Inn "The Steady Clock" (x=72..82, y=11..17) — 2 rooms
	carveBuilding(gmap, 72, 11, 82, 17)
	wallV(gmap, 77, 11, 17)              // vertical: Common | Private
	gmap.Set(77, 14, gamemap.MakeDoor()) // Common → Private
	gmap.Set(77, 17, gamemap.MakeDoor()) // south entrance

	// Watchtower (x=84..92, y=11..17) — 1 room
	carveBuilding(gmap, 84, 11, 92, 17)
	gmap.Set(88, 17, gamemap.MakeDoor()) // south entrance

	// ── SOUTHERN QUARTER — Market & Salvage ─────────────────────────────

	// Market Hall (x=20..30, y=38..45) — 2 stalls
	carveBuilding(gmap, 20, 38, 30, 45)
	wallV(gmap, 25, 38, 45)              // vertical: Stall A | Stall B
	gmap.Set(25, 42, gamemap.MakeDoor()) // between stalls
	gmap.Set(25, 38, gamemap.MakeDoor()) // north entrance

	// Salvage Shop "Korr's Temporal Salvage" (x=47..57, y=38..45) — 2 rooms
	carveBuilding(gmap, 47, 38, 57, 45)
	wallV(gmap, 52, 38, 45)              // vertical: Shop | Stockroom
	gmap.Set(52, 42, gamemap.MakeDoor()) // Shop → Stockroom
	gmap.Set(52, 38, gamemap.MakeDoor()) // north entrance

	// Settler Home South (x=34..43, y=38..45) — 2 rooms
	carveBuilding(gmap, 34, 38, 43, 45)
	wallV(gmap, 39, 38, 45)              // vertical: Living | Kitchen
	gmap.Set(39, 42, gamemap.MakeDoor()) // Living → Kitchen
	gmap.Set(39, 38, gamemap.MakeDoor()) // north entrance

	// ── Stairs down to Temporal Ruins dungeon ────────────────────────────
	stairsDownX, stairsDownY := 48, 34
	gmap.Set(stairsDownX, stairsDownY, gamemap.MakeStairsDown())

	// ── Rooms list (for findFreeSpawn) ───────────────────────────────────
	gmap.Rooms = []gamemap.Rect{
		// Workshop A
		{X1: 4, Y1: 4, X2: 7, Y2: 8},    // Main Bay
		{X1: 9, Y1: 4, X2: 12, Y2: 8},   // Parts Store
		// Workshop B
		{X1: 4, Y1: 12, X2: 12, Y2: 16}, // Chrono Lab
		// Workshop C
		{X1: 16, Y1: 4, X2: 19, Y2: 8},  // Forge
		{X1: 21, Y1: 4, X2: 24, Y2: 8},  // Storage
		// Residence A
		{X1: 73, Y1: 4, X2: 76, Y2: 8},  // Living
		{X1: 78, Y1: 4, X2: 81, Y2: 8},  // Bedroom
		// Residence B
		{X1: 85, Y1: 4, X2: 91, Y2: 8},  // Main Room
		// Inn
		{X1: 73, Y1: 12, X2: 76, Y2: 16}, // Common
		{X1: 78, Y1: 12, X2: 81, Y2: 16}, // Private
		// Watchtower
		{X1: 85, Y1: 12, X2: 91, Y2: 16}, // Interior
		// Chronolith approach
		{X1: 38, Y1: 12, X2: 58, Y2: 18}, // Chronolith area
		// Market plaza
		{X1: 30, Y1: 22, X2: 65, Y2: 33}, // Plaza
		// Market Hall
		{X1: 21, Y1: 39, X2: 24, Y2: 44}, // Stall A
		{X1: 26, Y1: 39, X2: 29, Y2: 44}, // Stall B
		// Salvage Shop
		{X1: 48, Y1: 39, X2: 51, Y2: 44}, // Shop
		{X1: 53, Y1: 39, X2: 56, Y2: 44}, // Stockroom
		// Settler Home South
		{X1: 35, Y1: 39, X2: 38, Y2: 44}, // Living
		{X1: 40, Y1: 39, X2: 42, Y2: 44}, // Kitchen
	}

	// ── Place NPCs ───────────────────────────────────────────────────────
	placeNPC := func(def assets.NPCDef, x, y int) {
		id := factory.NewNPC(w, def.Name, def.Glyph, component.NPCKind(def.Kind), def.Lines, x, y)
		// Attach movement schedule if one is defined.
		if sched, ok := assets.ChronolithsNPCSchedules[def.Name]; ok {
			mc := component.NPCMovement{
				MoveInterval: sched.MoveInterval,
			}
			for _, e := range sched.Entries {
				se := component.ScheduleEntry{
					StartTick: e.StartTick,
					Behavior:  component.MoveBehavior(e.Behavior),
					BoundsX1:  e.BoundsX1, BoundsY1: e.BoundsY1,
					BoundsX2:  e.BoundsX2, BoundsY2: e.BoundsY2,
					StandX:    e.StandX, StandY: e.StandY,
				}
				for _, wp := range e.Waypoints {
					se.Waypoints = append(se.Waypoints, component.Waypoint{X: wp[0], Y: wp[1]})
				}
				// For animal stationary entries with sentinel (-1,-1),
				// use the NPC's spawn position.
				if se.Behavior == component.MoveStationary && se.StandX < 0 {
					se.StandX = x
					se.StandY = y
				}
				mc.Schedule = append(mc.Schedule, se)
			}
			// Initialize to the correct schedule entry at tick 0.
			mc.ActiveIndex = 0
			mc.Behavior = mc.Schedule[0].Behavior
			if mc.Behavior == component.MoveWander {
				mc.BoundsX1 = mc.Schedule[0].BoundsX1
				mc.BoundsY1 = mc.Schedule[0].BoundsY1
				mc.BoundsX2 = mc.Schedule[0].BoundsX2
				mc.BoundsY2 = mc.Schedule[0].BoundsY2
			}
			w.Add(id, mc)
		}
	}

	// Named NPCs
	placeNPC(assets.ChronolithsNPCs[0], 12, 8)  // Construct 4471  — workshop A main bay
	placeNPC(assets.ChronolithsNPCs[1], 44, 14) // Chrono Keeper Essa — near Chronolith
	placeNPC(assets.ChronolithsNPCs[2], 50, 42) // Salvager Korr   — salvage shop
	placeNPC(assets.ChronolithsNPCs[3], 85, 8)  // Watcher Brin    — watchtower/residence B
	placeNPC(assets.ChronolithsNPCs[4], 72, 10) // Settler Mira    — near east residence
	// Animal
	placeNPC(assets.ChronolithsNPCs[5], 45, 25) // Temporal Hound  — market plaza

	// ── Place inscriptions ───────────────────────────────────────────────
	insc := assets.ChronolithsCityInscriptions
	factory.NewInscription(w, insc[0], 48, 12) // Chronolith approach
	factory.NewInscription(w, insc[1], 60, 22) // East side of plaza
	factory.NewInscription(w, insc[2], 48, 9)  // Near Chronolith
	factory.NewInscription(w, insc[3], 40, 28) // Plaza west
	factory.NewInscription(w, insc[4], 52, 36) // Near salvage shop

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

	// Trees — parks and borders
	for _, pos := range [][2]int{
		// West park (grass between workshops and plaza)
		{5, 20}, {14, 22}, {10, 25}, {18, 28}, {6, 30}, {22, 32},
		// East park (grass between residences and plaza)
		{75, 22}, {82, 25}, {78, 28}, {88, 30}, {80, 33}, {85, 22},
		// Around Chronolith
		{40, 5}, {56, 5}, {38, 10}, {58, 10},
		// South border area
		{10, 46}, {25, 46}, {65, 46}, {80, 46},
	} {
		pf("🌳", "Weathered Tree",
			"A tree that has stood in one time long enough to put down serious roots. The leaves never change season.",
			pos[0], pos[1])
	}

	// ── Chronolith area ─────────────────────────────────────────────────
	pf("⏳", "Temporal Marker",
		"A small amber obelisk marking the edge of the Chronolith's stable radius. Beyond this point, time has opinions.",
		42, 13)
	pf("⏳", "Temporal Marker",
		"An amber boundary stone. The air on one side feels different from the air on the other. Both sides insist they are correct.",
		54, 13)

	// ── Workshop A — Main Bay (x=4..7, y=4..8) ──────────────────────────
	pf("🔩", "Construct Parts",
		"Bins of gears, pistons, and temporal regulators. Some of the parts are from machines that haven't been built yet.",
		5, 6)
	pf("🛠️", "Repair Bench",
		"A heavy workbench scarred by centuries of Construct maintenance. The vise is adjusted to grip a mechanical arm.",
		6, 4)

	// ── Workshop A — Parts Store (x=9..12, y=4..8) ──────────────────────
	pf("🗃️", "Parts Catalogue",
		"A filing cabinet of Construct schematics. The drawer labelled TIMELINE SEVEN is locked. The lock is from Timeline Eight.",
		10, 6)

	// ── Workshop B — Chrono Lab (x=4..12, y=12..16) ─────────────────────
	pf("🔬", "Temporal Microscope",
		"A microscope that observes things at different speeds. The current slide shows a dust mote that has been falling for three centuries.",
		6, 14)
	pf("🧫", "Time Sample",
		"A sealed dish containing a sample of frozen time. It glows faintly amber. The label reads: DO NOT OBSERVE.",
		10, 14)

	// ── Workshop C — Forge (x=16..19, y=4..8) ───────────────────────────
	pf("🔧", "Calibration Tools",
		"Precision instruments for adjusting temporal mechanisms. The smallest wrench is designed for a gear tooth that exists in two moments.",
		17, 6)

	// ── Workshop C — Storage (x=21..24, y=4..8) ─────────────────────────
	pf("📦", "Salvage Crate",
		"Crates of temporal debris awaiting sorting. Some items are older than they will be.",
		22, 6)

	// ── Market plaza ─────────────────────────────────────────────────────
	pf("⛲", "Anchor Fountain",
		"A fountain at the plaza's center. The water falls at exactly the correct speed. This is more remarkable than it sounds.",
		48, 28)
	pf("📋", "Notice Board",
		"A board of civic notices. TEMPORAL ADVISORY: Daylight saving is not optional within city limits. Violators will be synchronized.",
		35, 24)
	pf("🛋️", "Stone Bench",
		"A bench facing the fountain. The stone is warm regardless of season. The Chronolith keeps everything steady, even furniture.",
		42, 30)
	pf("🛋️", "Stone Bench",
		"Another bench, this one facing east toward the badlands. Watchers sometimes sit here on their breaks, staring at nothing.",
		55, 30)

	// ── Residence A — Living (x=73..76, y=4..8) ─────────────────────────
	pf("🪴", "House Plant",
		"A potted fern that has been alive for exactly the right amount of time. Not a second more. Not a second less.",
		74, 6)

	// ── Residence A — Bedroom (x=78..81, y=4..8) ────────────────────────
	pf("🖼️", "Family Portrait",
		"A portrait of a family. The grandmother looks younger than the grandchild. Nobody in the painting seems bothered.",
		80, 6)

	// ── Residence B — Main Room (x=85..91, y=4..8) ──────────────────────
	pf("📚", "Bookshelf",
		"History books arranged chronologically. The chronology is debatable. Three volumes are dated before their own copyright.",
		88, 6)

	// ── Inn — Common Room (x=73..76, y=12..16) ──────────────────────────
	pf("🕯️", "Inn Candle",
		"A candle that burns at a perfectly regulated rate. The innkeeper is very proud of this. Time-stable flames are surprisingly rare.",
		74, 14)
	pf("🫗", "Ale Tap",
		"The inn's tap serves a brew called Anchor Ale. It always tastes like it was poured three seconds ago, regardless of when it was poured.",
		75, 12)

	// ── Inn — Private Room (x=78..81, y=12..16) ─────────────────────────
	pf("🗂️", "Guest Ledger",
		"The inn's guest register. Some guests have checked in after they checked out. The innkeeper has stopped questioning this.",
		79, 14)

	// ── Watchtower (x=85..91, y=12..16) ──────────────────────────────────
	pf("🔭", "Observation Scope",
		"A brass telescope pointed at the badlands. Through it, you can see frozen moments glinting like amber in the sun.",
		88, 14)
	pf("📡", "Temporal Sensor",
		"A device that measures temporal stability. The needle points firmly to STABLE. It does this with suspicious confidence.",
		87, 12)

	// ── Market Hall — Stall A (x=21..24, y=39..44) ──────────────────────
	pf("🪙", "Temporal Coin",
		"A coin frozen mid-flip in a display case. Both faces show the same result. The result changes when you look away.",
		22, 42)

	// ── Market Hall — Stall B (x=26..29, y=39..44) ──────────────────────
	pf("🌾", "Grain Sacks",
		"Grain from the stable zone. It grows at the correct rate, which makes it exotic in the Chronoliths.",
		27, 42)

	// ── Salvage Shop — Shop Floor (x=48..51, y=39..44) ──────────────────
	pf("🗃️", "Salvage Shelf",
		"Shelves of temporal salvage. Everything here was recovered from the badlands. Some items insist they were never lost.",
		49, 42)

	// ── Salvage Shop — Stockroom (x=53..56, y=39..44) ───────────────────
	pf("📦", "Sealed Crate",
		"A crate from Timeline Seven. The shipping label is dated to a year that hasn't happened. The contents are perfectly preserved.",
		54, 42)

	// ── Settler Home — Living (x=35..38, y=39..44) ──────────────────────
	pf("🛋️", "Armchair",
		"A comfortable chair positioned facing the window. From here you can see the Chronolith. It is always exactly where it was.",
		36, 42)

	// ── Settler Home — Kitchen (x=40..42, y=39..44) ─────────────────────
	pf("🫕", "Stew Pot",
		"A pot of stew that takes exactly as long to cook as the recipe says. In the Chronoliths, this is a luxury.",
		41, 42)

	// ── Return portal to Emberveil ───────────────────────────────────────
	// Small waypoint structure near the south market area.
	returnPortalX, returnPortalY := 65, 36
	gmap.Set(returnPortalX, returnPortalY, gamemap.MakeStairsDown())
	pf("⏰", "Return Gate",
		"A portal tuned to Emberveil's temporal frequency. Step through to return to the Prismatic Spire's city.",
		66, 36)
	factory.NewInscription(w, "TEMPORAL TRANSIT: Emberveil. Step onto the stairs to return to the Prismatic Spire.", 64, 36)

	return &Floor{
		Num:             100,
		World:           w,
		GMap:            gmap,
		Rng:             rng,
		SpawnX:          48,
		SpawnY:          18,
		StairsDownX:     stairsDownX,
		StairsDownY:     stairsDownY,
		StairsUpX:       -1, // no stairs up in the city
		StairsUpY:       -1,
		RespawnCooldown: -1,
		SafeZone:        true,
		Portals: map[[2]int]int{
			{stairsDownX, stairsDownY}:         101, // dungeon stairs → Temporal Ruins floor 1
			{returnPortalX, returnPortalY}: 0,   // return portal → Emberveil
		},
	}
}
