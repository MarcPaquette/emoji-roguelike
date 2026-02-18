package render

// Camera translates between world coordinates and screen coordinates.
// World X is multiplied by 2 because emoji occupy 2 terminal columns.
type Camera struct {
	OffsetX   int
	OffsetY   int
	ViewWidth int // in terminal columns
	ViewHeight int // in terminal rows
}

// NewCamera creates a camera centered on (cx, cy).
func NewCamera(cx, cy, viewW, viewH int) *Camera {
	c := &Camera{ViewWidth: viewW, ViewHeight: viewH}
	c.Center(cx, cy)
	return c
}

// Center repositions the camera so that world position (cx, cy) is in the middle.
func (c *Camera) Center(cx, cy int) {
	// ViewWidth is in columns; each world tile is 2 columns wide.
	c.OffsetX = cx - (c.ViewWidth/2)/2
	c.OffsetY = cy - c.ViewHeight/2
}

// WorldToScreen converts world (wx, wy) to screen (sx, sy).
// visible is false when the result falls outside the viewport.
func (c *Camera) WorldToScreen(wx, wy int) (sx, sy int, visible bool) {
	sx = (wx - c.OffsetX) * 2
	sy = wy - c.OffsetY
	visible = sx >= 0 && sx < c.ViewWidth && sy >= 0 && sy < c.ViewHeight
	return
}

// ScreenToWorld converts screen (sx, sy) to world coordinates.
func (c *Camera) ScreenToWorld(sx, sy int) (int, int) {
	return sx/2 + c.OffsetX, sy + c.OffsetY
}
