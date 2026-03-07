package assets

import "math"

// ─── XP Curve ────────────────────────────────────────────────────────────────

const MaxLevel = 100

// XPForLevel returns the cumulative XP needed to reach level n (n >= 2).
// Formula: floor(100 * 1.12^(n-1)).
func XPForLevel(n int) int {
	if n <= 1 {
		return 0
	}
	return int(math.Floor(100 * math.Pow(1.12, float64(n-1))))
}

// ─── XP Sources ──────────────────────────────────────────────────────────────

// XPForKill returns XP awarded for killing an enemy with the given threat cost.
func XPForKill(threatCost, floor int) int {
	return threatCost * (20 + floor*5)
}

// XPForEliteKill returns XP for killing a floor elite.
func XPForEliteKill(floor int) int {
	return 500 + floor*100
}

// XPForFloorEntry returns XP for first-time entry to a floor.
func XPForFloorEntry(floor int) int {
	return DungeonFloor(floor) * 150
}

const (
	XPForFurniture   = 50
	XPForPickup      = 10
	XPForInscription = 15
)

// ─── Class Growth ────────────────────────────────────────────────────────────

// ClassGrowth defines per-level stat scaling for a class.
type ClassGrowth struct {
	HPPerLevel float64 // MaxHP gained per level
	ATKPer5    int     // ATK gained every 5 levels
	DEFPer5    int     // DEF gained every 5 levels
}

// ClassGrowths maps class ID to its growth rates.
var ClassGrowths = map[string]ClassGrowth{
	"arcanist":  {HPPerLevel: 2.0, ATKPer5: 2, DEFPer5: 1},
	"revenant":  {HPPerLevel: 1.0, ATKPer5: 3, DEFPer5: 0},
	"construct": {HPPerLevel: 3.5, ATKPer5: 1, DEFPer5: 2},
	"dancer":    {HPPerLevel: 1.5, ATKPer5: 2, DEFPer5: 1},
	"oracle":    {HPPerLevel: 1.5, ATKPer5: 1, DEFPer5: 1},
	"symbiont":  {HPPerLevel: 2.5, ATKPer5: 2, DEFPer5: 1},
}

// GrowthForLevel returns cumulative (bonusHP, bonusATK, bonusDEF) for reaching
// the given level with the given class growth.
func GrowthForLevel(g ClassGrowth, level int) (hp, atk, def int) {
	if level <= 1 {
		return 0, 0, 0
	}
	gained := level - 1 // levels gained above 1
	hp = int(math.Floor(g.HPPerLevel * float64(gained)))
	atk = (gained / 5) * g.ATKPer5
	def = (gained / 5) * g.DEFPer5
	return hp, atk, def
}

// ─── Skill Tiers ─────────────────────────────────────────────────────────────

type SkillTier int

const (
	TierNovice SkillTier = iota // levels 1-9
	TierAdept                   // levels 10-24
	TierExpert                  // levels 25-44
	TierMaster                  // levels 45-69
	TierLegend                  // levels 70-100
)

// TierForLevel returns the skill tier available at the given level.
func TierForLevel(level int) SkillTier {
	switch {
	case level >= 70:
		return TierLegend
	case level >= 45:
		return TierMaster
	case level >= 25:
		return TierExpert
	case level >= 10:
		return TierAdept
	default:
		return TierNovice
	}
}

// BranchMilestones are levels where the player picks a specialization.
var BranchMilestones = []int{10, 25, 45, 70}

// IsBranchMilestone returns true if the given level is a specialization choice point.
func IsBranchMilestone(level int) bool {
	for _, m := range BranchMilestones {
		if level == m {
			return true
		}
	}
	return false
}

// ─── Skill Definitions ──────────────────────────────────────────────────────

// SkillKind describes what a skill does.
type SkillKind uint8

const (
	SkillPassiveStat    SkillKind = iota // permanent stat bonus
	SkillPassiveProc                     // on-event proc (kill heal, dodge, etc)
	SkillAbilityUpgrade                  // cooldown reduction, duration bonus
)

// SkillDef defines one learnable skill.
type SkillDef struct {
	ID          string
	Name        string
	Description string
	ClassID     string    // which class this belongs to
	Tier        SkillTier // required tier
	Branch      string    // "" = no branch required, "A" or "B"
	Kind        SkillKind

	// Effects (additive; only relevant fields set per skill)
	BonusATK       int
	BonusDEF       int
	BonusMaxHP     int
	BonusFOV       int
	DodgeChance    int
	KillHealBonus  int
	KillHealAdd    int
	ThornsDamage   int
	CooldownReduce int
	RegenReduce    int
}

// ─── Specialization Definitions ──────────────────────────────────────────────

// SpecDef defines a branch specialization choice.
type SpecDef struct {
	Branch      string // "A" or "B"
	Name        string
	Description string
}

// ClassSpecs maps class ID → milestone level → [2]SpecDef (A and B).
var ClassSpecs = map[string]map[int][2]SpecDef{
	"arcanist": {
		10: {
			{Branch: "A", Name: "Battlemage", Description: "Combat magic: ATK, lifesteal, damage procs"},
			{Branch: "B", Name: "Chronomancer", Description: "Time magic: FOV, dodge, cooldown reduction"},
		},
	},
	"revenant": {
		10: {
			{Branch: "A", Name: "Deathknight", Description: "Melee destruction: massive ATK, lifedrain"},
			{Branch: "B", Name: "Wraith", Description: "Evasive undead: dodge, thorns, sustain"},
		},
	},
	"construct": {
		10: {
			{Branch: "A", Name: "Juggernaut", Description: "Unstoppable tank: HP, DEF, thorns"},
			{Branch: "B", Name: "Overclocker", Description: "Glass cannon: ATK, cooldown, risk/reward"},
		},
	},
	"dancer": {
		10: {
			{Branch: "A", Name: "Bladedancer", Description: "Swift strikes: ATK, kill procs, aggression"},
			{Branch: "B", Name: "Shadowdancer", Description: "Evasion mastery: dodge, FOV, stealth"},
		},
	},
	"oracle": {
		10: {
			{Branch: "A", Name: "Seer", Description: "Foresight: FOV, dodge, awareness"},
			{Branch: "B", Name: "Mystic", Description: "Crystal power: ATK, DEF, ability upgrades"},
		},
	},
	"symbiont": {
		10: {
			{Branch: "A", Name: "Hivemind", Description: "Living weapon: ATK, kill heal, regen"},
			{Branch: "B", Name: "Carapace", Description: "Living armor: DEF, HP, thorns"},
		},
	},
}

// ─── Skill Registry ─────────────────────────────────────────────────────────

// AllSkills is the complete registry of learnable skills.
var AllSkills = []SkillDef{
	// ═══ ARCANIST ═══

	// Tier 0 — Novice (no branch)
	{ID: "arc_t0_atk", Name: "Arcane Attunement", Description: "+2 ATK", ClassID: "arcanist", Tier: TierNovice, Kind: SkillPassiveStat, BonusATK: 2},
	{ID: "arc_t0_hp", Name: "Mana Shield", Description: "+3 MaxHP", ClassID: "arcanist", Tier: TierNovice, Kind: SkillPassiveStat, BonusMaxHP: 3},
	{ID: "arc_t0_kh", Name: "Spell Echo", Description: "+10% kill heal chance", ClassID: "arcanist", Tier: TierNovice, Kind: SkillPassiveProc, KillHealAdd: 10},
	{ID: "arc_t0_cd", Name: "Rift Memory", Description: "-2 ability cooldown", ClassID: "arcanist", Tier: TierNovice, Kind: SkillAbilityUpgrade, CooldownReduce: 2},
	{ID: "arc_t0_fov", Name: "Crystalline Focus", Description: "+1 FOV", ClassID: "arcanist", Tier: TierNovice, Kind: SkillPassiveStat, BonusFOV: 1},

	// Tier 1 — Adept, Branch A (Battlemage)
	{ID: "arc_t1a_atk", Name: "War Magic", Description: "+3 ATK", ClassID: "arcanist", Tier: TierAdept, Branch: "A", Kind: SkillPassiveStat, BonusATK: 3},
	{ID: "arc_t1a_kh", Name: "Spell Siphon", Description: "+15% kill heal chance", ClassID: "arcanist", Tier: TierAdept, Branch: "A", Kind: SkillPassiveProc, KillHealAdd: 15},
	{ID: "arc_t1a_cd", Name: "Dimensional Strike", Description: "-3 ability cooldown", ClassID: "arcanist", Tier: TierAdept, Branch: "A", Kind: SkillAbilityUpgrade, CooldownReduce: 3},
	{ID: "arc_t1a_hp", Name: "Fortified Mind", Description: "+5 MaxHP", ClassID: "arcanist", Tier: TierAdept, Branch: "A", Kind: SkillPassiveStat, BonusMaxHP: 5},
	{ID: "arc_t1a_khr", Name: "Arcane Blade", Description: "+2 HP on kill", ClassID: "arcanist", Tier: TierAdept, Branch: "A", Kind: SkillPassiveProc, KillHealBonus: 2},

	// Tier 1 — Adept, Branch B (Chronomancer)
	{ID: "arc_t1b_fov", Name: "Temporal Sight", Description: "+2 FOV", ClassID: "arcanist", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusFOV: 2},
	{ID: "arc_t1b_dodge", Name: "Phase Step", Description: "10% dodge chance", ClassID: "arcanist", Tier: TierAdept, Branch: "B", Kind: SkillPassiveProc, DodgeChance: 10},
	{ID: "arc_t1b_def", Name: "Chrono Shield", Description: "+2 DEF", ClassID: "arcanist", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusDEF: 2},
	{ID: "arc_t1b_cd", Name: "Rift Mastery", Description: "-4 ability cooldown", ClassID: "arcanist", Tier: TierAdept, Branch: "B", Kind: SkillAbilityUpgrade, CooldownReduce: 4},
	{ID: "arc_t1b_hp", Name: "Time Dilation", Description: "+4 MaxHP", ClassID: "arcanist", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusMaxHP: 4},

	// ═══ REVENANT ═══

	// Tier 0 — Novice
	{ID: "rev_t0_atk", Name: "Death's Fury", Description: "+3 ATK", ClassID: "revenant", Tier: TierNovice, Kind: SkillPassiveStat, BonusATK: 3},
	{ID: "rev_t0_kh", Name: "Soul Drain", Description: "+2 HP on kill", ClassID: "revenant", Tier: TierNovice, Kind: SkillPassiveProc, KillHealBonus: 2},
	{ID: "rev_t0_hp", Name: "Undying Will", Description: "+3 MaxHP", ClassID: "revenant", Tier: TierNovice, Kind: SkillPassiveStat, BonusMaxHP: 3},
	{ID: "rev_t0_cd", Name: "Dark Pact", Description: "-2 ability cooldown", ClassID: "revenant", Tier: TierNovice, Kind: SkillAbilityUpgrade, CooldownReduce: 2},
	{ID: "rev_t0_thorns", Name: "Necrotic Aura", Description: "+1 thorns damage", ClassID: "revenant", Tier: TierNovice, Kind: SkillPassiveProc, ThornsDamage: 1},

	// Tier 1 — Adept, Branch A (Deathknight)
	{ID: "rev_t1a_atk", Name: "Annihilate", Description: "+4 ATK", ClassID: "revenant", Tier: TierAdept, Branch: "A", Kind: SkillPassiveStat, BonusATK: 4},
	{ID: "rev_t1a_kh", Name: "Deathstrike", Description: "+3 HP on kill", ClassID: "revenant", Tier: TierAdept, Branch: "A", Kind: SkillPassiveProc, KillHealBonus: 3},
	{ID: "rev_t1a_cd", Name: "Bargain Mastery", Description: "-3 ability cooldown", ClassID: "revenant", Tier: TierAdept, Branch: "A", Kind: SkillAbilityUpgrade, CooldownReduce: 3},
	{ID: "rev_t1a_hp", Name: "Blood Armor", Description: "+4 MaxHP", ClassID: "revenant", Tier: TierAdept, Branch: "A", Kind: SkillPassiveStat, BonusMaxHP: 4},
	{ID: "rev_t1a_atk2", Name: "Reaping Strikes", Description: "+3 ATK", ClassID: "revenant", Tier: TierAdept, Branch: "A", Kind: SkillPassiveStat, BonusATK: 3},

	// Tier 1 — Adept, Branch B (Wraith)
	{ID: "rev_t1b_dodge", Name: "Wraith Form", Description: "12% dodge chance", ClassID: "revenant", Tier: TierAdept, Branch: "B", Kind: SkillPassiveProc, DodgeChance: 12},
	{ID: "rev_t1b_thorns", Name: "Death's Echo", Description: "+2 thorns damage", ClassID: "revenant", Tier: TierAdept, Branch: "B", Kind: SkillPassiveProc, ThornsDamage: 2},
	{ID: "rev_t1b_def", Name: "Spectral Shield", Description: "+2 DEF", ClassID: "revenant", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusDEF: 2},
	{ID: "rev_t1b_kh", Name: "Life Tap", Description: "+2 HP on kill", ClassID: "revenant", Tier: TierAdept, Branch: "B", Kind: SkillPassiveProc, KillHealBonus: 2},
	{ID: "rev_t1b_hp", Name: "Ethereal Body", Description: "+5 MaxHP", ClassID: "revenant", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusMaxHP: 5},

	// ═══ CONSTRUCT ═══

	// Tier 0 — Novice
	{ID: "con_t0_hp", Name: "Reinforced Plating", Description: "+5 MaxHP", ClassID: "construct", Tier: TierNovice, Kind: SkillPassiveStat, BonusMaxHP: 5},
	{ID: "con_t0_def", Name: "Hardened Shell", Description: "+2 DEF", ClassID: "construct", Tier: TierNovice, Kind: SkillPassiveStat, BonusDEF: 2},
	{ID: "con_t0_regen", Name: "Efficient Repair", Description: "-2 regen interval", ClassID: "construct", Tier: TierNovice, Kind: SkillAbilityUpgrade, RegenReduce: 2},
	{ID: "con_t0_atk", Name: "Power Servos", Description: "+1 ATK", ClassID: "construct", Tier: TierNovice, Kind: SkillPassiveStat, BonusATK: 1},
	{ID: "con_t0_thorns", Name: "Reactive Armor", Description: "+1 thorns damage", ClassID: "construct", Tier: TierNovice, Kind: SkillPassiveProc, ThornsDamage: 1},

	// Tier 1 — Adept, Branch A (Juggernaut)
	{ID: "con_t1a_hp", Name: "Titan Frame", Description: "+8 MaxHP", ClassID: "construct", Tier: TierAdept, Branch: "A", Kind: SkillPassiveStat, BonusMaxHP: 8},
	{ID: "con_t1a_def", Name: "Fortress Protocol", Description: "+3 DEF", ClassID: "construct", Tier: TierAdept, Branch: "A", Kind: SkillPassiveStat, BonusDEF: 3},
	{ID: "con_t1a_thorns", Name: "Spiked Chassis", Description: "+2 thorns damage", ClassID: "construct", Tier: TierAdept, Branch: "A", Kind: SkillPassiveProc, ThornsDamage: 2},
	{ID: "con_t1a_regen", Name: "Nano-Repair", Description: "-2 regen interval", ClassID: "construct", Tier: TierAdept, Branch: "A", Kind: SkillAbilityUpgrade, RegenReduce: 2},
	{ID: "con_t1a_hp2", Name: "Ablative Layer", Description: "+6 MaxHP", ClassID: "construct", Tier: TierAdept, Branch: "A", Kind: SkillPassiveStat, BonusMaxHP: 6},

	// Tier 1 — Adept, Branch B (Overclocker)
	{ID: "con_t1b_atk", Name: "Overcharge Servos", Description: "+3 ATK", ClassID: "construct", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusATK: 3},
	{ID: "con_t1b_cd", Name: "Turbo Cycle", Description: "-4 ability cooldown", ClassID: "construct", Tier: TierAdept, Branch: "B", Kind: SkillAbilityUpgrade, CooldownReduce: 4},
	{ID: "con_t1b_atk2", Name: "Plasma Core", Description: "+2 ATK", ClassID: "construct", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusATK: 2},
	{ID: "con_t1b_kh", Name: "Salvage Protocol", Description: "+2 HP on kill", ClassID: "construct", Tier: TierAdept, Branch: "B", Kind: SkillPassiveProc, KillHealBonus: 2},
	{ID: "con_t1b_hp", Name: "Heat Sink", Description: "+4 MaxHP", ClassID: "construct", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusMaxHP: 4},

	// ═══ DANCER ═══

	// Tier 0 — Novice
	{ID: "dan_t0_atk", Name: "Momentum", Description: "+2 ATK", ClassID: "dancer", Tier: TierNovice, Kind: SkillPassiveStat, BonusATK: 2},
	{ID: "dan_t0_dodge", Name: "Evasive Step", Description: "8% dodge chance", ClassID: "dancer", Tier: TierNovice, Kind: SkillPassiveProc, DodgeChance: 8},
	{ID: "dan_t0_fov", Name: "Entropy Sense", Description: "+1 FOV", ClassID: "dancer", Tier: TierNovice, Kind: SkillPassiveStat, BonusFOV: 1},
	{ID: "dan_t0_hp", Name: "Chaos Guard", Description: "+3 MaxHP", ClassID: "dancer", Tier: TierNovice, Kind: SkillPassiveStat, BonusMaxHP: 3},
	{ID: "dan_t0_cd", Name: "Flow State", Description: "-3 ability cooldown", ClassID: "dancer", Tier: TierNovice, Kind: SkillAbilityUpgrade, CooldownReduce: 3},

	// Tier 1 — Adept, Branch A (Bladedancer)
	{ID: "dan_t1a_atk", Name: "Whirlwind", Description: "+3 ATK", ClassID: "dancer", Tier: TierAdept, Branch: "A", Kind: SkillPassiveStat, BonusATK: 3},
	{ID: "dan_t1a_kh", Name: "Blood Dance", Description: "+2 HP on kill", ClassID: "dancer", Tier: TierAdept, Branch: "A", Kind: SkillPassiveProc, KillHealBonus: 2},
	{ID: "dan_t1a_atk2", Name: "Razor Tempo", Description: "+2 ATK", ClassID: "dancer", Tier: TierAdept, Branch: "A", Kind: SkillPassiveStat, BonusATK: 2},
	{ID: "dan_t1a_cd", Name: "Blitz", Description: "-4 ability cooldown", ClassID: "dancer", Tier: TierAdept, Branch: "A", Kind: SkillAbilityUpgrade, CooldownReduce: 4},
	{ID: "dan_t1a_hp", Name: "Adrenaline", Description: "+4 MaxHP", ClassID: "dancer", Tier: TierAdept, Branch: "A", Kind: SkillPassiveStat, BonusMaxHP: 4},

	// Tier 1 — Adept, Branch B (Shadowdancer)
	{ID: "dan_t1b_dodge", Name: "Shadow Weave", Description: "12% dodge chance", ClassID: "dancer", Tier: TierAdept, Branch: "B", Kind: SkillPassiveProc, DodgeChance: 12},
	{ID: "dan_t1b_fov", Name: "Dark Vision", Description: "+2 FOV", ClassID: "dancer", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusFOV: 2},
	{ID: "dan_t1b_def", Name: "Shadow Cloak", Description: "+2 DEF", ClassID: "dancer", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusDEF: 2},
	{ID: "dan_t1b_cd", Name: "Vanishing Act", Description: "-3 ability cooldown", ClassID: "dancer", Tier: TierAdept, Branch: "B", Kind: SkillAbilityUpgrade, CooldownReduce: 3},
	{ID: "dan_t1b_hp", Name: "Ethereal Grace", Description: "+4 MaxHP", ClassID: "dancer", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusMaxHP: 4},

	// ═══ ORACLE ═══

	// Tier 0 — Novice
	{ID: "ora_t0_fov", Name: "Crystal Clarity", Description: "+2 FOV", ClassID: "oracle", Tier: TierNovice, Kind: SkillPassiveStat, BonusFOV: 2},
	{ID: "ora_t0_def", Name: "Prescient Guard", Description: "+1 DEF", ClassID: "oracle", Tier: TierNovice, Kind: SkillPassiveStat, BonusDEF: 1},
	{ID: "ora_t0_dodge", Name: "Foreseen Dodge", Description: "8% dodge chance", ClassID: "oracle", Tier: TierNovice, Kind: SkillPassiveProc, DodgeChance: 8},
	{ID: "ora_t0_cd", Name: "Future Sight", Description: "-3 ability cooldown", ClassID: "oracle", Tier: TierNovice, Kind: SkillAbilityUpgrade, CooldownReduce: 3},
	{ID: "ora_t0_hp", Name: "Crystal Shell", Description: "+3 MaxHP", ClassID: "oracle", Tier: TierNovice, Kind: SkillPassiveStat, BonusMaxHP: 3},

	// Tier 1 — Adept, Branch A (Seer)
	{ID: "ora_t1a_fov", Name: "All-Seeing Eye", Description: "+3 FOV", ClassID: "oracle", Tier: TierAdept, Branch: "A", Kind: SkillPassiveStat, BonusFOV: 3},
	{ID: "ora_t1a_dodge", Name: "Precognition", Description: "12% dodge chance", ClassID: "oracle", Tier: TierAdept, Branch: "A", Kind: SkillPassiveProc, DodgeChance: 12},
	{ID: "ora_t1a_def", Name: "Prophetic Shield", Description: "+2 DEF", ClassID: "oracle", Tier: TierAdept, Branch: "A", Kind: SkillPassiveStat, BonusDEF: 2},
	{ID: "ora_t1a_cd", Name: "Timeline Mastery", Description: "-4 ability cooldown", ClassID: "oracle", Tier: TierAdept, Branch: "A", Kind: SkillAbilityUpgrade, CooldownReduce: 4},
	{ID: "ora_t1a_hp", Name: "Crystal Fortitude", Description: "+5 MaxHP", ClassID: "oracle", Tier: TierAdept, Branch: "A", Kind: SkillPassiveStat, BonusMaxHP: 5},

	// Tier 1 — Adept, Branch B (Mystic)
	{ID: "ora_t1b_atk", Name: "Crystal Barrage", Description: "+3 ATK", ClassID: "oracle", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusATK: 3},
	{ID: "ora_t1b_def", Name: "Mystic Ward", Description: "+2 DEF", ClassID: "oracle", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusDEF: 2},
	{ID: "ora_t1b_cd", Name: "Resonant Focus", Description: "-3 ability cooldown", ClassID: "oracle", Tier: TierAdept, Branch: "B", Kind: SkillAbilityUpgrade, CooldownReduce: 3},
	{ID: "ora_t1b_kh", Name: "Crystal Leech", Description: "+2 HP on kill", ClassID: "oracle", Tier: TierAdept, Branch: "B", Kind: SkillPassiveProc, KillHealBonus: 2},
	{ID: "ora_t1b_hp", Name: "Arcane Vitality", Description: "+4 MaxHP", ClassID: "oracle", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusMaxHP: 4},

	// ═══ SYMBIONT ═══

	// Tier 0 — Novice
	{ID: "sym_t0_hp", Name: "Adaptive Membrane", Description: "+4 MaxHP", ClassID: "symbiont", Tier: TierNovice, Kind: SkillPassiveStat, BonusMaxHP: 4},
	{ID: "sym_t0_atk", Name: "Tendril Strike", Description: "+2 ATK", ClassID: "symbiont", Tier: TierNovice, Kind: SkillPassiveStat, BonusATK: 2},
	{ID: "sym_t0_regen", Name: "Accelerated Growth", Description: "-1 regen interval", ClassID: "symbiont", Tier: TierNovice, Kind: SkillAbilityUpgrade, RegenReduce: 1},
	{ID: "sym_t0_def", Name: "Carapace Layer", Description: "+1 DEF", ClassID: "symbiont", Tier: TierNovice, Kind: SkillPassiveStat, BonusDEF: 1},
	{ID: "sym_t0_kh", Name: "Consume", Description: "+1 HP on kill", ClassID: "symbiont", Tier: TierNovice, Kind: SkillPassiveProc, KillHealBonus: 1},

	// Tier 1 — Adept, Branch A (Hivemind)
	{ID: "sym_t1a_atk", Name: "Swarm Tendrils", Description: "+3 ATK", ClassID: "symbiont", Tier: TierAdept, Branch: "A", Kind: SkillPassiveStat, BonusATK: 3},
	{ID: "sym_t1a_kh", Name: "Devour", Description: "+3 HP on kill", ClassID: "symbiont", Tier: TierAdept, Branch: "A", Kind: SkillPassiveProc, KillHealBonus: 3},
	{ID: "sym_t1a_regen", Name: "Rapid Mitosis", Description: "-2 regen interval", ClassID: "symbiont", Tier: TierAdept, Branch: "A", Kind: SkillAbilityUpgrade, RegenReduce: 2},
	{ID: "sym_t1a_cd", Name: "Surge Catalyst", Description: "-3 ability cooldown", ClassID: "symbiont", Tier: TierAdept, Branch: "A", Kind: SkillAbilityUpgrade, CooldownReduce: 3},
	{ID: "sym_t1a_hp", Name: "Bio-Shield", Description: "+5 MaxHP", ClassID: "symbiont", Tier: TierAdept, Branch: "A", Kind: SkillPassiveStat, BonusMaxHP: 5},

	// Tier 1 — Adept, Branch B (Carapace)
	{ID: "sym_t1b_def", Name: "Hardened Carapace", Description: "+3 DEF", ClassID: "symbiont", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusDEF: 3},
	{ID: "sym_t1b_hp", Name: "Massive Growth", Description: "+7 MaxHP", ClassID: "symbiont", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusMaxHP: 7},
	{ID: "sym_t1b_thorns", Name: "Toxic Spines", Description: "+2 thorns damage", ClassID: "symbiont", Tier: TierAdept, Branch: "B", Kind: SkillPassiveProc, ThornsDamage: 2},
	{ID: "sym_t1b_regen", Name: "Deep Symbiosis", Description: "-2 regen interval", ClassID: "symbiont", Tier: TierAdept, Branch: "B", Kind: SkillAbilityUpgrade, RegenReduce: 2},
	{ID: "sym_t1b_def2", Name: "Organic Fortress", Description: "+2 DEF", ClassID: "symbiont", Tier: TierAdept, Branch: "B", Kind: SkillPassiveStat, BonusDEF: 2},
}

// SkillByID returns the SkillDef with the given ID, or nil if not found.
func SkillByID(id string) *SkillDef {
	for i := range AllSkills {
		if AllSkills[i].ID == id {
			return &AllSkills[i]
		}
	}
	return nil
}

// SkillsForClass returns all skills for the given class ID, optionally filtered
// by tier and branch. Pass branch="" to include both no-branch and all branches.
func SkillsForClass(classID string, tier SkillTier, branch string) []SkillDef {
	var result []SkillDef
	for _, s := range AllSkills {
		if s.ClassID != classID || s.Tier != tier {
			continue
		}
		if s.Branch == "" || s.Branch == branch {
			result = append(result, s)
		}
	}
	return result
}

// AvailableSkills returns skills the player can learn at a given level,
// excluding already-learned skills.
func AvailableSkills(classID string, level int, branch string, learned map[string]bool) []SkillDef {
	tier := TierForLevel(level)
	var result []SkillDef
	// Include current tier and all lower tiers (can still pick unlearned novice skills at higher tiers).
	for t := TierNovice; t <= tier; t++ {
		for _, s := range SkillsForClass(classID, t, branch) {
			if !learned[s.ID] {
				result = append(result, s)
			}
		}
	}
	return result
}
