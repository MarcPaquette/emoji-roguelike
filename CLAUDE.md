# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build ./...          # compile everything
go test ./...           # run all tests
go test ./internal/ecs/ # run tests for a single package
go test -run TestFoo ./internal/system/  # run one test by name
go mod tidy             # sync go.sum after changing dependencies
./emoji-roguelike       # run the game (requires an emoji-capable terminal)
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
| `CInventory` | 6 | `Inventory{Items, Capacity}` |
| `CEffects` | 7 | `Effects{Active []ActiveEffect}` |
| `CTagPlayer` | 8 | marker |
| `CTagBlocking` | 9 | marker |
| `CTagItem` | 10 | marker |
| `CTagStairs` | 11 | marker |
| `CInscription` | 12 | `Inscription{Text string}` |

**Next available:** 13. Never reuse a number.

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
```

### Dungeon generation (`internal/generate/`)
BSP tree splits the map recursively until leaves are ≤ `MaxLeafSize`. Each terminal leaf gets one room carved into it. `connectChildren` walks the tree and carves L-shaped (or Z/straight) corridors between sibling rooms. `Populate` places enemies against a `EnemyBudget` point pool and scatters items.

Difficulty is lerp'd over `t = (floor-1)/(MaxFloors-1)` in `game/levels.go`:
- Map grows 40×20 → 80×45
- `MaxLeafSize` shrinks 20 → 12 (more, smaller rooms)
- `EnemyBudget` grows 5 → 40

### Rendering (`internal/render/`)
**Critical — emoji are 2 terminal columns wide.** All world X coordinates are multiplied by 2 on the way to the screen (`sx = (wx - OffsetX) * 2`). `putGlyph` writes the leading rune via `SetContent` then fills column `x+1` with a space to prevent artifacts.

Tile glyphs are per-floor emoji defined in `render/colors.go` (`TileThemes[floorNum]`). Visible tiles use thematic emoji; explored-but-dark tiles use `🌑` (wall) / `🔲` (floor).

The HUD occupies the bottom 5 terminal rows. `DrawHUD` signature: `(w, playerID, floor int, className string, messages []string)`.

### FOV (`internal/system/fov.go`)
Recursive shadowcasting, 8 octants. **Variable roles matter:** `dy = -j` is the fixed row index; `dx` sweeps from `-j` to `0` within each row. The octant transform is `worldX = cx + dx*xx + dy*xy`. Mixing up which variable is fixed breaks the algorithm visibly (jagged non-circular shadows).

### Class system (`assets/theme.go`, `internal/game/classselect.go`)
`ClassDef` holds base stats, FOV radius, and passive flags (`KillRestoreHP`, `StartInvisible`, `StartRevealMap`, `StartItems`). The selection screen runs once before `loadFloor(1)`. `factory.NewPlayer` takes a `ClassDef` and applies stats/glyph directly. Passives fire in `loadFloor` (floor 1 only) and in `processAction`'s kill branch. `Game.fovRadius` is set from the class and passed to every `UpdateFOV` call.

### Game state machine (`internal/game/game.go`)
States: `StatePlaying`, `StateInventory`, `StateDead`, `StateVictory`, `StateClassSelect`. The main loop in `Run()` skips rendering when not in `StatePlaying`. Floor transitions preserve the player's current HP (saved before `ecs.NewWorld()`, restored after `NewPlayer`).

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

## Testing

Tests live alongside source in `_test.go` files. The test strategy is table-driven where multiple cases apply. Combat tests create a fresh world per iteration — do not share a long-lived defender entity across attack loops (HP will hit 0 mid-test).
