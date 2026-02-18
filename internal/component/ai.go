package component

import "emoji-roguelike/internal/ecs"

const CAI ecs.ComponentType = 5

// AIBehavior describes how an enemy acts each turn.
type AIBehavior uint8

const (
	BehaviorChase    AIBehavior = iota // move toward player, attack if adjacent
	BehaviorCowardly                   // flee when hurt
	BehaviorStationary                 // never moves
)

type AI struct {
	Behavior   AIBehavior
	SightRange int
}

func (AI) Type() ecs.ComponentType { return CAI }
