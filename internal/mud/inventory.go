package mud

import (
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/factory"
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// RunInventory opens the blocking inventory UI for a session.
// eventCh supplies keyboard events from the session's input goroutine.
// Modifies inventory in-place; writes back to ECS on exit.
// Returns true if a consumable was used (a turn was spent).
func (s *Server) RunInventory(sess *Session, eventCh <-chan tcell.Event) bool {
	s.mu.Lock()
	floor, ok := s.floors[sess.FloorNum]
	if !ok {
		s.mu.Unlock()
		return false
	}
	invComp := floor.World.Get(sess.PlayerID, component.CInventory)
	if invComp == nil {
		s.mu.Unlock()
		return false
	}
	inv := invComp.(component.Inventory)
	// Snapshot identity at inventory open: if the player dies/transitions
	// while the modal is open, we discard the stale local copy.
	snapshotFloor := sess.FloorNum
	snapshotPlayer := sess.PlayerID
	s.mu.Unlock()

	panel := 0
	cursor := 0
	statusMsg := ""
	turnUsed := false

	clamp := func() {
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

	save := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		// Only write back if the player hasn't died or changed floors.
		if sess.FloorNum != snapshotFloor || sess.PlayerID != snapshotPlayer {
			return
		}
		if f, ok2 := s.floors[sess.FloorNum]; ok2 {
			f.World.Add(sess.PlayerID, inv)
			recalcMaxHP(f.World, sess)
		}
	}

	for {
		clamp()
		drawInvScreen(sess.Screen, inv, panel, cursor, statusMsg)

		ev, ok := <-eventCh
		if !ok || ev == nil {
			save()
			return turnUsed
		}
		switch ev := ev.(type) {
		case *tcell.EventResize:
			sess.Screen.Sync()
		case *tcell.EventKey:
			statusMsg = ""
			switch ev.Key() {
			case tcell.KeyEscape:
				save()
				return turnUsed
			case tcell.KeyTab:
				panel = 1 - panel
				cursor = 0
			case tcell.KeyUp:
				cursor--
			case tcell.KeyDown:
				cursor++
			case tcell.KeyEnter:
				statusMsg = invEquipOrUnequip(&inv, panel, cursor)
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
					statusMsg = invEquipOrUnequip(&inv, panel, cursor)
				case 'u', 'U':
					msg, used := s.invUseConsumable(sess, &inv, panel, cursor)
					statusMsg = msg
					if used {
						turnUsed = true
						save()
						return turnUsed
					}
				case 'd', 'D':
					statusMsg = s.invDrop(sess, &inv, panel, &cursor)
				case 'i', 'I', 'q', 'Q':
					save()
					return turnUsed
				default:
					if ev.Rune() >= '1' && ev.Rune() <= '9' {
						idx := int(ev.Rune()-'0') - 1
						if idx < len(inv.Backpack) {
							panel = 0
							cursor = idx
						}
					}
				}
			}
		}
	}
}

func (s *Server) invUseConsumable(sess *Session, inv *component.Inventory, panel, cursor int) (string, bool) {
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
	s.mu.Lock()
	if f, ok := s.floors[sess.FloorNum]; ok {
		s.applyConsumableLocked(f, sess, item)
	}
	s.mu.Unlock()
	return fmt.Sprintf("Used %s.", item.Name), true
}

func (s *Server) invDrop(sess *Session, inv *component.Inventory, panel int, cursor *int) string {
	// Determine which item to drop from the local inventory copy first.
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
			item, inv.Head = inv.Head, component.Item{}
		case 1:
			if inv.Body.IsEmpty() {
				return "Nothing equipped here."
			}
			item, inv.Body = inv.Body, component.Item{}
		case 2:
			if inv.Feet.IsEmpty() {
				return "Nothing equipped here."
			}
			item, inv.Feet = inv.Feet, component.Item{}
		case 3:
			if inv.MainHand.IsEmpty() {
				return "Nothing equipped here."
			}
			item, inv.MainHand = inv.MainHand, component.Item{}
		case 4:
			if inv.OffHand.IsEmpty() {
				return "Nothing equipped here."
			}
			item, inv.OffHand = inv.OffHand, component.Item{}
		default:
			return "Invalid slot."
		}
	}

	// Single lock acquisition for both position read and item drop.
	s.mu.Lock()
	defer s.mu.Unlock()
	floor, ok := s.floors[sess.FloorNum]
	if !ok {
		return "Cannot drop here."
	}
	posComp := floor.World.Get(sess.PlayerID, component.CPosition)
	if posComp == nil {
		return "Cannot drop here."
	}
	pos := posComp.(component.Position)
	factory.DropItem(floor.World, item, pos.X, pos.Y)
	return fmt.Sprintf("Dropped %s.", item.Name)
}

// ─── inventory manipulation helpers ──────────────────────────────────────────

func invEquipOrUnequip(inv *component.Inventory, panel, cursor int) string {
	if panel == 0 {
		if cursor < 0 || cursor >= len(inv.Backpack) {
			return "Nothing selected."
		}
		if inv.Backpack[cursor].IsConsumable {
			return "Press [u] to use consumables."
		}
		return invEquip(inv, cursor)
	}
	return invUnequip(inv, cursor)
}

func invEquip(inv *component.Inventory, cursor int) string {
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
		extra := 0
		if !inv.MainHand.IsEmpty() {
			extra++
		}
		if !inv.OffHand.IsEmpty() {
			extra++
		}
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

func invUnequip(inv *component.Inventory, cursor int) string {
	if len(inv.Backpack) >= inv.Capacity {
		return "Backpack full — drop something first."
	}
	var item component.Item
	switch cursor {
	case 0:
		if inv.Head.IsEmpty() {
			return "Nothing equipped in HEAD slot."
		}
		item, inv.Head = inv.Head, component.Item{}
	case 1:
		if inv.Body.IsEmpty() {
			return "Nothing equipped in BODY slot."
		}
		item, inv.Body = inv.Body, component.Item{}
	case 2:
		if inv.Feet.IsEmpty() {
			return "Nothing equipped in FEET slot."
		}
		item, inv.Feet = inv.Feet, component.Item{}
	case 3:
		if inv.MainHand.IsEmpty() {
			return "Nothing equipped in WEAP slot."
		}
		item, inv.MainHand = inv.MainHand, component.Item{}
	case 4:
		if inv.OffHand.IsEmpty() {
			return "Nothing equipped in OFHND slot."
		}
		item, inv.OffHand = inv.OffHand, component.Item{}
	default:
		return "Invalid slot."
	}
	inv.Backpack = append(inv.Backpack, item)
	return fmt.Sprintf("Unequipped %s.", item.Name)
}

func removeAt(s []component.Item, i int) []component.Item {
	out := make([]component.Item, len(s)-1)
	copy(out, s[:i])
	copy(out[i:], s[i+1:])
	return out
}

func formatBonuses(item component.Item) string {
	if item.BonusATK == 0 && item.BonusDEF == 0 && item.BonusMaxHP == 0 {
		return ""
	}
	s := " ("
	if item.BonusATK != 0 {
		s += fmt.Sprintf("ATK%+d", item.BonusATK)
	}
	if item.BonusDEF != 0 {
		if len(s) > 2 {
			s += " "
		}
		s += fmt.Sprintf("DEF%+d", item.BonusDEF)
	}
	if item.BonusMaxHP != 0 {
		if len(s) > 2 {
			s += " "
		}
		s += fmt.Sprintf("HP%+d", item.BonusMaxHP)
	}
	return s + ")"
}

func slotLabel(slot component.ItemSlot) string {
	switch slot {
	case component.SlotHead:
		return "Head"
	case component.SlotBody:
		return "Body"
	case component.SlotFeet:
		return "Feet"
	case component.SlotOneHand:
		return "Main Hand"
	case component.SlotTwoHand:
		return "Two-Handed"
	case component.SlotOffHand:
		return "Off-Hand"
	}
	return "Consumable"
}

// ─── draw ─────────────────────────────────────────────────────────────────────

func drawInvScreen(screen tcell.Screen, inv component.Inventory, panel, cursor int, statusMsg string) {
	screen.Clear()
	sw, _ := screen.Size()
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

	put := func(x, y int, s string, style tcell.Style) { putText(screen, x, y, s, style) }

	title := fmt.Sprintf("INVENTORY  [Backpack %d/%d]", len(inv.Backpack), inv.Capacity)
	put(0, 0, title, yellow)
	hints := "[j/k] Move  [Tab] Switch  [e] Equip  [u] Use  [d] Drop  [Esc] Close"
	if len([]rune(hints)) < sw {
		put(sw-len([]rune(hints)), 0, hints, dim)
	}
	for x := range sw {
		screen.SetContent(x, 1, '─', nil, gray)
	}
	put(0, 2, "── EQUIPPED ──────────────────", white)
	put(mid, 2, "── BACKPACK ─────────────────", white)
	for y := 2; y <= 12; y++ {
		screen.SetContent(mid-1, y, '│', nil, gray)
	}

	equipSlots := []struct {
		label string
		item  component.Item
	}{
		{"HEAD ", inv.Head}, {"BODY ", inv.Body}, {"FEET ", inv.Feet},
		{"WEAP ", inv.MainHand}, {"OFHND", inv.OffHand},
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
			itemStr = slot.item.Glyph + " " + slot.item.Name + formatBonuses(slot.item)
		}
		put(0, row, fmt.Sprintf("%s%s %s", pfx, slot.label, itemStr), style)
	}

	atkB := inv.Head.BonusATK + inv.Body.BonusATK + inv.Feet.BonusATK + inv.MainHand.BonusATK + inv.OffHand.BonusATK
	defB := inv.Head.BonusDEF + inv.Body.BonusDEF + inv.Feet.BonusDEF + inv.MainHand.BonusDEF + inv.OffHand.BonusDEF
	hpB := inv.Head.BonusMaxHP + inv.Body.BonusMaxHP + inv.Feet.BonusMaxHP + inv.MainHand.BonusMaxHP + inv.OffHand.BonusMaxHP
	put(0, 8, fmt.Sprintf("  Equip bonus: ATK%+d DEF%+d HP%+d", atkB, defB, hpB), cyan)

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
		put(mid, row, fmt.Sprintf("%s[%d] %s %s%s%s", pfx, i+1, item.Glyph, item.Name, formatBonuses(item), tag), style)
	}
	if len(inv.Backpack) == 0 {
		put(mid, 3, "  (empty)", dim)
	}

	for x := range sw {
		screen.SetContent(x, 11, '─', nil, gray)
	}

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
		put(0, 12, fmt.Sprintf("%s — %s  ATK%+d DEF%+d MaxHP%+d",
			selItem.Name, slotLabel(selItem.Slot), selItem.BonusATK, selItem.BonusDEF, selItem.BonusMaxHP), white)
	}
	if statusMsg != "" {
		put(0, 13, statusMsg, green)
	}
	screen.Show()
}
