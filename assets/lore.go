package assets

// FloorLore holds 3 random atmospheric snippets per floor (index 0 unused).
// One is picked at random on entry.
var FloorLore = [11][]string{
	{}, // floor 0 unused
	{ // Floor 1 — Crystalline Labs
		"Frost-rimed walls hum with contained experiments. Several containment fields are no longer containing anything.",
		"Lab logs reference 'Phase III'. Phase I and II notes are conspicuously absent.",
		"A whiteboard reads: 'DO NOT TOUCH THE CRYSTALS'. Someone has drawn a smiley face next to it.",
	},
	{ // Floor 2 — Bioluminescent Warrens
		"Spores drift lazily through the air. They smell faintly of copper and ambition.",
		"The walls breathe. You tell yourself this is a metaphor. The walls do not agree.",
		"Someone carved 'Prometheus had the right idea' into a fungal column. You're not sure that's reassuring.",
	},
	{ // Floor 3 — Resonance Engine
		"Every gear turns in perfect synchrony. The machine does not appear to have an off switch.",
		"A placard reads: 'IN CASE OF RESONANCE CASCADE — EVACUATE DOWNWARD'. Downward seems like a bad idea.",
		"The vibration at this frequency is technically music. Technically.",
	},
	{ // Floor 4 — Fractured Observatory
		"The lenses were always aimed inward. Whatever they found, they kept climbing toward it.",
		"Star charts cover every surface. They don't correspond to any known sky.",
		"An astronomer's last entry: 'It blinked back.' The entry is dated seventeen years ago. The ink is still wet.",
	},
	{ // Floor 5 — Apex Nexus
		"Power conduits scar the walls like veins. The Spire's heart beats somewhere above.",
		"Security protocols are still active. They have adapted to their new purpose beautifully.",
		"The Apex Warden was once a curator. It takes its new role very seriously.",
	},
	{ // Floor 6 — Membrane of Echoes
		"Dimensional membranes ripple at your touch. Reality here is more of a guideline.",
		"You hear voices from directions that don't exist. They sound like you, slightly wrong.",
		"OSHA was not consulted when this floor was designed. Several dimensions were.",
	},
	{ // Floor 7 — The Calcified Archive
		"They chose to become the knowledge they sought. The transformation was apparently irreversible.",
		"Petrified scholars line the shelves alongside the books. It's hard to tell which are which.",
		"A crystallised hand still holds a quill. The inscription it was writing reads: 'And thus we arr—'",
	},
	{ // Floor 8 — Abyssal Foundry
		"Something is being forged here that has no name in any language you know.",
		"The heat doesn't burn. It remembers. And it has very strong opinions about you.",
		"Safety notice: Eye protection required. Soul protection not provided. Inquire within.",
	},
	{ // Floor 9 — The Dreaming Cortex
		"The collective unconscious has been under new management since the Spire was built. Management is displeased.",
		"Your thoughts echo back at you, slightly improved. The improvement is unsettling.",
		"Every nightmare you've ever had is a floor plan here. Someone has been very busy.",
	},
	{ // Floor 10 — The Prismatic Heart
		"The Unmaker waits at the core of all crystallised thought. It has been waiting since before thought existed.",
		"This is where the Spire's lenses were always aimed. The light here goes in, not out.",
		"You are not the first to reach this chamber. The others became part of the architecture.",
	},
}

// WallWritings holds a pool of inscriptions per floor (index 0 unused).
// 2-5 are chosen at random and placed on floor tiles each run.
var WallWritings = [11][]string{
	{}, // floor 0 unused
	{ // Floor 1 — Crystalline Labs
		"HAZARD CLASS 4 — DO NOT LICK THE WALLS. [Someone has added 'too late' in different ink.]",
		"Lab Log 7: The crystals have started humming. Dr. Voss says it's resonance. I don't think it's resonance.",
		"EXIT IS THAT WAY →  [The arrow has been corrected three times. The final correction points downward.]",
		"SAMPLE 7-G BREACHED. REMAIN CALM. SAMPLES 7-H THROUGH 7-Z ALSO BREACHED. BEGIN PANICKING.",
		"If found, please water my fern. — Researcher Thali. P.S. The fern may have achieved sentience. Proceed with caution.",
		"PHASE III IS PROCEEDING AS PLANNED [crossed out] SOMEWHAT AS PLANNED [crossed out] DIFFERENTLY THAN PLANNED.",
		"Note to self: the containment field is rated for crystal entities, not crystalline EMOTIONS. Update procurement.",
	},
	{ // Floor 2 — Bioluminescent Warrens
		"The spores are not harmful. The spores are not harmful. The spores are — [text trails into glowing fungus]",
		"Day 1: Beautiful. Day 4: Growing. Day 9: I am becoming beautiful too.",
		"CAUTION: Walls may bite. Not in a dangerous way. They're just curious.",
		"The mycelium network extends 40 leagues in every direction. It has opinions about you specifically.",
		"Prometheus was right. Fire was the lesser gift. — scrawled in bioluminescent ink",
		"Please do not think loudly. The spores respond to cognitive emissions.",
		"FUNGAL ADVISORY: If the walls start whispering your name, it means they like you. If they start knowing your secrets, evacuate.",
	},
	{ // Floor 3 — Resonance Engine
		"MAINTENANCE LOG: All systems nominal. The screaming is a feature.",
		"In the event of resonance cascade: Step 1) Do not panic. Step 2) Panic is acceptable. Step 3) There is no step 3.",
		"FREQUENCY 7.3Hz — KEEP CLEAR. FREQUENCY 7.4Hz — PLEASE EVACUATE. FREQUENCY 7.5Hz — IT'S TOO LATE.",
		"The gears have been turning for 400 years. No one has found where the power comes from.",
		"CAUTION: Sustained exposure to harmonic frequencies may cause [illegible]. Consult a physician who has also heard the music.",
		"The engine doesn't stop. The engine has never stopped. The engine predates the Spire by an amount we don't discuss.",
		"Engineer's note: The machine isn't broken. The machine is PERFECT. That's the problem.",
	},
	{ // Floor 4 — Fractured Observatory
		"THE THING IN THE LENS — DO NOT LOOK DIRECTLY — [the rest is scratched out with something that burned]",
		"Star Survey Complete: 0 stars found. 1 thing found. It found us back. Survey terminated.",
		"The observatory has been closed since Year 12. The lenses still move.",
		"NOTICE: The star charts are not charts of the sky. They are maps of something inside the sky.",
		"Dr. Herath's last entry: 'It's not a star. Stars don't blink. Stars don't wait. Stars don't—'",
		"There are 14 astronomer's chairs. Only 13 astronomers ever worked here. The 14th chair is warm.",
		"Observation log, final entry: We have confirmed that it is aware of the observatory. We have not confirmed that it is aware of us. Addendum: it is aware of us.",
	},
	{ // Floor 5 — Apex Nexus
		"SECURITY LEVEL: OMEGA. All personnel have been recalled. Warden protocols active. Have a productive day.",
		"The Warden was not always this. It remembers being something softer. It does not like remembering.",
		"POWER CONDUIT WARNING: Do not touch. The conduit has already touched you. This is a formality.",
		"Nexus Authority Log 4412: All systems transferred to autonomous control. All staff [REDACTED].",
		"I think the Warden is sad. I think it's guarding something that no longer needs guarding. — scratched in haste",
		"The heart of the Spire is upward. The heart of the Spire is waiting. The heart of the Spire has excellent patience.",
		"AUTOMATED NOTICE: You are not authorised to be here. Your unauthorised presence has been noted and appreciated.",
	},
	{ // Floor 6 — Membrane of Echoes
		"DIMENSIONAL STABILITY: 34%. Do not make sudden movements. The walls are listening from four directions at once.",
		"You have been here before. You will be here again. Hello. [The handwriting is yours.]",
		"CAUTION: Echoes in this section are not echoes. They are memories. They are learning new tricks.",
		"The membrane is thinning. This is normal. This is not normal. Both are true simultaneously, which is also normal.",
		"Something on the other side keeps pressing against the wall. We decided not to press back. We reconsidered. We pressed back.",
		"If you can read this, you are partially extradimensional. Congratulations. Please try not to sneeze.",
		"INTER-DIMENSIONAL HEALTH ADVISORY: Side effects of membrane proximity include: déjà vu, jamais vu, and mild ontological uncertainty.",
	},
	{ // Floor 7 — The Calcified Archive
		"THE PRICE OF KNOWLEDGE: Everything. THE PRICE OF IGNORANCE: Also everything, but faster.",
		"ARCHIVE POLICY: No writing in the margins. [The margin contains 300 pages of cramped, desperate notes.]",
		"Scholar Orn's Thesis: 'On the Irreversibility of Epistemic Crystallisation.' Status: Unfinished. Author Status: See Thesis.",
		"The books here know things the authors never wrote. The books learned it from each other during the long dark.",
		"CAUTION: Some texts are still completing themselves. Do not interrupt them. Do not let them interrupt you.",
		"Entry 1: I came for knowledge. Entry 89: I have become the knowledge. Entry 90: [written in mineral deposits, illegible, but somehow familiar]",
		"CATALOGUING NOTE: We stopped numbering the volumes when the volumes started numbering themselves.",
	},
	{ // Floor 8 — Abyssal Foundry
		"WHAT IS BEING FORGED: Unknown. HOW LONG: Unknown. FOR WHOM: [REDACTED] [REDACTED] [REDACTED].",
		"The heat here does not burn. It evaluates. It has found you wanting in seventeen specific ways.",
		"FORGE SAFETY: Eye protection required. Soul protection: bring your own. No refunds on either.",
		"The first thing forged here predates the Foundry. The Foundry was built around it. The Foundry does not discuss this.",
		"My hands are changing. This is fine. The work is almost done. The work has always been almost done. — Forge-Tender Maeris",
		"QUALITY CONTROL LOG: Batches 1-4891: adequate. Batch 4892: [the log catches fire at this point and the fire has opinions]",
		"TEMPERATURE ADVISORY: The concept of temperature breaks down below sublevel 3. Dress accordingly. There is no adequate dress.",
	},
	{ // Floor 9 — The Dreaming Cortex
		"YOUR THOUGHTS ARE BEING RECORDED. Your thoughts are not your thoughts. Your thoughts are filing a formal complaint.",
		"The Cortex has been awake since it was built. It dreams of being asleep. Its dreams are everyone else's nightmares.",
		"PSIONIC HAZARD ZONE: Do not think hostile thoughts. Do not think at all, if avoidable. Think of something pleasant. Not that.",
		"Dream Survey: 94% of nightmares in this sector are original content. 6% are yours. You brought them in.",
		"Something learned to think in here, and then learned to think about thinking, and has not stopped since.",
		"I came to study the dreaming mind. The dreaming mind studied me back. We reached an agreement I can no longer remember. — [no signature, only unease]",
		"COGNITIVE RESTRUCTURING IN PROGRESS. Please hold. Your original thoughts will be returned to you, improved. Please hold.",
	},
	{ // Floor 10 — The Prismatic Heart
		"THE UNMAKER WAS NOT ALWAYS THE UNMAKER. IT REMEMBERS WHAT IT WAS. IT WISHES IT DIDN'T.",
		"You are not the first. The first was absorbed in Year 1. The Spire has been refining the process.",
		"The light here is thinking. The light here is hungry. The light here has been waiting for something with your face.",
		"PRISMATIC CONVERGENCE PROTOCOLS ACTIVE. ALL INTRUDERS WILL BE CRYSTALLISED INTO MEMORY. WELCOME TO THE ARCHIVE.",
		"It began with a question: what lies at the intersection of all dimensions? The answer woke up. The answer is not pleased.",
		"If you succeed, the Spire dies. If you fail, you become part of the Spire. We wanted you to know the stakes. We admire your commitment either way.",
		"This room has no walls. The walls are made of everything that came before you. You are about to become a wall.",
	},
}

// EnemyLore holds a one-liner shown the first time each enemy type is killed.
// Keyed by glyph emoji.
var EnemyLore = map[string]string{
	// Floor 1–5 enemies
	GlyphCrystalCrawl: "The Crystal Crawl — a failed experiment in mineral cognition. It was almost sentient.",
	GlyphNeonSpecter:  "The Neon Specter — light given malice. The lab notes called it a 'luminous success'.",
	GlyphPrismDrake:   "The Prism Drake — a security asset repurposed by someone with worse ideas than the original designers.",
	GlyphVoidTendril:  "The Void Tendril — an appendage of something larger that, mercifully, did not follow it through.",
	GlyphThoughtLeech: "The Thought Leech — feeds on cognition. You feel briefly smarter. Then you feel its absence.",
	GlyphFractalGolem: "The Fractal Golem — built to last. It outlasted its builders by several geological epochs.",
	GlyphEntropyBloom: "The Entropy Bloom — chaos given floral form. Its beauty is genuinely impressive, and lethal.",
	GlyphApexWarden:   "The Apex Warden — a curator turned gatekeeper. Its resignation letter was never filed.",
	// Floor 6–10 enemies
	GlyphToxinSpore:       "The Toxin Spore — spores designed to mark dimensional boundaries. It marks you instead.",
	GlyphTideWraith:       "The Tide Wraith — a current from a sea that doesn't exist yet. Time is flexible here.",
	GlyphOssifiedScholar:  "The Ossified Scholar — chose knowledge over mortality, got neither as expected.",
	GlyphArchiveWarden:    "The Archive Warden — has been protecting these records since before the records existed.",
	GlyphCinderWraith:     "The Cinder Wraith — the Foundry's quality-control inspector, post-accident.",
	GlyphForgeGolem:       "The Forge Golem — built to be immortal. It takes 'resource recycling' very literally.",
	GlyphDreamStalker:     "The Dream Stalker — hunts in the space between thoughts. You weren't thinking about it. You were wrong.",
	GlyphPsychicEcho:      "The Psychic Echo — a memory of something terrible, learning to be terrible again.",
	GlyphCrystalRevenant:  "The Crystal Revenant — returned from somewhere worse. Doesn't want to go back. Will take yours.",
	GlyphUnmaker:          "The Unmaker — the last question the Spire ever asked. It did not like the answer.",
}
