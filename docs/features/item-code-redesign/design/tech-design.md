---
created: 2026-04-22
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: 事项编码体系重新设计

## Overview

本次变更涉及三个模型（Team、MainItem、SubItem）、两个 Repository 接口、两个 Service、一个数据迁移脚本，以及前端 6 个页面的展示更新。

核心技术决策：
- `NextCode()` / `NextSubCode()` 采用 **SELECT FOR UPDATE 悲观锁**，在事务内序列化同团队/同主事项的并发创建，消除 race condition
- Team.Code 字段在事项创建时**快照**写入编码，后续 Team.Code 变更不影响已有事项
- 数据迁移在单个 SQL 事务内完成，提供回滚脚本

## Architecture

### Layer Placement

| 层 | 变更内容 |
|----|---------|
| Model | Team 新增 `Code` 字段；MainItem.Code 列扩展；SubItem 新增 `Code` 字段 |
| Repository (interface) | `MainItemRepo.NextCode()` 签名不变，实现改为 SELECT FOR UPDATE；`SubItemRepo` 新增 `NextSubCode()` |
| Service | `TeamService.CreateTeam()` 接收 `Code` 字段；`MainItemService.Create()` / `ItemPoolService.Approve()` 调用路径不变；`SubItemService.Create()` 调用新 `NextSubCode()` |
| Handler | `TeamHandler.Create()` 绑定新 `Code` 字段；`teamToDTO()` 新增 `code` 字段 |
| Migration | 新增 `008_item_code_redesign.sql`：ALTER teams 加 code 列、ALTER main_items 扩展 code 列、ALTER sub_items 加 code 列、重写现有编码 |
| Frontend | 6 个页面编码展示值变更；TeamManagementPage 创建对话框新增 Code 输入框 |

### Component Diagram

```
TeamHandler.Create()
    └─ dto.CreateTeamReq{Name, Description, Code}
    └─ TeamService.CreateTeam()
           └─ TeamRepo.Create(team{Code})
                  └─ teams 表 (code VARCHAR(6) UNIQUE)

MainItemService.Create() / ItemPoolService.Approve()
    └─ MainItemRepo.NextCode(ctx, teamID)  ← SELECT FOR UPDATE on teams row
           └─ returns "{TEAM_CODE}-{seq:05d}"
    └─ MainItemRepo.Create(item{Code})
           └─ main_items 表 (code VARCHAR(12))

SubItemService.Create()
    └─ SubItemRepo.NextSubCode(ctx, mainItemID)  ← SELECT FOR UPDATE on main_items row
           └─ returns "{MAIN_CODE}-{sub_seq:02d}"
    └─ SubItemRepo.Create(sub{Code})
           └─ sub_items 表 (code VARCHAR(15), UNIQUE per main_item_id)
```

### Dependencies

无新增外部依赖。事务锁使用 GORM 现有的 `db.Transaction()` + `db.Set("gorm:query_option", "FOR UPDATE")` 或原生 `db.Raw()`。

## Interfaces

### Repository: MainItemRepo

`NextCode()` 签名不变，实现重写：

```go
// repository/main_item_repo.go — 接口不变
NextCode(ctx context.Context, teamID uint) (string, error)

// repository/gorm/main_item_repo.go — 实现重写
func (r *mainItemRepo) NextCode(ctx context.Context, teamID uint) (string, error) {
    var code string
    err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 锁定团队行，序列化同团队并发
        var team model.Team
        if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&team, teamID).Error; err != nil {
            return err
        }
        // 读取当前最大序号
        var maxSeq *int
        if err := tx.Model(&model.MainItem{}).
            Unscoped().
            Where("team_id = ?", teamID).
            Select("MAX(CAST(SUBSTR(code, LENGTH(?) + 2) AS INTEGER))", team.Code).
            Scan(&maxSeq).Error; err != nil {
            return err
        }
        seq := 1
        if maxSeq != nil {
            seq = *maxSeq + 1
        }
        code = fmt.Sprintf("%s-%05d", team.Code, seq)
        return nil
    })
    return code, err
}
```

### Repository: SubItemRepo

新增 `NextSubCode()` 方法：

```go
// repository/sub_item_repo.go — 接口新增
type SubItemRepo interface {
    // ... 现有方法 ...
    NextSubCode(ctx context.Context, mainItemID uint) (string, error)
}

// repository/gorm/sub_item_repo.go — 实现
func (r *subItemRepo) NextSubCode(ctx context.Context, mainItemID uint) (string, error) {
    var code string
    err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 锁定主事项行，序列化同主事项并发
        var mainItem model.MainItem
        if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&mainItem, mainItemID).Error; err != nil {
            return err
        }
        // 读取当前最大子序号
        var maxSubSeq *int
        if err := tx.Model(&model.SubItem{}).
            Where("main_item_id = ?", mainItemID).
            Select("MAX(CAST(SUBSTR(code, LENGTH(?) + 2) AS INTEGER))", mainItem.Code).
            Scan(&maxSubSeq).Error; err != nil {
            return err
        }
        subSeq := 1
        if maxSubSeq != nil {
            subSeq = *maxSubSeq + 1
        }
        code = fmt.Sprintf("%s-%02d", mainItem.Code, subSeq)
        return nil
    })
    return code, err
}
```

### DTO: CreateTeamReq

```go
// dto/team_dto.go
type CreateTeamReq struct {
    Name        string `json:"name"        binding:"required,max=100"`
    Description string `json:"description" binding:"max=500"`
    Code        string `json:"code"        binding:"required,min=2,max=6,alpha"`
}
```

> `alpha` 是 go-playground/validator 内置 tag，校验仅含字母字符。

### Handler: teamToDTO

```go
func teamToDTO(team *model.Team) gin.H {
    return gin.H{
        "id":          team.ID,
        "name":        team.Name,
        "description": team.Description,
        "code":        team.Code,   // 新增
        "pmId":        team.PmID,
        "createdAt":   team.CreatedAt,
        "updatedAt":   team.UpdatedAt,
    }
}
```

## Data Models

### model.Team

```go
type Team struct {
    BaseModel
    Name        string `gorm:"type:varchar(100);not null"                          json:"name"`
    Description string `gorm:"type:varchar(500)"                                   json:"description"`
    PmID        uint   `gorm:"not null"                                             json:"pmId"`
    Code        string `gorm:"type:varchar(6);not null;uniqueIndex:idx_team_code"  json:"code"`  // 新增
}
```

唯一索引名 `idx_team_code` 与迁移 SQL 保持一致。

### model.MainItem

```go
// Code 列从 varchar(10) 扩展到 varchar(12)
Code string `gorm:"type:varchar(12);not null;uniqueIndex:idx_main_item_team_code" json:"code"`
```

> 联合唯一索引 `idx_main_item_team_code`（team_id + code）已在 migration 006 中创建，本次仅扩展列宽。

### model.SubItem

```go
type SubItem struct {
    BaseModel
    // ... 现有字段 ...
    Code string `gorm:"type:varchar(15);not null;uniqueIndex:idx_sub_item_main_code" json:"code"`  // 新增
}
// uniqueIndex:idx_sub_item_main_code 需配合 main_item_id 构成联合唯一索引
// 实际 GORM tag: uniqueIndex:"idx_sub_item_main_code,composite:main_item_id"
```

### Migration: 008_item_code_redesign.sql

```sql
-- 008_item_code_redesign.sql

-- 1. teams 表新增 code 列
ALTER TABLE teams ADD COLUMN code VARCHAR(6) NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_teams_code ON teams(code);

-- 2. main_items.code 列扩展（SQLite 不支持 MODIFY COLUMN，通过重建表实现）
-- 注：GORM AutoMigrate 会处理列宽扩展，此处记录意图
-- ALTER TABLE main_items MODIFY COLUMN code VARCHAR(12) NOT NULL;

-- 3. sub_items 表新增 code 列
ALTER TABLE sub_items ADD COLUMN code VARCHAR(15) NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_sub_items_main_code ON sub_items(main_item_id, code);

-- 4. 数据迁移：为现有团队设置默认 code（需在应用层或手动执行）
-- 注：teams.code 迁移需要人工确认每个团队的 code 值，不在自动迁移中执行

-- 5. 重写 main_items.code（按 team_id 分组，组内按 id 排序）
-- 此步骤在迁移脚本中通过应用层执行（见 cmd/migrate/main.go）

-- 6. 为 sub_items 生成 code（按 main_item_id 分组，组内按 id 排序）
-- 此步骤在迁移脚本中通过应用层执行（见 cmd/migrate/main.go）
```

**迁移执行顺序**：
1. 手动为每个现有团队设置 `teams.code`（通过管理界面或 SQL）
2. 运行 `008_item_code_redesign.sql` 添加列和索引
3. 运行迁移程序重写 `main_items.code` 和生成 `sub_items.code`

**迁移程序逻辑**（`cmd/migrate/rewrite_codes.go`）：

```go
// 重写 main_items.code
// 按 team_id 分组，组内按 id 升序，分配 {team.Code}-{seq:05d}
// 包裹在单个事务内

// 生成 sub_items.code
// 按 main_item_id 分组，组内按 id 升序，基于 main_item.code 拼接 -{sub_seq:02d}
// 包裹在单个事务内
```

**回滚脚本**（`cmd/migrate/rollback_008.sql`）：

```sql
-- 回滚 008：还原 main_items.code 为 MI-XXXX 格式，删除新增列
-- 1. 还原 main_items.code（需备份原始值）
-- 2. DROP INDEX idx_sub_items_main_code
-- 3. ALTER TABLE sub_items DROP COLUMN code（SQLite 需重建表）
-- 4. DROP INDEX idx_teams_code
-- 5. ALTER TABLE teams DROP COLUMN code（SQLite 需重建表）
```

> SQLite 不支持 `DROP COLUMN`（SQLite 3.35.0+ 支持），生产环境使用 MySQL 无此限制。

## Error Handling

### 新增错误类型

```go
// pkg/errors/errors.go 新增
var ErrTeamCodeDuplicate = &AppError{Code: "TEAM_CODE_DUPLICATE", Message: "该编码已被使用", Status: 400}
var ErrTeamCodeRequired  = &AppError{Code: "TEAM_CODE_REQUIRED",  Message: "团队编码为必填项", Status: 400}
```

### 错误处理策略

| 场景 | 错误来源 | 处理方式 |
|------|---------|---------|
| Team.Code 格式不合法 | Gin binding `alpha,min=2,max=6` | `ErrValidation`（现有） |
| Team.Code 重复 | DB 唯一约束 → `isDuplicateKeyError()` | `ErrTeamCodeDuplicate`（新增） |
| NextCode() 事务失败 | DB 错误 | 透传，service 层不重试（锁已消除 race） |
| NextSubCode() 事务失败 | DB 错误 | 透传，service 层不重试 |
| SubItem.Code 唯一约束冲突 | DB 唯一约束（锁外异常路径） | `isDuplicateKeyError()` → service 层重试最多 3 次 |

`TeamRepo.Create()` 中检测唯一约束错误并映射到 `ErrTeamCodeDuplicate`，与现有 `isDuplicateKeyError()` 模式一致。

## Testing Strategy

工具链：后端使用 `github.com/stretchr/testify`（assert / require）；前端使用 `vitest` + `@testing-library/react`。

### Backend Unit Tests

| 测试文件 | 工具 | 测试内容 |
|---------|------|---------|
| `repository/gorm/main_item_repo_test.go` | testify | `NextCode()` 返回正确格式；序号递增；team code 快照（team code 变更后旧事项编码不变） |
| `repository/gorm/sub_item_repo_test.go` | testify | `NextSubCode()` 返回正确格式；子序号递增 |
| `service/team_service_test.go` | testify | `CreateTeam()` 传入 Code 字段；Code 重复返回 `ErrTeamCodeDuplicate` |
| `model/main_sub_item_test.go` | testify | SubItem.Code 字段存在；联合唯一索引生效 |

### Backend Integration Tests

运行环境：MySQL（SQLite 不支持行级锁，并发测试结果不可信）。

| 测试场景 | 验证内容 |
|---------|---------|
| 并发创建主事项 | 同一团队下 2 个 goroutine 并发调用 `Create()`，两者生成不同编码且均成功入库 |
| 并发创建子事项 | 同一主事项下 2 个 goroutine 并发调用，两者生成不同编码 |
| 迁移后编码格式 | `SELECT COUNT(*) FROM main_items WHERE code LIKE 'MI-%'` 返回 0 |
| 迁移后子事项编码 | `SELECT COUNT(*) FROM sub_items WHERE code IS NULL OR code = ''` 返回 0 |

### Frontend Unit Tests

| 测试文件 | 工具 | 测试内容 |
|---------|------|---------|
| `TeamManagementPage.test.tsx` | vitest + RTL | Code 输入框渲染；格式校验错误提示（空/长度/非字母）；后端返回 TEAM_CODE_DUPLICATE 时显示"该编码已被使用" |
| `api/teams.test.ts` | vitest | `createTeam()` 请求体含 `code` 字段；响应解析含 `code` |
| `api/subItems.test.ts` | vitest | 列表/详情响应解析含 `code` 字段 |

页面编码展示变更（ItemViewPage、TableViewPage、WeeklyViewPage、MainItemDetailPage、ItemPoolPage、SubItemDetailPage）均为纯展示值变更，不新增组件测试；由 API 层测试覆盖数据契约，人工验收页面渲染。

### Coverage Target

- `NextCode()` / `NextSubCode()` 实现：100%
- Team Code 校验路径（DTO binding + 唯一约束映射）：100%
- 前端 TeamManagementPage Code 校验分支：100%
- 迁移程序：手动演练（非自动化）

## Security Considerations

### Threat Model

- Team.Code 唯一性校验必须在后端执行，前端校验仅为 UX 反馈
- 编码格式（`^[A-Za-z]{2,6}$`）防止注入特殊字符进入编码字段

### Mitigations

- 后端 DTO binding tag `alpha` 拒绝非字母字符
- 数据库层唯一索引作为最终防线
- `isDuplicateKeyError()` 已覆盖 SQLite 和 MySQL 的唯一约束错误格式

## Open Questions

- [x] SQLite FOR UPDATE 支持：SQLite 不支持行级锁，`FOR UPDATE` 在 SQLite 下退化为表锁。开发/测试环境使用 SQLite 时并发测试结果可能与 MySQL 不同。**决策**：并发集成测试在 MySQL 环境下运行；SQLite 仅用于单元测试。
- [x] SubItem.Code 联合唯一索引 GORM tag：使用 `uniqueIndex:"idx_sub_item_main_code,composite:main_item_id"` 语法，需验证 GORM v1.31 支持此写法。备选：在迁移 SQL 中手动创建索引。

## Appendix

### Alternatives Considered

| 方案 | 优点 | 缺点 | 未选原因 |
|------|------|------|---------|
| SELECT MAX + 唯一索引重试 | 实现简单 | 高并发下重试次数不可控 | 已在 proposal 中排除 |
| 数据库序列（AUTO_INCREMENT 辅助表） | 无锁竞争 | 需额外表，跨 DB 兼容性差 | 过度设计，当前并发量低 |
| SELECT FOR UPDATE（选中） | 消除 race condition，锁粒度 per-team | SQLite 退化为表锁 | 接受，测试环境影响可控 |

### References

- `docs/proposals/item-code-redesign/proposal.md` — 业务背景和方案选型
- `docs/features/item-code-redesign/prd/prd-spec.md` — 功能需求
- `backend/internal/repository/gorm/main_item_repo.go:89` — 现有 NextCode() 实现
- `backend/migrations/006_main_items_team_code_index.sql` — 现有 team_id+code 联合唯一索引
