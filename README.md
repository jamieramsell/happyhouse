# Happyhouse 🏠

> A shared-household app for uni houses: a shared-cost **kitty**, a **fair chore rota**, and **spending & chore statistics**, installable on iOS and Android as a PWA.

[![Version](https://img.shields.io/badge/version-0.1.0-blue.svg)](CHANGELOG.md)
[![Versioning](https://img.shields.io/badge/versioning-semantic-brightgreen.svg)](https://semver.org)
[![Code Style](https://img.shields.io/badge/code%20style-Google%20%C2%B7%20Microsoft-blue.svg)](CLAUDE.md#code-style)
[![Go](https://img.shields.io/badge/go-1.24+-00ADD8.svg)](https://go.dev)
[![.NET](https://img.shields.io/badge/.NET-10-512BD4.svg)](https://dotnet.microsoft.com)
[![TypeScript](https://img.shields.io/badge/typescript-react%20PWA-3178C6.svg)](https://www.typescriptlang.org)

---

## Overview

Happyhouse solves the everyday friction of a shared house: who paid for what, who owes whom, and whose turn it is to do the bins. It brings shared costs, settlements, and a chore rota onto one installable app, with statistics for both.

It is also a deliberate learning project: polyglot **microservices** in **Go** and **C# / ASP.NET Core**, the repository pattern with dependency injection throughout, Docker Compose deployment, and an event-driven refactor with **NATS** along the way. The roadmap is phased so a working app exists early, even if the stretch milestones slip.

---

## Features

### Kitty (shared costs)
- **Expenses** with four split methods — `EQUAL`, `PERCENTAGE`, `EXACT`, `SHARES` — and largest-remainder rounding so shares always sum exactly to the total
- **Money movements** — directed transfers that shift balances but are excluded from spend statistics by construction
- **Live balances** — net position per member and a pairwise who-owes-whom view, derived from the ledger
- **Settlements** — review periods close, then a transfer-minimising algorithm (greedy `≤ n−1` transfers; exact bitmask DP as a stretch) suggests who pays whom; mark-paid auto-creates the movement
- **Recurring expenses** — rent, internet, subscriptions auto-materialised from a template

### Chores
- **Fair rota generation** — occurrences expanded from a recurrence model, assigned by a fairness-counter that balances per-chore and overall load, with optional effort weights
- **Swaps & reassignment** — any two members can swap; regeneration never clobbers completed or overridden occurrences
- **Completion tracking** — assignees mark done; a scheduled job marks missed chores after a grace period

### Statistics
- **Kitty** — total spend, by category, by payer, per-member share, trend over time (money movements excluded)
- **Chores** — completed/missed rates per member and per chore, streaks, and a "most reliable housemate" leaderboard
- Standard period picker everywhere: week / month / year / all-time / custom range

### Platform
- **Installable PWA** — one app on iOS and Android, offline app-shell, no app stores
- **Multi-household** — belong to more than one house; invite via rotatable share links; `admin` / `member` roles
- **Activity feed** — "Alexa added Toaster £28", "James completed Bins", "Rota regenerated"

Deferred (v2+): web push notifications, receipt photos, shopping list, email verification/reset, approval-based swaps.

---

## Architecture

Polyglot microservices behind a Traefik gateway, one Postgres instance with a database per service:

| Component | Tech |
|---|---|
| Frontend | React + TypeScript PWA (Vite) |
| `auth` service | Go |
| `household` service | Go *(Phase 2)* |
| `kitty` service | C# / ASP.NET Core |
| `chores` service | C# / ASP.NET Core *(Phase 6)* |
| Data | Postgres (one instance, one DB per service) |
| Edge | Traefik (TLS + path routing) |
| Events | NATS event bus *(Phase 5)* |
| Deployment | Docker Compose on a VPS |

The client talks REST/JSON to each service through Traefik path routing (`/api/v1/auth/*`, `/api/v1/kitty/*`, …). Services verify JWTs signed by `auth` via a JWKS endpoint. Statistics live inside the owning service. See [**`docs/DESIGN.md`**](docs/DESIGN.md) for the full design and rationale.

---

## Project Structure

```
happyhouse/
│
├── frontend/                       React + TypeScript PWA (Vite, vite-plugin-pwa)
│   ├── src/                        App.tsx, main.tsx, index.css, assets/
│   ├── public/                     favicon, PWA icons
│   ├── index.html
│   ├── vite.config.ts
│   ├── .oxlintrc.json
│   └── package.json
│
├── services/
│   ├── auth/                       Go
│   │   ├── cmd/server/main.go      entrypoint (graceful shutdown, manual DI)
│   │   ├── internal/api/
│   │   │   ├── router.go           /api/v1/auth routes + health endpoints
│   │   │   └── router_test.go
│   │   ├── go.mod                  module github.com/jamieramsell/happyhouse/services/auth
│   │   └── Dockerfile
│   └── kitty/                      C# / ASP.NET Core
│       ├── src/Kitty.Api/
│       │   ├── Program.cs          /api/v1/kitty routes + health endpoints
│       │   └── Kitty.Api.csproj
│       ├── tests/Kitty.Api.Tests/  xUnit + WebApplicationFactory<Program>
│       └── Dockerfile
│
├── deploy/
│   ├── docker-compose.dev.yml      Traefik + Postgres + services
│   ├── postgres/init.sql           per-service databases + roles
│   └── .env.example
│
├── docs/
│   ├── DESIGN.md                   architecture, services, domain model, algorithms
│   └── ROADMAP.md                  phased build plan with checklists
│
├── .github/
│   └── workflows/                  path-filtered CI: auth.yml, kitty.yml, frontend.yml
│
├── CHANGELOG.md
├── .gitignore
└── README.md
```

Planned (see `DESIGN.md`): `services/household/` and `services/chores/`, plus `libs/go-common` and `libs/dotnet-common` for shared JWT middleware, the recurrence model, error types, and event envelopes.

---

## Getting Started

### Prerequisites

- Docker + Docker Compose (for the full stack)
- Go 1.24+ · .NET 10 SDK · Node 22+ (for working on a single piece)

### Full stack

From the repo root:

```bash
cp deploy/.env.example deploy/.env
docker compose -f deploy/docker-compose.dev.yml --env-file deploy/.env up --build
```

Then open <http://localhost>; the app shell shows the `auth` and `kitty` services as **ok** once they're reachable through Traefik. The Traefik dashboard is at <http://localhost:8081>.

### Working on a single piece

| Piece | Command |
|---|---|
| Frontend | `cd frontend && npm install && npm run dev` |
| auth (Go) | `cd services/auth && go run ./cmd/server` → <http://localhost:8080/api/v1/auth/healthz> |
| kitty (C#) | `cd services/kitty && dotnet run --project src/Kitty.Api` → `/api/v1/kitty/healthz` |

Tests and checks:

```bash
cd services/auth  && go test ./... && go vet ./...
cd services/kitty && dotnet test tests/Kitty.Api.Tests/Kitty.Api.Tests.csproj
cd frontend       && npm run lint && npm run build
```

---

## Code Style

Each stack follows the canonical style guide for its language:

- Go — [Google Go Style Guide](https://google.github.io/styleguide/go/)
- C# — [Microsoft C# coding conventions](https://learn.microsoft.com/en-us/dotnet/csharp/fundamentals/coding-style/coding-conventions)
- TypeScript — [Google TypeScript Style Guide](https://google.github.io/styleguide/tsguide.html)

Style is enforced by each ecosystem's standard tooling (`gofmt`/`go vet`, `dotnet format`, `oxlint`). Per-service specifics (layering, interface naming, DI) are documented in [`CLAUDE.md`](CLAUDE.md).

---

## Versioning

This project uses [Semantic Versioning](https://semver.org/) (`MAJOR.MINOR.PATCH`):

- `MAJOR` — breaking API changes
- `MINOR` — new backwards-compatible features
- `PATCH` — backwards-compatible bug fixes

**Breaking changes** are flagged with `!` (e.g. `feat!:`, `fix!:`) and must have a corresponding issue opened first. Releases track the phased build plan in [ROADMAP.md](docs/ROADMAP.md). See [CHANGELOG.md](CHANGELOG.md) for the full release history.

---

## Contributing

This is a personal learning project, but the workflow is kept deliberately real. Please open an issue before submitting a pull request so the change can be discussed first.

Work flows from short-lived development branches up through per-milestone stable branches into `main`:

```
<milestone>/<label>/<name>  ──PR──▶  stable-<milestone>  ──PR──▶  main
```

Milestones map to the roadmap phases (`m0`–`m7`).

### Branching

- **Development branches** — `<milestone>/<label>/<name>`, where `<label>` is a Conventional Commit type (`feat`, `fix`, `refactor`, `docs`, `chore`, …).
  - `m1/feat/argon2id-hashing` — password hashing in milestone 1 (Phase 1, Auth)
  - `m3/refactor/split-rounding` — reworking largest-remainder rounding in milestone 3 (Phase 3, Kitty)
- **Stable branches** — `stable-<milestone>` (e.g. `stable-m1`, `stable-m3`). Development branches merge here via pull request once ready.
- **`main`** — a completed milestone is merged from its stable branch into `main` via a further pull request.

### Pull requests

1. Open an issue describing the change before starting work.
2. Branch from the relevant stable branch using the naming convention above.
3. Open a pull request targeting the appropriate branch (development → `stable-<milestone>`; completed milestone → `main`).
4. Every pull request must pass the automated checks (build, test, lint) before it can be merged — CI is path-filtered per service, so only the affected service rebuilds.

### Commits & style

- **[Conventional Commits](https://www.conventionalcommits.org/)** (`feat:`, `fix:`, `refactor:`, `chore:`, `docs:`, …).
- **Breaking changes** are flagged with `!` after the type/scope, before the colon — e.g. `feat!:`, `fix!:`. A breaking change must have a corresponding issue opened first.
- Each language follows its canonical style guide — [Google](https://google.github.io/styleguide/) for Go and TypeScript, [Microsoft's conventions](https://learn.microsoft.com/en-us/dotnet/csharp/fundamentals/coding-style/coding-conventions) for C# (see [Code Style](#code-style)).
- Releases follow **[Semantic Versioning](https://semver.org/)**.

---

## Status

🏗️ **Phase 0 complete** (`0.1.0`) — monorepo skeleton, Dockerised Traefik + Postgres stack, hello-world `auth` (Go) and `kitty` (C#) services, installable PWA shell, and path-filtered CI. Next up: [Phase 1 — Auth](docs/ROADMAP.md#phase-1--auth).
