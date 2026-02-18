package component

import "emoji-roguelike/internal/ecs"

const CHealth ecs.ComponentType = 2

type Health struct {
	Current, Max int
}

func (Health) Type() ecs.ComponentType { return CHealth }
