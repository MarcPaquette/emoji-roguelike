// emoji-roguelike-server starts an SSH server that pairs two players for a
// cooperative game session. Build:
//
//	go build -o emoji-roguelike-server ./cmd/server
//
// Usage:
//
//	./emoji-roguelike-server [--port 2222] [--key server_host_key]
//
// Connect from two terminals:
//
//	ssh -p 2222 localhost
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"emoji-roguelike/internal/game"
	internalssh "emoji-roguelike/internal/ssh"

	"github.com/gdamore/tcell/v2"
	gossh "github.com/gliderlabs/ssh"
	xssh "golang.org/x/crypto/ssh"
)

func main() {
	port := flag.Int("port", 2222, "SSH server port")
	keyFile := flag.String("key", "server_host_key", "Path to the PEM-encoded host key (auto-generated if absent)")
	flag.Parse()

	signer := loadOrCreateHostKey(*keyFile)
	l := newLobby()

	srv := &gossh.Server{
		Addr: fmt.Sprintf(":%d", *port),
		Handler: func(s gossh.Session) {
			l.handleSession(s)
		},
		// Accept PTY requests from any client.
		PtyCallback: func(_ gossh.Context, _ gossh.Pty) bool { return true },
		// Accept any authentication — appropriate for a private home server.
		// Add gossh.PublicKeyAuth or gossh.PasswordAuth options for real auth.
		HostSigners: []gossh.Signer{signer},
	}

	log.Printf("emoji-roguelike SSH server listening on :%d", *port)
	log.Printf("Connect from two terminals:  ssh -p %d -o StrictHostKeyChecking=no localhost", *port)
	log.Fatal(srv.ListenAndServe())
}

// ─── lobby ──────────────────────────────────────────────────────────────────

// lobby pairs incoming SSH sessions into cooperative game sessions.
// At most one game runs at a time; excess players are told to wait.
type lobby struct {
	mu     sync.Mutex
	waiter *waitEntry
}

// waitEntry represents a first player waiting for a partner.
type waitEntry struct {
	screen   tcell.Screen
	p2Ready  chan tcell.Screen // P1 receives P2's screen here
	gameDone chan struct{}     // closed when the game finishes
}

func newLobby() *lobby { return &lobby{} }

// handleSession is the gliderlabs SSH handler for one connection.
// It blocks for the duration of the connection so the SSH session stays open.
func (l *lobby) handleSession(s gossh.Session) {
	pty, winCh, hasPTY := s.Pty()
	if !hasPTY {
		fmt.Fprintln(s, "This game requires a PTY. Connect with: ssh -t -p 2222 <host>")
		return
	}

	// Determine the terminal type from the session environment.
	term := "xterm-256color"
	for _, env := range s.Environ() {
		if strings.HasPrefix(env, "TERM=") {
			term = env[5:]
			break
		}
	}

	// Create a tcell screen backed by this SSH session.
	// TERM must be set in the process environment before NewTerminfoScreenFromTty.
	tty := internalssh.NewSessionTty(s, pty, winCh)
	termMu.Lock()
	_ = os.Setenv("TERM", term)
	screen, err := tcell.NewTerminfoScreenFromTty(tty)
	termMu.Unlock()
	if err != nil {
		fmt.Fprintf(s, "Terminal setup failed: %v\n", err)
		return
	}
	if err := screen.Init(); err != nil {
		fmt.Fprintf(s, "Screen init failed: %v\n", err)
		return
	}

	l.mu.Lock()
	if l.waiter == nil {
		// First player: register and wait for a partner.
		w := &waitEntry{
			screen:   screen,
			p2Ready:  make(chan tcell.Screen, 1),
			gameDone: make(chan struct{}),
		}
		l.waiter = w
		l.mu.Unlock()

		showWaiting(screen)
		p2Screen := <-w.p2Ready

		// P1's goroutine drives the game.
		screens := [2]tcell.Screen{screen, p2Screen}
		game.NewCoopGame(screens).Run()
		close(w.gameDone)
	} else {
		// Second player: hand our screen to P1 and block until the game ends.
		w := l.waiter
		l.waiter = nil
		l.mu.Unlock()

		w.p2Ready <- screen
		<-w.gameDone // keep SSH session alive until game finishes
	}
}

// termMu protects os.Setenv("TERM") around screen creation.
// Safe because only one game runs at a time.
var termMu sync.Mutex

// showWaiting displays a "waiting for partner" message on the given screen.
func showWaiting(screen tcell.Screen) {
	screen.Clear()
	msg := "Waiting for second player..."
	w, h := screen.Size()
	x := (w - len(msg)) / 2
	y := h / 2
	style := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	for i, r := range msg {
		screen.SetContent(x+i, y, r, nil, style)
	}
	sub := "Connect another terminal:  ssh -p 2222 <host>"
	xs := (w - len(sub)) / 2
	for i, r := range sub {
		screen.SetContent(xs+i, y+2, r, nil, tcell.StyleDefault.Foreground(tcell.ColorGray))
	}
	screen.Show()
}

// ─── host key ───────────────────────────────────────────────────────────────

// loadOrCreateHostKey loads a PEM private key from path, or generates and
// persists a new ed25519 key if the file is absent or unreadable.
func loadOrCreateHostKey(path string) gossh.Signer {
	if data, err := os.ReadFile(path); err == nil {
		if signer, err := xssh.ParsePrivateKey(data); err == nil {
			log.Printf("Loaded host key from %s", path)
			return signer
		}
	}

	log.Printf("Generating new ed25519 host key → %s", path)
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("generate host key: %v", err)
	}
	signer, err := xssh.NewSignerFromKey(key)
	if err != nil {
		log.Fatalf("create signer: %v", err)
	}
	// Persist for next run (non-fatal if it fails).
	if pemBlock, err := xssh.MarshalPrivateKey(key, "emoji-roguelike server"); err == nil {
		_ = os.WriteFile(path, pem.EncodeToMemory(pemBlock), 0600)
	}
	return signer
}
