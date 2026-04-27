---
feature: "soft-delete-consistency"
---

# User Stories: Soft-Delete Consistency Fix

## Story 1: 删除的角色不再出现在列表中

**As a** 系统管理员
**I want to** 删除角色后，该角色立即从角色列表和权限分配中消失
**So that** 避免将已删除的角色分配给团队成员，造成权限混乱

**Acceptance Criteria:**
- Given 管理员已删除一个角色（deleted_flag=1 设置成功）
- When 管理员请求角色列表 API
- Then 该角色不出现在响应中

- Given 管理员删除了一个角色
- When 管理员立即刷新角色列表
- Then 该角色不在列表中，点击其 bizKey 返回 404

---

## Story 2: 删除的子项可以重建

**As a** 项目经理
**I want to** 删除子项后，能用相同的 item_code 创建新子项
**So that** 可以纠正错误删除而不会被唯一约束阻止

**Acceptance Criteria:**
- Given 某主项下存在子项，item_code 为 "P001-01"
- When PM 删除该子项
- Then 该子项从子项列表中消失（deleted_flag=1, deleted_time 已设置）

- Given 子项 "P001-01" 已被软删除（deleted_flag=1）
- When PM 在同一主项下创建新子项，item_code 为 "P001-01"
- Then 创建成功，不触发唯一约束错误

---

## Story 3: 通用查询防御未来删除功能

**As a** 开发人员
**I want to** 通用查询 helpers（FindByID/FindByIDs）自动过滤已删除记录
**So that** 未来为 User/MainItem/ItemPool 添加删除功能时，不需要修改查询层

**Acceptance Criteria:**
- Given User 实体的 deleted_flag=1
- When 调用 `FindByID[User]` 查询该记录
- Then 返回 NotFound 错误

- Given ProgressRecord 实体（不含 deleted_flag）
- When 调用 `FindByID[ProgressRecord]` 查询该记录
- Then 正常返回，不报 SQL 错误

---

## Story 4: 已删除团队成员不参与权限计算

**As a** 系统管理员
**I want to** 软删除的团队成员不参与权限检查和角色计数
**So that** 移除成员后其权限立即失效，角色成员计数准确

**Acceptance Criteria:**

**AC1 — HasPermission 排除已删除成员**
- Given 团队成员已被软删除（deleted_flag=1），且其角色包含某权限码
- When 系统检查该用户是否有该权限（HasPermission）
- Then 返回 false，不包含该权限

**AC2 — GetUserTeamPermissions 排除已删除成员**
- Given 用户 A 属于团队 T1（角色含 permission_x）和团队 T2（角色含 permission_y），其中 T1 中的成员记录已软删除（deleted_flag=1）
- When 系统查询用户 A 的所有团队权限（GetUserTeamPermissions）
- Then 返回结果仅包含 T2 的 permission_y，不包含 T1 的 permission_x

**AC3 — CountMembersByRoleID 不计入已删除成员**
- Given 某角色下有 3 个团队成员，其中 1 个已被软删除（deleted_flag=1）
- When 系统统计该角色的成员数（CountMembersByRoleID）
- Then 返回 2，而非 3

**AC4 — 边界条件：同一角色下多个已删除成员**
- Given 某角色下有 5 个团队成员，其中 3 个已被软删除
- When 调用 CountMembersByRoleID
- Then 返回 2

**AC5 — 边界条件：已删除成员的所有团队权限均被排除**
- Given 用户 A 在团队 T1 和 T2 中均为成员，用户 A 被软删除
- When 调用 GetUserTeamPermissions(userID=A)
- Then 返回空 map，不包含任何团队的权限
