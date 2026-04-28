---
feature: "permission-granularity"
---

# User Stories: 细化 user / role 权限粒度

## Story 1: member 在团队管理页查询角色列表

**As a** 持有 `member` 角色的团队成员
**I want to** 在团队管理页添加新成员时，能够看到角色下拉列表并选择角色
**So that** 我可以完成添加成员的操作，而不需要被授予角色管理权限

**Acceptance Criteria:**
- Given 我持有 `role:read` 权限但不持有 `role:create/update/delete`
- When 我打开团队管理页的"添加成员"对话框
- Then 角色下拉列表正常加载，显示所有可用角色

- Given 我持有 `role:read` 权限
- When 我尝试访问 `GET /admin/roles`
- Then 返回 200 和角色列表

- Given 我不持有 `role:read` 权限
- When 我尝试访问 `GET /admin/roles`
- Then 返回 403，角色下拉列表不可见

---

## Story 2: 自定义角色用户使用成员选择器

**As a** 持有自定义 `project_manager` 角色的用户
**I want to** 在任务分配时使用成员选择器列出团队成员
**So that** 我可以完成任务分配，而不会意外获得查看用户敏感信息（邮箱、手机号）的权限

**Acceptance Criteria:**
- Given 我持有 `user:list` 权限但不持有 `user:read`
- When 我访问 `GET /admin/users`
- Then 返回 200 和用户列表（不含敏感字段详情）

- Given 我持有 `user:list` 但不持有 `user:read`
- When 我尝试访问 `GET /admin/users/:userId`
- Then 返回 403，无法获取用户详情

- Given 我同时持有 `user:list` 和 `user:read`
- When 我访问 `GET /admin/users/:userId`
- Then 返回 200 和完整用户详情（含邮箱、手机号等字段）

---

## Story 3: pm 管理角色定义

**As a** 持有 `pm` 角色的项目经理
**I want to** 创建、编辑、删除自定义角色，并为角色分配权限码
**So that** 我可以灵活配置团队成员的权限组合

**Acceptance Criteria:**
- Given 我持有 `role:create` 权限
- When 我提交 `POST /admin/roles`（角色名称合法，权限码有效）
- Then 新角色创建成功，返回 201

- Given 我持有 `role:update` 权限
- When 我提交 `PUT /admin/roles/:id`
- Then 角色信息更新成功，返回 200

- Given 我持有 `role:delete` 权限，且目标角色无成员绑定
- When 我提交 `DELETE /admin/roles/:id`
- Then 角色删除成功，返回 200

- Given 我只持有 `role:read` 权限
- When 我尝试 `POST /admin/roles`
- Then 返回 403

---

## Story 4: pm 查看用户信息和分配角色

**As a** 持有 `pm` 角色的项目经理
**I want to** 查看用户列表、查看用户详情，并为用户分配角色
**So that** 我可以完成用户管理工作

**Acceptance Criteria:**
- Given 我持有 `user:list` 权限
- When 我访问 `GET /admin/users`
- Then 返回 200 和用户列表

- Given 我持有 `user:read` 权限
- When 我访问 `GET /admin/users/:userId`
- Then 返回 200 和完整用户详情

- Given 我持有 `user:assign_role` 权限
- When 我提交 `POST /admin/users`（为用户分配角色）
- Then 操作成功，返回 200

---

## Story 5: 管理员执行数据迁移

**As a** 系统管理员（运维）
**I want to** 通过迁移脚本将存量角色的旧权限码转换为新权限码
**So that** 现有用户的权限在迁移后语义等价，不出现功能中断

**Acceptance Criteria:**
- Given 数据库中存在持有 `user:manage_role` 的自定义角色
- When 步骤一迁移脚本执行完成
- Then 该角色的 `user:manage_role` 被替换为 `role:create` + `role:update` + `role:delete`，原码不再存在

- Given 数据库中存在持有旧语义 `user:read` 的角色
- When 步骤一迁移脚本执行完成
- Then 该角色的旧 `user:read` 被替换为 `user:list`，同时原子写入新 `user:read`

- Given 迁移脚本执行过程中发生错误
- When 事务回滚
- Then 数据库恢复到迁移前状态，无部分迁移残留

- Given 步骤一迁移完成
- When CI grep 断言运行
- Then 代码库中 `user:manage_role` 和旧语义 `user:read` 引用数为零，断言通过

---

## Story 6: 前端按权限码控制 UI 元素可见性

**As a** 任意角色的登录用户
**I want to** 只看到我有权限操作的 UI 元素（按钮、菜单、页面入口）
**So that** 界面简洁，不出现点击后报错的无效操作入口

**Acceptance Criteria:**
- Given 我不持有 `role:create` 权限
- When 我访问角色管理页
- Then "创建角色"按钮不显示

- Given 我不持有 `user:list` 权限
- When 我访问系统
- Then 用户管理页入口不显示

- Given 我持有 `role:read` 但不持有 `user:list`
- When 我打开团队管理页的"添加成员"对话框
- Then 角色下拉列表正常显示，但用户管理页入口不显示

- Given 前端隐藏了某个操作入口
- When 用户绕过前端直接调用对应 API
- Then 后端中间件返回 403（后端是安全防线，前端隐藏仅为 UX 优化）
