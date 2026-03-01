package telnet

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
)

// ── proto.go ────────────────────────────────────────────────────────────────

func TestNegotiateWritesExpectedBytes(t *testing.T) {
	var buf bytes.Buffer
	if err := Negotiate(&buf); err != nil {
		t.Fatalf("Negotiate: %v", err)
	}
	got := buf.Bytes()
	want := []byte{IAC, WILL, OptEcho, IAC, WILL, OptSGA, IAC, DO, OptNAWS}
	if !bytes.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNegotiateOrder(t *testing.T) {
	var buf bytes.Buffer
	_ = Negotiate(&buf)
	b := buf.Bytes()
	if len(b) != 9 {
		t.Fatalf("expected 9 bytes, got %d", len(b))
	}
	// ECHO before SGA before NAWS
	if b[2] != OptEcho {
		t.Errorf("first option should be Echo, got %#x", b[2])
	}
	if b[5] != OptSGA {
		t.Errorf("second option should be SGA, got %#x", b[5])
	}
	if b[8] != OptNAWS {
		t.Errorf("third option should be NAWS, got %#x", b[8])
	}
}

// ── filter.go ───────────────────────────────────────────────────────────────

func TestFilterPassthroughPlainData(t *testing.T) {
	r := bytes.NewReader([]byte("hello"))
	f := NewFilter(r)
	buf := make([]byte, 32)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Read: %v", err)
	}
	if string(buf[:n]) != "hello" {
		t.Errorf("got %q, want %q", buf[:n], "hello")
	}
}

func TestFilterStripsThreeByteIAC(t *testing.T) {
	data := []byte{IAC, WILL, OptEcho, 'A', 'B'}
	f := NewFilter(bytes.NewReader(data))
	buf := make([]byte, 32)
	n, _ := f.Read(buf)
	if string(buf[:n]) != "AB" {
		t.Errorf("got %q, want %q", buf[:n], "AB")
	}
}

func TestFilterStripsIACAtEnd(t *testing.T) {
	data := []byte{'X', IAC, DO, OptSGA}
	f := NewFilter(bytes.NewReader(data))
	buf := make([]byte, 32)
	n, _ := f.Read(buf)
	if string(buf[:n]) != "X" {
		t.Errorf("got %q, want %q", buf[:n], "X")
	}
}

func TestFilterHandlesNAWSSubnegotiation(t *testing.T) {
	data := []byte{IAC, SB, OptNAWS, 0, 80, 0, 24, IAC, SE}
	f := NewFilter(bytes.NewReader(data))
	buf := make([]byte, 32)
	n, _ := f.Read(buf)
	if n != 0 {
		t.Errorf("expected 0 app bytes, got %d: %q", n, buf[:n])
	}
	w, h := f.WindowSize()
	if w != 80 || h != 24 {
		t.Errorf("WindowSize = %dx%d, want 80x24", w, h)
	}
}

func TestFilterNAWSLargeSize(t *testing.T) {
	// 200x50 = (0x00C8, 0x0032)
	data := []byte{IAC, SB, OptNAWS, 0x00, 0xC8, 0x00, 0x32, IAC, SE}
	f := NewFilter(bytes.NewReader(data))
	buf := make([]byte, 32)
	f.Read(buf)
	w, h := f.WindowSize()
	if w != 200 || h != 50 {
		t.Errorf("WindowSize = %dx%d, want 200x50", w, h)
	}
}

func TestFilterSplitAcrossReads(t *testing.T) {
	// Use a pipe to control read boundaries.
	pr, pw := io.Pipe()
	f := NewFilter(pr)

	go func() {
		// Send IAC split across two writes.
		pw.Write([]byte{IAC})
		pw.Write([]byte{WILL, OptEcho, 'Z'})
		pw.Close()
	}()

	var result []byte
	buf := make([]byte, 32)
	for {
		n, err := f.Read(buf)
		result = append(result, buf[:n]...)
		if err != nil {
			break
		}
	}
	if string(result) != "Z" {
		t.Errorf("got %q, want %q", result, "Z")
	}
}

func TestFilterDoubleIACEscape(t *testing.T) {
	data := []byte{'A', IAC, IAC, 'B'}
	f := NewFilter(bytes.NewReader(data))
	buf := make([]byte, 32)
	n, _ := f.Read(buf)
	want := []byte{'A', 0xFF, 'B'}
	if !bytes.Equal(buf[:n], want) {
		t.Errorf("got %v, want %v", buf[:n], want)
	}
}

func TestFilterMultipleIACCommands(t *testing.T) {
	data := []byte{IAC, WILL, OptEcho, IAC, DO, OptSGA, 'd', 'a', 't', 'a'}
	f := NewFilter(bytes.NewReader(data))
	buf := make([]byte, 32)
	n, _ := f.Read(buf)
	if string(buf[:n]) != "data" {
		t.Errorf("got %q, want %q", buf[:n], "data")
	}
}

func TestFilterOnNAWSCallback(t *testing.T) {
	data := []byte{IAC, SB, OptNAWS, 0, 120, 0, 40, IAC, SE}
	f := NewFilter(bytes.NewReader(data))

	called := false
	f.OnNAWS(func() { called = true })

	buf := make([]byte, 32)
	f.Read(buf)

	if !called {
		t.Error("OnNAWS callback was not called")
	}
	w, h := f.WindowSize()
	if w != 120 || h != 40 {
		t.Errorf("WindowSize = %dx%d, want 120x40", w, h)
	}
}

func TestFilterDefaultWindowSize(t *testing.T) {
	f := NewFilter(bytes.NewReader(nil))
	w, h := f.WindowSize()
	if w != 80 || h != 24 {
		t.Errorf("default WindowSize = %dx%d, want 80x24", w, h)
	}
}

// ── tty.go ──────────────────────────────────────────────────────────────────

func TestTtyReadStripsIAC(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	tty := NewTelnetTty(server)
	go func() {
		client.Write([]byte{IAC, WILL, OptEcho})
		client.Write([]byte("hello"))
	}()

	buf := make([]byte, 32)
	n, err := tty.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if string(buf[:n]) != "hello" {
		t.Errorf("got %q, want %q", buf[:n], "hello")
	}
}

func TestTtyWrite(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	tty := NewTelnetTty(server)

	go func() {
		tty.Write([]byte("output"))
	}()

	buf := make([]byte, 32)
	n, err := client.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if string(buf[:n]) != "output" {
		t.Errorf("got %q, want %q", buf[:n], "output")
	}
}

func TestTtyWindowSizeDefault(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	tty := NewTelnetTty(server)
	ws, err := tty.WindowSize()
	if err != nil {
		t.Fatalf("WindowSize: %v", err)
	}
	if ws.Width != 80 || ws.Height != 24 {
		t.Errorf("got %dx%d, want 80x24", ws.Width, ws.Height)
	}
}

func TestTtyWindowSizeFromNAWS(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	tty := NewTelnetTty(server)

	go func() {
		// Send NAWS subnegotiation then app data.
		client.Write([]byte{IAC, SB, OptNAWS, 0, 132, 0, 43, IAC, SE, 'x'})
	}()

	buf := make([]byte, 32)
	tty.Read(buf) // triggers filter processing

	ws, err := tty.WindowSize()
	if err != nil {
		t.Fatalf("WindowSize: %v", err)
	}
	if ws.Width != 132 || ws.Height != 43 {
		t.Errorf("got %dx%d, want 132x43", ws.Width, ws.Height)
	}
}

func TestTtyNotifyResizeCallback(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	tty := NewTelnetTty(server)

	var mu sync.Mutex
	called := false
	tty.NotifyResize(func() {
		mu.Lock()
		called = true
		mu.Unlock()
	})

	go func() {
		client.Write([]byte{IAC, SB, OptNAWS, 0, 100, 0, 50, IAC, SE, 'y'})
	}()

	buf := make([]byte, 32)
	tty.Read(buf) // triggers NAWS processing

	mu.Lock()
	defer mu.Unlock()
	if !called {
		t.Error("resize callback was not called")
	}
}

func TestTtyStartSendsNegotiation(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	tty := NewTelnetTty(server)

	go func() {
		tty.Start()
	}()

	buf := make([]byte, 32)
	n, err := client.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	want := []byte{IAC, WILL, OptEcho, IAC, WILL, OptSGA, IAC, DO, OptNAWS}
	if !bytes.Equal(buf[:n], want) {
		t.Errorf("got %v, want %v", buf[:n], want)
	}
}

func TestTtyCloseClosesConn(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()

	tty := NewTelnetTty(server)
	tty.Close()

	// Reading from client should fail because server side is closed.
	buf := make([]byte, 1)
	_, err := client.Read(buf)
	if err == nil {
		t.Error("expected error reading from closed connection")
	}
}

func TestTtyStopDrainNoOps(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	tty := NewTelnetTty(server)
	if err := tty.Stop(); err != nil {
		t.Errorf("Stop: %v", err)
	}
	if err := tty.Drain(); err != nil {
		t.Errorf("Drain: %v", err)
	}
}

// ── handler.go ──────────────────────────────────────────────────────────────

func TestSanitizeTelnetNameLocalhost(t *testing.T) {
	cases := []struct {
		addr string
		want string
	}{
		{"127.0.0.1:12345", "Telnet"},
		{"[::1]:5555", "Telnet"},
		{"192.168.1.50:9999", "192.168.1.50"},
	}
	for _, tc := range cases {
		t.Run(tc.addr, func(t *testing.T) {
			got := sanitizeTelnetName(tc.addr)
			if got != tc.want {
				t.Errorf("sanitizeTelnetName(%q) = %q, want %q", tc.addr, got, tc.want)
			}
		})
	}
}

func TestSanitizeTelnetNameTruncatesLong(t *testing.T) {
	got := sanitizeTelnetName("abcdefghijklmnopqrstuvwxyz:1234")
	if len(got) > 16 {
		t.Errorf("name too long: %q (len %d)", got, len(got))
	}
}
