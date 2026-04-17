---
date: "2026-04-17"
design_path: "docs/features/pm-work-tracker/design/tech-design.md"
prd_path: "docs/features/pm-work-tracker/prd/prd-spec.md"
evaluator: Claude (automated)
---

# Design 评估报告

---

## 总评: A

```
╔═══════════════════════════════════════════════════════════════════╗
║                      DESIGN QUALITY REPORT                        ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  1. 架构清晰度 (Architecture Clarity)               Grade: A     ║
║     ├── 层级归属明确                                [A]          ║
║     ├── 组件图存在                                  [A]          ║
║     └── 依赖关系列出                                [A]          ║
║                                                                   ║
║  2. 接口与模型定义 (Interface & Model)               Grade: B     ║
║     ├── 接口有类型签名                              [A]          ║
║     ├── 模型有字段类型和约束                         [A]          ║
║     └── 可直接驱动实现                              [B]          ║
║                                                                   ║
║  3. 错误处理 (Error Handling)                        Grade: A     ║
║     ├── 错误类型定义                                [A]          ║
║     ├── 传播策略清晰                                [A]          ║
║     └── HTTP 状态码映射                             [A]          ║
║                                                                   ║
║  4. 测试策略 (Testing Strategy)                      Grade: A     ║
║     ├── 按层级分解                                  [A]          ║
║     ├── 覆盖率目标                                  [A]          ║
║     └── 测试工具指定                                [A]          ║
║                                                                   ║
║  5. 可拆解性 (Breakdown-Readiness) ★                Grade: A     ║
║     ├── 组件可枚举                                  [A]          ║
║     ├── 任务可推导                                  [A]          ║
║     └── PRD 验收标准覆盖                            [A]          ║
║                                                                   ║
║  6. 安全考量 (Security)                              Grade: A     ║
║     ├── 威胁模型                                    [A]          ║
║     └── 缓解措施                                    [A]          ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝
```

★ Breakdown-Readiness 是关键门控维度，直接决定能否进入 `/breakdown-tasks`

---

## 结构完整性

| Section                  | 状态      | 备注 |
| ------------------------ | --------- | ---- |
| Overview + 技术栈        | ✅        | React SPA + Go/Gin + SQLite/MySQL，部署拓扑图完整 |
| Architecture (层级+图)   | ✅        | 四层 ASCII 图 + 目录结构图双图并存 |
| Interfaces               | ✅        | 9 个 Service 接口，全部带类型签名 |
| Data Models              | ✅        | 6 个 GORM 模型，字段类型与约束完整 |
| Error Handling           | ✅        | AppError 类型定义 + HTTP 映射表 + TeamScope 传播流程 |
| Testing Strategy         | ✅        | 按层分解，覆盖率目标数值化，工具链明确 |
| Security Considerations  | ✅        | JWT 流程、RBAC 中间件、团队隔离、5 项威胁缓解 |
| Open Questions           | ✅        | 4 条待决策项，均已标注 v1 默认选择 |
| Alternatives Considered  | ✅        | 4 个方案对比，含 Pros/Cons/Why Not |

---

## 1. 架构清晰度 - Grade: A

| 检查项 | 状态 | 备注 |
|--------|------|------|
| 明确说明所属层级 | ✅ | Transport / Service / Repository / Database 四层明确 |
| 有组件图（ASCII/文字） | ✅ | 层级图 + 目录结构图均为 ASCII |
| 数据流向可追踪 | ✅ | Browser → Nginx → Go API → DB 拓扑图清晰 |
| 内外部依赖列出 | ✅ | 后端 4 个外部包 + 前端 4 个外部包均列出 |
| 与项目现有架构一致 | ✅ | 标准 Go 分层架构，前端 React + Vite 符合项目约定 |

**问题**: 无明显问题。前端 Gantt 库列出两个候选（dhtmlx-gantt / frappe-gantt），最终选型未确定。

**建议**: 在 Open Questions 中补充 Gantt 库选型决策，或在实现前确认，避免影响前端任务拆解。

---

## 2. 接口与模型定义 - Grade: B

| 检查项 | 状态 | 备注 |
|--------|------|------|
| 接口方法有参数类型 | ✅ | 所有 Go 接口方法均有完整参数类型 |
| 接口方法有返回类型 | ✅ | 所有方法均有返回类型，含 error |
| 模型字段有类型 | ✅ | 6 个 GORM 模型字段类型完整 |
| 模型字段有约束（not null、index 等） | ✅ | GORM tag 中 size/not null/index/default 均标注 |
| 所有主要组件都有定义 | ✅ | 9 个 Service 接口 + 6 个 Model 全覆盖 |
| 开发者可直接编码，无需猜测 | ⚠️ | DTO 结构体（dto.CreateMainItemReq 等）在接口中引用但未定义为 Go 代码 |

**问题**: 接口签名中大量引用 `dto.*` 类型（如 `dto.CreateMainItemReq`、`dto.MainItemFilter`、`dto.PageResult[T]`、`dto.WeeklyViewResult` 等），但 DTO 结构体定义未在设计文档中给出。开发者需要从 API Handbook 反推 Go 结构体，存在一定猜测空间（字段命名风格、可选字段处理方式）。

**建议**: 补充核心 DTO 结构体定义，至少覆盖：`PageResult[T]`（泛型分页容器）、`MainItemFilter`、`SubItemFilter`、`WeeklyViewResult`、`GanttResult`。其余 CRUD 请求 DTO 可由 API Handbook 直接推导，无需重复。

---

## 3. 错误处理 - Grade: A

| 检查项 | 状态 | 备注 |
|--------|------|------|
| 自定义错误类型或错误码定义 | ✅ | AppError 结构体 + 13 个预定义错误变量 |
| 层间传播策略明确 | ✅ | Service 返回 AppError，Handler 读取 Status 字段映射 HTTP 响应 |
| HTTP 状态码与错误类型映射 | ✅ | 完整映射表，覆盖 401/403/404/400/422/500 |
| 调用方行为说明 | ⚠️ | 服务端行为完整；前端对各错误码的处理行为（如 401 跳转登录、403 提示无权限）未在设计中显式说明，但 UI Design 中有部分覆盖 |

**问题**: 前端错误处理行为分散在 UI Design 文档中，技术设计本身未集中说明前端 axios 拦截器的统一错误处理策略（如 401 自动跳转、全局 message.error 兜底）。

**建议**: 在设计文档中补充一段前端错误处理约定，说明 axios 响应拦截器的行为规则，与后端错误码形成完整闭环。

---

## 4. 测试策略 - Grade: A

| 层级 | 测试类型 | 工具 | 覆盖率目标 | 状态 |
|------|----------|------|------------|------|
| Service (Go) | 单元测试 | testify/mock | ≥ 80% | ✅ |
| Repository (Go) | 集成测试 | go test + in-memory SQLite | ≥ 60% | ✅ |
| Handler (Go) | 集成测试 | go test + SQLite | ≥ 70% | ✅ |
| Frontend 组件 | 单元 + 集成 | Vitest + RTL + MSW | ≥ 60% | ✅ |

**问题**: E2E 测试明确标注为 v1 不做，合理。Handler 层测试工具未单独说明（推测与集成测试共用 SQLite），可更明确。

**建议**: 无需修改，测试策略已达 A 级标准。可选：在 Handler 测试中补充 `httptest` 的使用说明，明确 Handler 层测试是通过 HTTP 请求驱动还是直接调用函数。

---

## 5. 可拆解性 - Grade: A ★

| 检查项 | 状态 | 备注 |
|--------|------|------|
| 组件/模块可枚举（能列出清单） | ✅ | 9 个 Service、6 个 Model、8 个前端页面、3 个中间件、6 个 Repo 均可枚举 |
| 每个接口 → 可推导出实现任务 | ✅ | 每个 Service 接口方法 → 1 个实现任务；每个 Handler → 1 个路由注册任务 |
| 每个数据模型 → 可推导出 schema/迁移任务 | ✅ | 6 个 Model → 6 个迁移文件任务，约束和索引均已标注 |
| 无模糊边界（"共享逻辑"等） | ✅ | 层间调用规则明确："所有跨域调用通过 Service 接口，不直接跨 Handler/Repo" |
| PRD 验收标准在设计中均有体现 | ✅ | 见下方覆盖检查 |

**PRD 验收标准覆盖检查**:

| PRD 功能 | 设计覆盖 |
|---------|---------|
| 用户认证与鉴权（多角色） | ✅ AuthService + JWT + RBAC 中间件 |
| 团队管理（创建/邀请/移除/转让/解散） | ✅ TeamService 全覆盖 |
| 超级管理员全局视图与权限管理 | ✅ AdminService + SuperAdmin bypass |
| 主事项 CRUD + 归档 | ✅ MainItemService |
| 子事项 CRUD + 状态流转 8 种状态 | ✅ SubItemService.ChangeStatus |
| 进度记录追加式 + PM 修正 | ✅ ProgressService.Append + CorrectCompletion |
| 完成度自动汇总（加权平均） | ✅ MainItemService.RecalcCompletion，SubItem.Weight 字段 |
| 延期 ≥2 次自动升级 P1 + 重点事项 | ✅ SubItemService.ChangeStatus 中 DelayCount 逻辑 |
| 超期高亮 | ✅ isOverdue 字段（Gantt 响应）+ 前端红色标注 |
| 事项池（提交/分配/拒绝） | ✅ ItemPoolService，assign 原子事务 |
| 周视图（计划 vs 实际） | ✅ ViewService.WeeklyView |
| 甘特图视图 | ✅ ViewService.GanttView |
| 表格视图 + CSV 导出 | ✅ ViewService.TableView + TableExportCSV |
| 周报预览 + Markdown 导出（< 5s） | ✅ ReportService，ExportMarkdown 注释 `< 5s` |
| 团队数据隔离 | ✅ TeamScopeMiddleware + 所有 Repo 强制 team_id 过滤 |

**未覆盖的 PRD 验收标准**: 无

**问题**: Open Question 中"PM 修正完成度：原地更新 vs 插入新记录"尚未决策，但 ProgressRecord 模型已按原地更新设计（IsPMCorrect 标记）。若最终选择插入新记录，模型需调整。

**建议**: 在进入 breakdown-tasks 前关闭此 Open Question，确认 ProgressRecord 修正方案，避免实现中途改模型。

---

## 6. 安全考量 - Grade: A

| 检查项 | 状态 | 备注 |
|--------|------|------|
| 威胁模型识别 | ✅ | SQL 注入、密码泄露、Mass Assignment、CORS、暴力破解 5 类威胁 |
| 缓解措施具体 | ✅ | GORM 参数化查询、bcrypt cost≥12、显式 DTO、CORS 中间件、登录限速 10req/min |
| 与功能风险面匹配 | ✅ | 多租户隔离（TeamScopeMiddleware）、JWT 无状态设计、token 存内存而非 localStorage |

**问题**: JWT 登出为客户端丢弃 token，服务端无黑名单。设计已明确标注"v1 stateless design"，属于已知权衡，不扣分。

**建议**: 无需修改。若 v2 引入 refresh token，需补充 token 撤销机制设计。

---

## 优先改进项

| 优先级 | 维度 | 问题 | 建议操作 |
|--------|------|------|----------|
| P1 | 接口与模型定义 | 核心 DTO 结构体（PageResult、Filter 类型View 结果类型）未定义为 Go 代码 | 补充 5-8 个核心 DTO 结构体定义，可作为设计附录 |
| P2 | 可拆解性 | Open Question：PM 修正完成度方案未关闭，影响 ProgressRecord 模型最终形态 | 在 breakdown-tasks 前决策并更新模型定义 |
| P3 | 架构清晰度 | 前端 Gantt 库选型未定（dhtmlx-gantt vs frappe-gantt） | 确认选型后更新依赖列表，避免前端任务估时偏差 |

---

## 结论

- **可以进入 `/breakdown-tasks`**: 是
- **预计可拆解任务数**: ~45-55 个（后端：9 Service 实现 + 6 Model/Migration + Handler 路由 + 中间件；前端：8 页面 + 共享组件 + API Client + 状态管理；测试：各层测试套件）
- **建议**: 设计质量高，结构完整，PRD 全覆盖，可直接进入任务拆解；建议在拆解前补充核心 DTO 定义并关闭 PM 修正完成度的 Open Question。
