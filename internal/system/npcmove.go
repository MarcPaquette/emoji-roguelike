package system

import (
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/gamemap"
	"math/rand"
)

// ProcessNPCMovement advances all NPC movement schedules and positions for one tick.
// dayTick is GameTick % DayCycleTicks.
func ProcessNPCMovement(w *ecs.World, gmap *gamemap.GameMap, dayTick int, rng *rand.Rand) {
	for _, id := range w.Query(component.CNPC, component.CNPCMovement, component.CPosition) {
		mc := w.Get(id, component.CNPCMovement).(component.NPCMovement)
		pos := w.Get(id, component.CPosition).(component.Position)

		// Check schedule transition.
		newIdx := activeScheduleEntry(mc.Schedule, dayTick)
		if newIdx != mc.ActiveIndex {
			transitionSchedule(&mc, pos, newIdx)
		}

		// Speed gate: only move every MoveInterval ticks.
		mc.TickCounter++
		if mc.TickCounter < mc.MoveInterval {
			w.Add(id, mc)
			continue
		}
		mc.TickCounter = 0

		var dx, dy int
		switch mc.Behavior {
		case component.MoveStationary:
			// No movement.
		case component.MoveWander:
			dx, dy = wanderStep(mc, pos, gmap, rng)
		case component.MovePath:
			dx, dy = pathStep(&mc, pos)
		case component.MoveReturn:
			dx, dy = returnStep(&mc, pos)
		}

		if dx != 0 || dy != 0 {
			if tryMoveNPC(w, gmap, id, dx, dy) {
				mc.StuckCount = 0
			} else {
				mc.StuckCount++
				// Try alternate axis if primary blocked.
				altDX, altDY := 0, 0
				if dx != 0 && dy == 0 {
					altDY = rng.Intn(2)*2 - 1 // -1 or +1
				} else if dy != 0 && dx == 0 {
					altDX = rng.Intn(2)*2 - 1
				}
				if altDX != 0 || altDY != 0 {
					if tryMoveNPC(w, gmap, id, altDX, altDY) {
						mc.StuckCount = 0
					}
				}
				// Perpendicular jitter after 3+ stuck ticks.
				if mc.StuckCount >= 3 {
					jdx, jdy := 0, 0
					if dx != 0 {
						jdy = rng.Intn(2)*2 - 1
					} else {
						jdx = rng.Intn(2)*2 - 1
					}
					if tryMoveNPC(w, gmap, id, jdx, jdy) {
						mc.StuckCount = 0
					}
				}
			}
		}

		w.Add(id, mc)
	}
}

// tryMoveNPC attempts to move an NPC by (dx, dy).
// Returns false if destination is occupied by a blocking entity or not walkable.
func tryMoveNPC(w *ecs.World, gmap *gamemap.GameMap, id ecs.EntityID, dx, dy int) bool {
	posComp := w.Get(id, component.CPosition)
	if posComp == nil {
		return false
	}
	pos := posComp.(component.Position)
	nx, ny := pos.X+dx, pos.Y+dy

	if !gmap.IsWalkable(nx, ny) {
		return false
	}

	// Check for any blocking entity at destination.
	for _, eid := range w.Query(component.CTagBlocking, component.CPosition) {
		if eid == id {
			continue
		}
		epos := w.Get(eid, component.CPosition).(component.Position)
		if epos.X == nx && epos.Y == ny {
			return false
		}
	}

	w.Add(id, component.Position{X: nx, Y: ny})
	return true
}

// wanderStep picks a random cardinal direction within bounds.
func wanderStep(mc component.NPCMovement, pos component.Position, gmap *gamemap.GameMap, rng *rand.Rand) (int, int) {
	dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	d := dirs[rng.Intn(4)]
	nx, ny := pos.X+d[0], pos.Y+d[1]
	// Stay within wander bounds.
	if nx < mc.BoundsX1 || nx > mc.BoundsX2 || ny < mc.BoundsY1 || ny > mc.BoundsY2 {
		return 0, 0
	}
	if !gmap.IsWalkable(nx, ny) {
		return 0, 0
	}
	return d[0], d[1]
}

// pathStep computes a greedy step toward the current waypoint.
func pathStep(mc *component.NPCMovement, pos component.Position) (int, int) {
	if mc.PathDone || len(mc.Waypoints) == 0 {
		return 0, 0
	}
	wp := mc.Waypoints[mc.WaypointIdx]
	dx := sign(wp.X - pos.X)
	dy := sign(wp.Y - pos.Y)
	if dx == 0 && dy == 0 {
		// Arrived at waypoint.
		mc.WaypointIdx++
		if mc.WaypointIdx >= len(mc.Waypoints) {
			mc.PathDone = true
			return 0, 0
		}
		wp = mc.Waypoints[mc.WaypointIdx]
		dx = sign(wp.X - pos.X)
		dy = sign(wp.Y - pos.Y)
	}
	// Prefer axis with larger delta; move one axis at a time.
	adx := abs(wp.X - pos.X)
	ady := abs(wp.Y - pos.Y)
	if adx >= ady && dx != 0 {
		return dx, 0
	}
	if dy != 0 {
		return 0, dy
	}
	return dx, 0
}

// returnStep computes a greedy step toward the return target.
func returnStep(mc *component.NPCMovement, pos component.Position) (int, int) {
	dx := sign(mc.ReturnX - pos.X)
	dy := sign(mc.ReturnY - pos.Y)
	if dx == 0 && dy == 0 {
		// Arrived at return target.
		mc.Behavior = mc.ReturnThen
		return 0, 0
	}
	adx := abs(mc.ReturnX - pos.X)
	ady := abs(mc.ReturnY - pos.Y)
	if adx >= ady && dx != 0 {
		return dx, 0
	}
	if dy != 0 {
		return 0, dy
	}
	return dx, 0
}

// activeScheduleEntry returns the index of the schedule entry active at dayTick.
// The schedule wraps: if dayTick < first entry's StartTick, the last entry is active.
func activeScheduleEntry(schedule []component.ScheduleEntry, dayTick int) int {
	if len(schedule) == 0 {
		return 0
	}
	idx := len(schedule) - 1
	for i, e := range schedule {
		if dayTick >= e.StartTick {
			idx = i
		}
	}
	return idx
}

// transitionSchedule switches the NPC to a new schedule entry, potentially via MoveReturn.
func transitionSchedule(mc *component.NPCMovement, pos component.Position, newIdx int) {
	mc.ActiveIndex = newIdx
	entry := mc.Schedule[newIdx]

	switch entry.Behavior {
	case component.MoveStationary:
		// If NPC is far from stand position, return first.
		dist := abs(pos.X-entry.StandX) + abs(pos.Y-entry.StandY)
		if dist > 1 {
			mc.Behavior = component.MoveReturn
			mc.ReturnX = entry.StandX
			mc.ReturnY = entry.StandY
			mc.ReturnThen = component.MoveStationary
		} else {
			mc.Behavior = component.MoveStationary
		}

	case component.MoveWander:
		mc.BoundsX1 = entry.BoundsX1
		mc.BoundsY1 = entry.BoundsY1
		mc.BoundsX2 = entry.BoundsX2
		mc.BoundsY2 = entry.BoundsY2
		// If NPC is outside bounds, return to center first.
		if pos.X < entry.BoundsX1 || pos.X > entry.BoundsX2 ||
			pos.Y < entry.BoundsY1 || pos.Y > entry.BoundsY2 {
			mc.Behavior = component.MoveReturn
			mc.ReturnX = (entry.BoundsX1 + entry.BoundsX2) / 2
			mc.ReturnY = (entry.BoundsY1 + entry.BoundsY2) / 2
			mc.ReturnThen = component.MoveWander
		} else {
			mc.Behavior = component.MoveWander
		}

	case component.MovePath:
		mc.Waypoints = entry.Waypoints
		mc.WaypointIdx = 0
		mc.PathDone = false
		mc.Behavior = component.MovePath
	}

	mc.StuckCount = 0
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
