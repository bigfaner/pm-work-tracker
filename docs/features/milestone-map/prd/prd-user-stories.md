---
feature: "里程碑图"
---

# User Stories: 里程碑图

## Story 1: PM 创建里程碑图

**As a** PM（项目经理）
**I want to** 在里程碑图列表页创建里程碑图（名称+负责人+计划时间+描述）
**So that** 我能按项目或交付阶段组织里程碑，管理多条独立的时间线

**Acceptance Criteria:**
- Given 我拥有 milestone:create 权限，When 我点击页面右上角"+ 创建里程碑图"按钮，Then 弹出创建弹窗，标题为"创建里程碑图"，表单为空
- Given 我点击列表网格中的虚线创建卡片，Then 同样弹出创建弹窗
- Given 团队有 0 个里程碑图（空状态页），When 我点击空状态页的创建按钮，Then 弹出创建弹窗
- Given 创建弹窗已打开，When 我填写名称和负责人并点击创建，Then 里程碑图创建成功，状态为"规划中"，弹窗关闭，列表刷新显示新卡片
- Given 我不填名称直接提交，Then 表单显示"名称不能为空"错误提示且不提交
- Given 我提交名称为 101 字符，Then 表单显示"名称不能超过 100 字符"错误提示且不提交；When 恰好 100 字符，Then 创建成功
- Given 我不填负责人直接提交，Then 表单显示"请选择负责人"错误提示且不提交
- Given 我填写了计划开始时间和计划完成时间，When 计划完成时间早于计划开始时间，Then 表单显示"计划完成时间不得早于计划开始时间"错误提示
- Given 我不填描述和计划时间，When 我填写名称和负责人并提交，Then 创建成功（描述和计划时间为空）
- Given 提交中，When 请求未返回，Then 创建按钮显示 loading 状态，所有输入框禁用，防止重复提交
- Given 提交时后端返回 500 错误，Then 页面显示"创建失败，请重试"错误提示，弹窗不关闭，表单保留已填写数据
- Given 创建弹窗已打开，When 我点击取消或 × 按钮，Then 弹窗关闭，不执行任何操作
- Given 描述字段，Then 文本域高度 160px，支持 resize: vertical 调整

---

## Story 2: PM 编辑里程碑图信息

**As a** PM（项目经理）
**I want to** 编辑里程碑图的名称、描述、负责人和计划时间
**So that** 我能修正或调整里程碑图的基本信息

**Acceptance Criteria:**
- Given 里程碑图详情页已加载，When 我拥有 milestone:update 权限并点击标题区的编辑按钮，Then 弹出编辑弹窗（标题"编辑里程碑图"），表单预填当前值
- Given 编辑弹窗已打开，When 我修改名称或描述并点击保存，Then 变更立即保存，弹窗关闭，详情页列表刷新
- Given 编辑弹窗已打开，When 我修改计划完成时间使其早于计划开始时间并保存，Then 表单显示日期校验错误且不提交
- Given 提交中，When 请求未返回，Then 保存按钮显示 loading 状态，所有输入框禁用
- Given 两个 PM 同时编辑同一里程碑图，When 后保存者提交，Then 收到冲突提示"数据已被其他人修改，请刷新后重试"，不会静默覆盖
- Given 编辑弹窗已打开，When 我未做任何修改直接点击保存，Then 等同取消操作，弹窗关闭
- Given 编辑弹窗已打开，When 我点击取消或 × 按钮，Then 弹窗关闭，不执行任何操作

---

## Story 3: PM 删除里程碑图

**As a** PM（项目经理）
**I want to** 删除不再需要的里程碑图
**So that** 我能清理误创建或废弃的里程碑图，保持列表整洁

**Acceptance Criteria:**
- Given 里程碑图处于"规划中"状态且我拥有 milestone:delete 权限，When 我在详情页点击删除按钮，Then 弹出确认弹窗（标题"确定删除里程碑图 [名称]？"，描述"关联的 X 个里程碑和 Y 个事项将解除绑定，数据不可恢复"，确认按钮为 danger 样式）
- Given 确认弹窗已打开，When 我点击确认，Then 里程碑图及其下所有里程碑级联软删除，所有关联 MI 的 milestone_key 在同一事务内置空，页面跳转回列表视图
- Given 确认弹窗已打开，When 我点击取消，Then 弹窗关闭，不做任何操作
- Given 里程碑图处于"已评审"/"待实施"/"实施中"/"已完成"状态，When 我查看详情页，Then 删除按钮不显示（仅 planning 状态可删除，BR-4）
- Given 我没有 milestone:delete 权限，When 我查看详情页，Then 删除按钮不显示

---

## Story 4: PM 切换里程碑图状态

**As a** PM（项目经理）
**I want to** 推进或回退里程碑图状态（规划中/已评审/待实施/实施中/已完成）
**So that** 我能跟踪项目整体的推进阶段

**Acceptance Criteria:**
- Given 里程碑图处于"规划中"，When 我点击状态 Badge，Then 下拉菜单仅显示"已评审"选项
- Given 里程碑图处于"已评审"，When 我点击状态 Badge，Then 下拉菜单显示"待实施"和"规划中"（回退）选项
- Given 里程碑图处于"待实施"，When 我点击状态 Badge，Then 下拉菜单显示"实施中"和"已评审"（回退）选项
- Given 里程碑图处于"实施中"，When 我点击状态 Badge，Then 下拉菜单显示"已完成"和"待实施"（回退）选项
- Given 里程碑图处于"已完成"，When 我点击状态 Badge，Then 下拉菜单无可用选项（终态不可回退）
- Given 里程碑图下存在未完成的里程碑（非 completed/cancelled），When 我尝试切换至"已完成"，Then 状态变更被拒绝，提示"所有里程碑必须已完成才能标记完成"（BR-2）
- Given 里程碑图下所有里程碑均为终态（completed 或 cancelled），When 我尝试切换至"已完成"，Then 状态变更成功
- Given 状态变更请求已发送，When 后端返回错误，Then 显示错误提示，Badge 恢复原状态
- Given 我在列表页使用状态筛选器选择"实施中"，Then 列表仅显示"实施中"状态的里程碑图卡片
- Given 我在列表页使用负责人筛选选择某成员，Then 列表仅显示该成员负责的里程碑图卡片
- Given 我在列表页使用名称搜索框输入关键字（debounce 300ms），Then 列表仅显示名称包含关键字的里程碑图卡片

---

## Story 5: PM 创建里程碑

**As a** PM（项目经理）
**I want to** 在某个里程碑图的时间线视图内创建里程碑（名称+计划完成时间+描述）
**So that** 我能定义该图的交付阶段节点

**Acceptance Criteria:**
- Given 我拥有 milestone:create 权限且已进入某里程碑图的时间线视图，When 我点击"+ 创建里程碑"按钮，Then 弹出创建弹窗（标题"创建里程碑"），表单为空
- Given 创建弹窗已打开，When 我填写名称和计划完成时间并点击确认，Then 里程碑创建成功，状态为 not_started，完成度为 0，弹窗关闭，时间线刷新显示新节点
- Given 我不填名称直接提交，Then 表单显示"名称不能为空"错误提示且不提交
- Given 我提交名称为 101 字符，Then 表单显示"名称不能超过 100 字符"错误提示；When 恰好 100 字符，Then 创建成功
- Given 我不填计划完成时间直接提交，Then 表单显示"计划完成时间不能为空"错误提示且不提交
- Given 我不填描述，When 我填写名称和计划完成时间并提交，Then 创建成功（描述为空）
- Given 提交中，When 请求未返回，Then 确认按钮显示 loading 状态，所有输入框禁用
- Given 提交时后端返回 500 错误，Then 页面显示"创建失败，请重试"，弹窗不关闭，表单保留已填写数据
- Given 创建弹窗已打开，When 我点击取消或 × 按钮，Then 弹窗关闭，不执行任何操作

---

## Story 6: PM 编辑里程碑信息

**As a** PM（项目经理）
**I want to** 编辑里程碑的名称、计划完成时间和描述
**So that** 我能修正或调整阶段计划信息

**Acceptance Criteria:**
- Given 里程碑详情面板已打开且我拥有 milestone:update 权限，When 我点击面板中的编辑按钮，Then 弹出编辑弹窗（标题"编辑里程碑"），表单预填当前值
- Given 编辑弹窗已打开，When 我修改名称、计划完成时间或描述并保存，Then 变更立即保存，弹窗关闭，面板和时间线刷新
- Given 编辑弹窗已打开，When 我修改计划完成时间，Then 时间线上该节点的位置按新日期重新计算
- Given 提交中，When 请求未返回，Then 保存按钮显示 loading 状态，所有输入框禁用
- Given 两个 PM 同时编辑同一里程碑，When 后保存者提交，Then 收到冲突提示"数据已被其他人修改，请刷新后重试"，不会静默覆盖
- Given 编辑弹窗已打开，When 我未做任何修改直接点击保存，Then 等同取消操作，弹窗关闭
- Given 编辑弹窗已打开，When 我点击取消或 × 按钮，Then 弹窗关闭，不执行任何操作

---

## Story 7: PM 删除里程碑

**As a** PM（项目经理）
**I want to** 删除不再需要的里程碑
**So that** 我能清理误创建或废弃的阶段节点，保持里程碑列表整洁

**Acceptance Criteria:**
- Given 里程碑处于 not_started 状态且我拥有 milestone:delete 权限，When 我在详情面板点击"删除里程碑"按钮，Then 弹出确认弹窗（标题"确定删除里程碑 [名称]？"，描述"关联的 X 个事项将解除绑定，里程碑数据不可恢复"，确认按钮为 danger 样式）
- Given 里程碑处于 cancelled 状态且我拥有 milestone:delete 权限，When 我在详情面板点击"删除里程碑"按钮并确认，Then 里程碑软删除成功
- Given 确认删除，When 后端执行删除，Then 里程碑软删除，关联 MI 的 milestone_key 在同一事务内置空，面板关闭，时间线刷新
- Given 确认弹窗已打开，When 我点击取消，Then 弹窗关闭，不做任何操作
- Given 里程碑处于 in_progress 或 completed 状态，When 我查看详情面板，Then 删除按钮不显示（仅 not_started 和 cancelled 可删除，BR-4）
- Given 我没有 milestone:delete 权限，When 我查看详情面板，Then 删除按钮不显示

---

## Story 8: PM 切换里程碑状态

**As a** PM（项目经理）
**I want to** 通过点击状态 Badge 切换里程碑状态（not_started/in_progress/completed/cancelled）
**So that** 我能跟踪各阶段的生命周期进展

**Acceptance Criteria:**
- Given 里程碑处于 not_started，When 我点击状态 Badge，Then 下拉菜单显示 in_progress 和 cancelled 选项
- Given 里程碑处于 in_progress，When 我点击状态 Badge，Then 下拉菜单显示 completed 和 cancelled 选项
- Given 里程碑处于 completed，When 我点击状态 Badge，Then 下拉菜单仅显示 cancelled 选项
- Given 里程碑处于 cancelled，When 我点击状态 Badge，Then 无下拉菜单弹出（终态不可恢复）
- Given 里程碑下存在未完成（非终态）的 MI，When 我尝试将里程碑标记为 completed，Then 状态变更被拒绝，提示"所有关联事项必须已完成才能标记完成"（BR-1）
- Given 里程碑下所有 MI 均已处于终态（completed/cancelled），When 我将里程碑标记为 completed，Then 状态变更成功
- Given 我将里程碑从 not_started 或 in_progress 标记为 cancelled，Then 所有关联 MI 自动解绑（milestone_key 事务内置空），里程碑详情面板关联事项列表为空
- Given 里程碑已处于 cancelled 状态，When MI 编辑页尝试将该 MI 绑定到此里程碑，Then 绑定被拒绝（BR-3：cancelled 里程碑不可接收新 MI）
- Given 里程碑处于 cancelled 状态，When 我查看详情面板，Then 面板全局灰色调，关联 MI 列表为空，删除按钮可见（BR-4 允许删除 cancelled 里程碑）
- Given 状态变更请求已发送，When 后端返回错误，Then 显示错误提示，Badge 恢复原状态，面板和时间线不刷新

---

## Story 9: PM 通过里程碑图查看阶段进展

**As a** PM（项目经理）
**I want to** 在里程碑图列表和时间线视图中查看所有里程碑及其关联事项，支持筛选、重置和缩放
**So that** 我能一眼识别各阶段进展、延期风险，快速完成跨阶段进度盘点

**Acceptance Criteria:**

**列表视图：**
- Given 我访问 /milestones 页面，When 数据加载中，Then 显示 3 个骨架屏卡片
- Given 团队有 3+ 个里程碑图，When 数据加载完成，Then 列表视图正确渲染所有卡片，每张显示：名称+状态 Badge（第一行）、里程碑数量+事项数量+负责人左右对齐（第二行）、计划时间跨度+整体进度条+百分比（第三行）、里程碑节点缩略图（底部）
- Given 我悬停某张卡片，Then 卡片边框高亮，出现阴影效果
- Given 我点击某张卡片，Then 路由跳转至 /milestones/:mapId 进入时间线视图
- Given 我使用状态筛选器选择"实施中"，Then 仅显示实施中状态的卡片
- Given 我使用负责人筛选选择某成员，Then 仅显示该成员负责的卡片
- Given 我使用名称搜索框输入关键字（debounce 300ms），Then 仅显示名称匹配的卡片
- Given 筛选栏已设置筛选条件，When 我点击重置按钮，Then 所有筛选条件恢复默认
- Given 我点击刷新按钮，Then 列表重新加载，刷新按钮显示 loading 状态
- Given 团队有 0 个里程碑图，When 我访问页面，Then 显示空状态提示"暂无里程碑图"及创建按钮
- Given 后端 API 返回 500 错误，When 我访问页面，Then 显示"加载失败，请重试"提示和重试按钮，不显示空白页

**时间线视图：**
- Given 我从列表页进入某里程碑图的时间线视图，When 数据加载中，Then 显示骨架屏
- Given 数据加载完成，Then 页面从上到下渲染：面包屑（可点击"里程碑图"返回列表）→ 详情标题区（名称+可点击状态 Badge+编辑/删除按钮）→ 基本信息卡片（负责人+计划开始时间+计划完成时间+整体进度，分隔线下方描述最多三行+Tooltip）→ 筛选栏（搜索+StatusTagFilter+重置+刷新+创建里程碑按钮+缩放控件）→ 横向时间线
- Given 详情标题区的描述文本超过三行，When 我鼠标悬浮，Then Tooltip 展示完整描述内容
- Given 详情标题区的删除按钮，When 里程碑图状态非 planning，Then 删除按钮不显示
- Given 时间线渲染完成，When 里程碑按计划完成时间等比排列在时间轴上，Then 每个节点卡片显示状态色点+名称+完成度+计划完成时间+关联 MI 数量
- Given 时间线节点，When 我悬停节点（非选中），Then 显示 Tooltip（"X 个事项，Y 已完成"），节点背景高亮
- Given 时间线节点已渲染，When 我使用名称搜索（debounce 300ms），Then 仅显示名称匹配的节点，不匹配的隐藏
- Given 时间线节点已渲染，When 我使用 StatusTagFilter 切换状态标签（可多选 toggle），Then 仅显示匹配状态的节点
- Given 筛选栏已设置筛选条件，When 我点击重置按钮，Then 所有筛选条件恢复默认
- Given 我点击刷新按钮，Then 时间线重新加载
- Given 我点击缩放控件切换周/月/季，Then 时间轴刻度标签对应变化（周=每 7 天、月=每 30 天、季=每 90 天），里程碑节点位置不变，transition 200ms
- Given 里程碑节点过密（相邻节点 x 坐标差 < 184px），When 容器宽度不足以容纳所有节点，Then 时间线区域自动出现水平滚动条，可横向滚动查看所有节点
- Given 时间线节点下方显示关联 MI 条目，When 我点击某条 MI，Then 跳转到 /items/:mainItemId 主事项详情页
- Given 我在时间线上拖拽某 MI 条目到另一个里程碑节点，Then MI 换绑到目标里程碑（调用 API 更新 milestone_key），拖拽中 MI 显示 opacity-50、目标里程碑高亮；完成时显示撤销 toast（5s），原里程碑和新里程碑完成度重新计算
- Given 时间线上有 0 个里程碑（空），When 时间线视图加载完成，Then 显示"暂无里程碑"空状态和创建按钮
- Given 时间线数据加载失败，Then 显示错误提示+重试按钮
- Given 我点击面包屑"里程碑图"，Then 返回列表视图（路由跳转至 /milestones）

---

## Story 10: PM 在里程碑详情面板管理关联事项

**As a** PM（项目经理）
**I want to** 在里程碑详情面板中查看和操作关联事项（解绑、快速添加）
**So that** 我能在里程碑上下文中高效管理事项归属

**Acceptance Criteria:**
- Given 我在时间线上点击某里程碑节点，When 详情面板从右侧滑入（translate-x 300ms），Then 面板显示：名称行（名称+关闭按钮）→ 描述区域（第一行"描述"标签+状态 Badge+编辑按钮，第二行描述文本最多 6 行+Tooltip）→ 计划完成时间 → 进度（标签+进度条+百分比）→ 关联 MI 列表 → 危险操作区（删除按钮）
- Given 面板打开时数据加载中，Then 面板内显示骨架屏
- Given 描述文本超过 6 行，When 鼠标悬浮在描述文本上，Then Tooltip 展示完整描述内容
- Given 面板已打开，When 我点击面板外 overlay 区域，Then 面板关闭（slide-out 动画），焦点回到触发的里程碑节点
- Given 面板已打开，When 我按 Escape 键，Then 面板关闭，焦点回到触发的里程碑节点
- Given 面板已打开，When 我点击 × 关闭按钮，Then 面板关闭
- Given 面板已打开且我拥有 milestone:update 权限，When 我查看描述区域，Then 编辑按钮可见
- Given 我没有 milestone:update 权限，When 我查看面板，Then 编辑按钮不显示，状态 Badge 不可点击
- Given 面板内关联 MI 列表已渲染，When 我悬停某条 MI，Then 行右侧显示 × 解绑按钮
- Given 我点击某 MI 行右侧的 × 按钮，When 解绑成功，Then 该 MI 从列表移除，显示撤销 toast（5s 内点击撤销可恢复绑定），原里程碑完成度重新计算
- Given 我点击"+ 添加"按钮，Then 弹出主事项创建弹窗（复用 CreateMainItemDialog），所属里程碑字段自动预填当前里程碑名称且 disabled 不可修改
- Given 快速添加弹窗已打开，When 我填写标题、负责人、开始时间、预期完成时间并确认，Then MI 创建成功并自动绑定当前里程碑，弹窗关闭，面板 MI 列表刷新
- Given 快速添加弹窗的所属里程碑字段，Then 显示当前里程碑名称，disabled 状态，不可修改
- Given 快速添加弹窗提交中，Then 确认按钮 loading 状态，所有输入框禁用
- Given 快速添加弹窗已打开，When 我不填标题直接提交，Then 表单显示标题必填错误且不提交
- Given 快速添加弹窗已打开，When 我不填负责人直接提交，Then 表单显示负责人必填错误且不提交
- Given 快速添加弹窗已打开，When 我不填开始时间或预期完成时间直接提交，Then 表单显示对应必填错误且不提交
- Given 面板内 MI 列表，When 我点击某条 MI 的编号或标题，Then 跳转到 /items/:mainItemId 主事项详情页
- Given 里程碑处于 cancelled 状态，When 我查看面板，Then 面板全局灰色调（text-tertiary），关联 MI 列表为空（MI 已在取消时自动解绑），"+ 添加"按钮不显示
- Given 面板中删除按钮（位于"危险操作"区域），When 里程碑状态为 in_progress 或 completed，Then 删除按钮不显示（仅 not_started 和 cancelled 可见，BR-4）

---

## Story 11: PM 将事项分配到里程碑

**As a** PM（项目经理）
**I want to** 通过主事项编辑弹窗将 MainItem 绑定到里程碑、更改归属或解绑
**So that** 我能灵活管理事项的阶段归属

**Acceptance Criteria:**
- Given 我打开主事项编辑弹窗，When 弹窗加载完成，Then 看到"所属里程碑"下拉选择框（位于"负责人"下方），显示当前值或"未分配"
- Given MainItem 未分配到任何里程碑，When 我在下拉框选择一个里程碑并保存，Then MI 的 milestone_key 设为该里程碑的 bizKey，该里程碑完成度重新计算
- Given MainItem 已属于某个里程碑，When 我在下拉框选择"未分配"并保存，Then MI 的 milestone_key 置空，原里程碑完成度重新计算
- Given MainItem 已属于某个里程碑，When 我在下拉框选择另一个里程碑并保存，Then MI 换绑到新里程碑（自动解绑旧里程碑），两个里程碑完成度重新计算
- Given MainItem 已属于某个里程碑，When 我不做修改直接保存，Then 保持原值不变
- Given 团队下没有创建任何里程碑，When 我打开下拉框，Then 仅显示"未分配"选项
- Given 下拉框选项列表，When 里程碑处于 cancelled 状态，Then 该里程碑不在下拉选项中出现
- Given MI 处于终态（completed/closed），When 我尝试变更其 milestone_key 并保存，Then 操作被拒绝（BR-3：终态 MI 不可变更里程碑归属）
- Given MI 与目标里程碑属于不同团队，When 我尝试绑定，Then 操作被拒绝，提示"事项与里程碑不属于同一团队"

---

## Story 12: PM 在事项清单按里程碑筛选

**As a** PM（项目经理）
**I want to** 在事项清单页按里程碑筛选并查看事项的里程碑标签
**So that** 我能快速找到属于特定阶段的所有事项

**Acceptance Criteria:**
- Given 事项清单页已加载，When 里程碑功能可用，Then 筛选栏在现有"负责人"筛选器右侧显示"里程碑"下拉筛选器，默认值"全部"
- Given 下拉选项已加载，When 我展开下拉框，Then 显示"全部"+"未分配"+团队内所有非 cancelled 里程碑名称
- Given 下拉选项加载，When 里程碑处于 cancelled 状态，Then 该里程碑不在选项中出现
- Given 我选择某个里程碑，When 筛选生效，Then 列表仅显示属于该里程碑的 MI
- Given 我选择"未分配"，When 筛选生效，Then 列表仅显示 milestone_key 为空的 MI
- Given 我选择"全部"（默认），When 筛选生效，Then 显示所有 MI
- Given 筛选值传入非法 bizKey，When 筛选触发，Then 回退到"全部"，不产生错误
- Given 每条 MI 记录已渲染且已分配到里程碑，When MI 条目显示，Then 旁边显示里程碑名称 Badge（rounded-full，text-xs）
- Given 每条 MI 记录已渲染且未分配里程碑，When MI 条目显示，Then 不显示里程碑 Badge
- Given 我切换团队，When 新团队加载完成，Then 里程碑筛选器重置为"全部"，下拉选项刷新为新团队的里程碑列表
- Given 里程碑下拉选项加载失败，When 我展开下拉框，Then 显示"加载失败"且下拉框禁用，其余筛选器不受影响

---

## Story 13: PM 在表格视图查看里程碑列

**As a** PM（项目经理）
**I want to** 在表格视图中看到每个 MainItem 的里程碑归属列
**So that** 我能在表格中快速筛选和排序不同阶段的事项

**Acceptance Criteria:**
- Given 表格视图已加载，When 里程碑功能可用，Then 表格增加"里程碑"列（位于"标题"列和"优先级"列之间，列宽 w-32），已分配显示里程碑名称（text-secondary），未分配显示"-"（text-tertiary）
- Given 里程碑列已渲染，When 我点击列头排序图标选择升序（默认），Then 已分配 MI 按里程碑名称字母序升序排列，未分配 MI 排末尾
- Given 里程碑列已渲染，When 我点击列头排序图标选择降序，Then 已分配 MI 按里程碑名称字母序降序排列，未分配 MI 排末尾
- Given 里程碑列已渲染，When 我使用列头筛选选择某里程碑，Then 仅显示对应里程碑的 MI
- Given 筛选值传入非法 bizKey，When 筛选触发，Then 回退到"全部"
- Given MI 的 milestone_key 指向已被软删除的里程碑，When 我查看该 MI 的里程碑列，Then 显示"--"
- Given 里程碑数据加载失败，When 表格其余列正常渲染，Then 里程碑列显示"--"，不阻塞表格加载
- Given 表格视图未改变排序设置，When 页面首次加载，Then 里程碑列默认升序，未分配排末尾

---

## Story 14: 不同角色查看里程碑进展

**As a** 团队成员或管理层
**I want to** 以只读方式查看团队的里程碑图列表和时间线
**So that** 我能了解当前阶段和交付节点，或快速掌握各项目的交付进展

**Acceptance Criteria:**
- Given 我拥有 milestone:read 权限但无 create/update/delete 权限，When 我访问 /milestones 列表页，Then 我能看到完整的卡片列表和筛选功能，页面不显示"+ 创建里程碑图"按钮和虚线创建卡片
- Given 我拥有 milestone:read 但无 milestone:create，When 列表页有 0 个里程碑图，Then 显示空状态提示"暂无里程碑图"，不显示创建按钮
- Given 我拥有 milestone:read 但无 create/update/delete 权限，When 我进入时间线视图，Then 详情标题区不显示编辑和删除按钮，"+ 创建里程碑"按钮不显示或显示为禁用态
- Given 我拥有 milestone:read 但无 create/update/delete 权限，When 我点击里程碑节点打开详情面板，Then 面板不显示编辑按钮、删除按钮和"+ 添加"按钮，状态 Badge 不可点击，MI 行不显示 × 解绑按钮
- Given 我没有 milestone:read 权限，When 我访问 /milestones 页面，Then 页面返回 403 提示
- Given 我拥有 milestone:read 权限，When 后端 API 超时或返回 500，Then 页面显示"加载失败，请重试"提示和重试按钮，不显示空白页
- Given 我拥有 milestone:read 权限，When 我在时间线视图查看基本信息卡片和里程碑节点，Then 所有只读信息正常展示（名称、状态、进度、描述、关联 MI 列表），hover 交互（Tooltip、卡片高亮）正常工作
- Given 我拥有 milestone:read 权限，When 我在时间线上点击某条 MI，Then 正常跳转到 /items/:mainItemId 主事项详情页
