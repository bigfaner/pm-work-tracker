---
feature: "决策日志"
---

# 决策日志 — UI Functions

> Requirements layer: defines WHAT the UI must do. Not HOW it looks (that's ui-design.md).

## UI Scope

1. 决策 Timeline（主事项详情页内嵌区域）
2. 添加/编辑决策表单（Dialog）

---

## UI Function 1: 决策 Timeline

### Description

在主事项详情页中展示决策记录的时间线列表。已发布决策对所有团队成员可见，草稿仅录入人可见并带「草稿」标记。

### User Interaction Flow

1. 用户打开主事项详情页 → 页面滚动至决策记录区域
2. 系统加载决策列表（已发布 + 当前用户草稿），按时间倒序展示
3. 用户点击某条目 → 展开显示完整内容（替换 80 字符摘要）
4. 用户点击草稿条目的「编辑」按钮 → 打开预填表单

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| bizKey | number | API | 前端用于定位具体记录（编辑草稿用） |
| category | string | API | 预定义分类，显示为标签 Badge |
| tags | string[] | API | 自由标签，显示为 Badge 列表 |
| content | text | API | 摘要截取前 80 字符 + "..."，点击展开完整内容 |
| createdBy | number | API | 录入人 UserKey，需解析为姓名 |
| createTime | datetime | API | 创建时间 |
| status | string | API | "draft" 或 "published" |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Loading | 骨架屏 / "加载中..." | 首次加载、翻页加载 |
| Empty | "暂无决策记录" + 添加按钮（有权限时） | 列表为空 |
| Populated | 时间线列表 | 正常数据 |
| Error | "加载失败，请重试" | API 错误 |

### Validation Rules

- 列表查询时传入当前用户信息，后端按 CreatedBy 过滤草稿
- 每页 20 条，滚动到底部触发加载下一页

---

## UI Function 2: 添加/编辑决策表单

### Description

Dialog 形式的表单，用于新建决策或编辑草稿。支持「保存草稿」和「发布」两个操作。

### User Interaction Flow

1. 用户点击「添加决策」→ 打开空白表单 Dialog
2. 或用户点击草稿的「编辑」→ 打开预填当前草稿内容的表单 Dialog
3. 用户选择分类（下拉）、输入标签（多值）、填写内容（多行文本）
4. 标签输入时，输入框下方显示已有标签提示（recent tags dropdown）
5. 用户点击「保存草稿」→ 保存为草稿，关闭 Dialog，刷新列表
6. 用户点击「发布」→ 发布决策，关闭 Dialog，刷新列表
7. 校验不通过时，在对应字段下方显示错误提示

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| category | string | 预定义枚举 | 下拉选择，6 个选项 |
| tags | string[] | 用户输入 | 回车/逗号分隔，显示已有标签提示 |
| content | text | 用户输入 | 多行文本，≤ 2000 字符 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| New | 空白表单，「添加决策」标题 | 用户点击「添加决策」 |
| Edit | 预填草稿内容，「编辑决策」标题 | 用户点击草稿的「编辑」 |
| Submitting | 按钮显示 loading 状态 | 用户点击保存/发布 |
| Validation Error | 字段下方红色错误提示 | 校验不通过 |

### Validation Rules

| Rule | Field | Condition | Error Message |
|------|-------|-----------|---------------|
| Required | category | 未选择 | 「请选择分类」 |
| Required | content | 为空 | 「请输入决策内容」 |
| Max length | content | > 2000 字符 | 「内容不能超过 2000 字符」 |
| Max length | tags（单项） | > 20 字符 | 「标签不能超过 20 字符」 |
