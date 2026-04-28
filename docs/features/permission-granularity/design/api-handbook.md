---
created: 2026-04-28
related: design/tech-design.md
---

# API Handbook: 细化 user / role 权限粒度

## API Overview

本次变更不新增 API 接口，仅修改现有接口的权限码守卫。以下列出所有受影响接口的变更前后对比及完整契约。

## 受影响接口

### GET /admin/users — 列出用户

**Method**: `GET`
**Path**: `/v1/admin/users`
**Auth**: 登录用户，需要 `user:list`（原 `user:read`）

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| page | int | 否 | 页码，默认 1 |
| pageSize | int | 否 | 每页条数，默认 20 |

#### Response (200)

| Field | Type | Description |
|-------|------|-------------|
| items | User[] | 用户列表 |
| total | int | 总条数 |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 403 | ERR_FORBIDDEN | 缺少 user:list 权限 |

---

### GET /admin/users/:userId — 查看用户详情

**Method**: `GET`
**Path**: `/v1/admin/users/:userId`
**Auth**: 登录用户，需要 `user:read`（语义变更：原"列出用户"→新"查看用户详情"）

#### Response (200)

| Field | Type | Description |
|-------|------|-------------|
| bizKey | string | 用户唯一标识 |
| username | string | 用户名 |
| displayName | string | 显示名称 |
| email | string | 邮箱（敏感字段） |
| phone | string | 手机号（敏感字段） |
| isSuperAdmin | bool | 是否超管 |
| status | string | 账号状态 |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 403 | ERR_FORBIDDEN | 缺少 user:read 权限 |
| 404 | ERR_NOT_FOUND | 用户不存在 |

---

### POST /admin/users — 给用户分配角色

**Method**: `POST`
**Path**: `/v1/admin/users`
**Auth**: 登录用户，需要 `user:assign_role`（原 `user:manage_role`）

> 注：此路由在现有实现中用于给用户分配角色，非创建用户。

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| userId | string | 是 | 目标用户 bizKey |
| roleId | string | 是 | 角色 bizKey |

#### Response (200)

| Field | Type | Description |
|-------|------|-------------|
| — | — | 空 data，code=0 表示成功 |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 403 | ERR_FORBIDDEN | 缺少 user:assign_role 权限 |
| 404 | ERR_NOT_FOUND | 用户或角色不存在 |

---

### GET /admin/teams — 列出团队

**Method**: `GET`
**Path**: `/v1/admin/teams`
**Auth**: 登录用户，需要 `user:list`（原 `user:read`）

#### Response (200)

| Field | Type | Description |
|-------|------|-------------|
| items | Team[] | 团队列表 |
| total | int | 总条数 |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 403 | ERR_FORBIDDEN | 缺少 user:list 权限 |

---

### GET /admin/roles — 查看角色列表

**Method**: `GET`
**Path**: `/v1/admin/roles`
**Auth**: 登录用户，需要 `role:read`（原 `user:manage_role`）

#### Response (200)

| Field | Type | Description |
|-------|------|-------------|
| items | RoleListItem[] | 角色列表 |
| total | int | 总条数 |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 403 | ERR_FORBIDDEN | 缺少 role:read 权限 |

---

### GET /admin/roles/:id — 查看角色详情

**Method**: `GET`
**Path**: `/v1/admin/roles/:id`
**Auth**: 登录用户，需要 `role:read`（原 `user:manage_role`）

#### Response (200)

| Field | Type | Description |
|-------|------|-------------|
| bizKey | string | 角色唯一标识 |
| roleName | string | 角色名称 |
| roleDesc | string | 角色描述 |
| isPreset | bool | 是否预置角色 |
| permissions | PermissionItem[] | 权限码列表 |
| memberCount | int | 绑定成员数 |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 403 | ERR_FORBIDDEN | 缺少 role:read 权限 |
| 404 | ERR_NOT_FOUND | 角色不存在 |

---

### POST /admin/roles — 创建角色

**Method**: `POST`
**Path**: `/v1/admin/roles`
**Auth**: 登录用户，需要 `role:create`（原 `user:manage_role`）

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| roleName | string | 是 | 角色名称，1-50 字符，不可重复 |
| roleDesc | string | 否 | 角色描述，最多 200 字符 |
| permissionCodes | string[] | 否 | 权限码列表，空数组合法 |

#### Response (201)

| Field | Type | Description |
|-------|------|-------------|
| bizKey | string | 新角色唯一标识 |
| roleName | string | 角色名称 |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | ERR_VALIDATION | 参数校验失败 |
| 403 | ERR_FORBIDDEN | 缺少 role:create 权限 |
| 409 | ERR_ROLE_NAME_EXISTS | 角色名称已存在 |

---

### PUT /admin/roles/:id — 编辑角色

**Method**: `PUT`
**Path**: `/v1/admin/roles/:id`
**Auth**: 登录用户，需要 `role:update`（原 `user:manage_role`）

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| roleName | string | 否 | 新角色名称（预置角色不可改名） |
| roleDesc | string | 否 | 新描述 |
| permissionCodes | string[] | 否 | 全量覆盖权限码列表 |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 403 | ERR_FORBIDDEN | 缺少 role:update 权限，或尝试修改预置角色名称 |
| 404 | ERR_NOT_FOUND | 角色不存在 |

---

### DELETE /admin/roles/:id — 删除角色

**Method**: `DELETE`
**Path**: `/v1/admin/roles/:id`
**Auth**: 登录用户，需要 `role:delete`（原 `user:manage_role`）

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 403 | ERR_FORBIDDEN | 缺少 role:delete 权限，或目标为预置角色 |
| 404 | ERR_NOT_FOUND | 角色不存在 |
| 422 | ERR_ROLE_IN_USE | 角色下有成员绑定，无法删除 |

---

### GET /admin/permissions — 查看权限码注册表

**Method**: `GET`
**Path**: `/v1/admin/permissions`
**Auth**: 登录用户，需要 `role:read`（原 `user:manage_role`）

#### Response (200)

| Field | Type | Description |
|-------|------|-------------|
| — | ResourcePermissions[] | 按资源分组的权限码列表 |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 403 | ERR_FORBIDDEN | 缺少 role:read 权限 |

---

## Data Contracts

```ts
interface RoleListItem {
  bizKey: string
  roleName: string
  roleDesc: string
  isPreset: boolean
  permissionCount: number
  memberCount: number
  createTime: string  // ISO 8601
}

interface PermissionItem {
  code: string
  description: string
}

interface ResourcePermissions {
  resource: string
  permissions: PermissionItem[]
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| ERR_FORBIDDEN | 403 | 缺少对应权限码 |
| ERR_NOT_FOUND | 404 | 资源不存在 |
| ERR_VALIDATION | 400/422 | 参数校验失败 |
| ERR_ROLE_NAME_EXISTS | 409 | 角色名称重复 |
| ERR_ROLE_IN_USE | 422 | 角色有成员绑定，无法删除 |
| ERR_PRESET_ROLE_IMMUTABLE | 403 | 预置角色不可编辑或删除 |
