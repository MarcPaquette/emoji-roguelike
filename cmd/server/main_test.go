package main

import "testing"

func TestSanitizeName(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect string
	}{
		{"normal short name", "Alice", "Alice"},
		{"exactly 16 chars", "1234567890123456", "1234567890123456"},
		{"long name truncated", "ThisIsAVeryLongUsername", "ThisIsAVeryLongU"},
		{"control chars stripped", "he\x00ll\x1bo", "hello"},
		{"ansi escape partial", "he\x1b[31mllo", "he[31mllo"},
		{"empty input", "", ""},
		{"pure control chars", "\x00\x01\x02\x1b", ""},
		{"multi-byte runes truncated by byte limit", "日本語のテスト名前ですよね東京大阪京都", "日本語のテス"},
		{"emoji truncated by byte limit", "🎮Player🎮Name🎮", "🎮Player🎮Na"},
		{"mixed printable and control", "a\x00b\x01c\x02d", "abcd"},
		{"tabs stripped", "hello\tworld", "helloworld"},
		{"newlines stripped", "hello\nworld", "helloworld"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeName(tc.input)
			if got != tc.expect {
				t.Errorf("sanitizeName(%q) = %q, want %q", tc.input, got, tc.expect)
			}
		})
	}
}

func TestAllowedTerms(t *testing.T) {
	cases := []struct {
		name    string
		term    string
		allowed bool
	}{
		{"xterm-256color", "xterm-256color", true},
		{"tmux", "tmux", true},
		{"linux", "linux", true},
		{"vt100", "vt100", true},
		{"screen", "screen", true},
		{"rxvt-unicode-256color", "rxvt-unicode-256color", true},
		{"unknown term", "evil-term", false},
		{"path traversal", "../../../etc/passwd", false},
		{"empty string", "", false},
		{"xterm-kitty", "xterm-kitty", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := allowedTerms[tc.term]
			if got != tc.allowed {
				t.Errorf("allowedTerms[%q] = %v, want %v", tc.term, got, tc.allowed)
			}
		})
	}
}
