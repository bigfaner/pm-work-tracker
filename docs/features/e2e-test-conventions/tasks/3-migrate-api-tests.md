---
id: "3"
title: "迁移 API 测试到 tests/api/ + Vitest"
priority: "P0"
estimated_time: "2h"
complexity: "high"
dependencies: [2]
surface-key: "backend"
surface-type: "api"
breaking: false
type: "coding.refactor"
mainSession: false
---

# 3: 迁移 API 测试到 tests/api/ + Vitest

## Description

将 API 功能测试从 `tests/e2e/` 迁移到 `tests/api/`，运行器从 `@playwright/test` 切换到 `vitest`。这是本 proposal 的核心迁移工作。

迁移内容：
- 目录按 journey 组织：items/, roles/, teams/, users/, item-pool/, progress/, smoke/
- 所有文件 import 从 `@playwright/test` 改为 `vitest`
- 3 个使用 `request.newContext` 的文件改写为 `curl()`
- `.serial` 测试改为 `describe.sequential()`，检查失败跳过语义
- 创建 `vitest.config.ts`：`testTimeout: 30s`, `hookTimeout: 60s`, `sequence: { sequential: true }`

## Reference Files
- `docs/proposals/e2e-test-conventions/proposal.md` — 目标目录结构 (tests/api/), Playwright → Vitest API 兼容性矩阵, 具体规范清单 > API Functional Test
- `tests/shared/helpers.ts`: API 测试通过相对路径引用 shared helpers（如 `import { curl } from '../shared/helpers'`）(ref: Shared helpers 所有权规则)
- `docs/features/e2e-test-conventions/tasks/audit-manifest.md`: 使用 Task 1 产出的文件清单确定哪些文件归入 api/ (ref: 迁移前审计)

## Acceptance Criteria
- [ ] `tests/api/` 目录结构按 proposal 目标创建，包含 `vitest.config.ts`（testTimeout 30s, hookTimeout 60s, sequential）和 `package.json`
- [ ] API 测试文件全部迁移到 `tests/api/<journey>/` 子目录，所有 import 从 `@playwright/test` 改为 `vitest`
- [ ] 使用 `request.newContext` 的文件已改写为 `curl()`，共享 helpers 通过 `../shared/helpers` 引用
- [ ] `.serial` 测试已迁移为 `describe.sequential()`，失败传播语义已检查（Vitest `.sequential` 不自动跳过后续 tests）
- [ ] `npx vitest run` 从 `tests/api/` 执行，所有迁移后测试通过

## Implementation Notes
- 参考兼容性矩阵：`.serial` → `.sequential()` 失败传播行为不同
- `test.setTimeout(ms)` → `test('name', { timeout: ms }, fn)`
- 确认所有 API 测试文件无浏览器 API 依赖后再迁移

### Test Impact
- Affected test suite(s): tests/api/
- Expected fixture changes: 无 fixture 文件变更，仅运行器和 import 变更
- Risk level: high（~14 spec 文件，涉及 import/运行器/HTTP 客户端变更）
