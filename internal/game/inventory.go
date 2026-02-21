package game

import (
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/factory"
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// runInventoryScreen opens a blocking inventory UI.
// Returns true if a turn was spent (consumable used), false otherwise.
func (g *Game) runInventoryScreen() bool {
	invComp := g.world.Get(g.playerID, component.CInventory)
	if invComp == nil {
		return false
	}
	// Work on a local copy; write back on exit.
	inv := invComp.(component.Inventory)

	panel := 0  // 0 = backpack, 1 = equipment
	cursor := 0 // index within current panel
	statusMsg := ""
	turnUsed := false

	clampCursor := func() {
		if panel == 0 {
			if cursor < 0 {
				cursor = 0
			}
			if cursor >= len(inv.Backpack) {
				cursor = len(inv.Backpack) - 1
			}
			if cursor < 0 {
				cursor = 0
			}
		} else {
			if cursor < 0 {
				cursor = 0
			}
			if cursor > 4 {
				cursor = 4
			}
		}
	}

	for {
		clampCursor()
		g.drawInventoryScreen(inv, panel, cursor, statusMsg)

		ev := g.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			g.screen.Sync()
			continue
		case *tcell.EventKey:
			statusMsg = ""
			switch ev.Key() {
			case tcell.KeyEscape:
				// Save inventory back and close.
				g.world.Add(g.playerID, inv)
				g.recalcPlayerMaxHP()
				return turnUsed

			case tcell.KeyTab:
				panel = 1 - panel
				cursor = 0

			case tcell.KeyUp:
				cursor--
			case tcell.KeyDown:
				cursor++

			case tcell.KeyEnter:
				msg := g.invEquipOrUnequip(&inv, panel, cursor)
				statusMsg = msg

			default:
				switch ev.Rune() {
				case 'k', 'K':
					cursor--
				case 'j', 'J':
					cursor++
				case '\t':
					panel = 1 - panel
					cursor = 0
				case 'e', 'E':
					msg := g.invEquipOrUnequip(&inv, panel, cursor)
					statusMsg = msg
				case 'u', 'U':
					msg, used := g.invUseConsumable(&inv, panel, cursor)
					statusMsg = msg
					if used {
						turnUsed = true
						// Save and close after using a consumable.
						g.world.Add(g.playerID, inv)
						g.recalcPlayerMaxHP()
						return turnUsed
					}
				case 'd', 'D':
					msg := g.invDrop(&inv, panel, &cursor)
					statusMsg = msg
				case 'i', 'I', 'q', 'Q':
					g.world.Add(g.playerID, inv)
					g.recalcPlayerMaxHP()
					return turnUsed
				}
			}
		}
	}
}

// invEquipOrUnequip handles the equip/unequip action in the inventory.
func (g *Game) invEquipOrUnequip(inv *component.Inventory, panel, cursor int) string {
	if panel == 0 {
		// Equip from backpack
		if cursor < 0 || cursor >= len(inv.Backpack) {
			return "Nothing selected."
		}
		item := inv.Backpack[cursor]
		if item.IsConsumable {
			return "Press [u] to use consumables."
		}
		return g.invEquip(inv, cursor)
	}
	// Unequip from equipment panel
	return g.invUnequip(inv, cursor)
}

// invEquip moves a backpack item into its appropriate equipment slot.
func (g *Game) invEquip(inv *component.Inventory, cursor int) string {
	item := inv.Backpack[cursor]

	switch item.Slot {
	case component.SlotHead:
		old := inv.Head
		inv.Backpack = removeAt(inv.Backpack, cursor)
		inv.Head = item
		if !old.IsEmpty() {
			inv.Backpack = append(inv.Backpack, old)
		}
		return fmt.Sprintf("Equipped %s.", item.Name)

	case component.SlotBody:
		old := inv.Body
		inv.Backpack = removeAt(inv.Backpack, cursor)
		inv.Body = item
		if !old.IsEmpty() {
			inv.Backpack = append(inv.Backpack, old)
		}
		return fmt.Sprintf("Equipped %s.", item.Name)

	case component.SlotFeet:
		old := inv.Feet
		inv.Backpack = removeAt(inv.Backpack, cursor)
		inv.Feet = item
		if !old.IsEmpty() {
			inv.Backpack = append(inv.Backpack, old)
		}
		return fmt.Sprintf("Equipped %s.", item.Name)

	case component.SlotOneHand:
		old := inv.MainHand
		inv.Backpack = removeAt(inv.Backpack, cursor)
		inv.MainHand = item
		if !old.IsEmpty() {
			inv.Backpack = append(inv.Backpack, old)
		}
		return fmt.Sprintf("Equipped %s.", item.Name)

	case component.SlotTwoHand:
		// Count extra items that need to go to backpack.
		extra := 0
		if !inv.MainHand.IsEmpty() {
			extra++
		}
		if !inv.OffHand.IsEmpty() {
			extra++
		}
		// After removing cursor item: len-1 slots used. Need len-1+extra <= Capacity.
		if len(inv.Backpack)-1+extra > inv.Capacity {
			return "Not enough backpack space to swap."
		}
		inv.Backpack = removeAt(inv.Backpack, cursor)
		if !inv.OffHand.IsEmpty() {
			inv.Backpack = append(inv.Backpack, inv.OffHand)
			inv.OffHand = component.Item{}
		}
		if !inv.MainHand.IsEmpty() {
			inv.Backpack = append(inv.Backpack, inv.MainHand)
		}
		inv.MainHand = item
		return fmt.Sprintf("Equipped %s (two-handed).", item.Name)

	case component.SlotOffHand:
		if inv.MainHand.Slot == component.SlotTwoHand {
			return "Two-handed weapon occupies the off-hand slot."
		}
		old := inv.OffHand
		inv.Backpack = removeAt(inv.Backpack, cursor)
		inv.OffHand = item
		if !old.IsEmpty() {
			inv.Backpack = append(inv.Backpack, old)
		}
		return fmt.Sprintf("Equipped %s.", item.Name)
	}
	return "Cannot equip that."
}

// invUnequip moves an equipment slot item back to the backpack.
func (g *Game) invUnequip(inv *component.Inventory, cursor int) string {
	if len(inv.Backpack) >= inv.Capacity {
		return "Backpack full — drop something first."
	}
	var item component.Item
	switch cursor {
	case 0:
		if inv.Head.IsEmpty() {
			return "Nothing equipped in HEAD slot."
		}
		item = inv.Head
		inv.Head = component.Item{}
	case 1:
		if inv.Body.IsEmpty() {
			return "Nothing equipped in BODY slot."
		}
		item = inv.Body
		inv.Body = component.Item{}
	case 2:
		if inv.Feet.IsEmpty() {
			return "Nothing equipped in FEET slot."
		}
		item = inv.Feet
		inv.Feet = component.Item{}
	case 3:
		if inv.MainHand.IsEmpty() {
			return "Nothing equipped in WEAP slot."
		}
		item = inv.MainHand
		inv.MainHand = component.Item{}
	case 4:
		if inv.OffHand.IsEmpty() {
			return "Nothing equipped in OFHND slot."
		}
		item = inv.OffHand
		inv.OffHand = component.Item{}
	default:
		return "Invalid slot."
	}
	inv.Backpack = append(inv.Backpack, item)
	return fmt.Sprintf("Unequipped %s.", item.Name)
}

// invUseConsumable uses the selected consumable from the backpack.
// Returns a status message and whether a turn was spent.
func (g *Game) invUseConsumable(inv *component.Inventory, panel, cursor int) (string, bool) {
	if panel != 0 {
		return "Select a backpack item to use.", false
	}
	if cursor < 0 || cursor >= len(inv.Backpack) {
		return "Nothing selected.", false
	}
	item := inv.Backpack[cursor]
	if !item.IsConsumable {
		return "Equipment must be equipped, not used.", false
	}
	inv.Backpack = removeAt(inv.Backpack, cursor)
	g.applyConsumable(item)
	return fmt.Sprintf("Used %s.", item.Name), true
}

// invDrop drops the selected item at the player's position.
func (g *Game) invDrop(inv *component.Inventory, panel int, cursor *int) string {
	pos := g.playerPosition()
	var item component.Item

	if panel == 0 {
		if *cursor < 0 || *cursor >= len(inv.Backpack) {
			return "Nothing selected."
		}
		item = inv.Backpack[*cursor]
		inv.Backpack = removeAt(inv.Backpack, *cursor)
		if *cursor >= len(inv.Backpack) && *cursor > 0 {
			(*cursor)--
		}
	} else {
		switch *cursor {
		case 0:
			if inv.Head.IsEmpty() {
				return "Nothing equipped here."
			}
			item = inv.Head
			inv.Head = component.Item{}
		case 1:
			if inv.Body.IsEmpty() {
				return "Nothing equipped here."
			}
			item = inv.Body
			inv.Body = component.Item{}
		case 2:
			if inv.Feet.IsEmpty() {
				return "Nothing equipped here."
			}
			item = inv.Feet
			inv.Feet = component.Item{}
		case 3:
			if inv.MainHand.IsEmpty() {
				return "Nothing equipped here."
			}
			item = inv.MainHand
			inv.MainHand = component.Item{}
		case 4:
			if inv.OffHand.IsEmpty() {
				return "Nothing equipped here."
			}
			item = inv.OffHand
			inv.OffHand = component.Item{}
		default:
			return "Invalid slot."
		}
	}

	factory.DropItem(g.world, item, pos.X, pos.Y)
	return fmt.Sprintf("Dropped %s.", item.Name)
}

// drawInventoryScreen renders the full inventory UI.
func (g *Game) drawInventoryScreen(inv component.Inventory, panel, cursor int, statusMsg string) {
	g.screen.Clear()
	sw, sh := g.screen.Size()
	mid := sw / 2
	if mid < 30 {
		mid = 30
	}

	white := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	gray := tcell.StyleDefault.Foreground(tcell.ColorGray)
	yellow := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	cyan := tcell.StyleDefault.Foreground(tcell.ColorAqua)
	green := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	highlight := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorAqua)
	dim := tcell.StyleDefault.Foreground(tcell.ColorGray)
	_ = sh

	// Row 0: title + hint
	title := fmt.Sprintf("INVENTORY  [Backpack %d/%d]", len(inv.Backpack), inv.Capacity)
	g.putText(0, 0, title, yellow)
	hints := "[j/k] Move  [Tab] Switch  [e] Equip/Unequip  [u] Use  [d] Drop  [Esc] Close"
	if len(hints) < sw {
		g.putText(sw-len([]rune(hints)), 0, hints, dim)
	}

	// Row 1: separator
	for x := range sw {
		g.screen.SetContent(x, 1, '─', nil, gray)
	}

	// Row 2: column headers
	g.putText(0, 2, "── EQUIPPED ──────────────────", white)
	g.putText(mid, 2, "── BACKPACK ─────────────────", white)

	// Vertical divider
	for y := 2; y <= 12; y++ {
		g.screen.SetContent(mid-1, y, '│', nil, gray)
	}

	// Equipment panel (rows 3–7: 5 slots)
	equipSlots := []struct {
		label string
		item  component.Item
	}{
		{"HEAD ", inv.Head},
		{"BODY ", inv.Body},
		{"FEET ", inv.Feet},
		{"WEAP ", inv.MainHand},
		{"OFHND", inv.OffHand},
	}
	for i, slot := range equipSlots {
		row := 3 + i
		sel := panel == 1 && cursor == i
		style := white
		pfx := "  "
		if sel {
			style = highlight
			pfx = "► "
		}
		itemStr := "--"
		if !slot.item.IsEmpty() {
			bonuses := formatBonuses(slot.item)
			itemStr = slot.item.Glyph + " " + slot.item.Name + bonuses
		}
		line := fmt.Sprintf("%s%s %s", pfx, slot.label, itemStr)
		g.putText(0, row, line, style)
	}

	// Row 8: equipment bonus totals
	atkB := inv.Head.BonusATK + inv.Body.BonusATK + inv.Feet.BonusATK + inv.MainHand.BonusATK + inv.OffHand.BonusATK
	defB := inv.Head.BonusDEF + inv.Body.BonusDEF + inv.Feet.BonusDEF + inv.MainHand.BonusDEF + inv.OffHand.BonusDEF
	hpB := inv.Head.BonusMaxHP + inv.Body.BonusMaxHP + inv.Feet.BonusMaxHP + inv.MainHand.BonusMaxHP + inv.OffHand.BonusMaxHP
	g.putText(0, 8, fmt.Sprintf("  Equip bonus: ATK%+d DEF%+d HP%+d", atkB, defB, hpB), cyan)

	// Backpack panel (rows 3–10: up to 8 items)
	for i, item := range inv.Backpack {
		row := 3 + i
		if row > 10 {
			break
		}
		sel := panel == 0 && cursor == i
		style := white
		pfx := "  "
		if sel {
			style = highlight
			pfx = "► "
		}
		tag := ""
		if item.IsConsumable {
			tag = " [use]"
		}
		bonuses := formatBonuses(item)
		line := fmt.Sprintf("%s[%d] %s %s%s%s", pfx, i, item.Glyph, item.Name, bonuses, tag)
		g.putText(mid, row, line, style)
	}
	if len(inv.Backpack) == 0 {
		g.putText(mid, 3, "  (empty)", dim)
	}

	// Row 11: separator
	for x := range sw {
		g.screen.SetContent(x, 11, '─', nil, gray)
	}

	// Row 12: description of selected item or status message
	var selItem component.Item
	selEmpty := true
	if panel == 0 && cursor < len(inv.Backpack) {
		selItem = inv.Backpack[cursor]
		selEmpty = false
	} else if panel == 1 {
		switch cursor {
		case 0:
			selItem = inv.Head
		case 1:
			selItem = inv.Body
		case 2:
			selItem = inv.Feet
		case 3:
			selItem = inv.MainHand
		case 4:
			selItem = inv.OffHand
		}
		selEmpty = selItem.IsEmpty()
	}

	if !selEmpty {
		slotName := slotLabel(selItem.Slot)
		desc := fmt.Sprintf("%s — %s  ATK%+d DEF%+d MaxHP%+d",
			selItem.Name, slotName, selItem.BonusATK, selItem.BonusDEF, selItem.BonusMaxHP)
		g.putText(0, 12, desc, white)
	}
	if statusMsg != "" {
		g.putText(0, 13, statusMsg, green)
	}

	g.screen.Show()
}

// removeAt returns a new slice with the element at index i removed.
func removeAt(s []component.Item, i int) []component.Item {
	out := make([]component.Item, 0, len(s)-1)
	out = append(out, s[:i]...)
	out = append(out, s[i+1:]...)
	return out
}

// formatBonuses returns a compact bonus string for an item (e.g. " +4A +3D").
func formatBonuses(item component.Item) string {
	if item.IsConsumable {
		return ""
	}
	s := ""
	if item.BonusATK != 0 {
		s += fmt.Sprintf(" %+dA", item.BonusATK)
	}
	if item.BonusDEF != 0 {
		s += fmt.Sprintf(" %+dD", item.BonusDEF)
	}
	if item.BonusMaxHP != 0 {
		s += fmt.Sprintf(" %+dHP", item.BonusMaxHP)
	}
	return s
}

// slotLabel returns a human-readable slot name.
func slotLabel(slot component.ItemSlot) string {
	switch slot {
	case component.SlotConsumable:
		return "Consumable"
	case component.SlotHead:
		return "Head"
	case component.SlotBody:
		return "Body"
	case component.SlotFeet:
		return "Feet"
	case component.SlotOneHand:
		return "One-Hand Weapon"
	case component.SlotTwoHand:
		return "Two-Hand Weapon"
	case component.SlotOffHand:
		return "Off-Hand"
	}
	return "?"
}
