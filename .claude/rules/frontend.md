---
paths:
  - "frontend/src/**/*.{ts,tsx}"
---

# Frontend Critical Rules

Detailed conventions: `docs/conventions/frontend-architecture.md`, `docs/conventions/frontend-components.md`.

- Theme tokens only — never hardcoded Tailwind colors (emerald/red/amber/slate)
- UI components: `forwardRef` + `cn()`, no domain logic in `ui/`
- State management: zustand stores in `src/store/`, no Context/Redux
- One API module per domain entity in `src/api/`, uses shared `client` instance
