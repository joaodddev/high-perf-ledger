package ledger

import (
	"path/filepath"
	"testing"

	"github.com/joaodddev/high-perf-ledger/internal/codec"
	"github.com/joaodddev/high-perf-ledger/internal/event"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	store, err := NewStore(Config{
		WALPath:     filepath.Join(dir, "test.wal"),
		SnapshotDir: dir,
		Codec:       codec.NewJSONCodec(),
	})
	if err != nil {
		t.Fatalf("unexpected error creating store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestStore_AppendAndGetAccount(t *testing.T) {
	store := newTestStore(t)

	events := []event.Event{
		{AccountID: "acc-1", Type: event.AccountOpened},
		{AccountID: "acc-1", Type: event.Deposited, Amount: 10_000},
		{AccountID: "acc-1", Type: event.Withdrawn, Amount: 3_000},
	}

	for _, e := range events {
		if _, err := store.Append(e); err != nil {
			t.Fatalf("unexpected error appending: %v", err)
		}
	}

	acc, err := store.GetAccount("acc-1")
	if err != nil {
		t.Fatalf("unexpected error replaying account: %v", err)
	}
	if acc.Balance != 7_000 {
		t.Fatalf("expected balance 7000, got %d", acc.Balance)
	}
}

func TestStore_FiltersEventsByAccount(t *testing.T) {
	store := newTestStore(t)

	allEvents := []event.Event{
		{AccountID: "acc-1", Type: event.AccountOpened},
		{AccountID: "acc-2", Type: event.AccountOpened},
		{AccountID: "acc-1", Type: event.Deposited, Amount: 5_000},
		{AccountID: "acc-2", Type: event.Deposited, Amount: 1_000},
		{AccountID: "acc-1", Type: event.Deposited, Amount: 2_000},
	}

	for _, e := range allEvents {
		if _, err := store.Append(e); err != nil {
			t.Fatalf("unexpected error appending: %v", err)
		}
	}

	acc1, err := store.GetAccount("acc-1")
	if err != nil {
		t.Fatalf("unexpected error replaying acc-1: %v", err)
	}
	if acc1.Balance != 7_000 {
		t.Fatalf("expected acc-1 balance 7000, got %d", acc1.Balance)
	}

	acc2, err := store.GetAccount("acc-2")
	if err != nil {
		t.Fatalf("unexpected error replaying acc-2: %v", err)
	}
	if acc2.Balance != 1_000 {
		t.Fatalf("expected acc-2 balance 1000, got %d", acc2.Balance)
	}
}

func TestStore_GetAccount_NonExistentReturnsEmptyAccount(t *testing.T) {
	store := newTestStore(t)

	acc, err := store.GetAccount("does-not-exist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.Opened {
		t.Fatal("expected unopened account for non-existent ID")
	}
	if acc.Balance != 0 {
		t.Fatalf("expected zero balance, got %d", acc.Balance)
	}
}

func TestStore_ReplayIsConsistentAcrossMultipleReads(t *testing.T) {
	store := newTestStore(t)

	if _, err := store.Append(event.Event{AccountID: "acc-1", Type: event.AccountOpened}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := store.Append(event.Event{AccountID: "acc-1", Type: event.Deposited, Amount: 1_000}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	acc1, err := store.GetAccount("acc-1")
	if err != nil {
		t.Fatalf("unexpected error on first read: %v", err)
	}
	acc2, err := store.GetAccount("acc-1")
	if err != nil {
		t.Fatalf("unexpected error on second read: %v", err)
	}

	if acc1.Balance != acc2.Balance {
		t.Fatalf("replay inconsistency: first read balance=%d, second read balance=%d", acc1.Balance, acc2.Balance)
	}
}

func TestStore_SnapshotThenReplayMatchesFullReplay(t *testing.T) {
	store := newTestStore(t)

	events := []event.Event{
		{AccountID: "acc-1", Type: event.AccountOpened},
		{AccountID: "acc-1", Type: event.Deposited, Amount: 10_000},
		{AccountID: "acc-1", Type: event.Withdrawn, Amount: 2_000},
	}
	for _, e := range events {
		if _, err := store.Append(e); err != nil {
			t.Fatalf("unexpected error appending: %v", err)
		}
	}

	// baseline: full replay without any snapshot
	before, err := store.GetAccount("acc-1")
	if err != nil {
		t.Fatalf("unexpected error on pre-snapshot replay: %v", err)
	}

	if err := store.Snapshot("acc-1"); err != nil {
		t.Fatalf("unexpected error taking snapshot: %v", err)
	}

	// append more events after the snapshot point
	moreEvents := []event.Event{
		{AccountID: "acc-1", Type: event.Deposited, Amount: 5_000},
		{AccountID: "acc-1", Type: event.Withdrawn, Amount: 1_000},
	}
	for _, e := range moreEvents {
		if _, err := store.Append(e); err != nil {
			t.Fatalf("unexpected error appending post-snapshot: %v", err)
		}
	}

	after, err := store.GetAccount("acc-1")
	if err != nil {
		t.Fatalf("unexpected error on post-snapshot replay: %v", err)
	}

	expectedBalance := before.Balance + 5_000 - 1_000
	if after.Balance != expectedBalance {
		t.Fatalf("expected balance %d after snapshot+new events, got %d", expectedBalance, after.Balance)
	}
}

func TestStore_SnapshotOfNonExistentAccountIsEmpty(t *testing.T) {
	store := newTestStore(t)

	if err := store.Snapshot("ghost-account"); err != nil {
		t.Fatalf("unexpected error snapshotting non-existent account: %v", err)
	}

	acc, err := store.GetAccount("ghost-account")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.Opened {
		t.Fatal("expected ghost account to remain unopened")
	}
}

func TestStore_MultipleSnapshotsOverwritePrevious(t *testing.T) {
	store := newTestStore(t)

	if _, err := store.Append(event.Event{AccountID: "acc-1", Type: event.AccountOpened}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := store.Append(event.Event{AccountID: "acc-1", Type: event.Deposited, Amount: 1_000}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := store.Snapshot("acc-1"); err != nil {
		t.Fatalf("unexpected error on first snapshot: %v", err)
	}

	if _, err := store.Append(event.Event{AccountID: "acc-1", Type: event.Deposited, Amount: 2_000}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := store.Snapshot("acc-1"); err != nil {
		t.Fatalf("unexpected error on second snapshot: %v", err)
	}

	acc, err := store.GetAccount("acc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.Balance != 3_000 {
		t.Fatalf("expected balance 3000 after two snapshots, got %d", acc.Balance)
	}
}
