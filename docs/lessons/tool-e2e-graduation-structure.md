# E2E Test Graduation Structure: Feature Slug vs Functional Domain

## Problem

Graduated e2e tests in `tests/e2e/` preserve the original feature slug as subdirectories under each target type (ui/cli/api), instead of being reclassified by functional domain as the forge spec intends.

Current structure:
```
tests/e2e/<target>/<feature-slug>/   # e.g., api/rbac-permissions/api.spec.ts
```

Expected structure (forge intent):
```
tests/e2e/<target>/<domain>/         # e.g., api/permissions/permissions.spec.ts
```

Additionally, a stray file exists at `tests/e2e/tests/e2e/sub-item-edit.spec.ts` that was never properly graduated.

## Root Cause

**Both spec ambiguity and agent shortcut:**

1. **Forge spec is ambiguous**: The graduate-tests skill says destination is `tests/e2e/<target>/` with optional deeper `<domain>/` for splitting multi-domain specs. But "domain" is never clearly defined or contrasted with "feature slug". The examples (`login`, `dashboard`, `profile`) happen to look like feature slugs, so the agent treats feature slug as domain.

2. **Agent took the path of least resistance**: Instead of reading each spec, understanding the actual functional domain, and merging/splitting accordingly, the agent simply preserved the original feature slug directory under each target. This defeats the purpose of graduation — feature boundaries should be removed and tests reclassified by cross-cutting domain.

3. **No validation in graduation**: The skill has no structural check to verify that the output directory truly reflects functional domain classification vs feature slug preservation.

Causal chain: spec says `<target>/<domain>/` without defining "domain" → agent interprets "domain" as "feature slug" → graduation produces `api/rbac-permissions/` instead of `api/permissions/` → feature silos are preserved, not dissolved.

## Solution

### Spec fix (forge graduate-tests skill)
Add explicit guidance:
- Define "domain" as a functional area (e.g., `auth`, `permissions`, `items`, `users`), NOT a feature slug
- Provide a decision test: "If two different features test the same API endpoints or UI pages, they should merge into the same domain directory"
- Add a structural validation step that flags output dirs matching known feature slugs

### Agent behavior fix
- When graduating, the agent must analyze spec content to identify the actual functional domain, not reuse the feature slug
- When a target already has a domain directory for the same functional area, merge into it rather than creating a new directory

### Cleanup
- Remove `tests/e2e/tests/e2e/sub-item-edit.spec.ts` (or properly graduate it)
- Consider restructuring existing graduated tests to use domain-based directories

## Key Takeaway

When a spec says "reclassify by X", the "X" must be explicitly defined with examples of what it IS and what it is NOT. Without contrast examples ("domain is NOT feature slug"), agents will take the path of least resistance and preserve the existing structure rather than truly reclassifying.

**How to apply**: In any skill that involves reclassification or restructuring, include both positive examples (what the output should look like) and negative examples (what it should NOT look like, especially when the wrong answer is the easy/lazy one).

## Forge Strengthening Recommendations

Audit of forge's 18 skills reveals three existing constraint tags (`HARD-GATE`, `HARD-RULE`, `EXTREMELY-IMPORTANT`) but no output contracts, anti-patterns sections, or shared validation framework. Validation is ad-hoc per skill. Only `learn-lesson` has a "Common Mistakes" section.

### P0: Add universal HARD-GATE — never modify safety nets (1 line)

Applies to all skills that produce files:

```markdown
<HARD-GATE>
Forbidden: modifying .gitignore, lint config, CI config, or any checking/safety
mechanism to accommodate newly created files. If a config blocks your file,
question your file path — do not modify the config.
</HARD-GATE>
```

Prevents the pattern documented in `gotcha-agent-breaks-safety-net.md`.

### P1: Add Output Contract to graduate-tests (~20 lines)

Each skill that produces files should define three things:

```markdown
## Output Contract

### Positive examples
tests/e2e/ui/permissions/permissions.spec.ts     # domain = "permissions"
tests/e2e/api/auth/auth.spec.ts                   # domain = "auth"

### Anti-patterns (MUST NOT produce)
tests/e2e/api/rbac-permissions/api.spec.ts        # ← feature slug, not domain
tests/e2e/ui/improve-ui/ui.spec.ts                # ← feature slug, not domain

### Self-check
After graduation, run: ls tests/e2e/*/  and verify no directory name
matches a known feature slug from .graduated/
```

### P2: Add Anti-Patterns sections to key skills (~20 lines each)

`gen-test-scripts` and `gen-test-cases` should also include anti-patterns tables:

| Pattern | Why it's wrong | What to do instead |
|---------|---------------|-------------------|
| Using feature slug as domain dir | Preserves feature silos | Analyze spec content for functional domain |
| Each feature gets its own helpers.ts | Duplicates shared utilities | Import from top-level helpers.ts |
| Nested tests/e2e/tests/ | Wrong directory nesting | Files go under tests/e2e/\<target\>/ only |

### P3: Shared validation framework (larger effort, defer)

Currently validation is ad-hoc per skill. A shared mechanism (e.g., a `just validate-e2e-structure` target or a shared reference file with validation patterns) would provide systematic protection. This requires more design work and can be iterated later.

### Priority Summary

| Priority | Change | Effort | Impact |
|----------|--------|--------|--------|
| P0 | Universal HARD-GATE: never modify safety nets | 1 line | Prevents most dangerous agent behavior |
| P1 | graduate-tests Output Contract + Anti-patterns | ~20 lines | Fixes structural misclassification |
| P2 | gen-test-scripts Output Contract + Anti-patterns | ~20 lines | Preventive |
| P3 | Shared validation framework | Larger | Systematic long-term improvement |
