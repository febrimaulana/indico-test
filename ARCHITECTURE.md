# ARCHITECTURE

## 1. Architectural Design & Database Locking Strategy

**Schema.** `items(item_id PK, total_stock, reserved_stock)` stores counters; `available_stock` is
derived (`total - reserved`), so reads never scan reservations. `reservations(reservation_id PK,
user_id, item_id FK, quantity, status, expires_at, confirmed_at)` is the audit trail. `CHECK`
constraints (`reserved_stock >= 0`, `reserved_stock <= total_stock`, `quantity > 0`, status enum)
are the last line of defense: a logic bug cannot produce negative or oversold stock. Indexes: a
partial `(expires_at) WHERE status = 'active'` for the sweeper (EXPLAIN shows an index scan with no
sort), PKs for every hot-path lookup, plus `(item_id, status)` and `(user_id)` kept for future
per-item/per-user reads — `pg_stat_user_indexes` confirms the partial index and PKs carry all
current traffic.

**Locking.** Reserve is one transaction with one conditional statement:

```sql
UPDATE items SET reserved_stock = reserved_stock + $qty
WHERE item_id = $1 AND total_stock - reserved_stock >= $qty;
```

Zero rows means the item is missing (404) or stock is insufficient (409). I rejected
`SELECT ... FOR UPDATE` then `UPDATE`: it holds a row lock across a network round trip and queues
every writer behind a read that must precede the write. Under READ COMMITTED the predicate is
re-evaluated after lock acquisition (EvalPlanQual), so concurrent updates cannot double-spend;
SERIALIZABLE would only add retry storms.

Confirm is a guarded transition:

```sql
UPDATE reservations SET status = 'confirmed', confirmed_at = now()
WHERE reservation_id = $1 AND status = 'active' AND expires_at > now()
RETURNING item_id, quantity;
```

The loser of a double-confirm is classified (404 unknown, 409 already confirmed, 410 expired), and
physical stock (`total_stock` and `reserved_stock`) is decremented in the same transaction.

**Expiry.** A sweeper claims one batch per tick with `FOR UPDATE SKIP LOCKED`, aggregates per item
and releases in sorted `item_id` order (deadlock avoidance) inside one transaction. It runs on
startup and every 15s; correctness never depends on its timing because confirm re-checks
`expires_at` in SQL.

**Gin routing & boundaries.** Feature-first modular monolith: `internal/modules/inventory` owns
model/dto/errors/repository/service/handler/routes, `internal/platform` owns config, database,
HTTP envelope, validator, middleware and server. Handlers never touch SQL, services never import
GORM (a `TxManager` interface carries the transaction in the context), repositories hold no
business rules. On SIGINT/SIGTERM the server drains in-flight requests, then the sweeper stops,
then the pool closes — no request can observe a closed pool.

## 2. Distributed Scaling & Failure Modes (10 Instances)

- **Sweeper duplication** — `SKIP LOCKED` plus conditional transitions claim each row exactly
  once; extra instances only waste idle work, never double-release.
- **Pool exhaustion** — 10 × 25 pooled connections versus PostgreSQL's `max_connections` (~100)
  breaks first: fix with PgBouncer (transaction mode) or a smaller per-instance budget.
- **Hot-row contention** — all buyers of one flash item serialize on the same row. The fix ladder:
  rate limiting → write queue (Kafka) with retry responses → Redis `DECRBY` plus async persistence
  → sharded stock counters.
- **Clock skew** — expiry uses instance clocks; derive time from the database (`now()`) or NTP, and
  treat `expires_at` as advisory since confirm always re-validates in SQL.
- **Migrations** — golang-migrate takes an advisory lock, so 10 booting instances cannot corrupt
  versions.
- **Stateless redesign** — the HTTP tier holds no session state already. Extract the sweeper into a
  leader-elected worker (or `SKIP LOCKED` work stealing), add idempotency keys, serve stock reads
  from replicas with a short-TTL cache, and add lock-wait/pool metrics. Clients treat 409/410 as
  workflow states, not crashes.

## 3. Engineering Trade-offs & AI Transparency

- **Counters vs derived sums** — O(1) reads and precise locking cost one hot row per mutation;
  derived sums would remove the race but make reads scale with reservation volume.
- **Periodic sweep vs triggers/jobs** — ≤15s staleness and simple observability; the confirm guard
  closes the correctness gap.
- **Scope discipline (4–8h)** — no auth, rate limiting, metrics, CI or frontend tests. The 100%
  coverage gate is limited to `./internal/...` and driven by honest seams (injectable clock, ID
  generator, migrate/listen function variables) rather than mock-heavy or empty tests.
- **Frontend** — Axios + TanStack Query replace hand-rolled polling state (deduplication,
  abort-on-unmount, no 4xx retries, invalidation after mutations); `React.memo`/`useMemo`/
  `useCallback` are applied only where re-renders are measurable.
- **AI transparency** — suggestions were reviewed, not trusted. (1) An AI-proposed `SELECT ...
  FOR UPDATE` plus “sum of active reservations” availability query would hold locks across round
  trips and turn reads into scans; replaced with the single conditional `UPDATE` and a counter
  column. (2) AI-generated validation initially required the literal `usr_`/`item_` prefixes from
  the assignment examples, rejecting legitimate payloads such as `user123`; relaxed to
  length/charset rules plus existence checks (404). Both fixes are covered by tests.
