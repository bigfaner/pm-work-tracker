---
iteration: 4
date: 2026-04-28
scorer: forge-doc-scorer
total: 87/100
---

# Eval Report — Proposal: 细化 user / role 权限粒度

> Rubric not found at configured path. Standard proposal rubric applied (6 dimensions, 100 pts).
> Scored independently. No credit given for improvement from iteration 3 — only what is on the page now.

---

## Dimension Scores

### 1. Problem Statement — 17/20

**What works:** Two concrete problems are named with specific permission codes. The new paragraph identifies the affected user ("持有 `member` 角色的团队成员"), the specific broken flow ("团队管理页面添加新成员，该流程依赖角色下拉列表查询"), and the current workaround with its security implication ("将 `user:manage_role` 临时授予 `member` 角色，这同时开放了创建、编辑、删除角色定义的权限——远超实际需要"). This is a meaningful improvement over prior iterations.

**Deductions:**

- **-1 — No root cause analysis.** The proposal describes symptoms but not why the original design bundled these permissions together. Understanding the original intent would strengthen the case for splitting and reduce the risk of repeating the same mistake in future permission design.

- **-1 — Problem 2 lacks a concrete user scenario.** Problem 1 (role list bundling) now has a named user, a specific flow, and a workaround. Problem 2 (`user:read` split) is still described abstractly: "有些场景只需要列出用户（如成员选择器），有些场景需要查看完整用户信息（如用户管理页）". No user role is named, no broken flow is described, no workaround is mentioned. The asymmetry weakens the overall problem statement.

- **-1 — No frequency or scale data.** The workaround is now described, but there is no indication of how often this scenario occurs or how many users are affected. A reader cannot assess business priority from this section alone.

---

### 2. Proposed Solution — 23/25

**What works:** The `user:read` semantic collision is explicitly called out with a detailed two-step rollout plan, CI grep assertion, and rollback description. The data migration mapping includes a justification for the 1-to-3 expansion. The `role:*` resource split is well-motivated. The API endpoint mapping table covers all affected routes with old→new permission code pairs.

**Deductions:**

- **-1 — Two-step rollout is not reflected in scope or timeline.** The solution section describes a two-step deployment process with a deprecation window and CI gate between steps, but the Scope section makes no mention of this phasing. There is no indication of how long the deprecation window lasts or what triggers the step-2 gate beyond "CI grep 断言确认零残留".

- **-1 — Data migration manual fallback is fragile.** "若存在仅需只读角色列表的自定义角色，管理员应在迁移后手动移除多余的写权限，或在迁移前确认" is a human process with no enforcement mechanism. There is no proposed audit query to identify such roles before migration, and no post-migration verification step to confirm the manual cleanup happened.

---

### 3. Alternatives Considered — 12/15

**What works:** Four alternatives are present. Method D's rejection is reasoned: it correctly notes that the two-step process in Method A already internalizes the core idea. Method B's rejection includes a concrete failure scenario. Method C's rejection includes a specific functional blocking example.

**Deductions:**

- **-2 — Method C is still a strawman.** "保持现状，用文档说明" is obviously unacceptable. The rejection ("文档无法修复缺失的权限码") makes it more verbose but not more analytically valuable. Including an obviously-rejected option inflates the table without adding insight. Replace with a genuinely considered alternative that was rejected for non-obvious reasons.

- **-1 — Method B's rejection relies on a speculative future scenario.** "若未来需要'可创建角色但不可管理用户'的权限组合，仍无法表达" is a forward-looking argument, not a current concrete failure. The rejection would be stronger if grounded in a present limitation rather than a hypothetical future need.

---

### 4. Scope Definition — 12/15

**What works:** In/out scope lists are present and cover the main surfaces. The out-of-scope item for API interfaces now lists the 6 specific endpoints ("现有 6 个接口已覆盖全部 role:* 操作：`GET /admin/roles`、`GET /admin/roles/:id`、`POST /admin/roles`、`PUT /admin/roles/:id`、`DELETE /admin/roles/:id`、`GET /admin/permissions`，本次变更仅重新绑定权限码，无需新增路由"), making the sufficiency claim verifiable.

**Deductions:**

- **-2 — No mention of documentation or permission reference updates.** If permission codes change, any internal docs, onboarding guides, or API references that list permission codes are now stale. This is a real deliverable absent from scope.

- **-1 — Test coverage scope is unspecified.** The proposal mentions updating `VerifyPresetRoleCodes` in the risk table but does not include test updates in the in-scope list. If it's a risk, it's in scope — this inconsistency should be resolved.

---

### 5. Risk Assessment — 14/15

**What works:** Six risks are now present. Risk 6 (两步上线期间用户失去用户详情页访问权) is a new and well-handled addition — it correctly identifies the 403 window between step 1 and step 2, rates it Medium/Medium, and proposes a concrete atomic mitigation: "在步骤一的数据迁移脚本中，对所有持有旧 `user:read` 的角色同步写入新 `user:read`，即原子授予，无需等到步骤二手动操作". This was the highest-severity open gap from iteration 3.

**Deductions:**

- **-1 — `user:manage_role` → `user:assign_role` rename is not an explicit risk entry.** Risk 3 covers frontend reference updates via CI grep, which operationally catches this. But the rename is a breaking change for any code path checking the old code, and it is not rated with its own likelihood/impact or described with a specific rollback path distinct from the broader CI grep mitigation.

---

### 6. Success Criteria — 9/10

**What works:** Six criteria cover the main happy paths. "现有自定义角色的权限在迁移后语义等价" directly addresses migration correctness. "后端路由中间件全部使用新权限码，无残留 `user:manage_role` 引用" is verifiable. The test criterion is now specific and measurable: "新增至少 4 个针对 `role:*` 权限码的路由中间件测试，分别覆盖 `role:read`、`role:create`、`role:update`、`role:delete` 四个操作，验证无对应权限时返回 403".

**Deductions:**

- **-1 — No criterion for the two-step deployment gate.** Given the explicit two-step rollout described in the solution, there should be a criterion: "步骤一上线后，CI grep 断言确认零残留旧语义 `user:read` 引用，方可触发步骤二". Without this, the two-step process has no verifiable completion condition in the success criteria.

---

## Summary

Four targeted improvements were made this iteration: the Problem Statement now identifies the affected user and current workaround; the Scope section now lists the 6 specific endpoints backing the API sufficiency claim; Risk 6 addresses the access gap between deployment steps with a concrete atomic mitigation; and the test success criterion is now specific and measurable.

What remains weak: the Alternatives section still includes a strawman (Method C) that adds no analytical value; Problem 2 still lacks a concrete user scenario comparable to Problem 1; documentation updates are absent from scope; and the two-step deployment gate has no corresponding success criterion.
