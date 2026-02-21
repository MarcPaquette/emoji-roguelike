package component

import "emoji-roguelike/internal/ecs"

const CInventory ecs.ComponentType = 6

// Inventory holds a player's backpack and equipped items.
// Next available ComponentType: 14.
type Inventory struct {
	Backpack []Item // carried items, up to Capacity
	Capacity int
	// Equipment slots — zero Item (IsEmpty()==true) means nothing equipped.
	Head     Item
	Body     Item
	Feet     Item
	MainHand Item // SlotOneHand or SlotTwoHand
	OffHand  Item // blocked when MainHand.Slot == SlotTwoHand
}

func (Inventory) Type() ecs.ComponentType { return CInventory }
