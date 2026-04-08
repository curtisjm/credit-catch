# CreditCatch Backend

Go REST API for tracking and maximizing credit card benefits (statement credits, travel credits, dining credits, etc.).

## Architecture

```
cmd/server/main.go              Entry point, graceful shutdown, slog logging
internal/
  auth/
    jwt.go                      JWT issue/validate (HS256, configurable expiry)
    password.go                 bcrypt hash/compare (cost 12)
    refresh.go                  Refresh token rotation with theft detection
    oauth.go                    Google ID token verification
  config/config.go              Environment-based config (envconfig)
  credits/periods.go            Credit period generation engine
  database/database.go          pgx connection pool
  server/
    server.go                   Chi router, middleware stack, health endpoint
    middleware.go               JWT auth middleware
    auth.go                     Signup, login, refresh, logout handlers
    oauth.go                    OAuth handler (Google, Apple stub)
    cards.go                    Card catalog list/detail (public)
    user_cards.go               User card CRUD (authenticated)
    credits.go                  Credit tracking: list, current, mark used/unused
    dashboard.go                Dashboard: summary, annual, monthly aggregation
migrations/
  001_initial_schema.up.sql     10 tables: users, cards, credits, plaid, transactions
  002_oauth_and_statement_day   OAuth accounts + statement_close_day on user_cards
  003_refresh_tokens            Refresh token rotation table
../shared/docs/
  api-spec.yaml                 OpenAPI 3.1 specification (17 endpoints)
```

## Quick Start

### Prerequisites

- Go 1.22+
- Docker (for PostgreSQL)

### Setup

```bash
# Start PostgreSQL
make db-up

# Run migrations
make migrate-up

# Copy and configure environment
cp .env.example .env
# Edit .env: set JWT_SECRET to a secure random string

# Build and run
make run
```

The server starts at `http://localhost:8080`. Health check: `GET /health`.

### Development Commands

| Command | Description |
|---|---|
| `make build` | Compile to `bin/creditcatch-server` |
| `make run` | Build and run the server |
| `make test` | Run unit tests with race detector |
| `make test-integration` | Run integration tests (requires running Postgres) |
| `make lint` | Run `go vet` |
| `make migrate-up` | Apply database migrations |
| `make migrate-down` | Rollback last migration |
| `make db-up` | Start PostgreSQL via Docker Compose |
| `make db-down` | Stop PostgreSQL |

## API Overview

Base URL: `/api/v1`

### Public Endpoints

| Method | Path | Description |
|---|---|---|
| POST | `/auth/signup` | Create account (email, password, name) |
| POST | `/auth/login` | Login, returns access + refresh tokens |
| POST | `/auth/refresh` | Rotate refresh token, get new token pair |
| POST | `/auth/logout` | Revoke refresh token family |
| POST | `/auth/oauth` | OAuth login (Google ID token) |
| GET | `/cards` | List card catalog (paginated, filterable) |
| GET | `/cards/{card_id}` | Card detail with credit definitions |

### Authenticated Endpoints (Bearer JWT)

| Method | Path | Description |
|---|---|---|
| GET | `/me/cards` | List user's cards |
| POST | `/me/cards` | Add card (requires `statement_close_day`) |
| GET | `/me/cards/{id}` | Get user card detail |
| PATCH | `/me/cards/{id}` | Update card (nickname, dates, etc.) |
| DELETE | `/me/cards/{id}` | Remove card from collection |
| GET | `/me/credits` | List credit periods (paginated, filterable) |
| GET | `/me/credits/current` | Current active credits grouped by card |
| POST | `/me/credits/{id}/mark-used` | Mark credit as used |
| POST | `/me/credits/{id}/mark-unused` | Revert credit to unused |
| GET | `/me/dashboard/summary` | Current period summary |
| GET | `/me/dashboard/annual` | Annual breakdown by month |
| GET | `/me/dashboard/monthly` | Monthly breakdown by card |

Full API contract: [`shared/docs/api-spec.yaml`](../shared/docs/api-spec.yaml)

## Authentication

### Token Flow

1. **Signup/Login** returns `{access_token, refresh_token, user}`
2. **Access token** (JWT, 24h default): sent as `Authorization: Bearer <token>`
3. **Refresh token** (opaque, 30 days): sent to `POST /auth/refresh` to get a new pair
4. **Rotation**: each refresh consumes the old token and issues a new one
5. **Theft detection**: reusing a revoked refresh token revokes the entire token family

### OAuth

Clients handle the OAuth UI (Google Sign-In SDK, Sign in with Apple). Send the resulting ID token to:

```
POST /api/v1/auth/oauth
{"provider": "google", "id_token": "..."}
```

The backend verifies the token with the provider and either logs in an existing user or creates a new OAuth-only account (no password).

## Credit Period Engine

Credit periods are auto-generated when a user adds a card:

- **Monthly credits**: anchored to `statement_close_day`. If close day is 15, the period runs from the 16th to the 15th of the next month.
- **Annual credits**: anchored to `opened_date`. Annual cycle runs from the anniversary date to the day before the next anniversary.
- **One-time credits**: single period from opened date.

If no `opened_date` is provided, annual and one-time credit periods are **not** auto-generated. The user can still manually track them via the mark-used/unused endpoints.

Edge cases handled:
- Close day 31 in months with fewer days (clamps to last day of month)
- Leap year opened dates (Feb 29 clamps to Feb 28 in non-leap years)
- Year boundary crossings

## Database Schema

10 tables in the initial migration plus 2 support tables:

- **users** — accounts (email/password or OAuth-only)
- **oauth_accounts** — provider links (Google, Apple)
- **card_catalog** — credit card products
- **credit_definitions** — benefits per card (name, amount, period)
- **credit_match_rules** — future: auto-match transactions to credits
- **user_cards** — user's card collection with `statement_close_day`
- **credit_periods** — materialized tracking rows per benefit per period
- **refresh_tokens** — token rotation with family-based theft detection
- **notification_prefs**, **statement_uploads**, **plaid_items**, **transactions** — future features

## Testing

```bash
# Unit tests (no database required)
make test

# Integration tests (requires running Postgres with migrations applied)
DATABASE_URL="postgres://creditcatch:creditcatch@localhost:5432/creditcatch?sslmode=disable" \
JWT_SECRET=test-secret \
go test ./... -tags=integration -v

# Tests with coverage
go test ./... -cover
```

### Test Coverage

- **auth package**: JWT issue/validate, password hash/compare, token hashing (12 tests)
- **credits package**: monthly period edge cases, annual periods, leap years, clamp/days helpers (13 tests)
- **server package**: middleware auth validation, writeJSON, parseLimit, integration tests (9 unit + integration suite)

## Configuration

All configuration via environment variables:

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `JWT_SECRET` | Yes | — | HMAC secret for JWT signing |
| `PORT` | No | `8080` | HTTP listen port |
| `JWT_EXPIRY` | No | `24h` | Access token lifetime |
| `ENVIRONMENT` | No | `development` | `development` or `production` |
| `LOG_LEVEL` | No | `info` | Logging level |
| `PLAID_ENV` | No | `sandbox` | Plaid environment |
| `PLAID_CLIENT_ID` | No | — | Plaid client ID |
| `PLAID_SECRET` | No | — | Plaid secret key |
