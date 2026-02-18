package system

import (
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/gamemap"
)

// octant transform matrices.
// For each octant, a (dx, dy) sweep pair maps to a world offset via:
//   worldX = cx + dx*xx + dy*xy
//   worldY = cy + dx*yx + dy*yy
// where dx sweeps horizontally within the row and dy is the fixed row index.
// These match the standard RogueBasin recursive shadowcasting multipliers.
var octants = [8][4]int{
	{1, 0, 0, 1},
	{0, 1, 1, 0},
	{0, -1, 1, 0},
	{-1, 0, 0, 1},
	{-1, 0, 0, -1},
	{0, -1, -1, 0},
	{0, 1, -1, 0},
	{1, 0, 0, -1},
}

// UpdateFOV resets visibility and runs recursive shadowcasting from the player.
func UpdateFOV(w *ecs.World, gmap *gamemap.GameMap, playerID ecs.EntityID, radius int) {
	// Clear current visibility.
	for y := 0; y < gmap.Height; y++ {
		for x := 0; x < gmap.Width; x++ {
			gmap.At(x, y).Visible = false
		}
	}

	posComp := w.Get(playerID, component.CPosition)
	if posComp == nil {
		return
	}
	pos := posComp.(component.Position)

	// Origin is always visible.
	if gmap.InBounds(pos.X, pos.Y) {
		t := gmap.At(pos.X, pos.Y)
		t.Visible = true
		t.Explored = true
	}

	// Cast light in all 8 octants.
	for _, m := range octants {
		castLight(gmap, pos.X, pos.Y, 1, 1.0, 0.0, radius, m[0], m[1], m[2], m[3])
	}
}

// castLight casts light for one octant using recursive shadowcasting.
//
// Algorithm (matches Python RogueBasin reference):
//   - j is the current row (distance from origin along the main axis)
//   - dy = -j is fixed for the entire inner sweep (the row coordinate)
//   - dx sweeps from -j to 0 (the column coordinate within the row)
//   - world position: (cx + dx*xx + dy*xy,  cy + dx*yx + dy*yy)
//   - lSlope = (dx - 0.5) / (dy + 0.5)   rSlope = (dx + 0.5) / (dy - 0.5)
func castLight(gmap *gamemap.GameMap, cx, cy, row int, start, end float64, radius, xx, xy, yx, yy int) {
	if start < end {
		return
	}
	radiusSq := float64(radius * radius)
	newStart := start

	for j := row; j <= radius; j++ {
		dy := -j // fixed row index (always negative — moving away from origin)
		blocked := false

		for dx := -j; dx <= 0; dx++ {
			// Map sweep coordinates to world position.
			wx := cx + dx*xx + dy*xy
			wy := cy + dx*yx + dy*yy

			// Slope of the left and right edges of this cell.
			// dy is negative so (dy+0.5) and (dy-0.5) are both negative,
			// making the slopes positive for dx < 0 — slopes decrease toward 0
			// as dx moves right, matching the standard shadowcasting convention.
			lSlope := (float64(dx) - 0.5) / (float64(dy) + 0.5)
			rSlope := (float64(dx) + 0.5) / (float64(dy) - 0.5)

			if start < rSlope {
				continue // cell is to the right of our current shadow beam
			}
			if end > lSlope {
				break // cell is to the left; all remaining cells are too
			}

			// Light this cell if within the radius circle.
			if float64(dx*dx+dy*dy) < radiusSq && gmap.InBounds(wx, wy) {
				t := gmap.At(wx, wy)
				t.Visible = true
				t.Explored = true
			}

			opaque := !gmap.InBounds(wx, wy) || !gmap.IsTransparent(wx, wy)

			if blocked {
				if opaque {
					// Still inside a wall run — advance the shadow boundary.
					newStart = rSlope
				} else {
					// Transitioned wall→open: resume with updated start slope.
					blocked = false
					start = newStart
				}
			} else {
				if opaque && j < radius {
					// Hit a new wall — cast a child scan beyond it.
					blocked = true
					castLight(gmap, cx, cy, j+1, start, lSlope, radius, xx, xy, yx, yy)
					newStart = rSlope
				}
			}
		}
		if blocked {
			break // entire row was wall; no light beyond
		}
	}
}
