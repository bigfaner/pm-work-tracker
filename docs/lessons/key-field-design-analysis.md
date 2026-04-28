# Key 字段设计分析与问题排查

## 问题背景

用户报告：普通用户被授予子事项的所有权限后，切换子事项状态时报错 403 `{code: "FORBIDDEN", message: "insufficient permissions"}`。

## 根因分析

### 问题定位

`pmw_team_members.role_key` 字段存储了错误的值：
- **正确值**：角色的 `biz_key`（雪花 ID）
- **实际值**：部分记录存储了自增 `id`

### 数据流分析

```
前端传入 roleKey: "2048990585057972224" (bizKey 字符串)
    ↓
Handler: pkg.ParseID() → int64
    ↓
Service: 直接存储到 RoleKey 字段
    ↓
TeamScopeMiddleware: roleRepo.FindByBizKey(roleKey) 查询失败
    ↓
permCodes 为空 → 权限检查失败 → 403
```

### 脏数据来源

`tests/integration/helpers.go:596-601` 的 `findRoleIDByName` 函数返回 `role.ID`（自增 ID），而非 `role.BizKey`。

## 当前设计分析

### 数据库 JOIN 关系

```sql
-- team_repo.go:149-151
LEFT JOIN pmw_users ON pmw_users.id = pmw_team_members.user_key
LEFT JOIN pmw_roles ON pmw_roles.biz_key = pmw_team_members.role_key
LEFT JOIN pmw_teams ON pmw_teams.id = pmw_team_members.team_key
```

### 字段设计现状

| 字段 | 存储内容 | JOIN 方式 | 设计一致性 |
|------|----------|-----------|------------|
| `team_key` | 自增 ID | `pmw_teams.id = team_key` | ❌ 不一致 |
| `user_key` | 自增 ID | `pmw_users.id = user_key` | ❌ 不一致 |
| `pm_key` | 自增 ID | `pmw_users.id = pm_key` | ❌ 不一致 |
| `role_key` | BizKey | `pmw_roles.biz_key = role_key` | ✅ 一致 |
| `assignee_key` | BizKey | 前端传入 BizKey | ✅ 一致 |
| `proposer_key` | 自增 ID | 内部设置 | ❌ 不一致 |
| `author_key` | 自增 ID | 内部设置 | ❌ 不一致 |
| `main_item_key` | 自增 ID | `pmw_main_items.id = main_item_key` | ❌ 不一致 |
| `submitter_key` | 自增 ID | 内部设置 | ❌ 不一致 |

### 问题：设计不一致

**理论上，所有 `*Key` 字段都应该存储 `biz_key`**，原因：

1. **对外暴露**：前端使用 `biz_key`（雪花 ID）作为唯一标识
2. **数据迁移**：自增 ID 在不同环境间不一致，`biz_key` 全局唯一
3. **安全性**：自增 ID 可预测，`biz_key` 不可预测
4. **一致性**：避免混淆，统一使用 `biz_key`

### 为什么 `role_key` 必须存储 BizKey？

因为 `role_key` 是**用户可配置**的字段：
- 前端选择角色时传入 `roleKey: "204899..."`（BizKey 字符串）
- 后端直接存储这个 BizKey
- 查询时通过 `pmw_roles.biz_key = role_key` 关联

如果 `role_key` 存储自增 ID，会导致：
- 前端传入的 BizKey 无法直接使用
- 需要额外的查询将 BizKey 转换为 ID
- 与其他 `*Key` 字段的处理方式不一致

### 为什么其他 `*Key` 字段存储自增 ID 仍能工作？

因为代码中有一层转换：

```go
// Handler 层：将前端传入的 BizKey 解析为内部 ID
teamID := middleware.GetTeamID(c)  // 已经是解析后的 ID
userID := middleware.GetUserID(c)  // 已经是解析后的 ID

// Service 层：直接使用内部 ID
item := &model.MainItem{
    TeamKey: int64(teamID),  // 存储内部 ID
}
```

**但这违反了设计原则**：
- 字段名是 `*_key`，暗示应该存储 BizKey
- 实际存储的是自增 ID
- 命名与实际行为不符，容易产生混淆

## 影响范围评估

### 当前代码行为

| 入口 | 前端传入 | 存储值 | 是否正确 |
|------|----------|--------|----------|
| `InviteMember` | `roleKey: "204899..."` | BizKey | ✅ |
| `UpdateMemberRole` | `roleKey: "204899..."` | BizKey | ✅ |
| `CreateTeam` | 无（后端查找） | BizKey | ✅ |
| `TransferPM` | 无（后端查找） | BizKey | ✅ |
| `CreateMainItem` | `teamId` (BizKey) | ID (转换后) | ⚠️ 设计不一致 |
| `CreateSubItem` | `teamId` (BizKey) | ID (转换后) | ⚠️ 设计不一致 |
| 集成测试 | 直接创建 | ~~ID~~ → BizKey | ✅ 已修复 |

## 修复方案

### 已完成

1. **修复 `findRoleKeyByName`**：返回 `BizKey` 而非 `ID`
2. **修复测试用例**：`internal/model/role_test.go` 使用 `r.BizKey`
3. **清理脏数据**：更新数据库中 16 条错误记录

### 待评估：统一 `*Key` 字段设计

**方案 A：保持现状（混合模式）**
- `role_key`、`assignee_key` 存储 BizKey（用户可配置）
- `team_key`、`user_key`、`pm_key` 等存储自增 ID（内部设置）
- 优点：无需修改，现有代码已能工作
- 缺点：设计不一致，容易混淆

**方案 B：统一存储 `biz_key`**
- 所有 `*Key` 字段都存储 BizKey
- 修改 JOIN 查询为 `pmw_teams.biz_key = team_key`
- 优点：设计一致，符合对外暴露原则
- 缺点：需要修改大量代码和数据库

**建议**：采用方案 B，统一所有 `*Key` 字段存储 BizKey。

## 设计规范建议

### 命名约定

| 字段类型 | 命名 | 存储内容 | 说明 |
|----------|------|----------|------|
| 外键引用 | `*_key` | `biz_key` | 如 `role_key`、`team_key`、`user_key` |
| 内部引用 | `*_id` | 自增 ID | 如 `role_id`、`team_id`（如果需要） |

### 代码规范

```go
// ✅ 正确：存储 biz_key
member.RoleKey = &role.BizKey
member.TeamKey = team.BizKey
member.UserKey = user.BizKey

// ❌ 错误：存储自增 ID
member.RoleKey = &role.ID
member.TeamKey = int64(team.ID)
member.UserKey = int64(user.ID)
```

### JOIN 查询规范

```sql
-- ✅ 正确：使用 biz_key 关联
LEFT JOIN pmw_teams ON pmw_teams.biz_key = pmw_team_members.team_key

-- ❌ 错误：使用 id 关联
LEFT JOIN pmw_teams ON pmw_teams.id = pmw_team_members.team_key
```

### 测试规范

```go
// ✅ 正确：使用 biz_key
pmRoleKey := findRoleKeyByName(t, db, "pm")  // 返回 BizKey

// ❌ 错误：使用自增 ID
pmRoleID := findRoleIDByName(t, db, "pm")   // 返回 ID
```

## 迁移计划（方案 B）

### 阶段 1：修改代码

1. 修改所有 Service 层，存储 BizKey 而非 ID
2. 修改所有 JOIN 查询，使用 `biz_key` 关联
3. 更新所有测试用例

### 阶段 2：数据迁移

```sql
-- 迁移 team_key
UPDATE pmw_team_members tm
SET team_key = (SELECT biz_key FROM pmw_teams WHERE id = tm.team_key);

-- 迁移 user_key
UPDATE pmw_team_members tm
SET user_key = (SELECT biz_key FROM pmw_users WHERE id = tm.user_key);

-- 迁移 pm_key
UPDATE pmw_teams t
SET pm_key = (SELECT biz_key FROM pmw_users WHERE id = t.pm_key);

-- 其他表类似...
```

### 阶段 3：验证

1. 运行所有测试
2. 手动测试关键功能
3. 检查数据一致性

## 相关文件

- `backend/internal/model/team.go` - TeamMember 模型定义
- `backend/internal/model/main_item.go` - MainItem 模型定义
- `backend/internal/model/sub_item.go` - SubItem 模型定义
- `backend/internal/model/item_pool.go` - ItemPool 模型定义
- `backend/internal/model/progress_record.go` - ProgressRecord 模型定义
- `backend/internal/middleware/team_scope.go` - 权限加载逻辑
- `backend/internal/repository/gorm/team_repo.go` - JOIN 查询
- `backend/tests/integration/helpers.go` - 测试辅助函数
