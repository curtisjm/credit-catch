# Credit Catch

**Stop leaving money on the table.** Premium credit cards come with monthly and annual credits — Uber rides, streaming services, airline fees, hotel stays — but most cardholders forget to use them. That's hundreds of dollars in value lost each year, while you're still paying the annual fee.

Credit Catch tracks every credit across all your cards, detects when you've used them (automatically via Plaid or manually), and reminds you before they expire. Think of it as a dashboard for getting your money's worth.

## The Problem

A single premium card like the Amex Platinum comes with **over $1,500/year in credits** spread across 8+ categories with different reset dates, merchant restrictions, and expiration rules. Add a Chase Sapphire Reserve and a Citi Strata Premier and you're juggling 15+ credits across 3 different calendars.

Nobody tracks this well in their head. Most people lose 20-40% of their available credit value every year — that's real money subsidizing an annual fee that no longer pays for itself.

## How It Works

1. **Add your cards** — Select from a catalog of popular premium cards with pre-loaded benefit data
2. **Connect your bank** (optional) — Plaid integration automatically detects when you've used a credit based on transaction matching
3. **Track manually** — One tap to mark a credit as used
4. **Get reminded** — Push notifications and emails before credits expire
5. **See your ROI** — Dashboard shows total value extracted vs. annual fees paid

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go (Chi, pgx, golang-migrate) |
| Web | Next.js, TypeScript, Tailwind, shadcn/ui, Aceternity UI, Recharts |
| iOS | Swift, SwiftUI, MVVM |
| Database | PostgreSQL |
| Bank Integration | Plaid |
| Hosting | Fly.io |

## Project Structure

```
backend/          Go REST API
web/              Next.js web application
ios/              SwiftUI iOS app
shared/
  docs/           OpenAPI spec (contract between frontend and backend)
  seed/           Credit card benefits catalog
```

## Development

```bash
# Start Postgres
docker-compose up -d

# Run backend
cd backend && make migrate-up && make run

# Run web (separate terminal)
cd web && pnpm install && pnpm run dev
```

## Status

**Active development** — Phase 1 (manual tracking MVP) is in progress.

- [x] Go backend with auth, card catalog, credit tracking, dashboard APIs
- [x] Next.js web app with dark theme, dashboard, card management
- [x] SwiftUI iOS app scaffolding with networking layer
- [x] Credit card benefits catalog (18 cards, verified data)
- [x] OpenAPI spec for all Phase 1 endpoints
- [ ] Plaid integration for automated tracking
- [ ] Push notifications and email reminders
- [ ] Statement upload and parsing
- [ ] Annual ROI calculator

---

<details>
<summary><strong>Built with Gas Town — Multi-Agent AI Development</strong></summary>

### What is Gas Town?

This project is being developed using [Gas Town](https://steve-yegge.medium.com/gas-town-emergency-user-manual-cf0e4556d74b), a multi-agent orchestration system created by Steve Yegge. Gas Town coordinates multiple AI agents working on a shared codebase simultaneously, each with distinct roles and persistent identities.

I'm using this project as an opportunity to experiment with agent-driven development at scale — treating the AI agents as a managed engineering team rather than a single coding assistant.

### The Setup

**Town infrastructure:**
- **Mayor** — Cross-rig coordinator. Dispatches work, handles escalations, makes architectural decisions.
- **Deacon** — Town-level watchdog. Monitors health, manages background maintenance.
- **Witness** — Per-rig health monitor. Detects stuck agents, manages cleanup.
- **Refinery** — Merge queue processor. Validates and merges completed work to main.

**Crew (persistent, interactive agents):**
| Agent | Role |
|-------|------|
| `backend` | Go API architecture, data model, endpoint design |
| `web` | Next.js frontend, dashboard UX, component architecture |
| `ios` | SwiftUI app, native iOS patterns, networking layer |
| `cards` | Credit card domain expert — benefit research, matching rules, seed data accuracy |
| `sheriff` | PR review — standing orders to review work before merge |

**Polecats (ephemeral, autonomous workers):**
Three pooled workers (`rust`, `chrome`, `nitro`) that pick up well-defined implementation tasks. They spawn, execute, submit to the merge queue, and self-destruct. Used for repetitive work after the crew has established patterns.

### How Work Flows

1. **Design phase** — I work interactively with crew agents in tmux sessions, iterating on architecture and design decisions
2. **Dispatch** — The mayor creates beads (work items) and slings them to polecats for implementation
3. **Execution** — Polecats work autonomously on feature branches
4. **Review** — Sheriff crew reviews output before merge
5. **Merge** — Refinery processes the merge queue, validates, and lands changes on main

### What I've Learned

This section will be updated as the project progresses with observations about multi-agent development patterns, what works well, and what doesn't.

</details>

## License

[GPL-3.0](LICENSE)
