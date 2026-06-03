---
id: "1"
title: "迁移前审计：分类 tests/e2e/ 文件到 surface"
priority: "P0"
estimated_time: "30m"
dependencies: []
type: "doc"
mainSession: false
---

# 1: 迁移前审计：分类 tests/e2e/ 文件到 surface

## Description

在执行任何目录重组或框架迁移前，审计 `tests/e2e/` 下的所有测试文件，按 proposal 的分类规则判定每个文件的 surface 归属（api / web / infra / shared），并识别需要特殊处理的文件。

分类规则：
- 含浏览器 API（page, browser, locator 等）→ `web/`
- 仅 HTTP 请求（无浏览器）→ `api/`
- 使用 `runCli()` 的构建/静态检查 → `infra/`
- 跨 surface 共享工具 → `shared/`

## Reference Files
- `docs/proposals/e2e-test-conventions/proposal.md` — Problem, Evidence, 目标目录结构, Feasibility Assessment > Resource & Timeline

## Affected Files

### Create
| File | Description |
|------|-------------|
| `docs/features/e2e-test-conventions/tasks/audit-manifest.md` | 文件清单：每行 = 源文件 + 目标 surface + 特殊处理标记 |

### Modify
| File | Changes |
|------|---------|
| — | — |

### Delete
| File | Reason |
|------|--------|
| — | — |

## Acceptance Criteria
- [ ] 产出文件清单（audit-manifest.md），列出 `tests/e2e/` 下所有 .spec.ts 文件的 surface 归属判定（api / web / infra / shared）
- [ ] 标记需要特殊处理的文件：使用 `request.newContext` 的文件、`.serial` 测试、无 surface 后缀需人工判定的文件
- [ ] 清单与 proposal 的目标目录结构和 Infra 文件迁移映射交叉验证，无遗漏

## Implementation Notes
- 重点关注 20 个无 surface 后缀的文件，按分类规则逐个判定
- 已知 3 个文件使用 `request.newContext`，需在清单中标记
- 已知 10 个文件使用 `waitForTimeout()`（共 177 处），标记但不修复（Out of Scope）
