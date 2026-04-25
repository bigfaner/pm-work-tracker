---
feature: "jlc-schema-alignment"
status: tasks
---

# Feature: jlc-schema-alignment

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | Schema 对齐 JLC 规范，MySQL 兼容重写；涵盖字段重命名、软删机制替换、TEXT→VARCHAR、后端 Go 适配、前端 breaking change 同步部署 |
| User Stories | prd/prd-user-stories.md | 5 个用户故事，覆盖后端工程师（schema 执行、软删接口、查询过滤）、前端工程师（字段引用更新）、DBA（规范合规验证） |
| Tech Design | design/tech-design.md | 三层变更设计：schema DDL 重写、BaseModel 替换 + NotDeleted scope + SoftDelete 接口、前端 types 字段名更新；引入 bwmarrin/snowflake 生成 biz_key |

## Traceability

| PRD Section | Design Section | UI Component | Tasks |
|-------------|----------------|--------------|-------|
| 5.1 Schema 变更 | tech-design §Data Models | — | 1.2 |
| 5.2 后端 Go 适配（Repo 接口） | tech-design §Interfaces 3b, 3c | — | 1.1 |
| 5.2 后端 Go 适配（Snowflake + NotDeleted + 错误） | tech-design §Interfaces 2, 4, Error Handling | — | 1.3 |
| 5.2 后端 Go 适配（User 表） | tech-design §Interfaces 1, 3b, 5 | — | 2.1 |
| 5.2 后端 Go 适配（Team + TeamMember 表） | tech-design §Interfaces 1b, 3, 5b | — | 2.2 |
| 5.2 后端 Go 适配（MainItem model + repo） | tech-design §Interfaces 1, 3b | — | 2.3a |
| 5.2 后端 Go 适配（MainItem service + handler） | tech-design §Interfaces 3b, 5b | — | 2.3b |
| 5.2 后端 Go 适配（SubItem model + repo） | tech-design §Interfaces 1, 3, 3b | — | 2.4a |
| 5.2 后端 Go 适配（SubItem service + handler） | tech-design §Interfaces 3b, 5b | — | 2.4b |
| 5.2 后端 Go 适配（ItemPool 表） | tech-design §Interfaces 1, 3b, 5b | — | 2.5 |
| 5.2 后端 Go 适配（ProgressRecord + StatusHistory） | tech-design §Interfaces 1b | — | 2.6 |
| 5.3 前端字段更新（types） | tech-design §Interfaces 6 | types/index.ts | 3.1 |
| 5.3 前端字段更新（API 模块） | tech-design §Interfaces 3b | frontend/src/api/*.ts | 3.2 |
| 5.4 关联性需求（页面组件） | tech-design §Testing Strategy | frontend/src/pages/*.tsx | 3.3 |
| E2E 测试断言更新 | tech-design §Testing Strategy | — | T-test-1, T-test-2 |
