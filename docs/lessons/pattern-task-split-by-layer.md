---
category: pattern
triggers: ["task breakdown", "backend API implementation", "/breakdown-tasks"]
---

# Pattern: Split Backend API Tasks by Layer

When a single backend task spans Repo + Service + Handler + Routes for an entity with 7+ endpoints, agents stall during extended thinking blocks — they accumulate too much context (30+ file reads) and fail to produce tool calls within the watchdog window (600s).

## Rule

Split each entity's full-stack task into 3 subtasks:

1. **Infrastructure** — Repo interface, GORM implementation, helpers/extensions, route registration skeleton. ~1h.
2. **Basic CRUD** — Create + List + Get + Update endpoints (simple operations). ~1h.
3. **Complex Operations** — Delete (cascade/transaction), ChangeStatus (side effects), AvailableTransitions, computed fields. ~1.5h.

## Why

Forensic analysis (2026-05-12, task 2.1) showed the agent read 37 files in 3.3min, wrote only the repo interface, then stalled for >10min trying to plan 4 layers simultaneously. The thinking blocks grew progressively: 5s → 12s → 15s → indefinite.

Per-layer splitting keeps each task's implementation surface small enough for the agent to plan in a single thinking pass and iterate quickly.

## When to Apply

- Task estimated time > 2h
- Task covers both simple CRUD and complex operations (cascades, transactions, computed fields)
- Entity has 5+ endpoints
