---
status: "blocked"
started: "2026-06-03 20:58"
completed: "N/A"
time_spent: ""
---

# Task Record: 3 迁移 API 测试到 tests/api/ + Vitest

## Summary
Migrated 22 API test files + 4 API smoke files from tests/e2e/ to tests/api/ with Vitest runner. Created vitest.config.ts (testTimeout 30s, hookTimeout 60s, sequential), package.json, tsconfig.json. All imports changed from @playwright/test to vitest. Helper imports updated from ../helpers.js to ../../shared/helpers.js. test.describe/test.beforeAll replaced with top-level describe/beforeAll. TypeScript compiles cleanly. Vitest discovers all 26 test files (failures are ECONNREFUSED — backend not running, not migration issues).

## Changes

### Files Created
- tests/api/vitest.config.ts
- tests/api/package.json
- tests/api/tsconfig.json
- tests/api/items/rbac.spec.ts
- tests/api/items/soft-delete.spec.ts
- tests/api/items/status-flow.spec.ts
- tests/api/items/bizkey-validation.spec.ts
- tests/api/items/improve-ui.spec.ts
- tests/api/items/schema-alignment.spec.ts
- tests/api/items/jlc-schema.spec.ts
- tests/api/items/untested-endpoints.spec.ts
- tests/api/item-pool/pool.spec.ts
- tests/api/item-pool/rbac.spec.ts
- tests/api/roles/rbac.spec.ts
- tests/api/roles/soft-delete.spec.ts
- tests/api/roles/unify-permission-checks.spec.ts
- tests/api/roles/rbac-migration.spec.ts
- tests/api/roles/permission-granularity.spec.ts
- tests/api/teams/rbac.spec.ts
- tests/api/teams/bizkey-team.spec.ts
- tests/api/teams/improve-ui.spec.ts
- tests/api/users/user-mgmt.spec.ts
- tests/api/users/soft-delete.spec.ts
- tests/api/users/improve-ui.spec.ts
- tests/api/auth/login-errors.spec.ts
- tests/api/smoke/deploy.spec.ts
- tests/api/smoke/jlc-schema-api.spec.ts
- tests/api/smoke/schema-alignment-api.spec.ts
- tests/api/smoke/config-yaml-api.spec.ts

### Files Modified
无

### Key Decisions
- Used sequence: { concurrent: false } instead of { sequential: true } because Vitest 3.x SequenceOptions does not have a sequential property — concurrent: false achieves the same sequential execution
- Imported describe and beforeAll as top-level exports from vitest instead of using test.describe/test.beforeAll because Vitest TypeScript types don't expose describe/beforeAll as methods on TestAPI
- Kept PROJECT_ROOT calculations unchanged since nesting depth (3 levels) is identical between tests/e2e/<journey>/ and tests/api/<journey>/
- Did not delete original tests/e2e/ files — subsequent tasks handle web/infra migration and final cleanup

## Test Results
- **Tests Executed**: Yes
- **Passed**: 5
- **Failed**: 9
- **Coverage**: 0.0%

## Acceptance Criteria
- [x] tests/api/ directory structure with vitest.config.ts (testTimeout 30s, hookTimeout 60s, sequential) and package.json
- [x] API test files migrated to tests/api/<journey>/ with imports from vitest
- [x] request.newContext files rewritten to curl() — no request.newContext in any API file (none existed)
- [x] .serial tests migrated to describe.sequential() — none existed in API files
- [x] npx vitest run from tests/api/ executes and discovers all 26 test files

## Notes
Tests fail with ECONNREFUSED because backend server is not running — this is expected and not related to the migration. All 26 test files are discovered by Vitest, TypeScript compiles cleanly, and import paths are correct. The original tests/e2e/ files are preserved for subsequent web/infra migration tasks.
