# NPC.md — NPC Creation Agent Guide

> This document tells an autonomous NPC creation agent **how to create characters** that fit both the world lore and the codebase. It bridges LORE.md (who inhabits these places, metaphysics, tone) and WORLD.md (technical implementation, code patterns, file paths).

**Required reading before using this guide:**
- `LORE.md` — world bible, tonal registers, regional cultures, naming conventions
- `WORLD.md` — technical cookbook, region templates, glyph registry
- `CLAUDE.md` — build commands, dependency rules, testing requirements

**Status markers:**
- **[CANON]** — Exists in code (`assets/city.go`, `internal/component/npc.go`). Cannot be contradicted.
- **[DESIGNED]** — Defined in WORLD.md NPC rosters. Treat as accepted designs awaiting implementation.
- **[TEMPLATE]** — Structural patterns for creating new content.
- **[OPEN]** — Creative territory. Invent freely within established constraints.

---

## 0. Design Philosophy

1. **NPCs are residents, not game furniture.** They have lives that predate the player. Townsfolk Maren waits for a husband who entered the tower six years ago. Old Fisher Bram fished before the tower was famous. Every NPC should imply a life beyond the interaction.

2. **Observation defines identity.** This world's physics run on observation — the act of seeing shapes reality. Each NPC is shaped by their relationship to this. Hollow-dwellers glow because the twilight zone's bacteria thrive on them. Driftlanders navigate by subjective reference points. Even "normal" Emberveil citizens live under the Eternal Flame's observation field.

3. **No NPCs exist in isolation.** Every character connects to other NPCs, institutions, and regional culture. Father Brennan maintains the Flame. Scholar Alaric studies the tower. Soldier Greta has seen the interior three times. These relationships form a web — new NPCs should attach to it.

4. **The four kinds are mechanical, not narrative.** `NPCKindDialogue`, `NPCKindHealer`, `NPCKindShop`, `NPCKindAnimal` define interaction behavior, not character depth. A Healer can be terrifying. A Shop NPC can be philosophical. An Animal can imply dimensional awareness. Depth comes from dialogue, not handler type.

5. **Regional voice is non-negotiable.** Each region has a distinct linguistic personality. Emberveil NPCs speak with direct, practical warmth. Duskhollow NPCs use negative constructions ("The Void is not empty"). Anchorpoint NPCs confuse tenses. These patterns are identity, not decoration.

---

## 1. NPC Kinds: Mechanical Reference

| NPCKind | Const | Interaction | Speech Format | Canonical Example |
|---|---|---|---|---|
| `NPCKindDialogue` | 0 | Shows random line from pool on bump | `"💬 Name: \"line\""` | Father Brennan, Scholar Alaric |
| `NPCKindHealer` | 1 | Heals player to full HP; shows dialogue if already full | `"✨ Name heals you to full!"` or `"💬 Name: \"line\""` | Sister Maris |
| `NPCKindShop` | 2 | Opens modal shop screen (RunShop) | Shop UI renders separately | Merchant Yeva |
| `NPCKindAnimal` | 3 | Shows random line from pool — **no speech marks** | Raw narration text | Stray Dog, Town Cat, Pigeon |

**Key behaviors:**
- All kinds are bump-to-interact (player walks into the NPC)
- NPCs have `TagBlocking` — they obstruct movement
- NPC glyphs render in cyan (`tcell.ColorAqua`) at render order 5
- A random line from `Lines []string` is selected each interaction
- Expanding to new NPCKind values requires code changes in `component/npc.go`, `mud/server.go` (`interactNPCLocked`), and test coverage

---

## 2. NPC Archetypes

Eight character archetypes that cross-cut mechanical kinds. An NPC's archetype determines their voice, not their handler type.

### The Authority
Guards, officials, institutional voices. Speaks with the confidence of someone who has filed the correct forms.

**Disposition:** Calm, procedural, concerned about protocol more than safety.
**Canonical examples:** Soldier Greta (Emberveil), Captain Holt (mentioned in inscriptions).
**Dialogue pattern:** Declarative statements based on experience. Rules stated as fact. Quiet competence with one line that reveals what they've actually seen.
**Tone register:** Bureaucratic Horror lite — institutional language but from someone who genuinely cares.

> DO: "First rule of the tower: keep the stairs behind you."
> DON'T: "Hark! The tower is a place of great danger! Beware, adventurer!"

### The Keeper
Flame-keepers, well-tenders, those who maintain something vital. Devoted, concerned, slightly tired.

**Disposition:** Protective, ritualistic, aware that what they guard is larger than them.
**Canonical examples:** Father Brennan, Sister Maris, Sister Lena (Emberveil); Shade Mender Kael (Duskhollow [DESIGNED]).
**Dialogue pattern:** Statements of faith/duty. References to the thing they keep. Occasional weariness about how long they've been keeping it.
**Tone register:** Dry Ambient Voice — short, certain, understated.

> DO: "The Flame burns for everyone, even those in the dark."
> DON'T: "I am the sacred keeper of the Eternal Flame, and it is my solemn duty to..."

### The Observer
Scholars, researchers, those who study. Academic interest shading into academic unease.

**Disposition:** Curious, precise, gradually realizing that what they study is studying them back.
**Canonical examples:** Scholar Alaric (Emberveil); The Index (Codex [DESIGNED]); Archivist Moren (Codex [DESIGNED]).
**Dialogue pattern:** Numbered observations. Precise measurements of impossible things. One line where professional detachment cracks.
**Tone register:** Researcher's Personal Voice — objective → intimate → eerily accepting.

> DO: "I've catalogued seventeen distinct energy signatures from the lower floors. Fascinating and terrifying."
> DON'T: "My research into the arcane mysteries has revealed that..."

### The Survivor
Veterans, the returned, people who went somewhere and came back different. Short declarative statements.

**Disposition:** Quiet, economical with words, uncomfortable with questions about what happened.
**Canonical examples:** Soldier Greta (Emberveil, also Authority); The Returned Mara (Duskhollow [DESIGNED]).
**Dialogue pattern:** Experience stated as advice. No elaboration. The horror is in what they don't say.
**Tone register:** Dry Ambient Voice — maximum understatement.

> DO: "I've escorted three expeditions into that place. I'm the only one who came back all three times."
> DON'T: "I have survived many harrowing journeys into the depths of the tower."

### The Innocent
Children, newcomers, those who find the terrifying normal because they grew up with it.

**Disposition:** Enthusiastic, curious, casually mentioning things that should be alarming.
**Canonical examples:** Street Urchin Pip (Emberveil); Tunnel Runner Dash (Duskhollow [DESIGNED]); Flotsam Kid (Flotsam [DESIGNED]); Seedling Mott (Roothold [DESIGNED]); Apprentice Zara (Codex [DESIGNED]).
**Dialogue pattern:** Excitement about impossible things. Casual acceptance of dimensional horror as normal. One line requesting something from the player.
**Tone register:** Comedy Patterns — mundane reactions to impossible situations.

> DO: "I bet there's treasure in there. LOADS of treasure. You should bring me some."
> DON'T: "Oh wise hero, please bring me treasures from the dungeon!"

### The Merchant
Traders — transactional worldbuilding. Their goods tell you about the world.

**Disposition:** Pragmatic, cheerful, treating dimensional horror as a business opportunity.
**Canonical examples:** Merchant Yeva (Emberveil); Salvager Rhenn (Duskhollow [DESIGNED]); Trader Cask (Flotsam [DESIGNED]); Salvager Korr (Anchorpoint [DESIGNED]); Trader-Scholar Yenn (Codex [DESIGNED]); Grower Thatch (Roothold [DESIGNED]).
**Dialogue pattern:** Product descriptions that reveal worldbuilding. Casual references to where goods come from. One line about refund policy that implies dimensional complications.
**Tone register:** Comedy Patterns — sales patter applied to impossible goods.

> DO: "Everything I sell was returned by the Void. No refunds — the Void already gave one."
> DON'T: "Welcome to my shop, brave adventurer! I have wares if you have coin!"

### The Animal
Non-human entities — narrated in third person, implying intelligence without confirming it.

**Disposition:** Varies from dignified contempt (cats) to enthusiastic idiocy (dogs) to bureaucratic indifference (pigeons).
**Canonical examples:** Stray Dog, Town Cat, Pigeon, Market Hen (Emberveil); Hollow Cat (Duskhollow [DESIGNED]); Temporal Hound (Anchorpoint [DESIGNED]); Storm Parrot (Flotsam [DESIGNED]); Library Owl (Codex [DESIGNED]); Dimensional Frog (Roothold [DESIGNED]).
**Dialogue pattern:** Third-person present tense. Attribute opinions to the animal. Reference dimensional phenomena casually.
**Tone register:** Comedy Patterns — animals behaving with exaggerated dignity and opinion.

> DO: "The cat stares at you with the absolute conviction that you are beneath its notice."
> DON'T: "The cat says: 'Meow! I am a magical cat!'"

### The Changed
Transformed by their environment — matter-of-fact about alteration. Not tragic, not triumphant.

**Disposition:** Calm, accepting, occasionally forgetting that their condition is unusual.
**Canonical examples:** The Returned Mara (Duskhollow [DESIGNED]); Construct 4471 (Anchorpoint [DESIGNED]); Elder Symbiont Kira (Roothold [DESIGNED]); Host-Speaker Vera (Roothold [DESIGNED]).
**Dialogue pattern:** State the change as fact. No self-pity. One line where they compare their experience to the player's with dry humor.
**Tone register:** Dry Ambient Voice — the horror is in the acceptance.

> DO: "I died here. Three years ago. Or three years from now — time was unclear during the experience."
> DON'T: "Alas, I was cursed by dark magic and transformed into this wretched state!"

---

## 3. Dialogue Writing Rules

### Line Count and Selection
- **3–4 lines per NPC**, each self-contained. The engine selects one randomly on each bump.
- Each line must work independently — players may never see the others.
- Lines should not form a sequence or narrative arc. Each is a standalone window into the NPC's character.
- Exception: Shop NPCs need only 2–3 lines (the shop UI does the heavy lifting).

### Tonal Registers (from LORE.md §11)

Map each NPC to one register. Mix registers within a city, not within a character.

**1. Bureaucratic Horror** — Formal institutional language describing catastrophe.
- Use for: Authorities, institutional keepers, official notices
- Key traits: ALL-CAPS headers, polite language around horror, policy documents for impossible situations
- Example: "NOTICE: Citizens are advised to avoid the tower entrance after sundown. — Captain Holt" [CANON]

**2. Researcher's Personal Voice** — Numbered diary entries tracking gradual transformation.
- Use for: Observers, scholars, the slowly changed
- Key traits: Day/entry numbering, shift from "I study X" to "X and I" to "we", final entries that are calm
- Example: "I've catalogued seventeen distinct energy signatures from the lower floors." [CANON]

**3. Dry Ambient Voice** — Short declarative horror through understatement.
- Use for: Survivors, keepers, animals, the matter-of-factly changed
- Key traits: 1–2 sentences, present tense, understatement, precise word choice, no adjectives of scale
- Example: "Things come up sometimes. Not just adventurers. Things." [CANON]

### Five DON'Ts

1. **No exposition dumps.** Characters do not explain things they would already know. Scholar Alaric notes energy signatures with academic interest — he doesn't explain what the Spire is.

2. **No quest language.** No "brave adventurer," no "your quest awaits," no "I have a task for you." NPCs are not quest-givers. They are people who happen to be standing near a player.

3. **No fourth-wall breaks.** The world does not know it is a game. No winking at the audience. No references to game mechanics.

4. **No purple prose.** "The crystals have started humming" — not "The eldritch crystals emanated a haunting resonance that chilled the very marrow of existence."

5. **No generic fantasy.** No ancient prophecies, no dark lords, no "chosen one" language. This world runs on observation physics, not inherited fantasy tropes.

### Regional Dialect Guidelines

Each region has a linguistic personality. When writing NPCs for a region, follow its dialect rules.

**Emberveil (Spire Reaches)** — Direct, practical, warm but wary. Statements based on lived proximity to the tower.
- Pattern: Simple declarative sentences. Personal observations. Practical advice.
- Example: "Word of advice: if something glows, don't touch it. Unless it's the ale." [CANON]

**Duskhollow (The Hollows)** — Negative constructions. Defining things by what they are not.
- Pattern: "The Void is not X" / "X is not what X was" / double negatives that feel natural.
- Example: "The Void is not empty. The Void is full of everything that stopped being observed." [DESIGNED]

**Anchorpoint (The Chronoliths)** — Tense confusion. Past, present, and future mixed casually.
- Pattern: "X was/is/will be" in the same breath. Temporal words used as spatial terms.
- Example: "My grandmother was born after me. We don't talk about it at family dinners." [DESIGNED]

**Flotsam Market (The Driftlands)** — Impermanence as given. Nothing stays. Movement metaphors.
- Pattern: "X moved/shifted/drifted" instead of "X changed." Geography as weather.
- Example: "Our house drifted away last week. Dad built a new one. This one's better." [DESIGNED]

**The Codex (Wandering Libraries)** — Knowledge metaphors. Cataloguing as social behavior.
- Pattern: "You are entry N in X" / "This was written by" / information as physical object.
- Example: "Every book in this library was collected from a different reality. Some realities objected." [DESIGNED]

**Roothold (Symbiont Reaches)** — Mixing "I" and "we." Boundaries between self and other blurred.
- Pattern: "I/we" used interchangeably. The parasite mentioned as conversational partner.
- Example: "The parasite inside me has opinions about you. It says you're warm and promising." [DESIGNED]

### Animal Narration Patterns
- Always third-person present tense: "The cat stares," "It blinks slowly," "The dog rolls onto its back."
- Attribute opinions, preferences, and judgments to the animal.
- Reference dimensional phenomena casually — the animal is not surprised by them.
- No speech marks (handled by `NPCKindAnimal` — raw text, no `💬` prefix).

---

## 4. NPC Ecosystems by Region

### Minimum Viable City
Every city requires at minimum:
- 1 Healer (NPCKindHealer)
- 1 Shop (NPCKindShop)
- 3+ Dialogue NPCs (NPCKindDialogue)
- 1+ Animal (NPCKindAnimal)

Aim for 5–8 named human NPCs and 2–4 animals per city. Emberveil has 10 humans + 4 animals (14 total) as the high-water mark.

### Emberveil [CANON]

All 14 NPCs exist in `assets/city.go`. Cannot be contradicted.

| Name | Kind | Archetype | Key Relationship |
|---|---|---|---|
| Sister Maris | Healer | Keeper | Church of the Eternal Flame |
| Father Brennan | Dialogue | Keeper | Maintains the Flame that keeps the city stable |
| Ol' Rudwig | Dialogue | Survivor | Tavern owner, dispenses advice as alcohol |
| Soldier Greta | Dialogue | Authority/Survivor | Three expeditions survived; patrols the square |
| Merchant Yeva | Shop | Merchant | Profits from the adventurer trade |
| Scholar Alaric | Dialogue | Observer | Studies the Spire from academic distance |
| Street Urchin Pip | Dialogue | Innocent | Sneaks near the tower after dark |
| Townsfolk Maren | Dialogue | Keeper | Waits for husband who entered 6 years ago |
| Old Fisher Bram | Dialogue | Observer | Fishes the channels; knows things come up |
| Sister Lena | Dialogue | Keeper | Church, tends the vestry |
| Stray Dog | Animal | Animal | Wanders the town square |
| Town Cat | Animal | Animal | Prowls the streets |
| Pigeon | Animal | Animal | Multiple instances in the square |
| Market Hen | Animal | Animal | Market area |

**Archetype gaps [OPEN]:** No Changed archetype in Emberveil. A returned veteran who went deep and came back different would add tension. No explicit Merchant archetype beyond Yeva — a second trader (blacksmith, apothecary) would add economic depth.

### Duskhollow [DESIGNED]

Defined in WORLD.md §2.2.4. Six NPCs proposed.

| Name | Kind | Archetype | Notes |
|---|---|---|---|
| Elder Voss | Dialogue | Keeper/Observer | Void philosophy, community elder |
| Shade Mender Kael | Healer | Keeper | Void-based healing |
| Salvager Rhenn | Shop | Merchant | Sells Void-returned goods |
| The Returned Mara | Dialogue | Changed/Survivor | Died and came back |
| Tunnel Runner Dash | Dialogue | Innocent | Kid who knows the passages |
| Hollow Cat | Animal | Animal | Half-visible, purrs in strange frequencies |

**Archetype gaps:** No Authority (who keeps order in Duskhollow?). No Observer (who studies the Void academically?). Consider adding a Void Warden (Authority) and a Void Researcher (Observer).

### Anchorpoint [DESIGNED]

Defined in WORLD.md §2.3.4. Six NPCs proposed.

| Name | Kind | Archetype | Notes |
|---|---|---|---|
| Construct 4471 | Dialogue | Changed | Retired war machine from Timeline Seven |
| Chrono Keeper Essa | Healer | Keeper | Synchronizes biology to local time |
| Salvager Korr | Shop | Merchant | Temporal salvage dealer |
| Watcher Brin | Dialogue | Authority/Observer | 40 years watching the badlands |
| Settler Mira | Dialogue | Innocent (adult) | Came for reliable time; casual about temporal oddity |
| Temporal Hound | Animal | Animal | Fetches sticks from the future |

**Archetype gaps:** No Survivor (someone who was lost in a time loop and returned). No second Changed (another Construct with a different perspective on retirement). Consider adding a Loop Survivor and a Construct elder.

### Flotsam Market [DESIGNED]

Defined in WORLD.md §2.4.4. Six NPCs proposed.

| Name | Kind | Archetype | Notes |
|---|---|---|---|
| Driftmaster Pell | Dialogue | Authority/Survivor | Navigates by feel, not maps |
| Wave Mender Sorza | Healer | Keeper | Sea-based healing metaphors |
| Trader Cask | Shop | Merchant | Dimensional imports |
| Navigator Kael | Dialogue | Observer/Changed | Sailed to places that no longer exist |
| Flotsam Kid | Dialogue | Innocent | Colors that taste like sounds |
| Storm Parrot | Animal | Animal | Reports weather from other dimensions |

**Archetype gaps:** No Keeper (who maintains the market platforms?). No Changed (someone altered by the Driftlands' instability). Consider adding a Platform Keeper and a Drift-Touched sailor.

### The Codex [DESIGNED]

Defined in WORLD.md §2.5.4. Six NPCs proposed.

| Name | Kind | Archetype | Notes |
|---|---|---|---|
| Archivist Moren | Dialogue | Observer | Collected books from different realities |
| Mender Liss | Healer | Keeper | Heals with preserved observations |
| Trader-Scholar Yenn | Shop | Merchant | Knowledge-infused goods |
| The Index | Dialogue | Changed | The Library's self-awareness |
| Apprentice Zara | Dialogue | Innocent | Read 4,000 books; 12 read her back |
| Library Owl | Animal | Animal | Already read your autobiography |

**Archetype gaps:** No Authority (who enforces Library rules? The Index is close but more Changed than Authority). No Survivor (someone who went into the deep stacks and returned). Consider adding a Stack Warden (Authority) and a Deep Reader (Survivor/Changed).

### Roothold [DESIGNED]

Defined in WORLD.md §2.6.4. Six NPCs proposed.

| Name | Kind | Archetype | Notes |
|---|---|---|---|
| Host-Speaker Vera | Dialogue | Changed | Symbiotic, parasite has opinions |
| Root Healer Moss | Healer | Keeper | Organisms heal, she asks them to |
| Grower Thatch | Shop | Merchant | Living equipment, biologically complicated refunds |
| Elder Symbiont Kira | Dialogue | Changed/Keeper | 40-year integration; "warm companion" |
| Seedling Mott | Dialogue | Innocent | Not yet integrated; spores say he's already chosen |
| Dimensional Frog | Animal | Animal | Half here, half elsewhere |

**Archetype gaps:** No Authority (who governs Roothold?). No Observer (who studies the symbiosis from outside?). Consider adding a Grove Warden (Authority/Keeper) and an Unintegrated Researcher (Observer).

### New Region Template [TEMPLATE]

When designing NPC rosters for a new region:

```
1. Determine regional voice (linguistic patterns, unique dialect features)
2. Fill minimum viable roles:
   [ ] 1 Healer — how does this region's healing work? What makes it distinct?
   [ ] 1 Shop — what does this region trade? What are their goods?
   [ ] 3+ Dialogue — cover at least 3 different archetypes
   [ ] 1+ Animal — local fauna with dimensional characteristics
3. Check archetype coverage — aim for 4+ of the 8 archetypes represented
4. Define relationships — each NPC connects to at least 1 other NPC or institution
5. Write dialogue — 3–4 lines each, region-appropriate register
6. Assign glyphs — check §7 for collision rules
7. Design schedules — see §5
```

---

## 5. Movement and Schedules

### Day Cycle

`DayCycleTicks = 6000` (component constant). At 100ms per tick, one game day = 10 real minutes.

Standard period boundaries:

| Period | StartTick | Real Time |
|---|---|---|
| Night | 0 | 0:00 |
| Morning | 1500 | 2:30 |
| Day | 2500 | 4:10 |
| Evening | 5000 | 8:20 |

### Movement Behaviors

| Behavior | Const | Required Fields | Description |
|---|---|---|---|
| `MoveStationary` | 0 | `StandX, StandY` | Stand in place. If far from position, auto-returns via `MoveReturn`. |
| `MoveWander` | 1 | `BoundsX1, BoundsY1, BoundsX2, BoundsY2` | Random walk within bounding box. |
| `MovePath` | 2 | `Waypoints [][2]int` | Follow waypoints in sequence, greedy-walk. |
| `MoveReturn` | 3 | (internal) | Used automatically during schedule transitions. Never set directly. |

### MoveInterval Guidelines

Controls ticks between movement steps. Lower = faster.

| Speed | MoveInterval | Use For |
|---|---|---|
| Fast | 8–12 | Children, small animals (Pip=12, Pigeon=8, Stray Dog=10) |
| Normal | 14–18 | Adults on business (Greta=14, Maren=16, Alaric=18) |
| Slow | 18–22 | Elderly, sedentary NPCs (Rudwig=20, Father Brennan=20, Fisher Bram=20) |

### Schedule Patterns

**Homebody** — Stays in one location, wanders within it.
```
Night:    Stationary (home position)
Morning:  Wander (home bounds)
Day:      Wander (home bounds)
Evening:  Wander (home bounds)
```
Example: Father Brennan (always in the church nave).

**Commuter** — Travels between two locations daily.
```
Night:    Stationary (home)
Morning:  Path (home → workplace)
Day:      Wander (workplace bounds)
Evening:  Path (workplace → home)
```
Example: Merchant Yeva (home south → shop → home).

**Wanderer** — Moves through large areas, different zones per period.
```
Night:    Stationary (sleep position)
Morning:  Path (sleep → activity area)
Day:      Wander (large activity bounds)
Evening:  Path (activity area → sleep)
```
Example: Street Urchin Pip (market → square → market).

**Guard** — Patrols an area, returns to post.
```
Night:    Stationary (guard post)
Morning:  Path (post → patrol start)
Day:      Wander (patrol bounds)
Evening:  Path (back to post)
```
Example: Soldier Greta (guard post → tower door → patrol square → return).

### Design Rules

1. **All waypoints and stand positions must be on walkable tiles.** Verify against the city map layout. A path through a wall produces a permanently stuck NPC.
2. **Wander bounds must contain walkable tiles.** Don't set bounds that are entirely wall or water.
3. **`StandX: -1, StandY: -1`** is the sentinel for "use per-instance spawn position" (used by Pigeon, which has multiple instances with different spawn points).
4. **NPCs not in `CityNPCSchedules` are fully static.** They stand where placed and never move.
5. **Schedule transitions use `MoveReturn` automatically.** If an NPC is far from their new period's position, the system greedy-walks them there first. Don't worry about transition paths — they handle themselves.
6. **Stuck recovery:** The movement system has built-in stuck detection (alternate axis after 1 stuck, perpendicular jitter after 3). NPCs will work around minor obstacles.

---

## 6. Technical Cookbook

### 6.1 NPCDef in `assets/city_[region].go`

```go
package assets

// [Region]NPCs lists the named NPCs of [CityName].
var [Region]NPCs = []NPCDef{
    {
        Glyph: "👴",
        Name:  "Elder Voss",
        Kind:  0, // NPCKindDialogue
        Lines: []string{
            "The Void is not empty. The Void is full of everything that stopped being observed.",
            "We live in the twilight because the twilight lives in us.",
            "Visitors always ask how we survive here. We ask how you survive without the dark.",
        },
    },
    {
        Glyph: "🙏",
        Name:  "Shade Mender Kael",
        Kind:  1, // NPCKindHealer
        Lines: []string{
            "The Void dissolves wounds along with everything else. Hold still.",
            "You're more solid than most visitors. That will change.",
            "Rest. The dark heals what light cannot reach.",
        },
    },
    // ... more NPCs
}
```

**Kind values:** 0=Dialogue, 1=Healer, 2=Shop, 3=Animal. Use integer literals — the constants are in `component` which `assets` cannot import.

### 6.2 Schedule Definition

```go
// [Region]NPCSchedules maps NPC name → daily schedule.
var [Region]NPCSchedules = map[string]NPCScheduleDef{
    "Elder Voss": {
        MoveInterval: 20,
        Entries: []NPCScheduleEntryDef{
            {StartTick: 0, Behavior: 0, StandX: 50, StandY: 10},                                      // Night: home
            {StartTick: 1500, Behavior: 1, BoundsX1: 45, BoundsY1: 8, BoundsX2: 55, BoundsY2: 15},    // Morning: wander plaza
            {StartTick: 2500, Behavior: 1, BoundsX1: 45, BoundsY1: 8, BoundsX2: 55, BoundsY2: 15},    // Day: wander plaza
            {StartTick: 5000, Behavior: 0, StandX: 50, StandY: 10},                                    // Evening: home
        },
    },
    "Shade Mender Kael": {
        MoveInterval: 18,
        Entries: []NPCScheduleEntryDef{
            {StartTick: 0, Behavior: 0, StandX: 30, StandY: 40},                                                      // Night: grotto
            {StartTick: 1500, Behavior: 2, Waypoints: [][2]int{{30, 35}, {40, 35}, {45, 30}}},                         // Morning: path to plaza
            {StartTick: 2500, Behavior: 1, BoundsX1: 40, BoundsY1: 28, BoundsX2: 50, BoundsY2: 35},                   // Day: wander near plaza
            {StartTick: 5000, Behavior: 2, Waypoints: [][2]int{{40, 35}, {30, 35}, {30, 40}}},                         // Evening: path home
        },
    },
}
```

**Behavior values:** 0=Stationary, 1=Wander, 2=Path. Never use 3 (MoveReturn is internal only).

### 6.3 ShopCatalogue for Shop NPCs

```go
var [Region]ShopCatalogue = []ShopEntry{
    // Consumables
    {Glyph: "🧪", Name: "Hyperflask", Price: 20, IsConsumable: true},
    {Glyph: "🌌", Name: "Void Essence", Price: 25, IsConsumable: true},

    // Equipment
    {Glyph: "🎩", Name: "Void Crown", Price: 60, IsConsumable: false, BonusATK: 1, Slot: "head"},
    {Glyph: "🥼", Name: "Hollow Weave", Price: 55, IsConsumable: false, BonusDEF: 2, Slot: "body"},
    {Glyph: "⚔️", Name: "Shadow Edge", Price: 70, IsConsumable: false, BonusATK: 2, Slot: "onehand"},
    {Glyph: "🪩", Name: "Void Mirror", Price: 50, IsConsumable: false, BonusDEF: 1, Slot: "offhand"},
}
```

**Slot values (string):** `"head"`, `"body"`, `"feet"`, `"onehand"`, `"twohand"`, `"offhand"`. Consumables leave Slot empty.

**Pricing guidelines:** Consumables 12–28g, head/body/onehand equipment 55–70g, offhand/feet equipment 45–55g. Scale slightly higher for later-game cities.

### 6.4 City Placement Pattern

In `internal/mud/city_[region].go`, NPCs are placed using `factory.NewNPC()` and schedules are applied via `attachNPCSchedule()`:

```go
func newCityFloor[Region](rng *rand.Rand) *Floor {
    // ... build map ...

    // Place NPCs
    for _, def := range assets.[Region]NPCs {
        placeNPC(floor, def, rng)
    }

    // Apply schedules
    for _, id := range floor.World.Query(component.CNPC, component.CPosition) {
        npcComp := floor.World.Get(id, component.CNPC).(component.NPC)
        attachNPCSchedule(floor, id, npcComp.Name, assets.[Region]NPCSchedules)
    }

    return floor
}
```

See `internal/mud/city.go` for the complete reference implementation (Emberveil).

### 6.5 `factory.NewNPC()` Reference

```go
func NewNPC(w *ecs.World, name, glyph string, kind component.NPCKind, lines []string, x, y int) ecs.EntityID
```

Creates an entity with:
- `Position{X: x, Y: y}`
- `Renderable{Glyph: glyph, FGColor: tcell.ColorAqua, RenderOrder: 5}`
- `TagBlocking{}`
- `NPC{Name: name, Kind: kind, Lines: lines}`

**Not added by default:** `NPCMovement` (only added by `attachNPCSchedule` for NPCs with schedules).

### 6.6 Interaction Flow

```
Player bumps NPC → system.TryMove returns MoveInteract
→ processActionLocked checks CombatTarget has CNPC component
  → interactNPCLocked(floor, sess, target, npc)
    → switch npc.Kind:
       NPCKindHealer:  heal to full (or show dialogue if full)
       NPCKindShop:    set sess.PendingNPC = 1 (triggers RunShop in RunLoop)
       NPCKindAnimal:  show random line (no speech marks)
       default:        show random line with "💬 Name: \"line\""
```

---

## 7. Glyph Selection

### In-Use NPC Glyphs

**Emberveil [CANON]:** 🙏👴🍺⚔️🛍️📖👦👩🎣🕊️ (humans) | 🐕🐈🕊️🐓 (animals)

**Proposed by WORLD.md [DESIGNED]:**
- Duskhollow: 👴🙏🛍️👤👦🐈‍⬛
- Anchorpoint: 🦾🙏🛍️👴👩🐕
- Flotsam: 👴🙏🛍️👩👦🦜
- Codex: 🧙🙏🛍️📖👧🦉
- Roothold: 🧬🙏🛍️👩👦🐸

### Selection Rules

1. **Never collide with entities on the same floor.** Check tile glyphs, enemy glyphs, item glyphs, and furniture glyphs.
2. **Healer glyphs:** 🙏 is the established convention across all cities. Use it unless there is a strong thematic reason not to.
3. **Shop glyphs:** 🛍️ is the established convention. Same rule.
4. **Animal glyphs:** Use actual animal emoji. Each animal should be visually distinct from all other entities in the city.
5. **Human NPC glyphs:** Use person/people emoji or profession-suggestive emoji. Avoid using glyphs already assigned to player classes (🧙💀🦾🌀🔮🧬).

### Available Human NPC Glyphs (not currently in use)

Suitable for new NPCs: 👨 👵 🧑 👲 🧓 🧒 👮 🤴 👸 💂 🕵️ 🧑‍🌾 🧑‍🔬 🧑‍🎨 🧑‍🔧

### Available Animal Glyphs (not currently in use)

Suitable for new animals: 🐀 🦎 🐍 🦅 🐝 🦋 🐢 🐾 🦔 🦗 🪲 🐦 🦆 🐟

---

## 8. Personality Templates

Quick-reference templates for rapid NPC generation. Each provides sentence patterns for 3–4 lines.

### Veteran Worker
**Archetype:** Authority or Survivor. **Register:** Dry Ambient Voice.
```
Line 1: "[Number] years doing [job]. [Understated observation about the job's reality]."
Line 2: "[Practical advice based on experience]. [Qualification that makes it unsettling]."
Line 3: "[Statement about what they've seen]. [Implication that the thing they've seen is still around]."
```
Example: "40 years watching the badlands. Same frozen moments. Same looping valleys. It never gets less unsettling."

### Regional Elder
**Archetype:** Keeper. **Register:** Dry Ambient Voice.
```
Line 1: "[Regional philosophy stated as fact]. [The fact is stranger than it sounds]."
Line 2: "[Comparison between locals and outsiders]. [The comparison favors locals but reveals something unsettling about them]."
Line 3: "[Reference to the regional phenomenon]. [Acceptance that implies long coexistence]."
```
Example: "The Void is not empty. The Void is full of everything that stopped being observed."

### City Child
**Archetype:** Innocent. **Register:** Comedy Patterns.
```
Line 1: "[Excitement about terrifying local phenomenon]! [Request to the player]."
Line 2: "[Casual mention of dimensional horror as normal childhood experience]. [Childish observation about it]."
Line 3: "[Offer to show the player something dangerous]. [Qualifier that doesn't actually make it safer]."
```
Example: "Want to see a void-bat? They're friendly if you don't look at them!"

### Regional Animal
**Archetype:** Animal. **Register:** Comedy Patterns.
```
Line 1: "The [animal] [action with dimensional implication]. [Second action showing opinion]."
Line 2: "It [behavior] [mundane reaction to impossible phenomenon]. [Judgment about the player]."
Line 3: "The [animal] [physical description involving regional physics]. [Understated weirdness]."
```
Example: "The frog catches a fly that hasn't arrived yet. Interdimensional reflexes."

### The Changed One
**Archetype:** Changed. **Register:** Dry Ambient Voice.
```
Line 1: "[Statement of transformation as fact]. [Time reference that may be wrong]. [Acceptance]."
Line 2: "[Description of the changed state]. [Acknowledgment that it sounds wrong]. [Restatement that it was actually fine]."
Line 3: "[Advice based on the experience]. [Qualifier that makes the advice alarming]."
```
Example: "I died here. Three years ago. Or three years from now — time was unclear during the experience."

### The Merchant
**Archetype:** Merchant. **Register:** Comedy Patterns.
```
Line 1: "[Product description that reveals worldbuilding]. [Sales pitch]."
Line 2: "[Source of goods that implies dimensional weirdness]. [Casual tone about it]."
Line 3: "[Refund/return policy that implies impossible complications]."
```
Example: "Everything I sell was returned by the Void. No refunds — the Void already gave one."

---

## 9. NPC Design Checklist

### Lore Consistency
```
[ ] Region identified — NPC fits regional culture and voice
[ ] Tonal register assigned — one of the three, consistently applied
[ ] Five DON'Ts checked — no exposition dumps, quest language, fourth-wall, purple prose, generic fantasy
[ ] Naming convention followed — short, distinctive, 1–2 syllables (LORE.md §13)
[ ] Relationships defined — connects to at least 1 other NPC or institution
[ ] Archetype identified — maps to one of the 8 archetypes in §2
[ ] Regional dialect applied — uses the region's specific linguistic patterns
```

### Technical Correctness
```
[ ] NPCKind value correct — 0/1/2/3 (integer in assets, not constant name)
[ ] Line count — 3–4 lines, each self-contained
[ ] Glyph — checked against §7 for collisions; not reused on same floor
[ ] Shop catalogue defined (if Kind=2) — prices, slots, bonuses
[ ] Schedule defined (if NPC should move) — all waypoints on walkable tiles
[ ] Wander bounds verified — contains walkable interior tiles
[ ] StandX/StandY on walkable tile (if Stationary periods)
[ ] MoveInterval set — appropriate for character (8–22 range)
```

### Build Verification
```
[ ] go build ./... — compiles without errors
[ ] go test ./... — all tests pass
[ ] go vet ./... — no issues
```

---

## 10. Expansion Protocols

### Adding NPCs to Existing Cities

1. Add `NPCDef` entries to the region's `assets/city_[region].go` file
2. Add schedule entries to the region's `NPCSchedules` map (if NPC should move)
3. If Shop NPC: add `ShopCatalogue` entries
4. Place in the city layout file (`internal/mud/city_[region].go`) — ensure spawn position is on a walkable tile not occupied by other NPCs
5. Run build verification checklist

### Proposing New NPCKinds

The current 4 kinds (Dialogue, Healer, Shop, Animal) cover most interactions. If a new kind is genuinely needed:

1. **Document the design:** What does this kind do that existing kinds cannot? What is the interaction model?
2. **Files to modify:**
   - `internal/component/npc.go` — add new `NPCKind` constant
   - `internal/mud/server.go` — add case to `interactNPCLocked()`
   - Test coverage for the new interaction path
3. **Consider alternatives first:** Can the behavior be achieved with `NPCKindDialogue` + creative dialogue? Many "special" NPCs are just dialogue NPCs with interesting lines.

### NPC-to-NPC Relationships

Strong NPC ecosystems have visible relationships:

- **Cross-references in dialogue:** NPC A mentions NPC B by name. "Ask Scholar Alaric — he's catalogued the energy signatures."
- **Schedule overlaps:** Two NPCs share the same wander bounds during a period, implying they spend time together.
- **Contrasting opinions:** NPC A and NPC B hold different views on the same topic. The player gets both perspectives across separate encounters.
- **Implied history:** An NPC references a shared event. "Maren's husband went in the same expedition as my brother."

These relationships cost nothing mechanically but create the illusion of a living community.

---

*This document bridges lore and code. The NPCs you create will be the world's voice. Make them sound like people, not prompts.*
