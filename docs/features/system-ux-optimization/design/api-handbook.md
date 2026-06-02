---
created: "2026-06-02"
related: design/tech-design.md
---

# API Handbook: System UX Optimization Batch

## API Overview

3 个新增 API 端点 + 5 个现有端点增强。所有端点在 `TeamScopeMiddleware` 上下文内（需认证 + 团队成员关系），团队列表端点仅需认证。

## Endpoints

### Delete Main Item

**Method**: `DELETE`
**Path**: `/teams/:teamId/main-items/:itemId`
**Auth**: `main_item:delete`

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| :teamId | path (int64 bizKey) | yes | 团队标识 |
| :itemId | path (int64 bizKey) | yes | 主事项标识 |

#### Response (200)

| Field | Type | Description |
|-------|------|-------------|
| message | string | 固定值 "ok" |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 403 | FORBIDDEN | 无 main_item:delete 权限 |
| 404 | NOT_FOUND | 主事项不存在或已删除 |

---

### Delete Sub Item

**Method**: `DELETE`
**Path**: `/teams/:teamId/sub-items/:subId`
**Auth**: `sub_item:delete`

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| :teamId | path (int64 bizKey) | yes | 团队标识 |
| :subId | path (int64 bizKey) | yes | 子事项标识 |

#### Response (200)

| Field | Type | Description |
|-------|------|-------------|
| message | string | 固定值 "ok" |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 403 | FORBIDDEN | 无 sub_item:delete 权限 |
| 404 | NOT_FOUND | 子事项不存在或已删除 |

---

### Move Sub Item

**Method**: `PUT`
**Path**: `/teams/:teamId/sub-items/:subId/move`
**Auth**: `sub_item:update`

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| :teamId | path (int64 bizKey) | yes | 团队标识 |
| :subId | path (int64 bizKey) | yes | 子事项标识 |
| targetMainItemBizKey | body (string) | yes | 目标主事项 bizKey |

#### Response (200)

| Field | Type | Description |
|-------|------|-------------|
| newSubCode | string | 新编号（如 "T001-007"） |
| mainItemBizKey | string | 目标主事项 bizKey |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | BAD_REQUEST | 目标主事项已关闭 |
| 400 | BAD_REQUEST | 不能移动到同一主事项 |
| 404 | NOT_FOUND | 子事项不存在 |
| 404 | NOT_FOUND | 目标主事项不存在 |

---

### Enhanced List Main Items (过滤穿透 #10)

**Method**: `GET`
**Path**: `/teams/:teamId/main-items`
**Auth**: `main_item:read`

#### Enhanced Query Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| status | string (comma-separated) | no | 多状态筛选，如 `progressing,blocking` |
| assigneeKey | string | no | 负责人 bizKey，触发穿透 |

#### Enhanced Response Fields

每个 MainItem 对象额外返回：

| Field | Type | Description |
|-------|------|-------------|
| matchType | string? | `"direct"` / `"indirect"` / undefined |
| matchedSubItemIds | string[]? | indirect 时仅展示这些子事项 |

---

### Enhanced Gantt View (多状态过滤 #12)

**Method**: `GET`
**Path**: `/teams/:teamId/views/gantt`
**Auth**: `view:gantt`

#### Enhanced Query Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| status | string (comma-separated) | no | 多状态筛选，替代原有单选 status |

---

### Enhanced List Teams (成员过滤 #15)

**Method**: `GET`
**Path**: `/v1/teams`
**Auth**: JWT only

#### Enhanced Behavior

返回结果按当前用户成员关系过滤。仅返回用户所属的团队。SuperAdmin 返回所有团队。

---

## Data Contracts

### MoveResult

```json
{
  "newSubCode": "T001-007",
  "mainItemBizKey": "729384756"
}
```

### MainItemMatchInfo

```json
{
  "matchType": "indirect",
  "matchedSubItemIds": ["123456", "789012"]
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| BAD_REQUEST | 400 | 业务校验失败（目标已关闭、同一主事项） |
| FORBIDDEN | 403 | 无权限 |
| NOT_FOUND | 404 | 资源不存在或已删除 |
| INVALID_STATUS | 422 | 状态流转不合法 |
| CONFLICT | 409 | 并发冲突 |
