---
feature: "bizkey-unification"
status: tasks
---

# Feature: bizkey-unification

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 消除 Service 层 uint/int64 ID 混用问题，修复进度记录 team_key 数据污染，统一以 int64 bizKey 作为跨层边界标识符 |
| User Stories | prd/prd-user-stories.md | 3 个故事：进度记录写入正确 team_key、角色权限判断使用正确 bizKey、编译器约束防止同类 Bug |
| Tech Design | design/tech-design.md | Middleware 注入 int64 bizKey；8 个 Service 接口 teamID 参数改为 int64；修复 3 处具体 Bug；无 API/Schema 变更 |

## Traceability

| PRD Section | Design Section | UI Component | Tasks |
|-------------|----------------|--------------|-------|
| 5.4 rows 1 (middleware) | Interfaces §1 (GetTeamBizKey) | — | 1.1 |
| 5.4 rows 5-8 (ProgressService) | Interfaces §2 (ProgressService) | — | 1.2 |
| 5.4 rows 2-4 (TeamService) | Interfaces §3 (TeamService) | — | 1.3 |
| 5.4 rows 11,14 (ViewService, ReportService) | Interfaces §4-5 | — | 1.4 |
| 5.4 rows 9,10,13 (MainItem, SubItem, ItemPool) | Interfaces §6 | — | 1.5 |
| 5.4 rows 15-21 (Handler call sites) | Interfaces §7 | — | 2.1 |
| 5.5 bizKey 校验规则 | Middleware §1 (TeamScopeMiddleware) | — | 1.1 |
| User Stories AC (all 3) | Testing Strategy | — | 2.2, 2.gate |
