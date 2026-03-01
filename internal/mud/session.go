package mud

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/gamemap"
	"emoji-roguelike/internal/render"
	"sync"
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
)

// playerColors is the round-robin palette for distinguishing players visually.
var playerColors = []tcell.Color{
	tcell.ColorYellow,
	tcell.ColorFuchsia,
	tcell.ColorAqua,
	tcell.ColorLime,
	tcell.ColorOrange,
	tcell.ColorRed,
	tcell.ColorSilver,
	tcell.ColorWhite,
}

// Session holds all per-player state for one SSH connection.
type Session struct {
	ID    int
	Name  string // display name (SSH username or "Player N")
	Color tcell.Color
	Class assets.ClassDef

	// ECS identity — updated on floor transitions.
	PlayerID ecs.EntityID
	FloorNum int

	// Per-player persistent stats (survive floor transitions and respawns).
	FovRadius       int
	BaseMaxHP       int
	FurnitureATK    int
	FurnitureDEF    int
	FurnitureThorns int
	FurnitureKR     bool
	SpecialCooldown int
	Gold            int // current gold; earned by killing enemies, spent at shop

	// PendingNPC is set by the tick goroutine to trigger a shop modal.
	// Read and cleared in RunLoop's RenderCh handler (both under s.mu).
	PendingNPC ecs.EntityID

	// I/O
	Screen   tcell.Screen
	Renderer *render.Renderer

	// Per-player FOV snapshot: FovGrid[y][x] = visible from this player's perspective.
	FovGrid [][]bool

	// Pending action (last key wins).
	actionMu sync.Mutex
	pending  Action

	// Session state.
	Messages          []string
	RunLog            RunLog
	DiscoveredEnemies map[string]bool
	TurnCount         int

	// Render trigger: ticker sends here; session's goroutine drains and renders.
	RenderCh chan struct{}

	// deathCountdown > 0 means the player is dead and waiting to respawn.
	// Decremented each tick; when it reaches 0, respawn fires.
	// Accessed atomically: the tick goroutine writes under s.mu, while
	// the session goroutine reads without s.mu for UI gating.
	deathCountdown atomic.Int32

	// victory is set when the player defeats the final boss.
	// When true, the death countdown shows a victory screen instead of death,
	// and reaching 0 triggers an interactive victory modal instead of auto-respawn.
	victory atomic.Bool
}

// NewSession allocates a Session for a newly-connected player.
func NewSession(id int, name string, color tcell.Color, screen tcell.Screen) *Session {
	return &Session{
		ID:                id,
		Name:              name,
		Color:             color,
		Screen:            screen,
		Messages:          nil,
		DiscoveredEnemies: make(map[string]bool),
		RunLog: RunLog{
			EnemiesKilled: make(map[string]int),
			ItemsUsed:     make(map[string]int),
		},
		RenderCh: make(chan struct{}, 1),
	}
}

// GetDeathCountdown returns the current death countdown value.
// Safe to call from any goroutine.
func (s *Session) GetDeathCountdown() int { return int(s.deathCountdown.Load()) }

// SetDeathCountdown sets the death countdown. Caller should hold s.mu for
// consistency with other session state, but the field itself is atomic.
func (s *Session) SetDeathCountdown(v int) { s.deathCountdown.Store(int32(v)) }

// DecrDeathCountdown atomically decrements the death countdown and returns
// the new value.
func (s *Session) DecrDeathCountdown() int {
	return int(s.deathCountdown.Add(-1))
}

// IsVictory returns whether this session has a pending victory.
func (s *Session) IsVictory() bool { return s.victory.Load() }

// SetVictory marks this session as having achieved victory.
func (s *Session) SetVictory() { s.victory.Store(true) }

// ClearVictory clears the victory flag.
func (s *Session) ClearVictory() { s.victory.Store(false) }

// SetAction stores the player's most recent key action (last key wins).
func (s *Session) SetAction(a Action) {
	s.actionMu.Lock()
	s.pending = a
	s.actionMu.Unlock()
}

// TakeAction atomically retrieves and clears the pending action.
func (s *Session) TakeAction() Action {
	s.actionMu.Lock()
	a := s.pending
	s.pending = ActionNone
	s.actionMu.Unlock()
	return a
}

// AddMessage appends a message to the session's log, capping at 50 entries.
func (s *Session) AddMessage(msg string) {
	s.Messages = append(s.Messages, msg)
	if len(s.Messages) > 50 {
		s.Messages = s.Messages[len(s.Messages)-50:]
	}
}

// SnapshotFOV saves the current gmap.Tile.Visible state into s.FovGrid.
// Call this right after system.UpdateFOV, before the gmap state is clobbered
// by another player's FOV update.
func (s *Session) SnapshotFOV(gmap *gamemap.GameMap) {
	if len(s.FovGrid) != gmap.Height {
		s.FovGrid = make([][]bool, gmap.Height)
		for y := range s.FovGrid {
			s.FovGrid[y] = make([]bool, gmap.Width)
		}
	}
	for y := range gmap.Height {
		for x := range gmap.Width {
			s.FovGrid[y][x] = gmap.At(x, y).Visible
		}
	}
}

// ApplyFOV writes s.FovGrid back into gmap.Tile.Visible so that the renderer
// sees this player's field of view.
func (s *Session) ApplyFOV(gmap *gamemap.GameMap) {
	if s.FovGrid == nil {
		return
	}
	for y := range gmap.Height {
		for x := range gmap.Width {
			if y < len(s.FovGrid) && x < len(s.FovGrid[y]) {
				gmap.At(x, y).Visible = s.FovGrid[y][x]
			} else {
				gmap.At(x, y).Visible = false
			}
		}
	}
}
