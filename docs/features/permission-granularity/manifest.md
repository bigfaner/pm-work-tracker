---
feature: "permission-granularity"
status: tasks
---

# Feature: permission-granularity

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 将 user 资源权限码拆分为 user:list/read/update/assign_role，新增 role 资源（read/create/update/delete），更新路由绑定、预置角色和数据迁移方案 |
| User Stories | prd/prd-user-stories.md | 6 个用户故事，覆盖 member 查角色列表、自定义角色用户使用成员选择器、pm 管理角色/用户、管理员执行迁移、前端权限控制 |
| UI Functions | prd/prd-ui-functions.md | 3 个 UI 功能点：角色下拉选择器、用户管理页入口、角色管理页操作按钮权限控制 |
| Tech Design | design/tech-design.md | 权限码注册表、路由绑定、数据迁移函数、前端守卫的具体变更方案，含 PRD 覆盖映射 |
| API Handbook | design/api-handbook.md | 14 条受影响接口的权限码变更前后对比及完整契约 |

## Traceability

| PRD Section | Design Section | UI Component | Tasks |
|-------------|----------------|--------------|-------|
| 权限码注册表变更 | tech-design §Interfaces 1 | — | 1.1 |
| 数据迁移（两步上线） | tech-design §Interfaces 3 | — | 1.2 |
| 路由中间件绑定 | tech-design §Interfaces 2 + api-handbook | — | 1.3, 1.4, 1.5 |
| 预置角色权限调整 | tech-design §Interfaces 4 | — | 1.2 |
| 角色下拉选择器 | tech-design §Interfaces 8 | TeamManagementPage | 2.1, 2.3 |
| 用户管理页入口 | tech-design §Interfaces 6+7 | App.tsx + Sidebar.tsx | 2.1, 2.2 |
| 角色管理页按钮 | tech-design §Interfaces 9 | RoleManagementPage | 2.1, 2.4 |
| 前端权限测试 | tech-design §Testing Strategy | permission-driven-ui.test.tsx | 2.5 |
