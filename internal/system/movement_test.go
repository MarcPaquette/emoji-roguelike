package system

import (
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/gamemap"
	"testing"
)

func setupMoveWorld() (*ecs.World, *gamemap.GameMap, ecs.EntityID) {
	w := ecs.NewWorld()
	gmap := gamemap.New(10, 10)
	// Carve a small open area.
	for y := 1; y <= 8; y++ {
		for x := 1; x <= 8; x++ {
			gmap.Set(x, y, gamemap.MakeFloor())
		}
	}
	player := w.CreateEntity()
	w.Add(player, component.Position{X: 3, Y: 3})
	w.Add(player, component.TagBlocking{})
	return w, gmap, player
}

func TestTryMoveSucceeds(t *testing.T) {
	w, gmap, player := setupMoveWorld()
	result, _ := TryMove(w, gmap, player, 1, 0)
	if result != MoveOK {
		t.Fatalf("expected MoveOK, got %v", result)
	}
	pos := w.Get(player, component.CPosition).(component.Position)
	if pos.X != 4 || pos.Y != 3 {
		t.Fatalf("expected position (4,3), got (%d,%d)", pos.X, pos.Y)
	}
}

func TestTryMoveBlockedByWall(t *testing.T) {
	w, gmap, player := setupMoveWorld()
	// Move up into wall row (y=0).
	w.Add(player, component.Position{X: 3, Y: 1})
	result, _ := TryMove(w, gmap, player, 0, -1)
	if result != MoveBlocked {
		t.Fatalf("expected MoveBlocked, got %v", result)
	}
	pos := w.Get(player, component.CPosition).(component.Position)
	if pos.Y != 1 {
		t.Fatalf("position should be unchanged, got (%d,%d)", pos.X, pos.Y)
	}
}

func TestTryMoveIntoEntityReturnsAttack(t *testing.T) {
	w, gmap, player := setupMoveWorld()
	// Place a blocking enemy at (4,3).
	enemy := w.CreateEntity()
	w.Add(enemy, component.Position{X: 4, Y: 3})
	w.Add(enemy, component.TagBlocking{})

	result, target := TryMove(w, gmap, player, 1, 0)
	if result != MoveAttack {
		t.Fatalf("expected MoveAttack, got %v", result)
	}
	if target != enemy {
		t.Fatalf("expected target=%v, got %v", enemy, target)
	}
	// Player position should be unchanged.
	pos := w.Get(player, component.CPosition).(component.Position)
	if pos.X != 3 {
		t.Fatalf("player should not have moved, got (%d,%d)", pos.X, pos.Y)
	}
}
