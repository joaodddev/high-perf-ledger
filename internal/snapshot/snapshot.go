package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
)

type Snapshot struct {
	AccountID   string `json:"account_id"`
	Balance     int64  `json:"balance"`
	Opened      bool   `json:"opened"`
	LastEventID uint64 `json:"last_event_id"`
}

func Write(path string, s Snapshot) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("snapshot: marshal: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("snapshot: write temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("snapshot: atomic rename: %w", err)
	}

	return nil
}

func Read(path string) (Snapshot, bool, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Snapshot{}, false, nil
	}
	if err != nil {
		return Snapshot{}, false, fmt.Errorf("snapshot: read file: %w", err)
	}

	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return Snapshot{}, false, fmt.Errorf("snapshot: unmarshal: %w", err)
	}

	return s, true, nil
}
