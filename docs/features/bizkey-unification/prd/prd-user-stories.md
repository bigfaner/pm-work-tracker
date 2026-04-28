---
feature: "bizkey-unification"
---

# User Stories: BizKey Unification

## Story 1: 进度记录写入正确的 team_key

**As a** 系统用户（团队成员）
**I want to** 提交进度记录后，该记录能被正确地按团队查询到
**So that** 周报、进度统计等功能返回准确的数据，而不是静默地丢失或错配数据

**Acceptance Criteria:**
- Given 用户属于 bizKey 为 `123456789012345678` 的团队
- When 用户提交一条进度记录
- Then 数据库中该记录的 `team_key` 字段值为 `123456789012345678`（雪花 ID），而非该团队的内部自增 ID（如 `5`）

---

## Story 2: 角色权限判断使用正确的 bizKey

**As a** 团队 PM
**I want to** 邀请成员时系统能正确拒绝将 PM 角色分配给普通成员
**So that** 角色权限边界得到正确执行，不因 ID 类型混用而出现误判

**Acceptance Criteria:**
- Given 系统中 PM 角色的 bizKey 为 `987654321098765432`，内部 ID 为 `2`
- When PM 尝试邀请成员并指定角色 bizKey 为 `987654321098765432`
- Then 系统正确识别该角色为 PM 角色并返回 `ErrCannotAssignPMRole` 错误

---

## Story 3: 后端开发者新增功能时类型系统提供约束

**As a** 后端开发者
**I want to** 在编写新的 Service 方法时，编译器能阻止我将 uint 内部 ID 误传给期望 int64 bizKey 的参数
**So that** 不需要依赖代码审查来发现 uint/int64 混用问题，减少同类 Bug 引入

**Acceptance Criteria:**
- Given Service 接口方法签名中，来自外部输入的 ID 参数类型为 `int64`
- When 开发者尝试将 `uint` 类型的内部 ID 直接传入该参数
- Then 编译器报类型不匹配错误，构建失败，无法通过 `go build ./...`
