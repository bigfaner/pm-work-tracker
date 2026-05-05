---
scope: global
source: feature/code-conventions BIZ-001
---

# Feature Development Workflow

All features must follow the phase-gate development workflow. Each gate must pass before proceeding to the next.

## Phase-Gate Checklist

1. **PRD** → write requirements → pass `/eval-prd`
2. **Tech Design** → write architecture → pass `/eval-design`
3. **Task Breakdown** → decompose into executable tasks
4. **TDD Implementation** → for each task:
   - Write failing test
   - Implement minimal code to pass
   - Refactor if needed
   - Unit + integration tests pass
5. **E2E Tests** → feature-level end-to-end tests pass
6. **Lint Gate** → zero lint violations
7. **Consolidation** → `/consolidate-specs` to extract reusable knowledge
