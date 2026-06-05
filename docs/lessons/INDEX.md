---
description: "Lesson index — categorized anti-patterns and postmortems for on-demand loading"
---

# Lesson Index

Load relevant lessons BEFORE starting work in these areas. Lessons prevent recurring mistakes.

## By Category

### arch- — Architecture & Design Gaps

| File | When to Load |
|------|-------------|
| arch-bizkey-vs-internal-id.md | New entity with FK fields — BizKey vs internal ID confusion |
| arch-button-icon-convention.md | Building action buttons — icon+text pattern |
| arch-e2e-graduation-staging.md | Graduating e2e tests — directory reorganization |
| arch-task-executor-missing-e2e-step.md | Task executor flow — missing e2e verification step |
| arch-task-failure-recovery-loop.md | Task executor encounters test failures — must block + spawn fix tasks |
| arch-ui-integration-gap.md | Component task breakdown — must include page wiring task |

### debug- — Debugging Techniques

| File | When to Load |
|------|-------------|
| debug-e2e-beforeall-cascade.md | E2E beforeAll failures — cascading misleading errors |
| debug-e2e-zero-tests.md | E2E reports "0 tests" — layered import/config issues |

### gotcha- — Agent & Tool Pitfalls

| File | When to Load |
|------|-------------|
| gotcha-ac-self-report-without-verification.md | Accepting agent "done" claims without verification |
| gotcha-agent-breaks-safety-net.md | Agent commits files into nested .gitignore paths |
| gotcha-agent-browser-causes-test-timeout.md | E2E tests using agent-browser — timeout when not installed |
| gotcha-bizkey-vs-id-confusion.md | BizKey fields storing internal IDs — cascading 403 errors |
| gotcha-continue-replay-fresh-agent.md | Task CONTINUE dispatches fresh agent — re-executes completed work |
| gotcha-design-tech-skip-askuserquestion.md | /tech-design skill — skipped user confirmation |
| gotcha-e2e-script-generation.md | Generated e2e scripts — wrong paths/prefixes/ports |
| gotcha-eval-prd-use-zcode-agents.md | /eval-prd skill — must orchestrate scorer/reviser subagents |
| gotcha-hook-unbounded-test-timeout.md | Stop hooks — unbounded test suites hang past 600s |
| gotcha-parallel-subagent-rate-limit.md | Parallel subagent launches — max 3 concurrent to avoid API 529 |
| gotcha-no-subagent-for-sequential-work.md | Subagents for stateful sequential work — lost auth/context |
| gotcha-pipe-tail-buffers-test-output.md | Playwright output buffering — appears hung for 50+ min |
| gotcha-schema-alignment-cascading-test-failures.md | Schema changes cascade to unrelated test files |
| gotcha-subagent-partial-commits.md | Subagents leave 75+ files uncommitted after /run-tasks |
| gotcha-interface-blast-radius-dispatcher.md | Adding methods to shared interfaces — large blast radius appears as "stuck" |
| gotcha-claim-in-progress-priority.md | forge task claim prioritizes in_progress over pending — blocks dispatch loop |
| gotcha-forge-surfaces-path-matching.md | forge surfaces <path> needs surface key prefix — `.` or `docs/` won't match |
| gotcha-task-claim-priority-skip.md | forge task claim skips unblocked tasks — jumps to unrelated pipeline tasks |
| gotcha-tech-design-decision-archiving.md | Tech design must record naming/structural conventions |

### pattern- — Reusable Patterns

| File | When to Load |
|------|-------------|
| pattern-phase-gate-tasks.md | Task breakdown — model quality gates as explicit tasks |

### postmortem- — Postmortem Analysis

| File | When to Load |
|------|-------------|
| postmortem-jlc-schema-alignment.md | Large feature "done" but needed 12+ fix commits |

### tool- — Tool Usage

| File | When to Load |
|------|-------------|
| tool-bash-retry-loop.md | Bash retry loop — same failing command without strategy change |
| tool-e2e-graduation-structure.md | E2E graduation — reclassify by domain not feature slug |
| tool-fix-e2e-unknown-placeholder.md | Auto-generated fix tasks with vague "unknown" placeholders |

### Uncategorized

| File | When to Load |
|------|-------------|
| frontend-test-command.md | Running frontend tests — vitest --run flag duplication |
| hook-stop-e2e-blocking.md | Stop hook fires e2e check on unrelated work |
| key-field-design-analysis.md | Key-field design mismatch — 403 on status transitions |
| weekly-view-bug-fixes.md | Week picker selection shows no data |

## By Task Trigger

| Task | Lessons |
|------|---------|
| New entity / schema design | arch-bizkey-vs-internal-id, gotcha-bizkey-vs-id-confusion, key-field-design-analysis |
| Task breakdown | arch-ui-integration-gap, pattern-phase-gate-tasks, gotcha-tech-design-decision-archiving |
| E2E test generation | gotcha-e2e-script-generation, debug-e2e-zero-tests, debug-e2e-beforeall-cascade |
| E2E test graduation | arch-e2e-graduation-staging, tool-e2e-graduation-structure |
| Running /run-tasks | gotcha-subagent-partial-commits, gotcha-ac-self-report-without-verification, arch-task-failure-recovery-loop, gotcha-continue-replay-fresh-agent, gotcha-claim-in-progress-priority, gotcha-task-claim-priority-skip |
| Using subagents | gotcha-no-subagent-for-sequential-work, gotcha-agent-breaks-safety-net, gotcha-parallel-subagent-rate-limit |
| Schema migration | gotcha-schema-alignment-cascading-test-failures |
| Adding interface methods | gotcha-interface-blast-radius-dispatcher |
| Writing tech design | gotcha-design-tech-skip-askuserquestion, gotcha-tech-design-decision-archiving |
| Using eval skills | gotcha-eval-prd-use-zcode-agents |
| Hook configuration | gotcha-hook-unbounded-test-timeout, hook-stop-e2e-blocking |
| Frontend component work | arch-button-icon-convention, weekly-view-bug-fixes |
| Debugging test issues | tool-bash-retry-loop, gotcha-pipe-tail-buffers-test-output, frontend-test-command |
