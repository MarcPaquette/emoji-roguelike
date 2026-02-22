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
	Furniture    []FurnitureSpawn
}

// FurnitureSpawn describes one furniture piece to place.
type FurnitureSpawn struct {
	Entry FurnitureSpawnEntry
	X, Y  int
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

	// occupied tracks every position already claimed this pass so that no two
	// entities share a tile.
	type pt = [2]int
	occupied := make(map[pt]bool)
	pick := func(room gamemap.Rect) (int, int) {
		return pickFreeInRoom(room, cfg, occupied)
	}
	claim := func(x, y int) { occupied[pt{x, y}] = true }

	// Spawn the floor elite in a random placeable room (does not consume budget).
	if cfg.EliteEnemy != nil && len(placeable) > 0 {
		room := placeable[cfg.Rand.Intn(len(placeable))]
		x, y := pick(room)
		claim(x, y)
		result.Enemies = append(result.Enemies, EnemySpawn{Entry: *cfg.EliteEnemy, X: x, Y: y})
	}

	budget := cfg.EnemyBudget

	// Phase 1: guarantee one enemy in every placeable room (cheapest that fits budget).
	if len(cfg.EnemyTable) > 0 {
		for _, room := range placeable {
			aff := affordableEnemies(cfg.EnemyTable, budget)
			if len(aff) == 0 {
				break
			}
			entry := cheapestEntry(aff)
			x, y := pick(room)
			claim(x, y)
			result.Enemies = append(result.Enemies, EnemySpawn{Entry: entry, X: x, Y: y})
			budget -= entry.ThreatCost
		}
	}

	// Phase 2: spend remaining budget on random rooms/enemies (as before).
	for budget > 0 && len(cfg.EnemyTable) > 0 {
		if len(placeable) == 0 {
			break
		}
		room := placeable[cfg.Rand.Intn(len(placeable))]
		affordable := affordableEnemies(cfg.EnemyTable, budget)
		if len(affordable) == 0 {
			break
		}
		entry := affordable[cfg.Rand.Intn(len(affordable))]
		x, y := pick(room)
		claim(x, y)
		result.Enemies = append(result.Enemies, EnemySpawn{Entry: entry, X: x, Y: y})
		budget -= entry.ThreatCost
	}

	// Place items in random rooms.
	for i := 0; i < cfg.ItemCount && len(cfg.ItemTable) > 0; i++ {
		room := rooms[cfg.Rand.Intn(len(rooms))]
		entry := cfg.ItemTable[cfg.Rand.Intn(len(cfg.ItemTable))]
		x, y := pick(room)
		claim(x, y)
		result.Items = append(result.Items, ItemSpawn{Entry: entry, X: x, Y: y})
	}

	// Place equipment items in random rooms.
	for i := 0; i < cfg.EquipCount && len(cfg.EquipTable) > 0; i++ {
		room := rooms[cfg.Rand.Intn(len(rooms))]
		entry := cfg.EquipTable[cfg.Rand.Intn(len(cfg.EquipTable))]
		x, y := pick(room)
		claim(x, y)
		result.Equipment = append(result.Equipment, EquipSpawn{Entry: entry, X: x, Y: y})
	}

	// Place inscriptions, picking without replacement so no text repeats.
	pool := make([]string, len(cfg.InscriptionTexts))
	copy(pool, cfg.InscriptionTexts)
	cfg.Rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	count := min(cfg.InscriptionCount, len(pool))
	for i := 0; i < count; i++ {
		room := rooms[cfg.Rand.Intn(len(rooms))]
		x, y := pick(room)
		claim(x, y)
		result.Inscriptions = append(result.Inscriptions, InscriptionSpawn{Text: pool[i], X: x, Y: y})
	}

	// Place furniture in placeable rooms (skips player spawn and stairs rooms).
	if cfg.FurniturePerRoom > 0 && (len(cfg.CommonFurniture) > 0 || len(cfg.RareFurniture) > 0) {
		for _, room := range placeable {
			n := cfg.Rand.Intn(cfg.FurniturePerRoom) + 1
			for range n {
				var entry FurnitureSpawnEntry
				if len(cfg.RareFurniture) > 0 && cfg.Rand.Intn(100) < 15 {
					entry = cfg.RareFurniture[cfg.Rand.Intn(len(cfg.RareFurniture))]
				} else if len(cfg.CommonFurniture) > 0 {
					entry = cfg.CommonFurniture[cfg.Rand.Intn(len(cfg.CommonFurniture))]
				} else {
					continue
				}
				x, y := pick(room)
				claim(x, y)
				result.Furniture = append(result.Furniture, FurnitureSpawn{Entry: entry, X: x, Y: y})
			}
		}
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

// cheapestEntry returns the entry with the lowest ThreatCost from a non-empty slice.
func cheapestEntry(entries []EnemySpawnEntry) EnemySpawnEntry {
	best := entries[0]
	for _, e := range entries[1:] {
		if e.ThreatCost < best.ThreatCost {
			best = e
		}
	}
	return best
}

// pickFreeInRoom tries up to 20 times to find an unoccupied position inside
// room. If all attempts hit an occupied tile it falls back to any position
// (avoids an infinite loop in very crowded rooms).
func pickFreeInRoom(room gamemap.Rect, cfg *Config, occupied map[[2]int]bool) (int, int) {
	const maxAttempts = 20
	for range maxAttempts {
		x, y := randomInRoom(room, cfg)
		if !occupied[[2]int{x, y}] {
			return x, y
		}
	}
	return randomInRoom(room, cfg)
}

func randomInRoom(room gamemap.Rect, cfg *Config) (int, int) {
	// Shrink by 1 from each edge so nothing lands adjacent to a door.
	// Doors are carved just outside the room boundary (at Y1-1, Y2+1, X1-1, X2+1),
	// so the room's outermost row/column are the only door-adjacent positions.
	x1, y1 := room.X1+1, room.Y1+1
	x2, y2 := room.X2-1, room.Y2-1
	// Fall back to full room bounds for very small rooms.
	if x1 > x2 || y1 > y2 {
		x1, y1 = room.X1, room.Y1
		x2, y2 = room.X2, room.Y2
	}
	w := x2 - x1 + 1
	h := y2 - y1 + 1
	x := x1 + cfg.Rand.Intn(max(1, w))
	y := y1 + cfg.Rand.Intn(max(1, h))
	return x, y
}
