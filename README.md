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
| `z` | Use class ability |
| `,` | Pick up item |
| `i` | Open inventory |
| `>` | Descend stairs |
| `<` | Ascend stairs |
| `.` | Wait one turn |
| `q` / `Esc` | Quit (with confirmation) |

Inside the **inventory screen**, press the item's number key to use or equip it.

## Classes

Choose one at the start of each run:

| Emoji | Class | HP | ATK | DEF | Passive | Ability (`z`) |
|-------|-------|----|-----|-----|---------|---------------|
| 🧙 | Wandering Arcanist | 30 | 5 | 2 | Wild Magic: 30% chance per kill to restore 2 HP | Dimensional Rift — teleport to a random tile (12t) |
| 💀 | Void Revenant | 15 | 12 | 0 | Each kill restores 3 HP | Death's Bargain — spend 5 HP for +6 ATK 8 turns (15t) |
| 🦾 | Chrono Construct | 60 | 3 | 8 | Self-Repair: +1 HP every 8 turns | Overclock — +6 ATK 6 turns, 2 HP/turn burn (18t) |
| 🌀 | Entropy Dancer | 22 | 9 | 1 | — | Vanish — invisible 8 turns (20t, free per floor) |
| 🔮 | Crystal Oracle | 20 | 3 | 2 | — | Farsight — reveal entire floor (20t, free per floor) |
| 🧬 | Void Symbiont | 42 | 6 | 5 | Symbiotic Regen: +1 HP every 5 turns | Parasite Surge — +10 HP, +4 ATK 6 turns (12t, free per floor) |

Cooldowns shown as `(Nt)`. "Free per floor" means the cooldown resets on each new floor.

## Floors

Each floor has a unique name, tileset, and enemy roster. A floor elite (mini-boss) spawns on every level. The Unmaker ☄️ — the final boss — awaits on floor 10.

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

## MUD server (multiplayer SSH)

N players share a persistent world over SSH with tick-based updates.

```bash
# build and start the server
go build -o emoji-roguelike-server ./cmd/server
./emoji-roguelike-server            # listens on :2222

# connect from any number of terminals
ssh -p 2222 -o StrictHostKeyChecking=no localhost
```

Players spawn in **Emberveil** (Floor 0) — a safe starting city with NPCs, shops, and a healer. Kill enemies to earn gold, then return to the city to spend it. Death respawns you in Emberveil with gold reset.

The server auto-generates an ed25519 host key (`server_host_key`) on first run.

### City NPCs

NPCs follow daily schedules and move around the city. Bump into them to interact:

- **Shopkeepers** — buy equipment with gold
- **Healer** — restore HP
- **Dialogue NPCs** — lore and hints
- **Animals** — ambient flavor

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
