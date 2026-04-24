---
feature: "jlc-schema-alignment"
status: design
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
| 5.1 Schema 变更 | tech-design §Data Models | - | - |
| 5.2 后端 Go 适配 | tech-design §Interfaces 1-5 | - | - |
| 5.3 前端字段更新 | tech-design §Interfaces 6 | types/index.ts, pages/*.tsx | - |
