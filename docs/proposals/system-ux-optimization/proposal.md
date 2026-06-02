---
created: "2026-06-02"
author: "fanhuifeng"
status: Draft
intent: "new-feature"
---

# Proposal: System UX Optimization Batch

## Problem

PM Work Tracker 存在 10 项用户体验和功能缺陷，涵盖状态流转、表单交互、权限、删除、过滤等方面，影响日常使用效率。

### Evidence

- 状态切换失败时仅显示2秒 tooltip 提示"暂无可用流转"，用户无法理解失败原因
- 子事项编辑表单缺少开始时间字段，后端已支持但前端未暴露
- 主事项无删除功能（仅有归档），子事项删除未暴露为 API
- 待办事项转换表单的负责人字段无前端校验，提交后由后端报错
- 内置 member 角色用户登录后获取不到任何权限
- 卡片视图和表格视图的过滤器为单选，不支持多状态筛选，且不穿透子事项
- 新增表单关闭后不清空字段，再次打开显示残留数据
- 子事项列表无排序，数据以插入顺序展示

### Urgency

这些问题直接影响 PM 和团队成员的日常操作效率。#8 权限 bug 导致 member 角色用户完全无法使用系统。

## Proposed Solution

分两阶段实施 10 项优化：

**阶段一（Bug 修复 + 体验增强）：** 修复权限 bug，完善状态切换提示，补全表单校验和交互细节，统一排序和清空行为。

**阶段二（新功能）：** 新增子事项移动和过滤穿透功能。

### Innovation Highlights

- 过滤穿透：负责人筛选时，如果某子事项匹配，则连带展示其主事项。这需要后端支持子→父反向查询，卡片视图从客户端过滤改为服务端过滤
- 子事项移动时自动重新编号，保持编号与父事项的一致性

## Requirements Analysis

### Key Scenarios

**状态切换（#1）：**
- 用户点击状态流转按钮 → 后端返回不可流转原因 → 前端展示友好错误消息（如"该主事项下还有未完成的子事项，无法关闭"）
- 终态流转保留确认对话框

**编辑子事项开始时间（#2）：**
- 用户打开子事项编辑弹窗 → 看到"开始时间"字段 → 修改并保存

**PM 删除事项（#3）：**
- PM 点击删除按钮 → 确认对话框 → 软删除主事项及其所有子事项
- 非 PM 角色看不到删除按钮
- 子事项也可单独删除

**转换表单（#4, #7）：**
- 待办→子事项：开始时间默认当天，描述字段置灰且不可编辑
- 待办→主事项/子事项：负责人和优先级为必填，未填时提交按钮禁用

**子事项排序（#5）：**
- 主事项详情页的子事项列表按创建时间倒序（最新在前）

**表单清空（#6）：**
- 所有新增/转换表单关闭时（含取消）自动清空所有字段

**Member 角色权限（#8）：**
- member 角色用户登录后能获取到该角色的权限列表

**移动子事项（#9）：**
- 用户在子事项详情/编辑界面选择"移动到其他主事项" → 选择目标主事项 → 子事项重新编号并关联到新主事项
- 状态和负责人保持不变

**过滤穿透（#10）：**
- 状态过滤器改为多选（Checkbox Group 或 Multi-Select）
- 负责人过滤器：选中负责人 A 时，展示 A 负责的主事项 + 含 A 负责子事项的主事项（连同该子事项一起展示）
- 卡片视图和表格视图统一支持

### Non-Functional Requirements

- 删除操作需二次确认
- 过滤穿透查询性能：主事项 + 子事项联合过滤应在 500ms 内返回

### Constraints & Dependencies

- 后端已支持子事项 startDate 更新（SubItemUpdateReq.StartDate）
- 后端 TableView DTO 已支持状态多选（TableFilter.Status []string）
- 删除需新增权限码（main_item:delete, sub_item:delete）
- 子事项编号重新生成依赖现有的编号服务

## Alternatives & Industry Benchmarking

### Industry Solutions

项目管理工具（Jira, Linear, Asana）普遍支持：批量删除、子任务移动、多状态过滤。这些是基本功能而非创新。

### Comparison Table

| Approach | Source | Pros | Cons | Verdict |
|----------|--------|------|------|---------|
| Do nothing | — | 零开发成本 | 用户体验持续恶化，member 用户无法使用 | Rejected: 权限 bug 阻塞使用 |
| 仅修复 bug（#4-#8） | — | 最小范围 | 核心体验问题（删除、过滤、移动）未解决 | Rejected: 不完整 |
| 分阶段全量实施 | 本提案 | 按优先级交付，风险可控 | 阶段二较复杂 | **Selected: 平衡覆盖面和交付风险** |

## Feasibility Assessment

### Technical Feasibility

- 所有改动在现有技术栈内完成，无新依赖
- #9 子事项移动：需新增 API + 前端 UI + 编号重生成，中等工作量
- #10 过滤穿透：卡片视图需从客户端过滤改为服务端过滤，工作量较大

### Resource & Timeline

- 阶一 8 项：大部分为小改动，预计 1-2 天
- 阶段二 2 项：新功能开发，预计 2-3 天

### Dependency Readiness

- 后端已有子事项 startDate 更新、TableFilter 多选支持
- 编号服务已存在（用于创建时的编号生成）

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| 待办→子事项表单缺少 startDate 默认值 | Code inspection | Overturned: 代码已设置 `startDate: today`（ItemPoolPage.tsx:456），仅需修复描述字段置灰样式 |
| 子事项删除功能完全不存在 | Code inspection | Refined: Service 层已实现 soft-delete，但未暴露为 HTTP endpoint |
| 卡片视图过滤必须在客户端实现 | Assumption Flip | Overturned: 可以改为服务端过滤以支持穿透，但需新增 API |

## Scope

### In Scope

**阶段一：**
1. 状态切换失败时展示后端返回的错误消息（替代 tooltip）
2. 子事项编辑表单添加开始时间字段
3. PM 可软删除主事项（级联子事项）和子事项，需确认对话框
4. 待办→子事项表单描述字段添加置灰样式（disabled appearance）
5. 主事项详情页子事项列表倒序排列（ORDER BY id DESC）
6. 所有新增/转换表单关闭时清空字段
7. 转换表单（待办→主事项/子事项）负责人必填校验
8. 修复 member 角色用户登录后获取不到权限的 bug

**阶段二：**
9. 子事项移动到其他主事项（重新编号，保留状态和负责人）
10. 状态过滤器改为多选 + 负责人过滤穿透子事项（卡片视图和表格视图统一支持）

### Out of Scope

- 硬删除（永久清除数据）
- 删除通知/审计日志
- 批量操作（批量删除、批量移动）
- 子事项移动历史记录
- 卡片视图的拖拽排序
- 新增权限码的 seed 数据更新（由 RBAC 提案覆盖）

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| #8 权限 bug 根因不明（可能是 seed 数据或中间件逻辑） | M | H | 先定位根因（检查 member 角色的 RoleKey 是否为 nil）再修复 |
| #10 卡片视图改为服务端过滤，API 设计需与现有列表 API 兼容 | M | M | 复用 TableView 的后端过滤逻辑，或新增专用 API |
| #9 子事项移动时编号重生成可能与已有编号冲突 | L | M | 使用现有编号服务的序列逻辑，在事务中执行 |
| #3 级联删除可能导致误删子事项 | L | H | 确认对话框中明确提示"将同时删除 N 个子事项" |

## Success Criteria

- [ ] 状态切换被拒绝时，前端展示具体原因消息（非 tooltip），消息内容来自后端
- [ ] 子事项编辑弹窗包含开始时间字段，修改后能成功保存
- [ ] PM 角色可见删除按钮，点击后弹出确认对话框，确认后主事项及其子事项被软删除
- [ ] 待办→子事项表单的描述字段为 disabled 灰色样式，不可点击
- [ ] 主事项详情页子事项列表按 id 倒序排列
- [ ] 关闭新增/转换表单后再次打开，所有字段为空
- [ ] 转换表单未选负责人时提交按钮禁用，且负责人标签显示必填标记
- [ ] member 角色用户登录后能看到其权限范围内的菜单和功能
- [ ] 子事项可移动到其他主事项，移动后编号自动更新，状态不变
- [ ] 状态过滤器支持多选；负责人过滤器选中某人时，同时展示其负责的子事项所属的主事项

consistency_check_result:
  status: pass
  pairs_checked: 45
  conflicts_found: 0

## Next Steps

- Proceed to `/write-prd` to formalize requirements
