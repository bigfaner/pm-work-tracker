---
scope: global
source: feature/pm-work-tracker
verified: "2026-05-04"
---

# Status Machine

Source of truth: `backend/internal/pkg/status/status.go`, `transition.go`

## MainItem — 7 States

| Code | Display | Terminal |
|------|---------|----------|
| `pending` | 待开始 | No |
| `progressing` | 进行中 | No |
| `blocking` | 阻塞中 | No |
| `pausing` | 已暂停 | No |
| `reviewing` | 待验收 | No |
| `completed` | 已完成 | Yes |
| `closed` | 已关闭 | Yes |

### MainItem Transition Matrix

| From | To |
|------|----|
| `pending` | `progressing`, `closed` |
| `progressing` | `blocking`, `pausing`, `reviewing`, `closed` |
| `blocking` | `progressing` |
| `pausing` | `progressing`, `closed` |
| `reviewing` | `completed`, `progressing` |

## SubItem — 6 States

| Code | Display | Terminal |
|------|---------|----------|
| `pending` | 待开始 | No |
| `progressing` | 进行中 | No |
| `blocking` | 阻塞中 | No |
| `pausing` | 已暂停 | No |
| `completed` | 已完成 | Yes |
| `closed` | 已关闭 | Yes |

### SubItem Transition Matrix

| From | To |
|------|----|
| `pending` | `progressing`, `closed` |
| `progressing` | `blocking`, `pausing`, `completed`, `closed` |
| `blocking` | `progressing` |
| `pausing` | `progressing`, `closed` |

Any transition not listed above is invalid and returns `INVALID_STATUS` (422).

## ItemPool — 3 States

| Code | Display | Terminal |
|------|---------|----------|
| `pending` | 待分配 | No |
| `assigned` | 已分配 | Yes |
| `rejected` | 已拒绝 | Yes |

`assigned` and `rejected` are terminal. Assignment is atomic (single DB transaction).

## MilestoneMap -- 6 States

_Source: feature/milestone-map_

| Code | Display | Terminal |
|------|---------|----------|
| `planning` | 规划中 | No |
| `reviewed` | 已评审 | No |
| `ready` | 待实施 | No |
| `executing` | 实施中 | No |
| `completed` | 已完成 | Yes |
| `cancelled` | 已取消 | Yes |

### MilestoneMap Transition Matrix

| From | To |
|------|----|
| `planning` | `reviewed`, `cancelled` |
| `reviewed` | `ready`, `planning`, `cancelled` |
| `ready` | `executing`, `reviewed`, `cancelled` |
| `executing` | `completed`, `ready`, `cancelled` |

Terminal states (`completed`, `cancelled`) have no outgoing transitions.

Business rule filter: transition to `completed` is only available when all Milestones in the map are terminal (BR-2).

## Milestone -- 4 States

_Source: feature/milestone-map_

| Code | Display | Terminal |
|------|---------|----------|
| `not_started` | 未开始 | No |
| `in_progress` | 进行中 | No |
| `completed` | 已完成 | Yes |
| `cancelled` | 已取消 | Yes |

### Milestone Transition Matrix

| From | To |
|------|----|
| `not_started` | `in_progress`, `cancelled` |
| `in_progress` | `completed`, `cancelled` |
| `completed` | `cancelled`, `in_progress` |

`cancelled` is terminal with no outgoing transitions.

Business rule filters:
- Transition to `completed` only available when all related MainItems are terminal (BR-1).
- No transitions available when parent MilestoneMap is terminal (BR-5).

## Auto-P1 Rule (Aspirational)

PRD specifies: when DelayCount >= 2, upgrade SubItem priority to P1 and set IsKeyItem=true. **Not yet implemented** in service code. The `DelayCount` and `IsKeyItem` fields exist in the model but are not read or incremented by any service method.
