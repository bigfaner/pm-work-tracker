---
feature: "code-conventions"
generated: "2026-05-04"
status: draft
---

# Business Rules: Code Conventions

## Development Process

### BIZ-001: Feature Development Workflow

**Rule**: All features must follow the phase-gate development workflow: PRD → eval → tech design → eval → task breakdown → TDD implementation → e2e tests → consolidation.

**Context**: AI-assisted development without a structured workflow produces inconsistent quality. A mandatory checklist ensures no phase is skipped and each gate is validated before proceeding.

**Scope**: [CROSS]

**Source**: prd/prd-spec.md Section 5.6

Checklist:
1. PRD completed + passed `/eval-prd`
2. Tech design completed + passed `/eval-design`
3. Task breakdown completed
4. Each task follows TDD: write failing test → implement → refactor → tests pass
5. All tasks completed: feature-level e2e tests pass, lint clean, code committed
6. Post-completion: `/consolidate-specs` to extract reusable knowledge

### BIZ-002: Convention Enforcement via Lint

**Rule**: Every coding convention must have a corresponding automated lint rule for regression prevention. Conventions without enforcement are guidelines, not standards.

**Context**: Manually maintained conventions drift over time, especially with AI-generated code. Lint rules catch violations automatically and serve as executable documentation.

**Scope**: [CROSS]

**Source**: prd/prd-spec.md Section 5.5

Examples:
- `tagliatelle` enforces camelCase JSON tags
- `no-restricted-syntax` enforces theme tokens over hardcoded colors
- `dupl` detects code duplication above threshold

### BIZ-003: Code Deduplication Principle

**Rule**: When a code pattern appears 3+ times across files, extract it into a shared helper function or component. Each distinct helper/component should exist as exactly one copy.

**Context**: AI-generated code tends to repeat patterns inline rather than reusing shared code. The 3-copy threshold catches this before it becomes unmanageable.

**Scope**: [CROSS]

**Source**: prd/prd-spec.md Section 5.3

Thresholds applied in this feature:
- mapNotFound: 5 copies → 1 generic function
- Pagination defaults: 6 inline blocks → 1 helper
- Date parsing: 11 inline calls → 1 helper
- Textarea styling: 14 copies → 1 component
- PrioritySelect options: 21 copies → 1 component
