package ledger

import (
	"fmt"
	"path/filepath"

	"github.com/joaodddev/high-perf-ledger/internal/codec"
	"github.com/joaodddev/high-perf-ledger/internal/event"
	"github.com/joaodddev/high-perf-ledger/internal/snapshot"
	"github.com/joaodddev/high-perf-ledger/internal/wal"
)

type Store struct {
	writer      *wal.Writer
	reader      func() (*wal.Reader, error)
	snapshotDir string
}

type Config struct {
	WALPath     string
	SnapshotDir string
	Codec       codec.Codec
}

func NewStore(cfg Config) (*Store, error) {
	w, err := wal.NewWriter(wal.Config{Path: cfg.WALPath, Codec: cfg.Codec})
	if err != nil {
		return nil, fmt.Errorf("ledger: create writer: %w", err)
	}

	return &Store{
		writer:      w,
		reader:      func() (*wal.Reader, error) { return wal.NewReader(cfg.WALPath, cfg.Codec) },
		snapshotDir: cfg.SnapshotDir,
	}, nil
}

func (s *Store) Append(e event.Event) (uint64, error) {
	return s.writer.Append(e)
}

func (s *Store) snapshotPath(accountID string) string {
	return filepath.Join(s.snapshotDir, accountID+".snap")
}

func (s *Store) GetAccount(accountID string) (*Account, error) {
	r, err := s.reader()
	if err != nil {
		return nil, fmt.Errorf("ledger: open reader: %w", err)
	}
	defer r.Close()

	snap, found, err := snapshot.Read(s.snapshotPath(accountID))
	if err != nil {
		return nil, fmt.Errorf("ledger: read snapshot: %w", err)
	}

	if !found {
		allEvents, err := r.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("ledger: read all events: %w", err)
		}
		return Replay(accountID, filterByAccount(allEvents, accountID))
	}

	newEvents, err := r.ReadAfter(snap.LastEventID)
	if err != nil {
		return nil, fmt.Errorf("ledger: read events after snapshot: %w", err)
	}

	acc := &Account{
		ID:      snap.AccountID,
		Balance: snap.Balance,
		Opened:  snap.Opened,
		version: snap.LastEventID,
	}

	for _, e := range filterByAccount(newEvents, accountID) {
		if err := acc.Apply(e); err != nil {
			return nil, err
		}
	}

	return acc, nil
}

func (s *Store) Snapshot(accountID string) error {
	acc, err := s.GetAccount(accountID)
	if err != nil {
		return fmt.Errorf("ledger: replay before snapshot: %w", err)
	}

	snap := snapshot.Snapshot{
		AccountID:   acc.ID,
		Balance:     acc.Balance,
		Opened:      acc.Opened,
		LastEventID: acc.version,
	}

	return snapshot.Write(s.snapshotPath(accountID), snap)
}

func filterByAccount(events []event.Event, accountID string) []event.Event {
	var out []event.Event
	for _, e := range events {
		if e.AccountID == accountID {
			out = append(out, e)
		}
	}
	return out
}

func (s *Store) Close() error {
	return s.writer.Close()
}
