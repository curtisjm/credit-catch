# Authentication Flow

## Overview

Credit Catch uses JWT access tokens (short-lived) paired with opaque refresh tokens (long-lived, rotated on use). The backend implements token family tracking for stolen-token detection.

## Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/v1/auth/signup` | Create account |
| POST | `/api/v1/auth/login` | Sign in |
| POST | `/api/v1/auth/refresh` | Rotate tokens |
| POST | `/api/v1/auth/logout` | Revoke token family |

## Token Lifecycle

```
Signup/Login
  → Backend returns { access_token, refresh_token, user }
  → Frontend stores access_token in memory, refresh_token in localStorage
  → Frontend stores user object in localStorage

API Call
  → Authorization: Bearer <access_token>
  → If 401:
    → POST /api/v1/auth/refresh with { refresh_token }
    → If success: update tokens, retry original request
    → If fail: clear all tokens, redirect to /login

Logout
  → POST /api/v1/auth/logout with { refresh_token }
  → Clear all local storage
  → Redirect to /login
```

## Security Model

- **Access tokens**: JWT (HS256), contains `uid` claim, expires per server config (default 15min)
- **Refresh tokens**: 32-byte random hex, SHA-256 hashed before storage, 30-day validity
- **Token families**: Each refresh token belongs to a family. On rotation, old token is consumed and new one inherits the family. If a consumed token is replayed (indicating theft), the entire family is revoked.
- **Passwords**: bcrypt, cost 12
- **Validation**: email required (trimmed, lowercased), password minimum 8 characters

## Frontend Implementation

### Files
- `web/src/lib/auth.ts` — `login()`, `signup()`, `logout()`, `getUser()`, `isAuthenticated()`
- `web/src/lib/api.ts` — HTTP client with automatic 401 → refresh → retry
- `web/src/hooks/use-auth.ts` — React hook that redirects unauthenticated users to `/login`

### Auth Guard

The dashboard layout (`web/src/app/(dashboard)/layout.tsx`) uses the `useAuth()` hook. On mount, it checks `isAuthenticated()` (looks for access or refresh token). If neither exists, it redirects to `/login`. While checking, it shows a loading spinner.

### Storage Keys
- `refresh_token` — refresh token string in localStorage
- `credit_catch_user` — JSON-serialized user object in localStorage
- Access token is stored only in memory (not persisted across tabs/refreshes — refresh token handles rehydration)

## Error Handling

| Status | Frontend Behavior |
|--------|-------------------|
| 400 | Shows validation error from response body (`error` field) |
| 401 | "Invalid email or password" on login; triggers refresh on API calls |
| 409 | "An account with this email already exists" on signup |
| Other | Generic "Something went wrong" message |
| Network | "Unable to connect. Please try again." |
