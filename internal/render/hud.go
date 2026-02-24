package render

import (
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// DrawHUD renders the status bar and message log at the bottom of the screen.
// bonusATK and bonusDEF are the combined effect+equipment bonus values computed by game.go.
// abilityName is the class active ability name; abilityCooldown is turns remaining (0 = ready).
func (r *Renderer) DrawHUD(w *ecs.World, playerID ecs.EntityID, floor int, className string, messages []string, bonusATK, bonusDEF int, abilityName string, abilityCooldown int) {
	_, screenH := r.screen.Size()
	hudY := screenH - 5

	// Separator line.
	r.drawHLine(hudY, tcell.ColorGray)

	// Row 1: HP / ATK / DEF / Floor
	hpText := "HP: ?"
	if c := w.Get(playerID, component.CHealth); c != nil {
		hp := c.(component.Health)
		hpText = fmt.Sprintf("HP: %d/%d", hp.Current, hp.Max)
	}

	atkText := ""
	if c := w.Get(playerID, component.CCombat); c != nil {
		cb := c.(component.Combat)
		if bonusATK != 0 {
			atkText = fmt.Sprintf("  ATK:%d(%+d) DEF:%d(%+d)", cb.Attack, bonusATK, cb.Defense, bonusDEF)
		} else {
			atkText = fmt.Sprintf("  ATK:%d DEF:%d", cb.Attack, cb.Defense)
		}
	}

	classText := ""
	if className != "" {
		classText = fmt.Sprintf("[%s]  ", className)
	}
	floorText := fmt.Sprintf("  Floor: %d  The Prismatic Spire", floor)
	statusLine := classText + hpText + atkText + floorText
	r.drawText(0, hudY+1, statusLine, tcell.StyleDefault.Foreground(tcell.ColorWhite))

	// Row 2: equipped items
	inv := component.Inventory{}
	if c := w.Get(playerID, component.CInventory); c != nil {
		inv = c.(component.Inventory)
	}
	headG := "--"
	if !inv.Head.IsEmpty() {
		headG = inv.Head.Glyph
	}
	bodyG := "--"
	if !inv.Body.IsEmpty() {
		bodyG = inv.Body.Glyph
	}
	feetG := "--"
	if !inv.Feet.IsEmpty() {
		feetG = inv.Feet.Glyph
	}
	weapG := "--"
	if !inv.MainHand.IsEmpty() {
		weapG = inv.MainHand.Glyph
	}
	offG := "--"
	if !inv.OffHand.IsEmpty() {
		offG = inv.OffHand.Glyph
	}
	abilityStatus := ""
	if abilityName != "" {
		if abilityCooldown > 0 {
			abilityStatus = fmt.Sprintf("  [z]%s:%dt", abilityName, abilityCooldown)
		} else {
			abilityStatus = fmt.Sprintf("  [z]%s:RDY", abilityName)
		}
	}
	equipLine := fmt.Sprintf("HEAD:%s  BODY:%s  FEET:%s  WEAP:%s  OFHND:%s  [i]nventory%s",
		headG, bodyG, feetG, weapG, offG, abilityStatus)
	r.drawText(0, hudY+2, equipLine, tcell.StyleDefault.Foreground(tcell.ColorAqua))

	// Rows 3-4: last 2 wrapped message lines
	screenW, _ := r.screen.Size()
	var lines []string
	for _, msg := range messages {
		lines = append(lines, wrapText(msg, screenW)...)
	}
	lineStart := len(lines) - 2
	if lineStart < 0 {
		lineStart = 0
	}
	for i, line := range lines[lineStart:] {
		if i >= 2 {
			break
		}
		r.drawText(0, hudY+3+i, line, tcell.StyleDefault.Foreground(tcell.ColorLightYellow))
	}

	r.screen.Show()
}

// wrapText breaks text into lines that fit within width terminal columns,
// correctly accounting for wide characters (emoji) that occupy 2 columns.
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	runes := []rune(text)
	if runewidth.StringWidth(text) <= width {
		return []string{text}
	}
	var lines []string
	for len(runes) > 0 {
		col := 0
		split := 0
		lastSpace := -1
		for split < len(runes) {
			rw := runewidth.RuneWidth(runes[split])
			if col+rw > width {
				break
			}
			if runes[split] == ' ' {
				lastSpace = split
			}
			col += rw
			split++
		}
		if split == len(runes) {
			lines = append(lines, string(runes))
			break
		}
		if lastSpace > 0 {
			lines = append(lines, string(runes[:lastSpace]))
			runes = runes[lastSpace+1:] // skip the space
		} else {
			lines = append(lines, string(runes[:split]))
			runes = runes[split:]
		}
	}
	return lines
}

func (r *Renderer) drawHLine(y int, color tcell.Color) {
	w, _ := r.screen.Size()
	style := tcell.StyleDefault.Foreground(color)
	for x := 0; x < w; x++ {
		r.screen.SetContent(x, y, '─', nil, style)
	}
}

func (r *Renderer) drawText(x, y int, text string, style tcell.Style) {
	col := x
	for _, ch := range text {
		r.screen.SetContent(col, y, ch, nil, style)
		col++
	}
}
