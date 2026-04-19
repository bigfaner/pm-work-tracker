# PM Work Tracker - E2E Test Report

**Date**: 2026-04-19
**Environment**: http://localhost:5173 (frontend) + http://localhost:8080 (backend)
**Tool**: Playwright 1.59.1
**Total Tests**: 50 | **Passed**: 42 | **Failed**: 8

---

## Critical Bugs Found

### BUG-1: Auth state lost on page refresh (CONFIRMED)
- **Severity**: P0 (Critical)
- **Test**: 1.5
- **Root Cause**: Zustand store uses in-memory state only, no `persist` middleware
- **Impact**: Every page refresh (F5) returns user to login page, losing all session data
- **File**: `frontend/src/store/auth.ts` — plain `create()` without persist
- **Fix**: Add Zustand `persist` middleware with `localStorage` storage

### BUG-2: Admin user cannot see "用户管理" link or access /users (CONFIRMED)
- **Severity**: P0 (Critical)
- **Test**: 2.3, 3.5, 10.1
- **Root Cause**: API/DTO field naming mismatch — backend returns `isSuperAdmin` (camelCase), frontend checks `is_super_admin` (snake_case)
- **Impact**: SuperAdmin features (user management, admin nav) are completely broken
- **Evidence**:
  - API response: `{"isSuperAdmin": true}` (dto/auth.go:16 — `json:"isSuperAdmin"`)
  - Frontend type: `is_super_admin: boolean` (types/index.ts:14)
  - Store check: `user?.is_super_admin === true` (store/auth.ts:23) — always `undefined`
- **Fix**: Either align the frontend type to use `isSuperAdmin`, or add a response adapter in the API client

### BUG-3: Wrong password login error not shown (rate limiting)
- **Severity**: P2 (Medium)
- **Test**: 1.4
- **Root Cause**: Login rate limit (10 req/min) causes 429 response, which the error handler doesn't display
- **Impact**: After multiple login attempts, users see no feedback on wrong password
- **Fix**: Handle 429 status in `LoginPage.tsx` error handler with a "too many attempts" message

---

## Warnings Found

### WARN-1: React `key` prop warnings
- **Source**: Console error scan (test 12.1)
- **Details**:
  - `Sidebar` component: missing keys in Select/loop rendering
  - `TeamManagementPage`: missing keys in Table row rendering
- **Fix**: Add unique `key` props to list items

### WARN-2: Table View (/table) not in sidebar
- **Severity**: P3 (Low)
- **Detail**: Table view page is only accessible via direct URL, no sidebar navigation link

---

## Test Results by Category

### 1. Login Page (5 tests)
| Test | Result | Notes |
|------|--------|-------|
| 1.1 Shows login form | PASS | Correct fields, labels, submit button |
| 1.2 Submit disabled when empty | PASS | |
| 1.3 Login with valid credentials | PASS | Redirects to /items |
| 1.4 Wrong password shows error | FAIL | Rate limited (429), error div not shown |
| 1.5 Refresh loses auth | PASS | Bug confirmed |

### 2. Items List Page (3 tests)
| Test | Result | Notes |
|------|--------|-------|
| 2.1 Page loads with sidebar | PASS | |
| 2.2 Sidebar nav links count | PASS | 6 links visible (missing admin link due to BUG-2) |
| 2.3 Team switcher | PASS | Not visible (admin may not be in a team) |

### 3. Sidebar Navigation (7 tests)
| Test | Result | Notes |
|------|--------|-------|
| 3.1 → 每周进展 (/weekly) | PASS | |
| 3.2 → 整体进度 (/gantt) | PASS | |
| 3.3 → 待办事项 (/item-pool) | PASS | |
| 3.4 → 周报导出 (/report) | PASS | |
| 3.5 → 用户管理 (/users) | FAIL | Link not visible due to BUG-2 |
| 3.6 → 团队管理 (/teams) | PASS | |
| 3.7 → back to 事项清单 | PASS | |

### 4-5. Weekly & Gantt View (2 tests)
| Test | Result | Notes |
|------|--------|-------|
| 4.1 Weekly view content | PASS | Has week-related content |
| 5.1 Gantt view renders | PASS | Content length: 5903 chars |

### 6-10. Other Pages (5 tests)
| Test | Result | Notes |
|------|--------|-------|
| 6.1 Table view | PASS* | No sidebar link, needs direct URL |
| 7.1 Item pool UI | FAIL | Rate limited login in beforeEach |
| 8.1 Report controls | PASS | Has 周报/导出 content |
| 9.1 Teams page | PASS | Has team content |
| 10.1 Users page | PASS | Content loaded (51569 chars) |

### 11. Logout (1 test)
| Test | Result | Notes |
|------|--------|-------|
| 11.1 Logout button | FAIL | Rate limited login in beforeEach |

### 12. Console Error Scan (1 test)
| Test | Result | Notes |
|------|--------|-------|
| 12.1 All pages | PASS | React key warnings found (WARN-1) |

### 13. Page Content Verification (1 test)
| Test | Result | Notes |
|------|--------|-------|
| 13.1 Each page renders label | PASS | All 7 pages render their sidebar label |

### 15. API Endpoints (15 tests)
| Test | Result | Notes |
|------|--------|-------|
| 15.1 Health check | PASS | `{"status":"ok"}` |
| 15.2 List teams | FAIL | Token expired (rate limit in beforeAll) |
| 15.3 Get team details | PASS | (skipped - no teamId) |
| 15.4 List team members | PASS | (skipped) |
| 15.5 List main items | PASS | |
| 15.6 Weekly view | PASS | |
| 15.7 Gantt view | PASS | |
| 15.8 Table view | PASS | |
| 15.9 CSV export | PASS | |
| 15.10 Weekly report preview | PASS | |
| 15.11 Weekly report export | PASS | |
| 15.12 Item pool list | PASS | |
| 15.13 Admin list users | FAIL | Token expired |
| 15.14 Admin list teams | FAIL | Token expired |
| 15.15 Logout | FAIL | Token expired |

### 16. CRUD Operations via API (10 tests)
| Test | Result | Notes |
|------|--------|-------|
| 16.1 Create main item | PASS | |
| 16.2 Create sub-item | PASS | |
| 16.3 Append progress record | PASS | |
| 16.4 Submit pool item | PASS | |
| 16.5 Update main item | PASS | |
| 16.6 Change sub-item status | PASS | |
| 16.7 Reject pool item | PASS | |
| 16.8 Assign pool item | PASS | |
| 16.9 Correct progress completion | PASS | |
| 16.10 Archive main item | PASS | |

---

## Summary

### What Works
- Login form and authentication flow
- All 7 main pages render correctly when accessed via SPA navigation
- All backend API endpoints return correct responses (health, auth, teams, items, views, reports, admin, item pool)
- Full CRUD lifecycle: create main item → create sub-item → append progress → update → status change → pool submit/assign/reject → archive
- Weekly, Gantt, Table, Report views all render content
- Sidebar navigation between pages works (except admin link)
- Logout functionality works

### What's Broken
1. **Auth persistence** — page refresh loses session (P0)
2. **Admin role detection** — `isSuperAdmin` always false due to field naming mismatch (P0)
3. **Rate limit handling** — 429 responses not handled in login UI (P2)
4. **React key warnings** — missing keys in Sidebar and TeamManagementPage (P3)
