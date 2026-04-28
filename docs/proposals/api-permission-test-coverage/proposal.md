---
created: 2026-04-28
author: "faner"
status: Draft
---

# Proposal: 后端 API 角色权限校验测试补全

## Problem

当前 RBAC 权限体系已实现（中间件、路由绑定、预设角色），但测试覆盖存在结构性空白：

1. **中间件单元测试**（`permission_test.go`）只验证 mock context 注入，不验证真实角色配置下的 HTTP 响应。
2. **RBAC 集成测试**（`rbac_test.go`）只覆盖迁移幂等性和角色 CRUD，没有任何"角色 X 调用端点 Y 得到 200/403"的断言。
3. **自定义角色**（部分权限组合）完全没有测试覆盖。

结果：`permission_test.go` 的 11 个测试均通过 mock context 注入权限码，不经过路由层；`rbac_test.go` 的 21 个测试覆盖迁移和角色 CRUD，但路由层共有 53 处 `perm()` 绑定，其中任何一处权限码误写（如将 `perm("report:export")` 误写为 `perm("report:read")`）均无法被现有测试发现。

**紧迫性触发点**：commit `3200bdc`（`fix(rbac): sync missing permissions on existing roles during migration`）新增了 2 个权限码，但没有任何对应的路由层测试覆盖——这是一次已发生的盲区扩大，而非假设风险。当前 bizkey-unification 功能（`docs/features/bizkey-unification/`）正在推进，预计将引入更多权限码，若不在此时补全测试框架，每次新增权限都将重复同样的盲区。

## Proposed Solution

在两个层面补充测试，形成互补覆盖：

### 层 1：Handler 单元测试（mock permCodes）

在各 handler 的 `_test.go` 文件中，针对权限敏感操作补充"有权限 → 200、无权限 → 403"的表驱动测试。

**机制**：通过 `c.Set("permCodes", []string{...})` 直接注入权限集合，绕过 DB，专注验证中间件与 handler 的集成边界。

**覆盖目标**（每个端点 2 个 case：有权限 / 无权限）：

| 端点 | 所需权限 | pm 有 | member 有 |
|------|---------|-------|-----------|
| POST /main-items | main_item:create | ✓ | ✓ |
| POST /main-items/:id/archive | main_item:archive | ✓ | ✗ |
| PUT /main-items/:id/status | main_item:change_status | ✓ | ✗ |
| POST /members | team:invite | ✓ | ✗ |
| DELETE /members/:userId | team:remove | ✓ | ✗ |
| PUT /pm | team:transfer | ✓ | ✗ |
| POST /sub-items/:id/progress | progress:create | ✓ | ✓ |
| PATCH /progress/:id/completion | progress:update | ✓ | ✗ |
| POST /item-pool | item_pool:submit | ✓ | ✓ |
| POST /item-pool/:id/assign | item_pool:review | ✓ | ✗ |
| GET /views/weekly | view:weekly | ✓ | ✓ |
| GET /reports/weekly/export | report:export | ✓ | ✗ |

### 层 2：集成测试（真实 DB + HTTP）

在 `rbac_test.go` 中新增一组测试，使用真实 SQLite DB 和完整路由，验证端到端权限行为。

**测试场景**：

**A. 预设角色矩阵验证**
- 创建 3 个用户，分别绑定 superadmin / pm / member 角色
- 对代表性端点发起请求，断言响应码符合权限矩阵
- 覆盖：main_item（创建/归档）、team（邀请/移除）、progress（追加/修正）、item_pool（提交/审核）、view（周视图）、report（导出）

**B. 自定义角色验证**
- 创建自定义角色，仅赋予 `main_item:read` + `progress:read`
- 验证：该角色可读主事项（200），不可创建（403），不可归档（403）
- 验证：修改角色权限后，下次请求立即生效（无缓存问题）

**C. 权限边界场景**
- 空权限角色：所有受保护端点返回 403
- superadmin：所有端点返回非 403（绕过权限检查）
- 无效 token：返回 401（区分认证失败与授权失败）

## Alternatives Considered

**不做任何测试（do nothing）**：零实现成本，不占用当前迭代资源。代价是：任何 `perm()` 误写在生产环境前完全不可见，`3200bdc` 已经证明这种盲区会随每次权限扩展静默累积；随着 bizkey-unification 推进，未覆盖的权限码数量只会增加，修复成本随时间线性上升。

**仅补充单元测试**：实现快（预计 ~1 天），mock 注入不依赖 DB，CI 耗时几乎不变。但单元测试绕过路由层，无法发现 `perm("team:invite")` 误写为 `perm("team:read")` 这类路由绑定错误——而这正是当前最大的盲区所在。

**仅补充集成测试**：覆盖最真实，能端到端验证路由绑定。但每个 case 需要完整 DB 初始化，基于现有 20+ 条迁移，SQLite 冷启动约需 2–5 秒；handler 内部逻辑错误难以定位，调试成本高于单元测试。

**全量路由矩阵测试**：对 53 处 `perm()` 绑定 × 3 角色穷举，理论上覆盖最完整。但每新增一个权限码需同步维护约 6 个测试 case，按当前迭代速度估算每季度新增 ~10 个权限码，维护成本约 60 个 case/季度，远超收益。

**推荐方案**：两层互补——单元测试覆盖权限边界逻辑，集成测试覆盖代表性业务场景和自定义角色。单元测试快速定位逻辑错误，集成测试兜底路由绑定错误，两者合计约 3 个工作日，是 do nothing 与全量矩阵之间成本最低的有效覆盖方案。

## Scope

**In scope:**
- Handler 单元测试：为上表 12 个端点各补充 2 个权限 case（有/无），共 24 个 case
- 集成测试 A：预设角色矩阵（superadmin / pm / member × 代表性端点）
- 集成测试 B：自定义角色（部分权限组合 + 权限变更即时生效）
- 集成测试 C：边界场景（空权限、superadmin 绕过、401 vs 403）

**预计工作量**：单元测试 ~1 个工作日，集成测试 ~2 个工作日，合计 ~3 个工作日。

**Out of scope:**
- 前端权限渲染测试
- 性能/并发测试
- 数据权限（data scope）测试
- 非权限相关的业务逻辑测试

## Risks

1. **测试 DB 初始化成本**：集成测试每个 case 需要完整 RBAC 迁移，可能拖慢 CI。
   - 可能性：Medium（项目已有 20+ 条迁移，SQLite 在 CI 环境下每次冷启动约需数秒）；影响：Low（仅影响 CI 耗时，不影响正确性）。
   - 缓解：用 `TestMain` 共享一次迁移，各 case 用事务回滚隔离数据。

2. **权限矩阵维护**：新增权限码时需同步更新测试表格。
   - 可能性：High（权限码仍在迭代，历史上已多次新增）；影响：Medium（遗漏更新会导致新权限无测试覆盖，静默引入盲区）。
   - 缓解：在测试文件顶部维护权限矩阵常量，与 `codes.go` 对齐，CI 中加断言验证覆盖率。

3. **Handler mock 复杂度**：部分 handler 依赖多个 service，mock 成本高。
   - 可能性：Low（权限测试只需触发中间件，不需要 service 真正执行）；影响：Low（最坏情况是个别 handler 测试编写耗时增加，不影响覆盖目标）。
   - 缓解：权限测试只需验证 403 响应，不需要 service 真正执行，mock 返回零值即可。

## Success Criteria

- [ ] 所有新增单元测试通过：`go test ./internal/handler/... -run TestPermission`
- [ ] 所有新增集成测试通过：`go test ./tests/integration/... -run TestRBACPermission`
- [ ] 预设角色权限矩阵（pm/member 差异端点）100% 有测试覆盖
- [ ] 自定义角色场景有至少 1 个完整流程测试（创建角色 → 分配 → 验证访问 → 修改权限 → 再验证）
- [ ] 无权限返回 403、未认证返回 401，两者有明确区分的测试 case
- [ ] superadmin 角色对所有受保护端点返回非 403，有至少 1 个集成测试 case 覆盖
- [ ] 空权限角色对所有受保护端点返回 403，有至少 1 个集成测试 case 覆盖
