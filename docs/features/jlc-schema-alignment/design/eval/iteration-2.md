---
date: "2026-04-24"
doc_dir: "docs/features/jlc-schema-alignment/design/"
iteration: "2"
target_score: "90"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 2

**Score: 94/100** (target: 90)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  20      │  20      │ ✅         │
│    Layer placement explicit  │  7/7     │          │            │
│    Component diagram present │  7/7     │          │            │
│    Dependencies listed       │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  20      │  20      │ ✅         │
│    Interface signatures typed│  7/7     │          │            │
│    Models concrete           │  7/7     │          │            │
│    Directly implementable    │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  15      │  15      │ ✅         │
│    Error types defined       │  5/5     │          │            │
│    Propagation strategy clear│  5/5     │          │            │
│    HTTP status codes mapped  │  5/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  15      │  15      │ ✅         │
│    Per-layer test plan       │  5/5     │          │            │
│    Coverage target numeric   │  5/5     │          │            │
│    Test tooling named        │  5/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  19      │  20      │ ✅         │
│    Components enumerable     │  7/7     │          │            │
│    Tasks derivable           │  7/7     │          │            │
│    PRD AC coverage           │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │  5       │  10      │ ⚠️         │
│    Threat model present      │  3/5     │          │            │
│    Mitigations concrete      │  2/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  94      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness < 12/20 blocks progression to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Security Considerations | Multi-node deployment constraint buried in prose, not prominent in architecture decisions — "本设计假设单节点部署，worker-id 硬编码为 1" (line 928) should be in Architecture section as explicit constraint | -2 pts |
| Security Considerations | Logging middleware mitigation vague — "logging middleware 记录请求/响应时只记录 bizKey，不打印完整 model struct" (line 938) lacks implementation guidance (e.g., custom MarshalJSON, field filtering middleware) | -2 pts |
| Security Considerations | No explicit threat for SQL injection via bizKey parameter — handler parses `c.Param("itemId")` as int64, but no validation that it's actually a valid snowflake format | -1 pts |
| PRD AC coverage | "无外键约束（DDL 层面）" not explicitly mapped in PRD Coverage Map — DDL shows no FOREIGN KEY constraints, but Coverage Map doesn't explicitly state this PRD requirement is addressed | -1 pts |

---

## Attack Points

### Attack 1: Security Considerations — Multi-Node Deployment Constraint Not Architecturally Prominent

**Where**: "本设计假设单节点部署，worker-id 硬编码为 1。多节点部署需引入 worker-id 协调机制（如 etcd/Redis 分布式锁），超出本次迭代范围。" (line 928)

**Why it's weak**: This is a critical architectural constraint that affects system scalability, but it's buried in the Security Considerations section under a threat model table. A developer reading the Architecture section would not see this constraint and might deploy multiple instances, causing biz_key collisions. The constraint should be:
1. In the Architecture section as an explicit "Deployment Constraints" subsection
2. In the Dependencies section noting the snowflake library's single-node limitation
3. Referenced in the Open Questions as a resolved constraint

**What must improve**: Add explicit "Deployment Constraints" subsection in Architecture:
```markdown
### Deployment Constraints

- **Single-node only**: Current design hardcodes `worker-id=1` for snowflake generation. Multi-node deployment requires worker-id coordination (etcd/Redis) — out of scope for this iteration.
- **Violation consequence**: biz_key collisions will occur if multiple nodes share the same worker-id.
```

### Attack 2: Security Considerations — Logging Middleware Mitigation Lacks Implementation Guidance

**Where**: "logging middleware 记录请求/响应时只记录 bizKey，不打印完整 model struct" (line 938)

**Why it's weak**: The mitigation states what NOT to do (don't print full struct), but provides no implementation guidance. A developer implementing logging middleware would need to know:
1. How to filter model fields in structured logging (e.g., zap's `FieldFilter`, logrus hooks)
2. Whether to implement custom `MarshalJSON()` for models
3. Whether to use a custom logging wrapper that auto-filters sensitive fields
4. What logging library is used in the project (not mentioned in Dependencies)

Without this, the mitigation is aspirational rather than actionable.

**What must improve**: Add concrete implementation guidance:
```markdown
### Logging Implementation

- Use structured logging (e.g., zap) with field filtering
- Implement custom `MarshalJSON()` for BaseModel that omits internal fields:
  ```go
  func (b BaseModel) MarshalJSON() ([]byte, error) {
      return json.Marshal(struct {
          BizKey int64 `json:"bizKey"`
      }{BizKey: b.BizKey})
  }
  ```
- Or use logging middleware that explicitly extracts `bizKey` from context, never logs raw model
```

### Attack 3: PRD AC Coverage — "无外键约束" Not Explicitly Mapped

**Where**: PRD Coverage Map (lines 942-961) lists 18 PRD AC items, but "无外键约束（DDL 层面）" is not explicitly listed despite being a PRD requirement.

**Why it's weak**: The DDL in the design correctly shows no FOREIGN KEY constraints (only indexes), but the PRD Coverage Map doesn't explicitly state this requirement is addressed. This creates a traceability gap — a reviewer checking PRD coverage would need to manually inspect the DDL to verify this requirement. The Coverage Map should have an explicit row:
```
| 无外键约束（DDL 层面） | schema.sql | 所有表仅保留索引，无 FOREIGN KEY 约束 |
```

**What must improve**: Add explicit row in PRD Coverage Map for "无外键约束（DDL 层面）" requirement.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Breakdown-Readiness — FK Migration Strategy Incomplete | ✅ Yes | New "Migration Strategy" section (lines 963-1018) with two-phase migration: ADD COLUMN → backfill via JOIN → DROP COLUMN. Explicit SQL examples for each table. |
| Error Handling — Missing biz_key Duplicate Error Handling | ✅ Yes | New `ERR_DUPLICATE_BIZ_KEY` error type added (line 833), HTTP 409 status mapped, service layer implementation with `isMySQLDuplicateError()` helper (lines 850-875). |
| Security — Distributed Worker-ID Collision Not Addressed | ⚠️ Partially | Added explicit constraint text (line 928, 932) stating single-node assumption and multi-node coordination requirement. However, constraint is buried in Security section, not architecturally prominent. |

---

## Verdict

- **Score**: 94/100
- **Target**: 90/100
- **Gap**: +4 points (above target)
- **Breakdown-Readiness**: 19/20 — can proceed to /breakdown-tasks
- **Action**: Target reached. Design is implementation-ready. Minor improvements to Security section would strengthen architectural clarity.

---

## Detailed Analysis

### 1. Architecture Clarity (20/20)

**Layer placement explicit (7/7)**: The design clearly states the three-layer architecture (schema → model → repo → service → handler → frontend) with explicit layer boundaries. Each component's responsibility is well-defined. The data flow direction is clear with arrows in the diagram.

**Component diagram present (7/7)**: ASCII diagram at lines 41-77 shows all components and their relationships. Data flow direction is clear with arrows. All major components are represented: schema.sql, model/base.go, repo layer, service layer, snowflake generator, handler layer, and frontend types.

**Dependencies listed (6/6)**: Dependencies table at lines 79-89 lists all required packages with types and descriptions. All test dependencies are now included (`stretchr/testify`, `DATA-DOG/go-sqlmock`). Import paths are explicit.

### 2. Interface & Model Definitions (20/20)

**Interface signatures typed (7/7)**: All repo and service interfaces have complete Go type signatures (lines 252-481). Parameter types, return types, and error types are explicit. New `FindByBizKey` and `SoftDelete` methods are fully typed.

**Models concrete (7/7)**: BaseModel (lines 93-103) and all model structs have explicit field names, types, GORM tags, and JSON tags. Deviation models (ProgressRecord, StatusHistory, TeamMember) are clearly documented with rationale for not embedding BaseModel. No prose-only descriptions.

**Directly implementable (6/6)**: A developer can copy-paste the interface and model definitions and start implementing. No guessing required. The `isMySQLDuplicateError()` helper function is provided with implementation.

### 3. Error Handling (15/15)

**Error types defined (5/5)**: Three error types defined (`ErrValidation`, `ErrNotFound`, `ErrDuplicateBizKey`). All have explicit error codes, names, descriptions, and HTTP status mappings.

**Propagation strategy clear (5/5)**: Clear statement at lines 879-883: repo returns error → service transparent → handler uses `RespondError`. SoftDelete idempotency handling is explicit.

**HTTP status codes mapped (5/5)**: All three error types mapped to HTTP status codes (400, 404, 409). The duplicate biz_key scenario now has explicit 409 Conflict mapping.

### 4. Testing Strategy (15/15)

**Per-layer test plan (5/5)**: Table at lines 889-895 covers all layers with specific test types, tools, and what to test. Coverage targets are explicit per layer.

**Coverage target numeric (5/5)**: Explicit targets: backend 90%, frontend ≥70%. E2E scenarios are enumerated (5 scenarios, all must pass).

**Test tooling named (5/5)**: All test tools named with import paths: `go test`, `vitest`, `Playwright`, `sqlmock` (with import path `github.com/DATA-DOG/go-sqlmock`), `stretchr/testify`.

### 5. Breakdown-Readiness (19/20)

**Components enumerable (7/7)**: All components can be counted:
- 1 schema.sql
- 8 model files (base.go + 7 domain models)
- 6 repo files with new methods
- 5 service files with new methods
- 1 snowflake package
- 1 frontend types file
- Multiple frontend pages/components

**Tasks derivable (7/7)**: Each interface maps to implementation tasks. Migration strategy is now explicit with two-phase approach and SQL examples. FK data migration is fully derivable.

**PRD AC coverage (5/6)**: PRD Coverage Map at lines 942-961 addresses most PRD requirements. Minor gap: "无外键约束（DDL 层面）" is not explicitly listed in the Coverage Map, though the DDL correctly implements it.

### 6. Security Considerations (5/10)

**Threat model present (3/5)**: Five threats identified at lines 922-928. Good coverage of keyword conflict, ID enumeration, biz_key exposure, logging leakage, and worker-id collision. However:
- Multi-node deployment constraint is buried in prose, not architecturally prominent
- No threat for SQL injection via bizKey parameter validation

**Mitigations concrete (2/5)**: Each threat has a mitigation, but:
- Logging middleware mitigation is vague ("不打印完整 model struct") without implementation guidance
- Multi-node constraint is stated but not prominently positioned in architecture decisions
- No explicit validation for bizKey format in handler layer
