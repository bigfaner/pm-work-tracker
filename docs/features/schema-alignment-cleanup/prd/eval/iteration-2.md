# PRD Evaluation Report — Iteration 2

**Feature:** schema-alignment-cleanup
**Date:** 2026-04-26
**Documents evaluated:** `prd-spec.md`, `prd-user-stories.md`
**Total Score:** 95 / 100

---

## Previous Attack Points — Resolution Status

| # | Attack | Status | Resolution |
|---|--------|--------|------------|
| 1 | No per-item validation rules in functional spec tables | **Resolved** | All 24 items now have a "验证标准" column with concrete grep/test/compile verification commands |
| 2 | Round 2 (items 3-9) has no user story | **Resolved** | Story 2.5 added for backend developers covering dead code removal |
| 3 | Flow diagram shows sequential execution contradicting text | **Resolved** | Diagram revised with subgraphs, fan-out/fan-in connections, and "items 并行" labels |

---

## Dimension Scores

| Dimension | Score | Max |
|-----------|-------|-----|
| 1. Background & Goals | 19 | 20 |
| 2. Flow Diagrams | 20 | 20 |
| 3. Functional Specs | 18 | 20 |
| 4. User Stories | 20 | 20 |
| 5. Scope Clarity | 18 | 20 |

---

## Dimension 1: Background & Goals (19/20)

### Background three elements (7/7)
All three elements present and specific:
- **原因**: `jlc-schema-alignment` branch with 30+ commits, 2 silent bugs, ~15 code issues, ~8 architecture inconsistencies. Concrete impact statements for each.
- **对象**: "按 4 轮依赖顺序逐个修复 24 个已识别的问题，每个问题独立提交、独立测试"
- **人员**: Three user types with specific pain points: PM (broken features), backend devs (deprecated code, duplicate patterns), frontend devs (type inconsistencies, redundant conversions).

### Goals quantified (7/7)
All 6 goals have numeric targets: "2 个 bug 修复", "4 个 deprecated DTO、1 个 unused scope、3 处 dead code", "2 个重复接口、1 个重复分页、1 个重复 userToDTO、5 个重复状态记录调用点", "22 处冗余 String() 包装", specific table renames, specific scope alignment. No vague goals.

### Logical consistency (5/6)
**Deduction (-1):** The goals table claims "删除...1 个 unused scope" under "消除废弃代码", but no functional spec item maps to removing an unused scope. Round 2 items 3-9 are accounted for by "4 个 deprecated DTO" (item 3) and "3 处 dead code" (items 4-6), but the "1 个 unused scope" has no corresponding item. Item 22 (NotDeleted scope alignment) is in Round 4 under "架构调整" and is about consistent usage, not removal. Items 7-9 (GORM tags, console.error, test data) are also unaccounted for in the goals count. This creates a traceability gap between stated goals and deliverables.

---

## Dimension 2: Flow Diagrams (20/20)

### Mermaid diagram exists (7/7)
Flowchart present with proper Mermaid `flowchart TD` syntax.

### Main path complete (7/7)
Start to End covers all 4 rounds with test gates. Previous iteration's visual contradiction (sequential arrows vs parallel text) is now resolved — subgraphs use fan-out/fan-in connections with "items 并行" labels, correctly representing parallel execution within rounds.

### Decision points + error branches (6/6)
Four decision diamonds with four error branches that loop back. Complete coverage.

---

## Dimension 3: Functional Specs (18/20)

### Tables complete (7/7)
All 24 items across 4 rounds. All tables now include a "验证标准" column, providing consistent structure across rounds despite slightly different descriptive columns (问题/修复内容 vs 目标/文件 vs 目标/合并来源 — justified by each round's different nature).

### Field descriptions clear (6/7)
Most items are well-described with file locations and verification standards. **Deduction (-1):** Two items still require code inspection to fully understand the change:
- **Item 12** ("提取共享 `resolveBizKey` helper" from "7 个 handler 中的重复 resolve/parse 函数"): What does `resolveBizKey` do? What is the input/output signature? The PRD does not explain the helper's contract.
- **Item 15** ("提取状态记录 helper" from "5 个重复调用点"): What does "状态记录" mean in this context? What does the helper accept and return? Without this, the reader must inspect all 5 call sites.

### Validation rules explicit (5/6)
The new "验证标准" column is a major improvement — 18 of 24 items now have concrete grep commands or test commands as pass/fail criteria. **Deduction (-1):** Six items have weak or subjective verification:
- **Item 14**: "userToDTO 仅在共享文件中定义一次" — no grep to verify single definition
- **Item 15**: "状态记录调用点使用共享 helper" — no specific grep or test
- **Item 16**: "仅保留真正必要的调用（如 enum 转换）" — "truly necessary" is subjective, not a pass/fail criterion
- **Item 20**: "构造函数签名统一，无重复初始化逻辑" — subjective description, no grep/test
- **Item 21**: "单条和批量转换调用同一函数" — no verification command
- **Item 24**: "formatDate 仅在共享 utils 中定义一次" — no grep to verify count

These items rely on "tests pass" rather than concrete structural verification, which is insufficient when the test suite may not cover the specific refactoring concern.

---

## Dimension 4: User Stories (20/20)

### Coverage (7/7)
Three user types from background. PM gets Stories 1-2, backend dev gets Stories 2.5, 3, and 5, frontend dev gets Story 4. All types covered. Story 2.5 resolves the previous Round 2 gap.

### Format correct (7/7)
All 6 stories follow "As a / I want to / So that" format with concrete, actionable wants. No vague language.

### AC per story (6/6)
All stories have Given/When/Then acceptance criteria. Stories include grep-based verification commands as Then conditions where appropriate.

---

## Dimension 5: Scope Clarity (18/20)

### In-scope items concrete (7/7)
All 6 in-scope items are specific: 4 rounds with item number ranges, plus 2 process requirements. No vague areas.

### Out-of-scope lists deferred items (7/7)
Four deferred items named with specifics: performance optimization, file splitting, dialog component merging, and new features.

### Scope consistent with specs and stories (4/6)
**Deduction (-2):** Two consistency gaps remain:

1. **Goals-to-items mapping gap:** The goals table says "1 个 unused scope" but no functional spec item corresponds. Additionally, Round 2 items 7-9 (GORM tags, console.error, test data) are not accounted for in any goal row, creating an orphaned-work problem.

2. **Round 4 story gap:** Items 20-24 (ViewService constructor merge, single/batch VO merge, NotDeleted scope, TableRow type, formatDate utility) have limited or no user story traceability. Item 23 could fall under Story 4 (frontend type consistency), but items 20, 21, 22, and 24 have no story coverage. Story 3's "I want to 代码中每种模式只有一个实现" is broad enough to loosely cover some, but these items are in Round 4 (Architecture Alignment), not Round 3 (Pattern Unification), and no story explicitly mentions them.

---

## Top 3 Attack Points

### Attack 1: Functional Specs — 6 items with subjective or missing verification
**Severity:** 1 point lost. Items 14, 15, 16, 20, 21, and 24 have verification standards that cannot be objectively evaluated as pass/fail. Example: Item 16 says "仅保留真正必要的调用（如 enum 转换）" — "truly necessary" is a judgment call, not a testable criterion. Example: Item 14 says "userToDTO 仅在共享文件中定义一次" but provides no grep command to verify single definition (e.g., `grep -rn "func userToDTO" backend/internal/ | wc -l` should equal 1). Each of these 6 items needs a concrete structural verification command.

### Attack 2: Scope Clarity — Goals table has orphaned and phantom items
**Severity:** 2 points lost. The goals table claims "1 个 unused scope" under "消除废弃代码" but no functional spec item maps to this. Conversely, Round 2 items 7-9 (GORM tag removal, console.error replacement, test data fix) are not accounted for in any goal row. The goals table must reconcile with the functional spec: either remove the "unused scope" claim, or identify which item it maps to and adjust the goal description.

### Attack 3: Scope Clarity — 4 Round 4 items lack user story traceability
**Severity:** 2 points lost. Items 20 (ViewService constructor), 21 (single/batch VO merge), 22 (NotDeleted scope), and 24 (formatDate utility) have no corresponding user story. A stakeholder reviewing by user stories would miss 17% of the work. Adding a Story 6 for backend architecture consistency (covering items 20-22) and extending Story 4 to cover item 23 would close this gap.

---

## Verdict

**95/100** — Above the 90-point target. All three previous attack points were addressed: validation rules added to all 24 items, Story 2.5 added for Round 2, and flow diagram corrected for parallel execution. The remaining gaps are (1) 6 items with subjective verification, (2) goals-to-items mapping inconsistencies, and (3) 4 Round 4 items without story coverage. These are secondary issues that do not block implementation.
