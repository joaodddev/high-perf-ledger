package codec

import (
	"github.com/vmihailenco/msgpack/v5"

	"github.com/joaodddev/high-perf-ledger/internal/event"
)

type MsgpackCodec struct{}

func NewMsgpackCodec() *MsgpackCodec {
	return &MsgpackCodec{}
}

func (c *MsgpackCodec) Encode(e event.Event) ([]byte, error) {
	return msgpack.Marshal(e)
}

func (c *MsgpackCodec) Decode(data []byte) (event.Event, error) {
	var e event.Event
	if err := msgpack.Unmarshal(data, &e); err != nil {
		return event.Event{}, err
	}
	return e, nil
}

func (c *MsgpackCodec) Name() string {
	return "msgpack"
}
