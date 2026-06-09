# Copilot instructions for ebzer-api

Purpose: quick reference for Copilot sessions to understand, build, run and modify this repository.

---

## Quick commands

- Start dev server (creates DB and runs migrations):
  - ./server.sh start
  - or: go run cmd/server/main.go
- Stop / status / logs / restart / reset DB / open DB shell:
  - ./server.sh {stop|status|logs|restart|reset-db|db}
- Run integration tests (end-to-end):
  - ./test-api.sh
  - or: ./server.sh test (requires server running)
- Single endpoint checks (useful for targeted verification):
  - Health: curl -s http://localhost:3000/ping
  - DB ping: curl -s http://localhost:3000/dbping
  - Create one order (example):
    curl -s -X POST http://localhost:3000/api/orders -H "Content-Type: application/json" -d '{"description":"Test","amount_charged":"100.00","status":"confirmed","delivery_type":"pickup","client_name":"Test"}'

Notes: There are no Go unit tests (_test.go) in the repository; automated checks are performed via the shell integration tests (test-api.sh). No repository-wide linter config found.

---

## Environment & runtime

- Go version used: go 1.25.5 (go.mod)
- Module name: creaciones-api
- Primary runtime: Fiber v2 (github.com/gofiber/fiber/v2)
- DB: SQLite3 (mattn/go-sqlite3)
- DB path env var: SQLITE_DB_PATH (default: ./data/ebzer.db)
- Connection defaults: WAL journal, foreign_keys enabled, max open conns = 1

---

## High-level architecture (big picture)

- Clean Architecture style. Top-level mapping:
  - cmd/server/            — application entrypoint; wires dependencies and HTTP server
  - internal/db/           — DB connection, migrations, scanner types
    - internal/db/migrations — SQL migration files (applied automatically on start)
  - internal/<domain>/     — domain folders (orders, incomes, etc.)
    - models.go            — DB models / domain structs
    - dto.go               — request/response DTOs
    - repository.go        — persistence layer (DB access)
    - service.go           — business logic
    - handler.go           — HTTP handlers / route registration

- Wiring: cmd/server creates repositories -> services -> handlers and registers routes under /api.
- Migrations: Run automatically at startup via db.RunMigrations(..., "internal/db/migrations"). The system tracks applied migrations (schema_migrations).

---

## Key repository conventions and patterns

- Package layout: follow internal/<domain> with the same file set: models.go, dto.go, repository.go, service.go, handler.go. New domains should follow the same shape.
- Naming:
  - NewRepository, NewService, NewHandler constructors are expected and used to wire dependencies.
  - Handlers expose RegisterRoutes(routerGroup) to attach endpoints to a fiber.Group.
- DTOs are separated from models (dto.go) and used at HTTP boundaries; repositories work with models.
- Database:
  - migrations live in internal/db/migrations and are applied on startup — do not rely on external migration runners.
  - SQLITE_DB_PATH overrides DB file location for tests or CI.
  - SQLite-specific settings assume single-connection usage; avoid opening multiple pooled connections in application code.
- Scripts:
  - server.sh is the canonical management script (start/stop/test/logs/reset-db/db). Use it in CI/emulation for consistent behavior.
  - test-api.sh contains the integration test steps and can be run interactively; use it as the canonical E2E test suite.

---

## Files to inspect for changes affecting behavior

- cmd/server/main.go — wiring and startup flow (migrations + server settings)
- internal/db/* — connection, migration runner, scanner helpers
- internal/*/repository.go — SQL queries and schema assumptions

---

## AI assistant / other tools

- No existing AI assistant config files were found (CLAUDE.md, AGENTS.md, .cursorrules, .windsurfrules, CONVENTIONS.md, etc.).

---

If any of the above needs to be expanded (example: add Go unit test instructions, CI commands, or a linter config), say which area and Copilot will update this file.
