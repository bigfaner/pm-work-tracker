---
iteration: 2
date: 2026-04-28
scorer: forge-doc-scorer
total: 72/100
---

# Eval Report — Proposal: 细化 user / role 权限粒度

> Rubric not found at configured path. Standard proposal rubric applied (6 dimensions, 100 pts).
> Scored independently. No credit given for improvement from iteration 1 — only what is on the page now.

---

## Dimension Scores

### 1. Problem Statement — 14/20

**What works:** Two concrete problems are named with specific permission codes. The coupling issue is clearly articulated. The consequence is stated: "要么给用户过多权限…要么功能受限".

**Deductions:**

- **-3 — No evidence of real impact.** Still no mention of who is blocked, how often, or what workaround is currently in use. "目前无法给普通用户授予'只读角色列表'的权限" is a code-level observation, not a user impact statement.

- **-2 — No business/user framing.** Both problems are described from an engineer's perspective. There is no sentence explaining what a PM, member, or admin actually cannot do today as a result of this limitation.

- **-1 — No root cause analysis.** The proposal describes symptoms but not why the original design bundled these permissions together. Understanding the original intent would strengthen the case for splitting.

---

### 2. Proposed Solution — 20/25

**What works:** The `user:read` semantic collision is now explicitly called out with a detailed two-step rollout plan, CI grep assertion, and rollback description. The data migration mapping now includes a justification for the 1-to-3 expansion. The `role:*` resource split is well-motivated.

**Deductions:**

- **-3 — No API endpoint mapping.** "路由中间件：将原来绑定 `user:manage_role` 的角色管理接口改为对应的 `role:*` 权限码" names the change category but never lists which specific endpoints are affected. Which route currently requires `user:manage_role`? Which will require `role:read` vs `role:update`? Without this, the backend change scope is unverifiable and the reviewer cannot assess completeness.

- **-1 — Two-step rollout is not reflected in scope or timeline.** The solution section describes a two-step deployment process with a deprecation window and CI gate between steps, but the Scope section makes no mention of this phasing. There is no indication of how long the deprecation window lasts or what triggers the step-2 gate beyond "CI grep 断言确认零残留".

- **-1 — Data migration manual fallback is fragile.** "若存在仅需只读角色列表的自定义角色，管理员应在迁移后手动移除多余的写权限" is a human process with no enforcement mechanism. There is no proposed audit query to identify such roles before migration, and no post-migration verification step to confirm the manual cleanup happened.

---

### 3. Alternatives Considered — 8/15

**What works:** A comparison table exists with three options and explicit verdicts.

**Deductions:**

- **-4 — Alternatives B and C are strawmen.** Method C ("保持现状，用文档说明") is obviously unacceptable and adds no analytical value. Method B is dismissed with "长期不清晰" — a vague assertion, not a reasoned argument. Neither alternative was seriously stress-tested against the chosen approach.

- **-2 — No consideration of phased rollout as an alternative.** A natural alternative is: add the new codes first (non-breaking), migrate consumers, then deprecate old codes. This avoids the hard cutover risk entirely. The proposal actually describes a two-step process in the solution section, but never frames it as an alternative to a single-step migration.

- **-1 — No consideration of backward compatibility shim.** Keeping old codes as aliases during a transition window is a standard pattern for permission refactors. Its absence from the alternatives table is a gap, especially given the acknowledged `user:read` collision risk.

---

### 4. Scope Definition — 10/15

**What works:** In/out scope lists are present and cover the main surfaces.

**Deductions:**

- **-2 — "现有接口已满足需求" is an unsubstantiated claim.** The out-of-scope item "新增权限管理相关的 API 接口（现有接口已满足需求）" asserts sufficiency without evidence. If `role:read` is a new resource, does the existing role-list endpoint already enforce it? This needs verification, not assertion.

- **-2 — No mention of documentation or permission reference updates.** If permission codes change, any internal docs, onboarding guides, or API references that list permission codes are now stale. This is a real deliverable absent from scope.

- **-1 — Test coverage scope is unspecified.** The proposal mentions updating `VerifyPresetRoleCodes` in the risk table but does not include test updates in the in-scope list. This is an inconsistency — if it's a risk, it's in scope.

---

### 5. Risk Assessment — 12/15

**What works:** The `user:read` semantic collision is now the first risk, rated High/High, with a two-step mitigation, CI grep gate, and a concrete rollback description ("将 DB 中所有 `user:read` 恢复为 `user:list`，并回退代码至步骤一状态"). Risk 2 mitigation is now automated via CI grep rather than manual search. Risk 4 (`VerifyPresetRoleCodes` test failure) is a new and appropriate addition.

**Deductions:**

- **-1 — Cache/token invalidation is not mentioned.** If permissions are cached in JWT tokens or a session store, the migration will not take effect until tokens expire or are invalidated. A user holding a cached role with old `user:read` will continue to exercise the old (list) semantics until their token refreshes. This is a real deployment-window risk that is absent from the table.

- **-1 — Access gap during two-step deployment is unaddressed.** Step 1 migrates all `user:read` → `user:list`. Step 2 grants new `user:read` to roles that need it. Between steps, any user who previously relied on `user:read` to access user detail pages will lose that access. The risk table does not acknowledge this window or propose a mitigation (e.g., grant new `user:read` atomically in step 1 for roles that already held old `user:read`).

- **-1 — `user:assign_role` rename is not listed as a risk.** The rename of `user:manage_role` → `user:assign_role` is also a breaking change for any code path checking the old code. The CI grep covers it operationally, but it is not called out as an explicit risk with its own likelihood/impact rating.

---

### 6. Success Criteria — 8/10

**What works:** "现有自定义角色的权限在迁移后语义等价" is now present and directly addresses the migration correctness gap from iteration 1. Six criteria cover the main happy paths.

**Deductions:**

- **-1 — "所有相关测试通过" is not a criterion, it's a baseline.** Tests passing is a precondition for shipping, not a success criterion. Replace with something measurable: e.g., "新增 N 个针对 `role:*` 权限码的路由中间件测试，覆盖 read/create/update/delete 四个操作".

- **-1 — No criterion for the two-step deployment gate.** Given the explicit two-step rollout described in the solution, there should be a criterion: "步骤一上线后，CI grep 断言确认零残留旧语义 `user:read` 引用，方可触发步骤二". Without this, the two-step process has no verifiable completion condition in the success criteria.

---

## Summary

The proposal made targeted, effective improvements to the two highest-severity gaps from iteration 1: the `user:read` semantic collision is now explicitly flagged and mitigated, and the data migration mapping is now justified. Risk assessment and success criteria both improved meaningfully.

What remains weak: the Alternatives section is unchanged and still evaluates strawmen rather than real tradeoffs. The solution still lacks API endpoint mapping, making the backend change scope unverifiable. Two new risks surfaced in this review — cache/token invalidation and the access gap between deployment steps — that are not in the risk table. The scope section still asserts API sufficiency without evidence and omits documentation updates as a deliverable.
