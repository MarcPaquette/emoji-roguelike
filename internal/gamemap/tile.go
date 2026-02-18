package gamemap

// TileKind identifies the type of a map tile.
type TileKind uint8

const (
	TileWall       TileKind = iota
	TileFloor
	TileDoor
	TileStairsUp
	TileStairsDown
)

// Tile holds the kind and visibility state for one map cell.
type Tile struct {
	Kind        TileKind
	Walkable    bool
	Transparent bool
	Explored    bool
	Visible     bool
}

// MakeWall returns a blocking, opaque wall tile.
func MakeWall() Tile {
	return Tile{Kind: TileWall, Walkable: false, Transparent: false}
}

// MakeFloor returns a passable, transparent floor tile.
func MakeFloor() Tile {
	return Tile{Kind: TileFloor, Walkable: true, Transparent: true}
}

// MakeDoor returns a door tile (treated as floor for now).
func MakeDoor() Tile {
	return Tile{Kind: TileDoor, Walkable: true, Transparent: false}
}

// MakeStairsDown returns a downward staircase tile.
func MakeStairsDown() Tile {
	return Tile{Kind: TileStairsDown, Walkable: true, Transparent: true}
}

// MakeStairsUp returns an upward staircase tile.
func MakeStairsUp() Tile {
	return Tile{Kind: TileStairsUp, Walkable: true, Transparent: true}
}
