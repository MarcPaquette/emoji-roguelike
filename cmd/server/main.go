// emoji-roguelike-server starts an SSH MUD server supporting N simultaneous
// players in a shared dungeon. Build:
//
//	go build -o emoji-roguelike-server ./cmd/server
//
// Usage:
//
//	./emoji-roguelike-server [--port 2222] [--key server_host_key]
//
// Connect from any terminal:
//
//	ssh -p 2222 localhost
package main

import (
	cryptorand "crypto/rand"
	"crypto/ed25519"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	mathrand "math/rand"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	"emoji-roguelike/internal/mud"
	internalssh "emoji-roguelike/internal/ssh"

	"github.com/gdamore/tcell/v2"
	gossh "github.com/gliderlabs/ssh"
	xssh "golang.org/x/crypto/ssh"
)

// allowedTerms is the set of TERM values we accept from SSH clients.
// Anything not in this set is replaced with "xterm-256color".
var allowedTerms = map[string]bool{
	"xterm-256color":  true,
	"xterm":           true,
	"xterm-color":     true,
	"screen-256color": true,
	"screen":          true,
	"tmux-256color":   true,
	"tmux":            true,
	"linux":           true,
	"vt100":           true,
	"rxvt-unicode-256color": true,
}

const maxUsernameLen = 16

// sanitizeName cleans a username for display: strips non-printable runes and
// truncates to maxUsernameLen.
func sanitizeName(name string) string {
	var b strings.Builder
	for _, r := range name {
		if unicode.IsPrint(r) && !unicode.IsControl(r) {
			b.WriteRune(r)
			if b.Len() >= maxUsernameLen {
				break
			}
		}
	}
	s := b.String()
	// Truncate to maxUsernameLen runes (the byte check above is approximate
	// for multi-byte runes, so do a rune-level trim).
	runes := []rune(s)
	if len(runes) > maxUsernameLen {
		runes = runes[:maxUsernameLen]
	}
	return string(runes)
}

func main() {
	port := flag.Int("port", 2222, "SSH server port")
	keyFile := flag.String("key", "server_host_key", "Path to the PEM-encoded host key (auto-generated if absent)")
	flag.Parse()

	signer := loadOrCreateHostKey(*keyFile)
	rng := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	srv := mud.NewServer(rng)

	// Start the world ticker in a background goroutine.
	go srv.Run()

	sshSrv := &gossh.Server{
		Addr:        fmt.Sprintf(":%d", *port),
		IdleTimeout: 10 * time.Minute,
		MaxTimeout:  4 * time.Hour,
		Handler: func(s gossh.Session) {
			handleSession(srv, s)
		},
		PtyCallback: func(_ gossh.Context, _ gossh.Pty) bool { return true },
		HostSigners: []gossh.Signer{signer},
	}

	log.Printf("emoji-roguelike MUD server listening on :%d", *port)
	log.Printf("Connect with:  ssh -p %d -o StrictHostKeyChecking=no localhost", *port)
	log.Fatal(sshSrv.ListenAndServe())
}

// termMu serializes os.Setenv("TERM") around tcell screen creation.
// Multiple goroutines may create screens concurrently.
var termMu sync.Mutex

// handleSession is the gliderlabs SSH handler for one connection.
func handleSession(srv *mud.Server, s gossh.Session) {
	pty, winCh, hasPTY := s.Pty()
	if !hasPTY {
		fmt.Fprintln(s, "This game requires a PTY. Connect with: ssh -t -p 2222 <host>")
		return
	}

	term := "xterm-256color"
	for _, env := range s.Environ() {
		if strings.HasPrefix(env, "TERM=") {
			candidate := env[5:]
			if allowedTerms[candidate] {
				term = candidate
			}
			break
		}
	}

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
	defer screen.Fini()

	// Use SSH username as display name, fallback to remote address.
	name := sanitizeName(s.User())
	if name == "" || name == "git" {
		name = sanitizeName(s.RemoteAddr().String())
	}
	if name == "" {
		name = "Player"
	}

	// Class selection (blocking, before joining the world).
	cls, ok := mud.ClassSelect(screen)
	if !ok {
		return
	}

	sessID, color := srv.NextSessionID()
	sess := mud.NewSession(sessID, name, color, screen)
	sess.Class = cls
	sess.FovRadius = cls.FOVRadius
	sess.BaseMaxHP = cls.MaxHP
	sess.RunLog.Class = cls.Name

	if !srv.AddSession(sess) {
		fmt.Fprintln(s, "Server is full. Please try again later.")
		return
	}
	defer srv.RemoveSession(sess)

	srv.RunLoop(sess)
}

// ─── host key ────────────────────────────────────────────────────────────────

func loadOrCreateHostKey(path string) gossh.Signer {
	if data, err := os.ReadFile(path); err == nil {
		if signer, err := xssh.ParsePrivateKey(data); err == nil {
			log.Printf("Loaded host key from %s", path)
			return signer
		}
	}

	log.Printf("Generating new ed25519 host key → %s", path)
	_, key, err := ed25519.GenerateKey(cryptorand.Reader)
	if err != nil {
		log.Fatalf("generate host key: %v", err)
	}
	signer, err := xssh.NewSignerFromKey(key)
	if err != nil {
		log.Fatalf("create signer: %v", err)
	}
	if pemBlock, err := xssh.MarshalPrivateKey(key, "emoji-roguelike server"); err == nil {
		_ = os.WriteFile(path, pem.EncodeToMemory(pemBlock), 0600)
	}
	return signer
}
