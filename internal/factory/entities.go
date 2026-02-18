package factory

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/generate"

	"github.com/gdamore/tcell/v2"
)

// NewPlayer creates the player entity at (x, y) using the given class definition.
func NewPlayer(w *ecs.World, x, y int, class assets.ClassDef) ecs.EntityID {
	id := w.CreateEntity()
	w.Add(id, component.Position{X: x, Y: y})
	w.Add(id, component.Health{Current: class.MaxHP, Max: class.MaxHP})
	w.Add(id, component.Renderable{
		Glyph:       class.Emoji,
		FGColor:     tcell.ColorYellow,
		BGColor:     tcell.ColorDefault,
		RenderOrder: 10,
	})
	w.Add(id, component.Combat{Attack: class.Attack, Defense: class.Defense})
	w.Add(id, component.Inventory{Capacity: 10})
	w.Add(id, component.Effects{})
	w.Add(id, component.TagPlayer{})
	w.Add(id, component.TagBlocking{})
	return id
}

// NewEnemy creates an enemy entity from a spawn entry.
func NewEnemy(w *ecs.World, entry generate.EnemySpawnEntry, x, y int) ecs.EntityID {
	id := w.CreateEntity()
	w.Add(id, component.Position{X: x, Y: y})
	w.Add(id, component.Health{Current: entry.MaxHP, Max: entry.MaxHP})
	w.Add(id, component.Renderable{
		Glyph:       entry.Glyph,
		FGColor:     tcell.ColorRed,
		BGColor:     tcell.ColorDefault,
		RenderOrder: 5,
	})
	w.Add(id, component.Combat{Attack: entry.Attack, Defense: entry.Defense})
	w.Add(id, component.AI{Behavior: component.BehaviorChase, SightRange: entry.SightRange})
	w.Add(id, component.Effects{})
	w.Add(id, component.TagBlocking{})
	return id
}

// NewItem creates an item entity from a spawn entry.
func NewItem(w *ecs.World, entry generate.ItemSpawnEntry, x, y int) ecs.EntityID {
	id := w.CreateEntity()
	w.Add(id, component.Position{X: x, Y: y})
	w.Add(id, component.Renderable{
		Glyph:       entry.Glyph,
		FGColor:     tcell.ColorGreen,
		BGColor:     tcell.ColorDefault,
		RenderOrder: 2,
	})
	w.Add(id, component.TagItem{})
	return id
}

// NewItemByGlyph creates an item entity using a raw glyph string (for class start items).
func NewItemByGlyph(w *ecs.World, glyph string, x, y int) ecs.EntityID {
	return NewItem(w, generate.ItemSpawnEntry{Glyph: glyph, Name: glyph}, x, y)
}

// NewStairsDown creates a stairs-down entity.
func NewStairsDown(w *ecs.World, x, y int) ecs.EntityID {
	id := w.CreateEntity()
	w.Add(id, component.Position{X: x, Y: y})
	w.Add(id, component.Renderable{
		Glyph:       "🔽",
		FGColor:     tcell.ColorWhite,
		BGColor:     tcell.ColorDefault,
		RenderOrder: 1,
	})
	w.Add(id, component.TagStairs{})
	return id
}
