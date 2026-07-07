-- Provisions one database + one role per service, giving each service its own
-- isolated schema and credentials (database-per-service). Runs once, the first
-- time the Postgres data volume is created.
--
-- Dev-only passwords live here for convenience; production uses secrets and a
-- managed Postgres. The services do not connect yet in Phase 0 — this simply
-- has the databases ready for Phase 1 (auth) and Phase 3 (kitty).

CREATE ROLE auth_svc WITH LOGIN PASSWORD 'auth_dev_pw';
CREATE DATABASE auth_db OWNER auth_svc;

CREATE ROLE kitty_svc WITH LOGIN PASSWORD 'kitty_dev_pw';
CREATE DATABASE kitty_db OWNER kitty_svc;
