package component

import "emoji-roguelike/internal/ecs"

// ItemSlot categorises where an item can be equipped (or whether it is consumable).
type ItemSlot uint8

const (
	SlotConsumable ItemSlot = iota // 0 — single-use consumable
	SlotHead                       // 1
	SlotBody                       // 2
	SlotFeet                       // 3
	SlotOneHand                    // 4 — allows an off-hand item
	SlotTwoHand                    // 5 — blocks the off-hand slot
	SlotOffHand                    // 6
)

// Item is a plain value struct representing one item (consumable or equipment).
// It is stored by value inside Inventory — not as a separate ECS entity.
type Item struct {
	Name         string
	Glyph        string
	Slot         ItemSlot
	BonusATK     int
	BonusDEF     int
	BonusMaxHP   int
	IsConsumable bool
	EffectKind   uint8 // 0 = none; mirrors EffectKind constants
	EffectMag    int
	EffectDur    int
}

// IsEmpty returns true when this Item is the zero value (empty slot).
func (i Item) IsEmpty() bool { return i.Name == "" }

// CItem is the ECS component type for floor-item entities.
// The wrapped Item is copied into Inventory on pickup; the entity is then destroyed.
const CItem ecs.ComponentType = 13

// CItemComp wraps Item so it can be stored as an ECS component on floor entities.
type CItemComp struct{ Item }

func (CItemComp) Type() ecs.ComponentType { return CItem }
