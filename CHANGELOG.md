# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

_Next up: Phase 1 — Auth (register/login/refresh/JWKS)._

## [0.1.0] - 2026-07-08

Phase 0 — project skeleton.

### Added
- Monorepo layout (`frontend/`, `services/`, `deploy/`, `docs/`).
- Dockerised local stack via `deploy/docker-compose.dev.yml`: Traefik gateway
  (path routing on `/api/v1/...`) + Postgres with one database and credentials
  per service (`deploy/postgres/init.sql`).
- Hello-world `auth` service (Go) behind `/api/v1/auth/healthz` and `/readyz`,
  with graceful shutdown.
- Hello-world `kitty` service (C# / ASP.NET Core) behind `/api/v1/kitty/healthz`
  and `/readyz`.
- Installable Vite + React + TypeScript PWA shell, served through Traefik.
- Path-filtered GitHub Actions CI per service (`auth`, `kitty`, `frontend`).
- Design document (`docs/DESIGN.md`) and phased roadmap (`docs/ROADMAP.md`).

[Unreleased]: https://github.com/jamieramsell/happyhouse/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/jamieramsell/happyhouse/releases/tag/v0.1.0
