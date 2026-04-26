---
created: 2026-04-26
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: 数据库方言兼容性

## Overview

引入 `Dialect` 接口抽象 SQLite/MySQL 差异，修复 4 个已知不兼容点，新增 lint 检查防复发。改动集中在 repository 层（2 个 repo）和 migration 层（已有 if-else 分支），不影响 service/handler 层。

## Architecture

### Layer Placement

```
┌─────────────────────────────────────────────┐
│              pkg/dbutil (新增)                │
│  dialect.go      — Dialect 接口 + 工厂函数    │
│  dialect_test.go — 单元测试                   │
├─────────────────────────────────────────────┤
│           repository/gorm (修改)              │
│  main_item_repo.go — 注入 Dialect, NextCode   │
│  sub_item_repo.go  — 注入 Dialect, NextSubCode│
├─────────────────────────────────────────────┤
│           migration (修改)                    │
│  rbac.go — 导出 HasColumn, 用 isMySQL 分支    │
│           rebuildTeamMembersTable 增加分支     │
├─────────────────────────────────────────────┤
│           scripts (修改)                      │
│  lint-staged.sh — 新增 SQLite 关键字检测       │
└─────────────────────────────────────────────┘
```

### Dependencies

外部依赖：无新增。使用已有的 `gorm.io/gorm`、`github.com/glebarez/sqlite`、`gorm.io/driver/mysql`。

**内部导入关系**（新增的 import 标记为 ★）：

```
pkg/dbutil                    → gorm.io/gorm
repository/gorm/main_item_repo.go  → pkg/dbutil ★ (新增)
repository/gorm/sub_item_repo.go   → pkg/dbutil ★ (新增)
cmd/server/main.go                 → pkg/dbutil ★ (新增)
                                   → repository/gorm (已有)
migration/rbac.go                  → (无新增导入，复用已有 isMySQL 函数)
```

## Interfaces

### Interface 1: `Dialect`

屏蔽不同数据库的 SQL 语法差异。调用方不感知底层是 SQLite 还是 MySQL。

```go
// pkg/dbutil/dialect.go

// ColumnExpr is a string that represents a SQL column name or expression.
// Production callers must use pre-defined constants (e.g., ColCode).
// NewColumnExpr is exported for test use only — non-test callers should
// never need it.
type ColumnExpr string

// Pre-defined column expressions — the only valid inputs to Dialect methods.
const (
    ColCode ColumnExpr = "code"
)

// NewColumnExpr creates a ColumnExpr. Exported for use in tests only.
// Production code must use the constants above.
func NewColumnExpr(s string) ColumnExpr { return ColumnExpr(s) }

// Dialect abstracts SQL syntax differences between database engines.
type Dialect interface {
    // CastInt returns a CAST expression that produces an integer result.
    // SQLite: CAST(expr AS INTEGER)
    // MySQL:  CAST(expr AS SIGNED)
    CastInt(expr ColumnExpr) string

    // Substr returns a substring extraction expression.
    // SQLite: SUBSTR(str, start)
    // MySQL:  SUBSTRING(str, start)
    // start is 1-indexed (consistent across both databases).
    Substr(str ColumnExpr, start int) string

    // Now returns a datetime expression for the current timestamp.
    // SQLite: datetime('now')
    // MySQL:  CURRENT_TIMESTAMP
    Now() string
}

// NewDialect creates a Dialect based on the GORM Dialector name.
// Panics if db is nil (consistent with project panic-on-nil pattern).
func NewDialect(db *gorm.DB) Dialect
```

**实现**：两个未导出的结构体 `sqliteDialect{}` 和 `mysqlDialect{}`，`NewDialect` 根据 `db.Dialector.Name()` 返回对应实例。仅支持 `"sqlite"` 和 `"mysql"` 两个名称；其他值一律 panic（见 Error Handling 章节）。

**安全约束**：`CastInt` 和 `Substr` 的字符串参数通过 `ColumnExpr` 类型约束 — 见 Security Considerations 章节。

### Interface 2: `HasColumn` 导出

当前 `HasColumn` 是 `migration` 包的导出函数，内部硬编码了 `pragma_table_info`。改为根据 dialect 分支。

```go
// migration/rbac.go — HasColumn 已经导出，修改内部实现
func HasColumn(db *gorm.DB, table, column string) bool
```

行为变更：MySQL 走 `information_schema.columns`，SQLite 走 `pragma_table_info`（与已有的 `columnExists` 内部函数逻辑一致）。

## Data Models

无需新增数据模型。`Dialect` 是无状态接口，两个实现均为零大小结构体。

## Error Handling

### Panic-on-nil

`NewDialect(db)` 在 `db` 为 nil 时 panic，与项目 handler/service 构造器 panic-on-nil 模式一致。

### Unrecognized Dialect Fail-Fast

`NewDialect` 在 `db.Dialector.Name()` 既不是 `"sqlite"` 也不是 `"mysql"` 时 panic，附带类型化错误值。定义专用错误类型供测试断言：

```go
type UnsupportedDialectError struct {
    Name string
}

func (e UnsupportedDialectError) Error() string {
    return fmt.Sprintf("unsupported dialect: %s, only 'sqlite' and 'mysql' are supported", e.Name)
}
```

调用方式：`panic(UnsupportedDialectError{Name: name})`。测试通过 `recover()` 断言类型而非子串匹配：

```go
defer func() {
    r := recover()
    require.NotNil(t, r)
    _, ok := r.(dbutil.UnsupportedDialectError)
    require.True(t, ok, "expected UnsupportedDialectError, got %T: %v", r, r)
}()
```

这确保启动阶段立即暴露配置错误，而非在运行时产生令人困惑的 SQL 语法错误。

### Migration 层

`rebuildTeamMembersTable` 中 DDL 执行失败时返回 error 并回滚事务（现有行为不变）。

## Cross-Layer Data Map

Single-layer feature (backend only) — not applicable.

## Testing Strategy

### Per-Layer Test Plan

| Layer | Test Type | Tool | What to Test | Coverage Target |
|-------|-----------|------|--------------|-----------------|
| pkg/dbutil | Unit | go test + testify | NewDialect + 所有方法在两种方言下的输出 | 100% |
| repository/gorm | Unit (现有) | go test + testify | NextCode/NextSubCode 在 SQLite 下编号递增 | 现有测试通过 |
| repository/gorm | Unit (新增) | go test + testify | NextCode/NextSubCode 的 Select 子句在 mysqlDialect 下生成正确 SQL 字符串 | 100% of dialect paths |
| migration | Unit (现有) | go test + testify | HasColumn/rebuildTable 在 SQLite 下正确 | 现有测试通过 |

### Key Test Scenarios

1. **`NewDialect` 工厂**：
   - SQLite Dialector → 返回 `sqliteDialect`
   - MySQL Dialector → 返回 `mysqlDialect`
   - 未知 Dialector → panic（`"unsupported dialect: ..."`）
2. **`CastInt`**：
   - sqliteDialect: `CastInt(ColCode)` → `"CAST(code AS INTEGER)"`
   - mysqlDialect: `CastInt(ColCode)` → `"CAST(code AS SIGNED)"`
3. **`Substr`**：
   - sqliteDialect: `Substr(ColCode, 7)` → `"SUBSTR(code, 7)"`
   - mysqlDialect: `Substr(ColCode, 7)` → `"SUBSTRING(code, 7)"`
4. **`Now`**：
   - sqliteDialect: `Now()` → `"datetime('now')"`
   - mysqlDialect: `Now()` → `"CURRENT_TIMESTAMP"`
5. **NextCode Select 子句验证（mysqlDialect）**：
   - 构造 `mysqlDialect`，调用 `dialect.CastInt(dialect.Substr(dbutil.ColCode, 4))` → 验证结果字符串为 `"CAST(SUBSTRING(code, 4) AS SIGNED)"`
   - 同理验证 `sqliteDialect` 路径 → `"CAST(SUBSTR(code, 4) AS INTEGER)"`
   - 将该字符串拼入 `MAX(...)` 后验证完整表达式格式正确
6. **NextCode 端到端**（手动 MySQL 集成测试）：连续调用生成严格递增编号

### MySQL Repo Path Automated Test Plan

由于项目 CI 不包含 MySQL 实例，无法对 repo 层做 MySQL 端到端测试。采用以下分层策略确保 MySQL 路径正确性：

**层级 1 — Dialect 单元测试**（scenario 2-4）：证明 `mysqlDialect` 各方法输出符合 MySQL 语法。

**层级 2 — Select 子句字符串断言**（scenario 5）：在 repo 测试中新增用例，构造 `mysqlDialect` 实例，直接断言 `dialect.CastInt(dialect.Substr(dbutil.ColCode, N))` 拼接后的完整 `MAX(...)` 表达式字符串。具体实现：

```go
func TestNextCode_SelectClause_MySQL(t *testing.T) {
    d := dbutil.NewDialect(mysqlDB) // 或直接构造 mysqlDialect
    expr := "MAX(" + d.CastInt(d.Substr(dbutil.ColCode, 4)) + ")"
    assert.Equal(t, "MAX(CAST(SUBSTRING(code, 4) AS SIGNED))", expr)
}

func TestNextCode_SelectClause_SQLite(t *testing.T) {
    d := dbutil.NewDialect(sqliteDB) // 使用现有 SQLite 实例
    expr := "MAX(" + d.CastInt(d.Substr(dbutil.ColCode, 4)) + ")"
    assert.Equal(t, "MAX(CAST(SUBSTR(code, 4) AS INTEGER))", expr)
}
```

这两个测试保证 dialect 输出经 repo 层拼接后不会出现括号不匹配、函数名错误等问题。

**层级 3 — 手动集成验证**（scenario 6）：在本地 MySQL 实例运行 `NextCode` 连续调用，确认编号严格递增。记录验证步骤：`docker run mysql:8.0` → 配置 DSN → 运行 repo 测试 → 检查输出。

### Overall Coverage Target

`dialect.go` 100%；repo 层 dialect SQL 路径 100%（通过字符串断言覆盖 SQLite 和 MySQL 两条路径）；migration 层现有测试全部通过。

### 成功标准

前提条件：`config.yaml` 配置的是 MySQL。

- 上述单元测试全部通过
- `frontend/__tests__/e2e/` 内测试通过率 ≥ 95%

## Security Considerations

### SQL 注入风险与缓解

`CastInt` 和 `Substr` 将字符串参数直接嵌入 SQL 表达式（非参数化），理论上存在注入路径。

**威胁模型**：若 `CastInt` 或 `Substr` 的 `expr`/`str` 参数接受用户控制数据，攻击者可注入任意 SQL 片段。`start` 为 `int` 类型，无法承载注入载荷。`Now()` 无参数。

**缓解 — 类型级强制**：使用 `ColumnExpr` 类型（而非 `string`）作为 `CastInt`/`Substr` 的参数类型。`ColumnExpr` 的构造函数 `NewColumnExpr` 虽然 exported（供测试使用），但包外调用方需显式构造，形成代码审读时的视觉信号。当前所有调用点仅使用预定义常量 `ColCode`（值为 `"code"`），不接受用户输入。

**威胁 → 缓解对应**：

| Threat | Mitigation | Enforcement |
|--------|-----------|-------------|
| 调用方传入用户控制字符串 | `ColumnExpr` 类型约束 + 预定义常量 | 编译期类型检查：`string` 不能隐式传给 `ColumnExpr` 参数 |
| 新增调用点引入非列名表达式 | 代码审查 + `NewColumnExpr` 构造函数的视觉信号 | 非 `ColCode` 的调用需显式 `NewColumnExpr(...)` 并通过 CR 人工审核 |

**残留风险**：`NewColumnExpr` exported 供测试使用，恶意调用方理论上可构造任意 `ColumnExpr`。这是有意的 trade-off — 完全封闭类型（unexported 构造函数 + `interface{}` wrapper）会增加 ~20 行 boilerplate 并使测试变复杂，而实际攻击面为零（HTTP handler → service → repo 的调用链中不存在将用户输入传递到 `CastInt`/`Substr` 的路径）。

### HasColumn 注入面分析

`HasColumn(db *gorm.DB, table, column string)` 接受 `string` 参数（非 `ColumnExpr`），需要单独分析。

**参数化查询确认**：`HasColumn` 内部（及设计目标委托的 `columnExists`）使用 GORM `Raw(sql, args...)` 的参数化绑定：
- SQLite：`SELECT count(*) FROM pragma_table_info(?) WHERE name = ?` — `table` 和 `column` 作为绑定参数传入
- MySQL：`SELECT count(*) FROM information_schema.columns WHERE ... AND table_name = ? AND column_name = ?` — 同理

两个参数均通过 `?` 占位符传入，不走字符串拼接，因此不存在 SQL 注入风险。

**调用点枚举**：
- `migration/rbac.go` — `rebuildTeamMembersTable` 内部调用（固定表名 `"pmw_team_members"` + 固定列名字符串）
- `migration/rbac_test.go` — 测试代码，硬编码表名/列名

**用户输入可达性**：所有调用点的 `table` 和 `column` 参数均为代码中的字符串常量，不接受 HTTP 请求参数或任何外部输入。`HasColumn` 不暴露给 handler/service 层。

**为何使用 `string` 而非 `ColumnExpr`**：`HasColumn` 属于 migration 层（非 repo 层），其调用者仅传入固定常量。该函数签名已在代码中导出且被测试引用，引入 `ColumnExpr` 类型约束需要改动 migration 包的导入链，收益为零（无用户输入路径）。

## PRD Coverage Map

| PRD Requirement / AC | Design Component | Interface / Model |
|----------------------|------------------|-------------------|
| Story 1: 需求池转主事项 MySQL 下 200 | mainItemRepo + subItemRepo 注入 Dialect | `Dialect.CastInt` + `Dialect.Substr` |
| Story 2: RBAC 迁移 DDL 兼容 | rebuildTeamMembersTable 增加 isMySQL 分支 | 已有 `isMySQL()` + MySQL DDL |
| Story 3: lint 拦截硬编码 SQLite 关键字 | lint-staged.sh 新增 grep 检查（见下方 Lint Check Specification） | N/A (shell script) |
| Story 4: HasColumn MySQL 兼容 | HasColumn 内部增加 isMySQL 分支 | 已有 `isMySQL()` + `information_schema` |
| Story 5: SQLite 测试无回归 | 现有测试不变 | N/A |
| 方言辅助模块 | `pkg/dbutil/dialect.go` | `Dialect` interface + `NewDialect` |
| 代码规范 | `.claude/rules/` + `docs/lessons/` | N/A |

### Lint Check Specification (Story 3)

**目的**：阻止在 `repository/gorm` 层引入新的硬编码 SQLite 方言语法，强制使用 `pkg/dbutil.Dialect`。

**检测的关键字及模式**：

| Pattern | 匹配的 SQLite 方言 | 合法替代 |
|---------|-------------------|---------|
| `SUBSTR(` | `SUBSTR` 函数 | `dialect.Substr()` |
| `CAST(.+AS INTEGER)` | `CAST(... AS INTEGER)` 类型转换 | `dialect.CastInt()` |
| `datetime('now')` | SQLite 时间函数 | `dialect.Now()` |
| `pragma_table_info` | SQLite PRAGMA 系统表查询 | `migration.HasColumn()` |

**扫描范围**：仅 `backend/internal/repository/gorm/*.go`（repo 实现层）。`pkg/dbutil/`、`migration/`、`handler/`、`service/` 不在扫描范围内。

**排除逻辑**：
- 排除测试文件（`*_test.go`）— 测试中可能需要验证 SQLite 行为
- 排除 `pkg/dbutil/dialect.go` — 这是方言实现本身，必须包含 SQLite 关键字
- 不排除 `migration/` — 迁移层使用 `isMySQL()` 分支模式，不通过 `Dialect` 接口

**lint-staged.sh 实现**：

```bash
# 新增函数: check_sqlite_keywords
check_sqlite_keywords() {
    local pattern='SUBSTR\(|CAST.*AS INTEGER|datetime(.now.)|pragma_table_info'
    local files=$(git diff --cached --name-only -- 'backend/internal/repository/gorm/*.go' | grep -v '_test.go')
    if [ -n "$files" ]; then
        if grep -En "$pattern" $files; then
            echo "ERROR: Hardcoded SQLite syntax found in repo layer."
            echo "Use pkg/dbutil.Dialect instead. See docs/features/db-dialect-compat/."
            return 1
        fi
    fi
    return 0
}
```

**通过/失败示例**：

| File Content | Result | Reason |
|-------------|--------|--------|
| `Select("MAX(CAST(SUBSTR(code, ?) AS INTEGER))")` | FAIL | repo 层硬编码 SQLite CAST+SUBSTR |
| `Select("MAX(" + r.dialect.CastInt(r.dialect.Substr(dbutil.ColCode, 4)) + ")")` | PASS | 通过 dialect 接口调用 |
| `// SQLite uses SUBSTR for testing` | FAIL | 注释中包含关键字（grep 无法区分注释）— 此为已知限制，若误报则在行尾添加 `# noqa` |
| `SELECT pragma_table_info(...)` in `migration/rbac.go` | N/A (不扫描) | migration 不在扫描范围 |
| `return "SUBSTR(" + string(str) + ...)` in `dialect.go` | N/A (不扫描) | pkg/dbutil 不在扫描范围 |

**已知限制**：grep 无法区分 Go 字符串字面量和注释/代码，可能产生少量误报。对于合法包含关键字的行，使用行尾 `# nosqlite` 标记跳过。

## File Change Summary

| File | Change | Scope |
|------|--------|-------|
| `backend/internal/pkg/dbutil/dialect.go` | **新增** — ColumnExpr 类型 + 常量、Dialect 接口、两个实现、NewDialect 工厂 | ~50 行 |
| `backend/internal/pkg/dbutil/dialect_test.go` | **新增** — 单元测试 | ~60 行 |
| `backend/internal/repository/gorm/main_item_repo.go` | **修改** — 构造函数接收 Dialect，NextCode 使用 dialect | ~5 行改动 |
| `backend/internal/repository/gorm/sub_item_repo.go` | **修改** — 构造函数接收 Dialect，NextSubCode 使用 dialect | ~5 行改动 |
| `backend/internal/repository/main_item_repo.go` | **修改** — 接口不变（Dialect 不在接口中） | 无改动 |
| `backend/cmd/server/main.go` | **修改** — 创建 Dialect 并注入两个 repo | ~3 行改动 |
| `backend/internal/migration/rbac.go` | **修改** — rebuildTeamMembersTable 增加 MySQL DDL 分支；HasColumn 增加 MySQL 分支 | ~30 行改动 |
| `scripts/lint-staged.sh` | **修改** — 新增 SQLite 关键字 grep 检查 | ~10 行改动 |

### NextCode / NextSubCode SQL 重构详情

这是本 feature 的核心改动。两个 repo 的 `NextCode`/`NextSubCode` 方法需要将硬编码的 SQLite 语法替换为 dialect 调用。

**Before（`main_item_repo.go` — NextCode）：**

```go
// 硬编码 SQLite: CAST(...AS INTEGER), SUBSTR(...)
var maxCode int
err := r.db.Model(&model.MainItem{}).
    Where("team_id = ? AND code LIKE ?", teamID, prefix+"%").
    Select("MAX(CAST(SUBSTR(code, ?) AS INTEGER))", len(prefix)+1).
    Scan(&maxCode).Error
```

**After：**

```go
// dialect 替换: CastInt(Substr(...))
var maxCode int
err := r.db.Model(&model.MainItem{}).
    Where("team_id = ? AND code LIKE ?", teamID, prefix+"%").
    Select("MAX(" + r.dialect.CastInt(r.dialect.Substr(dbutil.ColCode, len(prefix)+1)) + ")").
    Scan(&maxCode).Error
```

**Before（`sub_item_repo.go` — NextSubCode）：**

```go
var maxNum int
err := r.db.Model(&model.SubItem{}).
    Where("main_item_id = ? AND code LIKE ?", mainItemID, prefix+"%").
    Select("MAX(CAST(SUBSTR(code, ?) AS INTEGER))", len(prefix)+1).
    Scan(&maxNum).Error
```

**After：**

```go
var maxNum int
err := r.db.Model(&model.SubItem{}).
    Where("main_item_id = ? AND code LIKE ?", mainItemID, prefix+"%").
    Select("MAX(" + r.dialect.CastInt(r.dialect.Substr(dbutil.ColCode, len(prefix)+1)) + ")").
    Scan(&maxNum).Error
```

关键变化：`SUBSTR(code, ?)` + GORM 占位符变为 `r.dialect.Substr(dbutil.ColCode, len(prefix)+1)` 直接生成完整表达式；`CAST(... AS INTEGER)` 变为 `r.dialect.CastInt(...)`。`len(prefix)+1` 从 GORM 参数变为直接传入 `Substr`，因此 `Select` 不再需要额外参数。

### DI 变更详情

`main.go` 中新增一行，两个 repo 构造函数各增加一个参数：

```go
// 新增
dialect := dbutil.NewDialect(db)

// 修改（加第二个参数）
mainItemRepo := gormrepo.NewGormMainItemRepo(db, dialect)
subItemRepo := gormrepo.NewGormSubItemRepo(db, dialect)
```

其余 repo 不受影响（它们不使用原始 SQL）。

### Migration 层变更详情

`rebuildTeamMembersTable` 中两处 `INTEGER PRIMARY KEY AUTOINCREMENT` DDL 增加 `isMySQL` 分支，使用 `BIGINT UNSIGNED NOT NULL AUTO_INCREMENT` + `PRIMARY KEY(id)` 格式（与 `rbacTableDDL` 中的模式一致）。

`HasColumn` 改为委托给已有的 `columnExists` 内部函数（已有 isMySQL 分支），消除重复逻辑。

## Open Questions

无。

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| 独立函数 (dbutil.CastInt(expr, db)) | 无需改构造函数/DI/测试 | 无法屏蔽方言差异，调用方仍需自己处理 | 用户选择了接口抽象方案 |
| GORM Callback 自动改写 SQL | 对业务代码透明 | ~150 行 boilerplate，正则脆弱 | 投入产出比不合理 |
