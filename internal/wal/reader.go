package wal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/joaodddev/high-perf-ledger/internal/codec"
	"github.com/joaodddev/high-perf-ledger/internal/ledger"
)

type Reader struct {
	file  *os.File
	codec codec.Codec
}

func NewReader(path string, c codec.Codec) (*Reader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("wal: open file for read: %w", err)
	}
	return &Reader{file: f, codec: c}, nil
}

func (r *Reader) ReadAll() ([]ledger.Event, error) {
	var events []ledger.Event

	for {
		lenBuf := make([]byte, 4)
		_, err := io.ReadFull(r.file, lenBuf)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("wal: read frame length: %w", err)
		}

		length := binary.BigEndian.Uint32(lenBuf)
		payload := make([]byte, length)
		if _, err := io.ReadFull(r.file, payload); err != nil {
			return nil, fmt.Errorf("wal: read frame payload: %w", err)
		}

		event, err := r.codec.Decode(payload)
		if err != nil {
			return nil, fmt.Errorf("wal: decode event: %w", err)
		}

		events = append(events, event)
	}

	return events, nil
}

func (r *Reader) Close() error {
	return r.file.Close()
}
