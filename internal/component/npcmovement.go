package component

import "emoji-roguelike/internal/ecs"

const CNPCMovement ecs.ComponentType = 17

// DayCycleTicks is the total ticks in one game day (6000 × 100ms = 10 real minutes).
const DayCycleTicks = 6000

// MoveBehavior describes how an NPC moves during a schedule period.
type MoveBehavior uint8

const (
	MoveStationary MoveBehavior = iota // stand in place
	MoveWander                         // random walk within bounding box
	MovePath                           // follow waypoints in sequence
	MoveReturn                         // internal: greedy-walk to target, then switch
)

// Waypoint is a single (X, Y) target for path-following.
type Waypoint struct{ X, Y int }

// ScheduleEntry defines one period of NPC behavior within the day cycle.
type ScheduleEntry struct {
	StartTick                           int
	Behavior                            MoveBehavior
	BoundsX1, BoundsY1, BoundsX2, BoundsY2 int // MoveWander
	Waypoints                           []Waypoint  // MovePath
	StandX, StandY                      int         // MoveStationary
}

// NPCMovement holds the full movement state for a scheduled NPC.
type NPCMovement struct {
	Schedule     []ScheduleEntry // sorted by StartTick; wraps at DayCycleTicks
	ActiveIndex  int
	Behavior     MoveBehavior
	TickCounter  int
	MoveInterval int // ticks between moves (10-20 typical)

	// Wander state
	BoundsX1, BoundsY1, BoundsX2, BoundsY2 int

	// Path state
	Waypoints   []Waypoint
	WaypointIdx int
	PathDone    bool

	// Return state (transition repositioning)
	ReturnX, ReturnY int
	ReturnThen       MoveBehavior

	StuckCount int
}

func (NPCMovement) Type() ecs.ComponentType { return CNPCMovement }
