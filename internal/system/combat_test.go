package system

import (
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
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

func TestAttackMissingComponents(t *testing.T) {
	// Attack returns an empty result when the attacker has no CCombat component.
	rng := rand.New(rand.NewSource(0))
	w := ecs.NewWorld()
	attacker := w.CreateEntity() // no CCombat
	defender := w.CreateEntity()
	w.Add(defender, component.Combat{Attack: 3, Defense: 1})
	w.Add(defender, component.Health{Current: 10, Max: 10})

	res := Attack(w, rng, attacker, defender)
	if res.Damage != 0 || res.Killed {
		t.Errorf("expected zero-value result for missing attacker component; got %+v", res)
	}
	hp := w.Get(defender, component.CHealth).(component.Health)
	if hp.Current != 10 {
		t.Errorf("defender HP should be unchanged; got %d", hp.Current)
	}
}

func TestAttackMinDamageIsOne(t *testing.T) {
	// When atk ≤ def, base is clamped to 1: damage = 1 + rand.Intn(3) → [1,3].
	rng := rand.New(rand.NewSource(7))
	for i := 0; i < 50; i++ {
		w, attacker, defender := makeCombatants(2, 10, 1000)
		res := Attack(w, rng, attacker, defender)
		if res.Damage < 1 || res.Damage > 3 {
			t.Errorf("iteration %d: damage %d out of range [1,3] when atk<def", i, res.Damage)
		}
	}
}

func TestEquipmentAttackBonus(t *testing.T) {
	// Equipment ATK bonus is factored into the attacker's effective attack.
	// attacker base ATK=3, weapon BonusATK=4, defender DEF=0:
	// damage = max(1, 7-0) + rand.Intn(3) → [7,9].
	rng := rand.New(rand.NewSource(0))
	for i := 0; i < 30; i++ {
		w := ecs.NewWorld()
		atk := w.CreateEntity()
		w.Add(atk, component.Combat{Attack: 3, Defense: 0})
		w.Add(atk, component.Inventory{MainHand: component.Item{BonusATK: 4}})
		def := w.CreateEntity()
		w.Add(def, component.Combat{Attack: 0, Defense: 0})
		w.Add(def, component.Health{Current: 1000, Max: 1000})

		res := Attack(w, rng, atk, def)
		if res.Damage < 7 || res.Damage > 9 {
			t.Errorf("iteration %d: damage %d out of range [7,9] with equipment ATK bonus", i, res.Damage)
		}
	}
}

func TestEquipmentDefenseBonus(t *testing.T) {
	// Equipment DEF bonus reduces incoming damage.
	// attacker ATK=10, defender base DEF=0, body armor BonusDEF=8:
	// damage = max(1, 10-8) + rand.Intn(3) → [2,4].
	rng := rand.New(rand.NewSource(5))
	for i := 0; i < 30; i++ {
		w := ecs.NewWorld()
		atk := w.CreateEntity()
		w.Add(atk, component.Combat{Attack: 10, Defense: 0})
		def := w.CreateEntity()
		w.Add(def, component.Combat{Attack: 0, Defense: 0})
		w.Add(def, component.Health{Current: 1000, Max: 1000})
		w.Add(def, component.Inventory{Body: component.Item{BonusDEF: 8}})

		res := Attack(w, rng, atk, def)
		if res.Damage < 2 || res.Damage > 4 {
			t.Errorf("iteration %d: damage %d out of range [2,4] with DEF bonus 8", i, res.Damage)
		}
	}
}

func TestSpecialAttackPoison(t *testing.T) {
	// SpecialKind=1 at 100% chance applies EffectPoison to the defender.
	rng := rand.New(rand.NewSource(0))
	w := ecs.NewWorld()
	atk := w.CreateEntity()
	w.Add(atk, component.Combat{Attack: 5, Defense: 0, SpecialKind: 1, SpecialChance: 100, SpecialMag: 3, SpecialDur: 5})
	def := w.CreateEntity()
	w.Add(def, component.Combat{Attack: 0, Defense: 0})
	w.Add(def, component.Health{Current: 1000, Max: 1000})
	w.Add(def, component.Effects{})

	res := Attack(w, rng, atk, def)
	if res.SpecialApplied != 1 {
		t.Errorf("expected SpecialApplied=1 (poison), got %d", res.SpecialApplied)
	}
	if !HasEffect(w, def, component.EffectPoison) {
		t.Error("expected EffectPoison to be applied to the defender")
	}
}

func TestSpecialAttackLifedrain(t *testing.T) {
	// SpecialKind=3 at 100% chance heals the attacker.
	rng := rand.New(rand.NewSource(0))
	w := ecs.NewWorld()
	atk := w.CreateEntity()
	w.Add(atk, component.Combat{Attack: 10, Defense: 0, SpecialKind: 3, SpecialChance: 100, SpecialMag: 10})
	w.Add(atk, component.Health{Current: 5, Max: 30}) // attacker starts at 5 HP
	def := w.CreateEntity()
	w.Add(def, component.Combat{Attack: 0, Defense: 0})
	w.Add(def, component.Health{Current: 1000, Max: 1000})

	res := Attack(w, rng, atk, def)
	if res.SpecialApplied != 3 {
		t.Errorf("expected SpecialApplied=3 (lifedrain), got %d", res.SpecialApplied)
	}
	if res.DrainedAmount < 1 {
		t.Errorf("DrainedAmount should be ≥ 1, got %d", res.DrainedAmount)
	}
	atkHP := w.Get(atk, component.CHealth).(component.Health)
	if atkHP.Current <= 5 {
		t.Errorf("attacker HP should have increased from 5 via lifedrain; got %d", atkHP.Current)
	}
	if atkHP.Current > atkHP.Max {
		t.Errorf("attacker HP %d exceeds max %d after lifedrain", atkHP.Current, atkHP.Max)
	}
}
