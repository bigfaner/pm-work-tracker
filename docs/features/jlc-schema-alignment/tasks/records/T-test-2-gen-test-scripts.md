---
status: "completed"
started: "2026-04-25 21:17"
completed: "2026-04-25 21:23"
time_spent: "~6m"
---

# Task Record: T-test-2 生成 e2e 测试脚本

## Summary
从 test-cases.md 生成可执行 TypeScript e2e 测试脚本：ui.spec.ts（1个UI测试）、api.spec.ts（6个API测试）、cli.spec.ts（5个CLI测试），以及 helpers.ts、package.json、tsconfig.json。所有 test() 包含追溯注释，共12条 Traceability 注释。

## Changes

### Files Created
- docs/features/jlc-schema-alignment/testing/scripts/helpers.ts
- docs/features/jlc-schema-alignment/testing/scripts/ui.spec.ts
- docs/features/jlc-schema-alignment/testing/scripts/api.spec.ts
- docs/features/jlc-schema-alignment/testing/scripts/cli.spec.ts
- docs/features/jlc-schema-alignment/testing/scripts/package.json
- docs/features/jlc-schema-alignment/testing/scripts/tsconfig.json

### Files Modified
无

### Key Decisions
- TC-001（UI）为 auth-required-test，顶层 before 调用 loginViaUI(page)
- TC-002~TC-007（API）为 auth-required-test，使用 authCurl（Bearer token）
- TC-008~TC-012（CLI）无需认证，直接使用 runCli()
- TC-007 通过读取前端源码文件验证 bizKey 路径构造，避免需要运行前端服务
- helpers.ts 中 loginViaUI 使用 sitemap 语义化 locator（getByLabel/getByRole）替代正则降级

## Test Results
- **Passed**: -1
- **Failed**: 0
- **Coverage**: N/A (task has no tests)

## Acceptance Criteria
- [x] testing/scripts/package.json 已创建
- [x] testing/scripts/helpers.ts 已创建
- [x] 至少一个 spec 文件已生成（ui.spec.ts / api.spec.ts / cli.spec.ts）
- [x] 每个 test() 包含追溯注释 // Traceability: TC-NNN → {PRD Source}

## Notes
无
