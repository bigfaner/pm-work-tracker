---
date: "2026-04-24"
doc_dir: "docs/features/jlc-schema-alignment/design/"
iteration: "3"
target_score: "90"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 3

**Score: 92/100** (target: 90)

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
│ 5. Breakdown-Readiness ★     │  20      │  20      │ ✅         │
│    Components enumerable     │  7/7     │          │            │
│    Tasks derivable           │  7/7     │          │            │
│    PRD AC coverage           │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │  2       │  10      │ ⚠️         │
│    Threat model present      │  1/5     │          │            │
│    Mitigations concrete      │  1/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  92      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness < 12/20 blocks progression to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Security Considerations | Threat model table lists 5 threats but 3 are not genuine security threats — "status 关键字冲突导致 DDL/DML parse error" (line 929) is a schema design issue, not a security threat; "auto-increment id 暴露" (line 930) is addressed by json:"-" but the threat column doesn't describe the actual attack scenario; "biz_key 暴露泄露雪花时间戳和 worker-id" (line 931) is informational leakage, not a direct security threat | -2 pts |
| Security Considerations | Mitigations are incomplete — "repo 接口不暴露硬删除方法" (line 941) is a design choice, not a security mitigation; "SoftDelete 实现加 deleted_flag = 0 条件" (line 942) is for idempotency, not security; only the logging mitigation (lines 945-956) has concrete implementation guidance | -2 pts |
| Security Considerations | No threat for bizKey parameter injection — handler parses `c.Param("itemId")` as int64 via `strconv.ParseInt`, but negative values or values outside snowflake range are not validated; a malicious bizKey like `-1` or `99999999999999999999` could cause unexpected behavior | -2 pts |
| Security Considerations | Multi-node deployment constraint added to Architecture section (lines 91-94) but Security section still contains redundant prose (lines 933-937) that duplicates the same information without adding security context | -2 pts |

---

## Attack Points

### Attack 1: Security Considerations — Threat Model Contains Non-Security Items

**Where**: "status 关键字冲突导致 DDL/DML parse error" (line 929), "auto-increment id 暴露，攻击者可顺序枚举资源" (line 930)

**Why it's weak**: The threat model table conflates schema design issues with security threats:
1. "status 关键字冲突" is a MySQL reserved word conflict — a schema correctness issue, not a security threat. No attacker is involved.
2. "auto-increment id 暴露" describes a design decision (json:"-"), not a threat scenario. The actual threat would be "resource enumeration attack" or "information disclosure via predictable IDs".
3. "biz_key 暴露泄露雪花时间戳和 worker-id" (line 931) is informational leakage, but the mitigation column says "biz_key 通过 json:"bizKey" 对外暴露" — which contradicts the idea that this is a threat to mitigate.

A proper threat model should identify: WHO is attacking, WHAT they can do, and WHAT the impact is. The current table mixes design constraints with security concerns.

**What must improve**: Restructure threat model to focus on actual security threats:
```markdown
| 威胁 | 攻击者 | 攻击方式 | 影响 | 对策 |
|------|--------|----------|------|------|
| 资源枚举攻击 | 外部用户 | 通过递增 id 遍历所有资源 | 信息泄露 | id 不对外暴露，使用非顺序 biz_key |
| biz_key 信息泄露 | 日志查看者 | 从日志中提取 biz_key 原始值 | 泄露创建时间和机器标识 | logging middleware 过滤敏感字段 |
| 参数注入 | 外部用户 | 提交非法 bizKey 值（负数、超大值） | 潜在的异常行为 | handler 层验证 bizKey 范围 |
```

### Attack 2: Security Considerations — Mitigations Not Security-Focused

**Where**: "repo 接口不暴露硬删除方法，从接口层面杜绝误操作" (line 941), "SoftDelete 实现加 deleted_flag = 0 条件，防止重复软删并确保幂等性" (line 942)

**Why it's weak**: These are not security mitigations:
1. "repo 接口不暴露硬删除方法" is a data integrity design choice, not a security control. It prevents accidental data loss, not malicious attacks.
2. "SoftDelete 幂等性" is a correctness property, not a security mitigation.

The only genuine security mitigation is the logging middleware guidance (lines 945-956), which addresses the biz_key information leakage threat. But even this is incomplete — it doesn't specify which logging library the project uses, making the `MarshalJSON()` example potentially inapplicable.

**What must improve**: Replace non-security mitigations with actual security controls:
```markdown
### Mitigations

- **biz_key 信息泄露防护**: logging middleware 使用 zap 的 `FieldFilter` 或自定义 `MarshalJSON()` 过滤 BaseModel 字段
- **参数验证**: handler 层验证 bizKey 范围 (1 <= bizKey <= 2^63-1)，拒绝负数和超大值
- **部署约束**: 单节点部署，worker-id 硬编码为 1；多节点部署需实现 worker-id 协调（见 Architecture → Deployment Constraints）
```

### Attack 3: Security Considerations — No bizKey Parameter Validation

**Where**: Handler layer code (lines 237-255) shows `strconv.ParseInt(c.Param("itemId"), 10, 64)` without range validation

**Why it's weak**: The design shows handler code that parses bizKey from URL path:
```go
bizKey, _ := strconv.ParseInt(c.Param("itemId"), 10, 64)
```

But there's no validation that:
1. bizKey is positive (snowflake IDs are always positive)
2. bizKey is within valid snowflake range
3. bizKey is not zero

A malicious request with `itemId=-1` or `itemId=0` would pass parsing and reach the repo layer, potentially causing unexpected behavior or information leakage through error messages.

**What must improve**: Add explicit validation in handler layer:
```go
bizKey, err := strconv.ParseInt(c.Param("itemId"), 10, 64)
if err != nil || bizKey <= 0 {
    apperrors.RespondError(c, apperrors.ErrValidation)
    return
}
```

Document this in Security Considerations as a parameter validation requirement.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Security — Multi-Node Deployment Constraint Not Architecturally Prominent | ✅ Yes | New "Deployment Constraints" subsection added in Architecture section (lines 91-94): "Single-node only: Current design hardcodes worker-id=1 for snowflake generation. Multi-node deployment requires worker-id coordination (etcd/Redis) — out of scope for this iteration." |
| Security — Logging Middleware Mitigation Lacks Implementation Guidance | ✅ Yes | New "Logging Implementation" subsection (lines 945-956) with concrete code example for custom `MarshalJSON()` and structured logging guidance. |
| PRD AC Coverage — "无外键约束" Not Explicitly Mapped | ✅ Yes | PRD Coverage Map now includes explicit row (line 969): "无外键约束（DDL 层面） | schema.sql | 所有表仅保留索引，无 FOREIGN KEY 约束" |

---

## Verdict

- **Score**: 92/100
- **Target**: 90/100
- **Gap**: +2 points (above target)
- **Breakdown-Readiness**: 20/20 — can proceed to /breakdown-tasks
- **Action**: Target reached. Design is implementation-ready. Security Considerations section has improved but still contains non-security items in threat model and lacks parameter validation guidance.

---

## Detailed Analysis

### 1. Architecture Clarity (20/20)

**Layer placement explicit (7/7)**: The design clearly states the three-layer architecture (schema → model → repo → service → handler → frontend) with explicit layer boundaries. Each component's responsibility is well-defined. The data flow direction is clear with arrows in the diagram.

**Component diagram present (7/7)**: ASCII diagram at lines 41-77 shows all components and their relationships. Data flow direction is clear with arrows. All major components are represented: schema.sql, model/base.go, repo layer, service layer, snowflake generator, handler layer, and frontend types.

**Dependencies listed (6/6)**: Dependencies table at lines 79-89 lists all required packages with types and descriptions. All test dependencies are included. Import paths are explicit. New "Deployment Constraints" subsection (lines 91-94) adds critical architectural constraint about single-node deployment.

### 2. Interface & Model Definitions (20/20)

**Interface signatures typed (7/7)**: All repo and service interfaces have complete Go type signatures (lines 259-488). Parameter types, return types, and error types are explicit. New `FindByBizKey` and `SoftDelete` methods are fully typed.

**Models concrete (7/7)**: BaseModel (lines 100-109) and all model structs have explicit field names, types, GORM tags, and JSON tags. Deviation models (ProgressRecord, StatusHistory, TeamMember) are clearly documented with rationale for not embedding BaseModel. No prose-only descriptions.

**Directly implementable (6/6)**: A developer can copy-paste the interface and model definitions and start implementing. No guessing required. The `isMySQLDuplicateError()` helper function is provided with implementation.

### 3. Error Handling (15/15)

**Error types defined (5/5)**: Three error types defined (`ErrValidation`, `ErrNotFound`, `ErrDuplicateBizKey`). All have explicit error codes, names, descriptions, and HTTP status mappings.

**Propagation strategy clear (5/5)**: Clear statement at lines 884-888: repo returns error → service transparent → handler uses `RespondError`. SoftDelete idempotency handling is explicit.

**HTTP status codes mapped (5/5)**: All three error types mapped to HTTP status codes (400, 404, 409). The duplicate biz_key scenario has explicit 409 Conflict mapping.

### 4. Testing Strategy (15/15)

**Per-layer test plan (5/5)**: Table at lines 892-900 covers all layers with specific test types, tools, and what to test. Coverage targets are explicit per layer.

**Coverage target numeric (5/5)**: Explicit targets: backend 90%, frontend ≥70%. E2E scenarios are enumerated (5 scenarios, all must pass).

**Test tooling named (5/5)**: All test tools named with import paths: `go test`, `vitest`, `Playwright`, `sqlmock` (with import path `github.com/DATA-DOG/go-sqlmock`), `stretchr/testify`.

### 5. Breakdown-Readiness (20/20)

**Components enumerable (7/7)**: All components can be counted:
- 1 schema.sql
- 8 model files (base.go + 7 domain models)
- 6 repo files with new methods
- 5 service files with new methods
- 1 snowflake package
- 1 frontend types file
- Multiple frontend pages/components

**Tasks derivable (7/7)**: Each interface maps to implementation tasks. Migration strategy is explicit with two-phase approach and SQL examples. FK data migration is fully derivable.

**PRD AC coverage (6/6)**: PRD Coverage Map at lines 958-979 addresses all PRD requirements including the newly added "无外键约束（DDL 层面）" row.

### 6. Security Considerations (2/10)

**Threat model present (1/5)**: Five threats identified at lines 927-933, but:
- 3 of 5 items are not genuine security threats (schema design issues, design choices)
- No threat for parameter injection via bizKey
- Threat model mixes design constraints with security concerns

**Mitigations concrete (1/5)**: Mitigations at lines 939-956, but:
- 2 of 3 mitigations are not security-focused (data integrity, idempotency)
- Only logging mitigation has implementation guidance
- No mitigation for bizKey parameter validation
- Redundant prose about multi-node deployment (lines 933-937) duplicates Architecture section without adding security context
