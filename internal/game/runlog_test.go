package game

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunLogDirXDGEnvOverride(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)

	dir, err := runLogDir()
	if err != nil {
		t.Fatalf("runLogDir returned error: %v", err)
	}
	want := filepath.Join(tmp, "emoji-roguelike")
	if dir != want {
		t.Errorf("dir = %q; want %q", dir, want)
	}
}

func TestRunLogDirDefaultFallback(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "") // force the fallback path

	dir, err := runLogDir()
	if err != nil {
		t.Skip("skipping: no user home directory available in test environment")
	}
	suffix := filepath.Join(".local", "share", "emoji-roguelike")
	if !strings.HasSuffix(dir, suffix) {
		t.Errorf("dir %q does not end with %q", dir, suffix)
	}
}

func TestSaveRunLog(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)

	log := RunLog{
		Victory:       false,
		Class:         "arcanist",
		FloorsReached: 3,
		TurnsPlayed:   42,
		CauseOfDeath:  "🦀",
		EnemiesKilled: map[string]int{"🦀": 2},
		ItemsUsed:     map[string]int{"🧪": 1},
	}
	saveRunLog(log)

	logPath := filepath.Join(tmp, "emoji-roguelike", "runs.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("runs.jsonl not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "arcanist") {
		t.Errorf("log file does not contain class name; got: %q", content)
	}
	if !strings.HasSuffix(content, "\n") {
		t.Errorf("log entry should end with newline; got: %q", content)
	}
}

func TestSaveRunLogAppendsMultiple(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)

	for i := range 3 {
		saveRunLog(RunLog{
			Class:         "arcanist",
			FloorsReached: i + 1,
			EnemiesKilled: map[string]int{},
			ItemsUsed:     map[string]int{},
		})
	}

	logPath := filepath.Join(tmp, "emoji-roguelike", "runs.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("runs.jsonl not found: %v", err)
	}
	// Each call appends one JSON line; count the newlines.
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 log lines, got %d", len(lines))
	}
}
