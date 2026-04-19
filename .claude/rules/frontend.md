---
paths:
  - "frontend/src/**/*.{ts,tsx}"
---

# Frontend Conventions

## Theme Tokens Only

Use CSS custom property theme tokens defined in `src/index.css`. Never use hardcoded Tailwind color classes.

```tsx
// ✅ Correct — theme tokens
<span className="text-success">Active</span>
<Badge className="bg-error text-error-text">P0</Badge>
<div className="text-secondary">Description</div>
<span className="text-warning">Warning</span>

// ❌ Wrong — hardcoded Tailwind colors
<span className="text-emerald-500">Active</span>
<Badge className="bg-red-600 text-red-100">P0</Badge>
<div className="text-slate-400">Description</div>
<span className="text-amber-500">Warning</span>
```

Token mapping:
- `emerald-*` → `success`, `success-text`, `success-bg`
- `red-*` → `error`, `error-text`, `error-bg`
- `amber-*` → `warning`, `warning-text`, `warning-bg`
- `slate-*` → `secondary`, `tertiary`

Enforced by ESLint `no-restricted-syntax` rule.

## API Module Pattern

One file per domain entity in `src/api/`. Uses shared axios `client` instance.

```ts
// ✅ Correct — src/api/mainItems.ts
import { client } from './client';

export const mainItemsApi = {
  list: (teamId: number, params?: MainItemFilter) =>
    client.get<PageResult<MainItem>>(`/api/v1/teams/${teamId}/main-items`, { params }),
  create: (teamId: number, data: MainItemCreateReq) =>
    client.post<MainItem>(`/api/v1/teams/${teamId}/main-items`, data),
};

// ❌ Wrong — inline fetch/axios in component
function ItemList() {
  fetch('/api/v1/teams/1/main-items').then(r => r.json());
}

// ❌ Wrong — multiple API calls in one file
// src/api/all.ts — contains teams + items + progress endpoints
```

Client auto-unwraps `{code: 0, data: ...}` envelope. 401 responses trigger redirect.

## UI Component Pattern (`src/components/ui/`)

Generic primitives with no domain knowledge. Use `forwardRef` + `cn()` for className merging.

```tsx
// ✅ Correct — src/components/ui/badge.tsx
import { forwardRef } from 'react';
import { cn } from '../lib/utils';

interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'default' | 'success' | 'error' | 'warning';
}

const Badge = forwardRef<HTMLSpanElement, BadgeProps>(
  ({ className, variant = 'default', ...props }, ref) => (
    <span
      ref={ref}
      className={cn(baseClasses, variantClasses[variant], className)}
      {...props}
    />
  )
);

// ❌ Wrong — domain-specific logic in UI component
function Badge({ priority }: { priority: string }) {
  const color = priority === 'P0' ? 'red' : 'green';  // no domain logic
}

// ❌ Wrong — hardcoded classes without cn()
<span className="fixed-class-string" ref={ref}>
```

## Shared Component Pattern (`src/components/shared/`)

Domain-aware components that compose UI primitives. Accept focused props.

```tsx
// ✅ Correct — src/components/shared/PriorityBadge.tsx
export function PriorityBadge({ priority }: { priority: string }) {
  const variant = priority === 'P0' ? 'error' : priority === 'P1' ? 'warning' : 'default';
  return <Badge variant={variant}>{priority}</Badge>;
}

// ❌ Wrong — inline badge styling in page component
function ItemRow({ item }) {
  return (
    <span className={item.priority === 'P0' ? 'bg-red-100 text-red-800' : 'bg-gray-100'}>
      {item.priority}
    </span>
  );
}
```

## State Management

Zustand stores in `src/store/`. One file per store, flat structure, no nested reducers.

```ts
// ✅ Correct — src/store/team.ts
import { create } from 'zustand';

interface TeamState {
  selectedTeamId: number | null;
  setSelectedTeam: (id: number) => void;
}

export const useTeamStore = create<TeamState>((set) => ({
  selectedTeamId: null,
  setSelectedTeam: (id) => set({ selectedTeamId: id }),
}));

// ❌ Wrong — React Context for global state
const TeamContext = createContext<TeamState>(...);

// ❌ Wrong — Redux boilerplate
const teamSlice = createSlice({ name: 'team', initialState, reducers: { ... } });
```

## Component Test Co-location

Test files live next to the component they test.

```
src/components/ui/badge.tsx
src/components/ui/badge.test.tsx
src/pages/ItemList.tsx
src/pages/ItemList.test.tsx
```

Use `vitest` + `@testing-library/react`. No Enzyme.
