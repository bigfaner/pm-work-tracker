# BizKey vs ID 命名约定

## 决策

所有外键字段统一使用 `*_key` 后缀存储 BizKey（雪花ID），`*_id` 后缀存储内部 ID（自增ID）。

## 背景

在 RBAC 实现过程中，`pmw_team_members.role_key` 字段出现了严重的混淆：
- 字段设计为存储 BizKey，但代码中多处错误使用了 `role.ID`
- 导致权限查询失败，所有非超级管理员用户收到 403 错误
- JOIN 查询条件不匹配（BizKey vs ID），返回空结果

## 约定

### 字段命名

| 后缀 | 存储内容 | 示例 |
|------|----------|------|
| `*_key` | BizKey（雪花ID，int64） | `role_key`, `team_key`, `user_key` |
| `*_id` | 内部 ID（自增ID，uint） | `role_id`, `team_id` |

### 使用规则

1. **存储层**：外键字段存储 BizKey，便于跨系统引用和数据迁移
2. **查询层**：JOIN 时必须匹配相同类型的键
   - `*_key` JOIN `biz_key`
   - `*_id` JOIN `id`
3. **业务层**：API 请求/响应使用 BizKey，内部查询先转换为 ID

### 代码示例

```go
// ✅ 正确：外键字段使用 BizKey
type TeamMember struct {
    RoleKey  *int64  `json:"roleKey"`  // 存储 role.BizKey
    TeamKey  int64   `json:"teamKey"`   // 存储 team.BizKey
    UserKey  int64   `json:"userKey"`   // 存储 user.BizKey
}

// ✅ 正确：JOIN 时匹配 BizKey
Joins("JOIN pmw_roles ON pmw_roles.biz_key = pmw_team_members.role_key")

// ✅ 正确：中间件中先转换 BizKey → ID
role, err := roleRepo.FindByBizKey(ctx, *member.RoleKey)
codes, err := roleRepo.ListPermissions(ctx, role.ID)
```

## 影响

- 所有新表的外键字段必须遵循此约定
- 现有表保持一致，不强制迁移
- 测试数据必须 BizKey ≠ ID，防止回归

## 相关文档

- 经验教训：`docs/lessons/gotcha-bizkey-vs-id-confusion.md`
