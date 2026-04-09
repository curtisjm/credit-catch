# Getting Started

## Prerequisites

- [Nix](https://nixos.org/download) with flakes enabled
- [direnv](https://direnv.net/) (optional but recommended)

## Setup

```bash
# Clone the repo
git clone git@github.com:curtisjm/credit-catch.git
cd credit-catch

# Enter the nix dev shell (or let direnv do it automatically)
nix develop

# Start PostgreSQL
db-start

# Run database migrations
cd backend && make migrate-up

# Seed the card catalog (18 cards, ~60 credit definitions)
make seed

# Start the Go backend (port 8080)
make run
```

In a separate terminal:

```bash
# Install web dependencies and start the dev server (port 3000)
cd web && pnpm install && pnpm dev
```

## First Run

1. Open http://localhost:3000
2. You'll be redirected to /login — click "Sign up" to create an account
3. Browse the card catalog at /cards and add a card (e.g. Chase Sapphire Reserve)
4. Credit periods are auto-generated — go to /credits to see them
5. Mark credits as used/unused — the /dashboard updates in real time

## Available Commands

### Nix Dev Shell

| Command | Description |
|---------|-------------|
| `db-start` | Start PostgreSQL |
| `db-stop` | Stop PostgreSQL |
| `db-reset` | Wipe and reinitialize the database |

### Backend (from `backend/` directory)

| Command | Description |
|---------|-------------|
| `make build` | Build the Go binary |
| `make run` | Build and run the server |
| `make test` | Run unit tests |
| `make test-integration` | Run integration tests (requires running DB) |
| `make migrate-up` | Apply all SQL migrations |
| `make migrate-down` | Revert all SQL migrations |
| `make seed` | Load card catalog from `shared/seed/card_catalog.json` |
| `make lint` | Run `go vet` |
| `make clean` | Remove build artifacts |

### Web (from `web/` directory)

| Command | Description |
|---------|-------------|
| `pnpm dev` | Start Next.js dev server (port 3000) |
| `pnpm build` | Production build |
| `pnpm lint` | ESLint |

## Teardown

```bash
# Stop the backend and web dev server with Ctrl-C
# Stop PostgreSQL
db-stop
```

## Environment Variables

Set automatically by the nix dev shell:

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://creditcatch@localhost:5432/creditcatch?host=...&sslmode=disable` | PostgreSQL connection string |
| `JWT_SECRET` | `dev-secret-do-not-use-in-production` | JWT signing key |
| `API_PORT` | `8080` | Backend HTTP port |
| `ENVIRONMENT` | `development` | Enables debug logging |
| `NEXT_PUBLIC_API_URL` | *(not set — defaults to `http://localhost:8080`)* | Backend URL for the web app |
