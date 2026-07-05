package codec

import "github.com/joaodddev/high-perf-ledger/internal/ledger"

type Codec interface {
	Encode(e ledger.Event) ([]byte, error)
	Decode(data []byte) (ledger.Event, error)

	Name() string
}
