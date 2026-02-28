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

// coopPlayer holds per-player state for a cooperative game session.
type coopPlayer struct {
	id                   ecs.EntityID
	screen               tcell.Screen
	renderer             *render.Renderer
	class                assets.ClassDef
	fovRadius            int
	baseMaxHP            int
	furnitureATK         int
	furnitureDEF         int
	furnitureThorns      int
	furnitureKillRestore bool
	specialCooldown      int // turns until z-ability can be used again
	// events receives all tcell events from the polling goroutine.
	events            chan tcell.Event
	alive             bool
	runLog            RunLog
	discoveredEnemies map[string]bool
}

// CoopGame is the shared game session for two players over SSH.
// A single ECS world, game map, and RNG are shared; each player has an
// independent tcell.Screen and per-player bonus state.
type CoopGame struct {
	world    *ecs.World
	gmap     *gamemap.GameMap
	floor    int
	rng      *rand.Rand
	state    GameState
	messages []string
	players  [2]*coopPlayer
}

// NewCoopGame creates a CoopGame backed by two already-initialized tcell screens.
func NewCoopGame(screens [2]tcell.Screen) *CoopGame {
	g := &CoopGame{
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
		state: StatePlaying,
	}
	for i, screen := range screens {
		g.players[i] = &coopPlayer{
			screen:            screen,
			events:            make(chan tcell.Event, 32),
			alive:             true,
			discoveredEnemies: make(map[string]bool),
			runLog: RunLog{
				EnemiesKilled: make(map[string]int),
				ItemsUsed:     make(map[string]int),
			},
		}
	}
	return g
}

// Run drives the cooperative game loop. Blocks until the game ends.
// Calls screen.Fini() on both screens before returning.
func (g *CoopGame) Run() {
	defer func() {
		for _, p := range g.players {
			p.screen.Fini()
		}
	}()

	// Class selection: run in parallel goroutines, one per player.
	// Each goroutine calls p.screen.PollEvent() directly (event goroutines not yet started).
	type classResult struct {
		idx int
		cls assets.ClassDef
	}
	results := make(chan classResult, 2)
	for i := range 2 {
		i, p := i, g.players[i]
		go func() {
			cls := coopClassSelect(p)
			results <- classResult{i, cls}
		}()
	}
	for range 2 {
		r := <-results
		g.players[r.idx].class = r.cls
		g.players[r.idx].fovRadius = r.cls.FOVRadius
		g.players[r.idx].baseMaxHP = r.cls.MaxHP
		g.players[r.idx].runLog.Class = r.cls.Name
	}

	// Start per-player event polling goroutines.
	for _, p := range g.players {
		p := p
		go func() {
			for {
				ev := p.screen.PollEvent()
				if ev == nil {
					return
				}
				p.events <- ev
			}
		}()
	}

	g.loadFloor(1)
	g.addMessage("Cooperative mode! Use hjklyubn or arrow keys to move. > to descend.")
	g.addMessage(fmt.Sprintf("P1: %s  P2: %s — take turns, then enemies act.",
		g.players[0].class.Name, g.players[1].class.Name))

	for g.state == StatePlaying {
		g.renderAll()

		prevFloor := g.floor

		// P1's turn.
		if g.players[0].alive {
			action := g.waitPlayerAction(g.players[0])
			if action == ActionQuit {
				g.state = StateDead
				break
			}
			turnUsed := g.processCoopAction(g.players[0], action)
			if g.state != StatePlaying {
				break
			}
			if g.floor != prevFloor {
				continue // floor changed — skip P2 and AI this round
			}
			if turnUsed {
				g.players[0].runLog.TurnsPlayed++
			}
		}

		if g.state != StatePlaying {
			break
		}
		g.renderAll()

		// P2's turn.
		if g.players[1].alive {
			action := g.waitPlayerAction(g.players[1])
			if action == ActionQuit {
				g.state = StateDead
				break
			}
			turnUsed := g.processCoopAction(g.players[1], action)
			if g.state != StatePlaying {
				break
			}
			if g.floor != prevFloor {
				continue
			}
			if turnUsed {
				g.players[1].runLog.TurnsPlayed++
			}
		}

		if g.state != StatePlaying {
			break
		}

		// Enemy AI + world tick after both players have acted.
		if g.players[0].alive || g.players[1].alive {
			g.tickWorld()
		}
	}

	// Stamp and save each player's run log.
	for _, p := range g.players {
		p.runLog.FloorsReached = g.floor
		p.runLog.Victory = g.state == StateVictory
		p.runLog.Timestamp = time.Now()
		if p.runLog.Victory {
			p.runLog.CauseOfDeath = ""
		}
		saveRunLog(p.runLog)
	}

	g.showCoopEndScreen()
}

// coopClassSelect blocks until the player selects a class on their screen.
// Calls p.screen.PollEvent() directly (before the event polling goroutine starts).
func coopClassSelect(p *coopPlayer) assets.ClassDef {
	selected := 0
	for {
		DrawClassSelectScreen(p.screen, selected)
		ev := p.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			p.screen.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyUp:
				selected = (selected - 1 + len(assets.Classes)) % len(assets.Classes)
			case tcell.KeyDown:
				selected = (selected + 1) % len(assets.Classes)
			case tcell.KeyEnter:
				return assets.Classes[selected]
			}
			switch ev.Rune() {
			case 'k', 'K':
				selected = (selected - 1 + len(assets.Classes)) % len(assets.Classes)
			case 'j', 'J':
				selected = (selected + 1) % len(assets.Classes)
			case '1', '2', '3', '4', '5', '6':
				idx := int(ev.Rune() - '1')
				if idx >= 0 && idx < len(assets.Classes) {
					return assets.Classes[idx]
				}
			}
		}
	}
}

// loadFloor generates and populates the given floor, creating player entities
// for all connected players. HP and inventory are preserved across transitions.
func (g *CoopGame) loadFloor(floor int) {
	// Save HP and inventory for live players before discarding the old world.
	type savedState struct {
		hp  int
		inv *component.Inventory
	}
	saved := [2]savedState{}
	if g.world != nil {
		for i, p := range g.players {
			if !p.alive || p.id == ecs.NilEntity {
				continue
			}
			if hpComp := g.world.Get(p.id, component.CHealth); hpComp != nil {
				saved[i].hp = hpComp.(component.Health).Current
			}
			if invComp := g.world.Get(p.id, component.CInventory); invComp != nil {
				inv := invComp.(component.Inventory)
				saved[i].inv = &inv
			}
		}
	}

	g.floor = floor
	g.world = ecs.NewWorld()

	cfg := levelConfig(floor, g.rng)
	gmap, px, py := generate.Generate(cfg)
	g.gmap = gmap

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
	for _, fs := range pop.Furniture {
		factory.NewFurniture(g.world, fs.Entry, fs.X, fs.Y)
	}

	// Player spawn colors: P1 yellow, P2 magenta so players can distinguish each other.
	playerColors := [2]tcell.Color{tcell.ColorYellow, tcell.ColorFuchsia}

	// Spawn positions: P1 at the map start, P2 one tile right (or at the same spot).
	spawnX := [2]int{px, px + 1}
	spawnY := [2]int{py, py}
	// If the +1 tile isn't walkable, put P2 on the same tile.
	if !gmap.InBounds(spawnX[1], spawnY[1]) || !gmap.IsWalkable(spawnX[1], spawnY[1]) {
		spawnX[1], spawnY[1] = px, py
	}

	for i, p := range g.players {
		if !p.alive {
			continue
		}
		p.id = factory.NewPlayer(g.world, spawnX[i], spawnY[i], p.class)

		// Override the default yellow with this player's colour.
		if rend := g.world.Get(p.id, component.CRenderable); rend != nil {
			r := rend.(component.Renderable)
			r.FGColor = playerColors[i]
			g.world.Add(p.id, r)
		}

		// Reapply persistent furniture combat bonuses.
		if p.furnitureATK != 0 || p.furnitureDEF != 0 {
			if cc := g.world.Get(p.id, component.CCombat); cc != nil {
				c := cc.(component.Combat)
				c.Attack += p.furnitureATK
				c.Defense += p.furnitureDEF
				g.world.Add(p.id, c)
			}
		}

		// Restore HP from previous floor (capped at max).
		if saved[i].hp > 0 && floor > 1 {
			hp := g.world.Get(p.id, component.CHealth).(component.Health)
			if saved[i].hp < hp.Max {
				hp.Current = saved[i].hp
			}
			g.world.Add(p.id, hp)
		}

		// Restore inventory from previous floor.
		if saved[i].inv != nil && floor > 1 {
			g.world.Add(p.id, *saved[i].inv)
			g.coopRecalcPlayerMaxHP(p)
		}

		// Floor 1 passives.
		if floor == 1 {
			for j, glyph := range p.class.StartItems {
				offsets := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
				ox, oy := offsets[j%len(offsets)][0], offsets[j%len(offsets)][1]
				ix, iy := spawnX[i]+ox, spawnY[i]+oy
				if !gmap.InBounds(ix, iy) || !gmap.IsWalkable(ix, iy) {
					ix, iy = spawnX[i], spawnY[i]
				}
				factory.NewItemByGlyph(g.world, glyph, ix, iy)
			}
		}

		// Reset ability cooldown on each floor entry for classes with AbilityFreeOnFloor.
		if p.class.AbilityFreeOnFloor {
			p.specialCooldown = 0
		}

		system.UpdateFOV(g.world, g.gmap, p.id, p.fovRadius)
		p.renderer = render.NewRenderer(p.screen, floor)
		p.renderer.CenterOn(spawnX[i], spawnY[i])
	}

	if floor == 1 {
		g.addMessage(fmt.Sprintf("You enter %s.", assets.FloorNames[floor]))
	} else {
		g.addMessage(fmt.Sprintf("You descend into %s (Floor %d).", assets.FloorNames[floor], floor))
	}
	if lore := assets.FloorLore[floor]; len(lore) > 0 {
		g.addMessage(lore[g.rng.Intn(len(lore))])
	}
}

// renderAll redraws the world and HUD on every connected player's screen,
// each centered on their own character.
func (g *CoopGame) renderAll() {
	for _, p := range g.players {
		if p.renderer == nil {
			continue
		}
		pos := g.coopPlayerPosition(p)
		p.renderer.CenterOn(pos.X, pos.Y)
		p.renderer.DrawFrame(g.world, g.gmap, p.id)
		equipATK, equipDEF := g.coopEquipBonuses(p)
		bonusATK := system.GetAttackBonus(g.world, p.id) + equipATK
		bonusDEF := system.GetDefenseBonus(g.world, p.id) + equipDEF
		p.renderer.DrawHUD(g.world, p.id, g.floor, p.class.Name, g.messages, bonusATK, bonusDEF, p.class.AbilityName, p.specialCooldown)
	}
}

// waitPlayerAction blocks until a meaningful action arrives on p.events.
// Resize events trigger a re-render and are otherwise consumed silently.
func (g *CoopGame) waitPlayerAction(p *coopPlayer) Action {
	for {
		ev, ok := <-p.events
		if !ok || ev == nil {
			return ActionQuit // disconnect
		}
		switch ev := ev.(type) {
		case *tcell.EventResize:
			p.screen.Sync()
			g.renderAll()
		case *tcell.EventKey:
			return keyToAction(ev)
		}
	}
}

// processCoopAction handles one player's action. It does NOT run enemy AI
// (that is deferred to tickWorld after both players have acted).
// Returns true if a game turn was consumed.
func (g *CoopGame) processCoopAction(p *coopPlayer, action Action) bool {
	// Stunned players lose their turn.
	if system.IsStunned(g.world, p.id) {
		g.addMessage(fmt.Sprintf("%s is stunned and cannot act!", p.class.Name))
		return true
	}

	switch action {
	case ActionWait:
		g.addMessage(fmt.Sprintf("%s waits.", p.class.Name))
		return true

	case ActionDescend:
		pos := g.coopPlayerPosition(p)
		tile := g.gmap.At(pos.X, pos.Y)
		if tile.Kind == gamemap.TileStairsDown {
			if g.floor >= MaxFloors {
				g.addMessage("There is nowhere further to descend.")
			} else {
				g.loadFloor(g.floor + 1)
			}
		} else {
			g.addMessage("There are no stairs down here.")
		}
		return false

	case ActionAscend:
		pos := g.coopPlayerPosition(p)
		tile := g.gmap.At(pos.X, pos.Y)
		if tile.Kind == gamemap.TileStairsUp && g.floor > 1 {
			g.loadFloor(g.floor - 1)
		} else {
			g.addMessage("There are no stairs up here.")
		}
		return false

	case ActionPickup:
		g.coopTryPickup(p)
		return true

	case ActionInventory:
		return g.coopRunInventory(p)

	case ActionSpecialAbility:
		if p.class.AbilityCooldown == 0 {
			g.addMessage(fmt.Sprintf("%s has no special ability.", p.class.Name))
		} else if p.specialCooldown > 0 {
			g.addMessage(fmt.Sprintf("%s: %s recharging (%d turns).", p.class.Name, p.class.AbilityName, p.specialCooldown))
		} else {
			g.useCoopSpecialAbility(p)
			p.specialCooldown = p.class.AbilityCooldown
			return true
		}
		return false

	default:
		dx, dy := actionToDelta(action)
		if dx == 0 && dy == 0 {
			return false
		}
		result, target := system.TryMove(g.world, g.gmap, p.id, dx, dy)
		switch result {
		case system.MoveOK:
			system.UpdateFOV(g.world, g.gmap, p.id, p.fovRadius)
			g.coopCheckInscription(p)
			return true

		case system.MoveInteract:
			g.coopInteractFurniture(p, target)
			return true

		case system.MoveAttack:
			// Don't attack teammates.
			for _, other := range g.players {
				if target == other.id {
					return false
				}
			}
			name := g.entityName(target)
			enemyPos := g.world.Get(target, component.CPosition).(component.Position)
			var lootDrops []component.LootEntry
			if lc := g.world.Get(target, component.CLoot); lc != nil {
				lootDrops = lc.(component.Loot).Drops
			}
			res := system.Attack(g.world, g.rng, p.id, target)
			p.runLog.DamageDealt += res.Damage
			if res.Killed {
				p.runLog.EnemiesKilled[name]++
				g.addMessage(fmt.Sprintf("%s kills the %s!", p.class.Name, name))
				if !p.discoveredEnemies[name] {
					p.discoveredEnemies[name] = true
					if lore, ok := assets.EnemyLore[name]; ok {
						g.addMessage(lore)
					}
				}
				for _, d := range lootDrops {
					if g.rng.Intn(100) < d.Chance {
						factory.NewItemByGlyph(g.world, d.Glyph, enemyPos.X, enemyPos.Y)
					}
				}
				if p.class.KillRestoreHP > 0 {
					g.coopRestorePlayerHP(p, p.class.KillRestoreHP)
				}
				if p.class.KillHealChance > 0 && g.rng.Intn(100) < p.class.KillHealChance {
					g.coopRestorePlayerHP(p, 2)
					g.addMessage(fmt.Sprintf("%s: Wild magic sparks! (+2 HP)", p.class.Name))
				}
				if p.furnitureKillRestore {
					g.coopRestorePlayerHP(p, 1)
				}
				g.checkCoopVictory()
			} else {
				g.addMessage(fmt.Sprintf("%s hits the %s for %d damage.", p.class.Name, name, res.Damage))
			}
			return true

		case system.MoveBlocked:
			pos := g.coopPlayerPosition(p)
			tx, ty := pos.X+dx, pos.Y+dy
			if g.gmap.InBounds(tx, ty) && g.gmap.At(tx, ty).Kind == gamemap.TileDoor {
				g.gmap.Set(tx, ty, gamemap.MakeFloor())
				system.UpdateFOV(g.world, g.gmap, p.id, p.fovRadius)
				g.addMessage(fmt.Sprintf("%s opens a door.", p.class.Name))
				return true
			}
		}
	}
	return false
}

// tickWorld applies poison/burn to all players, ticks effects, runs AI, and
// checks for player deaths and victory. Called once per round after both players act.
func (g *CoopGame) tickWorld() {
	// Poison and burn damage for each alive player.
	for _, p := range g.players {
		if p.alive {
			g.coopApplyPoisonDamage(p)
		}
	}
	system.TickEffects(g.world)

	// Per-player passive ticks: ability cooldown and regeneration.
	for _, p := range g.players {
		if !p.alive {
			continue
		}
		if p.specialCooldown > 0 {
			p.specialCooldown--
		}
		if p.class.PassiveRegen > 0 && p.runLog.TurnsPlayed > 0 && p.runLog.TurnsPlayed%p.class.PassiveRegen == 0 {
			g.coopRestorePlayerHP(p, 1)
		}
	}

	// Collect alive player IDs for multi-target AI.
	var pids []ecs.EntityID
	for _, p := range g.players {
		if p.alive && p.id != ecs.NilEntity {
			pids = append(pids, p.id)
		}
	}
	if len(pids) == 0 {
		return
	}

	hits := system.ProcessAI(g.world, g.gmap, pids, g.rng)

	// Attribute damage and apply thorns.
	// Use the combined thorns of both players (cooperative benefit).
	maxThorns := max(g.players[0].furnitureThorns, g.players[1].furnitureThorns)
	for _, h := range hits {
		if h.Damage > 0 {
			// Attribute damage directly via VictimID.
			for _, p := range g.players {
				if p.alive && p.id == h.VictimID {
					p.runLog.DamageTaken += h.Damage
					p.runLog.CauseOfDeath = h.EnemyGlyph
					break
				}
			}
			// Thorns: reflect damage back to the attacker.
			if maxThorns > 0 && h.AttackerID != ecs.NilEntity && g.world.Alive(h.AttackerID) {
				if hp := g.world.Get(h.AttackerID, component.CHealth); hp != nil {
					hpVal := hp.(component.Health)
					hpVal.Current -= maxThorns
					g.world.Add(h.AttackerID, hpVal)
				}
			}
		}
		g.handleCoopHitMessage(h)
	}

	// Check player deaths.
	for _, p := range g.players {
		if !p.alive {
			continue
		}
		hp := g.world.Get(p.id, component.CHealth)
		if hp == nil || hp.(component.Health).Current <= 0 {
			p.alive = false
			g.addMessage(fmt.Sprintf("%s has fallen!", p.class.Name))
		}
	}

	// All players dead → game over.
	if !g.players[0].alive && !g.players[1].alive {
		g.state = StateDead
	}

	g.checkCoopVictory()
}

func (g *CoopGame) handleCoopHitMessage(h system.EnemyHitResult) {
	switch h.SpecialApplied {
	case 1:
		g.addMessage(fmt.Sprintf("The %s poisons a player!", h.EnemyGlyph))
	case 2:
		g.addMessage(fmt.Sprintf("The %s weakens a player's attack!", h.EnemyGlyph))
	case 3:
		g.addMessage(fmt.Sprintf("The %s drains life force!", h.EnemyGlyph))
	case 4:
		g.addMessage(fmt.Sprintf("The %s stuns a player!", h.EnemyGlyph))
	case 5:
		g.addMessage(fmt.Sprintf("The %s shatters defenses!", h.EnemyGlyph))
	}
}


// ─── per-player helpers ────────────────────────────────────────────────────

func (g *CoopGame) coopPlayerPosition(p *coopPlayer) component.Position {
	c := g.world.Get(p.id, component.CPosition)
	if c == nil {
		return component.Position{}
	}
	return c.(component.Position)
}

func (g *CoopGame) coopRestorePlayerHP(p *coopPlayer, n int) {
	hpComp := g.world.Get(p.id, component.CHealth)
	if hpComp == nil {
		return
	}
	hp := hpComp.(component.Health)
	hp.Current += n
	if hp.Current > hp.Max {
		hp.Current = hp.Max
	}
	g.world.Add(p.id, hp)
}

func (g *CoopGame) coopApplyPoisonDamage(p *coopPlayer) {
	poisonDmg := system.GetPoisonDamage(g.world, p.id)
	burnDmg := system.GetSelfBurnDamage(g.world, p.id)
	total := poisonDmg + burnDmg
	if total <= 0 {
		return
	}
	hp := g.world.Get(p.id, component.CHealth)
	if hp == nil {
		return
	}
	h := hp.(component.Health)
	h.Current -= total
	g.world.Add(p.id, h)
	p.runLog.DamageTaken += total
	if poisonDmg > 0 {
		p.runLog.CauseOfDeath = "poison"
		g.addMessage(fmt.Sprintf("Poison burns through %s! (%d damage)", p.class.Name, poisonDmg))
	}
	if burnDmg > 0 {
		if p.runLog.CauseOfDeath == "" {
			p.runLog.CauseOfDeath = "self-burn"
		}
		g.addMessage(fmt.Sprintf("The resonance burns %s! (%d damage)", p.class.Name, burnDmg))
	}
}

func (g *CoopGame) coopEquipBonuses(p *coopPlayer) (atk, def int) {
	c := g.world.Get(p.id, component.CInventory)
	if c == nil {
		return 0, 0
	}
	inv := c.(component.Inventory)
	atk = inv.MainHand.BonusATK + inv.OffHand.BonusATK + inv.Head.BonusATK + inv.Body.BonusATK + inv.Feet.BonusATK
	def = inv.MainHand.BonusDEF + inv.OffHand.BonusDEF + inv.Head.BonusDEF + inv.Body.BonusDEF + inv.Feet.BonusDEF
	return atk, def
}

func (g *CoopGame) coopRecalcPlayerMaxHP(p *coopPlayer) {
	invComp := g.world.Get(p.id, component.CInventory)
	if invComp == nil {
		return
	}
	inv := invComp.(component.Inventory)
	bonus := inv.Head.BonusMaxHP + inv.Body.BonusMaxHP + inv.Feet.BonusMaxHP +
		inv.MainHand.BonusMaxHP + inv.OffHand.BonusMaxHP
	hpComp := g.world.Get(p.id, component.CHealth)
	if hpComp == nil {
		return
	}
	hp := hpComp.(component.Health)
	hp.Max = p.baseMaxHP + bonus
	if hp.Current > hp.Max {
		hp.Current = hp.Max
	}
	g.world.Add(p.id, hp)
}

func (g *CoopGame) coopCheckInscription(p *coopPlayer) {
	pos := g.coopPlayerPosition(p)
	for _, id := range g.world.Query(component.CInscription, component.CPosition) {
		ipos := g.world.Get(id, component.CPosition).(component.Position)
		if ipos.X == pos.X && ipos.Y == pos.Y {
			text := g.world.Get(id, component.CInscription).(component.Inscription).Text
			p.runLog.InscriptionsRead++
			g.addMessage("📝 " + text)
			return
		}
	}
}

// useCoopSpecialAbility fires the class active ability for a coop player.
func (g *CoopGame) useCoopSpecialAbility(p *coopPlayer) {
	switch p.class.ID {
	case "arcanist":
		g.coopTeleportPlayer(p)
		g.addMessage(fmt.Sprintf("%s: Dimensional Rift — reappears elsewhere!", p.class.Name))

	case "revenant":
		hpComp := g.world.Get(p.id, component.CHealth)
		if hpComp == nil {
			return
		}
		hp := hpComp.(component.Health)
		if hp.Current <= 5 {
			g.addMessage(fmt.Sprintf("%s: Too wounded to bargain with death!", p.class.Name))
			p.specialCooldown = 0 // refund cooldown
			return
		}
		hp.Current -= 5
		g.world.Add(p.id, hp)
		system.ApplyEffect(g.world, p.id, component.ActiveEffect{
			Kind: component.EffectAttackBoost, Magnitude: 6, TurnsRemaining: 8,
		})
		g.addMessage(fmt.Sprintf("%s: Death's Bargain! (-5 HP, +6 ATK for 8 turns)", p.class.Name))

	case "construct":
		system.ApplyEffect(g.world, p.id, component.ActiveEffect{
			Kind: component.EffectAttackBoost, Magnitude: 6, TurnsRemaining: 6,
		})
		system.ApplyEffect(g.world, p.id, component.ActiveEffect{
			Kind: component.EffectSelfBurn, Magnitude: 2, TurnsRemaining: 6,
		})
		g.addMessage(fmt.Sprintf("%s: Overclock! (+6 ATK for 6 turns, -2 HP/turn burn)", p.class.Name))

	case "dancer":
		system.ApplyEffect(g.world, p.id, component.ActiveEffect{
			Kind: component.EffectInvisible, Magnitude: 1, TurnsRemaining: 8,
		})
		g.addMessage(fmt.Sprintf("%s: Vanish! Invisible for 8 turns.", p.class.Name))

	case "oracle":
		for y := 0; y < g.gmap.Height; y++ {
			for x := 0; x < g.gmap.Width; x++ {
				if g.gmap.At(x, y).Walkable {
					g.gmap.At(x, y).Explored = true
				}
			}
		}
		g.addMessage(fmt.Sprintf("%s: Farsight — entire floor revealed!", p.class.Name))

	case "symbiont":
		g.coopRestorePlayerHP(p, 10)
		system.ApplyEffect(g.world, p.id, component.ActiveEffect{
			Kind: component.EffectAttackBoost, Magnitude: 4, TurnsRemaining: 6,
		})
		g.addMessage(fmt.Sprintf("%s: Parasite Surge! (+10 HP, +4 ATK for 6 turns)", p.class.Name))
	}
}

func (g *CoopGame) coopTeleportPlayer(p *coopPlayer) {
	rooms := g.gmap.Rooms
	if len(rooms) == 0 {
		return
	}
	room := rooms[g.rng.Intn(len(rooms))]
	x, y := room.Center()
	g.world.Add(p.id, component.Position{X: x, Y: y})
	system.UpdateFOV(g.world, g.gmap, p.id, p.fovRadius)
}

func (g *CoopGame) coopTryPickup(p *coopPlayer) {
	pos := g.coopPlayerPosition(p)
	for _, itemID := range g.world.Query(component.CTagItem, component.CPosition) {
		ipos := g.world.Get(itemID, component.CPosition).(component.Position)
		if ipos.X != pos.X || ipos.Y != pos.Y {
			continue
		}
		itemComp := g.world.Get(itemID, component.CItem)
		if itemComp == nil {
			g.addMessage("Strange item — cannot pick up.")
			return
		}
		item := itemComp.(component.CItemComp).Item
		invComp := g.world.Get(p.id, component.CInventory)
		if invComp == nil {
			return
		}
		inv := invComp.(component.Inventory)
		if len(inv.Backpack) >= inv.Capacity {
			g.addMessage("Backpack full! Drop something first.")
			return
		}
		inv.Backpack = append(inv.Backpack, item)
		g.world.Add(p.id, inv)
		g.world.DestroyEntity(itemID)
		g.addMessage(fmt.Sprintf("%s picks up %s.", p.class.Name, item.Name))
		return
	}
	g.addMessage("Nothing to pick up here.")
}

func (g *CoopGame) coopInteractFurniture(p *coopPlayer, id ecs.EntityID) {
	fc := g.world.Get(id, component.CFurniture)
	if fc == nil {
		return
	}
	f := fc.(component.Furniture)
	g.addMessage(fmt.Sprintf("%s %s: %s", f.Glyph, f.Name, f.Description))
	if f.Used {
		return
	}
	hasBonus := f.BonusATK != 0 || f.BonusDEF != 0 || f.BonusMaxHP != 0 ||
		f.HealHP != 0 || f.PassiveKind != 0
	if !hasBonus {
		return
	}
	if f.BonusATK != 0 {
		p.furnitureATK += f.BonusATK
		if cc := g.world.Get(p.id, component.CCombat); cc != nil {
			c := cc.(component.Combat)
			c.Attack += f.BonusATK
			g.world.Add(p.id, c)
		}
		g.addMessage(fmt.Sprintf("%s: Permanent ATK +%d!", p.class.Name, f.BonusATK))
	}
	if f.BonusDEF != 0 {
		p.furnitureDEF += f.BonusDEF
		if cc := g.world.Get(p.id, component.CCombat); cc != nil {
			c := cc.(component.Combat)
			c.Defense += f.BonusDEF
			g.world.Add(p.id, c)
		}
		g.addMessage(fmt.Sprintf("%s: Permanent DEF +%d!", p.class.Name, f.BonusDEF))
	}
	if f.BonusMaxHP != 0 {
		p.baseMaxHP += f.BonusMaxHP
		if hp := g.world.Get(p.id, component.CHealth); hp != nil {
			h := hp.(component.Health)
			h.Max += f.BonusMaxHP
			h.Current += f.BonusMaxHP
			if h.Current > h.Max {
				h.Current = h.Max
			}
			g.world.Add(p.id, h)
		}
		g.addMessage(fmt.Sprintf("%s: Permanent MaxHP +%d!", p.class.Name, f.BonusMaxHP))
	}
	if f.HealHP != 0 {
		g.coopRestorePlayerHP(p, f.HealHP)
		g.addMessage(fmt.Sprintf("Restored %d HP to %s!", f.HealHP, p.class.Name))
	}
	switch f.PassiveKind {
	case component.PassiveKeenEye:
		p.fovRadius++
		system.UpdateFOV(g.world, g.gmap, p.id, p.fovRadius)
		g.addMessage(fmt.Sprintf("%s: Vision expands permanently.", p.class.Name))
	case component.PassiveKillRestore:
		p.furnitureKillRestore = true
		g.addMessage(fmt.Sprintf("%s: Life force quickens on each kill.", p.class.Name))
	case component.PassiveThorns:
		p.furnitureThorns++
		g.addMessage(fmt.Sprintf("%s: Sharp crystals form beneath the skin.", p.class.Name))
	}
	f.Used = true
	g.world.Add(id, f)
}

// coopApplyConsumable applies a consumable item's effect to the given player.
func (g *CoopGame) coopApplyConsumable(p *coopPlayer, item component.Item) {
	glyph := item.Glyph
	p.runLog.ItemsUsed[glyph]++
	switch glyph {
	case assets.GlyphHyperflask:
		g.coopRestorePlayerHP(p, 15)
		g.addMessage("The Hyperflask restores 15 HP.")
	case assets.GlyphPrismShard:
		system.ApplyEffect(g.world, p.id, component.ActiveEffect{
			Kind: component.EffectAttackBoost, Magnitude: 3, TurnsRemaining: 10,
		})
		g.addMessage("The Prism Shard boosts ATK by 3 for 10 turns!")
	case assets.GlyphNullCloak:
		system.ApplyEffect(g.world, p.id, component.ActiveEffect{
			Kind: component.EffectInvisible, Magnitude: 1, TurnsRemaining: 12,
		})
		g.addMessage("The Null Cloak makes you invisible for 12 turns.")
	case assets.GlyphTesseract:
		g.coopTeleportPlayer(p)
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
		g.coopRestorePlayerHP(p, 20)
		g.addMessage("The Spore Draught mends wounds. (+20 HP)")
	case assets.GlyphResonanceCoil:
		system.ApplyEffect(g.world, p.id, component.ActiveEffect{
			Kind: component.EffectAttackBoost, Magnitude: 5, TurnsRemaining: 12,
		})
		g.addMessage("The Resonance Coil harmonises. (+5 ATK, 12 turns)")
	case assets.GlyphPrismaticWard:
		system.ApplyEffect(g.world, p.id, component.ActiveEffect{
			Kind: component.EffectDefenseBoost, Magnitude: 4, TurnsRemaining: 12,
		})
		g.addMessage("The Prismatic Ward refracts harm. (+4 DEF, 12 turns)")
	case assets.GlyphVoidEssence:
		system.ApplyEffect(g.world, p.id, component.ActiveEffect{
			Kind: component.EffectInvisible, Magnitude: 1, TurnsRemaining: 20,
		})
		g.addMessage("The Void Essence erases you from spacetime. (Invisible, 20 turns)")
	case assets.GlyphNanoSyringe:
		g.coopRestorePlayerHP(p, 30)
		g.addMessage("The Nano-Syringe floods your bloodstream. (+30 HP)")
	case assets.GlyphResonanceBurst:
		system.ApplyEffect(g.world, p.id, component.ActiveEffect{
			Kind: component.EffectAttackBoost, Magnitude: 8, TurnsRemaining: 8,
		})
		system.ApplyEffect(g.world, p.id, component.ActiveEffect{
			Kind: component.EffectSelfBurn, Magnitude: 2, TurnsRemaining: 8,
		})
		g.addMessage("Resonance Burst! (+8 ATK for 8 turns, -2 HP/turn burn)")
	case assets.GlyphPhaseRod:
		system.ApplyEffect(g.world, p.id, component.ActiveEffect{
			Kind: component.EffectDefenseBoost, Magnitude: 6, TurnsRemaining: 15,
		})
		g.addMessage("The Phase Rod envelops you. (+6 DEF, 15 turns)")
	case assets.GlyphApexCore:
		p.baseMaxHP += 3
		hp := g.world.Get(p.id, component.CHealth).(component.Health)
		hp.Max += 3
		hp.Current += 3
		g.world.Add(p.id, hp)
		g.addMessage("The Apex Core integrates into your biology. (+3 MaxHP permanently)")
	}
}

// ─── cooperative inventory screen ──────────────────────────────────────────

// coopRunInventory opens a blocking inventory UI on the player's own screen.
// Events are read from p.events (fed by the polling goroutine).
// Returns true if a game turn was consumed (consumable used).
func (g *CoopGame) coopRunInventory(p *coopPlayer) bool {
	invComp := g.world.Get(p.id, component.CInventory)
	if invComp == nil {
		return false
	}
	inv := invComp.(component.Inventory)
	panel := 0
	cursor := 0
	statusMsg := ""
	turnUsed := false

	clampCursor := func() {
		if panel == 0 {
			if cursor < 0 {
				cursor = 0
			}
			if cursor >= len(inv.Backpack) {
				cursor = len(inv.Backpack) - 1
			}
			if cursor < 0 {
				cursor = 0
			}
		} else {
			if cursor < 0 {
				cursor = 0
			}
			if cursor > 4 {
				cursor = 4
			}
		}
	}

	for {
		clampCursor()
		g.coopDrawInventoryScreen(p.screen, inv, panel, cursor, statusMsg)

		ev, ok := <-p.events
		if !ok || ev == nil {
			g.world.Add(p.id, inv)
			g.coopRecalcPlayerMaxHP(p)
			return turnUsed
		}
		switch ev := ev.(type) {
		case *tcell.EventResize:
			p.screen.Sync()
		case *tcell.EventKey:
			statusMsg = ""
			switch ev.Key() {
			case tcell.KeyEscape:
				g.world.Add(p.id, inv)
				g.coopRecalcPlayerMaxHP(p)
				return turnUsed
			case tcell.KeyTab:
				panel = 1 - panel
				cursor = 0
			case tcell.KeyUp:
				cursor--
			case tcell.KeyDown:
				cursor++
			case tcell.KeyEnter:
				statusMsg = g.coopInvEquipOrUnequip(p, &inv, panel, cursor)
			default:
				switch ev.Rune() {
				case 'k', 'K':
					cursor--
				case 'j', 'J':
					cursor++
				case '\t':
					panel = 1 - panel
					cursor = 0
				case 'e', 'E':
					statusMsg = g.coopInvEquipOrUnequip(p, &inv, panel, cursor)
				case 'u', 'U':
					msg, used := g.coopInvUseConsumable(p, &inv, panel, cursor)
					statusMsg = msg
					if used {
						turnUsed = true
						g.world.Add(p.id, inv)
						g.coopRecalcPlayerMaxHP(p)
						return turnUsed
					}
				case 'd', 'D':
					statusMsg = g.coopInvDrop(p, &inv, panel, &cursor)
				case 'i', 'I', 'q', 'Q':
					g.world.Add(p.id, inv)
					g.coopRecalcPlayerMaxHP(p)
					return turnUsed
				default:
					if ev.Rune() >= '1' && ev.Rune() <= '9' {
						idx := int(ev.Rune()-'0') - 1
						if idx < len(inv.Backpack) {
							panel = 0
							cursor = idx
						}
					}
				}
			}
		}
	}
}

func (g *CoopGame) coopInvEquipOrUnequip(_ *coopPlayer, inv *component.Inventory, panel, cursor int) string {
	if panel == 0 {
		if cursor < 0 || cursor >= len(inv.Backpack) {
			return "Nothing selected."
		}
		item := inv.Backpack[cursor]
		if item.IsConsumable {
			return "Press [u] to use consumables."
		}
		return coopInvEquip(inv, cursor)
	}
	return coopInvUnequip(inv, cursor)
}

func coopInvEquip(inv *component.Inventory, cursor int) string {
	item := inv.Backpack[cursor]
	switch item.Slot {
	case component.SlotHead:
		old := inv.Head
		inv.Backpack = removeAt(inv.Backpack, cursor)
		inv.Head = item
		if !old.IsEmpty() {
			inv.Backpack = append(inv.Backpack, old)
		}
		return fmt.Sprintf("Equipped %s.", item.Name)
	case component.SlotBody:
		old := inv.Body
		inv.Backpack = removeAt(inv.Backpack, cursor)
		inv.Body = item
		if !old.IsEmpty() {
			inv.Backpack = append(inv.Backpack, old)
		}
		return fmt.Sprintf("Equipped %s.", item.Name)
	case component.SlotFeet:
		old := inv.Feet
		inv.Backpack = removeAt(inv.Backpack, cursor)
		inv.Feet = item
		if !old.IsEmpty() {
			inv.Backpack = append(inv.Backpack, old)
		}
		return fmt.Sprintf("Equipped %s.", item.Name)
	case component.SlotOneHand:
		old := inv.MainHand
		inv.Backpack = removeAt(inv.Backpack, cursor)
		inv.MainHand = item
		if !old.IsEmpty() {
			inv.Backpack = append(inv.Backpack, old)
		}
		return fmt.Sprintf("Equipped %s.", item.Name)
	case component.SlotTwoHand:
		extra := 0
		if !inv.MainHand.IsEmpty() {
			extra++
		}
		if !inv.OffHand.IsEmpty() {
			extra++
		}
		if len(inv.Backpack)-1+extra > inv.Capacity {
			return "Not enough backpack space to swap."
		}
		inv.Backpack = removeAt(inv.Backpack, cursor)
		if !inv.OffHand.IsEmpty() {
			inv.Backpack = append(inv.Backpack, inv.OffHand)
			inv.OffHand = component.Item{}
		}
		if !inv.MainHand.IsEmpty() {
			inv.Backpack = append(inv.Backpack, inv.MainHand)
		}
		inv.MainHand = item
		return fmt.Sprintf("Equipped %s (two-handed).", item.Name)
	case component.SlotOffHand:
		if inv.MainHand.Slot == component.SlotTwoHand {
			return "Two-handed weapon occupies the off-hand slot."
		}
		old := inv.OffHand
		inv.Backpack = removeAt(inv.Backpack, cursor)
		inv.OffHand = item
		if !old.IsEmpty() {
			inv.Backpack = append(inv.Backpack, old)
		}
		return fmt.Sprintf("Equipped %s.", item.Name)
	}
	return "Cannot equip that."
}

func coopInvUnequip(inv *component.Inventory, cursor int) string {
	if len(inv.Backpack) >= inv.Capacity {
		return "Backpack full — drop something first."
	}
	var item component.Item
	switch cursor {
	case 0:
		if inv.Head.IsEmpty() {
			return "Nothing equipped in HEAD slot."
		}
		item = inv.Head
		inv.Head = component.Item{}
	case 1:
		if inv.Body.IsEmpty() {
			return "Nothing equipped in BODY slot."
		}
		item = inv.Body
		inv.Body = component.Item{}
	case 2:
		if inv.Feet.IsEmpty() {
			return "Nothing equipped in FEET slot."
		}
		item = inv.Feet
		inv.Feet = component.Item{}
	case 3:
		if inv.MainHand.IsEmpty() {
			return "Nothing equipped in WEAP slot."
		}
		item = inv.MainHand
		inv.MainHand = component.Item{}
	case 4:
		if inv.OffHand.IsEmpty() {
			return "Nothing equipped in OFHND slot."
		}
		item = inv.OffHand
		inv.OffHand = component.Item{}
	default:
		return "Invalid slot."
	}
	inv.Backpack = append(inv.Backpack, item)
	return fmt.Sprintf("Unequipped %s.", item.Name)
}

func (g *CoopGame) coopInvUseConsumable(p *coopPlayer, inv *component.Inventory, panel, cursor int) (string, bool) {
	if panel != 0 {
		return "Select a backpack item to use.", false
	}
	if cursor < 0 || cursor >= len(inv.Backpack) {
		return "Nothing selected.", false
	}
	item := inv.Backpack[cursor]
	if !item.IsConsumable {
		return "Equipment must be equipped, not used.", false
	}
	inv.Backpack = removeAt(inv.Backpack, cursor)
	g.coopApplyConsumable(p, item)
	return fmt.Sprintf("Used %s.", item.Name), true
}

func (g *CoopGame) coopInvDrop(p *coopPlayer, inv *component.Inventory, panel int, cursor *int) string {
	pos := g.coopPlayerPosition(p)
	var item component.Item
	if panel == 0 {
		if *cursor < 0 || *cursor >= len(inv.Backpack) {
			return "Nothing selected."
		}
		item = inv.Backpack[*cursor]
		inv.Backpack = removeAt(inv.Backpack, *cursor)
		if *cursor >= len(inv.Backpack) && *cursor > 0 {
			(*cursor)--
		}
	} else {
		switch *cursor {
		case 0:
			if inv.Head.IsEmpty() {
				return "Nothing equipped here."
			}
			item = inv.Head
			inv.Head = component.Item{}
		case 1:
			if inv.Body.IsEmpty() {
				return "Nothing equipped here."
			}
			item = inv.Body
			inv.Body = component.Item{}
		case 2:
			if inv.Feet.IsEmpty() {
				return "Nothing equipped here."
			}
			item = inv.Feet
			inv.Feet = component.Item{}
		case 3:
			if inv.MainHand.IsEmpty() {
				return "Nothing equipped here."
			}
			item = inv.MainHand
			inv.MainHand = component.Item{}
		case 4:
			if inv.OffHand.IsEmpty() {
				return "Nothing equipped here."
			}
			item = inv.OffHand
			inv.OffHand = component.Item{}
		default:
			return "Invalid slot."
		}
	}
	factory.DropItem(g.world, item, pos.X, pos.Y)
	return fmt.Sprintf("Dropped %s.", item.Name)
}

// coopDrawInventoryScreen renders the inventory UI onto the given screen.
func (g *CoopGame) coopDrawInventoryScreen(screen tcell.Screen, inv component.Inventory, panel, cursor int, statusMsg string) {
	screen.Clear()
	sw, _ := screen.Size()
	mid := sw / 2
	if mid < 30 {
		mid = 30
	}

	white := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	gray := tcell.StyleDefault.Foreground(tcell.ColorGray)
	yellow := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	cyan := tcell.StyleDefault.Foreground(tcell.ColorAqua)
	green := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	highlight := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorAqua)
	dim := tcell.StyleDefault.Foreground(tcell.ColorGray)

	put := func(x, y int, s string, style tcell.Style) {
		for _, r := range s {
			screen.SetContent(x, y, r, nil, style)
			x++
		}
	}

	title := fmt.Sprintf("INVENTORY  [Backpack %d/%d]", len(inv.Backpack), inv.Capacity)
	put(0, 0, title, yellow)
	hints := "[j/k] Move  [Tab] Switch  [e] Equip  [u] Use  [d] Drop  [Esc] Close"
	if len(hints) < sw {
		put(sw-len([]rune(hints)), 0, hints, dim)
	}
	for x := range sw {
		screen.SetContent(x, 1, '─', nil, gray)
	}
	put(0, 2, "── EQUIPPED ──────────────────", white)
	put(mid, 2, "── BACKPACK ─────────────────", white)
	for y := 2; y <= 12; y++ {
		screen.SetContent(mid-1, y, '│', nil, gray)
	}

	equipSlots := []struct {
		label string
		item  component.Item
	}{
		{"HEAD ", inv.Head}, {"BODY ", inv.Body}, {"FEET ", inv.Feet},
		{"WEAP ", inv.MainHand}, {"OFHND", inv.OffHand},
	}
	for i, slot := range equipSlots {
		row := 3 + i
		sel := panel == 1 && cursor == i
		style := white
		pfx := "  "
		if sel {
			style = highlight
			pfx = "► "
		}
		itemStr := "--"
		if !slot.item.IsEmpty() {
			itemStr = slot.item.Glyph + " " + slot.item.Name + formatBonuses(slot.item)
		}
		put(0, row, fmt.Sprintf("%s%s %s", pfx, slot.label, itemStr), style)
	}

	atkB := inv.Head.BonusATK + inv.Body.BonusATK + inv.Feet.BonusATK + inv.MainHand.BonusATK + inv.OffHand.BonusATK
	defB := inv.Head.BonusDEF + inv.Body.BonusDEF + inv.Feet.BonusDEF + inv.MainHand.BonusDEF + inv.OffHand.BonusDEF
	hpB := inv.Head.BonusMaxHP + inv.Body.BonusMaxHP + inv.Feet.BonusMaxHP + inv.MainHand.BonusMaxHP + inv.OffHand.BonusMaxHP
	put(0, 8, fmt.Sprintf("  Equip bonus: ATK%+d DEF%+d HP%+d", atkB, defB, hpB), cyan)

	for i, item := range inv.Backpack {
		row := 3 + i
		if row > 10 {
			break
		}
		sel := panel == 0 && cursor == i
		style := white
		pfx := "  "
		if sel {
			style = highlight
			pfx = "► "
		}
		tag := ""
		if item.IsConsumable {
			tag = " [use]"
		}
		put(mid, row, fmt.Sprintf("%s[%d] %s %s%s%s", pfx, i+1, item.Glyph, item.Name, formatBonuses(item), tag), style)
	}
	if len(inv.Backpack) == 0 {
		put(mid, 3, "  (empty)", dim)
	}

	for x := range sw {
		screen.SetContent(x, 11, '─', nil, gray)
	}

	var selItem component.Item
	selEmpty := true
	if panel == 0 && cursor < len(inv.Backpack) {
		selItem = inv.Backpack[cursor]
		selEmpty = false
	} else if panel == 1 {
		switch cursor {
		case 0:
			selItem = inv.Head
		case 1:
			selItem = inv.Body
		case 2:
			selItem = inv.Feet
		case 3:
			selItem = inv.MainHand
		case 4:
			selItem = inv.OffHand
		}
		selEmpty = selItem.IsEmpty()
	}
	if !selEmpty {
		put(0, 12, fmt.Sprintf("%s — %s  ATK%+d DEF%+d MaxHP%+d",
			selItem.Name, slotLabel(selItem.Slot), selItem.BonusATK, selItem.BonusDEF, selItem.BonusMaxHP), white)
	}
	if statusMsg != "" {
		put(0, 13, statusMsg, green)
	}
	screen.Show()
}

// ─── victory / end ──────────────────────────────────────────────────────────

func (g *CoopGame) checkCoopVictory() {
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
	g.addMessage("The Unmaker dissolves! The Spire's heart is yours — together!")
}

// showCoopEndScreen displays the end-of-game summary on all player screens.
func (g *CoopGame) showCoopEndScreen() {
	won := g.state == StateVictory

	white := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	gold := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	gray := tcell.StyleDefault.Foreground(tcell.ColorGray)
	dim := tcell.StyleDefault.Foreground(tcell.ColorLightYellow)
	green := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	red := tcell.StyleDefault.Foreground(tcell.ColorRed)

	// Build combined kill list.
	combined := make(map[string]int)
	for _, p := range g.players {
		for gl, cnt := range p.runLog.EnemiesKilled {
			combined[gl] += cnt
		}
	}
	type killEntry struct {
		glyph string
		count int
	}
	var kills []killEntry
	for gl, cnt := range combined {
		kills = append(kills, killEntry{gl, cnt})
	}
	sort.Slice(kills, func(i, j int) bool {
		if kills[i].count != kills[j].count {
			return kills[i].count > kills[j].count
		}
		return kills[i].glyph < kills[j].glyph
	})

	floorName := ""
	if g.floor >= 1 && g.floor <= MaxFloors {
		floorName = fmt.Sprintf("Floor %d — %s", g.floor, assets.FloorNames[g.floor])
	}

	totalDmgDealt := g.players[0].runLog.DamageDealt + g.players[1].runLog.DamageDealt
	totalDmgTaken := g.players[0].runLog.DamageTaken + g.players[1].runLog.DamageTaken

	draw := func(screen tcell.Screen) {
		screen.Clear()
		sw, _ := screen.Size()
		put := func(x, y int, s string, style tcell.Style) {
			for _, r := range s {
				screen.SetContent(x, y, r, nil, style)
				x++
			}
		}
		sep := func(y int) {
			for x := 0; x < sw; x++ {
				screen.SetContent(x, y, '─', nil, gray)
			}
		}
		label := func(y int, l, v string) {
			put(2, y, l, dim)
			put(22, y, v, white)
		}

		y := 1
		sep(y)
		y += 2

		if won {
			put(2, y, "THE PRISMATIC HEART IS SILENT", gold)
			badge := "[VICTORY]"
			put(sw-len(badge)-1, y, badge, green)
		} else {
			put(2, y, "THE SPIRE CLAIMS YOU BOTH", gold)
			badge := "[DEFEAT]"
			put(sw-len(badge)-1, y, badge, red)
		}
		y += 2

		for i, p := range g.players {
			label(y, fmt.Sprintf("P%d Class:", i+1), p.runLog.Class)
			y++
		}
		label(y, "Floor Reached:", floorName)
		y++

		totalTurns := g.players[0].runLog.TurnsPlayed + g.players[1].runLog.TurnsPlayed
		label(y, "Total Turns:", fmt.Sprintf("%d", totalTurns))
		y += 2

		totalKills := 0
		for _, e := range kills {
			totalKills += e.count
		}
		label(y, "Enemies Slain:", fmt.Sprintf("%d", totalKills))
		y++
		if len(kills) > 0 {
			breakdown := ""
			for _, e := range kills {
				breakdown += fmt.Sprintf("%s×%d  ", e.glyph, e.count)
			}
			maxRunes := sw - 6
			runes := []rune(breakdown)
			if len(runes) > maxRunes {
				runes = runes[:maxRunes]
			}
			put(4, y, string(runes), dim)
			y++
		}
		y++

		label(y, "Damage Dealt:", fmt.Sprintf("%d", totalDmgDealt))
		y++
		label(y, "Damage Taken:", fmt.Sprintf("%d", totalDmgTaken))
		y += 2

		if won {
			put(2, y, "The Unmaker is unmade. Together, you silenced the Spire.", green)
		}
		y += 2

		sep(y)
		y += 2

		put(2, y, "[Q] Quit", red)
		screen.Show()
	}

	// Draw on all screens and wait for either player to press Q.
	done := make(chan struct{})
	for _, p := range g.players {
		p := p
		draw(p.screen)
		go func() {
			for {
				ev, ok := <-p.events
				if !ok || ev == nil {
					close(done)
					return
				}
				if ev, ok := ev.(*tcell.EventKey); ok {
					r := ev.Rune()
					if r == 'q' || r == 'Q' || ev.Key() == tcell.KeyEscape {
						select {
						case <-done:
						default:
							close(done)
						}
						return
					}
				}
				if _, ok := ev.(*tcell.EventResize); ok {
					p.screen.Sync()
					draw(p.screen)
				}
			}
		}()
	}
	<-done
}

// addMessage appends a message to the shared log (capped at 50).
func (g *CoopGame) addMessage(msg string) {
	g.messages = append(g.messages, msg)
	if len(g.messages) > 50 {
		g.messages = g.messages[len(g.messages)-50:]
	}
}

func (g *CoopGame) entityName(id ecs.EntityID) string {
	rend := g.world.Get(id, component.CRenderable)
	if rend == nil {
		return "creature"
	}
	return rend.(component.Renderable).Glyph
}
