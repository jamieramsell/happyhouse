# CLAUDE.md

## Project Overview

A shared-household app for uni houses:

- 🏦 **Kitty** — shared costs with flexible splits and transfer-minimising settlements
- 🧹 **Chores** — a fair, effort-weighted rota
- 👛 **Wallet** — your own personal accounts, income, and budgets, with drift-free reconciliation against your real bank balance
- 📊 **Statistics** for all of the above

Installable on iOS and Android as a PWA.

It is equally a **learning project**: first hands-on experience with **microservices**, and learning **Go and C#** (coming from a Java background). The design deliberately pushes the boat out where it teaches something and stays pragmatic where it doesn't. It is in early development: the Phase 0 skeleton (Compose stack, hello-world services, PWA shell, CI) is complete, but the domains are mostly still to be built.

## Architecture

Polyglot microservices behind a Traefik gateway, one Postgres instance with a database per service, an event bus arriving mid-project:

| Service | Language / Stack | Responsibility |
|---|---|---|
| `auth` | Go | Users, credentials, JWT issuance, refresh tokens, profiles |
| `household` | Go | Households, membership, roles, invites, activity feed |
| `kitty` | C# / ASP.NET Core | Expenses, splits, money movements, balances, settlements, spend stats |
| `chores` | C# / ASP.NET Core | Chores, occurrences, rota, swaps, completions, chore stats |
| `wallet` | C# / ASP.NET Core | Income and expenses, money movements, actual vs expected balance, money stats |
| `frontend` | React + TypeScript PWA (Vite) | UI, service worker, offline shell |

Supporting pieces: **Traefik** (TLS + path routing, `/api/v1/...`), **Postgres** (one container, one DB + credentials per service), **NATS** (event bus, Phase 5). Statistics live **inside the owning service** — no separate reporting service. See [`docs/DESIGN.md`](docs/DESIGN.md) for the full rationale and [`docs/ROADMAP.md`](docs/ROADMAP.md) for the phased build plan.

## Repo Structure

```
happyhouse/
├── frontend/                       React + TypeScript PWA (Vite, vite-plugin-pwa)
│   ├── src/                        App.tsx, main.tsx, index.css, assets/
│   ├── public/                     favicon, PWA icons
│   ├── index.html
│   ├── vite.config.ts
│   ├── .oxlintrc.json              lint config (oxlint)
│   └── package.json
├── services/
│   ├── auth/                       Go — module github.com/jamieramsell/happyhouse/services/auth
│   │   ├── cmd/server/main.go      entrypoint (manual DI wiring lives here)
│   │   ├── internal/api/           HTTP layer (router.go, router_test.go)
│   │   └── go.mod
│   └── kitty/                      C# / ASP.NET Core
│       ├── src/Kitty.Api/          Program.cs, Kitty.Api.csproj
│       └── tests/Kitty.Api.Tests/  xUnit + WebApplicationFactory
├── deploy/
│   ├── docker-compose.dev.yml      Traefik + Postgres + services
│   ├── postgres/init.sql           per-service DBs + roles
│   └── .env.example
├── docs/
│   ├── DESIGN.md                   architecture, domain model, algorithms
│   └── ROADMAP.md                  phased build plan with checklists
├── .github/workflows/              per-service, path-filtered CI (auth.yml, kitty.yml, frontend.yml)
├── CHANGELOG.md
├── .gitignore
└── README.md
```

Planned but not yet present (see `DESIGN.md`): `services/household/` and `services/chores/` (same shape as their siblings), and `libs/go-common` + `libs/dotnet-common` for shared JWT middleware, the recurrence model, error types, and event envelopes.

## Running the Services

Full stack (requires Docker + Docker Compose), from the repo root:

```bash
cp deploy/.env.example deploy/.env
docker compose -f deploy/docker-compose.dev.yml --env-file deploy/.env up --build
```

Then open <http://localhost>; the app shell shows `auth` and `kitty` as **ok** once reachable through Traefik. The Traefik dashboard is at <http://localhost:8081>.

Working on a single piece without the full stack:

### auth (Go)
```bash
cd services/auth
go run ./cmd/server     # -> http://localhost:8080/api/v1/auth/healthz
go test ./...           # run tests
go vet ./...            # static checks (matches CI)
go build ./...
```

### kitty (C# / .NET)
```bash
cd services/kitty
dotnet run --project src/Kitty.Api                      # -> /api/v1/kitty/healthz
dotnet test tests/Kitty.Api.Tests/Kitty.Api.Tests.csproj
```

### frontend (React / Vite)
```bash
cd frontend
npm install
npm run dev             # Vite dev server
npm run lint            # oxlint
npm run build           # tsc -b && vite build
```

## Conventions

- **Commits**: Conventional Commits (`feat:`, `fix:`, `chore:`, `refactor:`, `docs:`, etc.). Breaking changes are marked with `<type>!:` (e.g. `feat!:`, `fix!:`) and should have a corresponding issue opened first.
- **Versioning**: Semantic Versioning; see [`CHANGELOG.md`](CHANGELOG.md). Releases track the phased roadmap.
- **Branching**: milestone-based feature branches, `<milestone>/<label>/<name>`, where `<milestone>` maps to a roadmap phase (`m0`–`m7`) and `<label>` is a Conventional Commit type (e.g. `m1/feat/argon2id-hashing`, `m3/refactor/split-rounding`). Development branches merge into `stable-<milestone>`; a completed milestone merges into `main`. Open an issue before opening a PR.
- **Code style**: Go and the frontend follow the [Google Style Guide](https://google.github.io/styleguide/); C# follows [Microsoft's C# coding conventions](https://learn.microsoft.com/en-us/dotnet/csharp/fundamentals/coding-style/coding-conventions). See [Code Style](#code-style) below.
- **API routing**: every service serves under `/api/v1/<service>/*` so Traefik path-routes without rewriting (`/api/v1/auth/*` → auth, `/api/v1/kitty/*` → kitty). Health endpoints: `/api/v1/<service>/healthz` and `/readyz`.
- **Go module root**: `github.com/jamieramsell/happyhouse/services/<service>`.
- **Money** is stored as integer pence (`amount_minor BIGINT`) with a `currency` column, never floats.
- **Derived, never stored as truth**: balances and stats are computed from the ledger on read, not persisted.

## Code Style

Each stack follows the canonical style guide for its language, enforced by that ecosystem's standard tooling:

- Go — [Google Go Style Guide](https://google.github.io/styleguide/go/) (`gofmt`, `go vet`)
- C# — [Microsoft C# coding conventions](https://learn.microsoft.com/en-us/dotnet/csharp/fundamentals/coding-style/coding-conventions) (`.editorconfig` + `dotnet format`, Roslyn analyzers)
- TypeScript — [Google TypeScript Style Guide](https://google.github.io/styleguide/tsguide.html) (`oxlint`)

The language-specific conventions below are how those guides apply to each service.

### Go (`auth`, `household`)

- **Style**: conform to the [Google Go Style Guide](https://google.github.io/styleguide/go/); code must be `gofmt` / `go vet` clean (CI runs `go vet ./...`, `go build ./...`, `go test ./...`).
- **Repository pattern**: repository **interfaces live in the domain package** (e.g. `UserRepository`, `HouseholdRepository`) — **no `I` prefix**, idiomatic Go. Postgres implementations go in an `infra/postgres` package using `pgx`.
- **Dependency injection**: **manual constructor injection wired in `main.go`** — no framework. Handlers depend on service structs; service structs depend on repo interfaces; tests inject in-memory fakes.
- **HTTP layer**: lives under `internal/api`; routes registered under the `/api/v1/<service>` prefix; JSON written via a small `writeJSON` helper.
- **Packages**: one package per concern; `internal/` for non-exported service code.
- **Tests**: standard library `testing`, co-located `_test.go` files in the same package (e.g. `router_test.go` beside `router.go`).

### C# (`kitty`, `chores`, `wallet`)

- **Style**: conform to [Microsoft's C# coding conventions](https://learn.microsoft.com/en-us/dotnet/csharp/fundamentals/coding-style/coding-conventions); keep `dotnet format` clean and let the built-in Roslyn analyzers/`.editorconfig` enforce it.
- **Interfaces**: prefix with `I` — idiomatic C# — e.g. `IExpenseRepository`, `IEventRepository`.
- **Classes / methods**: PascalCase; local variables camelCase.
- **Layered layout**: `Api` / `Application` / `Domain` / `Infrastructure` projects. Repository interfaces live in `Domain`; EF Core implementations in `Infrastructure`; wired through the built-in `Microsoft.Extensions.DependencyInjection` container (`services.AddScoped<IExpenseRepository, EfExpenseRepository>()`).
- **MVC layering**: Controller/endpoint → Service → Repository; keep business logic out of the HTTP layer.
- **Nullable** reference types and **implicit usings** are enabled; keep them on.
- **Tests**: xUnit; integration tests drive the app via `WebApplicationFactory<Program>` (one test class per unit under test, e.g. `HealthEndpointsTests`).

### TypeScript / React (`frontend`)

- **Style**: conform to the [Google TypeScript Style Guide](https://google.github.io/styleguide/tsguide.html), enforced as far as practical by `oxlint` (`react`, `typescript`, `oxc` plugins); `react/rules-of-hooks` is an error. Keep `npm run lint` and `npm run build` clean — both run in CI.
- **Files**: components PascalCase (`App.tsx`), config/util camelCase.
- **Types**: strict TypeScript (`tsc -b` is part of the build). API client types are intended to be generated from each service's OpenAPI spec (`openapi-typescript`) so frontend and backend can't silently drift.

## Current State

**Phase 0 is complete** (`0.1.0`): monorepo skeleton, Dockerised Traefik + Postgres stack, hello-world `auth` (Go) and `kitty` (C#) services behind health endpoints, an installable Vite + React PWA shell, and per-service path-filtered CI.

The established patterns to build on:
- Repository pattern with per-language DI (manual in Go, built-in container in C#), interfaces owned by the domain layer.
- Traefik path routing per service under `/api/v1/<service>`.
- One Postgres container, a separate database + credentials per service (`postgres/init.sql`).
- Path-filtered GitHub Actions so a change to one service doesn't rebuild the others.

Not yet built: the auth/household/kitty/chores domains proper, persistence and migrations (`golang-migrate` for Go, EF Core for C#), inter-service communication, the NATS event bus (Phase 5), and the shared libs. `docs/ROADMAP.md` tracks the phased plan; next up is **Phase 1 — Auth**.
