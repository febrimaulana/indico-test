# ARCHITECTURE

## 1. Architectural Design & Database Locking Strategy

**Schema.** `items(item_id PK, total_stock, reserved_stock)` stores the counters directly;
`available_stock` is always derived as `total_stock - reserved_stock`, so reads never scan
reservations. `reservations(reservation_id PK, user_id, item_id FK, quantity, status,
expires_at, confirmed_at)` is the audit trail. Database `CHECK` constraints
(`reserved_stock >= 0`, `reserved_stock <= total_stock`, `quantity > 0`, status enum) are the
last line of defense: even a logic bug cannot push the database into a negative or
oversold state. Indexes: a partial index on `(expires_at) WHERE status = 'active'` for the
sweeper, `(item_id, status)` for per-item lookups, and `(user_id)` for audits.

**Locking strategy.** Reserve is one transaction with one conditional statement:

```sql
UPDATE items SET reserved_stock = reserved_stock + $qty
WHERE item_id = $1 AND total_stock - reserved_stock >= $qty;
```

Zero rows updated means either the item does not exist (404) or stock is insufficient
(409). I deliberately rejected the classic `SELECT ... FOR UPDATE` then `UPDATE` pattern:
it holds a row lock across a network round trip and makes every writer queue on a
read that must precede the write. The single conditional update takes the row lock only
for the duration of one statement and, under PostgreSQL's READ COMMITTED, the predicate
is re-evaluated after lock acquisition (EvalPlanQual), so concurrent updates cannot
double-spend stock. Serializable isolation was unnecessary and would add retry storms
under flash-sale contention.

Confirm uses the same idea as a guarded state transition:

```sql
UPDATE reservations SET status = 'confirmed', confirmed_at = now()
WHERE reservation_id = $1 AND status = 'active' AND expires_at > now()
RETURNING item_id, quantity;
```

The loser of a double-confirm race gets zero rows and is classified (404 unknown, 409
already confirmed, 410 expired). Physical stock (`total_stock` and `reserved_stock`) is
decremented in the same transaction, so either both the reservation and the counters
change or neither does.

**Expiry.** Expired holds are returned by an in-process sweeper that claims one batch per
tick with `UPDATE ... WHERE reservation_id IN (SELECT ... FOR UPDATE SKIP LOCKED)`. Claiming
and releasing happen in one transaction, releases are aggregated per item and executed in
sorted `item_id` order (deadlock avoidance). The sweeper runs on startup and every 15s;
correctness does not depend on its timing because confirm re-checks `expires_at > now()`.

**Gin routing & boundaries.** Feature-first modular monolith: `internal/modules/inventory`
owns model/dto/errors/repository/service/handler/routes, while `internal/platform` owns
config, database, HTTP envelope, validator, middleware and server. Each module exposes a
service interface and registers its own routes; handlers never touch SQL, services never
import GORM, repositories never hold business rules. Transactions are mediated by a
`TxManager` interface that carries the transaction in the context, keeping GORM out of the
service layer (DIP).

**Graceful shutdown.** SIGINT/SIGTERM cancels the root context: `http.Server.Shutdown`
drains in-flight requests within `SHUTDOWN_TIMEOUT`, then the sweeper is canceled and only
after that is the connection pool closed — no request can observe a closed pool.

## 2. Distributed Scaling & Failure Modes (10 Instances)

- **Sweeper duplication** — with `FOR UPDATE SKIP LOCKED` and conditional transitions,
  multiple sweepers can race safely: each row is claimed exactly once and losers skip; the
  cost is only wasted idle work, never double release.
- **Connection pool exhaustion** — 10 × `DB_MAX_OPEN_CONNS` (default 25) = 250 connections
  versus PostgreSQL's `max_connections` (~100 default) is the first thing to break. Fix:
  PgBouncer in transaction mode, or a much smaller per-instance pool budget.
- **Hot-row contention** — all buyers of one flash item serialize on the same `items` row;
  throughput is bounded by single-row lock service time. Fixes, in increasing order of
  effort: application-level rate limiting; queueing writes (Kafka → worker) with a
  "please retry" response; moving the decrement to Redis (`DECRBY` with Lua for
  near-linear scaling) and syncing to PostgreSQL asynchronously; sharding one item's stock
  into N counters.
- **Clock skew** — expiry currently uses each instance's UTC clock. Across hosts, skew
  shifts the effective 5-minute window; the fix is to derive time from the database
  (`now()`) or enforce NTP, and to treat `expires_at` comparisons as advisory (confirm
  always re-validates in SQL).
- **Migrations** — golang-migrate takes an advisory lock, so booting 10 instances that run
  migrations is safe; only one applies, the rest wait and no-op.
- **Statelessness** — the HTTP tier already holds no session state (reservation state
  lives in PostgreSQL and the browser), so it scales horizontally behind a load balancer.
  Redesign for a fully distributed deployment: extract the sweeper into a dedicated
  worker with leader election (or a single scheduler with `SKIP LOCKED` work stealing),
  add idempotency keys on reserve/confirm, serve stock reads from replicas with a
  short-TTL cache, and add metrics/traces for lock wait and pool saturation. Clients
  should treat 409/410 as retryable workflow states, not crashes.

## 3. Engineering Trade-offs & AI Transparency

- **Denormalized counters vs derived sums**: counters make reads O(1) and locking precise,
  at the cost of every stock change touching one row; derived sums would remove the race
  entirely but make stock reads scale with reservation volume. For flash sales, the
  counter wins.
- **Periodic sweep vs database triggers/jobs**: the sweeper adds ≤15s of staleness in
  exchange for simple, observable logic; the confirm guard closes the correctness gap.
- **Scope discipline (4–8h window)**: no auth, rate limiting, metrics, CI, or frontend
  tests. The 100% coverage gate was limited to `./internal/...` and driven by cheap,
  honest seams (injectable clock, id generator, migrate/listen function variables) rather
  than mock-heavy tests or coverage-suppression tricks.
- **Frontend**: Axios + TanStack Query replaced hand-rolled fetch/polling state — polling,
  deduplication, abort-on-unmount, retry policy (no retries for 4xx) and cache invalidation
  after mutations fall out of the library. `React.memo`/`useMemo`/`useCallback` are applied
  only where re-renders are measurable (leaf cards), not everywhere.
- **AI transparency**: AI suggestions were reviewed, not trusted. Two concrete examples:
  (1) an AI-proposed `SELECT ... FOR UPDATE` + "sum of active reservations" availability
  query was flawed — it holds locks across round trips and turns reads into scans; we
  replaced it with the single conditional `UPDATE` and a counter column. (2) AI-generated
  validation initially required the literal `usr_`/`item_` prefixes from the assignment's
  examples, which would reject legitimate payloads such as `user123`; we relaxed it to
  length/charset rules plus existence checks (404), keeping the examples as suggestions
  rather than contract. Both fixes are covered by tests.
