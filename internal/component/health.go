package component

import "emoji-rougelike/internal/ecs"

const CHealth ecs.ComponentType = 2

type Health struct {
	Current, Max int
}

func (Health) Type() ecs.ComponentType { return CHealth }
