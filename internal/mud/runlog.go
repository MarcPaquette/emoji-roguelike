package mud

import (
	"encoding/json"
	"log/slog"
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
	GoldEarned       int            `json:"gold_earned"`
	CauseOfDeath     string         `json:"cause_of_death"`
}

// saveRunLog appends the completed run as a single JSON line to runs.jsonl.
// Errors are logged but never crash the server.
func saveRunLog(rl RunLog, logger *slog.Logger) {
	dir, err := runLogDir()
	if err != nil {
		logger.Warn("run log: cannot determine data dir", "error", err)
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		logger.Warn("run log: cannot create data dir", "error", err)
		return
	}
	f, err := os.OpenFile(filepath.Join(dir, "runs.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		logger.Warn("run log: cannot open file", "error", err)
		return
	}
	defer f.Close()
	data, err := json.Marshal(rl)
	if err != nil {
		logger.Warn("run log: cannot marshal JSON", "error", err)
		return
	}
	f.Write(data)         //nolint:errcheck
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
