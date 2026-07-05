package codec

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/joaodddev/high-perf-ledger/internal/codec/pb"
	"github.com/joaodddev/high-perf-ledger/internal/event"
)

type ProtoCodec struct{}

func NewProtoCodec() *ProtoCodec {
	return &ProtoCodec{}
}

func (c *ProtoCodec) Encode(e event.Event) ([]byte, error) {
	msg := &pb.Event{
		Id:                e.ID,
		AccountId:         e.AccountID,
		Type:              string(e.Type),
		Amount:            e.Amount,
		TimestampUnixNano: e.Timestamp.UnixNano(),
		Metadata:          e.Metadata,
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("codec: proto marshal: %w", err)
	}
	return data, nil
}

func (c *ProtoCodec) Decode(data []byte) (event.Event, error) {
	msg := &pb.Event{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return event.Event{}, fmt.Errorf("codec: proto unmarshal: %w", err)
	}

	return event.Event{
		ID:        msg.Id,
		AccountID: msg.AccountId,
		Type:      event.Type(msg.Type),
		Amount:    msg.Amount,
		Timestamp: timeFromUnixNano(msg.TimestampUnixNano),
		Metadata:  msg.Metadata,
	}, nil
}

func (c *ProtoCodec) Name() string {
	return "protobuf"
}

func timeFromUnixNano(nanos int64) (t time.Time) {
	return time.Unix(0, nanos)
}
