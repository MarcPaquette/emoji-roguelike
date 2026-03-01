package telnet

import (
	"io"
	"sync"
)

// FSM states for the IAC filter.
const (
	stateData   = iota
	stateIAC    // saw IAC, waiting for command byte
	stateOption // saw IAC + WILL/WONT/DO/DONT, waiting for option byte
	stateSB     // inside subnegotiation, consuming data bytes
	stateSBIAC  // inside subnegotiation, saw IAC (expecting SE or IAC escape)
)

// Filter wraps an io.Reader and strips telnet protocol bytes (IAC commands
// and subnegotiations), passing only application data through to Read callers.
// It also captures NAWS subnegotiation to track window size.
type Filter struct {
	r     io.Reader
	state int

	mu     sync.Mutex
	width  int
	height int
	nawsCb func()

	// subnegotiation accumulator
	sbOpt  byte
	sbData []byte
}

// NewFilter creates a Filter that strips IAC commands from r.
func NewFilter(r io.Reader) *Filter {
	return &Filter{
		r:      r,
		width:  80,
		height: 24,
	}
}

// WindowSize returns the current terminal dimensions.
// Defaults to 80x24 if no NAWS has been received.
func (f *Filter) WindowSize() (width, height int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.width, f.height
}

// OnNAWS registers a callback invoked when a NAWS update arrives.
func (f *Filter) OnNAWS(cb func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nawsCb = cb
}

// Read reads from the underlying reader, strips IAC protocol bytes,
// and returns only application data.
func (f *Filter) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	// We may need multiple underlying reads if all bytes are protocol data.
	for {
		buf := make([]byte, len(p))
		n, err := f.r.Read(buf)
		if n > 0 {
			out := f.process(buf[:n], p)
			if out > 0 {
				return out, err
			}
			// All bytes were protocol data; if no error, read again.
			if err != nil {
				return 0, err
			}
			continue
		}
		return 0, err
	}
}

// process runs raw bytes through the FSM and writes app data into dst.
// Returns the number of app bytes written.
func (f *Filter) process(raw []byte, dst []byte) int {
	out := 0
	for _, b := range raw {
		switch f.state {
		case stateData:
			if b == IAC {
				f.state = stateIAC
			} else {
				if out < len(dst) {
					dst[out] = b
					out++
				}
			}

		case stateIAC:
			switch b {
			case IAC:
				// Escaped 0xFF — emit literal 0xFF
				if out < len(dst) {
					dst[out] = 0xFF
					out++
				}
				f.state = stateData
			case WILL, WONT, DO, DONT:
				f.state = stateOption
			case SB:
				f.sbOpt = 0
				f.sbData = f.sbData[:0]
				f.state = stateSB
			default:
				// Unknown command byte — skip and return to data
				f.state = stateData
			}

		case stateOption:
			// Option byte consumed; back to data
			f.state = stateData

		case stateSB:
			if b == IAC {
				f.state = stateSBIAC
			} else {
				if f.sbOpt == 0 {
					f.sbOpt = b
				} else {
					f.sbData = append(f.sbData, b)
				}
			}

		case stateSBIAC:
			if b == SE {
				f.handleSubneg()
				f.state = stateData
			} else if b == IAC {
				// Escaped IAC inside subnegotiation
				f.sbData = append(f.sbData, 0xFF)
				f.state = stateSB
			} else {
				// Unexpected — treat as end of subneg
				f.state = stateData
			}
		}
	}
	return out
}

// handleSubneg processes a completed subnegotiation.
func (f *Filter) handleSubneg() {
	if f.sbOpt == OptNAWS && len(f.sbData) == 4 {
		w := int(f.sbData[0])<<8 | int(f.sbData[1])
		h := int(f.sbData[2])<<8 | int(f.sbData[3])
		f.mu.Lock()
		f.width = w
		f.height = h
		cb := f.nawsCb
		f.mu.Unlock()
		if cb != nil {
			cb()
		}
	}
}
