package system

import (
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/gamemap"
	"math"
	"math/rand"
)

// EnemyHitResult holds information about an enemy attack on the player.
type EnemyHitResult struct {
	EnemyGlyph     string
	SpecialApplied uint8
	DrainedAmount  int
	Damage         int
}

// ProcessAI runs one turn of AI for all AI-controlled entities and returns
// the results of any attacks made against the player.
func ProcessAI(w *ecs.World, gmap *gamemap.GameMap, playerID ecs.EntityID, rng *rand.Rand) []EnemyHitResult {
	playerPosComp := w.Get(playerID, component.CPosition)
	if playerPosComp == nil {
		return nil
	}
	playerPos := playerPosComp.(component.Position)

	var hits []EnemyHitResult
	for _, id := range w.Query(component.CAI, component.CPosition) {
		aiComp := w.Get(id, component.CAI).(component.AI)
		posComp := w.Get(id, component.CPosition).(component.Position)

		var attacked bool
		var res AttackResult
		var glyph string
		switch aiComp.Behavior {
		case component.BehaviorStationary:
			// never moves
		case component.BehaviorCowardly:
			attacked, res, glyph = cowardlyMove(w, gmap, id, posComp, playerPos, aiComp, rng)
		default:
			attacked, res, glyph = chaseMove(w, gmap, id, posComp, playerPos, aiComp, rng)
		}
		if attacked {
			hits = append(hits, EnemyHitResult{
				EnemyGlyph:     glyph,
				SpecialApplied: res.SpecialApplied,
				DrainedAmount:  res.DrainedAmount,
				Damage:         res.Damage,
			})
		}
	}
	return hits
}

func chaseMove(w *ecs.World, gmap *gamemap.GameMap, id ecs.EntityID,
	pos component.Position, playerPos component.Position, ai component.AI, rng *rand.Rand) (bool, AttackResult, string) {

	dx := playerPos.X - pos.X
	dy := playerPos.Y - pos.Y
	dist := math.Sqrt(float64(dx*dx + dy*dy))

	if dist > float64(ai.SightRange) {
		return false, AttackResult{}, ""
	}

	// Normalize to unit step.
	stepX, stepY := sign(dx), sign(dy)

	result, target := TryMove(w, gmap, id, stepX, 0)
	if result == MoveAttack {
		if w.Has(target, component.CTagPlayer) {
			glyph := enemyGlyph(w, id)
			res := Attack(w, rng, id, target)
			return true, res, glyph
		}
		return false, AttackResult{}, ""
	}
	if result == MoveOK {
		return false, AttackResult{}, ""
	}
	// Try vertical.
	result, target = TryMove(w, gmap, id, 0, stepY)
	if result == MoveAttack {
		if w.Has(target, component.CTagPlayer) {
			glyph := enemyGlyph(w, id)
			res := Attack(w, rng, id, target)
			return true, res, glyph
		}
	}
	return false, AttackResult{}, ""
}

func cowardlyMove(w *ecs.World, gmap *gamemap.GameMap, id ecs.EntityID,
	pos component.Position, playerPos component.Position, ai component.AI, rng *rand.Rand) (bool, AttackResult, string) {

	// Cowardly: attack if adjacent, otherwise flee.
	dx := playerPos.X - pos.X
	dy := playerPos.Y - pos.Y
	dist := math.Sqrt(float64(dx*dx + dy*dy))
	if dist > float64(ai.SightRange) {
		return false, AttackResult{}, ""
	}

	if dist <= 1.5 {
		// Adjacent — find player entity and attack.
		for _, playerEnt := range w.Query(component.CTagPlayer) {
			glyph := enemyGlyph(w, id)
			res := Attack(w, rng, id, playerEnt)
			return true, res, glyph
		}
		return false, AttackResult{}, ""
	}

	// Flee: move away from player.
	stepX, stepY := -sign(dx), -sign(dy)
	if TryMoveSimple(w, gmap, id, stepX, 0) == MoveOK {
		return false, AttackResult{}, ""
	}
	TryMoveSimple(w, gmap, id, 0, stepY)
	return false, AttackResult{}, ""
}

// enemyGlyph returns the glyph of an enemy entity (safe to call before Attack).
func enemyGlyph(w *ecs.World, id ecs.EntityID) string {
	c := w.Get(id, component.CRenderable)
	if c == nil {
		return "?"
	}
	return c.(component.Renderable).Glyph
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
