package system

import (
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"math/rand"
)

// AttackResult holds the outcome of one attack.
type AttackResult struct {
	Damage        int
	Killed        bool
	SpecialApplied uint8 // 0=none 1=poison 2=weaken 3=lifedrain
	DrainedAmount  int   // HP healed by lifedrain
}

// equipATKBonus returns total ATK bonus from equipped items (players only).
func equipATKBonus(w *ecs.World, id ecs.EntityID) int {
	c := w.Get(id, component.CInventory)
	if c == nil {
		return 0
	}
	inv := c.(component.Inventory)
	return inv.MainHand.BonusATK + inv.OffHand.BonusATK +
		inv.Head.BonusATK + inv.Body.BonusATK + inv.Feet.BonusATK
}

// equipDEFBonus returns total DEF bonus from equipped items (players only).
func equipDEFBonus(w *ecs.World, id ecs.EntityID) int {
	c := w.Get(id, component.CInventory)
	if c == nil {
		return 0
	}
	inv := c.(component.Inventory)
	return inv.MainHand.BonusDEF + inv.OffHand.BonusDEF +
		inv.Head.BonusDEF + inv.Body.BonusDEF + inv.Feet.BonusDEF
}

// Attack resolves one attack from attacker against defender.
// Damage formula: max(1, atk+bonus-def) + rand.Intn(3)
// If defender HP drops to ≤ 0, it is destroyed and Killed=true.
func Attack(w *ecs.World, rng *rand.Rand, attackerID, defenderID ecs.EntityID) AttackResult {
	atkComp := w.Get(attackerID, component.CCombat)
	defComp := w.Get(defenderID, component.CCombat)
	hpComp := w.Get(defenderID, component.CHealth)

	if atkComp == nil || defComp == nil || hpComp == nil {
		return AttackResult{}
	}

	cbt := atkComp.(component.Combat)
	def := defComp.(component.Combat).Defense
	hp := hpComp.(component.Health)

	atk := cbt.Attack + GetAttackBonus(w, attackerID) + equipATKBonus(w, attackerID)
	def += GetDefenseBonus(w, defenderID) + equipDEFBonus(w, defenderID)
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

	// Special attack: only triggered when defender is alive (not destroyed mid-attack
	// for lifedrain/weaken; poison can still be applied even if lethal hit — rare edge case).
	if cbt.SpecialKind != 0 && cbt.SpecialChance > 0 && rng.Intn(100) < cbt.SpecialChance {
		result.SpecialApplied = cbt.SpecialKind
		switch cbt.SpecialKind {
		case 1: // poison
			ApplyEffect(w, defenderID, component.ActiveEffect{
				Kind:           component.EffectPoison,
				Magnitude:      cbt.SpecialMag,
				TurnsRemaining: cbt.SpecialDur,
			})
		case 2: // weaken
			ApplyEffect(w, defenderID, component.ActiveEffect{
				Kind:           component.EffectWeaken,
				Magnitude:      cbt.SpecialMag,
				TurnsRemaining: cbt.SpecialDur,
			})
		case 3: // lifedrain
			drain := (dmg * cbt.SpecialMag) / 10
			if drain < 1 {
				drain = 1
			}
			result.DrainedAmount = drain
			// Heal the attacker.
			if atkHP := w.Get(attackerID, component.CHealth); atkHP != nil {
				ah := atkHP.(component.Health)
				ah.Current += drain
				if ah.Current > ah.Max {
					ah.Current = ah.Max
				}
				w.Add(attackerID, ah)
			}
		}
	}

	return result
}
