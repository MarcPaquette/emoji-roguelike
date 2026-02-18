package render

import (
	"emoji-rougelike/internal/component"
	"emoji-rougelike/internal/ecs"
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// DrawHUD renders the status bar and message log at the bottom of the screen.
// className is the selected class name shown at the start of the status line.
func (r *Renderer) DrawHUD(w *ecs.World, playerID ecs.EntityID, floor int, className string, messages []string) {
	_, screenH := r.screen.Size()
	hudY := screenH - 5

	// Separator line.
	r.drawHLine(hudY, tcell.ColorGray)

	// HP bar.
	hpText := "HP: ?"
	if c := w.Get(playerID, component.CHealth); c != nil {
		hp := c.(component.Health)
		hpText = fmt.Sprintf("HP: %d/%d", hp.Current, hp.Max)
	}

	// Combat stats.
	atkText := ""
	if c := w.Get(playerID, component.CCombat); c != nil {
		cb := c.(component.Combat)
		atkText = fmt.Sprintf("  ATK:%d DEF:%d", cb.Attack, cb.Defense)
	}

	classText := ""
	if className != "" {
		classText = fmt.Sprintf("[%s]  ", className)
	}
	floorText := fmt.Sprintf("  Floor: %d  The Prismatic Spire", floor)
	statusLine := classText + hpText + atkText + floorText
	r.drawText(0, hudY+1, statusLine, tcell.StyleDefault.Foreground(tcell.ColorWhite))

	// Message log (last 3 messages).
	start := len(messages) - 3
	if start < 0 {
		start = 0
	}
	for i, msg := range messages[start:] {
		r.drawText(0, hudY+2+i, msg, tcell.StyleDefault.Foreground(tcell.ColorLightYellow))
	}

	r.screen.Show()
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
