## Eval-contract Complete
**Final Score**: AVG 665/1000 (target: 850)
**Iterations Used**: 1/1

### Score Progression
| Iteration | Score | Delta |
|-----------|-------|-------|
| 1 | 665 avg | — |

### Per-Journey Contract Scores (Iteration 1)

| Journey (Contracts) | Score | Pass? | Completeness | Semantic Purity | Precondition Excl. | Fact Alignment | Surface Fitness | Internal Consistency |
|---------------------|-------|-------|-------------|----------------|-------------------|---------------|----------------|---------------------|
| milestone-lifecycle (6) | 665 | ❌ | 140 | 160 | 95 | 90 | 95 | 85 |
| milestone-map-lifecycle (8) | 665 | ❌ | 130 | 160 | 100 | 105 | 95 | 75 |

### Dimension Threshold Pass/Fail

| Dimension | Min Threshold | Journeys Passing | Journeys Failing |
|-----------|--------------|-----------------|-----------------|
| Completeness (200) | 120 | 2/2 | — |
| Semantic Purity (200) | 120 | 2/2 | — |
| Precondition Exclusivity (150) | 90 | 2/2 | — |
| Fact Alignment (150) | 90 | 2/2 | — |
| Surface Fitness (150) | 90 | 2/2 | — |
| Internal Consistency (150) | 90 | 0/2 | milestone-lifecycle (85), milestone-map-lifecycle (75) |

### Top Issues

1. **Internal Consistency below threshold (2/2)**: Wrong-status/missing-status preconditions, irrelevant invariants copy-pasted across steps, inconsistent state coverage vs PRD state machine.
2. **Missing derived outcomes (2/2)**: server-error, loading-state, cancel-dialog not covered in contracts despite being in journeys.
3. **Implementation coupling (2/2)**: API routes, HTTP methods, field names (assigneeBizKey, milestone_keys) in contract dimensions.
4. **Incomplete fact annotations (2/2)**: Contracts reference code files instead of PRD business rules.

### Outcome
**Target NOT reached** — 0/2 contract sets pass at 850/1000. Average 665/1000. Internal Consistency fails min threshold in both sets.
