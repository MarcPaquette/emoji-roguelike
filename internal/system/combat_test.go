package system

import (
	"emoji-rougelike/internal/component"
	"emoji-rougelike/internal/ecs"
	"math/rand"
	"testing"
)

func makeCombatants(atkVal, defVal, defHP int) (*ecs.World, ecs.EntityID, ecs.EntityID) {
	w := ecs.NewWorld()
	attacker := w.CreateEntity()
	w.Add(attacker, component.Combat{Attack: atkVal, Defense: 0})

	defender := w.CreateEntity()
	w.Add(defender, component.Combat{Attack: 0, Defense: defVal})
	w.Add(defender, component.Health{Current: defHP, Max: defHP})
	return w, attacker, defender
}

func TestAttackDamageRange(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < 50; i++ {
		// Fresh defender each iteration so it never dies.
		w, attacker, defender := makeCombatants(5, 2, 1000)
		hpBefore := w.Get(defender, component.CHealth).(component.Health).Current
		res := Attack(w, rng, attacker, defender)
		// Damage = max(1, 5-2) + rand.Intn(3) = 3 + [0,2] → [3,5]
		if res.Damage < 3 || res.Damage > 5 {
			t.Errorf("iteration %d: damage %d out of expected range [3,5]", i, res.Damage)
		}
		hpAfter := w.Get(defender, component.CHealth).(component.Health).Current
		if hpAfter != hpBefore-res.Damage {
			t.Errorf("HP not reduced correctly: before=%d after=%d damage=%d", hpBefore, hpAfter, res.Damage)
		}
		_ = attacker
	}
}

func TestAttackKillsDefender(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	// defender has 1 HP — any hit kills
	w, attacker, defender := makeCombatants(10, 0, 1)

	res := Attack(w, rng, attacker, defender)
	if !res.Killed {
		t.Fatal("expected Killed=true when defender HP reaches 0")
	}
	if w.Alive(defender) {
		t.Fatal("expected defender to be destroyed after kill")
	}
}
