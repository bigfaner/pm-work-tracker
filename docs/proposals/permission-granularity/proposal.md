---
created: 2026-04-28
author: "faner"
status: Draft
---

# Proposal: 细化 user / role 权限粒度

## Problem

当前权限码在 `user` 和角色管理两个维度上粒度过粗，导致两个实际问题：

1. **`user:manage_role` 把"查看角色列表"和"管理角色定义"捆绑在一起。** 在团队管理页面添加成员时，需要查询角色列表（下拉选择），但这不应该要求用户拥有管理角色的权限。目前无法给普通用户授予"只读角色列表"的权限。

2. **`user:read` 把"列出用户"和"查看用户详情"捆绑在一起。** 持有自定义 `project_manager` 角色的用户需要在任务分配时使用成员选择器（调用 `GET /admin/users`），但该角色不应能访问用户管理页面——`GET /admin/users/:userId` 返回完整用户信息，含邮箱、手机号等敏感字段。目前无法区分这两种访问：授予 `user:read` 则同时开放了用户详情接口；不授予则成员选择器不可用。实际绕过方式是授予 `user:read` 并在前端隐藏用户管理页入口，但后端接口仍可直接访问，权限控制形同虚设。

这导致：要么给用户过多权限（授予 `user:manage_role` 才能查角色列表），要么功能受限（无法在团队管理页正常选择角色）。

当前受影响的典型用户是持有 `member` 角色的团队成员：他们需要在团队管理页面添加新成员，该流程依赖角色下拉列表查询，但 `member` 角色默认不含 `user:manage_role`，导致角色列表不可见、添加成员流程中断。目前的实际绕过方式是将 `user:manage_role` 临时授予 `member` 角色，这同时开放了创建、编辑、删除角色定义的权限——远超实际需要。

## Proposed Solution

### 权限码变更

**user 资源（拆分 + 重命名）：**

| 旧权限码 | 新权限码 | 说明 |
|---------|---------|------|
| `user:read` | `user:list` | 列出用户（用于成员选择器等） |
| — | `user:read` | 查看用户详情（用户管理页） |
| `user:update` | `user:update` | 编辑用户信息（不变） |
| `user:manage_role` | `user:assign_role` | 给用户分配角色（原语义保留，重命名更准确） |

> **`user:read` 是硬重命名，不是复用。** 旧 `user:read`（列出用户）和新 `user:read`（查看用户详情）是不同语义。任何持有旧 `user:read` 的代码路径、前端权限守卫或缓存角色，在迁移后将静默执行"查看详情"语义——这是一个无报错的破坏性变更。必须采用两步上线：步骤一将所有存量 `user:read` 授权显式转换为 `user:list` 并验证无残留引用；步骤二才启用新 `user:read`。两步之间设置废弃窗口，CI 中添加 grep 断言确保旧语义引用清零。

**新增 role 资源（从 user 中独立出来）：**

| 权限码 | 说明 |
|-------|------|
| `role:read` | 查看角色列表和详情（团队管理页添加成员时需要） |
| `role:create` | 创建新角色 |
| `role:update` | 编辑角色名称、描述、权限码 |
| `role:delete` | 删除自定义角色 |

**关键设计原则：** `role:read` 是独立权限，拥有它不代表可以访问用户管理页面。用户管理页面的访问需要 `user:list` 或 `user:read`。

### 预置角色权限调整

**pm 角色新增：**
- `user:list`、`user:read`、`user:update`、`user:assign_role`
- `role:read`、`role:create`、`role:update`、`role:delete`

**member 角色：** 默认不含 user/role 权限。管理员可按需单独授予 `role:read`（如需要在团队管理页选择角色）。

**superadmin：** 绕过所有权限检查，无需变更。

### 后端变更范围

1. `permissions/codes.go`：更新 `user` 资源权限码，新增 `role` 资源
2. `migration/rbac.go`：更新 `seedPresetRoles` 中 pm/member 的权限码列表
3. 路由中间件：将原来绑定 `user:manage_role` 的接口改为对应的 `role:*` 权限码，具体映射如下：

   | 路由 | 旧权限码 | 新权限码 |
   |------|---------|---------|
   | `GET /admin/roles` | `user:manage_role` | `role:read` |
   | `GET /admin/roles/:id` | `user:manage_role` | `role:read` |
   | `POST /admin/roles` | `user:manage_role` | `role:create` |
   | `PUT /admin/roles/:id` | `user:manage_role` | `role:update` |
   | `DELETE /admin/roles/:id` | `user:manage_role` | `role:delete` |
   | `GET /admin/permissions` | `user:manage_role` | `role:read` |
   | `POST /admin/users` | `user:manage_role` | `user:assign_role` |
   | `GET /admin/users` | `user:read`（旧语义） | `user:list` |
   | `GET /admin/users/:userId` | `user:read`（旧语义） | `user:read`（新语义） |
   | `GET /admin/teams` | `user:read`（旧语义） | `user:list` |
4. 数据迁移：
   - `user:manage_role` → `role:create` + `role:update` + `role:delete`。映射依据：`user:manage_role` 的原始语义是"管理角色定义"，包含增删改三项操作，因此持有者应获得全部三个新码。若存在仅需只读角色列表的自定义角色，管理员应在迁移后手动移除多余的写权限，或在迁移前确认。
   - `user:read` → `user:list`（第一步上线时执行）；新 `user:read` 在第二步上线时按需授予，不自动继承。

### 前端变更范围

1. 权限判断逻辑：将 `user:manage_role` 替换为对应的 `role:*` 检查
2. 团队管理页：查询角色列表的权限检查改为 `role:read`
3. 用户管理页：入口权限检查改为 `user:list`

## Alternatives Considered

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| **方案 A：独立 role 资源 + user 细化（本方案）** | 语义清晰，职责分离，可独立授权 | 需要数据迁移，路由绑定需更新 | ✅ 采用 |
| 方案 B：只在 user 下加 `user:list_roles` | 改动最小，无需新资源 | role 管理操作（create/update/delete）仍挂在 `user` 资源下。具体失败场景：若未来需要"可创建角色但不可管理用户"的权限组合，仍无法表达——因为 `user:manage_role` 同时覆盖两者。每新增一个 role 操作，都要在 `user` 资源下打补丁，耦合问题在下一次需求时复现。 | ❌ 放弃 |
| 方案 C：向后兼容别名——保留旧 `user:read` 作为别名，同时映射到 `user:list` 和新 `user:read` 语义，过渡期结束后再删除 | 消费方无需立即迁移；旧码在过渡期内继续有效，不产生 403 | 别名本质上是"一个码授予两种权限"，这与本次拆分的目标相悖——持有别名的角色仍然无法区分"只列出用户"和"查看用户详情"。更关键的是，别名需要在中间件层做特殊分支处理，且必须在所有权限检查点同步维护，增加了实现复杂度；若清理不彻底，别名可能永久残留，拆分目标永远无法完成。本方案的两步上线已通过原子授予（步骤一同步写入新 `user:read`）解决了过渡期 403 问题，无需引入别名机制。 | ❌ 放弃 |
| 方案 D：分阶段上线（先加新码，再废弃旧码） | 无硬切换风险；旧码在过渡窗口内仍有效，消费方可按自己节奏迁移 | 过渡期内两套权限码并存，CI 断言需同时维护两套白名单；若清理不彻底，旧码可能永久残留。本方案的两步上线实际上已采用此思路处理 `user:read` 语义碰撞；对 `user:manage_role` → `role:*` 的迁移，额外保留别名会增加中间件复杂度，收益有限。 | ❌ 放弃（核心思路已内化到方案 A 的两步上线中） |

## Scope

### In Scope

- `permissions/codes.go` 权限码注册表更新
- `migration/rbac.go` 预置角色权限码同步
- 路由中间件权限码绑定更新
- 数据迁移：存量角色权限码替换
- 前端权限判断逻辑更新
- 内部文档更新：`docs/` 中所有列出权限码的文档（架构说明、API 参考、onboarding 指南）同步替换为新权限码

### Out of Scope

- 用户管理页面 UI 改版
- 角色管理页面 UI 改版
- 新增权限管理相关的 API 接口（现有 6 个接口已覆盖全部 role:* 操作：`GET /admin/roles`、`GET /admin/roles/:id`、`POST /admin/roles`、`PUT /admin/roles/:id`、`DELETE /admin/roles/:id`、`GET /admin/permissions`，本次变更仅重新绑定权限码，无需新增路由）
- 审计日志

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| `user:read` 语义碰撞：旧含义"列出用户"，新含义"查看用户详情"，持有旧权限的代码路径/前端守卫/缓存角色将静默执行错误语义 | High | High | 两步上线（见方案说明）；步骤一上线后 CI 中加 grep 断言，确认零残留旧语义引用后才执行步骤二；回滚方案：将 DB 中所有 `user:read` 恢复为 `user:list`，并回退代码至步骤一状态 |
| 数据迁移遗漏导致现有自定义角色权限丢失 | Low | High | 迁移脚本在事务中执行；迁移前对 `role_permissions` 表做快照备份；回滚脚本从快照恢复原始权限码行 |
| 前端权限码引用未全部更新导致功能不可用 | Medium | Medium | CI 中加 grep 断言：`user:manage_role` 和旧语义 `user:read` 引用数为零时才允许合并 |
| `VerifyPresetRoleCodes` 测试因权限码变更失败 | High | Low | 同步更新测试中的期望权限码列表 |
| JWT token 中缓存旧权限码导致迁移后权限未生效 | Low | Medium | 当前实现中 JWT Claims 仅含 `userID` 和 `username`，权限码在每次请求时由 `TeamScopeMiddleware` 从 DB 实时加载，不嵌入 token。因此本项目无此风险。但若未来引入权限码缓存层（Redis 等），迁移时必须同步清除缓存，或将缓存 TTL 设为不超过迁移窗口时长。 |
| 两步上线期间用户失去用户详情页访问权 | Medium | Medium | 步骤一将所有 `user:read` 迁移为 `user:list`，步骤二才按需授予新 `user:read`。两步之间，原本依赖旧 `user:read` 访问 `GET /admin/users/:userId` 的用户将收到 403。缓解方案：在步骤一的数据迁移脚本中，对所有持有旧 `user:read` 的角色同步写入新 `user:read`，即原子授予，无需等到步骤二手动操作。 |

## Success Criteria

- [ ] `role:read` 可单独授予用户，不附带用户管理页访问权
- [ ] 团队管理页添加成员时，拥有 `role:read` 的用户可正常查询角色列表
- [ ] pm 预置角色包含完整的 `user:*` 和 `role:*` 权限
- [ ] 现有自定义角色的权限在迁移后语义等价
- [ ] 后端路由中间件全部使用新权限码，无残留 `user:manage_role` 引用
- [ ] 新增至少 4 个针对 `role:*` 权限码的路由中间件测试，分别覆盖 `role:read`、`role:create`、`role:update`、`role:delete` 四个操作，验证无对应权限时返回 403

## Next Steps

- Proceed to `/write-prd` to formalize requirements
