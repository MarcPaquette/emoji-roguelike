package mud

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"math/rand"
	"testing"
)

// ─── City floor structure ─────────────────────────────────────────────────────

func TestCityFloorSafeZone(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	floor := newCityFloor(rng)
	if !floor.SafeZone {
		t.Error("city floor should have SafeZone=true")
	}
}

func TestCityFloorNoEnemies(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	floor := newCityFloor(rng)
	enemies := floor.World.Query(component.CAI)
	if len(enemies) != 0 {
		t.Errorf("city floor should have no enemies, got %d", len(enemies))
	}
}

func TestCityFloorHasStairsDownNoStairsUp(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	floor := newCityFloor(rng)

	foundDown, foundUp := false, false
	for y := range floor.GMap.Height {
		for x := range floor.GMap.Width {
			switch floor.GMap.At(x, y).Kind {
			case 4: // TileStairsDown
				foundDown = true
			case 3: // TileStairsUp
				foundUp = true
			}
		}
	}
	if !foundDown {
		t.Error("city floor should have stairs down")
	}
	if foundUp {
		t.Error("city floor should not have stairs up (it is the top level)")
	}
	if floor.StairsUpX != -1 || floor.StairsUpY != -1 {
		t.Errorf("city floor StairsUpX/Y should be -1, got (%d,%d)", floor.StairsUpX, floor.StairsUpY)
	}
}

func TestCityFloorHasNPCs(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	floor := newCityFloor(rng)
	npcs := floor.World.Query(component.CNPC)
	if len(npcs) == 0 {
		t.Error("city floor should have NPCs")
	}
}

func TestCityFloorRoomsPopulated(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	floor := newCityFloor(rng)
	if len(floor.GMap.Rooms) == 0 {
		t.Error("city floor should have rooms for findFreeSpawn")
	}
}

func TestCityFloorSpawnIsWalkable(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	floor := newCityFloor(rng)
	if !floor.GMap.IsWalkable(floor.SpawnX, floor.SpawnY) {
		t.Errorf("city spawn point (%d,%d) is not walkable", floor.SpawnX, floor.SpawnY)
	}
}

// ─── NPC interaction ──────────────────────────────────────────────────────────

// makeTestSessionOnCity creates a server with a city floor and a session on it.
func makeTestSessionOnCity(t *testing.T) (*Server, *Session) {
	t.Helper()
	rng := rand.New(rand.NewSource(42))
	srv := &Server{
		floors: map[int]*Floor{},
		rng:    rng,
	}
	srv.floors[0] = newCityFloor(rand.New(rand.NewSource(1)))
	sess := newTestSession(0, srv)

	srv.mu.Lock()
	srv.sessions = append(srv.sessions, sess)
	srv.spawnPlayerLocked(sess, 0)
	srv.mu.Unlock()
	return srv, sess
}

func TestNPCInteractDialogue(t *testing.T) {
	srv, sess := makeTestSessionOnCity(t)
	rng := rand.New(rand.NewSource(7))
	floor := srv.floors[0]

	npc := component.NPC{
		Name:  "Test NPC",
		Kind:  component.NPCKindDialogue,
		Lines: []string{"Hello, traveler."},
	}
	npcID := floor.World.CreateEntity()
	floor.World.Add(npcID, component.Position{X: 10, Y: 10})
	floor.World.Add(npcID, npc)

	srv.mu.Lock()
	floor.Rng = rng
	srv.interactNPCLocked(floor, sess, npcID, npc)
	srv.mu.Unlock()

	if len(sess.Messages) == 0 {
		t.Fatal("expected a message from NPC dialogue")
	}
}

func TestNPCInteractHealerPartialHP(t *testing.T) {
	srv, sess := makeTestSessionOnCity(t)
	floor := srv.floors[0]

	// Reduce player HP.
	srv.mu.Lock()
	hp := floor.World.Get(sess.PlayerID, component.CHealth).(component.Health)
	hp.Current = 5
	floor.World.Add(sess.PlayerID, hp)
	srv.mu.Unlock()

	npc := component.NPC{
		Name:  "Sister Maris",
		Kind:  component.NPCKindHealer,
		Lines: []string{"Be healed."},
	}
	npcID := floor.World.CreateEntity()

	srv.mu.Lock()
	floor.Rng = rand.New(rand.NewSource(1))
	srv.interactNPCLocked(floor, sess, npcID, npc)
	hpAfter := floor.World.Get(sess.PlayerID, component.CHealth).(component.Health)
	srv.mu.Unlock()

	if hpAfter.Current != hpAfter.Max {
		t.Errorf("healer should restore HP to max, got %d/%d", hpAfter.Current, hpAfter.Max)
	}
}

func TestNPCInteractHealerFullHP(t *testing.T) {
	srv, sess := makeTestSessionOnCity(t)
	floor := srv.floors[0]

	// Ensure player is at full HP.
	srv.mu.Lock()
	hp := floor.World.Get(sess.PlayerID, component.CHealth).(component.Health)
	srv.mu.Unlock()
	if hp.Current != hp.Max {
		t.Skip("player not at full HP, skip test")
	}

	prevMessages := len(sess.Messages)
	npc := component.NPC{
		Name:  "Sister Maris",
		Kind:  component.NPCKindHealer,
		Lines: []string{"You look healthy."},
	}
	npcID := floor.World.CreateEntity()

	srv.mu.Lock()
	floor.Rng = rand.New(rand.NewSource(1))
	srv.interactNPCLocked(floor, sess, npcID, npc)
	hpAfter := floor.World.Get(sess.PlayerID, component.CHealth).(component.Health)
	srv.mu.Unlock()

	// HP unchanged and a message was added.
	if hpAfter.Current != hp.Max {
		t.Errorf("full-HP healer should not change HP, got %d", hpAfter.Current)
	}
	if len(sess.Messages) <= prevMessages {
		t.Error("expected a message when healer is interacted at full HP")
	}
}

func TestNPCInteractShopSetsPendingNPC(t *testing.T) {
	srv, sess := makeTestSessionOnCity(t)
	floor := srv.floors[0]

	npc := component.NPC{
		Name:  "Merchant Yeva",
		Kind:  component.NPCKindShop,
		Lines: []string{"Welcome!"},
	}
	npcID := floor.World.CreateEntity()

	srv.mu.Lock()
	floor.Rng = rand.New(rand.NewSource(1))
	srv.interactNPCLocked(floor, sess, npcID, npc)
	pending := sess.PendingNPC
	srv.mu.Unlock()

	if pending == ecs.NilEntity {
		t.Error("shop NPC interaction should set PendingNPC != NilEntity")
	}
}

// ─── Gold system ──────────────────────────────────────────────────────────────

func TestGoldDropOnKill(t *testing.T) {
	srv := newTestServer()
	sess := newTestSession(0, srv)
	srv.AddSession(sess)

	// Transition to floor 1 so we can fight enemies.
	srv.mu.Lock()
	srv.transitionFloorLocked(sess, 1)
	floor1 := srv.floors[1]
	srv.mu.Unlock()

	// Add a weak enemy adjacent to player.
	srv.mu.Lock()
	posComp := floor1.World.Get(sess.PlayerID, component.CPosition).(component.Position)
	enemyID := floor1.World.CreateEntity()
	floor1.World.Add(enemyID, component.Position{X: posComp.X + 1, Y: posComp.Y})
	floor1.World.Add(enemyID, component.Health{Current: 1, Max: 1})
	floor1.World.Add(enemyID, component.Combat{Attack: 0, Defense: 0})
	floor1.World.Add(enemyID, component.AI{Behavior: component.BehaviorChase, SightRange: 5})
	floor1.World.Add(enemyID, component.Renderable{Glyph: "🦀"})
	floor1.World.Add(enemyID, component.TagBlocking{})
	floor1.World.Add(enemyID, component.Effects{})
	goldBefore := sess.Gold
	srv.processActionLocked(sess, ActionMoveE)
	goldAfter := sess.Gold
	srv.mu.Unlock()

	if goldAfter <= goldBefore {
		t.Errorf("killing enemy should award gold; before=%d after=%d", goldBefore, goldAfter)
	}
}

func TestRespawnResetsGoldAndReturnsToCity(t *testing.T) {
	srv := newTestServer()
	sess := newTestSession(0, srv)
	srv.AddSession(sess)

	// Give the session some gold.
	srv.mu.Lock()
	sess.Gold = 42
	sess.FloorNum = 3 // simulate being on a deep floor
	// Also create an entity so respawnLocked has something to clean up.
	srv.mu.Unlock()

	// Simulate respawn (which resets gold and returns to floor 0).
	srv.mu.Lock()
	// Create a dummy player entity on floor 3 for cleanup.
	if _, ok := srv.floors[3]; !ok {
		srv.floors[3] = newOpenFloor(3)
	}
	pid := srv.floors[3].World.CreateEntity()
	sess.PlayerID = pid
	srv.respawnLocked(sess)
	srv.mu.Unlock()

	if sess.Gold != 0 {
		t.Errorf("respawn should reset gold to 0, got %d", sess.Gold)
	}
	if sess.FloorNum != 0 {
		t.Errorf("respawn should return player to floor 0 (city), got %d", sess.FloorNum)
	}
}

// ─── Safe zone ────────────────────────────────────────────────────────────────

func TestSafeZonePreventsAttack(t *testing.T) {
	srv := newTestServer()
	sess := newTestSession(0, srv)
	srv.AddSession(sess)

	// Player is already on floor 0 (city, safe zone).
	floor := srv.floors[0]

	// Place a dummy enemy on the city floor (shouldn't normally happen, but for testing).
	srv.mu.Lock()
	posComp := floor.World.Get(sess.PlayerID, component.CPosition).(component.Position)
	enemyID := floor.World.CreateEntity()
	floor.World.Add(enemyID, component.Position{X: posComp.X + 1, Y: posComp.Y})
	floor.World.Add(enemyID, component.Health{Current: 10, Max: 10})
	floor.World.Add(enemyID, component.Combat{Attack: 0, Defense: 0})
	floor.World.Add(enemyID, component.TagBlocking{})
	floor.World.Add(enemyID, component.Effects{})
	floor.World.Add(enemyID, component.Renderable{Glyph: "🦀"})

	prevHP := floor.World.Get(enemyID, component.CHealth).(component.Health).Current
	srv.processActionLocked(sess, ActionMoveE)
	afterHP := floor.World.Get(enemyID, component.CHealth).(component.Health).Current
	srv.mu.Unlock()

	if afterHP != prevHP {
		t.Error("combat should be prevented in safe zone; enemy HP changed")
	}
}

func TestTickFloorSafeZoneSkipsAI(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	floor := newCityFloor(rng)
	floor.Rng = rand.New(rand.NewSource(42))

	screen := newSimScreen()
	sess := NewSession(0, "Player", playerColors[0], screen)
	sess.Class = assets.Classes[0]
	sess.FovRadius = sess.Class.FOVRadius
	sess.BaseMaxHP = sess.Class.MaxHP
	sess.FloorNum = 0

	// Add player entity to city floor.
	pid := floor.World.CreateEntity()
	floor.World.Add(pid, component.Position{X: 40, Y: 21})
	floor.World.Add(pid, component.Health{Current: 20, Max: 20})
	floor.World.Add(pid, component.Combat{Attack: 5, Defense: 2})
	floor.World.Add(pid, component.Effects{})
	floor.World.Add(pid, component.TagPlayer{})
	floor.World.Add(pid, component.TagBlocking{})
	sess.PlayerID = pid

	srv := &Server{
		floors:   map[int]*Floor{0: floor},
		sessions: []*Session{sess},
		rng:      rng,
	}

	hpBefore := floor.World.Get(pid, component.CHealth).(component.Health).Current

	srv.mu.Lock()
	srv.tickFloorLocked(floor)
	srv.mu.Unlock()

	hpAfter := floor.World.Get(pid, component.CHealth).(component.Health).Current
	// In safe zone, no AI means no damage (no enemies anyway, but SafeZone should skip AI path).
	if hpAfter < hpBefore {
		t.Errorf("safe zone tick should not reduce player HP; before=%d after=%d", hpBefore, hpAfter)
	}
}

// ─── Ascend from floor 1 returns to city ─────────────────────────────────────

func TestAscendFromFloor1ReturnsToCity(t *testing.T) {
	srv := newTestServer()
	sess := newTestSession(0, srv)
	srv.AddSession(sess)

	// Transition to floor 1.
	srv.mu.Lock()
	srv.transitionFloorLocked(sess, 1)
	floor1 := srv.floors[1]

	// Move player onto the stairs-up tile.
	floor1.World.Add(sess.PlayerID, component.Position{X: floor1.StairsUpX, Y: floor1.StairsUpY})

	// Perform ascend action.
	srv.processActionLocked(sess, ActionAscend)
	srv.mu.Unlock()

	if sess.FloorNum != 0 {
		t.Errorf("ascending from floor 1 should return to floor 0 (city), got floor %d", sess.FloorNum)
	}
}

// ─── Shop buying ──────────────────────────────────────────────────────────────

func TestShopBuyDeductsGold(t *testing.T) {
	srv, sess := makeTestSessionOnCity(t)

	// Give the session enough gold.
	sess.Gold = 100

	entry := assets.ShopCatalogue[0] // first item in shop
	priceExpected := entry.Price
	msg := srv.shopBuy(sess, assets.ShopCatalogue, 0)

	if sess.Gold != 100-priceExpected {
		t.Errorf("expected gold=%d after buy, got %d (msg: %s)", 100-priceExpected, sess.Gold, msg)
	}

	// Verify item is in backpack.
	srv.mu.Lock()
	floor := srv.floors[0]
	inv := floor.World.Get(sess.PlayerID, component.CInventory).(component.Inventory)
	srv.mu.Unlock()
	found := false
	for _, item := range inv.Backpack {
		if item.Name == entry.Name {
			found = true
		}
	}
	if !found {
		t.Errorf("bought item %q not found in backpack", entry.Name)
	}
}

func TestShopBuyInsufficientGold(t *testing.T) {
	srv, sess := makeTestSessionOnCity(t)
	sess.Gold = 0

	msg := srv.shopBuy(sess, assets.ShopCatalogue, 0)
	if sess.Gold != 0 {
		t.Error("gold should not change when purchase fails")
	}
	_ = msg // message content not asserted — just that it doesn't panic
}

// ─── Start items on city floor ────────────────────────────────────────────────

func TestSpawnStartItemsOnCityFloor(t *testing.T) {
	// Use the Symbiont class which has StartItems.
	var symbiont assets.ClassDef
	for _, c := range assets.Classes {
		if c.ID == "symbiont" {
			symbiont = c
			break
		}
	}
	if len(symbiont.StartItems) == 0 {
		t.Skip("symbiont has no start items")
	}

	rng := rand.New(rand.NewSource(42))
	srv := &Server{
		floors: map[int]*Floor{},
		rng:    rng,
	}
	srv.floors[0] = newCityFloor(rand.New(rand.NewSource(1)))

	screen := newSimScreen()
	sess := NewSession(0, "Test", playerColors[0], screen)
	sess.Class = symbiont
	sess.FovRadius = symbiont.FOVRadius
	sess.BaseMaxHP = symbiont.MaxHP
	sess.RunLog.Class = symbiont.Name
	sess.RunLog.EnemiesKilled = make(map[string]int)
	sess.RunLog.ItemsUsed = make(map[string]int)

	srv.mu.Lock()
	srv.sessions = append(srv.sessions, sess)
	srv.spawnPlayerLocked(sess, 0)
	floor := srv.floors[0]
	items := floor.World.Query(component.CTagItem)
	srv.mu.Unlock()

	if len(items) == 0 {
		t.Error("start items should be placed on city floor for symbiont class")
	}
}
