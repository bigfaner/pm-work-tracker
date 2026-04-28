---
feature: "integration-test-coverage"
generated: 2026-04-28
source: prd/prd-spec.md, prd/prd-user-stories.md
total_cases: 160
---

# Test Cases: Integration Test Coverage

All test cases are API type (backend-only Go integration + unit tests).

## F1: Item Lifecycle (api/item-lifecycle)

Target: `backend/tests/integration/item_lifecycle_test.go`

### MainItem CRUD

| Test ID | Source (PRD) | Description | Pre-conditions | Steps | Expected | Priority |
|---------|-------------|-------------|----------------|-------|----------|----------|
| TC-F1-001 | Story 1 AC1, F1 row 1 happy | PM creates MainItem with valid fields | Team + PM user exist | POST /teams/:id/main-items with title, priority, valid dates | 201, response contains created item with correct fields | P0 |
| TC-F1-002 | Story 1 AC1, F1 row 1 validation | Create MainItem missing title | Team + PM user exist | POST /teams/:id/main-items without title | 422, validation error | P0 |
| TC-F1-003 | Story 1 AC1, F1 row 1 validation | Create MainItem with invalid priority | Team + PM user exist | POST /teams/:id/main-items with priority "X99" | 422, validation error | P0 |
| TC-F1-004 | Story 1 AC1, F1 row 1 validation | Create MainItem with invalid date range | Team + PM user exist | POST /teams/:id/main-items with endDate before startDate | 422, validation error | P1 |
| TC-F1-005 | Story 1 AC7, F1 row 1 permission | Member creates MainItem | Team + member user exist | POST /teams/:id/main-items as member | 403, permission denied | P0 |
| TC-F1-006 | F1 row 2 happy | List MainItems with pagination | Team + PM user + items exist | GET /teams/:id/main-items?page=1&pageSize=10 | 200, paginated list with items and total | P0 |
| TC-F1-007 | F1 row 3 happy | Get MainItem detail | MainItem exists | GET /teams/:id/main-items/:itemId | 200, item detail with all fields | P0 |
| TC-F1-008 | F1 row 3 not-found | Get nonexistent MainItem | Team exists, item ID invalid | GET /teams/:id/main-items/99999 | 404, not found | P0 |
| TC-F1-009 | F1 row 3 permission | Non-team member gets MainItem detail | User not in team | GET /teams/:id/main-items/:itemId as outsider | 403, permission denied | P0 |
| TC-F1-010 | F1 row 4 happy | Update MainItem fields | MainItem in non-terminal state | PUT /teams/:id/main-items/:itemId with new title/priority | 200, updated fields reflected | P0 |
| TC-F1-011 | F1 row 4 validation | Update terminal-state MainItem | MainItem in completed state | PUT /teams/:id/main-items/:itemId with new title | 422, terminal state not editable | P1 |
| TC-F1-012 | F1 row 4 validation | Update MainItem with non-numeric assigneeKey | MainItem exists | PUT /teams/:id/main-items/:itemId with assigneeKey="abc" | 422, validation error | P1 |
| TC-F1-013 | F1 row 4 permission | Member updates MainItem | Member user in team | PUT /teams/:id/main-items/:itemId as member | 403, permission denied | P0 |
| TC-F1-014 | F1 row 4 not-found | Update nonexistent MainItem | Team exists, item ID invalid | PUT /teams/:id/main-items/99999 | 404, not found | P0 |

### Status & Archive

| Test ID | Source (PRD) | Description | Pre-conditions | Steps | Expected | Priority |
|---------|-------------|-------------|----------------|-------|----------|----------|
| TC-F1-015 | Story 1 AC4, F1 row 5 happy | Valid status transition on MainItem | MainItem in in-progress state | PUT /teams/:id/main-items/:itemId/status with target "reviewing" | 200, status updated | P0 |
| TC-F1-016 | Story 1 AC5, F1 row 5 validation | Invalid status transition new->completed | MainItem in new state | PUT status with target "completed" | 422, transition not allowed | P0 |
| TC-F1-017 | Story 1 AC4, F1 row 5 cascade | Terminal status cascades to sub-items | MainItem in-progress with sub-items | PUT status to terminal state | Sub-items auto-complete | P0 |
| TC-F1-018 | F1 row 5 permission | Member changes MainItem status | Member user in team | PUT status as member | 403, permission denied | P0 |
| TC-F1-019 | F1 row 5 not-found | Status change on nonexistent MainItem | Team exists, item ID invalid | PUT status on 99999 | 404, not found | P0 |
| TC-F1-020 | F1 row 6 happy | Available transitions for MainItem | MainItem in in-progress state | GET /teams/:id/main-items/:itemId/available-transitions | 200, list of valid transitions | P0 |
| TC-F1-021 | F1 row 6 cascade | Available transitions for terminal MainItem | MainItem in completed state | GET available-transitions | 200, empty list | P1 |
| TC-F1-022 | Story 1 AC6, F1 row 7 happy | Archive completed MainItem | MainItem in completed state | POST /teams/:id/main-items/:itemId/archive | 200, archive successful | P0 |
| TC-F1-023 | Story 1 AC6, F1 row 7 validation | Archive in-progress MainItem | MainItem in in-progress state | POST archive | 422, cannot archive non-completed item | P0 |
| TC-F1-024 | F1 row 7 not-found | Archive nonexistent MainItem | Team exists, item ID invalid | POST archive on 99999 | 404, not found | P0 |

### SubItem CRUD

| Test ID | Source (PRD) | Description | Pre-conditions | Steps | Expected | Priority |
|---------|-------------|-------------|----------------|-------|----------|----------|
| TC-F1-025 | Story 1 AC2, F1 row 8 happy | Create SubItem with weight | MainItem exists | POST /teams/:id/main-items/:itemId/sub-items with code, weight | 201, SubItem linked to MainItem | P0 |
| TC-F1-026 | Story 1 AC2, F1 row 8 validation | Create SubItem with weight <= 0 | MainItem exists | POST sub-items with weight=0 | 422, validation error | P0 |
| TC-F1-027 | F1 row 8 validation | Create SubItem with duplicate code | MainItem exists, SubItem with same code exists | POST sub-items with existing code | 422, validation error | P1 |
| TC-F1-028 | F1 row 8 permission | Member creates SubItem without sub_item:create permission | Member user in team | POST sub-items as member | 403, permission denied | P0 |
| TC-F1-029 | F1 row 8 not-found | Create SubItem on nonexistent MainItem | Team exists, main item ID invalid | POST sub-items on 99999 | 404, main item not found | P0 |
| TC-F1-030 | F1 row 9 happy | List SubItems for MainItem | MainItem with sub-items exists | GET /teams/:id/main-items/:itemId/sub-items | 200, list of sub-items | P0 |
| TC-F1-031 | F1 row 10 happy | Get SubItem detail | SubItem exists | GET /teams/:id/sub-items/:subId | 200, SubItem detail | P0 |
| TC-F1-032 | F1 row 10 not-found | Get nonexistent SubItem | Team exists, sub-item ID invalid | GET /teams/:id/sub-items/99999 | 404, not found | P0 |
| TC-F1-033 | F1 row 10 permission | Non-team member gets SubItem detail | User not in team | GET sub-items as outsider | 403, permission denied | P0 |
| TC-F1-034 | F1 row 11 happy | Update SubItem fields | SubItem exists | PUT /teams/:id/sub-items/:subId with new title | 200, updated fields | P0 |
| TC-F1-035 | F1 row 11 validation | Update SubItem with unresolvable assigneeKey | SubItem exists | PUT sub-items with assigneeKey pointing to non-existent bizKey | 422, validation error | P1 |
| TC-F1-036 | F1 row 11 not-found | Update nonexistent SubItem | Team exists, sub-item ID invalid | PUT /teams/:id/sub-items/99999 | 404, not found | P0 |

### SubItem Status & Assignee

| Test ID | Source (PRD) | Description | Pre-conditions | Steps | Expected | Priority |
|---------|-------------|-------------|----------------|-------|----------|----------|
| TC-F1-037 | F1 row 12 happy | Valid status transition on SubItem | SubItem in in-progress state | PUT /teams/:id/sub-items/:subId/status | 200, status updated | P0 |
| TC-F1-038 | F1 row 12 validation | Invalid status transition new->completed on SubItem | SubItem in new state | PUT status with target "completed" | 422, transition not allowed | P0 |
| TC-F1-039 | F1 row 12 cascade | Terminal SubItem status recalculates MainItem completion | SubItem completed, MainItem has multiple sub-items | PUT status to terminal | MainItem completion recalculated | P0 |
| TC-F1-040 | F1 row 12 not-found | Status change on nonexistent SubItem | Team exists, sub-item ID invalid | PUT status on 99999 | 404, not found | P0 |
| TC-F1-041 | F1 row 13 happy | Available transitions for SubItem | SubItem in in-progress state | GET available-transitions | 200, list of valid transitions | P0 |
| TC-F1-042 | F1 row 13 cascade | Available transitions for terminal SubItem | SubItem in completed state | GET available-transitions | 200, empty list | P1 |
| TC-F1-043 | F1 row 14 happy | Assign SubItem to team member | SubItem exists, member in team | PUT /teams/:id/sub-items/:subId/assignee with valid assigneeKey | 200, assignee set | P0 |
| TC-F1-044 | F1 row 14 permission | Non-member assigns SubItem | User not in team | PUT assignee as outsider | 403, permission denied | P0 |
| TC-F1-045 | F1 row 14 cascade | Clear SubItem assignee | SubItem with assignee set | PUT assignee with empty assigneeKey | 200, assignee cleared | P1 |

### Progress

| Test ID | Source (PRD) | Description | Pre-conditions | Steps | Expected | Priority |
|---------|-------------|-------------|----------------|-------|----------|----------|
| TC-F1-046 | Story 1 AC3, F1 row 15 happy | Append progress with completion=60 | SubItem with 0% completion | POST /teams/:id/sub-items/:subId/progress with completion=60 | 200, SubItem completion=60%, MainItem recalculated | P0 |
| TC-F1-047 | Story 1 AC3, F1 row 15 cascade | Progress 100% auto-transitions SubItem status | SubItem in in-progress at 80% | POST progress with completion=100 | SubItem auto-transitions to terminal status | P0 |
| TC-F1-048 | F1 row 15 validation | Regress completion (lower than previous) | SubItem at 60% | POST progress with completion=30 | 422, PROGRESS_REGRESSION | P0 |
| TC-F1-049 | F1 row 15 cascade | Completion rolls up to MainItem | MainItem with 2 SubItems (equal weight) | Append progress=100 to one SubItem | MainItem completion = 50% | P0 |
| TC-F1-050 | F1 row 16 happy | List progress records (reverse chronological) | SubItem with 3 progress records | GET /teams/:id/sub-items/:subId/progress | 200, records in reverse order | P0 |
| TC-F1-051 | F1 row 17 happy | Patch latest completion record | Progress record exists as latest | PATCH /teams/:id/progress/:recordId/completion with new value | 200, SubItem recalculated | P0 |
| TC-F1-052 | F1 row 17 not-found | Patch nonexistent progress record | Record ID invalid | PATCH /teams/:id/progress/99999/completion | 404, not found | P0 |
| TC-F1-053 | F1 row 17 cascade | Patch non-latest record does not cascade | Progress record exists but not latest | PATCH completion on older record | 200, SubItem not recalculated | P1 |

---

## F2: Item Pool (api/item-pool)

Target: `backend/tests/integration/item_pool_test.go`

| Test ID | Source (PRD) | Description | Pre-conditions | Steps | Expected | Priority |
|---------|-------------|-------------|----------------|-------|----------|----------|
| TC-F2-001 | Story 2 AC1, F2 row 1 happy | PM submits pool item | Team + PM user exist | POST /teams/:id/item-pool with title | 201, pool item status=pending | P0 |
| TC-F2-002 | Story 2 AC1, F2 row 1 validation | Submit pool item missing title | Team + PM user exist | POST item-pool without title | 422, validation error | P0 |
| TC-F2-003 | F2 row 1 validation | Submit pool item with title > 100 chars | Team + PM user exist | POST item-pool with 101-char title | 422, validation error | P1 |
| TC-F2-004 | F2 row 1 permission | Member submits pool item | Member user in team | POST item-pool as member | 403, permission denied | P0 |
| TC-F2-005 | F2 row 2 happy | List pool items with status filter | Pool items with various statuses exist | GET /teams/:id/item-pool?status=pending | 200, filtered list | P0 |
| TC-F2-006 | F2 row 3 happy | Get pool item detail | Pool item exists | GET /teams/:id/item-pool/:poolId | 200, pool item detail | P0 |
| TC-F2-007 | F2 row 3 not-found | Get nonexistent pool item | Team exists, pool ID invalid | GET /teams/:id/item-pool/99999 | 404, not found | P0 |
| TC-F2-008 | Story 2 AC2, F2 row 4 happy | Assign pool item to valid MainItem | Pending pool item + MainItem exist | POST /teams/:id/item-pool/:poolId/assign with mainItemId | 200, SubItem created, pool status=assigned | P0 |
| TC-F2-009 | Story 2 AC3, F2 row 4 cascade | Assign pool item to nonexistent MainItem | Pending pool item exists, MainItem does not | POST assign with invalid mainItemId | Rollback: pool status stays pending, no SubItem created | P0 |
| TC-F2-010 | F2 row 4 permission | Member assigns pool item | Member user in team | POST assign as member | 403, permission denied | P0 |
| TC-F2-011 | Story 2 AC4, F2 row 4 cascade | Assign already-processed pool item | Pool item in assigned status | POST assign again | 409, conflict | P0 |
| TC-F2-012 | F2 row 5 happy | Convert pool item to MainItem | Pending pool item exists | POST /teams/:id/item-pool/:poolId/convert-to-main | 200, MainItem created, pool status=assigned | P0 |
| TC-F2-013 | F2 row 5 permission | Member converts pool item | Member user in team | POST convert-to-main as member | 403, permission denied | P0 |
| TC-F2-014 | F2 row 5 cascade | Convert already-processed pool item | Pool item in assigned status | POST convert-to-main | 409, conflict | P0 |
| TC-F2-015 | Story 2 AC5, F2 row 6 happy | Reject pool item with reason | Pending pool item exists | POST /teams/:id/item-pool/:poolId/reject with reason | 200, pool status=rejected | P0 |
| TC-F2-016 | Story 2 AC6, F2 row 6 validation | Reject pool item without reason | Pending pool item exists | POST reject without reason | 422, validation error | P0 |
| TC-F2-017 | F2 row 6 permission | Member rejects pool item | Member user in team | POST reject as member | 403, permission denied | P0 |
| TC-F2-018 | F2 row 6 cascade | Reject already-processed pool item | Pool item in assigned status | POST reject | 409, conflict | P0 |

---

## F3: Team Management (api/team-management)

Target: `backend/tests/integration/team_management_test.go`

| Test ID | Source (PRD) | Description | Pre-conditions | Steps | Expected | Priority |
|---------|-------------|-------------|----------------|-------|----------|----------|
| TC-F3-001 | Story 3 AC1, F3 row 1 happy | Create team, creator auto PM | Logged-in user exists | POST /teams with name, code | 201, creator is PM member | P0 |
| TC-F3-002 | F3 row 1 validation | Create team with duplicate code | Team with same code exists | POST /teams with existing code | 422, validation error | P0 |
| TC-F3-003 | F3 row 2 happy | List user teams | User in multiple teams | GET /teams | 200, list of user teams | P0 |
| TC-F3-004 | F3 row 2 cascade | New user sees empty team list | Newly created user | GET /teams | 200, empty list | P1 |
| TC-F3-005 | F3 row 3 happy | Get team detail | User is team member | GET /teams/:id | 200, team detail | P0 |
| TC-F3-006 | F3 row 3 permission | Non-member gets team detail | User not in team | GET /teams/:id as outsider | 403, permission denied | P0 |
| TC-F3-007 | F3 row 3 not-found | Get nonexistent team | Team ID invalid | GET /teams/99999 | 404, not found | P0 |
| TC-F3-008 | F3 row 4 happy | Update team fields | PM user in team | PUT /teams/:id with new name/description | 200, updated fields | P0 |
| TC-F3-009 | F3 row 4 validation | Update team with empty name | PM user in team | PUT /teams/:id with name="" | 422, validation error | P0 |
| TC-F3-010 | F3 row 4 validation | Update team with name > 100 chars | PM user in team | PUT /teams/:id with 101-char name | 422, validation error | P1 |
| TC-F3-011 | F3 row 4 validation | Update team with description > 500 chars | PM user in team | PUT /teams/:id with 501-char description | 422, validation error | P1 |
| TC-F3-012 | F3 row 4 permission | Non-PM updates team | Member user in team | PUT /teams/:id as member | 403, permission denied | P0 |
| TC-F3-013 | Story 3 AC4, F3 row 5 happy | Disband team | PM user in team with items | DELETE /teams/:id | 200, team and items deleted | P0 |
| TC-F3-014 | F3 row 5 permission | Non-PM disbands team | Member user in team | DELETE /teams/:id as member | 403, permission denied | P0 |
| TC-F3-015 | F3 row 5 not-found | Disband nonexistent team | Team ID invalid | DELETE /teams/99999 | 404, not found | P0 |
| TC-F3-016 | F3 row 6 happy | Search users for invitation | PM user in team | GET /teams/:id/search-users?q=keyword | 200, matching users | P0 |
| TC-F3-017 | F3 row 6 permission | Member searches users | Member user in team | GET search-users as member | 403, permission denied | P0 |
| TC-F3-018 | F3 row 6 cascade | Search returns empty results | No matching users | GET search-users?q=nonexistent | 200, empty list | P1 |
| TC-F3-019 | Story 3 AC2, F3 row 7 happy | Invite member with role | PM user, target user exists | POST /teams/:id/members with userId, role | 200, member added with role | P0 |
| TC-F3-020 | F3 row 7 validation | Invite already-member user | PM user, target already in team | POST members with existing member userId | 409, conflict | P0 |
| TC-F3-021 | F3 row 7 permission | Member invites user | Member user in team | POST members as member | 403, permission denied | P0 |
| TC-F3-022 | F3 row 7 not-found | Invite nonexistent user | PM user, userId invalid | POST members with 99999 | 404, user not found | P0 |
| TC-F3-023 | Story 3 AC5, F3 row 8 happy | Remove member from team | PM user, member exists | DELETE /teams/:id/members/:userId | 200, member removed | P0 |
| TC-F3-024 | F3 row 8 validation | Remove PM from team | PM user | DELETE /teams/:id/members/:pmUserId | 422, PM cannot be removed | P0 |
| TC-F3-025 | F3 row 8 permission | Member removes member | Member user in team | DELETE members as member | 403, permission denied | P0 |
| TC-F3-026 | F3 row 8 not-found | Remove nonexistent member | PM user, userId invalid | DELETE /teams/:id/members/99999 | 404, not found | P0 |
| TC-F3-027 | Story 3 AC3, F3 row 9 happy | Change member role | PM user, member exists | PUT /teams/:id/members/:userId/role with new role | 200, role updated | P0 |
| TC-F3-028 | F3 row 9 validation | Change PM role | PM user | PUT members/:pmUserId/role | 403, PM role cannot be changed | P0 |
| TC-F3-029 | F3 row 9 permission | Member changes role | Member user in team | PUT members/:userId/role as member | 403, permission denied | P0 |
| TC-F3-030 | F3 row 9 not-found | Change role for nonexistent member | PM user, userId invalid | PUT /teams/:id/members/99999/role | 404, not found | P0 |

---

## F4: Admin User Management (api/admin-user)

Target: `backend/tests/integration/admin_user_test.go`

| Test ID | Source (PRD) | Description | Pre-conditions | Steps | Expected | Priority |
|---------|-------------|-------------|----------------|-------|----------|----------|
| TC-F4-001 | Story 4 AC5, F4 row 1 happy | SuperAdmin lists users | SuperAdmin user exists | GET /admin/users?page=1&pageSize=10 | 200, paginated user list | P0 |
| TC-F4-002 | F4 row 1 happy | SuperAdmin lists users with search | SuperAdmin user, users exist | GET /admin/users?search=keyword | 200, filtered list | P0 |
| TC-F4-003 | F4 row 1 permission | Non-SuperAdmin lists users | Regular user | GET /admin/users as regular user | 403, permission denied | P0 |
| TC-F4-004 | Story 4 AC1, F4 row 2 happy | SuperAdmin creates user | SuperAdmin user exists | POST /admin/users with username, password | 201, user created | P0 |
| TC-F4-005 | Story 4 AC2, F4 row 2 validation | Create user with duplicate username | SuperAdmin, username already exists | POST /admin/users with existing username | 409, conflict | P0 |
| TC-F4-006 | F4 row 2 permission | Non-SuperAdmin creates user | Regular user | POST /admin/users as regular user | 403, permission denied | P0 |
| TC-F4-007 | F4 row 3 happy | Get user detail | SuperAdmin, user exists | GET /admin/users/:userId | 200, user detail | P0 |
| TC-F4-008 | F4 row 3 not-found | Get nonexistent user | SuperAdmin exists | GET /admin/users/99999 | 404, not found | P0 |
| TC-F4-009 | F4 row 4 happy | Update user info | SuperAdmin, user exists | PUT /admin/users/:userId with displayName, email | 200, updated fields | P0 |
| TC-F4-010 | F4 row 4 validation | Update user with empty displayName | SuperAdmin, user exists | PUT /admin/users/:userId with displayName="" | 422, validation error | P0 |
| TC-F4-011 | F4 row 4 validation | Update user with displayName > 64 chars | SuperAdmin, user exists | PUT /admin/users/:userId with 65-char displayName | 422, validation error | P1 |
| TC-F4-012 | F4 row 4 validation | Update user with email > 100 chars | SuperAdmin, user exists | PUT /admin/users/:userId with 101-char email | 422, validation error | P1 |
| TC-F4-013 | F4 row 4 validation | Update user with nonexistent teamKey | SuperAdmin, user exists | PUT /admin/users/:userId with invalid teamKey | 422, validation error | P1 |
| TC-F4-014 | F4 row 4 not-found | Update nonexistent user | SuperAdmin exists | PUT /admin/users/99999 | 404, not found | P0 |
| TC-F4-015 | Story 4 AC3, F4 row 5 happy | SuperAdmin disables user | SuperAdmin, target user enabled | PUT /admin/users/:userId/status with disabled | 200, user disabled | P0 |
| TC-F4-016 | Story 4 AC4, F4 row 5 validation | SuperAdmin disables self | SuperAdmin user | PUT /admin/users/:selfId/status with disabled | 422, cannot disable self | P0 |
| TC-F4-017 | F4 row 5 not-found | Toggle status for nonexistent user | SuperAdmin exists | PUT /admin/users/99999/status | 404, not found | P0 |
| TC-F4-018 | F4 row 5 happy | SuperAdmin enables user | SuperAdmin, target user disabled | PUT /admin/users/:userId/status with enabled | 200, user enabled | P0 |
| TC-F4-019 | F4 row 6 happy | List teams with member count | SuperAdmin exists | GET /admin/teams | 200, teams with memberCount field | P0 |

---

## F5: Views & Reports (api/views-reports)

Target: `backend/tests/integration/views_reports_test.go`

| Test ID | Source (PRD) | Description | Pre-conditions | Steps | Expected | Priority |
|---------|-------------|-------------|----------------|-------|----------|----------|
| TC-F5-001 | Story 5 AC1, F5 row 1 happy | Weekly view with 3 items (1 completed, 2 in-progress) | Team with 3 items in those states | GET /teams/:id/views/weekly | 200, stats={NEW:0,completed:1,inProgress:2,overdue:0} | P0 |
| TC-F5-002 | Story 5 AC1, F5 row 1 format | Weekly view shows week-over-week delta | Team with items this week and last | GET /teams/:id/views/weekly | 200, delta field present | P1 |
| TC-F5-003 | F5 row 1 empty | Weekly view for empty team | Team with no items | GET /teams/:id/views/weekly | 200, all-zero stats | P0 |
| TC-F5-004 | F5 row 1 permission | Weekly view denied for non-member | User not in team | GET /teams/:id/views/weekly as outsider | 403, permission denied | P0 |
| TC-F5-005 | F5 row 2 happy | Gantt view with items and sub-items | Team with MainItems + SubItems | GET /teams/:id/views/gantt | 200, items with startDate/endDate, sub-items nested | P0 |
| TC-F5-006 | F5 row 2 format | Gantt view has status color mapping | Team with items in various statuses | GET /teams/:id/views/gantt | 200, status field mapped to color key | P1 |
| TC-F5-007 | F5 row 2 empty | Gantt view for empty team | Team with no items | GET /teams/:id/views/gantt | 200, empty array | P0 |
| TC-F5-008 | Story 5 AC2, F5 row 3 happy | Table view filter by status=completed | Team with mixed-status items | GET /teams/:id/views/table?status=completed | 200, only completed items | P0 |
| TC-F5-009 | F5 row 3 happy | Table view filter by overdue=true | Team with overdue items | GET /teams/:id/views/table?overdue=true | 200, only overdue non-completed items | P0 |
| TC-F5-010 | F5 row 3 format | Table view pagination total matches filter count | Team with many items | GET /teams/:id/views/table?status=completed&page=1 | 200, total matches actual count | P1 |
| TC-F5-011 | F5 row 3 empty | Table view for empty team | Team with no items | GET /teams/:id/views/table | 200, {items:[], total:0} | P0 |
| TC-F5-012 | Story 5 AC3, F5 row 4 happy | CSV export with data | Team with items | GET /teams/:id/views/table/export | 200, starts with UTF-8 BOM, has header + data rows | P0 |
| TC-F5-013 | F5 row 4 format | CSV export has correct content-type | Team with items | GET /teams/:id/views/table/export | Content-Type contains text/csv | P0 |
| TC-F5-014 | F5 row 4 empty | CSV export for empty team | Team with no items | GET /teams/:id/views/table/export | 200, BOM + header only, no data rows | P0 |
| TC-F5-015 | Story 5 AC4, F5 row 5 happy | Weekly report preview with sections | Team with activity this week | GET /teams/:id/reports/weekly/preview | 200, Markdown with ## Summary and ## Items | P0 |
| TC-F5-016 | F5 row 5 empty | Weekly report preview with no activity | Team with no activity | GET /teams/:id/reports/weekly/preview | 200, "no activity this week" | P0 |
| TC-F5-017 | Story 5 AC5, F5 row 6 happy | Weekly report export | Team with activity | GET /teams/:id/reports/weekly/export | 200, full Markdown with all sections | P0 |
| TC-F5-018 | F5 row 6 format | Report export has correct content-type | Team with activity | GET /teams/:id/reports/weekly/export | Content-Type: text/markdown | P0 |

---

## F6: Unit Test Gaps (api/unit-gaps)

Target: `backend/internal/handler/permission_handler_test.go` and edits to existing `*_service_test.go` files

| Test ID | Source (PRD) | Description | Pre-conditions | Steps | Expected | Priority |
|---------|-------------|-------------|----------------|-------|----------|----------|
| TC-F6-001 | Story 8 AC1 | GetPermissions handler returns permissions for role | Mock RoleService, valid role | Call GetPermissions handler | 200, permissions list | P0 |
| TC-F6-002 | Story 8 AC1 | GetPermissionCodes handler returns codes | Mock RoleService, valid role | Call GetPermissionCodes handler | 200, permission codes | P0 |
| TC-F6-003 | Story 8 AC2 | ConvertToMain creates MainItem and updates pool status | Mock repos, pending pool item | Call ConvertToMain service | MainItem created, pool status=assigned | P0 |
| TC-F6-004 | Story 8 AC2 | ConvertToMain on already-processed pool item | Mock repos, assigned pool item | Call ConvertToMain service | Error, no duplicate MainItem | P0 |
| TC-F6-005 | Story 8 AC3 | UpdateTeam with PM user updates fields | Mock teamRepo, PM user | Call UpdateTeam service | Fields updated successfully | P0 |
| TC-F6-006 | Story 8 AC3 | UpdateTeam denied for non-PM user | Mock teamRepo, member user | Call UpdateTeam service | Permission denied error | P0 |
| TC-F6-007 | Story 8 AC4 | ItemPoolService.GetByBizKey existing item | Mock repo, valid bizKey | Call GetByBizKey | Returns pool item | P0 |
| TC-F6-008 | Story 8 AC4 | ItemPoolService.GetByBizKey non-existing item | Mock repo, invalid bizKey | Call GetByBizKey | Returns not-found error | P0 |
| TC-F6-009 | Story 8 AC4 | ProgressService.GetByBizKey existing record | Mock repo, valid bizKey | Call GetByBizKey | Returns progress record | P0 |
| TC-F6-010 | Story 8 AC4 | ProgressService.GetByBizKey non-existing record | Mock repo, invalid bizKey | Call GetByBizKey | Returns not-found error | P0 |
| TC-F6-011 | Story 8 AC4 | SubItemService.GetByBizKey existing item | Mock repo, valid bizKey | Call GetByBizKey | Returns sub-item | P0 |
| TC-F6-012 | Story 8 AC4 | SubItemService.GetByBizKey non-existing item | Mock repo, invalid bizKey | Call GetByBizKey | Returns not-found error | P0 |

---

## Traceability Summary

### By PRD User Story

| Story | Test Cases | Count |
|-------|-----------|-------|
| Story 1: Item Lifecycle | TC-F1-001 through TC-F1-053 | 53 |
| Story 2: Item Pool | TC-F2-001 through TC-F2-018 | 18 |
| Story 3: Team Management | TC-F3-001 through TC-F3-030 | 30 |
| Story 4: Admin User | TC-F4-001 through TC-F4-019 | 19 |
| Story 5: Views & Reports | TC-F5-001 through TC-F5-018 | 18 |
| Story 6: Shared Helpers | Covered by F7 task, not test cases | -- |
| Story 7: Code Reviewer | Structural requirement, not test cases | -- |
| Story 8: Unit Test Gaps | TC-F6-001 through TC-F6-012 | 12 |
| **Total** | | **150** |

### By Type

| Type | Count |
|------|-------|
| API (integration) | 138 |
| API (unit) | 12 |
| **Total** | **150** |

### By Target File

| Target File | Test Case Range | Count |
|-------------|----------------|-------|
| `item_lifecycle_test.go` | TC-F1-001 -- TC-F1-053 | 53 |
| `item_pool_test.go` | TC-F2-001 -- TC-F2-018 | 18 |
| `team_management_test.go` | TC-F3-001 -- TC-F3-030 | 30 |
| `admin_user_test.go` | TC-F4-001 -- TC-F4-019 | 19 |
| `views_reports_test.go` | TC-F5-001 -- TC-F5-018 | 18 |
| `permission_handler_test.go` + `*_service_test.go` edits | TC-F6-001 -- TC-F6-012 | 12 |
| **Total** | | **150** |

### By Priority

| Priority | Count |
|----------|-------|
| P0 (happy path + critical edge cases) | 119 |
| P1 (format validation + less critical edges) | 31 |
| **Total** | **150** |
