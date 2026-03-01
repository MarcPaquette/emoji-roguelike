package telnet

import "io"

// Telnet protocol constants (RFC 854, RFC 1073).
const (
	IAC  byte = 0xFF // Interpret As Command
	WILL byte = 0xFB
	WONT byte = 0xFC
	DO   byte = 0xFD
	DONT byte = 0xFE
	SB   byte = 0xFA // Subnegotiation Begin
	SE   byte = 0xF0 // Subnegotiation End

	OptEcho byte = 0x01 // Echo
	OptSGA  byte = 0x03 // Suppress Go-Ahead
	OptNAWS byte = 0x1F // Negotiate About Window Size
)

// Negotiate writes the initial telnet option negotiation to w.
// Sequence: WILL ECHO, WILL SGA, DO NAWS (9 bytes total).
func Negotiate(w io.Writer) error {
	_, err := w.Write([]byte{
		IAC, WILL, OptEcho,
		IAC, WILL, OptSGA,
		IAC, DO, OptNAWS,
	})
	return err
}
