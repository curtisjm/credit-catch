# Backend: CORS & Seed Loader

## CORS Middleware

**File**: `backend/internal/server/server.go`
**Package**: `github.com/go-chi/cors`

CORS is configured as the first middleware in the chi router chain (before logging, auth, etc.) so that preflight `OPTIONS` requests return immediately without hitting other middleware.

### Allowed Origins
- `http://localhost:3000` — local Next.js dev server
- `https://*.fly.dev` — production Fly.io deployments

### Allowed Methods
GET, POST, PUT, PATCH, DELETE, OPTIONS

### Allowed Headers
Accept, Authorization, Content-Type

### Config
- `AllowCredentials: true` — needed for cookie-based flows (future)
- `MaxAge: 300` — browsers cache preflight for 5 minutes

## Seed Loader

**Entry point**: `backend/cmd/seed/main.go`
**Data source**: `shared/seed/card_catalog.json`
**Makefile target**: `make seed`

### What It Does

Reads the card catalog JSON and upserts into `card_catalog` and `credit_definitions` tables.

### Data Mapping

The seed JSON has a richer schema than the Phase 1 database. The seed loader maps fields as follows:

| JSON field | DB column | Notes |
|-----------|-----------|-------|
| `annual_fee` (dollars) | `annual_fee` (cents) | Multiplied by 100 |
| `credits[].value_annual` | `amount_cents` | Used when `per_period_max` is absent |
| `credits[].per_period_max` | `amount_cents` | Takes precedence for monthly credits |
| `credits[].frequency` | `period` | `monthly` → `monthly`, `annual` → `annual`, others → `annual` |
| `credits[].scope` | `category` | Joined with commas |
| `credits[].condition` or `.note` | `description` | First non-empty wins |

Fields like `earning`, `currency`, `point_value_cpp`, `auto_apply`, `expires`, `enrollment_required`, and `per_use_max` are not stored in Phase 1. They will be modeled in Phase 2 when the credit matching engine needs them.

### Idempotency

The seed loader uses `ON CONFLICT ... DO UPDATE` for cards (keyed on `issuer + name`) and `ON CONFLICT DO NOTHING` for credit definitions. It is safe to run multiple times.

### Usage

```bash
# From the backend/ directory (requires DATABASE_URL in env)
make seed

# Or directly
go run ./cmd/seed ../shared/seed/card_catalog.json
```

### Prerequisites
1. PostgreSQL running (`db-start`)
2. Migrations applied (`make migrate-up`)
3. `DATABASE_URL` environment variable set (automatic in nix dev shell)
