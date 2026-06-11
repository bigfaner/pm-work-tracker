---
id: "T-clean-code"
title: "Simplify and Clean Code"
priority: "P2"
estimated_time: "20min"
dependencies: ["4.gate"]
type: "code-quality.simplify"
surface-key: ""
surface-type: ""
---

## Acceptance Criteria
- [ ] All changed files pass compile, fmt, lint checks
- [ ] No dead code or unnecessary complexity introduced by this feature

Simplify and clean up code for the milestone-map feature.

## Discovery Strategy
1. Run `git diff --name-only main...HEAD` to identify files changed by this feature
2. Focus cleanup on changed files only
3. The skill resolves scope: git diff > feature context > user-specified paths

Do NOT clean up files outside this feature's scope.
