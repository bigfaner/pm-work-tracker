---
feature: "item-code-redesign"
status: design
---

# Feature: item-code-redesign

<!-- Status flow: prd → design → tasks → in-progress → done -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | 引入 Team Code 前缀重新设计主事项（TEAM-NNNNN）和子事项（TEAM-NNNNN-NN）编码体系，解决多团队场景下编码无法识别团队归属、子事项无持久编码两个问题 |
| User Stories | prd/prd-user-stories.md | 4 个用户故事：PM 创建团队设置 Code、团队成员通过编码前缀识别团队、子事项编码稳定引用、数据迁移后旧编码更新 |
| UI Functions | prd/prd-ui-functions.md | 4 个 UI 功能：创建团队对话框新增 Code 输入框（含校验）、团队列表新增 Code 列、主事项编码展示值变更、子事项编码来源从拼接改为字段读取 |
| Tech Design | design/tech-design.md | SELECT FOR UPDATE 悲观锁方案；Team/MainItem/SubItem 模型变更；migration 008；迁移程序设计；并发测试策略 |
| API Handbook | design/api-handbook.md | POST /teams 新增 code 字段；团队列表/详情响应新增 code；主事项 code 格式变更；子事项新增 code 字段 |

## Traceability

| PRD Section | Design Section | Tasks |
|-------------|----------------|-------|
| Team Code 字段 + 校验 | Data Models: model.Team；Interfaces: CreateTeamReq；Error Handling: ErrTeamCodeDuplicate | — |
| NextCode() / NextSubCode() 算法 | Interfaces: MainItemRepo.NextCode()；SubItemRepo.NextSubCode() | — |
| 数据迁移 | Data Models: Migration 008；Appendix: 迁移执行顺序 | — |
| 前端编码展示变更 | API Handbook: Data Contracts | — |
