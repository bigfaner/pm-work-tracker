---
feature: "code-conventions"
evaluated: 2026-04-19
---

# PRD Evaluation: code-conventions

## Overall Grade: A

## Dimension Scores

### Dimension 1: 背景与目标 — A

背景三要素完整且具体：

- **原因**：清晰阐述了两类结构性问题——编码规范缺失（CLAUDE.md 只有通用行为准则，无项目级规范）和已有代码质量低（列举了具体数字：25+ nil-service 检查、4 个 mapNotFound 副本、14 次 textarea 样式复制、20+ 硬编码颜色等）。还给出了时机判断（RBAC 基本完成，适合沉淀规范）。
- **对象**：明确列出三项交付物——规范文档体系、代码清理、lint 自动化。
- **人员**：用表格列出两类用户（AI 编码助手、开发者），各有明确的角色和核心诉求。

目标全部量化，共 6 项目标各有明确指标：100% 规范符合、0 个 snake_case JSON tag、重复代码副本减少 50%+、至少 3 个可复用组件、0 处硬编码颜色、lint 规则覆盖。背景中的问题与目标一一对应，逻辑一致。

### Dimension 2: 流程说明 — A

Mermaid 流程图完整存在，包含：

- **主流程**：审计 -> 编写规范 -> 更新 CLAUDE.md -> 配置 lint -> 后端清理 -> 前端清理 -> 验证
- **决策点**：3 个判断节点（后端测试通过?、前端测试通过?、所有测试通过 + lint 无违规?）
- **异常分支**：每个决策点都有"否"回退到上游步骤的分支
- **文字说明**：补充了规范建立流程、代码清理流程、开发流程规范三段描述

流程图覆盖了完整的工作路径，回退逻辑清晰。

### Dimension 3: 功能描述 — A

功能描述分为 6 个子节，全部使用了表格结构并包含具体数据：

- **5.1 规范文档体系**：两层结构（docs/ 和 .claude/rules/），每层用表格列出文档、定位/路径限定、内容，并给出关键约束（正反面示例、文件数 <= 6、行数 <= 200）
- **5.2 命名规范统一**：表格列出每层的当前状态和目标状态
- **5.3 后端去重**：表格列出 5 类重复模式、当前副本数、目标（含具体方案：nil-service 改为构造函数 panic、mapNotFound 统一为泛型函数等）
- **5.4 前端组件化与 UI 风格统一**：组件表格（3 个组件的问题和目标）、颜色表格（硬编码数量和目标）、补充规范
- **5.5 Lint 规则配置**：后端列出具体规则名（tagliatelle、govet、revive、dupl 含阈值），前端列出具体规则名（no-restricted-syntax、no-restricted-imports、naming-convention）
- **5.6 开发流程规范**：TDD 流程和 e2e 测试要求

作为开发者工具/代码质量类特性（非面向用户的 UI 功能），功能描述的完整性适配良好。每个子节都有现状数据、目标定义、表格呈现。关联性需求改动表（9 项）覆盖了所有受影响的模块和改动点。

### Dimension 4: User Stories — A

共 6 个 Story（Story 1 - Story 5 + Story 4.5），覆盖两类目标用户：

- **AI 编码助手**：Story 1（自动加载规范）
- **开发者**：Story 2-5 + 4.5（命名统一、去重、组件复用、UI 风格、TDD 流程）

格式全部符合 As a / I want / So that 标准结构。每个 Story 的动作具体（不是模糊的"改善代码质量"，而是"统一 JSON tag"、"消除重复样板代码"等具体操作）。

每个 Story 都有 Acceptance Criteria，采用 Given/When/Then 格式。AC 可验证（如"mapNotFound 只存在 1 份"、"0 处硬编码颜色"、"至少 3 个可复用组件"），且都包含了"全部现有测试通过"的回归保障。

### Dimension 5: Scope Clarity — A

In Scope 和 Out of Scope 都有明确定义：

- **In Scope**：15 项具体交付物，按 Phase 1/2/3 分组（规范文档与 Lint 配置 -> 后端清理 -> 前端清理），用 checkbox 列表，每项对应功能描述中的具体工作
- **Out of Scope**：5 项明确排除项（数据库列命名、业务逻辑变更、性能优化、新功能开发、zcode 插件修改）

Scope 与功能描述、User Stories 三者一致。In Scope 的每一项都能在功能描述中找到对应细节，Out of Scope 的排除项合理（如数据库列命名保持不变，因为由 ORM 自动映射，在命名规范统一中已说明理由）。阶段划分与流程图中的顺序一致。

### Dimension 6: UI Functions — N/A

本特性为开发者工具/代码质量特性，不涉及面向终端用户的界面功能，N/A 合理。

## Summary

**Top Issues:**

1. **AC 缺少边界条件的 Given/When/Then 格式化**：部分 Story 的 AC 虽然内容具体，但严格来看是"Given X / When Y / Then Z"的单条描述，未覆盖边界场景（如 Story 2 中如果清理过程中发现前端有依赖 snake_case 的非桥接代码该怎么办）。不过对于重构类需求这是可接受的。
2. **nil-service 检查方案在 5.3 节已明确为"构造函数中 panic"**，但 In Scope checklist 中只写了"统一 repo CRUD 模式"，未单独列出 nil-service 的处理项，可能被遗漏。
3. **缺少明确的验收顺序或里程碑检查点**：虽有三阶段划分，但缺少阶段间的验收标准（如 Phase 1 完成的标志是什么——lint 配置能检出至少 N 个现有违规?）。

**Recommendation:** Proceed to /design-tech. PRD 质量高，背景充分、目标量化、功能描述具体、User Stories 完整。上述三个问题属于设计阶段可以细化的层面，不阻碍技术设计。

## Improvement Suggestions

### P0 (Must Fix)

无。PRD 整体质量满足进入设计阶段的条件。

### P1 (Should Fix)

- In Scope checklist 建议为 nil-service 检查方案添加独立的 checkbox 项，与 5.3 节的描述保持一致。
- 每个阶段建议补充阶段完成的验收标准（如 Phase 1: lint 规则配置完成且能检出至少 X 个现有违规项）。

### P2 (Nice to Have)

- User Stories 可补充 Story 间的依赖关系说明（如 Story 3 依赖 Story 2 完成后再做）。
- 开发流程规范（5.6）可给出一个具体的 feature 开发 checklist 示例，帮助验证规范的实用性。
- 5.1 节可补充一个文档关系图，说明 docs/ 和 .claude/rules/ 之间的引用和覆盖关系。
