# INDICO — High-Throughput Inventory Reservation System & Mini Dashboard

Flash-sale ready inventory reservation: a **Go + Gin + PostgreSQL** backend with atomic,
no-oversell stock handling and automatic 5-minute reservation expiry, plus a
**React + Vite + TypeScript** mini dashboard for live tracking and confirmation.

Built for the INDICO Full Stack Take Home Test — see [`docs/INDICO-Fullstack-Test.md`](docs/INDICO-Fullstack-Test.md)
for the original assignment and [`ARCHITECTURE.md`](ARCHITECTURE.md) for design decisions.

---

## 🚀 Quick Start

> Prerequisite: **Docker** with Compose v2. Nothing else — no Go, no Node, no `.env` needed.

```bash
git clone <your-repo-url> indico
cd indico
docker compose up
```

First run builds the images (~2–4 minutes). When the logs settle, open:

| Service | URL |
|---|---|
| 🖥️ Mini Dashboard | http://localhost:3000 |
| 📚 Swagger UI (Try it out) | http://localhost:8080/swagger/index.html |
| ❤️ Health check | http://localhost:8080/api/v1/healthz |
| 🐘 PostgreSQL | `localhost:5432` (user/pass/db: `indico`) |

Migrations and seed data (`item_4021` with 100 stock, `item_4022` with 50) run
automatically on backend startup — the dashboard is usable immediately.

Port already taken? Override any port, for example:

```bash
POSTGRES_PORT=5433 docker compose up
```

## 🌐 Service URLs

| Service | Default port | Env override |
|---|---|---|
| Frontend (nginx) | `3000` | `FRONTEND_PORT` |
| Backend (Gin) | `8080` | `BACKEND_PORT` |
| PostgreSQL | `5432` | `POSTGRES_PORT` |

## ✨ Features

- **POST `/api/v1/inventory/reserve`** — atomic stock hold, expires automatically after 5 minutes.
- **POST `/api/v1/inventory/confirm`** — permanently commits a reservation and decrements physical stock.
- **GET `/api/v1/inventory/stock?item_id=...`** — real-time total/reserved/available breakdown.
- **No overselling**: single-statement conditional `UPDATE ... WHERE total_stock - reserved_stock >= qty`.
- **Automatic expiry cleanup**: background sweeper (`FOR UPDATE SKIP LOCKED`) + startup sweep + confirm-time guard.
- **Graceful shutdown**: SIGTERM drains HTTP requests, stops the sweeper, then closes the DB pool.
- **Live dashboard**: polling every 5s, manual refresh, 5-minute countdown, confirm button, and rendered Bahasa Indonesia error states.

## 📡 API Reference

All error responses share one envelope:

```json
{
  "status": "error",
  "error": {
    "code": "INSUFFICIENT_STOCK",
    "message": "stok tidak mencukupi",
    "details": [{ "field": "quantity", "rule": "gt", "message": "quantity harus lebih dari 0" }]
  }
}
```

| Error code | HTTP | Meaning |
|---|---|---|
| `INVALID_REQUEST` | 400 | malformed JSON / failed field validation |
| `ITEM_NOT_FOUND` | 404 | unknown `item_id` |
| `RESERVATION_NOT_FOUND` | 404 | unknown `reservation_id` |
| `INSUFFICIENT_STOCK` | 409 | available stock below requested quantity |
| `RESERVATION_ALREADY_CONFIRMED` | 409 | reservation was already committed |
| `RESERVATION_EXPIRED` | 410 | reservation passed its `expires_at` |
| `INTERNAL_ERROR` | 500 | unexpected server error |

### Reserve stock

```bash
curl -X POST http://localhost:8080/api/v1/inventory/reserve \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"usr_9981","item_id":"item_4021","quantity":2}'
```

```json
{
  "status": "success",
  "reservation_id": "res_883291",
  "item_id": "item_4021",
  "quantity": 2,
  "expires_at": "2026-07-20T16:35:00Z"
}
```

### Confirm reservation

```bash
curl -X POST http://localhost:8080/api/v1/inventory/confirm \
  -H 'Content-Type: application/json' \
  -d '{"reservation_id":"res_883291"}'
```

### Check stock

```bash
curl 'http://localhost:8080/api/v1/inventory/stock?item_id=item_4021'
```

```json
{ "item_id": "item_4021", "total_stock": 100, "reserved_stock": 15, "available_stock": 85 }
```

## 🏗️ Architecture

```
React + Vite (TS)  ── /api/* ──►  nginx (frontend container)
                                     │  reverse proxy
                                     ▼
                            Gin API (backend container)
                          middleware → handler → service
                                     │            │
                                     │      TxManager (GORM)
                                     ▼            ▼
                                 PostgreSQL (compose volume)
                                     ▲
                              Expiry Sweeper (in-process,
                              FOR UPDATE SKIP LOCKED)
```

- Backend: feature-first modular monolith (`internal/modules/inventory`) over a shared
  `internal/platform` kernel. SOLID boundaries: handler = HTTP, service = business rules,
  repository = SQL only.
- Frontend: feature-first (`src/features/inventory`), Axios + TanStack Query for every
  server interaction, `React.memo`/`useMemo`/`useCallback` and cleaned-up effects.

Full rationale in [`ARCHITECTURE.md`](ARCHITECTURE.md).

## ⚙️ Configuration

Every value has a default; override via environment or a root `.env` file.

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | backend HTTP port |
| `DATABASE_URL` | — | required by the backend process |
| `EXPIRY_DURATION` | `5m` | reservation lifetime |
| `SWEEP_INTERVAL` | `15s` | expired reservation sweep interval |
| `SWEEP_BATCH_SIZE` | `500` | max reservations per sweep |
| `SHUTDOWN_TIMEOUT` | `15s` | graceful drain budget |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `CORS_ALLOWED_ORIGINS` | `*` | comma-free list in v1 (`*` for dev) |
| `DB_MAX_OPEN_CONNS` / `DB_MAX_IDLE_CONNS` | `25` / `5` | connection pool |
| `DB_CONNECT_ATTEMPTS` / `DB_CONNECT_RETRY_DELAY` | `10` / `1s` | startup retry when Postgres is cold |

## 💻 Local Development (optional)

```bash
# 1. database only
docker compose up -d postgres

# 2. backend (http://localhost:8080, Swagger enabled)
cd backend
DATABASE_URL='postgres://indico:indico@localhost:5432/indico?sslmode=disable' go run ./cmd/api

# 3. frontend (http://localhost:5173, /api proxied to :8080)
cd frontend
npm install
npm run dev
```

## 🧪 Testing

```bash
cd backend
make test              # unit tests (integration auto-skips without a database)
make test-race         # go test -race -v ./...  (assignment requirement)
make test-integration  # spins up the compose postgres + runs integration/stress tests
make coverage          # unit + integration, fails below 100% of ./internal/...
```

Coverage gate: **100% of `./internal/...`** statements (unit + integration combined,
`-race`). The stress test (`TestIntegrationConcurrentReserveNeverOversells`) fires 200
concurrent reserves at stock 10 and proves exactly 10 succeed.

## 📁 Project Structure

```
indico/
├── docker-compose.yaml       # postgres + backend + frontend, one command
├── README.md · ARCHITECTURE.md · AGENTS.md (+ AGENT.md symlink)
├── docs/INDICO-Fullstack-Test.md
├── backend/
│   ├── cmd/api/              # entry point (thin)
│   ├── internal/
│   │   ├── app/              # composition root + router + lifecycle
│   │   ├── platform/         # config, database, httpx, validator, middleware, server, clock, idgen
│   │   └── modules/inventory # model, dto, errors, repository, service, worker, handler, routes
│   ├── migrations/           # embedded SQL migrations (golang-migrate)
│   └── docs/                 # generated Swagger (swag init)
└── frontend/src/features/inventory   # api, hooks, components, dashboard
```

## 🔧 Troubleshooting

| Symptom | Fix |
|---|---|
| `Bind for 0.0.0.0:5432 failed` | another Postgres is running: `POSTGRES_PORT=5433 docker compose up` |
| Want a clean database | `docker compose down -v && docker compose up` |
| Backend restarts / can't reach DB | `docker compose logs backend` — it retries the connection automatically |
| Regenerate Swagger after changing handlers | `cd backend && make swag` |
| Ports busy | `FRONTEND_PORT=3100 BACKEND_PORT=8181 docker compose up` |

## 🧭 Demo Script (walkthrough calls)

1. `docker compose up` → open http://localhost:3000
2. Watch the tracker poll (reserved 0 / available 100 for `item_4021`).
3. Submit the form (`usr_9981`, quantity 2) → reservation card + 5:00 countdown appears.
4. Press **Confirm Purchase** → stock total drops, tracker refreshes.
5. Try quantity 9999 → `INSUFFICIENT_STOCK` banner (stok tidak mencukupi).
6. Let a reservation expire (or stop reading) → countdown hits 0, UI shows the expired state,
   stock returns automatically via the sweeper.
