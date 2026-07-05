# high-perf-ledger

Pure Event Sourcing ledger (no framework): all state is derived from
an append-only event log on disk (WAL), with periodic snapshotting to
speed up replay. Includes a pluggable codec layer, benchmarked across
JSON, MessagePack, and Protobuf.

## Architecture

\`\`\`
Store.Append(event) → wal.Writer → [length-prefixed frame] → disk (.wal)
Store.GetAccount(id) → snapshot.Read (if exists) → wal.Reader.ReadAfter → Account.Apply (replay)
Store.Snapshot(id)   → GetAccount (full replay) → snapshot.Write (atomic rename)
\`\`\`

- `internal/event` — the `Event` type and its kinds; zero dependencies,
  the foundation every other package builds on
- `internal/codec` — pluggable `Encoder`/`Decoder`: JSON, MessagePack, Protobuf
- `internal/wal` — append-only writer/reader with length-prefixed framing
- `internal/snapshot` — atomic snapshot write (temp file + rename) and read
- `internal/ledger` — `Account` aggregate (pure business rules) and
  `Store` (orchestrates WAL + snapshot + replay)

## Why Event Sourcing here

Every state change is an immutable fact appended to the log — nothing
is ever updated in place. Current state (`Account.Balance`, etc.) is
always a *derived* value, reconstructed by replaying events through
`Account.Apply`. This gives a full audit trail for free and makes
`Replay` deterministic: the same event history always produces the
same state, proven by `TestStore_ReplayIsConsistentAcrossMultipleReads`.

## Snapshotting

Without snapshots, `GetAccount` would replay the entire WAL on every
call — fine for small logs, a real bottleneck as they grow. A snapshot
stores an account's already-replayed state plus the ID of the last
applied event; future reads only need to replay events after that
point (`wal.Reader.ReadAfter`).

**Known limitation:** `ReadAfter` still scans the full file from disk
before filtering in memory — it avoids redundant `Account.Apply` calls,
not the disk I/O itself. A byte-offset index (`eventID -> file position`)
would allow seeking directly; deferred since profiling didn't show this
as the actual bottleneck at this log size (see Benchmarks below).

## Benchmarks

Measured on: Intel Core i5-5200U @ 2.20GHz, single machine, local disk.

### Encode

| Codec | ns/op | B/op | allocs/op |
|---|---|---|---|
| JSON | 2629 | 560 | 10 |
| MessagePack | **1534** | 576 | **5** |
| Protobuf | 3367 | **416** | 14 |

### Decode

| Codec | ns/op | B/op | allocs/op |
|---|---|---|---|
| JSON | 6562 | 800 | 20 |
| MessagePack | **2340** | 560 | 12 |
| Protobuf | 2722 | 736 | 23 |

### Payload size (single representative event)

| Codec | Bytes |
|---|---|
| JSON | 207 |
| MessagePack | 155 |
| Protobuf | **112** |

**Takeaway:** MessagePack won on raw encode/decode speed and allocation
count in this benchmark — likely because Protobuf's reflection-based
Go implementation carries overhead that outweighs its compact wire
format at this message size. Protobuf still produces the smallest
payload, which matters more for network-bound or storage-bound
workloads than for CPU-bound serialization throughput. JSON is the
slowest and largest on every axis, as expected — kept as the default
for readability during development, not for production throughput.

This is a case where the "obviously more efficient" binary format
(Protobuf) doesn't win on every axis — worth measuring rather than
assuming.

## Design decisions

- **Amounts stored as `int64` cents**, never floats — avoids accumulated
  rounding error in a financial ledger.
- **Length-prefixed framing in the WAL** (`[4-byte length][payload]`):
  necessary because none of the three codecs have a native "where does
  this record end" delimiter when concatenated in a single binary file.
- **Snapshot writes use temp-file + atomic rename**: readers never
  observe a partially-written snapshot, even if the process crashes
  mid-write.
- **`Store.Append` does not validate business rules before writing**
  (e.g. sufficient balance) — invalid events are only caught at replay
  time via `Account.Apply`. A production system would typically
  replay-then-validate-then-append as a single atomic operation; this
  simplified version accepts that trade-off. Documented here rather
  than silently glossed over.
- **event → codec → wal → ledger** is a strict one-way dependency
  chain — this was violated once during development (`ledger` importing
  `wal`/`codec`, which imported `ledger` for

## Stack

Go 1.22 · encoding/json · vmihailenco/msgpack/v5 · google.golang.org/protobuf
