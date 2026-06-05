---
created: "2026-06-05"
tags: [testing, local-dev-deployment]
---

# forge surfaces path matching requires surface key prefix

## Problem

`forge surfaces <path>` returns "no surface found" when called with `.` or `docs/features/<slug>/` as the path argument, even though `.forge/config.yaml` defines surfaces `backend=api` and `frontend=web`.

## Root Cause

`forge surfaces <path>` uses **segment prefix matching** against surface **keys** (not types). The path must start with a configured surface key to match:
- `backend/` → matches key `backend` → returns `api`
- `frontend/` → matches key `frontend` → returns `web`
- `.` or `docs/` or `tests/` → no key match → exit 1

This is a 3-level causation:
1. `/run-tests` skill tries `forge surfaces .` or `forge surfaces <feature-dir>` — neither starts with a surface key
2. The CLI path matching compares path segments against configured surface keys, not against arbitrary directories
3. Without a matching surface key prefix, the CLI returns exit 1 with "no surface found"

## Solution

Use one of these approaches:
1. **No argument**: `forge surfaces` (lists all configured surfaces — `backend=api` and `frontend=web`)
2. **Key-prefixed path**: `forge surfaces backend/` or `forge surfaces frontend/`
3. **Task frontmatter**: Extract `surface-type` from the task's YAML frontmatter (preferred when a task is active)

For `/run-tests` without an active task, use `forge surfaces` (no args) to get all surfaces, then run tests for each surface separately.

## Reusable Pattern

When `forge surfaces <path>` fails, check:
1. Does the path start with a configured surface key? (`backend/`, `frontend/`)
2. If no path is known, use `forge surfaces` (no args) to list all surfaces
3. For task-based invocations, prefer extracting `surface-type` from task frontmatter over CLI path matching

## Example

```bash
# Correct — lists all surfaces
forge surfaces
# Output: backend=api
#         frontend=web

# Correct — matches surface key prefix
forge surfaces backend/
# Output: api

# WRONG — no surface key prefix
forge surfaces .
# Error: no surface found for path "."
```

## References

- `.forge/config.yaml` surfaces configuration
- `/run-tests` skill Step 1: Surface detection
