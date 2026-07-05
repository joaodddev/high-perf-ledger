package ledger

import "time"

type EventType string

const (
	EventAccountOpened    EventType = "account_opened"
	EventDeposited        EventType = "deposited"
	EventWithdrawn        EventType = "withdrawn"
	EventTransferSent     EventType = "transfer_sent"
	EventTransferReceived EventType = "transfer_received"
)

type Event struct {
	ID        uint64
	AccountID string
	Type      EventType
	Amount    int64
	Timestamp time.Time
	Metadata  map[string]string
}
