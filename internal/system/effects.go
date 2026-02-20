package system

import (
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
)

// TickEffects decrements all active effects by one turn and removes expired ones.
func TickEffects(w *ecs.World) {
	for _, id := range w.Query(component.CEffects) {
		eff := w.Get(id, component.CEffects).(component.Effects)
		active := eff.Active[:0]
		for _, e := range eff.Active {
			e.TurnsRemaining--
			if e.TurnsRemaining > 0 {
				active = append(active, e)
			}
		}
		eff.Active = active
		w.Add(id, eff)
	}
}

// ApplyEffect adds an effect to an entity, stacking if the kind already exists.
func ApplyEffect(w *ecs.World, id ecs.EntityID, eff component.ActiveEffect) {
	effs := component.Effects{}
	if c := w.Get(id, component.CEffects); c != nil {
		effs = c.(component.Effects)
	}
	// Replace existing effect of same kind if new duration is longer.
	for i, e := range effs.Active {
		if e.Kind == eff.Kind {
			if eff.TurnsRemaining > e.TurnsRemaining {
				effs.Active[i] = eff
			}
			w.Add(id, effs)
			return
		}
	}
	effs.Active = append(effs.Active, eff)
	w.Add(id, effs)
}

// HasEffect reports whether an entity currently has an effect of the given kind.
func HasEffect(w *ecs.World, id ecs.EntityID, kind component.EffectKind) bool {
	c := w.Get(id, component.CEffects)
	if c == nil {
		return false
	}
	for _, e := range c.(component.Effects).Active {
		if e.Kind == kind {
			return true
		}
	}
	return false
}

// GetAttackBonus returns the net attack modifier from active effects
// (EffectAttackBoost adds, EffectWeaken subtracts).
func GetAttackBonus(w *ecs.World, id ecs.EntityID) int {
	c := w.Get(id, component.CEffects)
	if c == nil {
		return 0
	}
	total := 0
	for _, e := range c.(component.Effects).Active {
		switch e.Kind {
		case component.EffectAttackBoost:
			total += e.Magnitude
		case component.EffectWeaken:
			total -= e.Magnitude
		}
	}
	return total
}

// GetDefenseBonus returns the net defense modifier from active EffectDefenseBoost effects.
func GetDefenseBonus(w *ecs.World, id ecs.EntityID) int {
	c := w.Get(id, component.CEffects)
	if c == nil {
		return 0
	}
	total := 0
	for _, e := range c.(component.Effects).Active {
		if e.Kind == component.EffectDefenseBoost {
			total += e.Magnitude
		}
	}
	return total
}

// GetPoisonDamage returns the total poison damage per turn from active effects.
func GetPoisonDamage(w *ecs.World, id ecs.EntityID) int {
	c := w.Get(id, component.CEffects)
	if c == nil {
		return 0
	}
	total := 0
	for _, e := range c.(component.Effects).Active {
		if e.Kind == component.EffectPoison {
			total += e.Magnitude
		}
	}
	return total
}
