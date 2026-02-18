package system

import (
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/gamemap"
	"math"
	"math/rand"
)

// ProcessAI runs one turn of AI for all AI-controlled entities.
func ProcessAI(w *ecs.World, gmap *gamemap.GameMap, playerID ecs.EntityID, rng *rand.Rand) {
	playerPosComp := w.Get(playerID, component.CPosition)
	if playerPosComp == nil {
		return
	}
	playerPos := playerPosComp.(component.Position)

	for _, id := range w.Query(component.CAI, component.CPosition) {
		aiComp := w.Get(id, component.CAI).(component.AI)
		posComp := w.Get(id, component.CPosition).(component.Position)

		switch aiComp.Behavior {
		case component.BehaviorStationary:
			// never moves
		case component.BehaviorCowardly:
			cowardlyMove(w, gmap, id, posComp, playerPos, aiComp, rng)
		default:
			chaseMove(w, gmap, id, posComp, playerPos, aiComp, rng)
		}
	}
}

func chaseMove(w *ecs.World, gmap *gamemap.GameMap, id ecs.EntityID,
	pos component.Position, playerPos component.Position, ai component.AI, rng *rand.Rand) {

	dx := playerPos.X - pos.X
	dy := playerPos.Y - pos.Y
	dist := math.Sqrt(float64(dx*dx + dy*dy))

	if dist > float64(ai.SightRange) {
		return
	}

	// Normalize to unit step.
	stepX, stepY := sign(dx), sign(dy)

	result, target := TryMove(w, gmap, id, stepX, 0)
	if result == MoveAttack {
		if w.Has(target, component.CTagPlayer) {
			Attack(w, rng, id, target)
		}
		return
	}
	if result == MoveOK {
		return
	}
	// Try vertical.
	result, target = TryMove(w, gmap, id, 0, stepY)
	if result == MoveAttack {
		if w.Has(target, component.CTagPlayer) {
			Attack(w, rng, id, target)
		}
	}
}

func cowardlyMove(w *ecs.World, gmap *gamemap.GameMap, id ecs.EntityID,
	pos component.Position, playerPos component.Position, ai component.AI, rng *rand.Rand) {

	// Cowardly: attack if adjacent, otherwise flee.
	dx := playerPos.X - pos.X
	dy := playerPos.Y - pos.Y
	dist := math.Sqrt(float64(dx*dx + dy*dy))
	if dist > float64(ai.SightRange) {
		return
	}

	if dist <= 1.5 {
		// Adjacent — find player entity and attack.
		for _, playerEnt := range w.Query(component.CTagPlayer) {
			Attack(w, rng, id, playerEnt)
		}
		return
	}

	// Flee: move away from player.
	stepX, stepY := -sign(dx), -sign(dy)
	if TryMoveSimple(w, gmap, id, stepX, 0) == MoveOK {
		return
	}
	TryMoveSimple(w, gmap, id, 0, stepY)
}

// TryMoveSimple is a convenience wrapper that discards the target.
func TryMoveSimple(w *ecs.World, gmap *gamemap.GameMap, id ecs.EntityID, dx, dy int) MoveResult {
	r, _ := TryMove(w, gmap, id, dx, dy)
	return r
}

func sign(v int) int {
	if v > 0 {
		return 1
	}
	if v < 0 {
		return -1
	}
	return 0
}
