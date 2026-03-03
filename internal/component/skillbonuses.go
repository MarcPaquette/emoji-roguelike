package component

import "emoji-roguelike/internal/ecs"

const CSkillBonuses ecs.ComponentType = 18

// SkillBonuses aggregates all skill-derived stat bonuses for a player entity.
type SkillBonuses struct {
	BonusATK       int // flat ATK added to combat
	BonusDEF       int // flat DEF added to combat
	BonusMaxHP     int // added to base MaxHP
	BonusFOV       int // added to FOV radius
	DodgeChance    int // 0-100, checked before damage
	KillHealBonus  int // flat HP restored on kill
	KillHealAdd    int // added to class KillHealChance %
	ThornsDamage   int // damage reflected per hit taken
	CooldownReduce int // subtracted from z-ability cooldown
	RegenReduce    int // subtracted from passive regen interval
}

func (SkillBonuses) Type() ecs.ComponentType { return CSkillBonuses }
