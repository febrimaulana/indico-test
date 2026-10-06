# AGENTS.md — Repository Rules

> This file is read by AI coding agents (and humans) before working in this repository.
> `AGENT.md` is a symlink to this file; this is the single source of truth.

## Project Overview

**High-Throughput Inventory Reservation System & Mini Dashboard** (INDICO take-home).

- **Backend (80%)**: Go + Gin + GORM + PostgreSQL. Real DB via Docker Compose — never SQLite/in-memory/mocks.
- **Frontend (20%)**: React + Vite + TypeScript + Tailwind CSS, strict. Axios + TanStack Query for all server state.
- **One command to run**: `docker compose up` (zero manual setup).

Read `README.md` for setup and `ARCHITECTURE.md` for design decisions.

## Golden Rules (NON-NEGOTIABLE)

1. **Strict stack.** Do not add/replace frameworks. Backend: Gin + GORM only for HTTP/ORM. Frontend: React + Vite only (no Next.js/Nuxt/Vue/etc).
2. **API JSON envelope is frozen:**
   - Success: `{"status":"success", ...}` (reserve/confirm) or the flat payload (stock) — exactly as SPEC in `docs/INDICO-Fullstack-Test.md`.
   - Error: `{"status":"error","error":{"code":"...","message":"...","details":[...]}}`.
3. **Error messages are in Bahasa Indonesia; error `code` is English.** Frozen codes:
   `INVALID_REQUEST`, `ITEM_NOT_FOUND`, `INSUFFICIENT_STOCK`, `RESERVATION_NOT_FOUND`, `RESERVATION_EXPIRED`, `RESERVATION_ALREADY_CONFIRMED`, `INTERNAL_ERROR`.
4. **Three-layer validation:** binding tags (format) → service (business rules) → PostgreSQL CHECK constraints (last line of defense). Never rely on one layer alone.
5. **Cross-module access only via service interfaces.** Never import another module's repository/model package. Never JOIN another module's tables.
6. **`internal/platform/` must never import `internal/modules/...`** (one-way dependency).
7. **Never edit an applied migration.** Add a new numbered file: `0000NN_<domain>_<description>.up.sql` + `.down.sql`.
8. **Frontend:** every HTTP call via `src/lib/axios.ts`; every piece of server state via TanStack Query; every `useEffect` must clean up.
9. **Quality gates before finishing any task:**
   - `cd backend && go test -race -v ./...` (green)
   - `cd backend && make coverage` (100% of `./internal/...`, unit + integration)
   - `cd backend && go vet ./...`
   - `cd frontend && npx tsc --noEmit && npm run build`
10. **Commit incrementally** (conventional commits, per phase). Never one massive commit.

## Repository Map

```
indico/
├── README.md               # human + AI friendly setup guide
├── ARCHITECTURE.md         # required deliverable (3 prompts, max 2 pages)
├── AGENTS.md               # this file
├── docker-compose.yaml     # postgres + backend + frontend (one command)
├── docs/
│   └── INDICO-Fullstack-Test.md   # the assignment converted to markdown
├── backend/                # Go + Gin + GORM  (see backend/AGENTS.md)
└── frontend/               # React + Vite + TS (see frontend/AGENTS.md)
```

## Adding a New Domain Module (backend)

1. Create `backend/migrations/0000NN_<domain>_*.up.sql` + `.down.sql`.
2. Create `backend/internal/modules/<domain>/` following the file convention
   (model.go, dto.go, errors.go, repository.go, repository_gorm.go, service.go, handler.go, routes.go, module.go + tests).
3. Register it in `backend/internal/app/app.go` (wiring) and `backend/internal/app/router.go` (one line: `xxx.RegisterRoutes(v1)`).
4. Create `frontend/src/features/<domain>/` if it has UI.
5. Add unit tests (success + minimum 2 negative cases), swagger annotations, and run `make swag`.
