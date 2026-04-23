---
feature: "weekly-stats-optimization"
status: tasks
---

# Feature: weekly-stats-optimization

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 将每周进展统计栏从 4 个卡片扩展为 7 个，新增 pending/pausing/overdue 三种状态统计，并为每个卡片添加 hover tooltip 说明统计口径 |
| User Stories | prd/prd-user-stories.md | 4 个用户故事：PM 快速识别风险事项、团队成员理解统计含义、移动端查看说明、键盘无障碍访问 |
| UI Functions | prd/prd-ui-functions.md | 3 个 UI 功能：7 卡片统计栏、统计卡片 Tooltip（含响应式 + 无障碍）、响应式布局 |
| Tech Design | design/tech-design.md | 后端 WeeklyStats DTO 新增 3 字段 + buildWeeklyGroups 计数逻辑扩展；前端 StatsBar 7 列 + Tooltip；isOverdue 参数扩展 |
| API Handbook | design/api-handbook.md | GET /v1/teams/:teamId/views/weekly 响应体 stats 字段扩展说明 |
| UI Design | ui/ui-design.md | StatsBar 7 卡片布局（Tailwind UI 风格）、Tooltip 交互规格（hover/click/keyboard）、响应式断点（xl/md/sm）、无障碍属性定义 |

## Traceability

| PRD Section | Design Section | Tasks |
|-------------|----------------|-------|
| 功能描述 5.1 统计栏 | tech-design.md § Interface 1, 2; ui-design.md § StatsBar | 1.1, 2.1, 3.1 |
| 功能描述 5.2 Tooltip 交互 | tech-design.md § Interface 6; ui-design.md § StatCard Tooltip | 3.2, 4.1 |
| 功能描述 5.3 响应式布局 | tech-design.md § Interface 6; ui-design.md § Responsive Layout | 3.2 |
| 功能描述 5.4 关联性需求改动 | tech-design.md § Interface 4, 5 | 1.2, 2.2 |
