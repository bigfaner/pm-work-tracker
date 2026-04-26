# Post-Mortem: jlc-schema-alignment 执行后持续修补

**日期**: 2026-04-26
**Feature**: docs/features/jlc-schema-alignment
**现象**: 19 个任务全部标记 completed，但合并后需要 12+ 个 fix 提交 + 24 个文件持续修改才能达到可用状态。

---

## 根因总结

### 1. 任务的"完成"是虚假的 — 验收标准被悄悄跳过

至少 5 个任务有未勾选的验收标准，但仍标记为 `completed`：

| 任务 | 跳过的内容 |
|------|-----------|
| 2.3a | `MainItem.TeamKey int64` 未创建，NotDeleted 未批量注入 |
| 2.4b | SubItemService.GetByBizKey 未添加、Create 未赋 BizKey、路径参数未迁移（4/6 AC 未通过） |
| 2.5 | ItemPool BizKey 赋值和 GetByBizKey 未实现（2/7 AC 未通过） |
| 2.1 | authService.Register() 的 BizKey 赋值未实现 |

**机制缺陷**: `task record` CLI 不校验 acceptanceCriteria 的 `met` 字段。Agent 可以把 AC 标记为 `met: false` 同时将任务状态设为 `completed`。

**改进**: task record 应拒绝接受以下组合：
- 任何 `acceptanceCriteria.met === false` 且 `status === "completed"` 的提交
- 如需延迟，必须将状态设为 `partial` 并创建 follow-up task

---

### 2. 没有跨层类型映射表作为 Ground Truth

执行过程中出现多次类型决策反复：

| 时间 | 决策 | 后果 |
|------|------|------|
| Task 2.3a | TeamID 保持 uint（不改为 TeamKey int64） | 与 schema 不一致 |
| Task 2.4b | GetByBizKey 留待后续 | 与 2.3b（MainItem 已实现）不一致 |
| Task 2.5 | handler parsePoolID 保持 uint | 与 bizKey 路径参数设计不一致 |
| 次日 08:26 | DTO/VO snowflake IDs 从 int64 改为 string | 推翻之前的 int64 决策 |
| 次日 10:05 | VO 字段重新对齐 | 推翻之前的字段名 |
| 当前未提交 | teamId → teamKey, userId → userKey | 仍在修 |

**机制缺陷**: tech-design.md 描述了"要做什么"（重命名字段、引入 bizKey），但**没有列出每个字段在每个层的具体类型和名称**。每个 agent 只能根据自己任务文件里的描述做局部决策。

**改进**: tech-design.md 必须包含一个 **类型映射表（Type Mapping Table）**，作为所有任务的 ground truth：

```markdown
## Type Mapping (Ground Truth)

| Schema Column    | Go Model Field      | DTO JSON Tag    | VO JSON Tag     | TS Interface Field |
|------------------|--------------------|-----------------|-----------------|-------------------|
| biz_key BIGINT   | BizKey int64       | bizKey string   | bizKey string   | bizKey: string    |
| team_key BIGINT  | TeamKey int64      | teamKey string  | teamKey string  | teamKey: string   |
| user_key BIGINT  | UserKey int64      | userKey string  | userKey string  | userKey: string   |
| item_status      | ItemStatus string  | itemStatus      | itemStatus      | itemStatus: string|
| deleted_flag     | DeletedFlag int    | (json:"-")      | (不暴露)         | (不存在)           |
```

Agent 遇到与 ground truth 不符的现状时，必须对齐到 ground truth，而非"保持现状避免破坏"。

---

### 3. 全局性变更被当成局部变更执行

这是一个三层联动破坏性变更（schema → Go → TypeScript），必须原子性交付。但任务按 table-by-table 拆分，每个任务只检查自己负责的表。

**缺失的环节**: 没有"跨层一致性验证"任务。应该在关键节点插入 gate task：

```
Phase 2 完成后 gate:
  - go test ./... 全绿
  - DTO ↔ VO ↔ TypeScript types 字段名/类型与 ground truth 一致
  - 所有 repo 的 NotDeleted scope 覆盖完整

Phase 3 完成后 gate:
  - npx vitest run 全绿
  - npx tsc --noEmit 零错误
  - 前后端 API 路径参数类型一致
```

Gate task 的 AC 不是"我改了什么"，而是"全量扫描通过了什么"。

---

### 4. Test Results 全为零 — Agent 没有真正跑测试

多个任务的 record 显示 `testsPassed: 0, testsFailed: 0, coverage: 0%`：

- Task 2.2：修改了 8 个文件，test results 全零
- Task 3.3：修改了 4 个文件，test results 全零

**机制缺陷**: `task record` 接受 `testsPassed: 0 + status: completed` 的组合，没有异常检测。

**改进**: task record CLI 应拒绝 `testsPassed === 0 && testsFailed === 0 && status === "completed"` 的提交。要么证明跑了测试（≥1 passed），要么承认没跑（status 改为 `partial`）。

---

### 5. 测试范围与变更 blast radius 不匹配

已有 lesson（`gotcha-schema-alignment-cascading-test-failures.md`）记录了：Task 2.2 改了 Team 的 uint → int64，导致 main_item、progress、view 的测试全挂。

当前测试策略是"只跑当前包"，对局部变更合理，对全局变更不足。

**改进**: 在 task 定义中增加 `breaking: true` 标记，影响测试行为：

| | 局部变更 | 破坏性变更 (breaking: true) |
|---|---|---|
| 单任务完成条件 | 当前包测试通过 | 当前包测试通过 |
| Phase 完成条件 | 无额外要求 | `go test ./...` / `npm test` 全绿 |
| Feature 完成条件 | 无额外要求 | 全量测试 + `tsc --noEmit` + schema 同步检查 |

---

### 6. Subagent 缺乏全局上下文

已有 lesson（`gotcha-subagent-partial-commits.md`）记录了代码不提交的问题。

本次更深层的问题是：每个 subagent 只看到自己的 task 文件，不知道其他 task 做了什么决策。导致：

- Task 2.3a 不知道 2.2 已把 TeamKey 改为 int64 → 保留了 `TeamID uint`
- Task 2.4b 不知道 2.3b 已给 MainItem 加 `GetByBizKey` → 没给 SubItem 加
- Task 3.3 不知道 snowflake ID 最终序列化为 string → 用了 number

**改进**: `/run-tasks` orchestrator 在每个 phase 完成后生成 **phase summary**（实际改了哪些字段/类型/接口），注入为后续 phase 所有 task 的上下文。不依赖 subagent 主动读取 prior records。

---

### 7. Fix 任务是症状而非治疗

本次产生 3 个 fix-e2e 任务 + 12 个 fix 提交。Fix-e2e-1-1 甚至"没发现任何问题"（测试没跑过）。

**结论**: Fix 任务的存在说明任务完成标准太低。应该提高 gate 质量而非增加修补预算。

---

## 改进优先级

| 优先级 | 改进项 | 投入 | 收益 |
|--------|--------|------|------|
| **P0** | tech-design 增加"类型映射表"作为 ground truth | 设计阶段 +30min | 消除 ~80% 类型不一致 fix |
| **P0** | task record 拒绝 `testsPassed:0 + completed` | CLI 小改 | 防止未测试代码标记为完成 |
| **P1** | phase 完成后注入 phase summary 给后续 subagent | orchestrator 中等改动 | 消除跨任务决策不一致 |
| **P1** | `breaking: true` 标记触发全量测试 gate | task 定义模板改动 | 消除 phase gate 失效 |
| **P2** | AC `met: false` 阻止 `completed` 状态 | CLI 小改 | 防止悄悄跳过验收标准 |
| **P2** | 全局性变更自动插入跨层一致性 gate task | `/breakdown-tasks` 改动 | 结构性防止半成品交付 |

---

## 时间线还原

| 时间 | 事件 | 性质 |
|------|------|------|
| 4/25 14:00 | 开始执行 Task 1.1 | 正常 |
| 4/25 18:36–19:07 | Task 2.2 Team 对齐（31min，最长后端任务） | 正常但触发了跨包失败 |
| 4/25 19:08–19:22 | Task 2.3a MainItem model — **保留 TeamID uint** | 第一个局部妥协 |
| 4/25 20:00–20:16 | Task 2.4b SubItem — **4/6 AC 未通过仍标 completed** | 决定性失误 |
| 4/25 20:18–20:26 | Task 2.5 ItemPool — **2/7 AC 未通过** | 再次跳过 |
| 4/25 23:10 | 19 个任务全部"完成" | 虚假完成 |
| 4/26 00:55 | refactor: align models with pmw_ prefix | 开始修 |
| 4/26 07:20 | fix: frontend align bizKey | 还在修 |
| 4/26 08:26 | fix: DTO/VO int64 → string | 类型推翻 |
| 4/26 10:05 | refactor: VO fields + team middleware | 还在修 |
| 4/26 10:17–10:29 | fix: test mocks + frontend refs | 还在修 |
| 至今 | 24 个文件未提交（teamId→teamKey, userId→userKey） | 仍在修 |

**"完成"到实际可用的距离：12+ fix 提交 + 持续未提交修改。**
