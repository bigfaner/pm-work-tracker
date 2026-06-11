---
feature: milestone-map
extracted: 2026-06-08
---

# Extracted Technical Specifications — milestone-map

## Status: Partially consolidated

Most milestone-map tech specs have been previously extracted. The following CROSS items remain:

### [CROSS] TECH-MS-001: Milestone Performance Targets

**Source**: PRD "Performance Requirements", design "Testing Strategy"

Performance targets for milestone-related pages and APIs:

| Metric | Target | Context |
|--------|--------|---------|
| MilestoneMap list page render | < 300ms | 10 milestone map cards |
| Timeline page initial render | < 500ms | 20 milestones + 200 MainItems |
| Milestone list API response | < 200ms | Standard CRUD API |
| Pagination threshold | 200 MainItems | Enable paged loading by milestone group when exceeded |

**Target file**: `docs/conventions/performance-targets.md` (append to existing)

### [CROSS] TECH-MS-002: Milestone Completion Calculation (Real-Time)

**Source**: PRD "Data Requirements", design "Computed Fields"

Milestone completion is calculated at GET time, not persisted:

- **Milestone completion**: average of all related MainItems' `completion` field (DECIMAL(5,2)). Empty milestone = 0.
- **MilestoneMap overall progress**: average of all related MainItems' `completion` across all milestones.
- **Calculation trigger**: on every GET request (real-time, not cached).
- **Related MI count**: also computed at GET time.

This extends the existing completion calculation pattern in DM-006/DM-008 (SubItem weighted average -> MainItem). The milestone layer adds a simple average on top.

**Target file**: `docs/conventions/data-model.md` (append new section DM-015)

### [CROSS] TECH-MS-003: Entity Hierarchy Extension (Milestone Binding)

**Source**: Design "Architecture", "Data Models"

The milestone feature extends the core entity hierarchy with two new entities and a many-to-one binding:

```
MilestoneMap -> Milestone -> MainItem (via milestone_key FK)
```

- MilestoneMap belongs to Team, contains zero or more Milestones.
- Milestone belongs to one MilestoneMap, binds zero or more MainItems (via `main_items.milestone_key` FK to `milestones.biz_key`).
- MainItem's `milestone_key` is nullable BIGINT with index, no DDL foreign key constraint.
- One MainItem can belong to at most one Milestone (many-to-one from MI perspective).

This extends DM-007 (Core Entity Hierarchy).

**Target file**: `docs/conventions/data-model.md` (update DM-007)

### Already Consolidated (no action needed)

| Spec | Target File | Status |
|------|-------------|--------|
| MilestoneMap 6-state machine | docs/conventions/status-machine.md | Already integrated |
| Milestone 4-state machine | docs/conventions/status-machine.md | Already integrated |
| milestone:* permission codes (4 codes) | docs/conventions/permission-codes.md | Already integrated |
| SD-007: Cascade soft-delete for MilestoneMap | docs/conventions/soft-delete.md | Already integrated |
| SD-008: Milestone soft-delete unbinds items | docs/conventions/soft-delete.md | Already integrated |
| Milestone error codes (8 codes) | docs/conventions/error-codes.md | Already integrated |
