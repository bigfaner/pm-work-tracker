---
created: 2026-04-19
author: faner
status: Draft
---

# Proposal: AI-Friendly Coding Conventions & Code Cleanup

## Problem

项目由 AI 辅助开发，当前存在两个核心问题：

1. **编码规范缺失**：CLAUDE.md 只有通用行为准则，没有项目级编码规范（命名规则、分层模式、组件化规则）。AI 每次会话无法获得一致的风格指引，导致同一项目内出现混合风格。
2. **已有代码质量低**：AI 生成的代码满足功能但存在大量重复和不一致：
   - 后端 model 层 JSON tag 用 snake_case，VO/DTO 层用 camelCase，前端手动桥接
   - Handler/Service/Repo 三层有大量重复样板代码（25+ 个 nil-service 检查、4 个 mapNotFound 副本、重复的分页和日期解析逻辑）
   - 前端页面组件缺少复用，相似 UI 模式在每个页面中重复实现

**Why now**: 项目功能开发进入稳定期（RBAC 基本完成），是沉淀规范、提升代码质量的合适时机。

## Proposed Solution

**规范先行，分批清理**：

### Phase 1: 建立编码规范（写入 zcode 标准文档）

| 文档 | 内容 |
|------|------|
| `docs/ARCHITECTURE.md` | 分层架构说明（Model → DTO/VO → Service → Handler → Router），每层职责和模板代码 |
| `docs/DECISIONS.md` | 技术决策记录：JSON tag 统一 camelCase、错误处理模式、分页模式等 |
| `docs/rules/` | 具体编码规则文件：<br>`naming.md` — 命名规范（JSON tag、变量、函数）<br>`patterns.md` — 分层模式（CRUD 模板、helper 提取规则）<br>`frontend.md` — 前端组件化规则（API 层、UI 组件、状态管理）|

`.claude/rules/` 作为编码规则载体（面向 AI，会话启动时自动加载，支持 `paths` 路径限定）。`docs/ARCHITECTURE.md` 和 `docs/DECISIONS.md` 保留在 `docs/` 下（面向人类）。

### Phase 2: 配置自动化 Lint 规则

- **Go 后端**：golangci-lint 配置，检查 JSON tag 命名、重复错误处理模式
- **TypeScript 前端**：ESLint 配置，检查 API 层命名、组件复用

### Phase 3: 清理现有代码

按后端→前端顺序：
1. **统一 JSON tag**：model 层迁移到 camelCase，去除前端手动桥接代码
2. **抽取公共 helper**：mapNotFound → 通用 `mapNotFound(err, targetErr)`；分页 → `setDefaultPagination(page)`；日期解析 → `parseOptionalDate(str)`
3. **统一 CRUD 模式**：repo 层提取泛型或接口减少重复
4. **前端组件化**：识别重复 UI 模式，抽取可复用组件

## Alternatives Considered

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| **A. 规范先行（推荐）** | 规范指导清理，减少返工；新代码立即受规范约束 | 前期需投入时间打磨规范 | ✅ Recommended |
| **B. 逐层重构** | 见效快，边做边定 | 无统一规范指导，可能方向不一致需要返工 | ❌ 返工风险高 |
| **C. Lint 驱动** | 最客观，可机器验证 | Go 自定义 lint 规则开发成本高，ROI 不合算 | ❌ ROI 低 |

## Scope

### In Scope
- 编写 `docs/ARCHITECTURE.md`、`docs/DECISIONS.md`（面向人类）
- 编写 `.claude/rules/*.md`（面向 AI，自动加载）
- 配置 golangci-lint 和 ESLint
- 后端：统一 JSON tag 为 camelCase
- 后端：抽取公共 helper（mapNotFound、pagination、dateParse）
- 后端：统一 repo CRUD 模式
- 前端：去除 snake_case 手动桥接
- 前端：抽取可复用 UI 组件

### Out of Scope
- 数据库列命名（保持 snake_case，由 GORM 自动映射）
- 业务逻辑变更
- 性能优化
- 新功能开发
- 修改 zcode 插件本身

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 规范文档过于理想化，实际编码时难以遵守 | Medium | High | 规范中提供具体代码示例和反例，让 AI 能直接参考 |
| 清理范围过大导致回归 bug | Medium | High | 每批清理后运行完整测试套件；分批提交 |
| lint 规则误报过多，开发体验差 | Low | Medium | 渐进式启用规则，先 warn 后 error |

## Success Criteria

- [ ] `docs/ARCHITECTURE.md`、`docs/DECISIONS.md` 已创建且内容完整
- [ ] `.claude/rules/*.md` 已创建，AI 会话自动加载
- [ ] golangci-lint 和 ESLint 配置完成，能检测命名违规
- [ ] 后端所有 model JSON tag 统一为 camelCase
- [ ] 后端重复代码减少 50% 以上（可量化：mapNotFound/pagination/dateParse 副本数）
- [ ] 前端无 snake_case 手动桥接代码
- [ ] 至少抽取 3 个可复用前端 UI 组件
- [ ] 全部现有测试通过

## Next Steps

- Proceed to `/write-prd` to formalize requirements
