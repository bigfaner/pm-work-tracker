---
id: "T-review-doc"
title: "Review Documentation Quality"
priority: "P1"
estimated_time: "30min"
dependencies: ["1", "8", "7", "9"]
type: "doc.review"
surface-key: ""
surface-type: ""
---

Review documentation quality for the e2e-test-conventions feature (quick mode).

## Acceptance Criteria
- [ ] All doc task deliverables reviewed for completeness, accuracy, and consistency with proposal
- [ ] Convention rule IDs (E-100~E-106, E-110~E-115, E-120~E-123) are unique, sequential, and non-overlapping
- [ ] Anti-pattern lists contain correct counts: api=8, web=7

## Acceptance Criteria Summary

The following acceptance criteria are pre-extracted from doc tasks. Use these as the review baseline.

### 1-migration-audit
- [ ] 产出文件清单（audit-manifest.md），列出 `tests/e2e/` 下所有 .spec.ts 文件的 surface 归属判定（api / web / infra / shared）
- [ ] 标记需要特殊处理的文件：使用 `request.newContext` 的文件、`.serial` 测试、无 surface 后缀需人工判定的文件
- [ ] 清单与 proposal 的目标目录结构和 Infra 文件迁移映射交叉验证，无遗漏


### 7-api-conventions
- [ ] `api/core.md` 包含 7 条规范（E-100~E-106），每条含 ID、规则名、正确示例、错误示例
- [ ] `api/core.md` 包含 8 条反模式清单，每条含反模式描述和替代方案
- [ ] `api/core.md` 的 `domains` frontmatter 已更新（含 testing-api, vitest, http-client 等关键词）


### 8-web-conventions
- [ ] `web/core.md` 包含 6 条规范（E-110~E-115），每条含 ID、规则名、正确示例、错误示例
- [ ] `web/core.md` 包含 7 条反模式清单，每条含反模式描述和替代方案
- [ ] `web/core.md` 的 `domains` frontmatter 已更新（含 testing-web, playwright, e2e 等关键词）


### 9-helpers-conventions
- [ ] `index.md` 包含 4 条 helpers 规范（E-120~E-123），每条含 ID、规则名、说明
- [ ] `index.md` 的 `domains` frontmatter 已更新（含 testing-helpers, config, token 等关键词）


## Discovery Strategy

Scan ONLY the following allowlist of directories for target documents:
- docs/features/e2e-test-conventions/ (prd/, design/, testing/, and any subdirectories)
- docs/proposals/e2e-test-conventions/

EXCLUDE the following from scanning — do NOT read or process these:
- tasks/ directory (task definitions are not deliverables)
- tasks/records/ directory (execution records are not deliverables)
- manifest.md (build artifact)
- index.json (build artifact)

Only .md files under the allowlist directories are target deliverables.
