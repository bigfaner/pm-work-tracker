---
feature: "code-quality-cleanup"
status: tasks
---

# Feature: code-quality-cleanup

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 4-phase progressive code quality cleanup: dead code removal, N+1 fix, dedup, file splits |
| User Stories | prd/prd-user-stories.md | 3 stories: faster list loads, onboarding clarity, maintainable backend |
| Tech Design | design/tech-design.md | Hybrid SQL strategy for N+1, generic pkg/repo helpers, shared frontend components, file split plan |
| API Handbook | design/api-handbook.md | Contract fixes: SubItemVO +statusName, remove delayCount, align TeamMemberResp/AdminTeam types |

## Traceability

| PRD Section | Design Section | Tasks |
|-------------|----------------|-------|
| Phase 1: Dead Code & Contracts (prd-spec §Phase 1) | Architecture → Phase 1 diagram, API Handbook → contract changes | 1.1 Backend dead code, 1.2 Frontend dead code, 1.3 SubItemVO +statusName, 1.4 Frontend contracts |
| Phase 2: N+1 Queries & Performance (prd-spec §Phase 2) | Interfaces → new repo methods, Data Models → new DTOs, Testing → benchmarks | 2.1 Batch repo methods, 2.2 Fix N+1 view_service, 2.3 Fix N+1 handlers, 2.4 SQL pushdown admin/team, 2.5 SQL pushdown gantt, 2.6 Frontend perf, 2.7 Benchmarks |
| Phase 3: Shared Components & Dedup (prd-spec §Phase 3) | Interfaces → pkg/repo + shared components, Data Models → filter helpers | 3.1 pkg/repo helpers, 3.2 Refactor repos, 3.3 Shared filter/util, 3.4 StatusTransitionDropdown, 3.5 Hooks/utils, 3.6 Replace inline dropdowns, 3.7 MemberSelect |
| Phase 4: File Splits & Readability (prd-spec §Phase 4) | Architecture → Phase 4 diagram | 4.1 Decompose WeeklyComparison, 4.2 Standardize handlers, 4.3 Move inline DTOs, 4.4 Split ItemViewPage, 4.5 Split MainItemDetailPage |
