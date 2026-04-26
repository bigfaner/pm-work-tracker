# 不要为顺序性工作启动 Subagent

## Problem

执行 `/gen-sitemap` 时，主 session 已完成登录、读取路由、拍摄快照等准备工作，
却在进入"逐页探索"阶段时启动了一个 subagent 来完成剩余工作。
结果：subagent 没有主 session 的 agent-browser 登录状态，需要重新认证，且增加了不必要的上下文切换。

## Root Cause

混淆了 subagent 的适用场景：

- **适合 subagent**：并行的独立任务、需要隔离上下文的大量搜索结果
- **不适合 subagent**：主 session 已经具备所有上下文、工具状态（如 agent-browser session）、且任务是顺序执行的

本质是"把自己能做的事外包出去"，而不是"把真正需要并行或隔离的事分出去"。

## Solution

在主 session 中直接继续执行，逐页调用 agent-browser：

```bash
npx agent-browser open http://localhost:5173/weekly
npx agent-browser wait --load networkidle
npx agent-browser snapshot -i
# 提取页面元素，写入 sitemap
```

## Key Takeaway

**启动 subagent 前问自己：**
1. 这个任务能在当前 session 直接完成吗？→ 如果能，就直接做
2. 当前 session 有 subagent 没有的状态（登录态、已读文件、已建上下文）吗？→ 如果有，不要用 subagent
3. 任务是顺序依赖的吗？→ 顺序任务不需要 subagent，并行独立任务才需要

Subagent 是为了**并行**和**隔离**，不是为了"把剩下的活甩出去"。
