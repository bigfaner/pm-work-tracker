---
title: "Milestone State Constraints"
domains: [milestone, status, transition, cascade, terminal, delete, unbind]
---

# Milestone State Constraints

Business rules governing the three-layer state model: MilestoneMap -> Milestone -> MainItem. Rules control when state transitions are allowed, how cascading works, and what constraints apply to cross-layer operations.

Source: feature/milestone-map

## Completion Preconditions

### BIZ-milestone-001: Milestone Completion Requires All Related Items Terminal

A Milestone can only be transitioned to `completed` when ALL related MainItems are in terminal status (`completed` or `closed`). Transitioning to `cancelled` auto-unbinds all related MainItems regardless of their status. Empty milestones (no related MainItems) can freely transition to `completed`.

**Enforcement**: Service layer, before status update. Returns `400 BAD_REQUEST` ("里程碑下存在未完成的事项") if non-terminal items exist.

**Side effect on cancelled**: When a Milestone transitions to `cancelled`, all related MainItems have their `milestone_key` set to NULL within the same database transaction.

**Source**: feature/milestone-map BR-1

### BIZ-milestone-002: MilestoneMap Completion Requires All Milestones Terminal

A MilestoneMap can only be transitioned to `completed` when ALL its Milestones are in terminal status (`completed` or `cancelled`). Transitioning to `cancelled` cascades: all non-terminal Milestones are set to `cancelled` (triggering BR-001 auto-unbind for each), and all related MainItems are unbound. Empty MilestoneMaps (no Milestones) can freely transition to `completed`.

**Enforcement**: Service layer, before status update. Returns `400 BAD_REQUEST` ("里程碑图下存在未完成的里程碑") if non-terminal milestones exist.

**Cascade on cancelled**: Within a single transaction:
1. Set all non-terminal Milestones to `cancelled`
2. Clear `milestone_key` on all MainItems associated with those Milestones

**Source**: feature/milestone-map BR-2

## Cross-Layer Move Constraints

### BIZ-milestone-003: MainItem Milestone Assignment Constraints

When updating a MainItem's `milestone_key` (bind/rebind/unbind), the following constraints apply:

1. **Terminal MainItem**: If the MainItem is in terminal status (`completed`/`closed`), its `milestone_key` cannot be changed. Returns `400 BAD_REQUEST`.
2. **Terminal Milestone**: If the target Milestone is in terminal status (`cancelled`), the binding is rejected. Returns `400 BAD_REQUEST`.
3. **Validation order**: Check MainItem status first, then check target Milestone status.

**Enforcement point**: `MainItemService.Update`, when `req.MilestoneKey` is provided. Requires `MilestoneRepo` dependency injection in MainItemService.

**Source**: feature/milestone-map BR-3

## Delete Constraints

### BIZ-milestone-004: Milestone and MilestoneMap Delete Constraints

**MilestoneMap deletion**: Only MilestoneMaps in `planning`, `reviewed`, or `ready` status can be deleted. MilestoneMaps in `executing`, `completed`, or `cancelled` status reject deletion. Returns `400 INVALID_PARAMS`.

On deletion: all Milestones within the map are soft-deleted, and all related MainItems' `milestone_key` is set to NULL within a single transaction. This cascading delete bypasses individual Milestone deletion constraints (BR-005).

**Milestone deletion**: Only Milestones in `not_started` or `cancelled` status can be deleted. Milestones in `in_progress` or `completed` reject deletion. Returns `400 INVALID_PARAMS`.

On deletion: all related MainItems' `milestone_key` is set to NULL within the same transaction.

**Source**: feature/milestone-map BR-4

## Parent-Terminal Downward Constraints

### BIZ-milestone-005: Parent-Terminal Blocks Child Operations

When a parent entity is in terminal status, child entities cannot change state or be created:

1. **Map terminal -> no Milestone create/update**: When `MilestoneMap.map_status` is `completed` or `cancelled`, creating new Milestones or changing Milestone status is rejected. Returns `400 MAP_IS_TERMINAL`.
2. **Map terminal -> no MI milestone_key change**: MainItems under Milestones of a terminal Map cannot change their `milestone_key` (stacks with BR-003).
3. **Milestone terminal -> no MI milestone_key change**: When `Milestone.milestone_status` is `completed` or `cancelled`, related MainItems cannot change their `milestone_key`. Returns `400 BAD_REQUEST`.

**Source**: feature/milestone-map BR-5

## Cascade Cancel

### BIZ-milestone-006: MilestoneMap Cancel Cascades to All Children

When a MilestoneMap transitions to `cancelled`:
1. All non-terminal Milestones are set to `cancelled` (triggers BR-001 auto-unbind per Milestone)
2. All MainItem `milestone_key` values for those Milestones are cleared

This is executed within a single database transaction to ensure atomicity.

**Source**: feature/milestone-map BR-6
