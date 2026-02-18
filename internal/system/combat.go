package system

import (
	"emoji-rougelike/internal/component"
	"emoji-rougelike/internal/ecs"
	"math/rand"
)

// AttackResult holds the outcome of one attack.
type AttackResult struct {
	Damage int
	Killed bool
}

// Attack resolves one attack from attacker against defender.
// Damage formula: max(1, atk-def) + rand.Intn(3)
// If defender HP drops to ≤ 0, it is destroyed and Killed=true.
func Attack(w *ecs.World, rng *rand.Rand, attackerID, defenderID ecs.EntityID) AttackResult {
	atkComp := w.Get(attackerID, component.CCombat)
	defComp := w.Get(defenderID, component.CCombat)
	hpComp := w.Get(defenderID, component.CHealth)

	if atkComp == nil || defComp == nil || hpComp == nil {
		return AttackResult{}
	}

	atk := atkComp.(component.Combat).Attack
	def := defComp.(component.Combat).Defense
	hp := hpComp.(component.Health)

	base := atk - def
	if base < 1 {
		base = 1
	}
	dmg := base + rng.Intn(3)

	hp.Current -= dmg
	w.Add(defenderID, hp)

	result := AttackResult{Damage: dmg}
	if hp.Current <= 0 {
		result.Killed = true
		w.DestroyEntity(defenderID)
	}
	return result
}
