package mud

import "github.com/gdamore/tcell/v2"

// putText writes a string to the screen starting at (x, y), one rune per
// column. It stops at the right edge of the screen (sw) to avoid overflow.
func putText(scr tcell.Screen, x, y int, s string, st tcell.Style) {
	sw, _ := scr.Size()
	for _, r := range s {
		if x >= sw {
			break
		}
		scr.SetContent(x, y, r, nil, st)
		x++
	}
}
