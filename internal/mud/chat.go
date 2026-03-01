package mud

import (
	"emoji-roguelike/internal/component"
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// Chat constants.
const (
	ChatRange        = 10 // Chebyshev distance for message delivery
	BubbleDuration   = 30 // ticks before a bubble disappears (~3s at 100ms/tick)
	MaxChatLength    = 60 // max runes a player can type
	MaxBubbleDisplay = 30 // max runes shown in a speech bubble
)

// ChatBubble represents a floating speech bubble above a player.
type ChatBubble struct {
	SenderID       int
	SenderName     string
	Text           string
	Color          tcell.Color
	TicksRemaining int
}

// chebyshev returns the Chebyshev (chessboard) distance between two points.
func chebyshev(x1, y1, x2, y2 int) int {
	dx := x1 - x2
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y2
	if dy < 0 {
		dy = -dy
	}
	return max(dx, dy)
}

// decayBubbles decrements TicksRemaining on each bubble and removes expired ones.
func decayBubbles(bubbles []ChatBubble) []ChatBubble {
	n := 0
	for i := range bubbles {
		bubbles[i].TicksRemaining--
		if bubbles[i].TicksRemaining > 0 {
			bubbles[n] = bubbles[i]
			n++
		}
	}
	return bubbles[:n]
}

// truncateBubbleText shortens text for display in speech bubbles.
func truncateBubbleText(text string) string {
	runes := []rune(text)
	if len(runes) <= MaxBubbleDisplay {
		return text
	}
	return string(runes[:MaxBubbleDisplay-1]) + "~"
}

// BroadcastChat sends a chat message from the sender to all nearby players.
// Caller must hold s.mu.
func (s *Server) BroadcastChat(sender *Session, text string) {
	floor, ok := s.floors[sender.FloorNum]
	if !ok {
		return
	}

	// Get sender position.
	posComp := floor.World.Get(sender.PlayerID, component.CPosition)
	if posComp == nil {
		return
	}
	senderPos := posComp.(component.Position)

	bubble := ChatBubble{
		SenderID:       sender.ID,
		SenderName:     sender.Name,
		Text:           truncateBubbleText(text),
		Color:          sender.Color,
		TicksRemaining: BubbleDuration,
	}

	// Sender always sees their own message.
	sender.AddMessage(fmt.Sprintf("You say: \"%s\"", text))
	sender.ChatBubbles = append(sender.ChatBubbles, bubble)

	// Deliver to nearby players on the same floor.
	for _, sess := range s.sessions {
		if sess == sender || sess.FloorNum != sender.FloorNum {
			continue
		}
		rposComp := floor.World.Get(sess.PlayerID, component.CPosition)
		if rposComp == nil {
			continue
		}
		rpos := rposComp.(component.Position)
		if chebyshev(senderPos.X, senderPos.Y, rpos.X, rpos.Y) <= ChatRange {
			sess.AddMessage(fmt.Sprintf("%s says: \"%s\"", sender.Name, text))
			sess.ChatBubbles = append(sess.ChatBubbles, bubble)
		}
	}
}

// RunChat handles the chat input modal. The world keeps ticking and rendering
// while the player types. Returns the typed message and true, or empty and false
// if cancelled.
func (s *Server) RunChat(sess *Session, eventCh <-chan tcell.Event) (string, bool) {
	var buf []rune

	for {
		// Draw the current world frame with chat input overlay.
		s.mu.Lock()
		s.RenderSession(sess)
		s.drawChatInput(sess, buf)
		s.mu.Unlock()
		sess.Screen.Show()

		select {
		case ev, ok := <-eventCh:
			if !ok {
				return "", false
			}
			switch ev := ev.(type) {
			case *tcell.EventResize:
				sess.Screen.Sync()
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEnter:
					if len(buf) == 0 {
						return "", false // empty message = cancel
					}
					return string(buf), true
				case tcell.KeyEscape:
					return "", false
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					if len(buf) > 0 {
						buf = buf[:len(buf)-1]
					}
				case tcell.KeyRune:
					if len(buf) < MaxChatLength {
						buf = append(buf, ev.Rune())
					}
				}
			}

		case <-sess.RenderCh:
			// World ticked — re-render with updated state.
			s.mu.Lock()
			s.RenderSession(sess)
			s.drawChatInput(sess, buf)
			s.mu.Unlock()
			sess.Screen.Show()
		}
	}
}

// drawChatInput renders the "Say: text_" prompt on the bottom row of the HUD.
func (s *Server) drawChatInput(sess *Session, buf []rune) {
	_, sh := sess.Screen.Size()
	y := sh - 1

	prompt := "Say: " + string(buf) + "_"
	style := tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	putText(sess.Screen, 0, y, prompt, style)
}

// drawChatBubbles renders active speech bubbles above sender positions.
// Caller must hold s.mu.
func (s *Server) drawChatBubbles(sess *Session, floor *Floor) {
	if len(sess.ChatBubbles) == 0 || sess.Renderer == nil {
		return
	}

	for _, bubble := range sess.ChatBubbles {
		// Find the sender's current position.
		var senderSess *Session
		for _, other := range s.sessions {
			if other.ID == bubble.SenderID && other.FloorNum == sess.FloorNum {
				senderSess = other
				break
			}
		}
		if senderSess == nil {
			continue
		}

		posComp := floor.World.Get(senderSess.PlayerID, component.CPosition)
		if posComp == nil {
			continue
		}
		pos := posComp.(component.Position)

		sx, sy, visible := sess.Renderer.WorldToScreen(pos.X, pos.Y)
		if !visible || sy <= 0 {
			continue
		}

		// Draw bubble 1 row above the sender, centered on their glyph.
		bubbleY := sy - 1
		text := bubble.Text
		textRunes := []rune(text)
		// Center on sx (which is the left column of the 2-wide emoji glyph).
		// sx+1 is the glyph center. Offset by half the text length.
		startX := sx + 1 - len(textRunes)/2

		style := tcell.StyleDefault.Foreground(bubble.Color).Background(tcell.ColorNavy)
		putText(sess.Screen, startX, bubbleY, text, style)
	}
}
