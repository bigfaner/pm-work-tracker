---
created: 2026-04-19
related: design/tech-design.md
---

# API Handbook: Improve UI

> 仅列出新增和变更的端点。未列出的端点保持原 API Handbook (`docs/features/pm-work-tracker/design/api-handbook.md`) 中的定义不变。

## Endpoints

---

### 创建用户

**Method**: `POST`
**Path**: `/api/v1/admin/users`
**Auth**: SuperAdmin

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| username | string | Yes | 账号名，3-64 字符，唯一 |
| displayName | string | Yes | 显示名称，1-64 字符 |
| email | string | No | 邮箱地址，最大 100 字符 |
| teamId | uint | No | 加入的团队 ID |
| canCreateTeam | bool | No | 创建团队权限，默认 false |

> **注意**: 无需传 password 字段。后端自动生成随机初始密码（12 位，含大小写字母+数字），通过响应返回，仅展示一次。

#### Response (201)

| Field | Type | Description |
|-------|------|-------------|
| id | uint | 用户 ID |
| username | string | 账号名 |
| displayName | string | 显示名称 |
| email | string | 邮箱 |
| canCreateTeam | bool | 创建团队权限 |
| status | string | "enabled" |
| teams | array | 所属团队列表（若传了 teamId） |
| initialPassword | string | 自动生成的初始密码（仅此次返回，后续不可查询） |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | VALIDATION_ERROR | 字段校验失败 |
| 422 | USER_EXISTS | 账号名已存在 |
| 404 | TEAM_NOT_FOUND | teamId 对应的团队不存在 |

---

### 获取单个用户

**Method**: `GET`
**Path**: `/api/v1/admin/users/:userId`
**Auth**: SuperAdmin

#### Response (200)

| Field | Type | Description |
|-------|------|-------------|
| id | uint | |
| username | string | |
| displayName | string | |
| email | string | |
| canCreateTeam | bool | |
| isSuperAdmin | bool | |
| status | string | "enabled" / "disabled" |
| teams | TeamSummary[] | 所属团队列表 |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 404 | USER_NOT_FOUND | 用户不存在 |

---

### 编辑用户

**Method**: `PUT`
**Path**: `/api/v1/admin/users/:userId`
**Auth**: SuperAdmin

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| displayName | string | No | 显示名称 |
| email | string | No | 邮箱 |
| canCreateTeam | bool | No | 创建团队权限 |
| teamId | uint | No | 调整所属团队 |

#### Response (200)

| Field | Type | Description |
|-------|------|-------------|
| id | uint | |
| username | string | |
| displayName | string | |
| email | string | |
| canCreateTeam | bool | |
| status | string | |
| teams | TeamSummary[] | |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 404 | USER_NOT_FOUND | 用户不存在 |
| 400 | VALIDATION_ERROR | 字段校验失败 |

---

### 变更用户状态

**Method**: `PUT`
**Path**: `/api/v1/admin/users/:userId/status`
**Auth**: SuperAdmin

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| status | string | Yes | "enabled" / "disabled" |

#### Response (200)

| Field | Type | Description |
|-------|------|-------------|
| id | uint | |
| username | string | |
| status | string | 变更后的状态 |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 404 | USER_NOT_FOUND | 用户不存在 |
| 422 | INVALID_STATUS | 无效的状态值 |
| 422 | CANNOT_DISABLE_SELF | 不能禁用自己 |

---

### 用户列表（增强）

**Method**: `GET`
**Path**: `/api/v1/admin/users`
**Auth**: SuperAdmin

原端点增强，增加搜索、筛选和返回字段。

#### Request（新增参数）

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| search | string | No | 按用户名/账号模糊搜索 |
| canCreateTeam | string | No | "true" / "false" 筛选 |
| page | int | No | 默认 1 |
| pageSize | int | No | 默认 20 |

#### Response (200)（增强字段）

| Field | Type | Description |
|-------|------|-------------|
| id | uint | |
| username | string | |
| displayName | string | |
| email | string | **新增** |
| canCreateTeam | bool | |
| isSuperAdmin | bool | |
| status | string | **新增**: "enabled" / "disabled" |
| teams | TeamSummary[] | **新增**: 所属团队列表 |

---

### 每周进展对比视图（重构）

**Method**: `GET`
**Path**: `/api/v1/teams/:teamId/views/weekly`
**Auth**: Team member

响应格式重构为上周/本周双列对比。

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| weekStart | string | Yes | ISO8601 date（当周周一） |

#### Response (200)

```json
{
  "weekStart": "2026-04-13",
  "weekEnd": "2026-04-19",
  "stats": {
    "activeSubItems": 5,
    "newlyCompleted": 2,
    "inProgress": 3,
    "blocked": 1
  },
  "groups": [
    {
      "mainItem": {
        "id": 1,
        "title": "用户认证模块开发",
        "priority": "P1",
        "startDate": "2026-04-01",
        "expectedEndDate": "2026-04-25",
        "completion": 58,
        "subItemCount": 4
      },
      "lastWeek": [
        {
          "id": 10,
          "title": "JWT Token 集成",
          "priority": "P2",
          "status": "进行中",
          "assigneeName": "李伟",
          "expectedEndDate": "2026-04-18",
          "completion": 40,
          "progressDescription": "Token 签发逻辑开发中"
        }
      ],
      "thisWeek": [
        {
          "id": 10,
          "title": "JWT Token 集成",
          "priority": "P2",
          "status": "进行中",
          "assigneeName": "李伟",
          "expectedEndDate": "2026-04-18",
          "completion": 70,
          "progressDescription": "Token 签发完成，黑名单联调中",
          "delta": 30,
          "isNew": false,
          "justCompleted": false
        },
        {
          "id": 15,
          "title": "OAuth2.0 对接",
          "priority": "P3",
          "status": "待开始",
          "assigneeName": "张明",
          "expectedEndDate": "2026-04-22",
          "completion": 0,
          "progressDescription": "",
          "delta": 0,
          "isNew": true,
          "justCompleted": false
        }
      ],
      "completedNoChange": [
        {
          "id": 8,
          "title": "登录页开发",
          "priority": "P1",
          "status": "已完成",
          "assigneeName": "王芳",
          "expectedEndDate": "2026-04-10",
          "completion": 100
        }
      ]
    }
  ]
}
```

**字段说明：**

| Field | Type | Description |
|-------|------|-------------|
| stats.activeSubItems | int | 本周有进度的子事项数 |
| stats.newlyCompleted | int | 本周新完成的子事项数 |
| stats.inProgress | int | 当前进行中的子事项数 |
| stats.blocked | int | 当前阻塞中的子事项数 |
| groups[].lastWeek | array | 上周快照（所有活跃子事项的状态） |
| groups[].thisWeek | array | 本周快照（含增量标记） |
| groups[].completedNoChange | array | 已完成且本周无变化的子事项（默认折叠） |
| SubItemSnapshot.delta | number | 进度增量（thisWeek.completion - lastWeek.completion），>0 时前端显示 "+N%" |
| SubItemSnapshot.isNew | bool | 上周快照中不存在 → 本周新增 |
| SubItemSnapshot.justCompleted | bool | 本周状态变为"已完成" |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | VALIDATION_ERROR | weekStart 格式错误或缺失 |
| 422 | FUTURE_WEEK_NOT_ALLOWED | weekStart 在未来 |

---

## Data Contracts

### TeamSummary

```json
{
  "id": 1,
  "name": "产品研发团队",
  "role": "pm"
}
```

### SubItemSnapshot

```json
{
  "id": 10,
  "title": "JWT Token 集成",
  "priority": "P2",
  "status": "进行中",
  "assigneeName": "李伟",
  "expectedEndDate": "2026-04-18",
  "completion": 70,
  "progressDescription": "Token 签发完成，黑名单联调中",
  "delta": 30,
  "isNew": false,
  "justCompleted": false
}
```

## Error Codes

新增错误码：

| Code | HTTP Status | Description |
|------|-------------|-------------|
| USER_EXISTS | 422 | 创建用户时账号名已存在 |
| USER_NOT_FOUND | 404 | 用户不存在 |
| USER_DISABLED | 403 | 账号已被禁用（登录时返回） |
| INVALID_STATUS | 422 | 无效的用户状态值 |
| CANNOT_DISABLE_SELF | 422 | 超管不能禁用自己 |
| TEAM_NOT_FOUND | 404 | 团队不存在（创建用户时指定） |
| FUTURE_WEEK_NOT_ALLOWED | 422 | 不允许选择未来周 |

原有错误码保持不变，参见 `docs/features/pm-work-tracker/design/api-handbook.md`。
