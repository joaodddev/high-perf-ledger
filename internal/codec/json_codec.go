package codec

import (
	"encoding/json"

	"github.com/joaodddev/high-perf-ledger/internal/event"
)

type JSONCodec struct{}

func NewJSONCodec() *JSONCodec {
	return &JSONCodec{}
}

func (c *JSONCodec) Encode(e event.Event) ([]byte, error) {
	return json.Marshal(e)
}

func (c *JSONCodec) Decode(data []byte) (event.Event, error) {
	var e event.Event
	if err := json.Unmarshal(data, &e); err != nil {
		return event.Event{}, err
	}
	return e, nil
}

func (c *JSONCodec) Name() string {
	return "json"
}
