package bench

import (
	"testing"
	"time"

	"github.com/joaodddev/high-perf-ledger/internal/codec"
	"github.com/joaodddev/high-perf-ledger/internal/event"
)

func sampleEvent() event.Event {
	return event.Event{
		ID:        12345,
		AccountID: "acc-9f3d2b1a",
		Type:      event.TransferSent,
		Amount:    150_000,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"to":        "acc-7e2c9f01",
			"reference": "invoice-8823",
			"channel":   "mobile",
		},
	}
}

func benchmarkEncode(b *testing.B, c codec.Codec) {
	e := sampleEvent()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := c.Encode(e); err != nil {
			b.Fatalf("encode error: %v", err)
		}
	}
}

func benchmarkDecode(b *testing.B, c codec.Codec) {
	e := sampleEvent()
	data, err := c.Encode(e)
	if err != nil {
		b.Fatalf("setup encode error: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := c.Decode(data); err != nil {
			b.Fatalf("decode error: %v", err)
		}
	}
}

func BenchmarkEncode_JSON(b *testing.B)     { benchmarkEncode(b, codec.NewJSONCodec()) }
func BenchmarkEncode_Msgpack(b *testing.B)  { benchmarkEncode(b, codec.NewMsgpackCodec()) }
func BenchmarkEncode_Protobuf(b *testing.B) { benchmarkEncode(b, codec.NewProtoCodec()) }

func BenchmarkDecode_JSON(b *testing.B)     { benchmarkDecode(b, codec.NewJSONCodec()) }
func BenchmarkDecode_Msgpack(b *testing.B)  { benchmarkDecode(b, codec.NewMsgpackCodec()) }
func BenchmarkDecode_Protobuf(b *testing.B) { benchmarkDecode(b, codec.NewProtoCodec()) }

func BenchmarkPayloadSize(b *testing.B) {
	e := sampleEvent()
	codecs := []codec.Codec{
		codec.NewJSONCodec(),
		codec.NewMsgpackCodec(),
		codec.NewProtoCodec(),
	}

	for _, c := range codecs {
		data, err := c.Encode(e)
		if err != nil {
			b.Fatalf("encode error for %s: %v", c.Name(), err)
		}
		b.Logf("%s payload size: %d bytes", c.Name(), len(data))
	}
}
