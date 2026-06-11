---
created: "2026-06-08"
tags: [testing, interface]
---

# E2E and API Tests Pass with Empty Seed Data

## Problem

After the full test pipeline ran (gen-contracts → gen-test-scripts → run-test-backend → run-test-frontend), all tests passed. However, manual QA revealed that no milestone map contained any milestones. Tests validated CRUD operations on empty containers without verifying the core user workflow: creating milestones within a milestone map.

## Root Cause

**L1 — Tests validate structure, not workflow completeness**: Generated test scripts test individual CRUD operations in isolation (create map, edit map, delete map) but don't test the multi-step workflow of creating a map, then adding milestones to it, then transitioning states.

**L2 — Seed data is minimal**: Test setup creates milestone maps but doesn't seed them with milestones. Tests assert "response is 200" and "data shape is correct" without checking that milestones exist within maps.

**L3 — Contract specs derived from journeys, journeys don't enforce data richness**: The journey eval scored 787/1000 and the contracts eval scored 665/1000, but the pipeline proceeded anyway (user bypassed target). Low-quality upstream artifacts produced low-quality tests that pass vacuously.

## Solution

- Test seed data must include realistic multi-entity scenarios (maps with milestones, milestones with items)
- E2E smoke tests should verify at least one end-to-end workflow, not just individual CRUD operations
- Consider adding a "golden path" test that exercises the full user workflow

## Reusable Pattern

- **Test passing ≠ feature working**: Tests that pass with zero child entities only prove the API doesn't crash on empty data.
- **Seed data matters more than test count**: 10 tests with realistic seed data > 50 tests with empty containers.
- **Low eval scores are warnings, not just numbers**: When journey/contract evals score below target, the downstream test quality is also compromised. Bypassing eval targets means accepting lower test effectiveness.

## Related Files

- `tests/frontend/milestone-map-lifecycle/`
- `tests/frontend/milestone-lifecycle/`
- `docs/features/milestone-map/testing/journeys/.eval-report.md`
