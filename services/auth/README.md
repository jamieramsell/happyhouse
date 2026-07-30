# auth

Users, credentials, JWT issuance, refresh tokens, and profiles for HappyHouse.

Go service, module `github.com/jamieramsell/happyhouse/services/auth`, served under `/api/v1/auth/*`. See [`../../docs/DESIGN.md`](../../docs/DESIGN.md) for the domain model and [`../../CLAUDE.md`](../../CLAUDE.md) for conventions.

## Running the service

```bash
cd services/auth
go run ./cmd/server   # -> http://localhost:8080/api/v1/auth/healthz
go test ./...
go vet ./...
```

## Database migrations

Schema is managed with [`golang-migrate`](https://github.com/golang-migrate/migrate). Migration files live under [`migrations/`](migrations/), numbered and forward-only, as `up`/`down` SQL pairs:

```
migrations/
├── 000001_create_users_table.up.sql
├── 000001_create_users_table.down.sql
├── 000002_create_refresh_tokens_table.up.sql
└── 000002_create_refresh_tokens_table.down.sql
```

### Install the CLI

The `migrate` CLI needs the Postgres driver compiled in.

```bash
brew install golang-migrate          # macOS; ships with the postgres driver
```

Or as a pinned Go binary:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Run migrations locally

Start Postgres (the full stack, or just the DB) so `auth_db` exists — it and the `auth_svc` role are provisioned by [`../../deploy/postgres/init.sql`](../../deploy/postgres/init.sql):

```bash
docker compose -f deploy/docker-compose.dev.yml --env-file deploy/.env up postgres
```

Point the CLI at the local database and apply all pending migrations. From the repo root:

```bash
export AUTH_DATABASE_URL='postgres://auth_svc:auth_dev_pw@localhost:5432/auth_db?sslmode=disable'

migrate -path services/auth/migrations -database "$AUTH_DATABASE_URL" up
```

Migrations are idempotent — re-running `up` when everything is applied is a no-op.

Common operations:

```bash
# Roll back the most recent migration
migrate -path services/auth/migrations -database "$AUTH_DATABASE_URL" down 1

# Show the current version
migrate -path services/auth/migrations -database "$AUTH_DATABASE_URL" version

# Scaffold a new migration pair
migrate create -ext sql -dir services/auth/migrations -seq <name>
```

> **Note:** `down` migrations drop tables and are destructive — they exist for local development only. Production is forward-only.
