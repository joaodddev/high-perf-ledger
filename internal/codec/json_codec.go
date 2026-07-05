package codec

import (
	"encoding/json"

	"github.com/joaodddev/high-perf-ledger/internal/ledger"
)

type JSONCodec struct{}

func NewJSONCodec() *JSONCodec {
	return &JSONCodec{}
}

func (c *JSONCodec) Encode(e ledger.Event) ([]byte, error) {
	return json.Marshal(e)
}

func (c *JSONCodec) Decode(data []byte) (ledger.Event, error) {
	var e ledger.Event
	if err := json.Unmarshal(data, &e); err != nil {
		return ledger.Event{}, err
	}
	return e, nil
}

func (c *JSONCodec) Name() string {
	return "json"
}
