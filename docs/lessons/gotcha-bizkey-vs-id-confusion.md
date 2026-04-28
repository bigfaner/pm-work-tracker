# BizKey vs ID Confusion: A Systemic Bug Pattern

## Problem

数据库中出现脏数据：`pmw_team_members.role_key` 字段存储了角色的内部自增 ID（如 `3`），而不是雪花 BizKey（如 `2048990585057972224`）。这导致权限查询失败，所有非超级管理员用户收到 403 错误。

**症状表现：**
- `record not found` 错误：`SELECT * FROM pmw_roles WHERE biz_key = 3`
- 权限检查失败，普通用户无法访问任何团队资源
- JOIN 查询返回空结果

## Root Cause

这是一个**系统性设计缺陷**，涉及多个层面的混淆：

### 1. 设计层面的混淆

**数据模型设计不一致：**
- `pmw_team_members.role_key` 字段设计为存储角色的 **BizKey**（雪花ID）
- 但代码中多处错误地使用了 `role.ID`（内部自增ID）

**为什么会这样？**
- RBAC 迁移时，`getRoleIDMap` 返回的是 `role.ID`，而不是 `role.BizKey`
- 迁移脚本 `rebuildTeamMembersTable` 直接使用了这个 map，导致历史数据写入错误的值

### 2. 代码层面的混淆

**三个关键错误点：**

```go
// ❌ 错误 1: CreateTeam - 使用 ID 而不是 BizKey
if pmRole, err := s.roleRepo.FindByName(ctx, "pm"); err == nil {
    roleKey := int64(pmRole.ID)  // 应该是 pmRole.BizKey
    member.RoleKey = &roleKey
}

// ❌ 错误 2: TransferPM - 新PM角色赋值
roleKey := int64(pmRole.ID)  // 应该是 pmRole.BizKey
newPMMember.RoleKey = &roleKey

// ❌ 错误 3: TransferPM - 旧PM角色赋值
roleKey := int64(memberRole.ID)  // 应该是 memberRole.BizKey
oldPMMember.RoleKey = &roleKey
```

### 3. 查询层面的混淆

**JOIN 条件错误：**
```go
// ❌ 错误: 用 role_key (BizKey) JOIN role.id (内部ID)
Joins("JOIN pmw_role_permissions ON pmw_role_permissions.role_id = pmw_team_members.role_key")

// ✅ 正确: 用 role_key (BizKey) JOIN role.biz_key (BizKey)
Joins("JOIN pmw_roles ON pmw_roles.biz_key = pmw_team_members.role_key")
```

### 4. 中间件层面的混淆

**权限加载逻辑错误：**
```go
// ❌ 错误: 把 BizKey 当作内部 ID 使用
codes, err := roleRepo.ListPermissions(ctx, uint(*member.RoleKey))

// ✅ 正确: 先通过 BizKey 查找角色，获取内部 ID
role, err := roleRepo.FindByBizKey(ctx, *member.RoleKey)
codes, err := roleRepo.ListPermissions(ctx, role.ID)
```

### 5. 因果链分析

```
症状: 权限查询返回空，用户收到 403
  ↓
直接原因: JOIN 条件不匹配 (BizKey vs ID)
  ↓
根本原因: 数据存储时就存错了值 (ID 而不是 BizKey)
  ↓
触发条件: 
  1. 迁移脚本 getRoleIDMap 返回 ID
  2. CreateTeam/TransferPM 使用 role.ID
  3. InviteMember 前端传值正确，但历史数据已错
```

## Solution

### 1. 代码修复

**统一使用 BizKey：**
```go
// ✅ 正确模式
if pmRole, err := s.roleRepo.FindByName(ctx, "pm"); err == nil {
    roleKey := pmRole.BizKey  // 使用 BizKey
    member.RoleKey = &roleKey
}
```

**JOIN 查询修正：**
```go
// ✅ 正确模式
Joins("JOIN pmw_roles ON pmw_roles.biz_key = pmw_team_members.role_key")
```

**中间件权限加载修正：**
```go
// ✅ 正确模式
role, err := roleRepo.FindByBizKey(ctx, *member.RoleKey)
codes, err := roleRepo.ListPermissions(ctx, role.ID)
```

### 2. 数据修复

```sql
-- 修复脏数据：将内部 ID 转换为 BizKey
UPDATE pmw_team_members 
SET role_key = (SELECT biz_key FROM pmw_roles WHERE pmw_roles.id = pmw_team_members.role_key)
WHERE role_key < 1000000000000000000 
  AND EXISTS (SELECT 1 FROM pmw_roles WHERE pmw_roles.id = pmw_team_members.role_key);
```

### 3. 测试修复

**测试必须使用真实的 BizKey：**
```go
// ❌ 错误: 测试中使用 ID
r := seedRole(t, db, "member", "Member role", true)
member.RoleKey = &r.ID  // ID 和 BizKey 相同，掩盖了 bug

// ✅ 正确: 测试中 BizKey ≠ ID
r.BizKey = int64(r.ID) + 1000  // 强制 BizKey ≠ ID
member.RoleKey = &r.BizKey     // 使用 BizKey
```

## Key Takeaway

### 核心原则

**永远不要混淆内部 ID 和外部 BizKey：**

1. **存储层**：外键字段存储 BizKey（雪花ID），便于跨系统引用
2. **查询层**：JOIN 时必须匹配相同类型的键（BizKey JOIN BizKey）
3. **业务层**：API 请求/响应使用 BizKey，内部查询先转换为 ID

### 检查清单

当涉及外键引用时，问自己三个问题：

1. **这个字段存储的是什么？** (BizKey 还是 ID)
2. **JOIN 条件是否匹配？** (BizKey JOIN BizKey, ID JOIN ID)
3. **测试是否覆盖了 BizKey ≠ ID 的场景？**

### 防御性编程

```go
// 添加断言或日志，帮助早期发现问题
if member.RoleKey != nil && *member.RoleKey < 1000000000000000000 {
    log.Warn("role_key appears to be internal ID, not BizKey")
}
```

### 为什么这个问题如此隐蔽？

1. **测试掩盖问题**：测试中 ID 和 BizKey 往往相同，bug 不易发现
2. **渐进式破坏**：只有新创建的数据会出错，历史数据可能正常
3. **错误传播**：一处错误会导致多处查询失败，难以定位根源
4. **命名混淆**：`role_key` 这个名字没有明确表示是 BizKey 还是 ID

### 最终建议

**在项目规范中明确定义：**
- 所有 `*_key` 字段存储 BizKey（雪花ID）
- 所有 `*_id` 字段存储内部 ID（自增ID）
- JOIN 查询时，确保键类型匹配
- 测试数据必须 BizKey ≠ ID，防止回归
