# user-analytics

Backend service that tracks user logins and returns **daily** and **monthly** unique user counts.

## Stack

- Go, [chi](https://github.com/go-chi/chi) HTTP router
- [ent](https://entgo.io/) ORM + SQLite (`modernc.org/sqlite`)
- Write-time rollup tables for fast count queries

## Quick start

```bash
go mod tidy
go mod vendor
go test ./...
make run        # http://localhost:8080
```

After changing ent schemas: `make generate`

After changing `LoginRepository` or `AnalyticsService` interfaces: `make mocks` (requires [mockery](https://github.com/vektra/mockery) v3)

Environment:

| Variable | Default | Description |
|----------|---------|-------------|
| `ADDR` | `:8080` | HTTP listen address |
| `DATABASE_PATH` | `./data/analytics.db` | SQLite file path |

## Database design

### Schema (3 tables)

```sql
CREATE TABLE user_logins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    login_time DATETIME NOT NULL,
    UNIQUE (user_id, login_time)
);
CREATE INDEX idx_user_logins_login_time ON user_logins (login_time);

CREATE TABLE daily_unique_users (
    date TEXT NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY (date, user_id)
);

CREATE TABLE monthly_unique_users (
    month TEXT NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY (month, user_id)
);
```

| Table | Role |
|-------|------|
| `user_logins` | Event log; `UNIQUE (user_id, login_time)` prevents double-counting on retry |
| `daily_unique_users` | One row per user per UTC day — fast daily counts |
| `monthly_unique_users` | One row per user per UTC month — fast monthly counts |

### Performant aggregation

Daily and monthly views use **`COUNT(*)` on rollup tables** keyed by `date` or `month`, not `COUNT(DISTINCT user_id)` over every login row. Uniqueness is enforced at ingest by the primary keys on the rollup tables.

### Ingest (write path)

Each login runs in one transaction:

1. Insert into `user_logins` (skip on duplicate).
2. Derive UTC `date` / `month` from `login_time`.
3. Insert into `daily_unique_users` and `monthly_unique_users` (skip on duplicate).

### Query (read path)

Service methods (assignment names in parentheses):

- `GetDailyUserCount(date)` (`GetDailyUniqueUsers`) → `SELECT COUNT(*) FROM daily_unique_users WHERE date = ?`
- `GetMonthlyUserCount(month)` (`GetMonthlyUniqueUsers`) → `SELECT COUNT(*) FROM monthly_unique_users WHERE month = ?`

HTTP: `GET /v1/daily-user-count?date=…` and `GET /v1/monthly-user-count?month=…` return the count as a JSON integer.

### Assumptions

- Calendar `date` / `month` use **UTC** (`YYYY-MM-DD`, `YYYY-MM`).
- `user_id` is a UUID (stored as `TEXT` in SQLite).
- Schema is applied via ent **auto-migrate** (`Schema.Create` on startup).

### SQLite for the take-home, PostgreSQL for production

This repo runs on **SQLite** (`DATABASE_PATH`, pure Go driver) so you can clone, test, and demo with no extra services. The assignment asks for a cloud-compatible **relational** database; the same design maps directly to **PostgreSQL** (`SERIAL`, `TIMESTAMP WITHOUT TIME ZONE`, native `UUID`, identical unique/primary keys). For production you would point ent at a Postgres DSN, use versioned migrations instead of startup auto-migrate, and keep the rollup + event model unchanged.

## API

OpenAPI spec: [docs/openapi.yaml](docs/openapi.yaml). Clickable requests: [api.http](api.http) (VS Code **REST Client** extension).

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/logins` | Record a login (`user_id` UUID, optional `login_time` RFC3339 UTC) |
| `GET` | `/v1/daily-user-count?date=YYYY-MM-DD` | Returns JSON integer — daily unique users |
| `GET` | `/v1/monthly-user-count?month=YYYY-MM` | Returns JSON integer — monthly unique users |

### Examples

```bash
# Ingest
curl -s -X POST http://localhost:8080/v1/logins \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"550e8400-e29b-41d4-a716-446655440000","login_time":"2026-06-03T10:00:00Z"}'

# Queries
curl -s 'http://localhost:8080/v1/daily-user-count?date=2026-06-03'
# 1

curl -s 'http://localhost:8080/v1/monthly-user-count?month=2026-06'
# 1
```

### Manual test walkthrough

Use a **fresh database** so counts are predictable (delete `./data/analytics.db` if you have run the server before).

1. Start the server: `make run`
2. Install the [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) extension and open [api.http](api.http)
3. Send these requests in order (use **Send Request** above each `###` block):
   - **Health** — expect `ok`
   - **Record login (user 1)** — `201`
   - **Record login (user 1 duplicate — idempotent)** — `201` (no extra unique user)
   - **Record login (user 2, same day)** — `201`
   - **Daily unique user count** — response body: `2`
   - **Monthly unique user count** — response body: `2`

The server log should show sentences like `Found 2 unique users on 2026-06-03.` and `Found 2 unique users in 2026-06.`

Skip **Record login (no login_time)** during this walkthrough unless you want counts tied to today’s UTC date.

## Project layout

```
cmd/server/          main
ent/schema/          ent entities
internal/api/        HTTP handlers + router
internal/domain/     shared errors and UTC date/month formats
internal/service/    business logic
internal/repository/ ent persistence
docs/openapi.yaml    API contract
api.http             REST Client sample requests
```

## Tests

```bash
go test ./...
```

- Unit tests: service and API layers use [mockery](https://github.com/vektra/mockery) mocks (`make mocks` to regenerate)
- Integration tests: repository against in-memory SQLite
