---
feature: "API Permission Test Coverage"
---

# User Stories: API Permission Test Coverage

## Story 1: 权限差异端点的单元测试

**As a** 开发者
**I want to** 对每个权限敏感端点运行"有权限 → 200 / 无权限 → 403"的单元测试
**So that** 路由层的 `perm()` 绑定错误能在 CI 中被立即发现，而不是逃逸到手工测试

**Acceptance Criteria:**
- Given 一个 handler 测试，通过 `c.Set("permCodes", []string{"main_item:archive"})` 注入权限
- When 调用 `POST /archive` 端点
- Then 响应码为 200

- Given 一个 handler 测试，注入空 `permCodes`（`[]string{}`）
- When 调用同一端点
- Then 响应码为 403

---

## Story 2: 预设角色矩阵的集成验证

**As a** 开发者
**I want to** 用真实 DB 和完整路由验证 superadmin/pm/member 三个预设角色的访问矩阵
**So that** 预设角色的权限配置与代码中的 `perm()` 绑定保持一致，任何不一致都能被测试捕获

**Acceptance Criteria:**
- Given 三个用户分别绑定 superadmin / pm / member 角色，团队已创建，目标资源已存在
- When 各用户调用 `POST /teams/:teamId/main-items/:id/archive`（需要 `main_item:archive`，member 无此权限）
- Then superadmin 和 pm 返回 200，member 返回 403

- Given 同上设置
- When 各用户调用 `POST /teams/:teamId/members`（需要 `team:invite`，member 无此权限）
- Then superadmin 和 pm 返回 200，member 返回 403

---

## Story 3: 自定义角色权限组合验证

**As a** 开发者
**I want to** 验证自定义角色（部分权限组合）的访问控制，以及权限变更后即时生效
**So that** 管理员在线调整角色权限后，系统行为立即反映变更，无需重启或缓存失效

**Acceptance Criteria:**
- Given 自定义角色仅有 `main_item:read` + `progress:read`，用户绑定该角色
- When 用户调用 `GET /main-items`
- Then 返回 200

- Given 同上
- When 用户调用 `POST /main-items`（需要 `main_item:create`）
- Then 返回 403

- Given 管理员为该角色新增 `main_item:create` 权限
- When 用户使用同一 token 立即调用 `POST /main-items`（不重新登录，不重新获取 token）
- Then 返回 200（权限从 DB 实时读取，无缓存层介入）

---

## Story 4: 权限边界场景验证

**As a** 代码审查者
**I want to** 确认空权限角色、superadmin 绕过、401 vs 403 三个边界场景均有测试覆盖
**So that** 权限系统的安全边界在代码审查时可以通过测试结果而非人工推断来验证

**Acceptance Criteria:**
- Given 用户绑定空权限角色（无任何权限码）
- When 调用任意受保护端点
- Then 返回 403

- Given superadmin 用户（绕过所有权限检查），目标资源已存在
- When 调用任意受保护端点
- Then 返回 200（不是权限拒绝；若返回 404/500 视为 fixture 缺失，不算通过）

- Given 请求携带无效 token
- When 调用任意端点
- Then 返回 401（而非 403，明确区分认证失败与授权失败）

---

## Story 5: 权限码覆盖率 CI 断言

**As a** 代码审查者
**I want to** 在 CI 中自动验证 `codes.go` 里每个权限码都出现在测试矩阵中
**So that** 新增权限码时若遗漏测试，CI 立即失败，不需要人工检查

**Acceptance Criteria:**
- Given `codes.go` 中定义了权限码 `foo:bar`，但该字符串未作为 `permCodes` 参数或测试矩阵值出现在测试文件中（注释或日志中的出现不计）
- When CI 运行权限码覆盖率断言步骤
- Then 构建失败，输出 `missing test coverage for: foo:bar`

- Given `codes.go` 中所有权限码均在 `middleware/permission_test.go` 或 `tests/integration/rbac_test.go` 中出现
- When CI 运行权限码覆盖率断言步骤
- Then 断言通过，构建继续
