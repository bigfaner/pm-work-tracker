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
| 5.1 Schema 变更 | tech-design §Data Models | — | 3.1 |
| 5.2 后端 Go 适配（BaseModel） | tech-design §Interfaces 1, 1b | — | 2.1 |
| 5.2 后端 Go 适配（Repo 接口） | tech-design §Interfaces 3b, 3c | — | 1.1 |
| 5.2 后端 Go 适配（Snowflake + 错误） | tech-design §Interfaces 4, Error Handling | — | 3.2 |
| 5.2 后端 Go 适配（NotDeleted + SoftDelete） | tech-design §Interfaces 2, 3 | — | 3.3 |
| 5.2 后端 Go 适配（FindByBizKey Repo） | tech-design §Interfaces 3b | — | 3.4 |
| 5.2 后端 Go 适配（Service 层） | tech-design §Interfaces 5, 5b | — | 3.5 |
| 5.2 后端 Go 适配（Handler 层） | tech-design §Interfaces 3b | — | 3.6 |
| 5.3 前端字段更新（types） | tech-design §Interfaces 6 | types/index.ts | 4.1 |
| 5.3 前端字段更新（API 模块） | tech-design §5.3 | frontend/src/api/*.ts | 4.2 |
| 5.4 关联性需求（页面组件） | tech-design §5.3 | frontend/src/pages/*.tsx | 4.3 |
| E2E 测试断言更新 | tech-design §Testing Strategy | — | T-test-1, T-test-2 |
