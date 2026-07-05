package ledger

import (
	"fmt"

	"github.com/joaodddev/high-perf-ledger/internal/codec"
	"github.com/joaodddev/high-perf-ledger/internal/event"
	"github.com/joaodddev/high-perf-ledger/internal/wal"
)

type Store struct {
	writer *wal.Writer
	reader func() (*wal.Reader, error)
}

type Config struct {
	WALPath string
	Codec   codec.Codec
}

func NewStore(cfg Config) (*Store, error) {
	w, err := wal.NewWriter(wal.Config{Path: cfg.WALPath, Codec: cfg.Codec})
	if err != nil {
		return nil, fmt.Errorf("ledger: create writer: %w", err)
	}
	return &Store{
		writer: w,
		reader: func() (*wal.Reader, error) { return wal.NewReader(cfg.WALPath, cfg.Codec) },
	}, nil
}

func (s *Store) Append(e event.Event) (uint64, error) {
	return s.writer.Append(e)
}

// TODO(perf): this scans and decodes the entire log file on every
// call, discarding events for other accounts. Fine for small logs,
// but becomes the main bottleneck as the WAL grows. Addressed in
// Phase 4 via periodic snapshotting + incremental replay.
func (s *Store) GetAccount(accountID string) (*Account, error) {
	r, err := s.reader()
	if err != nil {
		return nil, fmt.Errorf("ledger: open reader: %w", err)
	}
	defer r.Close()

	allEvents, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("ledger: read all events: %w", err)
	}

	filtered := filterByAccount(allEvents, accountID)
	return Replay(accountID, filtered)
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
