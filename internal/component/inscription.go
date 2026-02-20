package component

import "emoji-roguelike/internal/ecs"

const CInscription ecs.ComponentType = 12

// Inscription holds text etched onto a wall or floor tile.
type Inscription struct {
	Text string
}

func (Inscription) Type() ecs.ComponentType { return CInscription }
