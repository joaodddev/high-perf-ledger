package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
)

// Snapshot captures an account's fully-replayed state at a specific
// point in the event log, identified by LastEventID. Replaying only
// needs to apply events with ID > LastEventID after loading this.
type Snapshot struct {
	AccountID   string `json:"account_id"`
	Balance     int64  `json:"balance"`
	Opened      bool   `json:"opened"`
	LastEventID uint64 `json:"last_event_id"`
}

// Write persists a snapshot to disk, overwriting any existing file
// at path. Snapshots are small and infrequent, so plain JSON (not the
// pluggable WAL codec) is intentionally used here for simplicity —
// this is a distinct on-disk format from the event log itself.
func Write(path string, s Snapshot) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("snapshot: marshal: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("snapshot: write temp file: %w", err)
	}

	// Rename is atomic on POSIX filesystems: readers never observe a
	// partially-written snapshot file, even if the process crashes
	// mid-write (they'd see the old file or the new one, never a mix).
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("snapshot: atomic rename: %w", err)
	}

	return nil
}

// Read loads a snapshot from disk. Returns (Snapshot{}, false, nil)
// if no snapshot file exists yet — this is a normal, expected state
// for an account that hasn't been snapshotted.
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
