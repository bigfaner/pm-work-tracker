---
created: "2026-06-11"
author: fanhuifeng
status: Draft
intent: "enhancement"
---

# Proposal: Milestone Status Transition Guards

## Problem

里程碑图（MilestoneMap）和里程碑（Milestone）的状态转换缺少业务层约束，允许不符合实际项目管理流程的操作。例如：空里程碑可以进入进行中、里程碑可以跳过前序阶段直接推进、空图可以评审、已完成的里程碑可以随意重开。

### Evidence

当前实现中 BR-1~BR-6 覆盖了完成/删除/级联约束，但以下场景缺少守卫：

- 空里程碑（无挂载事项）可切换为 in_progress 或 completed
- 里程碑可跳过前序阶段：M1 未开始时 M2 可直接进入 in_progress
- 空里程碑图（无里程碑）可切换为 reviewed
- 里程碑图在 planning/reviewed 阶段即可变更里程碑状态
- completed 里程碑可重开为 in_progress
- 已评审后添加的里程碑无法与评审前里程碑区分管理
- ready/executing 阶段仍可增删里程碑结构

### Urgency

里程碑图功能已上线（feature completed），用户开始使用。缺少约束会导致数据不一致和操作混乱。越晚修复，已有数据的清理成本越高。

## Proposed Solution

为里程碑图和里程碑的状态转换添加 13 条业务约束规则，覆盖三个维度：

1. **里程碑前置条件**：空里程碑守卫、序列依赖、日期必填
2. **里程碑图→里程碑操作矩阵**：不同图状态下里程碑允许的操作不同
3. **转换矩阵修正**：禁止 completed → in_progress 重开

### 里程碑状态转换新矩阵

| From | To | 守卫条件 |
|------|----|---------|
| not_started | in_progress | 有事项 + 前序非 not_started + 图在 ready/executing + 有 expected_end_date |
| not_started | cancelled | 有事项 + 图非终态 |
| in_progress | completed | BR-1（所有事项终态） |
| in_progress | cancelled | 级联解绑事项 |
| completed | cancelled | 无额外条件 |
| cancelled | — | 终态 |

### 里程碑图状态→里程碑操作矩阵

| 图状态 | 新增里程碑 | 删除里程碑 | 编辑属性 | 状态变更 | 绑定/解绑事项 |
|--------|----------|----------|---------|---------|-------------|
| planning | YES | YES（BR-4） | YES | NO | YES |
| reviewed | YES | 仅评审后新增的 | YES | NO | YES |
| ready | NO | NO | YES（日期需校验） | YES | YES |
| executing | NO | NO | YES（日期需校验） | YES | YES |
| completed | NO（BR-5） | NO（BR-5） | NO（BR-5） | NO（BR-5） | NO（BR-5） |
| cancelled | NO（BR-5） | NO（BR-5） | NO（BR-5） | NO（BR-5） | NO（BR-5） |

### 里程碑图转换守卫

| 转换 | 守卫条件 |
|------|---------|
| planning → reviewed | 至少 1 个里程碑 |
| reviewed → planning | 无额外条件（保留所有里程碑） |
| reviewed → ready | 无额外条件 |
| ready → reviewed | 所有里程碑为 not_started |
| ready → executing | 无额外条件 |
| executing → ready | 无额外条件 |

### 序列规则

- 排序依据：里程碑的 expected_end_date（升序）
- 阻塞规则：如果前序里程碑为 not_started，后序里程碑不可进入 in_progress 或 completed
- 解除阻塞：cancelled（终态）的前序里程碑解除阻塞
- 日期编辑：允许修改日期，但修改后需重新校验序列合法性，不合法则拒绝

### Innovation Highlights

这不是创新性方案，而是项目管理的标准实践。核心思路来源于：
- 严格顺序流水线（类似 CI/CD pipeline 的 stage 依赖）
- 分阶段锁定（类似文档审批流程的 draft → reviewed → approved 锁定机制）
- 空状态守卫（防止无意义的状态推进）

## Requirements Analysis

### Key Scenarios

**正常流程：**
1. 创建图（planning）→ 添加里程碑（含日期）→ 绑定事项 → 评审（reviewed）→ 待实施（ready）→ 里程碑按序推进 → 全部完成 → 图完成（completed）

**约束拦截：**
2. 空里程碑尝试切换状态 → 422 错误，提示"请先添加事项"
3. M1 未开始时 M2 尝试开始 → 422 错误，提示"前序里程碑尚未开始"
4. 空图尝试评审 → 422 错误，提示"请先添加里程碑"
5. 评审后尝试删除评审前里程碑 → 422 错误
6. ready 状态尝试新增里程碑 → 422 错误
7. 有 in_progress 里程碑时回退到 reviewed → 422 错误
8. 修改日期导致序列冲突 → 422 错误

**边界情况：**
9. 评审后新增的里程碑，回退到 planning 后再次评审时，这些里程碑变为"评审前"里程碑
10. 多个里程碑 expected_end_date 相同时，序列如何判定
11. 里程碑取消后，后续里程碑自动解除阻塞

### Non-Functional Requirements

- 约束校验必须在现有 API 响应时间内完成（不增加显著延迟）
- 错误消息清晰可操作，用户能理解为什么被拒绝

### Constraints & Dependencies

- 依赖现有 milestone-map 功能（已完成）
- 依赖现有 status 转换引擎（`pkg/status/transition.go`）
- 需要区分"评审前"和"评审后"里程碑（需要追踪机制）
- 依赖 expected_end_date 字段（已存在）

## Alternatives & Industry Benchmarking

### Industry Solutions

主流项目管理工具（Jira、Linear、Asana）的状态约束通常是：
- 线性流水线：状态只能向前推进
- 前置条件：某些字段必填才能推进
- 阶段锁定：评审后结构不可变

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing | — | 零成本 | 数据不一致风险高，用户困惑 | Rejected: 约束缺失是实际问题 |
| 仅基础守卫 | 自定义 | 实现简单 | 覆盖不全，后续需要补 | Rejected: 分阶段增加迁移成本 |
| 完整约束系统 | Jira/Linear 模式 | 覆盖完整，行为可预测 | 实现工作量中等 | **Selected: 一次性解决，避免后续修补** |

## Feasibility Assessment

### Technical Feasibility

- 后端：在现有 service 层添加校验逻辑，扩展 AvailableTransitions 方法
- 前端：StatusTransitionDropdown 已调用 available-transitions API，前端基本无需改动
- 追踪评审前后里程碑：需要在里程碑图或里程碑上记录评审时间点

### Resource & Timeline

- 后端改动：service 层验证逻辑 + 错误码 ≈ 2-3 天
- 前端改动：错误提示更新 ≈ 0.5 天
- 测试：约束规则 E2E 测试 ≈ 1-2 天
- 总计：约 3-5 天

### Dependency Readiness

所有依赖功能已实现并上线。无外部依赖。

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| 空里程碑应该可以自由切换状态 | 5 Whys | Overturned: 空里程碑切换状态在实际项目中无意义，用户可能是误操作 |
| 里程碑可以任意顺序推进 | Assumption Flip | Overturned: 如果 M2 可以先于 M1 完成，时间线视图会产生逻辑矛盾 |
| 已完成的里程碑应该可以重开 | Stress Test | Refined: 重开会破坏后续里程碑的合法性（后续里程碑依赖前序完成），应禁止 |
| reviewed 状态下应完全锁定 | Assumption Flip | Refined: 完全锁定不现实（可能需要追加阶段），改为部分锁定（可新增但不可删原有关键字） |
| 日期相同的里程碑需要显式排序 | Occam's Razor | Confirmed: 日期相同时需要 tiebreaker，最简方案是使用创建时间或 bizKey 作为次要排序 |

## Scope

### In Scope

- 13 条约束规则的完整实现
- 里程碑转换矩阵更新（移除 completed → in_progress）
- 里程碑图状态→里程碑操作矩阵实现
- 序列校验逻辑（基于 expected_end_date）
- 评审前后里程碑区分追踪
- 新增/更新错误码
- 后端 service 层验证逻辑
- 前端错误提示展示
- AvailableTransitions API 扩展

### Out of Scope

- 新增 API 端点
- UI 组件重设计
- RBAC 权限变更
- 性能优化
- 数据迁移（现有数据不受影响，新约束只应用于后续操作）
- 相同日期里程碑的排序策略（后续迭代细化，当前使用创建时间作为 tiebreaker）

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 评审前后里程碑追踪增加 schema 复杂度 | M | M | 使用 reviewed_at 时间戳 + 里程碑 created_at 比较，避免新增字段 |
| 序列校验在里程碑数量大时性能下降 | L | M | 里程碑数量通常 <20，N² 校验可接受 |
| 约束过于严格影响用户灵活性 | M | M | ready/executing 状态保持属性和事项绑定灵活性 |
| reviewed → planning 回退后再次评审的追踪边界 | M | L | 明确规则：回退后再次评审时，所有已有里程碑变为"评审前" |

## Success Criteria

- [ ] 空里程碑（无事项）尝试切换任何状态时返回 422 错误，错误码 `EMPTY_MILESTONE`
- [ ] 前序里程碑为 not_started 时，后序里程碑无法切换为 in_progress 或 completed，返回 422 错误，错误码 `SEQUENCE_VIOLATION`
- [ ] 空里程碑图（0 个里程碑）无法切换为 reviewed，返回 422 错误，错误码 `EMPTY_MAP`
- [ ] completed 里程碑无法重开为 in_progress，转换矩阵中不再包含此路径
- [ ] planning/reviewed 状态下里程碑状态变更返回 422 错误，错误码 `MAP_STATE_BLOCKS_MILESTONE`
- [ ] reviewed 状态下删除评审前里程碑返回 422 错误，错误码 `REVIEWED_MILESTONE_LOCKED`
- [ ] ready/executing 状态下新增/删除里程碑返回 422 错误，错误码 `MAP_STRUCTURE_LOCKED`
- [ ] ready → reviewed 回退时如有非 not_started 里程碑返回 422 错误，错误码 `MILESTONE_NOT_RESETTABLE`
- [ ] 修改日期导致序列冲突时返回 422 错误，错误码 `DATE_SEQUENCE_CONFLICT`
- [ ] AvailableTransitions API 正确反映所有约束（前端无需额外判断）
- [ ] 所有约束规则有对应的后端单元测试和 E2E 测试
- [ ] cancelled 的前序里程碑不阻塞后序里程碑（序列验证正确跳过终态里程碑）

consistency_check_result:
  status: pass
  pairs_checked: 8
  conflicts_found: 0

## Next Steps

- Proceed to `/write-prd` to formalize requirements
