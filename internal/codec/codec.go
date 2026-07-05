package codec

import "github.com/joaodddev/high-perf-ledger/internal/event"

type Codec interface {
	Encode(e event.Event) ([]byte, error)
	Decode(data []byte) (event.Event, error)
	Name() string
}
