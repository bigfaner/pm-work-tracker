---
id: "5"
title: "迁移 Infra 测试到 tests/infra/"
priority: "P1"
estimated_time: "1h"
complexity: "low"
dependencies: [2]
surface-key: ""
surface-type: ""
breaking: false
type: "coding.refactor"
mainSession: false
---

# 5: 迁移 Infra 测试到 tests/infra/

## Description

将构建检查和静态分析测试从 `tests/e2e/` 迁移到 `tests/infra/`。这些文件使用 `runCli()` 执行 `go build`、`grep`、`lint-staged` 等命令，不是 CLI 功能测试。`-cli` 后缀改为 `-build` 以反映真实功能。

运行器从 Playwright 迁移到 Vitest（与 API 测试相同，但这些测试无 HTTP 调用，仅子进程执行）。

## Reference Files
- `docs/proposals/e2e-test-conventions/proposal.md` — 目标目录结构 (tests/infra/), Infra 文件迁移映射
- `docs/features/e2e-test-conventions/tasks/audit-manifest.md`: 使用 Task 1 产出的文件清单确定哪些文件归入 infra/ (ref: 迁移前审计)

## Acceptance Criteria
- [ ] `tests/infra/` 目录创建，包含 `vitest.config.ts`（testTimeout 15s）和 `package.json`
- [ ] 构建检查文件迁移到 `tests/infra/`，`-cli` 后缀改为 `-build`（如 `bizkey-cli` → `bizkey-build`，`jlc-schema-cli` → `schema-mysql`，`config-yaml-cli` → `config-yaml-build`，`unify-permission-checks-build` → `permission-checks-build`）
- [ ] `npx vitest run` 从 `tests/infra/` 执行，所有迁移后测试通过

## Implementation Notes
- 按 proposal 的 Infra 文件迁移映射操作
- `config-setup.ts`（非 spec）移入 `tests/shared/`
- import 从 `@playwright/test` 改为 `vitest`，`runCli()` 从 `../shared/helpers` 引用
