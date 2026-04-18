---
created: 2026-04-18
related: design/tech-design.md
---

# API Handbook: RBAC 权限体系

## API Overview

新增 8 个端点（6 个角色管理 + 1 个权限列表 + 1 个用户权限），修改 2 个现有端点（邀请成员、变更角色），删除 1 个端点（can-create-team）。

## Endpoints

### List Roles

**Method**: `GET`
**Path**: `/api/v1/admin/roles`
**Auth**: RequireSuperAdmin

#### Request (Query Parameters)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| search | string | No | 按角色名称模糊匹配 |
| preset | string | No | `all`(default) / `true` / `false` |
| page | int | No | 页码，默认 1 |
| page_size | int | No | 每页条数，默认 20 |

#### Response (200)

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 1,
        "name": "superadmin",
        "description": "系统最高权限角色",
        "is_preset": true,
        "permission_count": 0,
        "member_count": 1,
        "created_at": "2026-04-18T00:00:00Z"
      }
    ],
    "total": 3,
    "page": 1,
    "page_size": 20
  }
}
```

---

### Create Role

**Method**: `POST`
**Path**: `/api/v1/admin/roles`
**Auth**: RequireSuperAdmin

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | Yes | 2-50 字符，不可重复 |
| description | string | No | 最多 200 字符 |
| permission_codes | string[] | Yes | 至少 1 个，必须是系统定义的权限码 |

```json
{
  "name": "viewer",
  "description": "只读查看者",
  "permission_codes": ["main_item:read", "sub_item:read", "view:weekly"]
}
```

#### Response (201)

```json
{
  "code": 0,
  "data": {
    "id": 4,
    "name": "viewer",
    "description": "只读查看者",
    "is_preset": false,
    "created_at": "2026-04-18T10:00:00Z"
  }
}
```

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | VALIDATION_ERROR | 参数校验失败 |
| 409 | ROLE_DUPLICATE | 角色名称已存在 |

---

### Get Role

**Method**: `GET`
**Path**: `/api/v1/admin/roles/:roleId`
**Auth**: RequireSuperAdmin

#### Response (200)

```json
{
  "code": 0,
  "data": {
    "id": 2,
    "name": "pm",
    "description": "团队管理权限",
    "is_preset": true,
    "permissions": [
      {"code": "team:create", "resource": "team", "action": "create", "description": "创建团队"},
      {"code": "team:read", "resource": "team", "action": "read", "description": "查看团队信息"}
    ],
    "member_count": 5,
    "created_at": "2026-04-18T00:00:00Z",
    "updated_at": "2026-04-18T00:00:00Z"
  }
}
```

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 404 | ROLE_NOT_FOUND | 角色不存在 |

---

### Update Role

**Method**: `PUT`
**Path**: `/api/v1/admin/roles/:roleId`
**Auth**: RequireSuperAdmin

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | No | 2-50 字符，不可与其他角色重复 |
| description | string | No | 最多 200 字符 |
| permission_codes | string[] | No | 替换权限列表，至少 1 个 |

```json
{
  "name": "viewer",
  "description": "只读查看者（含甘特图）",
  "permission_codes": ["main_item:read", "sub_item:read", "view:weekly", "view:gantt"]
}
```

#### Response (200)

```json
{
  "code": 0,
  "data": {
    "id": 4,
    "name": "viewer",
    "description": "只读查看者（含甘特图）",
    "is_preset": false,
    "updated_at": "2026-04-18T11:00:00Z"
  }
}
```

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 404 | ROLE_NOT_FOUND | 角色不存在 |
| 409 | ROLE_DUPLICATE | 角色名称已存在 |

> 预置角色（pm、member）可编辑权限码，但不可修改 name。superadmin 角色不可编辑（返回 409 ROLE_PRESET）。

---

### Delete Role

**Method**: `DELETE`
**Path**: `/api/v1/admin/roles/:roleId`
**Auth**: RequireSuperAdmin

#### Response (200)

```json
{
  "code": 0,
  "data": null
}
```

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 404 | ROLE_NOT_FOUND | 角色不存在 |
| 409 | ROLE_PRESET | 预置角色不可删除 |
| 409 | ROLE_IN_USE | 角色正在被 N 个用户使用 |

---

### List Permissions

**Method**: `GET`
**Path**: `/api/v1/admin/permissions`
**Auth**: RequireSuperAdmin

#### Response (200)

```json
{
  "code": 0,
  "data": [
    {
      "resource": "team",
      "actions": [
        {"code": "team:create", "description": "创建团队"},
        {"code": "team:read", "description": "查看团队信息"},
        {"code": "team:update", "description": "编辑团队信息"}
      ]
    }
  ]
}
```

---

### Get Current User Permissions

**Method**: `GET`
**Path**: `/api/v1/me/permissions`
**Auth**: AuthMiddleware

#### Response (200)

```json
{
  "code": 0,
  "data": {
    "is_superadmin": false,
    "team_roles": {
      "1": "pm",
      "3": "member"
    },
    "team_permissions": {
      "1": ["team:create", "team:read", "team:update", "..."],
      "3": ["main_item:read", "sub_item:read", "..."]
    }
  }
}
```

> `team_roles` 和 `team_permissions` 的 key 是 teamID（字符串形式的 uint）。

---

### Invite Team Member (Modified)

**Method**: `POST`
**Path**: `/api/v1/teams/:teamId/members`
**Auth**: AuthMiddleware + RequirePermission(team:invite)

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| username | string | Yes | 用户账号 |
| role_id | uint | Yes | 角色ID，从系统角色列表中选择 |

```json
{
  "username": "zhangsan",
  "role_id": 3
}
```

> 变更：`role` 字符串字段替换为 `role_id` 数字字段。

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | VALIDATION_ERROR | role_id 不存在 |
| 404 | USER_NOT_FOUND | 用户不存在 |
| 409 | ALREADY_MEMBER | 用户已在团队中 |

---

### Change Team Member Role (New)

**Method**: `PUT`
**Path**: `/api/v1/teams/:teamId/members/:userId/role`
**Auth**: AuthMiddleware + RequirePermission(team:invite)

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| role_id | uint | Yes | 新角色ID |

```json
{
  "role_id": 2
}
```

#### Response (200)

```json
{
  "code": 0,
  "data": {
    "user_id": 5,
    "team_id": 1,
    "role_id": 2,
    "role_name": "pm"
  }
}
```

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | VALIDATION_ERROR | role_id 不存在 |
| 403 | NOT_TEAM_MEMBER | 目标用户不在该团队中 |
| 409 | ROLE_PRESET | 不能变更 superadmin 全局角色 |

---

## Data Contracts

### RoleListItem

| Field | Type | Description |
|-------|------|-------------|
| id | uint | 角色 ID |
| name | string | 角色名称 |
| description | string | 描述 |
| is_preset | bool | 是否预置角色 |
| permission_count | int | 权限码数量 |
| member_count | int | 使用该角色的成员数 |
| created_at | datetime | 创建时间 |

### PermissionGroup

| Field | Type | Description |
|-------|------|-------------|
| resource | string | 资源名（如 team） |
| actions | PermissionAction[] | 该资源下的操作列表 |

### PermissionAction

| Field | Type | Description |
|-------|------|-------------|
| code | string | 权限码（如 team:create） |
| description | string | 操作描述 |

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| ROLE_NOT_FOUND | 404 | 角色不存在 |
| ROLE_DUPLICATE | 409 | 角色名称已存在 |
| ROLE_PRESET | 409 | 预置角色不可删除/编辑 |
| ROLE_IN_USE | 409 | 角色正在被使用 |
| PERMISSION_DENIED | 403 | 权限不足 |
