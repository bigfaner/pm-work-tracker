---
feature: decision-log
source: prd/prd-spec.md, prd/prd-user-stories.md
---

# Business Rules: Decision Log

## BR-1: Draft/Published Lifecycle

A decision log has exactly two states: `draft` and `published`.

- **draft**: Only the creator can see and edit it. The creator may edit or publish it any number of times.
- **published**: Visible to all team members. Immutable -- cannot be edited, deleted, or reverted to draft.

State machine: `draft -> draft` (edit/save), `draft -> published` (publish). Published is a terminal state.

## BR-2: Draft Visibility

Drafts are private to their creator. When listing decision logs for a main item:

- Published decisions: visible to all team members.
- Drafts: only returned if `createdBy` matches the requesting user.

No team member (including admins) can see another user's drafts.

## BR-3: Permission Model

Decision log write operations reuse the `main_item:update` permission code. No separate permission codes are introduced.

| Operation | Permission Required |
|-----------|-------------------|
| List decision logs | Team membership only |
| Create decision log | `main_item:update` |
| Edit draft | `main_item:update` + owner |
| Publish draft | `main_item:update` + owner |

Users without `main_item:update` can view published decisions but see no add/edit buttons.

## BR-4: Predefined Categories

Decisions must have exactly one category from a fixed enum of six values:

| Value | Label |
|-------|-------|
| `technical` | Technical |
| `resource` | Resource |
| `requirement` | Requirement |
| `schedule` | Schedule |
| `risk` | Risk |
| `other` | Other |

Categories are not customizable.

## BR-5: Free-Form Tags

Decisions may have zero or more free-form tags (optional field).

- Each tag is a string with a maximum length of 20 characters.
- Tags are stored as a JSON array of strings.
- Tags are display-only metadata, not a query/filter dimension.

## BR-6: Content Validation

Decision content is required and must not exceed 2000 characters. The content field supports multi-line text (newlines).

## BR-7: Timeline Display

Decision logs are displayed in a timeline on the main item detail page:

- Sorted by creation time in descending order (newest first).
- Paginated at 20 items per page with lazy loading (scroll to load more).
- Each item shows: category badge, tags as badges, content summary (first 80 chars + "..."), creator name, timestamp, and status indicator (draft badge for drafts only).
- Clicking an item expands it to show full content.

## BR-8: Immutability Enforcement

Any attempt to modify a published decision (edit or re-publish) must return HTTP 403. This is enforced at the service layer, not just the UI.

Similarly, editing another user's draft must return HTTP 403, even if the caller has `main_item:update` permission.

## BR-9: Scope

Decision logs belong to main items only. Sub-items do not have decision logs.

Cross-item search and filtering of decision logs is out of scope. No notification mechanism for new decisions. No export functionality.

## BR-10: Performance Requirements

| Metric | Target |
|--------|--------|
| List API response time | < 200ms (for up to 100 records) |
| Supported concurrency | 50 QPS per main item |
| Records per main item | Expected < 200, no hard limit |
