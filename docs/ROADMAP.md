# Happyhouse — Build Roadmap

Each phase ends with something usable. Kitty lands before chores because it's the harder domain and the housemates' most urgent need; NATS lands mid-project so the "why event-driven" lesson is earned, not asserted. Wallet has no dependency on Households — it only needs Auth — so it's an **independent track**: build it in whatever order or interleaving suits, including alongside Phase 3.

| Phase | Deliverable | Learning focus |
|---|---|---|
| **0** | Monorepo skeleton, Compose + Traefik, hello-world Go & C# services, PWA shell, CI | Toolchains, Docker, routing |
| **1** | Auth service (register/login/refresh/JWKS) + frontend auth flow | Go, security, JWT |
| **2** | Wallet: personal accounts, income/expenses, transfers, inventory reconciliation, stats *(independent track — depends only on Auth)* | User-scoped C# service, bounded recurrence, reconciliation |
| **3** | Household service (CRUD, invites, roles) + household UI | Go service patterns, repo/DI in Go |
| **4** | Kitty core: expenses, 4 split methods, movements, live balances + UI | C#/ASP.NET, EF Core, money maths |
| **5** | Settlements: review schedules, greedy algorithm, mark-paid flow; kitty stats | Recurrence model, algorithms |
| **6** | NATS event bus: membership read-models replace sync calls; event-driven activity feed | Event-driven microservices |
| **7** | Chores: recurrence, rota generation, swaps, completions, stats + UI | Scheduling, fairness algorithms |
| **8** | Polish: recurring expenses, CSV export, offline caching, push notifications, exact settlement DP | PWA depth, stretch algorithms |

## Phase 0 — Skeleton ✅

- [x] Monorepo layout (`frontend/`, `services/`, `libs/`, `deploy/`, `docs/`)
- [x] `deploy/docker-compose.dev.yml`: Traefik + Postgres + placeholder services
- [x] Hello-world Go service (`services/auth`) behind `/api/v1/auth/healthz`
- [x] Hello-world ASP.NET Core service (`services/kitty`) behind `/api/v1/kitty/healthz`
- [x] Vite + React + TS PWA shell, installable, served via Traefik
- [x] GitHub Actions: path-filtered build + test per service

## Phase 1 — Auth

- [ ] Postgres schema + `golang-migrate` setup; users + refresh-token tables
- [ ] Register / login (argon2id), access JWT (EdDSA, ~15 min) + rotating refresh token (httpOnly cookie)
- [ ] JWKS endpoint; JWT-verification middleware in `libs/go-common` and `libs/dotnet-common`
- [ ] Frontend: register/login screens, in-memory access token, silent refresh
- [ ] Unit + Testcontainers integration tests

## Phase 2 — Wallet (independent track)

No dependency on Households/Kitty/Chores — only needs a verified JWT from Auth. Can be built any time after Phase 1, interleaved with Phase 3 if useful.

- [ ] Shared recurrence model (`times` / `every` / `unit` / `anchor` + optional `until` end-anchor for bounded recurrences — term-length jobs, fixed-duration passes) in both shared libs, property-tested
- [ ] Account CRUD (user-scoped: current / savings / investment / other)
- [ ] Income sources: recurring (amount/cycle, optionally `until`-bounded) and one-off, each with an **expected** amount distinct from what's actually received
- [ ] Recurring/one-off expenses (subscriptions, rent amortised to weekly-equivalent, bounded passes)
- [ ] Transfers between own accounts (weekly/monthly top-ups, roundup lump sums)
- [ ] Categorised expense ledger + budget lines (descriptive only — actual-vs-budget stat, no alerts)
- [ ] Inventory reconciliation: manual actual-balance entry vs computed expected balance, diff shown, adjusting entry to resync
- [ ] Stats: category spend + budget-vs-actual, income expected-vs-actual variance, account balance trend, projected-remaining/runway
- [ ] Frontend: accounts overview, income/expense setup, ledger entry, inventory flow, wallet stats screens

## Phase 3 — Households

- [ ] Household CRUD, membership, admin/member roles
- [ ] Invite codes / share links (rotatable, expiring)
- [ ] Internal membership-check endpoint (used by kitty/chores in v1, with ~60s cache)
- [ ] Activity feed (write API + list endpoint)
- [ ] Frontend: household switcher, create/join flows, member management

## Phase 4 — Kitty core

- [ ] Expense CRUD with categories (defaults + household-defined)
- [ ] Split methods: EQUAL / PERCENTAGE / EXACT / SHARES with validation
- [ ] Largest-remainder rounding (property-tested: shares always sum to total)
- [ ] Money movements (balance-affecting, stats-excluded by construction)
- [ ] Live balances (net per member + pairwise who-owes-who)
- [ ] Frontend: ledger, add-expense split UI, movements, balance summary

## Phase 5 — Settlements & kitty stats

- [ ] Review schedules per household (recurrence or manual)
- [ ] Settlement periods: close, snapshot, net balances
- [ ] Greedy transfer-minimisation (≤ n−1 transfers); mark-paid → auto money movement → `SETTLED`
- [ ] Settled-period immutability (corrections = adjusting entries)
- [ ] Stats endpoints: total / by category / by payer / per-member share / trend, over week / month / year / all-time / custom
- [ ] Frontend: settlement flow, stats screens (Recharts)

## Phase 6 — Event-driven refactor

- [ ] NATS in Compose; event envelopes in shared libs
- [ ] Household publishes `member.added` / `member.removed` / activity events
- [ ] Kitty + chores keep local membership read-models; drop sync membership calls
- [ ] Activity feed consumes events instead of direct posts

## Phase 7 — Chores

- [ ] Chore CRUD: category, recurrence, `people_required`, eligible members, effort weight
- [ ] Occurrence expansion job (~6-week rolling window)
- [ ] Fairness-counter rota assignment (weighted, per-chore + overall balance)
- [ ] Regeneration that never touches completed/overridden occurrences
- [ ] Swaps + manual reassignment (logged to activity feed)
- [ ] Completion tracking + scheduled `MISSED` marking (grace period)
- [ ] Chore stats: completed/missed per member & per chore, streaks, leaderboard
- [ ] Frontend: rota calendar, chore management, swap flow, stats

## Phase 8 — Polish & stretch

- [ ] Recurring expenses (template + recurrence → auto-materialised)
- [ ] CSV export of kitty ledger
- [ ] Offline caching (TanStack Query persistence) + PWA polish
- [ ] Web push notifications (chore due, review day, added-to-expense) via Go notification service
- [ ] Exact settlement optimisation (bitmask subset-sum DP for n ≤ ~12), greedy as baseline
- [ ] Balance write-off flow + leave-household guard
- [ ] Production deploy: VPS, Traefik TLS, GHCR images, deploy workflow, `deploy/VPS.md` runbook