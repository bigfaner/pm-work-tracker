---
feature: "system-ux-optimization"
journey: "member-permission-access"
risk_level: "High"
surface_types: ["web", "api"]
sources:
  - docs/features/system-ux-optimization/prd/prd-user-stories.md (Story 6)
  - docs/features/system-ux-optimization/prd/prd-spec.md (#6)
  - docs/features/system-ux-optimization/design/api-handbook.md
generated: "2026-06-04"
---

# Journey: member-permission-access

**Risk Level**: High

<!-- Risk Classification Criteria:
  High = Workflow involves auth regression testing and permission enforcement gaps
  A nil RoleKey bug silently broke access for an entire role class — security-adjacent regression risk.
-->

## Overview

Member-role user (including users with nil RoleKey) logs in and accesses the system with correct permission-scoped menus and functions, verifying the fix for the nil RoleKey permission query bug.

## Setup

- A team exists with member-role users
- A member user exists with a valid role_key and default permissions (item_pool:submit, main_item:list)
- A member user exists with an empty/null role_key column (the core bug scenario)
- The member role's default permission set includes item_pool:submit and main_item:list <!-- fact: prd-spec #6 -->
- A PM-role user exists with elevated permissions (main_item:delete, admin functions)

## Happy Path

### Step 1: Member user with valid role_key logs in
<!-- surface: web -->

**Precondition**: A member user with a valid role_key exists in the system <!-- fact: prd-spec #6 -->

**User Action**: Member user logs in with valid credentials

**Expected Result**: Authentication succeeds; user is redirected to the main dashboard; no permission errors occur during login

### Step 2: Member user with nil role_key logs in (core bug fix)
<!-- surface: web -->

**Precondition**: A member user exists with an empty/null role_key column in the database <!-- fact: prd-spec #6 — the nil RoleKey scenario is the core bug being fixed -->

**User Action**: Member user with nil role_key logs in with valid credentials

**Expected Result**: Authentication succeeds; the middleware falls back to querying the member role's default permission set; user is redirected to the main dashboard; the same menus and permissions are available as the user with a valid role_key

### Step 3: Verify menu visibility matches permission set
<!-- surface: web -->

**Precondition**: A member user is logged in (with either valid or nil role_key) <!-- fact: prd-spec #6, Story 6 -->

**User Action**: Member user views the navigation menu

**Expected Result**: Menu displays items corresponding to item_pool:submit and main_item:list permissions; menus outside the member role's permissions (e.g., delete, admin functions) are not visible

### Step 4: Access item listing page
<!-- surface: web -->

**Precondition**: A member user is logged in with main_item:list permission <!-- fact: Story 6 -->

**User Action**: Member user navigates to the item listing page

**Expected Result**: Page loads successfully; items within the user's team are displayed; no permission errors occur

### Step 5: API access with member permissions
<!-- surface: api -->

**Precondition**: A member user with nil role_key is authenticated <!-- fact: prd-spec #6 — confirms middleware fix works via API -->

**User Action**: An authenticated API request is sent to the item listing endpoint

**Expected Result**: The API returns item data within the user's team scope; no permission error is returned

## Edge Cases

### Step E1: Login with malformed credentials
<!-- surface: web -->

**Precondition**: The login form is displayed <!-- source: inferred — derived from Web surface `validation-error` mandatory outcome -->

**User Action**: User submits login with an empty password or malformed email format

**Expected Result**: A validation error message is displayed near the relevant field; no API request is sent; the form remains on screen for correction

### Step E2: Unauthenticated API request
<!-- surface: api -->

**Precondition**: An API request is sent without any authentication credentials <!-- source: inferred — derived from API surface `unauthorized` mandatory outcome -->

**User Action**: An API request is sent to a protected endpoint without a valid authentication token

**Expected Result**: The API returns an authentication error; no data is returned; no sensitive information is exposed in the response

### Step E3: Member user attempts PM-only action (API)
<!-- surface: api -->

**Precondition**: A member user is authenticated with default permissions that do not include main_item:delete <!-- source: inferred — derived from API surface `unauthorized` mandatory outcome -->

**User Action**: The member user sends an API request to a delete endpoint

**Expected Result**: The API returns an authorization error; no data is modified

### Step E4: Member user attempts PM-only action (Web)
<!-- surface: web -->

**Precondition**: A member user is viewing an item detail page that a PM user would see a delete button on

**User Action**: Member user views the item detail page

**Expected Result**: The delete button is not visible; no admin or PM-only controls are displayed

### Step E5: Session expired during navigation (Web)
<!-- surface: web -->

**Precondition**: A member user was previously authenticated and the session has since expired <!-- source: inferred — derived from Web surface `session-expired` mandatory outcome -->

**User Action**: Member user attempts to navigate to a page requiring authentication

**Expected Result**: The user is redirected to the login page; no partial data is displayed from the previous session

### Step E6: API request to non-existent resource
<!-- surface: api -->

**Precondition**: A member user is authenticated; the requested resource ID does not exist in the database <!-- source: inferred — derived from API surface `not-found` common boundary outcome -->

**User Action**: The member user sends an API request for a non-existent item ID

**Expected Result**: The API returns a "not found" error

### Step E7: API validation error for malformed request
<!-- surface: api -->

**Precondition**: An authenticated API request contains invalid or malformed parameters <!-- source: inferred — derived from API surface `validation-error` outcome -->

**User Action**: An API request is sent with missing required fields or invalid parameter values

**Expected Result**: The API returns a validation error response describing the invalid fields; no data is modified

## Journey Invariants

- Member-role users always receive at least item_pool:submit and main_item:list permissions after login, regardless of whether their role_key is valid or nil/empty
- The middleware correctly falls back to the member role's default permission set when role_key is nil or empty
- Menu items and actions visible to a user always match their role's permission set exactly
- API endpoints enforce the same permission checks as the UI — no UI-only permission enforcement
