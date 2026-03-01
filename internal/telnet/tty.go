package telnet

import (
	"net"

	"github.com/gdamore/tcell/v2"
)

// TelnetTty implements tcell.Tty over a raw TCP connection, using Filter to
// strip telnet protocol bytes from the input stream.
type TelnetTty struct {
	conn   net.Conn
	filter *Filter
}

// NewTelnetTty wraps a TCP connection as a tcell.Tty.
func NewTelnetTty(conn net.Conn) *TelnetTty {
	return &TelnetTty{
		conn:   conn,
		filter: NewFilter(conn),
	}
}

// Start sends the initial telnet negotiation to the client.
func (t *TelnetTty) Start() error {
	return Negotiate(t.conn)
}

// Stop is a no-op.
func (t *TelnetTty) Stop() error { return nil }

// Drain is a no-op.
func (t *TelnetTty) Drain() error { return nil }

// Read reads application data from the connection, stripping IAC protocol bytes.
func (t *TelnetTty) Read(p []byte) (int, error) {
	return t.filter.Read(p)
}

// Write sends output to the client connection.
func (t *TelnetTty) Write(p []byte) (int, error) {
	return t.conn.Write(p)
}

// Close closes the underlying TCP connection.
func (t *TelnetTty) Close() error {
	return t.conn.Close()
}

// WindowSize returns the current terminal dimensions from NAWS, or 80x24 default.
func (t *TelnetTty) WindowSize() (tcell.WindowSize, error) {
	w, h := t.filter.WindowSize()
	return tcell.WindowSize{Width: w, Height: h}, nil
}

// NotifyResize registers a callback invoked when a NAWS window size update arrives.
func (t *TelnetTty) NotifyResize(cb func()) {
	t.filter.OnNAWS(cb)
}
