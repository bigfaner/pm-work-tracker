---
created: 2026-04-19
related: design/tech-design.md
---

# API Handbook: RBAC 权限体系

## API Overview

新增 6 个管理端接口（角色 CRUD + 权限码列表）和 1 个用户端接口（当前用户权限查询）。修改 2 个现有接口（邀请成员、变更角色）。所有管理端接口需要 `user:manage_role` 权限。

## New Endpoints

### List Roles

**Method**: `GET`
**Path**: `/api/v1/admin/roles`
**Auth**: `user:manage_role`

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| search | string | No | 按角色名称模糊搜索 |
| is_preset | string | No | 筛选："all"(default) / "preset" / "custom" |
| page | int | No | 页码，默认 1 |
| page_size | int | No | 每页条数，默认 20 |

#### Response (200)

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 2,
        "name": "pm",
        "description": "团队管理权限",
        "is_preset": true,
        "permission_count": 22,
        "member_count": 5,
        "created_at": "2026-04-19T00:00:00Z"
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
**Auth**: `user:manage_role`

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | Yes | 角色名称，2-50 字符，不可重名 |
| description | string | No | 角色描述，最多 200 字符 |
| permission_codes | string[] | Yes | 权限码列表，至少 1 个 |

#### Response (200)

```json
{
  "code": 0,
  "data": {
    "id": 4,
    "name": "viewer",
    "description": "只读查看者",
    "is_preset": false,
    "permission_count": 3,
    "member_count": 0,
    "created_at": "2026-04-19T12:00:00Z"
  }
}
```

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | `ERR_INVALID_PERMISSION_CODE` | 包含不存在的权限码 |
| 409 | `ERR_ROLE_NAME_EXISTS` | 角色名称已存在 |
| 422 | `ERR_VALIDATION` | 名称长度不符或无权限码 |

---

### Get Role

**Method**: `GET`
**Path**: `/api/v1/admin/roles/:id`
**Auth**: `user:manage_role`

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
      {"code": "team:create", "description": "创建团队"},
      {"code": "team:read", "description": "查看团队信息"}
    ],
    "member_count": 5,
    "created_at": "2026-04-19T00:00:00Z"
  }
}
```

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 404 | `ERR_ROLE_NOT_FOUND` | 角色不存在 |

---

### Update Role

**Method**: `PUT`
**Path**: `/api/v1/admin/roles/:id`
**Auth**: `user:manage_role`

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | No | 角色名称（预置角色不可修改名称） |
| description | string | No | 角色描述 |
| permission_codes | string[] | No | 权限码列表（至少 1 个） |

#### Response (200)

```json
{
  "code": 0,
  "data": {
    "id": 4,
    "name": "viewer",
    "description": "只读查看者（含导出）",
    "is_preset": false,
    "permissions": [
      {"code": "team:read", "description": "查看团队信息"},
      {"code": "main_item:read", "description": "查看主事项"},
      {"code": "report:export", "description": "导出周报"}
    ],
    "member_count": 0,
    "created_at": "2026-04-19T12:00:00Z"
  }
}
```

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 403 | `ERR_PRESET_ROLE_IMMUTABLE` | superadmin 预置角色不可编辑 |
| 404 | `ERR_ROLE_NOT_FOUND` | 角色不存在 |
| 409 | `ERR_ROLE_NAME_EXISTS` | 角色名称已存在 |

**预置角色编辑规则**：
- `superadmin`：不可编辑名称、描述、权限（返回 403）
- `pm`/`member`：可编辑描述和权限勾选，不可修改名称，不可删除

---

### Delete Role

**Method**: `DELETE`
**Path**: `/api/v1/admin/roles/:id`
**Auth**: `user:manage_role`

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
| 403 | `ERR_PRESET_ROLE_IMMUTABLE` | 预置角色不可删除 |
| 422 | `ERR_ROLE_IN_USE` | 角色正在被 N 个用户使用 |

---

### List Permission Codes

**Method**: `GET`
**Path**: `/api/v1/admin/permissions`
**Auth**: `user:manage_role`

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
    },
    {
      "resource": "main_item",
      "actions": [
        {"code": "main_item:create", "description": "创建主事项"},
        {"code": "main_item:read", "description": "查看主事项"}
      ]
    }
  ]
}
```

---

### Get Current User Permissions

**Method**: `GET`
**Path**: `/api/v1/me/permissions`
**Auth**: 认证用户即可（无特定权限码要求）

#### Response (200)

```json
{
  "code": 0,
  "data": {
    "is_superadmin": false,
    "team_permissions": {
      "1": ["team:read", "team:update", "team:invite", "main_item:create"],
      "3": ["team:read", "main_item:read", "sub_item:create"]
    }
  }
}
```

**说明**：
- `is_superadmin`：对应用户的 `IsSuperAdmin` 标记
- `team_permissions`：key 为 team ID（字符串），value 为该团队角色拥有的权限码数组
- SuperAdmin 用户返回 `is_superadmin: true`，`team_permissions` 可以为空（前端知道是 superadmin 即可）

---

## Modified Endpoints

### Invite Member (Modified)

**Method**: `POST`
**Path**: `/api/v1/teams/:teamId/members`
**Auth**: `team:invite`

#### Request (Changed)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| username | string | Yes | 用户账号 |
| role_id | uint | Yes | 角色 ID（从系统角色列表中选择，不可为 superadmin） |

```json
{
  "username": "zhangsan",
  "role_id": 3
}
```

#### Response (200)

Unchanged.

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | `ERR_VALIDATION` | role_id 不能为 superadmin 角色ID |
| 404 | `ERR_USER_NOT_FOUND` | 用户不存在 |
| 404 | `ERR_ROLE_NOT_FOUND` | 角色不存在 |
| 422 | `ERR_ALREADY_MEMBER` | 用户已是团队成员 |

---

### Change Member Role (Modified)

**Method**: `PUT`
**Path**: `/api/v1/teams/:teamId/members/:memberId/role`
**Auth**: `team:invite`

#### Request (Changed)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| role_id | uint | Yes | 新角色 ID |

```json
{
  "role_id": 4
}
```

#### Response (200)

Unchanged.

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 404 | `ERR_ROLE_NOT_FOUND` | 角色不存在 |
| 403 | `ERR_CANNOT_MODIFY_SELF` | 不能变更自己的角色 |

---

### Create Team (Modified)

**Method**: `POST`
**Path**: `/api/v1/teams`
**Auth**: `team:create`（非团队上下文权限）

**变更**：从检查 `User.CanCreateTeam` 改为检查 `team:create` 权限码。SuperAdmin 自动拥有。

#### Request

Unchanged.

#### Response

Unchanged.

---

## Data Contracts

### RoleDetailResp

```go
type RoleDetailResp struct {
    ID              uint              `json:"id"`
    Name            string            `json:"name"`
    Description     string            `json:"description"`
    IsPreset        bool              `json:"is_preset"`
    Permissions     []PermissionItem  `json:"permissions,omitempty"`
    PermissionCount int               `json:"permission_count"`
    MemberCount     int64             `json:"member_count"`
    CreatedAt       time.Time         `json:"created_at"`
}
```

### PermissionItem

```go
type PermissionItem struct {
    Code        string `json:"code"`
    Description string `json:"description"`
}
```

### PermissionGroup

```go
type PermissionGroup struct {
    Resource string           `json:"resource"`
    Actions  []PermissionItem `json:"actions"`
}
```

### UserPermissionsResp

```go
type UserPermissionsResp struct {
    IsSuperAdmin    bool              `json:"is_superadmin"`
    TeamPermissions map[uint][]string `json:"team_permissions"`
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `ERR_ROLE_NOT_FOUND` | 404 | 角色不存在 |
| `ERR_ROLE_NAME_EXISTS` | 409 | 角色名称已被使用 |
| `ERR_ROLE_IN_USE` | 422 | 角色正在被使用，无法删除 |
| `ERR_PRESET_ROLE_IMMUTABLE` | 403 | 预置角色不可编辑或删除 |
| `ERR_INVALID_PERMISSION_CODE` | 400 | 提交了不存在的权限码 |
| `ERR_CANNOT_MODIFY_SELF` | 403 | 不能修改自己的角色 |
| `ERR_ALREADY_MEMBER` | 422 | 用户已是团队成员 |
| `ERR_FORBIDDEN` | 403 | 缺少所需权限码 |
