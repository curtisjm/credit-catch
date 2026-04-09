# Web App Architecture

## Stack

- **Framework**: Next.js 16 (App Router, React 19)
- **UI**: shadcn/ui (base-nova style, Base UI primitives)
- **Styling**: Tailwind CSS v4 with CSS custom properties
- **Font**: Geist Sans (body/headings), Geist Mono (data)
- **Package manager**: pnpm

## Project Structure

```
web/src/
├── app/
│   ├── layout.tsx           # Root layout — sets dark class, loads Geist fonts
│   ├── page.tsx             # Redirects to /dashboard
│   ├── globals.css          # Tailwind imports + Credit Catch theme tokens
│   ├── (auth)/
│   │   ├── layout.tsx       # Centered card layout for auth pages
│   │   ├── login/page.tsx   # Login form → POST /api/v1/auth/login
│   │   └── signup/page.tsx  # Signup form → POST /api/v1/auth/signup
│   └── (dashboard)/
│       ├── layout.tsx       # Auth guard + header + nav shell
│       ├── dashboard/page.tsx  # KPI cards + card breakdown + expiring alerts
│       ├── cards/page.tsx      # Card catalog + user card management
│       ├── credits/page.tsx    # Credit tracking with mark used/unused
│       └── settings/page.tsx   # Profile + notification prefs (stub)
├── components/
│   ├── ui/                  # shadcn/ui components (14 total)
│   ├── header.tsx           # Top nav bar with logo + nav + sign out
│   └── nav.tsx              # Dashboard/Cards/Credits/Settings navigation
├── hooks/
│   └── use-auth.ts          # Auth guard hook — redirects to /login if no session
├── lib/
│   ├── api.ts               # HTTP client with JWT auth + automatic 401 refresh
│   ├── auth.ts              # login/signup/logout + user persistence
│   ├── utils.ts             # cn() — shadcn class merging utility
│   └── mock-data.ts         # Mock data for development (can be removed)
└── types/
    └── api.ts               # TypeScript interfaces matching backend JSON shapes
```

## Theme

Dark-mode-first design. The `dark` class is hardcoded on `<html>` in the root layout.

| Token | Value | Usage |
|-------|-------|-------|
| `--background` | `#09090b` (zinc-950) | Page background |
| `--card` | `#18181b` (zinc-900) | Card/surface backgrounds |
| `--primary` | `#10b981` (emerald-500) | Buttons, links, active states |
| `--destructive` | `#ef4444` (red-500) | Error states, remove actions |
| `--success` | `#22c55e` (green-500) | "Used" badge |
| `--warning` | `#f59e0b` (amber-500) | "Expiring" badge, unused value |
| `--danger` | `#ef4444` (red-500) | "Expired" badge |
| `--radius` | `0.5rem` | Border radius base |

## Authentication Flow

1. User submits login/signup form
2. `auth.ts` calls `POST /api/v1/auth/{login,signup}`
3. Access token stored in memory (`AuthTokens` class), refresh token in `localStorage`
4. User object persisted in `localStorage` (key: `credit_catch_user`)
5. Dashboard layout wraps children in `useAuth()` hook — redirects to `/login` if no session
6. API client automatically retries on 401 using the refresh token
7. If refresh fails, tokens are cleared and user is redirected to `/login`

## API Client

`lib/api.ts` provides `api.get()`, `api.post()`, etc. with:
- Automatic `Authorization: Bearer` header injection
- Automatic token refresh on 401
- JSON serialization/deserialization
- `ApiError` class with `status` and `body` for error handling

All API paths use the `/api/v1/` prefix to match the backend routes.

## shadcn/ui Components

Installed via `pnpm dlx shadcn@latest`. The `Badge` component has been extended with custom variants (`success`, `warning`, `danger`, `info`) that use the theme's semantic colors.

Components: button, card, badge, table, tabs, input, label, select, dialog, dropdown-menu, navigation-menu, progress, separator, sheet.
