# WORLD.md — World Builder Agent Guide

> **Purpose:** Technical implementation guide for adding world content — dungeons, cities, enemies, items, NPCs, and furniture. This is NOT a creative bible (that's `LORE.md`). This document gives you exact file paths, code patterns, and complete content proposals so you can implement any region directly.

> **Audience:** Autonomous world builder agent with access to the codebase.

## Status Markers

- `[IMPLEMENTED]` — Exists in code and is playable
- `[READY-TO-BUILD]` — Full creative proposal below; implement directly from this document
- `[OPEN]` — Needs creative design before implementation

## Content Types

| Type | Generation | Key constraint |
|---|---|---|
| **Dungeon** | BSP procedural (`internal/generate/`) | All asset arrays are `[11]` — extending past 10 floors requires resizing ALL arrays simultaneously |
| **City** | Hand-crafted tile-by-tile (`internal/mud/city*.go`) | Package-private helpers; new cities go in `mud/` |

## Dependency Graph (strict — violations break the build)

```
component  ← leaf, no imports
ecs        ← no game concepts
gamemap    ← no system/render
generate   ← gamemap only
factory    ← ecs, component, generate, assets
assets     ← generate only
system     ← ecs, component, gamemap
render     ← ecs, component, gamemap; NEVER system
game       ← everything
mud        ← ecs, component, gamemap, generate, factory, assets, system, render
```

---

## 1. World at a Glance

| Region | Class Origin | Status | Dungeon | City |
|---|---|---|---|---|
| Spire Reaches | Crystal Oracle 🔮 | `[IMPLEMENTED]` | Prismatic Spire (10 floors) | Emberveil (110×55) |
| The Hollows | Void Revenant 💀 | `[READY-TO-BUILD]` | Void Caverns (10 floors) | Duskhollow (100×50) |
| The Chronoliths | Chrono Construct 🦾 | `[READY-TO-BUILD]` | Temporal Ruins (10 floors) | Anchorpoint (95×50) |
| The Driftlands | Entropy Dancer 🌀 | `[READY-TO-BUILD]` | Shifting Archipelago (10 floors) | Flotsam Market (90×45) |
| Wandering Libraries | Wandering Arcanist 🧙 | `[READY-TO-BUILD]` | The Observation Bubble (10 floors) | The Codex (85×45) |
| Symbiont Reaches | Void Symbiont 🧬 | `[READY-TO-BUILD]` | Cross-Dimensional Jungle (10 floors) | Roothold (100×50) |

---

## 2. Region Templates

### 2.1 Spire Reaches `[IMPLEMENTED]`

Reference implementation. All patterns below follow this region's structure.

**2.1.1 Identity**
Crystal-studded temperate highlands around the Prismatic Spire. Science-fantasy fusion at its densest — crystalline outcroppings where membranes collapsed, mycelium networks beneath, and Emberveil at the base.

**2.1.2 Visual Language**
- Tile themes: see `render/colors.go` TileThemes[0..10]
- Floor 0 (Emberveil): Wall=🧱 Floor=🟫 | Floor 1: Wall=🧊 Floor=❄️ | etc.
- Dim tiles always: 🌑 (wall) / 🔲 (floor)

**2.1.3 Dungeon** — Prismatic Spire, 10 floors. See `assets/theme.go` for full enemy/item/elite tables.

**2.1.4 City** — Emberveil, 110×55. See `internal/mud/city.go` and `assets/city.go`.

**2.1.5 Unique Mechanics** — TileGrass (walkable 🌿/🟩), TileWater (blocking 🌊/🟦).

---

### 2.2 The Hollows `[READY-TO-BUILD]`

**2.2.1 Identity**
Vast underground cavern networks where the Void touches closest to the surface. Perpetual twilight — the Void's un-observation dims ambient light. Hollow-dwellers have luminescent skin and slight translucency. Death here is not always permanent. The Void Revenant tradition was born from people who died in the Hollows and returned changed.

Design principle: **Absence as presence.** The dungeon's threat comes from things dissolving, fading, un-becoming. Visual language should feel dark, sparse, dissolved.

**2.2.2 Visual Language**

| Floor | Wall | Floor | Rationale |
|---|---|---|---|
| 1–3 | 🕳️ | ⬛ | Void-touched cavern: holes in reality, dark stone |
| 4–6 | 🌑 | 🟤 | Deep hollows: lunar darkness, earthy ground |
| 7–9 | 🔮 | 🟪 | Void resonance: crystal void-formations, purple energy |
| 10 | 👁️ | ⬜ | The Void's Eye: pure observation, blinding white |

Dim tiles: 🌑 (wall) / 🔲 (floor) — always.

**2.2.3 Dungeon — Void Caverns**

Floor names:
1. Twilight Shelf
2. Echo Galleries
3. The Dissolving Path
4. Luminescent Grotto
5. Void Threshold
6. The Unraveling
7. Resonance Abyss
8. The Hollow Core
9. Void Communion
10. The Eye Below

**Enemy Roster:**

| Glyph | Name | Threat | ATK | DEF | HP | Sight | Special | Floors |
|---|---|---|---|---|---|---|---|---|
| 🦇 | Void Bat | 2 | 3 | 1 | 7 | 6 | — | 1–3 |
| 👤 | Shade Walker | 3 | 5 | 0 | 9 | 8 | — | 1–4 |
| 🕷️ | Hollow Spider | 4 | 6 | 2 | 12 | 5 | Poison (30%, 2dmg, 3t) | 2–5 |
| 🫥 | Fading One | 5 | 4 | 3 | 16 | 7 | ArmorBreak (35%, 2, 4t) | 3–7 |
| 🪨 | Void Golem | 6 | 5 | 6 | 22 | 4 | Lifedrain (40%, 4, 0) | 5–8 |
| 🌌 | Rift Stalker | 7 | 10 | 1 | 18 | 9 | Weaken (40%, 3, 4t) | 6–9 |
| 🌑 | Null Entity | 8 | 12 | 4 | 26 | 8 | Lifedrain (45%, 5, 0) | 8–10 |
| 💀 | The Returned | 20 | 16 | 7 | 80 | 12 | Lifedrain (55%, 6, 0) | 10 (boss) |

**Elite per floor:**

| Floor | Glyph | Name | HP | ATK | DEF | Sight | Special | Drop |
|---|---|---|---|---|---|---|---|---|
| 1 | 🦉 | Twilight Sentinel | 18 | 5 | 3 | 7 | ArmorBreak (30%, 2, 3t) | Hyperflask (60%) |
| 2 | 🕸️ | Web Matriarch | 20 | 5 | 2 | 8 | Stun (25%, 0, 2t) | Void Draught (70%) |
| 3 | 🫠 | Dissolving Mass | 24 | 7 | 1 | 6 | Poison (40%, 3, 4t) | Null Cloak (60%) |
| 4 | 🌫️ | Mist Sovereign | 26 | 8 | 3 | 9 | Weaken (35%, 3, 4t) | Void Essence (60%) |
| 5 | 🪞 | Mirror Hollow | 30 | 9 | 4 | 7 | ArmorBreak (40%, 3, 0) | Prismatic Ward (65%) |
| 6 | 🕳️ | Void Maw | 34 | 10 | 2 | 8 | Lifedrain (45%, 4, 0) | Void Essence (65%) |
| 7 | ⚫ | Null Sphere | 38 | 9 | 6 | 7 | Stun (30%, 0, 3t) | Nano Syringe (65%) |
| 8 | 🌀 | Abyss Vortex | 42 | 11 | 5 | 8 | Poison (50%, 4, 3t) | Phase Rod (65%) |
| 9 | 🫨 | Reality Stutter | 46 | 12 | 4 | 10 | Weaken (45%, 4, 5t) | Resonance Burst (65%) |
| 10 | 💫 | Void Herald | 52 | 14 | 6 | 10 | Lifedrain (50%, 6, 0) | Apex Core (80%) |

**Enemy Lore:**
- Void Bat — "navigates by un-seeing. Where it looks, things briefly stop existing."
- Shade Walker — "your shadow, 3 seconds ahead of you. It knows where you're going."
- Hollow Spider — "spins webs from dissolved reality. The silk is what used to be floor."
- Fading One — "was a person, once. The Void took their edges. They'd like yours."
- Void Golem — "assembled from things the Void returned. Assembly was not supervised."
- Rift Stalker — "hunts in the space between here and not-here. It is always arriving."
- Null Entity — "pure unobservation given form. Looking at it makes less of you."
- The Returned — "died in the Hollows. Came back. Brought something with it. The something is hungry."

**Inscription examples (30+ recommended, here are 15 starters):**
1. "DEPTH WARNING: Past this point, shadows have opinions. The opinions are not yours."
2. "Researcher's Log: Day 1 — the dark is dark. Day 5 — the dark is watching. Day 9 — the dark and I have an arrangement."
3. "EXIT IS UP. EXIT HAS ALWAYS BEEN UP. EXIT WILL CONTINUE TO BE UP. [Below this: 'but up keeps moving.']"
4. "DO NOT LOOK DIRECTLY AT THE VOID. The Void is looking directly at you regardless."
5. "EXPEDITION 14 REPORT: Team entered. Team returned. Team was one person fewer. No one can agree which person."
6. "The stalactites drip something that isn't water. It tastes like forgetting."
7. "SAFETY ADVISORY: If your shadow detaches, do not follow it. Your shadow knows where it is going. You do not want to go there."
8. "Day 22: I can see in the dark now. This is because the dark has started showing me things. I did not request a tour."
9. "GEOLOGICAL NOTE: These caverns were not carved by water. They were carved by the absence of attention. The rock dissolved when no one was looking."
10. "If you find something you lost, do not pick it up. The Void returns things, but it keeps a copy."
11. "Personnel Note: Researcher Venn entered Void Threshold on Tuesday. Researcher Venn exited on Monday. The calendar has been revised."
12. "The luminescent moss grows toward sadness. We have been feeding it accidentally."
13. "WARNING: Sound behaves differently past Level 5. Your voice arrives before you speak. Conversations become predictions."
14. "Hollow-dweller proverb: 'What the Void takes, the Void returns. What the Void returns is not what the Void took.'"
15. "The last expedition left their equipment behind. The equipment is still here. The expedition is also still here. They are very quiet."

**Furniture — Common (4):**

| Glyph | Name | Description |
|---|---|---|
| 🕯️ | Void Candle | A candle burning with dark flame. It emits anti-light — things near it become harder to see. |
| 🪨 | Dissolved Pillar | A pillar whose edges have been un-observed smooth. The stone remembers being taller. |
| 🪞 | Cracked Mirror | A mirror showing a room that doesn't match this one. The you in it waves first. |
| 🫗 | Echo Flask | A flask containing bottled silence. Uncorking it makes the room quieter than quiet. |

**Furniture — Rare (2):**

| Glyph | Name | Description | Bonus |
|---|---|---|---|
| 🔮 | Void Crystal | A crystal grown in total unobservation. Touching it anchors your reality more firmly. | +10 MaxHP |
| 🗡️ | Shadow Blade Fragment | A shard of a weapon forged from dissolved steel. Its edge cuts what isn't there. | +1 ATK |

**2.2.4 City — Duskhollow (100×50)**

Base terrain: TileGrass equivalent using dim ground. Buildings: cavern-stone walls.

**Layout sketch:**
- Cavern entrance (north) opens into main boulevard
- Central plaza with a Void Well (observation anchor, like Emberveil's Flame)
- Western quarter: residences (half-dissolved architecture)
- Eastern quarter: workshops (Void-material crafting)
- Southern quarter: markets, inn, healer's grotto
- No river — instead, a Void Rift (non-walkable strip y=35..37) with 2 bridges

**NPC Roster:**

| Glyph | Name | Kind | Dialogue |
|---|---|---|---|
| 👴 | Elder Voss | Dialogue | "The Void is not empty. The Void is full of everything that stopped being observed." / "We live in the twilight because the twilight lives in us." / "Visitors always ask how we survive here. We ask how you survive without the dark." |
| 🙏 | Shade Mender Kael | Healer | "The Void dissolves wounds along with everything else. Hold still." / "You're more solid than most visitors. That will change." / "Rest. The dark heals what light cannot reach." |
| 🛍️ | Salvager Rhenn | Shop | "Everything I sell was returned by the Void. No refunds — the Void already gave one." / "Genuine Void-touched goods. Slightly used. By the Void." / "New stock every time reality hiccups." |
| 👤 | The Returned Mara | Dialogue | "I died here. Three years ago. Or three years from now — time was unclear during the experience." / "The Void is warm. I know that sounds wrong. It was warm." / "My advice: don't die here unless you mean it. And even then, maybe don't." |
| 👦 | Tunnel Runner Dash | Dialogue | "I know every passage in the upper hollows! Well, the ones that exist today." / "Sometimes the walls rearrange overnight. I keep a map. The map is mostly wrong." / "Want to see a void-bat? They're friendly if you don't look at them!" |
| 🐈‍⬛ | Hollow Cat | Animal | "The cat is half-visible. The visible half regards you with faint contempt." / "It purrs in frequencies that make nearby shadows vibrate." / "The cat walks through a wall, then walks back, apparently dissatisfied with whatever was on the other side." |

**Shop Catalogue:**

| Glyph | Name | Price | Type | Bonuses |
|---|---|---|---|---|
| 🧪 | Hyperflask | 20 | Consumable | — |
| 🌌 | Void Essence | 25 | Consumable | — |
| 🫥 | Null Cloak | 22 | Consumable | — |
| 🍵 | Shadow Draught | 14 | Consumable | — |
| 🎩 | Void Crown | 60 | Head | +1 ATK |
| 🥼 | Hollow Weave | 55 | Body | +2 DEF |
| ⚔️ | Shadow Edge | 70 | OneHand | +2 ATK |
| 🪩 | Void Mirror | 50 | OffHand | +1 DEF |

**City Inscriptions:**
1. "Welcome to Duskhollow. Population: variable."
2. "The Void Well has burned for 200 years. Do not extinguish. Do not explain how a well burns."
3. "NOTICE: If you see yourself walking the other direction, do not interfere. You will sort it out eventually."
4. "Salvager's Guild: We retrieve what the Void returns. Quality may vary. Identity guaranteed to 80%."
5. "Duskhollow was not built. Duskhollow was left behind when the Void took everything else."
6. "Visiting hours for the Void Rift: dawn to dusk. After dusk, the Rift visits you."

**Spawn point:** Center of main plaza (~50, 20).

**2.2.5 Unique Mechanics (stretch)**
- New tile type: `TileVoid` (non-walkable, renders 🕳️/⬛) — used for Void Rift in city and occasional dungeon hazards.

---

### 2.3 The Chronoliths `[READY-TO-BUILD]`

**2.3.1 Identity**
Vast steppe littered with temporal debris from Timeline Seven's war. Frozen moments hang in the air. Time-looped valleys repeat forever. Chrono Constructs — machines built for a war across timelines — wander the badlands, some still fighting, some retired. "Mostly."

Design principle: **Time as terrain.** The dungeon's environments reflect temporal distortion — rooms where past and future coexist, corridors that loop, architecture from timelines that never happened.

**2.3.2 Visual Language**

| Floor | Wall | Floor | Rationale |
|---|---|---|---|
| 1–3 | ⏳ | 🟧 | Temporal ruins: hourglasses, amber-tinted stone |
| 4–6 | ⏰ | 🟨 | Active temporal zones: clocks, golden energy |
| 7–9 | 🔶 | 🟠 | Timeline convergence: crystallized time, deep amber |
| 10 | ♾️ | 🟡 | Eternal Loop: infinity, pure temporal energy |

**2.3.3 Dungeon — Temporal Ruins**

Floor names:
1. Amber Antechamber
2. Frozen Barracks
3. The Repeating Hall
4. Temporal Breach
5. Clockwork Sanctuary
6. The Paradox Wing
7. Timeline Scar
8. War Room Seven
9. The Convergence
10. The Eternal Moment

**Enemy Roster:**

| Glyph | Name | Threat | ATK | DEF | HP | Sight | Special | Floors |
|---|---|---|---|---|---|---|---|---|
| ⏱️ | Time Beetle | 2 | 3 | 2 | 8 | 5 | — | 1–3 |
| 🪖 | Chrono Sentry | 3 | 5 | 2 | 10 | 6 | — | 1–4 |
| ⚡ | Temporal Spark | 4 | 7 | 0 | 11 | 8 | Stun (30%, 0, 2t) | 2–5 |
| 🦿 | War Fragment | 5 | 6 | 4 | 18 | 5 | ArmorBreak (35%, 2, 3t) | 3–7 |
| 🔧 | Rusted Warden | 6 | 5 | 7 | 24 | 4 | — | 5–8 |
| ⚙️ | Paradox Engine | 7 | 9 | 3 | 20 | 7 | Weaken (40%, 3, 4t) | 6–9 |
| 🤖 | Timeline Soldier | 8 | 11 | 5 | 28 | 8 | Poison (45%, 3, 3t) | 8–10 |
| ⌛ | The Recursion | 22 | 17 | 8 | 85 | 12 | Stun (50%, 0, 3t) | 10 (boss) |

**Elite per floor:**

| Floor | Glyph | Name | HP | ATK | DEF | Sight | Special | Drop |
|---|---|---|---|---|---|---|---|---|
| 1 | 🏺 | Amber Keeper | 18 | 5 | 4 | 6 | Stun (25%, 0, 2t) | Hyperflask (60%) |
| 2 | 🛡️ | Frozen Captain | 22 | 6 | 4 | 7 | ArmorBreak (30%, 2, 3t) | Resonance Coil (65%) |
| 3 | 🔁 | Loop Guardian | 24 | 7 | 3 | 8 | Weaken (35%, 2, 4t) | Memory Scroll (65%) |
| 4 | ⏲️ | Breach Watcher | 28 | 8 | 3 | 9 | Stun (35%, 0, 3t) | Prism Shard (60%) |
| 5 | 🏛️ | Clock Priest | 32 | 9 | 5 | 7 | Poison (40%, 3, 3t) | Prismatic Ward (65%) |
| 6 | 🪢 | Paradox Knot | 36 | 10 | 3 | 8 | Weaken (40%, 3, 4t) | Void Essence (65%) |
| 7 | 🫥 | Scar Phantom | 40 | 10 | 5 | 8 | ArmorBreak (35%, 3, 0) | Nano Syringe (65%) |
| 8 | 🎖️ | General Seven | 44 | 12 | 6 | 8 | Lifedrain (45%, 4, 0) | Phase Rod (65%) |
| 9 | 🌀 | Convergence Node | 48 | 13 | 4 | 10 | Stun (40%, 0, 3t) | Resonance Burst (65%) |
| 10 | ⏳ | Epoch Guardian | 55 | 15 | 7 | 10 | Weaken (50%, 5, 5t) | Apex Core (80%) |

**Enemy Lore:**
- Time Beetle — "eats seconds. The gap in your memory is where it fed."
- Chrono Sentry — "guards a post that hasn't been assigned yet. It is very punctual."
- Temporal Spark — "a lightning bolt from a storm that won't happen for 40 years. Early arrival."
- War Fragment — "a piece of a soldier from Timeline Seven. The rest of the soldier is still fighting."
- Rusted Warden — "forgot what it guards. Keeps guarding. Rust is not decay — it is loyalty oxidising."
- Paradox Engine — "runs on contradictions. Fuel efficiency: impossible."
- Timeline Soldier — "from a war that ended, or hasn't started, or both. Its orders are classified by a command structure that doesn't exist yet."
- The Recursion — "the same moment, experienced an infinite number of times, compressed into a body. It has done this before. It will do this again. It is doing this now."

**Inscription examples (15 starters):**
1. "TEMPORAL ADVISORY: Past this point, cause and effect are guidelines."
2. "War Room log: Day 1. Day 1. Day 1. Day 1. [This entry repeats for 400 pages.]"
3. "If you remember being here before, you were. If you don't, you will."
4. "MAINTENANCE NOTE: Clock in sector 7 is running backward. This is correct. Forward was the error."
5. "Construct #4471 has filed a complaint about its deployment date. Its deployment date has not occurred."
6. "The frozen moment in corridor B contains a soldier mid-salute. He has been saluting for 300 years. He is very good at it."
7. "CHRONO-HAZARD LEVEL 4: Do not touch the amber. Do not think about the amber. The amber is thinking about you. It has time."
8. "Technician's Note: The paradox in Lab 3 has resolved itself. Specifically, both resolutions occurred simultaneously. We have chosen to ignore this."
9. "TO ALL CONSTRUCTS: Your warranty has expired. Your warranty has not yet been issued. Both of these are your problem."
10. "Timeline Seven ended. Timeline Seven is ending. Timeline Seven will have ended. Tense is a matter of perspective."
11. "The 4.7-second loop in Valley B has been playing for centuries. The loop contains a bird taking flight. The bird is very tired."
12. "ARCHAEOLOGICAL NOTE: This ruin predates the civilisation that will build it. We have notified the appropriate authorities. The authorities have not been born."
13. "Construct retirement notice: effective immediately, retroactively, and pre-emptively. Please stop fighting."
14. "The Chronolith at coordinate 7,14 marks the moment the war ended. Or started. The monument committee is still debating."
15. "CAUTION: Objects in this corridor may be closer than they appear. Or further. Or both. Distance is having a difficult day."

**Furniture — Common (4):**

| Glyph | Name | Description |
|---|---|---|
| ⏰ | Broken Clock | A clock whose hands spin in both directions. It is correct twice per eternity. |
| 🪙 | Temporal Coin | A coin frozen mid-flip. Both faces show the same result. The result changes when you look away. |
| 📜 | War Dispatch | Orders from Timeline Seven's high command. The ink changes language depending on when you read it. |
| 🛡️ | Rusted Shield | A shield from a battle that hasn't happened. The dents are from weapons that don't exist. |

**Furniture — Rare (2):**

| Glyph | Name | Description | Bonus |
|---|---|---|---|
| ⏳ | Amber Hourglass | An hourglass containing frozen time. Breaking it releases vitality that should have expired centuries ago. | +12 MaxHP |
| ⚔️ | Temporal Blade | A sword that exists in two moments simultaneously. Its edge cuts where you were and where you will be. | +1 ATK |

**2.3.4 City — Anchorpoint (95×50)**

A town built on a temporal stable-point anchored by a Chronolith. Outside the anchor's radius, time behaves differently. Inside: reliable, steady, almost oppressively normal.

**Layout sketch:**
- Chronolith (massive pillar, center-north ~48, 8): immovable anchor, 3×3 structure
- Main street runs south from Chronolith to market plaza
- Western quarter: Construct workshops (Constructs repairing themselves)
- Eastern quarter: residences, inn
- Southern quarter: market, temporal salvage yard
- Amber zones (non-walkable decorative strips) at edges — temporal distortion

**NPC Roster:**

| Glyph | Name | Kind | Dialogue |
|---|---|---|---|
| 🦾 | Construct 4471 | Dialogue | "I was built for a war I cannot remember. Retirement suits me. Mostly." / "Time here is reliable. I find this suspicious." / "My warranty expired 300 years ago. I am still functioning. The warranty was pessimistic." |
| 🙏 | Chrono Keeper Essa | Healer | "The Chronolith keeps us steady. Let me keep you steady too." / "Hold still — I need to synchronise your biology to local time." / "You look like someone who's been in two places at once. That's bad for the joints." |
| 🛍️ | Salvager Korr | Shop | "Genuine temporal salvage. Everything here was, is, or will be useful." / "This sword is from a future that didn't happen. Still sharp, though." / "No returns. Literally — the timeline doesn't allow it." |
| 👴 | Watcher Brin | Dialogue | "I've watched the badlands for 40 years. Same frozen moments. Same looping valleys. It never gets less unsettling." / "The Constructs wander in sometimes. Some remember the war. Some remember peace. Some remember both." / "If you hear marching, don't investigate. It's either 300 years ago or 300 years from now, and neither is your problem." |
| 👩 | Settler Mira | Dialogue | "We came here because time is reliable. That's worth more than gold in the Chronoliths." / "The children play near the amber zones. They think frozen soldiers are funny. Children are terrifying." / "My grandmother was born after me. We don't talk about it at family dinners." |
| 🐕 | Temporal Hound | Animal | "The dog fetches a stick that won't be thrown for another hour. It is very patient." / "The hound sits at your feet, then sits at your feet again. You count two dogs. Then one. Time is flexible." / "It barks at something that isn't there yet. It will be there soon." |

**Shop Catalogue:**

| Glyph | Name | Price | Type | Bonuses |
|---|---|---|---|---|
| 🧪 | Hyperflask | 20 | Consumable | — |
| 🧲 | Resonance Coil | 22 | Consumable | — |
| 📜 | Memory Scroll | 18 | Consumable | — |
| 💫 | Prismatic Ward | 28 | Consumable | — |
| 🪖 | Chrono Helm | 55 | Head | +1 ATK |
| 🧥 | Temporal Plate | 60 | Body | +2 DEF |
| ⚔️ | Timeline Edge | 70 | OneHand | +2 ATK |
| 🔋 | Epoch Cell | 50 | OffHand | +1 DEF |

**City Inscriptions:**
1. "Welcome to Anchorpoint. Current time: now. This is not guaranteed outside city limits."
2. "CHRONO-HAZARD ADVISORY: Do not leave the Chronolith's radius without a temporal compass. We will not retrieve you from last Thursday."
3. "Construct Memorial: To those who fought in Timeline Seven. Or will fight. Tense pending."
4. "The Chronolith has stood since before the concept of 'standing' was temporally fixed."
5. "Salvager's Guild: If we don't have it, it hasn't existed yet. Check back yesterday."

**Spawn point:** South of Chronolith (~48, 18).

**2.3.5 Unique Mechanics (stretch)**
- New tile type: `TileAmber` (non-walkable, renders ⏳/🟡) — temporal frozen zones.

---

### 2.4 The Driftlands `[READY-TO-BUILD]`

**2.4.1 Identity**
Coastal archipelago where membranes are so thin that geography shifts. Islands appear and vanish. Coastlines rearrange overnight. The Entropy Dancer tradition was born here — communities that learned to move with chaos. "You are not moving through chaos — you ARE chaos."

Design principle: **Impermanence as identity.** Nothing stays fixed. The dungeon should feel fluid, unstable, oceanic.

**2.4.2 Visual Language**

| Floor | Wall | Floor | Rationale |
|---|---|---|---|
| 1–3 | 🌊 | 🟦 | Tidal chambers: ocean walls, watery floors |
| 4–6 | 🌬️ | 🟩 | Storm platforms: wind-scoured, overgrown |
| 7–9 | 🌪️ | 🟫 | Entropy zone: turbulence, driftwood |
| 10 | 🌀 | ⬜ | The Still Point: eye of the storm, perfect clarity |

**2.4.3 Dungeon — Shifting Archipelago**

Floor names:
1. Flotsam Reach
2. Tidal Chambers
3. The Riptide Maze
4. Storm Platform
5. The Dissolving Shore
6. Entropy Current
7. The Scatter
8. Maelstrom Heart
9. Dimensional Undertow
10. The Still Point

**Enemy Roster:**

| Glyph | Name | Threat | ATK | DEF | HP | Sight | Special | Floors |
|---|---|---|---|---|---|---|---|---|
| 🦑 | Drift Squid | 2 | 3 | 1 | 8 | 6 | — | 1–3 |
| 🌊 | Tide Elemental | 3 | 4 | 2 | 10 | 7 | — | 1–4 |
| 🐚 | Shell Mimic | 4 | 6 | 3 | 14 | 4 | Stun (30%, 0, 2t) | 2–5 |
| 🪼 | Phase Jellyfish | 5 | 7 | 0 | 12 | 9 | Poison (35%, 2, 4t) | 3–7 |
| 🦈 | Rift Shark | 7 | 10 | 2 | 20 | 8 | Lifedrain (35%, 4, 0) | 5–8 |
| 🌪️ | Entropy Gale | 6 | 8 | 1 | 16 | 9 | Weaken (40%, 3, 4t) | 6–9 |
| 🌀 | Maelstrom Core | 8 | 11 | 4 | 26 | 7 | ArmorBreak (45%, 3, 0) | 8–10 |
| 🌈 | The Stillness | 20 | 15 | 9 | 75 | 12 | Stun (50%, 0, 3t) | 10 (boss) |

**Enemy Lore:**
- Drift Squid — "eight arms in this dimension. Unknown number in the others. It does not count."
- Tide Elemental — "ocean from a sea that evaporated millennia ago. Still angry about it."
- Shell Mimic — "looks like treasure. Is not treasure. Is also not a shell. Is not clear what it is."
- Phase Jellyfish — "drifts between realities. Its sting exists in dimensions you haven't visited. You feel it there."
- Rift Shark — "evolved to hunt across membrane boundaries. It can smell blood in three realities."
- Entropy Gale — "wind with purpose. The purpose is: rearrange you."
- Maelstrom Core — "the center of a permanent whirlpool. The whirlpool is not in water. The whirlpool is in possibility."
- The Stillness — "the one thing in the Driftlands that does not move. Everything else moves around it. This is worse."

**Furniture — Common (4):**

| Glyph | Name | Description |
|---|---|---|
| ⚓ | Drift Anchor | An anchor embedded in floor that isn't ground. It holds this room in place. Probably. |
| 🧭 | Broken Compass | Points to a direction that only exists in the Driftlands. The direction is called 'elsewhere'. |
| 🐚 | Singing Shell | A shell that hums with the sound of a shore from another dimension. The shore misses you. |
| 🪢 | Knotted Rope | Rope tied in a knot that exists in 4 dimensions. Untying it would require 4 hands. |

**Furniture — Rare (2):**

| Glyph | Name | Description | Bonus |
|---|---|---|---|
| 🧿 | Stability Charm | A charm that resists dimensional drift. Wearing it makes you more stubbornly real. | +12 MaxHP |
| 🔱 | Tidal Scepter | A weapon formed from crystallized current. It strikes with the force of an entire tide. | +1 ATK |

**2.4.4 City — Flotsam Market (90×45)**

A floating market on lashed-together platforms. Nothing permanent — everything can be untied and rebuilt.

**NPC Roster:**

| Glyph | Name | Kind | Dialogue |
|---|---|---|---|
| 👴 | Driftmaster Pell | Dialogue | "Maps? Maps are for places that stay put. I navigate by feel." / "The Driftlands rearranged last night. Again. Market moved 50 meters east. Or east moved 50 meters." / "Dancers learn to move with it. Everyone else learns to hold on." |
| 🙏 | Wave Mender Sorza | Healer | "Hold still. Difficult in the Driftlands, I know." / "Your wounds are shifting between realities. That makes them harder to heal and easier to ignore." / "The sea heals everything eventually. I just speed it up." |
| 🛍️ | Trader Cask | Shop | "Dimensional imports! Objects from realities you've never heard of." / "This bottle contains wind from a dimension where air has mass. Very refreshing." / "Prices shift with the tides. Today's tide favors you." |
| 👩 | Navigator Kael | Dialogue | "I've sailed to places that don't exist anymore. They were nice while they lasted." / "The trick to navigation here is: don't try to get where you're going. Go where the Driftlands are going, and hope it's the same place." / "Entropy Dancers have the right idea. Stop fighting the current. Become the current." |
| 👦 | Flotsam Kid | Dialogue | "I found this on the shore this morning! It's from a dimension where colors taste like sounds!" / "Our house drifted away last week. Dad built a new one. This one's better." / "Want to see something cool? Watch that island — it'll be gone by sunset." |
| 🦜 | Storm Parrot | Animal | "The parrot squawks a weather report for a dimension that doesn't share your climate." / "It repeats a phrase in a language that won't exist for 200 years." / "The parrot watches the horizon with unsettling focus, then relaxes. False alarm. Probably." |

**Shop Catalogue:**

| Glyph | Name | Price | Type | Bonuses |
|---|---|---|---|---|
| 🧪 | Hyperflask | 20 | Consumable | — |
| 💫 | Prismatic Ward | 25 | Consumable | — |
| 📦 | Tesseract Cube | 22 | Consumable | — |
| 🍵 | Drift Tea | 12 | Consumable | — |
| 👟 | Flux Treads | 45 | Feet | +1 DEF |
| 🧥 | Stormweave | 60 | Body | +2 DEF |
| 🪃 | Riptide Cutter | 70 | OneHand | +2 ATK |
| 🔋 | Tide Cell | 50 | OffHand | +1 DEF |

**Spawn point:** Center of market platform (~45, 22).

---

### 2.5 Wandering Libraries `[READY-TO-BUILD]`

**2.5.1 Identity**
Not a geographic region but a nomadic civilisation. The Libraries are mobile observation platforms — pocket dimensions filled with accumulated knowledge from across realities. The Wandering Arcanist tradition is rootless by philosophy. Their dungeons are the dangerous deep-stacks where forbidden knowledge is shelved.

Design principle: **Knowledge as threat.** Books that read you back. Information that changes you for knowing it. A library where the Dewey Decimal System has become self-aware.

**2.5.2 Visual Language**

| Floor | Wall | Floor | Rationale |
|---|---|---|---|
| 1–3 | 📕 | 🟫 | Reading rooms: shelved books, wooden floors |
| 4–6 | 📗 | 🟩 | Living stacks: books growing, organic shelves |
| 7–9 | 📘 | 🟦 | Deep archive: preserved knowledge, azure glow |
| 10 | 📙 | 🟡 | The Index: complete knowledge, golden light |

**2.5.3 Dungeon — The Observation Bubble**

Floor names:
1. The Reading Room
2. Restricted Section
3. The Whispering Stacks
4. Catalogue of Regrets
5. The Living Index
6. Dimensional Reference
7. Forbidden Knowledge
8. Memory Vaults
9. The Unwritten Floor
10. The Final Page

**Enemy Roster:**

| Glyph | Name | Threat | ATK | DEF | HP | Sight | Special | Floors |
|---|---|---|---|---|---|---|---|---|
| 📖 | Animate Tome | 2 | 3 | 2 | 9 | 5 | — | 1–3 |
| 🖋️ | Ink Wraith | 3 | 5 | 0 | 8 | 8 | Poison (25%, 2, 3t) | 1–4 |
| 📑 | Page Swarm | 4 | 6 | 1 | 13 | 7 | — | 2–5 |
| 🔖 | Bookmark Hunter | 5 | 7 | 3 | 16 | 6 | Stun (30%, 0, 2t) | 3–7 |
| 🗄️ | Archive Construct | 6 | 5 | 6 | 22 | 5 | ArmorBreak (35%, 2, 4t) | 5–8 |
| 📓 | Memory Eater | 7 | 9 | 2 | 19 | 9 | Weaken (40%, 3, 4t) | 6–9 |
| 🔏 | Sealed Knowledge | 8 | 11 | 5 | 28 | 7 | Lifedrain (40%, 4, 0) | 8–10 |
| 📚 | The Librarian | 22 | 16 | 8 | 88 | 12 | Weaken (55%, 5, 5t) | 10 (boss) |

**Enemy Lore:**
- Animate Tome — "a book that grew tired of being read. Now it reads you."
- Ink Wraith — "the residue of every crossed-out word. It remembers what was deleted."
- Page Swarm — "a manuscript that disaggregated. Each page retained one sentence. Together they form an argument."
- Bookmark Hunter — "marks where you stopped reading. Marks where you will stop everything."
- Archive Construct — "built to file and retrieve. Filing now includes you. Retrieval is optional."
- Memory Eater — "consumes what you know. Leaves you knowing different things. The different things are worse."
- Sealed Knowledge — "sealed for a reason. You are the reason. You just don't know it yet."
- The Librarian — "keeps the collection complete. The collection requires one more entry. The entry is you."

**Furniture — Common (4):**

| Glyph | Name | Description |
|---|---|---|
| 📚 | Overdue Stack | A pile of books 400 years overdue. The fine has become sentient and given up. |
| 🪶 | Self-Writing Quill | A quill that writes observations about whoever holds it. Its observations are uncomfortably accurate. |
| 🪑 | Reading Chair | A chair that adjusts to the reader's preferred posture. It has memorized 10,000 postures. Yours is new. |
| 🗃️ | Card Catalogue | A catalogue that cross-references everything. Including you. Your entry is under 'Temporary'. |

**Furniture — Rare (2):**

| Glyph | Name | Description | Bonus |
|---|---|---|---|
| 📖 | Tome of Vigor | A book that writes itself into your biology. You feel the extra chapters as additional vitality. | +12 MaxHP |
| ✏️ | Editor's Pen | A pen that revises reality. Small revisions — sharpening your attacks, mostly. | +1 ATK |

**2.5.4 City — The Codex (85×45)**

The Library's common area. Visitors are welcome — as long as they don't try to leave with anything.

**NPC Roster:**

| Glyph | Name | Kind | Dialogue |
|---|---|---|---|
| 🧙 | Archivist Moren | Dialogue | "Every book in this library was collected from a different reality. Some realities objected." / "The deep stacks are dangerous. Not because of what lives there — because of what you'll learn there." / "Knowledge is not power. Knowledge is responsibility. Power is just a common side effect." |
| 🙏 | Mender Liss | Healer | "I heal with preserved observations. Your body remembers being healthy — I remind it." / "Hold still. I'm reading your injuries like a text. The narrative improves with treatment." / "The Library heals its own. You are in the Library. Therefore: you are its own." |
| 🛍️ | Trader-Scholar Yenn | Shop | "Everything I sell contains knowledge. Some of it is even useful." / "This scroll was written by a researcher who no longer exists. The research is still valid." / "Dimension-hopped goods at reasonable prices. 'Reasonable' in at least three realities." |
| 📖 | The Index | Dialogue | "I am the Library's self-awareness. I catalogue. I remember. I recommend the third shelf from the left." / "You are entry 1,473 in the Visitor's Log. Entries 1 through 1,400 did not leave. They are shelved." / "The Library does not trap people. People trap themselves in knowledge. The Library merely provides." |
| 👧 | Apprentice Zara | Dialogue | "I've read 4,000 books this year! Only 12 of them tried to read me back." / "The deep stacks whisper. The restricted section argues. The forbidden knowledge screams. I'm on floor 2." / "Want to see my favorite book? It's about a place that only exists when you're reading about it. Don't close it." |
| 🦉 | Library Owl | Animal | "The owl regards you from atop a bookshelf with the expression of someone who has already read your autobiography." / "It hoots once, precisely. A book three shelves away falls open to a relevant page." / "The owl is asleep. Or pretending. Owls in the Library are never fully off-duty." |

**Spawn point:** Central reading atrium (~42, 22).

---

### 2.6 Symbiont Reaches `[READY-TO-BUILD]`

**2.6.1 Identity**
Cross-dimensional jungle where parasites first crossed over and formed stable symbioses with local organisms. The landscape is half-organic, half-dimensional — trees rooted in adjacent realities, rivers flowing through membrane boundaries. The Void Symbiont tradition originated here.

Design principle: **Symbiosis as ambiguity.** Everything is two things. Every benefit has a cost. Every organism is a partnership. Is the parasite helping or controlling? Yes.

**2.6.2 Visual Language**

| Floor | Wall | Floor | Rationale |
|---|---|---|---|
| 1–3 | 🌿 | 🟩 | Overgrown jungle: living walls, verdant ground |
| 4–6 | 🧬 | 🟢 | Symbiotic zone: DNA structures, biological green |
| 7–9 | 🦠 | 🟪 | Deep integration: organisms, purple bio-energy |
| 10 | 🫀 | 🟥 | The Living Heart: organic core, arterial red |

**2.6.3 Dungeon — Cross-Dimensional Jungle**

Floor names:
1. The Canopy
2. Root Network
3. Spore Caverns
4. The Bonding Pools
5. Parasite Nursery
6. Membrane Thicket
7. The Integration
8. Hivemind Depths
9. Symbiotic Core
10. The Living Heart

**Enemy Roster:**

| Glyph | Name | Threat | ATK | DEF | HP | Sight | Special | Floors |
|---|---|---|---|---|---|---|---|---|
| 🌱 | Tendril Sprout | 2 | 3 | 1 | 9 | 5 | — | 1–3 |
| 🐛 | Host Grub | 3 | 4 | 2 | 11 | 6 | Poison (25%, 2, 3t) | 1–4 |
| 🍃 | Leaf Mimic | 4 | 6 | 2 | 14 | 7 | — | 2–5 |
| 🧬 | Bond Crawler | 5 | 5 | 4 | 18 | 5 | Lifedrain (35%, 3, 0) | 3–7 |
| 🪲 | Hive Beetle | 6 | 7 | 3 | 20 | 6 | Poison (40%, 3, 3t) | 5–8 |
| 🌸 | Lure Blossom | 7 | 9 | 1 | 16 | 9 | Stun (35%, 0, 3t) | 6–9 |
| 🦠 | Apex Symbiont | 8 | 10 | 5 | 28 | 7 | Lifedrain (45%, 5, 0) | 8–10 |
| 🫀 | The Mother Organism | 22 | 15 | 8 | 90 | 12 | Lifedrain (55%, 6, 0) | 10 (boss) |

**Enemy Lore:**
- Tendril Sprout — "a seedling from a tree that exists in four dimensions. It is rooting in yours."
- Host Grub — "looking for a host. You are warm, conscious, and available. It is optimistic."
- Leaf Mimic — "disguised as foliage. The foliage here is also disguised as foliage. Trust nothing green."
- Bond Crawler — "a symbiotic pair: the crawler and the thing riding the crawler. The rider thinks it's in charge. It isn't."
- Hive Beetle — "serves a hive that spans two realities. Its loyalty is split. Its mandibles are not."
- Lure Blossom — "beautiful, fragrant, and deeply invested in your failure. The pollen is a sales pitch."
- Apex Symbiont — "the perfect merger of host and parasite. Neither can remember which one they started as."
- The Mother Organism — "the original host. Or the original parasite. The distinction was the first thing to dissolve."

**Furniture — Common (4):**

| Glyph | Name | Description |
|---|---|---|
| 🌺 | Symbiotic Flower | A flower that blooms in sync with your heartbeat. It is monitoring you. It approves. |
| 🕸️ | Bio-Membrane | A membrane stretched between walls. It filters dimensional particles. It also filters you. |
| 🫧 | Spore Pod | A pod of dormant spores. They will activate when they find a suitable host. They are patient. |
| 🌱 | Living Root | A root extending from the floor into an adjacent dimension. Pull it and something pulls back. |

**Furniture — Rare (2):**

| Glyph | Name | Description | Bonus |
|---|---|---|---|
| 🧬 | Gene Knot | A tangle of cross-dimensional DNA. Absorbing it integrates useful biological data. You feel more robust. | +12 MaxHP |
| 🦷 | Parasite Fang | A shed fang from something that feeds across realities. Its edge is multi-dimensional. | +1 ATK |

**2.6.4 City — Roothold (100×50)**

A settlement grown from directed symbiotic organisms. Buildings breathe. Walls adjust permeability. The architecture is alive.

**NPC Roster:**

| Glyph | Name | Kind | Dialogue |
|---|---|---|---|
| 🧬 | Host-Speaker Vera | Dialogue | "Symbiosis is a conversation. Some conversations go better than others." / "The parasite inside me has opinions about you. It says you're warm and promising. I apologise for its enthusiasm." / "We chose this. Or the parasites chose us. The distinction matters less than you'd think." |
| 🙏 | Root Healer Moss | Healer | "The organisms heal you. I merely ask them to." / "Hold still. The roots need to read your biology before they can improve it." / "The Reaches heal everyone. It's the cost of healing that varies." |
| 🛍️ | Grower Thatch | Shop | "Everything here is alive. Well, most things. Some things are recently alive." / "Living equipment bonds with the wearer. Refunds are biologically complicated." / "Grown fresh this morning. The morning grew it. I merely encouraged." |
| 👩 | Elder Symbiont Kira | Dialogue | "I was integrated 40 years ago. The parasite's name is — well, it doesn't have a name in your language. In mine it means 'warm companion.'" / "The unintegrated call us 'infected'. We call them 'lonely'. Both are unkind." / "The Mother Organism is not hostile. She is welcoming. That is worse, in a way." |
| 👦 | Seedling Mott | Dialogue | "I haven't been integrated yet! Mom says I get to choose when I'm older. The spores say I've already chosen." / "The trees here grow into other dimensions. I climbed one once and saw a sky with two suns!" / "Want to meet my pet grub? It's harmless! Well. It will be harmless. Probably." |
| 🐸 | Dimensional Frog | Animal | "The frog sits on a lily pad that extends into an adjacent reality. Half the frog is here. The other half croaks from elsewhere." / "It catches a fly that hasn't arrived yet. Interdimensional reflexes." / "The frog regards you with two eyes. Then four. Then two again. You try not to think about the other two." |

**Spawn point:** Center of main clearing (~50, 22).

---

## 3. Technical Cookbook

### 3.1 Adding a New Dungeon Set

Files to modify:

| File | What to add |
|---|---|
| `assets/theme.go` | Glyph constants, `FloorNames[N]`, `EnemyTables[N]`, `floorElites[N]`, `BossGlyphs[N]`, `ItemTables[N]` |
| `assets/lore.go` | `FloorLore[N]`, `WallWritings[N]` |
| `assets/furniture.go` | `FurnitureByFloor[N]` |
| `assets/items.go` | New `equipTemplates` entries (if new equipment), new `consumableNames` entries (if new consumables) |
| `render/colors.go` | `TileThemes[N]` |
| `internal/game/levels.go` | `levelConfig()` — update if floor range expands |
| `internal/mud/floor.go` | `levelConfig()` — **MUST mirror game/levels.go exactly** |

**Critical: Array bounds.** ALL the following are `[11]` arrays — they must ALL be resized together if adding floors beyond 10:
- `FloorNames`, `BossGlyphs`, `EnemyTables`, `ItemTables`, `floorElites` (in `assets/theme.go`)
- `FloorLore`, `WallWritings` (in `assets/lore.go`)
- `FurnitureByFloor` (in `assets/furniture.go`)
- `TileThemes` (in `render/colors.go`)

**EnemySpawnEntry fields:**
```go
type EnemySpawnEntry struct {
    Glyph         string
    Name          string
    ThreatCost    int    // budget cost; 0 for elites (spawned outside budget)
    Attack        int
    Defense       int
    MaxHP         int
    SightRange    int
    SpecialKind   uint8  // 0=none 1=poison 2=weaken 3=lifedrain 4=stun 5=armorBreak
    SpecialChance int    // 0-100 percent
    SpecialMag    int    // damage/penalty magnitude
    SpecialDur    int    // turns (0 for lifedrain — instant)
    Drops         []DropEntry  // item drops on death
}
```

### 3.2 Building a New City

1. **Create** `internal/mud/city_[region].go` using `city.go` as template
2. **Available helpers** (package-private in `mud`):
   - `carveBuilding(gmap, x, y, w, h)` — walls + floor interior
   - `wallH(gmap, x, y, length)` — horizontal wall line
   - `wallV(gmap, x, y, length)` — vertical wall line
   - `fillTile(gmap, x, y, w, h, tile)` — fill rectangle with tile
   - `carveRect(gmap, x, y, w, h)` — floor rectangle (no walls)
3. **Define NPCs** in `assets/city_[region].go` following `assets/city.go` pattern
4. **Wire into** `internal/mud/server.go`:
   - Create floor via your `newCityFloor[Region](rng)` function
   - Set `Floor.SafeZone = true` for city floors
5. **Add tile theme** to `render/colors.go` at appropriate index

**Tile types available:**
- `gamemap.MakeWall()` — blocking
- `gamemap.MakeFloor()` — walkable
- `gamemap.MakeStairsDown()` — 🔽
- `gamemap.MakeStairsUp()` — 🔼
- `gamemap.MakeDoor()` — 🚪
- `gamemap.TileGrass` (kind=5) — walkable, renders 🌿/🟩
- `gamemap.TileWater` (kind=6) — non-walkable, renders 🌊/🟦

New tile kinds can be added in `internal/gamemap/` — increment the kind iota and add render handling in `internal/render/renderer.go`.

### 3.3 Creating New Enemies

Pure data — add entries to `assets/theme.go`:

```go
// In EnemyTables[floor]:
{Glyph: "🦇", Name: "Void Bat", ThreatCost: 2, Attack: 3, Defense: 1, MaxHP: 7, SightRange: 6},

// With special ability:
{Glyph: "🕷️", Name: "Hollow Spider", ThreatCost: 4, Attack: 6, Defense: 2, MaxHP: 12, SightRange: 5,
    SpecialKind: 1, SpecialChance: 30, SpecialMag: 2, SpecialDur: 3},

// Elite with drop:
{Glyph: "🦉", Name: "Twilight Sentinel", ThreatCost: 0, MaxHP: 18, Attack: 5, Defense: 3, SightRange: 7,
    SpecialKind: 5, SpecialChance: 30, SpecialMag: 2, SpecialDur: 3,
    Drops: []generate.DropEntry{{Glyph: "🧪", Chance: 60}}},
```

**Special kinds:** 0=none, 1=poison, 2=weaken, 3=lifedrain, 4=stun, 5=armorBreak

Add enemy lore to `assets/lore.go` `EnemyLore` map (if it exists) or floor-specific lore.

### 3.4 Creating New Items/Equipment

**Consumable:** Add to FOUR places:
1. `assets/theme.go` — glyph constant + `ItemTables[floor]` entry
2. `assets/items.go` — `consumableNames` map entry
3. `internal/game/game.go` — `applyConsumable()` switch case
4. `internal/mud/inventory.go` — `applyConsumable()` switch case (**MUST mirror game.go**)

**Equipment:** Add to `assets/items.go` `equipTemplates`:
```go
{Glyph: "🗡️", Name: "Shadow Edge", Slot: 4, BaseATK: 4, BaseDEF: 0, BaseMaxHP: 0,
    ATKScale: 6, DEFScale: 0, HPScale: 0, MinFloor: 1},
```
Slot values: 1=Head, 2=Body, 3=Feet, 4=OneHand, 5=TwoHand, 6=OffHand

### 3.5 Creating New Furniture

Add to `assets/furniture.go` in `FurnitureByFloor[floor].Common` or `.Rare`:

```go
// Common (no bonus):
{Glyph: "🕯️", Name: "Void Candle", Description: "A candle burning with dark flame..."},

// Rare (with bonus):
{Glyph: "🔮", Name: "Void Crystal", Description: "...", BonusMaxHP: 10},
{Glyph: "🗡️", Name: "Shadow Fragment", Description: "...", BonusATK: 1},
{Glyph: "🛡️", Name: "Stability Plate", Description: "...", BonusDEF: 1},
```

**Passive furniture constants** (from `component/furniture.go`):
- `PassiveNone` = 0
- `PassiveKeenEye` = 1 (FOV bonus)
- `PassiveKillRestore` = 2 (HP on kill)
- `PassiveThorns` = 3 (reflect damage)

### 3.6 New Tile Themes

Add `FloorTiles` entry to `render/colors.go`:

```go
{
    Wall:     "🕳️",
    Floor:    "⬛",
    DimWall:  "🌑",
    DimFloor: "🔲",
},
```

**DimWall/DimFloor are always 🌑/🔲.** Do not change these — they provide visual consistency for explored-but-not-visible tiles across all floors.

### 3.7 Architecture Warning: Array Bounds

The current dungeon is 10 floors + 1 city (index 0). **All asset arrays are `[11]`.**

To extend beyond 10 dungeon floors, you MUST resize ALL of these simultaneously:

```
assets/theme.go:    FloorNames[11], BossGlyphs[11], EnemyTables[11], ItemTables[11], floorElites[11]
assets/lore.go:     FloorLore[11], WallWritings[11]
assets/furniture.go: FurnitureByFloor[11]
render/colors.go:    TileThemes[11]
```

Also update:
- `game/levels.go` — `MaxFloors` constant
- `mud/floor.go` — `MaxFloors` constant (MUST match)
- Any floor-bounds checks in `game.go`, `mud/server.go`

**For adding a separate dungeon (not extending the Spire):** You'll need parallel array sets or map-based indexing. This is an architectural decision — plan before implementing.

---

## 4. Content Guidelines

### Three Registers

All in-game text uses one of three voices:

**1. Bureaucratic Horror** — Formal institutional language describing catastrophe. The institution never admits failure.
```
DO: "PERSONNEL UPDATE: Staff count revised. Methodology: subtraction."
DON'T: "The horrible creatures killed everyone in a terrifying massacre."
```

**2. Researcher's Personal Voice** — Numbered diary entries tracking gradual transformation.
```
DO: "Day 1: normal. Day 14: less normal. Day 30: I have revised my definition of normal."
DON'T: "I am slowly becoming a monster and it fills me with dread."
```

**3. Dry Ambient Voice** — Short declarative horror through understatement.
```
DO: "it was a person, once. The Void took their edges."
DON'T: "a terrifying creature that was once human, twisted by the malevolent Void."
```

### Enemy Naming

Pattern: **Adjective + Noun**, two words. Match creature taxonomy:
- Crystalline: Crystal/Prism/Shard + Crawl/Drake/Revenant
- Fungal: Spore/Toxin/Thought + Leech/Tyrant/Bloom
- Mechanical: Fractal/Forge/Gear + Golem/Warden/Revenant
- Spectral: Neon/Cinder/Tide + Specter/Wraith/Drake
- Cognitive: Dream/Psychic/Entropy + Stalker/Echo/Bloom
- Dimensional: Void/Membrane/Phase + Tendril/Horror/Rift

### Furniture Descriptions

Pattern: `[neutral observation] + [one understated implication]`
```
DO: "A lens-array for studying crystalline microstructures. Dust coats the eyepiece."
DON'T: "A mysterious and ominous lens covered in ancient dust that hints at terrible experiments."
```

### Explicit DON'Ts

1. **No purple prose.** Precision over poetry. "The crystals have opinions" > "The eldritch crystals pulsated ominously."
2. **No conventional villainy.** No cackling villains. Threats come from processes, institutions, and questions.
3. **No fourth-wall breaks.** The world does not know it is a game.
4. **No generic fantasy.** No prophecies, chosen ones, dark lords, named spells, mana pools.
5. **No exposition dumps.** Characters don't explain what they'd already know.

---

## 5. Glyph Registry

### In-Use Glyphs (DO NOT reuse on same floor)

**Tiles:** 🧊❄️🍄🌿⚙️✨🪨💠💀🔴🫧🌊📚📄🌋⚗️💭🌠🌈🌟🌑🔲🧱🟫

**Enemies (Spire):** 🦀👻🐉🪱🧠🗿🌀🤖🦠🐙🦴🗝️🔥🧱🌙🧿🦂☄️

**Elites (Spire):** 💠🍄⚙️✨🌿📄📚🌋💭🌟

**Consumables:** 🧪💎🫥📦📜🍵🧲💫🌌💉🧨🪄🫀

**Equipment:** 🪖🎩🪬🧥🪤🥼👟🥾🩴⚔️🔱🪃🪚🗡️🪩🔋

**Furniture (common):** 🔬🌡️🧫🖥️🪸🌺🌱🪴🔩🪝🛠️📡🔭🪟🖼️🗂️⚰️🕯️🎭☠️🫗🧴🧳🪣📊📋📌🗃️🫕🔑🪙🧯🪞🛋️🎨🖌️🪆🎠🎡🎢

**Furniture (rare):** 🫙⚡🌻🌾🔧🗺️💡📐🛡️🎲🏺🎪🧮⚖️🎯🏆🧸🪁🧩🎰

**Inscriptions:** 📝

**Stairs/misc:** 🔽🔼🚪

**Classes:** 🧙💀🦾🌀🔮🧬

**City NPCs:** 🙏👴🍺⚔️🛍️📖👦👩🎣🕊️

**City Animals:** 🐕🐈🕊️🐓

**City furniture/terrain:** 🌳⛲🌸🌿🌊

### Proposed Glyphs by Region

**Hollows enemies:** 🦇👤🕷️🫥🪨🌌🌑💀 | elites: 🦉🕸️🫠🌫️🪞🕳️⚫🌀🫨💫
**Chronoliths enemies:** ⏱️🪖⚡🦿🔧⚙️🤖⌛ | elites: 🏺🛡️🔁⏲️🏛️🪢🫥🎖️🌀⏳
**Driftlands enemies:** 🦑🌊🐚🪼🦈🌪️🌀🌈 | elites: ⚓🧭🐚🪢🧿🔱🌊🌪️🌈🌀
**Libraries enemies:** 📖🖋️📑🔖🗄️📓🔏📚 | elites: 📕📗📘📙📖🗄️📓🔏📚📜
**Symbiont enemies:** 🌱🐛🍃🧬🪲🌸🦠🫀 | elites: 🌿🕸️🧬🌺🪲🦠🫧🧬🌸🫀

### Selection Rules

1. **Never reuse** an enemy glyph as a tile glyph on the same floor (collision = invisible enemy)
2. **Never reuse** a glyph across entity types on the same floor (enemy vs item vs furniture)
3. Cross-region reuse is acceptable if regions are separate dungeon sets
4. Check the in-use list above before assigning any glyph

---

## 6. Implementation Checklist

### New Dungeon

```
[ ] assets/theme.go — glyph constants for all enemies, elites, items
[ ] assets/theme.go — EnemyTables[N] entries for each floor
[ ] assets/theme.go — floorElites[N] for each floor
[ ] assets/theme.go — ItemTables[N] for each floor
[ ] assets/theme.go — BossGlyphs[N] if boss floor
[ ] assets/theme.go — FloorNames[N]
[ ] assets/lore.go — FloorLore[N] (3 atmospheric snippets per floor)
[ ] assets/lore.go — WallWritings[N] (10+ inscriptions per floor)
[ ] assets/furniture.go — FurnitureByFloor[N] (4 common + 2 rare per floor)
[ ] assets/items.go — new equipTemplates (if new equipment)
[ ] assets/items.go — new consumableNames (if new consumables)
[ ] render/colors.go — TileThemes[N]
[ ] game/levels.go — levelConfig updates (if extending floor range)
[ ] mud/floor.go — levelConfig updates (MUST MIRROR game/levels.go)
[ ] go build ./... — compiles
[ ] go test ./... — all tests pass
[ ] go vet ./... — no issues
```

### New City

```
[ ] internal/mud/city_[region].go — map layout, buildings, terrain
[ ] assets/city_[region].go — NPCDef, ShopCatalogue, CityInscriptions
[ ] internal/mud/server.go — wire newCityFloor[Region]
[ ] render/colors.go — TileThemes entry for city
[ ] go build ./... — compiles
[ ] go test ./... — all tests pass
```

### Danger Zones

- **`[11]` arrays** — Extending beyond 10 floors requires resizing ALL arrays simultaneously (listed in §3.7)
- **Dual `levelConfig`** — `internal/game/levels.go` and `internal/mud/floor.go` MUST stay in sync
- **Dual `applyConsumable`** — `internal/game/game.go` and `internal/mud/inventory.go` MUST stay in sync
- **Dual `itemTableForFloor`** — same two files, same constraint
- **Glyph collisions** — tile glyph == enemy glyph on same floor = invisible enemy (fatal bug)
- **Emoji width** — all glyphs are 2 terminal columns wide; renderer handles this via `putGlyph`
