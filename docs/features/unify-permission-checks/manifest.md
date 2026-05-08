---
feature: "unify-permission-checks"
status: design
---

# Feature: unify-permission-checks

<!-- Status flow: prd → design → tasks → in-progress → completed -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 前后端统一权限码校验；保留 is_super_admin 列作为 permCodes 加载标志；移除 RequirePermission tier-1 短路、handler-level bypass、前端 isSuperAdmin |
| User Stories | prd/prd-user-stories.md | 7 个用户故事：自定义角色编辑子事项、变更状态、SuperAdmin 全码校验、跨团队访问、前端 isSuperAdmin 移除、PM 操作不受影响、自定义角色添加进度 |
| Tech Design | design/tech-design.md | RequirePermission 移除 tier-1 短路；TeamScopeMiddleware SuperAdmin 注入全部 29 码；handler/service 层 bypass 移除；seedPresetRoles 增加 29 码 |
| API Handbook | design/api-handbook.md | 响应体移除 isSuperAdmin；/me/permissions SuperAdmin 返回全部 29 码；授权统一走 permCodes |

## Traceability

| PRD Section | Design Section | UI Component | Tasks |
|-------------|----------------|--------------|-------|
| 5.1 Seed & Data Changes | Interface 6: seedPresetRoles + Interface 3: GetUserPermissions | — | |
| 5.2 Backend Middleware Changes | Interface 1: RequirePermission + Interface 2: TeamScopeMiddleware | — | |
| 5.3 Backend Handler Changes | Interface 4: TeamService, Interface 5: ProgressService | — | |
| 5.4 Backend Service Changes | Interface 4: TeamService — Simplified Signatures | — | |
| 5.5 Backend VO/DTO Changes | Cross-Layer Data Map | — | |
| 5.6 Frontend Changes | Cross-Layer Data Map + Frontend File Enumeration | — | |
| 5.7 Related Changes | Interface 6: seedPresetRoles | — | |
