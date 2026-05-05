---
created: "2026-05-04"
tags: [testing, architecture]
---

# Task-executor Should Verify Feature E2E for Frontend Tasks

## Problem

After implementing decision-log UI changes (collapsible timeline, icon buttons, card headers), the e2e tests had multiple failures (heading strict mode, selector ambiguity). These were only caught when the user manually ran `just test-e2e` mid-session. The Stop hook's `task all-completed` would eventually catch them too, but by then the agent's context is lost — fixing requires re-reading all the UI changes.

## Root Cause

Causal chain (5 levels):

1. **Symptom**: E2e breakage from UI changes goes undetected until user manually runs `just test-e2e`
2. **Direct cause**: Task-executor quality gate only runs `just test` (unit/integration), not `just test-e2e`
3. **Code cause**: `task-executor.md` Step 3 defines quality gate as `compile → fmt → lint → test` with no e2e step
4. **Process cause**: Forge's testing lifecycle has a gap between task-level and project-level:
   - Task-level: `just test` per task (agent has context)
   - **GAP**: no feature e2e per task
   - Project-level: `just test-e2e` via Stop hook (agent context lost)
5. **Root cause**: The three-layer testing lifecycle was designed assuming e2e changes only happen during T-test-3 (`/run-e2e-tests`). But implementation tasks in Phase 2-4 also modify UI/API, breaking e2e specs that were already generated in T-test-1/2/3.

**Why T-test-3 doesn't cover this**: T-test-3 generates and runs e2e specs against the current UI. Later implementation tasks (Phase 2-4) change the UI, invalidating those specs. There's no mechanism to re-verify feature e2e after each implementation task.

## Forge Improvement Proposal

### Current architecture (3-layer, with gap)

```
Task-level:      just test [scope]              ← agent has context, but no e2e
     ↓
  [GAP]          feature e2e not verified per task
     ↓
Project-level:   just test-e2e                  ← Stop hook, agent context lost
```

### Proposed: Extend run-tasks dispatcher with conditional e2e verification

**File to modify**: `forge/commands/run-tasks.md`

Add a new step between Step 3 (Verify Record) and Step 4 (Context Check):

```
### Step 3.5: Feature E2E Verification (conditional)

If the completed task had SCOPE=frontend or SCOPE=all:

1. Extract feature slug from current task context
2. Check if e2e specs exist:
   ls tests/e2e/features/<slug>/ 2>/dev/null || \
   ls tests/e2e/items/*decision-log* 2>/dev/null  # graduated specs
3. If specs exist, run:
   just e2e-setup
   just test-e2e --feature <slug>
4. On failure: dispatch error-fixer with failure context (same as Step 5 pattern)
5. On success or no specs: continue to Step 4
```

**Why in the dispatcher, not task-executor**:
- Precedent: Step 5 (Breaking Task Gate) already runs `just test` directly in the dispatcher
- Server lifecycle: `just e2e-setup` needs to run before e2e tests; dispatcher manages this better than the agent
- Error recovery: Dispatcher can dispatch error-fixer on failure, keeping the agent separation clean

**Why conditional (not always)**:
- Backend-only tasks (`scope: backend`) never affect e2e selectors
- Tasks before T-test-3 have no e2e specs yet (skip gracefully)
- Minimizes pipeline latency for unaffected tasks

### Alternative: Extend task-executor Step 1 with e2e awareness

**File to modify**: `forge/agents/task-executor.md`

In Step 1 (Read Task Definition), add:

```
**E2E Surface Check**: If SCOPE is frontend or all:
- Check for existing e2e specs related to the task's affected components
- Look in tests/e2e/features/<slug>/ and tests/e2e/<module>/
- If specs exist, note them for awareness during Step 2 (TDD Implementation)
- When modifying UI/API components that have e2e coverage, update selectors/assertions in the same commit
```

This is a softer change — no new verification step, but makes the agent proactive about maintaining e2e specs.

### Recommended: Both changes together

1. **Dispatcher Step 3.5** (systematic): Catches e2e breakage per-task, dispatches error-fixer while context is fresh
2. **Task-executor e2e awareness** (preventive): Agent updates e2e specs proactively, reducing breakage in the first place

The Stop hook remains as the final safety net.

## Reusable Pattern

### Per-task e2e verification closes the feedback loop:

```
Before:
  Task completes → unit tests pass → [GAP] → Stop hook catches e2e failure
                                                   ↑ agent context lost, hard to fix

After:
  Task completes → unit tests pass → feature e2e verified → Stop hook (safety net)
                                           ↑ error-fixer has fresh context
```

### When to run feature e2e per task:

| Condition | Run e2e? | Reason |
|-----------|----------|--------|
| scope=frontend, specs exist | **Yes** | UI changes may break selectors |
| scope=all, specs exist | **Yes** | Full-stack changes may break API/UI |
| scope=backend, specs exist | No | Backend changes caught by unit tests |
| scope=frontend, no specs yet | No | T-test-3 hasn't generated specs yet |

## Example

```markdown
# run-tasks.md — proposed Step 3.5

### Step 3.5: Feature E2E Verification (conditional)

If the completed task had `SCOPE: frontend` or `SCOPE: all`:

```bash
# Check for existing e2e specs (staging or graduated)
FEATURE_SLUG=$(task feature --slug)
SPEC_DIR=$(ls -d tests/e2e/features/${FEATURE_SLUG}/ 2>/dev/null || \
           ls -d tests/e2e/items/*${FEATURE_SLUG}* 2>/dev/null || echo "")

if [ -n "$SPEC_DIR" ]; then
    just e2e-setup && just test-e2e --feature ${FEATURE_SLUG}
    # On failure: dispatch error-fixer (same as Step 5 pattern)
fi
```

Skip if SCOPE is backend or no specs exist.
```

## Related Files

- `forge/agents/task-executor.md` — task-executor agent (Step 3 quality gate)
- `forge/commands/run-tasks.md` — /run-tasks dispatcher (dispatch loop)
- `forge/hooks/guide.md` — Testing Lifecycle section (three-layer architecture)
- `forge/hooks/hooks.json` — Stop hook triggers `task all-completed`
