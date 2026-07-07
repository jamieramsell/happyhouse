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

## Status

📐 **Design phase** — no application code yet. Implementation begins with [Phase 0](docs/ROADMAP.md#phase-0--skeleton).
