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

// Game is the top-level orchestrator.
type Game struct {
	screen        tcell.Screen
	renderer      *render.Renderer
	world         *ecs.World
	gmap          *gamemap.GameMap
	playerID      ecs.EntityID
	rng           *rand.Rand
	floor         int
	state         GameState
	messages      []string
	selectedClass assets.ClassDef
	fovRadius     int
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
		floor:  1,
	}
	return g, nil
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
	g.world = ecs.NewWorld()

	cfg := levelConfig(floor, g.rng)
	gmap, px, py := generate.Generate(cfg)
	g.gmap = gmap

	// Populate enemies and items.
	pop := generate.Populate(gmap, cfg)
	for _, es := range pop.Enemies {
		factory.NewEnemy(g.world, es.Entry, es.X, es.Y)
	}
	for _, is := range pop.Items {
		factory.NewItem(g.world, is.Entry, is.X, is.Y)
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
			// Spread items in different directions so they don't stack.
			offsets := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
			ox, oy := offsets[i%len(offsets)][0], offsets[i%len(offsets)][1]
			ix, iy := px+ox, py+oy
			if !gmap.InBounds(ix, iy) || !gmap.IsWalkable(ix, iy) {
				ix, iy = px, py // fall back to player tile
			}
			factory.NewItemByGlyph(g.world, glyph, ix, iy)
		}
	}

	// Update FOV using the class-specific radius.
	system.UpdateFOV(g.world, g.gmap, g.playerID, g.fovRadius)

	// Create/update renderer.
	g.renderer = render.NewRenderer(g.screen, floor)
	g.renderer.CenterOn(px, py)

	if floor == 1 {
		g.addMessage(fmt.Sprintf("You enter the %s as a %s.", assets.FloorNames[floor], g.selectedClass.Name))
	} else {
		g.addMessage(fmt.Sprintf("You descend into the %s (Floor %d).", assets.FloorNames[floor], floor))
	}
}

// Run is the main game loop.
func (g *Game) Run() {
	defer g.screen.Fini()

	// Show class selection before loading any floor.
	if !g.runClassSelect() {
		return
	}

	g.loadFloor(1)
	g.addMessage("Use hjklyubn or arrow keys to move. > to descend.")

	for g.state != StateDead && g.state != StateVictory {
		// Render.
		playerPos := g.playerPosition()
		g.renderer.CenterOn(playerPos.X, playerPos.Y)
		g.renderer.DrawFrame(g.world, g.gmap, g.playerID)
		g.renderer.DrawHUD(g.world, g.playerID, g.floor, g.selectedClass.Name, g.messages)

		// Wait for input.
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

	// Show end screen.
	g.showEndScreen()
}

// processAction handles one player action and optionally runs enemy AI.
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
		// Movement.
		dx, dy := actionToDelta(action)
		if dx != 0 || dy != 0 {
			result, target := system.TryMove(g.world, g.gmap, g.playerID, dx, dy)
			switch result {
			case system.MoveOK:
				turnUsed = true
				system.UpdateFOV(g.world, g.gmap, g.playerID, g.fovRadius)
			case system.MoveAttack:
				res := system.Attack(g.world, g.rng, g.playerID, target)
				name := g.entityName(target)
				if res.Killed {
					g.addMessage(fmt.Sprintf("You kill the %s!", name))
					// Void Revenant passive: restore HP on kill.
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
		system.TickEffects(g.world)
		system.ProcessAI(g.world, g.gmap, g.playerID, g.rng)
		g.checkPlayerDead()
	}
}

// restorePlayerHP adds n HP to the player, capped at max.
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
	switch glyph {
	case assets.GlyphHyperflask:
		restore := 10
		g.restorePlayerHP(restore)
		g.addMessage(fmt.Sprintf("The Hyperflask restores %d HP.", restore))

	case assets.GlyphPrismShard:
		system.ApplyEffect(g.world, g.playerID, component.ActiveEffect{
			Kind: component.EffectAttackBoost, Magnitude: 3, TurnsRemaining: 5,
		})
		g.addMessage("The Prism Shard boosts your ATK by 3 for 5 turns!")

	case assets.GlyphNullCloak:
		system.ApplyEffect(g.world, g.playerID, component.ActiveEffect{
			Kind: component.EffectInvisible, Magnitude: 1, TurnsRemaining: 8,
		})
		g.addMessage("The Null Cloak makes you invisible for 8 turns.")

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
	if g.floor == MaxFloors {
		for _, id := range g.world.Query(component.CRenderable) {
			rend := g.world.Get(id, component.CRenderable).(component.Renderable)
			if rend.Glyph == assets.GlyphApexWarden {
				return // still alive
			}
		}
		g.state = StateVictory
		g.addMessage("The Apex Warden falls! The Spire's power is yours!")
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

func (g *Game) showEndScreen() {
	g.screen.Clear()
	msg := "You have died. The Spire claims you."
	if g.state == StateVictory {
		msg = "You have reached the Apex Engine. The Spire is yours!"
	}
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	for i, ch := range msg {
		g.screen.SetContent(i, 5, ch, nil, style)
	}
	prompt := "Press any key to exit."
	for i, ch := range prompt {
		g.screen.SetContent(i, 7, ch, nil, style)
	}
	g.screen.Show()
	g.screen.PollEvent()
}
