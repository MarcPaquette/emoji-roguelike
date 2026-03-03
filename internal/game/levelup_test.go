package game

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/component"
	"math/rand"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func newLevelTestGame(t *testing.T, classID string) *Game {
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
	for _, c := range assets.Classes {
		if c.ID == classID {
			g.selectedClass = c
			break
		}
	}
	g.fovRadius = g.selectedClass.FOVRadius
	g.floorsVisited[1] = true // suppress floor-entry XP
	g.loadFloor(1)
	return g
}

func TestGrantXPTriggersLevelUp(t *testing.T) {
	g := newLevelTestGame(t, "arcanist")
	if g.playerLevel != 1 {
		t.Fatalf("initial level = %d; want 1", g.playerLevel)
	}
	// Grant exactly enough for level 2.
	g.grantXP(assets.XPForLevel(2))
	if g.playerLevel != 2 {
		t.Errorf("level after XP grant = %d; want 2", g.playerLevel)
	}
	if g.pendingLevels != 1 {
		t.Errorf("pendingLevels = %d; want 1", g.pendingLevels)
	}
}

func TestMultipleLevelUpsFromLargeXP(t *testing.T) {
	g := newLevelTestGame(t, "arcanist")
	// Grant a massive amount of XP — should reach multiple levels.
	g.grantXP(10000)
	if g.playerLevel <= 5 {
		t.Errorf("level = %d; expected > 5 from 10000 XP", g.playerLevel)
	}
	if g.pendingLevels != g.playerLevel-1 {
		t.Errorf("pendingLevels = %d; want %d", g.pendingLevels, g.playerLevel-1)
	}
}

func TestStatGrowthAppliesPerClass(t *testing.T) {
	for _, c := range assets.Classes {
		t.Run(c.ID, func(t *testing.T) {
			g := newLevelTestGame(t, c.ID)
			baseCombat := g.world.Get(g.playerID, component.CCombat).(component.Combat)
			baseHP := g.world.Get(g.playerID, component.CHealth).(component.Health)

			// Grant enough XP for level 6 (should trigger ATK/DEF growth at level 6 for most classes).
			g.grantXP(assets.XPForLevel(6) + 1)

			growth := assets.ClassGrowths[c.ID]
			wantHP, wantATK, wantDEF := assets.GrowthForLevel(growth, g.playerLevel)

			gotHP := g.world.Get(g.playerID, component.CHealth).(component.Health)
			if gotHP.Max != baseHP.Max+wantHP {
				t.Errorf("MaxHP = %d; want %d (base %d + growth %d)", gotHP.Max, baseHP.Max+wantHP, baseHP.Max, wantHP)
			}

			gotCombat := g.world.Get(g.playerID, component.CCombat).(component.Combat)
			if gotCombat.Attack != baseCombat.Attack+wantATK {
				t.Errorf("ATK = %d; want %d (base %d + growth %d)", gotCombat.Attack, baseCombat.Attack+wantATK, baseCombat.Attack, wantATK)
			}
			if gotCombat.Defense != baseCombat.Defense+wantDEF {
				t.Errorf("DEF = %d; want %d (base %d + growth %d)", gotCombat.Defense, baseCombat.Defense+wantDEF, baseCombat.Defense, wantDEF)
			}
		})
	}
}

func TestSkillBonusesCompute(t *testing.T) {
	g := newLevelTestGame(t, "arcanist")
	g.learnSkill("arc_t0_atk") // +2 ATK
	g.learnSkill("arc_t0_hp")  // +3 MaxHP

	sb := g.computeSkillBonuses()
	if sb.BonusATK != 2 {
		t.Errorf("BonusATK = %d; want 2", sb.BonusATK)
	}
	if sb.BonusMaxHP != 3 {
		t.Errorf("BonusMaxHP = %d; want 3", sb.BonusMaxHP)
	}
}

func TestSkillBonusesStackMultiple(t *testing.T) {
	g := newLevelTestGame(t, "revenant")
	g.learnSkill("rev_t0_atk")    // +3 ATK
	g.learnSkill("rev_t0_thorns") // +1 thorns

	sb := g.computeSkillBonuses()
	if sb.BonusATK != 3 {
		t.Errorf("BonusATK = %d; want 3", sb.BonusATK)
	}
	if sb.ThornsDamage != 1 {
		t.Errorf("ThornsDamage = %d; want 1", sb.ThornsDamage)
	}
}

func TestSkillBonusesWrittenToEntity(t *testing.T) {
	g := newLevelTestGame(t, "arcanist")
	g.learnSkill("arc_t0_atk") // +2 ATK

	c := g.world.Get(g.playerID, component.CSkillBonuses)
	if c == nil {
		t.Fatal("CSkillBonuses component not found on player")
	}
	sb := c.(component.SkillBonuses)
	if sb.BonusATK != 2 {
		t.Errorf("entity SkillBonuses.BonusATK = %d; want 2", sb.BonusATK)
	}
}

func TestSkillsSurviveFloorTransition(t *testing.T) {
	g := newLevelTestGame(t, "arcanist")
	g.learnSkill("arc_t0_atk")
	g.grantXP(500) // level up a bit

	origLevel := g.playerLevel
	origSkills := len(g.learnedSkills)

	g.floorsVisited[2] = true // suppress floor-entry XP
	g.loadFloor(2)

	if g.playerLevel != origLevel {
		t.Errorf("level changed from %d to %d after floor transition", origLevel, g.playerLevel)
	}
	if len(g.learnedSkills) != origSkills {
		t.Errorf("skills changed from %d to %d after floor transition", origSkills, len(g.learnedSkills))
	}

	// Verify skill bonuses are applied to new entity.
	c := g.world.Get(g.playerID, component.CSkillBonuses)
	if c == nil {
		t.Fatal("CSkillBonuses not on player after floor transition")
	}
	sb := c.(component.SkillBonuses)
	if sb.BonusATK != 2 {
		t.Errorf("SkillBonuses.BonusATK = %d; want 2 after floor transition", sb.BonusATK)
	}
}

func TestSkillsResetOnResetForRun(t *testing.T) {
	g := newLevelTestGame(t, "arcanist")
	g.learnSkill("arc_t0_atk")
	g.grantXP(1000)

	g.resetForRun()

	if g.playerLevel != 1 {
		t.Errorf("playerLevel = %d; want 1 after reset", g.playerLevel)
	}
	if g.playerXP != 0 {
		t.Errorf("playerXP = %d; want 0 after reset", g.playerXP)
	}
	if len(g.learnedSkills) != 0 {
		t.Errorf("learnedSkills = %v; want empty after reset", g.learnedSkills)
	}
	if g.branch != "" {
		t.Errorf("branch = %q; want empty after reset", g.branch)
	}
	if g.pendingLevels != 0 {
		t.Errorf("pendingLevels = %d; want 0 after reset", g.pendingLevels)
	}
}

func TestLevelCap(t *testing.T) {
	g := newLevelTestGame(t, "arcanist")
	g.grantXP(1_000_000_000) // massive XP
	if g.playerLevel > assets.MaxLevel {
		t.Errorf("playerLevel = %d; should not exceed %d", g.playerLevel, assets.MaxLevel)
	}
}

func TestPendingLevelsAccumulate(t *testing.T) {
	g := newLevelTestGame(t, "arcanist")
	g.grantXP(assets.XPForLevel(5) + 1) // should reach at least level 5
	if g.pendingLevels < 4 {
		t.Errorf("pendingLevels = %d; want >= 4", g.pendingLevels)
	}
}

func TestEffectiveCooldownReduction(t *testing.T) {
	g := newLevelTestGame(t, "arcanist")
	baseCd := g.selectedClass.AbilityCooldown // 12
	g.learnSkill("arc_t0_cd")                // -2 cooldown

	cd := g.effectiveCooldown()
	if cd != baseCd-2 {
		t.Errorf("effectiveCooldown = %d; want %d", cd, baseCd-2)
	}
}

func TestEffectiveCooldownMinimum(t *testing.T) {
	g := newLevelTestGame(t, "arcanist")
	// Learn multiple cooldown reduction skills.
	g.learnSkill("arc_t0_cd")  // -2
	g.learnSkill("arc_t1a_cd") // -3
	g.learnSkill("arc_t1b_cd") // -4 (normally inaccessible but testing the floor)

	cd := g.effectiveCooldown()
	if cd < 1 {
		t.Errorf("effectiveCooldown = %d; should be >= 1", cd)
	}
}

func TestEffectiveFOVRadius(t *testing.T) {
	g := newLevelTestGame(t, "arcanist")
	baseFOV := g.fovRadius
	g.learnSkill("arc_t0_fov") // +1 FOV

	if g.effectiveFOVRadius() != baseFOV+1 {
		t.Errorf("effectiveFOVRadius = %d; want %d", g.effectiveFOVRadius(), baseFOV+1)
	}
}

func TestEffectiveRegenInterval(t *testing.T) {
	g := newLevelTestGame(t, "construct")
	baseRegen := g.selectedClass.PassiveRegen // 8
	g.learnSkill("con_t0_regen")              // -2 regen interval

	ri := g.effectiveRegenInterval()
	if ri != baseRegen-2 {
		t.Errorf("effectiveRegenInterval = %d; want %d", ri, baseRegen-2)
	}
}

func TestEffectiveRegenIntervalMin(t *testing.T) {
	g := newLevelTestGame(t, "symbiont") // PassiveRegen=5
	g.learnSkill("sym_t0_regen")         // -1
	g.learnSkill("sym_t1a_regen")        // -2
	g.learnSkill("sym_t1b_regen")        // -2

	ri := g.effectiveRegenInterval()
	if ri < 1 {
		t.Errorf("effectiveRegenInterval = %d; should be >= 1", ri)
	}
}

func TestZeroXPNoLevelUp(t *testing.T) {
	g := newLevelTestGame(t, "arcanist")
	g.grantXP(0)
	if g.playerLevel != 1 {
		t.Errorf("playerLevel = %d; want 1 after 0 XP", g.playerLevel)
	}
}

func TestFloorEntryXPGrantedOnce(t *testing.T) {
	g := newLevelTestGame(t, "arcanist")
	// Manually reset floorsVisited to simulate fresh state.
	g.floorsVisited = make(map[int]bool)
	g.loadFloor(1)
	xpAfterFirst := g.playerXP

	// Load floor 1 again — should not grant more XP.
	g.loadFloor(1)
	if g.playerXP != xpAfterFirst {
		t.Errorf("XP after second floor-1 entry = %d; want %d (no duplicate)", g.playerXP, xpAfterFirst)
	}
}
