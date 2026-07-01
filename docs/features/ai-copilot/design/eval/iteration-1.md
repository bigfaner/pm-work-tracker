# Design Evaluation — Iteration 1

**Doc**: `docs/features/ai-copilot/design/` (tech-design.md + 9 sub-modules + schema.sql)
**Rubric**: `forge/eval/rubrics/design.md` (1000 pts, 7 dimensions)
**Iteration**: 1 of 3
**Target**: 900

---

## Phase 1 — Reasoning Audit (Pre-Score Anchors)

| # | Anchor (independent judgment) | Verified? |
|---|-------------------------------|-----------|
| A1 | Problem → Solution: design covers 6 entities × 4 ops via Planner + 4 Executors (writer/reader/updater/mover). Maps cleanly to PRD scope. | ✅ |
| A2 | Claim "AI 不直接写库" — `commit_create` Action tool is invoked by Writer Executor *during execute phase*, persisting a real entity (e.g. `MainItemService.Create` returns `bizKey=MI-0023`) **before** the user clicks 提交 on the form card (request-model.md §6.1). PRD Flow step 7 requires write API to be called only *after* user submits. The "form card" in this design is a *post-creation confirmation card*, not the "预填表单卡片" the PRD describes. | ❌ Major contradiction |
| A3 | Claim "Schema 单一来源" — references `entity-schemas.md` but `SchemaLoader` interface signature is never shown (only `schemaLoader.Load(entityType)` in prose, interfaces.md §4-5). | ⚠ Partial |
| A4 | Claim "完全无状态" — TurnContext is goroutine-scoped, but state is rebuilt from `copilot_turns.confirmed_intent` JSON. Verified. | ✅ |
| A5 | Claim "可熔断" — `feature_flags` 30s cache + `agent_call_logs` cost rollup. Cost rollup query (`MonthlyCost`) is described but no cron job or materialized rollup is in schema. | ⚠ Partial |
| A6 | PRD AC: confidence band (≥0.7 / 0.4-0.7 / <0.4) handling. **Not mentioned anywhere** in design. | ❌ AC gap |
| A7 | PRD AC: quota-exceeded → "降级为关键词匹配模式". Design returns `ERR_COPILOT_QUOTA_EXCEEDED` SSE error event (`fallbackAction: use_form`); keyword-matching mode is named once in security.md L177 but never designed. | ❌ AC gap |
| A8 | PRD AC: 50-turn session cap. Design explicitly deviates ("取消，改为 token 预算"). Deviation disclosed in tech-design.md §1.5, but PRD Story 7 still enumerates the 50-turn behavior as an AC. | ⚠ Deviation noted, but AC still enumerated in PRD |
| A9 | PRD AC: concurrent field edit (`last-write-wins + diff展示`). Design has PATCH /messages/:id but no concurrency model, no diff workflow. | ❌ AC gap |
| A10 | Required doc `api-handbook.md` referenced in tech-design.md L34 but file does **not exist** in design dir. | ❌ Missing deliverable |
| A11 | Convention AB-001 declares `teamBizKey int64` at Service boundary; design uses `bizKey string` (VARCHAR(36)) throughout (`GetByBizKey(ctx, bizKey string)`, schema `biz_key VARCHAR(36)`). Project-wide convention conflict. | ❌ Convention violation |
| A12 | Self-contradiction in error handling: tech-design.md §4.1 lists `ERR_COPILOT_INPUT_TOO_LONG` with HTTP Status "—" (no code) yet §1.5/Security requires truncation + continue. The error table row is dead. | ⚠ Internal |

---

## Phase 2 — Rubric Scoring

### 1. Architecture Clarity (170 pts) → **150/170**

| Criterion | Score | Notes |
|-----------|-------|-------|
| Layer placement explicit | 58/60 | `internal/copilot/` subtree fully enumerated (tech-design.md §2.1, lines 145-197). Each sub-package职责清晰 (handler/orchestrator/agent/prompt/provider/sse/model/repository/service/tools). Existing untouched code listed explicitly. Only -2: `tools/` package placement is described but its dependency on `service/` (Dispatcher) creates a circular risk (service/dispatcher.go uses FormCard; tools/commit.go uses Dispatcher) — not acknowledged. |
| Component diagram present | 55/60 | tech-design.md §1.2 (high-level ASCII) + §2.2 (call-flow ASCII) + llm-integration.md §1.1 (end-to-end ASCII). Diagrams are clear. -5: no diagram showing the 3-tier status coordination (Session/Turn/Message) — only state-machines.md prose. -0 for agent-architecture.md §2 tool matrix (excellent). |
| Dependencies listed | 37/50 | tech-design.md §2.3 lists internal deps + claims "新增外部依赖：无（GLM 通过标准 HTTPS 调用，不引入 SDK）". But: (a) testing-strategy.md §7.1 uses `actions/setup-go@v4` with `go-version: '1.26'` — Go 1.26 does not exist (latest stable is 1.23 as of design date); (b) frontend testing imports `jest.spyOn` but project uses Vitest (per CLAUDE.md) — jest is not in deps; (c) `regexp` Go stdlib usage in sanitizer is fine, but no version pin. -13 for these unchecked/incorrect dependency claims. |

### 2. Interface & Model Definitions (170 pts) → **127/170**

(er-diagram.md exists → use "db-schema: yes" row.)

| Criterion | Score | Notes |
|-----------|-------|-------|
| Interface signatures typed | 32/40 | Most interfaces typed (interfaces.md §1-10). -8: `Tool.ParametersSchema() map[string]any` and `ToolResult.Data map[string]any` are untyped bags — function-calling tools need a JSON Schema, not `map[string]any`. The Planner emission tool schemas (`submitRewriteSchema`, `submitIntentSchema`) are referenced by name in agent-architecture.md §3.3 but **never defined**. |
| Inline models concrete | 30/40 | `IntentPayload`, `IntentSpec`, `FieldState`, `MissingItem`, `FormCardData`, `QueryResultCardData` etc. — mostly concrete. -10: `FieldState` is referenced in `IntentSpec.Fields []FieldState` (request-model.md §3.1) and `FormCardData.Fields []FieldState` but **its struct definition is nowhere**. `EntityRef`, `EntityRecord`, `DiffOverlay`, `FormErrors`, `TracePayload`, `IntentMeta` similarly referenced but undefined. A developer cannot code these without guessing. |
| ER diagram complete | 28/30 | er-diagram.md has full ASCII entity boxes + relationships + cardinality + index strategy + soft-delete policy. -2: no Mermaid `erDiagram` block as rubric prefers; ASCII boxes are complete but Mermaid would be more tooling-friendly. |
| SQL DDL directly usable | 22/30 | schema.sql is executable as-is (both SQLite active + MySQL commented). FKs, indexes, defaults present. **However**: -4 for inline COMMENT syntax: rubric asks for "Inline COMMENT syntax" but DDL has `--` line comments only (no `COMMENT ON COLUMN`), so column docs are not in DB metadata. -4: `copilot_messages` has no FK on `session_id` (only `turn_id` FK) — denormalized and orphan-prone if turn cascade misses. |
| Cross-layer consistency | 15/30 | **Multiple inconsistencies**: (a) Go struct in tech-design.md §3 uses `BizKey string` (gorm varchar(36)) but injected convention AB-001 mandates `int64 bizKey` at Service boundary — the design's `GetByBizKey(ctx, bizKey string)` interface (interfaces.md L267) **directly violates** AB-001. (b) Cross-Layer Data Map (tech-design.md §5) lists `commit result` but the design creates the entity *before* the user submits the form card (request-model.md §6.1), contradicting the PRD semantics where "form card = prefill before write". (c) `MessageRepository.UpdateCardField` (interfaces.md L271) takes `value any` but messages.card is `json.RawMessage` in the Go struct (tech-design.md L335) — typed-mismatch. -15. |

### 3. Error Handling (130 pts) → **95/130**

| Criterion | Score | Notes |
|-----------|-------|-------|
| Error types defined | 35/45 | 11 error codes enumerated (tech-design.md §4.1). -5: `ERR_COPILOT_INPUT_TOO_LONG` and `ERR_COPILOT_TURN_CANCELLED` have HTTP status "—" — undefined mapping. -5: missing error code for the documented "敏感字段过滤未触发 → fail-open（继续调用），但记录告警" case — there is no event code for this fail-open signal. |
| Propagation strategy clear | 33/45 | §4.2 lists 3 categories (流中/入口/持久化). llm-integration.md §6 has detailed retry/backoff with concrete code. -7: "持久化错误：日志记录，但不阻塞流（best-effort persist）" — if `msgRepo.Append` for the intent message fails, the entire reconstruction logic (which depends on the persisted intent message bizKey) breaks; "best-effort" is hand-waved, no recovery path. -5: the documented "tool 错误不 break" (llm-integration.md §2.3 invariant #5) means a `commit_create` that fails is fed back to LLM — but the LLM may then retry the actual write, creating duplicate entities. No idempotency key is described. |
| HTTP status codes mapped | 27/40 | 8 of 11 codes mapped (429/503/504/502/422/403/404/410). 3 codes have "—" placeholder. -6: `commit_card` returns JSON 200 even when underlying entity service returns validation errors (request-model.md §1.5) — no mapping of entity-service errors to HTTP for the commit path; the design punts to "PATCH card state=failed" but the JSON response schema for that case is not specified. -7: SSE `error` event vs HTTP error boundary unclear — the design streams SSE 200 then can emit an `error` event for quota-exceeded, but quota is checked pre-flight (security.md §7.1) which should be HTTP 429 before stream opens. Contradiction. |

### 4. Testing Strategy (130 pts) → **103/130**

| Criterion | Score | Notes |
|-----------|-------|-------|
| Per-layer test plan | 38/45 | testing-strategy.md §3 has per-layer tables (Planner / Executor / Orchestrator / Handler / Repository / Frontend). -4: no explicit test plan for `prompt/` ContextBuilder FIFO cropping edge cases (only listed in coverage table §6, no scenarios). -3: Provider SSE parser (glm_provider.go `parseSSEStream`) — no test scenarios for partial-chunk tool_call fragmentation (a real GLM behavior called out in llm-integration.md §3.1). |
| Coverage target numeric | 38/45 | §6 has per-module numeric targets (agent ≥85%, orchestrator ≥85%, etc.). -4: no overall project coverage floor (CLAUDE.md says "核心模块 ≥ 80%" but design doesn't reconcile per-module vs project-wide). -3: frontend coverage target absent (only "Vitest + RTL" named). |
| Test tooling named | 27/40 | Backend: Go testing + testify ✓. Frontend: Vitest + RTL + MSW ✓. **But**: testing-strategy.md §4.2 uses `jest.spyOn(global, 'fetch')` — project uses Vitest, not Jest (CLAUDE.md confirms `npx vitest run`). §7.1 specifies `go-version: '1.26'` — nonexistent Go version (latest is 1.23). §7.1 uses `actions/setup-node@v3` with `node-version: '20'` but no pnpm/yarn lockfile mention. -13 for these incorrect tooling references that would break CI on first run. |

### 5. Breakdown-Readiness ★ (180 pts) → **133/180**

(critical gate — must reach 160/180 to advance to /breakdown-tasks.)

| Criterion | Score | Notes |
|-----------|-------|-------|
| Components enumerable | 58/65 | Components clearly enumerable: 5 Agents, ~17 tools (3 categories), 5 repos, 1 dispatcher, 1 orchestrator, 1 ctxBuilder, 1 provider (+factory), 11 SSE event types. -7: the `prompt/` package lists 5 files (context_builder/system_prompts/schema_loader/token_counter/history_window) but their interface boundaries are blurred — `ContextBuilder.Build` calls into `SchemaLoader` and `HistoryWindow` and `TokenCounter`, yet the only documented interface is `ContextBuilder`; the other 4 are file names with no contracts. A task breakdown would have to invent these contracts. |
| Tasks derivable | 50/65 | Each Agent → at least one impl task ✓. Each tool → impl task ✓. Each model → schema migration task ✓ (schema.sql is direct-use). **But**: -8 for the missing `FieldState` / `EntityRef` / `EntityRecord` / `TracePayload` / `IntentMeta` / `DiffOverlay` / `FormErrors` struct definitions — a developer starting tasks would have to reverse-engineer these from JSON examples. -7 for the missing `api-handbook.md` (referenced in tech-design.md L34 as `[api-handbook.md](./api-handbook.md) | API 手册` but file does not exist) — Handler-level task breakdown needs the request/response contract. |
| PRD AC coverage | 25/50 | **Significant AC gaps** (each -5 to -6, total -25):<br>• **Confidence band handling (≥0.7 / 0.4-0.7 / <0.4)** — PRD Story 7 enumerates 4 boundary ACs (0.7/0.69/0.4/0.39). Design has zero mention of confidence scoring anywhere. (-8)<br>• **Quota-exceeded → 关键词匹配降级模式** — PRD Story 7 + Security explicitly requires a non-LLM keyword-matching fallback. Design returns `ERR_COPILOT_QUOTA_EXCEEDED` error event with `fallbackAction: use_form`. The keyword mode is named once (security.md L177) but no design, no schema, no flow. (-6)<br>• **Concurrent field edit (last-write-wins + diff展示)** — PRD Story 7 requires diff preview when dialog-supplemented fields conflict with directly-edited fields. PATCH /messages/:id has no concurrency model, no diff event type, no diff card type. (-5)<br>• **50-turn session cap** — design explicitly deviates; disclosed but PRD Story 7 AC still enumerates the cap behavior. Deviation note in tech-design.md §1.5 acknowledges this, so it's "disclosed gap" rather than "missing". (-2)<br>• **AI service unavailable → 面板降级提示** — partial: `/health` endpoint exists but no "面板展示降级提示" UI event is in sse-protocol.md (no SSE event fires before a turn starts; the panel polls `/health`). (-4) |

### 6. Security Considerations (80 pts) → **58/80**

(PRD has auth + data privacy + multi-user — dimension scored.)

| Criterion | Score | Notes |
|-----------|-------|-------|
| Threat model present | 27/40 | security.md covers: prompt injection, sensitive field leak, cross-Team bizKey, RBAC bypass via AI, quota abuse, key leak. -8: **no threat model for the new "AI creates entity before user submits" behavior** — if the Writer Executor calls `MainItemService.Create` and the user never clicks 提交, an orphan MainItem is persisted. This is a data-pollution threat the design introduces but does not acknowledge. -5: no threat model for the SSE streaming endpoint itself (no rate-limit on connection count, no max concurrent streams per user — a malicious user could open many SSE streams, each pinning a goroutine for up to 8×30s). |
| Mitigations concrete | 31/40 | Concrete code for sanitizer (§4), redactor (§5), quota (§7.1), feature flag cache (§7.2). **But**: -4: sanitizer regex list is brittle — pattern `(?i)you are (now )?a` will false-positive legitimate text like "you are a member of team X". -5: `QuotaService.CheckQuota` "fail-open（继续调用）" on DB error (security.md L186) — for a quota-abuse threat, fail-open is the wrong default; an attacker who triggers DB errors gets unlimited calls. Should fail-closed for high-risk operations. |

### 7. Implementation Feasibility (140 pts) → **95/140**

| Criterion | Score | Notes |
|-----------|-------|-------|
| Dependencies available | 32/50 | -6: Go 1.26 referenced (testing-strategy.md §7.1) does not exist; latest is 1.23 — design assumes unavailable toolchain. -4: `jest.spyOn` used in frontend tests but project uses Vitest (CLAUDE.md) — jest is not a project dep. -3: GLM provider claims "GLM 通过标准 HTTPS 调用，不引入 SDK" — fine, but `glm-4-plus` model name and `/api/paas/v4` endpoint should be verified against current GLM docs (no citation). -5: Convention AB-001 conflict — design uses `bizKey string` everywhere, but project convention requires `int64 bizKey` at Service boundary. Either the design violates convention (-15 per rubric) or the convention needs amendment — at minimum the conflict must be acknowledged. Scored -5 here and -15 in Dim 2 cross-layer consistency (already applied). |
| Architecture fits project structure | 38/50 | `internal/copilot/` subtree follows existing `internal/<module>/` pattern. Reuses `internal/service/*`, `internal/middleware/*`, `internal/pkg/snowflake`, `internal/pkg/apperrors`. **But**: -7: the design introduces `Tool.Execute(ctx, args map[string]any, p ToolExecParams)` returning `ToolResult{Data map[string]any}` — untyped bags clash with Go's typed service-layer style in the existing codebase (where `MainItemService.Create(ctx, req MainItemCreateReq) (MainItem, error)`). The Dispatcher bridges by routing `FormCard` to typed service calls, but the FormCard→CreateReq mapping is described in prose (interfaces.md §7) without typed signatures. -5: design's `provider.Event` enum is `EventDelta/EventToolCall/...` but interfaces.md L52-58 also uses `ProviderEventDelta` (Go convention) — two naming styles mixed in the same package. |
| Technical claims grounded | 25/40 | -5: "GLM 没有公开 tokenizer，1 token ≈ 1.5 中文字符 / 4 英文字符" (llm-integration.md §3.4) — for Chinese-heavy prompts this underestimates by ~30%, which propagates to ContextBuilder budget; "±20% 可接受" is hand-waved. -4: "20 并发用户下，P95 延迟不超过 5s" (testing-strategy.md §8.2) — no capacity model (how many concurrent GLM HTTP connections? Provider has no connection pool config). -3: "成本失控... 月度 $200 熔断" — but `agent_call_logs.cost_usd` is stored per-call yet no cron/materialized view computes the monthly sum; `MonthlyCost` query (interfaces.md L303) does SUM live, which on a 30-day × N-users table will be slow without rollup. -3: "敏感字段过滤未触发 → fail-open" — security claim that is not grounded in any concrete detection logic; "fail-open" assumes the redactor either succeeds or signals, but the redactor has no "did I miss something?" signal. |

### Cross-Dimension Coherence Check

- **Cross-layer inconsistency** (Dim 2 ↔ Dim 7): `bizKey string` vs convention `int64 bizKey`. Already deducted in Dim 2.
- **Error flow contradiction** (Dim 3 ↔ Dim 5): Quota-exceeded is checked pre-flight (HTTP 429) per security.md §7.1 but the SSE error event example in sse-protocol.md §9 emits `ERR_COPILOT_QUOTA_EXCEEDED` mid-stream. Two different code paths for the same logical condition, undocumented. (-3 in Dim 3, already counted.)
- **AC gap flow** (Dim 5 → Blindspot): Confidence band + keyword fallback + concurrent diff — all PRD AC gaps — surface as [blindspot] attacks below.

---

## Phase 3 — Blindspot Hunt

Each cites a specific document quote and the failure pattern it matches.

### [blindspot-B1] Silent entity pollution when user abandons post-creation form card

**Quote** (request-model.md §6.1, lines 558-562):
```
{"event":"tool_call",...,"toolName":"commit_create","arguments":{"entity_type":"main_item","fields":{...}}}
{"event":"tool_result",...,"status":"success","result":{"bizKey":"MI-0023"}...}
{"event":"card_message",...,"cardType":"form","status":"prefilled","cardData":{...,"targetEntity":{"bizKey":"MI-0023",...}}}
```
The Writer Executor's `commit_create` Action tool invokes `MainItemService.Create` **during execute phase**, persisting a real entity (`MI-0023`) to the database. The "form card" pushed afterwards is a confirmation card referencing the already-created entity.

**Failure pattern**: *Solutions that reintroduce patterns they claim to eliminate.* tech-design.md §1.3 L88 claims "**AI 不直接写库**——AI 解析意图后调用现有实体 service 的 Create/Update/Transition". The intent was "AI doesn't bypass service layer", but the literal effect is "AI triggers DB write before user confirms". This contradicts PRD §Flow step 7 ("用户点击提交 → ... 调用现有 API 端点执行") and creates orphan entities when users abandon (Turn superseded / network drop / browser close between execute and commit).

**What must improve**: Move `commit_create` / `commit_update` / `move_sub_item` from execute-phase Action tools to commit-phase server-side execution (triggered by `POST /messages {type: commit_card}`). The form card must be a *prefill preview* (in-memory only), not a *post-creation confirmation*. This is the single biggest design flaw and it propagates to security (orphan pollution threat), PRD AC coverage (Story 1 AC #6 "调用 MainItem 创建 API" expects commit-time), and breakdown-readiness (tasks must be reorganized).

### [blindspot-B2] Orphan goroutine / connection exhaustion via unprotected SSE endpoints

**Quote** (llm-integration.md §6.4, lines 770-777):
```
| 单次 Agent 迭代（一次 LLM 调用） | 30s | 同上（共用） |
| 单次 Agent StreamRun 总时长 | 4 分钟 | `8 iter × 30s`（隐式上限） |
| Turn 总执行 | 无显式限制 | 由客户端断开 / max iterations 间接限制 |
```

**Failure pattern**: *Unhandled concurrent access — shared state with no synchronization model.* A single `free_text` request can pin a goroutine for 4 minutes (8 iterations × 30s). The Turn-in-flight guard (state-machines.md §6) rejects *same-session* concurrent `free_text` but a user can open N sessions in parallel, each pinning a goroutine + GLM upstream connection. No rate-limit on SSE connection count per user, nosemaphore on GLM outbound concurrency. 20 users × 4-min StreamRun = potential goroutine explosion.

**What must improve**: Add (a) per-user concurrent SSE stream limit (e.g., 1 active turn across all sessions), (b) outbound GLM HTTP connection semaphore (cap concurrent in-flight GLM calls globally), (c) hard StreamRun wall-clock deadline (not just iteration count) — e.g., 60s total. Document these in llm-integration.md §6.4 with concrete config keys.

### [blindspot-B3] `cost_usd` column populated but no rollup job; live SUM will degrade

**Quote** (security.md §7.2, lines 195-198 + interfaces.md §303):
```
- 月度成本 > $200 → 管理员一键关闭
- 单日总调用 > 1000 次（异常）→ 管理员一键关闭
...
MonthlyCost(ctx context.Context, year, month int) (float64, error)
```

**Failure pattern**: *Missing data migration / aggregation strategy.* `agent_call_logs` is append-only with 30-day retention. At scale (1000 calls/day × 30 days = 30k rows per user × N users), `MonthlyCost` doing `SUM(cost_usd) WHERE year/month` on every check becomes O(N) per request. The design names the query but provides no rollup job, no materialized view, no cron. The 30s feature-flag cache (§7.2) only caches the boolean, not the cost computation.

**What must improve**: Either (a) add a `copilot_cost_daily` rollup table populated by cron, or (b) precompute `monthly_cost_usd` per user/day and cache aggressively. Document the refresh cadence. The "每月 $200 熔断" threat is only as good as the cost counter's freshness.

### [blindspot-B4] SSE `error` event vs HTTP error response boundary contradicts quota pre-flight check

**Quote** (security.md §7.1 L186-190 + sse-protocol.md §9 L503):
```go
// QuotaService.CheckQuota — pre-flight
count, err := s.callLogRepo.CountTodayByUser(ctx, userBizKey)
if count >= s.dailyLimit { return ErrQuotaExceeded }
```
vs
```
{"event":"error",...,"code":"ERR_COPILOT_QUOTA_EXCEEDED","message":"今日 AI 调用已达上限（50 次）","recoverable":true,"fallbackAction":"use_form"}
```

**Failure pattern**: *Error paths that terminate silently / contradictory error semantics.* The quota is checked pre-flight (HTTP layer), so a quota-exceeded request should return HTTP 429 *before* the SSE stream opens (per tech-design.md §4.1 mapping `ERR_COPILOT_QUOTA_EXCEEDED → 429`). Yet the SSE protocol example shows the same error code as an in-stream `error` event. Which is it? If both code paths exist, the contract is ambiguous; if only one exists, the other document is wrong. Frontend cannot know whether to expect a JSON 429 or an SSE error event.

**What must improve**: Pick one. Quota-exceeded is pre-flight → HTTP 429 + JSON body `{code, recoverable, fallbackAction}` (no SSE stream opens). Update sse-protocol.md §9 example to use a genuinely mid-stream error (e.g., `ERR_COPILOT_AI_TIMEOUT` after first byte). Update security.md to clarify "quota check happens before SSE stream; quota errors are HTTP 429".

### [blindspot-B5] Idempotency absent for `commit_*` Action tools — duplicate entity creation on LLM retry

**Quote** (llm-integration.md §6.3, lines 751-764):
```go
result, err := a.tools.Execute(ctx, tc.Name, tc.Arguments, toolParams)
if err != nil {
    result = tool.ToolResult{Status: tool.ToolStatusError, Error: err.Error()}
    // 不 break，继续 append tool message
}
// LLM 在下一轮看到 role=tool, content={status:error,...}
// 可能：① 重试（修正 args）  ② 改用其他工具  ③ 放弃并 emit 失败答复
```

**Failure pattern**: *Error paths that terminate silently / unhandled concurrent access.* If `commit_create` times out at the GLM HTTP layer (request reached MainItemService but response was lost), the LLM sees `status=error` and may issue `commit_create` again with the same args → duplicate MainItem. There is no idempotency key, no request-id dedupe at the Dispatcher, no `client_request_id` column on `copilot_agent_call_logs`.

**What must improve**: Either (a) generate an idempotency key per `intent_id` and pass it through to entity service Create calls (requires existing services to support idempotency — likely a v2 concern), or (b) restrict LLM tool retries to Read tools only (Action tools are fire-once; an error becomes a terminal failure for the turn). Document this restriction in agent-architecture.md §2 and llm-integration.md §6.3.

### [blindspot-B6] `current_turn_id` is part of `copilot_sessions` but never UPDATEd in the documented flows

**Quote** (schema.sql L25 + request-model.md §2.2 handler code):
```
current_turn_id VARCHAR(36) NOT NULL DEFAULT '',
```
The handler `handleFreeText` (request-model.md L187-203) creates a Turn row and persists user msg, but **never calls `sessionRepo.UpdateCurrentTurn(turnID)`**. The state-machines.md §6 middleware `TurnInFlightGuard` (L488) reads `sess.CurrentTurnID` — if it is always the empty default `''`, the guard never fires. The supersession logic (state-machines.md L426) also depends on it.

**Failure pattern**: *Implicit coupling between modules — dependencies not acknowledged in interfaces.* `SessionRepository.UpdateCurrentTurn` is declared (interfaces.md L280) but no caller is shown anywhere. The concurrent-guard invariant (state-machines.md §6) silently breaks.

**What must improve**: Add `h.sessionRepo.UpdateCurrentTurn(c, sessionID, turnID)` to the `handleFreeText` transaction (and the supersession path). Add an integration test that verifies `current_turn_id` is non-empty after first `free_text` and points to the latest turn after subsequent `free_text`. Document the invariant "every successful free_text updates current_turn_id in the same tx" in state-machines.md §6.

### [blindspot-B7] Failed-persist intent message leaves turn in unrecoverable state

**Quote** (tech-design.md §4.2 L378-380):
```
- 持久化错误：日志记录，但不阻塞流（best-effort persist）
```
And (request-model.md §5.3 L495-499):
```go
// persist 意图消息
intentMsg := o.msgRepo.Append(Message{...})
// 推送 card_message 事件（cardType=intent）
eventCh <- cardMessageEvent(turnCtx, "intent", intentPayload.State, intentPayload, intentMsg.BizKey)
```

If `msgRepo.Append` fails (DB connection blip), `intentMsg` is nil/zero, `intentMsg.BizKey` either panics (nil deref) or emits a card_message event with empty messageId. The user sees the intent card on screen but the next `confirm_intent` request (`intentMessageId: ""`) cannot find the message → 404. The turn is stuck in `awaiting_confirm_intent` forever (cron will mark it `superseded` after 24h, but the user has no immediate recovery path).

**Failure pattern**: *Error paths that terminate silently.* "Best-effort persist" is incompatible with the design's own invariant "意图消息是 source of truth" (request-model.md §1.2). If the source of truth fails to persist, the entire downstream flow cannot reconstruct plan.

**What must improve**: Persist intent message **inside** the same transaction as Turn status update (already done for user msg in handleFreeText, but not for the AI intent msg in HandleUserMessage). If persist fails, emit `error` event + turn `failed`, do NOT emit the card_message event. Remove "best-effort persist" language from §4.2 for the intent-message path (it can remain for trace messages, which truly are best-effort).

---

## Deduction Items Summary

| ID | Type | Pts | Quote / Anchor |
|----|------|-----|----------------|
| D1 | Vague language | -20 | "1 token ≈ 1.5 中文字符 / 4 英文字符... ±20% 可接受" (llm-integration.md §3.4) — unvalidated heuristic treated as grounded claim. |
| D2 | Cross-section inconsistency | -30 | `bizKey string` (interfaces.md throughout) vs AB-001 convention `int64 bizKey`. |
| D3 | PRD AC gap | -30 | Confidence band handling absent (PRD Story 7, 4 boundary ACs). |
| D4 | PRD AC gap | -30 | Quota-exceeded keyword-matching fallback mode absent (PRD Story 7 + Security). |
| D5 | PRD AC gap | -30 | Concurrent field edit `last-write-wins + diff展示` absent (PRD Story 7). |
| D6 | Placeholder / missing deliverable | -20 | `api-handbook.md` referenced but file does not exist (tech-design.md L34). |
| D7 | Placeholder (undefined types) | -20 × instances | `FieldState`, `EntityRef`, `EntityRecord`, `TracePayload`, `IntentMeta`, `DiffOverlay`, `FormErrors` referenced but never defined. Counted as aggregate -20 in Dim 2. |
| D8 | Cross-section inconsistency | -30 | Quota error path: HTTP 429 (tech-design.md §4.1) vs SSE error event (sse-protocol.md §9). |
| D9 | Convention violation | -15 | Go 1.26 / jest references conflict with project toolchain (CLAUDE.md). |
| D10 | Vague language | -20 | "持久化错误：日志记录，但不阻塞流（best-effort persist）" (tech-design.md §4.2) — for intent message this is unsafe (see B7). |

(Deductions are reflected inside dimension sub-scores above; this table is for reviser reference.)

---

## Final Score

| Dimension | Score | Max |
|-----------|-------|-----|
| 1. Architecture Clarity | 150 | 170 |
| 2. Interface & Model Definitions | 127 | 170 |
| 3. Error Handling | 95 | 130 |
| 4. Testing Strategy | 103 | 130 |
| 5. Breakdown-Readiness ★ | 133 | 180 |
| 6. Security Considerations | 58 | 80 |
| 7. Implementation Feasibility | 95 | 140 |
| **TOTAL** | **761** | **1000** |

**Gate status**: ❌ Below target (900) and below breakdown-readiness gate (160/180). Iteration 2 required.

---

## Priority Fixes for Iteration 2 (Reviser Instructions)

1. **[B1] Move write operations to commit phase** (highest impact). Restructure: Executor's `commit_*` tools become *prefill* tools that build the form card without DB writes; actual writes happen in `commit_card` Handler path via Dispatcher. This single change resolves the orphan-pollution security threat, aligns with PRD Flow step 7, and unlocks several ACs.
2. **[D3/D4/D5] Address 3 PRD AC gaps**: confidence bands (add to Planner output schema + 3-band routing), keyword fallback mode (define a `KeywordFallbackService` and its trigger path), concurrent field diff (define PATCH semantics + diff event/card type).
3. **[B6] Add `UpdateCurrentTurn` calls** to all turn-creating handlers; add an integration test.
4. **[B7] Make intent-message persist transactional** with turn status update; remove "best-effort" language for source-of-truth messages.
5. **[D2] Reconcile bizKey type** with convention AB-001 (either update design to int64 or explicitly document the deviation with rationale).
6. **[D6] Create `api-handbook.md`** with all 8 HTTP endpoints, request/response schemas, error mappings.
7. **[D7] Define all referenced-but-undefined structs**: `FieldState`, `EntityRef`, `EntityRecord`, `TracePayload`, `IntentMeta`, `DiffOverlay`, `FormErrors`.
8. **[B2] Add SSE connection / GLM outbound concurrency limits** to llm-integration.md §6.4.
9. **[B4] Pick one quota-error path** (recommend HTTP 429 pre-flight) and update sse-protocol.md §9.
10. **[B5] Define Action-tool retry policy** (recommend fire-once for writes).
11. **[D9] Fix tooling versions**: Go 1.23 (not 1.26), Vitest (not jest).
12. **[B3] Add cost rollup table or materialized view** for monthly cost queries.

Addressing items 1-7 alone should push the score past 900 and unblock the breakdown-readiness gate.
