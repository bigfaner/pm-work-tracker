---
feature: "Integration Test Coverage"
---

# User Stories: Integration Test Coverage

## Story 1: 开发者验证 Item 生命周期完整流程

**As a** 开发者
**I want to** 通过集成测试验证 MainItem 创建 → SubItem 创建 → Progress 追加 → Status 变更 → Archive 的完整生命周期
**So that** 确保 item 管理域的 17 个端点在 handler→middleware→service→repository 全链路下行为正确，防止状态机转换和完成度级联计算出现回归

**Acceptance Criteria:**
- Given 一个团队和 PM 用户, When PM 创建 MainItem 并附带有效字段, Then 返回 201 且数据库中可查询到该记录
- Given 一个已存在的 MainItem, When PM 创建 SubItem 并指定权重, Then SubItem 关联到 MainItem，MainItem 完成度保持正确初始值
- Given 一个完成度为 0% 的 SubItem, When PM 追加 Progress (completion=60), Then SubItem 完成度更新为 60，MainItem 完成度按权重重算
- Given 一个 in-progress 的 MainItem, When PM 请求有效状态转换 (如 in-progress→completed), Then 状态变更成功且所有子项自动标记为终端状态
- Given 一个 in-progress 的 MainItem, When PM 请求无效状态转换 (如 new→completed), Then 返回 422 且状态不变
- Given 一个 completed 的 MainItem, When PM 归档该 item, Then 返回 200；当 item 处于 in-progress 时归档返回 422
- Given 一个 member 角色用户, When 尝试创建/更新/变更状态 MainItem, Then 返回 403

---

## Story 2: 开发者验证 Item Pool 审查流程

**As a** 开发者
**I want to** 通过集成测试验证 Item Pool 的提交→分配/转为子项/拒绝完整流程
**So that** 确保池项状态互斥（不可重复处理）、分配事务回滚（无效主项时恢复池状态）、转为独立主项的正确性

**Acceptance Criteria:**
- Given 一个团队, When PM 提交池项, Then 返回 201 且池状态为 pending
- Given 一个 pending 的池项, When PM 分配到有效 MainItem, Then 创建 SubItem + 池状态更新为 assigned
- Given 一个 pending 的池项, When PM 分配到不存在的 MainItem, Then 操作回滚：池状态仍为 pending，无 SubItem 创建
- Given 一个已处理（assigned）的池项, When PM 再次尝试分配/拒绝/转为子项, Then 返回 409
- Given 一个 pending 的池项, When PM 拒绝且提供原因, Then 池状态更新为 rejected
- Given 一个 pending 的池项, When PM 拒绝但未提供原因, Then 返回 422

---

## Story 3: 开发者验证团队管理流程

**As a** 开发者
**I want to** 通过集成测试验证团队创建→邀请成员→角色变更→移除成员→解散团队的完整流程
**So that** 确保 PM 专属操作保护、成员角色正确分配、团队解散级联删除等业务逻辑在全链路下正确执行

**Acceptance Criteria:**
- Given 一个已登录用户, When 创建团队, Then 返回 201 且创建者自动成为 PM
- Given 一个团队, When PM 邀请用户并指定角色, Then 用户成为团队成员并拥有对应角色权限
- Given 一个团队, When PM 变更成员角色, Then 成员权限立即生效（下次请求反映新角色）
- Given 一个团队, When PM 解散团队, Then 团队及其所有项目被删除
- Given 一个团队, When 尝试移除 PM, Then 返回 422
- Given 一个团队, When member 角色用户尝试邀请/移除/变更角色, Then 返回 403

---

## Story 4: 开发者验证后台用户管理流程

**As a** 开发者
**I want to** 通过集成测试验证管理员创建用户→编辑信息→切换状态的完整流程
**So that** 确保用户管理后台的权限控制（仅 SuperAdmin）、重复用户名检测、禁止禁用自身等边界条件正确执行

**Acceptance Criteria:**
- Given 一个 SuperAdmin, When 创建用户（有效用户名+密码）, Then 返回 201
- Given 一个 SuperAdmin, When 创建用户但用户名已存在, Then 返回 409
- Given 一个 SuperAdmin, When 禁用某用户, Then 用户状态更新为 disabled
- Given 一个 SuperAdmin, When 尝试禁用自身, Then 返回 422
- Given 一个普通用户, When 尝试访问 /admin/users, Then 返回 403

---

## Story 5: 开发者验证视图和报表数据正确性

**As a** 开发者
**I want to** 通过集成测试验证 Weekly/Gantt/Table 视图和 CSV/Report 导出的数据正确性
**So that** 确保聚合统计、过滤逻辑、导出格式等数据密集型端点在真实数据下返回预期结果

**Acceptance Criteria:**
- Given 3 项创建（其中 1 项 completed、2 项 in-progress）, When 请求周视图, Then stats 为 `{NEW:0, completed:1, inProgress:2, overdue:0}`
- Given 混合状态的项目, When 请求 `?status=completed`, Then 仅返回已完成项
- Given 项目数据, When 请求 CSV 导出, Then 响应以 UTF-8 BOM 开头，首行为表头
- Given 空团队, When 请求任何视图端点, Then 返回空数据（零统计/空数组/BOM+表头）
- Given 有活动的项目, When 请求周报预览, Then Markdown 含 `## Summary` 和 `## Items` 段落

---

## Story 6: 开发者提取共享测试辅助函数

**As a** 开发者
**I want to** 从现有集成测试（`auth_isolation_test.go`、`progress_completion_test.go`）和 F1 编写过程中提取 10 个共享辅助函数到独立 `helpers.go` 文件
**So that** F2-F5 的测试编写可直接复用 DB 初始化、路由构建、登录、种子数据等基础设施，避免每个流程文件重复实现相同逻辑

**Acceptance Criteria:**
- Given F1 测试编写完成, When 从中提取 `setupTestDB`/`setupTestRouter`/`loginAs`/`signTokenWithClaims`/`seedProgressData`/`appendProgress`/`seedPoolData`/`seedReportData` 8 个现有辅助函数, Then 函数签名与 F7 规格表一致，原始文件中的重复定义已删除
- Given F1 测试中发现可复用模式, When 提取 `createTeamWithMembers` 和 `createMainItem` 2 个新辅助函数, Then 签名符合 F7 规格表且带 GoDoc 注释
- Given helpers.go 文件创建完成, When 运行 F1 测试, Then 所有测试仍通过（提取不改变行为）

---

## Story 7: 代码审查者通过增量 PR 审查测试代码

**As a** 代码审查者
**I want to** 每个 Feature（F1-F6）作为独立测试 PR 提交，F7 helpers 作为独立基础设施 PR 提交
**So that** 我能在每次 PR 审查中聚焦一个业务域的测试逻辑，无需一次性理解全部 150+ 用例

**Acceptance Criteria:**
- Given 收到一个测试 PR, When 审查 diff, Then 测试函数名包含业务语义可识别被测场景
- Given 收到一个测试 PR, When 审查 diff, Then 同一流程的端点测试在相邻函数中可追踪用户操作序列
- Given 收到 F7 helpers PR, When 审查文件, Then 包含 F7 规格表所列全部 10 个辅助函数
- Given 收到 F7 helpers PR, When 审查文件, Then 现有测试文件中的重复辅助函数定义已删除
- Given 审查全部 PR 合并后的代码库, When 检查测试覆盖, Then 各 Feature 测试文件独立且覆盖对应业务域

---

## Story 8: 开发者补全单元测试缺口

**As a** 开发者
**I want to** 为 permission_handler、ConvertToMain、UpdateTeam、GetByBizKey 等未测试方法补充单元测试
**So that** 所有 service/handler 层公开方法都有至少一个单元测试，消除测试盲区

**Acceptance Criteria:**
- Given permission_handler.go 完全无测试, When 补充测试后, Then `GetPermissions` 和 `GetPermissionCodes` 各有至少一个通过测试
- Given ItemPoolService.ConvertToMain 无测试, When 补充测试后, Then 事务性场景（创建主项+更新池状态）有至少一个通过测试
- Given TeamService.UpdateTeam 无测试, When 补充测试后, Then PM 权限检查和字段更新各有至少一个通过测试
- Given 3 个 GetByBizKey 方法无测试, When 补充后, Then 各方法存在/不存在场景各有至少一个通过测试
