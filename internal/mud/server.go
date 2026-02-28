// Package mud implements a tick-based MUD server for emoji-roguelike.
// N players connect over SSH; each gets their own session. A single ticker
// goroutine advances the shared ECS world every TickInterval, consuming one
// queued action per player. Rendering happens in each session's own goroutine
// triggered by the ticker.
package mud

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/factory"
	"emoji-roguelike/internal/gamemap"
	"emoji-roguelike/internal/render"
	"emoji-roguelike/internal/system"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

// TickInterval is the wall-clock period between world ticks.
const TickInterval = 500 * time.Millisecond

// DeathTicks is the number of ticks a dead player waits before respawning.
const DeathTicks = 4

// EnemyRespawnDelay is the number of ticks after a floor is cleared before
// a new enemy wave is spawned (~30 seconds at 500 ms/tick).
const EnemyRespawnDelay = 60

// Server manages all floors and sessions for the MUD.
type Server struct {
	mu       sync.Mutex
	floors   map[int]*Floor
	sessions []*Session
	nextID   int
	rng      *rand.Rand
}

// NextSessionID returns a unique session ID and an assigned player color.
// Safe to call concurrently.
func (s *Server) NextSessionID() (int, tcell.Color) {
	s.mu.Lock()
	id := s.nextID
	s.nextID++
	color := playerColors[id%len(playerColors)]
	s.mu.Unlock()
	return id, color
}

// NewServer creates a Server and pre-generates floor 0 (the city of Emberveil).
func NewServer(rng *rand.Rand) *Server {
	s := &Server{
		floors: make(map[int]*Floor),
		rng:    rng,
	}
	s.floors[0] = newCityFloor(rand.New(rand.NewSource(rng.Int63())))
	return s
}

// Run starts the ticker loop. Blocks until the process exits.
func (s *Server) Run() {
	ticker := time.NewTicker(TickInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.tick()
	}
}

// AddSession registers a new session and spawns the player entity on floor 0 (the city).
// Must be called after the session's Class is set.
// The caller must NOT hold s.mu.
func (s *Server) AddSession(sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions = append(s.sessions, sess)
	s.spawnPlayerLocked(sess, 0)
	globalMessage(s.sessions, fmt.Sprintf("🌟 %s has arrived in Emberveil!", sess.Name))
}

// RemoveSession deregisters a session and removes the player entity from the world.
func (s *Server) RemoveSession(sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove entity from its floor.
	if floor, ok := s.floors[sess.FloorNum]; ok && sess.PlayerID != ecs.NilEntity {
		floor.World.DestroyEntity(sess.PlayerID)
	}

	// Remove from session list.
	for i, other := range s.sessions {
		if other == sess {
			s.sessions = append(s.sessions[:i], s.sessions[i+1:]...)
			break
		}
	}
	globalMessage(s.sessions, fmt.Sprintf("👋 %s has left the dungeon.", sess.Name))
}

// SignalRender sends a non-blocking render signal to all sessions.
// Called after tick() completes so sessions can redraw.
func (s *Server) signalRender() {
	for _, sess := range s.sessions {
		select {
		case sess.RenderCh <- struct{}{}:
		default:
		}
	}
}

// ─── Tick ────────────────────────────────────────────────────────────────────

func (s *Server) tick() {
	s.mu.Lock()

	// 1. Process one pending action per live player.
	for _, sess := range s.sessions {
		if sess.DeathCountdown > 0 {
			sess.DeathCountdown--
			if sess.DeathCountdown == 0 {
				s.respawnLocked(sess)
			}
			continue
		}
		action := sess.TakeAction()
		if action != ActionNone {
			s.processActionLocked(sess, action)
		}
	}

	// 2. Tick each active floor (effects, AI, passive regen, death checks).
	for _, floor := range s.floors {
		s.tickFloorLocked(floor)
	}

	s.mu.Unlock()

	// 3. Signal all sessions to render (outside the lock so slow SSH writes
	// don't block the next tick from starting).
	s.signalRender()
}

// tickFloorLocked advances AI and effects for one floor.
// Caller must hold s.mu.
func (s *Server) tickFloorLocked(floor *Floor) {
	// Safe zones (e.g. the city) skip combat and AI but still tick player cooldowns.
	if floor.SafeZone {
		for _, sess := range s.sessions {
			if sess.FloorNum != floor.Num || sess.DeathCountdown != 0 {
				continue
			}
			if sess.SpecialCooldown > 0 {
				sess.SpecialCooldown--
			}
			sess.TurnCount++
			if sess.Class.PassiveRegen > 0 && sess.TurnCount%sess.Class.PassiveRegen == 0 {
				restoreHP(floor.World, sess.PlayerID, 1)
			}
			sess.RunLog.TurnsPlayed++
		}
		return
	}

	// Collect live player IDs on this floor.
	var playerIDs []ecs.EntityID
	for _, sess := range s.sessions {
		if sess.FloorNum == floor.Num && sess.DeathCountdown == 0 && sess.PlayerID != ecs.NilEntity {
			playerIDs = append(playerIDs, sess.PlayerID)
		}
	}

	// Apply poison/burn to all players on this floor.
	for _, sess := range s.sessions {
		if sess.FloorNum == floor.Num && sess.DeathCountdown == 0 {
			s.applyDoTLocked(floor, sess)
		}
	}

	// Tick effects (reduces all duration counters).
	system.TickEffects(floor.World)

	// Per-player: ability cooldown and passive regen.
	for _, sess := range s.sessions {
		if sess.FloorNum != floor.Num || sess.DeathCountdown != 0 {
			continue
		}
		if sess.SpecialCooldown > 0 {
			sess.SpecialCooldown--
		}
		sess.TurnCount++
		if sess.Class.PassiveRegen > 0 && sess.TurnCount%sess.Class.PassiveRegen == 0 {
			restoreHP(floor.World, sess.PlayerID, 1)
		}
		sess.RunLog.TurnsPlayed++
	}

	// Run AI (no-op if no players).
	if len(playerIDs) == 0 {
		return
	}

	hits := system.ProcessAI(floor.World, floor.GMap, playerIDs, floor.Rng)

	// Process hits: attribute damage, apply thorns, generate messages.
	for _, h := range hits {
		if h.Damage > 0 {
			// Attribute damage directly via VictimID.
			if sess := s.sessionByPlayerID(h.VictimID); sess != nil {
				sess.RunLog.DamageTaken += h.Damage
				sess.RunLog.CauseOfDeath = h.EnemyGlyph
			}
			// Thorns: reflect damage to the attacker.
			if h.AttackerID != ecs.NilEntity && floor.World.Alive(h.AttackerID) {
				maxThorns := 0
				for _, sess := range s.sessions {
					if sess.FloorNum == floor.Num && sess.FurnitureThorns > maxThorns {
						maxThorns = sess.FurnitureThorns
					}
				}
				if maxThorns > 0 {
					if hp := floor.World.Get(h.AttackerID, component.CHealth); hp != nil {
						hpVal := hp.(component.Health)
						hpVal.Current -= maxThorns
						floor.World.Add(h.AttackerID, hpVal)
					}
				}
			}
		}
		// Send hit messages to all players on the floor.
		if msg := hitMessage(h, s.victimName(h.VictimID)); msg != "" {
			floorMessage(s.sessions, floor.Num, msg)
		}
	}

	// Check for player deaths.
	for _, sess := range s.sessions {
		if sess.FloorNum != floor.Num || sess.DeathCountdown != 0 {
			continue
		}
		hp := floor.World.Get(sess.PlayerID, component.CHealth)
		if hp == nil || hp.(component.Health).Current <= 0 {
			floorMessage(s.sessions, floor.Num, fmt.Sprintf("💀 %s has fallen!", sess.Name))
			sess.RunLog.Timestamp = time.Now()
			saveRunLog(sess.RunLog)
			sess.DeathCountdown = DeathTicks
			// Entity stays in world while countdown runs so others can see the
			// corpse position; it's cleaned up in respawnLocked.
		}
	}

	// Enemy respawn: when the floor is cleared and players are present,
	// start a countdown; spawn a new wave when it expires.
	enemyCount := len(floor.World.Query(component.CAI))
	if len(playerIDs) > 0 && enemyCount == 0 {
		if floor.RespawnCooldown < 0 {
			// Floor just cleared — start the countdown.
			floor.RespawnCooldown = EnemyRespawnDelay
		} else if floor.RespawnCooldown > 0 {
			floor.RespawnCooldown--
		} else { // == 0: time to spawn
			s.respawnEnemiesLocked(floor)
			floor.RespawnCooldown = -1
		}
	} else if enemyCount > 0 {
		floor.RespawnCooldown = -1 // reset when enemies are alive
	}
}

// hitMessage returns the floor-visible message for an enemy special attack.
func hitMessage(h system.EnemyHitResult, victimName string) string {
	switch h.SpecialApplied {
	case 1:
		return fmt.Sprintf("The %s poisons %s!", h.EnemyGlyph, victimName)
	case 2:
		return fmt.Sprintf("The %s weakens %s's attack!", h.EnemyGlyph, victimName)
	case 3:
		return fmt.Sprintf("The %s drains %s's life force!", h.EnemyGlyph, victimName)
	case 4:
		return fmt.Sprintf("The %s stuns %s!", h.EnemyGlyph, victimName)
	case 5:
		return fmt.Sprintf("The %s shatters %s's defenses!", h.EnemyGlyph, victimName)
	}
	return ""
}

// victimName returns the display name of the player with the given entity ID,
// or "a player" if not found.
func (s *Server) victimName(id ecs.EntityID) string {
	for _, sess := range s.sessions {
		if sess.PlayerID == id {
			return sess.Name
		}
	}
	return "a player"
}

// ─── Action Processing ───────────────────────────────────────────────────────

// processActionLocked handles one player action.
// Caller must hold s.mu.
func (s *Server) processActionLocked(sess *Session, action Action) {
	floor, ok := s.floors[sess.FloorNum]
	if !ok {
		return
	}

	// Stunned players lose their turn.
	if system.IsStunned(floor.World, sess.PlayerID) {
		sess.AddMessage("You are stunned and cannot act!")
		return
	}

	switch action {
	case ActionWait:
		sess.AddMessage("You wait.")

	case ActionDescend, ActionUseStairs:
		posComp := floor.World.Get(sess.PlayerID, component.CPosition)
		if posComp == nil {
			return
		}
		pos := posComp.(component.Position)
		tile := floor.GMap.At(pos.X, pos.Y)
		switch {
		case tile.Kind == gamemap.TileStairsDown && sess.FloorNum < MaxFloors:
			s.transitionFloorLocked(sess, sess.FloorNum+1)
			return
		case tile.Kind == gamemap.TileStairsDown && sess.FloorNum >= MaxFloors:
			sess.AddMessage("There is nowhere further to descend.")
		case tile.Kind == gamemap.TileStairsUp && sess.FloorNum > 0:
			s.transitionFloorLocked(sess, sess.FloorNum-1)
			return
		case action == ActionDescend:
			sess.AddMessage("There are no stairs down here.")
		default:
			sess.AddMessage("There are no stairs here.")
		}

	case ActionAscend:
		posComp := floor.World.Get(sess.PlayerID, component.CPosition)
		if posComp == nil {
			return
		}
		pos := posComp.(component.Position)
		tile := floor.GMap.At(pos.X, pos.Y)
		if tile.Kind == gamemap.TileStairsUp && sess.FloorNum > 0 {
			s.transitionFloorLocked(sess, sess.FloorNum-1)
		} else {
			sess.AddMessage("There are no stairs up here.")
		}

	case ActionPickup:
		s.tryPickupLocked(floor, sess)

	case ActionSpecialAbility:
		if sess.Class.AbilityCooldown == 0 {
			sess.AddMessage("No special ability.")
		} else if sess.SpecialCooldown > 0 {
			sess.AddMessage(fmt.Sprintf("%s recharging (%d turns).", sess.Class.AbilityName, sess.SpecialCooldown))
		} else {
			s.useSpecialAbilityLocked(floor, sess)
			sess.SpecialCooldown = sess.Class.AbilityCooldown
		}

	default:
		dx, dy := actionToDelta(action)
		if dx == 0 && dy == 0 {
			return
		}
		result, target := system.TryMove(floor.World, floor.GMap, sess.PlayerID, dx, dy)
		switch result {
		case system.MoveOK:
			system.UpdateFOV(floor.World, floor.GMap, sess.PlayerID, sess.FovRadius)
			sess.SnapshotFOV(floor.GMap)
			s.checkInscriptionLocked(floor, sess)

		case system.MoveInteract:
			if npc := floor.World.Get(target, component.CNPC); npc != nil {
				s.interactNPCLocked(floor, sess, target, npc.(component.NPC))
			} else {
				s.interactFurnitureLocked(floor, sess, target)
			}

		case system.MoveAttack:
			// Don't attack other players.
			if s.isPlayerEntity(target) {
				return
			}
			// No combat in safe zones.
			if floor.SafeZone {
				sess.AddMessage("Emberveil is a place of peace. Violence is forbidden here.")
				return
			}
			name := entityGlyph(floor.World, target)
			enemyPos := floor.World.Get(target, component.CPosition).(component.Position)
			var lootDrops []component.LootEntry
			if lc := floor.World.Get(target, component.CLoot); lc != nil {
				lootDrops = lc.(component.Loot).Drops
			}
			res := system.Attack(floor.World, floor.Rng, sess.PlayerID, target)
			sess.RunLog.DamageDealt += res.Damage
			if res.Killed {
				sess.RunLog.EnemiesKilled[name]++
				gold := floor.Rng.Intn(4) + 1
				sess.Gold += gold
				sess.RunLog.GoldEarned += gold
				floorMessage(s.sessions, floor.Num, fmt.Sprintf("%s kills the %s! (+%d💰)", sess.Name, name, gold))
				if !sess.DiscoveredEnemies[name] {
					sess.DiscoveredEnemies[name] = true
					if lore, ok := assets.EnemyLore[name]; ok {
						sess.AddMessage(lore)
					}
				}
				for _, d := range lootDrops {
					if floor.Rng.Intn(100) < d.Chance {
						factory.NewItemByGlyph(floor.World, d.Glyph, enemyPos.X, enemyPos.Y)
						sess.AddMessage(fmt.Sprintf("The %s drops something!", name))
					}
				}
				if sess.Class.KillRestoreHP > 0 {
					restoreHP(floor.World, sess.PlayerID, sess.Class.KillRestoreHP)
					sess.AddMessage(fmt.Sprintf("The kill feeds you. (+%d HP)", sess.Class.KillRestoreHP))
				}
				if sess.Class.KillHealChance > 0 && floor.Rng.Intn(100) < sess.Class.KillHealChance {
					restoreHP(floor.World, sess.PlayerID, 2)
					sess.AddMessage("Wild magic sparks! (+2 HP)")
				}
				if sess.FurnitureKR {
					restoreHP(floor.World, sess.PlayerID, 1)
				}
				s.checkVictoryLocked(floor, sess)
			} else {
				sess.AddMessage(fmt.Sprintf("You hit the %s for %d damage.", name, res.Damage))
			}

		case system.MoveBlocked:
			posComp := floor.World.Get(sess.PlayerID, component.CPosition)
			if posComp == nil {
				return
			}
			pos := posComp.(component.Position)
			tx, ty := pos.X+dx, pos.Y+dy
			if floor.GMap.InBounds(tx, ty) && floor.GMap.At(tx, ty).Kind == gamemap.TileDoor {
				floor.GMap.Set(tx, ty, gamemap.MakeFloor())
				system.UpdateFOV(floor.World, floor.GMap, sess.PlayerID, sess.FovRadius)
				sess.SnapshotFOV(floor.GMap)
				sess.AddMessage("You open the door.")
			}
		}
	}
}

// ─── Floor Transitions ───────────────────────────────────────────────────────

// transitionFloorLocked moves a session to a different floor, generating the
// target floor lazily if needed.
// Caller must hold s.mu.
func (s *Server) transitionFloorLocked(sess *Session, targetFloor int) {
	oldFloor, hasOld := s.floors[sess.FloorNum]

	// Save HP and inventory before destroying old entity.
	savedHP := -1
	var savedInv *component.Inventory
	if hasOld && sess.PlayerID != ecs.NilEntity {
		if hp := oldFloor.World.Get(sess.PlayerID, component.CHealth); hp != nil {
			savedHP = hp.(component.Health).Current
		}
		if inv := oldFloor.World.Get(sess.PlayerID, component.CInventory); inv != nil {
			v := inv.(component.Inventory)
			savedInv = &v
		}
		oldFloor.World.DestroyEntity(sess.PlayerID)
	}

	// Get or create the target floor.
	floor, ok := s.floors[targetFloor]
	if !ok {
		floor = newFloor(targetFloor, rand.New(rand.NewSource(s.rng.Int63())))
		s.floors[targetFloor] = floor
	}

	// Direction-aware spawn: descending → near stairs up; ascending → near stairs down.
	fromFloor := sess.FloorNum
	sess.FloorNum = targetFloor
	if targetFloor > sess.RunLog.FloorsReached {
		sess.RunLog.FloorsReached = targetFloor
	}
	var spawnX, spawnY int
	if targetFloor > fromFloor {
		sx, sy := floor.StairsUpX, floor.StairsUpY
		if sx < 0 {
			sx, sy = floor.SpawnX, floor.SpawnY
		}
		spawnX, spawnY = findFreeSpawn(floor, s.sessions, sx, sy)
	} else {
		spawnX, spawnY = findFreeSpawn(floor, s.sessions, floor.StairsDownX, floor.StairsDownY)
	}
	sess.PlayerID = factory.NewPlayer(floor.World, spawnX, spawnY, sess.Class)

	// Apply player color.
	if rend := floor.World.Get(sess.PlayerID, component.CRenderable); rend != nil {
		r := rend.(component.Renderable)
		r.FGColor = sess.Color
		floor.World.Add(sess.PlayerID, r)
	}

	// Reapply furniture combat bonuses.
	if sess.FurnitureATK != 0 || sess.FurnitureDEF != 0 {
		if cc := floor.World.Get(sess.PlayerID, component.CCombat); cc != nil {
			c := cc.(component.Combat)
			c.Attack += sess.FurnitureATK
			c.Defense += sess.FurnitureDEF
			floor.World.Add(sess.PlayerID, c)
		}
	}

	// Restore HP (capped at max).
	if savedHP > 0 {
		hp := floor.World.Get(sess.PlayerID, component.CHealth).(component.Health)
		if savedHP < hp.Max {
			hp.Current = savedHP
		}
		floor.World.Add(sess.PlayerID, hp)
	}

	// Restore inventory.
	if savedInv != nil {
		floor.World.Add(sess.PlayerID, *savedInv)
		recalcMaxHP(floor.World, sess)
	}

	// AbilityFreeOnFloor: reset cooldown on each floor entry.
	if sess.Class.AbilityFreeOnFloor {
		sess.SpecialCooldown = 0
	}

	system.UpdateFOV(floor.World, floor.GMap, sess.PlayerID, sess.FovRadius)
	sess.SnapshotFOV(floor.GMap)
	sess.Renderer = render.NewRenderer(sess.Screen, targetFloor)
	sess.Renderer.CenterOn(spawnX, spawnY)

	if targetFloor == 0 {
		sess.AddMessage(fmt.Sprintf("%s returns to Emberveil.", sess.Name))
	} else if targetFloor > fromFloor {
		sess.AddMessage(fmt.Sprintf("%s descends into %s (Floor %d).", sess.Name, assets.FloorNames[targetFloor], targetFloor))
	} else {
		sess.AddMessage(fmt.Sprintf("%s ascends to %s (Floor %d).", sess.Name, assets.FloorNames[targetFloor], targetFloor))
	}
	if lore := assets.FloorLore[targetFloor]; len(lore) > 0 {
		sess.AddMessage(lore[floor.Rng.Intn(len(lore))])
	}
}

// spawnPlayerLocked creates a player entity on the given floor for a new session.
// Caller must hold s.mu.
func (s *Server) spawnPlayerLocked(sess *Session, floorNum int) {
	floor, ok := s.floors[floorNum]
	if !ok {
		floor = newFloor(floorNum, rand.New(rand.NewSource(s.rng.Int63())))
		s.floors[floorNum] = floor
	}

	sess.FloorNum = floorNum
	sx, sy := findFreeSpawn(floor, s.sessions, floor.SpawnX, floor.SpawnY)
	sess.PlayerID = factory.NewPlayer(floor.World, sx, sy, sess.Class)

	// Override glyph color to this player's assigned color.
	if rend := floor.World.Get(sess.PlayerID, component.CRenderable); rend != nil {
		r := rend.(component.Renderable)
		r.FGColor = sess.Color
		floor.World.Add(sess.PlayerID, r)
	}

	// Apply any pre-existing furniture bonuses (e.g. on respawn).
	if sess.FurnitureATK != 0 || sess.FurnitureDEF != 0 {
		if cc := floor.World.Get(sess.PlayerID, component.CCombat); cc != nil {
			c := cc.(component.Combat)
			c.Attack += sess.FurnitureATK
			c.Defense += sess.FurnitureDEF
			floor.World.Add(sess.PlayerID, c)
		}
	}

	// Spawn class start items adjacent to the player (city spawn only).
	if floorNum == 0 {
		for i, glyph := range sess.Class.StartItems {
			offsets := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
			ox, oy := offsets[i%len(offsets)][0], offsets[i%len(offsets)][1]
			ix, iy := sx+ox, sy+oy
			if !floor.GMap.InBounds(ix, iy) || !floor.GMap.IsWalkable(ix, iy) {
				ix, iy = sx, sy
			}
			factory.NewItemByGlyph(floor.World, glyph, ix, iy)
		}
	}

	if sess.RunLog.FloorsReached < floorNum {
		sess.RunLog.FloorsReached = floorNum
	}

	system.UpdateFOV(floor.World, floor.GMap, sess.PlayerID, sess.FovRadius)
	sess.SnapshotFOV(floor.GMap)
	sess.Renderer = render.NewRenderer(sess.Screen, floorNum)
	sess.Renderer.CenterOn(sx, sy)
}

// respawnLocked resets a dead session and returns them to Emberveil (floor 0).
// Caller must hold s.mu.
func (s *Server) respawnLocked(sess *Session) {
	// Destroy old entity if still present.
	if floor, ok := s.floors[sess.FloorNum]; ok && sess.PlayerID != ecs.NilEntity {
		floor.World.DestroyEntity(sess.PlayerID)
	}

	// Reset per-run stats but keep class and furniture bonuses.
	sess.RunLog = RunLog{
		EnemiesKilled: make(map[string]int),
		ItemsUsed:     make(map[string]int),
		Class:         sess.Class.Name,
	}
	sess.DiscoveredEnemies = make(map[string]bool)
	sess.TurnCount = 0
	sess.SpecialCooldown = 0
	sess.FurnitureATK = 0
	sess.FurnitureDEF = 0
	sess.FurnitureThorns = 0
	sess.FurnitureKR = false
	sess.FovRadius = sess.Class.FOVRadius
	sess.BaseMaxHP = sess.Class.MaxHP
	sess.PlayerID = ecs.NilEntity
	sess.Gold = 0

	sess.AddMessage("You respawn in Emberveil...")
	s.spawnPlayerLocked(sess, 0)
}

// ─── Per-session rendering ────────────────────────────────────────────────────

// RenderSession renders the current world state to a session's screen.
// Must be called while holding s.mu (to safely access ECS and gmap).
func (s *Server) RenderSession(sess *Session) {
	floor, ok := s.floors[sess.FloorNum]
	if !ok || sess.Renderer == nil {
		return
	}

	if sess.DeathCountdown > 0 {
		drawDeathScreen(sess, sess.DeathCountdown)
		return
	}

	// Apply this session's FOV to the shared gamemap.
	sess.ApplyFOV(floor.GMap)

	// Center camera on player.
	posComp := floor.World.Get(sess.PlayerID, component.CPosition)
	if posComp != nil {
		pos := posComp.(component.Position)
		sess.Renderer.CenterOn(pos.X, pos.Y)
	}

	sess.Renderer.DrawFrame(floor.World, floor.GMap, sess.PlayerID)

	// Compute bonuses for HUD.
	equipATK, equipDEF := equipBonuses(floor.World, sess.PlayerID)
	bonusATK := system.GetAttackBonus(floor.World, sess.PlayerID) + equipATK
	bonusDEF := system.GetDefenseBonus(floor.World, sess.PlayerID) + equipDEF

	// Embed player count and gold in className field for HUD display.
	className := fmt.Sprintf("%s [%d online] 💰%d", sess.Class.Name, len(s.sessions), sess.Gold)

	sess.Renderer.DrawHUD(floor.World, sess.PlayerID, sess.FloorNum, className,
		sess.Messages, bonusATK, bonusDEF, sess.Class.AbilityName, sess.SpecialCooldown)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func restoreHP(w *ecs.World, id ecs.EntityID, n int) {
	hp := w.Get(id, component.CHealth)
	if hp == nil {
		return
	}
	h := hp.(component.Health)
	h.Current += n
	if h.Current > h.Max {
		h.Current = h.Max
	}
	w.Add(id, h)
}

func recalcMaxHP(w *ecs.World, sess *Session) {
	invComp := w.Get(sess.PlayerID, component.CInventory)
	if invComp == nil {
		return
	}
	inv := invComp.(component.Inventory)
	bonus := inv.Head.BonusMaxHP + inv.Body.BonusMaxHP + inv.Feet.BonusMaxHP +
		inv.MainHand.BonusMaxHP + inv.OffHand.BonusMaxHP
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

func equipBonuses(w *ecs.World, id ecs.EntityID) (atk, def int) {
	c := w.Get(id, component.CInventory)
	if c == nil {
		return 0, 0
	}
	inv := c.(component.Inventory)
	atk = inv.MainHand.BonusATK + inv.OffHand.BonusATK + inv.Head.BonusATK + inv.Body.BonusATK + inv.Feet.BonusATK
	def = inv.MainHand.BonusDEF + inv.OffHand.BonusDEF + inv.Head.BonusDEF + inv.Body.BonusDEF + inv.Feet.BonusDEF
	return atk, def
}

func entityGlyph(w *ecs.World, id ecs.EntityID) string {
	c := w.Get(id, component.CRenderable)
	if c == nil {
		return "creature"
	}
	return c.(component.Renderable).Glyph
}

func (s *Server) sessionByPlayerID(pid ecs.EntityID) *Session {
	for _, sess := range s.sessions {
		if sess.PlayerID == pid {
			return sess
		}
	}
	return nil
}

// findFreeSpawn returns the nearest walkable cell to (x, y) that is not
// already occupied by another player on the same floor.
func findFreeSpawn(floor *Floor, sessions []*Session, x, y int) (int, int) {
	occupied := func(tx, ty int) bool {
		if !floor.GMap.InBounds(tx, ty) || !floor.GMap.IsWalkable(tx, ty) {
			return true
		}
		for _, sess := range sessions {
			if sess.FloorNum != floor.Num || sess.PlayerID == ecs.NilEntity {
				continue
			}
			if pos := floor.World.Get(sess.PlayerID, component.CPosition); pos != nil {
				p := pos.(component.Position)
				if p.X == tx && p.Y == ty {
					return true
				}
			}
		}
		return false
	}
	if !occupied(x, y) {
		return x, y
	}
	for r := 1; r <= 10; r++ {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if !occupied(x+dx, y+dy) {
					return x + dx, y + dy
				}
			}
		}
	}
	return x, y // final fallback
}


func (s *Server) isPlayerEntity(id ecs.EntityID) bool {
	for _, sess := range s.sessions {
		if sess.PlayerID == id {
			return true
		}
	}
	return false
}

func (s *Server) checkVictoryLocked(floor *Floor, sess *Session) {
	if floor.Num != MaxFloors {
		return
	}
	bossGlyph := assets.BossGlyphs[floor.Num]
	if bossGlyph == "" {
		return
	}
	for _, id := range floor.World.Query(component.CRenderable) {
		rend := floor.World.Get(id, component.CRenderable).(component.Renderable)
		if rend.Glyph == bossGlyph {
			return // boss still alive
		}
	}
	sess.RunLog.Victory = true
	floorMessage(s.sessions, floor.Num, fmt.Sprintf("🏆 %s has defeated the boss! The Spire's heart is theirs!", sess.Name))
	sess.RunLog.Timestamp = time.Now()
	saveRunLog(sess.RunLog)
	sess.DeathCountdown = DeathTicks // reuse death countdown for victory → respawn
}

// applyDoTLocked applies poison and self-burn damage to one session's player.
func (s *Server) applyDoTLocked(floor *Floor, sess *Session) {
	poisonDmg := system.GetPoisonDamage(floor.World, sess.PlayerID)
	burnDmg := system.GetSelfBurnDamage(floor.World, sess.PlayerID)
	total := poisonDmg + burnDmg
	if total <= 0 {
		return
	}
	hp := floor.World.Get(sess.PlayerID, component.CHealth)
	if hp == nil {
		return
	}
	h := hp.(component.Health)
	h.Current -= total
	floor.World.Add(sess.PlayerID, h)
	sess.RunLog.DamageTaken += total
	if poisonDmg > 0 {
		sess.RunLog.CauseOfDeath = "poison"
		sess.AddMessage(fmt.Sprintf("Poison burns through you! (%d damage)", poisonDmg))
	}
	if burnDmg > 0 {
		if sess.RunLog.CauseOfDeath == "" {
			sess.RunLog.CauseOfDeath = "self-burn"
		}
		sess.AddMessage(fmt.Sprintf("The resonance burns you! (%d damage)", burnDmg))
	}
}

// checkInscriptionLocked shows any wall text at the player's current position.
func (s *Server) checkInscriptionLocked(floor *Floor, sess *Session) {
	posComp := floor.World.Get(sess.PlayerID, component.CPosition)
	if posComp == nil {
		return
	}
	pos := posComp.(component.Position)
	for _, id := range floor.World.Query(component.CInscription, component.CPosition) {
		ipos := floor.World.Get(id, component.CPosition).(component.Position)
		if ipos.X == pos.X && ipos.Y == pos.Y {
			text := floor.World.Get(id, component.CInscription).(component.Inscription).Text
			sess.RunLog.InscriptionsRead++
			sess.AddMessage("📝 " + text)
			return
		}
	}
}

// tryPickupLocked picks up an item at the player's position.
func (s *Server) tryPickupLocked(floor *Floor, sess *Session) {
	posComp := floor.World.Get(sess.PlayerID, component.CPosition)
	if posComp == nil {
		return
	}
	pos := posComp.(component.Position)
	for _, itemID := range floor.World.Query(component.CTagItem, component.CPosition) {
		ipos := floor.World.Get(itemID, component.CPosition).(component.Position)
		if ipos.X != pos.X || ipos.Y != pos.Y {
			continue
		}
		itemComp := floor.World.Get(itemID, component.CItem)
		if itemComp == nil {
			sess.AddMessage("Strange item — cannot pick up.")
			return
		}
		item := itemComp.(component.CItemComp).Item
		invComp := floor.World.Get(sess.PlayerID, component.CInventory)
		if invComp == nil {
			return
		}
		inv := invComp.(component.Inventory)
		if len(inv.Backpack) >= inv.Capacity {
			sess.AddMessage("Backpack full! Drop something first.")
			return
		}
		inv.Backpack = append(inv.Backpack, item)
		floor.World.Add(sess.PlayerID, inv)
		floor.World.DestroyEntity(itemID)
		sess.AddMessage(fmt.Sprintf("You pick up %s. [i] to open inventory.", item.Name))
		return
	}
	sess.AddMessage("Nothing to pick up here.")
}

// interactNPCLocked handles a player bumping into an NPC.
// Caller must hold s.mu.
func (s *Server) interactNPCLocked(floor *Floor, sess *Session, _ ecs.EntityID, npc component.NPC) {
	switch npc.Kind {
	case component.NPCKindHealer:
		hp := floor.World.Get(sess.PlayerID, component.CHealth).(component.Health)
		if hp.Current == hp.Max {
			line := npc.Lines[floor.Rng.Intn(len(npc.Lines))]
			sess.AddMessage(fmt.Sprintf("💬 %s: \"%s\"", npc.Name, line))
			return
		}
		old := hp.Current
		hp.Current = hp.Max
		floor.World.Add(sess.PlayerID, hp)
		sess.AddMessage(fmt.Sprintf("✨ %s heals you to full! (%d → %d HP)", npc.Name, old, hp.Max))

	case component.NPCKindShop:
		sess.PendingNPC = 1 // any non-zero signals RunShop; exact ID not needed

	case component.NPCKindAnimal:
		line := npc.Lines[floor.Rng.Intn(len(npc.Lines))]
		sess.AddMessage(line) // no speech marks for animals

	default: // NPCKindDialogue
		line := npc.Lines[floor.Rng.Intn(len(npc.Lines))]
		sess.AddMessage(fmt.Sprintf("💬 %s: \"%s\"", npc.Name, line))
	}
}

// interactFurnitureLocked applies a furniture bonus to the player.
func (s *Server) interactFurnitureLocked(floor *Floor, sess *Session, id ecs.EntityID) {
	fc := floor.World.Get(id, component.CFurniture)
	if fc == nil {
		return
	}
	f := fc.(component.Furniture)
	sess.AddMessage(fmt.Sprintf("%s %s: %s", f.Glyph, f.Name, f.Description))
	if f.IsRepeatable {
		return // atmospheric furniture — description only, no bonus
	}
	if f.Used {
		return
	}
	hasBonus := f.BonusATK != 0 || f.BonusDEF != 0 || f.BonusMaxHP != 0 ||
		f.HealHP != 0 || f.PassiveKind != 0
	if !hasBonus {
		return
	}
	if f.BonusATK != 0 {
		sess.FurnitureATK += f.BonusATK
		if cc := floor.World.Get(sess.PlayerID, component.CCombat); cc != nil {
			c := cc.(component.Combat)
			c.Attack += f.BonusATK
			floor.World.Add(sess.PlayerID, c)
		}
		sess.AddMessage(fmt.Sprintf("Permanent ATK +%d!", f.BonusATK))
	}
	if f.BonusDEF != 0 {
		sess.FurnitureDEF += f.BonusDEF
		if cc := floor.World.Get(sess.PlayerID, component.CCombat); cc != nil {
			c := cc.(component.Combat)
			c.Defense += f.BonusDEF
			floor.World.Add(sess.PlayerID, c)
		}
		sess.AddMessage(fmt.Sprintf("Permanent DEF +%d!", f.BonusDEF))
	}
	if f.BonusMaxHP != 0 {
		sess.BaseMaxHP += f.BonusMaxHP
		if hp := floor.World.Get(sess.PlayerID, component.CHealth); hp != nil {
			h := hp.(component.Health)
			h.Max += f.BonusMaxHP
			h.Current += f.BonusMaxHP
			if h.Current > h.Max {
				h.Current = h.Max
			}
			floor.World.Add(sess.PlayerID, h)
		}
		sess.AddMessage(fmt.Sprintf("Permanent MaxHP +%d!", f.BonusMaxHP))
	}
	if f.HealHP != 0 {
		restoreHP(floor.World, sess.PlayerID, f.HealHP)
		sess.AddMessage(fmt.Sprintf("Restored %d HP!", f.HealHP))
	}
	switch f.PassiveKind {
	case component.PassiveKeenEye:
		sess.FovRadius++
		system.UpdateFOV(floor.World, floor.GMap, sess.PlayerID, sess.FovRadius)
		sess.SnapshotFOV(floor.GMap)
		sess.AddMessage("Your vision expands permanently.")
	case component.PassiveKillRestore:
		sess.FurnitureKR = true
		sess.AddMessage("You feel your life force quicken on each kill.")
	case component.PassiveThorns:
		sess.FurnitureThorns++
		sess.AddMessage("Sharp crystals form beneath your skin.")
	}
	f.Used = true
	floor.World.Add(id, f)
}

// useSpecialAbilityLocked fires the class active ability.
func (s *Server) useSpecialAbilityLocked(floor *Floor, sess *Session) {
	switch sess.Class.ID {
	case "arcanist":
		rooms := floor.GMap.Rooms
		if len(rooms) == 0 {
			return
		}
		room := rooms[floor.Rng.Intn(len(rooms))]
		x, y := room.Center()
		floor.World.Add(sess.PlayerID, component.Position{X: x, Y: y})
		system.UpdateFOV(floor.World, floor.GMap, sess.PlayerID, sess.FovRadius)
		sess.SnapshotFOV(floor.GMap)
		sess.AddMessage("Dimensional Rift tears open — you reappear elsewhere!")

	case "revenant":
		hpComp := floor.World.Get(sess.PlayerID, component.CHealth)
		if hpComp == nil {
			return
		}
		hp := hpComp.(component.Health)
		if hp.Current <= 5 {
			sess.AddMessage("Too wounded to bargain with death!")
			sess.SpecialCooldown = 0
			return
		}
		hp.Current -= 5
		floor.World.Add(sess.PlayerID, hp)
		system.ApplyEffect(floor.World, sess.PlayerID, component.ActiveEffect{
			Kind: component.EffectAttackBoost, Magnitude: 6, TurnsRemaining: 8,
		})
		sess.AddMessage("Death's Bargain struck! (-5 HP, +6 ATK for 8 turns)")

	case "construct":
		system.ApplyEffect(floor.World, sess.PlayerID, component.ActiveEffect{
			Kind: component.EffectAttackBoost, Magnitude: 6, TurnsRemaining: 6,
		})
		system.ApplyEffect(floor.World, sess.PlayerID, component.ActiveEffect{
			Kind: component.EffectSelfBurn, Magnitude: 2, TurnsRemaining: 6,
		})
		sess.AddMessage("Overclock engaged! (+6 ATK for 6 turns, -2 HP/turn burn)")

	case "dancer":
		system.ApplyEffect(floor.World, sess.PlayerID, component.ActiveEffect{
			Kind: component.EffectInvisible, Magnitude: 1, TurnsRemaining: 8,
		})
		sess.AddMessage("Vanish! You fade from perception for 8 turns.")

	case "oracle":
		for y := range floor.GMap.Height {
			for x := range floor.GMap.Width {
				if floor.GMap.At(x, y).Walkable {
					floor.GMap.At(x, y).Explored = true
				}
			}
		}
		sess.AddMessage("Farsight! The entire floor is revealed.")

	case "symbiont":
		restoreHP(floor.World, sess.PlayerID, 10)
		system.ApplyEffect(floor.World, sess.PlayerID, component.ActiveEffect{
			Kind: component.EffectAttackBoost, Magnitude: 4, TurnsRemaining: 6,
		})
		sess.AddMessage("Parasite Surge! (+10 HP, +4 ATK for 6 turns)")
	}
}

// applyConsumableLocked applies a consumed item's effect.
func (s *Server) applyConsumableLocked(floor *Floor, sess *Session, item component.Item) {
	glyph := item.Glyph
	sess.RunLog.ItemsUsed[glyph]++
	switch glyph {
	case assets.GlyphHyperflask:
		restoreHP(floor.World, sess.PlayerID, 15)
		sess.AddMessage("The Hyperflask restores 15 HP.")
	case assets.GlyphPrismShard:
		system.ApplyEffect(floor.World, sess.PlayerID, component.ActiveEffect{
			Kind: component.EffectAttackBoost, Magnitude: 3, TurnsRemaining: 10,
		})
		sess.AddMessage("The Prism Shard boosts your ATK by 3 for 10 turns!")
	case assets.GlyphNullCloak:
		system.ApplyEffect(floor.World, sess.PlayerID, component.ActiveEffect{
			Kind: component.EffectInvisible, Magnitude: 1, TurnsRemaining: 12,
		})
		sess.AddMessage("The Null Cloak makes you invisible for 12 turns.")
	case assets.GlyphTesseract:
		rooms := floor.GMap.Rooms
		if len(rooms) > 0 {
			room := rooms[floor.Rng.Intn(len(rooms))]
			x, y := room.Center()
			floor.World.Add(sess.PlayerID, component.Position{X: x, Y: y})
			system.UpdateFOV(floor.World, floor.GMap, sess.PlayerID, sess.FovRadius)
			sess.SnapshotFOV(floor.GMap)
		}
		sess.AddMessage("The Tesseract Cube warps you to a random location!")
	case assets.GlyphMemoryScroll:
		for y := range floor.GMap.Height {
			for x := range floor.GMap.Width {
				if floor.GMap.At(x, y).Walkable {
					floor.GMap.At(x, y).Explored = true
				}
			}
		}
		sess.AddMessage("The Memory Scroll reveals the entire floor.")
	case assets.GlyphSporeDraught:
		restoreHP(floor.World, sess.PlayerID, 20)
		sess.AddMessage("The Spore Draught mends your wounds. (+20 HP)")
	case assets.GlyphResonanceCoil:
		system.ApplyEffect(floor.World, sess.PlayerID, component.ActiveEffect{
			Kind: component.EffectAttackBoost, Magnitude: 5, TurnsRemaining: 12,
		})
		sess.AddMessage("The Resonance Coil harmonises. (+5 ATK, 12 turns)")
	case assets.GlyphPrismaticWard:
		system.ApplyEffect(floor.World, sess.PlayerID, component.ActiveEffect{
			Kind: component.EffectDefenseBoost, Magnitude: 4, TurnsRemaining: 12,
		})
		sess.AddMessage("The Prismatic Ward refracts harm. (+4 DEF, 12 turns)")
	case assets.GlyphVoidEssence:
		system.ApplyEffect(floor.World, sess.PlayerID, component.ActiveEffect{
			Kind: component.EffectInvisible, Magnitude: 1, TurnsRemaining: 20,
		})
		sess.AddMessage("The Void Essence erases you from spacetime. (Invisible, 20 turns)")
	case assets.GlyphNanoSyringe:
		restoreHP(floor.World, sess.PlayerID, 30)
		sess.AddMessage("The Nano-Syringe floods your bloodstream. (+30 HP)")
	case assets.GlyphResonanceBurst:
		system.ApplyEffect(floor.World, sess.PlayerID, component.ActiveEffect{
			Kind: component.EffectAttackBoost, Magnitude: 8, TurnsRemaining: 8,
		})
		system.ApplyEffect(floor.World, sess.PlayerID, component.ActiveEffect{
			Kind: component.EffectSelfBurn, Magnitude: 2, TurnsRemaining: 8,
		})
		sess.AddMessage("Resonance Burst! (+8 ATK, -2 HP/turn burn, 8 turns)")
	case assets.GlyphPhaseRod:
		system.ApplyEffect(floor.World, sess.PlayerID, component.ActiveEffect{
			Kind: component.EffectDefenseBoost, Magnitude: 6, TurnsRemaining: 15,
		})
		sess.AddMessage("The Phase Rod envelops you in prismatic shielding. (+6 DEF, 15 turns)")
	case assets.GlyphApexCore:
		sess.BaseMaxHP += 3
		hp := floor.World.Get(sess.PlayerID, component.CHealth).(component.Health)
		hp.Max += 3
		hp.Current += 3
		floor.World.Add(sess.PlayerID, hp)
		sess.AddMessage("The Apex Core integrates into your biology. (+3 MaxHP permanently)")
	}
}

// respawnEnemiesLocked spawns a partial enemy wave on the given floor.
// Used when a cleared floor has active players and the respawn timer fires.
// Caller must hold s.mu.
func (s *Server) respawnEnemiesLocked(floor *Floor) {
	cfg := levelConfig(floor.Num, floor.Rng)
	if len(cfg.EnemyTable) == 0 || len(floor.GMap.Rooms) == 0 {
		return
	}

	budget := max(cfg.EnemyBudget/3, 3)

	// Skip the first room (player spawn area) when picking placement rooms.
	startIdx := 1
	if len(floor.GMap.Rooms) < 2 {
		startIdx = 0
	}
	rooms := floor.GMap.Rooms[startIdx:]

	for attempts := 0; budget > 0 && attempts < 30; attempts++ {
		entry := cfg.EnemyTable[floor.Rng.Intn(len(cfg.EnemyTable))]
		if entry.ThreatCost > budget {
			continue
		}
		room := rooms[floor.Rng.Intn(len(rooms))]
		cx, cy := room.Center()
		factory.NewEnemy(floor.World, entry, cx, cy)
		budget -= entry.ThreatCost
	}

	floorMessage(s.sessions, floor.Num, "🌑 The dungeon stirs as new threats emerge from the shadows...")
}
