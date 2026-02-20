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
	Class            string
	FloorsReached    int
	TurnsPlayed      int
	EnemiesKilled    map[string]int // glyph → kill count
	ItemsUsed        map[string]int // glyph → use count
	InscriptionsRead int
	DamageDealt      int
	DamageTaken      int
	CauseOfDeath     string // last thing that hurt the player ("poison" or enemy glyph)
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
	g.discoveredEnemies = make(map[string]bool)
	g.runLog = RunLog{
		EnemiesKilled: make(map[string]int),
		ItemsUsed:     make(map[string]int),
	}
}

// loadFloor generates and populates the given floor.
// On transitions (floor > 1) the player's current HP is preserved.
func (g *Game) loadFloor(floor int) {
	// Preserve HP across floor transitions.
	savedHP := -1
	if g.world != nil && g.playerID != ecs.NilEntity {
		if hpComp := g.world.Get(g.playerID, component.CHealth); hpComp != nil {
			savedHP = hpComp.(component.Health).Current
		}
	}

	g.floor = floor
	if floor > g.runLog.FloorsReached {
		g.runLog.FloorsReached = floor
	}
	g.world = ecs.NewWorld()

	cfg := levelConfig(floor, g.rng)
	gmap, px, py := generate.Generate(cfg)
	g.gmap = gmap

	// Populate enemies, items, and inscriptions.
	pop := generate.Populate(gmap, cfg)
	for _, es := range pop.Enemies {
		factory.NewEnemy(g.world, es.Entry, es.X, es.Y)
	}
	for _, is := range pop.Items {
		factory.NewItem(g.world, is.Entry, is.X, is.Y)
	}
	for _, ins := range pop.Inscriptions {
		factory.NewInscription(g.world, ins.Text, ins.X, ins.Y)
	}

	// Create player using the selected class definition.
	g.playerID = factory.NewPlayer(g.world, px, py, g.selectedClass)

	// Restore HP from previous floor (capped at class max).
	if savedHP > 0 && floor > 1 {
		hp := g.world.Get(g.playerID, component.CHealth).(component.Health)
		if savedHP < hp.Max {
			hp.Current = savedHP
		}
		g.world.Add(g.playerID, hp)
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
			g.renderer.DrawHUD(g.world, g.playerID, g.floor, g.selectedClass.Name, g.messages)

			ev := g.screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventResize:
				g.screen.Sync()
				continue
			case *tcell.EventKey:
				action := keyToAction(ev)
				if action == ActionQuit {
					return
				}
				g.processAction(action)
			}
		}

		if !g.showEndScreen() {
			return
		}
	}
}

// processAction handles one player action and optionally advances enemy AI.
func (g *Game) processAction(action Action) {
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
				// Capture name/glyph BEFORE Attack() which may destroy the entity.
				name := g.entityName(target)
				glyph := name
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
			switch h.SpecialApplied {
			case 1:
				g.addMessage(fmt.Sprintf("The %s poisons you!", h.EnemyGlyph))
			case 2:
				g.addMessage(fmt.Sprintf("The %s weakens your attack!", h.EnemyGlyph))
			case 3:
				g.addMessage(fmt.Sprintf("The %s drains your life force! (+%d HP to enemy)", h.EnemyGlyph, h.DrainedAmount))
			}
		}
		g.checkPlayerDead()
	}
}

func (g *Game) applyPoisonDamage() {
	dmg := system.GetPoisonDamage(g.world, g.playerID)
	if dmg <= 0 {
		return
	}
	hp := g.world.Get(g.playerID, component.CHealth)
	if hp == nil {
		return
	}
	h := hp.(component.Health)
	h.Current -= dmg
	g.world.Add(g.playerID, h)
	g.runLog.DamageTaken += dmg
	g.runLog.CauseOfDeath = "poison"
	g.addMessage(fmt.Sprintf("Poison burns through you! (%d damage)", dmg))
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
			rend := g.world.Get(itemID, component.CRenderable)
			name := "item"
			if rend != nil {
				name = rend.(component.Renderable).Glyph
			}
			g.applyItem(itemID)
			g.world.DestroyEntity(itemID)
			g.addMessage(fmt.Sprintf("You pick up %s.", name))
			return
		}
	}
	g.addMessage("Nothing to pick up here.")
}

// applyItem immediately uses an item (all items are consumable).
func (g *Game) applyItem(itemID ecs.EntityID) {
	rend := g.world.Get(itemID, component.CRenderable)
	if rend == nil {
		return
	}
	glyph := rend.(component.Renderable).Glyph
	g.runLog.ItemsUsed[glyph]++
	switch glyph {
	case assets.GlyphHyperflask:
		restore := 15
		g.restorePlayerHP(restore)
		g.addMessage(fmt.Sprintf("The Hyperflask restores %d HP.", restore))

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
		restore := 20
		g.restorePlayerHP(restore)
		g.addMessage(fmt.Sprintf("The Spore Draught mends your wounds with living mycelium. (+%d HP)", restore))

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
	}
}

// checkInscription displays any wall-writing at the player's current position.
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
