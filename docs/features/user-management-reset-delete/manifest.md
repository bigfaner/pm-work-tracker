---
feature: "user-management-reset-delete"
status: prd
---

# Feature: user-management-reset-delete

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 超级管理员重置密码（弹窗输入新密码）和软删除用户（列表过滤+登录拦截）的功能需求，含流程图、表单校验规则和安全约束 |
| User Stories | prd/prd-user-stories.md | 5 个用户故事，覆盖重置密码、输入校验、软删除、防误删自身、非超管鉴权 |
| UI Functions | prd/prd-ui-functions.md | 3 个 UI 功能：列表操作列扩展、重置密码弹窗、删除确认弹窗 |

## Traceability

| PRD Section | Design Section | UI Component | Tasks |
|-------------|----------------|--------------|-------|
| 重置密码功能 | — | 重置密码弹窗 | — |
| 软删除用户功能 | — | 删除确认弹窗 | — |
| 用户列表操作列 | — | UserManagementPage 操作列 | — |
| 登录鉴权关联改动 | — | — | — |
