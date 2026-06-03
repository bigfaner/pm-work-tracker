---
id: "4"
title: "迁移 Web 测试到 tests/web/"
priority: "P1"
estimated_time: "1.5h"
complexity: "medium"
dependencies: [2]
surface-key: "frontend"
surface-type: "web"
breaking: false
type: "coding.refactor"
mainSession: false
---

# 4: 迁移 Web 测试到 tests/web/

## Description

将 Web E2E 测试从 `tests/e2e/` 移动到 `tests/web/`，保留 Playwright 作为运行器。创建独立的 `playwright.config.ts` 和 `package.json`。

目录按 journey 组织：items/, roles/, teams/, users/, auth/, weekly/, item-pool/, smoke/。跨 surface 测试（如 full-e2e.spec.ts）归 web/smoke/。

## Reference Files
- `docs/proposals/e2e-test-conventions/proposal.md` — 目标目录结构 (tests/web/), Developer Observable Behavior
- `tests/shared/helpers.ts`: Web 测试中非 Playwright 依赖的 helpers 通过 `../shared/helpers` 引用 (ref: Shared helpers 所有权规则)
- `docs/features/e2e-test-conventions/tasks/audit-manifest.md`: 使用 Task 1 产出的文件清单确定哪些文件归入 web/ (ref: 迁移前审计)

## Acceptance Criteria
- [ ] `tests/web/` 目录结构按 proposal 目标创建，包含 `playwright.config.ts` 和 `package.json`
- [ ] Web 测试文件迁移到 `tests/web/<journey>/` 子目录，import 路径更新（shared helpers 通过 `../shared/helpers`，Playwright-specific helpers 保留在 `tests/web/helpers/`）
- [ ] 所有 Web 测试在 `tests/web/` 位置下通过（`npx playwright test` 从 tests/web/ 执行）

## Implementation Notes
- Playwright-specific helpers（如 `login(page)`, `screenshot(page, tcId)`）移入 `tests/web/helpers/`
- 跨 surface 测试（full-e2e.spec.ts）归 `web/smoke/`，因主导 surface 为 web
- 复杂度判定：涉及 ~10 个 spec 文件，均为文件移动 + import 路径调整，属机械性变更，但文件数量较多故定 medium

### Test Impact
- Affected test suite(s): tests/web/
- Expected fixture changes: 无 fixture 变更，仅位置和 import 路径调整
- Risk level: medium
