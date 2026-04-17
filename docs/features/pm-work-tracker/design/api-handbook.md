---
created: 2026-04-17
related: design/tech-design.md
---

# API Handbook: PM Work Tracker

## API Overview

Base URL: `/api/v1`  
Auth: `Authorization: Bearer <jwt>`  
Content-Type: `application/json`

Response envelope:
```json
// success
{ "code": 0, "data": { ... } }
// paginated
{ "code": 0, "data": { "items": [...], "total": 100, "page": 1, "pageSize": 20 } }
// error
{ "code": "ERROR_CODE", "message": "human readable message" }
```

---

## Auth

### POST /api/v1/auth/login

**Auth**: public

#### Request
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| username | string | ✓ | 账号 |
| password | string | ✓ | 密码 |

#### Response (200)
| Field | Type | Description |
|-------|------|-------------|
| token | string | JWT token (24h expiry) |
| user.id | uint | 用户 ID |
| user.username | string | 账号 |
| user.displayName | string | 姓名 |
| user.isSuperAdmin | bool | 是否超级管理员 |
| user.canCreateTeam | bool | 是否有创建团队权限 |

#### Error Responses
| Status | Code | Description |
|--------|------|-------------|
| 400 | VALIDATION_ERROR | 字段缺失 |
| 401 | UNAUTHORIZED | 账号或密码错误 |

---

### POST /api/v1/auth/logout

**Auth**: any authenticated

#### Request
(empty body)

#### Response (200)
```json
{ "code": 0, "data": null }
```

---

## Teams

### POST /api/v1/teams

**Auth**: user with `canCreateTeam=true`

#### Request
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | ✓ | 团队名称，max 100 |
| description | string | | 团队描述，max 500 |

#### Response (201)
| Field | Type | Description |
|-------|------|-------------|
| id | uint | 团队 ID |
| name | string | 团队名称 |
| description | string | 描述 |
| pmId | uint | PM 用户 ID |
| createdAt | string | ISO8601 |

#### Error Responses
| Status | Code | Description |
|--------|------|-------------|
| 403 | FORBIDDEN | 无创建团队权限 |

---

### GET /api/v1/teams

**Auth**: authenticated (SuperAdmin sees all; others see own teams)

#### Query Params
(none)

#### Response (200)
Array of team objects (same schema as POST response).

---

### GET /api/v1/teams/:teamId

**Auth**: team member

#### Response (200)
| Field | Type | Description |
|-------|------|-------------|
| id | uint | |
| name | string | |
| description | string | |
| pmId | uint | |
| pm.displayName | string | PM 姓名 |
| memberCount | int | 成员数 |
| mainItemCount | int | 主事项数 |
| createdAt | string | |

---

### PUT /api/v1/teams/:teamId

**Auth**: PM

#### Request
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | | 团队名称 |
| description | string | | 描述 |

#### Response (200)
Updated team object.

---

### DELETE /api/v1/teams/:teamId

**Auth**: PM or SuperAdmin

#### Request
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| confirmName | string | ✓ | 必须与团队名称完全一致 |

#### Response (200)
```json
{ "code": 0, "data": null }
```

#### Error Responses
| Status | Code | Description |
|--------|------|-------------|
| 422 | VALIDATION_ERROR | confirmName 不匹配 |

---

### GET /api/v1/teams/:teamId/members

**Auth**: team member

#### Response (200)
Array of:
| Field | Type | Description |
|-------|------|-------------|
| userId | uint | |
| displayName | string | 姓名 |
| username | string | 账号 |
| role | string | "pm" \| "member" |
| joinedAt | string | ISO8601 |

---

### POST /api/v1/teams/:teamId/members

**Auth**: PM

#### Request
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| username | string | ✓ | 被邀请用户账号 |
| role | string | ✓ | "member" (PM 角色通过 transfer 设置) |

#### Error Responses
| Status | Code | Description |
|--------|------|-------------|
| 404 | USER_NOT_FOUND | 用户不存在 |
| 422 | ALREADY_MEMBER | 已是团队成员 |

---

### DELETE /api/v1/teams/:teamId/members/:userId

**Auth**: PM

#### Error Responses
| Status | Code | Description |
|--------|------|-------------|
| 422 | CANNOT_REMOVE_SELF | PM 不能移除自己 |

---

### PUT /api/v1/teams/:teamId/pm

**Auth**: PM (transfer PM role)

#### Request
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| newPmUserId | uint | ✓ | 新 PM 的用户 ID（必须是团队成员） |

#### Error Responses
| Status | Code | Description |
|--------|------|-------------|
| 422 | NOT_TEAM_MEMBER | 目标用户不是团队成员 |

---

## MainItems

### POST /api/v1/teams/:teamId/main-items

**Auth**: PM

#### Request
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| title | string | ✓ | 标题，max 100 |
| priority | string | ✓ | "P1"\|"P2"\|"P3" |
| assigneeId | uint | | 负责人用户 ID |
| startDate | string | | ISO8601 date |
| expectedEndDate | string | | ISO8601 date，不早于 startDate |

#### Response (201)
Full MainItem object (see Data Contracts).

---

### GET /api/v1/teams/:teamId/main-items

**Auth**: team member

#### Query Params
| Param | Type | Description |
|-------|------|-------------|
| priority | string | P1\|P2\|P3，多值逗号分隔 |
| status | string | 状态值，多值逗号分隔 |
| assigneeId | uint | 负责人筛选 |
| archived | bool | false（默认）\| true |
| page | int | 默认 1 |
| pageSize | int | 默认 20，max 100 |

#### Response (200)
Paginated list of MainItem objects.

---

### GET /api/v1/teams/:teamId/main-items/:itemId

**Auth**: team member

#### Response (200)
MainItem object + `subItems` array (summary fields only).

---

### PUT /api/v1/teams/:teamId/main-items/:itemId

**Auth**: PM

#### Request
Same optional fields as POST (all optional).

#### Response (200)
Updated MainItem object.

---

### POST /api/v1/teams/:teamId/main-items/:itemId/archive

**Auth**: PM

#### Request
(empty body)

#### Error Responses
| Status | Code | Description |
|--------|------|-------------|
| 422 | ARCHIVE_NOT_ALLOWED | 状态不是"已完成"或"已关闭" |

---

## SubItems

### POST /api/v1/teams/:teamId/main-items/:mainId/sub-items

**Auth**: PM or any team member

#### Request
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| title | string | ✓ | 标题，max 100 |
| description | string | | 描述 |
| priority | string | ✓ | "P1"\|"P2"\|"P3" |
| assigneeId | uint | ✓ | 负责人用户 ID |
| startDate | string | | ISO8601 date |
| expectedEndDate | string | | ISO8601 date |

#### Response (201)
Full SubItem object. Initial status: `待开始`.

---

### GET /api/v1/teams/:teamId/main-items/:mainId/sub-items

**Auth**: team member

#### Query Params
| Param | Type | Description |
|-------|------|-------------|
| priority | string | 多值逗号分隔 |
| status | string | 多值逗号分隔 |
| assigneeId | uint | |
| page | int | 默认 1 |
| pageSize | int | 默认 20 |

#### Response (200)
Paginated list of SubItem objects.

---

### GET /api/v1/teams/:teamId/sub-items/:itemId

**Auth**: team member

#### Response (200)
Full SubItem object.

---

### PUT /api/v1/teams/:teamId/sub-items/:itemId

**Auth**: PM or assignee

#### Request
| Field | Type | Description |
|-------|------|-------------|
| title | string | |
| description | string | |
| priority | string | |
| expectedEndDate | string | |

---

### PUT /api/v1/teams/:teamId/sub-items/:itemId/status

**Auth**: PM or assignee

#### Request
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| status | string | ✓ | 目标状态 |

#### Error Responses
| Status | Code | Description |
|--------|------|-------------|
| 422 | INVALID_STATUS | 非法状态流转 |

---

### PUT /api/v1/teams/:teamId/sub-items/:itemId/assignee

**Auth**: PM

#### Request
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| assigneeId | uint | ✓ | 新负责人用户 ID |

---

## ProgressRecords

### POST /api/v1/teams/:teamId/sub-items/:itemId/progress

**Auth**: team member (assignee or any member)

#### Request
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| completion | float | ✓ | 0-100，不能低于上一条记录（PM 修正除外） |
| achievement | string | | 成果 |
| blocker | string | | 卡点 |
| lesson | string | | 经验 |

#### Response (201)
ProgressRecord object.

#### Error Responses
| Status | Code | Description |
|--------|------|-------------|
| 422 | PROGRESS_REGRESSION | 完成度低于上一条记录 |

---

### GET /api/v1/teams/:teamId/sub-items/:itemId/progress

**Auth**: team member

#### Response (200)
Array of ProgressRecord objects, ordered by `createdAt` ASC.

---

### PATCH /api/v1/teams/:teamId/progress/:recordId/completion

**Auth**: PM only

#### Request
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| completion | float | ✓ | 修正后的完成度 0-100 |

#### Response (200)
Updated ProgressRecord object with `isPMCorrect: true`.

---

## ItemPool

### POST /api/v1/teams/:teamId/item-pool

**Auth**: any team member

#### Request
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| title | string | ✓ | 标题，max 100 |
| background | string | | 背景 |
| expectedOutput | string | | 预期产出 |

#### Response (201)
ItemPool object. Initial status: `待分配`.

---

### GET /api/v1/teams/:teamId/item-pool

**Auth**: PM or member

#### Query Params
| Param | Type | Description |
|-------|------|-------------|
| status | string | 待分配\|已分配\|已拒绝 |
| page | int | 默认 1 |
| pageSize | int | 默认 20 |

#### Response (200)
Paginated list of ItemPool objects.

---

### GET /api/v1/teams/:teamId/item-pool/:poolId

**Auth**: PM or member

#### Response (200)
Full ItemPool object.

---

### POST /api/v1/teams/:teamId/item-pool/:poolId/assign

**Auth**: PM

#### Request
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| mainItemId | uint | ✓ | 挂载的主事项 ID |
| assigneeId | uint | ✓ | 子事项负责人 |

#### Response (200)
```json
{ "code": 0, "data": { "subItemId": 123 } }
```

Atomically: updates ItemPool status → `已分配`, creates SubItem under mainItemId.

#### Error Responses
| Status | Code | Description |
|--------|------|-------------|
| 422 | ITEM_ALREADY_PROCESSED | 事项已分配或已拒绝 |
| 404 | ITEM_NOT_FOUND | 主事项不存在 |

---

### POST /api/v1/teams/:teamId/item-pool/:poolId/reject

**Auth**: PM

#### Request
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| reason | string | ✓ | 拒绝原因，max 200 |

#### Response (200)
Updated ItemPool object with status `已拒绝`.

---

## Views

### GET /api/v1/teams/:teamId/views/weekly

**Auth**: team member

#### Query Params
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| weekStart | string | ✓ | ISO8601 date (Monday of the week) |

#### Response (200)
```json
{
  "weekStart": "2026-04-13",
  "weekEnd": "2026-04-19",
  "groups": [
    {
      "mainItem": { "id": 1, "title": "...", "completion": 60 },
      "newlyCompleted": [ SubItemWithProgress ],
      "hasProgress": [ SubItemWithProgress ],
      "noChangeFromLastWeek": [ SubItemSummary ]
    }
  ]
}
```

`SubItemWithProgress` includes the SubItem fields + `progressThisWeek` array (records created in the week range).

---

### GET /api/v1/teams/:teamId/views/gantt

**Auth**: team member

#### Query Params
| Param | Type | Description |
|-------|------|-------------|
| status | string | 筛选主事项状态 |

#### Response (200)
```json
{
  "items": [
    {
      "id": 1, "title": "...", "priority": "P1",
      "startDate": "2026-04-01", "expectedEndDate": "2026-04-30",
      "completion": 60, "status": "进行中", "isOverdue": false,
      "subItems": [
        { "id": 10, "title": "...", "startDate": "...", "expectedEndDate": "...", "completion": 80, "status": "待验收" }
      ]
    }
  ]
}
```

SubItems are returned but hidden by default on the frontend (expand on click).

---

### GET /api/v1/teams/:teamId/views/table

**Auth**: team member

#### Query Params
| Param | Type | Description |
|-------|------|-------------|
| type | string | "main"\|"sub"\|"" (all) |
| priority | string | 多值逗号分隔 |
| status | string | 多值逗号分隔 |
| assigneeId | uint | |
| sortBy | string | 字段名 |
| sortOrder | string | "asc"\|"desc" |
| page | int | 默认 1 |
| pageSize | int | 默认 50 |

#### Response (200)
Paginated list of TableRow objects:
| Field | Type | Description |
|-------|------|-------------|
| id | uint | |
| type | string | "main"\|"sub" |
| code | string | 编号 |
| title | string | |
| priority | string | |
| assigneeId | uint | |
| assigneeName | string | |
| status | string | |
| completion | float | |
| expectedEndDate | string | |
| actualEndDate | string | |

---

### GET /api/v1/teams/:teamId/views/table/export

**Auth**: team member

Same query params as table view (no pagination). Returns CSV file.

**Response**: `Content-Type: text/csv`, `Content-Disposition: attachment; filename="items-export.csv"`

#### Error Responses
| Status | Code | Description |
|--------|------|-------------|
| 422 | NO_DATA | 当前筛选条件下无数据 |

---

## Reports

### GET /api/v1/teams/:teamId/reports/weekly/preview

**Auth**: PM or member

#### Query Params
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| weekStart | string | ✓ | ISO8601 date (Monday) |

#### Response (200)
```json
{
  "weekStart": "2026-04-13",
  "weekEnd": "2026-04-19",
  "sections": [
    {
      "mainItem": { "id": 1, "title": "...", "completion": 60 },
      "subItems": [
        {
          "id": 10, "title": "...", "completion": 80,
          "achievements": ["成果1", "成果2"],
          "blockers": ["卡点1"]
        }
      ]
    }
  ]
}
```

#### Error Responses
| Status | Code | Description |
|--------|------|-------------|
| 422 | NO_DATA | 所选周暂无数据 |

---

### GET /api/v1/teams/:teamId/reports/weekly/export

**Auth**: PM or member

Same query params as preview. Returns Markdown file within 5 seconds.

**Response**: `Content-Type: text/markdown`, `Content-Disposition: attachment; filename="weekly-report-2026-W16.md"`

---

## Admin

### GET /api/v1/admin/users

**Auth**: SuperAdmin

#### Query Params
| Param | Type | Description |
|-------|------|-------------|
| page | int | 默认 1 |
| pageSize | int | 默认 50 |

#### Response (200)
Paginated list of:
| Field | Type | Description |
|-------|------|-------------|
| id | uint | |
| username | string | 账号 |
| displayName | string | 姓名 |
| canCreateTeam | bool | 创建团队权限 |
| isSuperAdmin | bool | |

---

### PUT /api/v1/admin/users/:userId/can-create-team

**Auth**: SuperAdmin

#### Request
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| canCreateTeam | bool | ✓ | 授予或撤销权限 |

#### Error Responses
| Status | Code | Description |
|--------|------|-------------|
| 422 | CANNOT_MODIFY_SELF | 不能修改自己的权限 |

---

### GET /api/v1/admin/teams

**Auth**: SuperAdmin

#### Query Params
| Param | Type | Description |
|-------|------|-------------|
| page | int | 默认 1 |
| pageSize | int | 默认 50 |

#### Response (200)
Paginated list of:
| Field | Type | Description |
|-------|------|-------------|
| id | uint | |
| name | string | |
| pm.displayName | string | PM 姓名 |
| memberCount | int | |
| mainItemCount | int | |
| createdAt | string | |

---

## Data Contracts

### MainItem Object
```json
{
  "id": 1,
  "code": "MI-0001",
  "title": "接入新支付渠道",
  "priority": "P1",
  "proposerId": 2,
  "proposerName": "张三",
  "assigneeId": 3,
  "assigneeName": "李四",
  "startDate": "2026-04-01",
  "expectedEndDate": "2026-04-30",
  "actualEndDate": null,
  "status": "进行中",
  "completion": 45.5,
  "isKeyItem": false,
  "delayCount": 0,
  "archivedAt": null,
  "createdAt": "2026-04-01T09:00:00Z",
  "updatedAt": "2026-04-15T14:30:00Z"
}
```

### SubItem Object
```json
{
  "id": 10,
  "mainItemId": 1,
  "title": "对接支付宝 SDK",
  "description": "...",
  "priority": "P2",
  "assigneeId": 3,
  "assigneeName": "李四",
  "startDate": "2026-04-01",
  "expectedEndDate": "2026-04-15",
  "actualEndDate": null,
  "status": "进行中",
  "completion": 60,
  "isKeyItem": false,
  "delayCount": 0,
  "createdAt": "2026-04-01T09:00:00Z",
  "updatedAt": "2026-04-15T14:30:00Z"
}
```

### ProgressRecord Object
```json
{
  "id": 100,
  "subItemId": 10,
  "authorId": 3,
  "authorName": "李四",
  "completion": 60,
  "achievement": "完成了 SDK 初始化和沙箱环境联调",
  "blocker": "正式环境证书申请中",
  "lesson": "沙箱和正式环境配置差异较大，需提前确认",
  "isPMCorrect": false,
  "createdAt": "2026-04-15T18:00:00Z"
}
```

### ItemPool Object
```json
{
  "id": 50,
  "title": "优化首页加载速度",
  "background": "用户反馈首页加载超过 3 秒",
  "expectedOutput": "首页 LCP < 1.5 秒",
  "submitterId": 5,
  "submitterName": "王五",
  "status": "待分配",
  "assignedMainId": null,
  "assignedSubId": null,
  "assigneeId": null,
  "rejectReason": "",
  "reviewedAt": null,
  "createdAt": "2026-04-16T10:00:00Z"
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| UNAUTHORIZED | 401 | JWT 缺失或过期 |
| FORBIDDEN | 403 | 权限不足 |
| NOT_TEAM_MEMBER | 403 | 不是该团队成员 |
| TEAM_NOT_FOUND | 404 | 团队不存在 |
| USER_NOT_FOUND | 404 | 用户不存在 |
| ITEM_NOT_FOUND | 404 | 事项不存在 |
| VALIDATION_ERROR | 400 | 请求参数校验失败 |
| INVALID_STATUS | 422 | 非法状态流转 |
| ARCHIVE_NOT_ALLOWED | 422 | 归档条件不满足 |
| PROGRESS_REGRESSION | 422 | 完成度低于上一条记录 |
| ITEM_ALREADY_PROCESSED | 422 | 事项已处理 |
| CANNOT_REMOVE_SELF | 422 | 不能移除自己 |
| CANNOT_MODIFY_SELF | 422 | 不能修改自己的权限 |
| ALREADY_MEMBER | 422 | 已是团队成员 |
| NO_DATA | 422 | 所选范围无数据 |
| INTERNAL_ERROR | 500 | 服务器内部错误 |
