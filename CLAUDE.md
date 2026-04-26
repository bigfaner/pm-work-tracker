---
paths:
  backend: "backend/"
  frontend: "frontend/"
  docs: "docs/"
  tasks: "docs/features/pm-work-tracker/tasks/"
---

# CLAUDE.md

Behavioral guidelines to reduce common LLM coding mistakes. Merge with project-specific instructions as needed.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

**These guidelines are working if:** fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.

## 5. Key Documents

```
docs/
├── ARCHITECTURE.md              # 分层架构总览
├── DECISIONS.md                 # 技术决策记录
├── lessons/                     # 经验教训（bug 修复、踩坑记录）
│   ├── frontend-test-command.md # 前端测试命令正确用法
│   └── weekly-view-bug-fixes.md # 周视图时区/ISO周计算 bug 修复
└── features/<slug>/             # Feature 工作区
    ├── manifest.md              # Feature 索引 & 可追溯性映射
    ├── prd/                     # 需求文档
    ├── design/                  # 技术设计 & API 文档
    ├── ui/                      # UI 设计规格
    └── tasks/                   # 任务定义 & 执行记录
```

### Running Tests

```bash
# Backend
cd backend && go test ./...

# Frontend — use either, do NOT append --run
npm test          # package.json script already includes --run
npx vitest run    # equivalent
```

### Schema Consistency

`backend/migrations/SQLite-schema.sql` and `backend/migrations/MySql-schema.sql` must always stay in sync. Any DDL change (add table, add column, change type) must be applied to **both** files.
