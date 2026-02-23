package game

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/component"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// newTestCoopGame creates a CoopGame with simulation screens and pre-assigns
// classes so loadFloor can be called immediately without class selection.
func newTestCoopGame() *CoopGame {
	screens := [2]tcell.Screen{
		newSimScreen(),
		newSimScreen(),
	}
	g := NewCoopGame(screens)
	// Assign the first two classes directly (skips interactive class-select).
	for i, p := range g.players {
		p.class = assets.Classes[i%len(assets.Classes)]
		p.fovRadius = p.class.FOVRadius
		p.baseMaxHP = p.class.MaxHP
		p.runLog.Class = p.class.Name
	}
	return g
}

// newSimScreen creates an initialized 80×24 simulation screen.
func newSimScreen() tcell.Screen {
	ss := tcell.NewSimulationScreen("UTF-8")
	ss.SetSize(80, 24)
	_ = ss.Init()
	return ss
}

// TestCoopGameLoadFloor verifies that loadFloor creates a world with exactly
// two live player entities (one per player).
func TestCoopGameLoadFloor(t *testing.T) {
	g := newTestCoopGame()
	g.loadFloor(1)

	if g.world == nil {
		t.Fatal("world is nil after loadFloor")
	}
	if g.gmap == nil {
		t.Fatal("gmap is nil after loadFloor")
	}

	// Both players must have a valid entity ID.
	for i, p := range g.players {
		if p.id == 0 {
			t.Errorf("player %d has zero entity ID", i)
		}
		if !g.world.Alive(p.id) {
			t.Errorf("player %d entity is not alive", i)
		}
	}

	// Both must have health components.
	for i, p := range g.players {
		hpComp := g.world.Get(p.id, component.CHealth)
		if hpComp == nil {
			t.Fatalf("player %d has no health component", i)
		}
		hp := hpComp.(component.Health)
		if hp.Current <= 0 {
			t.Errorf("player %d starts with non-positive HP: %d", i, hp.Current)
		}
		if hp.Max <= 0 {
			t.Errorf("player %d starts with non-positive MaxHP: %d", i, hp.Max)
		}
	}

	// Players must have different entity IDs.
	if g.players[0].id == g.players[1].id {
		t.Error("both players share the same entity ID")
	}

	// Both must have the CTagPlayer marker.
	for i, p := range g.players {
		if !g.world.Has(p.id, component.CTagPlayer) {
			t.Errorf("player %d missing CTagPlayer", i)
		}
	}
}

// TestCoopGameLoadFloorPreservesHP verifies that player HP is preserved
// across floor transitions (load floor 1 then floor 2).
func TestCoopGameLoadFloorPreservesHP(t *testing.T) {
	g := newTestCoopGame()
	g.loadFloor(1)

	// Damage P1 so we can verify preservation.
	p1 := g.players[0]
	hp := g.world.Get(p1.id, component.CHealth).(component.Health)
	savedHP := hp.Current - 3
	if savedHP < 1 {
		savedHP = 1
	}
	hp.Current = savedHP
	g.world.Add(p1.id, hp)

	g.loadFloor(2)

	// P1's HP should be restored to savedHP (not full max).
	hpAfter := g.world.Get(p1.id, component.CHealth).(component.Health)
	if hpAfter.Current != savedHP {
		t.Errorf("P1 HP after floor transition = %d; want %d", hpAfter.Current, savedHP)
	}
}

// TestCoopGameTurnOrderMovement verifies that processCoopAction for both
// players advances their positions independently in the shared world.
func TestCoopGameTurnOrderMovement(t *testing.T) {
	g := newTestCoopGame()
	g.loadFloor(1)

	// Record initial positions.
	pos0Before := g.coopPlayerPosition(g.players[0])
	pos1Before := g.coopPlayerPosition(g.players[1])

	// Give P1 a wait action (turn used, position unchanged).
	turnUsed := g.processCoopAction(g.players[0], ActionWait)
	if !turnUsed {
		t.Error("ActionWait should use a turn")
	}

	// P1 position must not change on wait.
	pos0After := g.coopPlayerPosition(g.players[0])
	if pos0After != pos0Before {
		t.Errorf("P1 position changed on wait: %v → %v", pos0Before, pos0After)
	}

	// P2 also waits.
	turnUsed = g.processCoopAction(g.players[1], ActionWait)
	if !turnUsed {
		t.Error("ActionWait for P2 should use a turn")
	}

	// P2 position must not change.
	pos1After := g.coopPlayerPosition(g.players[1])
	if pos1After != pos1Before {
		t.Errorf("P2 position changed on wait: %v → %v", pos1Before, pos1After)
	}
}

// TestCoopGamePickup verifies that each player can independently pick up an item.
func TestCoopGamePickup(t *testing.T) {
	g := newTestCoopGame()
	g.loadFloor(1)

	p := g.players[0]
	pos := g.coopPlayerPosition(p)

	// Spawn an item at P1's feet.
	factory := assets.GlyphHyperflask
	_ = factory // just use the constant

	// Use the game's factory function to place an item at player's position.
	// We call coopTryPickup to test pickup logic without needing a real item on the map.
	// First verify the player starts with an empty backpack.
	invComp := g.world.Get(p.id, component.CInventory)
	if invComp == nil {
		t.Fatal("player has no inventory")
	}
	inv := invComp.(component.Inventory)
	initialCount := len(inv.Backpack)

	// Place an item at the player's position directly into the world.
	itemID := g.world.CreateEntity()
	g.world.Add(itemID, component.Position{X: pos.X, Y: pos.Y})
	g.world.Add(itemID, component.TagItem{})
	g.world.Add(itemID, component.CItemComp{Item: component.Item{
		Name:         "Test Flask",
		Glyph:        assets.GlyphHyperflask,
		Slot:         component.SlotConsumable,
		IsConsumable: true,
	}})

	g.coopTryPickup(p)

	inv2 := g.world.Get(p.id, component.CInventory).(component.Inventory)
	if len(inv2.Backpack) != initialCount+1 {
		t.Errorf("expected backpack size %d after pickup, got %d", initialCount+1, len(inv2.Backpack))
	}
	// Item entity should be destroyed.
	if g.world.Alive(itemID) {
		t.Error("item entity should be destroyed after pickup")
	}
}

// TestCoopGameMessages verifies that messages are shared across both players.
func TestCoopGameMessages(t *testing.T) {
	g := newTestCoopGame()
	g.addMessage("hello coop")

	if len(g.messages) == 0 {
		t.Fatal("no messages in shared log")
	}
	if g.messages[len(g.messages)-1] != "hello coop" {
		t.Errorf("unexpected last message: %q", g.messages[len(g.messages)-1])
	}
}

// TestCoopGameBothPlayersAlive verifies the initial alive state.
func TestCoopGameBothPlayersAlive(t *testing.T) {
	g := newTestCoopGame()
	g.loadFloor(1)

	for i, p := range g.players {
		if !p.alive {
			t.Errorf("player %d should be alive at game start", i)
		}
	}
}

// TestCoopGamePlayersDifferentColors verifies that the two player renderables
// have distinct foreground colors.
func TestCoopGamePlayersDifferentColors(t *testing.T) {
	g := newTestCoopGame()
	g.loadFloor(1)

	rend0 := g.world.Get(g.players[0].id, component.CRenderable).(component.Renderable)
	rend1 := g.world.Get(g.players[1].id, component.CRenderable).(component.Renderable)

	if rend0.FGColor == rend1.FGColor {
		t.Errorf("both players have the same FGColor (%v); they should be distinct", rend0.FGColor)
	}
}
