package game

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/factory"
	"emoji-roguelike/internal/gamemap"
	"emoji-roguelike/internal/generate"
	"emoji-roguelike/internal/render"
	"emoji-roguelike/internal/system"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
)

// GameState tracks the main state machine.
type GameState uint8

const (
	StatePlaying GameState = iota
	StateInventory
	StateDead
	StateVictory
	StateClassSelect
)

// RunLog records statistics gathered during one run.
type RunLog struct {
	Timestamp        time.Time      `json:"timestamp"`
	Victory          bool           `json:"victory"`
	Class            string         `json:"class"`
	FloorsReached    int            `json:"floors_reached"`
	TurnsPlayed      int            `json:"turns_played"`
	EnemiesKilled    map[string]int `json:"enemies_killed"` // glyph → kill count
	ItemsUsed        map[string]int `json:"items_used"`     // glyph → use count
	InscriptionsRead int            `json:"inscriptions_read"`
	DamageDealt      int            `json:"damage_dealt"`
	DamageTaken      int            `json:"damage_taken"`
	CauseOfDeath     string         `json:"cause_of_death"` // last thing that hurt the player ("poison" or enemy glyph)
}

// Game is the top-level orchestrator.
type Game struct {
	screen            tcell.Screen
	renderer          *render.Renderer
	world             *ecs.World
	gmap              *gamemap.GameMap
	playerID          ecs.EntityID
	rng               *rand.Rand
	floor             int
	state             GameState
	messages          []string
	selectedClass     assets.ClassDef
	fovRadius         int
	baseMaxHP         int // base MaxHP from class, used by recalcPlayerMaxHP
	discoveredEnemies map[string]bool
	runLog            RunLog
}

// New creates and returns a Game with screen initialized.
// Floor loading is deferred until after class selection in Run().
func New() (*Game, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("create screen: %w", err)
	}
	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("init screen: %w", err)
	}
	screen.EnableMouse()

	g := &Game{
		screen: screen,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	g.resetForRun()
	return g, nil
}

// resetForRun clears all per-run state in preparation for a fresh start.
func (g *Game) resetForRun() {
	g.floor = 1
	g.state = StatePlaying
	g.messages = nil
	g.world = nil
	g.gmap = nil
	g.baseMaxHP = 0
	g.discoveredEnemies = make(map[string]bool)
	g.runLog = RunLog{
		EnemiesKilled: make(map[string]int),
		ItemsUsed:     make(map[string]int),
	}
}

// loadFloor generates and populates the given floor.
// On transitions (floor > 1) the player's current HP and inventory are preserved.
func (g *Game) loadFloor(floor int) {
	// Preserve HP and inventory across floor transitions.
	savedHP := -1
	var savedInv *component.Inventory
	if g.world != nil && g.playerID != ecs.NilEntity {
		if hpComp := g.world.Get(g.playerID, component.CHealth); hpComp != nil {
			savedHP = hpComp.(component.Health).Current
		}
		if invComp := g.world.Get(g.playerID, component.CInventory); invComp != nil {
			inv := invComp.(component.Inventory)
			savedInv = &inv
		}
	}

	// Set base MaxHP from class on the first floor.
	if floor == 1 {
		g.baseMaxHP = g.selectedClass.MaxHP
	}

	g.floor = floor
	if floor > g.runLog.FloorsReached {
		g.runLog.FloorsReached = floor
	}
	g.world = ecs.NewWorld()

	cfg := levelConfig(floor, g.rng)
	gmap, px, py := generate.Generate(cfg)
	g.gmap = gmap

	// Populate enemies, items, inscriptions, and equipment.
	pop := generate.Populate(gmap, cfg)
	for _, es := range pop.Enemies {
		factory.NewEnemy(g.world, es.Entry, es.X, es.Y)
	}
	for _, is := range pop.Items {
		factory.NewItem(g.world, is.Entry, is.X, is.Y)
	}
	for _, eq := range pop.Equipment {
		factory.NewEquipItem(g.world, eq.Entry, floor, g.rng, eq.X, eq.Y)
	}
	for _, ins := range pop.Inscriptions {
		factory.NewInscription(g.world, ins.Text, ins.X, ins.Y)
	}

	// Create player using the selected class definition.
	g.playerID = factory.NewPlayer(g.world, px, py, g.selectedClass)

	// Restore HP from previous floor (capped at max).
	if savedHP > 0 && floor > 1 {
		hp := g.world.Get(g.playerID, component.CHealth).(component.Health)
		if savedHP < hp.Max {
			hp.Current = savedHP
		}
		g.world.Add(g.playerID, hp)
	}

	// Restore inventory from previous floor.
	if savedInv != nil && floor > 1 {
		g.world.Add(g.playerID, *savedInv)
		g.recalcPlayerMaxHP()
	}

	// Apply class passive effects on floor 1 only.
	if floor == 1 {
		if g.selectedClass.StartInvisible > 0 {
			system.ApplyEffect(g.world, g.playerID, component.ActiveEffect{
				Kind:           component.EffectInvisible,
				Magnitude:      1,
				TurnsRemaining: g.selectedClass.StartInvisible,
			})
		}
		if g.selectedClass.StartRevealMap {
			for y := 0; y < gmap.Height; y++ {
				for x := 0; x < gmap.Width; x++ {
					if gmap.At(x, y).Walkable {
						gmap.At(x, y).Explored = true
					}
				}
			}
		}
		// Spawn class start items adjacent to the player.
		for i, glyph := range g.selectedClass.StartItems {
			offsets := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
			ox, oy := offsets[i%len(offsets)][0], offsets[i%len(offsets)][1]
			ix, iy := px+ox, py+oy
			if !gmap.InBounds(ix, iy) || !gmap.IsWalkable(ix, iy) {
				ix, iy = px, py
			}
			factory.NewItemByGlyph(g.world, glyph, ix, iy)
		}
	}

	system.UpdateFOV(g.world, g.gmap, g.playerID, g.fovRadius)
	g.renderer = render.NewRenderer(g.screen, floor)
	g.renderer.CenterOn(px, py)

	if floor == 1 {
		g.addMessage(fmt.Sprintf("You enter the %s as a %s.", assets.FloorNames[floor], g.selectedClass.Name))
	} else {
		g.addMessage(fmt.Sprintf("You descend into %s (Floor %d).", assets.FloorNames[floor], floor))
	}
	if lore := assets.FloorLore[floor]; len(lore) > 0 {
		g.addMessage(lore[g.rng.Intn(len(lore))])
	}
}

// Run is the main game loop. Supports multiple consecutive runs via Try Again.
func (g *Game) Run() {
	defer g.screen.Fini()

	for {
		g.resetForRun()

		if !g.runClassSelect() {
			return
		}
		g.runLog.Class = g.selectedClass.Name

		g.loadFloor(1)
		g.addMessage("Use hjklyubn or arrow keys to move. > to descend.")

		for g.state != StateDead && g.state != StateVictory {
			playerPos := g.playerPosition()
			g.renderer.CenterOn(playerPos.X, playerPos.Y)
			g.renderer.DrawFrame(g.world, g.gmap, g.playerID)
			// Compute equipment + effect bonuses for HUD display.
			equipATK, equipDEF := g.equipBonuses()
			bonusATK := system.GetAttackBonus(g.world, g.playerID) + equipATK
			bonusDEF := system.GetDefenseBonus(g.world, g.playerID) + equipDEF
			g.renderer.DrawHUD(g.world, g.playerID, g.floor, g.selectedClass.Name, g.messages, bonusATK, bonusDEF)

			ev := g.screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventResize:
				g.screen.Sync()
				continue
			case *tcell.EventKey:
				action := keyToAction(ev)
				if action == ActionQuit {
					if g.confirmQuit(func() {
						playerPos := g.playerPosition()
						g.renderer.CenterOn(playerPos.X, playerPos.Y)
						g.renderer.DrawFrame(g.world, g.gmap, g.playerID)
						equipATK, equipDEF := g.equipBonuses()
						bonusATK := system.GetAttackBonus(g.world, g.playerID) + equipATK
						bonusDEF := system.GetDefenseBonus(g.world, g.playerID) + equipDEF
						g.renderer.DrawHUD(g.world, g.playerID, g.floor, g.selectedClass.Name, g.messages, bonusATK, bonusDEF)
					}) {
						return
					}
					continue
				}
				g.processAction(action)
			}
		}

		g.runLog.Victory = g.state == StateVictory
		g.runLog.Timestamp = time.Now()
		if g.runLog.Victory {
			g.runLog.CauseOfDeath = ""
		}
		saveRunLog(g.runLog)

		if !g.showEndScreen() {
			return
		}
	}
}

// processAction handles one player action and optionally advances enemy AI.
func (g *Game) processAction(action Action) {
	// If stunned, skip all player actions and just run a world tick.
	if system.IsStunned(g.world, g.playerID) {
		g.addMessage("You are stunned and cannot act!")
		g.runLog.TurnsPlayed++
		g.applyPoisonDamage()
		system.TickEffects(g.world)
		hits := system.ProcessAI(g.world, g.gmap, g.playerID, g.rng)
		for _, h := range hits {
			if h.Damage > 0 {
				g.runLog.DamageTaken += h.Damage
				g.runLog.CauseOfDeath = h.EnemyGlyph
			}
			g.handleSpecialHitMessage(h)
		}
		g.checkPlayerDead()
		return
	}

	turnUsed := false

	switch action {
	case ActionWait:
		turnUsed = true
		g.addMessage("You wait.")

	case ActionDescend:
		pos := g.playerPosition()
		tile := g.gmap.At(pos.X, pos.Y)
		if tile.Kind == gamemap.TileStairsDown {
			if g.floor >= MaxFloors {
				g.addMessage("There is nowhere further to descend.")
			} else {
				g.loadFloor(g.floor + 1)
				return
			}
		} else {
			g.addMessage("There are no stairs down here.")
		}

	case ActionAscend:
		pos := g.playerPosition()
		tile := g.gmap.At(pos.X, pos.Y)
		if tile.Kind == gamemap.TileStairsUp && g.floor > 1 {
			g.loadFloor(g.floor - 1)
			return
		} else {
			g.addMessage("There are no stairs up here.")
		}

	case ActionPickup:
		g.tryPickup()
		turnUsed = true

	case ActionInventory:
		turnUsed = g.runInventoryScreen()

	default:
		dx, dy := actionToDelta(action)
		if dx != 0 || dy != 0 {
			result, target := system.TryMove(g.world, g.gmap, g.playerID, dx, dy)
			switch result {
			case system.MoveOK:
				turnUsed = true
				system.UpdateFOV(g.world, g.gmap, g.playerID, g.fovRadius)
				g.checkInscription()
			case system.MoveAttack:
				// Capture name/glyph/position/loot BEFORE Attack() which may destroy the entity.
				name := g.entityName(target)
				glyph := name
				enemyPos := g.world.Get(target, component.CPosition).(component.Position)
				var lootDrops []component.LootEntry
				if lc := g.world.Get(target, component.CLoot); lc != nil {
					lootDrops = lc.(component.Loot).Drops
				}
				res := system.Attack(g.world, g.rng, g.playerID, target)
				g.runLog.DamageDealt += res.Damage
				if res.Killed {
					g.runLog.EnemiesKilled[glyph]++
					g.addMessage(fmt.Sprintf("You kill the %s!", name))
					if !g.discoveredEnemies[glyph] {
						g.discoveredEnemies[glyph] = true
						if lore, ok := assets.EnemyLore[glyph]; ok {
							g.addMessage(lore)
						}
					}
					for _, d := range lootDrops {
						if g.rng.Intn(100) < d.Chance {
							factory.NewItemByGlyph(g.world, d.Glyph, enemyPos.X, enemyPos.Y)
							g.addMessage(fmt.Sprintf("The %s drops something!", name))
						}
					}
					if g.selectedClass.KillRestoreHP > 0 {
						g.restorePlayerHP(g.selectedClass.KillRestoreHP)
						g.addMessage(fmt.Sprintf("The kill feeds you. (+%d HP)", g.selectedClass.KillRestoreHP))
					}
					g.checkVictory()
				} else {
					g.addMessage(fmt.Sprintf("You hit the %s for %d damage.", name, res.Damage))
				}
				turnUsed = true
			case system.MoveBlocked:
				// no message for walking into walls
			}
		}
	}

	if turnUsed {
		g.runLog.TurnsPlayed++
		g.applyPoisonDamage()
		system.TickEffects(g.world)
		hits := system.ProcessAI(g.world, g.gmap, g.playerID, g.rng)
		for _, h := range hits {
			if h.Damage > 0 {
				g.runLog.DamageTaken += h.Damage
				g.runLog.CauseOfDeath = h.EnemyGlyph
			}
			g.handleSpecialHitMessage(h)
		}
		g.checkPlayerDead()
	}
}

func (g *Game) handleSpecialHitMessage(h system.EnemyHitResult) {
	switch h.SpecialApplied {
	case 1:
		g.addMessage(fmt.Sprintf("The %s poisons you!", h.EnemyGlyph))
	case 2:
		g.addMessage(fmt.Sprintf("The %s weakens your attack!", h.EnemyGlyph))
	case 3:
		g.addMessage(fmt.Sprintf("The %s drains your life force! (+%d HP to enemy)", h.EnemyGlyph, h.DrainedAmount))
	case 4:
		g.addMessage(fmt.Sprintf("The %s stuns you! (skip next turn)", h.EnemyGlyph))
	case 5:
		g.addMessage(fmt.Sprintf("The %s shatters your defenses!", h.EnemyGlyph))
	}
}

func (g *Game) applyPoisonDamage() {
	poisonDmg := system.GetPoisonDamage(g.world, g.playerID)
	burnDmg := system.GetSelfBurnDamage(g.world, g.playerID)
	totalDmg := poisonDmg + burnDmg
	if totalDmg <= 0 {
		return
	}
	hp := g.world.Get(g.playerID, component.CHealth)
	if hp == nil {
		return
	}
	h := hp.(component.Health)
	h.Current -= totalDmg
	g.world.Add(g.playerID, h)
	g.runLog.DamageTaken += totalDmg
	if poisonDmg > 0 {
		g.runLog.CauseOfDeath = "poison"
		g.addMessage(fmt.Sprintf("Poison burns through you! (%d damage)", poisonDmg))
	}
	if burnDmg > 0 {
		if g.runLog.CauseOfDeath == "" {
			g.runLog.CauseOfDeath = "self-burn"
		}
		g.addMessage(fmt.Sprintf("The resonance burns you! (%d damage)", burnDmg))
	}
}

func (g *Game) restorePlayerHP(n int) {
	hpComp := g.world.Get(g.playerID, component.CHealth)
	if hpComp == nil {
		return
	}
	hp := hpComp.(component.Health)
	hp.Current += n
	if hp.Current > hp.Max {
		hp.Current = hp.Max
	}
	g.world.Add(g.playerID, hp)
}

func (g *Game) tryPickup() {
	pos := g.playerPosition()
	items := g.world.Query(component.CTagItem, component.CPosition)
	for _, itemID := range items {
		ipos := g.world.Get(itemID, component.CPosition).(component.Position)
		if ipos.X == pos.X && ipos.Y == pos.Y {
			// Read the item data from CItemComp.
			itemComp := g.world.Get(itemID, component.CItem)
			if itemComp == nil {
				g.addMessage("Strange item — cannot pick up.")
				return
			}
			item := itemComp.(component.CItemComp).Item

			// Check backpack capacity.
			invComp := g.world.Get(g.playerID, component.CInventory)
			if invComp == nil {
				return
			}
			inv := invComp.(component.Inventory)
			if len(inv.Backpack) >= inv.Capacity {
				g.addMessage("Backpack full! Drop something first.")
				return
			}

			// Add to backpack and destroy floor entity.
			inv.Backpack = append(inv.Backpack, item)
			g.world.Add(g.playerID, inv)
			g.world.DestroyEntity(itemID)
			g.addMessage(fmt.Sprintf("You pick up %s. [i] to open inventory.", item.Name))
			return
		}
	}
	g.addMessage("Nothing to pick up here.")
}

// applyConsumable uses a consumable item from inventory (called by inventory screen).
func (g *Game) applyConsumable(item component.Item) {
	glyph := item.Glyph
	g.runLog.ItemsUsed[glyph]++
	switch glyph {
	case assets.GlyphHyperflask:
		g.restorePlayerHP(15)
		g.addMessage("The Hyperflask restores 15 HP.")

	case assets.GlyphPrismShard:
		system.ApplyEffect(g.world, g.playerID, component.ActiveEffect{
			Kind: component.EffectAttackBoost, Magnitude: 3, TurnsRemaining: 10,
		})
		g.addMessage("The Prism Shard boosts your ATK by 3 for 10 turns!")

	case assets.GlyphNullCloak:
		system.ApplyEffect(g.world, g.playerID, component.ActiveEffect{
			Kind: component.EffectInvisible, Magnitude: 1, TurnsRemaining: 12,
		})
		g.addMessage("The Null Cloak makes you invisible for 12 turns.")

	case assets.GlyphTesseract:
		g.teleportPlayer()
		g.addMessage("The Tesseract Cube warps you to a random location!")

	case assets.GlyphMemoryScroll:
		for y := 0; y < g.gmap.Height; y++ {
			for x := 0; x < g.gmap.Width; x++ {
				if g.gmap.At(x, y).Walkable {
					g.gmap.At(x, y).Explored = true
				}
			}
		}
		g.addMessage("The Memory Scroll reveals the entire floor.")

	case assets.GlyphSporeDraught:
		g.restorePlayerHP(20)
		g.addMessage("The Spore Draught mends your wounds with living mycelium. (+20 HP)")

	case assets.GlyphResonanceCoil:
		system.ApplyEffect(g.world, g.playerID, component.ActiveEffect{
			Kind: component.EffectAttackBoost, Magnitude: 5, TurnsRemaining: 12,
		})
		g.addMessage("The Resonance Coil harmonises with your strikes. (+5 ATK, 12 turns)")

	case assets.GlyphPrismaticWard:
		system.ApplyEffect(g.world, g.playerID, component.ActiveEffect{
			Kind: component.EffectDefenseBoost, Magnitude: 4, TurnsRemaining: 12,
		})
		g.addMessage("The Prismatic Ward refracts incoming harm. (+4 DEF, 12 turns)")

	case assets.GlyphVoidEssence:
		system.ApplyEffect(g.world, g.playerID, component.ActiveEffect{
			Kind: component.EffectInvisible, Magnitude: 1, TurnsRemaining: 20,
		})
		g.addMessage("The Void Essence erases you from local spacetime. (Invisible, 20 turns)")

	case assets.GlyphNanoSyringe:
		g.restorePlayerHP(30)
		g.addMessage("The Nano-Syringe floods your bloodstream with healing agents. (+30 HP)")

	case assets.GlyphResonanceBurst:
		system.ApplyEffect(g.world, g.playerID, component.ActiveEffect{
			Kind: component.EffectAttackBoost, Magnitude: 8, TurnsRemaining: 8,
		})
		system.ApplyEffect(g.world, g.playerID, component.ActiveEffect{
			Kind: component.EffectSelfBurn, Magnitude: 2, TurnsRemaining: 8,
		})
		g.addMessage("Resonance Burst! (+8 ATK for 8 turns, but -2 HP/turn burn)")

	case assets.GlyphPhaseRod:
		system.ApplyEffect(g.world, g.playerID, component.ActiveEffect{
			Kind: component.EffectDefenseBoost, Magnitude: 6, TurnsRemaining: 15,
		})
		g.addMessage("The Phase Rod envelops you in prismatic shielding. (+6 DEF, 15 turns)")

	case assets.GlyphApexCore:
		g.baseMaxHP += 3
		hp := g.world.Get(g.playerID, component.CHealth).(component.Health)
		hp.Max += 3
		hp.Current += 3
		g.world.Add(g.playerID, hp)
		g.addMessage("The Apex Core integrates into your biology. (+3 MaxHP permanently)")
	}
}

// checkInscription displays any wall-writing at the player's current position.
// equipBonuses returns the total ATK and DEF bonus from all equipped items.
func (g *Game) equipBonuses() (atk, def int) {
	c := g.world.Get(g.playerID, component.CInventory)
	if c == nil {
		return 0, 0
	}
	inv := c.(component.Inventory)
	atk = inv.MainHand.BonusATK + inv.OffHand.BonusATK +
		inv.Head.BonusATK + inv.Body.BonusATK + inv.Feet.BonusATK
	def = inv.MainHand.BonusDEF + inv.OffHand.BonusDEF +
		inv.Head.BonusDEF + inv.Body.BonusDEF + inv.Feet.BonusDEF
	return atk, def
}

// recalcPlayerMaxHP recalculates the player's MaxHP from baseMaxHP + equipment bonuses.
func (g *Game) recalcPlayerMaxHP() {
	invComp := g.world.Get(g.playerID, component.CInventory)
	if invComp == nil {
		return
	}
	inv := invComp.(component.Inventory)
	bonus := inv.Head.BonusMaxHP + inv.Body.BonusMaxHP + inv.Feet.BonusMaxHP +
		inv.MainHand.BonusMaxHP + inv.OffHand.BonusMaxHP
	hpComp := g.world.Get(g.playerID, component.CHealth)
	if hpComp == nil {
		return
	}
	hp := hpComp.(component.Health)
	hp.Max = g.baseMaxHP + bonus
	if hp.Current > hp.Max {
		hp.Current = hp.Max
	}
	g.world.Add(g.playerID, hp)
}

func (g *Game) checkInscription() {
	pos := g.playerPosition()
	for _, id := range g.world.Query(component.CInscription, component.CPosition) {
		ipos := g.world.Get(id, component.CPosition).(component.Position)
		if ipos.X == pos.X && ipos.Y == pos.Y {
			text := g.world.Get(id, component.CInscription).(component.Inscription).Text
			g.runLog.InscriptionsRead++
			g.addMessage("📝 " + text)
			return
		}
	}
}

func (g *Game) teleportPlayer() {
	rooms := g.gmap.Rooms
	if len(rooms) == 0 {
		return
	}
	room := rooms[g.rng.Intn(len(rooms))]
	x, y := room.Center()
	g.world.Add(g.playerID, component.Position{X: x, Y: y})
	system.UpdateFOV(g.world, g.gmap, g.playerID, g.fovRadius)
}

func (g *Game) checkPlayerDead() {
	hp := g.world.Get(g.playerID, component.CHealth)
	if hp == nil || hp.(component.Health).Current <= 0 {
		g.state = StateDead
	}
}

func (g *Game) checkVictory() {
	if g.floor != MaxFloors {
		return
	}
	bossGlyph := assets.BossGlyphs[g.floor]
	if bossGlyph == "" {
		return
	}
	for _, id := range g.world.Query(component.CRenderable) {
		rend := g.world.Get(id, component.CRenderable).(component.Renderable)
		if rend.Glyph == bossGlyph {
			return // still alive
		}
	}
	g.state = StateVictory
	g.addMessage("The Unmaker dissolves into prismatic light. The Spire's heart is yours!")
}

// confirmQuit draws a small overlay asking the player to confirm quitting.
// redraw is called first each iteration to paint the background.
// Returns true if the player confirms (Y), false otherwise (N / Escape).
func (g *Game) confirmQuit(redraw func()) bool {
	sw, sh := g.screen.Size()
	msg := "Quit? [Y]es / [N]o"
	style := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDefault).Bold(true)
	for {
		redraw()
		sw, sh = g.screen.Size()
		x := (sw - len(msg)) / 2
		y := sh / 2
		g.putText(x, y, msg, style)
		g.screen.Show()

		ev := g.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			g.screen.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
				return false
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'y', 'Y':
					return true
				case 'n', 'N':
					return false
				}
			}
		}
	}
}

func (g *Game) playerPosition() component.Position {
	c := g.world.Get(g.playerID, component.CPosition)
	if c == nil {
		return component.Position{}
	}
	return c.(component.Position)
}

func (g *Game) entityName(id ecs.EntityID) string {
	rend := g.world.Get(id, component.CRenderable)
	if rend == nil {
		return "creature"
	}
	return rend.(component.Renderable).Glyph
}

func (g *Game) addMessage(msg string) {
	g.messages = append(g.messages, msg)
	if len(g.messages) > 50 {
		g.messages = g.messages[len(g.messages)-50:]
	}
}

// putText writes a string to the screen at (x, y), one column per rune.
func (g *Game) putText(x, y int, s string, style tcell.Style) {
	for _, r := range s {
		g.screen.SetContent(x, y, r, nil, style)
		x++
	}
}

// showEndScreen renders the run summary and returns true if the player
// wants to try again, false to quit.
func (g *Game) showEndScreen() bool {
	won := g.state == StateVictory

	// Pre-compute kill breakdown sorted by count descending.
	type killEntry struct {
		glyph string
		count int
	}
	var kills []killEntry
	for gl, cnt := range g.runLog.EnemiesKilled {
		kills = append(kills, killEntry{gl, cnt})
	}
	sort.Slice(kills, func(i, j int) bool {
		if kills[i].count != kills[j].count {
			return kills[i].count > kills[j].count
		}
		return kills[i].glyph < kills[j].glyph
	})
	totalKills := 0
	for _, e := range kills {
		totalKills += e.count
	}

	totalItems := 0
	for _, c := range g.runLog.ItemsUsed {
		totalItems += c
	}

	floorName := ""
	if g.runLog.FloorsReached >= 1 && g.runLog.FloorsReached <= MaxFloors {
		floorName = fmt.Sprintf("Floor %d — %s",
			g.runLog.FloorsReached, assets.FloorNames[g.runLog.FloorsReached])
	}

	white := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	gold  := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	gray  := tcell.StyleDefault.Foreground(tcell.ColorGray)
	dim   := tcell.StyleDefault.Foreground(tcell.ColorLightYellow)
	green := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	red   := tcell.StyleDefault.Foreground(tcell.ColorRed)

	for {
		g.screen.Clear()
		sw, _ := g.screen.Size()

		sep := func(y int) {
			for x := 0; x < sw; x++ {
				g.screen.SetContent(x, y, '─', nil, gray)
			}
		}
		// label prints a left-aligned key at column 2 and value at column 22.
		label := func(y int, l, v string) {
			g.putText(2, y, l, dim)
			g.putText(22, y, v, white)
		}

		y := 1
		sep(y); y += 2

		// Title + outcome badge.
		if won {
			g.putText(2, y, "THE PRISMATIC HEART IS SILENT", gold)
			badge := "[VICTORY]"
			g.putText(sw-len(badge)-1, y, badge, green)
		} else {
			g.putText(2, y, "THE SPIRE CLAIMS YOU", gold)
			badge := "[DEFEAT]"
			g.putText(sw-len(badge)-1, y, badge, red)
		}
		y += 2

		// Core stats.
		label(y, "Class:", g.runLog.Class); y++
		label(y, "Floor Reached:", floorName); y++
		label(y, "Turns Survived:", fmt.Sprintf("%d", g.runLog.TurnsPlayed)); y += 2

		// Kill count + breakdown.
		label(y, "Enemies Slain:", fmt.Sprintf("%d", totalKills)); y++
		if len(kills) > 0 {
			breakdown := ""
			for _, e := range kills {
				breakdown += fmt.Sprintf("%s×%d  ", e.glyph, e.count)
			}
			// Trim to fit screen (rune-based, emoji count as 1 here but render wider).
			maxRunes := sw - 6
			runes := []rune(breakdown)
			if len(runes) > maxRunes {
				runes = runes[:maxRunes]
			}
			g.putText(4, y, string(runes), dim)
			y++
		}
		y++

		label(y, "Items Used:", fmt.Sprintf("%d", totalItems)); y++
		label(y, "Inscriptions Read:", fmt.Sprintf("%d", g.runLog.InscriptionsRead)); y += 2

		label(y, "Damage Dealt:", fmt.Sprintf("%d", g.runLog.DamageDealt)); y++
		label(y, "Damage Taken:", fmt.Sprintf("%d", g.runLog.DamageTaken)); y += 2

		// Outcome line.
		if won {
			g.putText(2, y, "The Unmaker is unmade. The Spire falls silent.", green)
		} else if g.runLog.CauseOfDeath == "poison" {
			label(y, "Killed By:", "poison")
		} else if g.runLog.CauseOfDeath != "" {
			label(y, "Killed By:", g.runLog.CauseOfDeath)
		}
		y += 2

		sep(y); y += 2

		g.putText(2, y, "[R] Try Again", green)
		g.putText(18, y, "[Q] Quit", red)

		g.screen.Show()

		ev := g.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			g.screen.Sync()
			continue // redraw on resize
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'r', 'R':
					return true
				case 'q', 'Q':
					return false
				}
			case tcell.KeyEscape:
				return false
			}
		}
	}
}
