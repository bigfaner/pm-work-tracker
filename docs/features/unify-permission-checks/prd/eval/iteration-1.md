---
date: "2026-05-08"
doc_dir: "docs/features/unify-permission-checks/prd/"
iteration: 1
target_score: 90
scoring_mode: "Mode B (no UI)"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 1

**Score: 86/100** (target: 90, mode: Mode B (no UI))

```
┌─────────────────────────────────────────────────────────────────┐
│                       PRD QUALITY SCORECARD                      │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Background & Goals        │  15      │  15      │ ✅         │
│    Three elements            │  5/5     │          │            │
│    Goals quantified          │  4/4     │          │            │
│    Logical consistency       │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Flow Diagrams             │  18      │  20      │ ⚠️         │
│    Mermaid diagram exists    │  7/7     │          │            │
│    Main path complete        │  7/7     │          │            │
│    Decision + error branches │  4/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3b. Flow Completeness (B)    │  17      │  20      │ ⚠️         │
│    Flow steps complete       │  7/7     │          │            │
│    Data flow documented      │  7/7     │          │            │
│    Exception handling        │  3/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  23      │  30      │ ⚠️         │
│    Coverage per user type    │  5/7     │          │            │
│    Format correct            │  7/7     │          │            │
│    AC per story (G/W/T)      │  6/6     │          │            │
│    AC verifiability          │  5/10    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Scope Clarity             │  13      │  15      │ ⚠️         │
│    In-scope concrete         │  5/5     │          │            │
│    Out-of-scope explicit     │  4/4     │          │            │
│    Consistent with specs     │  4/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  86      │  100     │ ⚠️         │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

> **Mode B** (prd-ui-functions.md absent): Dimension 3 evaluates Flow Completeness from prd-spec.md Flow Description.
> Sub-criteria: Flow steps describe complete business process /7, Data flow documented /7, Exception handling and edge cases /6.

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| prd-spec.md: Flow Diagram (mermaid) | Missing error branch for AuthMiddleware failure (unauthenticated request) and TeamScopeMiddleware failure (non-member, non-superadmin) | -2 pts (Dim 2) |
| prd-spec.md: Flow Description | No documented error paths for migration failure scenarios, concurrent requests during deployment, or rollback behavior in the flow description | -3 pts (Dim 3) |
| prd-user-stories.md | No story covering progress_handler bypass removal despite it being in scope (prd-spec.md 5.3 #2 explicitly lists progress_handler) | -2 pts (Dim 4) |
| prd-user-stories.md | Story 3 AC is a single monolithic Given/When/Then covering 14 operations -- not individually verifiable | -2 pts (Dim 4) |
| prd-user-stories.md | Zero error-path ACs across all 5 stories -- no AC verifies what happens when permission is correctly denied after removal | -3 pts (Dim 4) |
| prd-spec.md + prd-user-stories.md | Scope lists progress_handler and team_service PM identity changes, but no user story covers these behavioral scenarios | -2 pts (Dim 5) |

---

## Attack Points

### Attack 1: User Stories — Zero error-path ACs, missing coverage for progress_handler

**Where**: prd-user-stories.md -- all 5 stories have only happy-path ACs. Also, `prd-spec.md` line 141: "handler/progress_handler.go: 移除 pmFlag 参数和 isPMOrSuperAdmin() 调用" has no corresponding user story.

**Why it's weak**: Every story's "Then" clause describes only success (200, "编辑成功", "编译通过"). There is not a single AC that verifies permission denial still works correctly after bypass removal. This is a critical regression risk -- removing bypass logic could accidentally grant or deny access. Additionally, the progress_handler bypass removal (explicitly listed in Functional Specs 5.3 #2) has no user story, leaving a behavioral gap. The scope-scope mismatch (progress_handler in scope but no story) indicates incomplete story coverage.

**What must improve**: (1) Add at least 2 error-path ACs: one verifying a user without `sub_item:update` still gets 403, and one verifying a non-superadmin without the required code gets 403 on team operations. (2) Add a user story for progress_handler bypass removal (e.g., "As a custom role user with `progress:create`, I want to add progress to any sub-item without being blocked by PM checks"). (3) Break Story 3's monolithic AC into at least 2-3 specific ACs covering distinct operation categories.

### Attack 2: Flow Diagrams — Missing error branches for middleware failures

**Where**: prd-spec.md mermaid diagram (lines 95-117) -- only one error node (`Forbidden[返回 403 FORBIDDEN]`), no branch for authentication failure or team-scope failure.

**Why it's weak**: The diagram shows `Start -> Auth -> TeamScope -> PermCheck -> Handler -> Success`, but only PermCheck has an error branch. AuthMiddleware can fail (invalid/expired token -> 401). TeamScopeMiddleware can fail for a non-superadmin non-member (-> 403). These are real error paths that users will encounter. A flow diagram in a PRD should depict all significant outcomes, not just the permission-check failure. The absence of these branches understates the system's actual error surface.

**What must improve**: Add two more error branches to the mermaid diagram: (1) After Auth node, add a decision for auth validity -> 401 Unauthorized; (2) After TeamScope node for non-superadmin path, add a decision for team membership -> 403 if not a member. This gives the full picture of all rejection points.

### Attack 3: Flow Completeness — Exception handling and edge cases insufficiently documented

**Where**: prd-spec.md "Flow Description" section (lines 69-117) and "Other Notes" (lines 184-205). The only error handling mentioned is the 403 in the permission check. The Data Requirements section says "Migration 必须是事务性的" but no flow or fallback describes what happens on failure.

**Why it's weak**: The "current flow" and "target flow" descriptions are purely linear -- they describe the happy path in prose and only mention the 403 at the permission check. Critical edge cases are unaddressed: (1) What happens during the migration window if a request arrives while `is_super_admin` column has been dropped but team_members records haven't been inserted yet? (2) What if the migration fails mid-transaction -- is there a rollback strategy documented? (3) What happens if an existing SuperAdmin user has no team_members record after migration (data integrity check)? The "Monitoring Requirements" section says "无新增监控需求" -- but removing a core authorization path absolutely warrants monitoring for 403 spikes, especially during rollout.

**What must improve**: (1) Add a "Failure Scenarios" subsection to Flow Description covering: migration rollback, deployment ordering (backend before frontend or vice versa), and the degraded state if migration partially fails. (2) Add an edge case: what if a user was SuperAdmin but is not found during migration (orphan user). (3) Revise monitoring requirements to include a baseline 403-rate check during rollout, since removing bypass logic is a high-risk authorization change.

---

## Verdict

- **Score**: 86/100
- **Target**: 90/100
- **Gap**: 4 points
- **Action**: Continue to iteration 2
