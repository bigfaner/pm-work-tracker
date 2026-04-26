---
date: "2026-04-26"
doc_dir: "docs/features/db-dialect-compat/design/"
iteration: "2"
target_score: "85"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 2

**Score: 87/100** (target: 85)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  18      │  20      │ ✅         │
│    Layer placement explicit  │  7/7     │          │            │
│    Component diagram present │  7/7     │          │            │
│    Dependencies listed       │  4/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  19      │  20      │ ✅         │
│    Interface signatures typed│  7/7     │          │            │
│    Models concrete           │  6/7     │          │            │
│    Directly implementable    │  6/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  13      │  15      │ ⚠️         │
│    Error types defined       │  4/5     │          │            │
│    Propagation strategy clear│  5/5     │          │            │
│    HTTP status codes mapped  │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  12      │  15      │ ⚠️         │
│    Per-layer test plan       │  4/5     │          │            │
│    Coverage target numeric   │  4/5     │          │            │
│    Test tooling named        │  4/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  18      │  20      │ ✅         │
│    Components enumerable     │  7/7     │          │            │
│    Tasks derivable           │  6/7     │          │            │
│    PRD AC coverage           │  5/6     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │  7       │  10      │ ⚠️         │
│    Threat model present      │  4/5     │          │            │
│    Mitigations concrete      │  3/5     │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  87      │  100     │ ✅         │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness 18/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Dependencies section | Internal module import relationships not stated — no explicit list of which packages import `pkg/dbutil` and what `pkg/dbutil` imports | -2 pts (Dependencies listed) |
| Data Models section | Concrete struct definitions (`type sqliteDialect struct{}`) and repo struct field additions (`dialect Dialect`) are still only described in prose, not shown as Go code | -1 pt (Models concrete) |
| Error Handling | No custom error type or error variable is defined for the panic-on-unrecognized-dialect case; panic messages are described in prose but no `var ErrUnsupportedDialect` or similar sentinel is shown | -1 pt (Error types defined) |
| Error Handling | No explicit error-to-HTTP-status mapping table; relies on "existing behavior" prose | -1 pt (HTTP status codes mapped) |
| Testing Strategy | The `scripts/lint-staged.sh` change (Story 3) has no test plan — how to verify the lint blocks bad commits and passes good ones | -1 pt (Per-layer test plan) |
| Testing Strategy | No numeric coverage target for modified repo code paths (NextCode/NextSubCode dialect-using paths); "现有测试通过" is regression, not coverage | -1 pt (Coverage target numeric) |
| Testing Strategy | Test tooling named only as "go test + testify" — no mention of how the MySQL dialect path is tested in repos without a real MySQL instance (mock? integration? manual only?) | -1 pt (Test tooling named) |
| NextCode/NextSubCode detail | The DI wiring for `main.go` is shown but the repo struct definition change (adding `dialect Dialect` field) is only implied by the constructor signature change, never shown as a concrete struct | -1 pt (Tasks derivable) |
| PRD Coverage Map | Story 3 lint check grep pattern, exclusion logic, and false-positive avoidance are still not specified | -1 pt (PRD AC coverage) |
| Security section | SQL injection mitigation relies on caller discipline and comments, not type-level enforcement or validation; `CastInt(expr string)` accepts arbitrary strings | -1 pt (Threat model present) |
| Security section | No concrete enforcement mechanism (e.g., input validation, allowlist, type constraint) — only documentation-based mitigation | -2 pts (Mitigations concrete) |

---

## Attack Points

### Attack 1: [Testing Strategy — MySQL repo path testing gap]

**Where**: Testing Strategy table, row "repository/gorm": "NextCode/NextSubCode 在 SQLite 下编号递增 — 现有测试通过"
**Why it's weak**: The entire feature exists to make NextCode/NextSubCode work on MySQL, yet the test plan only covers "SQLite 下编号递增". The Key Test Scenarios section lists "NextCode 端到端（手动 MySQL 集成测试）" as scenario 5, but this is marked as manual — there is no automated test plan for the MySQL dialect SQL path in the repository layer. The `dialect.go` unit tests prove the Dialect interface produces correct SQL fragments, but nothing proves those fragments are correctly integrated into the GORM query chain for NextCode on MySQL. A developer could break the MySQL integration (e.g., concatenation order, missing parentheses in `MAX(...)`) and no automated test would catch it.
**What must improve**: Add a plan for automated testing of the MySQL path in repos — either a unit test that constructs a `mysqlDialect` and verifies the generated SQL string in the `Select` clause, or an integration test strategy with a MySQL test container. At minimum, state the risk explicitly and define the manual verification checklist.

### Attack 2: [Security — No enforcement mechanism for SQL injection constraint]

**Where**: Security Considerations: "这一约束已在接口定义中以注释形式记录，未来调用者必须遵守"
**Why it's weak**: The security argument reduces to "we wrote a comment saying don't pass user input." Comments are not enforceable. The `CastInt(expr string)` signature accepts any string, and nothing — no type system, no validation, no allowlist — prevents a future developer from passing user-controlled data. The doc acknowledges the risk exists ("将字符串参数直接嵌入 SQL 表达式") but the mitigation is entirely documentary. For a design document, this is acceptable as a known trade-off, but the mitigation score must reflect that the mitigation is weak by design.
**What must improve**: Either add a concrete enforcement mechanism (input validation in CastInt/Substr checking against known column patterns, or a typed `ColumnExpr` string type that can only be constructed from constants), or explicitly document this as an accepted risk with a threat model that explains why the comment-only approach is sufficient given the call-site audit.

### Attack 3: [Breakdown-Readiness — Story 3 lint check implementation undefined]

**Where**: PRD Coverage Map, Story 3: "lint-staged.sh 新增 SQLite 关键字 grep 检查"
**Why it's weak**: The PRD specifies two acceptance criteria for Story 3: (a) "repo 层写入硬编码 SQLite 关键字时提交被拦截" and (b) "使用 dialect 包时提交通过（无假阳性）". The design says "新增 SQLite 关键字 grep 检查" but never defines: (1) which exact SQLite keywords are checked (SUBSTR? CAST? datetime? pragma_? all of them?), (2) which file paths are scanned (only `repository/gorm/*.go`? all `.go` files?), (3) how `dialect.go` itself is excluded since it must contain these strings in its output, (4) how test files are handled. A developer tasked with implementing this must make all these decisions from scratch.
**What must improve**: Add a "Lint Check Specification" subsection that defines: the exact keyword list, the file glob pattern, the exclusion logic (e.g., `grep -r 'SUBSTR\|CAST.*AS INTEGER\|datetime\|pragma_' backend/internal/repository/gorm/*.go | grep -v '_test.go'`), and the expected pass/fail examples.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Core SQL query change not shown (NextCode/NextSubCode before/after) | ✅ Yes | New "NextCode / NextSubCode SQL 重构详情" section (lines 172-218) provides complete before/after Go code for both `NextCode` and `NextSubCode`, including the exact `Select` clause rewrite from hardcoded SQLite to `r.dialect.CastInt(r.dialect.Substr(...))` |
| Attack 2: Silent SQLite fallback for unknown dialect names | ✅ Yes | Error Handling section now states: "既不是 'sqlite' 也不是 'mysql' 时 panic，附带清晰的错误信息" with example message. This is a fail-fast approach, not silent fallback. Key Test Scenarios also include "未知 Dialector → panic" test case |
| Attack 3: Security claim not enforced by interface design | ❌ Partially | Security section now includes "SQL 注入风险分析" paragraph that explicitly analyzes the injection path and documents the constraint. However, the enforcement mechanism is still comment-only — no validation, type constraint, or allowlist was added. The analysis is better but the mitigation remains documentary |

---

## Verdict

- **Score**: 87/100
- **Target**: 85/100
- **Gap**: 0 points (target reached)
- **Breakdown-Readiness**: 18/20 — can proceed to `/breakdown-tasks`
- **Action**: Target reached. The document addresses all three iteration-1 attacks substantively. The remaining gaps (MySQL repo test strategy, Story 3 lint spec, security enforcement mechanism) are minor and can be refined during implementation.
