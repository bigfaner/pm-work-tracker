---
date: "2026-04-30"
doc_dir: "docs/features/e2e-test-scripts-rebuild/design/"
iteration: "2"
target_score: "80"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 2

**Score: 91/100** (target: 80)

```
┌─────────────────────────────────────────────────────────────────┐
│                     DESIGN QUALITY SCORECARD                     │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Architecture Clarity      │  18      │  20      │ ✅          │
│    Layer placement explicit  │   7/7    │          │            │
│    Component diagram present │   5/7    │          │            │
│    Dependencies listed       │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Interface & Model Defs    │  17      │  20      │ ✅          │
│    Interface signatures typed│   7/7    │          │            │
│    Models concrete           │   6/7    │          │            │
│    Directly implementable    │   4/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Error Handling            │  15      │  15      │ ✅          │
│    Error types defined       │   5/5    │          │            │
│    Propagation strategy clear│   5/5    │          │            │
│    HTTP status codes mapped  │   5/5    │          │ N/A        │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. Testing Strategy          │  12      │  15      │ ⚠️          │
│    Per-layer test plan       │   5/5    │          │            │
│    Coverage target numeric   │   4/5    │          │            │
│    Test tooling named        │   3/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Breakdown-Readiness ★     │  20      │  20      │ ✅          │
│    Components enumerable     │   7/7    │          │            │
│    Tasks derivable           │   7/7    │          │            │
│    PRD AC coverage           │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 6. Security Considerations   │   9      │  10      │ ✅          │
│    Threat model present      │   5/5    │          │            │
│    Mitigations concrete      │   4/5    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  91      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

★ Breakdown-Readiness 20/20 — can proceed to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| Architecture / Component Diagram | `validateSpec.ts` and `updatePackageJson.ts` are called by the Orchestrator but absent from the component diagram — only the 3 forge skills appear as Orchestrator children | -2 pts |
| Interface & Model / Models concrete | `FeatureClassification` uses object literal syntax (`FeatureClassification = { slug: string ... }`) rather than TypeScript `interface` — inconsistent with the rest of the doc | -1 pt |
| Interface & Model / Directly implementable | `executeFeature` "shells out to the three forge skills in sequence" — the shelling-out mechanism is unspecified (child_process.exec? skill API? Claude subagent call?). A developer cannot implement this without guessing the execution context | -2 pts |
| Testing Strategy / Test tooling named | "ls / read" listed as the tool for graduation marker checks — this is a file system operation, not a test tool. "grep / AST scan" in the table is inconsistent with Interface 2 which states "no AST library required, just regex" | -2 pts |
| Testing Strategy / Coverage target numeric | Coverage targets are pass-rate thresholds (≥80%, 100%) not code line/branch coverage percentages. No line or branch coverage target stated for the orchestrator scripts themselves | -1 pt |
| Security / Mitigations concrete | "Credentials are rotated after each release cycle by the team lead" — no rotation schedule, no technical enforcement mechanism, no tooling named. This is a people-process control stated without any verifiable technical backing | -1 pt |

---

## Attack Points

### Attack 1: Interface & Model Definitions — executeFeature shelling-out mechanism is unspecified

**Where**: "Implementor: the per-feature loop script at `docs/features/e2e-test-scripts-rebuild/testing/scripts/run-all.ts`. It shells out to the three forge skills in sequence."

**Why it's weak**: The forge skills (`/gen-test-scripts`, `/run-e2e-tests`, `/graduate-tests`) are Claude skills, not CLI binaries. "Shells out" implies `child_process.exec` or `child_process.spawn`, but forge skills are invoked through the Claude Code harness, not via a shell command. A developer reading this cannot determine whether `run-all.ts` calls a CLI wrapper, makes an API call, or triggers a subagent. This is the single most critical implementation gap in the document — the entire orchestration mechanism is hand-waved with "shells out."

**What must improve**: Specify the invocation mechanism explicitly. If the orchestrator is itself a Claude session that calls skills via the Skill tool, say so. If there is a CLI wrapper (e.g., `claude /gen-test-scripts --slug foo`), name it and show the call signature. If `run-all.ts` is a TypeScript script that calls a programmatic API, show the import and call. The current text leaves the most important implementation detail undefined.

---

### Attack 2: Architecture Clarity — validateSpec.ts and updatePackageJson.ts are invisible in the component diagram

**Where**: The component diagram shows: `Orchestrator → /gen-test-scripts`, `Orchestrator → /run-e2e-tests`, `Orchestrator → /graduate-tests`. The diagram caption states "Orchestrator writes: · KNOWN_FAILURES.md · tests/e2e/package.json (after all features complete, via updatePackageJson)."

**Why it's weak**: `validateSpec.ts` and `updatePackageJson.ts` are defined as full typed interfaces with named implementors, yet neither appears as a node in the component diagram. A developer reading only the diagram sees three forge skills and nothing else — the two internal helper scripts that the Orchestrator also invokes are completely absent. The diagram is inconsistent with the Interfaces section. The phrase "via updatePackageJson" in the diagram caption is a dangling reference to a component that has no box.

**What must improve**: Add `validate-spec.ts` and `update-package-json.ts` as component boxes under the Orchestrator, with directed edges showing when they are called (validate-spec before /graduate-tests; update-package-json after all features complete). The diagram should be the authoritative component inventory — if a component has a typed interface, it must appear in the diagram.

---

### Attack 3: Testing Strategy — "ls / read" is not a test tool and the table contradicts Interface 2

**Where**: Testing Strategy table, row "Graduation marker": Tool = "ls / read". Row "Script conformance": Tool = "grep / AST scan". Interface 2 states: "Uses Node.js `fs.readFileSync` + regex scan (no AST library required for these three checks)."

**Why it's weak**: Two problems. First, "ls / read" is a file system operation, not a test tool — it has no assertion semantics, no pass/fail output, no test runner integration. The rubric requires "specific test libraries/frameworks named." Second, the table says "grep / AST scan" for script conformance, but Interface 2 explicitly says no AST library is needed and the check is regex-based. The table and the interface section contradict each other. A developer implementing the test layer would not know whether to reach for an AST parser or just use regex.

**What must improve**: Replace "ls / read" with the actual assertion mechanism — e.g., "`node:assert` + `fs.existsSync`" or "`node:test` with file existence check." Replace "grep / AST scan" with "`fs.readFileSync` + regex" to match Interface 2. Every tool entry in the testing table should name something a developer can `import` or `require`.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Three of four "interfaces" are not interfaces (no typed signatures) | ✅ | Interfaces 2 and 3 now have full typed signatures (`validateSpec(filePath: string): ValidationResult`, `updatePackageJson(specPaths: SpecPaths): void`). Old "Graduated Spec File Contract" and "KNOWN_FAILURES.md Entry Format" correctly moved to "File Format Conventions" section. |
| Attack 2: Security threat model lists non-threats; `admin/admin123` as mitigation | ✅ | Threat model rewritten as a 4-row table with concrete threats (config.yaml commit, spec injection, CI secret exposure, overly broad permissions). `admin/admin123` removed entirely. Countermeasures are specific. |
| Attack 3: Component diagram is a pipeline diagram, not a component diagram | ✅ | New diagram shows Orchestrator box, three forge skill boxes, with labeled read/write edges. Partial credit deducted only because `validateSpec.ts` and `updatePackageJson.ts` are missing from the diagram. |

---

## Verdict

- **Score**: 91/100
- **Target**: 80/100
- **Gap**: 0 points (target exceeded by 11)
- **Breakdown-Readiness**: 20/20 — can proceed to `/breakdown-tasks`
- **Action**: Target reached. All three iteration-1 attacks addressed. Remaining gaps (shelling-out mechanism, diagram completeness, test tooling naming) are minor and do not block task breakdown.
