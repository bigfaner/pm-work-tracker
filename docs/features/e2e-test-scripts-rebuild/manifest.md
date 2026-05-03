---
feature: "e2e-test-scripts-rebuild"
status: tasks
---

# Feature: e2e-test-scripts-rebuild

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 为 11 个有 test-cases.md 但未 graduate 的 feature 重新生成符合 forge 规范的 e2e 脚本，并 graduate 到 tests/e2e/ 回归套件 |
| User Stories | prd/prd-user-stories.md | 开发者运行完整回归套件、维护者 graduate 单个 feature、维护者处理通过率不足的 feature |
| Tech Design | design/tech-design.md | 纯工具链工作流：/gen-test-scripts → /run-e2e-tests → /graduate-tests，逐 feature 执行，含脚本合规契约、package.json 更新规范、KNOWN_FAILURES.md 格式 |

## Traceability

| PRD Section | Design Section | UI Component | Tasks |
|-------------|----------------|--------------|-------|
| 脚本生成规范 | Interfaces §2 Graduated Spec File Contract | — | 1.1, 2.1, 2.2, 2.3, 2.4 |
| Graduate 规范 | Interface §3 package.json Updater + Model §1 Graduation Marker | — | 1.2, 3.1 |
| 已知失败处理 | Error Handling ERR_PASS_RATE_LOW + Convention §2 KNOWN_FAILURES.md | — | 2.1, 2.2, 2.3, 2.4 |
| 流程说明 | Architecture Component Diagram | — | 2.1, 2.2, 2.3, 2.4, 2.gate |
| Story 1: npm test 覆盖全部 feature | Interface §3 package.json Updater | — | 3.1, 3.gate |
| Story 2: graduate 单个 feature | Interface §1 executeFeature | — | 2.1, 2.2, 2.3, 2.4 |
| Story 3: 处理通过率不足 | Error Handling ERR_PASS_RATE_LOW | — | 2.1, 2.2, 2.3, 2.4 |
