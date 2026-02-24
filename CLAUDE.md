# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build ./...                              # compile both binaries (game + server)
go build -o emoji-roguelike-server ./cmd/server  # build SSH server separately
go test ./...                              # run all tests
go test ./internal/ecs/                    # run tests for a single package
go test -run TestFoo ./internal/system/    # run one test by name
go mod tidy                                # sync go.sum after changing dependencies
./emoji-roguelike                          # run single-player game (requires emoji-capable terminal)
./emoji-roguelike-server                   # start SSH co-op server (listens on :2222)
```

The binary requires a terminal with full emoji support (kitty, GNOME Terminal, iTerm2). Plain xterm will render emoji incorrectly.

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
| `CLoot` | 14 | `Loot{Drops []LootEntry}` — drop table for elites/enemies |
| `CFurniture` | 15 | `Furniture{Glyph, Name, Description, BonusATK/DEF/MaxHP, HealHP, PassiveKind, Used}` |

**Next available:** 16. Never reuse a number.

### Effect kinds (`component/effects.go`)
`EffectKind` iota values used by `ActiveEffect`:

| Const | Value | Notes |
|---|---|---|
| `EffectAttackBoost` | 0 | +Magnitude ATK for Duration turns |
| `EffectInvisible` | 1 | enemies ignore the entity |
| `EffectRevealMap` | 2 | unused at runtime; kept for legacy |
| `EffectPoison` | 3 | -Magnitude HP/turn |
| `EffectWeaken` | 4 | -Magnitude ATK |
| `EffectDefenseBoost` | 5 | +Magnitude DEF |
| `EffectSelfBurn` | 6 | player self-inflicts -Magnitude HP/turn |
| `EffectStun` | 7 | player cannot act for Duration turns |
| `EffectArmorBreak` | 8 | -Magnitude DEF on the target |

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
assets     ← generate only
ssh        ← tcell, gliderlabs/ssh only; no game packages
cmd/server ← game, ssh, tcell, gliderlabs/ssh
```

### Dungeon generation (`internal/generate/`)
BSP tree splits the map recursively until leaves are ≤ `MaxLeafSize`. Each terminal leaf gets one room carved into it. `connectChildren` walks the tree and carves L-shaped (or Z/straight) corridors between sibling rooms. `Populate` places enemies against a `EnemyBudget` point pool and scatters items, equipment, inscriptions, and furniture.

**Floor elites**: each floor spawns exactly one elite enemy (defined in `assets/theme.go::floorElites`) outside the normal budget. Elites carry a `CLoot` component with a drop table; loot is spawned on the enemy's tile at death.

**Doors**: rooms are surrounded by wall tiles; BSP carves a random door tile (`TileDoor`) into each room wall during connection. Bumping into a `TileDoor` opens it (converts it to a floor tile) without spending a turn on attack.

Difficulty is lerp'd over `t = (floor-1)/(MaxFloors-1)` in `game/levels.go`:
- Map grows 40×20 → 90×50
- `MaxLeafSize` shrinks 20 → 10 (more, smaller rooms)
- `EnemyBudget` grows 5 → 55
- `ItemCount` grows 3 → 8, `EquipCount` grows 1 → 3

### Rendering (`internal/render/`)
**Critical — emoji are 2 terminal columns wide.** All world X coordinates are multiplied by 2 on the way to the screen (`sx = (wx - OffsetX) * 2`). `putGlyph` writes the leading rune via `SetContent` then fills column `x+1` with a space to prevent artifacts.

Tile glyphs are per-floor emoji defined in `render/colors.go` (`TileThemes[floorNum]`). Visible tiles use thematic emoji; explored-but-dark tiles use `🌑` (wall) / `🔲` (floor).

The HUD occupies the bottom 5 terminal rows. `DrawHUD` signature:
```go
func (r *Renderer) DrawHUD(w *ecs.World, playerID ecs.EntityID, floor int,
    className string, messages []string,
    bonusATK, bonusDEF int, abilityName string, abilityCooldown int)
```
`bonusATK`/`bonusDEF` include both equipment and active-effect totals. `abilityName` and `abilityCooldown` drive the `[z] AbilityName (N)` line in the HUD.

### FOV (`internal/system/fov.go`)
Recursive shadowcasting, 8 octants. **Variable roles matter:** `dy = -j` is the fixed row index; `dx` sweeps from `-j` to `0` within each row. The octant transform is `worldX = cx + dx*xx + dy*xy`. Mixing up which variable is fixed breaks the algorithm visibly (jagged non-circular shadows).

### Class system (`assets/theme.go`, `internal/game/classselect.go`)
`ClassDef` holds base stats, FOV radius, and passive/active flags:

| Field | Effect |
|---|---|
| `KillRestoreHP` | Restore N HP on each kill (Void Revenant: 3) |
| `KillHealChance` | % chance to restore 2 HP on each kill (Arcanist: 30%) |
| `PassiveRegen` | Restore 1 HP every N turns (Construct: 8, Symbiont: 5) |
| `StartItems` | Glyphs of items spawned near player on floor 1 (Symbiont gets flask/shard/cloak) |
| `AbilityName` | Display name shown in HUD and class screen |
| `AbilityDesc` | One-liner shown on class selection screen |
| `AbilityCooldown` | Turns between uses of the `z` key (0 = no ability) |
| `AbilityFreeOnFloor` | Reset cooldown to 0 on each new floor entry |

The six classes and their active abilities (`z` key):

| Class | ID | Ability | Cooldown |
|---|---|---|---|
| 🧙 Wandering Arcanist | `arcanist` | Dimensional Rift — teleport to random room | 12 |
| 💀 Void Revenant | `revenant` | Death's Bargain — spend 5 HP for +6 ATK (8 turns) | 15 |
| 🦾 Chrono Construct | `construct` | Overclock — +6 ATK for 6 turns, -2 HP/turn burn | 18 |
| 🌀 Entropy Dancer | `dancer` | Vanish — invisible for 8 turns | 20 (free each floor) |
| 🔮 Crystal Oracle | `oracle` | Farsight — reveal entire floor | 20 (free each floor) |
| 🧬 Void Symbiont | `symbiont` | Parasite Surge — +10 HP, +4 ATK for 6 turns | 12 (free each floor) |

`Game.fovRadius` is set from the class and passed to every `UpdateFOV` call. Passives fire in `loadFloor` (floor 1 only) and in `processAction`'s kill branch.

### Furniture system (`internal/component/furniture.go`, `assets/furniture.go`)
Each floor defines two furniture pools: `Common` (decorative, no bonus) and `Rare` (grants a one-time bonus). The populator (`generate/populator.go`) places 1–2 pieces per room, with rare pieces appearing at roughly 25% probability.

Bumping into a furniture entity triggers `interactFurniture`, which shows the flavour text and—if unused—applies the bonus and marks it `Used`. Furniture bonuses persist across floor transitions (tracked in `Game.furnitureATK/DEF/Thorns/KillRestore`).

Passive kinds (`component/furniture.go`):

| Const | Value | Effect |
|---|---|---|
| `PassiveKeenEye` | 1 | +1 FOV radius permanently |
| `PassiveKillRestore` | 2 | restore 1 HP on each kill |
| `PassiveThorns` | 3 | reflect 1 damage to each attacker |

### Enemy special attacks (`internal/generate/bsp.go`, `internal/system/ai.go`)
`EnemySpawnEntry.SpecialKind` encodes the type of special attack:

| Value | Name | Effect |
|---|---|---|
| 0 | none | normal melee |
| 1 | poison | applies `EffectPoison` for `SpecialDur` turns at `SpecialMag` HP/turn |
| 2 | weaken | applies `EffectWeaken` (-`SpecialMag` ATK) for `SpecialDur` turns |
| 3 | lifedrain | steals HP from player and heals the attacker |
| 4 | stun | applies `EffectStun` for `SpecialDur` turns (player skips their turn) |
| 5 | armorBreak | applies `EffectArmorBreak` (-`SpecialMag` DEF) for `SpecialDur` turns |

### Game state machine (`internal/game/game.go`)
States: `StatePlaying`, `StateInventory`, `StateDead`, `StateVictory`, `StateClassSelect`. The main loop in `Run()` skips rendering when not in `StatePlaying`. Floor transitions preserve the player's current HP, inventory, and accumulated furniture bonuses (saved before `ecs.NewWorld()`, restored after `NewPlayer`).

### Keybindings (`internal/game/input.go`)
All bindings handled by `keyToAction`:

| Key(s) | Action |
|---|---|
| `↑ ↓ ← →` / `k j h l` / `8 2 6 4` | Move cardinal (numpad layout with numlock) |
| `y u b n` / `7 9 1 3` | Move diagonally (NW NE SW SE) |
| `. 5` | Wait one turn |
| `, g G` | Pick up item at current tile |
| `i I` | Open inventory screen |
| `Enter` | Use stairs (descend or ascend depending on tile) |
| `> <` | Descend / ascend stairs (explicit) |
| `z Z` | Use class special ability |
| `? ` | Show help screen overlay |
| `q Q Esc` | Quit (with confirmation prompt) |

### Run logging (`internal/game/runlog.go`)
Every completed run (death or victory) is appended as one JSON line to `~/.local/share/emoji-roguelike/runs.jsonl` (XDG: `$XDG_DATA_HOME/emoji-roguelike/runs.jsonl`). The `RunLog` struct (defined in `game.go`) is what gets serialised:

```
timestamp, victory, class, floors_reached, turns_played,
enemies_killed{glyph→count}, items_used{glyph→count},
inscriptions_read, damage_dealt, damage_taken, cause_of_death
```

`saveRunLog()` silently discards I/O errors so a full disk can never crash the game. Fields are set throughout gameplay; `Victory` and `Timestamp` are stamped in `Run()` just before `showEndScreen()`.

Quick analysis with standard tools:

```bash
jq -r '.victory' ~/.local/share/emoji-roguelike/runs.jsonl | sort | uniq -c
jq -r '.cause_of_death' ~/.local/share/emoji-roguelike/runs.jsonl | sort | uniq -c | sort -rn
```

### SSH co-op (`cmd/server/`, `internal/ssh/`, `internal/game/coop.go`)
Two players can share one game session over SSH. Architecture:

- `internal/ssh/SessionTty` implements `tcell.Tty` backed by a `gliderlabs/ssh` session, bridging SSH I/O to tcell's screen abstraction.
- `cmd/server/main.go` runs a lobby that pairs the first two connecting clients. Each client gets its own `tcell.Screen`; the pair is handed to `game.NewCoopGame(screens).Run()`.
- `CoopGame` shares a single ECS `World`, `GameMap`, and `RNG`. Per-player state (`coopPlayer`) holds the entity ID, renderer, class, furniture bonuses, and the event channel.

**Co-op turn order**: P1 acts → P2 acts → enemy AI runs. If a floor transition occurs during either player's turn, the remaining player/AI steps are skipped for that round.

**AI targeting in co-op**: enemies chase whichever live player is nearest to the map centroid (`nearestLivePlayerID`). Thorns reflect using the maximum of both players' thorn values (cooperative benefit). Players cannot attack each other.

**Lobby**: at most one game runs at a time. A second P1 must wait in the lobby until P2 connects; additional connections beyond a live pair are rejected with a wait message.

Build and connect:
```bash
go build -o emoji-roguelike-server ./cmd/server
./emoji-roguelike-server          # listens on :2222
ssh -p 2222 -o StrictHostKeyChecking=no localhost  # connect (run twice)
```

The server auto-generates an ed25519 host key (`server_host_key`) on first run. Player 1 renders in yellow, Player 2 in fuchsia.

## Testing

Tests live alongside source in `_test.go` files. **Run `go test ./...` after every code change — all tests must pass before considering a task complete.**

### Required test coverage

Every package that contains logic (not just data structs) must have tests. When adding or modifying code:

- Write tests for the new/changed behaviour before or alongside the implementation.
- If modifying an existing function, check whether existing tests cover the new path; add cases if not.
- Packages currently with tests: `ecs`, `gamemap`, `generate`, `system`, `factory`, `game`.

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
- The SSH server lobby (`cmd/server/`) — requires live network connections.
