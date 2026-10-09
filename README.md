# Sovereign Settlement Ledger (SSL)

A settlement engine in Go + PostgreSQL, built to prove correctness
under concurrency and regulatory constraint — not to ship features fast.

**Core invariant:** `available_balance = balance − SUM(pending_holds) ≥ 0`
Every design decision in this repo exists to keep that true under
concurrent access, not just in the happy path.

## What's demonstrated here

- **Concurrency-safe worker pool**: priority `select` handling
  non-blocking context cancellation, bounded shutdown under a
  Kubernetes-style termination grace period, race-free under
  `go run -race` with mid-run interrupt.
- **Write-skew anomaly**, reproduced and closed:
  - Under `RepeatableRead`, two concurrent holds against the same
    account, on disjoint rows, both pass validation independently
    and together violate the invariant.
  - Under `Serializable` + a transaction-scoped retry loop, the
    second transaction aborts (`SQLSTATE 40001`), retries against a
    fresh snapshot, and correctly rejects.
  - See `experiments/writeskew/main.go`.

## Why this matters

Most CRUD apps never notice write skew because most invariants are
single-row. A settlement invariant is cross-row and cross-transaction
by nature — this repo is a deliberate demonstration of the isolation
level and retry discipline that class of invariant requires.

## Running it

```bash
psql ssl -f db/schema.sql
psql ssl -f db/reset.sql        # reset before every run
go run ./experiments/writeskew
```

## Structure

```
main.go, settler.go        engine (concurrency block)
db/schema.sql, db/reset.sql
experiments/writeskew/     write-skew scenario (isolated main)
```

## Status

- [x] Concurrency block (priority select, bounded shutdown, LIFO defers)
- [x] Write-skew anomaly (RR reproduction, Serializable + retry close)
- [x] Dirty read anomaly (simulated uncommitted-state mutation, MVCC bypass)
- [ ] Non-repeatable read anomaly
- [ ] Phantom anomaly
- [ ] Property-based tests (`pgregory.net/rapid`) generating interleavings
- [ ] `specs/` — invariant in precise English
