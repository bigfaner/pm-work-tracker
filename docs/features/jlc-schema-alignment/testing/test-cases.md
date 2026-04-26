---
feature: "jlc-schema-alignment"
sources:
  - prd/prd-user-stories.md
  - prd/prd-spec.md
generated: "2026-04-25"
---

# Test Cases: jlc-schema-alignment

## Summary

| Type | Count |
|------|-------|
| UI   | 1     |
| API  | 6     |
| CLI  | 5     |
| **Total** | **12** |

---

## UI Test Cases

## TC-001: 主事项详情页 URL 使用 bizKey 导航
- **Source**: Story 6 / AC-1
- **Type**: UI
- **Target**: ui/items-detail
- **Test ID**: ui/items-detail/main-item-detail-url-uses-bizkey
- **Pre-conditions**: 用户已登录；后端已部署新版本 API（响应含 bizKey，不含 id）；事项清单页存在至少一条主事项
- **Route**: /items/:mainItemId
- **Element**: L-003, E-005
- **Steps**:
  1. 导航至 `/items`（事项清单页）
  2. 点击任意主事项行进入详情页
- **Expected**: 浏览器地址栏 URL 格式为 `/items/{bizKey}`，其中 bizKey 为 16 位左右的雪花算法数字（如 `/items/1234567890123456`），不为自增短整数（如 `/items/1`）；详情页正常加载，面包屑导航（E-021）显示正确
- **Priority**: P0

---

## API Test Cases

## TC-002: SoftDelete API 调用设置 deleted_flag 和 deleted_time
- **Source**: Story 2 / AC-1
- **Type**: API
- **Target**: api/soft-delete
- **Test ID**: api/soft-delete/soft-delete-sets-deleted-flag-and-time
- **Pre-conditions**: 数据库中存在一条 deleted_flag=0 的主事项记录；已获取该记录的 bizKey
- **Steps**:
  1. 发送 `DELETE /api/v1/teams/:teamId/main-items/:bizKey`
  2. 直接查询数据库该记录的 deleted_flag 和 deleted_time 字段
- **Expected**: HTTP 响应状态码为 200 或 204；数据库中该记录 `deleted_flag=1`，`deleted_time` 为当前时间（非默认值 `1970-01-01 08:00:00`）；记录未被物理删除（行仍存在）
- **Priority**: P0

## TC-003: 已软删记录不出现在列表和详情 API 响应中
- **Source**: Story 3 / AC-1
- **Type**: API
- **Target**: api/main-items
- **Test ID**: api/main-items/deleted-records-excluded-from-responses
- **Pre-conditions**: 数据库中存在 deleted_flag=1 的主事项记录（bizKey 已知）
- **Steps**:
  1. 发送 `GET /api/v1/teams/:teamId/main-items`（列表接口）
  2. 发送 `GET /api/v1/teams/:teamId/main-items/:bizKey`（详情接口，使用已软删记录的 bizKey）
- **Expected**: 列表响应中不包含 deleted_flag=1 的记录；详情接口返回 404 或业务错误码，不返回已删除记录数据
- **Priority**: P0

## TC-004: 资源 API 响应使用新字段名
- **Source**: Story 4 / AC-1
- **Type**: API
- **Target**: api/resources
- **Test ID**: api/resources/response-uses-new-field-names
- **Pre-conditions**: 后端已部署新版本；数据库中存在主事项、子事项、待办事项、用户各至少一条记录
- **Steps**:
  1. 发送 `GET /api/v1/teams/:teamId/main-items`
  2. 发送 `GET /api/v1/teams/:teamId/sub-items`（或主事项下子事项列表）
  3. 发送 `GET /api/v1/teams/:teamId/item-pools`
  4. 发送 `GET /api/v1/users`（或当前用户信息接口）
- **Expected**: 所有响应体中：`status` 字段已替换为 `itemStatus`（主/子事项）、`poolStatus`（待办事项）、`userStatus`（用户）；`createdAt` 已替换为 `createTime`；`updatedAt` 已替换为 `dbUpdateTime`；不存在 `deletedAt` 字段
- **Priority**: P0

## TC-005: 资源 API 响应包含 bizKey 且不含 id
- **Source**: Story 4 / AC-1; Story 6 / AC-1
- **Type**: API
- **Target**: api/resources
- **Test ID**: api/resources/response-contains-bizkey-excludes-id
- **Pre-conditions**: 后端已部署新版本；数据库中存在至少一条主事项记录
- **Steps**:
  1. 发送 `GET /api/v1/teams/:teamId/main-items`
  2. 检查响应体中每个资源对象的字段列表
- **Expected**: 每个资源对象包含 `bizKey` 字段（int64 类型数值）；不包含 `id` 字段（json:"-" 不对外暴露）
- **Priority**: P0

## TC-006: 后端通过 bizKey 路径参数正确定位记录
- **Source**: Story 6 / AC-3
- **Type**: API
- **Target**: api/main-items
- **Test ID**: api/main-items/backend-resolves-record-by-bizkey
- **Pre-conditions**: 数据库中存在一条主事项，bizKey 已知（如 1234567890123456）
- **Steps**:
  1. 发送 `GET /api/v1/teams/:teamId/main-items/1234567890123456`（路径参数为 bizKey 值）
  2. 检查响应体
- **Expected**: HTTP 200；响应体中 `bizKey` 字段值与路径参数一致；返回正确的主事项数据（标题、状态等字段正确）
- **Priority**: P0

## TC-007: 前端 API 模块使用 bizKey 构造请求路径
- **Source**: Story 6 / AC-2
- **Type**: API
- **Target**: api/frontend-module
- **Test ID**: api/frontend-module/api-module-uses-bizkey-in-path-params
- **Pre-conditions**: 前端已更新 API 模块代码
- **Steps**:
  1. 在前端 API 模块（`frontend/src/api/mainItems.ts` 等）中检查 GET/PUT/DELETE 请求的路径构造逻辑
  2. 执行前端单元测试，验证 API 调用路径
- **Expected**: 所有资源的 GET/PUT/DELETE 请求路径使用 `bizKey` 作为路径参数（如 `/api/v1/teams/1/main-items/${item.bizKey}`）；不使用 `item.id` 构造路径；前端单元测试中 mock 验证路径包含 bizKey 值
- **Priority**: P0

---

## CLI Test Cases

## TC-008: schema.sql 在 MySQL 8.0 上执行无语法错误
- **Source**: Story 1 / AC-1
- **Type**: CLI
- **Target**: cli/schema-execute
- **Test ID**: cli/schema-execute/schema-executes-on-mysql-80-without-errors
- **Pre-conditions**: 已安装 MySQL 8.0 实例；具备数据库创建权限
- **Steps**:
  1. 创建测试数据库：`CREATE DATABASE test_schema_alignment;`
  2. 执行 schema：`mysql -u root -p test_schema_alignment < backend/migrations/schema.sql`
  3. 检查执行输出
- **Expected**: 命令退出码为 0；无 `ERROR 1064` 或其他 MySQL 语法错误；所有表成功创建；每张业务表包含字段：`create_time`、`db_update_time`、`deleted_flag`、`deleted_time`、`biz_key`
- **Priority**: P0

## TC-009: schema.sql 表结构符合 JLC 规范
- **Source**: Story 5 / AC-1
- **Type**: CLI
- **Target**: cli/schema-compliance
- **Test ID**: cli/schema-compliance/schema-tables-comply-with-jlc-standard
- **Pre-conditions**: schema.sql 已在 MySQL 8.0 上成功执行（TC-008 通过）
- **Steps**:
  1. 执行：`SELECT TABLE_NAME, TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA='test_schema_alignment';`
  2. 执行：`SELECT TABLE_NAME, INDEX_NAME FROM information_schema.STATISTICS WHERE TABLE_SCHEMA='test_schema_alignment';`
  3. 执行：`SELECT TABLE_NAME, COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='test_schema_alignment' AND COLUMN_NAME='status';`
- **Expected**: 所有业务表的 TABLE_COMMENT 非空；所有索引名称以 `idx_` 或 `uk_` 开头；不存在名为 `status` 的字段（已重命名为 user_status / item_status / pool_status）
- **Priority**: P1

## TC-010: schema.sql 无 TEXT 字段和非规范数值类型
- **Source**: Story 5 / AC-1; Spec 5.1（TEXT→VARCHAR、REAL→DECIMAL）
- **Type**: CLI
- **Target**: cli/schema-types
- **Test ID**: cli/schema-types/schema-has-no-text-or-float-fields
- **Pre-conditions**: schema.sql 已在 MySQL 8.0 上成功执行（TC-008 通过）
- **Steps**:
  1. 执行：`SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='test_schema_alignment' AND DATA_TYPE IN ('text','mediumtext','longtext','float','double','real');`
- **Expected**: 查询结果为空（0 行）；原 7 个 TEXT 字段已全部替换为 VARCHAR；原 `completion REAL` 已替换为 `DECIMAL(5,2)`
- **Priority**: P1

## TC-011: 后端测试套件全部通过
- **Source**: Story 3 / AC-2
- **Type**: CLI
- **Target**: cli/go-test
- **Test ID**: cli/go-test/backend-test-suite-passes-with-zero-failures
- **Pre-conditions**: 后端 Go 代码已完成 model/repo/service 层适配
- **Steps**:
  1. 在项目根目录执行：`cd backend && go test ./...`
- **Expected**: 命令退出码为 0；输出中无 `FAIL` 行；所有测试包显示 `ok`
- **Priority**: P0

## TC-012: 前端测试套件全部通过
- **Source**: Story 4 / AC-2
- **Type**: CLI
- **Target**: cli/npm-test
- **Test ID**: cli/npm-test/frontend-test-suite-passes-with-zero-failures
- **Pre-conditions**: 前端代码已完成 API 模块和组件字段引用更新
- **Steps**:
  1. 在项目根目录执行：`cd frontend && npm test`
- **Expected**: 命令退出码为 0；输出中无测试失败（0 failed）；所有测试文件通过
- **Priority**: P0

---

## Traceability

| TC ID | Source | Type | Target | Priority |
|-------|--------|------|--------|----------|
| TC-001 | Story 6 / AC-1 | UI | ui/items-detail | P0 |
| TC-002 | Story 2 / AC-1 | API | api/soft-delete | P0 |
| TC-003 | Story 3 / AC-1 | API | api/main-items | P0 |
| TC-004 | Story 4 / AC-1 | API | api/resources | P0 |
| TC-005 | Story 4 / AC-1; Story 6 / AC-1 | API | api/resources | P0 |
| TC-006 | Story 6 / AC-3 | API | api/main-items | P0 |
| TC-007 | Story 6 / AC-2 | API | api/frontend-module | P0 |
| TC-008 | Story 1 / AC-1 | CLI | cli/schema-execute | P0 |
| TC-009 | Story 5 / AC-1 | CLI | cli/schema-compliance | P1 |
| TC-010 | Story 5 / AC-1; Spec 5.1 | CLI | cli/schema-types | P1 |
| TC-011 | Story 3 / AC-2 | CLI | cli/go-test | P0 |
| TC-012 | Story 4 / AC-2 | CLI | cli/npm-test | P0 |
