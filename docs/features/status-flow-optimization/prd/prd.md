# PRD: Status Flow Optimization

## Background

PM Work Tracker 的事项（MainItem/SubItem）状态管理存在 6 个已验证的问题，影响团队日常状态跟踪：

1. **前后端状态值不一致** — 前端用 "未开始/待评审/延期"，后端用 "待开始/待验收/已延期"。约 30% 的状态值无法正确渲染（`StatusBadge` 回退到 default 样式）
2. **MainItem 无状态机** — Update 接口直接修改 status 字段，任意值均可写入（`main_item_service.go:88-89`）
3. **StatusDropdown 不生效** — ItemViewPage 的状态下拉菜单未绑定 onClick，选择后无实际效果（`ItemViewPage.tsx:1070-1087`）
4. **无主子事项联动** — 子事项状态变化不影响主事项状态
5. **无状态变更记录** — 无法追溯状态变化历史
6. **"挂起"语义模糊** — 无注释说明含义，与 "阻塞中" 无区分

> **不延期的成本**：P3 导致事项总览页（团队主入口）状态切换完全不可用。P1 导致约 30% 状态值无法正确渲染。

## Goals

1. **统一状态枚举体系** — code（英文，存储传输）+ name（中文，显示），消除前后端不一致
2. **建立状态机** — MainItem/SubItem 各自拥有合法流转规则，非法转换被拒绝
3. **实现主子联动** — 子事项状态变化自动评估并调整主事项状态
4. **记录变更历史** — 每次状态变更留痕，支持审计追溯
5. **修复前端功能** — StatusDropdown 绑定 API、动态过滤合法目标、统一显示

### Success Metrics

- 所有状态值在前后端一致渲染（0 个 fallback）
- 0 个非法状态转换通过 API
- 100% 的状态变更被记录

## Scope

### In-Scope

- 状态枚举 code/name 体系（MainItem 7 种，SubItem 6 种）
- MainItem ChangeStatus 方法 + 状态机
- SubItem ChangeStatus 更新（移除待验收/已延期，挂起改为 pausing）
- 主子事项状态联动（含增删子事项触发）
- RecalcCompletion 与联动并存协作
- 终态副作用（completion=100 + actual_end_date）
- status_histories 表及记录逻辑
- available-transitions API 端点
- 前端 StatusDropdown 功能修复 + 动态过滤 + 确认对话框
- 前端状态显示统一（中文硬编码替换为英文 code）
- 前端延期标记（计算值，非状态）
- 前端联动失败 toast 通知

### Out-of-Scope

- 状态变更通知（邮件/站内信）
- 状态历史查看 UI 页面（仅提供 API）
- 甘特图/周报/表格视图中的状态相关调整

### Future Enhancements

1. **中优先级** — 终态撤销/重开：允许 PM 将 completed/closed 的事项重新打开
2. **低优先级** — 状态变更原因：blocking/pausing/reviewing→progressing 时要求填写原因

## Requirements

### R1: Status Enumeration

状态值改为 code + name 双字段。code 为英文（存储传输），name 为中文（显示）。

命名规则：持续状态用进行时（-ing），终态用过去时（-ed）。`pending` 视为例外。

**MainItem 状态集（7 种）：**

| code | name | type | description |
|------|------|------|-------------|
| `pending` | 待开始 | 持续 | 初始状态 |
| `progressing` | 进行中 | 持续 | 正在处理 |
| `blocking` | 阻塞中 | 持续 | 被外部因素阻塞 |
| `pausing` | 已暂停 | 持续 | 因更高优先级事项暂停 |
| `reviewing` | 待验收 | 持续 | 等待 PM 验收确认 |
| `completed` | 已完成 | 终态 | PM 验收通过，不可逆转 |
| `closed` | 已关闭 | 终态 | 不再追踪，不可逆转 |

**SubItem 状态集（6 种）：**

| code | name | type | description |
|------|------|------|-------------|
| `pending` | 待开始 | 持续 | 初始状态 |
| `progressing` | 进行中 | 持续 | 正在处理 |
| `blocking` | 阻塞中 | 持续 | 被外部因素阻塞 |
| `pausing` | 已暂停 | 持续 | 因更高优先级事项暂停 |
| `completed` | 已完成 | 终态 | 执行人标记完成，不可逆转 |
| `closed` | 已关闭 | 终态 | 不再追踪，不可逆转 |

> SubItem 无 reviewing，执行人可直接从 progressing 标记 completed。现有 SubItem 的 "待验收" 状态将在实现中移除。

**延期是计算值，不是状态。** 事项是否延期由 `expected_end_date < 当前日期` 计算得出，前端据此展示延期标记。

### R2: Status Transition Rules

**硬约束**：MainItem/SubItem 的 Update 接口不再允许修改 status 字段。所有状态变更必须通过 ChangeStatus 方法经状态机校验。

**自转换拒绝**：新状态与当前状态相同时返回错误。

**终态副作用**：转换到终态时，update fields 中一并设置 `completion=100` + `actual_end_date=当前时间`，并在 status_histories 中记录 completed_at。

**MainItem 流转：**

```
pending     → progressing, closed
progressing → blocking, pausing, reviewing, closed
blocking    → progressing
pausing     → progressing, closed
reviewing   → completed (仅 PM), progressing (仅 PM)
completed / closed → 终态，不可逆转
```

**SubItem 流转：**

```
pending     → progressing, closed
progressing → blocking, pausing, completed, closed
blocking    → progressing
pausing     → progressing, closed
completed / closed → 终态，不可逆转
```

> blocking/pausing 恢复后只能回到 progressing。如需变为其他持续状态需两步操作，保持流转图简单可预测。

**无子事项的主事项**：没有子事项时，`progressing → reviewing` 允许 PM 手动触发，后续 reviewing → completed 仍需 PM 验收。

### R3: Main-Sub Item Linkage

子事项状态变化（含新增、删除）时，按优先级从高到低匹配评估主事项状态，命中即停止：

| Priority | Condition | MainItem Target | Constraint |
|----------|-----------|-----------------|------------|
| 1 (最高) | 所有子事项均为 completed/closed，且至少一个 completed | `reviewing` | 主事项当前状态须允许 → reviewing |
| 2 | 所有子事项均为 closed（无 completed） | `closed` | 主事项当前状态须允许 → closed |
| 3 | 所有子事项均为 pausing（或 pausing + closed） | `pausing` | 主事项当前状态须允许 → pausing |
| 4 | 存在任意 blocking 子事项（且非全部终态） | `blocking` | 主事项当前为 pending 或 progressing |
| 5 | 存在任意 progressing 子事项 | `progressing` | 主事项当前为 pending |

**联动规则：**

- **新增子事项**：新增后立即触发联动。若主事项在 reviewing，新增 pending 子事项导致退回 progressing（此场景 reviewing → progressing 由联动触发，不限于 PM 手动）
- **删除子事项**：删除后对剩余子事项触发联动
- **无子事项**：不触发任何联动
- **联动失败**：主事项当前状态不允许转换到目标时，主事项状态不变，但 status_histories 记录意图及失败原因，前端展示 toast 通知用户

### R4: RecalcCompletion Coordination

现有 RecalcCompletion（按子事项 weight 加权平均计算主事项 completion）与新增联动并存：

| Mechanism | Trigger | Responsibility |
|-----------|---------|---------------|
| RecalcCompletion | 子事项 completion 值变化或变为 completed | 重算主事项 completion 百分比 |
| 状态联动 | 子事项 status 变化（含增删） | 调整主事项 status 字段 |

子事项变为 completed 时两者同时触发，固定顺序：先 RecalcCompletion（更新百分比），再联动评估（更新状态）。

### R5: Status Change Log

新增 `status_histories` 表：

| Column | Type | Description |
|--------|------|-------------|
| id | uint | 主键 |
| item_type | string | "main_item" / "sub_item" |
| item_id | uint | 事项 ID |
| from_status | string | 变更前状态 code |
| to_status | string | 变更后状态 code |
| changed_by | uint | 操作人 ID |
| is_auto | bool | 是否联动自动触发 |
| remark | string | 备注信息（联动失败原因等） |
| completed_at | timestamp | 终态转换时记录实际完成时间 |
| created_at | timestamp | 记录创建时间 |

### R6: Frontend

- **StatusDropdown 绑定**：onClick 调用后端 ChangeStatus API
- **动态过滤**：通过 `GET /api/v1/teams/:teamId/items/:itemId/available-transitions` 获取合法目标状态
- **状态显示统一**：StatusBadge 使用 code-to-name 映射
- **延期标记**：`expected_end_date < 当前日期` 且非终态时展示延期标记
- **PM 角色控制**：reviewing → completed/progressing 仅 PM 可见
- **不可逆确认**：closed/completed 操作前弹出确认对话框
- **联动失败通知**：子事项状态变更成功但主事项联动失败时，展示 toast 警告

## Phased Delivery

### Phase 1: State Machine Foundation

后端状态枚举体系 + 状态机核心，前端统一状态显示。

**Deliverables:**
- 状态枚举定义（code + name 映射）
- MainItem ChangeStatus + 状态机（10 条合法转换）
- SubItem ChangeStatus 更新（移除废弃状态，9 条合法转换）
- 移除 MainItemUpdateReq.Status 字段
- 移除 MainItem/SubItem 的 DelayCount 字段
- available-transitions API 端点
- 前端 StatusBadge 统一为 code-to-name 映射
- 前端 STATUS_OPTIONS 替换为英文 code

**Verify:** AC-1 ~ AC-5, AC-18, AC-23

### Phase 2: Linkage & Side Effects

主子联动 + 终态副作用 + 状态变更日志。

**Deliverables:**
- 主子事项状态联动（5 级优先级规则）
- 新增/删除子事项触发联动
- 终态副作用（completion=100 + actual_end_date）
- RecalcCompletion 协作（固定执行顺序）
- per-MainItem 互斥锁（并发控制）
- status_histories 表 + repo + service
- 联动失败 toast 通知

**Verify:** AC-6 ~ AC-15

### Phase 3: Frontend UX Polish

前端交互完善。

**Deliverables:**
- StatusDropdown 绑定 ChangeStatus API
- StatusDropdown 动态过滤（基于 available-transitions）
- 延期标记（计算值 UI）
- PM 角色控制（reviewing 操作可见性）
- 不可逆操作确认对话框

**Verify:** AC-16, AC-17, AC-19 ~ AC-21

## Acceptance Criteria

### State Machine

- **AC-1**: MainItem 7 个状态码和 SubItem 6 个状态码在数据库存储为英文 code，API 响应返回 code + name
- **AC-2**: MainItem 每条合法转换（10 条）通过 ChangeStatus 成功；非法转换返回 400 且数据库不变
- **AC-3**: SubItem 每条合法转换（9 条）通过 ChangeStatus 成功；非法转换返回 400
- **AC-4**: 自转换（新状态=当前状态）返回错误
- **AC-5**: MainItemUpdateReq 无 Status 字段，通过 Update 传入 status 不生效

### Terminal Side Effects

- **AC-6**: 进入 completed/closed 时 completion=100、actual_end_date=当前时间

### Linkage

- **AC-7**: 所有子事项均为 completed/closed 且至少一个 completed 时，主事项自动转 reviewing
- **AC-8**: 联动 5 个优先级场景均有对应测试验证主事项目标状态正确
- **AC-9**: reviewing 状态下新增 pending 子事项，主事项退回 progressing
- **AC-10**: 删除子事项后重新触发联动，主事项状态正确
- **AC-11**: 无子事项时不触发联动
- **AC-12**: 联动目标不被允许时主事项不变，status_histories 记录意图及原因，前端展示 toast

### RecalcCompletion

- **AC-13**: 子事项变 completed 时先 RecalcCompletion 再联动评估（顺序可由集成测试断言中间状态验证）

### Status History

- **AC-14**: 每次 ChangeStatus 成功后 status_histories 新增一条完整记录
- **AC-15**: 联动触发的变更 is_auto=true，手动触发的 is_auto=false

### API

- **AC-23**: available-transitions API 对 MainItem/SubItem 的每个当前状态返回正确的合法目标状态集合

### Frontend

- **AC-16**: StatusDropdown onClick 调用 ChangeStatus API，选择后状态实际改变
- **AC-17**: StatusDropdown 仅显示 available-transitions API 返回的合法选项
- **AC-18**: StatusBadge 使用 code-to-name 映射，不使用硬编码中文值
- **AC-19**: expected_end_date < 当前日期 且非终态时 UI 显示延期标记；终态不显示
- **AC-20**: reviewing → completed/progressing 仅 PM 角色可见
- **AC-21**: closed/completed 操作前弹出确认对话框，取消不执行

## Risk Assessment

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|-----------|--------|------------|
| R1 | API 破坏性变更 — 移除 UpdateReq.Status | High | Medium | 前端同步移除 Update 中 status 赋值。项目前后端同仓无外部消费者 |
| R2 | reviewing 下新增子事项的静默回退 | Medium | Medium | 前端新增子事项时若主事项 reviewing 则弹出提示；联动失败展示 toast |
| R3 | RecalcCompletion 与联动执行顺序错误 | Low | High | 固定顺序（先 Recalc 后联动）+ 集成测试覆盖完整路径 |
| R4 | 并发子事项变更竞态 | Low | Medium | 应用层 per-MainItem 互斥锁（sync.Mutex map 按 mainItemID 分片） |
| R5 | 前端状态硬编码替换回归 | Medium | Medium | 逐文件替换 + 每个文件替换后运行测试；前端 AC 覆盖所有显示场景 |

## Alternatives Considered

### A1: Unified Status Set (MainItem + SubItem share all 7 states)

Rejected. SubItem adding reviewing introduces unnecessary PM verification overhead. Current product design: PM manages MainItems, executors self-manage SubItems.

### A2: Event-Driven Linkage vs Synchronous

Rejected. Single-process app with SQLite; strong consistency via synchronous calls is more appropriate than introducing message queue infrastructure.

### A3: Overdue as Status vs Computed Value

Rejected. Overdue as status doubles the state space (each state needs +overdue variant) and requires manual recovery. Computed value auto-resolves when dates change.

### A4: blocking/pausing Recovery to Any State vs progressing Only

Rejected. Two-step recovery cost is lower than maintaining complex transition graphs. Real-world scenarios almost always resume to progressing.
