# Design System: PM Work Tracker

> Extracted from: http://localhost:5173 (source code analysis)
> Date: 2026-04-26
> Based on: Custom (Tailwind CSS v4 + Radix UI + CVA)

## Visual Theme & Atmosphere

A compact, information-dense project management interface built on a blue-primary palette with slate neutrals. The design prioritizes data density over whitespace — 13px body text, tight spacing, and dense table layouts keep work items visible at a glance. White card surfaces sit on a cool-gray page background, separated by subtle borders rather than shadows. The overall feel is clean, functional, and professional — closer to a developer tool than a consumer app.

## Color Palette

| Role | Value | Usage |
|------|-------|-------|
| Background | `#f8fafc` | Page background, table headers, hover states |
| Surface | `#ffffff` | Cards, sidebar, inputs, dialogs |
| Border | `#e2e8f0` | Card borders, dividers, table borders |
| Border Dark | `#cbd5e1` | Input borders, secondary button borders |
| Text Primary | `#0f172a` | Headings, titles, primary body text |
| Text Secondary | `#475569` | Descriptions, table cells, dropdown items |
| Text Tertiary | `#94a3b8` | Placeholders, subtle labels, table headers |
| Accent | `#2563eb` | Primary buttons, active tabs, sidebar brand |
| Accent Light | `#3b82f6` | Progress fill, today marker, links |
| Accent Hover | `#1d4ed8` | Primary button hover, active nav text |
| Accent BG | `#eff6ff` | Active nav background, avatar fallback, light fills |
| Accent Ring | `#bfdbfe` | Focus ring color |
| Success | `#3b82f6` | Success/completed status (blue, not green) |
| Success BG | `#eff6ff` | Success status background |
| Success Text | `#1d4ed8` | Success status text |
| Warning | `#d97706` | Warning/on-hold status |
| Warning BG | `#fffbeb` | Warning status background |
| Warning Text | `#92400e` | Warning status text |
| Error | `#dc2626` | Error/blocked status |
| Error BG | `#fef2f2` | Error status background |
| Error Text | `#991b1b` | Error status text |

## Typography

| Role | Font | Weight | Size | Usage |
|------|------|--------|------|-------|
| Page Title | System stack | 600 | 18px | Sidebar heading, dialog title |
| Login Title | System stack | 600 | 20px | Login page title |
| Section Heading | System stack | 500 | 14px | Form section labels |
| Body / Form Label | System stack | 500 | 14px | Form labels, general text |
| Component Text | System stack | 500 | 13px | Buttons, inputs, table cells, dropdown items |
| Table Header | System stack | 500 | 12px | Column headers (uppercase tracking-wider) |
| Badge / Caption | System stack | 500 | 12px | Status badges, pagination labels |
| Tooltip | System stack | 400 | 12px | Tooltip text (white on dark bg) |

System font stack: `-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif`

## Components

### Buttons

- **Primary**: bg `#2563eb`, text white, rounded-lg (8px), h-10 px-4, hover bg `#1d4ed8`, focus ring `#3b82f6`
- **Secondary**: bg white, text primary, border `#cbd5e1`, rounded-lg, hover bg-alt
- **Warning**: transparent bg, border `warning-text/40`, text warning-text
- **Danger**: transparent bg, border `error-text/40`, text error-text
- **Ghost**: transparent bg, text secondary, hover bg-alt
- **Icon**: 36x36px square, rounded-lg
- **Sizes**: sm (h-8 px-3 text-xs) / default (h-10 px-4) / lg (h-11 px-6)
- **Transition**: 150ms all

### Cards

- Background: `#ffffff`
- Border: 1px solid `#e2e8f0`
- Border Radius: 12px (rounded-xl)
- Shadow: `0 1px 2px 0 rgb(0 0 0 / 0.05)`
- Header: flex between, p-4 px-5, border-b
- Content: p-5
- Footer: flex, p-3 px-5, border-t, bg `#f8fafc`

### Inputs

- Background: `#ffffff`
- Border: 1px solid `#cbd5e1`
- Border Radius: 6px (rounded-md)
- Height: 40px (h-10)
- Padding: px-3 py-2
- Text: 13px, color primary
- Placeholder: color tertiary
- Focus: border `#3b82f6`, ring-2 ring `#bfdbfe`
- Shadow: `0 1px 2px 0 rgb(0 0 0 / 0.05)`
- Disabled: opacity 50%, cursor not-allowed

### Navigation (Sidebar)

- Layout: Fixed left sidebar, 240px wide
- Background: white, border-r
- Brand area: px-4 py-5, border-b
- Nav items: px-2 py-3 area, gap-2.5 between items
- Active item: bg `#eff6ff`, text `#1d4ed8`, font-medium
- Inactive item: text secondary, hover bg `#f8fafc`
- User section: px-4 py-3, border-t, avatar + name + logout

### Tables

- Wrapper: relative, w-full, overflow-auto
- Header: bg `#f8fafc`, cells px-4 py-2.5, text-xs uppercase tracking-wider, color tertiary
- Body cells: px-4 py-3, text 13px, color secondary
- Rows: border-b, hover bg `#f8fafc`

### Dialogs

- Overlay: fixed inset-0, bg black/50, z-50, fade animation
- Content: fixed centered, z-50, rounded-xl (12px), bg white, shadow-lg, zoom+fade animation
- Sizes: sm (400px), md (480px), lg (560px)
- Header: p-5 pb-0, title text-lg font-semibold
- Body: p-5
- Footer: flex justify-end, gap-2, px-5 pb-5

### Badges

- Base: inline-flex, rounded-full, px-2.5 py-0.5, text-xs font-medium
- Status variants map directly to semantic colors (success/warning/error + bg/text)
- Priority: high (error), medium (warning), low (default)

## Layout

- Sidebar width: 240px (fixed)
- Main content: ml-240px, padding 24px (p-6), bg `#f8fafc`
- Dialog sizes: sm=400px, md=480px, lg=560px
- Grid: Single-column flow, no explicit grid columns
- Section padding: 24px desktop
- Component spacing: 16-20px (mb-4 to mb-5) between form groups

## Depth & Elevation

| Level | Shadow | Usage |
|-------|--------|-------|
| 0 | none | Sidebar, nav items |
| 1 | `0 1px 2px 0 rgb(0 0 0 / 0.05)` | Cards, inputs, selects |
| 2 | `0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)` | Toasts, tooltips, dropdowns |
| 3 | `0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)` | Dialogs, popovers |

## Do's and Don'ts

| Do | Don't |
|----|-------|
| Use theme tokens (`text-primary`, `bg-bg-alt`, `border-border`) | Use hardcoded Tailwind color classes (`text-slate-500`, `bg-blue-100`) |
| Use `cn()` for conditional class merging | Use string concatenation for classes |
| Use CVA variants for component style variations | Use inline ternaries for multi-property variants |
| Use semantic tokens (`success`, `warning`, `error`) for status colors | Use `emerald-*`, `red-*`, `amber-*` directly |
| Use 13px (`text-[13px]`) for dense UI elements | Mix 13px and 14px in the same context |
| Use `rounded-lg` for buttons, `rounded-xl` for cards | Use border-radius values not in the scale |

## Responsive Behavior

| Breakpoint | Behavior |
|-----------|---------|
| Mobile (<768px) | Not optimized — desktop-first tool |
| Desktop (>1024px) | Fixed 240px sidebar + fluid main content |

## Signature Patterns

1. **Blue = Success**: Completed/done states use blue (`#3b82f6`), not green — matching the primary brand color
2. **13px Dense Text**: Body text at 13px instead of 14px, optimized for information density in tables and forms
3. **Border-Delineated Cards**: White cards on light-gray background with borders (not shadows) as the primary separation mechanism
4. **Status Badge System**: Every work status maps to a distinct badge variant with matched bg/text semantic colors
5. **ESLint-Enforced Theme**: Custom ESLint rule bans hardcoded Tailwind color classes, forcing all colors through theme tokens
