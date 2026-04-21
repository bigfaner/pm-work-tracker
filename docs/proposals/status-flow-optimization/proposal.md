# Proposal: Status Flow Optimization

## Problem

当前事项（MainItem/SubItem）的状态管理存在以下问题：

1. **前后端状态值不一致** — 前端用 "未开始/待评审/延期"，后端用 "待开始/待验收/已延期"，导致显示和逻辑混乱
2. **MainItem 无状态机** — Update 接口直接修改 status 字段，无合法转换校验
3. **StatusDropdown 不生效** — 前端状态下拉菜单未绑定 onClick 事件，选择后不会真正改变状态
4. **无主子事项联动** — 子事项状态变化不影响主事项状态
5. **无状态变更记录** — 无法追溯状态变化历史（谁在何时将状态从 A 改为 B）
6. **缺少"暂停"语义** — 现有"挂起"含义模糊，需要更明确的"因更高优先级事项而暂停"

### 代码验证证据

以上 6 个问题均可通过代码审计直接验证，以下是各问题对应的代码位置和影响范围：

**P1（前后端状态值不一致）：** 后端 SubItem 有 8 种状态码（`model/sub_item.go:8`），MainItem 默认值为 `"待开始"`（`model/main_item.go:19`）。前端 `StatusBadge` 使用 7 种中文值（`未开始/进行中/待评审/已完成/已关闭/阻塞中/延期`，见 `StatusBadge.tsx:5-12`），其中 `待评审` 和 `延期` 在后端不存在对应值。后端用 `待验收` 和 `已延期`，前端用 `待评审` 和 `延期`。**影响**：当后端返回 `"已延期"` 或 `"待验收"` 时，`StatusBadge` 的 `statusVariantMap` 找不到匹配项，回退到 `default` 样式，视觉上无法区分状态。涉及组件：`StatusBadge.tsx`、`ItemViewPage.tsx:52`（`STATUS_OPTIONS`）、`MainItemDetailPage.tsx:58`（`STATUS_OPTIONS`）、`TableViewPage.tsx:33`（`STATUS_OPTIONS`）。

**P2（MainItem 无状态机）：** `main_item_service.go:88-89` 的 `Update` 方法直接将 `req.Status` 写入数据库，无转换合法性校验。任意状态值均可通过 Update API 写入，包括无效值或非法跳转（如 `pending→completed`）。

**P3（StatusDropdown 不生效）：** `ItemViewPage.tsx:1070-1087` 中的 `StatusDropdown` 组件，其 `DropdownMenuItem` 未绑定 `onClick` 或 `onSelect` 处理器，点击选项后不会触发 API 调用。对比 `MainItemDetailPage.tsx:361`，子事项的 `DropdownMenuItem` 已绑定 `onSelect` 调用 `statusChangeMutation.mutate`，但 `ItemViewPage` 中未做同样绑定。**影响**：用户在事项总览页尝试切换状态，UI 有下拉交互但无实际效果，构成功能性 bug。

**P4（无主子事项联动）：** `sub_item_service.go:148-153` 中 `ChangeStatus` 仅在子事项变为 `"已完成"` 时调用 `s.mainItemSvc.RecalcCompletion`（重算完成百分比），未根据子事项状态变化调整主事项状态。

**P5（无状态变更记录）：** 代码库中不存在 `status_histories` 表或相关模型/仓库/服务，无法回答"谁在何时将状态从 A 改为 B"。

**P6（"挂起"语义模糊）：** 后端 `allowedTransitions`（`sub_item_service.go:35-43`）中 `"挂起"` 状态可转到 `"进行中"` 和 `"已关闭"`，但无注释或文档说明其含义，且与 `"阻塞中"`（外部阻塞）无区分。

> **不延期的成本**：P3 导致事项总览页（ItemViewPage，团队所有成员的主入口）状态切换完全不可用。P1 导致约 30% 的状态值（`待验收`/`已延期` 在 SubItem 中出现，`待评审` 在前端定义）无法正确渲染。这两个问题直接影响日常状态跟踪流程。

## Proposal

### 1. 状态枚举体系统一

状态值改为 code + name 双字段结构，code 为英文（存储和传输），name 为中文（显示）。MainItem 和 SubItem 拥有独立的状态集。

> 命名规则：持续状态使用进行时语态（-ing），终态使用过去时语法（-ed）。`pending` 作为通用初始状态词汇视为例外。

> **延期是计算值，不是状态。** 事项是否延期由 `expected_end_date < 当前日期` 计算得出，不作为状态枚举值。前端可根据计算结果在 UI 上展示延期标记。

**MainItem 状态集（7种）：**

| code | name | 类型 | 说明 |
|------|------|------|------|
| `pending` | 待开始 | 持续 | 初始状态 |
| `progressing` | 进行中 | 持续 | 正在处理 |
| `blocking` | 阻塞中 | 持续 | 被外部因素阻塞 |
| `pausing` | 已暂停 | 持续 | 因更高优先级事项暂停 |
| `reviewing` | 待验收 | 持续 | 等待PM验收确认 |
| `completed` | 已完成 | 终态 | PM验收通过，不可逆转 |
| `closed` | 已关闭 | 终态 | 不再追踪，不可逆转 |

**SubItem 状态集（6种）：**

| code | name | 类型 | 说明 |
|------|------|------|------|
| `pending` | 待开始 | 持续 | 初始状态 |
| `progressing` | 进行中 | 持续 | 正在处理 |
| `blocking` | 阻塞中 | 持续 | 被外部因素阻塞 |
| `pausing` | 已暂停 | 持续 | 因更高优先级事项暂停 |
| `completed` | 已完成 | 终态 | 执行人标记完成，不可逆转 |
| `closed` | 已关闭 | 终态 | 不再追踪，不可逆转 |

> SubItem 无 `reviewing`，执行人可直接从 `progressing` 标记 `completed`。现有 SubItem 的 `待验收` 状态及对应测试用例将在实现中移除。

### 2. 状态流转规则

> **硬约束**：MainItem/SubItem 的 Update 接口不再允许修改 status 字段。所有状态变更必须通过 ChangeStatus 方法，经状态机校验后才能生效。Update 接口中的 status 字段将被移除。
>
> **自转换**：调用 ChangeStatus 时新状态与当前状态相同，视为无效操作，返回错误。
>
> **终态副作用**：事项转换到终态（`completed` 或 `closed`）时，ChangeStatus 方法在更新数据库记录时一并执行：
> - 子事项：直接在 SubItem 的 update fields 中将 `completion` 设为 100、`actual_end_date` 设为当前时间
> - 主事项：同理，直接在 MainItem 的 update fields 中将 `completion` 设为 100、`actual_end_date` 设为当前时间
> - 在 `status_histories` 表中同步记录该实际完成时间

**MainItem 流转：**

```
pending     → progressing, closed
progressing → blocking, pausing, reviewing, closed
blocking    → progressing
pausing     → progressing, closed
reviewing   → completed (仅PM), progressing (仅PM)
completed / closed → 终态，不可逆转
```

> `reviewing` 的两条出口均仅限 PM 操作：验收通过 → `completed`，验收不通过打回 → `progressing`。

**SubItem 流转：**

```
pending     → progressing, closed
progressing → blocking, pausing, completed, closed
blocking    → progressing
pausing     → progressing, closed
completed / closed → 终态，不可逆转
```

> `blocking`/`pausing` 恢复后只能回到 `progressing`。如需变为其他持续状态，需两步操作（先恢复再转换），保持流转图简单可预测。

**无子事项的主事项流转：**

主事项没有子事项时，`progressing → reviewing` 允许由 PM 手动触发（跳过"全部子事项完成"的联动前提），后续 `reviewing → completed` 仍需 PM 验收。

### 3. 主子事项状态联动

当子事项状态变化（含新增、删除子事项）时，重新评估主事项状态。联动按优先级从高到低匹配，命中即停止：

| 优先级 | 触发条件 | 主事项目标状态 | 约束 |
|--------|----------|---------------|------|
| 1（最高） | 所有子事项均为 `completed` 或 `closed`，且至少有一个 `completed` | `reviewing` | 主事项当前状态须允许 → reviewing |
| 2 | 所有子事项均为 `closed`（无 `completed`） | `closed` | 主事项当前状态须允许 → closed |
| 3 | 所有子事项均为 `pausing`（或 `pausing` + `closed`） | `pausing` | 主事项当前状态须允许 → pausing |
| 4 | 存在任意 `blocking` 子事项（且非全部终态） | `blocking` | 主事项当前为 `pending` 或 `progressing` |
| 5 | 存在任意 `progressing` 子事项 | `progressing` | 主事项当前为 `pending` |

**联动规则说明：**

- **新增子事项**：新增后立即触发联动评估。如主事项在 `reviewing`，新增一个 `pending` 子事项导致不再满足"全部终态"，主事项退回 `progressing`（此场景特例：`reviewing → progressing` 由联动触发，不限于 PM 手动）。
- **删除子事项**：删除后对剩余子事项触发联动评估。
- **无子事项**：不触发任何联动。主事项状态完全由手动操作控制。
- **混合状态**：按优先级从高到低匹配。例如 1 completed + 1 blocking + 1 pending → 不满足优先级 1-3，满足优先级 4（存在 blocking）→ 主事项 blocking。
- **联动失败处理**：若主事项当前状态不允许转换到目标状态，联动不生效，但写入 status_history（记录联动意图及失败原因）供排查。

### 4. RecalcCompletion（现有逻辑）与新增联动的协作

**RecalcCompletion** 是现有方法（`main_item_service.go:139`），职责是按子事项 weight 加权平均计算主事项的 `completion` 百分比（0-100）。当前在 SubItem 变为 `completed` 时被调用。

本次新增的状态联动不会替代 RecalcCompletion，两者并存、职责不同：

| 机制 | 来源 | 触发时机 | 职责 |
|------|------|---------|------|
| RecalcCompletion | 现有，保留 | 子事项 `completion` 值变化或变为 `completed` | 重算主事项的 `completion` 百分比（数值字段） |
| 状态联动 | 新增 | 子事项 status 变化（含增删） | 调整主事项的 `status` 字段（状态枚举） |

子事项变为 `completed` 时同时触发两者：先 RecalcCompletion 更新百分比，再执行联动评估更新状态。

> **注意**：当子事项进入终态时，其自身的 `completion` 被直接设为 100（见第 2 节终态副作用）。随后 RecalcCompletion 仍会基于所有子事项的 completion（含这个 100）重新计算主事项的加权完成度。当主事项自身进入终态时，其 completion 也被直接设为 100，不再依赖 RecalcCompletion。

### 5. 状态变更日志

新增 `status_histories` 表，记录每次状态变更：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| item_type | string | "main_item" / "sub_item" |
| item_id | uint | 事项 ID |
| from_status | string | 变更前状态 code |
| to_status | string | 变更后状态 code |
| changed_by | uint | 操作人 ID |
| is_auto | bool | 是否为联动自动触发 |
| remark | string | 备注信息（联动失败原因等） |
| completed_at | timestamp | 终态转换时记录实际完成时间 |
| created_at | timestamp | 记录创建时间 |

### 6. 前端修复

- **StatusDropdown 绑定事件**：onClick 调用后端 ChangeStatus API
- **动态过滤目标状态**：前端根据当前状态和流转规则，只显示合法的目标状态选项
- **状态显示统一**：StatusBadge 使用 code-to-name 映射显示中文名称
- **延期标记**：根据 `expected_end_date < 当前日期` 且状态非终态，在 UI 上展示延期标记（如红色标记或角标），不再是独立状态
- **PM 角色控制**：主事项的 `reviewing → completed` 和 `reviewing → progressing` 仅对 PM 角色可见/可操作
- **不可逆操作确认**：`closed`、`completed`（MainItem）操作前弹出确认对话框
- **流转规则同步**：前端通过后端 API 获取当前事项的合法目标状态列表（`GET /api/v1/teams/:teamId/items/:itemId/available-transitions`），而非前端硬编码流转规则

### 7. 代码变更要点

#### 7.1 后端状态枚举替换

**移除 SubItem 废弃状态（`已延期`/`待验收`/`挂起`）的代码：**

| 文件 | 函数/常量 | 变更 |
|------|-----------|------|
| `service/sub_item_service.go:35-43` | `allowedTransitions` map | 移除 `"已延期"` 和 `"待验收"` 条目，`"挂起"` 改为 `"pausing"`，所有中文 key 替换为英文 code |
| `service/sub_item_service.go:53` | `Create()` 中 `Status: "待开始"` | 改为 `Status: "pending"` |
| `service/sub_item_service.go:128-136` | `ChangeStatus()` 中 `newStatus == "已延期"` 分支 | 整段删除（delay_count 自增 + priority 升级逻辑） |
| `service/sub_item_service.go:139` | `ChangeStatus()` 中 `newStatus == "已完成"` 分支 | 改为 `newStatus == "completed"`，新增 `completion=100` |
| `service/main_item_service.go:49` | `Create()` 中 `Status: "待开始"` | 改为 `Status: "pending"` |
| `service/main_item_service.go:88-89` | `Update()` 中 `req.Status` 分支 | 删除整个 `if req.Status != nil` 块 |
| `service/main_item_service.go:117` | `Archive()` 中 `item.Status != "已完成"` 判断 | 改为 `item.Status != "completed"` |
| `model/sub_item.go:23` | `DelayCount` 字段 | 移除（延期改为计算值） |
| `model/main_item.go:22` | `DelayCount` 字段 | 移除 |

#### 7.2 后端测试需更新的文件

| 文件 | 涉及测试数 | 变更说明 |
|------|-----------|---------|
| `service/sub_item_service_test.go` | 22 个测试 | 所有中文状态值替换为英文 code；删除 `TestChangeStatus_已延期_*`（3 个）；删除 `TestChangeStatus_待验收_*`（3 个）；删除 `TestChangeStatus_挂起_*`（2 个） |
| `service/main_item_service_test.go` | 16 个测试 | `"待开始"` → `"pending"`，`"已完成"` → `"completed"`，`"已关闭"` → `"closed"`，`"进行中"` → `"progressing"` |
| `service/view_service_test.go` | ~40 处状态引用 | 所有测试数据中的中文状态值替换为英文 code |
| `service/item_pool_service_test.go` | 1 处 | `"待开始"` → `"pending"` |
| `service/report_service_test.go` | 3 处 | 状态值替换 |
| `repository/gorm/main_item_repo_test.go` | ~20 处 | 状态值替换 |
| `repository/gorm/sub_item_repo_test.go` | ~15 处 | 状态值替换 |
| `repository/gorm/progress_repo_test.go` | 2 处 | 状态值替换 |
| `handler/main_item_handler_test.go` | ~5 处 | 状态值替换，移除 Update 请求中的 status 字段 |
| `handler/sub_item_handler_test.go` | ~10 处 | 状态值替换 |
| `model/main_sub_item_test.go` | 待确认 | 状态值替换 |

#### 7.3 前端中文硬编码替换

| 文件 | 当前值 | 替换为 |
|------|--------|--------|
| `components/shared/StatusBadge.tsx:5-12` | `statusVariantMap` 的 7 个中文 key | 英文 code key + code-to-name 映射 |
| `components/shared/StatusBadge.test.tsx` | 7 个中文状态测试数据 | 英文 code + 中文名显示验证 |
| `pages/ItemViewPage.tsx:52` | `STATUS_OPTIONS` 数组（7 个中文值） | 英文 code 数组 |
| `pages/MainItemDetailPage.tsx:58` | `STATUS_OPTIONS` 数组（7 个中文值） | 英文 code 数组 |
| `pages/TableViewPage.tsx:33` | `STATUS_OPTIONS` 数组（7 个中文值） | 英文 code 数组 |
| `pages/ItemViewPage.tsx:1070-1087` | `StatusDropdown` 组件 | 绑定 onClick/ChangeStatus API 调用 |
| `mocks/handlers.ts` | 中文状态值 | 英文 code |

#### 7.4 后端新增代码

- `dto/item_dto.go`：`MainItemUpdateReq` 移除 `Status *string` 字段（第 96 行）
- 新增 MainItem `ChangeStatus` 方法及状态机（`service/main_item_service.go`）
- 新增 per-MainItem 互斥锁 map（用于 R5 并发控制）
- 新增 `status_histories` 表、model、repo、service
- 新增 `available-transitions` API endpoint（`handler/main_item_handler.go`）

## Alternatives Considered

### A1: 统一状态集（MainItem 与 SubItem 共用同一套状态）

将 `reviewing` 也加入 SubItem，使主/子事项使用完全相同的 7 个状态。

| 维度 | 统一 7 状态 | 当前方案（MainItem 7 / SubItem 6） |
|------|-----------|----------------------------------|
| 模型一致性 | 高，一套状态枚举 | 中，两套枚举但共享 6 个 code |
| 验收流程 | 子事项也需要 PM 验收，增加流程负担 | 子事项由执行人直接标记完成 |
| 联动复杂度 | 主/子都有 reviewing，联动判断更复杂 | 子事项 completed 直接触发主事项 reviewing |
| 适用场景 | 适合 PM 深度管控子任务的团队 | 适合子事项由执行人自治、PM 只验收主事项的团队 |

**结论：** 拒绝。当前项目的产品设计是 PM 管控主事项、执行人自治子事项，统一状态集会为子事项引入不必要的验收环节。

### A2: 事件驱动联动 vs. 同步联动

用消息队列或事件总线解耦子事项变更与主事项联动。

| 维度 | 事件驱动 | 当前方案（同步联动） |
|------|---------|-------------------|
| 一致性 | 最终一致，存在短暂窗口期主/子状态不一致 | 强一致，同一事务内完成 |
| 复杂度 | 引入消息队列依赖，需处理重试、幂等 | 无额外基础设施，代码在 service 层直接调用 |
| 性能 | 子事项 API 响应更快 | 子事项 API 包含联动计算，响应略慢 |
| 适用场景 | 高并发、微服务架构 | 单体应用、团队规模小 |

**结论：** 拒绝。当前项目是单体应用，并发量低，强一致性比性能更重要。事件驱动引入的基础设施成本不合理。

### A3: 延期作为状态 vs. 计算值

将"延期"保留为状态枚举值。

| 维度 | 延期作为状态 | 当前方案（计算值） |
|------|-----------|-----------------|
| 状态膨胀 | 每个状态需要 +延期变体（如 progressing-overdue），状态数翻倍 | 状态数不变 |
| 延期恢复 | 需要手动将状态从"已延期"改回，容易遗漏 | 日期到达后自动消失，无需手动操作 |
| 查询灵活性 | 需要遍历所有延期相关状态码 | 一条查询条件（expected_end_date < now AND status 非终态） |
| 用户理解 | "延期"是主观判断，硬编码为状态可能与实际不符 | 延期标记是提示信息，不阻塞流转 |

**结论：** 拒绝。延期是时间推导的结果，不是事项的固有属性。作为状态会导致状态膨胀和手动恢复负担。

### A4: blocking/pausing 可恢复到任意持续状态 vs. 仅恢复到 progressing

| 维度 | 恢复到任意状态 | 仅恢复到 progressing |
|------|-------------|--------------------|
| 操作步数 | 一步到位 | 阻塞解除后需两步（→ progressing → 目标状态） |
| 流转图复杂度 | 每个 blocking/pausing 有 3-4 条出口 | 每个 blocking/pausing 仅 1 条出口 |
| 真实场景 | 少见"从阻塞直接到暂停"的合理场景 | 大多数场景是阻塞/暂停解除后回到执行 |
| 可预测性 | 低，组合爆炸 | 高，恢复路径唯一 |

**结论：** 拒绝任意恢复。两步操作的成本低于维护复杂流转图的成本，且与真实使用场景匹配。

## Risk Assessment

| # | 风险 | 可能性 | 影响 | 缓解措施 |
|---|------|--------|------|---------|
| R1 | **数据迁移失败** — 现有事项处于已废弃状态（"已延期"/"挂起"）无法映射到新状态码 | 中 | 高 | 编写一次性迁移脚本，映射规则：`已延期→progressing`（延期由计算值替代）、`挂起→pausing`、`待验收→reviewing`（MainItem）/ `progressing`（SubItem）。迁移前备份数据库，迁移后运行验证脚本检查所有事项状态值是否在合法枚举范围内 |
| R2 | **API 破坏性变更** — 移除 `MainItemUpdateReq.Status` 字段导致现有调用方报错 | 高 | 中 | 移除前搜索代码库确认所有调用方。前端在实现 ChangeStatus API 时同步移除 Update 中的 status 赋值。由于项目前后端在同一仓库且无外部消费者，影响可控 |
| R3 | **reviewing 状态下新增子事项的静默回退** — 用户在主事项 reviewing 时新增子事项，主事项自动退回 progressing，用户可能不理解 | 中 | 中 | 前端在新增子事项时，若主事项处于 reviewing，弹出提示："添加子事项将使主事项退回到进行中"。联动失败日志也会记录此回退事件。**API 层面**：后端 SubItem Create 接口（`sub_item_handler.go`）同样会触发联动（联动逻辑在 service 层，不依赖前端），因此即使通过 API 直接创建子项，联动仍会执行。前端确认对话框仅为 UX 提示，不是联动的唯一防线 |
| R4 | **RecalcCompletion 与联动执行顺序错误** — 子事项 completed 时两者同时触发，顺序错误会导致主事项状态不一致 | 低 | 高 | 在 ChangeStatus 中固定执行顺序：先 RecalcCompletion（更新百分比），再联动评估（更新状态）。编写集成测试覆盖"子事项→completed→主事项 completion=100 且 status=reviewing"的完整路径 |
| R5 | **并发子事项变更竞态** — 多个子事项几乎同时变更状态，触发多次联动评估，可能导致主事项状态反复切换或最终不一致 | 低 | 中 | 项目使用 glebarez/sqlite（纯 Go 驱动），不支持 `SELECT ... FOR UPDATE` 行锁。改用应用层 per-MainItem 互斥锁（`sync.Mutex` map，按 mainItemID 分片）序列化同一主事项的联动评估。ChangeStatus 获取目标主事项的互斥锁后再执行状态更新 + 联动，确保同一主事项的联动串行执行。此方案在单进程部署（SQLite 限制）下完全可行 |

## Acceptance Criteria

以下每条标准均为可验证的通过/失败判定，覆盖所有 In-Scope 条目。

### 状态枚举与状态机

- **AC-1**: MainItem 的 7 个状态码（`pending/progressing/blocking/pausing/reviewing/completed/closed`）和 SubItem 的 6 个状态码（`pending/progressing/blocking/pausing/completed/closed`）在数据库中存储为英文 code，API 响应中返回 code + name 双字段
- **AC-2**: MainItem 的每条合法转换路径（共 10 条）均可通过 ChangeStatus 成功执行；所有非法转换（如 `pending→completed`、`completed→progressing`）返回 400 错误，且数据库中状态不变
- **AC-3**: SubItem 的每条合法转换路径（共 9 条）均可通过 ChangeStatus 成功执行；所有非法转换返回 400 错误
- **AC-4**: 调用 ChangeStatus 时新状态与当前状态相同，返回错误（自转换拒绝）
- **AC-5**: `MainItemUpdateReq` 中无 `Status` 字段，通过 Update 接口传入 status 不生效（字段被忽略或返回错误）

### 终态副作用

- **AC-6**: 事项进入 `completed` 或 `closed` 时，`completion` 被设为 100、`actual_end_date` 被设为当前时间，在数据库中可验证

### 主子事项联动

- **AC-7**: 当所有子事项均为 `completed`/`closed` 且至少一个 `completed` 时，主事项自动转为 `reviewing`（假设主事项当前状态允许）
- **AC-8**: 联动优先级规则覆盖所有 5 个优先级场景，每个优先级有对应的测试用例验证主事项目标状态正确
- **AC-9**: 主事项处于 `reviewing` 时新增一个 `pending` 子事项，主事项退回 `progressing`
- **AC-10**: 删除子事项后，对剩余子事项重新触发联动评估，主事项状态正确更新
- **AC-11**: 主事项无子事项时，不触发任何联动
- **AC-12**: 联动目标状态不被当前状态允许时，主事项状态不变，但 `status_histories` 中记录联动意图及失败原因

### RecalcCompletion 协作

- **AC-13**: 子事项变为 `completed` 时，先执行 RecalcCompletion 更新主事项 completion 百分比，再执行联动评估更新主事项状态（顺序可通过集成测试断言中间状态验证）

### 状态变更日志

- **AC-14**: 每次调用 ChangeStatus 成功后，`status_histories` 表中新增一条记录，包含正确的 `item_type`、`item_id`、`from_status`、`to_status`、`changed_by`、`is_auto` 字段值
- **AC-15**: 联动触发的自动状态变更，`is_auto` 为 true；手动调用 ChangeStatus 触发的变更，`is_auto` 为 false

### 前端

- **AC-16**: StatusDropdown onClick 调用后端 ChangeStatus API，选择状态后事项状态实际改变（可通过刷新页面或重新查询验证）
- **AC-17**: StatusDropdown 仅显示当前事项的合法目标状态选项（通过 `available-transitions` API 获取），非法状态不在下拉选项中
- **AC-18**: StatusBadge 使用 code-to-name 映射显示中文名称，不再使用前端硬编码中文值
- **AC-19**: 当 `expected_end_date < 当前日期` 且状态为非终态时，UI 显示延期标记；终态事项不显示延期标记
- **AC-20**: MainItem 的 `reviewing→completed` 和 `reviewing→progressing` 操作仅对 PM 角色可见
- **AC-21**: 执行 `closed` 或 `completed`（MainItem）操作前弹出确认对话框，取消则不执行

### 数据迁移

- **AC-22**: 现有处于废弃状态（"已延期"/"挂起"/"待验收"）的事项全部迁移到新状态码，迁移后无事项的状态值不在合法枚举范围内

## Scope

### In-Scope

- 状态枚举 code/name 体系设计
- MainItem ChangeStatus 方法 + 状态机
- SubItem ChangeStatus 更新（移除待验收/已延期，挂起→pausing）
- 主子事项状态联动（含增删子事项）
- RecalcCompletion 与联动并存
- 终态副作用（completion=100 + actual_end_date）
- status_histories 表及记录逻辑
- 前端延期标记（计算值，非状态）
- 前端 StatusDropdown 功能修复 + 动态过滤 + 确认对话框
- 前端状态显示统一（中文硬编码→英文 code）
- 前端通过 API 获取合法目标状态

### Out-of-Scope

- 状态变更通知（如邮件/站内信）— 可作为后续增强
- 状态历史查看 UI 页面（仅提供 API）
- 甘特图/周报/表格视图中的状态相关调整（可独立处理）

### 后续增强优先级

1. **中优先级** — 终态撤销/重开：允许 PM 将 `completed`/`closed` 的事项重新打开
2. **低优先级** — 状态变更原因：`blocking`/`pausing`/`reviewing → progressing` 时要求填写原因

## Impact

- **后端**：model 层新增枚举定义、service 层新增 MainItem.ChangeStatus + 联动逻辑 + 终态副作用、`MainItemUpdateReq`（`dto/item_dto.go:96`）移除 `Status` 字段、移除 `ChangeStatus` 中 `newStatus == "已延期"` 分支（`sub_item_service.go:128-136`，含 `delay_count` 自增和 `is_key_item`/`priority` 自动升级逻辑）、移除 `MainItem` 和 `SubItem` model 的 `DelayCount` 字段（`main_item.go:22`、`sub_item.go:23`）、新增 status_history repo/service、新增 available-transitions API
- **前端**：`StatusBadge.tsx` 的 `statusVariantMap`（7 个中文 key）替换为英文 code + code-to-name 映射；`ItemViewPage.tsx:52`、`MainItemDetailPage.tsx:58`、`TableViewPage.tsx:33` 三处 `STATUS_OPTIONS` 数组替换为英文 code；`ItemViewPage.tsx:1070-1087` 的 `StatusDropdown` 绑定 ChangeStatus API 调用；`ItemFilters.tsx` 的 `statusOptions` prop 改为接收英文 code；新增延期计算标记逻辑；不可逆操作增加确认对话框
- **数据库**：新增 `status_histories` 表；`main_items` 和 `sub_items` 表移除 `delay_count` 列；`status` 列值从中文迁移为英文 code
