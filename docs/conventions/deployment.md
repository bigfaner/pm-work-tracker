---
scope: backend, frontend, build
source: feature/single-binary-deploy
---

# Single Binary Deployment Conventions

## TECH-deploy-001: embed.FS for Static Assets

The Go binary embeds the frontend `dist/` directory via `//go:embed dist` in package `backend/web`. The binary serves both API routes and static frontend assets with zero external dependencies (no Nginx, no file system reads).

`web.ValidateAssets()` checks that `dist/index.html` exists at startup. If missing, the process exits with a fatal error and a hint to run the build script.

**Why**: Compile-time embedding locks frontend and backend to the same version. No deployment mismatch is possible. Simplifies deployment to a single binary.

**Example**:
```go
// backend/web/embed.go
package web

import "embed"

//go:embed dist
var FS embed.FS
```

## TECH-deploy-002: Cache Headers

Static assets and SPA entry point use different caching strategies:

| Path | Cache-Control | Why |
|------|---------------|-----|
| `{BP}/assets/*` | `max-age=31536000, immutable` | Vite adds content hashes to filenames; cached forever |
| `{BP}/` (SPA entry) | `no-cache` | index.html changes on every deploy; must revalidate |
| `{BP}/*path` (SPA fallback) | `no-cache` | Same reason as entry |

**Why**: Content-hashed assets never change, so aggressive caching is safe. The SPA entry must not be cached or users will see stale JS/CSS references after deployment.

## TECH-deploy-003: Route Priority Order

Routes are registered in this priority order, ensuring API routes always take precedence:

1. `GET /health` — health check (no base path, no auth)
2. `{BP}/v1/*` — API routes (auth + permission middleware)
3. `{BP}/assets/*filepath` — static file serving (no auth)
4. `{BP}/` — SPA entry page (no auth)
5. `{BP}/*path` — SPA fallback catch-all (no auth)

**Why**: API routes must match before SPA fallback. If a request matches both an API pattern and a frontend route, the API handler wins. This ordering prevents the SPA catch-all from swallowing API requests.

## TECH-deploy-004: BASE_PATH Configuration

`BASE_PATH` is a single configurable prefix applied to all routes (API, static, SPA). It is read from the backend config file (`server.base_path`) and injected into the frontend at build time via `VITE_BASE_PATH`.

- Backend: Gin RouterGroup prefix (`r.Group(basePath)`)
- Frontend build: `vite.config.ts` sets `base: VITE_BASE_PATH || '/'`
- Frontend runtime: `BrowserRouter basename={VITE_BASE_PATH || '/'}`
- Frontend API client: `baseURL` = `${VITE_BASE_PATH ?? ''}/v1`

Default is empty string (local dev). Production uses a path like `/pm` for shared-domain deployments.

**Why**: Single config point for all routing. No risk of frontend/backend path mismatch.

## TECH-deploy-005: API Prefix Convention

API routes use `{BASE_PATH}/v1/*` (no `/api` prefix). The `/api` layer is eliminated — the version segment `/v1` alone distinguishes API routes from frontend routes.

**Before**: `/api/v1/teams/1/main-items`
**After**: `/v1/teams/1/main-items` (or `/pm/v1/teams/1/main-items` with BASE_PATH=/pm)

**Why**: Removing `/api` eliminates redundant path nesting. The `/v1` prefix is sufficient to differentiate API calls from frontend routes.

## TECH-deploy-006: Build Script Branch Validation

`scripts/build.sh` validates that the git branch matches the target environment:

- `dev` build requires the `dev` branch
- `prod` build requires the `main` branch

Branch mismatch exits with a non-zero code and an error message.

The script uses `set -e` so any step failure (npm ci, npm run build, go build) stops immediately.

**Why**: Prevents accidental deployment of development code to production or vice versa. Compile-time branch enforcement is a lightweight guard.
