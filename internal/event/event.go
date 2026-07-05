package event

import "time"

type Type string

const (
	AccountOpened    Type = "account_opened"
	Deposited        Type = "deposited"
	Withdrawn        Type = "withdrawn"
	TransferSent     Type = "transfer_sent"
	TransferReceived Type = "transfer_received"
)

type Event struct {
	ID        uint64
	AccountID string
	Type      Type
	Amount    int64
	Timestamp time.Time
	Metadata  map[string]string
}
