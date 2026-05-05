---
scope: global
source: feature/code-conventions BIZ-002, TECH-008
---

# Lint Enforcement Standards

Every coding convention must have a corresponding automated lint rule. Conventions without enforcement are guidelines, not standards.

## Backend (golangci-lint)

### BIZ-lint-001: camelCase JSON Tags

**Rule**: `tagliatelle` linter enforces `json: camel` on all struct tags.

**Rationale**: Mixed snake_case/camelCase JSON tags cause frontend mapping bugs. Automated enforcement prevents regression.

### BIZ-lint-002: Duplicate Code Detection

**Rule**: `dupl` linter with threshold 80 detects copy-pasted code blocks.

**Rationale**: Catches when a helper should have been extracted but was inlined instead.

### TECH-lint-001: Configuration Location

```yaml
# backend/.golangci.yml
linters:
  enable:
    - tagliatelle
    - dupl
linters-settings:
  tagliatelle:
    rules:
      json: camel
  dupl:
    threshold: 80
```

## Frontend (ESLint)

### BIZ-lint-003: No Hardcoded Colors

**Rule**: `no-restricted-syntax` blocks Tailwind color classes: `emerald-*`, `red-*`, `amber-*`, `slate-*`.

**Rationale**: Forces use of theme tokens, enabling theme changes via CSS variables.

### TECH-lint-002: Configuration Location

```js
// frontend/eslint.config.js
'no-restricted-syntax': [
  'error',
  { selector: 'Literal[value=/\\b(emerald|red|amber|slate)-\\d/]',
    message: 'Use theme tokens instead of hardcoded colors.' },
]
```
