# Content map

Shared ledger between the SSL build and the social content engine. Both Claude scopes read and write
this file. Git is the sync layer: pull at the start of a session, commit after any change.

**Naming:** `build-<thing>` for back-room/technical reels, `story-<nn>-<thing>` for front-door/personal.
Old bank IDs kept as aliases so W39-W42 reviews still resolve.

Last updated: 20 Oct 2026

---

## The map

| Build block | Build status | Reel | Reel status | Spoiler guard |
|---|---|---|---|---|
| Write-skew demo (`experiments/writeskew`) | closed | `build-write-skew` (alias B1) | footage shot, unscripted | **Do not show `available: -20` or `ERROR 40001` anywhere until this ships.** Both are its entire payload. Terminal in other reels is texture only: a glimpse of green, no numbers. |
| Dirty read (`experiments/dirtyread`) | closed | unassigned | — | — |
| SELECT-then-INSERT atomicity | closed | `build-select-insert` (alias B2) | not written | ships after write-skew |
| Isolation pendulum (RC/RR/SSI decision rule) | closed 16 Oct | `build-isolation-pendulum` (alias B7) | filed in bank | ships after write-skew and select-insert. It is the *method*; the anomalies come first |
| Retry loop / congestion control | in progress | teaser inside `build-write-skew` | — | retry is the close and the next-episode tease, never the payload of the write-skew reel |
| ACID property tests (pgregory.net/rapid) | not started | unassigned | — | — |
| `settler.go` graceful shutdown | closed | unassigned | — | candidate once the back-room lane has room |

## Shipped

| Reel | Alias | Date | IG | Notes |
|---|---|---|---|---|
| `story-red-yellow-green-01` | F5-R | 26 Sep 2026 | `DdzGJshzgN-` | 4,786 plays, 29.3% skip, 73 saves. Best performer. Pinned. |
| `build-hidden-dependency` | F7 | 4 Oct 2026 | `DeFjdG6MtRW` | 1,687 plays, 30.1% skip, 7 saves. Caption recounted an incident, saves collapsed. |
| `story-restart-01-pivot` | F12a | 11 Oct 2026 | `DeXypLYsrr9` | 275 plays, 64.4% skip. Question opener. Pinned on profile. |
| `story-restart-02-ai-shift` | F12b | 18 Oct 2026 | `Dep3MdysSU9` | 1,239 plays @16h, 40.7% skip, 15 saves. Claim opener. |

## Unblocking rule

A back-room reel becomes filmable when its build block closes **and** the terminal output is legible
at 24-28pt in a narrow window. Closing a block is therefore the trigger to script, not to film.

When a reel ships, record what became public in the Shipped table. That is what prevents spending a
payload as a cutaway.

## Open

- `build-write-skew` reporting fix: on the VIOLATED run `main.go` prints `total holds: 120 (balance 100)`
  with no depth, while the REJECTED path already prints `available=40`. Also `(balance 100)` and
  `total <= 100` hardcode the balance in two places. Fix before filming: find the arithmetic expression
  whose sign carries the verdict and whose magnitude carries the depth.
- No master video saved for `story-restart-02-ai-shift`. Export to `03_SHIP/` before publishing, always.
