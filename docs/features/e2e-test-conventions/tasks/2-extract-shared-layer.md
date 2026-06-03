---
id: "2"
title: "提取 shared 层到 tests/shared/"
priority: "P0"
estimated_time: "1h"
complexity: "low"
dependencies: [1]
surface-key: ""
surface-type: ""
breaking: false
type: "coding.refactor"
mainSession: false
---

# 2: 提取 shared 层到 tests/shared/

## Description

将 `tests/e2e/helpers.ts` 和 `tests/e2e/config.yaml` 移动到 `tests/shared/`，作为跨 surface 的共享工具层。旧路径保留 re-export shim 以确保迁移期间其他文件仍能引用。

Shared helpers 禁止 import `@playwright/test` 或任何浏览器 API，仅允许纯 Node.js 代码。

## Reference Files
- `docs/proposals/e2e-test-conventions/proposal.md` — 目标目录结构 (tests/shared/), Shared helpers 所有权规则, Feasibility Assessment > Resource & Timeline
- `tests/e2e/helpers.ts`: 提取到 tests/shared/helpers.ts，确认无 Playwright 依赖 (ref: Shared helpers 所有权规则)
- `tests/e2e/config.yaml`: 移入 tests/shared/config.yaml (ref: 目标目录结构)

## Acceptance Criteria
- [ ] `tests/shared/helpers.ts` 创建完成，内容来自原 `tests/e2e/helpers.ts`，`grep '@playwright/test' tests/shared/` 输出为空
- [ ] `tests/shared/config.yaml` 从 `tests/e2e/config.yaml` 移入
- [ ] 旧路径 `tests/e2e/helpers.ts` 和 `tests/e2e/config.yaml` 保留 re-export shim（`export * from '../shared/helpers'`），确保其他文件引用不断裂

## Implementation Notes
- 如 `helpers.ts` 中包含 Playwright-specific 函数（如 `login(page)`），这些函数留在原处（后续 Web 迁移时移入 `tests/web/helpers/`），shared 只提取纯 Node 函数
- shim 是临时措施，Task 6 会清理
