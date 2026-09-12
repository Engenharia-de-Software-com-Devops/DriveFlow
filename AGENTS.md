# DriveFlow Agent Guide

## Commands

- Install both workspaces: `make instalar`.
- Run the standard suite: `make testar`.
- Run Postgres integration tests: `make testar-integracao`; it starts only the `db` Compose service and sets `DATABASE_URL` for the Go tests.
- Check backend formatting and static analysis: `make verificar`. `gofmt -l .` only reports unformatted files, so run `cd backend && gofmt -w <files>` before rechecking when it reports output.
- Run a focused Go test from `backend/`: `go test ./internal/locacao -run TestName -count=1`.
- Run frontend tests from `frontend/`: `npm test -- --run`; focus a file with `npm test -- src/App.test.tsx`.
- Start the full production-like stack: `docker compose up --build`; start hot-reload containers: `docker compose -f docker-compose.dev.yml up --build`.

## Architecture

- `backend/main.go` selects the repository and starts the API. Without `DATABASE_URL`, it uses in-memory storage: local data does not persist and Postgres integration tests skip.
- Keep business rules in `backend/internal/locacao/`, HTTP request/response translation in `backend/internal/api/`, and memory/Postgres repositories plus migrations in `backend/internal/armazenamento/`.
- The React SPA is in `frontend/src/`; put API-client changes in `frontend/src/api.js` and UI changes in `frontend/src/componentes/`.
- Production Compose builds with `REACT_APP_API_URL=""`; `frontend/docker/nginx.conf` proxies relative `/api` and `/health` requests to the API. Do not replace these relative calls with a container hostname.

## Invariants

- Represent all money as integer centavos across API, domain, persistence, and UI conversion; never introduce floating-point currency.
- Every vehicle or rental lookup must be scoped by `empresaID`; this enforces tenant isolation.
- Reservation conflicts are enforced both by domain logic and a PostgreSQL exclusion constraint. Preserve both layers.
- Migrations are embedded in the Go binary and run automatically at API startup. Add a new `NNNN_descricao.up.sql` and matching `.down.sql`; never edit an already-applied migration. Human-review migrations before production.

## Workflow

- After completing any change, run `make verificar` and `make testar`.
- Before a PR, run `make verificar` and `make testar`; changes to behavior require tests.
- Work branches target `dev`. Do not push directly to `main`, `prod`, or `homo`; use a PR reviewed by another contributor.

## Collaboration

- Ask the developer when requirements or intended behavior are unclear; do not guess or assume.
- Follow TDD: write tests for the expected behavior and edge cases before implementation.
- Write code identifiers, including functions, variables, types, classes, new filenames, and comments, in English. Preserve existing domain names and commands unless a change is required.
- Do not run Git operations, including commits, automatically. The developer performs them unless they explicitly request otherwise.
