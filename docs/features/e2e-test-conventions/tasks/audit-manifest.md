# Audit Manifest: tests/e2e/ Surface Classification

Generated: 2026-06-03
Source: `tests/e2e/` (52 spec files)
Classification rules from: `docs/proposals/e2e-test-conventions/proposal.md`

## Summary

| Surface | Count | Notes |
|---------|-------|-------|
| api     | 22    | Pure HTTP tests, no browser APIs |
| web     | 20    | Uses page/browser/locator APIs |
| infra   | 7     | Uses runCli/execSync for build/static checks |
| shared  | 0     | No spec files; 1 helper (config-setup.ts) + helpers.ts + config.yaml |

**Special handling flags:**

| Flag | Files | Count |
|------|-------|-------|
| `request.newContext` | item-list-fixes.spec.ts, item-list.spec.ts, progress-auto-status.spec.ts | 3 |
| `.serial` | item-list-fixes.spec.ts, item-list.spec.ts, progress-auto-status.spec.ts, weekly-view.spec.ts | 4 |
| `waitForTimeout` (anti-pattern, mark only) | 13 files, 166 total occurrences | 166 |
| No surface suffix (needs human-judged classification) | 20 files | 20 |

## File Classification

### api/ (22 files) — API Functional Tests, migrate to Vitest

| # | Source Path | Target Path | Flags | Notes |
|---|-------------|-------------|-------|-------|
| 1 | `items/rbac-api.spec.ts` | `tests/api/items/rbac.spec.ts` | | HTTP only, curl() |
| 2 | `items/soft-delete-api.spec.ts` | `tests/api/items/soft-delete.spec.ts` | | HTTP only, curl() |
| 3 | `items/status-flow-api.spec.ts` | `tests/api/items/status-flow.spec.ts` | | HTTP only, curl() |
| 4 | `items/bizkey-validation-api.spec.ts` | `tests/api/items/bizkey-validation.spec.ts` | | HTTP only |
| 5 | `items/improve-ui-api.spec.ts` | `tests/api/items/improve-ui.spec.ts` | | HTTP only |
| 6 | `items/schema-alignment-api.spec.ts` | `tests/api/items/schema-alignment.spec.ts` | | HTTP only, curl() |
| 7 | `items/jlc-schema-api.spec.ts` | `tests/api/items/jlc-schema.spec.ts` | | HTTP + runCli (mixed; primarily API endpoint validation, runCli for schema file check) |
| 8 | `items/untested-endpoints.spec.ts` | `tests/api/items/untested-endpoints.spec.ts` | | HTTP only, curl(); no surface suffix |
| 9 | `item-pool/pool-api.spec.ts` | `tests/api/item-pool/pool.spec.ts` | | HTTP only, curl() |
| 10 | `item-pool/rbac-api.spec.ts` | `tests/api/item-pool/rbac.spec.ts` | | HTTP only, curl() |
| 11 | `roles/rbac-api.spec.ts` | `tests/api/roles/rbac.spec.ts` | | HTTP only, curl() |
| 12 | `roles/soft-delete-api.spec.ts` | `tests/api/roles/soft-delete.spec.ts` | | HTTP only, curl() |
| 13 | `roles/unify-permission-checks-api.spec.ts` | `tests/api/roles/unify-permission-checks.spec.ts` | | HTTP only, curl() |
| 14 | `roles/rbac-cli.spec.ts` | `tests/api/roles/rbac-migration.spec.ts` | | No runCli despite -cli suffix; pure HTTP test for RBAC data migration |
| 15 | `roles/permission-granularity.spec.ts` | `tests/api/roles/permission-granularity.spec.ts` | | HTTP only, curl(); no surface suffix |
| 16 | `teams/rbac-api.spec.ts` | `tests/api/teams/rbac.spec.ts` | | HTTP only, curl() |
| 17 | `teams/bizkey-team-api.spec.ts` | `tests/api/teams/bizkey-team.spec.ts` | | HTTP only, curl() |
| 18 | `teams/improve-ui-api.spec.ts` | `tests/api/teams/improve-ui.spec.ts` | | HTTP only, curl() |
| 19 | `users/user-mgmt-api.spec.ts` | `tests/api/users/user-mgmt.spec.ts` | | HTTP only, curl() |
| 20 | `users/soft-delete-api.spec.ts` | `tests/api/users/soft-delete.spec.ts` | | HTTP only, curl() |
| 21 | `users/improve-ui-api.spec.ts` | `tests/api/users/improve-ui.spec.ts` | | HTTP only, curl() |
| 22 | `auth/login-errors.spec.ts` | `tests/api/auth/login-errors.spec.ts` | | HTTP only, curl(); no surface suffix |

### web/ (20 files) — Web E2E Tests, stay on Playwright

| # | Source Path | Target Path | Flags | Notes |
|---|-------------|-------------|-------|-------|
| 1 | `items/item-list.spec.ts` | `tests/web/items/item-list.spec.ts` | `request.newContext`, `.serial`, `waitForTimeout(54)` | Full browser + request API; largest waitForTimeout consumer |
| 2 | `items/item-list-fixes.spec.ts` | `tests/web/items/item-list-fixes.spec.ts` | `request.newContext`, `.serial`, `waitForTimeout(22)` | Browser + request API |
| 3 | `items/status-flow-ui.spec.ts` | `tests/web/items/status-flow-ui.spec.ts` | `waitForTimeout(16)` | Browser automation |
| 4 | `items/jlc-schema-ui.spec.ts` | `tests/web/items/jlc-schema-ui.spec.ts` | | Browser automation |
| 5 | `items/improve-ui-items.spec.ts` | `tests/web/items/improve-ui-items.spec.ts` | | No surface suffix; browser automation (page.goto, page.waitForLoadState) |
| 6 | `items/refresh-button.spec.ts` | `tests/web/items/refresh-button.spec.ts` | `waitForTimeout(1)` | No surface suffix; browser (locator, getByTestId) |
| 7 | `items/status-flow-dynamic.spec.ts` | `tests/web/items/status-flow-dynamic.spec.ts` | `waitForTimeout(5)` | No surface suffix; browser (page.getByRole) |
| 8 | `items/sub-item-edit.spec.ts` | `tests/web/items/sub-item-edit.spec.ts` | | No surface suffix; imports Page type |
| 9 | `items/view-pages.spec.ts` | `tests/web/items/view-pages.spec.ts` | `waitForTimeout(2)` | No surface suffix; browser (page.goto, page.locator) |
| 10 | `item-pool/item-pool.spec.ts` | `tests/web/item-pool/pool.spec.ts` | `waitForTimeout(34)` | No surface suffix; browser (page.locator) + fetch for setup |
| 11 | `roles/rbac-ui.spec.ts` | `tests/web/roles/rac.spec.ts` | `waitForTimeout(1)` | Browser automation |
| 12 | `roles/member-permissions.spec.ts` | `tests/web/roles/member-permissions.spec.ts` | | No surface suffix; browser (page) + fetch for setup |
| 13 | `teams/team-management.spec.ts` | `tests/web/teams/team-management.spec.ts` | | No surface suffix; browser automation |
| 14 | `teams/team-detail.spec.ts` | `tests/web/teams/team-detail.spec.ts` | `waitForTimeout(2)` | No surface suffix; browser-heavy (144 browser API calls) + fetch for setup |
| 15 | `users/user-mgmt-ui.spec.ts` | `tests/web/users/user-mgmt.spec.ts` | `waitForTimeout(3)` | Browser automation |
| 16 | `auth/improve-ui-auth.spec.ts` | `tests/web/auth/improve-ui-auth.spec.ts` | | No surface suffix; browser (page.goto, page.waitForLoadState) |
| 17 | `progress/progress-auto-status.spec.ts` | `tests/web/progress/progress-auto-status.spec.ts` | `request.newContext`, `.serial`, `waitForTimeout(6)` | Browser (page.locator) + request API |
| 18 | `smoke/full-e2e.spec.ts` | `tests/web/smoke/full-e2e.spec.ts` | `waitForTimeout(3)` | Cross-surface smoke; browser-heavy login flow → web/smoke/ per proposal |
| 19 | `weekly/weekly-view.spec.ts` | `tests/web/weekly/weekly-view.spec.ts` | `.serial`, `waitForTimeout(17)` | Browser automation |
| 20 | `infra/schema-alignment-ui.spec.ts` | `tests/web/smoke/schema-alignment-ui.spec.ts` | | Static file scan (no browser), but validates UI code; per proposal mapping |

### infra/ (7 files) — Build & Static Checks, migrate to Vitest

| # | Source Path | Target Path | Flags | Notes |
|---|-------------|-------------|-------|-------|
| 1 | `infra/bizkey-cli.spec.ts` | `tests/infra/bizkey-build.spec.ts` | Rename -cli → -build | Uses runCli; per proposal mapping |
| 2 | `infra/config-yaml-cli.spec.ts` | `tests/infra/config-yaml-build.spec.ts` | Rename -cli → -build | Uses runServer; per proposal mapping |
| 3 | `infra/jlc-schema-cli.spec.ts` | `tests/infra/schema-mysql.spec.ts` | Rename | Uses runCli (10 calls); per proposal mapping |
| 4 | `infra/lint-keywords-cli.spec.ts` | `tests/infra/lint-keywords.spec.ts` | Rename -cli → -build | Uses runCli (11 calls); per proposal mapping |
| 5 | `infra/e2e-rebuild-cli.spec.ts` | `tests/infra/e2e-rebuild.spec.ts` | Rename -cli → -build | Uses runCli; per proposal mapping |
| 6 | `infra/unify-permission-checks-build.spec.ts` | `tests/infra/permission-checks-build.spec.ts` | | Uses runCli; already has -build suffix |
| 7 | `infra/config-yaml-api.spec.ts` | `tests/api/smoke/config-yaml-api.spec.ts` | | Per proposal: goes to tests/api/smoke/ (uses curl, API-only) |

### infra/ — API Smoke (moved to api/smoke/) — 3 files

| # | Source Path | Target Path | Flags | Notes |
|---|-------------|-------------|-------|-------|
| 8 | `infra/deploy-smoke.spec.ts` | `tests/api/smoke/deploy.spec.ts` | | Per proposal mapping; uses curl(), API-only |
| 9 | `infra/jlc-schema-api.spec.ts` | `tests/api/smoke/jlc-schema-api.spec.ts` | | Per proposal mapping; API endpoint checks |
| 10 | `infra/schema-alignment-api.spec.ts` | `tests/api/smoke/schema-alignment-api.spec.ts` | | Per proposal mapping; API + curl |

> **Note**: infra/ actually contains 11 spec files, but 4 map to api/smoke/ or web/smoke/ per the proposal. The remaining 7 are true build/static checks staying in infra/.

### shared/ (non-spec files) — Shared Infrastructure

| # | Source Path | Target Path | Notes |
|---|-------------|-------------|-------|
| 1 | `tests/e2e/helpers.ts` | `tests/shared/helpers.ts` | Shared helpers; must not import @playwright/test post-migration |
| 2 | `tests/e2e/config.yaml` | `tests/shared/config.yaml` | Shared config |
| 3 | `tests/e2e/infra/config-setup.ts` | `tests/shared/config-setup.ts` | Per proposal mapping; shared config helper |

### Graduated markers (non-spec, archive only)

`tests/e2e/.graduated/` contains 12 marker directories (no spec files). These are not migrated; they serve as historical records.

## No-Suffix File Classification (20 files)

These files lack `-api`, `-ui`, or `-cli` suffix. Classification by rule application:

| # | File | Rule Applied | Surface | Browser APIs | HTTP Only | runCli |
|---|------|-------------|---------|-------------|-----------|--------|
| 1 | `items/untested-endpoints.spec.ts` | HTTP only (curl) | api | No | Yes | No |
| 2 | `items/improve-ui-items.spec.ts` | Browser (page.goto) | web | Yes | No | No |
| 3 | `items/refresh-button.spec.ts` | Browser (locator) | web | Yes | No | No |
| 4 | `items/status-flow-dynamic.spec.ts` | Browser (page.getByRole) | web | Yes | No | No |
| 5 | `items/sub-item-edit.spec.ts` | Browser (Page type) | web | Yes | No | No |
| 6 | `items/view-pages.spec.ts` | Browser (page.goto) | web | Yes | No | No |
| 7 | `items/item-list.spec.ts` | Browser + request.newContext | web | Yes | No | No |
| 8 | `items/item-list-fixes.spec.ts` | Browser + request.newContext | web | Yes | No | No |
| 9 | `item-pool/item-pool.spec.ts` | Browser (page.locator) | web | Yes | Mixed | No |
| 10 | `roles/permission-granularity.spec.ts` | HTTP only (curl) | api | No | Yes | No |
| 11 | `roles/member-permissions.spec.ts` | Browser (page) | web | Yes | Mixed | No |
| 12 | `teams/team-management.spec.ts` | Browser (page) | web | Yes | No | No |
| 13 | `teams/team-detail.spec.ts` | Browser (page) | web | Yes | Mixed | No |
| 14 | `auth/login-errors.spec.ts` | HTTP only (curl) | api | No | Yes | No |
| 15 | `auth/improve-ui-auth.spec.ts` | Browser (page.goto) | web | Yes | No | No |
| 16 | `progress/progress-auto-status.spec.ts` | Browser + request.newContext | web | Yes | No | No |
| 17 | `smoke/full-e2e.spec.ts` | Browser (page.goto) | web | Yes | No | No |
| 18 | `weekly/weekly-view.spec.ts` | Browser (page) | web | Yes | No | No |
| 19 | `infra/deploy-smoke.spec.ts` | curl only (API smoke) | api | No | Yes | No |
| 20 | `infra/unify-permission-checks-build.spec.ts` | runCli (build check) | infra | No | No | Yes |

## Cross-Validation with Proposal

### Proposal's Infra 文件迁移映射 check:

| Proposal Mapping | Actual File | Status |
|-----------------|-------------|--------|
| `tests/api/smoke/deploy-smoke.spec.ts` | `infra/deploy-smoke.spec.ts` | MATCH |
| `tests/api/smoke/jlc-schema-api.spec.ts` | `infra/jlc-schema-api.spec.ts` | MATCH |
| `tests/api/smoke/schema-alignment-api.spec.ts` | `infra/schema-alignment-api.spec.ts` | MATCH |
| `tests/api/smoke/config-yaml-api.spec.ts` | `infra/config-yaml-api.spec.ts` | MATCH |
| `tests/web/smoke/schema-alignment-ui.spec.ts` | `infra/schema-alignment-ui.spec.ts` | MATCH |
| `tests/infra/config-yaml-build.spec.ts` (原 config-yaml-cli) | `infra/config-yaml-cli.spec.ts` | MATCH |
| `tests/infra/bizkey-build.spec.ts` (原 bizkey-cli) | `infra/bizkey-cli.spec.ts` | MATCH |
| `tests/infra/schema-mysql.spec.ts` (原 jlc-schema-cli) | `infra/jlc-schema-cli.spec.ts` | MATCH |
| `tests/infra/lint-keywords.spec.ts` | `infra/lint-keywords-cli.spec.ts` | MATCH |
| `tests/infra/e2e-rebuild.spec.ts` | `infra/e2e-rebuild-cli.spec.ts` | MATCH |
| `tests/infra/permission-checks-build.spec.ts` (原 unify-permission-checks-build) | `infra/unify-permission-checks-build.spec.ts` | MATCH |
| `tests/shared/config-setup.ts` | `infra/config-setup.ts` | MATCH |

All 12 entries from proposal's Infra 文件迁移映射 verified. No omissions.

### Proposal's target directory structure check:

Cross-referencing the proposal's target tree against the 52 classified files:

| Proposal Target Dir | Expected Files | Actual Files Classified | Status |
|---------------------|---------------|------------------------|--------|
| `tests/api/items/` | 8 | 8 | MATCH |
| `tests/api/item-pool/` | 2 | 2 | MATCH |
| `tests/api/roles/` | 4 | 5 (includes rbac-cli → rbac-migration) | DIFFERS: +1 (rbac-cli classified as API, not in proposal tree) |
| `tests/api/teams/` | 3 | 3 | MATCH |
| `tests/api/users/` | 3 | 3 | MATCH |
| `tests/api/auth/` | 0 in proposal | 1 (login-errors) | DIFFERS: +1 (proposal tree does not list auth/ under api/) |
| `tests/api/smoke/` | 4 | 4 | MATCH |
| `tests/web/items/` | 2 in proposal | 9 | DIFFERS: +7 (no-suffix files classified as web) |
| `tests/web/roles/` | 1 in proposal | 2 | DIFFERS: +1 (member-permissions no suffix) |
| `tests/web/teams/` | 2 in proposal | 2 | MATCH |
| `tests/web/users/` | 1 in proposal | 1 | MATCH |
| `tests/web/auth/` | 1 in proposal | 2 | DIFFERS: +1 (improve-ui-auth no suffix) |
| `tests/web/item-pool/` | 1 in proposal | 1 | MATCH |
| `tests/web/weekly/` | 1 in proposal | 1 | MATCH |
| `tests/web/smoke/` | 1 in proposal | 2 | DIFFERS: +1 (schema-alignment-ui from infra) |
| `tests/web/progress/` | 0 in proposal | 1 | DIFFERS: +1 (progress-auto-status) |
| `tests/infra/` | 6 | 6 | MATCH |

**DIFFERS explanations**: The proposal's target tree shows only representative examples, not exhaustive file lists. The extra files are valid classifications per the proposal's own rules (browser API -> web, HTTP only -> api). No files are unaccounted for.
