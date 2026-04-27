---
created: 2026-04-27
related: design/tech-design.md
---

# API Handbook: Decision Log

## API Overview

Decision Log API provides CRUD operations for decision records attached to main items. Routes are nested under the main-items resource. Write operations require `main_item:update` permission; read operations require team membership only.

## Endpoints

### Create Decision Log

**Method**: `POST`
**Path**: `/api/v1/teams/:teamId/main-items/:mainId/decision-logs`
**Auth**: `main_item:update`

Creates a new decision log entry. Can be saved as draft or published immediately.

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| category | string | Yes | One of: `technical`, `resource`, `requirement`, `schedule`, `risk`, `other` |
| tags | string[] | No | Free-form tags. Each tag ≤ 20 chars. Sent as JSON array. |
| content | string | Yes | Decision content. ≤ 2000 chars. |
| logStatus | string | Yes | `"draft"` or `"published"` |

#### Request Example

```json
{
  "category": "technical",
  "tags": ["缓存策略", "性能优化"],
  "content": "决定采用 Redis 缓存热点数据，预期命中率 > 80%",
  "logStatus": "draft"
}
```

#### Response (201)

| Field | Type | Description |
|-------|------|-------------|
| bizKey | string | Unique identifier for this decision log |
| mainItemKey | string | Parent main item's BizKey |
| category | string | Predefined category |
| tags | string[] | Free-form tags |
| content | string | Full decision content |
| logStatus | string | `"draft"` or `"published"` |
| createdBy | string | Creator's user BizKey |
| creatorName | string | Creator's display name |
| createTime | string | ISO 8601 (RFC3339) |
| updateTime | string | ISO 8601 (RFC3339) |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | VALIDATION_ERROR | Missing/invalid category, content, or status |
| 404 | ITEM_NOT_FOUND | Parent main item doesn't exist |

---

### List Decision Logs

**Method**: `GET`
**Path**: `/api/v1/teams/:teamId/main-items/:mainId/decision-logs`
**Auth**: Team membership (no specific permission)

Returns published decisions (all team members) + current user's drafts, ordered by `createTime` DESC.

#### Request

No request body. Parameters from path and auth context.

#### Response (200)

| Field | Type | Description |
|-------|------|-------------|
| items | DecisionLog[] | Array of decision log entries |
| total | number | Total count of matching entries |
| page | number | Current page number |
| size | number | Page size |

#### Response Example

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "bizKey": "1893456789012345678",
        "mainItemKey": "1893456789012345001",
        "category": "technical",
        "tags": ["缓存策略", "性能优化"],
        "content": "决定采用 Redis 缓存热点数据...",
        "logStatus": "published",
        "createdBy": "1893456789012345901",
        "creatorName": "张三",
        "createTime": "2026-04-27T14:30:00+08:00",
        "updateTime": "2026-04-27T14:30:00+08:00"
      }
    ],
    "total": 1,
    "page": 1,
    "size": 20
  }
}
```

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 404 | ITEM_NOT_FOUND | Parent main item doesn't exist |

---

### Update Decision Log (Draft Only)

**Method**: `PUT`
**Path**: `/api/v1/teams/:teamId/main-items/:mainId/decision-logs/:id`
**Auth**: `main_item:update`

Updates a draft decision log. Only the creator can update their own drafts. Published decisions cannot be updated.

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| category | string | Yes | One of: `technical`, `resource`, `requirement`, `schedule`, `risk`, `other` |
| tags | string[] | No | Free-form tags. Each tag ≤ 20 chars. |
| content | string | Yes | Decision content. ≤ 2000 chars. |

#### Response (200)

Same structure as Create response (updated decision log VO).

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | VALIDATION_ERROR | Missing/invalid category or content |
| 403 | FORBIDDEN | Decision is published, or caller is not the creator |
| 404 | DECISION_LOG_NOT_FOUND | Decision log doesn't exist |

---

### Publish Decision Log

**Method**: `PATCH`
**Path**: `/api/v1/teams/:teamId/main-items/:mainId/decision-logs/:id/publish`
**Auth**: `main_item:update`

Transitions a draft decision to published status. Only the creator can publish their own draft. Once published, the decision becomes immutable.

#### Request

No request body.

#### Response (200)

Same structure as Create response (with `logStatus: "published"`).

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 403 | FORBIDDEN | Decision is already published, or caller is not the creator |
| 404 | DECISION_LOG_NOT_FOUND | Decision log doesn't exist |

---

## Data Contracts

### DecisionLog (shared response shape)

| Field | Type | Description |
|-------|------|-------------|
| bizKey | string | Snowflake ID (string-formatted int64) |
| mainItemKey | string | Parent main item's BizKey |
| category | string | Enum: technical, resource, requirement, schedule, risk, other |
| tags | string[] | Free-form tags (parsed from JSON storage) |
| content | string | Full decision text |
| logStatus | string | Enum: draft, published |
| createdBy | string | Creator's user BizKey |
| creatorName | string | Creator's display name |
| createTime | string | RFC3339 timestamp |
| updateTime | string | RFC3339 timestamp |

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| VALIDATION_ERROR | 400 | Invalid or missing request fields |
| FORBIDDEN | 403 | Insufficient permissions or ownership mismatch |
| ITEM_NOT_FOUND | 404 | Parent main item not found |
| DECISION_LOG_NOT_FOUND | 404 | Decision log not found |
