package codec

import (
	"testing"
	"time"

	"github.com/joaodddev/high-perf-ledger/internal/event"
)

func TestMsgpackCodec_RoundTrip(t *testing.T) {
	c := NewMsgpackCodec()
	original := event.Event{
		ID:        42,
		AccountID: "acc-1",
		Type:      event.Deposited,
		Amount:    5_000,
		Timestamp: time.Now().Truncate(time.Second), // avoid sub-second precision mismatches
		Metadata:  map[string]string{"note": "test"},
	}

	data, err := c.Encode(original)
	if err != nil {
		t.Fatalf("unexpected error encoding: %v", err)
	}

	decoded, err := c.Decode(data)
	if err != nil {
		t.Fatalf("unexpected error decoding: %v", err)
	}

	if decoded.ID != original.ID || decoded.AccountID != original.AccountID ||
		decoded.Type != original.Type || decoded.Amount != original.Amount {
		t.Fatalf("round-trip mismatch: got %+v, want %+v", decoded, original)
	}
	if !decoded.Timestamp.Equal(original.Timestamp) {
		t.Fatalf("timestamp mismatch: got %v, want %v", decoded.Timestamp, original.Timestamp)
	}
}

func TestMsgpackCodec_Name(t *testing.T) {
	c := NewMsgpackCodec()
	if c.Name() != "msgpack" {
		t.Fatalf("expected name 'msgpack', got %q", c.Name())
	}
}
