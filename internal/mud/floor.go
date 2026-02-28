package mud

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/factory"
	"emoji-roguelike/internal/gamemap"
	"emoji-roguelike/internal/generate"
	"fmt"
	"math"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

const MaxFloors = 10

// Floor holds the shared game state for one dungeon level.
// All players on the same floor share one World and GameMap.
type Floor struct {
	Num   int
	World *ecs.World
	GMap  *gamemap.GameMap
	Rng   *rand.Rand

	// SpawnX/SpawnY is the default player spawn point (first room center).
	SpawnX, SpawnY int

	// StairsDownX/Y is where the down stairs tile is placed (last room center).
	StairsDownX, StairsDownY int

	// StairsUpX/Y is where the up stairs tile is placed (first room center).
	// Both are -1 for floor 0 (city) which has no up stairs.
	StairsUpX, StairsUpY int

	// RespawnCooldown drives enemy wave respawning.
	// -1 = idle (enemies present or cooldown not started)
	//  N > 0 = ticks remaining before wave spawns
	//  0 = spawn wave this tick
	RespawnCooldown int

	// SafeZone disables combat and AI ticking (used for the starting city).
	SafeZone bool
}

// newFloor generates a fresh dungeon floor using the same level config as the
// single-player and coop modes.
func newFloor(num int, rng *rand.Rand) *Floor {
	cfg := levelConfig(num, rng)
	gmap, px, py := generate.Generate(cfg)
	w := ecs.NewWorld()

	pop := generate.Populate(gmap, cfg)
	for _, es := range pop.Enemies {
		factory.NewEnemy(w, es.Entry, es.X, es.Y)
	}
	for _, is := range pop.Items {
		factory.NewItem(w, is.Entry, is.X, is.Y)
	}
	for _, eq := range pop.Equipment {
		factory.NewEquipItem(w, eq.Entry, num, rng, eq.X, eq.Y)
	}
	for _, ins := range pop.Inscriptions {
		factory.NewInscription(w, ins.Text, ins.X, ins.Y)
	}
	for _, fs := range pop.Furniture {
		factory.NewFurniture(w, fs.Entry, fs.X, fs.Y)
	}

	// Derive stair positions from generated rooms.
	stairsDownX, stairsDownY := px, py // fallback if only one room
	if len(gmap.Rooms) > 1 {
		last := gmap.Rooms[len(gmap.Rooms)-1]
		stairsDownX, stairsDownY = last.Center()
	}
	stairsUpX, stairsUpY := -1, -1 // -1 = not present (floor 0 / city)
	if num > 0 {
		stairsUpX, stairsUpY = px, py // stairs up shares spawn tile
	}
	// Floor 1 in MUD: manually place stairs-up tile so players can return to the city.
	// (bsp.go places stairs-up only for num > 1; floor 1 needs special handling here.)
	if num == 1 {
		gmap.Set(px, py, gamemap.MakeStairsUp())
	}

	return &Floor{
		Num:             num,
		World:           w,
		GMap:            gmap,
		Rng:             rng,
		SpawnX:          px,
		SpawnY:          py,
		StairsDownX:     stairsDownX,
		StairsDownY:     stairsDownY,
		StairsUpX:       stairsUpX,
		StairsUpY:       stairsUpY,
		RespawnCooldown: -1,
	}
}

// levelConfig mirrors the single-player levelConfig from game/levels.go.
func levelConfig(floor int, rng *rand.Rand) *generate.Config {
	t := 0.0
	if MaxFloors > 1 {
		t = float64(floor-1) / float64(MaxFloors-1)
	}
	return &generate.Config{
		MapWidth:         lerpi(40, 90, t),
		MapHeight:        lerpi(20, 50, t),
		MinLeafSize:      8,
		MaxLeafSize:      lerpi(20, 10, t),
		SplitRatio:       0.5,
		MinRoomSize:      4,
		RoomPadding:      1,
		CorridorStyle:    generate.CorridorLShaped,
		FloorNumber:      floor,
		EnemyBudget:      lerpi(5, 55, t),
		ItemCount:        lerpi(3, 8, t),
		EquipCount:       lerpi(1, 3, t),
		EnemyTable:       assets.EnemyTables[floor],
		ItemTable:        itemTableForFloor(floor),
		EquipTable:       assets.EquipTablesForFloor(floor),
		InscriptionTexts: assets.WallWritings[floor],
		InscriptionCount: 2 + rng.Intn(4),
		EliteEnemy:       assets.FloorElite(floor),
		CommonFurniture:  assets.FurnitureByFloor[floor].Common,
		RareFurniture:    assets.FurnitureByFloor[floor].Rare,
		FurniturePerRoom: 2,
		Rand:             rng,
	}
}

func itemTableForFloor(floor int) []generate.ItemSpawnEntry {
	base := assets.ItemTables[floor]
	var extra []generate.ItemSpawnEntry
	if floor >= 3 {
		extra = append(extra, generate.ItemSpawnEntry{Glyph: assets.GlyphResonanceBurst, Name: "Resonance Burst"})
	}
	if floor >= 5 {
		extra = append(extra, generate.ItemSpawnEntry{Glyph: assets.GlyphNanoSyringe, Name: "Nano-Syringe"})
	}
	if floor >= 6 {
		extra = append(extra, generate.ItemSpawnEntry{Glyph: assets.GlyphPhaseRod, Name: "Phase Rod"})
	}
	if floor >= 8 {
		extra = append(extra, generate.ItemSpawnEntry{Glyph: assets.GlyphApexCore, Name: "Apex Core"})
	}
	if len(extra) == 0 {
		return base
	}
	combined := make([]generate.ItemSpawnEntry, len(base)+len(extra))
	copy(combined, base)
	copy(combined[len(base):], extra)
	return combined
}

func lerpi(a, b int, t float64) int {
	return int(math.Round(float64(a) + t*float64(b-a)))
}

// globalMessage broadcasts a message to all sessions regardless of floor.
func globalMessage(sessions []*Session, msg string) {
	for _, s := range sessions {
		s.AddMessage(msg)
	}
}

// floorMessage broadcasts to all sessions on the given floor.
func floorMessage(sessions []*Session, floorNum int, msg string) {
	for _, s := range sessions {
		if s.FloorNum == floorNum {
			s.AddMessage(msg)
		}
	}
}

// drawDeathScreen renders a simple "You died" overlay to the session's screen.
func drawDeathScreen(sess *Session, ticksLeft int) {
	sess.Screen.Clear()
	w, h := sess.Screen.Size()
	style := tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
	dim := tcell.StyleDefault.Foreground(tcell.ColorGray)
	msgs := []string{
		"💀 You have fallen! 💀",
		fmt.Sprintf("Respawning in %d...", ticksLeft),
	}
	for i, msg := range msgs {
		x := (w - len([]rune(msg))) / 2
		y := h/2 - 1 + i
		if x < 0 {
			x = 0
		}
		st := style
		if i > 0 {
			st = dim
		}
		for _, r := range msg {
			if x >= w {
				break
			}
			sess.Screen.SetContent(x, y, r, nil, st)
			x++
		}
	}
	sess.Screen.Show()
}

// ColorName returns a human-readable name for the color (for display purposes).
func ColorName(c tcell.Color) string {
	switch c {
	case tcell.ColorYellow:
		return "yellow"
	case tcell.ColorFuchsia:
		return "fuchsia"
	case tcell.ColorAqua:
		return "cyan"
	case tcell.ColorLime:
		return "lime"
	case tcell.ColorOrange:
		return "orange"
	case tcell.ColorRed:
		return "red"
	case tcell.ColorSilver:
		return "silver"
	default:
		return "white"
	}
}
