package component

import "emoji-roguelike/internal/ecs"

const CEffects ecs.ComponentType = 7

// EffectKind describes what an active effect does.
type EffectKind uint8

const (
	EffectAttackBoost EffectKind = iota
	EffectInvisible
	EffectRevealMap
)

// ActiveEffect is a timed status applied to an entity.
type ActiveEffect struct {
	Kind          EffectKind
	Magnitude     int
	TurnsRemaining int
}

type Effects struct {
	Active []ActiveEffect
}

func (Effects) Type() ecs.ComponentType { return CEffects }
