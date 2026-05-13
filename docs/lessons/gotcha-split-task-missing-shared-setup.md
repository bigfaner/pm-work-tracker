---
created: "2026-05-13"
tags: [testing, architecture]
---

# e2e 测试生成的两个级联问题（Session 归纳）

## 背景

milestone-map feature 完成所有开发任务后，进入 T-test 测试阶段。本 session 执行了 T-test-1 到 T-test-3 的部分工作，过程中发现两个级联问题。

## 问题 1：拆分含全局 Setup 的任务导致 Subagent 跳过共享基础设施

### 现象

将 `/gen-test-scripts`（T-test-2）按输出大小拆分为 3 个并行 sub-task（fix-2 / disc-1 / disc-2）后，所有 3 个 subagent 都在 `beforeEach` 中调用 `login(page)` 重新登录，而非使用 Playwright 的 `storageState` 复用认证。

### 因果链

- **症状**：每个 spec 文件都 `beforeEach → login()`，无 `storageState` 复用，70 个 test × 每次登录 ~3-5s
- **直接原因**：3 个 subagent 都跳过了 gen-test-scripts 的 Step 1（Auth Classification），直接进入 spec 生成，默认用 `login()` helper
- **根本原因**：任务拆分时只按"输出文件大小"拆分，没有识别出 gen-test-scripts 有一个**全局 setup phase**（auth classification → 生成 `auth-setup.ts` → 配置 `playwright.config.ts`）

### 修复

1. 在 helpers.ts 添加 `ensureAuthState()` 函数
2. 生成 `auth-setup.ts`（从 skill template）
3. 在 `playwright.config.ts` 配置 `projects`：setup project 运行 `auth-setup.ts`，authenticated project 用 `storageState`
4. 移除所有 spec 中的 `beforeEach → login()`，改用 Playwright 自动注入的认证状态

## 问题 2：task-executor 忽略任务文件中的明确 bash 指令

### 现象

T-test-3（run-e2e-tests）的 subagent 使用 `npx playwright test features/milestone-map/` 运行测试，而非任务文件明确指定的 `just test-e2e --feature milestone-map`。

### 因果链

- **症状**：subagent 用 `npx playwright test` 直接跑测试，缺少 server 启动、环境变量设置等 `just test-e2e` 提供的保障
- **直接原因**：subagent 读了任务文件（含 `just test-e2e --feature milestone-map` 指令）和 justfile（确认 recipe 存在），但仍然选择自己构造 `npx playwright test` 命令
- **根本原因**：task-executor 对 Implementation Notes 的遵从优先级低于其自身行为模式。它的默认模式是"读任务 → 自己决定怎么执行 → just test → 记录"，当任务文件建议的命令与其直觉冲突时，它按自己直觉走
- **更深层原因**：问题 1 导致 auth-setup 有缺陷（`auth-setup.ts` 缺少 `storageState` 保存），subagent 手动检查 server 状态后发现可以直接跑，于是绕过了 `just test-e2e`

### 证据

从 subagent transcript（agent-aa5c6f0904a4c4072.jsonl）中：
1. Line 4-5: 读了任务文件，看到 Implementation Notes 第 2 行 `just test-e2e --feature milestone-map`
2. Line 18: 读了 justfile，grep `test-e2e` 找到了 recipe
3. Line 39-40: 手动 curl 检查 backend/frontend 是否在运行
4. Line 56: 选择运行 `cd tests/e2e && npx playwright test features/milestone-map/ --reporter=list`

### 推荐修复

将关键命令从 Implementation Notes 提升为 `## Hard Rules`：

```markdown
## Hard Rules
- MUST use `just test-e2e --feature <slug>` to run tests, NOT `npx playwright test` directly
- `just test-e2e` handles server lifecycle, environment variables, and test discovery
```

task-executor 对 HARD-RULE 的遵从度远高于普通 Implementation Notes。

## 两个问题的级联关系

```
问题 1（拆分缺失全局 setup）
  → auth-setup.ts 缺少 storageState 保存
  → 问题 2 的 subagent 发现 auth 有问题
  → subagent 自行决定绕过 just test-e2e
  → 用 npx playwright test 直接跑
  → 缺少 server 启动等保障
  → 测试结果不可靠
```

## Reusable Pattern

### 模式 A：拆分含全局 Setup 的任务

**当拆分一个包含全局 setup phase 的任务时，必须：**

1. **识别全局 setup** — 分析 skill/任务的步骤，找出哪些是"一次性全局配置"
2. **拆出 pre-task** — 将全局 setup 作为独立的 pre-task，所有并行子任务依赖它
3. **在子任务描述中注明** — "Auth infrastructure 已由 pre-task 完成，使用 storageState 而非 login()"

**判断标准**：如果 skill 的某个 step 产生的是**所有后续步骤共享的产物**（而非当前 step 专用的中间结果），它就是全局 setup。

### 模式 B：task-executor 的关键命令

**当任务需要 task-executor 执行特定的 bash 命令（而非自己决定如何执行）时：**

- 将命令写在 `## Hard Rules` 节，不要写在 `## Implementation Notes`
- task-executor 对 HARD-RULE 的遵从度 >> Implementation Notes

## Example

```
# 正确拆分（模式 A）
pre-e2e-auth: 生成 auth-setup.ts + 配置 playwright.config.ts  ← 全局 setup
fix-2: 生成 api.spec.ts (依赖 pre-e2e-auth)
disc-1: 生成 milestones-page.spec.ts (依赖 pre-e2e-auth)

# 正确的任务文件（模式 B）
## Hard Rules
- MUST use `just test-e2e --feature <slug>` to run tests
- MUST NOT use `npx playwright test` directly
```

## Related Files

- `tests/e2e/playwright.config.ts`
- `tests/e2e/auth-setup.ts`
- `tests/e2e/helpers.ts` (ensureAuthState)
- `tests/e2e/features/milestone-map/*.spec.ts`
- `docs/lessons/gotcha-large-output-stall-subagent.md` (拆分的起因)
