package system

import (
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/gamemap"
	"math/rand"
	"testing"
)

// openMap creates a w×h map that is entirely passable floor.
func openMap(w, h int) *gamemap.GameMap {
	gmap := gamemap.New(w, h)
	for y := range h {
		for x := range w {
			gmap.Set(x, y, gamemap.MakeFloor())
		}
	}
	return gmap
}

// newAIWorld creates a minimal world with a player at (px, py) on a 20×20 open map.
func newAIWorld(px, py int) (*ecs.World, *gamemap.GameMap, ecs.EntityID) {
	w := ecs.NewWorld()
	gmap := openMap(20, 20)
	player := w.CreateEntity()
	w.Add(player, component.Position{X: px, Y: py})
	w.Add(player, component.TagPlayer{})
	w.Add(player, component.TagBlocking{})
	w.Add(player, component.Combat{Attack: 3, Defense: 1})
	w.Add(player, component.Health{Current: 30, Max: 30})
	w.Add(player, component.Effects{})
	return w, gmap, player
}

// addEnemy adds a minimal chase-behavior enemy entity to the world.
func addEnemy(w *ecs.World, x, y int, behavior component.AIBehavior, sightRange int) ecs.EntityID {
	id := w.CreateEntity()
	w.Add(id, component.Position{X: x, Y: y})
	w.Add(id, component.AI{Behavior: behavior, SightRange: sightRange})
	w.Add(id, component.Combat{Attack: 4, Defense: 0})
	w.Add(id, component.Health{Current: 20, Max: 20})
	w.Add(id, component.Renderable{Glyph: "🦀"})
	w.Add(id, component.TagBlocking{})
	w.Add(id, component.Effects{})
	return id
}

func TestProcessAINoPlayerPosition(t *testing.T) {
	// Player exists but has no CPosition — ProcessAI must return nil without panicking.
	w := ecs.NewWorld()
	gmap := openMap(10, 10)
	player := w.CreateEntity()
	w.Add(player, component.TagPlayer{})
	// Deliberately no CPosition.

	rng := rand.New(rand.NewSource(0))
	hits := ProcessAI(w, gmap, []ecs.EntityID{player}, rng)
	if hits != nil {
		t.Errorf("expected nil hits when player has no position; got %v", hits)
	}
}

func TestAIStationaryNeverAttacks(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	w, gmap, player := newAIWorld(5, 5)
	enemy := addEnemy(w, 6, 5, component.BehaviorStationary, 10)
	startPos := w.Get(enemy, component.CPosition).(component.Position)

	hits := ProcessAI(w, gmap, []ecs.EntityID{player}, rng)
	if len(hits) != 0 {
		t.Errorf("stationary enemy should never attack; got %d hit(s)", len(hits))
	}
	endPos := w.Get(enemy, component.CPosition).(component.Position)
	if endPos != startPos {
		t.Errorf("stationary enemy moved from %v to %v", startPos, endPos)
	}
}

func TestAIChaseAttacksAdjacentPlayer(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	w, gmap, player := newAIWorld(5, 5)
	// Place enemy 1 tile east of player (distance = 1, within sight).
	addEnemy(w, 6, 5, component.BehaviorChase, 10)

	hits := ProcessAI(w, gmap, []ecs.EntityID{player}, rng)
	if len(hits) != 1 {
		t.Fatalf("adjacent chase enemy should attack player; got %d hit(s)", len(hits))
	}
	if hits[0].Damage < 1 {
		t.Errorf("hit damage should be ≥ 1; got %d", hits[0].Damage)
	}
}

func TestAIChaseIgnoresOutOfRangePlayer(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	// Player at (0,0), enemy at (10,10): distance ≈ 14, sight range = 3.
	w, gmap, player := newAIWorld(0, 0)
	enemy := addEnemy(w, 10, 10, component.BehaviorChase, 3)
	startPos := w.Get(enemy, component.CPosition).(component.Position)

	hits := ProcessAI(w, gmap, []ecs.EntityID{player}, rng)
	if len(hits) != 0 {
		t.Errorf("out-of-range enemy should not attack; got %d hit(s)", len(hits))
	}
	endPos := w.Get(enemy, component.CPosition).(component.Position)
	if endPos != startPos {
		t.Errorf("out-of-range enemy should not move; was %v, now %v", startPos, endPos)
	}
}

func TestAICowardlyAttacksWhenAdjacent(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	w, gmap, player := newAIWorld(5, 5)
	// dist = 1 ≤ 1.5 → attacks immediately.
	addEnemy(w, 6, 5, component.BehaviorCowardly, 10)

	hits := ProcessAI(w, gmap, []ecs.EntityID{player}, rng)
	if len(hits) != 1 {
		t.Fatalf("adjacent cowardly enemy should attack player; got %d hit(s)", len(hits))
	}
}

func TestAICowardlyFleesWhenNotAdjacent(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	// Player at (5,5), enemy at (8,5): dist=3, within sight but dist>1.5 → flee.
	w, gmap, player := newAIWorld(5, 5)
	enemy := addEnemy(w, 8, 5, component.BehaviorCowardly, 10)

	hits := ProcessAI(w, gmap, []ecs.EntityID{player}, rng)
	if len(hits) != 0 {
		t.Errorf("non-adjacent cowardly enemy should not attack; got %d hit(s)", len(hits))
	}
	// stepX = -sign(playerX - enemyX) = -sign(5-8) = -sign(-3) = 1 → enemy moves east.
	endPos := w.Get(enemy, component.CPosition).(component.Position)
	if endPos.X <= 8 {
		t.Errorf("cowardly enemy should have moved east (away from player at x=5); got x=%d", endPos.X)
	}
}

func TestAIMultiPlayerTargetsNearest(t *testing.T) {
	// Two players: near at (2,0), far at (15,0).
	// One chase enemy at (3,0) with sight range 5.
	// Only the near player is within sight; enemy should attack them.
	rng := rand.New(rand.NewSource(0))
	w := ecs.NewWorld()
	gmap := openMap(20, 20)

	makePlayer := func(x, y int) ecs.EntityID {
		p := w.CreateEntity()
		w.Add(p, component.Position{X: x, Y: y})
		w.Add(p, component.TagPlayer{})
		w.Add(p, component.TagBlocking{})
		w.Add(p, component.Combat{Attack: 3, Defense: 1})
		w.Add(p, component.Health{Current: 30, Max: 30})
		w.Add(p, component.Effects{})
		return p
	}

	nearPlayer := makePlayer(2, 0)
	farPlayer := makePlayer(15, 0)

	// Enemy at (3,0), sight range 5 — nearPlayer (dist=1) is in range, farPlayer (dist=12) is not.
	addEnemy(w, 3, 0, component.BehaviorChase, 5)

	hits := ProcessAI(w, gmap, []ecs.EntityID{nearPlayer, farPlayer}, rng)
	if len(hits) == 0 {
		t.Fatal("expected enemy to attack the near player")
	}
	if hits[0].VictimID != nearPlayer {
		t.Errorf("expected VictimID=%v (nearPlayer), got %v", nearPlayer, hits[0].VictimID)
	}
	if hits[0].Damage < 1 {
		t.Errorf("expected damage >= 1, got %d", hits[0].Damage)
	}
	// Far player should not have taken damage.
	farHP := w.Get(farPlayer, component.CHealth).(component.Health).Current
	if farHP != 30 {
		t.Errorf("far player should not have taken damage; HP=%d", farHP)
	}
}
