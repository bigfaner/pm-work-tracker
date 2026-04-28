---
feature: "bizkey-unification"
status: prd
---

# Feature: bizkey-unification

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 消除 Service 层 uint/int64 ID 混用问题，修复进度记录 team_key 数据污染，统一以 int64 bizKey 作为跨层边界标识符 |
| User Stories | prd/prd-user-stories.md | 3 个故事：进度记录写入正确 team_key、角色权限判断使用正确 bizKey、编译器约束防止同类 Bug |

## Traceability

| PRD Section | Design Section | UI Component | Tasks |
|-------------|----------------|--------------|-------|
| 5.4 关联性需求改动 | — | — | — |
