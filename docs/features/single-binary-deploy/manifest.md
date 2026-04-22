---
feature: single-binary-deploy
status: tasks
---

# Manifest: 前后端单体构建

## Documents

| Type | Path | Summary |
|------|------|---------|
| PRD Spec | prd/prd-spec.md | 构建脚本、BASE_PATH 路由设计、静态文件服务、SPA 兜底、启动校验、缓存策略 |
| User Stories | prd/prd-user-stories.md | 5 个用户故事：一键构建、零依赖运行、SPA 路由直访、启动配置校验、本地开发不受影响 |
| Tech Design | design/tech-design.md | embed.FS 嵌入前端产物，路由前缀重构（去掉 /api），静态 handler，BASE_PATH 配置，构建脚本设计 |

## Traceability

| PRD Section | Design Section | Tasks |
|-------------|----------------|-------|
| §5.1 构建脚本 | Build Script + scripts/build.sh | 7.1 |
| §5.2 BASE_PATH 与路径设计 | Route Changes + Config 变更 | 1.1, 4.1 |
| §5.2 路由规则 | Route Changes（注册顺序表） | 4.1 |
| §5.2 缓存头策略 | 缓存策略表 | 3.1 |
| §5.2 启动校验 | Error Handling（启动校验失败） | 2.1, 5.1 |
| §5.3 关联性需求改动 | Frontend Changes + .gitignore Additions | 6.1, 7.2 |
| 安全性需求 | Security Considerations | 1.1, 7.2 |
