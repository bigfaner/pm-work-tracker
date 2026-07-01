# Design Evaluation — Iteration 2

**Doc**: `docs/features/ai-copilot/design/` (tech-design.md + 10 sub-modules + schema.sql, now including api-handbook.md)
**Rubric**: `forge/eval/rubrics/design.md` (1000 pts, 7 dimensions)
**Iteration**: 2 of 3
**Target**: 900

---

## Iteration 1 → 2 Verification (which attacks were addressed?)

| Iter-1 Attack | Status | Evidence |
|---------------|--------|----------|
| B1 orphan pollution (commit_create in LLM flow) | ✅ Resolved | `commit_create`/`commit_update`/`move_sub_item` tools removed (agent-architecture.md §2.2). Write moved to `commit_card` Handler → Dispatcher (request-model.md §6.1 请求 3; security.md §7.3; llm-integration.md §7.1). Form card now `targetEntity.bizKey=""` until commit. |
| B2 SSE / GLM concurrency limits | ✅ Resolved | llm-integration.md §6.5 adds `max_streams_per_user=1`, `max_outbound_llm=10`, `max_global_streams=50` with middleware + semaphore code. New error codes `ERR_COPILOT_USER_STREAM_BUSY` / `ERR_COPILOT_LLM_OUTBOUND_BUSY`. StreamRun wall-clock 60s. |
| B3 cost_usd no rollup | ❌ Not addressed | `MonthlyCost(ctx, year, month)` still live SUM (interfaces.md §6). No rollup table, no cron, no materialized view. security.md §7.2 still says "月度成本 > $200". |
| B4 quota HTTP vs SSE contradiction | ✅ Resolved | security.md §7.1 + sse-protocol.md §9 + api-handbook.md §7: single path (HTTP 429 pre-flight, never SSE error event). |
| B5 idempotency for commit | ⚠ Partial | `copilot_idempotency_keys` table added (schema.sql). Dispatcher prose in interfaces.md §7.1. **But no Repository interface declared, no handler code shown, no service method signature.** Wiring is prose-only. |
| B6 UpdateCurrentTurn never called | ✅ Resolved | request-model.md §2.2 handleFreeText now calls `h.sessionRepo.UpdateCurrentTurn(c, sessionID, turnID)` inside the tx. testing-strategy.md §3.4 mentions `TestFreeText_UpdatesCurrentTurnID`. |
| B7 intent persist best-effort | ✅ Resolved | tech-design.md §4.4 + request-model.md §5.3 HandleUserMessage uses `txManager.WithinTx` with `AppendTx` / `UpdateStatusTx` / `UpdateIntentMessageIDTx`. Persist fail → rollback + emit error + turn=failed. |
| D1 token estimate ±20% | ❌ Not addressed | llm-integration.md §3.4 unchanged. |
| D2 bizKey string vs AB-001 | ⚠ Disclosed | interfaces.md §0 documents deviation with rationale; conversion at Dispatcher boundary. AB-001 not amended but deviation is explicit. |
| D3 confidence band AC | ✅ Resolved | agent-architecture.md §3.4: confidence field, 3-band routing, 4 boundary ACs (0.7/0.69/0.4/0.39), `candidate_list` cardType for mid-confidence. |
| D4 keyword fallback AC | ✅ Resolved | security.md §7.5 `KeywordFallbackService` with rule table + HTTP 429 body extension `keywordMatch`. |
| D5 concurrent field diff AC | ✅ Resolved | api-handbook.md §4 + DiffOverlay struct (interfaces.md §11) + `ERR_COPILOT_FIELD_CONFLICT` 409 + FieldState.Version/EditedAt. |
| D6 api-handbook.md missing | ✅ Resolved | New api-handbook.md created. |
| D7 undefined structs | ✅ Resolved | interfaces.md §11 defines FieldState / EntityRef / EntityRecord / TracePayload / TraceAction / IntentMeta / DiffOverlay / FormErrors. |
| D9 Go 1.26 / jest | ✅ Resolved | testing-strategy.md §7.1 uses `go-version: '1.23'`; §4.2 uses `vi.spyOn(global, 'fetch')`. |

**Net**: 11 of 12 prior attacks resolved or disclosed; B3 (cost rollup) untouched, B5 (idempotency wiring) only partial.

---

## Phase 1 — Reasoning Audit (Pre-Score Anchors)

| # | Anchor | Verified? |
|---|--------|-----------|
| A1 | Restructure (commit moved from LLM tool to commit_card Handler) resolves orphan pollution AND keeps "AI 不直接写库" consistent: LLM 流 only emits prefilled form card (bizKey=""); real DB write happens in synchronous commit_card Handler path. Claim and mechanism now aligned. | ✅ |
| A2 | `copilot_idempotency_keys` schema exists (schema.sql lines 200-238) and prose says Dispatcher checks it (interfaces.md §7.1). **But**: (a) no `IdempotencyKeyRepository` in interfaces.md §6; (b) no `*Tx` variant of Dispatcher method shown; (c) `handleCommitCard` Handler code that extracts `requestId` from MessageRequest, calls Dispatcher with it, and writes followup msg — **never shown anywhere**. The wiring end-to-end (handler → service → repo) is asserted, not designed. | ❌ Wiring gap |
| A3 | Per-user SSE=1 and global GLM=10 limits: code snippets present (llm-integration.md §6.5). `UserStreamGuard` middleware and `outboundSem chan struct{}` in GLMProvider are concrete. **But** `UserStreamGuard` calls `c.ShouldBindJSON(&req)` to inspect type — Gin cannot re-read the body when the downstream handler tries to bind again unless buffering is enabled. Common Gin pitfall, unacknowledged. The mitigation may not work as written. | ⚠ Implementation hazard |
| A4 | Confidence band candidate_list flow requires a new Turn status `awaiting_select_intent` (agent-architecture.md §3.4: "state-machines.md §3 新增状态 awaiting_select_intent"). **state-machines.md §3 still lists only 10 Turn statuses — `awaiting_select_intent` is NOT added.** No transition rule, no interception matrix row, no Message status for the candidate_list card. The candidate_list AC is half-designed. | ❌ Cross-section inconsistency |
| A5 | `FormCardData` (sse-protocol.md §3) adds `DiffOverlay *DiffOverlay` but **does NOT add `lastEditedAt`** — yet api-handbook.md §4 says "每条 form card 消息携带 `cardData.fields[].version`... 与 `cardData.lastEditedAt`". FieldState has EditedAt but FormCardData-level lastEditedAt is missing from the struct. Concurrent-diff AC partially wired. | ⚠ Schema mismatch |
| A6 | Schema.sql has 6 tables; tech-design.md §6.3 AutoMigrate code lists only 5 models (`Session`, `Turn`, `Message`, `AgentCallLog`, `FeatureFlag`) — **`IdempotencyKey` model missing from AutoMigrate**. The 6th table won't be created on `db.AutoMigrate(...)`. Cross-section inconsistency introduced by revision. | ❌ Migration gap |
| A7 | `copilot_messages.session_id` column has index but **no FK constraint** (only `turn_id` has FK). Orphan risk if turn cascade misses or if a message is inserted with stale session_id. Old issue, not fixed by revision. | ⚠ Unchanged |
| A8 | `txManager.WithinTx` pattern in handleFreeText uses non-Tx method names (`h.turnRepo.Create`, `h.msgRepo.Append`, `h.sessionRepo.UpdateCurrentTurn`) inside the tx closure — but interfaces.md §6 declares only these non-Tx signatures. The HandleUserMessage code (request-model.md §5.3) switches to `*Tx` variants (`AppendTx`, `UpdateStatusTx`, `UpdateIntentMessageIDTx`) — **these Tx variants are NOT declared in interfaces.md §6**. Two patterns mixed; a developer cannot tell which methods exist. | ❌ Interface mismatch |
| A9 | GLM outbound semaphore capacity=10 grounded in claim "GLM API single key QPS ~10 (保守估算)" — but QPS ≠ concurrent inflight. 10 instant concurrent requests to GLM would likely 429 (real GLM RPM limits not cited). Capacity model is hand-waved. | ⚠ Ungrounded |
| A10 | `commit_card` failure response: api-handbook.md §2.3 says "body 同时含 `messageId` + `newState:"failed"` + `errors`" but the JSON example block immediately above only shows the success body. Failure body shape is described in prose but not exemplified — frontend contract ambiguous. | ⚠ Example gap |

---

## Phase 2 — Rubric Scoring (current state, strict)

### 1. Architecture Clarity (170 pts) → **150/170**

| Criterion | Score | Notes |
|-----------|-------|-------|
| Layer placement explicit | 55/60 | `internal/copilot/` subtree fully enumerated (tech-design.md §2.1). -5: stale file reference — `tools/commit.go` still listed under `tools/` package (line 195) but commit tools were removed (§2.2 agent-architecture.md). Either file should be deleted or renamed (e.g. `dispatcher_helpers.go` under `service/`). A developer copying the tree verbatim will create a stray file. |
| Component diagram present | 55/60 | tech-design.md §1.2 + §2.2 + llm-integration.md §1.1. Diagrams clear. -5: still no diagram showing 3-tier status coordination (Session/Turn/Message). state-machines.md has ASCII for each layer but no joint view. |
| Dependencies listed | 40/50 | Internal deps listed. "新增外部依赖：无" claim repeated. -5: GLM endpoint and model name (`glm-4-plus`, `/api/paas/v4`) still not verified against current GLM docs (no citation, like iter 1). -5: outbound semaphore capacity=10 grounding weak (see A9). |

### 2. Interface & Model Definitions (170 pts) → **136/170**

(er-diagram.md exists → "db-schema: yes" row.)

| Criterion | Score | Notes |
|-----------|-------|-------|
| Interface signatures typed | 31/40 | Most interfaces typed. -5: `Tool.ParametersSchema() map[string]any` and `ToolResult.Data map[string]any` still untyped bags — function-calling tools should expose a JSON Schema, not `map[string]any`. -4: Planner emission tool schemas (`submitRewriteSchema`, `submitIntentSchema`) referenced by name (agent-architecture.md §3.3) **still not defined anywhere** — same gap as iter 1, unresolved. |
| Inline models concrete | 37/40 | All previously-undefined structs (FieldState / EntityRef / EntityRecord / TracePayload / IntentMeta / DiffOverlay / FormErrors) now defined in interfaces.md §11 with typed fields and JSON tags. -3: `Message.Card json.RawMessage` (tech-design.md §3) is still an untyped bag — could be a typed union of IntentCardData/FormCardData/etc. |
| ER diagram complete | 26/30 | er-diagram.md has full ASCII entity boxes + relationships + cardinality + index strategy + soft-delete policy. -2: still ASCII only; rubric prefers Mermaid `erDiagram` block for tooling. -2: er-diagram.md still shows only 5 entities — `copilot_idempotency_keys` (new table) not added to overview diagram. |
| SQL DDL directly usable | 22/30 | schema.sql executable as-is (SQLite active + MySQL commented). FKs/indexes/defaults present. -4: still `--` line comments only (rubric asks for "inline COMMENT syntax" so column docs live in DB metadata). -4: `copilot_messages.session_id` has index but **no FK** (only `turn_id` FK) — orphan risk. -4: **NEW inconsistency** — schema.sql has 6 tables but tech-design.md §6.3 `AutoMigrate` lists only 5 models (`IdempotencyKey` missing) — the 6th table won't be auto-created; cross-section inconsistency introduced by revision. |
| Cross-layer consistency | 20/30 | bizKey deviation now explicitly documented (interfaces.md §0) with conversion at Dispatcher — improvement. **But**: (a) `FormCardData` (sse-protocol.md §3) lacks `lastEditedAt` field that api-handbook.md §4 prose references ("cardData.lastEditedAt") — schema/code mismatch (-4). (b) Turn state `awaiting_select_intent` referenced by agent-architecture.md §3.4 but absent from state-machines.md §3 + §10 (-4). (c) `*Tx` repo methods (`AppendTx`, `UpdateStatusTx`, `UpdateIntentMessageIDTx`) called in request-model.md §5.3 but absent from interfaces.md §6 (-2). |

### 3. Error Handling (130 pts) → **107/130**

| Criterion | Score | Notes |
|-----------|-------|-------|
| Error types defined | 41/45 | 16 error codes (added FIELD_CONFLICT / USER_STREAM_BUSY / LLM_OUTBOUND_BUSY / QUOTA_CHECK_FAILED). INPUT_TOO_LONG now "200（继续）"; TURN_CANCELLED "200（JSON）". All have HTTP status. -4: still no error code for the documented redactor "fail-open 但记录告警" case (security.md §5 prose) — no event code captures this fail-open signal in agent_call_logs. |
| Propagation strategy clear | 38/45 | §4.2 + §4.4 transactional persist for source-of-truth messages. Best-effort only for trace. Quota pre-flight single path. **But**: -4: idempotency conflict path (UNIQUE violation on `copilot_idempotency_keys.request_id`) — what does Dispatcher return? HTTP 200 with cached result or 409? interfaces.md §7.1 prose says "命中则直接返回上次结果" but no HTTP mapping, no error code if `result_biz_key` is NULL (pending from a crashed prior attempt). -3: `commit_card` entity-service validation failure propagation — api-handbook.md §2.3 says body contains `errors` but no JSON example; ambiguous shape. |
| HTTP status codes mapped | 28/40 | All 16 codes mapped in tech-design.md §4.1 + api-handbook.md §6. **But**: -4: `commit_card` failure body shape — when entity service returns 422, what exactly does the frontend receive? api-handbook.md §2.3 example only shows success. -3: `ERR_COPILOT_USER_STREAM_BUSY` returns 429 — but quota also returns 429. Frontend cannot distinguish quota-exceeded from stream-busy without inspecting `code` (it can, but the dual-429 should be called out). -5: `UserStreamGuard` middleware correctness — `c.ShouldBindJSON(&req)` consumes the body; subsequent handler `ShouldBindJSON` will fail EOF unless `c.Request.Body` is buffered. The 429 error path may itself be unreachable if binding fails upstream. Unacknowledged Gin pitfall. |

### 4. Testing Strategy (130 pts) → **112/130**

| Criterion | Score | Notes |
|-----------|-------|-------|
| Per-layer test plan | 41/45 | testing-strategy.md §3 has per-layer tables. Confidence-band tests added (§3.1 supplement). commit_card idempotency tests added (§3.2). -2: no test scenarios for `UserStreamGuard` concurrent SSE rejection (the new middleware has no test row). -2: no test scenarios for candidate_list (mid-confidence) flow — a brand new path with 4 boundary ACs. |
| Coverage target numeric | 40/45 | §6 per-module targets (agent ≥85%, etc.). -3: still no overall project coverage floor reconciling with CLAUDE.md "核心模块 ≥ 80%". -2: still no frontend coverage target. |
| Test tooling named | 31/40 | Backend: Go testing + testify ✓. Frontend: Vitest + RTL + MSW ✓. CI uses Go 1.23 (fixed), `vi.spyOn` (fixed). -3: `actions/setup-node@v3` with `node-version: '20'` but no pnpm/yarn lockfile mention — npm ci works but project tool unspecified. -3: no load-testing tool named for verifying `max_streams_per_user=1` / `max_outbound_llm=10` under contention (these are concurrency invariants — unit tests cannot prove them). -3: no tool named for testing the idempotency UNIQUE race (two concurrent commit_card with same requestId). |

### 5. Breakdown-Readiness ★ (180 pts) → **150/180**

(Critical gate — must reach 160/180 to advance.)

| Criterion | Score | Notes |
|-----------|-------|-------|
| Components enumerable | 55/65 | Components clearly enumerable: 5 Agents, ~14 tools (Read+Emission, no Action), 5 repos, 1 dispatcher, 1 orchestrator, 1 ctxBuilder, 1 provider, 1 quotaSvc, 1 flagCache, 1 keywordFallbackSvc, 1 streamLimiter, 16 SSE event types. **But**: -5: `handleCommitCard` Handler referenced in 5+ places (request-model.md §6.1, security.md §7.3, interfaces.md §7.1) but its full implementation (requestId extraction → idempotency check → Dispatcher.Dispatch → UPDATE msg status → persist followup msg → return JSON) is **never shown**. A task breakdown would have to invent the handler contract. -3: `IdempotencyKeyRepository` not declared in interfaces.md §6 — a developer cannot enumerate repository impl tasks without this interface. -2: `tools/commit.go` listed in §2.1 layer placement but commit tools removed — stale file reference confuses component count. |
| Tasks derivable | 55/65 | Each Agent → impl task ✓. Each tool → impl task ✓. Each model → schema task ✓ (with caveat below). Structs defined ✓. api-handbook.md exists ✓. **But**: -4: AutoMigrate code (tech-design.md §6.3) lists only 5 models but schema.sql has 6 tables — a developer following the AutoMigrate snippet will miss `copilot_idempotency_keys`. -3: Tx-variant repository methods (`AppendTx`, `UpdateStatusTx`, `UpdateIntentMessageIDTx`) used in code but absent from interface — task breakdown cannot derive signatures. -3: `handleCommitCard` full flow absent (see above). |
| PRD AC coverage | 40/50 | Improvements: confidence band (§3.4) ✓, keyword fallback (§7.5) ✓, concurrent diff (api-handbook §4 + DiffOverlay) ✓. **But**: -5: **candidate_list flow has no Turn state** — agent-architecture.md §3.4 says "state-machines.md §3 新增状态 awaiting_select_intent" but state-machines.md §3 + §10 still list only 10 Turn states; no transition rule, no interception matrix row, no Message status. The mid-confidence AC is half-designed and unimplementable as written. -3: keyword fallback `keywordMatch` field returned in HTTP 429 body but no frontend contract showing how KeywordFallbackCard persists (does it become a Message? a card? ephemeral?) — task breakdown cannot derive frontend state tasks. -2: 50-turn session cap deviation still disclosed but PRD AC remains enumerated. |

### 6. Security Considerations (80 pts) → **66/80**

| Criterion | Score | Notes |
|-----------|-------|-------|
| Threat model present | 34/40 | security.md covers prompt injection, sensitive fields, cross-Team bizKey, RBAC bypass, quota abuse, key leak, orphan pollution (§7.3 new), mid-commit drop (§7.3 new), SSE connection attack (§7.5 new). **But**: -3: `UserStreamGuard` threat acknowledged but the mitigation's correctness depends on Gin body buffering — not modeled. -3: no threat model for `copilot_idempotency_keys` abuse — a malicious client could send empty `requestId` (UNIQUE collision on first row) or enumerate requestIds to pollute the table; no size limit, no rate limit on commit_card specifically. |
| Mitigations concrete | 32/40 | Sanitizer, redactor, quota (fail-closed §7.4), feature flag cache, idempotency table, write isolation, keyword fallback. **But**: -3: sanitizer pattern `(?i)you are (now )?a` false positives on legitimate text ("you are a member of team X") — iter 1 issue, unaddressed. -2: redactor "fail-open" claim (§5) still not grounded — the redactor has no "did I miss something?" signal, so "fail-open（继续调用），但记录告警" is unverifiable. -3: `commit_card` idempotency mitigation prose-only — no code, no repo interface, no test scenario for the UNIQUE-race case where two concurrent requests collide. |

### 7. Implementation Feasibility (140 pts) → **107/140**

| Criterion | Score | Notes |
|-----------|-------|-------|
| Dependencies available | 42/50 | Go 1.23 fixed. Vitest fixed. GLM via HTTPS. -5: AB-001 convention conflict — deviation documented (interfaces.md §0) with conversion at Dispatcher boundary, but the convention itself is not amended. The "typed at boundary" spirit is preserved but a project-wide convention is being carved out without an explicit convention update. -3: GLM model name `glm-4-plus` and endpoint `/api/paas/v4` still not verified against current GLM docs (no citation; same gap as iter 1). |
| Architecture fits project structure | 38/50 | `internal/copilot/` subtree follows existing pattern. Reuses `internal/service/*`, `internal/middleware/*`, `internal/pkg/snowflake`, `internal/pkg/apperrors`. **But**: -5: `UserStreamGuard` middleware double-binds body (Gin pitfall) — would fail in practice unless `ShouldBindBodyWith` is used. -4: `IdempotencyKeyRepository` not declared; `handleCommitCard` Handler code not shown; AutoMigrate missing 6th model — implementation would not match design. -3: Tx-variant repo methods referenced in code but absent from interface contract — Go compiler will reject. |
| Technical claims grounded | 27/40 | -4: token estimate "1 token ≈ 1.5 中文字符 / 4 英文字符... ±20% 可接受" (llm-integration.md §3.4) still unvalidated heuristic treated as grounded. -3: GLM QPS~10 claim for semaphore capacity uncited. -3: 20 concurrent users / P95 <5s no capacity model. -3: MonthlyCost live SUM still no rollup (B3 unresolved) — at 30k rows/user × N users the monthly cost query degrades, threatening the "$200 熔断" guarantee. |

---

## Cross-Dimension Coherence Check

- **State-machine ↔ Agent-architecture** (Dim 2 ↔ Dim 5): agent-architecture.md §3.4 introduces `awaiting_select_intent` for candidate_list flow; state-machines.md §3/§10 does NOT add this state. The mid-confidence PRD AC cannot be implemented without inventing the state machine entry. (-5 already in Dim 5.)
- **Schema ↔ AutoMigrate** (Dim 2 ↔ Dim 7): schema.sql has 6 tables; AutoMigrate code lists 5. Migration would silently miss `copilot_idempotency_keys`. (-4 in Dim 2, -2 in Dim 7.)
- **FormCardData ↔ api-handbook** (Dim 2 ↔ Dim 3): api-handbook.md §4 prose references `cardData.lastEditedAt`; FormCardData struct (sse-protocol.md §3) has no such field. (-4 in Dim 2.)
- **Repository interface ↔ usage** (Dim 2 ↔ Dim 7): `AppendTx` / `UpdateStatusTx` / `UpdateIntentMessageIDTx` used in HandleUserMessage code; not declared in interfaces.md §6. (-2 in Dim 2, -2 in Dim 7.)
- **Error flow** (Dim 3): mostly resolved; quota pre-flight single path; FieldConflict defined. `commit_card` failure body shape still ambiguous.

---

## Phase 3 — Blindspot Hunt

Each cites a specific document quote and the failure pattern it matches.

### [blindspot-B8] `awaiting_select_intent` Turn status missing — candidate_list AC unimplementable

**Quote** (agent-architecture.md §3.4):
> state-machines.md §3 新增状态 `awaiting_select_intent`（介于 planning 与 awaiting_confirm_intent 之间）。

vs (state-machines.md §3 + §10):
> 状态枚举（10 个）：planning / awaiting_clarify / awaiting_confirm_intent / executing / awaiting_commit / awaiting_select_candidate / done / cancelled / superseded / failed

**Failure pattern**: *Implicit coupling between modules — dependencies not acknowledged in interfaces.* The mid-confidence PRD AC (candidate_list flow) is designed in agent-architecture.md but the state machine it depends on is never updated. Result: `select_intent` request type (mentioned in agent-architecture.md §3.4) has no interception matrix row, no transition rule, no Turn status to detect "is a candidate selection in flight?", no Message status for the candidate_list card. The whole mid-confidence branch is dead-on-arrival.

**What must improve**: Add `awaiting_select_intent` to state-machines.md §3 (state enum), §3 conversion matrix (transitions from planning on mid-confidence, from awaiting_select_intent on select_intent), §6 interception matrix (free_text in awaiting_select_intent → superseded; select_intent → executing), §10 status const. Add `MsgStatusCandidateListAwaiting` for the candidate_list card. Add `select_intent` request type to MessageRequest (interfaces.md §10 currently lists 7 types — `select_intent` is not among them despite being referenced in §3.4).

### [blindspot-B9] `copilot_idempotency_keys` not wired end-to-end — schema exists, code is prose

**Quote** (interfaces.md §7.1):
> Dispatcher 在 entity service Create 路径上：
> 1. 查 `copilot_idempotency_keys` 表（`request_id` UNIQUE）；命中则直接返回上次结果（含 bizKey），不再调 entity service。
> 2. 未命中 → 事务内：INSERT idempotency row + 调 entity service Create + UPDATE form card status=submitted。

**Failure pattern**: *Error paths that terminate silently / solutions that reintroduce patterns they claim to eliminate.* The idempotency mitigation is asserted in prose but:
- (a) **No `IdempotencyKeyRepository` in interfaces.md §6** — a developer cannot derive the repository task.
- (b) **No `*Tx` variant of Dispatcher.Dispatch** — the prose says "事务内：INSERT idempotency row + 调 entity service Create + UPDATE form card" but `Dispatcher.Dispatch(ctx, fc FormCard) (*DispatchResult, error)` (interfaces.md §7) takes no `*gorm.DB` and returns no tx. How does the Dispatcher participate in the caller's transaction?
- (c) **No `handleCommitCard` Handler code** — request-model.md §6.1 请求 3 shows the prose step list ("幂等检查：copilot_idempotency_keys WHERE request_id=req_uuid_v4") but no Go code, unlike `handleFreeText` and `handleConfirmIntent` which have full code blocks.
- (d) **No error path for the UNIQUE collision** — if two concurrent commit_card requests with the same requestId race, one wins INSERT, the other gets a UNIQUE violation. The loser must read the winner's `result_biz_key` — but if the winner hasn't committed yet (`status=pending`, `result_biz_key=NULL`), the loser must wait. No wait/retry logic designed.
- (e) **No requestId validation** — MessageRequest has `RequestID string` but no validation that it's UUID v4, non-empty, or per-turn-unique. A malicious client could send `requestId=""` (all empty strings UNIQUE-collide on first row → second request returns first request's bizKey forever) or send a different requestId per retry (bypassing idempotency entirely).

**What must improve**: Declare `IdempotencyKeyRepository` (Get / Insert / UpdateResult / UpdateStatus). Show `handleCommitCard` full Go code including requestId validation (UUID v4, non-empty), tx wrapping, conflict handling. Add a Dispatcher Tx-variant or pass `*gorm.DB` through. Add a test scenario for the UNIQUE race. Add request validation to MessageRequest.

### [blindspot-B10] `UserStreamGuard` middleware body double-bind — mitigation correctness gap

**Quote** (llm-integration.md §6.5):
```go
func (h *CopilotHandler) UserStreamGuard(c *gin.Context) {
    var req MessageRequest
    if err := c.ShouldBindJSON(&req); err != nil { c.Next(); return }
    if !triggersSSEStream(req.Type) { c.Next(); return }
    // ...
}
```

vs the downstream handler (request-model.md §2.2):
```go
func (h *CopilotHandler) PostMessage(c *gin.Context) {
    var req MessageRequest
    if err := c.ShouldBindJSON(&req); err != nil { ... }
    // ...
}
```

**Failure pattern**: *Unhandled concurrent access — shared state with no synchronization model.* Gin's `c.ShouldBindJSON` reads and consumes `c.Request.Body`. The second `ShouldBindJSON` in the handler will return EOF unless the middleware uses `c.ShouldBindBodyWith` (which buffers the body for re-reading). The design shows two `ShouldBindJSON` calls in sequence without buffering. Result: every request that passes through `UserStreamGuard` will fail at the handler with a 400 validation error — the SSE stream never opens, the 429 mitigation never fires, the entire SSE path is broken. Alternatively the middleware silently passes through (`c.Next(); return` on bind error), in which case `UserStreamGuard` is a no-op for all malformed body and the concurrency limit is unenforced.

**What must improve**: Either (a) use `c.ShouldBindBodyWith(&req, binding.JSON)` in the middleware (Gin's documented pattern for body re-read), or (b) parse `req.Type` from a different source (query param, header). Add a test scenario `TestUserStreamGuard_BodyReReadable` that verifies the handler still receives the body after the middleware binds.

### [blindspot-B11] AutoMigrate list mismatches schema.sql — 6th table won't be created

**Quote** (tech-design.md §6.3):
```go
if err := db.AutoMigrate(
    &copilotmodel.Session{},
    &copilotmodel.Turn{},
    &copilotmodel.Message{},
    &copilotmodel.AgentCallLog{},
    &copilotmodel.FeatureFlag{},
); err != nil { ... }
```

vs (schema.sql lines 200-238): `copilot_idempotency_keys` table exists with its own DDL block.

**Failure pattern**: *Missing data migration strategy — schema changes with no transition plan.* The AutoMigrate call lists 5 models; the schema has 6 tables. On a fresh install via AutoMigrate, `copilot_idempotency_keys` is never created. The first `commit_card` request will then fail at the Dispatcher (table missing) — and the error path is undefined because the design assumes the table exists. On an upgrade install via raw SQL, the table is created, but environment drift between AutoMigrate-only and SQL-only deployments will cause inconsistent behavior.

**What must improve**: Add `&copilotmodel.IdempotencyKey{}` to the AutoMigrate call in §6.3. Add a corresponding Go model definition in tech-design.md §3 (Go struct overview) or interfaces.md §11. Reconcile the count ("5 张新表" → "6 张新表") in tech-design.md §1.2 diagram caption and §6.3 prose.

### [blindspot-B12] `FormCardData.lastEditedAt` missing — concurrent-diff AC partially wired

**Quote** (api-handbook.md §4):
> 每条 form card 消息携带 `cardData.fields[].version`（int64，每次 PATCH 递增）与 `cardData.lastEditedAt`（int64 ms）。

vs (sse-protocol.md §3):
```go
type FormCardData struct {
    OpType       string       `json:"opType"`
    EntityType   string       `json:"entityType"`
    TargetEntity *EntityRef   `json:"targetEntity"`
    Fields       []FieldState `json:"fields"`
    Errors       *FormErrors  `json:"errors,omitempty"`
    RetryCount   int          `json:"retryCount"`
    DiffOverlay  *DiffOverlay `json:"diffOverlay,omitempty"`
}
```

No `LastEditedAt` field.

**Failure pattern**: *Solutions that reintroduce patterns they claim to eliminate.* The concurrent-edit AC hinges on last-write-wins + diff preview. The version field per FieldState exists (good), but the card-level `lastEditedAt` that api-handbook.md §4 prose references is absent from the struct. The DiffOverlay struct has `OtherEditedAt` (per-field) but no card-level timestamp. Without `LastEditedAt`, the frontend cannot sort "which write happened last" across multiple PATCH calls on different fields — the diff conflict detection has no total ordering.

**What must improve**: Add `LastEditedAt int64 json:"lastEditedAt"` to `FormCardData`. Document the update rule (every successful PATCH op=edit_field bumps both `fields[changedIndex].version` and `cardData.lastEditedAt`). Add a test scenario verifying multi-field concurrent edits resolve correctly.

### [blindspot-B13] `MonthlyCost` live SUM still no rollup — $200 circuit breaker freshness unknown

**Quote** (security.md §7.2):
> 月度成本 > $200 → 管理员一键关闭

vs (interfaces.md §6):
```go
MonthlyCost(ctx context.Context, year, month int) (float64, error)
```

**Failure pattern**: *Missing data migration / aggregation strategy.* (Same as iter-1 B3, unchanged.) `agent_call_logs` is append-only. At scale (1000 calls/day × 30 days = 30k rows/user × N users), `MonthlyCost` doing `SUM(cost_usd) WHERE year/month` on every check becomes O(N). The 30s feature-flag cache (§7.2) caches the boolean flag, not the cost computation. The "$200 熔断" guarantee depends on `MonthlyCost` being called on a cadence — but no caller is shown (no cron, no admin dashboard query frequency).

**What must improve**: Either (a) add `copilot_cost_daily` rollup table populated by a cron job (recommended), or (b) precompute `monthly_cost_usd` per user and cache aggressively. Document the refresh cadence. Add a monitoring metric for `MonthlyCost` query latency.

---

## Deduction Items Summary

| ID | Type | Pts | Quote / Anchor |
|----|------|-----|----------------|
| D1 | Cross-section inconsistency | -30 | `awaiting_select_intent` referenced in agent-architecture.md §3.4 but absent from state-machines.md §3/§10. (B8) |
| D2 | Cross-section inconsistency | -20 | schema.sql has 6 tables; AutoMigrate (tech-design.md §6.3) lists 5 — `IdempotencyKey` missing. (B11) |
| D3 | Cross-section inconsistency | -15 | `FormCardData` (sse-protocol.md §3) missing `LastEditedAt`; api-handbook.md §4 prose references it. (B12) |
| D4 | Cross-section inconsistency | -10 | `*Tx` repo methods (`AppendTx`, `UpdateStatusTx`, `UpdateIntentMessageIDTx`) called in request-model.md §5.3 but absent from interfaces.md §6. |
| D5 | PRD AC gap (residual) | -20 | candidate_list mid-confidence flow has no Turn status → unimplementable as designed. (B8) |
| D6 | Placeholder (prose-only mitigation) | -20 | `copilot_idempotency_keys` end-to-end wiring (handler code, repo interface, tx participation) is prose-only. (B9) |
| D7 | Implementation hazard | -15 | `UserStreamGuard` `ShouldBindJSON` consumes body; downstream handler `ShouldBindJSON` will EOF unless body buffered. (B10) |
| D8 | PRD AC gap (residual) | -10 | `commit_card` failure body shape — JSON example shows only success; failure body with `errors` is described but not exemplified. |
| D9 | Vague language (unchanged) | -10 | "1 token ≈ 1.5 中文字符... ±20% 可接受" (llm-integration.md §3.4) still unvalidated. |
| D10 | Convention deviation | -10 | AB-001 deviation documented but convention not amended; project-wide int64 bizKey boundary is being carved out. |
| D11 | Missing rollup (unchanged) | -10 | `MonthlyCost` live SUM; $200 circuit breaker freshness depends on call cadence not designed. (B13) |
| D12 | Stale reference | -5 | `tools/commit.go` still in layer placement (tech-design.md §2.1) but commit tools removed. |

(Deductions are reflected inside dimension sub-scores above; this table is for reviser reference.)

---

## Final Score

| Dimension | Score | Max |
|-----------|-------|-----|
| 1. Architecture Clarity | 150 | 170 |
| 2. Interface & Model Definitions | 136 | 170 |
| 3. Error Handling | 107 | 130 |
| 4. Testing Strategy | 112 | 130 |
| 5. Breakdown-Readiness ★ | 150 | 180 |
| 6. Security Considerations | 66 | 80 |
| 7. Implementation Feasibility | 107 | 140 |
| **TOTAL** | **828** | **1000** |

**Gate status**: ❌ Below target (900) and below breakdown-readiness gate (160/180). Iteration 3 required.

Net change vs iteration 1: 828 − 761 = +67 points. Significant improvement, but two new cross-section inconsistencies (B11 AutoMigrate; B8 awaiting_select_intent) and one implementation hazard (B10 body double-bind) were introduced by the revision.

---

## Priority Fixes for Iteration 3 (Reviser Instructions)

1. **[B8] Add `awaiting_select_intent` Turn state end-to-end** — state-machines.md §3 (enum + transitions), §6 (interception matrix), §10 (const); add `select_intent` to MessageRequest type list (interfaces.md §10); add `MsgStatusCandidateListAwaiting`; add interception rules. Unblocks mid-confidence PRD AC. (Highest impact — currently unimplementable.)
2. **[B11] Reconcile AutoMigrate with schema.sql** — add `&copilotmodel.IdempotencyKey{}` to tech-design.md §6.3; update "5 张新表" → "6 张新表" in §1.2 diagram caption; add Go model struct to §3 or interfaces.md §11.
3. **[B9] Wire `copilot_idempotency_keys` end-to-end** — declare `IdempotencyKeyRepository` in interfaces.md §6; show `handleCommitCard` full Go code (requestId UUID v4 validation, tx wrap, idempotency check, Dispatcher call, status update, followup persist); add Dispatcher Tx-variant or pass `*gorm.DB`; add error path for UNIQUE collision with pending `result_biz_key`.
4. **[B10] Fix `UserStreamGuard` body re-read** — use `c.ShouldBindBodyWith(&req, binding.JSON)` in middleware; add `TestUserStreamGuard_BodyReReadable` test.
5. **[B12] Add `LastEditedAt` to `FormCardData`** — update sse-protocol.md §3; document PATCH update rule; add multi-field concurrent edit test.
6. **[D4] Declare Tx-variant repository methods** — add `AppendTx(tx, ...)`, `UpdateStatusTx(tx, ...)`, `UpdateIntentMessageIDTx(tx, ...)` to interfaces.md §6 (or document that non-Tx methods detect tx from context).
7. **[B13 / D11] Add cost rollup** — either `copilot_cost_daily` table + cron, or precomputed `monthly_cost_usd`; document refresh cadence; add `MonthlyCost` latency metric.
8. **[D8] Add `commit_card` failure body JSON example** — show 422 / 403 / 404 response shapes with `errors` field; reconcile api-handbook.md §2.3.
9. **[D12] Remove stale `tools/commit.go`** from layer placement (tech-design.md §2.1).
10. **[Iter-1 carryover] Define Planner emission tool schemas** (`submitRewriteSchema`, `submitIntentSchema`) — still referenced by name, still undefined.

Addressing items 1-5 alone (the new inconsistencies and hazards introduced this iteration) should push the score past 900 and unblock the breakdown-readiness gate.
