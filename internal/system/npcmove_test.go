package system

import (
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/gamemap"
	"math/rand"
	"testing"
)

// makeOpenMap creates a fully walkable WxH map.
func makeOpenMap(w, h int) *gamemap.GameMap {
	gmap := gamemap.New(w, h)
	for y := range h {
		for x := range w {
			gmap.Set(x, y, gamemap.MakeFloor())
		}
	}
	return gmap
}

// makeNPC creates a minimal NPC entity with position and blocking tag.
func makeNPC(w *ecs.World, x, y int) ecs.EntityID {
	id := w.CreateEntity()
	w.Add(id, component.Position{X: x, Y: y})
	w.Add(id, component.TagBlocking{})
	w.Add(id, component.NPC{Name: "Test", Kind: component.NPCKindDialogue})
	return id
}

func TestActiveScheduleEntry(t *testing.T) {
	schedule := []component.ScheduleEntry{
		{StartTick: 0},
		{StartTick: 1500},
		{StartTick: 2500},
		{StartTick: 5000},
	}

	cases := []struct {
		name    string
		dayTick int
		want    int
	}{
		{"at start", 0, 0},
		{"before morning", 1499, 0},
		{"at morning", 1500, 1},
		{"mid morning", 2000, 1},
		{"at day", 2500, 2},
		{"mid day", 3500, 2},
		{"at evening", 5000, 3},
		{"end of day", 5999, 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := activeScheduleEntry(schedule, tc.dayTick)
			if got != tc.want {
				t.Errorf("activeScheduleEntry(schedule, %d) = %d, want %d", tc.dayTick, got, tc.want)
			}
		})
	}
}

func TestActiveScheduleEntryEmpty(t *testing.T) {
	got := activeScheduleEntry(nil, 100)
	if got != 0 {
		t.Errorf("activeScheduleEntry(nil, 100) = %d, want 0", got)
	}
}

func TestWanderStaysInBounds(t *testing.T) {
	world := ecs.NewWorld()
	gmap := makeOpenMap(30, 30)
	rng := rand.New(rand.NewSource(42))

	id := makeNPC(world, 15, 15)
	mc := component.NPCMovement{
		Schedule: []component.ScheduleEntry{
			{StartTick: 0, Behavior: component.MoveWander,
				BoundsX1: 10, BoundsY1: 10, BoundsX2: 20, BoundsY2: 20},
		},
		ActiveIndex:  0,
		Behavior:     component.MoveWander,
		MoveInterval: 1,
		BoundsX1:     10, BoundsY1: 10, BoundsX2: 20, BoundsY2: 20,
	}
	world.Add(id, mc)

	for range 100 {
		ProcessNPCMovement(world, gmap, 0, rng)
		pos := world.Get(id, component.CPosition).(component.Position)
		if pos.X < 10 || pos.X > 20 || pos.Y < 10 || pos.Y > 20 {
			t.Fatalf("NPC wandered out of bounds to (%d,%d)", pos.X, pos.Y)
		}
	}
}

func TestPathReachesWaypoints(t *testing.T) {
	world := ecs.NewWorld()
	gmap := makeOpenMap(30, 30)
	rng := rand.New(rand.NewSource(42))

	id := makeNPC(world, 5, 5)
	waypoints := []component.Waypoint{{X: 10, Y: 5}, {X: 10, Y: 10}, {X: 15, Y: 10}}
	mc := component.NPCMovement{
		Schedule: []component.ScheduleEntry{
			{StartTick: 0, Behavior: component.MovePath, Waypoints: waypoints},
		},
		ActiveIndex:  0,
		Behavior:     component.MovePath,
		MoveInterval: 1,
		Waypoints:    waypoints,
		WaypointIdx:  0,
	}
	world.Add(id, mc)

	for range 200 {
		ProcessNPCMovement(world, gmap, 0, rng)
	}

	pos := world.Get(id, component.CPosition).(component.Position)
	if pos.X != 15 || pos.Y != 10 {
		t.Errorf("NPC did not reach final waypoint: got (%d,%d), want (15,10)", pos.X, pos.Y)
	}
	mc2 := world.Get(id, component.CNPCMovement).(component.NPCMovement)
	if !mc2.PathDone {
		t.Error("PathDone should be true after reaching final waypoint")
	}
}

func TestPathBlockedUsesAlternateAxis(t *testing.T) {
	world := ecs.NewWorld()
	gmap := makeOpenMap(20, 20)
	rng := rand.New(rand.NewSource(42))

	// Place a blocker directly in the NPC's path.
	blocker := world.CreateEntity()
	world.Add(blocker, component.Position{X: 6, Y: 5})
	world.Add(blocker, component.TagBlocking{})

	id := makeNPC(world, 5, 5)
	waypoints := []component.Waypoint{{X: 10, Y: 5}}
	mc := component.NPCMovement{
		Schedule: []component.ScheduleEntry{
			{StartTick: 0, Behavior: component.MovePath, Waypoints: waypoints},
		},
		ActiveIndex:  0,
		Behavior:     component.MovePath,
		MoveInterval: 1,
		Waypoints:    waypoints,
		WaypointIdx:  0,
	}
	world.Add(id, mc)

	// Run a few ticks — the NPC should try alternate axes.
	for range 5 {
		ProcessNPCMovement(world, gmap, 0, rng)
	}

	pos := world.Get(id, component.CPosition).(component.Position)
	// Should have moved (not stuck at original position).
	if pos.X == 5 && pos.Y == 5 {
		t.Error("NPC should have moved via alternate axis to avoid blocker")
	}
}

func TestMoveReturnReachesTarget(t *testing.T) {
	world := ecs.NewWorld()
	gmap := makeOpenMap(20, 20)
	rng := rand.New(rand.NewSource(42))

	id := makeNPC(world, 10, 10)
	mc := component.NPCMovement{
		Schedule: []component.ScheduleEntry{
			{StartTick: 0, Behavior: component.MoveStationary, StandX: 5, StandY: 5},
		},
		ActiveIndex:  0,
		Behavior:     component.MoveReturn,
		MoveInterval: 1,
		ReturnX:      5,
		ReturnY:      5,
		ReturnThen:   component.MoveStationary,
	}
	world.Add(id, mc)

	for range 20 {
		ProcessNPCMovement(world, gmap, 0, rng)
	}

	pos := world.Get(id, component.CPosition).(component.Position)
	if pos.X != 5 || pos.Y != 5 {
		t.Errorf("NPC did not return to target: got (%d,%d), want (5,5)", pos.X, pos.Y)
	}
	mc2 := world.Get(id, component.CNPCMovement).(component.NPCMovement)
	if mc2.Behavior != component.MoveStationary {
		t.Errorf("Behavior should be MoveStationary after return, got %d", mc2.Behavior)
	}
}

func TestTryMoveNPCBlocked(t *testing.T) {
	cases := []struct {
		name   string
		setup  func(w *ecs.World, gmap *gamemap.GameMap)
		npcX   int
		npcY   int
		dx, dy int
	}{
		{
			name: "wall",
			setup: func(w *ecs.World, gmap *gamemap.GameMap) {
				// Default map has walls at edges; leave (0,0) as wall.
			},
			npcX: 1, npcY: 1, dx: -1, dy: 0,
		},
		{
			name: "other NPC blocking",
			setup: func(w *ecs.World, gmap *gamemap.GameMap) {
				other := w.CreateEntity()
				w.Add(other, component.Position{X: 6, Y: 5})
				w.Add(other, component.TagBlocking{})
			},
			npcX: 5, npcY: 5, dx: 1, dy: 0,
		},
		{
			name: "player blocking",
			setup: func(w *ecs.World, gmap *gamemap.GameMap) {
				player := w.CreateEntity()
				w.Add(player, component.Position{X: 5, Y: 4})
				w.Add(player, component.TagBlocking{})
				w.Add(player, component.TagPlayer{})
			},
			npcX: 5, npcY: 5, dx: 0, dy: -1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			world := ecs.NewWorld()
			gmap := makeOpenMap(20, 20)
			// Put wall at (0,*) for wall test.
			for y := range 20 {
				gmap.Set(0, y, gamemap.MakeWall())
			}
			tc.setup(world, gmap)

			npcID := world.CreateEntity()
			world.Add(npcID, component.Position{X: tc.npcX, Y: tc.npcY})
			world.Add(npcID, component.TagBlocking{})

			ok := tryMoveNPC(world, gmap, npcID, tc.dx, tc.dy)
			if ok {
				t.Error("tryMoveNPC should have returned false")
			}
			pos := world.Get(npcID, component.CPosition).(component.Position)
			if pos.X != tc.npcX || pos.Y != tc.npcY {
				t.Errorf("position should not have changed, got (%d,%d)", pos.X, pos.Y)
			}
		})
	}
}

func TestTryMoveNPCSuccess(t *testing.T) {
	world := ecs.NewWorld()
	gmap := makeOpenMap(20, 20)

	npcID := world.CreateEntity()
	world.Add(npcID, component.Position{X: 5, Y: 5})
	world.Add(npcID, component.TagBlocking{})

	ok := tryMoveNPC(world, gmap, npcID, 1, 0)
	if !ok {
		t.Fatal("tryMoveNPC should have returned true")
	}
	pos := world.Get(npcID, component.CPosition).(component.Position)
	if pos.X != 6 || pos.Y != 5 {
		t.Errorf("position should be (6,5), got (%d,%d)", pos.X, pos.Y)
	}
}

func TestScheduleTransition(t *testing.T) {
	world := ecs.NewWorld()
	gmap := makeOpenMap(30, 30)
	rng := rand.New(rand.NewSource(42))

	id := makeNPC(world, 15, 15)
	mc := component.NPCMovement{
		Schedule: []component.ScheduleEntry{
			{StartTick: 0, Behavior: component.MoveStationary, StandX: 15, StandY: 15},
			{StartTick: 1500, Behavior: component.MoveWander,
				BoundsX1: 10, BoundsY1: 10, BoundsX2: 20, BoundsY2: 20},
		},
		ActiveIndex:  0,
		Behavior:     component.MoveStationary,
		MoveInterval: 1,
	}
	world.Add(id, mc)

	// Tick at dayTick=0 — should stay stationary.
	ProcessNPCMovement(world, gmap, 0, rng)
	mc2 := world.Get(id, component.CNPCMovement).(component.NPCMovement)
	if mc2.Behavior != component.MoveStationary {
		t.Errorf("expected MoveStationary at tick 0, got %d", mc2.Behavior)
	}

	// Tick at dayTick=1500 — should transition to wander.
	ProcessNPCMovement(world, gmap, 1500, rng)
	mc3 := world.Get(id, component.CNPCMovement).(component.NPCMovement)
	if mc3.Behavior != component.MoveWander {
		t.Errorf("expected MoveWander at tick 1500, got %d", mc3.Behavior)
	}
	if mc3.ActiveIndex != 1 {
		t.Errorf("expected ActiveIndex 1, got %d", mc3.ActiveIndex)
	}
}

func TestStationaryNeverMoves(t *testing.T) {
	world := ecs.NewWorld()
	gmap := makeOpenMap(20, 20)
	rng := rand.New(rand.NewSource(42))

	id := makeNPC(world, 10, 10)
	mc := component.NPCMovement{
		Schedule: []component.ScheduleEntry{
			{StartTick: 0, Behavior: component.MoveStationary, StandX: 10, StandY: 10},
		},
		ActiveIndex:  0,
		Behavior:     component.MoveStationary,
		MoveInterval: 1,
	}
	world.Add(id, mc)

	for range 100 {
		ProcessNPCMovement(world, gmap, 0, rng)
		pos := world.Get(id, component.CPosition).(component.Position)
		if pos.X != 10 || pos.Y != 10 {
			t.Fatalf("stationary NPC moved to (%d,%d)", pos.X, pos.Y)
		}
	}
}

func TestNPCDoesNotCrossWater(t *testing.T) {
	world := ecs.NewWorld()
	gmap := makeOpenMap(20, 20)
	// Place a row of water.
	for x := range 20 {
		gmap.Set(x, 10, gamemap.MakeWater())
	}
	rng := rand.New(rand.NewSource(42))

	id := makeNPC(world, 10, 9)
	mc := component.NPCMovement{
		Schedule: []component.ScheduleEntry{
			{StartTick: 0, Behavior: component.MoveWander,
				BoundsX1: 0, BoundsY1: 0, BoundsX2: 19, BoundsY2: 19},
		},
		ActiveIndex:  0,
		Behavior:     component.MoveWander,
		MoveInterval: 1,
		BoundsX1:     0, BoundsY1: 0, BoundsX2: 19, BoundsY2: 19,
	}
	world.Add(id, mc)

	for range 200 {
		ProcessNPCMovement(world, gmap, 0, rng)
		pos := world.Get(id, component.CPosition).(component.Position)
		if pos.Y == 10 {
			t.Fatal("NPC walked onto water tile")
		}
	}
}
