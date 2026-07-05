package ledger

import "fmt"

type Account struct {
	ID      string
	Balance int64 // cents
	Opened  bool

	version uint64
}

func NewAccount(id string) *Account {
	return &Account{ID: id}
}

func (a *Account) Apply(e Event) error {
	switch e.Type {
	case EventAccountOpened:
		if a.Opened {
			return fmt.Errorf("ledger: account %s already opened", a.ID)
		}
		a.Opened = true

	case EventDeposited:
		if !a.Opened {
			return fmt.Errorf("ledger: cannot deposit into unopened account %s", a.ID)
		}
		a.Balance += e.Amount

	case EventWithdrawn:
		if !a.Opened {
			return fmt.Errorf("ledger: cannot withdraw from unopened account %s", a.ID)
		}
		if a.Balance < e.Amount {
			return fmt.Errorf("ledger: insufficient funds in account %s", a.ID)
		}
		a.Balance -= e.Amount

	case EventTransferSent:
		if a.Balance < e.Amount {
			return fmt.Errorf("ledger: insufficient funds for transfer from account %s", a.ID)
		}
		a.Balance -= e.Amount

	case EventTransferReceived:
		a.Balance += e.Amount

	default:
		return fmt.Errorf("ledger: unknown event type %q", e.Type)
	}

	a.version = e.ID
	return nil
}

func Replay(id string, events []Event) (*Account, error) {
	acc := NewAccount(id)
	for _, e := range events {
		if err := acc.Apply(e); err != nil {
			return nil, err
		}
	}
	return acc, nil
}
