---
title: 事项编码体系重新设计
slug: item-code-redesign
status: draft
created: 2026-04-22
---

## Problem

当前事项编码存在以下问题：

1. **主事项编码格式固定**：`MI-0001`，前缀 `MI-` 无业务含义，无法从编码识别所属团队
2. **子事项无编码**：SubItem 没有任何人类可读的标识符，只能依赖数据库 ID
3. **Team 无缩写**：团队模型缺少短代码字段，无法作为编码前缀

## Proposal

重新设计事项编码体系，引入团队缩写作为前缀，并为子事项增加编码。

### 编码格式

| 层级 | 格式 | 示例 | 组成 |
|------|------|------|------|
| Team Code | `{code}` | `FEAT` | 2~6位字母，全局唯一 |
| 主事项 | `{team_code}-{seq}` | `FEAT-00001` | 团队缩写 + 5位自增序号 |
| 子事项 | `{team_code}-{seq}-{sub_seq}` | `FEAT-00001-01` | 主事项编码 + 2位自增序号 |

### 核心规则

- **不可变**：编码生成后不再变更，即使团队 code 被修改
- **Team Code 快照**：创建事项时，将当时的 team code 写入事项编码，后续 team code 变更不影响已有编码
- **序号范围**：主事项序号 per-team 自增；子事项序号 per-main-item 从 01 开始
- **Team Code 格式**：2~6位大小写字母组合，全局唯一

### 变更范围

**Team 模型**：新增 `Code` 字段（varchar(6)，全局唯一索引）

**MainItem 模型**：
- `Code` 列从 `varchar(10)` 扩展到 `varchar(12)`
- 编码格式从 `MI-NNNN` 改为 `{team_code}-NNNNN`
- `NextCode()` 逻辑改用 team code 作为前缀

**SubItem 模型**：
- 新增 `Code` 字段（varchar(15)），per-main-item 唯一
- 新增 `NextSubCode()` 逻辑，基于主事项编码生成子事项编码

**前端**：
- 更新 TypeScript 类型定义
- 更新各页面（ItemViewPage、TableViewPage、WeeklyViewPage、MainItemDetailPage）中编码的展示和搜索
- ItemPoolPage 中关联事项编码的展示

**数据迁移**：
- 直接切换，不需要兼容旧 `MI-XXXX` 格式

## Out of Scope

- 编码的自定义前缀或可配置格式
- 编码的批量重命名
- 已删除事项编码的回收复用
- 编码的跨团队唯一性（编码仅在团队内唯一）

## Open Questions

（已通过对话全部确认，无遗留问题）
