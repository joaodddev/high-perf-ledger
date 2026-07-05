package ledger

import (
	"fmt"

	"github.com/joaodddev/high-perf-ledger/internal/codec"
	"github.com/joaodddev/high-perf-ledger/internal/wal"
)

type Store struct {
	writer *wal.Writer
	reader func() (*wal.Reader, error) // factory, since Reader opens its own file handle
}

type Config struct {
	WALPath string
	Codec   codec.Codec
}

func NewStore(cfg Config) (*Store, error) {
	w, err := wal.NewWriter(wal.Config{
		Path:  cfg.WALPath,
		Codec: cfg.Codec,
	})
	if err != nil {
		return nil, fmt.Errorf("ledger: create writer: %w", err)
	}

	return &Store{
		writer: w,
		reader: func() (*wal.Reader, error) {
			return wal.NewReader(cfg.WALPath, cfg.Codec)
		},
	}, nil
}

func (s *Store) Append(e Event) (uint64, error) {
	return s.writer.Append(e)
}

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

func filterByAccount(events []Event, accountID string) []Event {
	var out []Event
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
