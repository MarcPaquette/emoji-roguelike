package telnet

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"

	"emoji-roguelike/internal/mud"

	"github.com/gdamore/tcell/v2"
)

// HandleConnection manages a single telnet client connection through class
// selection, session creation, and the game loop. It mirrors the SSH
// handleSession flow in cmd/server/main.go.
func HandleConnection(conn net.Conn, srv *mud.Server, logger *slog.Logger, termMu *sync.Mutex) {
	remoteAddr := conn.RemoteAddr().String()
	logger.Info("telnet connection", "remote", remoteAddr)

	tty := NewTelnetTty(conn)

	termMu.Lock()
	_ = os.Setenv("TERM", "xterm-256color")
	screen, err := tcell.NewTerminfoScreenFromTty(tty)
	termMu.Unlock()
	if err != nil {
		logger.Error("telnet terminal setup failed", "remote", remoteAddr, "error", err)
		fmt.Fprintf(conn, "Terminal setup failed: %v\r\n", err)
		conn.Close()
		return
	}
	if err := screen.Init(); err != nil {
		logger.Error("telnet screen init failed", "remote", remoteAddr, "error", err)
		fmt.Fprintf(conn, "Screen init failed: %v\r\n", err)
		conn.Close()
		return
	}
	defer screen.Fini()
	defer conn.Close()

	// Derive a display name from the remote address.
	name := sanitizeTelnetName(remoteAddr)

	cls, ok := mud.ClassSelect(screen)
	if !ok {
		logger.Info("telnet disconnected during class select", "remote", remoteAddr)
		return
	}

	sessID, color := srv.NextSessionID()
	sess := mud.NewSession(sessID, name, color, screen)
	sess.Class = cls
	sess.FovRadius = cls.FOVRadius
	sess.BaseMaxHP = cls.MaxHP
	sess.RunLog.Class = cls.Name

	if !srv.AddSession(sess) {
		fmt.Fprintf(conn, "Server is full. Please try again later.\r\n")
		return
	}
	defer srv.RemoveSession(sess)

	srv.RunLoop(sess)
}

// sanitizeTelnetName produces a short display name from a remote address.
func sanitizeTelnetName(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	if host == "::1" || host == "127.0.0.1" || host == "localhost" {
		return "Telnet"
	}
	const maxLen = 16
	if len(host) > maxLen {
		host = host[:maxLen]
	}
	return host
}
