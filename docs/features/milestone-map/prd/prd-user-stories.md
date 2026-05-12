---
feature: "里程碑图"
---

# User Stories: 里程碑图

## Story 1a: PM 创建里程碑

**As a** PM（项目经理）
**I want to** 创建里程碑（名称+计划日期）
**So that** 我能按项目交付阶段组织事项

**Acceptance Criteria:**
- Given 我拥有 milestone:create 权限，When 我在里程碑图页面点击创建并填写名称+计划日期，Then 里程碑创建成功，状态为 not_started，完成度为 0
- Given 我拥有 milestone:create 权限，When 我提交名称恰好 100 字符的里程碑，Then 创建成功；When 我提交名称为 101 字符，Then 表单显示"名称不能超过 100 字符"错误提示且不提交
- Given 我拥有 milestone:create 权限，When 我不填名称直接提交，Then 表单显示"名称不能为空"错误提示且不提交
- Given 我拥有 milestone:create 权限，When 我填写名称后提交但后端返回 500 错误，Then 页面显示"创建失败，请重试"，表单保留已填写数据不丢失

---

## Story 1b: PM 编辑里程碑信息

**As a** PM（项目经理）
**I want to** 编辑里程碑的名称和计划日期
**So that** 我能修正或调整阶段计划信息

**Acceptance Criteria:**
- Given 里程碑已存在且我拥有 milestone:update 权限，When 我修改名称或计划日期，Then 变更立即保存并在时间线上反映
- Given 里程碑已存在且我拥有 milestone:update 权限，When 两个 PM 同时编辑同一里程碑并保存，Then 后保存者收到冲突提示"数据已被其他人修改，请刷新后重试"，不会静默覆盖

---

## Story 1c: PM 删除里程碑

**As a** PM（项目经理）
**I want to** 删除不再需要的里程碑
**So that** 我能清理误创建或废弃的阶段节点，保持里程碑列表整洁

**Acceptance Criteria:**
- Given 里程碑处于任意状态且我拥有 milestone:delete 权限，When 我删除该里程碑，Then 里程碑软删除，所有关联 MI 的 milestone_key 在同一事务内置空

---

## Story 1d: PM 切换里程碑状态

**As a** PM（项目经理）
**I want to** 切换里程碑状态（not_started/in_progress/completed/cancelled）
**So that** 我能跟踪各阶段的生命周期进展

**Acceptance Criteria:**
- Given 里程碑处于 not_started 或 in_progress 状态，When 我切换状态，Then 状态变更成功且时间线节点样式对应更新
- Given 里程碑处于 completed 状态，When 我尝试切换到 in_progress 状态，Then 操作被拒绝并提示"completed 状态不可回退到 in_progress"
- Given 里程碑处于 cancelled 状态，When 我尝试切换到任何其他状态，Then 操作被拒绝并提示"已取消的里程碑不可恢复"
- Given 里程碑处于 completed 或 in_progress 状态，When 我将其标记为 cancelled，Then 状态变为 cancelled，关联 MI 自动解绑

---

## Story 2: PM 将事项分配到里程碑

**As a** PM（项目经理）
**I want to** 通过时间线拖拽或主事项编辑页将 MainItem 绑定到里程碑、更改归属、或解绑
**So that** 我能灵活管理事项的阶段归属，保持阶段与事项的对齐

**Acceptance Criteria:**
- Given MainItem 未分配到任何里程碑，When 我在主事项编辑页选择一个里程碑，Then MI 的 milestone_key 设为该里程碑的 bizKey，该里程碑完成度更新
- Given MainItem 已属于里程碑 M1，When 我在时间线将 MI 拖至 M2，Then MI 的 milestone_key 更新为 M2 的 bizKey，M1 和 M2 的完成度均重新计算
- Given MainItem 已属于某个里程碑，When 我将其拖至时间线空白区域或在编辑页清空里程碑字段，Then MI 的 milestone_key 置空，原里程碑完成度重新计算
- Given MainItem 已属于某个里程碑，When 我在事项清单页筛选该里程碑，Then 列表只显示属于该里程碑的事项
- Given 我将 MainItem 拖拽至目标里程碑，When 后端 API 返回网络错误，Then 拖拽操作回滚，MI 停留在原里程碑位置，页面显示"操作失败，请重试"
- Given 团队下没有创建任何里程碑，When 我打开主事项编辑页的里程碑选择器，Then 下拉框仅显示"未分配"选项，无其他可选项

---

## Story 3: PM 通过时间线图查看阶段进展

**As a** PM（项目经理）
**I want to** 在时间线图页面查看所有里程碑及其关联事项，支持缩放（周/月/季）和拖拽交互
**So that** 我能一眼识别各阶段进展、延期风险，快速完成跨阶段进度盘点

**Acceptance Criteria:**
- Given 团队有 3+ 个里程碑，When 我访问 /milestones 页面，Then 时间线正确渲染所有里程碑节点，每个节点显示名称、计划日期、状态图标和完成度
- Given 时间线已渲染，When 我点击缩放控件切换周/月/季，Then 时间轴刻度标签对应变化，里程碑和事项位置重新排列
- Given 时间线已渲染，When 我点击某里程碑节点，Then 弹出详情面板显示该里程碑信息和关联 MI 列表
- Given 时间线已渲染，When 我悬停某里程碑节点，Then 显示关联事项数量和完成概况
- Given 团队有 >200 个 MI，When 我访问时间线页面，Then 每个里程碑下仅显示前 20 条 MI，底部显示"还有 N 条"按钮，点击后展开该里程碑下的全部 MI
- Given 团队有 0 个里程碑，When 我访问时间线页面，Then 显示空状态提示"暂无里程碑，点击创建"及创建按钮
- Given 团队有里程碑但后端 GET /milestones 返回 500 错误，When 我访问时间线页面，Then 显示"加载失败，请重试"提示和重试按钮，不显示空白页
- Given 里程碑处于 in_progress 状态但无关联 MI，When 我在时间线查看该里程碑，Then 完成度显示 0，节点样式仍为 in_progress（不因 0 关联 MI 降级为 not_started）

---

## Story 4: 团队成员查看里程碑进展

**As a** 团队成员
**I want to** 查看团队的里程碑时间线，了解当前阶段和交付节点
**So that** 我能清楚自己的工作处于哪个阶段，理解整体交付计划

**Acceptance Criteria:**
- Given 我拥有 milestone:read 权限，When 我访问 /milestones 页面，Then 我能看到完整的时间线视图（只读，无创建/编辑/删除按钮）
- Given 我没有 milestone:create 权限，When 我查看时间线，Then 创建按钮显示为禁用态（灰色+tooltip "无权限"）
- Given 我拥有 milestone:read 权限，When 时间线数据加载时后端 API 超时，Then 页面显示"加载失败"提示和重试按钮，不显示空白页
- Given 我没有 milestone:create 权限且团队有 0 个里程碑，When 我访问 /milestones 页面，Then 显示空状态提示"暂无里程碑"，不显示创建按钮
- Given 团队有已取消（cancelled）的里程碑，When 我查看时间线，Then 已取消的里程碑节点以灰色样式显示，无关联 MI，点击后详情面板显示"已取消"状态标签

---

## Story 5: 表格视图展示里程碑维度

**As a** PM（项目经理）
**I want to** 在表格视图中看到每个 MainItem 的里程碑归属列
**So that** 我能在表格中快速筛选和排序不同阶段的事项

**Acceptance Criteria:**
- Given 表格视图已加载，When 里程碑功能可用，Then 表格增加"里程碑"列，显示 MI 所属里程碑名称（未分配则显示"-"）
- Given 表格视图已有里程碑列，When 我按里程碑列筛选，Then 只显示对应里程碑的 MI
- Given 表格视图已有里程碑列，When 里程碑数据加载失败，Then 里程碑列显示"--"，其余列正常渲染，不阻塞表格加载
- Given MI 的 milestone_key 指向已被软删除的里程碑，When 我在表格视图查看该 MI，Then 里程碑列显示"--"（不显示已删除里程碑名称，不报错）
- Given 表格视图已有里程碑列，When 我按里程碑列排序选择降序，Then 已分配里程碑的 MI 按里程碑名称降序排列，未分配的 MI 排在末尾
- Given 表格视图已有里程碑列且未改变排序设置，When 页面加载，Then 已分配里程碑的 MI 按里程碑名称升序（默认）排列，未分配的 MI 排在末尾
- Given 表格视图已按里程碑筛选为 M1 且按里程碑列排序为降序，When 结果显示，Then 仅 M1 下的 MI 按名称降序排列，其他里程碑的 MI 不显示

---

## Story 6: 管理层查看项目里程碑进度概览

**As a** 管理层
**I want to** 通过时间线图以只读方式查看项目里程碑的整体进度和阶段分布
**So that** 我能快速掌握各项目的交付进展，无需深入事项细节

**Acceptance Criteria:**
- Given 我拥有 milestone:read 权限且无 create/update/delete 权限，When 我访问 /milestones 页面，Then 我能看到完整时间线视图，页面不显示创建/编辑/删除/状态切换按钮
- Given 团队有多个里程碑，When 我查看时间线，Then 每个里程碑节点显示名称、计划日期、状态和完成度，我能通过缩放切换周/月/季视图
- Given 我没有 milestone:read 权限，When 我访问 /milestones 页面，Then 页面返回 403 提示
- Given 团队有 0 个里程碑，When 我访问时间线页面，Then 显示空状态提示"暂无里程碑"，不显示创建按钮
- Given 我拥有 milestone:read 权限，When 后端 GET /milestones 超时或返回 500，Then 页面显示"加载失败，请重试"提示和重试按钮，不显示空白页
