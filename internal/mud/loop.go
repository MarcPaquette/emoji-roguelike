package mud

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/game"

	"github.com/gdamore/tcell/v2"
)

// ClassSelect blocks on the session's screen until the player picks a class.
// Returns false if the player disconnects without selecting.
func ClassSelect(screen tcell.Screen) (assets.ClassDef, bool) {
	selected := 0
	for {
		game.DrawClassSelectScreen(screen, selected)
		ev := screen.PollEvent()
		if ev == nil {
			return assets.ClassDef{}, false
		}
		switch ev := ev.(type) {
		case *tcell.EventResize:
			screen.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyUp:
				selected = (selected - 1 + len(assets.Classes)) % len(assets.Classes)
			case tcell.KeyDown:
				selected = (selected + 1) % len(assets.Classes)
			case tcell.KeyEnter:
				return assets.Classes[selected], true
			case tcell.KeyEscape:
				return assets.ClassDef{}, false
			}
			switch ev.Rune() {
			case 'k', 'K':
				selected = (selected - 1 + len(assets.Classes)) % len(assets.Classes)
			case 'j', 'J':
				selected = (selected + 1) % len(assets.Classes)
			case 'q', 'Q':
				return assets.ClassDef{}, false
			case '1', '2', '3', '4', '5', '6':
				idx := int(ev.Rune() - '1')
				if idx >= 0 && idx < len(assets.Classes) {
					return assets.Classes[idx], true
				}
			}
		}
	}
}

// RunLoop is the per-session goroutine. It reads input, triggers renders, and
// handles modal screens (inventory, help). Blocks until the player disconnects.
func (s *Server) RunLoop(sess *Session) {
	// Start an async input reader goroutine.
	eventCh := make(chan tcell.Event, 32)
	go func() {
		for {
			ev := sess.Screen.PollEvent()
			if ev == nil {
				close(eventCh)
				return
			}
			eventCh <- ev
		}
	}()

	for {
		select {
		case ev, ok := <-eventCh:
			if !ok {
				return // screen closed / disconnected
			}
			switch ev := ev.(type) {
			case *tcell.EventResize:
				sess.Screen.Sync()
				// Trigger immediate re-render after resize.
				select {
				case sess.RenderCh <- struct{}{}:
				default:
				}
			case *tcell.EventKey:
				action := keyToAction(ev)
				switch action {
				case ActionQuit:
					if confirmQuit(sess, eventCh) {
						return
					}
					// Player cancelled — re-render and continue.
					select {
					case sess.RenderCh <- struct{}{}:
					default:
					}
				case ActionInventory:
					if sess.GetDeathCountdown() == 0 {
						s.RunInventory(sess, eventCh)
						// Trigger re-render after inventory closes.
						select {
						case sess.RenderCh <- struct{}{}:
						default:
						}
					}
				case ActionHelp:
					if sess.GetDeathCountdown() == 0 {
						runHelp(sess, eventCh)
						select {
						case sess.RenderCh <- struct{}{}:
						default:
						}
					}
				default:
					sess.SetAction(action)
				}
			}

		case <-sess.RenderCh:
			s.mu.Lock()
			pendingNPC := sess.PendingNPC
			sess.PendingNPC = 0 // clear under lock
			s.RenderSession(sess)
			s.mu.Unlock()
			sess.Screen.Show()

			// Interactive victory screen: countdown finished, waiting for input.
			if sess.IsVictory() && sess.GetDeathCountdown() == 0 {
				if s.runVictory(sess, eventCh) {
					// Player chose to restart — respawn to Emberveil.
					s.mu.Lock()
					sess.ClearVictory()
					s.respawnLocked(sess)
					s.mu.Unlock()
					select {
					case sess.RenderCh <- struct{}{}:
					default:
					}
				} else {
					return // player chose to quit
				}
				continue
			}

			if pendingNPC != 0 && sess.GetDeathCountdown() == 0 {
				s.RunShop(sess, eventCh)
				// Trigger re-render after shop closes.
				select {
				case sess.RenderCh <- struct{}{}:
				default:
				}
			}
		}
	}
}

// runHelp shows a keybinding reference overlay. Any key dismisses it.
func runHelp(sess *Session, eventCh <-chan tcell.Event) {
	lines := []string{
		"── Movement ──────────────────────────",
		"  Arrow keys / hjkl   Cardinal",
		"  yubn / 7 9 1 3      Diagonal",
		"  5 / .               Wait a turn",
		"",
		"── Actions ───────────────────────────",
		"  g / ,               Pick up item",
		"  i                   Inventory",
		"  Enter               Use stairs",
		"  z                   Special ability",
		"",
		"── Stairs (alternate) ────────────────",
		"  >                   Descend",
		"  <                   Ascend",
		"",
		"── Game ──────────────────────────────",
		"  q / Esc             Disconnect",
		"  ?                   This help",
		"",
		"  [any key to close]",
	}

	header := " Controls "
	width := 42
	hdrStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	bodyStyle := tcell.StyleDefault.Foreground(tcell.ColorSilver)
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)

	draw := func() {
		sess.Screen.Clear()
		sw, sh := sess.Screen.Size()
		boxH := len(lines) + 3
		x0 := (sw - width) / 2
		y0 := (sh - boxH) / 2

		for col := x0; col < x0+width; col++ {
			sess.Screen.SetContent(col, y0, '─', nil, borderStyle)
			sess.Screen.SetContent(col, y0+boxH-1, '─', nil, borderStyle)
		}
		for row := y0; row < y0+boxH; row++ {
			sess.Screen.SetContent(x0, row, '│', nil, borderStyle)
			sess.Screen.SetContent(x0+width-1, row, '│', nil, borderStyle)
		}
		sess.Screen.SetContent(x0, y0, '┌', nil, borderStyle)
		sess.Screen.SetContent(x0+width-1, y0, '┐', nil, borderStyle)
		sess.Screen.SetContent(x0, y0+boxH-1, '└', nil, borderStyle)
		sess.Screen.SetContent(x0+width-1, y0+boxH-1, '┘', nil, borderStyle)

		hx := x0 + (width-len([]rune(header)))/2
		for i, r := range header {
			sess.Screen.SetContent(hx+i, y0, r, nil, hdrStyle)
		}
		for i, line := range lines {
			x := x0 + 2
			for _, r := range line {
				sess.Screen.SetContent(x, y0+1+i, r, nil, bodyStyle)
				x++
			}
		}
		sess.Screen.Show()
	}

	for {
		draw()
		ev, ok := <-eventCh
		if !ok {
			return
		}
		switch ev.(type) {
		case *tcell.EventResize:
			sess.Screen.Sync()
		case *tcell.EventKey:
			return
		}
	}
}

// runVictory blocks on input until the player presses [R] (restart) or [Q] (quit).
// Returns true for restart, false for quit/disconnect.
func (s *Server) runVictory(sess *Session, eventCh <-chan tcell.Event) bool {
	for {
		// Render the victory screen (which includes [R]/[Q] prompt since countdown == 0).
		s.mu.Lock()
		drawVictoryScreen(sess)
		s.mu.Unlock()
		sess.Screen.Show()

		ev, ok := <-eventCh
		if !ok {
			return false // disconnected
		}
		switch ev := ev.(type) {
		case *tcell.EventResize:
			sess.Screen.Sync()
		case *tcell.EventKey:
			switch ev.Rune() {
			case 'r', 'R':
				return true
			case 'q', 'Q':
				return false
			}
			if ev.Key() == tcell.KeyEscape {
				return false
			}
		}
	}
}

// confirmQuit shows a "Really quit? (y/n)" prompt. Returns true if confirmed.
func confirmQuit(sess *Session, eventCh <-chan tcell.Event) bool {
	prompt := " Really disconnect? (y/n) "
	width := len([]rune(prompt)) + 4
	hdrStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)

	draw := func() {
		sess.Screen.Clear()
		sw, sh := sess.Screen.Size()
		boxH := 3
		x0 := (sw - width) / 2
		y0 := (sh - boxH) / 2

		for col := x0; col < x0+width; col++ {
			sess.Screen.SetContent(col, y0, '─', nil, borderStyle)
			sess.Screen.SetContent(col, y0+boxH-1, '─', nil, borderStyle)
		}
		for row := y0; row < y0+boxH; row++ {
			sess.Screen.SetContent(x0, row, '│', nil, borderStyle)
			sess.Screen.SetContent(x0+width-1, row, '│', nil, borderStyle)
		}
		sess.Screen.SetContent(x0, y0, '┌', nil, borderStyle)
		sess.Screen.SetContent(x0+width-1, y0, '┐', nil, borderStyle)
		sess.Screen.SetContent(x0, y0+boxH-1, '└', nil, borderStyle)
		sess.Screen.SetContent(x0+width-1, y0+boxH-1, '┘', nil, borderStyle)

		px := x0 + 2
		for _, r := range prompt {
			sess.Screen.SetContent(px, y0+1, r, nil, hdrStyle)
			px++
		}
		sess.Screen.Show()
	}

	for {
		draw()
		ev, ok := <-eventCh
		if !ok {
			return true // disconnected
		}
		switch ev := ev.(type) {
		case *tcell.EventResize:
			sess.Screen.Sync()
		case *tcell.EventKey:
			switch ev.Rune() {
			case 'y', 'Y':
				return true
			default:
				return false
			}
		}
	}
}
