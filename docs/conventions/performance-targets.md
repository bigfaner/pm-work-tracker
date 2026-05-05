---
scope: global
source: feature/pm-work-tracker
---

# Performance Targets

## Report Export

- **Weekly report Markdown export must complete in < 5 seconds.**
- This covers `ReportService.ExportMarkdown` end-to-end: data aggregation + Markdown generation.
- Context: PRD specifies this as a quantified requirement for weekly reporting workflow (PRD 6).

## Weekly View

- **Weekly view page load must complete in < 2 seconds.**
- Covers `ViewService.WeeklyView` query + response serialization.
- Context: PRD targets "常规列表页加载时间 < 2 秒" for all list pages including weekly view (PRD 6).

## List Pages (General)

- **All standard list pages must load in < 2 seconds.**
- Applies to MainItem list, SubItem list, ItemPool list, Table view.
- Default page size: 20 items (Table view: 50 items).

## JWT Token Expiry

- **JWT tokens expire 24 hours after issuance.**
- No refresh token mechanism in v1 -- stateless design.
- Client stores token in memory (not localStorage) and attaches as `Authorization: Bearer <token>`.
- JWT secret loaded from `JWT_SECRET` environment variable (minimum 32 bytes), never hardcoded.
- Context: Tech design decision -- 24h expiry balances security and usability without refresh token complexity.

## Concurrent Access

- **Multiple users can operate simultaneously without data conflicts.**
- GORM handles row-level locking for updates. Progress append is synchronous.
- Context: PRD specifies "多人同时操作不冲突，数据实时一致" (PRD 6).

## Login Rate Limiting

- **Login endpoint rate limited to 10 requests per minute per IP.**
- Enforced via Gin rate-limit middleware on `/api/v1/auth/login`.
- Context: Security mitigation against brute-force attacks.
