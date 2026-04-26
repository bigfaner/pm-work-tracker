---
feature: schema-alignment-cleanup
---

# User Stories: Schema Alignment Post-Refactoring Cleanup

## Story 1: 指派子项功能恢复

**As a** PM 用户
**I want to** 在子项上点击"指派"后能看到指派人被保存
**So that** 我可以将工作分配给具体负责人并追踪

**Acceptance Criteria:**
- Given 一个未指派的子项
- When 用户选择一个成员并点击"指派"
- Then 子项的 `assignee_key` 列被正确更新（而非错误地写入 `assignee_id` 列）
- And 刷新页面后，指派人名称仍然显示在子项上

---

## Story 2: 按人筛选功能恢复

**As a** PM 用户
**I want to** 在项目列表中按负责人筛选项目
**So that** 我可以快速查看某个成员负责的所有工作项

**Acceptance Criteria:**
- Given 项目列表页面，存在多个已指派给不同成员的项目
- When 用户从"按人筛选"下拉框选择某个成员
- Then 列表仅显示指派给该成员的项目（而非返回全部或空结果）
- And 该行为在 MySQL 和 SQLite 上均一致

---

## Story 2.5: 清理死代码以降低认知负担

**As a** 后端开发者
**I want to** 代码库中不再包含 deprecated DTO、无效 nil 检查、冗余 GORM tags 等死代码
**So that** 我阅读代码时不需要在脑中过滤已废弃的内容，减少误用 deprecated API 的风险

**Acceptance Criteria:**
- Given `item_dto.go` 中存在 4 个标记为 deprecated 的 DTO
- When 删除完成后
- Then `grep -rn "Deprecated" backend/internal/dto/item_dto.go` 返回零结果
- And `go build ./...` 编译通过（无引用断裂）

---

## Story 3: 代码库可维护性提升

**As a** 后端开发者
**I want to** 代码中每种模式只有一个实现（分页、事务接口、bizKey 解析、DTO 转换、状态记录）
**So that** 我修改功能时只需改动一处，不需要在 5-7 个重复位置同步修改

**Acceptance Criteria:**
- Given 代码库中存在重复的 `TransactionDB` / `dbTransactor` 接口
- When 合并为单一共享接口后
- Then `grep -r "dbTransactor" backend/` 返回零结果
- And 所有使用方编译通过且测试通过

---

## Story 4: 前端类型一致性

**As a** 前端开发者
**I want to** 前端类型定义与后端 API 响应完全一致（ID 全部为 string bizKey，字段名统一）
**So that** 我不需要在 form state 和 API 调用之间维护 `assigneeId` ↔ `assigneeKey` 的映射

**Acceptance Criteria:**
- Given 前端存在 `PermissionData.teamPermissions: Record<number, string[]>` 和表单字段 `assigneeId`
- When 类型对齐完成后
- Then `teamPermissions` 键类型为 `string`
- And `grep -rn "assigneeId" frontend/src/` 返回零结果
- And `npx vitest run` 全部通过

---

## Story 5: 表命名规范统一

**As a** 后端开发者
**I want to** 所有数据库表使用统一的 `pmw_` 前缀
**So that** 我可以通过前缀快速识别项目相关的表，避免与其他系统混淆

**Acceptance Criteria:**
- Given `roles` 和 `role_permissions` 表缺少 `pmw_` 前缀
- When 重命名完成
- Then `Role.TableName()` 返回 `"pmw_roles"`
- And `RolePermission.TableName()` 返回 `"pmw_role_permissions"`
- And `go test ./internal/model/ -run TestRole` 通过
