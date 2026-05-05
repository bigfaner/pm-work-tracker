---
scope: frontend
source: feature/improve-ui BIZ-007
---

# Frontend UX Conventions

## Empty State Convention

All list/dashboard pages must display an empty state message with guidance when no data matches the current filter. Never show a blank page.

Example messages:
- Item list: "暂无事项" + create button
- Team list: "暂无团队" + create button
- Search results: "未找到匹配结果" + suggest adjusting filters

## Layout

Desktop-only, minimum resolution 1280px. No mobile adaptation.
