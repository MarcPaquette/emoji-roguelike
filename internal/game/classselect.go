package game

import (
	"emoji-roguelike/assets"
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// runClassSelect shows the class selection screen and blocks until the player
// picks a class. Returns false if the player quits without selecting.
func (g *Game) runClassSelect() bool {
	selected := 0
	for {
		g.drawClassSelect(selected)
		ev := g.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			g.screen.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyUp:
				selected = (selected - 1 + len(assets.Classes)) % len(assets.Classes)
			case tcell.KeyDown:
				selected = (selected + 1) % len(assets.Classes)
			case tcell.KeyEnter:
				g.selectedClass = assets.Classes[selected]
				g.fovRadius = g.selectedClass.FOVRadius
				return true
			case tcell.KeyEscape:
				if g.confirmQuit(func() { g.drawClassSelect(selected) }) {
					return false
				}
			}
			switch ev.Rune() {
			case 'k', 'K':
				selected = (selected - 1 + len(assets.Classes)) % len(assets.Classes)
			case 'j', 'J':
				selected = (selected + 1) % len(assets.Classes)
			case 'q', 'Q':
				if g.confirmQuit(func() { g.drawClassSelect(selected) }) {
					return false
				}
			case '1', '2', '3', '4', '5', '6':
				idx := int(ev.Rune()-'1')
				if idx >= 0 && idx < len(assets.Classes) {
					g.selectedClass = assets.Classes[idx]
					g.fovRadius = g.selectedClass.FOVRadius
					return true
				}
			}
		}
	}
}

// drawClassSelect renders the full class selection UI to the screen.
func (g *Game) drawClassSelect(selected int) {
	g.screen.Clear()
	w, _ := g.screen.Size()

	titleStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(180, 100, 255)).Bold(true)
	normalStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	dimStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)
	highlightStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.NewRGBColor(180, 100, 255))
	statStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(150, 220, 255))
	passiveStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(255, 200, 50))

	centerText := func(y int, text string, style tcell.Style) {
		x := (w - len([]rune(text))) / 2
		if x < 0 {
			x = 0
		}
		drawScreenText(g.screen, x, y, text, style)
	}

	centerText(1, "✨ THE PRISMATIC SPIRE ✨", titleStyle)
	centerText(2, "Choose your dimensional class", dimStyle)

	// Each class occupies 4 lines + 1 blank = 5 rows. Start at row 4.
	startY := 4
	for i, class := range assets.Classes {
		y := startY + i*5
		prefix := "  "
		lineStyle := normalStyle
		if i == selected {
			prefix = "► "
			lineStyle = highlightStyle
		}

		// Line 1: number + emoji + name
		nameLine := fmt.Sprintf("%s[%d] %s %s", prefix, i+1, class.Emoji, class.Name)
		drawScreenText(g.screen, 2, y, nameLine, lineStyle)

		// Line 2: lore (indented, dimmed)
		loreLine := fmt.Sprintf("      \"%s\"", class.Lore)
		drawScreenText(g.screen, 2, y+1, loreLine, dimStyle)

		// Line 3: stats
		statsLine := fmt.Sprintf("      HP:%-3d ATK:%-2d DEF:%-2d FOV:%-2d", class.MaxHP, class.Attack, class.Defense, class.FOVRadius)
		drawScreenText(g.screen, 2, y+2, statsLine, statStyle)

		// Line 4: passive
		passiveLine := fmt.Sprintf("      Passive: %s", class.PassiveDesc)
		drawScreenText(g.screen, 2, y+3, passiveLine, passiveStyle)
	}

	hintsY := startY + len(assets.Classes)*5 + 1
	centerText(hintsY, "[j/k or ↑/↓] Navigate   [1-6] Quick-select   [Enter] Confirm   [q] Quit", dimStyle)

	g.screen.Show()
}

// drawScreenText writes a string to the screen at (x, y) with the given style.
func drawScreenText(screen tcell.Screen, x, y int, text string, style tcell.Style) {
	col := x
	for _, ch := range text {
		screen.SetContent(col, y, ch, nil, style)
		col++
	}
}
