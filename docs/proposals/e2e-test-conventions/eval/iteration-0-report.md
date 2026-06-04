# Iteration 0 Report — Pre-Revision (Freeform Findings)

```yaml
iteration: 0
title: "Pre-Revision (Freeform Findings)"
```

## ATTACK_POINTS

- **[high]** Playwright serial 与 Vitest sequential 在失败传播行为上不兼容，迁移会静默改变测试语义 | quote: "API 测试文件基本只需改 import 和文件名后缀（-api 前缀去掉），curl() 和 helpers 不变" | improvement: 添加 Playwright-to-Vitest API 兼容性矩阵，明确 .serial 行为差异的迁移方案

- **[high]** CI 将在迁移后立即中断，且无过渡计划 | quote: "CI pipeline 调整（目录变更后的 CI 适配另议）" | improvement: 将 CI 过渡计划纳入 in-scope，或在 Key Risks 中标注为已知的即时影响并给出过渡方案

- **[high]** 反模式目录未完整覆盖证据表中已识别的问题模式 | quote: "4 种登录方式（localStorage 注入/表单/API token/cache）" | improvement: 扩展反模式列表，补充缺失的 test data cleanup、混合断言风格等

- **[medium]** shared helpers 边界未定义，可能导致 API 测试传递依赖 Playwright | quote: "共享 helpers 提取到 tests/shared/" | improvement: 定义 shared/helpers.ts 归属规则，将 Playwright-specific helpers 拆分到 tests/web/

- **[medium]** 迁移时间线假设 API 测试中不存在 Playwright fixture 模式 | quote: "目录重组 + API 框架迁移：约 2-3 小时（~49 个测试文件的路径和 import 调整）" | improvement: 增加迁移前审计步骤，调整时间线估算

- **[medium]** 增量迁移期间无过渡状态计划 | quote: "逐个 surface 迁移并验证" | improvement: 定义迁移顺序，使用 re-export shim 保持向后兼容

- **[medium]** 规范编写后缺乏防止 spec drift 的持续机制 | quote: "先完成迁移，再从最终代码提炼规范" | improvement: 添加 drift 检测机制（CI grep 检查或季度审查）

## BORDERLINE_FINDINGS

(none)

## SKIPPED_FINDINGS

- (suggestion) 添加 Playwright-to-Vitest API 兼容性矩阵 — 已由 ATTACK_POINTS 第一条覆盖
- (suggestion) 定义 shared/helpers.ts 归属规则 — 已由 ATTACK_POINTS 第四条覆盖
- (suggestion) 将 CI 过渡计划纳入迁移范围 — 已由 ATTACK_POINTS 第二条覆盖
- (suggestion) 扩展反模式列表 — 已由 ATTACK_POINTS 第三条覆盖
- (suggestion) 定义迁移顺序以保持每个提交后测试可运行 — 已由 ATTACK_POINTS 第六条覆盖
- (suggestion) 添加 spec drift 检测机制 — 已由 ATTACK_POINTS 第七条覆盖

## Classification Audit

- Factual correction: 3 findings (serial behavior, CI breakage, anti-pattern coverage)
- Structural suggestion: 4 findings (helpers boundary, timeline, transition plan, drift mechanism)
- Subjective preference: 6 findings (suggestions, merged into existing attack points)

## Rubric

(all dimensions): N/A
