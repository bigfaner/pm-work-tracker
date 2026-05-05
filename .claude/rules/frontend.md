---
paths:
  - "frontend/src/**/*.{ts,tsx}"
---

# Frontend Conventions

Detailed conventions in `docs/conventions/frontend-architecture.md`, `docs/conventions/frontend-components.md`, `docs/conventions/frontend-ux.md`.

Key rules enforced by linters:

- Theme tokens only (never hardcoded Tailwind colors: emerald/red/amber/slate)
- One API module per domain entity in `src/api/`, uses shared `client` instance
- UI components: `forwardRef` + `cn()`, no domain logic in `ui/`
- State management: zustand stores in `src/store/`, no Context/Redux
- Test files co-located, vitest + @testing-library/react
