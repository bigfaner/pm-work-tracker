# Standard Task IDs Collide Across Features

## Problem

T-test-2 (gen-test-scripts) in `permission-granularity` feature generated scripts under `docs/features/integration-test-coverage/testing/scripts/` instead of `docs/features/permission-granularity/testing/scripts/`.

## Root Cause

The standard test task chain uses fixed IDs: `T-test-1` (gen-test-cases) and `T-test-2` (gen-test-scripts). The `integration-test-coverage` feature (PR #18, merged April 28) previously executed T-test-2 and committed results to its own directory.

When `permission-granularity` feature's T-test-2 ran, the task executor detected the existing commit (`51a3be7`) and concluded "task already completed on 2026-04-28", skipping regeneration and referencing the wrong feature's output.

**Causal chain:** identical task ID → agent finds prior commit → skips execution → files land in wrong feature directory.

## Solution

No code fix needed for this specific case. For future features, either:
1. Use feature-scoped task IDs (e.g., `pg-T-test-1` instead of `T-test-1`) to prevent collision, or
2. Task executor should always verify output path matches current feature slug before declaring "already completed"

## Key Takeaway

When `/breakdown-tasks` generates standard test tasks (T-test-1, T-test-2), these IDs are not namespaced to the feature. If a previous feature ran the same IDs and was merged, the task executor will find the old commit and skip execution, writing output to the wrong feature directory. Verify the output path after any "already completed" detection.
