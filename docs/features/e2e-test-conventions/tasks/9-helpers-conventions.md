---
id: "9"
title: "更新 Helpers 规范和 conventions 索引（index.md）"
priority: "P2"
estimated_time: "30m"
dependencies: [7, 8]
type: "doc"
mainSession: false
---

# 9: 更新 Helpers 规范和 conventions 索引（index.md）

## Description

在 `docs/conventions/testing/index.md` 中追加 Helpers 工具库规范（E-120~E-123），涵盖配置来源、Token 缓存、子进程执行、调试截图。同时更新 index.md 的 `domains` frontmatter。

## Reference Files
- `docs/proposals/e2e-test-conventions/proposal.md` — 具体规范清单 > Helpers 工具库规范（index.md）
- `docs/conventions/testing/index.md`: 追加 E-120~E-123 规范 + 更新 domains frontmatter (ref: 具体规范清单 > Helpers 工具库规范)

## Affected Files

### Create
| File | Description |
|------|-------------|
| — | — |

### Modify
| File | Changes |
|------|---------|
| `docs/conventions/testing/index.md` | 追加 4 条 helpers 规范（E-120~E-123）+ 更新 domains frontmatter |

### Delete
| File | Reason |
|------|--------|
| — | — |

## Acceptance Criteria
- [ ] `index.md` 包含 4 条 helpers 规范（E-120~E-123），每条含 ID、规则名、说明
- [ ] `index.md` 的 `domains` frontmatter 已更新（含 testing-helpers, config, token 等关键词）

## Implementation Notes
- 此任务是 conventions 文档更新的最后一步，确保所有三个文件（api/core.md, web/core.md, index.md）的规范编号连续、无重叠
