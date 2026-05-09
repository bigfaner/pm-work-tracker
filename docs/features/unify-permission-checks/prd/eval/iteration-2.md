---
date: "2026-05-08"
doc_dir: "docs/features/unify-permission-checks/prd/"
iteration: 2
target_score: 90
scoring_mode: "Mode B (no UI)"
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 2

**Score: 95/100** (target: 90, mode: Mode B (no UI))

```
┌─────────────────────────────────────────────────────────────────┐
│                       PRD QUALITY SCORECARD                      │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 1. Background & Goals        │  15      │  15      │ ✅         │
│    Three elements            │  5/5     │          │            │
│    Goals quantified          │  4/4     │          │            │
│    Logical consistency       │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 2. Flow Diagrams             │  20      │  20      │ ✅         │
│    Mermaid diagram exists    │  7/7     │          │            │
│    Main path complete        │  7/7     │          │            │
│    Decision + error branches │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 3b. Flow Completeness (B)    │  20      │  20      │ ✅         │
│    Flow steps complete       │  7/7     │          │            │
│    Data flow documented      │  7/7     │          │            │
│    Exception handling        │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 4. User Stories              │  26      │  30      │ ⚠️         │
│    Coverage per user type    │  6/7     │          │            │
│    Format correct            │  7/7     │          │            │
│    AC per story (G/W/T)      │  6/6     │          │            │
│    AC verifiability          │  7/10    │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ 5. Scope Clarity             │  14      │  15      │ ⚠️         │
│    In-scope concrete         │  5/5     │          │            │
│    Out-of-scope explicit     │  4/4     │          │            │
│    Consistent with specs     │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┬────────────┤
│ TOTAL                        │  95      │  100     │ ✅         │
└─────────────────────────────────────────────────────────────────┘
```

> **Mode B** (prd-ui-functions.md absent): Dimension 3 evaluates Flow Completeness from prd-spec.md Flow Description.
> Sub-criteria: Flow steps describe complete business process /7, Data flow documented /7, Exception handling and edge cases /6.

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| prd-user-stories.md: Background Users table | PM role is listed in Background with team_handler PM-identity-to-permission-code conversion affecting it, but no story verifies PM operations work post-change (PM validation logic is fundamentally altered in team_service) | -1 pt (Dim 4 coverage) |
| prd-user-stories.md: Story 3 ACs | AC 3a/3b/3c "Then 每项操作均返回成功，等同于移除 IsSuperAdmin 字段前" uses comparative language referencing pre-change state rather than explicit expected outcomes; requires baseline knowledge to verify | -1 pt (Dim 4 verifiability) |
| prd-user-stories.md: Story 5 | Only AC is build-time check ("编译通过且所有测试通过") -- does not verify runtime behavior that SuperAdmin user actually sees correct UI elements after isSuperAdmin removal | -1 pt (Dim 4 verifiability) |
| prd-user-stories.md + prd-spec.md | DB migration (scope item: "DB migration：删除 users.is_super_admin 列") has detailed Failure Scenarios in prd-spec.md but no user story covers the migration transition behavior for existing SuperAdmin users | -1 pt (Dim 4 verifiability) |
| prd-spec.md + prd-user-stories.md | team_service.go changes (5.4 #1: "PM 身份校验改为权限码校验") affects PM behavior but no story explicitly verifies PM role operations still function correctly after the identity-check-to-permission-code conversion | -1 pt (Dim 5 consistency) |

---

## Attack Points

### Attack 1: User Stories -- Story 5 AC is a build-time check, not behavioral verification

**Where**: prd-user-stories.md Story 5: "Given 代码库中不再存在 isSuperAdmin 引用 / When TypeScript 编译和所有测试运行 / Then 编译通过且所有测试通过"

**Why it's weak**: This AC only verifies that code compiles and tests pass after removing `isSuperAdmin`. It does not verify any runtime behavior: does a SuperAdmin user see the correct permission-controlled UI elements in the browser? Does the PermissionGuard correctly gate visibility for all 29 permission codes that replaced the single boolean? A refactoring that removes all `isSuperAdmin` references but accidentally breaks PermissionGuard logic would pass this AC. The "I want" statement says "UI 可见性由权限码驱动" but the AC never verifies that UI visibility actually works -- only that the code compiles.

**What must improve**: Add at least one runtime AC: e.g., "Given a SuperAdmin user with all 29 permission codes / When they view a team page / Then all admin-level UI elements (disband team, transfer PM, etc.) are visible" and/or "Given a custom role user with only sub_item:view / When they view a sub-item / Then the edit button is hidden." This bridges the gap between "code compiles" and "feature works."

### Attack 2: User Stories -- PM role affected by team_service identity-check change but no story covers it

**Where**: prd-spec.md 5.4 #1: "service/team_service.go: ListTeams 移除 isSuperAdmin 参数；PM 身份校验改为权限码校验（team:update 等）" -- no corresponding user story. prd-spec.md Background Users table lists PM with "无变化" but the PM identity validation in team_service is being fundamentally altered from identity-based to permission-code-based.

**Why it's weak**: The scope says "PM 身份校验改为权限码校验" which means PM users' team management operations will now be validated through a different mechanism. The Background claims "无变化" for PMs, but this claim is unverified. A PM who has `team:update` by role assignment should see no change, but if the permission-code-check implementation differs from the PM-identity-check in any edge case (e.g., a PM who was transferred PM role but whose permission codes weren't updated), the PM would experience a regression. No story or AC catches this scenario. Additionally, the user type coverage is incomplete: 4 user types in Background, stories cover 3 (custom role, SuperAdmin, frontend dev), PM is the gap.

**What must improve**: Either (a) add a story: "As a PM, I want to continue managing my team (update, disband, transfer PM) after the PM identity check is replaced by permission-code check, so that I experience no functional regression" with ACs covering team management operations, or (b) if PM truly has no behavior change, explicitly state in the Background why PM is not affected and remove PM from the Users table to avoid the coverage expectation.

### Attack 3: User Stories -- Story 3 ACs use comparative language instead of explicit expected outcomes

**Where**: prd-user-stories.md Story 3 AC 3a: "Then 每项操作均返回成功，等同于移除 IsSuperAdmin 字段前". Same pattern in AC 3b and 3c.

**Why it's weak**: The "Then" clause says "equivalent to before IsSuperAdmin field removal" which is a comparative assertion rather than an explicit, self-contained expected outcome. To verify this AC, a tester needs to know exactly what the pre-removal behavior was for all 14+ operations. This creates an implicit dependency on baseline knowledge that isn't documented in the AC itself. A stronger AC would state: "Then 返回 200 OK for each of the following operations: 创建团队 (201), 更新团队 (200), 解散团队 (200), ..." with explicit expected status codes. The "14 类操作" from the Goals section could be enumerated to make the AC self-verifying.

**What must improve**: Rewrite Story 3 ACs to list explicit expected outcomes instead of "等同于移除前". For AC 3a: enumerate the specific team operations and their expected status codes (e.g., "创建团队返回 201, 更新团队返回 200, 解散团队返回 200, 移除成员返回 200, 修改角色返回 200, 转让PM返回 200"). This makes the AC independently verifiable without requiring baseline knowledge of the pre-change system.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Zero error-path ACs, missing progress_handler story | ✅ | Stories 1, 2, 4, 6 now all have error-path ACs (403 FORBIDDEN for missing permission codes). Story 6 was added for progress_handler bypass removal. Story 3 ACs were decomposed from 1 monolithic AC into 3 categorical ACs (3a/3b/3c). |
| Attack 2: Missing error branches in flow diagram (auth failure, team-scope failure) | ✅ | Mermaid diagram now includes: AuthValid decision diamond with "No -> 返回 401 UNAUTHORIZED" branch, IsTeamMember decision diamond with "No -> 返回 403 非团队成员" branch. Three error paths total. |
| Attack 3: Exception handling insufficiently documented in Flow Description | ✅ | "Failure Scenarios" table added with 3 detailed scenarios: migration partial failure (transaction rollback), concurrent requests during deployment (ordering strategy), orphan SuperAdmin users (validation + warning log). Monitoring requirements updated to include 403-rate baseline monitoring during rollout. |

---

## Verdict

- **Score**: 95/100
- **Target**: 90/100
- **Gap**: 0 points (target exceeded by 5)
- **Action**: Target reached. Iteration complete.
