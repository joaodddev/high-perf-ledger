package wal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joaodddev/high-perf-ledger/internal/codec"
	"github.com/joaodddev/high-perf-ledger/internal/ledger"
)

func tempWALPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "test.wal")
}

func TestWriter_AppendAssignsSequentialIDs(t *testing.T) {
	path := tempWALPath(t)
	w, err := NewWriter(Config{Path: path, Codec: codec.NewJSONCodec()})
	if err != nil {
		t.Fatalf("unexpected error creating writer: %v", err)
	}
	defer w.Close()

	id1, err := w.Append(ledger.Event{AccountID: "acc-1", Type: ledger.EventAccountOpened})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	id2, err := w.Append(ledger.Event{AccountID: "acc-1", Type: ledger.EventDeposited, Amount: 1000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id1 != 1 || id2 != 2 {
		t.Fatalf("expected sequential IDs 1,2 got %d,%d", id1, id2)
	}
}

func TestWriter_Reader_RoundTrip(t *testing.T) {
	path := tempWALPath(t)
	c := codec.NewJSONCodec()

	w, err := NewWriter(Config{Path: path, Codec: c})
	if err != nil {
		t.Fatalf("unexpected error creating writer: %v", err)
	}

	events := []ledger.Event{
		{AccountID: "acc-1", Type: ledger.EventAccountOpened},
		{AccountID: "acc-1", Type: ledger.EventDeposited, Amount: 5000},
		{AccountID: "acc-1", Type: ledger.EventWithdrawn, Amount: 2000},
	}

	for _, e := range events {
		if _, err := w.Append(e); err != nil {
			t.Fatalf("unexpected error appending: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("unexpected error closing writer: %v", err)
	}

	r, err := NewReader(path, c)
	if err != nil {
		t.Fatalf("unexpected error creating reader: %v", err)
	}
	defer r.Close()

	got, err := r.ReadAll()
	if err != nil {
		t.Fatalf("unexpected error reading: %v", err)
	}

	if len(got) != len(events) {
		t.Fatalf("expected %d events, got %d", len(events), len(got))
	}
	for i, e := range got {
		if e.Type != events[i].Type || e.Amount != events[i].Amount {
			t.Fatalf("event %d mismatch: got %+v, want type=%s amount=%d", i, e, events[i].Type, events[i].Amount)
		}
		if e.ID != uint64(i+1) {
			t.Fatalf("event %d expected ID %d, got %d", i, i+1, e.ID)
		}
	}
}

func TestWriter_FsyncDoesNotError(t *testing.T) {
	path := tempWALPath(t)
	w, err := NewWriter(Config{Path: path, Codec: codec.NewJSONCodec(), Fsync: true})
	if err != nil {
		t.Fatalf("unexpected error creating writer: %v", err)
	}
	defer w.Close()

	if _, err := w.Append(ledger.Event{AccountID: "acc-1", Type: ledger.EventAccountOpened}); err != nil {
		t.Fatalf("unexpected error with fsync enabled: %v", err)
	}
}

func TestReader_EmptyFile(t *testing.T) {
	path := tempWALPath(t)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("unexpected error creating empty file: %v", err)
	}
	f.Close()

	r, err := NewReader(path, codec.NewJSONCodec())
	if err != nil {
		t.Fatalf("unexpected error creating reader: %v", err)
	}
	defer r.Close()

	events, err := r.ReadAll()
	if err != nil {
		t.Fatalf("unexpected error reading empty file: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events from empty file, got %d", len(events))
	}
}
