package mud

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// RunLog records statistics for one MUD life (connect → death or victory).
type RunLog struct {
	Timestamp        time.Time      `json:"timestamp"`
	Victory          bool           `json:"victory"`
	Class            string         `json:"class"`
	FloorsReached    int            `json:"floors_reached"`
	TurnsPlayed      int            `json:"turns_played"`
	EnemiesKilled    map[string]int `json:"enemies_killed"`
	ItemsUsed        map[string]int `json:"items_used"`
	InscriptionsRead int            `json:"inscriptions_read"`
	DamageDealt      int            `json:"damage_dealt"`
	DamageTaken      int            `json:"damage_taken"`
	CauseOfDeath     string         `json:"cause_of_death"`
}

// saveRunLog appends the completed run as a single JSON line to runs.jsonl.
// Errors are silently discarded so a disk problem never crashes the server.
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
	f.Write(data)      //nolint:errcheck
	f.Write([]byte("\n")) //nolint:errcheck
}

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
