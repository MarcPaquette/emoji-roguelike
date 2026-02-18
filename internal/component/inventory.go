package component

import "emoji-roguelike/internal/ecs"

const CInventory ecs.ComponentType = 6

type Inventory struct {
	Items    []ecs.EntityID
	Capacity int
}

func (Inventory) Type() ecs.ComponentType { return CInventory }
