# Backend Rules (Go + Gin + GORM)

Feature-first modular monolith. One domain = one folder in `internal/modules/<domain>/`.

## Required file convention inside a module

| File | Responsibility |
|---|---|
| `model.go` | GORM entities / domain types. No HTTP, no business rules. |
| `dto.go` | Request/response structs + `binding` and `example` tags. Contract only. |
| `errors.go` | Sentinel errors for this module (e.g. `ErrInsufficientStock`). |
| `repository.go` | Repository **interfaces** only. |
| `repository_gorm.go` | GORM implementation. Queries only — NO business rules. |
| `service.go` | Business logic. MUST NOT import `gorm.io/gorm`. Transactions via `platform/database.TxManager`. |
| `expiry_worker.go` | Background worker (if any). Ticker + context; stops on shutdown. |
| `handler.go` | Gin handlers: bind → validate → service → response. NO DB access. |
| `routes.go` | The ONLY place this module registers routes. |
| `module.go` | Constructor wiring: repo → service → handler. |
| `*_test.go` | Unit tests colocated, fake repositories, no network/DB. |

## Query & Concurrency Rules

- Reserve MUST be a single atomic conditional update:
  `UPDATE items SET reserved_stock = reserved_stock + $qty WHERE item_id = $1 AND total_stock - reserved_stock >= $qty RETURNING ...`.
  Never read-then-write without locking.
- Confirm MUST use a conditional status transition guarded by `expires_at > now()` in the same transaction that decrements stock.
- Expiry sweeper MUST use `FOR UPDATE SKIP LOCKED` batches (safe for multiple instances).
- Lock order rule: touch `reservations` before `items`; when updating multiple items in one transaction, update them ordered by `item_id` (deadlock avoidance).
- Every new query must be justified with an index. Add the index in a migration.
- Always pass `c.Request.Context()` (or worker ctx) into GORM via `WithContext`.

## Error & Response Rules

- Sentinel errors in `errors.go`; handlers map them to HTTP + frozen error codes.
- Validation messages in Bahasa Indonesia; codes in English (see root AGENTS.md).
- Additional endpoints beyond the spec are allowed only as platform endpoints (`/api/v1/healthz`, `/swagger/*`).

## Swagger Rules

- Every endpoint MUST have swagger annotations: `@Summary`, `@Description` (including business rules), `@Param`, `@Success`, and at least 2 `@Failure` entries.
- After changing handlers/DTOs: run `make swag` (regenerates `docs/` — commit the output).
- Swagger UI must stay reachable at `/swagger/index.html`.

## Testing Policy

- Unit tests: fakes for repository/TxManager/Clock/IDGenerator. No DB, no network.
- Integration tests (`integration_test.go`): require `TEST_DATABASE_URL`; auto-`t.Skip` when unset. Truncate tables + re-seed between tests.
- New endpoint/behavior = success test + minimum 2 negative tests.
- Coverage gate: `make coverage` must report 100% for `./internal/...` (unit + integration combined). `cmd/` is excluded.
- Reachability seams are mandatory for 100% coverage: inject Clock, ID generator, random reader, migrate/listen function vars. Do NOT write fake/empty tests to inflate coverage.

## Commands

```bash
make run               # local run
make test              # unit only
make test-race         # go test -race -v ./...
make test-integration  # needs docker compose postgres + TEST_DATABASE_URL
make coverage          # unit + integration, fails < 100% on ./internal/...
make swag              # regenerate swagger docs
```
