package ssh

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	gossh "github.com/gliderlabs/ssh"
)

// SessionTty implements tcell.Tty backed by a gliderlabs/ssh session.
// Each connected SSH client gets its own SessionTty → tcell.Screen pair.
type SessionTty struct {
	session gossh.Session
	mu      sync.Mutex
	window  gossh.Window
	winCh   <-chan gossh.Window
	cb      func() // resize callback registered by tcell
}

// NewSessionTty wraps a gliderlabs SSH session as a tcell Tty.
// pty holds the initial window size; winCh delivers subsequent resize events.
func NewSessionTty(s gossh.Session, pty gossh.Pty, winCh <-chan gossh.Window) *SessionTty {
	return &SessionTty{
		session: s,
		window:  pty.Window,
		winCh:   winCh,
	}
}

// Read reads raw bytes from the SSH session's stdin (keyboard input).
func (t *SessionTty) Read(b []byte) (int, error) { return t.session.Read(b) }

// Write writes rendered output to the SSH session's stdout.
func (t *SessionTty) Write(b []byte) (int, error) { return t.session.Write(b) }

// Close closes the SSH session channel.
func (t *SessionTty) Close() error { return t.session.Close() }

// Start is a no-op — the SSH channel is already open.
func (t *SessionTty) Start() error { return nil }

// Stop is a no-op — the SSH channel is managed by the server handler goroutine.
func (t *SessionTty) Stop() error { return nil }

// Drain is a no-op — SSH flushes writes immediately.
func (t *SessionTty) Drain() error { return nil }

// WindowSize returns the current terminal dimensions.
func (t *SessionTty) WindowSize() (tcell.WindowSize, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return tcell.WindowSize{Width: t.window.Width, Height: t.window.Height}, nil
}

// NotifyResize registers a callback invoked on every window resize event.
// It also starts a goroutine that drains the window-change channel for the
// lifetime of the session.
func (t *SessionTty) NotifyResize(cb func()) {
	t.mu.Lock()
	t.cb = cb
	t.mu.Unlock()

	go func() {
		for win := range t.winCh {
			t.mu.Lock()
			t.window = win
			localCb := t.cb
			t.mu.Unlock()
			if localCb != nil {
				localCb()
			}
		}
	}()
}
