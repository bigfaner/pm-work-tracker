---
feature: "jlc-schema-alignment"
---

# User Stories: Schema 对齐嘉立创数据库开发规范（MySQL 兼容）

## Story 1: 后端工程师执行 MySQL 兼容 Schema

**As a** 后端开发工程师
**I want to** 在 MySQL 8.0 上无报错执行新的 `schema.sql`
**So that** 生产环境数据库可以基于规范化的 schema 初始化，不再受 SQLite 语法限制

**Acceptance Criteria:**
- Given 已安装 MySQL 8.0 实例
- When 执行新的 `schema.sql`
- Then 所有建表语句成功执行，0 个语法错误，所有表包含 `create_time`、`db_update_time`、`deleted_flag`、`deleted_time`、`biz_key` 字段

---

## Story 2: 后端工程师使用软删除接口

**As a** 后端开发工程师
**I want to** 通过 repo 层封闭的 `SoftDelete(ctx, id)` 方法执行软删除
**So that** 无法在 repo 外部意外调用硬删除，数据安全有制度保障

**Acceptance Criteria:**
- Given 任意业务表的 repo 接口
- When 调用 `SoftDelete(ctx, id)`
- Then 数据库执行 `UPDATE ... SET deleted_flag=1, deleted_time=NOW()`，记录不被物理删除
- Given repo 接口定义
- When 检查接口方法列表
- Then 不存在 `HardDelete` 或直接暴露 `db.Delete()` 的方法

---

## Story 3: 后端工程师查询时自动过滤软删记录

**As a** 后端开发工程师
**I want to** 所有 repo 查询自动通过 `NotDeleted()` scope 过滤软删记录
**So that** 不会因遗漏过滤条件而将已删除数据返回给用户

**Acceptance Criteria:**
- Given 数据库中存在 `deleted_flag=1` 的记录
- When 调用任意 repo 的 `Find`/`First`/`List` 方法
- Then 返回结果中不包含 `deleted_flag=1` 的记录
- Given 后端测试套件
- When 执行 `go test ./...`
- Then 0 个测试失败

---

## Story 4: 前端工程师消费更新后的 API 字段

**As a** 前端开发工程师
**I want to** 使用更新后的 JSON 字段名（`itemStatus`、`createTime`、`dbUpdateTime` 等）访问 API 响应
**So that** 前端组件与后端 API 保持一致，不出现字段名不匹配导致的 undefined 错误

**Acceptance Criteria:**
- Given 后端已部署新版本 API
- When 前端调用任意资源接口（main-items、sub-items、item-pools、users）
- Then 响应体中 `status` 字段已替换为 `userStatus`/`itemStatus`/`poolStatus`，`createdAt`/`updatedAt` 已替换为 `createTime`/`dbUpdateTime`，响应体中不存在 `bizKey` 字段
- Given 前端测试套件
- When 执行 `npm test`
- Then 0 个测试失败

---

## Story 5: DBA 验证 schema 符合 JLC 规范

**As a** DBA / 运维
**I want to** 确认新 schema 完整符合 JLCZD-03-016 强制规范
**So that** 数据库可以通过合规审查，不存在规范违反项

**Acceptance Criteria:**
- Given 新的 `schema.sql`
- When 逐条对照 JLC 规范检查
- Then 所有表有 `COMMENT`，所有索引符合 `idx_`/`uk_` 命名，无 `status` 关键字直接用作字段名，无 TEXT 字段，无 FLOAT/DOUBLE 用于数值字段
