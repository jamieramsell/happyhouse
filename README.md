# Happyhouse 🏠

A shared-household app for uni houses: track shared costs in a **kitty** (flexible splits, transfer-minimising settlements), run a **fair chore rota**, and see **spending & chore statistics** — installable on iOS and Android as a PWA.

Also a deliberate learning project: polyglot **microservices** in **Go** and **C#/ASP.NET Core**, repository pattern + dependency injection throughout, Docker Compose deployment, and an event-driven refactor with NATS along the way.

## Documentation

- [**Design document**](docs/DESIGN.md) — architecture, services, domain model, algorithms
- [**Roadmap**](docs/ROADMAP.md) — phased build plan with checklists

## Architecture at a glance

| Component | Tech |
|---|---|
| Frontend | React + TypeScript PWA (Vite) |
| auth service | Go |
| household service | Go |
| kitty service | C# / ASP.NET Core |
| chores service | C# / ASP.NET Core |
| Data | Postgres (one DB per service) |
| Edge | Traefik (TLS + routing) |
| Deployment | Docker Compose on a VPS |

## Getting started (local dev)

Requires Docker + Docker Compose.

```bash
cp deploy/.env.example deploy/.env
docker compose -f deploy/docker-compose.dev.yml --env-file deploy/.env up --build
```

Then open <http://localhost> — the app shell shows the `auth` and `kitty` services
as **ok** once they're reachable through Traefik. The Traefik dashboard is at
<http://localhost:8081>.

Working on a single piece without the full stack:

| Piece | Command |
|---|---|
| Frontend | `cd frontend && npm install && npm run dev` |
| auth (Go) | `cd services/auth && go run ./cmd/server` → <http://localhost:8080/api/v1/auth/healthz> |
| kitty (C#) | `cd services/kitty && dotnet run --project src/Kitty.Api` → `/api/v1/kitty/healthz` |

## Status

🏗️ **Phase 0 complete** — monorepo skeleton, Dockerised Traefik + Postgres stack,
hello-world `auth` (Go) and `kitty` (C#) services, installable PWA shell, and
path-filtered CI. Next up: [Phase 1 — Auth](docs/ROADMAP.md#phase-1--auth).
