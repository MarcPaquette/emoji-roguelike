package mud

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/gamemap"
	"math/rand"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

func newSimScreen() tcell.Screen {
	ss := tcell.NewSimulationScreen("UTF-8")
	ss.SetSize(80, 24)
	_ = ss.Init()
	return ss
}

func newTestServer() *Server {
	rng := rand.New(rand.NewSource(42))
	return NewServer(rng)
}

func newTestSession(id int, _ *Server) *Session {
	screen := newSimScreen()
	_, color := id, playerColors[id%len(playerColors)]
	sess := NewSession(id, "TestPlayer", color, screen)
	sess.Class = assets.Classes[0]
	sess.FovRadius = sess.Class.FOVRadius
	sess.BaseMaxHP = sess.Class.MaxHP
	sess.RunLog.Class = sess.Class.Name
	return sess
}

// ─── Session action queue ─────────────────────────────────────────────────────

func TestSessionActionQueueEmpty(t *testing.T) {
	sess := &Session{RenderCh: make(chan struct{}, 1)}
	if got := sess.TakeAction(); got != ActionNone {
		t.Errorf("expected ActionNone on empty queue, got %v", got)
	}
}

func TestSessionActionQueueSetTake(t *testing.T) {
	sess := &Session{RenderCh: make(chan struct{}, 1)}
	sess.SetAction(ActionMoveN)
	if got := sess.TakeAction(); got != ActionMoveN {
		t.Errorf("expected ActionMoveN, got %v", got)
	}
	// After taking, queue is clear.
	if got := sess.TakeAction(); got != ActionNone {
		t.Errorf("expected ActionNone after take, got %v", got)
	}
}

func TestSessionActionQueueLastKeyWins(t *testing.T) {
	sess := &Session{RenderCh: make(chan struct{}, 1)}
	sess.SetAction(ActionMoveE)
	sess.SetAction(ActionMoveW)
	if got := sess.TakeAction(); got != ActionMoveW {
		t.Errorf("expected last-set action ActionMoveW, got %v", got)
	}
}

// ─── FOV snapshot / apply ─────────────────────────────────────────────────────

func TestFovSnapshotApplyRoundTrip(t *testing.T) {
	gmap := gamemap.New(10, 10)
	for y := range 10 {
		for x := range 10 {
			gmap.Set(x, y, gamemap.MakeFloor())
		}
	}
	gmap.At(2, 3).Visible = true
	gmap.At(7, 8).Visible = true

	sess := &Session{RenderCh: make(chan struct{}, 1)}
	sess.SnapshotFOV(gmap)

	// Clobber the map's visibility.
	for y := range 10 {
		for x := range 10 {
			gmap.At(x, y).Visible = false
		}
	}

	sess.ApplyFOV(gmap)

	if !gmap.At(2, 3).Visible {
		t.Error("expected (2,3) visible after ApplyFOV")
	}
	if !gmap.At(7, 8).Visible {
		t.Error("expected (7,8) visible after ApplyFOV")
	}
	if gmap.At(0, 0).Visible {
		t.Error("expected (0,0) invisible after ApplyFOV")
	}
}

func TestFovApplyNilGridNoOp(t *testing.T) {
	gmap := gamemap.New(5, 5)
	for y := range 5 {
		for x := range 5 {
			gmap.Set(x, y, gamemap.MakeFloor())
			gmap.At(x, y).Visible = true
		}
	}
	sess := &Session{RenderCh: make(chan struct{}, 1)} // FovGrid is nil
	sess.ApplyFOV(gmap)                               // must not panic
	// Visibility is unchanged (ApplyFOV is a no-op with nil grid).
}

// ─── findFreeSpawn ────────────────────────────────────────────────────────────

func newOpenFloor(num int) *Floor {
	gmap := gamemap.New(20, 20)
	for y := range 20 {
		for x := range 20 {
			gmap.Set(x, y, gamemap.MakeFloor())
		}
	}
	gmap.Rooms = []gamemap.Rect{{X1: 1, Y1: 1, X2: 18, Y2: 18}}
	return &Floor{
		Num:   num,
		World: ecs.NewWorld(),
		GMap:  gmap,
	}
}

func TestFindFreeSpawnNoOccupants(t *testing.T) {
	floor := newOpenFloor(1)
	x, y := findFreeSpawn(floor, nil, 5, 5)
	if x != 5 || y != 5 {
		t.Errorf("expected (5,5), got (%d,%d)", x, y)
	}
}

func TestFindFreeSpawnOccupied(t *testing.T) {
	floor := newOpenFloor(1)

	// Place a player entity at (5,5).
	playerID := floor.World.CreateEntity()
	floor.World.Add(playerID, component.Position{X: 5, Y: 5})

	sess := &Session{
		FloorNum: 1,
		PlayerID: playerID,
		RenderCh: make(chan struct{}, 1),
	}
	sessions := []*Session{sess}

	x, y := findFreeSpawn(floor, sessions, 5, 5)
	if x == 5 && y == 5 {
		t.Error("expected a different position when (5,5) is occupied")
	}
	if !floor.GMap.IsWalkable(x, y) {
		t.Errorf("spawn position (%d,%d) is not walkable", x, y)
	}
}

func TestFindFreeSpawnNilSessions(t *testing.T) {
	floor := newOpenFloor(1)
	// nil sessions slice: no occupants → desired position returned.
	x, y := findFreeSpawn(floor, nil, 3, 7)
	if x != 3 || y != 7 {
		t.Errorf("expected (3,7), got (%d,%d)", x, y)
	}
}

// ─── Server session lifecycle ─────────────────────────────────────────────────

func TestServerAddRemoveSession(t *testing.T) {
	srv := newTestServer()
	sess := newTestSession(0, srv)

	srv.AddSession(sess)

	if sess.FloorNum != 1 {
		t.Errorf("expected FloorNum=1, got %d", sess.FloorNum)
	}
	if sess.PlayerID == ecs.NilEntity {
		t.Error("expected a valid PlayerID after AddSession")
	}

	srv.mu.Lock()
	floor := srv.floors[1]
	posComp := floor.World.Get(sess.PlayerID, component.CPosition)
	srv.mu.Unlock()

	if posComp == nil {
		t.Fatal("expected player to have a Position component")
	}

	srv.RemoveSession(sess)

	srv.mu.Lock()
	alive := floor.World.Alive(sess.PlayerID)
	count := len(srv.sessions)
	srv.mu.Unlock()

	if alive {
		t.Error("expected player entity destroyed after RemoveSession")
	}
	if count != 0 {
		t.Errorf("expected 0 sessions, got %d", count)
	}
}

func TestServerMultipleSessionsDistinctColors(t *testing.T) {
	srv := newTestServer()
	sess0 := newTestSession(0, srv)
	sess1 := newTestSession(1, srv)
	if sess0.Color == sess1.Color {
		t.Error("consecutive sessions should have distinct colors")
	}
	srv.AddSession(sess0)
	srv.AddSession(sess1)
	defer srv.RemoveSession(sess0)
	defer srv.RemoveSession(sess1)
}

// ─── Stairs placement ─────────────────────────────────────────────────────────

func TestStairsUpOnNonFirstFloor(t *testing.T) {
	for floorNum := 2; floorNum <= 5; floorNum++ {
		rng := rand.New(rand.NewSource(int64(floorNum) * 7))
		floor := newFloor(floorNum, rng)

		found := false
		for y := range floor.GMap.Height {
			for x := range floor.GMap.Width {
				if floor.GMap.At(x, y).Kind == gamemap.TileStairsUp {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("floor %d: expected stairs up tile but none found", floorNum)
		}
	}
}

func TestNoStairsUpOnFloor1(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	floor := newFloor(1, rng)

	for y := range floor.GMap.Height {
		for x := range floor.GMap.Width {
			if floor.GMap.At(x, y).Kind == gamemap.TileStairsUp {
				t.Errorf("floor 1 should not have stairs up (found at %d,%d)", x, y)
			}
		}
	}
}

func TestStairsDownOnAllFloors(t *testing.T) {
	for floorNum := 1; floorNum <= 5; floorNum++ {
		rng := rand.New(rand.NewSource(int64(floorNum) * 13))
		floor := newFloor(floorNum, rng)

		found := false
		for y := range floor.GMap.Height {
			for x := range floor.GMap.Width {
				if floor.GMap.At(x, y).Kind == gamemap.TileStairsDown {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("floor %d: expected stairs down tile but none found", floorNum)
		}
	}
}

// ─── Floor transition ─────────────────────────────────────────────────────────

func TestFloorTransitionPreservesHP(t *testing.T) {
	srv := newTestServer()
	sess := newTestSession(0, srv)
	srv.AddSession(sess)

	// Reduce player HP on floor 1.
	srv.mu.Lock()
	floor1 := srv.floors[1]
	if hp := floor1.World.Get(sess.PlayerID, component.CHealth); hp != nil {
		h := hp.(component.Health)
		h.Current = 5
		floor1.World.Add(sess.PlayerID, h)
	}
	// Descend to floor 2.
	srv.transitionFloorLocked(sess, 2)
	floor2 := srv.floors[2]
	hpComp := floor2.World.Get(sess.PlayerID, component.CHealth)
	srv.mu.Unlock()

	if hpComp == nil {
		t.Fatal("player has no Health component on floor 2")
	}
	h := hpComp.(component.Health)
	if h.Current != 5 {
		t.Errorf("expected HP=5 after descend, got %d", h.Current)
	}
}

func TestFloorTransitionDescendLandsNearStairsUp(t *testing.T) {
	srv := newTestServer()
	sess := newTestSession(0, srv)
	srv.AddSession(sess)

	srv.mu.Lock()
	srv.transitionFloorLocked(sess, 2)
	floor2 := srv.floors[2]
	posComp := floor2.World.Get(sess.PlayerID, component.CPosition)
	srv.mu.Unlock()

	if posComp == nil {
		t.Fatal("player has no Position on floor 2")
	}
	pos := posComp.(component.Position)
	// Stairs up are at floor2.StairsUpX/Y (== SpawnX/Y for floor 2).
	// Player should be within a small radius.
	srv.mu.Lock()
	ux, uy := floor2.StairsUpX, floor2.StairsUpY
	srv.mu.Unlock()

	dx := pos.X - ux
	dy := pos.Y - uy
	dist := dx*dx + dy*dy
	if dist > 25 { // within 5 tiles
		t.Errorf("player (%d,%d) too far from stairs up (%d,%d)", pos.X, pos.Y, ux, uy)
	}
}

// ─── Enemy respawn ────────────────────────────────────────────────────────────

func TestRespawnEnemiesLocked(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	floor := newFloor(3, rng) // floor 3 has a varied enemy table

	// Clear all enemies.
	for _, id := range floor.World.Query(component.CAI) {
		floor.World.DestroyEntity(id)
	}
	if got := len(floor.World.Query(component.CAI)); got != 0 {
		t.Fatalf("expected 0 enemies after clear, got %d", got)
	}

	srv := &Server{
		floors:   map[int]*Floor{3: floor},
		sessions: nil,
		rng:      rng,
	}
	srv.respawnEnemiesLocked(floor)

	newCount := len(floor.World.Query(component.CAI))
	if newCount == 0 {
		t.Error("expected enemies after respawnEnemiesLocked")
	}
}

func TestFloorRespawnCooldownInitiallyIdle(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	floor := newFloor(1, rng)
	if floor.RespawnCooldown != -1 {
		t.Errorf("expected RespawnCooldown=-1 on new floor, got %d", floor.RespawnCooldown)
	}
}

// ─── Multi-player AI targeting ────────────────────────────────────────────────

func TestMultiPlayerAITargetsNearest(t *testing.T) {
	// Place Alice close to an enemy, Bob far away.
	// After tickFloorLocked, only Alice should have taken damage.
	floor := newOpenFloor(1)
	floor.Rng = rand.New(rand.NewSource(42))

	makeSessionPlayer := func(id int, name string, x, y int) *Session {
		screen := newSimScreen()
		sess := NewSession(id, name, playerColors[id%len(playerColors)], screen)
		sess.Class = assets.Classes[0]
		sess.FovRadius = sess.Class.FOVRadius
		sess.BaseMaxHP = sess.Class.MaxHP
		sess.RunLog.EnemiesKilled = make(map[string]int)
		sess.RunLog.ItemsUsed = make(map[string]int)
		sess.FloorNum = 1

		pid := floor.World.CreateEntity()
		floor.World.Add(pid, component.Position{X: x, Y: y})
		floor.World.Add(pid, component.TagPlayer{})
		floor.World.Add(pid, component.TagBlocking{})
		floor.World.Add(pid, component.Combat{Attack: 3, Defense: 1})
		floor.World.Add(pid, component.Health{Current: 20, Max: 20})
		floor.World.Add(pid, component.Effects{})
		sess.PlayerID = pid
		return sess
	}

	alice := makeSessionPlayer(0, "Alice", 5, 5) // near enemy
	bob := makeSessionPlayer(1, "Bob", 18, 18)   // far from enemy

	// Enemy adjacent to Alice at (6,5), sight range 10.
	enemy := floor.World.CreateEntity()
	floor.World.Add(enemy, component.Position{X: 6, Y: 5})
	floor.World.Add(enemy, component.AI{Behavior: component.BehaviorChase, SightRange: 10})
	floor.World.Add(enemy, component.Combat{Attack: 5, Defense: 0})
	floor.World.Add(enemy, component.Health{Current: 20, Max: 20})
	floor.World.Add(enemy, component.Renderable{Glyph: "🦀"})
	floor.World.Add(enemy, component.TagBlocking{})
	floor.World.Add(enemy, component.Effects{})

	srv := &Server{
		floors:   map[int]*Floor{1: floor},
		sessions: []*Session{alice, bob},
		rng:      rand.New(rand.NewSource(42)),
	}

	srv.mu.Lock()
	srv.tickFloorLocked(floor)
	srv.mu.Unlock()

	aliceHP := floor.World.Get(alice.PlayerID, component.CHealth).(component.Health).Current
	bobHP := floor.World.Get(bob.PlayerID, component.CHealth).(component.Health).Current

	// Alice (adjacent to enemy) should have taken damage.
	if aliceHP >= 20 {
		t.Errorf("Alice should have taken damage; HP=%d", aliceHP)
	}
	// Bob (far away, also within sight range 10 but farther than Alice) should take no damage.
	// Enemy picks nearest: Alice at dist=1 vs Bob at dist≈18.
	if bobHP != 20 {
		t.Errorf("Bob should not have taken damage; HP=%d", bobHP)
	}
	// Damage should be attributed to Alice's run log.
	if alice.RunLog.DamageTaken == 0 {
		t.Error("Alice's run log should record damage taken")
	}
	if bob.RunLog.DamageTaken != 0 {
		t.Error("Bob's run log should not record damage taken")
	}
}
