package wal

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"

	"github.com/joaodddev/high-perf-ledger/internal/codec"
	"github.com/joaodddev/high-perf-ledger/internal/ledger"
)

type Writer struct {
	mu     sync.Mutex
	file   *os.File
	codec  codec.Codec
	fsync  bool
	nextID uint64
}

type Config struct {
	Path  string
	Codec codec.Codec

	Fsync bool
}

func NewWriter(cfg Config) (*Writer, error) {
	f, err := os.OpenFile(cfg.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("wal: open file: %w", err)
	}

	return &Writer{
		file:  f,
		codec: cfg.Codec,
		fsync: cfg.Fsync,
	}, nil
}

func (w *Writer) Append(e ledger.Event) (uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.nextID++
	e.ID = w.nextID

	payload, err := w.codec.Encode(e)
	if err != nil {
		return 0, fmt.Errorf("wal: encode event: %w", err)
	}

	frame := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(frame[:4], uint32(len(payload)))
	copy(frame[4:], payload)

	if _, err := w.file.Write(frame); err != nil {
		return 0, fmt.Errorf("wal: write frame: %w", err)
	}

	if w.fsync {
		if err := w.file.Sync(); err != nil {
			return 0, fmt.Errorf("wal: fsync: %w", err)
		}
	}

	return e.ID, nil
}

func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}
