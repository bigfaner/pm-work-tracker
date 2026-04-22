---
feature: "Code Quality Cleanup"
---

# User Stories: Code Quality Cleanup

## Story 1: Faster List Page Loading

**As an** end user of the PM Work Tracker
**I want to** load list pages (item pool, progress records, table view) quickly even with 100+ records
**So that** I can work efficiently without waiting for data to load

**Acceptance Criteria:**

- Given the item pool has 200+ records, when I open the item pool list page, then data loads with p95 response time under 200ms (or at least 40% faster than before cleanup)
- Given the progress list has 150+ records, when I open the progress list page, then data loads with p95 response time under 200ms (or at least 40% faster than before cleanup)

---

## Story 2: Understandable Codebase for Onboarding

**As a** developer joining the team
**I want to** find only relevant, actively-used code when reading through the codebase
**So that** I can quickly understand the system without being confused by dead code, duplicated logic, or contract mismatches

**Acceptance Criteria:**

- Given I search for a type definition, when I look at the frontend types, then every declared field matches what the backend actually returns
- Given I look at a component file, when I open it, then it contains at most 300 lines focused on one responsibility
- Given I encounter a shared pattern (status dropdown, member select), when I look for it, then I find one shared implementation rather than 3+ copies

---

## Story 3: Maintainable Backend for Feature Development

**As a** backend developer
**I want to** add new features without navigating around duplicated boilerplate and N+1 query traps
**So that** I can ship features faster and with fewer performance bugs

**Acceptance Criteria:**

- Given I need to add a new list endpoint, when I look at existing list endpoints for reference, then they all use O(1) association queries (no per-item DB calls)
- Given I need to create a new repository, when I look at existing repos, then they share a generic `FindByID` pattern with zero duplicated boilerplate
- Given I need to add a new handler, when I look at existing handlers, then they all follow the same constructor pattern (panic-on-nil)

---

## Story 4: Navigable Frontend After Restructuring

**As a** frontend developer on the team
**I want to** find components, hooks, and utilities in dedicated modules with clear import paths
**So that** I can locate and reuse shared code without reading through 1000+ line files

**Acceptance Criteria:**

- Given I need a status transition dropdown, when I import `StatusTransitionDropdown`, then it works identically across ItemViewPage, MainItemDetailPage, and SubItemDetailPage
- Given I need to format a date, when I import `formatDate`, then there is exactly one shared implementation (not 4+ copies)
- Given I open any page component file, when I check its line count, then it does not exceed 300 lines
- Given I need to find a dialog component, when I look in the page directory, then it exists as a separate file (e.g., `CreateMainItemDialog.tsx`) rather than being inline
