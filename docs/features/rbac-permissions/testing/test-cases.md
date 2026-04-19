---
feature: "rbac-permissions"
sources:
  - prd/prd-user-stories.md
  - prd/prd-spec.md
  - prd/prd-ui-functions.md
generated: "2026-04-19"
---

# Test Cases: rbac-permissions

## Summary

| Type | Count |
|------|-------|
| UI   | 33   |
| API  | 22  |
| CLI  | 7  |
| **Total** | **62** |

---

## UI Test Cases

### TC-001: 角色列表展示完整信息
- **Source**: Story 1 / AC-1
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录
- **Steps**:
  1. 点击导航中"角色管理"菜单
  2. 等待角色列表页加载完成
- **Expected**: 列表展示所有系统角色，每行包含角色名称、描述、权限数量、使用人数、类型（预置/自定义）和创建时间
- **Priority**: P0

### TC-002: 创建新角色成功并出现在列表
- **Source**: Story 1 / AC-2
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录，在角色管理页面
- **Steps**:
  1. 点击"创建角色"按钮
  2. 填写角色名称和描述
  3. 勾选至少一个权限码
  4. 点击"保存"
- **Expected**: 新角色创建成功，出现在角色列表中；邀请成员时角色下拉列表包含该新角色
- **Priority**: P0

### TC-003: 编辑角色权限即时生效
- **Source**: Story 1 / AC-3
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录，存在一个非 superadmin 角色
- **Steps**:
  1. 选择该角色并点击编辑
  2. 修改权限勾选（增加或移除权限码）
  3. 保存
- **Expected**: 权限即时更新；所有使用该角色的用户权限同步生效
- **Priority**: P0

### TC-004: 编辑角色名称和描述
- **Source**: Story 1 / AC-4
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录，存在一个非 superadmin 角色
- **Steps**:
  1. 选择该自定义角色并点击编辑
  2. 修改角色名称和描述
  3. 保存
- **Expected**: 角色名称和描述更新成功，列表显示新信息
- **Priority**: P0

### TC-005: 删除无用户的自定义角色
- **Source**: Story 1 / AC-5
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录，存在一个非预置角色且无用户使用
- **Steps**:
  1. 找到该自定义角色
  2. 点击"删除"按钮
  3. 在确认弹窗中确认删除
- **Expected**: 角色被删除，不再出现在列表中
- **Priority**: P0

### TC-006: 有用户的角色无法删除
- **Source**: Story 1 / AC-6
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录，存在一个有用户正在使用的角色
- **Steps**:
  1. 找到该角色
  2. 尝试点击"删除"按钮
- **Expected**: 删除按钮置灰并显示 Tooltip 提示"该角色正在被 N 个用户使用，无法删除"
- **Priority**: P0

### TC-007: superadmin 预置角色不可编辑删除
- **Source**: Story 1 / AC-7
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录
- **Steps**:
  1. 在角色列表中找到 superadmin 预置角色
  2. 查看操作列
- **Expected**: 编辑和删除按钮不可用（不显示或置灰）
- **Priority**: P0

### TC-008: pm/member 预置角色可编辑权限不可删除
- **Source**: Story 1 / AC-8
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录
- **Steps**:
  1. 找到 pm 或 member 预置角色
  2. 点击编辑按钮
  3. 尝试修改权限勾选
  4. 尝试修改角色名称
  5. 查看删除按钮
- **Expected**: 权限勾选可修改并保存成功；角色名称字段为 disabled 状态不可修改；删除按钮不可用
- **Priority**: P0

### TC-009: 查看系统权限码列表
- **Source**: Story 1 / AC-9
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录，在角色管理页面
- **Steps**:
  1. 点击"查看权限列表"按钮
  2. 查看弹出的权限列表视图
- **Expected**: 展示所有系统权限码，按资源分组（团队管理、主事项、子事项等），每个权限码显示权限码字符串和操作描述；全部只读不可编辑
- **Priority**: P1

### TC-010: 邀请成员时展示角色列表（排除 superadmin）
- **Source**: Story 2 / AC-1
- **Type**: UI
- **Pre-conditions**: 以 PM 角色身份登录，在团队成员管理页面
- **Steps**:
  1. 点击"邀请成员"按钮
  2. 搜索并选择一个用户
  3. 查看角色下拉列表
- **Expected**: 下拉列表包含 pm、member 及管理员创建的自定义角色，不包含 superadmin
- **Priority**: P0

### TC-011: 变更成员角色后 UI 刷新
- **Source**: Story 2 / AC-3
- **Type**: UI
- **Pre-conditions**: PM 在团队成员管理页面，团队中有其他成员
- **Steps**:
  1. 点击某成员角色旁的"变更"按钮
  2. 从下拉列表中选择新角色
  3. 等待变更完成
- **Expected**: 该成员的角色显示更新为新角色名称；该成员的前端 UI 在下次权限刷新时同步更新
- **Priority**: P0

### TC-012: PM 不能变更自己的角色
- **Source**: Story 2 / AC-4
- **Type**: UI
- **Pre-conditions**: 以 PM 角色身份登录，查看团队成员列表
- **Steps**:
  1. 在成员列表中找到自己的行
- **Expected**: 自己的角色行不显示"变更"按钮（需超级管理员操作）
- **Priority**: P0

### TC-013: 非 PM 非超管看不到邀请和变更按钮
- **Source**: Story 2 / AC-5
- **Type**: UI
- **Pre-conditions**: 以 member 角色身份登录
- **Steps**:
  1. 进入团队成员管理页面
- **Expected**: "邀请成员"按钮和"变更角色"按钮不可见
- **Priority**: P1

### TC-014: Member 看不到 PM 权限按钮
- **Source**: Story 3 / AC-1
- **Type**: UI
- **Pre-conditions**: 以 member 角色加入某团队
- **Steps**:
  1. 查看该团队详情页面
- **Expected**: 看不到"邀请成员"、"编辑团队信息"、"转让 PM"等需要 PM 权限的按钮
- **Priority**: P0

### TC-015: PM 能看到所有管理按钮
- **Source**: Story 3 / AC-2
- **Type**: UI
- **Pre-conditions**: 以 PM 角色加入某团队
- **Steps**:
  1. 查看该团队详情页面
- **Expected**: 能看到团队管理的所有按钮：邀请、移除、转让 PM、编辑事项等
- **Priority**: P0

### TC-016: 无 view:gantt 权限甘特图入口不显示
- **Source**: Story 3 / AC-3
- **Type**: UI
- **Pre-conditions**: 用户角色不包含 `view:gantt` 权限码
- **Steps**:
  1. 查看导航菜单
- **Expected**: 甘特图入口不显示
- **Priority**: P0

### TC-017: 角色修改后路由切换时 UI 刷新
- **Source**: Story 3 / AC-4
- **Type**: UI
- **Pre-conditions**: 管理员修改了用户的角色权限
- **Steps**:
  1. 用户切换到其他路由页面
  2. 观察按钮/菜单的显示状态
- **Expected**: UI 自动更新，根据新权限显示/隐藏相应的功能
- **Priority**: P1

### TC-018: 超管所有按钮可见不受团队限制
- **Source**: Story 3 / AC-5
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录
- **Steps**:
  1. 查看任意团队页面
- **Expected**: 所有操作按钮都可见，不受团队角色限制
- **Priority**: P0

### TC-019: 跨团队切换权限按钮变化
- **Source**: Story 3 / AC-6
- **Type**: UI
- **Pre-conditions**: 用户在 A 团队是 pm 角色，在 B 团队是 member 角色
- **Steps**:
  1. 在 A 团队查看按钮状态
  2. 切换到 B 团队
  3. 再次查看按钮状态
- **Expected**: A 团队显示"邀请成员"等 PM 按钮；切换到 B 团队后这些按钮消失，只显示 member 权限范围内的按钮
- **Priority**: P0

### TC-020: 无 user:read 权限用户管理菜单不显示
- **Source**: Story 3 / AC-7
- **Type**: UI
- **Pre-conditions**: 用户没有任何团队角色包含 `user:read` 权限
- **Steps**:
  1. 查看侧边栏导航
- **Expected**: "用户管理"菜单不显示
- **Priority**: P1

### TC-021: 无 team:create 权限创建团队按钮不可见
- **Source**: Story 5 / AC-2
- **Type**: UI
- **Pre-conditions**: 用户所有团队角色都不包含 `team:create` 权限码
- **Steps**:
  1. 查看团队管理页面
- **Expected**: "创建团队"按钮不可见
- **Priority**: P0

### TC-022: 超管创建团队按钮始终可见
- **Source**: Story 5 / AC-3
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录
- **Steps**:
  1. 查看团队管理页面
- **Expected**: "创建团队"按钮始终可见
- **Priority**: P1

### TC-023: B 团队邀请按钮不显示（member 角色）
- **Source**: Story 8 / AC-2
- **Type**: UI
- **Pre-conditions**: 用户在 A 团队是 pm 角色，在 B 团队是 member 角色
- **Steps**:
  1. 切换到 B 团队上下文
  2. 查看 B 团队详情页面
- **Expected**: "邀请成员"按钮不显示
- **Priority**: P0

### TC-024: 角色增加权限后前端显示新入口
- **Source**: Story 10 / AC-2
- **Type**: UI
- **Pre-conditions**: 用户 A 的角色不包含 `view:gantt` 权限码
- **Steps**:
  1. 超管编辑该角色，增加 `view:gantt` 权限码并保存
  2. 用户 A 刷新权限（路由切换或轮询触发）
- **Expected**: 用户 A 的前端在权限刷新后显示甘特图入口（无需重新登录）
- **Priority**: P1

### TC-025: 权限取消后前端隐藏按钮
- **Source**: Story 10 / AC-3
- **Type**: UI
- **Pre-conditions**: 用户 A 正在使用需要 `team:invite` 权限的功能
- **Steps**:
  1. 超管取消用户 A 角色的 `team:invite` 权限并保存
  2. 用户 A 进行下一次操作或路由切换
- **Expected**: 前端隐藏"邀请成员"按钮；在此之前当前会话的操作仍可继续
- **Priority**: P1

### TC-026: 角色列表页加载骨架屏状态
- **Source**: UI Function 1 / States
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录
- **Steps**:
  1. 进入角色管理页面
  2. 观察页面加载过程
- **Expected**: 加载期间表格行显示骨架屏（3-5 行），筛选栏不可操作
- **Priority**: P1

### TC-027: 角色列表页空状态提示
- **Source**: UI Function 1 / States
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录，系统中只有预置角色
- **Steps**:
  1. 进入角色管理页面
  2. 筛选"自定义"角色
- **Expected**: 显示居中提示"暂无自定义角色"，并提供创建按钮
- **Priority**: P2

### TC-028: 角色列表页错误状态与重试
- **Source**: UI Function 1 / States
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录
- **Steps**:
  1. 模拟网络错误导致加载失败
  2. 查看错误提示
  3. 点击"重试"按钮
- **Expected**: 表格上方显示红色错误提示条；点击重试后重新加载数据
- **Priority**: P2

### TC-029: 角色名称校验（2-50 字符，不可重名）
- **Source**: Spec 5.1 / 表单字段规则
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录，在角色编辑表单
- **Steps**:
  1. 输入 1 个字符的角色名称 → 观察校验提示
  2. 输入 51 个字符的角色名称 → 观察校验提示
  3. 输入与已有角色相同的名称 → 观察校验提示
  4. 输入 2-50 字符且不重名的名称 → 观察校验通过
- **Expected**: 不符合长度时显示校验提示；重名时显示"角色名称已存在"；合法输入时无错误提示
- **Priority**: P1

### TC-030: 描述字符数限制（最多 200 字符）
- **Source**: Spec 5.1 / 表单字段规则
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录，在角色编辑表单
- **Steps**:
  1. 输入超过 200 字符的描述文本
- **Expected**: 显示字符数超限校验提示
- **Priority**: P2

### TC-031: 权限勾选至少选择一个
- **Source**: Spec 5.1 / 表单字段规则
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录，在角色编辑表单
- **Steps**:
  1. 取消所有权限勾选
  2. 点击保存
- **Expected**: 显示校验提示"至少选择 1 个权限"
- **Priority**: P1

### TC-032: 搜索角色名称筛选列表
- **Source**: Spec 5.1 / 搜索条件
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录，角色列表中存在多个角色
- **Steps**:
  1. 在搜索框输入角色名称关键字
  2. 等待 300ms 防抖
- **Expected**: 列表实时筛选，仅显示名称包含关键字的角色
- **Priority**: P2

### TC-033: 筛选预置/自定义角色
- **Source**: Spec 5.1 / 搜索条件
- **Type**: UI
- **Pre-conditions**: 以超级管理员身份登录，角色列表中存在预置和自定义角色
- **Steps**:
  1. 选择筛选下拉中的"预置"
  2. 观察列表变化
  3. 选择"自定义"
  4. 观察列表变化
- **Expected**: 选择"预置"仅显示 superadmin、pm、member；选择"自定义"仅显示用户创建的角色
- **Priority**: P2

---

## API Test Cases

### TC-034: 邀请用户加入团队并分配角色
- **Source**: Story 2 / AC-2
- **Type**: API
- **Pre-conditions**: PM 身份，目标用户不在团队中
- **Steps**:
  1. 调用 `POST /api/v1/teams/:teamId/members`，传入 userId 和 roleId
- **Expected**: 用户成功加入团队；用户获得对应角色的权限；返回 200
- **Priority**: P0

### TC-035: 拥有 team:create 权限创建团队成功
- **Source**: Story 5 / AC-1
- **Type**: API
- **Pre-conditions**: 用户角色包含 `team:create` 权限码
- **Steps**:
  1. 调用创建团队 API
- **Expected**: 团队创建成功，返回 200
- **Priority**: P0

### TC-036: 拥有 main_item:create 权限创建主事项
- **Source**: Story 6 / AC-1
- **Type**: API
- **Pre-conditions**: 用户在团队中的角色包含 `main_item:create` 权限码
- **Steps**:
  1. 在该团队中调用创建主事项 API
- **Expected**: 主事项创建成功
- **Priority**: P0

### TC-037: 拥有 sub_item:assign 权限分配负责人
- **Source**: Story 6 / AC-2
- **Type**: API
- **Pre-conditions**: 用户角色包含 `sub_item:assign` 权限码
- **Steps**:
  1. 调用 `PUT /api/v1/teams/:teamId/sub-items/:subId/assignee` 分配负责人
- **Expected**: 分配成功
- **Priority**: P0

### TC-038: 拥有 item_pool:review 权限审核事项
- **Source**: Story 6 / AC-3
- **Type**: API
- **Pre-conditions**: 用户角色包含 `item_pool:review` 权限码
- **Steps**:
  1. 调用事项池审核/分配/拒绝 API
- **Expected**: 操作成功
- **Priority**: P0

### TC-039: 无 main_item:archive 权限归档返回 403
- **Source**: Story 6 / AC-4
- **Type**: API
- **Pre-conditions**: 用户角色不包含 `main_item:archive` 权限码
- **Steps**:
  1. 直接调用归档主事项 API（绕过前端按钮隐藏）
- **Expected**: 返回 403 Forbidden
- **Priority**: P0

### TC-040: 拥有 progress:update 权限修正进度
- **Source**: Story 6 / AC-5
- **Type**: API
- **Pre-conditions**: 用户角色包含 `progress:update` 权限码
- **Steps**:
  1. 调用修正团队成员进度记录 API
- **Expected**: 修正成功
- **Priority**: P0

### TC-041: 拥有 sub_item:create 权限创建子事项
- **Source**: Story 7 / AC-1
- **Type**: API
- **Pre-conditions**: 用户角色包含 `sub_item:create` 权限码
- **Steps**:
  1. 调用创建子事项 API
- **Expected**: 子事项创建成功
- **Priority**: P1

### TC-042: 拥有 progress:create 权限追加进度
- **Source**: Story 7 / AC-2
- **Type**: API
- **Pre-conditions**: 用户角色包含 `progress:create` 权限码
- **Steps**:
  1. 调用追加进度记录 API
- **Expected**: 进度记录创建成功
- **Priority**: P1

### TC-043: 拥有 item_pool:submit 权限提交事项
- **Source**: Story 7 / AC-3
- **Type**: API
- **Pre-conditions**: 用户角色包含 `item_pool:submit` 权限码
- **Steps**:
  1. 调用提交事项到事项池 API
- **Expected**: 提交成功
- **Priority**: P1

### TC-044: 无 team:invite 权限 API 返回 403
- **Source**: Story 7 / AC-4, Story 9 / AC-1
- **Type**: API
- **Pre-conditions**: 用户角色不包含 `team:invite` 权限码
- **Steps**:
  1. 直接调用 `POST /api/v1/teams/:teamId/members` 接口
- **Expected**: 后端返回 403 Forbidden
- **Priority**: P0

### TC-045: 被分配者可编辑自己的子事项
- **Source**: Story 7 / AC-5
- **Type**: API
- **Pre-conditions**: 用户角色包含 `sub_item:update` 权限码；该子事项已分配给该用户
- **Steps**:
  1. 调用编辑子事项 API
- **Expected**: 编辑成功（被分配者额外访问权限生效）
- **Priority**: P0

### TC-046: 非被分配者编辑子事项被拒绝
- **Source**: Story 7 / AC-6
- **Type**: API
- **Pre-conditions**: 用户角色包含 `sub_item:update` 权限码；该子事项未分配给该用户
- **Steps**:
  1. 调用编辑子事项 API
- **Expected**: 编辑被拒绝，返回 403（需 assignee 或 PM 权限）
- **Priority**: P0

### TC-047: A 团队 pm 角色邀请成员成功
- **Source**: Story 8 / AC-1
- **Type**: API
- **Pre-conditions**: 用户在 A 团队是 pm 角色，在 B 团队是 member 角色
- **Steps**:
  1. 在 A 团队上下文调用邀请成员 API
- **Expected**: 邀请成功
- **Priority**: P0

### TC-048: A 团队权限不跨团队到 B 团队
- **Source**: Story 8 / AC-3
- **Type**: API
- **Pre-conditions**: 用户在 A 团队是 pm 角色
- **Steps**:
  1. 使用 A 团队的 pm 权限尝试操作 B 团队的资源（如邀请 B 团队成员）
- **Expected**: 操作被拒绝，返回 403（权限不跨团队生效）
- **Priority**: P0

### TC-049: 非团队成员调用团队 API 返回 403
- **Source**: Story 9 / AC-2
- **Type**: API
- **Pre-conditions**: 用户不是该团队成员
- **Steps**:
  1. 直接调用该团队内任何 API
- **Expected**: 后端返回 403 Forbidden
- **Priority**: P0

### TC-050: 非超管调用 admin API 返回 403
- **Source**: Story 9 / AC-3
- **Type**: API
- **Pre-conditions**: 用户不是超级管理员
- **Steps**:
  1. 直接调用 `GET /api/v1/admin/users` 接口
- **Expected**: 后端返回 403 Forbidden
- **Priority**: P0

### TC-051: 无 sub_item:assign 权限直接调用 API 返回 403
- **Source**: Story 9 / AC-4
- **Type**: API
- **Pre-conditions**: 用户角色不包含 `sub_item:assign` 权限码
- **Steps**:
  1. 直接调用 `PUT /api/v1/teams/:teamId/sub-items/:subId/assignee` 接口
- **Expected**: 后端返回 403 Forbidden
- **Priority**: P0

### TC-052: 角色权限取消后后端即时生效
- **Source**: Story 10 / AC-1
- **Type**: API
- **Pre-conditions**: 用户 A 的角色包含 `team:invite` 权限码
- **Steps**:
  1. 超管编辑该角色，取消 `team:invite` 权限码并保存
  2. 用户 A 调用 `POST /api/v1/teams/:teamId/members` 接口
- **Expected**: 后端返回 403，无需用户重新登录
- **Priority**: P0

### TC-053: 获取用户权限 API 返回正确数据
- **Source**: Spec 5.6 / 权限获取方式
- **Type**: API
- **Pre-conditions**: 用户已登录，拥有 JWT token
- **Steps**:
  1. 调用 `GET /api/me/permissions`
- **Expected**: 返回 `{ is_superadmin: boolean, team_permissions: { team_id: [permission_codes] } }` 格式数据
- **Priority**: P0

### TC-054: 权限 API 响应时间 < 200ms
- **Source**: Spec / 性能需求
- **Type**: API
- **Pre-conditions**: 用户已登录
- **Steps**:
  1. 调用 `GET /api/me/permissions`
  2. 测量响应时间
- **Expected**: 响应时间 < 200ms
- **Priority**: P2

### TC-055: 权限检查中间件响应时间 < 10ms
- **Source**: Spec / 性能需求
- **Type**: API
- **Pre-conditions**: 用户已登录，请求需要权限校验的 API
- **Steps**:
  1. 调用需要权限检查的 API
  2. 测量中间件权限校验耗时
- **Expected**: 权限检查响应时间 < 10ms
- **Priority**: P2

---

## CLI Test Cases

### TC-056: 迁移超级管理员用户到 superadmin 角色
- **Source**: Story 4 / AC-1
- **Type**: CLI
- **Pre-conditions**: 数据库中存在 `users.is_super_admin = true` 的用户
- **Steps**:
  1. 执行迁移脚本
  2. 检查该用户是否绑定了 superadmin 预置角色
- **Expected**: 迁移后该用户绑定为 superadmin 预置角色，可继续管理所有团队
- **Priority**: P0

### TC-057: 迁移 PM 团队成员到 pm 角色
- **Source**: Story 4 / AC-2
- **Type**: CLI
- **Pre-conditions**: 数据库中存在 `team_members.role = "pm"` 的团队成员
- **Steps**:
  1. 执行迁移脚本
  2. 检查这些成员的角色绑定
- **Expected**: 这些成员被绑定为 pm 预置角色，在原团队保持 PM 权限
- **Priority**: P0

### TC-058: 迁移普通成员到 member 角色
- **Source**: Story 4 / AC-3
- **Type**: CLI
- **Pre-conditions**: 数据库中存在 `team_members.role = "member"` 的团队成员
- **Steps**:
  1. 执行迁移脚本
  2. 检查这些成员的角色绑定
- **Expected**: 这些成员被绑定为 member 预置角色
- **Priority**: P0

### TC-059: 迁移 can_create_team 用户保留权限
- **Source**: Story 4 / AC-4
- **Type**: CLI
- **Pre-conditions**: 数据库中存在 `users.can_create_team = true` 的用户
- **Steps**:
  1. 执行迁移脚本
  2. 检查这些用户角色的权限码
- **Expected**: 这些用户的角色包含 `team:create` 权限码
- **Priority**: P0

### TC-060: 迁移失败完整回滚
- **Source**: Story 4 / AC-5
- **Type**: CLI
- **Pre-conditions**: 数据库处于迁移前状态
- **Steps**:
  1. 模拟迁移过程中某步骤失败
  2. 检查数据库状态
- **Expected**: 整个迁移回滚，数据库恢复到迁移前状态，系统保持原有行为
- **Priority**: P0

### TC-061: 迁移脚本幂等执行
- **Source**: Story 4 / AC-6
- **Type**: CLI
- **Pre-conditions**: 迁移脚本已执行成功一次
- **Steps**:
  1. 再次执行迁移脚本
  2. 检查数据状态
- **Expected**: 脚本正常完成，不产生重复数据或错误
- **Priority**: P0

### TC-062: 迁移后移除旧字段
- **Source**: Story 4 / AC-7
- **Type**: CLI
- **Pre-conditions**: 迁移已完成
- **Steps**:
  1. 启动系统
  2. 检查 users 表结构
- **Expected**: users 表上不再存在 `is_super_admin` 和 `can_create_team` 字段
- **Priority**: P0

---

## Traceability

| TC ID | Source | Type | Priority |
|-------|--------|------|----------|
| TC-001 | Story 1 / AC-1 | UI | P0 |
| TC-002 | Story 1 / AC-2 | UI | P0 |
| TC-003 | Story 1 / AC-3 | UI | P0 |
| TC-004 | Story 1 / AC-4 | UI | P0 |
| TC-005 | Story 1 / AC-5 | UI | P0 |
| TC-006 | Story 1 / AC-6 | UI | P0 |
| TC-007 | Story 1 / AC-7 | UI | P0 |
| TC-008 | Story 1 / AC-8 | UI | P0 |
| TC-009 | Story 1 / AC-9 | UI | P1 |
| TC-010 | Story 2 / AC-1 | UI | P0 |
| TC-011 | Story 2 / AC-3 | UI | P0 |
| TC-012 | Story 2 / AC-4 | UI | P0 |
| TC-013 | Story 2 / AC-5 | UI | P1 |
| TC-014 | Story 3 / AC-1 | UI | P0 |
| TC-015 | Story 3 / AC-2 | UI | P0 |
| TC-016 | Story 3 / AC-3 | UI | P0 |
| TC-017 | Story 3 / AC-4 | UI | P1 |
| TC-018 | Story 3 / AC-5 | UI | P0 |
| TC-019 | Story 3 / AC-6 | UI | P0 |
| TC-020 | Story 3 / AC-7 | UI | P1 |
| TC-021 | Story 5 / AC-2 | UI | P0 |
| TC-022 | Story 5 / AC-3 | UI | P1 |
| TC-023 | Story 8 / AC-2 | UI | P0 |
| TC-024 | Story 10 / AC-2 | UI | P1 |
| TC-025 | Story 10 / AC-3 | UI | P1 |
| TC-026 | UI Function 1 / States | UI | P1 |
| TC-027 | UI Function 1 / States | UI | P2 |
| TC-028 | UI Function 1 / States | UI | P2 |
| TC-029 | Spec 5.1 / 表单字段规则 | UI | P1 |
| TC-030 | Spec 5.1 / 表单字段规则 | UI | P2 |
| TC-031 | Spec 5.1 / 表单字段规则 | UI | P1 |
| TC-032 | Spec 5.1 / 搜索条件 | UI | P2 |
| TC-033 | Spec 5.1 / 搜索条件 | UI | P2 |
| TC-034 | Story 2 / AC-2 | API | P0 |
| TC-035 | Story 5 / AC-1 | API | P0 |
| TC-036 | Story 6 / AC-1 | API | P0 |
| TC-037 | Story 6 / AC-2 | API | P0 |
| TC-038 | Story 6 / AC-3 | API | P0 |
| TC-039 | Story 6 / AC-4 | API | P0 |
| TC-040 | Story 6 / AC-5 | API | P0 |
| TC-041 | Story 7 / AC-1 | API | P1 |
| TC-042 | Story 7 / AC-2 | API | P1 |
| TC-043 | Story 7 / AC-3 | API | P1 |
| TC-044 | Story 7 / AC-4, Story 9 / AC-1 | API | P0 |
| TC-045 | Story 7 / AC-5 | API | P0 |
| TC-046 | Story 7 / AC-6 | API | P0 |
| TC-047 | Story 8 / AC-1 | API | P0 |
| TC-048 | Story 8 / AC-3 | API | P0 |
| TC-049 | Story 9 / AC-2 | API | P0 |
| TC-050 | Story 9 / AC-3 | API | P0 |
| TC-051 | Story 9 / AC-4 | API | P0 |
| TC-052 | Story 10 / AC-1 | API | P0 |
| TC-053 | Spec 5.6 / 权限获取方式 | API | P0 |
| TC-054 | Spec / 性能需求 | API | P2 |
| TC-055 | Spec / 性能需求 | API | P2 |
| TC-056 | Story 4 / AC-1 | CLI | P0 |
| TC-057 | Story 4 / AC-2 | CLI | P0 |
| TC-058 | Story 4 / AC-3 | CLI | P0 |
| TC-059 | Story 4 / AC-4 | CLI | P0 |
| TC-060 | Story 4 / AC-5 | CLI | P0 |
| TC-061 | Story 4 / AC-6 | CLI | P0 |
| TC-062 | Story 4 / AC-7 | CLI | P0 |
