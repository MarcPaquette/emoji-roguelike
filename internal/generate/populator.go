package generate

import (
	"emoji-roguelike/internal/gamemap"
)

// SpawnPoint holds a world coordinate where an entity should appear.
type SpawnPoint struct {
	X, Y int
}

// InscriptionSpawn describes one wall-writing to place.
type InscriptionSpawn struct {
	Text string
	X, Y int
}

// PopulateResult is returned by Populate with entity spawn data.
type PopulateResult struct {
	Enemies      []EnemySpawn
	Items        []ItemSpawn
	Equipment    []EquipSpawn
	Inscriptions []InscriptionSpawn
}

// EnemySpawn describes one enemy to create.
type EnemySpawn struct {
	Entry EnemySpawnEntry
	X, Y  int
}

// ItemSpawn describes one item to create.
type ItemSpawn struct {
	Entry ItemSpawnEntry
	X, Y  int
}

// EquipSpawn describes one equipment item to create.
type EquipSpawn struct {
	Entry EquipSpawnEntry
	X, Y  int
}

// Populate places enemies and items in the generated rooms.
func Populate(gmap *gamemap.GameMap, cfg *Config) PopulateResult {
	var result PopulateResult

	// Skip first room (player spawn) and last room (stairs).
	rooms := gmap.Rooms
	if len(rooms) <= 2 {
		return result
	}
	placeable := rooms[1 : len(rooms)-1]

	// Place enemies within budget.
	budget := cfg.EnemyBudget
	for budget > 0 && len(cfg.EnemyTable) > 0 {
		// Pick a random room.
		if len(placeable) == 0 {
			break
		}
		room := placeable[cfg.Rand.Intn(len(placeable))]
		// Pick a random enemy that fits budget.
		affordable := affordableEnemies(cfg.EnemyTable, budget)
		if len(affordable) == 0 {
			break
		}
		entry := affordable[cfg.Rand.Intn(len(affordable))]
		x, y := randomInRoom(room, cfg)
		result.Enemies = append(result.Enemies, EnemySpawn{Entry: entry, X: x, Y: y})
		budget -= entry.ThreatCost
	}

	// Place items in random rooms.
	for i := 0; i < cfg.ItemCount && len(cfg.ItemTable) > 0; i++ {
		room := rooms[cfg.Rand.Intn(len(rooms))]
		entry := cfg.ItemTable[cfg.Rand.Intn(len(cfg.ItemTable))]
		x, y := randomInRoom(room, cfg)
		result.Items = append(result.Items, ItemSpawn{Entry: entry, X: x, Y: y})
	}

	// Place equipment items in random rooms.
	for i := 0; i < cfg.EquipCount && len(cfg.EquipTable) > 0; i++ {
		room := rooms[cfg.Rand.Intn(len(rooms))]
		entry := cfg.EquipTable[cfg.Rand.Intn(len(cfg.EquipTable))]
		x, y := randomInRoom(room, cfg)
		result.Equipment = append(result.Equipment, EquipSpawn{Entry: entry, X: x, Y: y})
	}

	// Place inscriptions, picking without replacement so no text repeats.
	pool := make([]string, len(cfg.InscriptionTexts))
	copy(pool, cfg.InscriptionTexts)
	cfg.Rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	count := min(cfg.InscriptionCount, len(pool))
	for i := 0; i < count; i++ {
		room := rooms[cfg.Rand.Intn(len(rooms))]
		x, y := randomInRoom(room, cfg)
		result.Inscriptions = append(result.Inscriptions, InscriptionSpawn{Text: pool[i], X: x, Y: y})
	}

	return result
}

func affordableEnemies(table []EnemySpawnEntry, budget int) []EnemySpawnEntry {
	var out []EnemySpawnEntry
	for _, e := range table {
		if e.ThreatCost <= budget {
			out = append(out, e)
		}
	}
	return out
}

func randomInRoom(room gamemap.Rect, cfg *Config) (int, int) {
	w := room.X2 - room.X1 + 1
	h := room.Y2 - room.Y1 + 1
	x := room.X1 + cfg.Rand.Intn(max(1, w))
	y := room.Y1 + cfg.Rand.Intn(max(1, h))
	return x, y
}
