package generate

import (
	"emoji-roguelike/internal/gamemap"
	"math/rand"
	"testing"
)

// makeRoomedMap builds a GameMap pre-populated with the given number of rooms.
func makeRoomedMap(rooms int) *gamemap.GameMap {
	gmap := gamemap.New(80, 40)
	for i := range rooms {
		x := 2 + i*10
		r := gamemap.Rect{X1: x, Y1: 2, X2: x + 6, Y2: 8}
		gmap.Rooms = append(gmap.Rooms, r)
		for y := r.Y1; y <= r.Y2; y++ {
			for rx := r.X1; rx <= r.X2; rx++ {
				gmap.Set(rx, y, gamemap.MakeFloor())
			}
		}
	}
	return gmap
}

// makeBaseConfig returns a Config with a simple enemy/item/equip table and no inscriptions.
func makeBaseConfig(budget, itemCount, equipCount int) *Config {
	return &Config{
		EnemyBudget: budget,
		ItemCount:   itemCount,
		EquipCount:  equipCount,
		EnemyTable: []EnemySpawnEntry{
			{Glyph: "🦀", ThreatCost: 2, MaxHP: 8},
			{Glyph: "👻", ThreatCost: 5, MaxHP: 12},
		},
		ItemTable:  []ItemSpawnEntry{{Glyph: "🧪", Name: "Hyperflask"}},
		EquipTable: []EquipSpawnEntry{{Glyph: "⚔️", Name: "Shard Blade", Slot: 4, BaseATK: 2}},
		Rand:       rand.New(rand.NewSource(42)),
	}
}

func TestPopulateNoop_TooFewRooms(t *testing.T) {
	// ≤ 2 rooms: Populate skips enemy placement and returns empty result.
	for rooms := range 3 {
		gmap := makeRoomedMap(rooms)
		cfg := makeBaseConfig(10, 3, 2)
		result := Populate(gmap, cfg)
		if len(result.Enemies) != 0 {
			t.Errorf("rooms=%d: expected 0 enemies, got %d", rooms, len(result.Enemies))
		}
	}
}

func TestPopulateBudgetRespected(t *testing.T) {
	// Total threat cost of spawned enemies must never exceed the budget.
	for _, budget := range []int{5, 10, 20, 40} {
		gmap := makeRoomedMap(6)
		cfg := makeBaseConfig(budget, 0, 0)
		cfg.Rand = rand.New(rand.NewSource(int64(budget)))
		result := Populate(gmap, cfg)

		total := 0
		for _, e := range result.Enemies {
			total += e.Entry.ThreatCost
		}
		if total > budget {
			t.Errorf("budget=%d: placed enemies total threat=%d exceeds budget", budget, total)
		}
	}
}

func TestPopulateItemCount(t *testing.T) {
	gmap := makeRoomedMap(5)
	cfg := makeBaseConfig(0, 4, 0)
	result := Populate(gmap, cfg)
	if len(result.Items) != 4 {
		t.Errorf("expected 4 items, got %d", len(result.Items))
	}
}

func TestPopulateEquipCount(t *testing.T) {
	gmap := makeRoomedMap(5)
	cfg := makeBaseConfig(0, 0, 3)
	result := Populate(gmap, cfg)
	if len(result.Equipment) != 3 {
		t.Errorf("expected 3 equipment items, got %d", len(result.Equipment))
	}
}

func TestPopulateInscriptionsNoRepeats(t *testing.T) {
	texts := []string{"text A", "text B", "text C", "text D", "text E"}
	gmap := makeRoomedMap(5)
	cfg := makeBaseConfig(0, 0, 0)
	cfg.InscriptionTexts = texts
	cfg.InscriptionCount = 3
	result := Populate(gmap, cfg)

	if len(result.Inscriptions) != 3 {
		t.Fatalf("expected 3 inscriptions, got %d", len(result.Inscriptions))
	}
	seen := make(map[string]bool)
	for _, ins := range result.Inscriptions {
		if seen[ins.Text] {
			t.Errorf("duplicate inscription text: %q", ins.Text)
		}
		seen[ins.Text] = true
	}
}

func TestPopulateInscriptionCountCappedByPool(t *testing.T) {
	// When requested count exceeds pool size, only pool-size inscriptions are placed.
	gmap := makeRoomedMap(5)
	cfg := makeBaseConfig(0, 0, 0)
	cfg.InscriptionTexts = []string{"only one"}
	cfg.InscriptionCount = 10
	result := Populate(gmap, cfg)
	if len(result.Inscriptions) != 1 {
		t.Errorf("expected 1 inscription (limited by pool size), got %d", len(result.Inscriptions))
	}
}

func TestAffordableEnemiesFilter(t *testing.T) {
	table := []EnemySpawnEntry{
		{Glyph: "A", ThreatCost: 2},
		{Glyph: "B", ThreatCost: 5},
		{Glyph: "C", ThreatCost: 10},
	}
	cases := []struct {
		budget int
		want   int
	}{
		{1, 0},  // none affordable
		{3, 1},  // only A (cost 2)
		{5, 2},  // A and B (costs 2, 5)
		{10, 3}, // all three
	}
	for _, tc := range cases {
		got := affordableEnemies(table, tc.budget)
		if len(got) != tc.want {
			t.Errorf("budget=%d: got %d affordable entries; want %d", tc.budget, len(got), tc.want)
		}
	}
}

func TestCheapestEntry(t *testing.T) {
	cases := []struct {
		name    string
		entries []EnemySpawnEntry
		wantCost int
	}{
		{"single entry", []EnemySpawnEntry{{Glyph: "A", ThreatCost: 5}}, 5},
		{"first is cheapest", []EnemySpawnEntry{{Glyph: "A", ThreatCost: 2}, {Glyph: "B", ThreatCost: 5}}, 2},
		{"last is cheapest", []EnemySpawnEntry{{Glyph: "A", ThreatCost: 8}, {Glyph: "B", ThreatCost: 3}}, 3},
		{"middle is cheapest", []EnemySpawnEntry{{Glyph: "A", ThreatCost: 5}, {Glyph: "B", ThreatCost: 1}, {Glyph: "C", ThreatCost: 9}}, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := cheapestEntry(tc.entries)
			if got.ThreatCost != tc.wantCost {
				t.Errorf("got ThreatCost=%d; want %d", got.ThreatCost, tc.wantCost)
			}
		})
	}
}

func TestPopulateEveryRoomGetsEnemy(t *testing.T) {
	// With enough budget, every placeable room should contain at least one enemy.
	// 5 rooms → 3 placeable (first=player spawn, last=stairs).
	gmap := makeRoomedMap(5)
	placeableCount := len(gmap.Rooms) - 2 // 3
	cfg := makeBaseConfig(100, 0, 0)      // budget far exceeds cost of filling all rooms
	result := Populate(gmap, cfg)

	if len(result.Enemies) < placeableCount {
		t.Errorf("expected at least %d enemies (one per placeable room), got %d", placeableCount, len(result.Enemies))
	}
}

func TestPopulateEveryRoomGetsEnemyTightBudget(t *testing.T) {
	// Budget exactly covers the cheapest enemy (cost 2) for each of 3 placeable rooms.
	// Enemy table has only cost-2 entries so every room should still fill.
	gmap := makeRoomedMap(5)
	placeableCount := len(gmap.Rooms) - 2 // 3
	cfg := &Config{
		EnemyBudget: placeableCount * 2, // exactly enough for one cheap enemy per room
		EnemyTable:  []EnemySpawnEntry{{Glyph: "🦀", ThreatCost: 2, MaxHP: 8}},
		Rand:        rand.New(rand.NewSource(42)),
	}
	result := Populate(gmap, cfg)
	if len(result.Enemies) < placeableCount {
		t.Errorf("expected at least %d enemies, got %d", placeableCount, len(result.Enemies))
	}
}

func TestPopulateEliteSpawned(t *testing.T) {
	// When EliteEnemy is set, exactly one instance of it must appear in the result.
	elite := &EnemySpawnEntry{Glyph: "💠", Name: "Shardmind", ThreatCost: 0, MaxHP: 20}
	gmap := makeRoomedMap(5)
	cfg := makeBaseConfig(0, 0, 0)
	cfg.EliteEnemy = elite
	result := Populate(gmap, cfg)

	eliteCount := 0
	for _, e := range result.Enemies {
		if e.Entry.Glyph == elite.Glyph {
			eliteCount++
		}
	}
	if eliteCount != 1 {
		t.Errorf("expected exactly 1 elite enemy (%s), got %d", elite.Glyph, eliteCount)
	}
}

func TestPopulateEliteNilNoExtra(t *testing.T) {
	// When EliteEnemy is nil, no elite should appear.
	gmap := makeRoomedMap(5)
	cfg := makeBaseConfig(0, 0, 0)
	cfg.EliteEnemy = nil
	result := Populate(gmap, cfg)
	// With budget=0 and no elite, there should be zero enemies.
	if len(result.Enemies) != 0 {
		t.Errorf("expected 0 enemies with nil elite and zero budget, got %d", len(result.Enemies))
	}
}
