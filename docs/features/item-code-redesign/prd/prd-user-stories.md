---
feature: "事项编码体系重新设计"
---

# User Stories: 事项编码体系重新设计

## Story 1: PM 创建团队时设置团队编码

**As a** PM（项目经理）
**I want to** 在创建团队时设置一个 2~6 位的英文字母编码
**So that** 该团队下的所有事项编码都带有可识别的团队前缀，多团队场景下无需额外说明团队归属

**Acceptance Criteria:**
- Given 我在创建团队对话框中填写 Name 和 Description
- When 我在 Code 输入框中输入 2~6 位英文字母并提交
- Then 团队创建成功，团队列表页新增一列显示该 Code 值

- Given 我输入的 Code 少于 2 位、超过 6 位或包含非字母字符
- When 我失焦或点击提交
- Then 输入框下方显示 "编码须为 2~6 位英文字母"，表单不提交

- Given 我输入的 Code 与已有团队 Code 相同
- When 我点击提交
- Then 后端返回错误，输入框下方显示 "该编码已被使用"，表单不提交

---

## Story 2: 团队成员通过编码前缀识别事项所属团队

**As a** 团队成员
**I want to** 在 ItemViewPage、TableViewPage、WeeklyViewPage 等页面看到带团队前缀的事项编码（如 `FEAT-00001`）
**So that** 在多团队协作时，我可以直接从编码判断事项所属团队，无需额外确认

**Acceptance Criteria:**
- Given 我所在团队的 Code 为 `FEAT`
- When 我在 ItemViewPage 查看主事项列表
- Then 每条主事项的编码 Badge 显示 `FEAT-NNNNN` 格式（5 位序号左补零）

- Given 我在搜索框输入 `FEAT-`
- When 搜索执行
- Then 结果列表仅包含该团队的事项（前缀过滤）

- Given 我在搜索框输入完整编码 `FEAT-00001`
- When 搜索执行
- Then 结果列表仅包含该编码对应的一条事项

---

## Story 3: 团队成员通过子事项编码在周报中稳定引用

**As a** 团队成员
**I want to** 在 MainItemDetailPage 的子事项表格中看到每个子事项的持久编码（如 `FEAT-00001-01`）
**So that** 我可以在周报、进度追踪、跨页面沟通中稳定引用子事项，不依赖数据库 ID

**Acceptance Criteria:**
- Given 主事项编码为 `FEAT-00001`，其下有 3 个子事项
- When 我打开 MainItemDetailPage
- Then 子事项表格中每行显示 `FEAT-00001-01`、`FEAT-00001-02`、`FEAT-00001-03`，编码来自 SubItem.Code 字段，非前端拼接

- Given 子事项已创建并有编码
- When 我在 SubItemDetailPage 查看该子事项
- Then 页面显示的编码与 MainItemDetailPage 中一致（同一字段，非拼接）

- Given 数据迁移完成后
- When 查询 `SELECT COUNT(*) FROM sub_items WHERE code IS NULL OR code = ''`
- Then 返回 0

---

## Story 4: 数据迁移后旧事项编码被更新为新格式

**As a** PM
**I want to** 在系统升级后，所有已有主事项和子事项的编码自动更新为新格式
**So that** 无需手动重新创建事项，历史数据可以无缝延续

**Acceptance Criteria:**
- Given 迁移前存在 `MI-0001`、`MI-0002` 等旧格式编码
- When 迁移脚本执行完成
- Then `SELECT COUNT(*) FROM main_items WHERE code LIKE 'MI-%'` 返回 0；所有编码更新为 `{TEAM_CODE}-NNNNN` 格式

- Given 迁移脚本执行中途出错
- When 事务回滚
- Then 数据库恢复到迁移前状态，无部分更新的脏数据
