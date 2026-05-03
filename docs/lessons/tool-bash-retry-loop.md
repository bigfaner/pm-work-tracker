# Bash 工具重试循环陷阱

## Problem

在 Bash 工具中执行命令失败后，反复重试相同命令而不改变策略。导致连续 10+ 次执行同一失败命令，浪费时间且无进展。

具体案例：运行 `node --import tsx/esm --test cli/e2e-test-scripts-rebuild/cli.spec.ts` 需要从 `tests/e2e/` 目录执行（tsx 只安装在该目录），但 CWD 始终是项目根目录，导致 `node --test` 找不到文件。

## Root Cause

Claude Code 的 Bash 工具每次调用都会重置工作目录到项目根目录。`cd` 不会在调用之间保持。

**因果链**:
1. 症状：命令返回 "Could not find" 错误
2. 直接原因：路径 `cli/e2e-test-scripts-rebuild/cli.spec.ts` 相对于根目录解析，而不是 `tests/e2e/`
3. 深层原因：需要 `cd tests/e2e && node ...` 组合命令，但未使用
4. 行为原因：遇到重复失败后没有停下来分析错误，而是无脑重试

## Solution

**用组合命令代替分开的 cd 和执行**:

```bash
# ✅ 正确 — 单次调用中 cd + 执行
cd tests/e2e && node --import tsx/esm --test cli/e2e-test-scripts-rebuild/cli.spec.ts

# ❌ 错误 — cd 不会跨调用保持
cd tests/e2e
node --import tsx/esm --test cli/e2e-test-scripts-rebuild/cli.spec.ts
```

**重试规则**:
- 同一命令连续失败 **2 次** 后必须停止，分析错误原因
- 改变策略（换路径、换参数、换方法）后再尝试
- 绝不无变化地重试第 3 次

## Key Takeaway

1. Bash 工具的 CWD 每次重置 — 需要 `cd X && cmd` 组合命令
2. 同一命令失败 2 次 = 需要改变策略的信号，不是继续重试的信号
3. 失败时先读错误信息、分析原因、再决定下一步，而非盲目重试
