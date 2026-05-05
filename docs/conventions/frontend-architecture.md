---
scope: frontend
source: feature/improve-ui TECH-004, TECH-005
---

# Frontend Architecture

## TECH-fe-001: shadcn/ui Component Pattern

UI primitives in `src/components/ui/` follow the shadcn/ui pattern:
- Radix UI primitives + Tailwind CSS
- `forwardRef` + `cn()` className merge
- Component source code is copied into the project (not imported as library)

Component directory structure:
- `ui/` — Generic primitives (button, input, select, dialog, badge, etc.)
- `layout/` — Layout components (AppLayout, Sidebar)
- `shared/` — Domain-aware compositions (StatusBadge, PriorityBadge, etc.)

## TECH-fe-002: Dependency Stack

| Purpose | Library |
|---------|---------|
| CSS | Tailwind CSS v4 |
| UI Primitives | Radix UI |
| Class merging | clsx + tailwind-merge (via `cn()`) |
| Variant management | class-variance-authority |
| Icons | lucide-react |
| Routing | react-router-dom |
| Data fetching | @tanstack/react-query |
| State management | zustand |
| HTTP client | axios |
| Date handling | dayjs |
| Charts/Gantt | frappe-gantt (style override with Tailwind) |

Do not add alternative libraries for these purposes without a documented decision in `docs/decisions/`.
