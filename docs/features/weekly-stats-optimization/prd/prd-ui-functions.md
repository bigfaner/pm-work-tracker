---
feature: "weekly-stats-optimization"
---

# 每周进展统计优化 — UI Functions

> Requirements layer: defines WHAT the UI must do. Not HOW it looks (that's ui-design.md).

## UI Scope

- 每周进展页（`/weekly`）顶部统计栏（StatsBar 组件）

---

## UI Function 1: 7 卡片统计栏

### Description

将现有 4 个统计卡片扩展为 7 个，覆盖所有决策相关状态维度。每个卡片展示一个数字和对应的标签名称。

### User Interaction Flow

1. 用户访问每周进展页，选择周次
2. 系统加载数据，统计栏显示 7 个卡片，每个卡片展示对应状态的子事项数量
3. 数据加载中时，卡片显示加载占位状态

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| activeSubItems | number | `stats.activeSubItems` | 本周活跃总数 |
| newlyCompleted | number | `stats.newlyCompleted` | 本周新完成数 |
| inProgress | number | `stats.inProgress` | 进行中数 |
| blocked | number | `stats.blocked` | 阻塞中数 |
| pending | number | `stats.pending` | 未开始数 |
| pausing | number | `stats.pausing` | 暂停中数 |
| overdue | number | `stats.overdue` | 逾期中数 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 加载中 | 卡片显示骨架屏或 "-" 占位 | 接口请求未返回 |
| 已加载 | 显示实际数字 | 接口返回成功 |
| 空数据 | 数字显示 0 | 当周无对应状态事项 |
| 错误 | 所有卡片数字显示 "-"，统计栏顶部显示"加载失败，请刷新" | 接口返回非 2xx 或请求超时 |

### Validation Rules

- 所有数字为非负整数
- 卡片顺序固定：本周活跃 → 本周新完成 → 进行中 → 阻塞中 → 未开始 → 暂停中 → 逾期中

---

## UI Function 2: 统计卡片 Tooltip

### Description

每个统计卡片支持 hover（桌面端）或点击（移动端）展示 tooltip，说明该卡片的统计口径。

### User Interaction Flow

**桌面端：**
1. 用户将鼠标移入卡片区域
2. 延迟 300ms 后显示 tooltip（防止鼠标划过时频繁弹出）
3. 鼠标移出卡片区域后 tooltip 隐藏

**移动端：**
1. 用户点击卡片
2. tooltip 展开显示
3. 用户再次点击同一卡片，tooltip 收起
4. 用户点击其他区域，tooltip 收起

**键盘：**
1. 用户 Tab 键聚焦卡片
2. tooltip 立即可见（无延迟）
3. 用户 Tab 离开卡片，tooltip 隐藏

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| tooltipText | string | 静态配置 | 每个卡片对应固定的 tooltip 文本，见下表 |

**Tooltip 文本定义（字符串须与此完全一致）：**

| 卡片 | Tooltip 文本 |
|------|-------------|
| 本周活跃 | 本周有进展记录，或计划周期与本周重叠的子事项总数 |
| 本周新完成 | 本周内实际完成（actualEndDate 落在本周）的子事项数 |
| 进行中 | 状态为"进行中"且本周活跃的子事项数 |
| 阻塞中 | 状态为"阻塞中"且本周活跃的子事项数 |
| 未开始 | 已创建但尚未启动（状态为 pending）且本周活跃的子事项数 |
| 暂停中 | 状态为"暂停中"且本周活跃的子事项数 |
| 逾期中 | 计划截止日在本周结束前已过、尚未完成/关闭且本周活跃的子事项数 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 隐藏 | tooltip 不可见 | 默认状态 |
| 显示 | tooltip 浮层可见，内容为对应说明文本 | hover 300ms / 点击 / Tab 聚焦 |

### Validation Rules

- tooltip 位置：默认显示在卡片正上方（placement=top）；若上方空间不足则自动翻转至下方
- 每个卡片元素须携带 `aria-describedby` 属性，其值指向对应 tooltip 内容节点的 `id`
- tooltip 内容节点须有唯一 `id`，供 `aria-describedby` 引用

---

## UI Function 3: 响应式布局

### Description

统计栏根据视口宽度自动调整卡片排列方式，确保在不同设备上均可正常展示。

### User Interaction Flow

用户在不同设备/窗口宽度下访问页面时，统计栏自动适配布局，无需用户操作。

### Data Requirements

无额外数据需求，复用 UI Function 1 的数据。

### States

| State | Display | Trigger |
|-------|---------|---------|
| 宽屏（≥1280px） | 7 列单行 | 视口宽度 ≥ 1280px |
| 中屏（768–1279px） | 4+3 两行（flex-wrap） | 视口宽度 768–1279px |
| 窄屏（<768px） | 2 列多行 | 视口宽度 < 768px |

### Validation Rules

- 任何断点下，卡片内容（数字 + 标签）不得被截断
- 任何断点下，统计栏不得产生横向滚动条
