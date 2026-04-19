---
feature: "rbac-permissions"
---

# User Stories: RBAC 权限体系

## Story 1: 超级管理员在线管理角色

**As a** 超级管理员
**I want to** 在管理界面创建、编辑和删除角色，并为角色勾选权限码
**So that** 我可以灵活组合权限，适应业务变化，无需等待开发者改代码发版

**Acceptance Criteria:**
- Given 我已以超级管理员身份登录
- When 我进入角色管理页面
- Then 我可以看到所有系统角色列表（名称、描述、权限数量、使用人数）

- Given 我在角色管理页面
- When 我点击"创建角色"并填写名称、描述，勾选权限码后提交
- Then 新角色创建成功，出现在角色列表中，立即可用于邀请成员时选择

- Given 我选择了一个已有角色（非 superadmin）
- When 我修改其权限勾选并保存
- Then 该角色的权限即时更新，所有使用该角色的用户权限同步生效

- Given 我选择了一个已有角色（非 superadmin）
- When 我修改其名称和描述并保存
- Then 角色信息更新成功

- Given 我选择了一个自定义角色（非预置角色）
- When 该角色没有用户使用
- Then 我可以删除该角色

- Given 我选择了一个角色
- When 该角色有用户正在使用
- Then 系统提示"该角色正在被 N 个用户使用，无法删除"

- Given 我查看 superadmin 预置角色
- When 我尝试编辑或删除
- Then 系统不允许操作，编辑和删除按钮不可用

- Given 我查看 pm 或 member 预置角色
- When 我尝试编辑
- Then 我可以修改权限勾选，但不能修改角色名称，也不能删除

- Given 我在角色管理页面
- When 我点击"查看权限列表"
- Then 系统展示所有权限码，按资源分组，只读不可编辑

---

## Story 2: PM 在邀请成员时指定角色

**As a** PM（团队负责人）
**I want to** 在邀请用户加入团队时为其指定角色
**So that** 不同成员根据其职责拥有不同的操作权限，团队管理更灵活

**Acceptance Criteria:**
- Given 我在团队成员管理页面
- When 我搜索并选择一个用户准备邀请
- Then 系统展示可选的角色列表（pm、member 及管理员创建的自定义角色，不包括 superadmin）

- Given 我选择了用户和角色
- When 我确认邀请
- Then 用户加入团队并获得对应角色的权限

- Given 一个团队成员已加入团队
- When 我变更其角色
- Then 该成员的权限即时更新，前端 UI 同步刷新

- Given 我是团队的 PM
- When 我尝试变更自己的角色
- Then 系统拒绝操作（"变更角色"按钮不可用，需超级管理员操作）

- Given 我不是 PM 也不是超级管理员
- When 我尝试邀请成员或变更角色
- Then 操作被拒绝（按钮不可见）

---

## Story 3: 前端根据权限动态渲染 UI

**As a** 系统用户
**I want to** 系统界面只显示我有权限使用的功能和按钮
**So that** 我不会看到无法操作的按钮或被拒绝的操作，体验更清晰

**Acceptance Criteria:**
- Given 我以 member 角色加入某团队
- When 我查看该团队页面
- Then 我看不到"邀请成员"、"编辑团队信息"等需要 pm 权限的按钮

- Given 我以 pm 角色加入某团队
- When 我查看该团队页面
- Then 我能看到团队管理的所有按钮（邀请、移除、转让 PM、编辑事项等）

- Given 我没有 `view:gantt` 权限
- When 我查看导航菜单
- Then 甘特图入口不显示

- Given 我的角色被管理员修改
- When 我切换路由或轮询触发权限刷新
- Then UI 自动更新，显示/隐藏相应的功能

- Given 我是超级管理员
- When 我查看任何团队页面
- Then 所有操作按钮都可见，不受团队角色限制

- Given 我在 A 团队是 pm 角色，在 B 团队是 member 角色
- When 我从 A 团队切换到 B 团队
- Then A 团队的"邀请成员"等 PM 按钮消失，B 团队只显示 member 权限范围内的按钮

- Given 我没有任何团队角色包含 `user:read` 权限
- When 我查看侧边栏
- Then "用户管理"菜单不显示

---

## Story 4: 现有数据无缝迁移到 RBAC

**As a** 超级管理员
**I want to** 系统通过迁移脚本自动将现有用户权限映射到新的 RBAC 模型
**So that** 升级后所有用户的行为与升级前完全一致，无需手动重新配置

**Acceptance Criteria:**
- Given 迁移前存在 users.is_super_admin = true 的用户
- When 迁移脚本执行完成
- Then 这些用户被绑定为 superadmin 预置角色，可继续管理所有团队

- Given 迁移前存在 team_members.role = "pm" 的团队成员
- When 迁移脚本执行完成
- Then 这些成员被绑定为 pm 预置角色，在原团队保持 PM 权限

- Given 迁移前存在 team_members.role = "member" 的团队成员
- When 迁移脚本执行完成
- Then 这些成员被绑定为 member 预置角色

- Given 迁移前存在 users.can_create_team = true 的用户
- When 迁移脚本执行完成
- Then 这些用户的角色包含 team:create 权限码

- Given 迁移脚本执行失败
- When 发生任何错误
- Then 整个迁移回滚，数据库恢复到迁移前状态，系统保持原有行为，可排查后重新执行

- Given 迁移脚本已执行成功
- When 再次执行迁移脚本
- Then 脚本正常完成，不产生重复数据或错误（幂等性）

- Given 迁移完成后
- When 系统启动
- Then users 表上不再存在 is_super_admin 和 can_create_team 字段

---

## Story 5: 团队创建权限控制

**As a** 有创建团队权限的用户
**I want to** 在拥有权限的情况下创建新团队
**So that** 我可以按需组建新的工作团队

**Acceptance Criteria:**
- Given 我的某个团队角色包含 `team:create` 权限码
- When 我点击"创建团队"
- Then 我可以成功创建团队

- Given 我的任何团队角色都不包含 `team:create` 权限码
- When 我查看团队管理页面
- Then "创建团队"按钮不可见

- Given 我是超级管理员
- When 我查看团队管理页面
- Then "创建团队"按钮始终可见

---

## Story 6: PM 的权限驱动操作

**As a** PM（团队负责人）
**I want to** 通过角色的权限码执行团队管理操作
**So that** 我的操作能力由权限码精确控制，而非依赖"PM"角色名

**Acceptance Criteria:**
- Given 我的角色包含 `main_item:create` 权限码
- When 我在团队中创建主事项
- Then 创建成功

- Given 我的角色包含 `sub_item:assign` 权限码
- When 我为子事项分配负责人
- Then 分配成功

- Given 我的角色包含 `item_pool:review` 权限码
- When 我审核、分配或拒绝事项池中的事项
- Then 操作成功

- Given 我的角色不包含 `main_item:archive` 权限码
- When 我尝试归档主事项
- Then 按钮不可见，直接调用 API 返回 403

- Given 我的角色包含 `progress:update` 权限码
- When 我修正团队成员的进度记录
- Then 修正成功

---

## Story 7: Member 的受限操作

**As a** 团队成员（member 角色）
**I want to** 在权限范围内完成日常工作
**So that** 我可以提交事项、更新进度，但不能执行管理操作

**Acceptance Criteria:**
- Given 我的角色包含 `sub_item:create` 权限码
- When 我创建子事项
- Then 创建成功

- Given 我的角色包含 `progress:create` 权限码
- When 我追加进度记录
- Then 记录创建成功

- Given 我的角色包含 `item_pool:submit` 权限码
- When 我提交事项到事项池
- Then 提交成功

- Given 我的角色不包含 `team:invite` 权限码
- When 我查看团队成员页面
- Then "邀请成员"按钮不显示，直接调用 API 返回 403

- Given 我的角色包含 `sub_item:update` 权限码
- And 该子事项分配给了我
- When 我编辑该子事项
- Then 编辑成功（被分配者额外访问权限）

- Given 我的角色包含 `sub_item:update` 权限码
- And 该子事项未分配给我
- When 我尝试编辑该子事项
- Then 编辑被拒绝（需 assignee 或 PM 权限）

---

## Story 8: 跨团队权限隔离

**As a** 同时属于多个团队的用户
**I want to** 在不同团队中拥有不同的操作权限
**So that** 我在 A 团队的管理能力不会延伸到 B 团队

**Acceptance Criteria:**
- Given 我在 A 团队是 pm 角色，在 B 团队是 member 角色
- When 我在 A 团队邀请成员
- Then 邀请成功

- Given 我在 A 团队是 pm 角色，在 B 团队是 member 角色
- When 我在 B 团队尝试邀请成员
- Then "邀请成员"按钮不显示

- Given 我在 A 团队是 pm 角色
- When 我尝试用 A 团队的 pm 权限操作 B 团队的资源
- Then 操作被拒绝（权限不跨团队生效）

---

## Story 9: 后端权限强制执行

**As a** 系统管理员
**I want to** 后端中间件始终强制执行权限检查
**So that** 即使前端遗漏了权限渲染，恶意用户也无法通过直接调用 API 绕过权限

**Acceptance Criteria:**
- Given 一个用户的角色不包含 `team:invite` 权限码
- When 该用户直接调用 `POST /api/v1/teams/:teamId/members` 接口
- Then 后端返回 403 Forbidden

- Given 一个用户不是团队成员
- When 该用户直接调用团队内任何 API
- Then 后端返回 403 Forbidden（非团队成员）

- Given 一个非超级管理员用户
- When 该用户直接调用 `GET /api/v1/admin/users` 接口
- Then 后端返回 403 Forbidden

- Given 一个用户的角色不包含 `sub_item:assign` 权限码
- When 该用户直接调用 `PUT /api/v1/teams/:teamId/sub-items/:subId/assignee` 接口
- Then 后端返回 403 Forbidden

---

## Story 10: 角色编辑即时生效

**As a** 超级管理员
**I want to** 编辑角色的权限后变更即时生效
**So that** 权限调整不需要等待 token 过期或用户重新登录

**Acceptance Criteria:**
- Given 用户 A 的角色包含 `team:invite` 权限码
- When 我编辑该角色，取消 `team:invite` 权限码并保存
- Then 用户 A 的下一次请求即被拒绝 `team:invite` 相关操作（无需重新登录）

- Given 用户 A 的角色不包含 `view:gantt` 权限码
- When 我编辑该角色，增加 `view:gantt` 权限码并保存
- Then 用户 A 的前端在权限刷新后显示甘特图入口（无需重新登录）

- Given 用户 A 正在使用需要 `team:invite` 权限的功能
- When 我取消该权限并保存
- Then 用户 A 的下一次操作或路由切换时，前端隐藏"邀请成员"按钮；在此之前的当前会话操作仍可继续
