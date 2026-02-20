package component

import "emoji-roguelike/internal/ecs"

const CCombat ecs.ComponentType = 4

type Combat struct {
	Attack        int
	Defense       int
	SpecialKind   uint8 // 0=none 1=poison 2=weaken 3=lifedrain
	SpecialChance int   // 0-100 percent
	SpecialMag    int   // poison/weaken magnitude or lifedrain percent*10
	SpecialDur    int   // turns the player effect lasts
}

func (Combat) Type() ecs.ComponentType { return CCombat }
