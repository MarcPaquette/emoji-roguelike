# emoji-roguelike

A terminal roguelike where every entity is an emoji. Descend 10 floors of the **Prismatic Spire** — an ancient research station piercing the membrane between dimensions — and destroy the Unmaker at the summit to claim victory.

```
🧙 vs 🦀    🧪💎🪄    🔽🔼
```

## Requirements

An emoji-capable terminal is **required**. Recommended:

- [Kitty](https://sw.kovidgoyal.net/kitty/)
- GNOME Terminal
- iTerm2

Plain `xterm` will render emoji incorrectly (emoji are 2 columns wide).

## Build & run

```bash
go build ./...        # compile single-player and server binaries
./emoji-roguelike     # start single-player game
```

## Controls

| Key | Action |
|-----|--------|
| `↑ ↓ ← →` or `k j h l` | Move north/south/west/east |
| `y u b n` | Move diagonally (NW NE SW SE) |
| `,` | Pick up item |
| `i` | Open inventory |
| `>` | Descend stairs |
| `<` | Ascend stairs |
| `.` | Wait one turn |
| `q` / `Esc` | Quit (with confirmation) |

Inside the **inventory screen**, press the item's number key to use or equip it.

## Classes

Choose one at the start of each run:

| Emoji | Class | HP | ATK | DEF | Passive |
|-------|-------|----|-----|-----|---------|
| 🧙 | Wandering Arcanist | 30 | 5 | 2 | — |
| 💀 | Void Revenant | 15 | 12 | 0 | Each kill restores 3 HP |
| 🦾 | Chrono Construct | 60 | 3 | 8 | — (the stats are the passive) |
| 🌀 | Entropy Dancer | 22 | 9 | 1 | Invisible to enemies for 8 turns |
| 🔮 | Crystal Oracle | 20 | 3 | 2 | Entire floor revealed from the start |
| 🧬 | Void Symbiont | 42 | 6 | 5 | Starts with Hyperflask, Prism Shard, and Null Cloak |

## Floors

Each floor has a unique name, tileset, and enemy roster. A floor elite (elite mini-boss) spawns on every level. The Unmaker 🔥 — the final boss — awaits on floor 10.

| Floor | Name | Elite |
|-------|------|-------|
| 1 | Crystalline Labs | 💠 Shardmind |
| 2 | Bioluminescent Warrens | 🍄 Spore Tyrant |
| 3 | Resonance Engine | ⚙️ Gear Revenant |
| 4 | Fractured Observatory | ✨ Prism Specter |
| 5 | Apex Nexus | 🌿 Tendril Overmind |
| 6 | Membrane of Echoes | 📄 Membrane Horror |
| 7 | The Calcified Archive | 📚 Petrified Scholar |
| 8 | Abyssal Foundry | 🌋 Magma Revenant |
| 9 | The Dreaming Cortex | 💭 Somnivore |
| 10 | The Prismatic Heart | 🌟 Prismatic Horror + ☄️ The Unmaker |

## Items

Consumables and equipment are scattered across every floor. New items become available as you descend.

**Consumables** (use from inventory): 🧪💎🫥📦📜🍵🧲💫🌌💉🧨🪄🫀

**Equipment slots:** Head / Body / Feet / Main Hand / Off-Hand. Stats scale with floor depth. Two-hand weapons occupy both weapon slots.

## Furniture

Each floor's rooms contain interactive furniture. Bump into a piece to activate it — once only. Effects include ATK/DEF/HP bonuses and passives such as:

- **Keen Eye** — extended field of view
- **Kill Restore** — HP restored on each kill
- **Thorns** — reflect damage back to attackers

## SSH co-op (experimental)

Two players can share one game session over SSH.

```bash
# build and start the server
go build -o emoji-roguelike-server ./cmd/server
./emoji-roguelike-server                  # listens on :2222

# connect from two terminals (any order)
ssh -p 2222 -o StrictHostKeyChecking=no localhost
```

Player 1 renders in yellow, Player 2 in fuchsia. Turn order: P1 → P2 → enemies. At most one active game at a time; additional connections wait in the lobby.

The server auto-generates an ed25519 host key (`server_host_key`) on first run.

## Run history

Every completed run is appended as a JSON line to:

```
~/.local/share/emoji-roguelike/runs.jsonl
```

Quick analysis:

```bash
# win/loss breakdown
jq -r '.victory' ~/.local/share/emoji-roguelike/runs.jsonl | sort | uniq -c

# most common causes of death
jq -r '.cause_of_death' ~/.local/share/emoji-roguelike/runs.jsonl | sort | uniq -c | sort -rn
```

## Development

```bash
go build ./...              # compile everything
go test ./...               # run all tests (must pass before committing)
go test ./internal/ecs/     # run tests for a single package
go mod tidy                 # sync go.sum after changing dependencies
```

See [CLAUDE.md](CLAUDE.md) for full architecture notes.
