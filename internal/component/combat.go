package component

import "emoji-roguelike/internal/ecs"

const CCombat ecs.ComponentType = 4

type Combat struct {
	Attack  int
	Defense int
}

func (Combat) Type() ecs.ComponentType { return CCombat }
