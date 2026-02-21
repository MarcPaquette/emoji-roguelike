package gamemap

import "testing"

func TestInBounds(t *testing.T) {
	m := New(10, 8)
	cases := []struct {
		x, y    int
		want    bool
	}{
		{0, 0, true},
		{9, 7, true},
		{-1, 0, false},
		{10, 0, false},
		{0, 8, false},
	}
	for _, c := range cases {
		got := m.InBounds(c.x, c.y)
		if got != c.want {
			t.Errorf("InBounds(%d,%d)=%v, want %v", c.x, c.y, got, c.want)
		}
	}
}

func TestIsWalkable(t *testing.T) {
	m := New(5, 5)
	// all walls initially
	if m.IsWalkable(2, 2) {
		t.Error("wall tile should not be walkable")
	}
	m.Set(2, 2, MakeFloor())
	if !m.IsWalkable(2, 2) {
		t.Error("floor tile should be walkable")
	}
	// out of bounds
	if m.IsWalkable(-1, 0) {
		t.Error("out-of-bounds should not be walkable")
	}
}

func TestRectCenter(t *testing.T) {
	r := Rect{X1: 0, Y1: 0, X2: 4, Y2: 4}
	cx, cy := r.Center()
	if cx != 2 || cy != 2 {
		t.Errorf("expected center (2,2), got (%d,%d)", cx, cy)
	}
}

func TestRectIntersects(t *testing.T) {
	a := Rect{0, 0, 4, 4}
	b := Rect{3, 3, 7, 7}
	c := Rect{5, 5, 9, 9}
	if !a.Intersects(b) {
		t.Error("a and b should intersect")
	}
	if a.Intersects(c) {
		t.Error("a and c should not intersect")
	}
}

func TestAt(t *testing.T) {
	m := New(5, 5)
	// Default tiles are walls; At returns a pointer into the map.
	if m.At(2, 3).Kind != TileWall {
		t.Fatal("expected TileWall at (2,3) before any Set")
	}
	m.Set(2, 3, MakeFloor())
	if m.At(2, 3).Kind != TileFloor {
		t.Fatal("Set should be reflected by subsequent At")
	}
}

func TestIsTransparent(t *testing.T) {
	cases := []struct {
		name string
		tile Tile
		x, y int
		want bool
	}{
		{"wall is opaque", MakeWall(), 2, 2, false},
		{"floor is transparent", MakeFloor(), 2, 2, true},
		{"out-of-bounds x=-1", MakeWall(), -1, 0, false},
		{"out-of-bounds y=-1", MakeWall(), 0, -1, false},
		{"out-of-bounds beyond width", MakeWall(), 10, 2, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := New(5, 5)
			if tc.x >= 0 && tc.y >= 0 && tc.x < 5 && tc.y < 5 {
				m.Set(tc.x, tc.y, tc.tile)
			}
			if got := m.IsTransparent(tc.x, tc.y); got != tc.want {
				t.Errorf("IsTransparent(%d,%d) = %v; want %v", tc.x, tc.y, got, tc.want)
			}
		})
	}
}
