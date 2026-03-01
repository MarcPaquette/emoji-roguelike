package render

import (
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/gamemap"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Renderer draws the game world onto a tcell screen.
type Renderer struct {
	screen  tcell.Screen
	camera  *Camera
	floor   int // 1-indexed floor number for color selection
}

// NewRenderer creates a Renderer for the given screen.
func NewRenderer(screen tcell.Screen, floor int) *Renderer {
	w, h := screen.Size()
	// Reserve bottom 5 rows for the HUD.
	viewH := h - 5
	return &Renderer{
		screen: screen,
		camera: NewCamera(0, 0, w, viewH),
		floor:  floor,
	}
}

// SetFloor updates the floor theme index.
func (r *Renderer) SetFloor(floor int) { r.floor = floor }

// CenterOn recenters the camera on world position (x, y).
func (r *Renderer) CenterOn(x, y int) { r.camera.Center(x, y) }

// WorldToScreen converts world coordinates to screen coordinates.
// visible is false when the position falls outside the viewport.
func (r *Renderer) WorldToScreen(wx, wy int) (sx, sy int, visible bool) {
	return r.camera.WorldToScreen(wx, wy)
}

// DrawFrame renders tiles, entities, and the HUD.
func (r *Renderer) DrawFrame(w *ecs.World, gmap *gamemap.GameMap, playerID ecs.EntityID) {
	r.screen.Clear()
	r.drawMap(gmap)
	r.drawEntities(w, gmap)
}

// drawMap renders all visible/explored tiles using per-floor emoji glyphs.
func (r *Renderer) drawMap(gmap *gamemap.GameMap) {
	fi := r.floor
	if fi < 0 || fi >= len(TileThemes) {
		fi = 1
	}
	theme := TileThemes[fi]
	style := tcell.StyleDefault.Background(tcell.ColorBlack)

	for y := 0; y < gmap.Height; y++ {
		for x := 0; x < gmap.Width; x++ {
			tile := gmap.At(x, y)
			if !tile.Visible && !tile.Explored {
				continue
			}
			sx, sy, onScreen := r.camera.WorldToScreen(x, y)
			if !onScreen {
				continue
			}

			var glyph string
			if tile.Visible {
				switch tile.Kind {
				case gamemap.TileWall:
					glyph = theme.Wall
				case gamemap.TileFloor:
					glyph = theme.Floor
				case gamemap.TileDoor:
					glyph = "🚪"
				case gamemap.TileStairsDown:
					glyph = "🔽"
				case gamemap.TileStairsUp:
					glyph = "🔼"
				case gamemap.TileGrass:
					glyph = "🟩"
				case gamemap.TileWater:
					glyph = "🟦"
				default:
					glyph = theme.Floor
				}
			} else {
				// Explored but currently dark.
				switch tile.Kind {
				case gamemap.TileWall:
					glyph = theme.DimWall
				case gamemap.TileDoor:
					glyph = "🚪"
				case gamemap.TileStairsDown:
					glyph = "🔽"
				case gamemap.TileStairsUp:
					glyph = "🔼"
				case gamemap.TileGrass:
					glyph = "🟩"
				case gamemap.TileWater:
					glyph = "🟦"
				default:
					glyph = theme.DimFloor
				}
			}

			r.putGlyph(sx, sy, glyph, style)
		}
	}
}

// renderableEntity holds sorting info for entity rendering.
type renderableEntity struct {
	id    ecs.EntityID
	order int
	pos   component.Position
	rend  component.Renderable
}

// drawEntities renders all entities with Renderable + Position, ordered by RenderOrder.
func (r *Renderer) drawEntities(w *ecs.World, gmap *gamemap.GameMap) {
	ids := w.Query(component.CRenderable, component.CPosition)
	entities := make([]renderableEntity, 0, len(ids))

	for _, id := range ids {
		posComp := w.Get(id, component.CPosition)
		rendComp := w.Get(id, component.CRenderable)
		if posComp == nil || rendComp == nil {
			continue
		}
		pos := posComp.(component.Position)
		rend := rendComp.(component.Renderable)
		// Only draw entities on visible tiles.
		if gmap.InBounds(pos.X, pos.Y) && !gmap.At(pos.X, pos.Y).Visible {
			continue
		}
		entities = append(entities, renderableEntity{id: id, order: rend.RenderOrder, pos: pos, rend: rend})
	}

	// Sort ascending by render order (lower = drawn first / behind).
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].order < entities[j].order
	})

	for _, e := range entities {
		sx, sy, onScreen := r.camera.WorldToScreen(e.pos.X, e.pos.Y)
		if !onScreen {
			continue
		}
		style := tcell.StyleDefault.Foreground(e.rend.FGColor).Background(tcell.ColorBlack)
		r.putGlyph(sx, sy, e.rend.Glyph, style)
	}
}

// putGlyph draws a single glyph (ASCII or multi-rune emoji) at screen position (x, y).
func (r *Renderer) putGlyph(x, y int, glyph string, style tcell.Style) {
	runes := []rune(glyph)
	if len(runes) == 0 {
		return
	}
	mainc := runes[0]
	var combc []rune
	if len(runes) > 1 {
		combc = runes[1:]
	}
	r.screen.SetContent(x, y, mainc, combc, style)
	if runewidth.StringWidth(glyph) == 2 {
		// Fill the second column to avoid rendering artifacts.
		r.screen.SetContent(x+1, y, ' ', nil, style)
	}
}
