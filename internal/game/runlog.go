package game

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// saveRunLog appends the completed run as a single JSON line to runs.jsonl.
// Errors are silently discarded so a disk problem never crashes the game.
func saveRunLog(log RunLog) {
	dir, err := runLogDir()
	if err != nil {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(dir, "runs.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	data, err := json.Marshal(log)
	if err != nil {
		return
	}
	f.Write(data)    //nolint:errcheck — best-effort write
	f.Write([]byte("\n")) //nolint:errcheck
}

// runLogDir returns the directory where run logs are stored.
// Follows XDG Base Directory spec: $XDG_DATA_HOME/emoji-roguelike,
// defaulting to ~/.local/share/emoji-roguelike.
func runLogDir() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "emoji-roguelike"), nil
}
