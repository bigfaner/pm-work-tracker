---
feature: "e2e-test-scripts-rebuild"
---

# User Stories: E2E Test Scripts Rebuild & Graduation

## Story 1: 开发者运行完整回归套件

**As a** 开发者
**I want to** 运行 `npm test` 并覆盖所有已完成 feature 的回归测试
**So that** 能在提交代码前发现对已有功能的破坏

**Acceptance Criteria:**
- Given `tests/e2e/` 中所有 11 个 feature 已 graduate
- When 开发者在 `tests/e2e/` 目录运行 `npm test`
- Then 命令以退出码 0 完成，输出每个 spec 文件的通过/跳过/失败计数，且 ≥80% 脚本通过

---

## Story 2: 维护者 graduate 单个 feature 的测试脚本

**As a** 维护者
**I want to** 对一个 feature 重新生成脚本并 graduate 到回归套件
**So that** 该 feature 的测试覆盖从 feature 工作区迁移到持续运行的回归保护中

**Acceptance Criteria:**
- Given feature 有 `test-cases.md` 且尚未 graduate
- When 维护者依次运行 `/gen-test-scripts`、`/run-e2e-tests`（通过率 ≥80%）、`/graduate-tests`
- Then `tests/e2e/.graduated/<slug>` 标记文件存在，graduated 脚本出现在 `tests/e2e/<target>/` 下，且不含对 `testing/scripts/` 的导入

---

## Story 3: 维护者处理通过率不足的 feature

**As a** 维护者
**I want to** 在 graduate 前记录无法修复的失败断言
**So that** 不合格的脚本不会污染回归套件，同时保留问题追踪

**Acceptance Criteria:**
- Given 某 feature 脚本运行后通过率 < 80%
- When 维护者将失败断言记录到 `tests/e2e/KNOWN_FAILURES.md` 并确认原因
- Then graduate 可继续进行，`KNOWN_FAILURES.md` 包含 feature slug、失败描述和负责人，且无未记录的失败断言进入回归套件
