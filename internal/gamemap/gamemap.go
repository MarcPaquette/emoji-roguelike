package gamemap

// Rect is an axis-aligned rectangle used for rooms.
type Rect struct {
	X1, Y1, X2, Y2 int
}

// Center returns the center point of the rectangle.
func (r Rect) Center() (int, int) {
	return (r.X1 + r.X2) / 2, (r.Y1 + r.Y2) / 2
}

// Intersects reports whether r overlaps other (inclusive edges).
func (r Rect) Intersects(other Rect) bool {
	return r.X1 <= other.X2 && r.X2 >= other.X1 &&
		r.Y1 <= other.Y2 && r.Y2 >= other.Y1
}

// GameMap holds the tile grid and room list for one dungeon level.
type GameMap struct {
	Width, Height int
	Tiles         [][]Tile
	Rooms         []Rect
}

// New creates a GameMap filled with walls.
func New(width, height int) *GameMap {
	tiles := make([][]Tile, height)
	for y := range tiles {
		tiles[y] = make([]Tile, width)
		for x := range tiles[y] {
			tiles[y][x] = MakeWall()
		}
	}
	return &GameMap{Width: width, Height: height, Tiles: tiles}
}

// InBounds reports whether (x, y) is within the map boundaries.
func (m *GameMap) InBounds(x, y int) bool {
	return x >= 0 && x < m.Width && y >= 0 && y < m.Height
}

// At returns a pointer to the tile at (x, y). Panics if out of bounds.
func (m *GameMap) At(x, y int) *Tile {
	return &m.Tiles[y][x]
}

// Set replaces the tile at (x, y).
func (m *GameMap) Set(x, y int, t Tile) {
	m.Tiles[y][x] = t
}

// IsWalkable returns true when (x, y) is in bounds and walkable.
func (m *GameMap) IsWalkable(x, y int) bool {
	if !m.InBounds(x, y) {
		return false
	}
	return m.Tiles[y][x].Walkable
}

// IsTransparent returns true when (x, y) is in bounds and transparent.
func (m *GameMap) IsTransparent(x, y int) bool {
	if !m.InBounds(x, y) {
		return false
	}
	return m.Tiles[y][x].Transparent
}
