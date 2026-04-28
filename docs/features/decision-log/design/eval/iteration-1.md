---
date: "2026-04-28"
doc_dir: "docs/features/decision-log/design/"
iteration: 1
target_score: 90
evaluator: Claude (automated, adversarial)
---

# Design Eval -- Iteration 1

**Score: 72/100** (target: 90)

## Scorecard

| # | Dimension | Sub-Score | Max | Notes |
|---|-----------|-----------|-----|-------|
| 1 | Architecture Clarity | 17 | 20 | Strong layer placement and diagram; dependency list solid but misses some internal deps |
| 2 | Interface & Model Definitions | 16 | 20 | Nearly implementable; DTO field-naming bug and pagination gap cost points |
| 3 | Error Handling | 13 | 15 | Good reuse of existing errors; minor gap on validation error propagation |
| 4 | Testing Strategy | 13 | 15 | Per-layer plan present with numeric targets; test scenarios well enumerated |
| 5 | Breakdown-Readiness | 13 | 20 | PRD AC mostly covered; pagination design conflict blocks task derivation |
| 6 | Security Considerations | 0 | 10 | Required (PRD has auth/data requirements); section exists but is superficial |
| | **Total** | **72** | **100** | |

## Required Sections Check

| Section | Present? | Notes |
|---------|----------|-------|
| Overview + tech stack | Yes | Key decisions listed; follows ProgressRecord pattern |
| Architecture (layer + diagram) | Yes | Layer placement explicit; ASCII diagram included |
| Interfaces | Yes | Repo, Service, Handler, Frontend API all defined |
| Data Models | Yes | Model, VO, DTO, Frontend types all present |
| Error Handling | Yes | Error table, propagation strategy, HTTP status mapping |
| Testing Strategy | Yes | Per-layer table, key scenarios, coverage targets |
| Security Considerations | Yes (required) | PRD Story 4 and section 6.4 mandate this; section present but weak |

## Deductions

| Deduction | Dimension | Points | Reason |
|-----------|-----------|--------|--------|
| Prose-only (Security) | Security Considerations | -5 | Threat model is a prose table with no concrete implementation details; no code or config references |
| PRD AC gap | Breakdown-Readiness | -3 | PRD 5.1 specifies pagination (20/page, lazy loading) but Service interface returns `[]model.DecisionLog` with no pagination parameters |
| Design inconsistency | Interface & Model Definitions | -3 | API handbook shows paginated list response (`items`, `total`, `page`, `size`) but service interface returns unpaginated `[]model.DecisionLog` |
| DTO naming bug | Interface & Model Definitions | -1 | `DecisionLogCreateReq.Status` field name contradicts JSON tag `logStatus`; inconsistent with model field name `LogStatus` |

## Attack Points

### Attack 1: Breakdown-Readiness -- Pagination design conflict blocks implementation

**Where**: tech-design.md line 77 -- `List(ctx context.Context, mainItemID uint, userID uint) ([]model.DecisionLog, error)` vs api-handbook.md line 80-84 -- paginated response with `items`, `total`, `page`, `size`

**Why it's weak**: The service interface signature returns a flat `[]model.DecisionLog` with no pagination parameters (no offset, no page, no size). Meanwhile the API handbook documents a paginated response envelope. The PRD (section 5.1) explicitly requires "20/page with lazy loading." A developer implementing from this design would have to guess: does the service return everything and the handler slices it? Does the repo accept pagination? Should `PageResult` from `dto` be used? The Cross-Layer Data Map (line 248) also omits any pagination fields. This gap makes it impossible to derive a clean repo method signature and handler implementation without asking clarifying questions.

**What must improve**: (1) Add `page` and `pageSize` (or `offset`/`limit`) parameters to the service `List` interface. (2) Change the return type to `(*dto.PageResult[model.DecisionLog], error)` or equivalent, matching the established `PageResult[T]` pattern in `dto/item_dto.go`. (3) Add pagination fields to the repo `ListByItem` interface. (4) Update the Cross-Layer Data Map to include pagination fields.

### Attack 2: Security Considerations -- Threat model is superficial with no implementation specificity

**Where**: tech-design.md lines 282-294 -- "Threat Model" and "Mitigations" sections

**Why it's weak**: The PRD's Story 4 is entirely about permission isolation and access control verification, making security a first-class requirement. The threat model lists four generic threats with one-sentence mitigations, but none reference specific code or middleware. For example:
- "Cross-team access" mitigation says "Team scope middleware + TeamKey in model" but does not specify whether the middleware is applied at the route group level or checked per-handler-method. The design does not show the route registration code that would apply this middleware.
- "Content injection" says "No HTML rendering; content displayed as text" -- this is a frontend concern with no reference to how React renders it (e.g., is `dangerouslySetInnerHTML` explicitly banned? Is there a sanitization library?).
- The design mentions `main_item:update` permission but does not show the route registration snippet that applies the permission middleware, leaving the developer to infer from existing patterns.
- No mention of input sanitization for the `tags` field, which accepts arbitrary strings and stores them as JSON. The design explicitly says "No backend validation on individual tag length -- frontend handles this" (line 182), which means a direct API call could inject tags of any length, bypassing the PRD's "each tag <= 20 chars" constraint (PRD section 5.3).

**What must improve**: (1) Add the route registration code snippet showing `middleware.Permission("main_item:update")` applied to the write routes. (2) Move tag length validation to the backend DTO `binding` tag instead of relying on frontend-only enforcement. (3) Specify how team-scoping is enforced at the route level (which middleware, on which route group). (4) For each threat, reference the specific line of code or middleware that mitigates it.

### Attack 3: Interface & Model Definitions -- DTO field naming inconsistency and Tags serialization mismatch

**Where**: tech-design.md lines 168-173 -- `DecisionLogCreateReq` struct, and line 182 -- Tags validation note

**Why it's weak**: Two concrete problems:
1. The `DecisionLogCreateReq` struct has a field named `Status` with JSON tag `"logStatus"`. The model field is `LogStatus`. Every other field is consistent between model/VO/DTO (`Category` -> `"category"`, `Content` -> `"content"`). This field should be named `LogStatus` to match, otherwise Go developers will be confused about why one field breaks the naming convention.
2. The DTO declares `Tags` as `string` type (`Tags string \`json:"tags"\``) with a note that it "arrives as JSON string from frontend (already serialized)." But the frontend type definition (line 200) declares `tags: string[]` -- an array, not a string. The API handbook request example (api-handbook.md line 36) shows `"tags": ["tag1", "tag2"]` -- an array. So the frontend sends a JSON array, but the backend DTO expects a JSON string. This is a serialization contract violation. Either the frontend must serialize tags to a string before sending (undocumented), or the backend DTO must accept `[]string` and serialize internally. The design does not resolve this ambiguity.

**What must improve**: (1) Rename `Status` to `LogStatus` in `DecisionLogCreateReq` for consistency. (2) Explicitly document the Tags serialization contract: does the frontend send `tags: "['a','b']"` (string) or `tags: ['a','b']` (array)? If string, show the frontend serialization code. If array, change the DTO type and document where JSON serialization to TEXT column happens (in the service or repo layer).

## Verdict
- **Score**: 72/100
- **Target**: 90/100
- **Breakdown-Readiness**: 13/20
- **Action**: Continue to iteration 2. Critical fixes needed: (1) resolve pagination design conflict across service/handler/repo/API handbook, (2) harden security section with implementation-specific mitigations and backend tag validation, (3) fix DTO field naming and Tags serialization contract.
