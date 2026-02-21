package component

import "emoji-roguelike/internal/ecs"

const CLoot ecs.ComponentType = 14

// LootEntry describes one possible item drop with a percentage chance.
type LootEntry struct {
	Glyph  string
	Chance int // 0–100
}

// Loot holds the drop table for an entity.
type Loot struct {
	Drops []LootEntry
}

func (Loot) Type() ecs.ComponentType { return CLoot }
