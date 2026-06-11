---
id: "disc-1"
title: "Lesson: E2E test hardcoded IP/port anti-pattern"
priority: "P1"
dependencies: []
status: 
type: "doc"
---

# disc-1: Lesson: E2E test hardcoded IP/port anti-pattern

Record lesson: when fixing E2E test navigation failures (page.goto relative URL fails with 'invalid URL'), the correct fix is to import baseUrl from helpers.js (which reads from config.yaml), NOT to hardcode http://127.0.0.1:5173. Hardcoding breaks port flexibility, multi-env support, and CI. The helpers already export baseUrl with config fallback. Lesson category: testing, gotcha.
