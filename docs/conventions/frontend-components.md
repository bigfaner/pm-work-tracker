---
scope: frontend
source: feature/code-conventions TECH-005–007
verified: "2026-05-04"
---

# Frontend Component Conventions

Reusable component patterns and theme token usage rules.

## TECH-frontend-001: Textarea Component

**Location**: `src/components/ui/textarea.tsx`

**Pattern**: `forwardRef<HTMLTextAreaElement>` + `cn()` className merge. Mirrors `input.tsx`.

**Base classes**: `flex w-full rounded-md border border-border-dark bg-white px-3 py-2 text-[13px] text-primary shadow-sm placeholder:text-tertiary focus:outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-200 min-h-[120px] resize-y`

**Usage**:
```tsx
import { Textarea } from '@/components/ui/textarea';

<Textarea placeholder="Enter description..." value={text} onChange={handleChange} />
```

## TECH-frontend-002: PrioritySelectItems Component

**Location**: `src/components/shared/PrioritySelectItems.tsx`

**Pattern**: Renders exactly 3 `<SelectItem>` options: P1, P2, P3. No props. Caller provides `<Select>`, `<SelectTrigger>`, `<SelectContent>`.

**Usage**:
```tsx
<Select value={priority} onValueChange={setPriority}>
  <SelectTrigger><SelectValue placeholder="Priority" /></SelectTrigger>
  <SelectContent>
    <PrioritySelectItems />
  </SelectContent>
</Select>
```

## TECH-frontend-003: Color Token Mapping

**Rule**: All UI colors use CSS custom property theme tokens defined in `src/index.css`. Hardcoded Tailwind color classes are prohibited.

**Mapping**:

| Hardcoded          | Theme Token Family    |
|--------------------|-----------------------|
| `emerald-*`        | `success`, `success-text`, `success-bg` |
| `red-*`            | `error`, `error-text`, `error-bg`       |
| `amber-*`          | `warning`, `warning-text`, `warning-bg` |
| `slate-*`          | `secondary`, `tertiary`                 |

**Theme token value change**: `--color-success` series changed from green to light blue:
- `--color-success`: `#3b82f6` (primary-500)
- `--color-success-bg`: `#eff6ff` (primary-50)
- `--color-success-text`: `#1d4ed8` (primary-700)
