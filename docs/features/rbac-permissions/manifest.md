---
feature: "rbac-permissions"
status: tasks
---

# Feature: rbac-permissions

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | RBAC 权限体系：团队级角色 + superadmin 预置角色，权限码由代码定义，角色在线可配 |
| User Stories | prd/prd-user-stories.md | 10 个用户故事：角色管理(1)、邀请角色(2)、权限渲染(3)、数据迁移(4)、团队创建(5)、PM 操作(6)、Member 受限(7)、跨团队隔离(8)、后端强制(9)、即时生效(10) |
| UI Functions | prd/prd-ui-functions.md | 6 个 UI 功能：角色列表、角色编辑表单、权限码浏览、邀请时角色选择、权限驱动渲染、变更成员角色 |
| Tech Design | design/tech-design.md | 统一 RequirePermission 中间件 + Go 权限码注册表 + roles/role_permissions 表 + IsSuperAdmin flag |
| API Handbook | design/api-handbook.md | 6 个新接口（角色 CRUD + 权限码列表 + 用户权限）+ 3 个修改接口（邀请/变更角色/创建团队） |

## Traceability

| PRD Section | User Story | UI Function | Tech Design | API Handbook | Tasks |
|-------------|------------|-------------|-------------|--------------|-------|
| 5.1 角色管理 | Story 1 | UI 1, UI 2, UI 3 | RoleService + RoleRepo + RequirePermission | Role CRUD + Permission Codes | 1.1, 1.2, 2.1, 3.1, 4.2, 5.2 |
| 5.2 团队成员角色管理 | Story 2 | UI 4, UI 6 | TeamMember.RoleID + invite 改造 | Invite/Change Role | 4.3, 5.4 |
| 5.3 前端权限渲染 | Story 3 | UI 5 | PermissionGuard + useHasPermission | GET /me/permissions | 5.1, 5.3 |
| 5.4 预置角色定义 | Story 1, 2 | — | Seed Data + is_preset flag | — | 1.3 |
| 5.5 数据迁移 | Story 4 | — | Migration script | — | 1.3 |
| 5.6 JWT Claims | — | — | Claims 移除 Role，加 Username | — | 4.3 |
| 5.7 关联改动 | Story 5 | UI 5 | Router 权限映射表 | Create Team 改造 | 4.1, 4.3 |
