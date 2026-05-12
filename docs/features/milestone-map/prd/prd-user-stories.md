---
feature: "里程碑图"
---

# User Stories: 里程碑图

## Story 1: PM 创建里程碑图

**As a** PM（项目经理）
**I want to** 创建里程碑图（名称+描述）
**So that** 我能按项目或交付阶段组织里程碑，管理多条独立的时间线

**Acceptance Criteria:**
- Given 我拥有 milestone:create 权限，When 我在里程碑图列表页点击创建并填写名称+描述，Then 里程碑图创建成功，状态为"规划中"
- Given 我拥有 milestone:create 权限，When 我提交名称恰好 100 字符的里程碑图，Then 创建成功；When 我提交名称为 101 字符，Then 表单显示"名称不能超过 100 字符"错误提示且不提交
- Given 我拥有 milestone:create 权限，When 我不填名称直接提交，Then 表单显示"名称不能为空"错误提示且不提交

---

## Story 2: PM 编辑里程碑图信息

**As a** PM（项目经理）
**I want to** 编辑里程碑图的名称和描述
**So that** 我能修正或调整里程碑图的基本信息

**Acceptance Criteria:**
- Given 里程碑图已存在且我拥有 milestone:update 权限，When 我修改名称或描述，Then 变更立即保存并在列表中反映

---

## Story 3: PM 切换里程碑图状态

**As a** PM（项目经理）
**I want to** 推进里程碑图状态（规划中/已评审/待实施/实施中/已完成）
**So that** 我能跟踪项目整体的推进阶段

**Acceptance Criteria:**
- Given 里程碑图处于"规划中"状态，When 我点击状态 Badge，Then 下拉菜单仅显示"已评审"选项
- Given 里程碑图处于"实施中"状态，When 我点击状态 Badge，Then 下拉菜单显示"待实施"和"已完成"选项
- Given 里程碑图处于"已完成"状态，When 我点击状态 Badge，Then 下拉菜单无可用选项，显示"暂无可用流转"
- Given 我在里程碑图列表页，When 我使用状态筛选器选择"实施中"，Then 列表仅显示"实施中"状态的里程碑图卡片

---

## Story 4a: PM 创建里程碑

**As a** PM（项目经理）
**I want to** 在某个里程碑图内创建里程碑（名称+计划完成时间）
**So that** 我能定义该图的交付阶段节点

**Acceptance Criteria:**
- Given 我拥有 milestone:create 权限且已进入某里程碑图的时间线视图，When 我点击创建并填写名称+计划完成时间，Then 里程碑创建成功，状态为 not_started，完成度为 0
- Given 我拥有 milestone:create 权限，When 我提交名称恰好 100 字符的里程碑，Then 创建成功；When 我提交名称为 101 字符，Then 表单显示"名称不能超过 100 字符"错误提示且不提交
- Given 我拥有 milestone:create 权限，When 我不填名称直接提交，Then 表单显示"名称不能为空"错误提示且不提交
- Given 我拥有 milestone:create 权限，When 我填写名称后提交但后端返回 500 错误，Then 页面显示"创建失败，请重试"，表单保留已填写数据不丢失

---

## Story 4b: PM 编辑里程碑信息

**As a** PM（项目经理）
**I want to** 编辑里程碑的名称和计划完成时间
**So that** 我能修正或调整阶段计划信息

**Acceptance Criteria:**
- Given 里程碑已存在且我拥有 milestone:update 权限，When 我修改名称或计划完成时间，Then 变更立即保存并在时间线上反映
- Given 里程碑已存在且我拥有 milestone:update 权限，When 两个 PM 同时编辑同一里程碑并保存，Then 后保存者收到冲突提示"数据已被其他人修改，请刷新后重试"，不会静默覆盖

---

## Story 4c: PM 删除里程碑

**As a** PM（项目经理）
**I want to** 删除不再需要的里程碑
**So that** 我能清理误创建或废弃的阶段节点，保持里程碑列表整洁

**Acceptance Criteria:**
- Given 里程碑处于任意状态且我拥有 milestone:delete 权限，When 我删除该里程碑，Then 里程碑软删除，所有关联 MI 的 milestone_key 在同一事务内置空

---

## Story 5: PM 切换里程碑状态

**As a** PM（项目经理）
**I want to** 通过点击状态 Badge 切换里程碑状态（not_started/in_progress/completed/cancelled）
**So that** 我能跟踪各阶段的生命周期进展

**Acceptance Criteria:**
- Given 里程碑处于 not_started 或 in_progress 状态，When 我点击状态 Badge 并选择目标状态，Then 状态变更成功且时间线节点样式对应更新
- Given 里程碑处于 completed 状态，When 我点击状态 Badge，Then 下拉菜单仅显示"已取消"选项，无 in_progress 回退选项
- Given 里程碑处于 cancelled 状态，When 我点击状态 Badge，Then 下拉菜单无可用选项，显示"暂无可用流转"
- Given 里程碑处于 completed 或 in_progress 状态，When 我将其标记为 cancelled，Then 状态变为 cancelled，关联 MI 自动解绑

---

## Story 6: PM 将事项分配到里程碑

**As a** PM（项目经理）
**I want to** 通过主事项编辑页将 MainItem 绑定到里程碑、更改归属、或解绑
**So that** 我能灵活管理事项的阶段归属，保持阶段与事项的对齐

**Acceptance Criteria:**
- Given MainItem 未分配到任何里程碑，When 我在主事项编辑页选择一个里程碑，Then MI 的 milestone_key 设为该里程碑的 bizKey，该里程碑完成度更新
- Given MainItem 已属于某个里程碑，When 我在编辑页清空里程碑字段，Then MI 的 milestone_key 置空，原里程碑完成度重新计算
- Given MainItem 已属于某个里程碑，When 我在事项清单页筛选该里程碑，Then 列表只显示属于该里程碑的事项
- Given 团队下没有创建任何里程碑，When 我打开主事项编辑页的里程碑选择器，Then 下拉框仅显示"未分配"选项，无其他可选项

---

## Story 7: PM 在里程碑详情面板管理关联事项

**As a** PM（项目经理）
**I want to** 在里程碑详情面板中查看关联事项、快速添加事项、解绑事项
**So that** 我能在里程碑上下文中高效管理事项归属

**Acceptance Criteria:**
- Given 里程碑详情面板已打开，When 我点击某 MI 行右侧的 × 按钮，Then 该 MI 与里程碑解绑，MI 列表刷新，显示撤销 toast
- Given 里程碑详情面板已打开，When 我点击"+ 添加"按钮，Then 弹出主事项创建表单，所属里程碑自动预填为当前里程碑且不可修改
- Given 快速添加表单已打开，When 我填写标题、负责人、开始时间、预期完成时间并确认，Then MI 创建成功并自动绑定当前里程碑，详情面板 MI 列表刷新
- Given 快速添加表单已打开，When 所属里程碑字段显示，Then 字段为 disabled 状态，显示当前里程碑名称，不可修改

---

## Story 8: PM 通过里程碑图查看阶段进展

**As a** PM（项目经理）
**I want to** 在里程碑图列表和时间线视图中查看所有里程碑及其关联事项，支持状态筛选和缩放
**So that** 我能一眼识别各阶段进展、延期风险，快速完成跨阶段进度盘点

**Acceptance Criteria:**
- Given 团队有 3+ 个里程碑图，When 我访问 /milestones 页面，Then 列表视图正确渲染所有里程碑图卡片，每张卡片显示名称、状态、里程碑数量、事项数量、整体进度
- Given 列表视图已渲染，When 我按状态筛选选择"实施中"，Then 仅显示实施中状态的里程碑图卡片
- Given 我点击某张里程碑图卡片，When 进入时间线视图，Then 时间线正确渲染该图的所有里程碑节点，每个节点显示名称、计划完成时间、状态和完成度
- Given 时间线已渲染，When 我点击缩放控件切换周/月/季，Then 时间轴刻度标签对应变化，里程碑和事项位置重新排列
- Given 团队有 0 个里程碑图，When 我访问 /milestones 页面，Then 显示空状态提示"暂无里程碑图"及创建按钮
- Given 团队有里程碑图但后端 API 返回 500 错误，When 我访问 /milestones 页面，Then 显示"加载失败，请重试"提示和重试按钮，不显示空白页

---

## Story 9: 团队成员查看里程碑进展

**As a** 团队成员
**I want to** 查看团队的里程碑图列表和时间线，了解当前阶段和交付节点
**So that** 我能清楚自己的工作处于哪个阶段，理解整体交付计划

**Acceptance Criteria:**
- Given 我拥有 milestone:read 权限，When 我访问 /milestones 页面，Then 我能看到完整的列表视图和时间线视图（只读，无创建/编辑/删除按钮）
- Given 我没有 milestone:create 权限，When 我查看列表视图，Then 创建按钮显示为禁用态（灰色+tooltip "无权限"）
- Given 我没有 milestone:create 权限且团队有 0 个里程碑图，When 我访问 /milestones 页面，Then 显示空状态提示"暂无里程碑图"，不显示创建按钮

---

## Story 10: 表格视图展示里程碑维度

**As a** PM（项目经理）
**I want to** 在表格视图中看到每个 MainItem 的里程碑归属列
**So that** 我能在表格中快速筛选和排序不同阶段的事项

**Acceptance Criteria:**
- Given 表格视图已加载，When 里程碑功能可用，Then 表格增加"里程碑"列，显示 MI 所属里程碑名称（未分配则显示"-"）
- Given 表格视图已有里程碑列，When 我按里程碑列筛选，Then 只显示对应里程碑的 MI
- Given 表格视图已有里程碑列，When 里程碑数据加载失败，Then 里程碑列显示"--"，其余列正常渲染，不阻塞表格加载
- Given MI 的 milestone_key 指向已被软删除的里程碑，When 我在表格视图查看该 MI，Then 里程碑列显示"--"
- Given 表格视图已有里程碑列，When 我按里程碑列排序选择降序，Then 已分配里程碑的 MI 按里程碑名称降序排列，未分配的 MI 排在末尾
- Given 表格视图已有里程碑列且未改变排序设置，When 页面加载，Then 已分配里程碑的 MI 按里程碑名称升序（默认）排列，未分配的 MI 排在末尾

---

## Story 11: 管理层查看项目里程碑进度概览

**As a** 管理层
**I want to** 通过里程碑图列表以只读方式查看各项目的整体进度和阶段分布
**So that** 我能快速掌握各项目的交付进展，无需深入事项细节

**Acceptance Criteria:**
- Given 我拥有 milestone:read 权限且无 create/update/delete 权限，When 我访问 /milestones 页面，Then 我能看到完整列表视图，页面不显示创建/编辑/删除按钮
- Given 我没有 milestone:read 权限，When 我访问 /milestones 页面，Then 页面返回 403 提示
- Given 团队有 0 个里程碑图，When 我访问 /milestones 页面，Then 显示空状态提示"暂无里程碑图"，不显示创建按钮
- Given 我拥有 milestone:read 权限，When 后端 API 超时或返回 500，Then 页面显示"加载失败，请重试"提示和重试按钮，不显示空白页
