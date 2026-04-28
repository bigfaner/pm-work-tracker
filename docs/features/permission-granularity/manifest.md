---
feature: "permission-granularity"
status: prd
---

# Feature: permission-granularity

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 将 user 资源权限码拆分为 user:list/read/update/assign_role，新增 role 资源（read/create/update/delete），更新路由绑定、预置角色和数据迁移方案 |
| User Stories | prd/prd-user-stories.md | 6 个用户故事，覆盖 member 查角色列表、自定义角色用户使用成员选择器、pm 管理角色/用户、管理员执行迁移、前端权限控制 |
| UI Functions | prd/prd-ui-functions.md | 3 个 UI 功能点：角色下拉选择器、用户管理页入口、角色管理页操作按钮权限控制 |

## Traceability

| PRD Section | Design Section | UI Component | Tasks |
|-------------|----------------|--------------|-------|
| 权限码注册表变更 | — | — | — |
| 路由中间件绑定 | — | — | — |
| 预置角色权限调整 | — | — | — |
| 数据迁移（两步上线） | — | — | — |
| 角色下拉选择器 | — | 团队管理页 AddMemberDialog | — |
| 用户管理页入口 | — | 导航菜单 | — |
| 角色管理页按钮 | — | RoleManagementPage | — |
