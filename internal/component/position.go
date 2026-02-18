package component

import "emoji-roguelike/internal/ecs"

const CPosition ecs.ComponentType = 1

type Position struct {
	X, Y int
}

func (Position) Type() ecs.ComponentType { return CPosition }
