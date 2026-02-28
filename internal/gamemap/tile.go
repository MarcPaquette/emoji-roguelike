package gamemap

// TileKind identifies the type of a map tile.
type TileKind uint8

const (
	TileWall       TileKind = iota
	TileFloor
	TileDoor
	TileStairsUp
	TileStairsDown
	TileGrass  // walkable outdoor terrain (parks, fields)
	TileWater  // non-walkable water (rivers, lakes)
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

// MakeDoor returns a closed door tile (blocks movement and sight).
func MakeDoor() Tile {
	return Tile{Kind: TileDoor, Walkable: false, Transparent: false}
}

// MakeStairsDown returns a downward staircase tile.
func MakeStairsDown() Tile {
	return Tile{Kind: TileStairsDown, Walkable: true, Transparent: true}
}

// MakeStairsUp returns an upward staircase tile.
func MakeStairsUp() Tile {
	return Tile{Kind: TileStairsUp, Walkable: true, Transparent: true}
}

// MakeGrass returns a walkable, transparent grass tile (outdoor terrain).
func MakeGrass() Tile {
	return Tile{Kind: TileGrass, Walkable: true, Transparent: true}
}

// MakeWater returns a non-walkable, transparent water tile (rivers, lakes).
func MakeWater() Tile {
	return Tile{Kind: TileWater, Walkable: false, Transparent: true}
}
