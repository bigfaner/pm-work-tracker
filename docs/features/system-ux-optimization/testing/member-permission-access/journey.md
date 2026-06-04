---
feature: "system-ux-optimization"
journey: "member-permission-access"
risk_level: "Medium"
surface_types: ["web", "api"]
sources:
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 6)
  - docs/features/system-ux-optimization/prd/prd-spec.md
generated: "2026-06-04"
---

# Journey: member-permission-access

**Risk Level**: Medium

<!-- Risk Classification Criteria:
  Medium = Workflow involves multi-step interaction without irreversible side effects
  Login and menu access validation does not mutate data but involves multi-step auth flow.
-->

## Overview

Member-role user logs in and accesses the system with correct permission-scoped menus and functions, verifying the fix for the nil RoleKey permission query bug.

## Setup

- A team member with member role exists in the system
- The member role has default permissions including item_pool:submit and main_item:list
- The backend permission middleware has been fixed to handle nil RoleKey correctly

## Happy Path

### Step 1: Member user login

**User Action**: Member-role user logs into the system with valid credentials

**Expected Result**: Authentication succeeds; user is redirected to the main dashboard; no permission errors occur during login

### Step 2: Verify menu visibility

**User Action**: Member user views the navigation menu after login

**Expected Result**: Menu displays at minimum: todo item submission (item_pool:submit) and item listing (main_item:list); menus outside the member role's permissions (e.g., delete, admin functions) are not visible

### Step 3: Access item listing page

**User Action**: Member user navigates to the item listing page

**Expected Result**: Page loads successfully; items within the user's team scope are displayed; no permission denied errors

### Step 4: Access todo submission page

**User Action**: Member user navigates to the todo item submission page

**Expected Result**: Page loads successfully; submission form is functional; no permission denied errors

## Edge Cases

### Step 1b: Login with no role assigned

**Precondition**: A user exists in the team member list with an empty role_key

**User Action**: Admin views the team member list

**Expected Result**: No team member records exist with empty role_key; the system ensures all members have a valid role assigned

### Step 2b: API access without valid token

**Precondition**: No authentication token is provided

**User Action**: API request is sent to GET /api/menus without a Bearer token

**Expected Result**: API returns 401 Unauthorized; no menu data is returned; no sensitive information is leaked

### Step 3b: Member user attempts PM-only action

**Precondition**: Member user is logged in

**User Action**: Member user sends a DELETE request via API to an item endpoint

**Expected Result**: API returns 403 Forbidden; no data is modified; UI does not expose the delete button

### Step 4b: Expired session during navigation

**Precondition**: Member user's session has expired

**User Action**: Member user attempts to navigate to a protected page

**Expected Result**: User is redirected to login page or shown a session-expired message; no partial data is displayed

## Journey Invariants

- Member-role users always receive at least item_pool:submit and main_item:list permissions after login
- Users with empty/null role_key are never left without permissions -- the middleware correctly queries the member role's default permission set
- Menu items and actions visible to a user always match their role's permission set exactly
- API endpoints enforce the same permission checks as the UI -- no UI-only permission enforcement
