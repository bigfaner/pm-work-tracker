---
iteration: 5
date: 2026-04-28
scorer: forge-doc-scorer
total: 92/100
---

# Eval Report — Proposal: 细化 user / role 权限粒度

> Rubric not found at configured path. Standard proposal rubric applied (6 dimensions, 100 pts).
> Scored independently. No credit given for improvement from iteration 4 — only what is on the page now.

---

## Dimension Scores

### 1. Problem Statement — 18/20

**What works:** Both problems now have named users, specific broken flows, described workarounds, and stated consequences. Problem 2 is now fully parallel to Problem 1: the `project_manager` role is named, the task-assignment member selector is the specific flow, the workaround (grant `user:read` + hide frontend entry) is described, and the security consequence is explicit: "后端接口仍可直接访问，权限控制形同虚设". This closes the asymmetry flagged in iteration 4.

**Deductions:**

- **-1 — No root cause analysis.** The proposal still describes symptoms without explaining why the original design bundled these permissions together. Understanding the original intent would strengthen the case for splitting and reduce the risk of repeating the same mistake in future permission design.

- **-1 — No frequency or scale data.** Both workarounds are now described, but there is no indication of how often these scenarios occur or how many users are affected. A reader cannot assess business priority from this section alone.

---

### 2. Proposed Solution — 23/25

**What works:** The `user:read` semantic collision is explicitly called out with a detailed two-step rollout plan, CI grep assertion, and rollback description. The data migration mapping includes a justification for the 1-to-3 expansion. The `role:*` resource split is well-motivated. The API endpoint mapping table covers all affected routes with old→new permission code pairs.

**Deductions:**

- **-1 — Two-step rollout is not reflected in scope or timeline.** The solution section describes a two-step deployment process with a deprecation window and CI gate between steps, but the Scope section makes no mention of this phasing. There is no indication of how long the deprecation window lasts or what triggers the step-2 gate beyond "CI grep 断言确认零残留".

- **-1 — Data migration manual fallback is fragile.** "若存在仅需只读角色列表的自定义角色，管理员应在迁移后手动移除多余的写权限，或在迁移前确认" is a human process with no enforcement mechanism. There is no proposed audit query to identify such roles before migration, and no post-migration verification step to confirm the manual cleanup happened.

---

### 3. Alternatives Considered — 14/15

**What works:** Method C has been replaced with a genuinely considered alternative: backward-compatible aliases for `user:read`. This is a real engineering approach, not a strawman. The rejection is substantive: "别名本质上是'一个码授予两种权限'，这与本次拆分的目标相悖" and correctly identifies that the two-step rollout already solves the 403 problem Method C was trying to address, making the alias mechanism redundant. Method D's rejection correctly notes that the two-step process in Method A already internalizes the core idea.

**Deductions:**

- **-1 — Method B's rejection relies on a speculative future scenario.** "若未来需要'可创建角色但不可管理用户'的权限组合，仍无法表达" is a forward-looking argument, not a current concrete failure. The rejection would be stronger if grounded in a present limitation rather than a hypothetical future need.

---

### 4. Scope Definition — 14/15

**What works:** In/out scope lists are present and cover the main surfaces. The out-of-scope item for API interfaces lists the 6 specific endpoints, making the sufficiency claim verifiable. Documentation updates are now explicitly in scope: "内部文档更新：`docs/` 中所有列出权限码的文档（架构说明、API 参考、onboarding 指南）同步替换为新权限码" — this closes the gap from iteration 4.

**Deductions:**

- **-1 — Test coverage scope is unspecified.** The proposal mentions updating `VerifyPresetRoleCodes` in the risk table but does not include test updates in the in-scope list. If it's a risk, it's in scope — this inconsistency should be resolved.

---

### 5. Risk Assessment — 14/15

**What works:** Six risks are present and well-handled. Risk 6 (两步上线期间用户失去用户详情页访问权) correctly identifies the 403 window between step 1 and step 2, rates it Medium/Medium, and proposes a concrete atomic mitigation. Risk 5 (JWT token caching) correctly identifies that the current implementation has no token-cache risk while flagging the future risk if a cache layer is introduced.

**Deductions:**

- **-1 — `user:manage_role` → `user:assign_role` rename is not an explicit risk entry.** Risk 3 covers frontend reference updates via CI grep, which operationally catches this. But the rename is a breaking change for any code path checking the old code, and it is not rated with its own likelihood/impact or described with a specific rollback path distinct from the broader CI grep mitigation.

---

### 6. Success Criteria — 9/10

**What works:** Six criteria cover the main happy paths. "现有自定义角色的权限在迁移后语义等价" directly addresses migration correctness. "后端路由中间件全部使用新权限码，无残留 `user:manage_role` 引用" is verifiable. The test criterion is specific and measurable: "新增至少 4 个针对 `role:*` 权限码的路由中间件测试，分别覆盖 `role:read`、`role:create`、`role:update`、`role:delete` 四个操作，验证无对应权限时返回 403".

**Deductions:**

- **-1 — No criterion for the two-step deployment gate.** Given the explicit two-step rollout described in the solution, there should be a criterion: "步骤一上线后，CI grep 断言确认零残留旧语义 `user:read` 引用，方可触发步骤二". Without this, the two-step process has no verifiable completion condition in the success criteria.

---

## Summary

Three targeted improvements were made this iteration: Problem 2 now has a named user, specific scenario, workaround, and security consequence — closing the asymmetry with Problem 1; Method C was replaced with a genuine backward-compatible alias alternative with a substantive rejection; and documentation updates were added explicitly to scope.

What remains weak: the two-step deployment gate still has no corresponding success criterion; test coverage is absent from the in-scope list despite being flagged as a risk; Method B's rejection is still grounded in a speculative future scenario rather than a present limitation; and neither problem has frequency or scale data to support business prioritization.
