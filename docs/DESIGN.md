# Happyhouse — Design Document

## Context

Happyhouse is a shared-household app for a uni house (~5 users): shared-cost ("kitty") tracking with flexible splits and transfer-minimising settlement, a fairness-balanced chore rota, and statistics for both. It is equally a **learning project**: gaining first hands-on experience with **microservices**, and learning **Go and C#** (coming from a Java background). The design therefore deliberately pushes the boat out where it teaches something, and stays pragmatic where it doesn't.

Foundational decisions:

- **Client**: PWA (installable on iOS + Android, no app stores, builds on existing TS baseline)
- **Backend**: polyglot microservices in **Go + C#**
- **Hosting**: cheap VPS + Docker Compose, Traefik/Caddy for HTTPS
- **Auth**: self-built (email+password, argon2id, JWTs) — fine for a trusted user base, maximum learning

---

## 1. System architecture

```
                        ┌─────────────────────────────┐
   Phone / Laptop       │   VPS (Docker Compose)      │
 ┌───────────────┐      │  ┌───────────────────────┐  │
 │  Happyhouse   │HTTPS │  │ Traefik (TLS, routing)│  │
 │  PWA (React)  ├──────┼──►     /api/v1/...       │  │
 └───────────────┘      │  └───┬────┬────┬────┬────┘  │
                        │      │    │    │    │       │
                        │   ┌──▼─┐┌─▼──┐┌▼───┐┌▼────┐ │
                        │   │auth││hshd││kitty││chore│ │
                        │   │(Go)││(Go)││(C#) ││(C#) │ │
                        │   └──┬─┘└─┬──┘└┬───┘└┬────┘ │
                        │      └────┴────┴─────┘      │
                        │   ┌────────▼──────────┐     │
                        │   │ Postgres (1 inst, │     │
                        │   │ 1 DB per service) │     │
                        │   └───────────────────┘     │
                        │   [NATS event bus — Phase 5]│
                        └─────────────────────────────┘
```

### 1.1 Services and language assignment

| Service | Language | Why this language | Owns |
|---|---|---|---|
| **auth** | Go | Small, security-critical, minimal surface — plays to Go's strengths; a gentle first Go service | Users, credentials, JWT issuance, refresh tokens, profiles |
| **household** | Go | CRUD-heavy but simple domain; second Go service cements the language | Households, membership, roles, invites, activity feed |
| **kitty** | C# / ASP.NET Core | Richest domain logic (splits, rounding, settlement algorithm, review periods) — benefits from C#'s expressive type system, LINQ, and a Java-familiar framework | Expenses, splits, money movements, balances, settlements, review schedules, spend stats |
| **chores** | C# / ASP.NET Core | Second algorithm-heavy domain (recurrence, rota fairness) | Chores, occurrences, rota, swaps, completions, chore stats |
| **frontend** | TypeScript / React | PWA, served as static files by Traefik/Caddy | UI, service worker, offline shell |

A notification service (web push / email) is a **Phase 7** addition in Go, consuming events.

Statistics live **inside the owning service** (kitty stats in kitty, chore stats in chores) — a separate reporting service would force cross-service data joins for zero benefit at this scale.

### 1.2 Communication

- **Client ↔ services**: REST/JSON through Traefik path routing (`/api/v1/auth/*` → auth, `/api/v1/households/*` → household, etc.). Each service ships an **OpenAPI spec**; TypeScript client types are generated from the specs (`openapi-typescript`), so frontend and backend can't silently drift.
- **Service ↔ service (v1)**: direct internal HTTP on the Compose network. The main need is membership checks ("is user X in household Y?") — kitty/chores call household's internal endpoint with a short (~60s) in-memory cache.
- **Service ↔ service (Phase 5 learning milestone)**: introduce **NATS** and event-carried state transfer. Household publishes `member.added` / `member.removed`; kitty and chores maintain local membership read-models and drop the synchronous calls. This is the classic monolith-habits → event-driven refactor and is deliberately scheduled *after* the app works, so the *why* of event-driven design is experienced, not just asserted.
- **Auth between services**: auth service signs JWTs with an **EdDSA/RS256 private key**; every other service verifies with the public key (exposed via a JWKS endpoint) — no shared secrets, no call to auth on every request. A small shared middleware library per language (`libs/go-common`, `libs/dotnet-common`) handles verification.

### 1.3 Data

- **One Postgres container**, but a **separate database + separate credentials per service** — real database-per-service isolation semantics (no cross-DB joins possible) without paying for four instances of RAM on a £5 VPS. Migrating a service to its own instance later is a connection-string change, which is the whole point of the repository pattern.
- **Money is stored as integer pence** (`amount_minor BIGINT`) with a `currency` column (always `GBP` for now). Never floats.
- **All quantities derived, never stored as truth**: balances are computed from the expense/movement ledger; stats are aggregated on read. At 5 users this is trivially fast and eliminates a whole class of drift bugs. (Materialise later only if ever needed.)
- Migrations: `golang-migrate` for Go services, EF Core migrations for C# services — each service owns and versions its own schema.

### 1.4 Repository pattern & dependency injection

Two explicit base requirements of the project:

- **Go services**: repository **interfaces live in the domain package** (e.g. `UserRepository`, `HouseholdRepository`), implemented in an `infra/postgres` package using `pgx`. Wiring is **manual constructor injection in `main.go`** — idiomatic Go DI (no framework magic); handlers depend on service structs, service structs depend on repo interfaces. Tests inject in-memory fakes.
- **C# services**: repository interfaces in the `Domain` project, EF Core implementations in `Infrastructure`, registered in the built-in **`Microsoft.Extensions.DependencyInjection`** container (`services.AddScoped<IExpenseRepository, EfExpenseRepository>()`). Controllers/endpoints take dependencies by constructor. This is the canonical ASP.NET Core layered layout (Api / Application / Domain / Infrastructure) and will feel familiar from Spring.
- Result: swapping Postgres for anything else means writing one new `infra` implementation per service and changing one registration line — no domain or handler code changes.

---

## 2. Domain design

### 2.1 Shared recurrence model (used by kitty reviews, chores, recurring expenses)

One model satisfies "daily, weekly, monthly, annually, custom, or every N units, or K times every N units":

```
Recurrence {
  times:  int      // occurrences per cycle, e.g. 2
  every:  int      // cycle length, e.g. 3
  unit:   DAY | WEEK | MONTH | YEAR
  anchor: date     // when the first cycle starts
}
// "weekly"                 = {1, 1, WEEK}
// "once every 2 days"      = {1, 2, DAY}
// "twice every 3 months"   = {2, 3, MONTH}
// plus a MANUAL sentinel (no auto-generation)
```

Occurrence expansion: a cycle spans `every × unit` from the anchor; `times` occurrences are spaced evenly within each cycle. Month/year arithmetic uses "same day-of-month, clamped" (Jan 31 + 1 month = Feb 28/29). This model is implemented **once per language** in the shared lib and property-tested hard, because everything leans on it.

### 2.2 Auth & users

- Register (email, display name, password → argon2id), login → short-lived access JWT (~15 min) + rotating refresh token (httpOnly cookie, revocable, stored hashed).
- JWT claims: `sub` (user id), `name`; **household roles are not in the token** (membership changes mid-session) — services check membership per request (v1: HTTP+cache; Phase 5: local read-model).
- v2: email verification + password reset (needs an SMTP dependency, so deferred), avatar.

### 2.3 Households

- Users can belong to **multiple households**. Roles: `admin` / `member` (admins edit household settings, review schedule, remove members).
- Joining via **invite code / share link** (rotatable, expiring) — right for the "flatmates on a group chat" flow.
- **Leaving guard**: a member with a non-zero kitty balance can't leave until settled or the balance is explicitly written off (household service asks kitty via internal call).
- **Activity feed** per household: "Alexa added Toaster £28", "James completed Bins", "Rota regenerated" — cheap to build (each service posts events to household service, or reads from NATS later) and hugely useful for trust in a shared-money app.

### 2.4 Kitty — expenses, splits, movements, balances

**Expense**: single payer (v1), amount (pence), category (household-defined list + defaults: Groceries, Bills, Household, Social, Other), description, date, participants, split method:

| Method | Input | Validation |
|---|---|---|
| `EQUAL` | just participants | — |
| `PERCENTAGE` | % per participant | must sum to 100.00 |
| `EXACT` | £x.xx per participant | must sum to expense amount |
| `SHARES` | ratio parts per participant (e.g. 2:1:1) | positive integers |

**Rounding** (the classic bug source): compute each share in fractional pence, floor, then distribute leftover pence by **largest remainder** with a deterministic tie-break (participant id) — shares always sum exactly to the total and re-computation is stable.

**Money movement**: a directed transfer `from → to, amount, note, optional link to an expense` (e.g. Alexa buys a toaster for £28, James transfers £10 towards it). It **shifts balances but is excluded from all spend statistics** — modelled as a distinct entity, not a special-cased expense, so the exclusion is structural rather than a filter someone can forget.

**Balances**: from the ledger, compute each member's net position (`paid − owed ± movements`). Pairwise "who owes who" view derived the same way.

**Recurring expenses**: rent, internet, streaming subs — a template + recurrence that auto-materialises expenses. Reuses the recurrence model; cheap win, very high real-world value for a uni house.

**Editing/deleting** expenses in an **open** period is allowed and recorded in the activity feed; expenses in a settled period are immutable (correction = new adjusting entry). This keeps historical settlements truthful.

### 2.5 Settlement — the debt-minimisation algorithm

- Each household has a **review schedule** (recurrence or manual). When a review is due (or triggered manually), a **Settlement Period** closes: it snapshots all not-yet-settled expenses/movements and computes net balances.
- **Algorithm (v1 — greedy)**: compute net balance per member; repeatedly match the largest debtor with the largest creditor for `min(|debt|, |credit|)`. Guarantees ≤ *n−1* transfers, runs in O(n log n), and is what Splitwise-class apps ship.
- **Algorithm (stretch — exact)**: true minimum transfer count is NP-hard, but for n ≤ ~12 a **subset-sum DP over bitmasks** finds zero-sum subgroups that settle internally, each saving one transfer. Flagged as an explicit algorithms-learning milestone with the greedy as the correctness baseline.
- Suggested transfers are presented; members **mark them paid** (each mark auto-creates a money movement); when all are paid the period is `SETTLED`. New expenses meanwhile accrue to the next open period.

### 2.6 Chores — rota generation and fairness

**Chore**: name, category, recurrence, `people_required` per occurrence, eligible members (who contributes), optional **effort weight** (1–5, default 1 — so "clean the oven" ≠ "wipe the table" in fairness maths).

**Rota generation**:
- Expand recurrences into dated **occurrences** over a rolling window (generated ~6 weeks ahead by a scheduled job).
- Assign via **fairness-counter greedy**: every member has a running weighted-assignment score per chore and overall; each occurrence takes the `people_required` eligible members with the lowest scores (ties → least-recently-assigned, then stable id order). This balances both per-chore ("everyone does bins equally") and overall load.
- **Regeneration** (member joins/leaves, chore edited) only touches **future, non-overridden, non-completed** occurrences — history and manual swaps are never clobbered.

**Overrides/swaps**: any two members can swap assignments; direct reassignment allowed; both logged to the activity feed. (Approval flows are v2 — housemates coordinate in person anyway.)

**Completion & missed**: assignees mark done; a scheduled job marks occurrences `MISSED` when the deadline + grace period (configurable, default = until next occurrence) passes. Missed chores feed stats — deliberately no auto-punishment; the leaderboard is the social pressure.

### 2.7 Statistics

Standard period picker everywhere: **week / month / year / all-time / custom range**.

- **Kitty** (excludes money movements by construction): total spend, by category, by payer, per-member consumed share, trend over time, largest expenses.
- **Chores**: completed / missed counts and rate per member, per chore, current streaks, "most reliable housemate" leaderboard.
- Served as JSON aggregates by the owning service; rendered client-side (Recharts or similar).

### 2.8 Added features summary (beyond the original brief)

Included in scope: activity feed, recurring expenses, effort-weighted chores, invite links, balance write-off, settled-period immutability, CSV export (kitty ledger — trivial, great for trust).

Deferred (v2+): push notifications (chore due / review day / added-to-expense), receipt photos, shopping list, email verification/reset, approval-based swaps.

---

## 3. Frontend (PWA)

- **React + TypeScript + Vite**, `vite-plugin-pwa` (installable, offline app-shell + cached last-known data via TanStack Query persistence), **TanStack Router + Query**, Tailwind CSS, Recharts for stats.
- Auth: access token in memory, refresh token in httpOnly cookie; silent refresh on 401.
- Mobile-first layouts (it will live on phones); works on desktop for free.
- Key screens: Login/Register → Household switcher → Household home (balance summary + upcoming chores) → Kitty (ledger, add expense with split UI, movements, settlement flow) → Chores (rota calendar, chore list, swap flow) → Stats → Household settings (members, invites, categories, review schedule).

## 4. Repository layout (monorepo)

```
happyhouse/
  README.md
  docs/               DESIGN.md (this doc), ROADMAP.md, adr/ (decision records)
  frontend/           React + TS PWA
  services/
    auth/             Go        cmd/ internal/{domain,app,infra,http}
    household/        Go        (same shape)
    kitty/            C#        src/{Api,Application,Domain,Infrastructure}, tests/
    chores/           C#        (same shape)
  libs/
    go-common/        JWT middleware, recurrence, errors, event envelopes
    dotnet-common/    same for C#
  deploy/
    docker-compose.yml, docker-compose.dev.yml, traefik/, .env.example, VPS.md
  .github/workflows/  per-service CI (path-filtered), deploy workflow
```

## 5. Testing & CI

- **Unit tests** concentrate where the risk is: split rounding, recurrence expansion, settlement greedy (+ property tests: shares sum to total; settlement transfers zero all balances), rota fairness (property: max−min weighted load bounded).
- **Integration tests** per service against real Postgres via **Testcontainers** (exists for both Go and .NET) — exercising the repository implementations.
- **E2E**: a small Playwright suite against `docker compose up` (register → household → expense → settle → chore → complete).
- **CI**: GitHub Actions, path-filtered per service (a kitty change doesn't rebuild auth); build + test + docker image push (GHCR); deploy = SSH to VPS, `docker compose pull && up -d`.

## 6. Risks / honest caveats

- **Polyglot microservices for 5 users is over-engineering** — that's the point (learning), but the roadmap phases are ordered so a working app exists early even if later phases slip.
- **iOS PWA limits**: no background sync; push requires the user to Add to Home Screen first. Acceptable; a native client remains a future option since everything is API-first.
- **Two languages = two of everything** (test setup, CI, shared lib). Mitigated by the shared-lib pattern and keeping service shapes symmetrical.

See [ROADMAP.md](./ROADMAP.md) for the phased build plan.
