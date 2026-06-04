---
id: "6"
title: "清理迁移残留 + CI 适配"
priority: "P1"
estimated_time: "1h"
complexity: "medium"
dependencies: [3, 4, 5]
surface-key: ""
surface-type: ""
breaking: false
type: "coding.cleanup"
mainSession: false
---

# 6: 清理迁移残留 + CI 适配

## Description

所有 surface 迁移完成后：(1) 删除 re-export shim 和空的 `tests/e2e/` 目录；(2) 适配 CI 按 surface 分独立 job。

CI 适配范围（仅此三项，不做全面重构）：
- (a) 按 surface 分 job（api 启动后端，web 启动全栈，infra 无需服务）
- (b) 更新 glob 匹配新目录路径
- (c) 移除旧的统一 e2e job

## Reference Files
- `docs/proposals/e2e-test-conventions/proposal.md` — Key Risks (CI 中断), Feasibility Assessment > Resource & Timeline
- `tests/e2e/`: 确认目录为空后删除 (ref: 目标目录结构)
- CI 配置文件（GitHub Actions / 其他）: 按 surface 分 job，更新 glob (ref: Key Risks)

## Acceptance Criteria
- [ ] Re-export shim 文件（`tests/e2e/helpers.ts`, `tests/e2e/config.yaml`）已删除
- [ ] `tests/e2e/` 目录已删除（确认为空后）
- [ ] CI 有独立 api/web/infra job，旧统一 e2e job 已移除

## Implementation Notes
- CI 过渡策略：同 PR，(1) 新旧 job 并行 → (2) 逐 surface 更新 glob → (3) 移除旧 job
- 复杂度判定：CI 配置变更需要理解现有 CI 结构，涉及多文件修改，定 medium
