---
created: 2026-04-19
updated: 2026-04-22
author: faner
status: Draft
---

# Proposal: Code Quality — Performance, Duplication & Readability

## Problem

项目由 AI 辅助开发，当前存在三类问题，按严重程度排序：

### 1. 性能问题（新发现）

| 问题 | 位置 | 影响 |
|------|------|------|
| N+1 查询 | `view_service.go:871` `resolveAssigneeNames` | 每个唯一 assignee 发一次 `FindByID`，10 个 assignee = 10 次串行 DB 查询 |
| 全表内存分页 | `view_service.go:573-631` `TableView` | 先把所有 main_items + sub_items 加载到内存，再在 Go 里切片分页 |
| 重复全表扫描 | `view_service.go:634-706` `TableExportCSV` | 与 `TableView` 完全相同的 fetch+filter 逻辑，独立执行 |
| 双构造函数 | `view_service.go:34-50` | `NewViewService` / `NewViewServiceWithUserRepo` 两个构造函数指向同一结构体，混乱 |
| 全局 mutex map 泄漏 | `main_item_service.go:35` `linkageMuMap` | 每个 MainItem ID 创建一个 mutex，永不回收，随数据增长无限增大 |

### 2. 重复代码（原有问题）

- Handler/Service/Repo 三层有重复样板：`mapNotFound` 副本、重复的分页和日期解析逻辑
- `TableView` 和 `TableExportCSV` 共享约 60 行相同的 fetch+filter 代码

### 3. 规范缺失（原有问题）

- CLAUDE.md 只有通用行为准则，没有项目级编码规范
- AI 每次会话无法获得一致的风格指引

**Why now（后果，非时机）**：`linkageMuMap` 代码中无删除路径，进程生命周期内只增不减；GORM 日志在开发环境中已可观测到 `resolveAssigneeNames` 触发的串行查询。若推迟 6 个月：按当前迭代速度 main_items 预计增长至 500+，届时每次 `TableView` 请求将加载 500+ 行到内存再切片，`linkageMuMap` 将持有 500+ 个永不释放的 mutex 条目；无规范约束下 AI 会话已出现过 snake_case JSON tag（model 层历史遗留），同类风格漂移将持续累积。

## Proposed Solution

**Phase 0 先修性能，再清理重复，最后固化规范。**

### Phase 0: 修复性能问题（优先）

| 任务 | 改动 |
|------|------|
| 批量查询 assignee | 在 `UserRepo` 增加 `FindByIDs(ctx, ids []uint)` 方法；`resolveAssigneeNames` 改为单次 `WHERE id IN (...)` |
| DB 层分页 | `TableView` 改为将 filter/pagination 下推到 repo 层，避免全表加载 |
| 提取 `fetchTableRows` | 将 `TableView` 和 `TableExportCSV` 共同的 fetch+filter 逻辑提取为私有方法 |
| 合并构造函数 | 删除 `NewViewService`，只保留 `NewViewServiceWithUserRepo`，重命名为 `NewViewService` |
| 限制 linkageMuMap | 改用 `sync.Map` 或在 `EvaluateLinkage` 完成后清理 mutex，防止无限增长 |

**完成后的可观测状态**：GORM 查询日志中 `resolveAssigneeNames` 从 N 条串行查询变为 1 条 `WHERE id IN (...)` 查询；`TableView` 不再全表加载；所有现有测试通过。

### Phase 1: 建立编码规范

| 文档 | 内容 |
|------|------|
| `docs/ARCHITECTURE.md` | 分层架构说明（Model → DTO/VO → Service → Handler → Router），每层职责和可复制的 CRUD 模板代码 |
| `docs/DECISIONS.md` | 技术决策记录：JSON tag 统一 camelCase、错误处理模式、分页模式等 |
| `.claude/rules/naming.md` | 命名规范（JSON tag、变量、函数）|
| `.claude/rules/patterns.md` | 分层模式（CRUD 模板、helper 提取规则）|
| `.claude/rules/frontend.md` | 前端组件化规则（API 层、UI 组件、状态管理）|

**完成后的可观测状态**：开发者打开 `docs/ARCHITECTURE.md` 可直接复制 CRUD 模板代码，无需查阅现有代码推断惯例；在新 Claude Code 会话中执行 `@rules/naming.md`，规则内容出现在上下文中（行为验证，而非仅检查文件存在）；用一段含 snake_case JSON tag 的代码片段提问，Claude Code 应指出违规并引用 naming.md 中的规则。

### Phase 2: 配置自动化 Lint 规则

- **Go 后端**：golangci-lint 启用 `tagliatelle` 规则，检测 snake_case JSON tag；启用 `dupl` 检测重复代码块
- **TypeScript 前端**：ESLint 新增规则，检查 API 层文件命名（`camelCase.ts`）和组件导出命名（PascalCase）

**完成后的可观测状态**：`golangci-lint run ./...` 在检测到 snake_case JSON tag 时以非零退出码退出，阻断 CI；ESLint 对 API 层命名违规同样返回非零退出码。开发者在本地提交前运行 lint 即可发现违规，无需等待 code review。

### Phase 3: 清理现有重复代码

按后端→前端顺序：
1. **抽取公共 helper**：`mapNotFound` → 通用 `MapNotFound(err, targetErr)`；分页 → `ApplyPaginationDefaults`；日期解析 → `ParseDate`
2. **统一 CRUD 模式**：repo 层减少重复
3. **前端组件化**：识别重复 UI 模式，抽取可复用组件（具体组件名在 Phase 3 开始前确定）

**完成后的可观测状态**：`grep -r "ErrRecordNotFound" backend/internal` 返回 ≤10 处（当前 21 处）；lint 在清理后全量通过，确认无回归。

## Alternatives Considered

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| **A. 四阶段捆绑（推荐）** | Phase 0 修复的是实现 bug，Phases 1–3 修复的是产生 bug 的过程缺陷；若只做 Phase 0，下一个 AI 会话在没有规范和 lint 约束的情况下仍会引入新的 N+1 查询或 snake_case tag，Phase 0 的修复会被逐步侵蚀；捆绑的核心理由是"Phase 0 修复 + Phases 1–3 防止复发"，而非仅仅"同根同源"；Phase 3 清理在 lint 就位后执行，避免二次触碰同一文件 | 总工作量比单做 Phase 0 大；Phase 3 工期较长 | ✅ Recommended |
| **A'. 仅做 Phase 0** | 工作量最小，立即消除性能风险 | 不解决复发风险：历史上已有 AI 会话引入 snake_case tag（model 层遗留），无规范和 lint 约束下同类问题会在 Phase 0 修复后的新会话中重新出现；3 个月后需要重新做 Phase 1–3，届时要二次触碰已改过的文件 | ❌ 修复会被侵蚀 |
| **B. 重复优先** | 先消除 21 个重复副本，后续改动减少重复文件触碰 | N+1 查询和全表内存分页持续存在，数据量增长后响应时间线性恶化；性能问题比重复代码更紧迫 | ❌ 风险持续 |
| **C. 全部一次（无分阶段）** | 一次性解决 | 单 PR 改动范围横跨性能、文档、lint、清理，难以 review，回归风险高 | ❌ 难以 review |
| **D. 不做（Do Nothing）** | 零成本 | 6 个月内：`linkageMuMap` 随 main_items 增长持续泄漏（500+ mutex 条目）；`TableView` 每次请求内存加载量线性增长；AI 会话无规范约束，snake_case/camelCase 混用将继续扩散，修复成本随时间递增 | ❌ 技术债加速累积 |

**四阶段捆绑的依赖关系**：Phase 0 独立可先行；Phase 1（规范文档）是 Phase 2（lint 配置）的前置条件——lint 规则必须与文档中的决策对齐；Phase 2 是 Phase 3（清理）的前置条件——清理后需要 lint 全量验证无回归。Phase 1 可与 Phase 0 并行。

## Scope

### In Scope

| Phase | 内容 | 规模 | 是否可并行 |
|-------|------|------|-----------|
| Phase 0 | `UserRepo.FindByIDs`、`TableView` DB 分页、`fetchTableRows` 提取、构造函数合并、`linkageMuMap` 限制 | M（约 3–5 天） | 可与 Phase 1 并行启动 |
| Phase 1 | 编写 `docs/ARCHITECTURE.md`、`docs/DECISIONS.md`、`.claude/rules/*.md` | S（约 1–2 天） | 可与 Phase 0 并行 |
| Phase 2 | 配置 golangci-lint（`tagliatelle`、`dupl`）和 ESLint 命名规则 | S（约 1 天） | 须在 Phase 1 完成后启动 |
| Phase 3 | 后端 helper 提取；前端可复用组件抽取（≥3 个，名称在 Phase 3 开始前确定） | L（约 5–7 天） | 须在 Phase 2 完成后启动 |

**总工期目标**：Phase 0 + Phase 1 并行，随后 Phase 2 → Phase 3 串行，目标在 **3 个 sprint（约 3 周）** 内完成全部四个阶段。

**顺序约束**：Phase 0 和 Phase 1 可并行；Phase 2 必须在 Phase 1 之后（lint 规则须与规范文档对齐）；Phase 3 必须在 Phase 2 之后（清理后需 lint 全量验证）。

### Out of Scope
- 数据库列命名（保持 snake_case，由 GORM 自动映射）
- 业务逻辑变更
- 新功能开发
- 修改 zcode 插件本身
- CI/CD 流水线集成 lint（lint 本地可用即满足本提案目标，CI 集成作为后续独立任务）
- 现有违规代码的批量迁移（Phase 2 lint 规则以 warn 模式启用，不强制修复存量违规）

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| `TableView` 改 DB 分页后行为差异 | Medium | High | 改动前后对比测试；现有 repo 测试覆盖 |
| `linkageMuMap` 改动引入并发 bug | Low | High | 保守方案：改用 `sync.Map`，不改锁语义 |
| 规范文档过于理想化 | Medium | Medium | 规范中提供具体代码示例和反例 |
| 清理范围过大导致回归 | Medium | High | 每批清理后运行完整测试套件；分批提交 |
| Phase 2 lint 规则因存量违规过多而长期停留在 warn 模式，无法切换为 error 强制执行 | Medium | High | Phase 2 启动前先统计存量违规数量（`golangci-lint run --enable tagliatelle ./... 2>&1 \| grep -c "json-camel"`）；若违规数 >20，在 Phase 2 内增加一个 cleanup 子任务，将存量违规降至 ≤5 后再切换为 error 模式 |

## Success Criteria

### Phase 0
- [ ] `UserRepo` 新增 `FindByIDs` 方法，`resolveAssigneeNames` 改为单次批量查询
- [ ] `TableView` 分页下推到 repo 层，不再全表加载
- [ ] `TableView` 和 `TableExportCSV` 共享 `fetchTableRows` 私有方法
- [ ] `NewViewService` 只有一个构造函数
- [ ] `linkageMuMap` 不再无限增长
- [ ] 全部现有测试通过

### Phase 1–3
- [ ] `docs/ARCHITECTURE.md` 已创建，包含以下全部章节：分层架构图、各层职责说明、CRUD 模板代码、错误处理模式
- [ ] `docs/DECISIONS.md` 已创建，包含以下全部章节：JSON tag 规则、分页模式、日期解析模式、错误映射模式
- [ ] `.claude/rules/naming.md`、`patterns.md`、`frontend.md` 已创建，可在 AI 会话中通过 `@rules` 引用加载
- [ ] golangci-lint 新增 `tagliatelle` 规则，`go lint ./...` 能检测出 snake_case JSON tag 违规
- [ ] 后端重复副本从当前 21 个（12 个 repo 层内联 `ErrRecordNotFound` 检查 + 1 个 `parsePagination` 重复函数 + 8 个 service 层日期解析样板）减少到 ≤10 个（减少 ≥50%）
- [ ] 至少抽取 3 个可复用前端 UI 组件（需在 Phase 3 开始前列出具体组件名）
- [ ] 全部现有测试通过

## Next Steps

- Proceed to `/write-prd` to formalize requirements
