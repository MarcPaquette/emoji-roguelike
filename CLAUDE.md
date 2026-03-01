# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Task Tracking (beads)

Use `bd` (beads) for **all** task tracking and planning. Do NOT use TodoWrite, TaskCreate, or markdown files for task management.

```bash
bd ready                          # find available work (no blockers)
bd list --status=open             # all open issues
bd show <id>                      # detailed view with dependencies
bd create --title="..." --description="..." --type=task|bug|feature --priority=2
bd update <id> --status=in_progress  # claim work
bd close <id>                     # mark complete
bd close <id1> <id2> ...          # close multiple at once
bd dep add <issue> <depends-on>   # add dependency
bd blocked                        # show blocked issues
```

**Workflow:** `bd ready` → `bd show <id>` → `bd update <id> --status=in_progress` → do work → `bd close <id>` → git commit & push.

Priority is 0–4 (0=critical, 2=medium, 4=backlog). Do NOT use `bd edit` (opens $EDITOR, blocks agents).

## Commands

```bash
go build ./...          # compile everything
go test ./...           # run all tests
go test ./internal/ecs/ # run tests for a single package
go test -run TestFoo ./internal/system/  # run one test by name
go vet ./...            # static analysis (run before committing)
go fix ./...            # update deprecated API usage
go mod tidy             # sync go.sum after changing dependencies
./emoji-roguelike       # run single-player (requires emoji-capable terminal)

# MUD server
go build -o emoji-roguelike-server ./cmd/server
./emoji-roguelike-server            # listens on :2222
ssh localhost -p 2222               # connect as a player
```

Go version: **1.25.4** — `min`/`max` builtins and `for i := range N` available.

Both binaries require a terminal with full emoji support (kitty, GNOME Terminal, iTerm2). Plain xterm will render emoji incorrectly.

## Architecture

### ECS core (`internal/ecs/`)
`World` is the sole data store. Every game object is an `EntityID` (uint64). Components are stored in `map[ComponentType]map[EntityID]Component`. `World.Query(types...)` returns all alive entities that possess every listed component type, using the smallest store as the candidate set to minimise iteration.

### Component types (`internal/component/`)
Pure data structs — zero logic. Each implements `Type() ComponentType`. The iota sequence:

| Const | Value | Type |
|---|---|---|
| `CPosition` | 1 | `Position{X, Y int}` |
| `CHealth` | 2 | `Health{Current, Max int}` |
| `CRenderable` | 3 | `Renderable{Glyph, FGColor, BGColor, RenderOrder}` |
| `CCombat` | 4 | `Combat{Attack, Defense int}` |
| `CAI` | 5 | `AI{Behavior, SightRange}` |
| `CInventory` | 6 | `Inventory{Backpack []Item, Capacity int, Head/Body/Feet/MainHand/OffHand Item}` |
| `CEffects` | 7 | `Effects{Active []ActiveEffect}` |
| `CTagPlayer` | 8 | marker |
| `CTagBlocking` | 9 | marker |
| `CTagItem` | 10 | marker |
| `CTagStairs` | 11 | marker |
| `CInscription` | 12 | `Inscription{Text string}` |
| `CItem` | 13 | `CItemComp{Item}` — wraps `component.Item` value for floor entities |
| `CLoot` | 14 | `Loot` — loot drop data |
| `CFurniture` | 15 | `Furniture{Glyph, Name, Description, BonusATK/DEF/MaxHP, HealHP, PassiveKind, Used, IsRepeatable}` |
| `CNPC` | 16 | `NPC{Name, Kind, DialogueLines, Glyph}` — city NPCs (Dialogue/Healer/Shop/Animal) |

**Next available:** 17. Never reuse a number.

### Dependency rule (strict)
```
component  ← leaf, no imports from this module
ecs        ← no game concepts
gamemap    ← no system/render
generate   ← gamemap only; never system or render
factory    ← ecs, component, generate, assets
system     ← ecs, component, gamemap
render     ← ecs, component, gamemap; NEVER imports system
game       ← everything
mud        ← ecs, component, gamemap, generate, factory, assets, system, render
ssh        ← no game concepts (tcell.Tty adapter only)
assets     ← generate only
```

### Dungeon generation (`internal/generate/`)
BSP tree splits the map recursively until leaves are ≤ `MaxLeafSize`. Each terminal leaf gets one room carved into it. `connectChildren` walks the tree and carves L-shaped (or Z/straight) corridors between sibling rooms. `Populate` places enemies against a `EnemyBudget` point pool and scatters items.

Difficulty is lerp'd over `t = (floor-1)/(MaxFloors-1)` in `game/levels.go`:
- Map grows 40×20 → 90×50
- `MaxLeafSize` shrinks 20 → 10 (more, smaller rooms)
- `EnemyBudget` grows 5 → 55

### Rendering (`internal/render/`)
**Critical — emoji are 2 terminal columns wide.** All world X coordinates are multiplied by 2 on the way to the screen (`sx = (wx - OffsetX) * 2`). `putGlyph` writes the leading rune via `SetContent` then fills column `x+1` with a space to prevent artifacts.

Tile glyphs are per-floor emoji defined in `render/colors.go` (`TileThemes[floorNum]`). Visible tiles use thematic emoji; explored-but-dark tiles use `🌑` (wall) / `🔲` (floor).

The HUD occupies the bottom 5 terminal rows. `DrawHUD` signature: `(w, playerID, floor int, className string, messages []string, bonusATK, bonusDEF int, abilityName string, abilityCooldown int)`.

### FOV (`internal/system/fov.go`)
Recursive shadowcasting, 8 octants. **Variable roles matter:** `dy = -j` is the fixed row index; `dx` sweeps from `-j` to `0` within each row. The octant transform is `worldX = cx + dx*xx + dy*xy`. Mixing up which variable is fixed breaks the algorithm visibly (jagged non-circular shadows).

### Class system (`assets/theme.go`, `internal/game/classselect.go`)
`ClassDef` holds base stats, FOV radius, passive fields (`KillHealChance`, `PassiveRegen`, `StartItems`), and active ability fields (`AbilityName`, `AbilityCooldown`, `AbilityFreeOnFloor`). The selection screen runs once before `loadFloor(1)`. `factory.NewPlayer` takes a `ClassDef` and applies stats/glyph directly. `Game.fovRadius` is set from the class and passed to every `UpdateFOV` call.

Active abilities fire on `z` key (`ActionSpecialAbility`). `Game.specialCooldown` tracks turns remaining. Classes with `AbilityFreeOnFloor=true` get cooldown reset each floor. `KillHealChance` (percentage) restores HP on kill. `PassiveRegen` (N turns) restores 1 HP every N turns.

### Game state machine (`internal/game/game.go`)
States: `StatePlaying`, `StateInventory`, `StateDead`, `StateVictory`, `StateClassSelect`. The main loop in `Run()` skips rendering when not in `StatePlaying`. Floor transitions preserve the player's current HP (saved before `ecs.NewWorld()`, restored after `NewPlayer`).

### Run logging
Two `RunLog` structs exist — single-player (`internal/game/runlog.go`) and MUD (`internal/mud/runlog.go`). Both append one JSON line per completed run to `~/.local/share/emoji-roguelike/runs.jsonl`. The MUD version adds a `gold_earned` field. `saveRunLog()` silently discards I/O errors.

```bash
jq -r '.victory' ~/.local/share/emoji-roguelike/runs.jsonl | sort | uniq -c
jq -r '.cause_of_death' ~/.local/share/emoji-roguelike/runs.jsonl | sort | uniq -c | sort -rn
```

### Furniture system (`internal/component/furniture.go`, `assets/furniture.go`)
Bump-to-interact objects placed in dungeon rooms (1–2 per room, 15% rare chance). Grant one-time bonuses (ATK/DEF/MaxHP/Heal) or passives (KeenEye, KillRestore, Thorns). `IsRepeatable=true` for city furniture (shows description every bump, no bonus). Furniture blocks movement; enemies cannot pass through. Bonuses persist across floors via `Game.furnitureATK/DEF/Thorns/KillRestore`.

### MUD server (`internal/mud/`, `internal/ssh/`, `cmd/server/`)
Multiplayer SSH server. N players share a persistent world with tick-based updates (500ms). One ticker goroutine + one goroutine per session. Key files:
- `mud/server.go` — `Server`, `tick()`, action processing, floor transitions
- `mud/session.go` — per-player state, FOV snapshot/apply
- `mud/floor.go` — `Floor` struct, `newFloor()`, lazy floor generation
- `mud/city.go` — Emberveil (Floor 0), 110×55 hand-crafted starting city with NPCs/shops
- `mud/shop.go` — shop UI, `Session.Gold` currency (earned by killing enemies)
- `mud/inventory.go` — modal inventory (reads from `eventCh`)
- `ssh/tty.go` — `SessionTty` implements `tcell.Tty` over SSH

Floor 0 is a safe zone (no combat/AI). Players spawn at city center, respawn there on death (gold reset). `Session.PendingNPC` triggers shop/healer/dialogue interactions. Floor 1+ have stairs UP to return to city.

### NPC system (`component/npc.go`, `assets/city.go`)
`CNPC=16`. NPCKind: Dialogue, Healer, Shop, Animal. Created via `factory.NewNPC()`. City NPCs defined in `assets/city.go` (`CityNPCs`, `CityAnimals`, `ShopCatalogue`).

## Testing

Tests live alongside source in `_test.go` files. **Run `go test ./...` after every code change — all tests must pass before considering a task complete.**

### Required test coverage

Every package that contains logic (not just data structs) must have tests. When adding or modifying code:

- Write tests for the new/changed behaviour before or alongside the implementation.
- If modifying an existing function, check whether existing tests cover the new path; add cases if not.
- Packages currently with tests: `ecs`, `gamemap`, `generate`, `system`, `factory`, `game`, `mud`.

### Test writing rules

1. **Table-driven for multiple cases** — use a `cases []struct{...}` slice and loop with `t.Run(tc.name, ...)`.
2. **Fresh world per iteration** — never share a mutable ECS world across table rows; a dead entity in one row corrupts later rows.
3. **Test behaviour, not internals** — assert on observable side-effects (HP reduced, entity destroyed, map tile lit) rather than internal state.
4. **Name tests clearly** — `TestFOVWallBlocksLight`, not `TestFOV3`. Use present-tense descriptions.
5. **Helper functions for setup** — extract repeated world/map setup into unexported helpers (`makePlayerAt`, `newEffectsWorld`, `openMap`) so test bodies stay short.
6. **Avoid `t.Fatal` inside loops** — use `t.Errorf` so all iterations are reported, not just the first failure.
7. **Seed RNGs deterministically** — `rand.New(rand.NewSource(42))`. Using multiple seeds (0–9) catches edge cases in randomised logic.
8. **Test error paths** — missing components, zero budget, empty tables, out-of-bounds calls.

### What NOT to test

- Pure data structs in `component/` (no logic to test).
- Rendering (`render/`) — requires a live terminal screen.
- Interactive UI loops (`game/inventory.go`, `game/classselect.go`) — input-driven, not unit-testable.
- `saveRunLog` I/O errors — the function intentionally discards them; testing silent discard adds no value.
