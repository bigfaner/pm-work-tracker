---
feature: "config-yaml"
status: tasks
---

# Feature: config-yaml

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 后端配置系统从扁平环境变量迁移到结构化 YAML 配置，按功能域分组，支持环境变量覆盖 |
| User Stories | prd/prd-user-stories.md | 5 个用户故事：YAML 配置管理、环境变量覆盖、启动验证、初始管理员创建、连接池调优 |
| Tech Design | design/tech-design.md | 嵌套结构体 + yaml.v3 + 反射 env 覆盖，-config flag 指定路径，InitDB/SeedAdmin 重构 |

## Traceability

| PRD Section | Design Section | Tasks |
|-------------|----------------|-------|
| 配置域与字段 | Data Models > Config Struct | 1.1 |
| 流程说明 > 业务流程 | Interfaces > LoadConfig | 1.2, 1.3, 1.4 |
| 环境变量覆盖规则 | Data Models > Env Override | 1.3 |
| 验证规则 | Error Handling > Validation | 1.4 |
| 数据库连接池 | Interfaces > InitDB | 2.1 |
| 初始管理员创建 | Interfaces > SeedAdmin | 2.2 |
| 安全性需求 | Security Considerations | 1.1 (gitignore), 2.2 (bcrypt) |
| Breaking Changes | Breaking Changes | 3.1, 3.2 |
