package mud

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/component"
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// RunShop opens the blocking shop UI for a session.
// eventCh supplies keyboard events from the session's input goroutine.
// The player presses a–h to buy an item, Esc/q to close.
func (s *Server) RunShop(sess *Session, eventCh <-chan tcell.Event) {
	catalogue := assets.ShopCatalogue
	cursor := 0
	statusMsg := ""

	for {
		drawShopScreen(sess.Screen, catalogue, cursor, sess.Gold, statusMsg)

		ev, ok := <-eventCh
		if !ok || ev == nil {
			return
		}
		statusMsg = ""
		switch ev := ev.(type) {
		case *tcell.EventResize:
			sess.Screen.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
				return
			case tcell.KeyUp:
				if cursor > 0 {
					cursor--
				}
			case tcell.KeyDown:
				if cursor < len(catalogue)-1 {
					cursor++
				}
			case tcell.KeyEnter:
				statusMsg = s.shopBuy(sess, catalogue, cursor)
			default:
				r := ev.Rune()
				switch {
				case r == 'q' || r == 'Q':
					return
				case r == 'k' || r == 'K':
					if cursor > 0 {
						cursor--
					}
				case r == 'j' || r == 'J':
					if cursor < len(catalogue)-1 {
						cursor++
					}
				case r >= 'a' && r <= 'h':
					idx := int(r - 'a')
					if idx < len(catalogue) {
						cursor = idx
						statusMsg = s.shopBuy(sess, catalogue, idx)
					}
				}
			}
		}
	}
}

// shopBuy attempts to purchase the item at the given catalogue index.
func (s *Server) shopBuy(sess *Session, catalogue []assets.ShopEntry, idx int) string {
	if idx < 0 || idx >= len(catalogue) {
		return "Nothing selected."
	}
	entry := catalogue[idx]
	if sess.Gold < entry.Price {
		return fmt.Sprintf("Not enough gold. (%d💰 needed, you have %d💰)", entry.Price, sess.Gold)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	floor, ok := s.floors[sess.FloorNum]
	if !ok {
		return "Cannot buy here."
	}
	invComp := floor.World.Get(sess.PlayerID, component.CInventory)
	if invComp == nil {
		return "Cannot buy here."
	}
	inv := invComp.(component.Inventory)
	if len(inv.Backpack) >= inv.Capacity {
		return "Backpack full! Drop something first."
	}

	item := shopEntryToItem(entry)
	inv.Backpack = append(inv.Backpack, item)
	floor.World.Add(sess.PlayerID, inv)
	sess.Gold -= entry.Price
	return fmt.Sprintf("Bought %s %s. (%d💰 remaining)", entry.Glyph, entry.Name, sess.Gold)
}

// shopEntryToItem converts a ShopEntry into a component.Item.
func shopEntryToItem(e assets.ShopEntry) component.Item {
	if e.IsConsumable {
		return component.Item{
			Name:         e.Name,
			Glyph:        e.Glyph,
			Slot:         component.SlotConsumable,
			IsConsumable: true,
		}
	}
	slot := slotStringToItemSlot(e.Slot)
	return component.Item{
		Name:         e.Name,
		Glyph:        e.Glyph,
		Slot:         slot,
		BonusATK:     e.BonusATK,
		BonusDEF:     e.BonusDEF,
		BonusMaxHP:   e.BonusMaxHP,
		IsConsumable: false,
	}
}

// slotStringToItemSlot converts the string slot name used in ShopEntry to
// the component.ItemSlot constant.
func slotStringToItemSlot(s string) component.ItemSlot {
	switch s {
	case "head":
		return component.SlotHead
	case "body":
		return component.SlotBody
	case "feet":
		return component.SlotFeet
	case "onehand":
		return component.SlotOneHand
	case "twohand":
		return component.SlotTwoHand
	case "offhand":
		return component.SlotOffHand
	}
	return component.SlotConsumable
}

// drawShopScreen renders the shop modal to the session's screen.
func drawShopScreen(screen tcell.Screen, items []assets.ShopEntry, cursor, gold int, statusMsg string) {
	screen.Clear()
	sw, _ := screen.Size()

	white := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	gray := tcell.StyleDefault.Foreground(tcell.ColorGray)
	yellow := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	green := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	highlight := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorAqua)
	dim := tcell.StyleDefault.Foreground(tcell.ColorGray)

	put := func(x, y int, s string, style tcell.Style) { putText(screen, x, y, s, style) }

	put(0, 0, fmt.Sprintf("🛍️ YEVA'S PROVISIONS  [You have %d💰]", gold), yellow)
	hints := "[j/k] Move  [a-h] Buy  [Enter] Buy selected  [Esc] Close"
	if len([]rune(hints)) < sw {
		put(sw-len([]rune(hints)), 0, hints, dim)
	}
	for x := range sw {
		screen.SetContent(x, 1, '─', nil, gray)
	}
	put(0, 2, "  #  Item                          Price", white)
	for x := range sw {
		screen.SetContent(x, 3, '─', nil, gray)
	}

	for i, item := range items {
		row := 4 + i
		sel := cursor == i
		style := white
		pfx := "  "
		if sel {
			style = highlight
			pfx = "► "
		}
		tag := "consumable"
		if !item.IsConsumable {
			tag = "equip"
			if item.BonusATK != 0 {
				tag += fmt.Sprintf(" ATK%+d", item.BonusATK)
			}
			if item.BonusDEF != 0 {
				tag += fmt.Sprintf(" DEF%+d", item.BonusDEF)
			}
			if item.BonusMaxHP != 0 {
				tag += fmt.Sprintf(" HP%+d", item.BonusMaxHP)
			}
		}
		line := fmt.Sprintf("%s[%c] %s %-20s  %3d💰  [%s]",
			pfx, 'a'+rune(i), item.Glyph, item.Name, item.Price, tag)
		put(0, row, line, style)
	}

	for x := range sw {
		screen.SetContent(x, 4+len(items), '─', nil, gray)
	}
	if statusMsg != "" {
		put(0, 5+len(items), statusMsg, green)
	}
	screen.Show()
}
