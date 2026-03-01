package mud

import (
	"emoji-roguelike/assets"
	"emoji-roguelike/internal/component"
	"emoji-roguelike/internal/ecs"
	"emoji-roguelike/internal/gamemap"
	"math/rand"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ─── Chebyshev distance ──────────────────────────────────────────────────────

func TestChebyshevDistance(t *testing.T) {
	cases := []struct {
		name                 string
		x1, y1, x2, y2, want int
	}{
		{"same point", 5, 5, 5, 5, 0},
		{"horizontal", 0, 0, 3, 0, 3},
		{"vertical", 0, 0, 0, 7, 7},
		{"diagonal", 0, 0, 4, 4, 4},
		{"mixed", 1, 2, 4, 9, 7},
		{"negative coords", -3, -3, 3, 3, 6},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := chebyshev(tc.x1, tc.y1, tc.x2, tc.y2)
			if got != tc.want {
				t.Errorf("chebyshev(%d,%d,%d,%d) = %d, want %d",
					tc.x1, tc.y1, tc.x2, tc.y2, got, tc.want)
			}
		})
	}
}

// ─── Decay bubbles ───────────────────────────────────────────────────────────

func TestDecayBubblesRemovesExpired(t *testing.T) {
	bubbles := []ChatBubble{
		{SenderName: "A", TicksRemaining: 1},  // will be removed (decremented to 0)
		{SenderName: "B", TicksRemaining: 5},  // will remain at 4
		{SenderName: "C", TicksRemaining: 2},  // will remain at 1
	}
	result := decayBubbles(bubbles)
	if len(result) != 2 {
		t.Fatalf("expected 2 remaining, got %d", len(result))
	}
	if result[0].SenderName != "B" || result[0].TicksRemaining != 4 {
		t.Errorf("result[0] = %+v, want B with 4 ticks", result[0])
	}
	if result[1].SenderName != "C" || result[1].TicksRemaining != 1 {
		t.Errorf("result[1] = %+v, want C with 1 tick", result[1])
	}
}

func TestDecayBubblesAllExpired(t *testing.T) {
	bubbles := []ChatBubble{
		{TicksRemaining: 1},
		{TicksRemaining: 1},
	}
	result := decayBubbles(bubbles)
	if len(result) != 0 {
		t.Errorf("expected empty, got %d bubbles", len(result))
	}
}

func TestDecayBubblesEmptySlice(t *testing.T) {
	result := decayBubbles(nil)
	if len(result) != 0 {
		t.Errorf("expected empty, got %d bubbles", len(result))
	}
	result = decayBubbles([]ChatBubble{})
	if len(result) != 0 {
		t.Errorf("expected empty, got %d bubbles", len(result))
	}
}

// ─── Bubble text truncation ──────────────────────────────────────────────────

func TestBubbleTextTruncation(t *testing.T) {
	short := "Hello!"
	if got := truncateBubbleText(short); got != short {
		t.Errorf("short text should be unchanged, got %q", got)
	}

	exact := strings.Repeat("a", MaxBubbleDisplay)
	if got := truncateBubbleText(exact); got != exact {
		t.Errorf("exact-length text should be unchanged, got length %d", len([]rune(got)))
	}

	long := strings.Repeat("x", MaxBubbleDisplay+10)
	got := truncateBubbleText(long)
	runes := []rune(got)
	if len(runes) != MaxBubbleDisplay {
		t.Errorf("truncated length = %d, want %d", len(runes), MaxBubbleDisplay)
	}
	if runes[len(runes)-1] != '~' {
		t.Errorf("truncated text should end with '~', got %q", got)
	}
}

// ─── Broadcast helpers ───────────────────────────────────────────────────────

// chatTestSetup creates a server with a flat open floor and N sessions placed
// at the given positions. Returns the server and sessions slice.
func chatTestSetup(positions [][2]int) (*Server, []*Session) {
	rng := rand.New(rand.NewSource(42))
	srv := &Server{
		floors:   make(map[int]*Floor),
		sessions: nil,
		rng:      rng,
	}

	// Create a simple open floor.
	w := ecs.NewWorld()
	gmap := gamemap.New(50, 50)
	for y := range 50 {
		for x := range 50 {
			gmap.Set(x, y, gamemap.MakeFloor())
		}
	}
	floor := &Floor{
		Num:             1,
		World:           w,
		GMap:            gmap,
		Rng:             rng,
		RespawnCooldown: -1,
	}
	srv.floors[1] = floor

	var sessions []*Session
	for i, pos := range positions {
		screen := newSimScreen()
		color := playerColors[i%len(playerColors)]
		sess := NewSession(i, "Player"+string(rune('A'+i)), color, screen)
		sess.Class = assets.Classes[0]
		sess.FloorNum = 1

		// Create player entity at position.
		pid := w.CreateEntity()
		w.Add(pid, component.Position{X: pos[0], Y: pos[1]})
		w.Add(pid, component.TagPlayer{})
		sess.PlayerID = pid

		srv.sessions = append(srv.sessions, sess)
		sessions = append(sessions, sess)
	}

	return srv, sessions
}

func TestBroadcastChatWithinRange(t *testing.T) {
	// Two players 5 tiles apart (within ChatRange=10).
	srv, sessions := chatTestSetup([][2]int{{10, 10}, {15, 10}})
	sender, receiver := sessions[0], sessions[1]

	srv.BroadcastChat(sender, "Hello!")

	// Sender gets "You say:" message.
	found := false
	for _, msg := range sender.Messages {
		if strings.Contains(msg, "You say:") && strings.Contains(msg, "Hello!") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("sender should see 'You say:' message, got %v", sender.Messages)
	}

	// Receiver gets "PlayerA says:" message.
	found = false
	for _, msg := range receiver.Messages {
		if strings.Contains(msg, "PlayerA says:") && strings.Contains(msg, "Hello!") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("receiver should see 'PlayerA says:' message, got %v", receiver.Messages)
	}

	// Both should have bubbles.
	if len(sender.ChatBubbles) != 1 {
		t.Errorf("sender should have 1 bubble, got %d", len(sender.ChatBubbles))
	}
	if len(receiver.ChatBubbles) != 1 {
		t.Errorf("receiver should have 1 bubble, got %d", len(receiver.ChatBubbles))
	}
}

func TestBroadcastChatOutOfRange(t *testing.T) {
	// Two players 15 tiles apart (outside ChatRange=10).
	srv, sessions := chatTestSetup([][2]int{{5, 5}, {20, 5}})
	sender, receiver := sessions[0], sessions[1]

	srv.BroadcastChat(sender, "Can you hear me?")

	// Sender always gets their message.
	if len(sender.ChatBubbles) != 1 {
		t.Errorf("sender should have 1 bubble, got %d", len(sender.ChatBubbles))
	}

	// Receiver should NOT get the message.
	if len(receiver.ChatBubbles) != 0 {
		t.Errorf("receiver should have 0 bubbles, got %d", len(receiver.ChatBubbles))
	}
	for _, msg := range receiver.Messages {
		if strings.Contains(msg, "says:") {
			t.Errorf("out-of-range receiver should not get chat message, got %v", receiver.Messages)
			break
		}
	}
}

func TestBroadcastChatDifferentFloor(t *testing.T) {
	srv, sessions := chatTestSetup([][2]int{{10, 10}, {10, 10}})
	sender, receiver := sessions[0], sessions[1]

	// Move receiver to a different floor.
	receiver.FloorNum = 2

	srv.BroadcastChat(sender, "Anyone here?")

	if len(sender.ChatBubbles) != 1 {
		t.Errorf("sender should have 1 bubble, got %d", len(sender.ChatBubbles))
	}
	if len(receiver.ChatBubbles) != 0 {
		t.Errorf("different-floor receiver should have 0 bubbles, got %d", len(receiver.ChatBubbles))
	}
}

func TestBroadcastChatSenderAlwaysReceives(t *testing.T) {
	// Single player — sender is the only one.
	srv, sessions := chatTestSetup([][2]int{{10, 10}})
	sender := sessions[0]

	srv.BroadcastChat(sender, "Echo!")

	if len(sender.ChatBubbles) != 1 {
		t.Errorf("sender should always get their own bubble, got %d", len(sender.ChatBubbles))
	}
	found := false
	for _, msg := range sender.Messages {
		if strings.Contains(msg, "You say:") {
			found = true
		}
	}
	if !found {
		t.Errorf("sender should see own message")
	}
}

func TestMultipleBubblesPerSession(t *testing.T) {
	// 4 players at the same spot; 3 send messages, receiver gets 3 bubbles.
	srv, sessions := chatTestSetup([][2]int{{10, 10}, {10, 10}, {10, 10}, {10, 10}})
	receiver := sessions[3]

	for i := range 3 {
		srv.BroadcastChat(sessions[i], "msg")
	}

	if len(receiver.ChatBubbles) != 3 {
		t.Errorf("receiver should have 3 bubbles from 3 senders, got %d", len(receiver.ChatBubbles))
	}
}

// ─── Action mapping ──────────────────────────────────────────────────────────

func TestChatActionMapping(t *testing.T) {
	for _, r := range []rune{'t', 'T'} {
		ev := tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
		if got := keyToAction(ev); got != ActionChat {
			t.Errorf("keyToAction(%q) = %v, want ActionChat", r, got)
		}
	}
}
