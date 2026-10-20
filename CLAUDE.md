#SSL - working rules ## Tiers
🔴 load-bearing (anything touching the invariant, concurrency, money paths): I write it by hand.
No agent writes 🔴 code. Agents may ask questions, review, or explain, never produce it.
🟡 review-assisted. 🟢 boilerplate, paste freely.

## Definition of done for a 🔴 block
Before reporting a block closed, update `docs/content-map.md`:
flip its build row to closed, and note whether it unblocks a reel.
That file is the shared ledger with the content engine. Respect its spoiler guards:
values reserved for an unshipped reel must not appear in any other output.

## Invariant
`balance - SUM(pending_holds) >= 0`, predicate-shaped, so SSL runs SERIALIZABLE with a retry loop.
Serializable detects, the retry loop paces.

## Commands
DB: `postgres://localhost/ssl`
Reset: `psql ssl -e -f db/reset.sql`
Check: `psql ssl -c "SELECT balance - COALESCE((SELECT SUM(amount) FROM pending_holds WHERE account_id=1),0) AS available FROM accounts WHERE id=1;"`
Run: `go run ./experiments/writeskew`

## Commits
Conventional commits: `type(scope): lowercase subject`, body explains the mechanism, not the diff.
