---
feature: "weekly-stats-optimization"
---

# API Handbook: 每周进展统计优化

## GET /api/v1/teams/:teamId/views/weekly

### Request

```
GET /api/v1/teams/:teamId/views/weekly?weekStart=YYYY-MM-DD
Authorization: Bearer <token>
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `teamId` | path | ✓ | 团队 ID |
| `weekStart` | query | ✓ | 周一日期，格式 `YYYY-MM-DD` |

### Response (变更后)

```json
{
  "code": 0,
  "data": {
    "weekStart": "2026-04-21",
    "weekEnd": "2026-04-27",
    "stats": {
      "activeSubItems": 12,
      "newlyCompleted": 3,
      "inProgress": 5,
      "blocked": 1,
      "pending": 2,
      "pausing": 1,
      "overdue": 4
    },
    "groups": [ ... ]
  }
}
```

### stats 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `activeSubItems` | int | 本周活跃子事项总数（含新完成） |
| `newlyCompleted` | int | 本周新完成（actualEndDate ∈ [weekStart, weekEnd]） |
| `inProgress` | int | status=progressing 且本周活跃 |
| `blocked` | int | status=blocking 且本周活跃 |
| `pending` | int | status=pending 且本周活跃（新增） |
| `pausing` | int | status=pausing 且本周活跃（新增） |
| `overdue` | int | expectedEndDate < weekEnd AND status ∉ {completed,closed} AND 本周活跃（新增） |

### 向后兼容性

新增的 3 个字段（`pending`、`pausing`、`overdue`）在旧客户端中会被忽略。Go struct 零值保证即使计算逻辑未触发，字段也会序列化为 `0` 而非 `null`。

### Error Responses

| Status | Code | 场景 |
|--------|------|------|
| 400 | `VALIDATION_ERROR` | weekStart 格式错误或非周一 |
| 400 | `VALIDATION_ERROR` | weekStart 在未来 |
| 401 | `UNAUTHORIZED` | token 无效或缺失 |
| 403 | `FORBIDDEN` | 用户不属于该团队 |
