---
created: 2026-05-12
related: design/tech-design.md
---

# API Handbook: 里程碑图

## API Overview

里程碑图功能新增两组 RESTful API：MilestoneMap（里程碑图 CRUD + 状态切换）和 Milestone（里程碑 CRUD + 状态切换）。所有端点嵌套在团队路由下，受 RBAC 权限控制。

权限码：
- `milestone:create` — 创建里程碑图和里程碑
- `milestone:read` — 查看里程碑图和里程碑
- `milestone:update` — 编辑里程碑图/里程碑信息及状态切换
- `milestone:delete` — 删除里程碑图和里程碑

---

## MilestoneMap Endpoints

### Create MilestoneMap

**Method**: `POST`
**Path**: `/api/v1/teams/:teamId/milestone-maps`
**Auth**: `milestone:create`

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| mapName | string | Yes | 里程碑图名称，1-100 字符 |
| mapDesc | string | No | 里程碑图描述 |
| assigneeBizKey | int64 | Yes | 负责人 bizKey |
| planStartDate | string | No | 计划开始时间 (YYYY-MM-DD) |
| expectedEndDate | string | No | 计划完成时间 (YYYY-MM-DD) |

#### Response (201)

| Field | Type | Description |
|-------|------|-------------|
| bizKey | string | 里程碑图 bizKey |
| teamKey | string | 所属团队 bizKey |
| mapName | string | 里程碑图名称 |
| mapDesc | string | 描述 |
| creatorKey | string | 创建者 bizKey |
| creatorName | string | 创建者姓名 |
| assigneeKey | string | 负责人 bizKey |
| assigneeName | string | 负责人姓名 |
| planStartDate | string | 计划开始时间 |
| expectedEndDate | string | 计划完成时间 |
| mapStatus | string | 状态码 |
| statusName | string | 状态显示名 |
| milestoneCount | number | 里程碑数量（0） |
| itemCount | number | 关联事项数量（0） |
| overallProgress | number | 整体进度（0） |
| createTime | string | 创建时间 |
| dbUpdateTime | string | 更新时间 |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | INVALID_PARAMS | 名称校验失败（空/超长） |
| 403 | FORBIDDEN | 无 milestone:create 权限 |

---

### List MilestoneMaps

**Method**: `GET`
**Path**: `/api/v1/teams/:teamId/milestone-maps`
**Auth**: `milestone:read`

#### Query Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | No | 按名称模糊搜索 |
| assigneeKey | string | No | 按负责人筛选，bizKey |
| status | string | No | 按状态筛选：planning/reviewed/ready/executing/completed/cancelled |
| page | number | No | 页码，默认 1 |
| pageSize | number | No | 每页数量，默认 20 |

#### Response (200)

| Field | Type | Description |
|-------|------|-------------|
| items | MilestoneMap[] | 里程碑图列表 |
| total | number | 总数 |
| page | number | 当前页 |
| size | number | 每页数量 |

每个 MilestoneMap 包含 Create 响应中的所有字段，外加 computed 字段（milestoneCount, itemCount, overallProgress）。

---

### Get MilestoneMap

**Method**: `GET`
**Path**: `/api/v1/teams/:teamId/milestone-maps/:mapId`
**Auth**: `milestone:read`

#### Response (200)

同 Create 响应结构，加上所有里程碑的摘要信息。

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 404 | NOT_FOUND | 里程碑图不存在 |

---

### Update MilestoneMap

**Method**: `PUT`
**Path**: `/api/v1/teams/:teamId/milestone-maps/:mapId`
**Auth**: `milestone:update`

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| mapName | string | No | 里程碑图名称，1-100 字符 |
| mapDesc | string | No | 描述 |
| assigneeBizKey | *int64 | No | 负责人 bizKey |
| planStartDate | *string | No | 计划开始时间 |
| expectedEndDate | *string | No | 计划完成时间 |

所有字段可选，仅更新传入的字段。

#### Response (200)

更新后的 MilestoneMap。

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | INVALID_PARAMS | 名称校验失败 |
| 404 | NOT_FOUND | 里程碑图不存在 |
| 409 | CONFLICT | 数据已被其他人修改，请刷新后重试 |

---

### Delete MilestoneMap

**Method**: `DELETE`
**Path**: `/api/v1/teams/:teamId/milestone-maps/:mapId`
**Auth**: `milestone:delete`

#### Response (200)

```json
{ "message": "deleted" }
```

**副作用**: 同一事务内软删除该图下所有里程碑，并解绑所有关联 MI（milestone_key 置空）。

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | INVALID_PARAMS | 里程碑图状态不允许删除（仅 planning/reviewed/ready 可删除） |
| 404 | NOT_FOUND | 里程碑图不存在 |

---

### Change MilestoneMap Status

**Method**: `PUT`
**Path**: `/api/v1/teams/:teamId/milestone-maps/:mapId/status`
**Auth**: `milestone:update`

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| status | string | Yes | 目标状态 |

#### Response (200)

更新后的 MilestoneMap。

**副作用**: 切换至 cancelled 时，级联取消所有非终态里程碑并自动解绑所有关联 MI（BR-6）。

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | BAD_REQUEST | 里程碑图下存在未完成的里程碑，无法标记为已完成（BR-2） |
| 422 | INVALID_STATUS | 无效的状态转换 |

---

### Get MilestoneMap Available Transitions

**Method**: `GET`
**Path**: `/api/v1/teams/:teamId/milestone-maps/:mapId/available-transitions`
**Auth**: `milestone:read`

#### Response (200)

```json
{"transitions": ["reviewed", "cancelled"]}
```

---

## Milestone Endpoints

### Create Milestone

**Method**: `POST`
**Path**: `/api/v1/teams/:teamId/milestone-maps/:mapId/milestones`
**Auth**: `milestone:create`

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| milestoneName | string | Yes | 里程碑名称，1-100 字符 |
| milestoneDesc | string | No | 描述 |
| expectedEndDate | string | Yes | 计划完成时间，ISO 8601 日期 |

#### Response (201)

| Field | Type | Description |
|-------|------|-------------|
| bizKey | string | 里程碑 bizKey |
| teamKey | string | 所属团队 bizKey |
| milestoneMapKey | string | 所属里程碑图 bizKey |
| milestoneName | string | 里程碑名称 |
| milestoneDesc | string | 描述 |
| expectedEndDate | string | 计划完成时间 |
| milestoneStatus | string | 状态码 |
| statusName | string | 状态显示名 |
| completion | number | 完成度（0） |
| relatedMICount | number | 关联事项数量（0） |
| createTime | string | 创建时间 |
| dbUpdateTime | string | 更新时间 |

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | INVALID_PARAMS | 名称/日期校验失败 |
| 400 | MAP_IS_TERMINAL | 里程碑图处于终态（completed/cancelled），不可创建里程碑 |
| 404 | NOT_FOUND | 所属里程碑图不存在 |
| 409 | DUPLICATE_NAME | 同一里程碑图下已存在同名里程碑 |

---

### List Milestones (by Map)

**Method**: `GET`
**Path**: `/api/v1/teams/:teamId/milestone-maps/:mapId/milestones`
**Auth**: `milestone:read`

#### Response (200)

```json
{
  "items": [Milestone],
  "total": 5
}
```

非分页，返回该里程碑图下所有未删除里程碑，按 expected_end_date 升序排列。

---

### List Milestones (by Team)

**Method**: `GET`
**Path**: `/api/v1/teams/:teamId/milestones`
**Auth**: `milestone:read`

用于 UF-4（事项清单筛选）、UF-5（编辑选择器）、UF-6（表格列）、UF-1 时间线视图筛选。

#### Query Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | No | 按名称模糊搜索 |
| status | string | No | 按状态筛选：not_started/in_progress/completed/cancelled |
| excludeCancelled | boolean | No | 是否排除已取消里程碑，默认 true |

#### Response (200)

```json
{
  "items": [Milestone],
  "total": 12
}
```

---

### Get Milestone

**Method**: `GET`
**Path**: `/api/v1/teams/:teamId/milestones/:milestoneId`
**Auth**: `milestone:read`

#### Response (200)

Milestone 详情，含关联 MI 列表摘要。

---

### Update Milestone

**Method**: `PUT`
**Path**: `/api/v1/teams/:teamId/milestones/:milestoneId`
**Auth**: `milestone:update`

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| milestoneName | string | No | 里程碑名称 |
| milestoneDesc | *string | No | 描述 |
| expectedEndDate | string | No | 计划完成时间 |

#### Response (200)

更新后的 Milestone。

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | INVALID_PARAMS | 名称校验失败 |
| 404 | NOT_FOUND | 里程碑不存在 |
| 409 | CONFLICT | 数据已被其他人修改 |

---

### Delete Milestone

**Method**: `DELETE`
**Path**: `/api/v1/teams/:teamId/milestones/:milestoneId`
**Auth**: `milestone:delete`

#### Response (200)

```json
{ "message": "deleted" }
```

**副作用**: 同一事务内解绑所有关联 MI（milestone_key 置空）。

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | INVALID_PARAMS | 里程碑状态不允许删除（仅 not_started/cancelled 可删除） |
| 404 | NOT_FOUND | 里程碑不存在 |

---

### Change Milestone Status

**Method**: `PUT`
**Path**: `/api/v1/teams/:teamId/milestones/:milestoneId/status`
**Auth**: `milestone:update`

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| status | string | Yes | 目标状态 |

#### Response (200)

更新后的 Milestone。

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | BAD_REQUEST | 里程碑下存在未完成的事项，无法标记为已完成（BR-1） |
| - | Side Effect | 切换至 cancelled 时，自动解绑所有关联 MI |
| 422 | INVALID_STATUS | 无效的状态转换 |

---

### Get Milestone Available Transitions

**Method**: `GET`
**Path**: `/api/v1/teams/:teamId/milestones/:milestoneId/available-transitions`
**Auth**: `milestone:read`

#### Response (200)

```json
{"transitions": ["in_progress", "cancelled"]}
```

---

## Data Contracts

### MilestoneMapStatus

| Code | Display | Terminal |
|------|---------|----------|
| `planning` | 规划中 | No |
| `reviewed` | 已评审 | No |
| `ready` | 待实施 | No |
| `executing` | 实施中 | No |
| `completed` | 已完成 | Yes |
| `cancelled` | 已取消 | Yes |

### MilestoneStatus

| Code | Display | Terminal |
|------|---------|----------|
| `not_started` | 未开始 | No |
| `in_progress` | 进行中 | No |
| `completed` | 已完成 | Yes |
| `cancelled` | 已取消 | Yes |

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| INVALID_PARAMS | 400 | 请求参数校验失败 |
| MAP_IS_TERMINAL | 400 | 里程碑图处于终态，不可创建/修改里程碑 |
| DUPLICATE_NAME | 409 | 同一里程碑图下已存在同名里程碑 |
| INVALID_STATUS | 422 | 状态转换不合法 |
| NOT_FOUND | 404 | 资源不存在或已删除 |
| CONFLICT | 409 | 并发冲突（里程碑编辑） |
| FORBIDDEN | 403 | 无权限 |
