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
	AttackerID     ecs.EntityID // entity that landed the attack (for thorns)
	VictimID       ecs.EntityID // player entity that was hit
	SpecialApplied uint8
	DrainedAmount  int
	Damage         int
}

// ProcessAI runs one turn of AI for all AI-controlled entities and returns
// the results of any attacks made against the player(s).
func ProcessAI(w *ecs.World, gmap *gamemap.GameMap, playerIDs []ecs.EntityID, rng *rand.Rand) []EnemyHitResult {
	if len(playerIDs) == 0 {
		return nil
	}

	var hits []EnemyHitResult
	for _, id := range w.Query(component.CAI, component.CPosition) {
		aiComp := w.Get(id, component.CAI).(component.AI)
		posComp := w.Get(id, component.CPosition).(component.Position)

		_, targetPos, inRange := nearestPlayer(w, playerIDs, posComp, aiComp.SightRange)
		if !inRange {
			continue
		}

		var attacked bool
		var res AttackResult
		var glyph string
		var victimID ecs.EntityID
		switch aiComp.Behavior {
		case component.BehaviorStationary:
			// never moves
		case component.BehaviorCowardly:
			attacked, res, glyph, victimID = cowardlyMove(w, gmap, id, posComp, targetPos, aiComp, rng)
		default:
			attacked, res, glyph, victimID = chaseMove(w, gmap, id, posComp, targetPos, aiComp, rng)
		}
		if attacked {
			hits = append(hits, EnemyHitResult{
				EnemyGlyph:     glyph,
				AttackerID:     id,
				VictimID:       victimID,
				SpecialApplied: res.SpecialApplied,
				DrainedAmount:  res.DrainedAmount,
				Damage:         res.Damage,
			})
		}
	}
	return hits
}

// nearestPlayer returns the ID and position of the player from playerIDs
// that is closest to enemyPos and within sightRange.
// Returns ecs.NilEntity and zero Position if none qualify.
func nearestPlayer(w *ecs.World, playerIDs []ecs.EntityID,
	enemyPos component.Position, sightRange int) (ecs.EntityID, component.Position, bool) {
	best := ecs.NilEntity
	var bestPos component.Position
	bestDist := math.MaxFloat64
	for _, pid := range playerIDs {
		if pid == ecs.NilEntity {
			continue
		}
		pc := w.Get(pid, component.CPosition)
		if pc == nil {
			continue
		}
		pos := pc.(component.Position)
		dx := float64(pos.X - enemyPos.X)
		dy := float64(pos.Y - enemyPos.Y)
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist <= float64(sightRange) && dist < bestDist {
			best = pid
			bestPos = pos
			bestDist = dist
		}
	}
	return best, bestPos, best != ecs.NilEntity
}

func chaseMove(w *ecs.World, gmap *gamemap.GameMap, id ecs.EntityID,
	pos component.Position, playerPos component.Position, ai component.AI, rng *rand.Rand) (bool, AttackResult, string, ecs.EntityID) {

	dx := playerPos.X - pos.X
	dy := playerPos.Y - pos.Y
	dist := math.Sqrt(float64(dx*dx + dy*dy))

	if dist > float64(ai.SightRange) {
		return false, AttackResult{}, "", ecs.NilEntity
	}

	// Normalize to unit step.
	stepX, stepY := sign(dx), sign(dy)

	result, target := TryMove(w, gmap, id, stepX, 0)
	if result == MoveAttack {
		if w.Has(target, component.CTagPlayer) {
			glyph := enemyGlyph(w, id)
			res := Attack(w, rng, id, target)
			return true, res, glyph, target
		}
		return false, AttackResult{}, "", ecs.NilEntity
	}
	if result == MoveOK {
		return false, AttackResult{}, "", ecs.NilEntity
	}
	// Try vertical.
	result, target = TryMove(w, gmap, id, 0, stepY)
	if result == MoveAttack {
		if w.Has(target, component.CTagPlayer) {
			glyph := enemyGlyph(w, id)
			res := Attack(w, rng, id, target)
			return true, res, glyph, target
		}
	}
	return false, AttackResult{}, "", ecs.NilEntity
}

func cowardlyMove(w *ecs.World, gmap *gamemap.GameMap, id ecs.EntityID,
	pos component.Position, playerPos component.Position, ai component.AI, rng *rand.Rand) (bool, AttackResult, string, ecs.EntityID) {

	// Cowardly: attack if adjacent, otherwise flee.
	dx := playerPos.X - pos.X
	dy := playerPos.Y - pos.Y
	dist := math.Sqrt(float64(dx*dx + dy*dy))
	if dist > float64(ai.SightRange) {
		return false, AttackResult{}, "", ecs.NilEntity
	}

	if dist <= 1.5 {
		// Adjacent — find player entity and attack.
		for _, playerEnt := range w.Query(component.CTagPlayer) {
			glyph := enemyGlyph(w, id)
			res := Attack(w, rng, id, playerEnt)
			return true, res, glyph, playerEnt
		}
		return false, AttackResult{}, "", ecs.NilEntity
	}

	// Flee: move away from player.
	stepX, stepY := -sign(dx), -sign(dy)
	if TryMoveSimple(w, gmap, id, stepX, 0) == MoveOK {
		return false, AttackResult{}, "", ecs.NilEntity
	}
	TryMoveSimple(w, gmap, id, 0, stepY)
	return false, AttackResult{}, "", ecs.NilEntity
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
