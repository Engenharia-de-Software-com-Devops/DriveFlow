# DriveFlow Agent Guide

## Commands

- Install both workspaces: `make instalar`.
- Run the standard suite: `make testar`.
- Run Postgres integration tests: `make testar-integracao`; it starts only the `db` Compose service and sets `DATABASE_URL` for the Go tests.
- Check backend formatting and static analysis: `make verificar`. `gofmt -l .` only reports unformatted files, so run `cd backend && gofmt -w <files>` before rechecking when it reports output.
- Run a focused Go test from `backend/`: `go test ./tests/unidade/usecases -run TestName -count=1`.
- Run a focused frontend test from `frontend/`: `CI=true npm test -- --watchAll=false --testPathPattern=src/api.test.js`.
- Start the full production-like stack: `docker compose up --build`; start hot-reload containers: `docker compose -f docker-compose.dev.yml up --build`.

## Architecture

- `backend/cmd/app/main.go` wires the layers and starts the API. Without `DATABASE_URL`, it uses in-memory storage: local data does not persist.
- Keep the domain model, domain errors, and pricing rules in `backend/internal/entities/`; business rules and the repository port in `backend/internal/usecases/`; HTTP request/response translation in `backend/internal/delivery/http/`; and the memory/Postgres repositories plus migrations in `backend/internal/repository/`.
- Dependencies point inward, and `go list -deps ./internal/<pkg>` is how you check it: `entities` imports nothing from the project; `usecases` imports `entities` and the `pkg/id` helper; `repository` imports only `entities`, satisfying the `usecases.Repository` port structurally, since Go interfaces are implicit; `delivery/http` imports `usecases` and `entities`. Never import `delivery` or `repository` from `usecases`.
- The React SPA is in `frontend/src/`; put API-client changes in `frontend/src/api.js` and UI changes in `frontend/src/componentes/`.
- Production Compose builds with `VITE_API_URL=""`; `frontend/docker/nginx.conf` proxies relative `/api` and `/health` requests to the API. Do not replace these relative calls with a container hostname.

## Tests

- Backend tests live in `backend/tests/`, never next to the code they exercise. The tree mirrors the layers and splits on the only thing that changes how you run them: whether a database has to be up.
- `tests/unidade/{entities,usecases,repository,delivery}/` runs without a database and without Docker; it is what `make testar` runs.
- `tests/integracao/` requires PostgreSQL. Every file there starts with `//go:build integracao` on the first line, before `package`. Without the tag the file is excluded from the build, which is what keeps `go test ./...` free of database dependencies.
- Shared setup that both levels need goes in `backend/tests/apoio/` as a regular (non-`_test.go`) file: a `_test.go` file is invisible to other packages.
- A test in a separate directory only sees the exported API of the package under test. If a test genuinely needs an unexported identifier, it goes back beside the code as a `_test.go` file.
- See `backend/tests/README.md` for which folder a new test belongs in.

## Invariants

- Represent all money as integer centavos across API, domain, persistence, and UI conversion; never introduce floating-point currency.
- Every vehicle or rental lookup must be scoped by `CompanyID` (`empresa_id` on the wire); this enforces tenant isolation.
- Reservation conflicts are enforced both by domain logic and a PostgreSQL exclusion constraint. Preserve both layers.
- Migrations are embedded in the Go binary and run automatically at API startup. Add a new `NNNN_descricao.up.sql` and matching `.down.sql`; never edit an already-applied migration. Human-review migrations before production.

## Workflow

- After completing any change, run `make verificar` and `make testar`.
- Before a PR, run `make verificar` and `make testar`; changes to behavior require tests. Also run `make testar-integracao` when the change touches SQL, a migration, or the repository.
- Work branches target `dev`. Do not push directly to `main`, `prod`, or `homo`; use a PR reviewed by another contributor.

## Collaboration

- Ask the developer when requirements or intended behavior are unclear; do not guess or assume.
- Follow TDD: write tests for the expected behavior and edge cases before implementation.
- Write code identifiers, including functions, variables, types, classes, new filenames, and comments, in English. Preserve existing domain names and commands unless a change is required.
- Do not run Git operations, including commits, automatically. The developer performs them unless they explicitly request otherwise.
