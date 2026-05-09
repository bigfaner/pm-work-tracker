---
created: 2026-05-08
author: "faner"
status: Draft
---

# Proposal: 统一权限码校验，移除 bypass 模式

## Problem

权限系统中存在两层 bypass 逻辑，导致自定义角色（如 ext-member）即使拥有对应权限码也会被 403 拒绝。

### Evidence

1. **ext-member 403 bug**（todos.txt #63）：自定义角色 ext-member 拥有 `sub_item:update`，但编辑子事项时 handler 内的 `isPMOrSuperAdmin()` 返回 false，触发 assignee 所有权检查，若非 assignee 则 403。
2. **SuperAdmin bypass 是硬编码短路**：`RequirePermission` 中间件直接 `c.Next()` 跳过，不走权限码。SuperAdmin 角色在 DB 中没有任何权限码记录。
3. **team_handler 中 5 处 "充当 PM" 逻辑**：SuperAdmin 通过 `IsSuperAdmin` 标志获取团队 PM 的 BizKey 来绕过服务层的 PM 身份校验。
4. **progress_handler 使用 pmFlag**：`isPMOrSuperAdmin()` 决定是否能给任何子事项添加进度。

### Urgency

自定义角色是已上线功能（RBAC 系统），但 bypass 模式使自定义角色的权限形同虚设。每新增一个 handler 都需要手动添加 bypass 逻辑，容易遗漏。现在修复可避免更多 bypass 代码累积。

## Proposed Solution

**移除所有 bypass，统一走权限码校验：**

1. **删除 `User.IsSuperAdmin` 字段** -- 从 model、DB schema、VO/DTO、前端类型中移除。SuperAdmin 不再是特殊用户属性，而是一个拥有全部 29 个权限码的预设角色。
2. **`superadmin` 预设角色获得全部权限码** -- 在 RBAC seed 中为 superadmin 角色写入所有 29 个权限码记录。
3. **移除 `isPMOrSuperAdmin()` 函数及其所有调用** -- sub_item_handler 和 progress_handler 不再做 assignee/PM 身份检查。拥有 `sub_item:update` 即可编辑任何子事项；拥有 `progress:create` 即可添加进度。
4. **移除 `RequirePermission` 中的 SuperAdmin 短路** -- 所有请求统一走 permCodes 线性扫描。
5. **TeamScopeMiddleware 保留 superadmin 角色的自动团队注入** -- 检测用户的角色 key 为 superadmin 时，自动注入团队成员身份（不要求 `team_members` 表有记录），保持跨团队访问能力。
6. **team_handler 中的 "充当 PM" 逻辑替换为权限码驱动** -- 拥有 `team:update`/`team:invite`/`team:remove` 等权限码即可执行操作，服务层不再校验 PM 身份，改为校验权限码。
7. **前端同步移除 `isSuperAdmin`** -- auth store、types、PermissionGuard、页面组件全部改用权限码判断。

### User Impact

以下描述各角色在变更前后可观察到的行为差异：

| Persona | Before | After |
|---------|--------|-------|
| **SuperAdmin** | 可执行所有操作（通过硬编码短路）；前端看到 "SuperAdmin" 标签 | 可执行所有操作（通过 superadmin 角色的 29 个权限码）；前端不再显示 "SuperAdmin" 标签，改为基于权限码的 UI 元素可见性。**无功能性变化** |
| **PM** | 所有操作正常 | 所有操作正常。**无变化** |
| **自定义角色（如 ext-member）** | 拥有 `sub_item:update` 但编辑子事项时返回 403；拥有 `sub_item:assign` 但分配负责人时返回 403 | 拥有 `sub_item:update` 即可成功编辑任何子事项；拥有 `sub_item:assign` 即可成功分配负责人。**核心修复** |
| **member** | member 角色已持有 `progress:create`，可为任何子事项添加进度 | 行为不变。**无变化** |

## Alternatives Considered

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| Do nothing | 零成本，无风险 | 自定义角色仍被 bypass 逻辑阻塞；每新增 handler 都需手动维护 bypass；ext-member bug 无法根治 | Rejected: 绕过模式与 RBAC 设计目标矛盾 |
| 仅修复 ext-member bug | 改动小（调整 isPMOrSuperAdmin 判断条件） | 治标不治本；SuperAdmin bypass 仍在；未来其他自定义角色会遇到同类问题 | Rejected: 修补而非根治 |
| 渐进式迁移：Phase 1 保留 IsSuperAdmin 作为只读标记，Phase 2 再移除 | 分阶段降风险 | 两套逻辑并存增加复杂度；迁移窗口期维护成本高 | Rejected: 当前系统仅 4 个预设角色（superadmin/pm/member/ext-member）、29 个权限码、~15 个 handler，一次到位的改动量可控 |

## Scope

### In Scope

- 删除 `User.IsSuperAdmin` 字段（model、migration、VO/DTO）
- superadmin 预设角色 seed 全部 29 个权限码
- 移除 `RequirePermission` 中间件的 SuperAdmin 短路
- 移除 `isPMOrSuperAdmin()` 函数及所有调用点（sub_item_handler、progress_handler）
- 移除 sub_item_handler 中的 assignee 所有权检查
- 替换 team_handler 中 5 处 "SuperAdmin 充当 PM" 逻辑为权限码驱动（Update: L99、Disband: L128、RemoveMember: L194、UpdateMemberRole: L233、TransferPM: L276）
- TeamScopeMiddleware 为 superadmin 角色自动注入团队成员身份
- 服务层（team_service 等）将 PM 身份校验改为权限码校验
- DB migration：删除 `users.is_super_admin` 列，将原 SuperAdmin 用户绑定到 superadmin 角色
- 前端：从 types、auth store、PermissionGuard、页面组件中移除 `isSuperAdmin`，改用权限码
- 更新 `docs/conventions/permission-codes.md` 中的 SuperAdmin Bypass Rule 章节

### Out of Scope

- 新增权限码（当前 29 个够用）
- 角色管理 UI 的变更
- 数据范围（data scope）的变更（all_teams / own_teams / self_only 不变）
- member 角色权限调整
- `todos.txt` #65（member 角色获取权限问题）

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 现有 SuperAdmin 用户迁移失败 | Low | High | migration 中将 `is_super_admin=true` 的用户自动绑定到 superadmin 角色；事务保证原子性 |
| SuperAdmin 角色未加入某团队导致无法访问 | Medium（因 SuperAdmin 可能不属于任何 team_members 记录，而 TeamScopeMiddleware 的自动注入是新增逻辑） | High | TeamScopeMiddleware 对 superadmin 角色自动注入团队身份，无需 `team_members` 记录；新增集成测试验证 SuperAdmin 在未加入任何团队时仍可访问所有团队的资源 |
| 前端遗漏 isSuperAdmin 引用导致运行时错误 | Medium | Medium | Grep 全量扫描后逐个修复；CI 构建时 TypeScript 类型检查会捕获大部分遗漏 |
| 服务层 PM 校验移除后，非 PM 角色持有 team 管理权限码执行 PM-only 操作 | Low | Medium | 经审查 `rbac.go` seed 数据，当前 3 个预设角色中仅 PM 持有 `team:update`、`team:invite`、`team:remove`、`team:delete`、`team:transfer`（member 仅持有 `team:read`）。自定义角色由 PM 创建，PM 自行决定是否授予这些权限码——这是 RBAC 的设计意图，不是回归。若后续需限制，可在角色管理 UI 增加 "仅限 PM 角色" 标记 |
| progress 的 pmFlag 移除后，非 PM 可以为任何子事项添加进度 | Medium | Low（member 角色已持有 `progress:create`，此行为在当前系统中已存在） | 接受此行为作为 RBAC 设计意图的一部分：持有 `progress:create` 的角色可以为任何可见子事项添加进度。若后续需要限制为仅 assignee 可添加进度，需新增 `progress:create_own` 权限码，不在本次范围内 |

## Success Criteria

- [ ] 自定义角色（如 ext-member）拥有 `sub_item:update` 后可编辑任何子事项，不再 403
- [ ] `User.IsSuperAdmin` 字段从 model、DB schema、VO/DTO 中完全移除
- [ ] `isPMOrSuperAdmin()` 函数从代码中完全移除（0 引用）
- [ ] `RequirePermission` 中间件不再有 SuperAdmin 短路逻辑
- [ ] `backend/internal/middleware/` 中 Grep `isSuperAdmin` 或 `IsSuperAdmin` 结果为 0
- [ ] SuperAdmin 角色用户可逐项执行以下操作（等同移除前）：创建团队、更新团队信息、解散团队、邀请成员、移除成员、修改成员角色、转让 PM、创建/读取/编辑/归档事项、创建/读取/编辑/分配/变更状态子事项、创建/读取/编辑进度、提交/审核池、查看周报/甘特图/表格视图、导出报表、查看用户
- [ ] team_handler 的 5 处 "充当 PM" 逻辑全部替换为权限码校验：非 PM 但持有 `team:update` 的角色可更新团队信息，持有 `team:invite` 的角色可邀请/修改角色，持有 `team:remove` 的角色可移除成员，持有 `team:transfer` 的角色可转让 PM
- [ ] 前端 `isSuperAdmin` 引用全部移除，TypeScript 编译通过
- [ ] 现有测试全部通过（含 rbac_test、permission_test、handler 测试）

## Next Steps

- Proceed to `/write-prd` to formalize requirements
