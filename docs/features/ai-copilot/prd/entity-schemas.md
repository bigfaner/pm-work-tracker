---
feature: "AI Copilot 对话助手"
---

# Entity Schemas（前端渲染契约）

> 定义 6 个实体（MainItem / SubItem / Milestone / MilestoneMap / ProgressRecord / ItemPool）的字段 schema，驱动聊天面板中表单卡片（UF-3）与结果卡片（UF-4）的渲染。
> **Schema 仅描述"如何渲染与展示"，不含业务逻辑**——状态机转移、RBAC、跨字段校验仍由后端服务层负责，前端通过现有 API 复用。

## 设计原则

1. **schema = 视图模型**：定义字段集 + 控件类型 + 展示槽位；业务规则（状态机、权限、校验）由后端服务层负责。
2. **role 按渲染语义命名**：不叫 `planStartDate` / `expectedEndDate`（业务字段名），而叫 `date`（渲染语义）；同一 role 可被多业务字段共用，避免 role 集合爆炸。
3. **一份 schema 同时驱动 form 与 result**：字段定义在 form 与 result 间天然对齐，避免"展示了一个表单里没有的字段"这类错配。
4. **新增实体 = 加 schema 文件**：渲染器零改动；后端实现新实体的 CRUD/状态机/RBAC（已有业务逻辑层），前端 ENTITY_SCHEMAS 加 entry 即可。

## Role 字段语义角色（11 个）

每个字段在 schema 中声明其 `role`，渲染器按 role 把字段塞进对应槽位。Role 集合**收敛**——6 个实体的所有字段都能映射到这 11 个 role 之一。

| Role | 槽位（result） | 控件（form） | 展示形态（result） | 出现实体 |
|------|--------------|-------------|------------------|---------|
| `title` | head 行左 | Input | 13px 500 text-primary | 全部（ProgressRecord 的 title 取自 subItem.title） |
| `code` | head 行右 | readonly（自动生成） | 11px text-tertiary | 全部 |
| `priority` | fields 行 | Select [P1,P2,P3] | badge（P0=error / P1=warning / P2,P3=neutral） | MainItem, ItemPool |

> **P0 的特殊性**：表单可选优先级为 P1/P2/P3（用户创建时不可选 P0）；P0 仅出现在**查询结果展示**与 `PRIORITY_BADGE` badge 映射里，用于渲染历史/系统注入的 P0 数据（如 QUERY_DATA_P0 mock）。即"P0 可被查询展示，但不可经 Copilot 创建"。
| `status` | fields 行 | Select（实体特定状态集） | badge（success/neutral/warning/error，按状态语义映射） | MainItem, SubItem, Milestone, MilestoneMap, ItemPool |
| `assignee` | fields 行 | Select（团队成员） | 👤 {name} | MainItem, SubItem |
| `submitter` | fields 行 | Select（默认当前用户，提交后 readonly） | 👤 {name} | ItemPool |
| `date` | meta 行 | DatePicker | 📅 {label} {yyyy/MM/dd} | 全部（planStart/expectedEnd/createdAt 等） |
| `parent` | meta 行 | Select（父实体列表） | 📁 {label} {parent title} | SubItem→MainItem, Milestone→MilestoneMap |
| `team` | meta 行 | Select（用户所属 Team） | 👥 {teamName} | MilestoneMap |
| `progress` | progress 条 | Number 0–100 | progress bar + label（如"75% · 3/4 子任务"） | MainItem（聚合自 SubItem）, SubItem（直接字段）, Milestone（聚合自 SubItem）, MilestoneMap（聚合自 Milestone） |
| `text` | text 块 | Textarea | 12px text-secondary，超长截断 1–2 行 | 全部（description / background / expectedOutput / achievement） |

### 槽位映射（result 卡片）

```
┌─────────────────────────────────────────┐
│ {title}              {code}          ›  │ ← head 行
├─────────────────────────────────────────┤
│ {priority badge} {status badge} 👤 {assignee} │ ← fields 行
├─────────────────────────────────────────┤
│ 📅 截止 {date}  📁 {parent}  👥 {team}    │ ← meta 行
├─────────────────────────────────────────┤
│ ████████░░ 75% · 3/4 子任务             │ ← progress 条（可选）
├─────────────────────────────────────────┤
│ 描述：xxxx                               │ ← text 块（可选，可多行）
└─────────────────────────────────────────┘
```

## Schema 字段定义格式

```yaml
fields:
  - name: title            # 业务字段名（与后端 API 字段对齐）
    role: title            # 渲染语义角色（11 选 1）
    control: input         # 表单控件类型：input/select/datepicker/textarea/number/readonly
    required: true         # 是否必填（默认 false）
    label: "标题"          # 显示标签
    options: [...]         # select 类控件的静态选项
    source: teamMembers    # select 类控件的动态数据源（从后端预加载）
    derived: false         # AI 推导值（form 模式渲染"AI 推断"标记）
    default: null          # 默认值
    value_from: null       # 值来自关联实体字段（如 ProgressRecord.title 来自 subItem.title）

result_slots:
  head: [title, code]            # 简写：role/name 列表
  fields: [priority, status, assignee]
  meta:                          # 详细写法：每项含 field + icon + label
    - {field: expectedEndDate, icon: 📅, label: 截止}
    - {field: milestoneKey, icon: 🏁, label: 里程碑}
  progress:                      # 进度条（可选）：声明数据源 + label 模板
    source: subitems             # direct（直接字段）/ subitems / subitems_of_milestone / milestones
    field: completion            # source=direct 时取此字段
    tpl: "{percent}% · {completed}/{total} 子任务"
  text: [description]            # 长文本块（按顺序展示）
```

---

## 实体 Schema

### MainItem（主事项）

```yaml
fields:
  - {name: title,           role: title,    control: input,     required: true,  label: "标题"}
  - {name: description,     role: text,     control: textarea,                   label: "描述"}
  - {name: priority,        role: priority, control: select,    options: [P1,P2,P3], required: true, label: "优先级", derived: true}
  - {name: status,          role: status,   control: select,    options: [todo,in_progress,paused,completed,cancelled], label: "状态"}
  - {name: assignee,        role: assignee, control: select,    source: teamMembers, label: "负责人", derived: true}
  - {name: planStartDate,   role: date,     control: datepicker,                  label: "计划开始"}
  - {name: expectedEndDate, role: date,     control: datepicker, required: true,   label: "预期截止"}
  - {name: milestoneKey,    role: parent,   control: select,    source: teamMilestones, required: true, label: "里程碑"}
result_slots:
  head: [title, code]
  fields: [priority, status, assignee]
  meta:
    - {field: expectedEndDate, icon: 📅, label: 截止}
    - {field: milestoneKey, icon: 🏁, label: 里程碑}
  progress: {source: subitems, tpl: "{percent}% · {completed}/{total} 子任务"}
  text: [description]
```

### SubItem（子事项）

```yaml
fields:
  - {name: title,         role: title,    control: input,     required: true, label: "标题"}
  - {name: parent,        role: parent,   control: select,    source: teamMainItems, required: true, label: "父事项"}
  - {name: description,   role: text,     control: textarea,                  label: "描述"}
  - {name: status,        role: status,   control: select,    options: [todo,in_progress,paused,completed,cancelled], label: "状态"}
  - {name: assignee,      role: assignee, control: select,    source: teamMembers, label: "负责人", derived: true}
  - {name: completion,    role: progress, control: number,    min: 0, max: 100, label: "完成度"}
  - {name: achievement,   role: text,     control: textarea,                  label: "达成说明"}
result_slots:
  head: [title, code]
  fields: [status, assignee]
  meta:
    - {field: parent, icon: 📁, label: 父事项}
    - {field: expectedEndDate, icon: 📅, label: 截止}
  progress: {source: direct, field: completion, tpl: "{value}%"}
  text: [achievement]
```

### Milestone（里程碑）

```yaml
fields:
  - {name: title,           role: title,    control: input,     required: true, label: "标题"}
  - {name: parent,          role: parent,   control: select,    source: teamMilestoneMaps, required: true, label: "里程碑图"}
  - {name: status,          role: status,   control: select,    options: [planned,in_progress,completed,cancelled], label: "状态"}
  - {name: expectedEndDate, role: date,     control: datepicker,                  label: "预期截止"}
  - {name: description,     role: text,     control: textarea,                   label: "描述"}
result_slots:
  head: [title, code]
  fields: [status]
  meta:
    - {field: parent, icon: 📁, label: 里程碑图}
    - {field: expectedEndDate, icon: 📅, label: 截止}
  progress: {source: subitems_of_milestone, tpl: "{percent}% · {completed}/{total} 子事项"}
  text: [description]
```

### MilestoneMap（里程碑图）

```yaml
fields:
  - {name: title,           role: title,    control: input,     required: true, label: "标题"}
  - {name: team,            role: team,     control: select,    source: userTeams, required: true, label: "Team"}
  - {name: status,          role: status,   control: select,    options: [planned,active,completed,cancelled], label: "状态"}
  - {name: expectedEndDate, role: date,     control: datepicker,                  label: "预期截止"}
  - {name: description,     role: text,     control: textarea,                   label: "描述"}
result_slots:
  head: [title, code]
  fields: [status, team]
  meta:
    - {field: expectedEndDate, icon: 📅, label: 截止}
    - {field: milestoneCount, icon: 🏆, label: "{value} 个里程碑"}
  progress: {source: milestones, tpl: "{percent}% · {completed}/{total} 里程碑"}
  text: [description]
```

### ProgressRecord（进度记录）

> 特殊：`title` 角色由关联实体的字段派生（取 subItem.title）；ProgressRecord 本身无独立标题字段。

```yaml
fields:
  - {name: subItem,    role: parent,   control: select,    source: teamSubItems, required: true, label: "子任务"}
  - {name: completion, role: progress, control: number,    min: 0, max: 100, required: true, label: "完成度"}
  - {name: achievement, role: text,    control: textarea,                   label: "达成说明"}
  - {name: createdAt,  role: date,     control: readonly,                   label: "记录时间"}
result_slots:
  head:
    - {field: title, value_from: subItem.title}
    - code
  fields: []
  meta:
    - {field: createdAt, icon: 📅, label: 记录时间}
  progress: {source: direct, field: completion, tpl: "{value}%"}
  text: [achievement]
```

### ItemPool（待办事项池）

```yaml
fields:
  - {name: title,          role: title,     control: input,     required: true, label: "标题"}
  - {name: background,     role: text,      control: textarea,  required: true, label: "背景"}
  - {name: expectedOutput, role: text,      control: textarea,  required: true, label: "预期产出"}
  - {name: priority,       role: priority,  control: select,    options: [P1,P2,P3], label: "优先级"}
  - {name: submitter,      role: submitter, control: select,    source: currentUser, label: "提交人", default_note: "默认当前登录用户，提交后 readonly"}
  - {name: status,         role: status,    control: select,    options: [pending,triaged,accepted,rejected], label: "状态"}
  - {name: createdAt,      role: date,      control: readonly,                   label: "提交时间"}
result_slots:
  head: [title, code]
  fields: [priority, status, submitter]
  meta:
    - {field: createdAt, icon: 📅, label: 提交时间}
  text: [background, expectedOutput]
```

---

## 扩展：新增实体流程

1. 在本文档新增实体 schema 段（fields + result_slots）
2. 后端实现该实体的 CRUD / 状态机 / RBAC（已有业务逻辑层，直接复用现有 service）
3. 前端 `ENTITY_SCHEMAS` 加对应 entry（与本文档一致）
4. 渲染器自动支持 form/result 两种模式，**零改动**

若新实体的字段无法映射到现有 11 个 role，先评估是否需扩 role 集合（如 `cost` 金额、`tags` 标签、`attachment` 附件），再加 role + 槽位 + 渲染分支。**避免按业务字段名扩展 role**（不要加 `planStartDate`、`expectedOutput` 这类）。

## 与业务逻辑的边界

| 由 schema 决定（前端渲染契约） | 由后端业务层决定（前端复用 API） |
|------------------------------|------------------------------|
| 字段集（哪些字段渲染） | 状态机转移规则（available-transitions 端点） |
| 控件类型（input/select/datepicker 等） | RBAC 权限校验 |
| 展示槽位（head/fields/meta/progress/text） | 跨字段校验（如"截止日期不能早于开始日期"） |
| 必填规则（required） | 字段值合法选项（如 priority 仅 P1/P2/P3 的兜底校验） |
| label 文案 | 业务实体间的引用合法性（如 assignee 必须是 Team 成员） |
| 派生字段标记（derived） | AI 推导值的生成（由 AI 代理负责） |

## 引用

- 被 [`prd-spec.md`](./prd-spec.md) In Scope / Functional Specs 引用为"实体字段渲染契约"
- 被 [`prd-ui-functions.md`](./prd-ui-functions.md) UF-3 / UF-4 引用为字段集与展示规则来源
- 被 [`../ui/ui-design.md`](../ui/ui-design.md) Component 3 / Component 4 引用为渲染策略
- 前端 `ENTITY_SCHEMAS` 常量按本文档实现（见 `ui/prototype/app.js`）
