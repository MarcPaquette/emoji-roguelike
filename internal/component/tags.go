package component

import "emoji-roguelike/internal/ecs"

const (
	CTagPlayer   ecs.ComponentType = 8
	CTagBlocking ecs.ComponentType = 9
	CTagItem     ecs.ComponentType = 10
	CTagStairs   ecs.ComponentType = 11
)

// TagPlayer marks the player-controlled entity.
type TagPlayer struct{}

func (TagPlayer) Type() ecs.ComponentType { return CTagPlayer }

// TagBlocking marks an entity that occupies its tile (blocks movement).
type TagBlocking struct{}

func (TagBlocking) Type() ecs.ComponentType { return CTagBlocking }

// TagItem marks a pickup item on the map.
type TagItem struct{}

func (TagItem) Type() ecs.ComponentType { return CTagItem }

// TagStairs marks staircase entities.
type TagStairs struct{}

func (TagStairs) Type() ecs.ComponentType { return CTagStairs }
