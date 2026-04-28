---
iteration: 3
date: 2026-04-28
scorer: forge-doc-scorer
total: 80/100
---

# Eval Report — Proposal: 细化 user / role 权限粒度

> Rubric not found at configured path. Standard proposal rubric applied (6 dimensions, 100 pts).
> Scored independently. No credit given for improvement from iteration 2 — only what is on the page now.

---

## Dimension Scores

### 1. Problem Statement — 14/20

**What works:** Two concrete problems are named with specific permission codes. The coupling issue is clearly articulated. A specific user-facing scenario is given: "在团队管理页面添加成员时，需要查询角色列表（下拉选择），但这不应该要求用户拥有管理角色的权限". The consequence is stated: "要么给用户过多权限…要么功能受限".

**Deductions:**

- **-3 — No evidence of real impact.** "目前无法给普通用户授予'只读角色列表'的权限" is still a code-level observation. There is no mention of who is blocked, how often this scenario occurs, or what workaround is currently in use. A proposal should establish that the pain is real and felt by real users.

- **-2 — No business/user framing.** Both problems are described from an engineer's perspective. There is no sentence explaining what a PM, member, or admin actually cannot do today as a result of this limitation. The reader cannot assess business priority from this section alone.

- **-1 — No root cause analysis.** The proposal describes symptoms but not why the original design bundled these permissions together. Understanding the original intent would strengthen the case for splitting and reduce the risk of repeating the same mistake.

---

### 2. Proposed Solution — 23/25

**What works:** The `user:read` semantic collision is explicitly called out with a detailed two-step rollout plan, CI grep assertion, and rollback description. The data migration mapping includes a justification for the 1-to-3 expansion. The `role:*` resource split is well-motivated. The API endpoint mapping table is now present and covers all affected routes with old→new permission code pairs — this is a significant improvement that makes the backend change scope verifiable.

**Deductions:**

- **-1 — Two-step rollout is not reflected in scope or timeline.** The solution section describes a two-step deployment process with a deprecation window and CI gate between steps, but the Scope section makes no mention of this phasing. There is no indication of how long the deprecation window lasts or what triggers the step-2 gate beyond "CI grep 断言确认零残留".

- **-1 — Data migration manual fallback is fragile.** "若存在仅需只读角色列表的自定义角色，管理员应在迁移后手动移除多余的写权限，或在迁移前确认" is a human process with no enforcement mechanism. There is no proposed audit query to identify such roles before migration, and no post-migration verification step to confirm the manual cleanup happened.

---

### 3. Alternatives Considered — 12/15

**What works:** A fourth alternative (Method D: 分阶段上线) has been added and addresses the phased rollout gap from iteration 2. Method D's rejection is reasoned: it correctly notes that the two-step process in Method A already internalizes the core idea, and that keeping aliases for `user:manage_role` → `role:*` would add middleware complexity for limited gain. Method B's rejection is now more detailed with a concrete failure scenario. Method C's rejection now includes a specific functional blocking example.

**Deductions:**

- **-2 — Method C is still a strawman.** "保持现状，用文档说明" is obviously unacceptable. The added detail in the rejection ("文档无法修复缺失的权限码") makes the rejection more verbose but not more analytically valuable. Including an obviously-rejected option inflates the table without adding insight.

- **-1 — Method B's rejection relies on a speculative future scenario.** "若未来需要'可创建角色但不可管理用户'的权限组合，仍无法表达" is a forward-looking argument, not a current concrete failure. The rejection would be stronger if grounded in a present limitation rather than a hypothetical future need.

---

### 4. Scope Definition — 10/15

**What works:** In/out scope lists are present and cover the main surfaces. The five in-scope items map directly to the backend and frontend change sections.

**Deductions:**

- **-2 — "现有接口已满足需求" is an unsubstantiated claim.** The out-of-scope item "新增权限管理相关的 API 接口（现有接口已满足需求）" asserts sufficiency without evidence. If `role:read` is a new resource, does the existing role-list endpoint already enforce it? The API endpoint mapping table in the solution section implies existing routes are being re-bound, but this does not confirm that no new endpoints are needed. This needs verification, not assertion.

- **-2 — No mention of documentation or permission reference updates.** If permission codes change, any internal docs, onboarding guides, or API references that list permission codes are now stale. This is a real deliverable absent from scope.

- **-1 — Test coverage scope is unspecified.** The proposal mentions updating `VerifyPresetRoleCodes` in the risk table but does not include test updates in the in-scope list. If it's a risk, it's in scope — this inconsistency should be resolved.

---

### 5. Risk Assessment — 13/15

**What works:** Five risks are now present. The `user:read` semantic collision is Risk 1 with High/High rating, two-step mitigation, CI grep gate, and a concrete rollback description. Risk 5 (JWT token caching) is a new and well-handled addition — it correctly identifies that the current implementation loads permissions from DB on each request and therefore has no token-cache risk, while flagging the future risk if a cache layer is introduced.

**Deductions:**

- **-1 — Access gap during two-step deployment is unaddressed.** Step 1 migrates all `user:read` → `user:list`. Step 2 grants new `user:read` to roles that need it. Between steps, any user who previously relied on `user:read` to access user detail pages will lose that access. The solution section acknowledges "新 `user:read` 在第二步上线时按需授予，不自动继承" but does not frame this as a risk or propose mitigation (e.g., grant new `user:read` atomically in step 1 for roles that already held old `user:read`). This is absent from the risk table.

- **-1 — `user:manage_role` → `user:assign_role` rename is not an explicit risk entry.** Risk 3 covers frontend reference updates via CI grep, which operationally catches this. But the rename is a breaking change for any code path checking the old code, and it is not rated with its own likelihood/impact or described with a specific rollback path distinct from the broader CI grep mitigation.

---

### 6. Success Criteria — 8/10

**What works:** Six criteria cover the main happy paths. "现有自定义角色的权限在迁移后语义等价" directly addresses migration correctness. "后端路由中间件全部使用新权限码，无残留 `user:manage_role` 引用" is verifiable.

**Deductions:**

- **-1 — "所有相关测试通过" is not a criterion, it's a baseline.** Tests passing is a precondition for shipping, not a success criterion. Replace with something measurable: e.g., "新增 N 个针对 `role:*` 权限码的路由中间件测试，覆盖 read/create/update/delete 四个操作".

- **-1 — No criterion for the two-step deployment gate.** Given the explicit two-step rollout described in the solution, there should be a criterion: "步骤一上线后，CI grep 断言确认零残留旧语义 `user:read` 引用，方可触发步骤二". Without this, the two-step process has no verifiable completion condition in the success criteria.

---

## Summary

The proposal made one targeted, high-value improvement this iteration: the API endpoint mapping table is now present and makes the backend change scope verifiable — this was the highest-severity open gap from iteration 2. The JWT token caching risk is also now correctly handled with a nuanced explanation.

What remains weak: the Problem Statement, Scope, and Success Criteria sections are unchanged from iteration 2 and carry the same deductions. The access gap between deployment steps — users losing `user:read` detail-page access between step 1 and step 2 — is still not in the risk table despite being acknowledged in the solution text. The Alternatives section still includes a strawman (Method C) that adds no analytical value.
