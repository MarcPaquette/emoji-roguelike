package component

import "emoji-roguelike/internal/ecs"

const CFurniture ecs.ComponentType = 15

// Passive ability constants for furniture bonuses.
const (
	PassiveNone        = 0
	PassiveKeenEye     = 1 // +1 FOV radius permanently
	PassiveKillRestore = 2 // restore 1 HP on each kill
	PassiveThorns      = 3 // reflect 1 damage when struck
)

// Furniture is a decorative entity that may grant a one-time permanent bonus
// when the player steps onto it.
type Furniture struct {
	Glyph       string
	Name        string
	Description string
	BonusATK    int
	BonusDEF    int
	BonusMaxHP  int
	HealHP      int
	PassiveKind int  // one of the Passive* constants above
	Used        bool // prevents repeat bonus triggers
}

func (Furniture) Type() ecs.ComponentType { return CFurniture }
