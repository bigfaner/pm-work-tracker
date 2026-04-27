# 软删除表的唯一索引必须包含 deleted_flag + deleted_time

**日期:** 2026-04-28
**状态:** 已实施

## 背景

软删除（soft-delete）场景下，业务唯一字段的值在删除后可能需要复用。例如：
- 用户 A（username=zhangsan）被软删除后，新用户 B 可以用 `zhangsan` 注册
- 团队邀请码 `ABC123` 软删除后可被新团队复用

如果唯一索引仅覆盖业务字段，已删除记录会阻止新记录使用相同值。

## 决策

所有 soft-delete 表的业务唯一索引（`uk_biz_key` 除外）必须追加 `deleted_flag, deleted_time` 两列。

`biz_key` 是全局唯一业务键（系统生成），不受软删除影响，因此 `uk_biz_key` 不需要追加。

## 实施清单

以下索引已按此规则建立，SQLite 与 MySQL schema 一致：

| 表 | 唯一索引 | 包含 deleted 列 |
|---|---|---|
| pmw_users | `uk_users_username_deleted(username, deleted_flag, deleted_time)` | ✅ |
| pmw_teams | `uk_teams_code_deleted(team_code, deleted_flag, deleted_time)` | ✅ |
| pmw_team_members | `uk_team_user_deleted(team_key, user_key, deleted_flag, deleted_time)` | ✅ |
| pmw_main_items | `uk_main_items_team_code_deleted(team_key, item_code, deleted_flag, deleted_time)` | ✅ |
| pmw_sub_items | `uk_sub_items_main_code(main_item_key, item_code, deleted_flag, deleted_time)` | ✅ |
| pmw_roles | `uk_roles_name(role_name, deleted_flag, deleted_time)` | ✅ |
| pmw_item_pools | 仅有 `uk_biz_key`，无业务唯一约束 | N/A |

**不适用此规则的表：**
- `pmw_progress_records` — append-only，无软删除
- `pmw_status_histories` — append-only，无软删除
- `pmw_role_permissions` — 无软删除列，角色删除时权限级联清理
