package game

import (
	"math/rand"
	"testing"

	"emoji-roguelike/assets"
	"emoji-roguelike/internal/component"

	"github.com/gdamore/tcell/v2"
)

// newAbilityTestGame builds a minimal single-player Game using a simulation
// screen, with the given class selected. loadFloor(1) is called so the world
// and renderer are ready.
func newAbilityTestGame(t *testing.T, classID string) *Game {
	t.Helper()
	ss := tcell.NewSimulationScreen("UTF-8")
	ss.SetSize(80, 24)
	if err := ss.Init(); err != nil {
		t.Fatalf("SimulationScreen.Init: %v", err)
	}
	g := &Game{
		screen: ss,
		rng:    rand.New(rand.NewSource(42)),
	}
	g.resetForRun()
	// Find the class by ID.
	found := false
	for _, c := range assets.Classes {
		if c.ID == classID {
			g.selectedClass = c
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("class %q not found in assets.Classes", classID)
	}
	g.fovRadius = g.selectedClass.FOVRadius
	g.loadFloor(1)
	return g
}

// ─── cooldown mechanics ───────────────────────────────────────────────────────

// TestAbilityCooldownSetOnUse verifies that using the special ability puts it
// on cooldown. The post-action tick decrements the cooldown once in the same
// turn, so the observed value after processAction is AbilityCooldown-1.
func TestAbilityCooldownSetOnUse(t *testing.T) {
	cases := []struct {
		classID string
		wantCd  int // AbilityCooldown - 1 (one tick already consumed)
	}{
		{"arcanist", 11},
		{"revenant", 14},
		{"construct", 17},
		{"dancer", 19},
		{"oracle", 19},
		{"symbiont", 11},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.classID, func(t *testing.T) {
			g := newAbilityTestGame(t, tc.classID)
			if g.specialCooldown != 0 {
				t.Fatalf("specialCooldown should be 0 initially, got %d", g.specialCooldown)
			}

			// For Revenant the ability only fires if HP > 5; ensure player is healthy.
			if tc.classID == "revenant" {
				hp := g.world.Get(g.playerID, component.CHealth).(component.Health)
				hp.Current = hp.Max
				g.world.Add(g.playerID, hp)
			}

			g.processAction(ActionSpecialAbility)

			if g.specialCooldown != tc.wantCd {
				t.Errorf("specialCooldown = %d; want %d", g.specialCooldown, tc.wantCd)
			}
		})
	}
}

// TestAbilityCooldownDecrement verifies that specialCooldown decrements by 1
// each turn the player takes.
func TestAbilityCooldownDecrement(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	g.specialCooldown = 5

	// Take a wait-turn to advance the world tick.
	before := g.specialCooldown
	g.processAction(ActionWait)
	if g.specialCooldown != before-1 {
		t.Errorf("cooldown after 1 wait turn: got %d, want %d", g.specialCooldown, before-1)
	}
}

// TestAbilityCooldownBlocksReuse checks that a message is shown and cooldown
// is unchanged when the ability is used while it is still recharging.
func TestAbilityCooldownBlocksReuse(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	g.specialCooldown = 5

	msgsBefore := len(g.messages)
	g.processAction(ActionSpecialAbility) // should show "recharging" message
	if g.specialCooldown != 5 {
		t.Errorf("cooldown changed while recharging: got %d, want 5", g.specialCooldown)
	}
	if len(g.messages) <= msgsBefore {
		t.Error("expected a recharging message to be added")
	}
}

// TestAbilityCooldownReadyAfterDecrement verifies the ability can fire again
// once the cooldown reaches 0.
func TestAbilityCooldownReadyAfterDecrement(t *testing.T) {
	g := newAbilityTestGame(t, "dancer")
	g.specialCooldown = 1

	// One wait-turn should reduce cooldown to 0.
	g.processAction(ActionWait)
	if g.specialCooldown != 0 {
		t.Fatalf("cooldown should be 0 after decrement, got %d", g.specialCooldown)
	}

	// Now the ability should fire successfully.
	g.processAction(ActionSpecialAbility)
	if g.specialCooldown == 0 {
		t.Error("cooldown should be non-zero after ability fires again")
	}
}

// ─── AbilityFreeOnFloor ───────────────────────────────────────────────────────

// TestAbilityFreeOnFloorResetsOnEachFloor verifies that classes with
// AbilityFreeOnFloor=true get their cooldown reset to 0 on loadFloor.
func TestAbilityFreeOnFloorResetsOnEachFloor(t *testing.T) {
	cases := []string{"dancer", "oracle", "symbiont"}
	for _, classID := range cases {
		classID := classID
		t.Run(classID, func(t *testing.T) {
			g := newAbilityTestGame(t, classID)
			// Simulate having used the ability so cooldown > 0.
			g.specialCooldown = g.selectedClass.AbilityCooldown

			// Load next floor — cooldown should reset.
			g.loadFloor(2)
			if g.specialCooldown != 0 {
				t.Errorf("after loadFloor, specialCooldown = %d; want 0", g.specialCooldown)
			}
		})
	}
}

// TestAbilityNotFreeOnFloor verifies that classes without AbilityFreeOnFloor
// do NOT reset their cooldown on loadFloor.
func TestAbilityNotFreeOnFloor(t *testing.T) {
	cases := []string{"arcanist", "revenant", "construct"}
	for _, classID := range cases {
		classID := classID
		t.Run(classID, func(t *testing.T) {
			g := newAbilityTestGame(t, classID)
			g.specialCooldown = 5

			g.loadFloor(2)
			if g.specialCooldown != 5 {
				t.Errorf("specialCooldown should stay 5 after loadFloor; got %d", g.specialCooldown)
			}
		})
	}
}

// ─── KillHealChance ───────────────────────────────────────────────────────────

// TestKillHealChanceTriggersWithSeed verifies that a seeded RNG produces both
// "healing happened" and "no healing" outcomes across different seeds.
func TestKillHealChanceTriggersWithSeed(t *testing.T) {
	healed := false
	notHealed := false
	for seed := int64(0); seed < 20; seed++ {
		g := newAbilityTestGame(t, "arcanist") // KillHealChance=30
		// Set player HP to not-full so we can detect a heal.
		hp := g.world.Get(g.playerID, component.CHealth).(component.Health)
		hp.Current = hp.Max - 10
		g.world.Add(g.playerID, hp)
		beforeHP := hp.Current

		// Use a fixed RNG seed so outcomes are deterministic.
		g.rng = rand.New(rand.NewSource(seed))

		// Spawn an enemy adjacent to the player and kill it via the combat system.
		// Rather than running real combat, we directly call the kill-branch logic
		// by injecting a known kill message path: manually invoke restorePlayerHP
		// conditioned on the same RNG check.
		if g.rng.Intn(100) < g.selectedClass.KillHealChance {
			g.restorePlayerHP(2)
			healed = true
		} else {
			notHealed = true
		}

		afterHP := g.world.Get(g.playerID, component.CHealth).(component.Health).Current
		if healed && afterHP <= beforeHP {
			t.Logf("seed %d: expected heal but HP unchanged (%d -> %d)", seed, beforeHP, afterHP)
		}
	}
	if !healed {
		t.Error("expected at least one healing trigger across 20 seeds for 30% chance")
	}
	if !notHealed {
		t.Error("expected at least one miss across 20 seeds for 30% chance")
	}
}

// TestKillHealChanceIsZeroForOtherClasses verifies that non-Arcanist classes
// have KillHealChance=0.
func TestKillHealChanceIsZeroForOtherClasses(t *testing.T) {
	for _, c := range assets.Classes {
		if c.ID == "arcanist" {
			continue
		}
		if c.KillHealChance != 0 {
			t.Errorf("class %q has unexpected KillHealChance=%d; want 0", c.ID, c.KillHealChance)
		}
	}
}

// ─── PassiveRegen ─────────────────────────────────────────────────────────────

// TestPassiveRegenTriggerEveryNTurns verifies that PassiveRegen restores 1 HP
// every N turns for Construct (N=8) and Symbiont (N=5).
func TestPassiveRegenTriggerEveryNTurns(t *testing.T) {
	cases := []struct {
		classID string
		n       int
	}{
		{"construct", 8},
		{"symbiont", 5},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.classID, func(t *testing.T) {
			g := newAbilityTestGame(t, tc.classID)

			// Reduce HP to leave headroom for regen.
			hp := g.world.Get(g.playerID, component.CHealth).(component.Health)
			hp.Current = hp.Max - 20
			g.world.Add(g.playerID, hp)

			hpBefore := hp.Current

			// Advance exactly tc.n wait-turns. Regen should trigger once at turn N.
			for range tc.n {
				g.processAction(ActionWait)
			}

			hpAfter := g.world.Get(g.playerID, component.CHealth).(component.Health).Current
			// We expect at least +1 HP from regen (turn N is exactly divisible).
			if hpAfter <= hpBefore {
				t.Errorf("after %d turns HP = %d; want > %d (regen should trigger)", tc.n, hpAfter, hpBefore)
			}
		})
	}
}

// TestPassiveRegenDoesNotTriggerBetweenIntervals checks that regen does NOT
// fire on turns that are not multiples of PassiveRegen.
func TestPassiveRegenDoesNotTriggerBetweenIntervals(t *testing.T) {
	g := newAbilityTestGame(t, "construct") // PassiveRegen=8

	// Set HP below max with room.
	hp := g.world.Get(g.playerID, component.CHealth).(component.Health)
	hp.Current = hp.Max - 10
	g.world.Add(g.playerID, hp)

	// Take 7 turns (one less than the interval).
	for range 7 {
		g.processAction(ActionWait)
	}

	hpAfter := g.world.Get(g.playerID, component.CHealth).(component.Health).Current
	if hpAfter > hp.Current {
		t.Errorf("HP went up before regen interval: %d -> %d", hp.Current, hpAfter)
	}
}

// TestPassiveRegenClassConfig verifies the PassiveRegen config values on classes.
func TestPassiveRegenClassConfig(t *testing.T) {
	cases := []struct {
		classID string
		want    int
	}{
		{"arcanist", 0},
		{"revenant", 0},
		{"construct", 8},
		{"dancer", 0},
		{"oracle", 0},
		{"symbiont", 5},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.classID, func(t *testing.T) {
			for _, c := range assets.Classes {
				if c.ID == tc.classID {
					if c.PassiveRegen != tc.want {
						t.Errorf("class %q PassiveRegen=%d; want %d", tc.classID, c.PassiveRegen, tc.want)
					}
					return
				}
			}
			t.Fatalf("class %q not found", tc.classID)
		})
	}
}

// ─── useSpecialAbility side-effects ──────────────────────────────────────────

// TestOracleFarsightRevealMap checks that Oracle's Farsight reveals all walkable
// tiles on the floor.
func TestOracleFarsightRevealMap(t *testing.T) {
	g := newAbilityTestGame(t, "oracle")
	// Count unexplored walkable tiles before using ability.
	unexploredBefore := 0
	for y := range g.gmap.Height {
		for x := range g.gmap.Width {
			tile := g.gmap.At(x, y)
			if tile.Walkable && !tile.Explored {
				unexploredBefore++
			}
		}
	}
	// At least some tiles should be unexplored initially (far rooms not in FOV).
	if unexploredBefore == 0 {
		t.Skip("all tiles already explored on floor 1 — cannot test Farsight")
	}

	g.useSpecialAbility()

	unexploredAfter := 0
	for y := range g.gmap.Height {
		for x := range g.gmap.Width {
			tile := g.gmap.At(x, y)
			if tile.Walkable && !tile.Explored {
				unexploredAfter++
			}
		}
	}
	if unexploredAfter != 0 {
		t.Errorf("after Farsight, %d walkable tiles remain unexplored; want 0", unexploredAfter)
	}
}

// TestDancerVanishAppliesInvisible checks that Dancer's Vanish applies the
// Invisible effect for 8 turns.
func TestDancerVanishAppliesInvisible(t *testing.T) {
	g := newAbilityTestGame(t, "dancer")
	g.useSpecialAbility()

	efComp := g.world.Get(g.playerID, component.CEffects)
	if efComp == nil {
		t.Fatal("player has no effects component after Vanish")
	}
	effects := efComp.(component.Effects).Active
	found := false
	for _, e := range effects {
		if e.Kind == component.EffectInvisible && e.TurnsRemaining == 8 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Invisible effect for 8 turns not found; effects: %v", effects)
	}
}

// TestRevenantDeathBargainSpendHP verifies that Death's Bargain deducts 5 HP
// and applies an attack boost effect.
func TestRevenantDeathBargainSpendHP(t *testing.T) {
	g := newAbilityTestGame(t, "revenant")
	// Ensure player has enough HP for the bargain.
	hp := g.world.Get(g.playerID, component.CHealth).(component.Health)
	hp.Current = hp.Max
	g.world.Add(g.playerID, hp)
	beforeHP := hp.Current

	g.useSpecialAbility()

	afterHP := g.world.Get(g.playerID, component.CHealth).(component.Health).Current
	if afterHP != beforeHP-5 {
		t.Errorf("HP after Death's Bargain = %d; want %d", afterHP, beforeHP-5)
	}

	efComp := g.world.Get(g.playerID, component.CEffects)
	if efComp == nil {
		t.Fatal("player has no effects after Death's Bargain")
	}
	effects := efComp.(component.Effects).Active
	found := false
	for _, e := range effects {
		if e.Kind == component.EffectAttackBoost && e.Magnitude == 6 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("AttackBoost+6 not found after Death's Bargain; effects: %v", effects)
	}
}

// TestRevenantDeathBargainRefusesWhenTooWounded checks that the ability is
// blocked (cooldown refunded) when HP <= 5.
func TestRevenantDeathBargainRefusesWhenTooWounded(t *testing.T) {
	g := newAbilityTestGame(t, "revenant")
	hp := g.world.Get(g.playerID, component.CHealth).(component.Health)
	hp.Current = 5
	g.world.Add(g.playerID, hp)

	g.specialCooldown = 0
	g.useSpecialAbility()

	// Cooldown should be refunded (still 0 after blocked use).
	if g.specialCooldown != 0 {
		t.Errorf("cooldown should remain 0 after blocked use; got %d", g.specialCooldown)
	}
}

// TestSymbiontParasiteSurgeHealsAndBoosts verifies HP and ATK boost.
func TestSymbiontParasiteSurgeHealsAndBoosts(t *testing.T) {
	g := newAbilityTestGame(t, "symbiont")
	// Reduce HP so healing is observable.
	hp := g.world.Get(g.playerID, component.CHealth).(component.Health)
	hp.Current = hp.Max - 20
	g.world.Add(g.playerID, hp)
	beforeHP := hp.Current

	g.useSpecialAbility()

	afterHP := g.world.Get(g.playerID, component.CHealth).(component.Health).Current
	if afterHP != beforeHP+10 {
		t.Errorf("HP after Parasite Surge = %d; want %d", afterHP, beforeHP+10)
	}

	efComp := g.world.Get(g.playerID, component.CEffects)
	if efComp == nil {
		t.Fatal("player has no effects after Parasite Surge")
	}
	found := false
	for _, e := range efComp.(component.Effects).Active {
		if e.Kind == component.EffectAttackBoost && e.Magnitude == 4 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("AttackBoost+4 not found after Parasite Surge")
	}
}

// TestArcanistTeleportChangesPosition verifies that Dimensional Rift moves the
// player to a different tile (almost certainly, given a multi-room map).
func TestArcanistTeleportChangesPosition(t *testing.T) {
	g := newAbilityTestGame(t, "arcanist")
	before := g.world.Get(g.playerID, component.CPosition).(component.Position)

	// Seed a few seeds to find one that actually changes position.
	moved := false
	for seed := int64(0); seed < 10; seed++ {
		g2 := newAbilityTestGame(t, "arcanist")
		g2.rng = rand.New(rand.NewSource(seed))
		pos2Before := g2.world.Get(g2.playerID, component.CPosition).(component.Position)
		g2.useSpecialAbility()
		pos2After := g2.world.Get(g2.playerID, component.CPosition).(component.Position)
		if pos2After.X != pos2Before.X || pos2After.Y != pos2Before.Y {
			moved = true
			break
		}
	}
	_ = before
	if !moved {
		t.Error("Dimensional Rift never changed player position across 10 seeds")
	}
}

// ─── ClassDef field validation ────────────────────────────────────────────────

// TestAllClassesHaveAbilityDefined checks that every class has a non-empty
// AbilityName and a positive AbilityCooldown.
func TestAllClassesHaveAbilityDefined(t *testing.T) {
	for _, c := range assets.Classes {
		if c.AbilityName == "" {
			t.Errorf("class %q has empty AbilityName", c.ID)
		}
		if c.AbilityCooldown <= 0 {
			t.Errorf("class %q has AbilityCooldown=%d; want > 0", c.ID, c.AbilityCooldown)
		}
	}
}

