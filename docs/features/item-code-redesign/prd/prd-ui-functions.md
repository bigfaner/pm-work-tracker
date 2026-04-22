---
feature: "事项编码体系重新设计"
---

# 事项编码体系重新设计 — UI Functions

> Requirements layer: defines WHAT the UI must do. Not HOW it looks (that's ui-design.md).

## UI Scope

- TeamManagementPage — 创建团队对话框（新增 Code 输入框）
- TeamManagementPage — 团队列表（新增 Code 列）
- ItemViewPage — 主事项列表（编码 Badge 值变更）
- TableViewPage — 主事项表格行（编码展示值变更）
- WeeklyViewPage — 周视图行首（编码值变更）
- MainItemDetailPage — 标题旁 Badge（编码值变更）+ 子事项表格（编码来源变更）
- SubItemDetailPage — 子事项编码展示（来源变更）
- ItemPoolPage — 关联主事项编码展示（值变更）

---

## UI Function 1: 创建团队对话框 — Code 输入框

### Description

在现有创建团队对话框中，在 Name 和 Description 字段之后新增一个 "团队编码" 输入框，供 PM 在创建团队时设置唯一的团队缩写。

### User Interaction Flow

1. PM 点击 "创建团队" 按钮，对话框打开
2. PM 填写 Name、Description，然后在 Code 输入框中输入 2~6 位英文字母
3. 输入框失焦时，前端校验格式（正则 `^[A-Za-z]{2,6}$`）
4. 格式不合法 → 输入框下方显示 "编码须为 2~6 位英文字母"
5. PM 点击提交 → 后端校验唯一性
6. Code 重复 → 输入框下方显示 "该编码已被使用"
7. 校验通过 → 团队创建成功，对话框关闭，团队列表刷新

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| code | string | 用户输入 | 2~6 位英文字母，提交前 trim |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 空 | 占位符 "如 FEAT、CORE" | 初始状态 |
| 输入中 | 正常输入框样式 | 用户输入 |
| 格式错误 | 红色边框 + 下方错误文字 | 失焦且格式不合法 |
| 重复错误 | 红色边框 + 下方错误文字 | 提交后后端返回唯一性冲突 |
| 合法 | 正常样式 | 格式校验通过 |

### Validation Rules

| 序号 | 条件 | 触发 | 提示语 |
|------|------|------|--------|
| 1 | 为空 | 提交 | 团队编码为必填项 |
| 2 | 长度 < 2 或 > 6，或含非字母字符 | 失焦 / 提交 | 编码须为 2~6 位英文字母 |
| 3 | 与已有 Code 重复 | 提交（后端返回） | 该编码已被使用 |

---

## UI Function 2: 团队列表 — Code 列

### Description

在 TeamManagementPage 的团队列表中新增一列，展示每个团队的 Code 值。

### User Interaction Flow

1. 用户进入 TeamManagementPage
2. 团队列表加载，每行显示 Code 列
3. 新建团队成功后，列表刷新，新团队行的 Code 列显示刚设置的值

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| code | string | Team.Code 字段 | 只读展示，不可编辑 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 加载中 | 与现有列表加载态一致 | 页面初始化 |
| 已加载 | 每行显示 Code 值（如 `FEAT`） | 数据返回 |

---

## UI Function 3: 主事项编码展示（各页面）

### Description

ItemViewPage、TableViewPage、WeeklyViewPage、MainItemDetailPage、ItemPoolPage 中的主事项编码展示，值从 `MI-NNNN` 变更为 `{TEAM_CODE}-NNNNN`。组件和布局不变，仅数据源值变更。

### User Interaction Flow

无交互变更，纯展示。用户在搜索框输入 `FEAT-` 可过滤该团队事项，输入完整编码精确匹配。

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| code | string | MainItem.Code 字段 | 格式 `{TEAM_CODE}-NNNNN` |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 正常 | Badge / span 展示编码值 | 数据加载完成 |

---

## UI Function 4: 子事项编码展示（MainItemDetailPage + SubItemDetailPage）

### Description

MainItemDetailPage 子事项表格中，每行展示子事项编码（如 `FEAT-00001-01`），来源从运行时拼接改为读取 SubItem.Code 字段。SubItemDetailPage 同理。

### User Interaction Flow

无交互变更，纯展示。

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| code | string | SubItem.Code 字段 | 格式 `{TEAM_CODE}-NNNNN-NN`；原运行时拼接逻辑删除 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 正常 | Badge 展示子事项编码 | 数据加载完成 |
| 编码为空 | 不应出现（迁移后所有 SubItem 均有 Code） | — |
