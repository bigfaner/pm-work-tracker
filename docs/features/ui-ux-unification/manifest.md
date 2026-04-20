---
feature: "ui-ux-unification"
status: tasks
---

# Feature: ui-ux-unification

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 前端 UI/UX 统一改造：菜单重排、表格按钮图标化、WeekPicker 组件（Safari 兼容+快速切换）、详情页布局统一、链接高亮、追加进度按钮位置调整、每周进展进度数字、按钮文案统一，共 8 个改造点 |
| User Stories | prd/prd-user-stories.md | 6 个用户故事，覆盖 Safari 用户、PM、团队成员、管理员的核心使用场景 |
| UI Functions | prd/prd-ui-functions.md | 8 个 UI 功能点的交互流程、数据需求和状态定义 |
| Tech Design | design/tech-design.md | 纯前端改造，新建 WeekPicker 组件和 weekUtils.ts，涉及 16 个文件，无后端变更 |

## Traceability

| PRD Section | Design Section | Tasks |
|-------------|----------------|-------|
| F1 导航菜单重排 | F1 — Sidebar 菜单重排 | [2.1](tasks/2.1-sidebar-reorder.md) |
| F2 表格操作按钮统一 | F2 — 表格操作按钮加图标 | [2.2](tasks/2.2-table-button-icons.md) |
| F3 WeekPicker 组件 | F3 — WeekPicker 组件（新建） | [1.1](tasks/1.1-week-utils.md), [1.2](tasks/1.2-week-picker.md), [2.3](tasks/2.3-replace-week-picker.md) |
| F4 详情页信息布局统一 | F4 — 详情页信息布局统一 | [2.4](tasks/2.4-detail-layout.md) |
| F5 可跳转文字链接高亮 | F5 — 可跳转文字链接高亮 | [2.5](tasks/2.5-link-highlight.md) |
| F6 追加进度按钮位置 | F6 — 「追加进度」按钮位置调整 | [2.6](tasks/2.6-move-append-btn.md) |
| F7 每周进展子事项进度显示 | F7 — 每周进展子事项进度显示 | [2.7](tasks/2.7-weekly-completion.md) |
| F8 按钮文案统一 | F8 — 按钮文案统一 | [2.8](tasks/2.8-rename-button.md) |
| — | Testing Strategy | [3.1](tasks/3.1-e2e-update.md) |

