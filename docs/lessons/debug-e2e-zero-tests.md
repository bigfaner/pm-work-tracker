# node:test e2e 测试 "0 tests" 问题排查指南

## Problem

使用 node:test 编写的 e2e 测试脚本，执行后报告 "0 tests, 0 passed, 0 failed"。
经过 3 轮修复才解决，每轮修了不同的问题。

## Root Cause

三个独立问题叠加导致：

### 1. before 钩子异常导致整组取消

`node:test` 的 `before()` 钩子如果抛异常，会**取消该 suite 下所有子测试**，
报告为 0 tests 而非 1 failure。这是 node:test 的设计决策——before 失败意味着
测试环境未就绪，后续测试无意义。

**触发条件**：`before` 中调用 `fetch()` 获取 auth token，后端未启动时抛
`TypeError: fetch failed` (ECONNREFUSED)。

### 2. tsx 不激活 node:test runner

`tsx file.ts` 只做 TypeScript 转译+执行，**不会激活 node 的 `--test` 模式**。
`node:test` 的 `describe`/`test`/`before` 函数虽然能 import，但不会被注册到
test runner，导致 "0 tests"。

**正确用法**：`node --import tsx --test file.ts`

### 3. 缺少 "test" npm script

package.json 有 `test:api` 和 `test:all` 但没有 `"test"`。`npm test` 查找
精确的 `"test"` key，找不到时执行 npm 默认行为（直接退出，无输出）。

**hook 调用路径**：stop hook → `npm test` → 查找 scripts.test → 未找到 → 退出

## Solution

三处修复：
1. `before` 钩子中 wrap auth call 在 try/catch 中，失败时 fallback 到未认证模式
2. package.json scripts 使用 `node --import tsx --test api.spec.ts`
3. 确保 package.json 有 `"test"` script 入口

## Key Takeaway

生成 e2e 测试脚本时，必须遵循以下检查清单：

1. **npm script 入口**：`package.json` 必须有 `"test"` 脚本（不是 `test:xxx`）
2. **node:test 激活**：必须用 `node --test`（或 `--import tsx --test`），不能只用 `tsx`
3. **before/after 钩子容错**：`before()` 中的网络调用必须 try/catch，失败不应取消所有测试
4. **配置缺失降级**：config.yaml 等配置文件不存在时，应返回默认值而非 throw

### 调试技巧

当遇到 "0 tests" 时，按以下顺序排查：
```
npm test 有 "test" script? → node --test 激活? → before 钩子报错? → config 找不到?
```
