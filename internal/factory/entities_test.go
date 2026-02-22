package factory

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/generate"
	"math/rand"
	"testing"
)

// testClass is a minimal ClassDef used across factory tests.
var testClass = assets.ClassDef{
	ID:      "test",
	Emoji:   "🧙",
	MaxHP:   20,
	Attack:  4,
	Defense: 2,
}

func TestNewPlayerComponents(t *testing.T) {
	w := ecs.NewWorld()
	id := NewPlayer(w, 5, 3, testClass)

	if !w.Alive(id) {
		t.Fatal("player entity must be alive")
	}

	pos := w.Get(id, component.CPosition)
	if pos == nil {
		t.Fatal("player must have CPosition")
	}
	if p := pos.(component.Position); p.X != 5 || p.Y != 3 {
		t.Errorf("position = (%d,%d); want (5,3)", p.X, p.Y)
	}

	hp := w.Get(id, component.CHealth)
	if hp == nil {
		t.Fatal("player must have CHealth")
	}
	if h := hp.(component.Health); h.Current != testClass.MaxHP || h.Max != testClass.MaxHP {
		t.Errorf("HP = %d/%d; want %d/%d", h.Current, h.Max, testClass.MaxHP, testClass.MaxHP)
	}

	cbt := w.Get(id, component.CCombat)
	if cbt == nil {
		t.Fatal("player must have CCombat")
	}
	if c := cbt.(component.Combat); c.Attack != testClass.Attack || c.Defense != testClass.Defense {
		t.Errorf("Combat = {atk:%d def:%d}; want {atk:%d def:%d}", c.Attack, c.Defense, testClass.Attack, testClass.Defense)
	}

	if w.Get(id, component.CInventory) == nil {
		t.Error("player must have CInventory")
	}
	if w.Get(id, component.CEffects) == nil {
		t.Error("player must have CEffects")
	}
	if !w.Has(id, component.CTagPlayer) {
		t.Error("player must have CTagPlayer")
	}
	if !w.Has(id, component.CTagBlocking) {
		t.Error("player must have CTagBlocking")
	}
}

func TestNewPlayerInventoryCapacity(t *testing.T) {
	w := ecs.NewWorld()
	id := NewPlayer(w, 0, 0, testClass)
	inv := w.Get(id, component.CInventory).(component.Inventory)
	if inv.Capacity != 8 {
		t.Errorf("player inventory capacity = %d; want 8", inv.Capacity)
	}
}

func TestNewEnemyComponents(t *testing.T) {
	entry := generate.EnemySpawnEntry{
		Glyph:      "🦀",
		Name:       "Crystal Crawl",
		ThreatCost: 2,
		Attack:     3,
		Defense:    2,
		MaxHP:      8,
		SightRange: 5,
	}
	w := ecs.NewWorld()
	id := NewEnemy(w, entry, 7, 9)

	if !w.Alive(id) {
		t.Fatal("enemy entity must be alive")
	}
	if pos := w.Get(id, component.CPosition); pos == nil {
		t.Fatal("enemy must have CPosition")
	} else if p := pos.(component.Position); p.X != 7 || p.Y != 9 {
		t.Errorf("position = (%d,%d); want (7,9)", p.X, p.Y)
	}

	hp := w.Get(id, component.CHealth)
	if hp == nil {
		t.Fatal("enemy must have CHealth")
	}
	if h := hp.(component.Health); h.Max != entry.MaxHP {
		t.Errorf("max HP = %d; want %d", h.Max, entry.MaxHP)
	}

	ai := w.Get(id, component.CAI)
	if ai == nil {
		t.Fatal("enemy must have CAI")
	}
	if a := ai.(component.AI); a.SightRange != entry.SightRange {
		t.Errorf("SightRange = %d; want %d", a.SightRange, entry.SightRange)
	}

	if !w.Has(id, component.CTagBlocking) {
		t.Error("enemy must have CTagBlocking")
	}
	if w.Get(id, component.CEffects) == nil {
		t.Error("enemy must have CEffects")
	}
}

func TestNewItemComponents(t *testing.T) {
	entry := generate.ItemSpawnEntry{Glyph: "🧪", Name: "Hyperflask"}
	w := ecs.NewWorld()
	id := NewItem(w, entry, 4, 6)

	if w.Get(id, component.CPosition) == nil {
		t.Fatal("item must have CPosition")
	}
	if !w.Has(id, component.CTagItem) {
		t.Error("item must have CTagItem")
	}
	ci := w.Get(id, component.CItem)
	if ci == nil {
		t.Fatal("item must have CItem")
	}
	item := ci.(component.CItemComp).Item
	if !item.IsConsumable {
		t.Error("item created by NewItem must be consumable")
	}
	if item.Glyph != entry.Glyph {
		t.Errorf("Glyph = %q; want %q", item.Glyph, entry.Glyph)
	}
	if item.Name != entry.Name {
		t.Errorf("Name = %q; want %q", item.Name, entry.Name)
	}
}

func TestNewEquipItemScaling(t *testing.T) {
	// Equipment ATK bonus grows with floor number.
	entry := generate.EquipSpawnEntry{
		Glyph:    "⚔️",
		Name:     "Shard Blade",
		Slot:     4,
		BaseATK:  2,
		ATKScale: 5,
	}
	w1 := ecs.NewWorld()
	id1 := NewEquipItem(w1, entry, 1, rand.New(rand.NewSource(0)), 0, 0)
	bonus1 := w1.Get(id1, component.CItem).(component.CItemComp).Item.BonusATK

	w2 := ecs.NewWorld()
	id2 := NewEquipItem(w2, entry, 10, rand.New(rand.NewSource(0)), 0, 0)
	bonus2 := w2.Get(id2, component.CItem).(component.CItemComp).Item.BonusATK

	if bonus2 <= bonus1 {
		t.Errorf("floor-10 ATK bonus (%d) should exceed floor-1 bonus (%d)", bonus2, bonus1)
	}
}

func TestNewEquipItemNotConsumable(t *testing.T) {
	entry := generate.EquipSpawnEntry{Glyph: "⚔️", Name: "Shard Blade", Slot: 4}
	w := ecs.NewWorld()
	id := NewEquipItem(w, entry, 1, rand.New(rand.NewSource(0)), 0, 0)
	item := w.Get(id, component.CItem).(component.CItemComp).Item
	if item.IsConsumable {
		t.Error("equipment items must not be consumable")
	}
}

func TestDropItemCreatesFloorEntity(t *testing.T) {
	item := component.Item{
		Name:         "Hyperflask",
		Glyph:        "🧪",
		Slot:         component.SlotConsumable,
		IsConsumable: true,
	}
	w := ecs.NewWorld()
	id := DropItem(w, item, 3, 3)

	if !w.Has(id, component.CTagItem) {
		t.Error("dropped item must have CTagItem")
	}
	ci := w.Get(id, component.CItem)
	if ci == nil {
		t.Fatal("dropped item must have CItem")
	}
	if ci.(component.CItemComp).Item.Name != item.Name {
		t.Errorf("item Name = %q; want %q", ci.(component.CItemComp).Item.Name, item.Name)
	}
}

func TestNewInscriptionComponents(t *testing.T) {
	w := ecs.NewWorld()
	id := NewInscription(w, "Hello, dungeon", 5, 7)

	if !w.Alive(id) {
		t.Fatal("inscription entity must be alive")
	}
	pos := w.Get(id, component.CPosition)
	if pos == nil {
		t.Fatal("inscription must have CPosition")
	}
	if p := pos.(component.Position); p.X != 5 || p.Y != 7 {
		t.Errorf("position = (%d,%d); want (5,7)", p.X, p.Y)
	}
	ins := w.Get(id, component.CInscription)
	if ins == nil {
		t.Fatal("inscription must have CInscription")
	}
	if ins.(component.Inscription).Text != "Hello, dungeon" {
		t.Errorf("Text = %q; want %q", ins.(component.Inscription).Text, "Hello, dungeon")
	}
}

func TestNewStairsDownComponents(t *testing.T) {
	w := ecs.NewWorld()
	id := NewStairsDown(w, 8, 12)

	if !w.Alive(id) {
		t.Fatal("stairs entity must be alive")
	}
	pos := w.Get(id, component.CPosition)
	if pos == nil {
		t.Fatal("stairs must have CPosition")
	}
	if p := pos.(component.Position); p.X != 8 || p.Y != 12 {
		t.Errorf("position = (%d,%d); want (8,12)", p.X, p.Y)
	}
	if !w.Has(id, component.CTagStairs) {
		t.Error("stairs entity must have CTagStairs")
	}
}

func TestNewFurnitureComponents(t *testing.T) {
	entry := generate.FurnitureSpawnEntry{
		Glyph:       "🔬",
		Name:        "Microscope",
		Description: "A dusty microscope.",
		BonusATK:    0,
		BonusDEF:    0,
		BonusMaxHP:  8,
	}
	w := ecs.NewWorld()
	id := NewFurniture(w, entry, 3, 7)

	if !w.Alive(id) {
		t.Fatal("furniture entity must be alive")
	}
	pos := w.Get(id, component.CPosition)
	if pos == nil {
		t.Fatal("furniture must have CPosition")
	}
	if p := pos.(component.Position); p.X != 3 || p.Y != 7 {
		t.Errorf("position = (%d,%d); want (3,7)", p.X, p.Y)
	}
	fc := w.Get(id, component.CFurniture)
	if fc == nil {
		t.Fatal("furniture must have CFurniture")
	}
	f := fc.(component.Furniture)
	if f.Glyph != entry.Glyph {
		t.Errorf("Glyph = %q; want %q", f.Glyph, entry.Glyph)
	}
	if f.Name != entry.Name {
		t.Errorf("Name = %q; want %q", f.Name, entry.Name)
	}
	if f.Description != entry.Description {
		t.Errorf("Description = %q; want %q", f.Description, entry.Description)
	}
	if f.BonusMaxHP != entry.BonusMaxHP {
		t.Errorf("BonusMaxHP = %d; want %d", f.BonusMaxHP, entry.BonusMaxHP)
	}
	if f.Used {
		t.Error("new furniture must not be Used")
	}
}

func TestNewFurnitureHasRenderable(t *testing.T) {
	entry := generate.FurnitureSpawnEntry{Glyph: "🌡️", Name: "Thermometer", Description: "Frozen."}
	w := ecs.NewWorld()
	id := NewFurniture(w, entry, 0, 0)
	rend := w.Get(id, component.CRenderable)
	if rend == nil {
		t.Fatal("furniture must have CRenderable")
	}
	if r := rend.(component.Renderable); r.Glyph != entry.Glyph {
		t.Errorf("Renderable.Glyph = %q; want %q", r.Glyph, entry.Glyph)
	}
}

func TestNewFurnitureNotBlocking(t *testing.T) {
	// Furniture must NOT have CTagBlocking (interaction handled via MoveInteract).
	entry := generate.FurnitureSpawnEntry{Glyph: "🔬", Name: "Microscope", Description: "A dusty microscope."}
	w := ecs.NewWorld()
	id := NewFurniture(w, entry, 0, 0)
	if w.Has(id, component.CTagBlocking) {
		t.Error("furniture must not have CTagBlocking")
	}
}
