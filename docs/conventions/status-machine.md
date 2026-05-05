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

## Auto-P1 Rule (Aspirational)

PRD specifies: when DelayCount >= 2, upgrade SubItem priority to P1 and set IsKeyItem=true. **Not yet implemented** in service code. The `DelayCount` and `IsKeyItem` fields exist in the model but are not read or incremented by any service method.
