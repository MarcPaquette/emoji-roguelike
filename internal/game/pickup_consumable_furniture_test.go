package game

import (
	"math/rand"
	"testing"

	"emoji-roguelike/assets"
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/system"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

// placeFloorItem creates a consumable floor-item entity at the given position.
func placeFloorItem(g *Game, glyph, name string, x, y int) ecs.EntityID {
	id := g.world.CreateEntity()
	g.world.Add(id, component.Position{X: x, Y: y})
	g.world.Add(id, component.TagItem{})
	g.world.Add(id, component.CItemComp{Item: component.Item{
		Name:         name,
		Glyph:        glyph,
		Slot:         component.SlotConsumable,
		IsConsumable: true,
	}})
	return id
}

// placeFurniture creates a furniture entity directly in the ECS world.
func placeFurniture(g *Game, f component.Furniture) ecs.EntityID {
	id := g.world.CreateEntity()
	g.world.Add(id, f)
	return id
}

// playerPos returns the current player position.
func playerPos(g *Game) component.Position {
	return g.world.Get(g.playerID, component.CPosition).(component.Position)
}

// playerHP returns the current player Health component.
func playerHP(g *Game) component.Health {
	return g.world.Get(g.playerID, component.CHealth).(component.Health)
}

// playerCombat returns the player's Combat component.
func playerCombat(g *Game) component.Combat {
	return g.world.Get(g.playerID, component.CCombat).(component.Combat)
}

// hasMessage returns true if any message in g.messages contains the substring.
func hasMessage(g *Game, sub string) bool {
	for _, m := range g.messages {
		if len(m) >= len(sub) {
			for i := 0; i <= len(m)-len(sub); i++ {
				if m[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// ─── tryPickup ────────────────────────────────────────────────────────────────

func TestTryPickupSucceeds(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	pos := playerPos(g)

	itemID := placeFloorItem(g, assets.GlyphHyperflask, "Hyperflask", pos.X, pos.Y)

	g.tryPickup()

	// Item entity should be destroyed.
	if g.world.Alive(itemID) {
		t.Error("floor item entity should be destroyed after pickup")
	}
	// Item should be in the backpack.
	inv := g.world.Get(g.playerID, component.CInventory).(component.Inventory)
	if len(inv.Backpack) != 1 {
		t.Fatalf("backpack length = %d; want 1", len(inv.Backpack))
	}
	if inv.Backpack[0].Glyph != assets.GlyphHyperflask {
		t.Errorf("backpack[0] glyph = %q; want %q", inv.Backpack[0].Glyph, assets.GlyphHyperflask)
	}
	// Pickup message should appear.
	if !hasMessage(g, "You pick up") {
		t.Error("expected pickup message containing 'You pick up'")
	}
}

func TestTryPickupNothingHere(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	msgsBefore := len(g.messages)

	g.tryPickup()

	if len(g.messages) <= msgsBefore {
		t.Error("expected a 'Nothing to pick up here' message")
	}
	if !hasMessage(g, "Nothing to pick up here") {
		t.Errorf("expected 'Nothing to pick up here' message; got %v", g.messages[msgsBefore:])
	}
}

func TestTryPickupBackpackFull(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	pos := playerPos(g)

	// Fill the backpack to capacity.
	inv := g.world.Get(g.playerID, component.CInventory).(component.Inventory)
	for range inv.Capacity {
		inv.Backpack = append(inv.Backpack, component.Item{Name: "Filler", Glyph: "🧪"})
	}
	g.world.Add(g.playerID, inv)

	// Place an item at the player's position.
	itemID := placeFloorItem(g, assets.GlyphHyperflask, "Hyperflask", pos.X, pos.Y)

	g.tryPickup()

	// Item should NOT be picked up — entity still alive.
	if !g.world.Alive(itemID) {
		t.Error("floor item should remain when backpack is full")
	}
	if !hasMessage(g, "Backpack full") {
		t.Error("expected 'Backpack full' message")
	}
}

func TestTryPickupMissingItemComp(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	pos := playerPos(g)

	// Create a floor-item entity that has CTagItem and CPosition but no CItemComp.
	id := g.world.CreateEntity()
	g.world.Add(id, component.Position{X: pos.X, Y: pos.Y})
	g.world.Add(id, component.TagItem{})

	g.tryPickup()

	if !hasMessage(g, "Strange item") {
		t.Error("expected 'Strange item' message for entity missing CItemComp")
	}
}

// ─── applyConsumable ──────────────────────────────────────────────────────────

func TestApplyConsumableTracksItemsUsed(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	item := component.Item{Glyph: assets.GlyphHyperflask, Name: "Hyperflask", IsConsumable: true}

	g.applyConsumable(item)
	g.applyConsumable(item)

	if g.runLog.ItemsUsed[assets.GlyphHyperflask] != 2 {
		t.Errorf("ItemsUsed = %d; want 2", g.runLog.ItemsUsed[assets.GlyphHyperflask])
	}
}

func TestApplyConsumableHyperflaskRestoresHP(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	hp := playerHP(g)
	hp.Current = hp.Max - 20
	g.world.Add(g.playerID, hp)
	before := hp.Current

	g.applyConsumable(component.Item{Glyph: assets.GlyphHyperflask})

	after := playerHP(g).Current
	if after != before+15 {
		t.Errorf("HP after Hyperflask = %d; want %d", after, before+15)
	}
}

func TestApplyConsumableSporeDraughtRestoresHP(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	hp := playerHP(g)
	hp.Current = hp.Max - 30
	g.world.Add(g.playerID, hp)
	before := hp.Current

	g.applyConsumable(component.Item{Glyph: assets.GlyphSporeDraught})

	after := playerHP(g).Current
	if after != before+20 {
		t.Errorf("HP after Spore Draught = %d; want %d", after, before+20)
	}
}

func TestApplyConsumableNanoSyringeRestoresHP(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	hp := playerHP(g)
	hp.Current = hp.Max - 40
	g.world.Add(g.playerID, hp)
	before := hp.Current

	g.applyConsumable(component.Item{Glyph: assets.GlyphNanoSyringe})

	after := playerHP(g).Current
	if after != before+30 {
		t.Errorf("HP after Nano-Syringe = %d; want %d", after, before+30)
	}
}

func TestApplyConsumableHyperflaskCapsAtMax(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	// HP already full.
	hp := playerHP(g)
	before := hp.Max

	g.applyConsumable(component.Item{Glyph: assets.GlyphHyperflask})

	after := playerHP(g).Current
	if after != before {
		t.Errorf("HP should not exceed Max; got %d, max %d", after, before)
	}
}

func TestApplyConsumablePrismShardAppliesATKBoost(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")

	g.applyConsumable(component.Item{Glyph: assets.GlyphPrismShard})

	bonus := system.GetAttackBonus(g.world, g.playerID)
	if bonus != 3 {
		t.Errorf("ATK bonus after Prism Shard = %d; want 3", bonus)
	}
}

func TestApplyConsumableNullCloakAppliesInvisible(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")

	g.applyConsumable(component.Item{Glyph: assets.GlyphNullCloak})

	if !system.HasEffect(g.world, g.playerID, component.EffectInvisible) {
		t.Error("Null Cloak should apply Invisible effect")
	}
}

func TestApplyConsumableMemoryScrollRevealsMap(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")

	g.applyConsumable(component.Item{Glyph: assets.GlyphMemoryScroll})

	for y := range g.gmap.Height {
		for x := range g.gmap.Width {
			tile := g.gmap.At(x, y)
			if tile.Walkable && !tile.Explored {
				t.Errorf("tile (%d,%d) is walkable but unexplored after Memory Scroll", x, y)
				return
			}
		}
	}
}

func TestApplyConsumableResonanceCoilAppliesATKBoost(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")

	g.applyConsumable(component.Item{Glyph: assets.GlyphResonanceCoil})

	bonus := system.GetAttackBonus(g.world, g.playerID)
	if bonus != 5 {
		t.Errorf("ATK bonus after Resonance Coil = %d; want 5", bonus)
	}
}

func TestApplyConsumablePrismaticWardAppliesDEFBoost(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")

	g.applyConsumable(component.Item{Glyph: assets.GlyphPrismaticWard})

	bonus := system.GetDefenseBonus(g.world, g.playerID)
	if bonus != 4 {
		t.Errorf("DEF bonus after Prismatic Ward = %d; want 4", bonus)
	}
}

func TestApplyConsumableVoidEssenceAppliesInvisible(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")

	g.applyConsumable(component.Item{Glyph: assets.GlyphVoidEssence})

	if !system.HasEffect(g.world, g.playerID, component.EffectInvisible) {
		t.Error("Void Essence should apply Invisible effect")
	}
}

func TestApplyConsumableResonanceBurstAppliesATKAndBurn(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")

	g.applyConsumable(component.Item{Glyph: assets.GlyphResonanceBurst})

	atkBonus := system.GetAttackBonus(g.world, g.playerID)
	if atkBonus != 8 {
		t.Errorf("ATK bonus after Resonance Burst = %d; want 8", atkBonus)
	}
	burnDmg := system.GetSelfBurnDamage(g.world, g.playerID)
	if burnDmg != 2 {
		t.Errorf("SelfBurn damage after Resonance Burst = %d; want 2", burnDmg)
	}
}

func TestApplyConsumablePhaseRodAppliesDEFBoost(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")

	g.applyConsumable(component.Item{Glyph: assets.GlyphPhaseRod})

	bonus := system.GetDefenseBonus(g.world, g.playerID)
	if bonus != 6 {
		t.Errorf("DEF bonus after Phase Rod = %d; want 6", bonus)
	}
}

func TestApplyConsumableApexCoreIncreasesPermanentMaxHP(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	baseBefore := g.baseMaxHP
	hp := playerHP(g)
	maxBefore := hp.Max
	curBefore := hp.Current

	g.applyConsumable(component.Item{Glyph: assets.GlyphApexCore})

	if g.baseMaxHP != baseBefore+3 {
		t.Errorf("baseMaxHP = %d; want %d", g.baseMaxHP, baseBefore+3)
	}
	after := playerHP(g)
	if after.Max != maxBefore+3 {
		t.Errorf("hp.Max = %d; want %d", after.Max, maxBefore+3)
	}
	if after.Current != curBefore+3 {
		t.Errorf("hp.Current = %d; want %d", after.Current, curBefore+3)
	}
}

func TestApplyConsumableTesseractChangesPosition(t *testing.T) {
	// Try several seeds until we find one that moves the player.
	moved := false
	for seed := int64(0); seed < 10; seed++ {
		g := newAbilityTestGame(t, "arcanist")
		g.rng = rand.New(rand.NewSource(seed))
		before := playerPos(g)

		g.applyConsumable(component.Item{Glyph: assets.GlyphTesseract})

		after := playerPos(g)
		if after.X != before.X || after.Y != before.Y {
			moved = true
			break
		}
	}
	if !moved {
		t.Error("Tesseract Cube never moved player across 10 seeds")
	}
}

// ─── interactFurniture ────────────────────────────────────────────────────────

func TestInteractFurnitureBonusATK(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	atkBefore := playerCombat(g).Attack
	furATKBefore := g.furnitureATK

	fid := placeFurniture(g, component.Furniture{
		Glyph: "🔬", Name: "Lab Bench", Description: "Your reflexes sharpen.", BonusATK: 2,
	})

	g.interactFurniture(fid)

	if g.furnitureATK != furATKBefore+2 {
		t.Errorf("furnitureATK = %d; want %d", g.furnitureATK, furATKBefore+2)
	}
	if playerCombat(g).Attack != atkBefore+2 {
		t.Errorf("player Attack = %d; want %d", playerCombat(g).Attack, atkBefore+2)
	}
	// Verify Used flag is set.
	f := g.world.Get(fid, component.CFurniture).(component.Furniture)
	if !f.Used {
		t.Error("furniture should be marked Used after first interaction")
	}
}

func TestInteractFurnitureBonusDEF(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	defBefore := playerCombat(g).Defense
	furDEFBefore := g.furnitureDEF

	fid := placeFurniture(g, component.Furniture{
		Glyph: "🛡️", Name: "Shield Rack", Description: "You feel sturdier.", BonusDEF: 3,
	})

	g.interactFurniture(fid)

	if g.furnitureDEF != furDEFBefore+3 {
		t.Errorf("furnitureDEF = %d; want %d", g.furnitureDEF, furDEFBefore+3)
	}
	if playerCombat(g).Defense != defBefore+3 {
		t.Errorf("player Defense = %d; want %d", playerCombat(g).Defense, defBefore+3)
	}
}

func TestInteractFurnitureBonusMaxHP(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	baseBefore := g.baseMaxHP
	hp := playerHP(g)
	maxBefore := hp.Max
	curBefore := hp.Current

	fid := placeFurniture(g, component.Furniture{
		Glyph: "💊", Name: "Med Station", Description: "Your body adapts.", BonusMaxHP: 4,
	})

	g.interactFurniture(fid)

	if g.baseMaxHP != baseBefore+4 {
		t.Errorf("baseMaxHP = %d; want %d", g.baseMaxHP, baseBefore+4)
	}
	after := playerHP(g)
	if after.Max != maxBefore+4 {
		t.Errorf("hp.Max = %d; want %d", after.Max, maxBefore+4)
	}
	if after.Current != curBefore+4 {
		t.Errorf("hp.Current = %d; want %d", after.Current, curBefore+4)
	}
}

func TestInteractFurnitureHealHP(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	hp := playerHP(g)
	hp.Current = hp.Max - 10
	g.world.Add(g.playerID, hp)
	before := hp.Current

	fid := placeFurniture(g, component.Furniture{
		Glyph: "🌿", Name: "Healing Moss", Description: "The moss soothes you.", HealHP: 5,
	})

	g.interactFurniture(fid)

	after := playerHP(g).Current
	if after != before+5 {
		t.Errorf("HP after HealHP furniture = %d; want %d", after, before+5)
	}
}

func TestInteractFurniturePassiveKeenEye(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	radiusBefore := g.fovRadius

	fid := placeFurniture(g, component.Furniture{
		Glyph: "🔭", Name: "Telescope", Description: "Your sight extends.", PassiveKind: component.PassiveKeenEye,
	})

	g.interactFurniture(fid)

	if g.fovRadius != radiusBefore+1 {
		t.Errorf("fovRadius = %d; want %d", g.fovRadius, radiusBefore+1)
	}
}

func TestInteractFurniturePassiveKillRestore(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")

	fid := placeFurniture(g, component.Furniture{
		Glyph: "💀", Name: "Skull Altar", Description: "Your life quickens on kills.", PassiveKind: component.PassiveKillRestore,
	})

	g.interactFurniture(fid)

	if !g.furnitureKillRestore {
		t.Error("furnitureKillRestore should be true after PassiveKillRestore furniture")
	}
}

func TestInteractFurniturePassiveThorns(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	thornsBefore := g.furnitureThorns

	fid := placeFurniture(g, component.Furniture{
		Glyph: "🌵", Name: "Crystal Spine", Description: "Spines form beneath your skin.", PassiveKind: component.PassiveThorns,
	})

	g.interactFurniture(fid)

	if g.furnitureThorns != thornsBefore+1 {
		t.Errorf("furnitureThorns = %d; want %d", g.furnitureThorns, thornsBefore+1)
	}
}

func TestInteractFurnitureAlreadyUsed(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	atkBefore := playerCombat(g).Attack
	furATKBefore := g.furnitureATK

	fid := placeFurniture(g, component.Furniture{
		Glyph: "🔬", Name: "Lab Bench", Description: "Nothing new here.", BonusATK: 2, Used: true,
	})

	g.interactFurniture(fid)

	// No bonus should be applied because it's already Used.
	if g.furnitureATK != furATKBefore {
		t.Errorf("furnitureATK changed on used furniture: %d -> %d", furATKBefore, g.furnitureATK)
	}
	if playerCombat(g).Attack != atkBefore {
		t.Errorf("player Attack changed on used furniture: %d -> %d", atkBefore, playerCombat(g).Attack)
	}
}

func TestInteractFurnitureBonusAppliedOnlyOnce(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	atkBefore := playerCombat(g).Attack

	fid := placeFurniture(g, component.Furniture{
		Glyph: "🔬", Name: "Lab Bench", Description: "Sharpens reflexes.", BonusATK: 2,
	})

	g.interactFurniture(fid)
	g.interactFurniture(fid) // second interaction should be a no-op

	if playerCombat(g).Attack != atkBefore+2 {
		t.Errorf("player Attack after 2 interactions = %d; want %d (bonus applied once)", playerCombat(g).Attack, atkBefore+2)
	}
}

func TestInteractFurnitureNoBonus(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	atkBefore := playerCombat(g).Attack
	msgsBefore := len(g.messages)

	fid := placeFurniture(g, component.Furniture{
		Glyph: "🪟", Name: "Window", Description: "A cracked viewport.",
	})

	g.interactFurniture(fid)

	// No combat changes.
	if playerCombat(g).Attack != atkBefore {
		t.Error("furniture with no bonus should not change player stats")
	}
	// But the description message should still appear.
	if len(g.messages) <= msgsBefore {
		t.Error("expected description message even for no-bonus furniture")
	}
}

func TestInteractFurnitureMissingComponent(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	msgsBefore := len(g.messages)

	// Create an entity that has no CFurniture component.
	id := g.world.CreateEntity()

	g.interactFurniture(id)

	// Should silently return — no message, no panic.
	if len(g.messages) != msgsBefore {
		t.Error("interactFurniture with no component should produce no messages")
	}
}
