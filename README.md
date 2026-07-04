# high-perf-ledger

Pure Event Sourcing Ledger (no framework): all state is derived from an append-only event log on disk (WAL), with periodic snapshotting to speed up replay.

## Goal

Measure and optmize read/write performance of a hand-built event store, comparing serialization strategies (JSON vs Protobuf vs MessagePack) under real load.
