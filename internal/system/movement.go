package system

import (
	"emoji-rougelike/internal/component"
	"emoji-rougelike/internal/ecs"
	"emoji-rougelike/internal/gamemap"
)

// MoveResult describes the outcome of a TryMove call.
type MoveResult uint8

const (
	MoveOK      MoveResult = iota // position updated
	MoveBlocked                   // wall or out-of-bounds
	MoveAttack                    // bumped a blocking entity
)

// TryMove attempts to move entity id by (dx, dy) on gmap.
// Returns the outcome and (if MoveAttack) the target entity.
func TryMove(w *ecs.World, gmap *gamemap.GameMap, id ecs.EntityID, dx, dy int) (MoveResult, ecs.EntityID) {
	posComp := w.Get(id, component.CPosition)
	if posComp == nil {
		return MoveBlocked, ecs.NilEntity
	}
	pos := posComp.(component.Position)
	nx, ny := pos.X+dx, pos.Y+dy

	// Check for blocking entities at destination.
	for _, other := range w.Query(component.CTagBlocking, component.CPosition) {
		if other == id {
			continue
		}
		otherPos := w.Get(other, component.CPosition).(component.Position)
		if otherPos.X == nx && otherPos.Y == ny {
			return MoveAttack, other
		}
	}

	// Check map walkability.
	if !gmap.IsWalkable(nx, ny) {
		return MoveBlocked, ecs.NilEntity
	}

	// Move.
	w.Add(id, component.Position{X: nx, Y: ny})
	return MoveOK, ecs.NilEntity
}
