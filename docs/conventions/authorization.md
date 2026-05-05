---
scope: global
source: feature/improve-ui BIZ-001
---

# Authorization Rules

## Page-Level Access Control

Access control follows a strict role hierarchy:

| Role | Can Access |
|------|-----------|
| SuperAdmin | User management, all teams, team detail, member management |
| PM | Own team detail, member management for own team |
| Team Member | Read-only views for own team |

## Navigation Rules

Non-authorized users see the navigation entry **hidden** (not disabled, not 403). This prevents confusion about features they don't have access to.

## API Enforcement

Backend middleware chain: `Auth → TeamScope → RequireRole`. Team data isolation via `teamID` filtering in queries.

---

## Permission Refresh Strategy

_Source: feature/rbac-permissions_

Frontend permission data is refreshed on these triggers:

| Trigger | Behavior |
|---------|----------|
| Login success | Immediately call `fetchPermissions()` |
| Route change | If last fetch was >5 minutes ago, re-fetch |
| 403 response | Immediately re-fetch (permissions may have changed) |
| Manual | Refresh button available on team member management page |

**Why:** Permissions can change when an admin edits a role or changes a member's role. The 5-minute stale threshold balances freshness with API call frequency. The 403-triggered refresh handles the case where permissions were revoked mid-session.

**Implementation:** Auth store tracks `permissionsLoadedAt` timestamp. `fetchPermissions()` calls `GET /api/v1/me/permissions` and updates the store.

## Role Binding at Team Invite Time

_Source: feature/rbac-permissions_

When a user is invited to a team, the inviting PM/SuperAdmin selects a role from the system-defined role list (excluding superadmin). The role is stored in `team_members.role_id` (FK to `roles` table).

**Key rules:**
- A user can have different roles in different teams.
- Role assignment happens only at invite time or via explicit role change by PM/SuperAdmin.
- `superadmin` role cannot be assigned via team invite -- it is a global role managed separately.
- When changing a member's role, the new role must exist and not be `superadmin`.

**Why:** Separates team-level role management from global SuperAdmin management. Each team operates independently for role assignments.
