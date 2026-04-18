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
- Then 新角色创建成功，出现在角色列表中，立即可用于分配给用户

- Given 我选择了一个已有角色
- When 我修改其权限勾选并保存
- Then 该角色的权限即时更新，所有使用该角色的用户权限同步生效

- Given 我选择了一个自定义角色（非预置角色）
- When 该角色没有用户使用
- Then 我可以删除该角色

- Given 我选择了一个角色
- When 该角色有用户正在使用
- Then 系统提示"该角色正在被 N 个用户使用，无法删除"

---

## Story 2: PM 在邀请成员时指定角色

**As a** PM（团队负责人）
**I want to** 在邀请用户加入团队时为其指定角色
**So that** 不同成员根据其职责拥有不同的操作权限，团队管理更灵活

**Acceptance Criteria:**
- Given 我在团队成员管理页面
- When 我搜索并选择一个用户准备邀请
- Then 系统展示可选的角色列表（pm、member 及管理员创建的自定义角色）

- Given 我选择了用户和角色
- When 我确认邀请
- Then 用户加入团队并获得对应角色的权限

- Given 一个团队成员已加入团队
- When 我变更其角色
- Then 该成员的权限即时更新

---

## Story 3: 前端根据权限动态渲染 UI

**As a** 系统用户
**I want to** 系统界面只显示我有权限使用的功能和按钮
**So that** 我不会看到无法操作的灰色按钮或被拒绝的操作，体验更清晰

**Acceptance Criteria:**
- Given 我以 member 角色登录
- When 我查看团队页面
- Then 我看不到"邀请成员"、"编辑团队信息"等需要 pm 权限的按钮

- Given 我以 pm 角色登录
- When 我查看团队页面
- Then 我能看到团队管理的所有按钮（邀请、移除、转让 PM）

- Given 我没有 `view:gantt` 权限
- When 我查看导航菜单
- Then 甘特图入口不显示

- Given 我的角色被管理员修改
- When 我下一次操作触发权限刷新
- Then UI 自动更新，显示/隐藏相应的功能

---

## Story 4: 现有数据无缝迁移到 RBAC

**As a** 超级管理员
**I want to** 系统通过迁移脚本自动将现有用户权限映射到新的 RBAC 模型
**So that** 升级后所有用户的行为与升级前完全一致，无需手动重新配置

**Acceptance Criteria:**
- Given 迁移前存在 is_super_admin=true 的用户
- When 迁移脚本执行完成
- Then 这些用户被映射为 superadmin 全局角色，可继续管理所有团队

- Given 迁移前存在 is_pm=true 的团队成员
- When 迁移脚本执行完成
- Then 这些成员被映射为 pm 角色，在原团队保持 PM 权限

- Given 迁移前存在 is_pm=false 的团队成员
- When 迁移脚本执行完成
- Then 这些成员被映射为 member 角色

- Given 迁移脚本执行失败
- When 发生任何错误
- Then 整个迁移回滚，数据库恢复到迁移前状态
