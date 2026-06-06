---
created: "2026-06-05"
tags: [testing, architecture]
---

# Move sub-item 422: contract missing HTTP method + repo whitelist gap

## Problem

The Move sub-item API returned 422 `INVALID_FIELD` in production. Three independent test layers all failed to catch this: E2E tests used the wrong HTTP method, service unit tests mocked the repo, and the API handbook was not consulted during test generation.

## Root Cause

Three-level causation chain:

1. **Contract files don't specify HTTP methods** — The contract `step-2-select-target-and-confirm.md` describes the move operation semantically but never mentions `PUT`. The `api-handbook.md` correctly defines `PUT /teams/:teamId/sub-items/:subId/move`, but the contract doesn't reference this.

2. **LLM-generated test scripts guess the method** — `/gen-test-scripts` produces test code from contracts. Without an explicit HTTP method in the contract, the LLM defaults to `POST` for "action" operations. The generated `sub-item-move.spec.ts` uses `curl('POST', .../move')` everywhere (lines 139, 251), but the route is registered as `PUT` in `router.go:136`.

3. **Service tests mock the repo interface** — `sub_item_service_test.go` uses a mock `SubItemRepo` whose `Update()` accepts any fields map. The real GORM repo's `UpdateFields` has an `acceptedFields` whitelist that rejects `main_item_key` and `item_code` with `ErrInvalidField`.

## Solution

Two fixes applied:
1. Add `main_item_key` and `item_code` to the SubItem `acceptedFields` whitelist in `pkg/repo/helpers.go`.
2. Fix E2E test to use `PUT` method (matches the registered route).

Prevention: contracts should reference the api-handbook endpoint (method + path) so that test generators use the correct HTTP method.

## Reusable Pattern

When contracts describe API operations:
- Include the HTTP method and path from the api-handbook (e.g., `<!-- endpoint: PUT /teams/:teamId/sub-items/:subId/move -->`)
- This gives the test script generator an authoritative source instead of relying on LLM guesswork

When adding new `repo.Update` calls with previously-unused field keys:
- Check `acceptedFields` in `pkg/repo/helpers.go` for the target model
- Add any new field keys to the whitelist

## Example

```markdown
<!-- Contract should reference the endpoint -->
- Operation: Move sub-item to another parent
- Endpoint: PUT /teams/:teamId/sub-items/:subId/move (api-handbook Move Sub Item)
```

## References

- `backend/internal/pkg/repo/helpers.go` — `acceptedFields` whitelist
- `backend/internal/handler/router.go:136` — route registered as `PUT`
- `docs/features/system-ux-optimization/design/api-handbook.md:72` — correctly specifies `PUT`
- `tests/sub-item-move/sub-item-move.spec.ts:139,251` — incorrectly uses `POST`
- `docs/features/system-ux-optimization/testing/sub-item-move/contracts/` — no HTTP method specified
