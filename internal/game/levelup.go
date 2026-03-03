package game

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/component"
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

// grantXP adds XP and checks for level-ups.
func (g *Game) grantXP(amount int) {
	if amount <= 0 {
		return
	}
	g.playerXP += amount
	g.checkLevelUp()
}

// checkLevelUp processes all pending level-ups from accumulated XP.
func (g *Game) checkLevelUp() {
	for g.playerLevel < assets.MaxLevel {
		needed := assets.XPForLevel(g.playerLevel + 1)
		if g.playerXP < needed {
			break
		}
		g.playerLevel++
		g.pendingLevels++
		g.applyLevelGrowth()
		g.addMessage(fmt.Sprintf("You reached level %d!", g.playerLevel))
	}
	if g.pendingLevels > 0 {
		g.addMessage("LEVEL UP! Press [x] to choose a skill.")
	}
}

// applyLevelGrowth applies the stat growth for the current level.
func (g *Game) applyLevelGrowth() {
	growth, ok := assets.ClassGrowths[g.selectedClass.ID]
	if !ok {
		return
	}

	// Compute delta between current level and previous level growth.
	curHP, curATK, curDEF := assets.GrowthForLevel(growth, g.playerLevel)
	prevHP, prevATK, prevDEF := assets.GrowthForLevel(growth, g.playerLevel-1)
	dHP := curHP - prevHP
	dATK := curATK - prevATK
	dDEF := curDEF - prevDEF

	if dHP > 0 {
		g.baseMaxHP += dHP
		if hp := g.world.Get(g.playerID, component.CHealth); hp != nil {
			h := hp.(component.Health)
			h.Max += dHP
			h.Current += dHP
			g.world.Add(g.playerID, h)
		}
	}
	if dATK > 0 || dDEF > 0 {
		if cc := g.world.Get(g.playerID, component.CCombat); cc != nil {
			c := cc.(component.Combat)
			c.Attack += dATK
			c.Defense += dDEF
			g.world.Add(g.playerID, c)
		}
	}
}

// computeSkillBonuses sums all learned skills into a SkillBonuses struct.
func (g *Game) computeSkillBonuses() component.SkillBonuses {
	var sb component.SkillBonuses
	for _, id := range g.learnedSkills {
		s := assets.SkillByID(id)
		if s == nil {
			continue
		}
		sb.BonusATK += s.BonusATK
		sb.BonusDEF += s.BonusDEF
		sb.BonusMaxHP += s.BonusMaxHP
		sb.BonusFOV += s.BonusFOV
		sb.DodgeChance += s.DodgeChance
		sb.KillHealBonus += s.KillHealBonus
		sb.KillHealAdd += s.KillHealAdd
		sb.ThornsDamage += s.ThornsDamage
		sb.CooldownReduce += s.CooldownReduce
		sb.RegenReduce += s.RegenReduce
	}
	return sb
}

// applySkillBonuses writes the computed skill bonuses to the player entity.
func (g *Game) applySkillBonuses() {
	sb := g.computeSkillBonuses()
	g.skillBonusATK = sb.BonusATK
	g.skillBonusDEF = sb.BonusDEF
	g.skillBonusMaxHP = sb.BonusMaxHP
	g.skillBonusFOV = sb.BonusFOV
	if g.world != nil && g.playerID != 0 {
		g.world.Add(g.playerID, sb)
	}
}

// learnSkill adds a skill and recomputes bonuses.
func (g *Game) learnSkill(skillID string) {
	g.learnedSkills = append(g.learnedSkills, skillID)
	g.applySkillBonuses()
	// Recalc MaxHP to include new skill HP bonus.
	g.recalcPlayerMaxHP()
}

// effectiveFOVRadius returns the base FOV plus skill bonuses.
func (g *Game) effectiveFOVRadius() int {
	return g.fovRadius + g.skillBonusFOV
}

// effectiveCooldown returns the class ability cooldown minus skill reductions (min 1).
func (g *Game) effectiveCooldown() int {
	cd := g.selectedClass.AbilityCooldown - g.computeSkillBonuses().CooldownReduce
	if cd < 1 && g.selectedClass.AbilityCooldown > 0 {
		cd = 1
	}
	return cd
}

// effectiveRegenInterval returns the passive regen interval minus skill reductions (min 1).
func (g *Game) effectiveRegenInterval() int {
	base := g.selectedClass.PassiveRegen
	if base <= 0 {
		return 0
	}
	interval := base - g.computeSkillBonuses().RegenReduce
	if interval < 1 {
		interval = 1
	}
	return interval
}

// ─── Level-Up UI ────────────────────────────────────────────────────────────

// runLevelUpScreen shows the modal skill selection UI.
// Returns true if the player picked a skill (turn consumed).
func (g *Game) runLevelUpScreen() bool {
	if g.pendingLevels <= 0 {
		return false
	}

	// Check for specialization milestone first.
	if assets.IsBranchMilestone(g.playerLevel) && g.branch == "" {
		g.runSpecializationScreen()
	}

	// Build available skill pool.
	learned := make(map[string]bool, len(g.learnedSkills))
	for _, id := range g.learnedSkills {
		learned[id] = true
	}
	available := assets.AvailableSkills(g.selectedClass.ID, g.playerLevel, g.branch, learned)
	if len(available) == 0 {
		g.pendingLevels--
		g.addMessage("No skills available to learn.")
		return false
	}

	// Pick 3 random skills to offer.
	offered := pickNSkills(available, 3, g.rng)

	selected := 0
	for {
		DrawLevelUpScreen(g.screen, offered, selected, g.playerLevel, g.pendingLevels)
		g.screen.Show()

		ev := g.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			g.screen.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyUp:
				selected = (selected - 1 + len(offered)) % len(offered)
			case tcell.KeyDown:
				selected = (selected + 1) % len(offered)
			case tcell.KeyEnter:
				g.learnSkill(offered[selected].ID)
				g.pendingLevels--
				g.addMessage(fmt.Sprintf("Learned %s: %s", offered[selected].Name, offered[selected].Description))
				return true
			case tcell.KeyEscape:
				return false // dismiss without choosing
			}
			switch ev.Rune() {
			case 'k', 'K':
				selected = (selected - 1 + len(offered)) % len(offered)
			case 'j', 'J':
				selected = (selected + 1) % len(offered)
			}
		}
	}
}

// runSpecializationScreen shows the branch choice modal.
func (g *Game) runSpecializationScreen() {
	specs, ok := assets.ClassSpecs[g.selectedClass.ID]
	if !ok {
		return
	}
	pair, ok := specs[g.playerLevel]
	if !ok {
		return
	}

	selected := 0
	for {
		DrawSpecScreen(g.screen, pair, selected, g.playerLevel)
		g.screen.Show()

		ev := g.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			g.screen.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyUp:
				selected = 1 - selected
			case tcell.KeyDown:
				selected = 1 - selected
			case tcell.KeyEnter:
				g.branch = pair[selected].Branch
				g.addMessage(fmt.Sprintf("Specialization chosen: %s!", pair[selected].Name))
				return
			}
			switch ev.Rune() {
			case 'k', 'K':
				selected = 1 - selected
			case 'j', 'J':
				selected = 1 - selected
			}
		}
	}
}

// ─── Shared Rendering (used by both single-player and MUD) ──────────────────

// DrawLevelUpScreen renders the skill selection modal.
func DrawLevelUpScreen(screen tcell.Screen, skills []assets.SkillDef, selected, level, pending int) {
	screen.Clear()
	sw, sh := screen.Size()

	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	normalStyle := tcell.StyleDefault.Foreground(tcell.ColorSilver)
	selectedStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
	descStyle := tcell.StyleDefault.Foreground(tcell.ColorAqua)
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)

	width := 50
	boxH := len(skills)*3 + 5
	x0 := (sw - width) / 2
	y0 := (sh - boxH) / 2

	// Border
	for col := x0; col < x0+width; col++ {
		screen.SetContent(col, y0, '─', nil, borderStyle)
		screen.SetContent(col, y0+boxH-1, '─', nil, borderStyle)
	}
	for row := y0; row < y0+boxH; row++ {
		screen.SetContent(x0, row, '│', nil, borderStyle)
		screen.SetContent(x0+width-1, row, '│', nil, borderStyle)
	}
	screen.SetContent(x0, y0, '┌', nil, borderStyle)
	screen.SetContent(x0+width-1, y0, '┐', nil, borderStyle)
	screen.SetContent(x0, y0+boxH-1, '└', nil, borderStyle)
	screen.SetContent(x0+width-1, y0+boxH-1, '┘', nil, borderStyle)

	title := fmt.Sprintf(" Level %d — Choose a Skill (%d pending) ", level, pending)
	tx := x0 + (width-len([]rune(title)))/2
	putScreenText(screen, tx, y0, title, titleStyle)

	y := y0 + 2
	for i, s := range skills {
		style := normalStyle
		prefix := "  "
		if i == selected {
			style = selectedStyle
			prefix = "> "
		}
		putScreenText(screen, x0+2, y, prefix+s.Name, style)
		putScreenText(screen, x0+4, y+1, s.Description, descStyle)
		y += 3
	}

	putScreenText(screen, x0+2, y0+boxH-2, "[j/k] select  [Enter] learn  [Esc] later", normalStyle)
}

// DrawSpecScreen renders the specialization choice modal.
func DrawSpecScreen(screen tcell.Screen, pair [2]assets.SpecDef, selected, level int) {
	screen.Clear()
	sw, sh := screen.Size()

	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	normalStyle := tcell.StyleDefault.Foreground(tcell.ColorSilver)
	selectedStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
	descStyle := tcell.StyleDefault.Foreground(tcell.ColorAqua)
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)

	width := 50
	boxH := 11
	x0 := (sw - width) / 2
	y0 := (sh - boxH) / 2

	for col := x0; col < x0+width; col++ {
		screen.SetContent(col, y0, '─', nil, borderStyle)
		screen.SetContent(col, y0+boxH-1, '─', nil, borderStyle)
	}
	for row := y0; row < y0+boxH; row++ {
		screen.SetContent(x0, row, '│', nil, borderStyle)
		screen.SetContent(x0+width-1, row, '│', nil, borderStyle)
	}
	screen.SetContent(x0, y0, '┌', nil, borderStyle)
	screen.SetContent(x0+width-1, y0, '┐', nil, borderStyle)
	screen.SetContent(x0, y0+boxH-1, '└', nil, borderStyle)
	screen.SetContent(x0+width-1, y0+boxH-1, '┘', nil, borderStyle)

	title := fmt.Sprintf(" Level %d — Choose Specialization ", level)
	tx := x0 + (width-len([]rune(title)))/2
	putScreenText(screen, tx, y0, title, titleStyle)

	for i, spec := range pair {
		y := y0 + 2 + i*3
		style := normalStyle
		prefix := "  "
		if i == selected {
			style = selectedStyle
			prefix = "> "
		}
		putScreenText(screen, x0+2, y, prefix+spec.Name, style)
		putScreenText(screen, x0+4, y+1, spec.Description, descStyle)
	}

	putScreenText(screen, x0+2, y0+boxH-2, "[j/k] select  [Enter] choose", normalStyle)
}

// putScreenText is a helper to write text directly to a tcell.Screen.
func putScreenText(screen tcell.Screen, x, y int, s string, style tcell.Style) {
	for _, r := range s {
		screen.SetContent(x, y, r, nil, style)
		x++
	}
}

// pickNSkills returns up to n randomly selected skills from the pool.
func pickNSkills(pool []assets.SkillDef, n int, rng *rand.Rand) []assets.SkillDef {
	if len(pool) <= n {
		return pool
	}
	// Fisher-Yates partial shuffle.
	perm := make([]assets.SkillDef, len(pool))
	copy(perm, pool)
	for i := range n {
		j := i + rng.Intn(len(perm)-i)
		perm[i], perm[j] = perm[j], perm[i]
	}
	return perm[:n]
}
