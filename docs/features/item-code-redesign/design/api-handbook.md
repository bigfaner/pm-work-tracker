---
created: 2026-04-22
related: design/tech-design.md
---

# API Handbook: 事项编码体系重新设计

## API Overview

本次变更涉及以下 API 改动：
1. `POST /v1/teams` — 请求体新增 `code` 字段；响应新增 `code` 字段
2. `GET /v1/teams` — 响应列表项新增 `code` 字段
3. `GET /v1/teams/:teamId` — 响应新增 `code` 字段
4. 主事项和子事项相关接口的响应中，`code` 字段值格式变更（`MI-NNNN` → `TEAM-NNNNN` / `TEAM-NNNNN-NN`）

子事项创建接口（`POST /v1/teams/:teamId/sub-items`）无签名变更，响应中新增 `code` 字段。

---

## Endpoints

### 创建团队

**Method**: `POST`
**Path**: `/v1/teams`
**Auth**: 已登录用户（`team:create` 权限）

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | 是 | 团队名称，max 100 |
| description | string | 否 | 团队描述，max 500 |
| code | string | 是 | 团队编码，2~6 位英文字母，全局唯一 |

```json
{
  "name": "Feature Team",
  "description": "负责新功能开发",
  "code": "FEAT"
}
```

#### Response (201)

| Field | Type | Description |
|-------|------|-------------|
| id | number | 团队 ID |
| name | string | 团队名称 |
| description | string | 团队描述 |
| code | string | 团队编码（新增） |
| pmId | number | PM 用户 ID |
| createdAt | string | 创建时间 RFC3339 |
| updatedAt | string | 更新时间 RFC3339 |

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "name": "Feature Team",
    "description": "负责新功能开发",
    "code": "FEAT",
    "pmId": 42,
    "createdAt": "2026-04-22T10:00:00Z",
    "updatedAt": "2026-04-22T10:00:00Z"
  }
}
```

#### Error Responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | VALIDATION_ERROR | code 为空、长度不符、含非字母字符 |
| 400 | TEAM_CODE_DUPLICATE | code 与已有团队重复 |
| 401 | UNAUTHORIZED | 未登录 |

---

### 获取团队列表

**Method**: `GET`
**Path**: `/v1/teams`
**Auth**: 已登录用户

#### Response (200)

列表项新增 `code` 字段：

| Field | Type | Description |
|-------|------|-------------|
| id | number | 团队 ID |
| name | string | 团队名称 |
| description | string | 团队描述 |
| code | string | 团队编码（新增） |
| pmId | number | PM 用户 ID |
| pmDisplayName | string | PM 显示名 |
| createdAt | string | 创建时间 |
| updatedAt | string | 更新时间 |

---

### 获取团队详情

**Method**: `GET`
**Path**: `/v1/teams/:teamId`
**Auth**: 团队成员

#### Response (200)

响应新增 `code` 字段（与创建团队响应结构一致）。

---

### 创建主事项（响应变更）

**Method**: `POST`
**Path**: `/v1/teams/:teamId/main-items`
**Auth**: 团队成员（`main_item:create` 权限）

请求体无变更。响应中 `code` 字段值格式从 `MI-NNNN` 变为 `{TEAM_CODE}-NNNNN`：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "code": "FEAT-00001",
    ...
  }
}
```

---

### 创建子事项（响应变更）

**Method**: `POST`
**Path**: `/v1/teams/:teamId/sub-items`
**Auth**: 团队成员（`main_item:update` 权限）

请求体无变更。响应新增 `code` 字段：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "code": "FEAT-00001-01",
    "mainItemId": 1,
    ...
  }
}
```

---

## Data Contracts

### Team 对象（含 code 字段）

```typescript
interface Team {
  id: number
  name: string
  description: string
  code: string        // 新增，如 "FEAT"
  pmId: number
  createdAt: string
  updatedAt: string
}
```

### MainItem 对象（code 格式变更）

```typescript
interface MainItem {
  id: number
  code: string        // 格式从 "MI-0001" 变为 "FEAT-00001"
  teamId: number
  title: string
  // ... 其他字段不变
}
```

### SubItem 对象（新增 code 字段）

```typescript
interface SubItem {
  id: number
  code: string        // 新增，如 "FEAT-00001-01"
  mainItemId: number
  teamId: number
  title: string
  // ... 其他字段不变
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| TEAM_CODE_DUPLICATE | 400 | 团队编码已被其他团队使用 |
| VALIDATION_ERROR | 400 | 请求参数校验失败（含 code 格式不合法） |
