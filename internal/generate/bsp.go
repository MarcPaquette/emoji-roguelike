package generate

import (
	"emoji-roguelike/internal/gamemap"
	"math/rand"
)

// CorridorStyle selects the shape of connecting tunnels.
type CorridorStyle uint8

const (
	CorridorLShaped   CorridorStyle = iota
	CorridorZShaped
	CorridorStraight
)

// DropEntry describes one item that may drop from a defeated enemy.
type DropEntry struct {
	Glyph  string
	Chance int // 0–100
}

// EnemySpawnEntry describes one possible enemy spawn with its threat cost.
type EnemySpawnEntry struct {
	Glyph         string
	Name          string
	ThreatCost    int
	Attack        int
	Defense       int
	MaxHP         int
	SightRange    int
	SpecialKind   uint8 // 0=none 1=poison 2=weaken 3=lifedrain 4=stun 5=armorBreak
	SpecialChance int   // 0-100 percent
	SpecialMag    int   // magnitude (poison dmg/turn, weaken atk penalty, lifedrain % * 10, armorBreak DEF penalty)
	SpecialDur    int   // turns the status effect lasts
	Drops         []DropEntry
}

// ItemSpawnEntry describes one possible item spawn.
type ItemSpawnEntry struct {
	Glyph string
	Name  string
}

// EquipSpawnEntry describes one possible equipment spawn.
// Slot uses the same numeric values as component.ItemSlot to avoid a circular import.
type EquipSpawnEntry struct {
	Glyph                        string
	Name                         string
	Slot                         uint8 // 1=Head 2=Body 3=Feet 4=OneHand 5=TwoHand 6=OffHand
	BaseATK, BaseDEF, BaseMaxHP  int
	ATKScale, DEFScale, HPScale  int
	MinFloor                     int
}

// Config drives procedural generation for one floor.
type Config struct {
	MapWidth, MapHeight int
	MinLeafSize         int
	MaxLeafSize         int
	SplitRatio          float64
	MinRoomSize         int
	RoomPadding         int
	CorridorStyle       CorridorStyle
	FloorNumber          int
	EnemyBudget          int
	ItemCount            int
	EquipCount           int
	EnemyTable           []EnemySpawnEntry
	ItemTable            []ItemSpawnEntry
	EquipTable           []EquipSpawnEntry
	InscriptionTexts     []string // pool of wall-writing texts to draw from
	InscriptionCount     int      // how many to place (typically 2-5)
	EliteEnemy           *EnemySpawnEntry // if non-nil, always spawned once in a random placeable room
	Rand                 *rand.Rand
}

// bspLeaf is a node in the BSP tree.
type bspLeaf struct {
	X, Y, W, H  int
	left, right *bspLeaf
	room        *gamemap.Rect
}

// split divides the leaf into two children, returning false when leaf is too small.
func (l *bspLeaf) split(cfg *Config) bool {
	if l.left != nil || l.right != nil {
		return false // already split
	}
	// Decide split direction: horizontal when taller, vertical when wider.
	splitH := cfg.Rand.Intn(2) == 0
	if l.W > l.H && float64(l.W)/float64(l.H) >= 1.25 {
		splitH = false
	} else if l.H > l.W && float64(l.H)/float64(l.W) >= 1.25 {
		splitH = true
	}

	maxSize := l.H
	if !splitH {
		maxSize = l.W
	}
	if maxSize <= cfg.MinLeafSize*2 {
		return false // too small to split
	}

	lo := cfg.MinLeafSize
	hi := maxSize - cfg.MinLeafSize
	if lo >= hi {
		return false
	}
	split := lo + cfg.Rand.Intn(hi-lo+1)

	if splitH {
		l.left = &bspLeaf{X: l.X, Y: l.Y, W: l.W, H: split}
		l.right = &bspLeaf{X: l.X, Y: l.Y + split, W: l.W, H: l.H - split}
	} else {
		l.left = &bspLeaf{X: l.X, Y: l.Y, W: split, H: l.H}
		l.right = &bspLeaf{X: l.X + split, Y: l.Y, W: l.W - split, H: l.H}
	}
	return true
}

// createRooms recursively carves rooms inside terminal leaves.
func (l *bspLeaf) createRooms(gmap *gamemap.GameMap, cfg *Config) {
	if l.left != nil || l.right != nil {
		if l.left != nil {
			l.left.createRooms(gmap, cfg)
		}
		if l.right != nil {
			l.right.createRooms(gmap, cfg)
		}
		return
	}
	// Terminal leaf — place a room.
	pad := cfg.RoomPadding
	minW := cfg.MinRoomSize
	minH := cfg.MinRoomSize

	availW := l.W - 2*pad
	availH := l.H - 2*pad
	if availW < minW {
		availW = minW
	}
	if availH < minH {
		availH = minH
	}

	rw := minW + cfg.Rand.Intn(max(1, availW-minW+1))
	rh := minH + cfg.Rand.Intn(max(1, availH-minH+1))

	// Clamp to leaf bounds
	if rw > l.W-2*pad {
		rw = l.W - 2*pad
	}
	if rh > l.H-2*pad {
		rh = l.H - 2*pad
	}
	if rw < 3 {
		rw = 3
	}
	if rh < 3 {
		rh = 3
	}

	rx := l.X + pad + cfg.Rand.Intn(max(1, l.W-rw-2*pad+1))
	ry := l.Y + pad + cfg.Rand.Intn(max(1, l.H-rh-2*pad+1))

	// Safety clamp to map bounds (leave 1-tile border).
	if rx < 1 {
		rx = 1
	}
	if ry < 1 {
		ry = 1
	}
	if rx+rw >= gmap.Width {
		rw = gmap.Width - rx - 1
	}
	if ry+rh >= gmap.Height {
		rh = gmap.Height - ry - 1
	}
	if rw < 3 || rh < 3 {
		return
	}

	room := gamemap.Rect{X1: rx, Y1: ry, X2: rx + rw - 1, Y2: ry + rh - 1}
	l.room = &room

	// Carve floor tiles.
	for y := room.Y1; y <= room.Y2; y++ {
		for x := room.X1; x <= room.X2; x++ {
			gmap.Set(x, y, gamemap.MakeFloor())
		}
	}
	gmap.Rooms = append(gmap.Rooms, room)
}

// getRoom returns the room nearest to this leaf's center (from children if split).
func (l *bspLeaf) getRoom() *gamemap.Rect {
	if l.room != nil {
		return l.room
	}
	var lRoom, rRoom *gamemap.Rect
	if l.left != nil {
		lRoom = l.left.getRoom()
	}
	if l.right != nil {
		rRoom = l.right.getRoom()
	}
	if lRoom == nil {
		return rRoom
	}
	if rRoom == nil {
		return lRoom
	}
	return lRoom // just pick one
}

// connectChildren carves corridors between the two children of a split leaf.
func (l *bspLeaf) connectChildren(gmap *gamemap.GameMap, cfg *Config) {
	if l.left == nil || l.right == nil {
		return
	}
	l.left.connectChildren(gmap, cfg)
	l.right.connectChildren(gmap, cfg)

	lRoom := l.left.getRoom()
	rRoom := l.right.getRoom()
	if lRoom == nil || rRoom == nil {
		return
	}
	lCX, lCY := lRoom.Center()
	rCX, rCY := rRoom.Center()
	carveCorridor(gmap, lCX, lCY, rCX, rCY, cfg)
}

// Generate runs BSP generation and returns the populated map plus player start.
func Generate(cfg *Config) (*gamemap.GameMap, int, int) {
	gmap := gamemap.New(cfg.MapWidth, cfg.MapHeight)

	root := &bspLeaf{X: 0, Y: 0, W: cfg.MapWidth, H: cfg.MapHeight}

	// Build BSP tree.
	leaves := []*bspLeaf{root}
	splitAny := true
	for splitAny {
		splitAny = false
		var next []*bspLeaf
		for _, leaf := range leaves {
			if leaf.left != nil || leaf.right != nil {
				next = append(next, leaf.left, leaf.right)
				continue
			}
			if leaf.W > cfg.MaxLeafSize || leaf.H > cfg.MaxLeafSize ||
				cfg.Rand.Float64() > 0.25 {
				if leaf.split(cfg) {
					next = append(next, leaf.left, leaf.right)
					splitAny = true
					continue
				}
			}
			next = append(next, leaf)
		}
		leaves = next
	}

	root.createRooms(gmap, cfg)
	root.connectChildren(gmap, cfg)

	// Player starts in center of first room.
	px, py := 1, 1
	if len(gmap.Rooms) > 0 {
		px, py = gmap.Rooms[0].Center()
	}

	// Place stairs down in last room.
	if len(gmap.Rooms) > 1 {
		last := gmap.Rooms[len(gmap.Rooms)-1]
		sx, sy := last.Center()
		gmap.Set(sx, sy, gamemap.MakeStairsDown())
	}

	return gmap, px, py
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
