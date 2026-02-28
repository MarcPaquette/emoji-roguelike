package factory

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/generate"
	"math/rand"

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
	w.Add(id, component.Inventory{Capacity: 8})
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
	w.Add(id, component.Combat{
		Attack:        entry.Attack,
		Defense:       entry.Defense,
		SpecialKind:   entry.SpecialKind,
		SpecialChance: entry.SpecialChance,
		SpecialMag:    entry.SpecialMag,
		SpecialDur:    entry.SpecialDur,
	})
	w.Add(id, component.AI{Behavior: component.BehaviorChase, SightRange: entry.SightRange})
	w.Add(id, component.Effects{})
	w.Add(id, component.TagBlocking{})
	if len(entry.Drops) > 0 {
		drops := make([]component.LootEntry, len(entry.Drops))
		for i, d := range entry.Drops {
			drops[i] = component.LootEntry{Glyph: d.Glyph, Chance: d.Chance}
		}
		w.Add(id, component.Loot{Drops: drops})
	}
	return id
}

// NewItem creates a consumable item entity from a spawn entry.
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
	w.Add(id, component.CItemComp{Item: component.Item{
		Name:         entry.Name,
		Glyph:        entry.Glyph,
		Slot:         component.SlotConsumable,
		IsConsumable: true,
	}})
	return id
}

// NewItemByGlyph creates a consumable item entity using a raw glyph string (for class start items).
func NewItemByGlyph(w *ecs.World, glyph string, x, y int) ecs.EntityID {
	name := assets.ConsumableName(glyph)
	return NewItem(w, generate.ItemSpawnEntry{Glyph: glyph, Name: name}, x, y)
}

// NewEquipItem creates an equipment item entity with floor-scaled stats.
func NewEquipItem(w *ecs.World, entry generate.EquipSpawnEntry, floor int, rng *rand.Rand, x, y int) ecs.EntityID {
	t := 0.0
	if floor > 1 {
		t = float64(floor-1) / 9.0
	}
	variant := rng.Intn(3) // 0, 1, or 2

	bonusATK := entry.BaseATK + int(float64(entry.ATKScale)*t)
	if entry.ATKScale > 0 {
		bonusATK += variant
	}
	bonusDEF := entry.BaseDEF + int(float64(entry.DEFScale)*t)
	if entry.DEFScale > 0 {
		bonusDEF += variant
	}
	bonusHP := entry.BaseMaxHP + int(float64(entry.HPScale)*t)
	if entry.HPScale > 0 {
		bonusHP += variant * 2
	}

	id := w.CreateEntity()
	w.Add(id, component.Position{X: x, Y: y})
	w.Add(id, component.Renderable{
		Glyph:       entry.Glyph,
		FGColor:     tcell.ColorAqua,
		BGColor:     tcell.ColorDefault,
		RenderOrder: 2,
	})
	w.Add(id, component.TagItem{})
	w.Add(id, component.CItemComp{Item: component.Item{
		Name:         entry.Name,
		Glyph:        entry.Glyph,
		Slot:         component.ItemSlot(entry.Slot),
		BonusATK:     bonusATK,
		BonusDEF:     bonusDEF,
		BonusMaxHP:   bonusHP,
		IsConsumable: false,
	}})
	return id
}

// DropItem creates a floor-item entity from an Item value (used when dropping from inventory).
func DropItem(w *ecs.World, item component.Item, x, y int) ecs.EntityID {
	id := w.CreateEntity()
	w.Add(id, component.Position{X: x, Y: y})
	color := tcell.ColorGreen
	if !item.IsConsumable {
		color = tcell.ColorAqua
	}
	w.Add(id, component.Renderable{
		Glyph:       item.Glyph,
		FGColor:     color,
		BGColor:     tcell.ColorDefault,
		RenderOrder: 2,
	})
	w.Add(id, component.TagItem{})
	w.Add(id, component.CItemComp{Item: item})
	return id
}

// NewInscription creates a wall-writing entity at (x, y).
// The player reads it by stepping onto the tile.
func NewInscription(w *ecs.World, text string, x, y int) ecs.EntityID {
	id := w.CreateEntity()
	w.Add(id, component.Position{X: x, Y: y})
	w.Add(id, component.Renderable{
		Glyph:       "📝",
		FGColor:     tcell.ColorLightBlue,
		BGColor:     tcell.ColorDefault,
		RenderOrder: 1,
	})
	w.Add(id, component.Inscription{Text: text})
	return id
}

// NewFurniture creates a decorative furniture entity that may grant a one-time bonus.
func NewFurniture(w *ecs.World, entry generate.FurnitureSpawnEntry, x, y int) ecs.EntityID {
	id := w.CreateEntity()
	w.Add(id, component.Position{X: x, Y: y})
	w.Add(id, component.Renderable{
		Glyph:       entry.Glyph,
		FGColor:     tcell.ColorYellow,
		BGColor:     tcell.ColorDefault,
		RenderOrder: 1,
	})
	w.Add(id, component.Furniture{
		Glyph:       entry.Glyph,
		Name:        entry.Name,
		Description: entry.Description,
		BonusATK:    entry.BonusATK,
		BonusDEF:    entry.BonusDEF,
		BonusMaxHP:  entry.BonusMaxHP,
		HealHP:      entry.HealHP,
		PassiveKind: entry.PassiveKind,
	})
	return id
}

// NewNPC creates a non-hostile, interactable NPC entity at (x, y).
func NewNPC(w *ecs.World, name, glyph string, kind component.NPCKind, lines []string, x, y int) ecs.EntityID {
	id := w.CreateEntity()
	w.Add(id, component.Position{X: x, Y: y})
	w.Add(id, component.Renderable{
		Glyph:       glyph,
		FGColor:     tcell.ColorAqua,
		BGColor:     tcell.ColorDefault,
		RenderOrder: 5,
	})
	w.Add(id, component.TagBlocking{})
	w.Add(id, component.NPC{Name: name, Kind: kind, Lines: lines})
	return id
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
