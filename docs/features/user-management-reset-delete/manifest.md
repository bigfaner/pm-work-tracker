---
feature: "user-management-reset-delete"
status: tasks
---

# Feature: user-management-reset-delete

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 超级管理员重置密码（弹窗输入新密码）和软删除用户（列表过滤+登录拦截）的功能需求，含流程图、表单校验规则和安全约束 |
| User Stories | prd/prd-user-stories.md | 6 个用户故事，覆盖重置密码、输入校验、软删除、防误删自身、非超管鉴权、复制账号密码 |
| UI Functions | prd/prd-ui-functions.md | 4 个 UI 功能：列表操作列扩展、重置密码弹窗、删除确认弹窗、复制账号与密码 |
| Tech Design | design/tech-design.md | 三层架构设计：Handler→Service→Repo，前端 Radix Dialog + React Query，复用 DeletedFlag |
| API Handbook | design/api-handbook.md | 2 个新端点（PUT password, DELETE user）+ 3 个现有端点行为变更 |

## Traceability

| PRD Section | Design Section | UI Component | Tasks |
|-------------|----------------|--------------|-------|
| Story 1: 重置密码 happy path | AdminService.ResetPassword + AdminHandler.ResetPassword | 重置密码弹窗 | 1.1, 2.1, 2.2, 3.2 |
| Story 1: 后端错误处理 | Handler error mapping | 重置密码弹窗 | 2.2, 3.2 |
| Story 2: 密码输入校验 | Frontend form validation + Gin binding | 重置密码弹窗 | 1.1, 3.2 |
| Story 3: 软删除 happy path | AdminService.SoftDeleteUser + AdminHandler.DeleteUser | 删除确认弹窗 | 1.1, 1.2, 2.1, 2.2, 3.3 |
| Story 3: 列表状态过期 | Service ErrUserNotFound | 删除确认弹窗 | 1.1, 2.1, 3.3 |
| Story 4: 防止误删自身 | Service self-delete check | 列表操作列 | 2.1, 3.3 |
| Story 5: 非超管鉴权 | Permission middleware + frontend isSuperAdmin | 列表操作列 | 2.2, 3.2, 3.3 |
| Story 6: 复制账号密码 | Clipboard API | 创建用户结果弹窗 | 3.3 |
| 5.1: 列表过滤已删除 | NotDeleted scope on ListFiltered | — | 1.1 |
| 5.5: 登录拒绝已删除 | AuthService.DeletedFlag check | — | 1.2 |
| 5.5: JWT 拒绝已删除 | AuthMiddleware.DeletedFlag check | — | 1.2 |
