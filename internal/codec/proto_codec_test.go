package codec

import (
	"testing"
	"time"

	"github.com/joaodddev/high-perf-ledger/internal/event"
)

func TestProtoCodec_RoundTrip(t *testing.T) {
	c := NewProtoCodec()
	original := event.Event{
		ID:        7,
		AccountID: "acc-2",
		Type:      event.Withdrawn,
		Amount:    3_000,
		Timestamp: time.Now().Truncate(time.Nanosecond),
		Metadata:  map[string]string{"reason": "atm"},
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

func TestProtoCodec_Name(t *testing.T) {
	c := NewProtoCodec()
	if c.Name() != "protobuf" {
		t.Fatalf("expected name 'protobuf', got %q", c.Name())
	}
}
