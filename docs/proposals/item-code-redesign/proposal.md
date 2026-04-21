---
title: 事项编码体系重新设计
slug: item-code-redesign
status: draft
created: 2026-04-22
---

## Problem

当前事项编码存在以下问题：

1. **主事项编码无法区分团队**：`NextCode()` 硬编码返回 `MI-0001`（见 `main_item_repo.go:102`），所有团队共享同一前缀。系统已有 195 次提交、6 个 feature 开发周期，即将从单团队内测过渡到多团队使用。当团队数 > 1 时，`MI-0042` 无法告知用户该事项属于哪个团队——每次口头沟通都需额外确认团队归属，事项越多辨识成本越高。该问题不是假设：`NextCode()` 被 2 个 service 调用点（`main_item_service.go:76`、`item_pool_service.go:141`）共享，说明编码生成路径已被多处依赖，改得越晚影响面越大。

2. **子事项无持久编码**：SubItem 模型没有 `Code` 字段（见 `sub_item.go`）。前端已有 **2 处**运行时拼接临时标识符（`MainItemDetailPage.tsx:407`、`SubItemDetailPage.tsx:153`），格式为 `SI-${itemId}-${subId}`——这证明子事项有可读标识的需求，但当前方案依赖数据库 ID 且不稳定（删除后 ID 不可复用，无法在周报或进度追踪中稳定引用）。临时拼接还意味着同一子事项在不同页面可能呈现不同格式。

3. **Team 模型缺少缩写字段**：`Team` 结构体（见 `team.go`）仅有 `Name` 和 `Description`，没有可作为编码前缀的短代码字段。

## Proposal

重新设计事项编码体系，引入团队缩写作为前缀，并为子事项增加编码。

### 编码格式

| 层级 | 格式 | 示例 | 组成 |
|------|------|------|------|
| Team Code | `{code}` | `FEAT` | 2~6位字母，全局唯一 |
| 主事项 | `{team_code}-{seq}` | `FEAT-00001` | 团队缩写 + 5位自增序号 |
| 子事项 | `{team_code}-{seq}-{sub_seq}` | `FEAT-00001-01` | 主事项编码 + 2位自增序号 |

### 核心规则

- **不可变**：编码生成后不再变更，即使团队 code 被修改
- **Team Code 快照**：创建事项时，将当时的 team code 写入事项编码，后续 team code 变更不影响已有编码
- **序号范围**：主事项序号 per-team 自增；子事项序号 per-main-item 从 01 开始
- **Team Code 格式**：2~6位大小写字母组合，全局唯一

### 变更范围

**Team 模型**：新增 `Code` 字段（varchar(6)，全局唯一索引）

**MainItem 模型**：
- `Code` 列从 `varchar(10)` 扩展到 `varchar(12)`
- 编码格式从 `MI-NNNN` 改为 `{team_code}-NNNNN`
- `NextCode()` 逻辑改用 team code 作为前缀

**SubItem 模型**：
- 新增 `Code` 字段（varchar(15)），per-main-item 唯一
- 新增 `NextSubCode()` 逻辑，基于主事项编码生成子事项编码

### 用户可见行为

**Team Code 管理**：在现有 `TeamManagementPage` 的创建团队对话框中增加一个 "团队编码" 输入框（与 name、description 并列）。输入规则：2~6 位英文字母，必填。校验反馈——输入不符合格式时，输入框下方显示 "编码须为 2~6 位英文字母"；与已有团队编码重复时，显示 "该编码已被使用"。创建成功后，团队列表页每行增加一列显示编码。

**各页面编码展示变化**：

| 页面 | 当前行为 | 变更后行为 |
|------|----------|-----------|
| ItemViewPage | 显示 `MI-0001`（Badge 组件，`<Badge className="font-mono">`）；搜索按 `code.toLowerCase().includes(q)` 过滤 | 显示 `FEAT-00001`（同一 Badge 组件）；搜索逻辑不变，输入 `FEAT-` 可按团队过滤，输入完整编码精确匹配到单条 |
| TableViewPage | 行内显示 `MI-0001`（`<span className="font-mono text-xs">`）；搜索同上 | 显示 `FEAT-00001`；同一组件，仅数据源变更 |
| WeeklyViewPage | 行首显示 `{mainItem.code} {mainItem.title}`（`<span>` 内联） | 显示 `FEAT-00001 标题`，格式不变，仅编码值变更 |
| MainItemDetailPage | 标题旁 Badge 显示 `MI-0001`；子事项表格中运行时生成 `SI-001-01` 临时编码 | 标题旁 Badge 显示 `FEAT-00001`；子事项行显示 `FEAT-00001-01`（来自 SubItem.Code 字段，非拼接） |
| ItemPoolPage | 关联主事项显示 `{mi.code} {mi.title}` | 显示 `FEAT-00001 标题`，格式不变 |

**搜索行为**：所有页面的搜索框行为不变（`code.toLowerCase().includes(q)`）。用户可输入 `FEAT-` 过滤某团队事项，或输入完整编码 `FEAT-00001` 精确匹配。无模糊搜索或前缀自动补全。

### 数据迁移

- 直接切换，不需要兼容旧 `MI-XXXX` 格式
- 为已有 MainItem 的 code 列重写为 `{TEAM_CODE}-NNNNN`（按 team_id 分组，每组内按 id 排序分配递增序号）
- 为已有 SubItem 生成 code（按 main_item_id 分组，组内按 id 排序，基于所属主事项编码拼接 `-NN` 后缀）

## Alternatives Analysis

### A. Do Nothing（保持 MI-NNNN 格式）

保持当前 `MI-NNNN` 格式不变，子事项继续使用数据库 ID。

- **优点**：零迁移风险，零开发成本，现有索引和搜索逻辑不变
- **缺点**：子事项始终无人类可读标识符；主事项编码无法区分团队；随着多团队使用，所有团队共享 `MI-` 前缀会导致沟通混淆（"你看一下 MI-0042" 需要额外说明是哪个团队）
- **结论**：不可接受。子事项编码是后续功能（周报关联、进度追踪）的前置依赖，且当前编码在多团队场景下已造成辨识困难

### B. UUID 作为事项标识

使用 UUID（如 `550e8400-e29b-41d4-a716-446655440000`）替代自增编码。

- **优点**：全局唯一无需协调，天然支持分布式生成，无 race condition
- **缺点**：不可读、不可口述、搜索困难；用户在会议中无法快速引用事项；需要全量替换所有前端展示逻辑；索引性能差（36 字符随机字符串 B-tree 效率低）
- **结论**：排除。事项编码的核心使用场景是人与人之间口头/文字快速引用，UUID 完全违背这一目标

### C. 全局序号（不按团队分区）

使用全局自增序号，前缀仍为团队缩写：`FEAT-00001`、`CORE-00002`、`FEAT-00003`（序号跨团队连续）。

- **优点**：实现简单，无需 per-team 的 `MAX(code)` 查询，不存在 per-team 的 race condition
- **缺点**：不同团队共享序号空间，从编码序号无法推断团队内事项数量；跨团队的 `MAX(code)` 查询反而成为全局竞争热点，比 per-team 序号竞争更严重
- **结论**：排除。per-team 序号将竞争范围缩小到单团队内，而当前系统以团队为操作边界（每个用户只在一个团队内操作），per-team 序号更合理

### D. 选中方案：Team Code + per-team 5 位序号

即本提案描述的方案。

**关键设计决策的理由**：

| 决策 | 理由 |
|------|------|
| 快照 team code（非实时引用） | 编码不可变原则：如果 team code 修改后旧事项编码跟着变，会导致已分享/归档的编码失效 |
| per-team 序号（非全局） | 团队是操作边界，per-team 将并发竞争限制在单团队内；5 位序号支持单个团队 99999 个事项 |
| 5 位宽度（非 4 位） | 4 位仅支持 9999 个事项/团队，5 位支持 99999。当前已有团队接近百条事项，4 位余量不足 |
| 直接切换（非双格式兼容） | 系统尚在内测阶段，用户基数极小，兼容旧格式的复杂度不值得付出 |

**选中方案的代价**：code 生成逻辑更复杂（需查 team code + per-team MAX）；MainItem.Code 列需从 varchar(10) 扩展到 varchar(12)；SubItem 需新增 Code 列及索引；数据迁移需重写所有现有编码。这些代价在内测阶段（数据量小、可接受停机迁移）是合理的。

## Risks

| 风险 | 可能性 | 影响 | 缓解措施 |
|------|--------|------|----------|
| **数据迁移失败**：ALTER TABLE 扩展 Code 列 + 重写所有 MI-XXXX 编码 + 为 SubItem 生成编码时中途出错 | 低 | 高：编码数据损坏导致事项无法检索 | 迁移脚本包裹在事务内；迁移前全量备份数据库；编写迁移回滚脚本（还原 varchar(10) + 恢复旧编码值 + 删除 SubItem.code 列） |
| **NextCode 并发 race condition**：两个并发请求同时读到相同 `MAX(code)`，生成重复编码 | 中 | 高：违反唯一约束导致创建失败或数据不一致 | 利用现有 `idx_team_code` 唯一索引（team_id + code 联合唯一）作为最终防线，重复插入会返回唯一约束错误，由 service 层重试（最多 3 次）；长期方案可改用 `SELECT ... FOR UPDATE` 或数据库序列 |
| **NextSubCode 并发竞争**：同一主事项下并发创建子事项时序号冲突 | 中 | 中：子事项创建失败 | SubItem 的 Code 字段加 per-main-item 唯一索引，配合 service 层重试逻辑（最多 3 次）；唯一索引 + 重试是实际防线，不依赖并发量低的假设 |
| **迁移后旧编码引用失效**：用户已分享的 MI-XXXX 链接或书签无法访问 | 低 | 低：内测阶段，外部引用极少 | 前端搜索框增加对旧格式的模糊匹配重定向（如检测到 `MI-` 前缀自动跳转到新编码）；或直接接受——内测用户可通过事项列表重新定位 |
| **Team Code 缺失或无效**：团队创建时未设置 Code，或 Code 包含非字母字符，导致事项编码生成失败 | 低 | 高：该团队下无法创建事项 | 后端 Team 创建接口校验 Code 必填 + 正则 `^[A-Za-z]{2,6}$` + 全局唯一；数据库层加 CHECK 约束 + 唯一索引作为双重保障 |

## Out of Scope

- 编码的自定义前缀或可配置格式
- 编码的批量重命名
- 已删除事项编码的回收复用
- 编码的跨团队唯一性（编码仅在团队内唯一）

## Success Criteria

| 验收项 | 可测试条件 |
|--------|-----------|
| Team Code 字段 | Team 模型包含 Code 字段，varchar(6)，全局唯一索引生效；重复 Code 插入返回唯一约束错误 |
| Team Code 校验 | 创建团队时 Code 为空、长度 < 2 或 > 6、包含非字母字符、与已有 Code 重复，均返回 400 错误及对应提示信息 |
| 主事项编码格式 | 创建主事项时生成 `{TEAM_CODE}-NNNNN` 格式编码（如 `FEAT-00001`），5 位序号左补零；per-team 严格递增 |
| 子事项编码格式 | 创建子事项时生成 `{TEAM_CODE}-{MAIN_SEQ}-{NN}` 格式编码（如 `FEAT-00001-01`），2 位序号左补零；per-main-item 严格递增 |
| 编码不可变 | 更新 team code 后，该团队下已有事项编码不变（单元测试验证） |
| SubItem 迁移 | 迁移后 `SELECT COUNT(*) FROM sub_items WHERE code IS NULL OR code = ''` 返回 0；每个 SubItem 的 code 值符合 `{TEAM_CODE}-{MAIN_SEQ}-{NN}` 格式 |
| MainItem 迁移 | 所有 `main_items.code` 中的 `MI-XXXX` 格式被替换为 `{TEAM_CODE}-NNNNN`；迁移后 `SELECT COUNT(*) FROM main_items WHERE code LIKE 'MI-%'` 返回 0 |
| 前端展示 — 主事项 | ItemViewPage、TableViewPage 中每行主事项的 Badge 渲染 `{TEAM_CODE}-NNNNN`（如 `FEAT-00001`）；WeeklyViewPage 行首渲染 `{code} {title}`；MainItemDetailPage 标题旁 Badge 渲染 `{TEAM_CODE}-NNNNN` |
| 前端展示 — 子事项 | MainItemDetailPage 子事项表格中，每个子事项行的 Badge 渲染 `{TEAM_CODE}-{MAIN_SEQ}-{NN}`（如 `FEAT-00001-01`），取自 SubItem.Code 字段，非前端拼接 |
| 前端搜索 | 在 ItemViewPage 搜索框输入完整编码 `FEAT-00001`，结果列表仅包含该编码对应的一条事项（精确匹配）；输入 `FEAT-`，结果列表包含该团队所有事项（前缀过滤） |
| 并发编码生成 | 同一团队下并发创建 2 个主事项，两者生成不同编码且均成功入库（测试：启动 2 个 goroutine 同时调用 `Create`，断言两个 code 值不同且均非空）；同一主事项下并发创建 2 个子事项，同理生成不同编码 |
| 现有测试通过 | 迁移后所有现有单元测试和集成测试通过（NextCode 相关测试需更新期望值） |

## Open Questions

（已通过对话全部确认，无遗留问题）
