---
feature: "status-flow-optimization"
status: tasks
---

# Feature: status-flow-optimization

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd.md | 统一事项状态枚举体系（code+name），建立 MainItem/SubItem 状态机，实现主子联动、状态变更日志，修复前端 StatusDropdown |
| Tech Design | design/design.md | 新增 pkg/status 枚举包、StatusHistory 审计模型、ChangeStatus 状态机、EvaluateLinkage 联动引擎、per-MainItem 互斥锁 |
| Design Eval | design/design-eval.md | 评估得分 B，Breakdown-Readiness A，22 条 AC 全覆盖 |

## Traceability

| PRD Section | Design Section | Tasks |
|-------------|----------------|-------|
| R1: Status Enumeration | New Package: pkg/status | 1.1, 1.7 |
| R2: Status Transition Rules | Modified Service: MainItemService, SubItemService | 1.4, 1.5 |
| R3: Main-Sub Item Linkage | LinkageResult Struct, EvaluateLinkage | 2.1 |
| R4: RecalcCompletion Coordination | Modified Service: SubItemService | 2.1 |
| R5: Status Change Log | New Model: StatusHistory, New Service: StatusHistoryService | 1.3, 1.4, 1.5, 2.1 |
| R6: Frontend | Frontend Changes | 1.7, 3.1 |
| AC-1 Status codes | pkg/status, VO Changes | 1.1, 1.4, 1.5, 1.7 |
| AC-2 MainItem transitions | MainItem ChangeStatus | 1.4, 1.6 |
| AC-3 SubItem transitions | SubItem ChangeStatus update | 1.5, 1.6 |
| AC-4 Self-transition | ChangeStatus flow step 2 | 1.4, 1.5, 1.6 |
| AC-5 Remove UpdateReq.Status | Modified DTO | 1.2 |
| AC-6 Terminal side effects | ChangeStatus step 6 | 1.4, 1.5 |
| AC-7~AC-12 Linkage | EvaluateLinkage | 2.1 |
| AC-13 RecalcCompletion order | Execution Flow | 2.1 |
| AC-14, AC-15 Status history | StatusHistoryService | 1.3, 1.4, 1.5, 2.1 |
| AC-16~AC-18 Frontend display | StatusBadge, StatusDropdown | 1.7, 3.1 |
| AC-19 Overdue badge | Frontend Changes > Overdue Badge | 1.7, 3.1 |
| AC-20 PM-only reviewing | Security Considerations | 1.4, 3.1 |
| AC-21 Confirmation dialog | Frontend Changes > StatusDropdown | 3.1 |
| AC-23 Available-transitions API | AvailableTransitions methods | 1.4, 1.5 |
