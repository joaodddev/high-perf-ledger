package ledger

import (
	"testing"

	"github.com/joaodddev/high-perf-ledger/internal/event"
)

func TestAccount_ReplayDepositsAndWithdrawals(t *testing.T) {
	events := []event.Event{
		{ID: 1, AccountID: "acc-1", Type: event.AccountOpened},
		{ID: 2, AccountID: "acc-1", Type: event.Deposited, Amount: 10_000},
		{ID: 3, AccountID: "acc-1", Type: event.Withdrawn, Amount: 3_000},
	}

	acc, err := Replay("acc-1", events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.Balance != 7_000 {
		t.Fatalf("expected balance 7000, got %d", acc.Balance)
	}
	if acc.version != 3 {
		t.Fatalf("expected version 3, got %d", acc.version)
	}
}

func TestAccount_RejectsDoubleOpen(t *testing.T) {
	events := []event.Event{
		{ID: 1, AccountID: "acc-1", Type: event.AccountOpened},
		{ID: 2, AccountID: "acc-1", Type: event.AccountOpened},
	}

	_, err := Replay("acc-1", events)
	if err == nil {
		t.Fatal("expected error when opening account twice")
	}
}

func TestAccount_RejectsWithdrawalBeforeOpen(t *testing.T) {
	events := []event.Event{
		{ID: 1, AccountID: "acc-1", Type: event.Withdrawn, Amount: 100},
	}

	_, err := Replay("acc-1", events)
	if err == nil {
		t.Fatal("expected error when withdrawing from unopened account")
	}
}

func TestAccount_RejectsOverdraft(t *testing.T) {
	events := []event.Event{
		{ID: 1, AccountID: "acc-1", Type: event.AccountOpened},
		{ID: 2, AccountID: "acc-1", Type: event.Deposited, Amount: 1_000},
		{ID: 3, AccountID: "acc-1", Type: event.Withdrawn, Amount: 5_000},
	}

	_, err := Replay("acc-1", events)
	if err == nil {
		t.Fatal("expected error when withdrawal exceeds balance")
	}
}

func TestAccount_TransferSentAndReceived(t *testing.T) {
	sender := []event.Event{
		{ID: 1, AccountID: "acc-1", Type: event.AccountOpened},
		{ID: 2, AccountID: "acc-1", Type: event.Deposited, Amount: 5_000},
		{ID: 3, AccountID: "acc-1", Type: event.TransferSent, Amount: 2_000, Metadata: map[string]string{"to": "acc-2"}},
	}
	receiver := []event.Event{
		{ID: 1, AccountID: "acc-2", Type: event.AccountOpened},
		{ID: 2, AccountID: "acc-2", Type: event.TransferReceived, Amount: 2_000, Metadata: map[string]string{"from": "acc-1"}},
	}

	senderAcc, err := Replay("acc-1", sender)
	if err != nil {
		t.Fatalf("unexpected error replaying sender: %v", err)
	}
	if senderAcc.Balance != 3_000 {
		t.Fatalf("expected sender balance 3000, got %d", senderAcc.Balance)
	}

	receiverAcc, err := Replay("acc-2", receiver)
	if err != nil {
		t.Fatalf("unexpected error replaying receiver: %v", err)
	}
	if receiverAcc.Balance != 2_000 {
		t.Fatalf("expected receiver balance 2000, got %d", receiverAcc.Balance)
	}
}

func TestAccount_RejectsUnknownEventType(t *testing.T) {
	events := []event.Event{
		{ID: 1, AccountID: "acc-1", Type: "unknown_event"},
	}

	_, err := Replay("acc-1", events)
	if err == nil {
		t.Fatal("expected error for unknown event type")
	}
}
