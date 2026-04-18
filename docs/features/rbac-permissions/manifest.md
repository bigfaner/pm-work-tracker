---
feature: "rbac-permissions"
status: design
---

# Feature: rbac-permissions

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | RBAC 权限体系：团队级角色 + 全局 superadmin，权限码由代码定义，角色在线可配 |
| User Stories | prd/prd-user-stories.md | 4 个用户故事：角色管理、邀请时指定角色、权限渲染、数据迁移 |
| UI Functions | prd/prd-ui-functions.md | 5 个 UI 功能：角色列表、角色编辑表单、权限码浏览、邀请时角色选择、权限驱动渲染 |
| Tech Design | design/tech-design.md | roles + role_permissions 数据模型，权限码常量注册表，RequirePermission 中间件，内存权限缓存 |
| API Handbook | design/api-handbook.md | 8 个新端点 + 2 个修改端点 + 1 个删除端点，含完整请求/响应格式 |
| UI Design | ui/ui-design.md | Tailwind UI 风格，5 个组件：角色列表、角色编辑模态框、权限浏览、邀请角色选择、权限驱动渲染 |

## Traceability

| PRD Section | Design Section | UI Design | Tasks |
|-------------|----------------|-----------|-------|
| 5.1 角色管理 | roles + role_permissions 模型, RoleRepo, RoleService, role_handler | 角色列表页 + 角色编辑表单 | |
| 5.2 团队成员角色管理 | team_members.role_id FK, InviteMemberReq.role_id, ChangeMemberRole API | 邀请成员角色选择 | |
| 5.3 前端权限渲染 | GET /api/me/permissions, 权限驱动渲染 | 权限驱动 UI 渲染 + 受控元素映射 | |
| 5.4 预置角色定义 | permissions 常量注册表, 数据库种子数据 | 权限码浏览视图 | |
| 5.5 数据迁移 | Migration 事务脚本 | | |
| 5.6 JWT Claims | JWT Claims 精简, RequireSuperAdmin 中间件 | | |
| 5.7 关联改动 | Middleware Route Mapping 表 | | |
