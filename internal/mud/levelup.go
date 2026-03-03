package mud

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/game"
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

// grantXPLocked adds XP to a session and checks for level-ups.
// Caller must hold s.mu.
func grantXPLocked(sess *Session, amount int) {
	if amount <= 0 {
		return
	}
	sess.XP += amount
	for sess.Level < assets.MaxLevel {
		needed := assets.XPForLevel(sess.Level + 1)
		if sess.XP < needed {
			break
		}
		sess.Level++
		sess.PendingLevels++
		sess.AddMessage(fmt.Sprintf("You reached level %d!", sess.Level))
	}
	if sess.PendingLevels > 0 {
		sess.AddMessage("LEVEL UP! Press [x] to choose a skill.")
	}
}

// applyLevelGrowthLocked applies cumulative stat growth for the current level.
// Caller must hold s.mu.
func applyLevelGrowthLocked(w *ecs.World, sess *Session) {
	growth, ok := assets.ClassGrowths[sess.Class.ID]
	if !ok {
		return
	}
	hp, atk, def := assets.GrowthForLevel(growth, sess.Level)
	// BaseMaxHP starts at class base; add growth HP.
	sess.BaseMaxHP = sess.Class.MaxHP + hp

	if sess.PlayerID == ecs.NilEntity {
		return
	}

	// Apply ATK/DEF growth on top of class base + furniture.
	if cc := w.Get(sess.PlayerID, component.CCombat); cc != nil {
		c := cc.(component.Combat)
		c.Attack = sess.Class.Attack + atk + sess.FurnitureATK
		c.Defense = sess.Class.Defense + def + sess.FurnitureDEF
		w.Add(sess.PlayerID, c)
	}
}

// computeSessionSkillBonuses sums all learned skills into a SkillBonuses.
func computeSessionSkillBonuses(sess *Session) component.SkillBonuses {
	var sb component.SkillBonuses
	for _, id := range sess.LearnedSkills {
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

// applySkillBonusesLocked writes skill bonuses to the player entity.
func applySkillBonusesLocked(w *ecs.World, sess *Session) {
	if sess.PlayerID == ecs.NilEntity {
		return
	}
	sb := computeSessionSkillBonuses(sess)
	w.Add(sess.PlayerID, sb)
}

// effectiveCooldown returns the cooldown for the session's class ability.
func effectiveCooldown(sess *Session) int {
	cd := sess.Class.AbilityCooldown - computeSessionSkillBonuses(sess).CooldownReduce
	if cd < 1 && sess.Class.AbilityCooldown > 0 {
		cd = 1
	}
	return cd
}

// effectiveRegenInterval returns the regen interval for the session.
func effectiveRegenInterval(sess *Session) int {
	base := sess.Class.PassiveRegen
	if base <= 0 {
		return 0
	}
	interval := base - computeSessionSkillBonuses(sess).RegenReduce
	if interval < 1 {
		interval = 1
	}
	return interval
}

// effectiveFOVRadius returns the FOV radius including skill bonuses.
func effectiveFOVRadius(sess *Session) int {
	return sess.FovRadius + computeSessionSkillBonuses(sess).BonusFOV
}

// RunLevelUp runs the modal level-up skill selection.
// Blocks on eventCh until the player picks a skill or dismisses.
func (s *Server) RunLevelUp(sess *Session, eventCh <-chan tcell.Event) {
	if sess.PendingLevels <= 0 {
		return
	}

	// Check for specialization milestone first.
	if assets.IsBranchMilestone(sess.Level) && sess.Branch == "" {
		s.RunSpecialization(sess, eventCh)
	}

	learned := make(map[string]bool, len(sess.LearnedSkills))
	for _, id := range sess.LearnedSkills {
		learned[id] = true
	}
	available := assets.AvailableSkills(sess.Class.ID, sess.Level, sess.Branch, learned)
	if len(available) == 0 {
		sess.PendingLevels--
		sess.AddMessage("No skills available to learn.")
		return
	}

	rng := rand.New(rand.NewSource(int64(sess.ID) + int64(sess.Level)))
	offered := pickNSkillsMud(available, 3, rng)

	selected := 0
	for {
		game.DrawLevelUpScreen(sess.Screen, offered, selected, sess.Level, sess.PendingLevels)
		sess.Screen.Show()

		ev, ok := <-eventCh
		if !ok {
			return
		}
		switch ev := ev.(type) {
		case *tcell.EventResize:
			sess.Screen.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyUp:
				selected = (selected - 1 + len(offered)) % len(offered)
			case tcell.KeyDown:
				selected = (selected + 1) % len(offered)
			case tcell.KeyEnter:
				s.mu.Lock()
				sess.LearnedSkills = append(sess.LearnedSkills, offered[selected].ID)
				sess.PendingLevels--
				if floor, ok := s.floors[sess.FloorNum]; ok {
					applySkillBonusesLocked(floor.World, sess)
					applyLevelGrowthLocked(floor.World, sess)
					recalcMaxHPWithSkills(floor.World, sess)
				}
				sess.AddMessage(fmt.Sprintf("Learned %s: %s", offered[selected].Name, offered[selected].Description))
				s.mu.Unlock()
				return
			case tcell.KeyEscape:
				return
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

// RunSpecialization runs the modal branch specialization choice.
func (s *Server) RunSpecialization(sess *Session, eventCh <-chan tcell.Event) {
	specs, ok := assets.ClassSpecs[sess.Class.ID]
	if !ok {
		return
	}
	pair, ok := specs[sess.Level]
	if !ok {
		return
	}

	selected := 0
	for {
		game.DrawSpecScreen(sess.Screen, pair, selected, sess.Level)
		sess.Screen.Show()

		ev, ok := <-eventCh
		if !ok {
			return
		}
		switch ev := ev.(type) {
		case *tcell.EventResize:
			sess.Screen.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyUp:
				selected = 1 - selected
			case tcell.KeyDown:
				selected = 1 - selected
			case tcell.KeyEnter:
				sess.Branch = pair[selected].Branch
				sess.AddMessage(fmt.Sprintf("Specialization chosen: %s!", pair[selected].Name))
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

// recalcMaxHPWithSkills recalculates MaxHP including skill bonuses.
func recalcMaxHPWithSkills(w *ecs.World, sess *Session) {
	invComp := w.Get(sess.PlayerID, component.CInventory)
	if invComp == nil {
		return
	}
	inv := invComp.(component.Inventory)
	sb := computeSessionSkillBonuses(sess)
	bonus := inv.Head.BonusMaxHP + inv.Body.BonusMaxHP + inv.Feet.BonusMaxHP +
		inv.MainHand.BonusMaxHP + inv.OffHand.BonusMaxHP + sb.BonusMaxHP
	hpComp := w.Get(sess.PlayerID, component.CHealth)
	if hpComp == nil {
		return
	}
	hp := hpComp.(component.Health)
	hp.Max = sess.BaseMaxHP + bonus
	if hp.Current > hp.Max {
		hp.Current = hp.Max
	}
	w.Add(sess.PlayerID, hp)
}

func pickNSkillsMud(pool []assets.SkillDef, n int, rng *rand.Rand) []assets.SkillDef {
	if len(pool) <= n {
		return pool
	}
	perm := make([]assets.SkillDef, len(pool))
	copy(perm, pool)
	for i := range n {
		j := i + rng.Intn(len(perm)-i)
		perm[i], perm[j] = perm[j], perm[i]
	}
	return perm[:n]
}
