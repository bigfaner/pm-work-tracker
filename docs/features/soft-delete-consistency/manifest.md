---
feature: "soft-delete-consistency"
status: tasks
---

# Feature: soft-delete-consistency

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 统一为所有 BaseModel 实体的查询方法添加 NotDeleted scope，修复 SubItem 软删除实现和唯一索引 |
| User Stories | prd/prd-user-stories.md | 4 个用户故事：角色列表过滤、子项重建、通用查询防御、权限计算排除 |
| Tech Design | design/tech-design.md | 运行时类型切换 + 统一 TeamMember 软删除 + 8 文件变更清单 |

## Traceability

| PRD Section | Design Section | Tasks |
|-------------|----------------|-------|
| 5.4 #1 通用 helpers | Decision 1: isSoftDeletable[T]() | 1.1 |
| 5.4 #2 User repo | Complete Change List: user_repo.go | 2.1 |
| 5.4 #3 Team repo | Decision 2: Unify TeamMember + Change List | 2.2 |
| 5.4 #4 MainItem repo | Complete Change List: main_item_repo.go | 2.3 |
| 5.4 #5 SubItem repo | Modified SoftDelete + Change List | 2.4, 3.1 |
| 5.4 #6 ItemPool repo | Complete Change List: item_pool_repo.go | 2.4 |
| 5.4 #7 Role repo join 查询 | Complete Change List: role_repo.go | 2.5 |
| 5.4 #8 Schema 修复 | Schema: pmw_sub_items unique index | 3.2 |
| TeamMember FindMember/CountMembers | Decision 2 + Coverage Map | 2.2 |
| TeamMember RemoveMember | Modified: RemoveMember | 3.1 |
