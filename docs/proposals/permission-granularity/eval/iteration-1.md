---
iteration: 1
date: 2026-04-28
scorer: forge-doc-scorer
total: 62/100
---

# Eval Report — Proposal: 细化 user / role 权限粒度

> Rubric not found at configured path. Standard proposal rubric applied (6 dimensions, 100 pts).

---

## Dimension Scores

### 1. Problem Statement — 14/20

**What works:** Two concrete problems are named with specific permission codes. The coupling issue is clearly articulated.

**Deductions:**

- **-3 — No evidence of real impact.** The problem is stated as a code-level observation ("目前无法给普通用户授予'只读角色列表'的权限"), but there is no mention of who is blocked, how often, or what workaround is currently in use. A proposal should establish that the pain is real and felt.

- **-2 — No business/user framing.** Both problems are described from an engineer's perspective. There is no sentence explaining what a PM, member, or admin actually cannot do today as a result.

- **-1 — No root cause analysis.** The proposal describes symptoms but not why the original design bundled these permissions together. Understanding the original intent would strengthen the case for splitting.

---

### 2. Proposed Solution — 16/25

**What works:** The old→new permission code mapping table is clear. The `role:*` resource split is well-motivated.

**Deductions:**

- **-4 — Semantic collision on `user:read` is a critical unaddressed risk.** The proposal renames old `user:read` → `user:list`, then reuses the name `user:read` for a new, different permission ("查看用户详情"). Any existing code, token, or cached permission set that holds `user:read` will silently mean something different after migration. This is the most dangerous part of the design and receives zero discussion.

- **-3 — No API endpoint mapping.** The proposal lists permission codes but never maps them to actual routes. Which endpoint currently requires `user:manage_role`? Which will require `role:read` vs `role:update`? Without this, the backend change scope (item 3) is unverifiable.

- **-2 — Data migration logic is underspecified.** "将现有角色中的 `user:manage_role` 替换为 `role:create`、`role:update`、`role:delete`" — this is a 1-to-3 expansion. The proposal does not justify why every holder of `user:manage_role` should receive all three new codes. That is a privilege escalation assumption, not a neutral migration.

---

### 3. Alternatives Considered — 8/15

**What works:** A comparison table exists with three options.

**Deductions:**

- **-4 — Alternatives B and C are strawmen.** Method C ("保持现状") is obviously unacceptable and adds no analytical value. Method B is dismissed with "长期不清晰" — a vague assertion, not a reasoned argument. Neither alternative was seriously stress-tested.

- **-2 — No consideration of phased rollout.** A natural alternative is: add the new codes first (non-breaking), migrate consumers, then deprecate old codes. This avoids the hard cutover risk entirely and is not mentioned.

- **-1 — No consideration of backward compatibility shim.** Keeping old codes as aliases during a transition window is a standard pattern for permission refactors. Its absence from the alternatives table is a gap.

---

### 4. Scope Definition — 10/15

**What works:** In/out scope lists are present and cover the main surfaces.

**Deductions:**

- **-2 — "现有接口已满足需求" is an unsubstantiated claim.** The out-of-scope item "新增权限管理相关的 API 接口（现有接口已满足需求）" asserts sufficiency without evidence. If `role:read` is a new resource, does the existing role-list endpoint already enforce it? This needs verification, not assertion.

- **-2 — No mention of documentation or permission reference updates.** If permission codes change, any internal docs, onboarding guides, or API references that list permission codes are now stale. This is a real deliverable that is absent from scope.

- **-1 — Test coverage scope is unspecified.** The proposal mentions updating `VerifyPresetRoleCodes` in the risk table but does not include test updates in the in-scope list. This is an inconsistency.

---

### 5. Risk Assessment — 8/15

**What works:** Risk table format is clean. Three risks are identified with likelihood/impact ratings.

**Deductions:**

- **-4 — The highest-impact risk is missing entirely.** The `user:read` semantic rename (old meaning: list users; new meaning: view user detail) is a silent breaking change. Any code path, frontend guard, or cached role that checks `user:read` will now enforce a different permission than intended. This is High likelihood, High impact, and is not in the risk table at all.

- **-2 — Risk 1 mitigation is vague.** "迁移前备份，提供回滚脚本" — what does the rollback script do? Does it restore the old permission codes in the DB? Does it revert the code? A mitigation that says "provide rollback" without describing it is not actionable.

- **-1 — Risk 2 mitigation is manual and fragile.** "全局搜索 `user:manage_role` / `user:read` 引用，逐一替换" is a human process. No automated lint rule, CI check, or grep-in-CI is proposed to enforce completeness.

---

### 6. Success Criteria — 6/10

**What works:** Six criteria are listed covering the main happy paths.

**Deductions:**

- **-2 — No criterion covers the migration correctness.** The most important post-migration check — "existing custom roles retain semantically equivalent permissions after migration" — is absent. The criteria only check new behavior, not preservation of old behavior.

- **-1 — "所有相关测试通过" is not a criterion, it's a baseline.** Tests passing is a precondition for shipping, not a success criterion. Replace with something measurable: e.g., "新增 N 个针对 role:* 权限码的路由中间件测试".

- **-1 — No criterion for the `user:read` rename correctness.** Given the semantic collision risk, there should be an explicit criterion: "no existing code path checks `user:read` with the old (list) semantics after migration".

---

## Summary

The proposal is coherent and solves a real problem, but it underestimates the complexity of its own migration. The `user:read` rename is a silent semantic breaking change that is neither flagged as a risk nor addressed in success criteria. The data migration logic makes an unjustified privilege-expansion assumption. Alternatives are not seriously evaluated. The document reads like an internal note rather than a reviewable proposal — it needs API-level specificity, a harder look at migration safety, and measurable success criteria.
