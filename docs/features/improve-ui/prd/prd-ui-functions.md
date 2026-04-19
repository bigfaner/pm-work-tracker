---
feature: "improve-ui"
---

# Improve UI — UI Functions

> Requirements layer: defines WHAT the UI must do, based on the 13-page prototype.
> Only pages with significant changes from the original design include detailed specs.
> Unchanged pages reference the original PRD (`docs/features/pm-work-tracker/prd/prd-ui-functions.md`).

## UI Scope

13 页面，分为三类：

| 类别 | 页面 | 变化程度 |
|------|------|----------|
| 核心业务 | 登录、事项清单、主事项详情、子事项详情、事项池 | 事项清单有结构变化，其余视觉重做 |
| 数据视图 | 全量表格、甘特图、每周进展、周报导出 | 全量表格为新增独立页，每周进展为重点变更，其余视觉重做 |
| 管理 | 团队管理、团队详情、用户管理 | 用户管理和团队详情为拆分重构，团队管理视觉重做 |

---

## UI Function 1: 登录页（login）

### Description
用户通过账号密码登录系统。视觉风格重做，交互逻辑不变。

### User Interaction Flow
1. 输入账号 → 输入密码 → 点击"登录"
2. 验证通过 → 跳转事项清单
3. 验证失败 → 显示"账号或密码错误"，清空密码框

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 账号 | string | 用户输入 | 必填 |
| 密码 | string | 用户输入 | 必填，密码掩码 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 初始 | 登录按钮禁用 | 页面加载 |
| 可提交 | 登录按钮启用 | 账号和密码均非空 |
| 错误 | 红色错误提示文本 | 提交后验证失败 |
| 加载中 | 按钮显示加载态 | 提交中 |

### Validation Rules
- 两个字段均非空时才允许提交
- 错误提示不暴露具体是账号还是密码错误

---

## UI Function 2: 事项清单（main-items）⭐ 结构变更

### Description
系统默认着陆页，展示当前团队的所有活跃主事项。**新增 Summary/Detail 视图切换**。

### User Interaction Flow
1. 默认进入 Summary 卡片视图
2. 可切换到 Detail 表格视图
3. 两种视图共用筛选条件
4. Summary 视图：点击卡片展开查看子事项，无限滚动加载
5. Detail 视图：表格展示所有字段，分页浏览
6. 点击主事项标题 → 跳转主事项详情
7. 点击子事项标题 → 跳转子事项详情
8. 点击状态徽章 → 内联状态变更下拉
9. 点击"创建主事项" → 弹窗创建
10. 点击"创建子事项" → 弹窗创建

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 事项 ID | string | 后端 | 等宽字体 |
| 标题 | string | 后端 | 可点击链接 |
| 优先级 | string | 后端 | P1/P2/P3 徽章 |
| 负责人 | string | 后端 | 含头像 |
| 进度 | number | 后端 | 百分比，进度条 |
| 状态 | string | 后端 | 7 种状态徽章 |
| 开始日期 | date | 后端 | - |
| 预期完成 | date | 后端 | 逾期标红 |
| 实际完成 | date | 后端 | - |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Summary 视图 | 卡片列表，无限滚动 | 默认 / 点击 Summary |
| Detail 视图 | 数据表格，分页 | 点击 Detail |
| 空列表 | 空状态提示 | 无事项数据 |
| 加载中 | 骨架屏/加载动画 | 数据请求中 |

### Validation Rules
- 视图切换时保持当前筛选条件
- Summary 视图无限滚动：每次加载 5 条
- Detail 视图分页：默认 20 条/页
- 状态变更：内联下拉选择，即时生效

---

## UI Function 3: 主事项详情（mainitem-detail）

### Description
单个主事项的完整信息展示，含圆形进度、成果/卡点汇总、子事项表格。视觉重做，交互不变。

### User Interaction Flow
1. 面包屑导航：事项 > [主事项标题]
2. 查看基本信息（负责人、日期、描述）
3. 查看圆形进度指示器
4. 展开/折叠成果和卡点汇总
5. 查看子事项表格，可按状态/负责人筛选
6. 点击子事项标题 → 跳转子事项详情
7. 编辑主事项 / 归档

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 事项 ID | string | 后端 | 徽章 |
| 标题 | string | 后端 | - |
| 优先级 | string | 后端 | P1/P2/P3 |
| 状态 | string | 后端 | 状态徽章 |
| 负责人 | string | 后端 | 含头像 |
| 预期完成 | date | 后端 | 逾期标红 |
| 开始日期 | date | 后端 | - |
| 实际完成 | date | 后端 | - |
| 描述 | string | 后端 | - |
| 进度 | number | 后端 | 圆形进度 + 线性进度 |
| 成果汇总 | string[] | 后端 | 绿色文本 |
| 卡点汇总 | string[] | 后端 | 红色文本 |

### States
无新增状态。与原设计一致。

### Validation Rules
无新增规则。与原设计一致。

---

## UI Function 4: 子事项详情（subitem-detail）

### Description
单个子事项的完整信息展示，含进度条和进度时间线。视觉重做，交互不变。

### User Interaction Flow
1. 面包屑导航：事项 > [主事项标题] > [子事项标题]
2. 查看子事项信息（ID、父事项链接、优先级、负责人、状态等）
3. 查看进度条
4. 查看进度时间线（按日期倒序）
5. 点击"追加进度" → 弹窗填写百分比、成果、卡点

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 子事项 ID | string | 后端 | - |
| 父事项 | string | 后端 | 可点击链接 |
| 优先级 | string | 后端 | P1/P2/P3 |
| 负责人 | string | 后端 | 含头像 |
| 状态 | string | 后端 | 状态徽章 |
| 预期完成 | date | 后端 | - |
| 当前进度 | number | 后端 | 百分比，进度条 |
| 描述 | string | 后端 | - |
| 进度记录 | array | 后端 | 日期、百分比、成果、卡点 |

### States
无新增状态。与原设计一致。

### Validation Rules
- 追加进度：百分比 ≥ 上次记录值

---

## UI Function 5: 事项池（item-pool）

### Description
团队成员提交的待处理事项收件箱。PM 可转换为正式事项或拒绝。视觉重做，交互不变。

### User Interaction Flow
1. 查看待处理事项列表（无限滚动）
2. 筛选：按状态（待分配/已分配/已拒绝）
3. 搜索：按标题/ID
4. 转换为主事项 → 弹窗填写优先级、负责人、日期
5. 转换为子事项 → 弹窗选择父事项、填写详情
6. 拒绝 → 弹窗填写拒绝原因

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 池事项 ID | string | 后端 | POOL-xxx 格式 |
| 标题 | string | 后端 | - |
| 状态 | string | 后端 | 待分配/已分配/已拒绝 |
| 提交时间 | datetime | 后端 | - |
| 描述 | string | 后端 | - |
| 左边框色 | string | 前端计算 | 待分配蓝色，已分配灰色，已拒绝红色 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 待分配 | 蓝色左边框，完整操作按钮 | 默认 |
| 已分配 | 灰色半透明，显示关联的主事项 | 转换操作后 |
| 已拒绝 | 显示拒绝原因 | 拒绝操作后 |

### Validation Rules
- 转换为子事项时必须选择父事项
- 拒绝原因必填

---

## UI Function 6: 全量表格（table-view）⭐ 新增独立页

### Description
跨主/子事项的统一数据表格，支持多维筛选和 CSV 导出。原为事项清单的内嵌功能，现为独立页面。

### User Interaction Flow
1. 查看所有主/子事项的统一表格
2. 筛选：按标题、类型（main/sub）、优先级、负责人、状态
3. 分页：可选 5/10/20/50 条/页
4. 点击标题 → 跳转对应详情页
5. 点击"导出 CSV" → 下载当前筛选结果

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 类型 | string | 前端计算 | main/sub 徽章 |
| ID | string | 后端 | 等宽字体 |
| 标题 | string | 后端 | 可点击链接 |
| 优先级 | string | 后端 | P1/P2/P3 徽章 |
| 负责人 | string | 后端 | 含头像 |
| 进度 | number | 后端 | 百分比 |
| 状态 | string | 后端 | 状态徽章 |
| 开始日期 | date | 后端 | - |
| 预期完成 | date | 后端 | 逾期标红 |
| 实际完成 | date | 后端 | - |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 默认 | 完整表格，默认分页 | 页面加载 |
| 筛选中 | 按筛选条件过滤结果 | 输入筛选条件 |
| 空结果 | 空状态提示 | 无匹配数据 |

### Validation Rules
- 逾期日期（预期完成 < 今天且未完成）标红
- 标题链接根据类型跳转不同页面（main → mainitem-detail，sub → subitem-detail）

---

## UI Function 7: 甘特图（gantt-view）

### Description
时间线甘特图视图。视觉重做，交互不变。

### User Interaction Flow
1. 选择日期范围（默认：当前周 ±2 周）
2. 查看主事项时间条，可展开/折叠子事项
3. 今日标记线（蓝色竖线）
4. 搜索筛选事项
5. 加载更多事项

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 事项 ID | string | 后端 | - |
| 标题 | string | 后端 | 标签面板 |
| 开始日期 | date | 后端 | 条形起点 |
| 结束日期 | date | 后端 | 条形终点 |
| 完成百分比 | number | 后端 | 条形填充比例 |
| 状态 | string | 后端 | 条形颜色（进行中/已完成/逾期/无日期） |

### States
无新增状态。与原设计一致。

### Validation Rules
- 无日期的事项显示灰色虚线条
- 工作日浅蓝背景，周末灰色背景

---

## UI Function 8: 每周进展（weekly-view）⭐ 重点变更

### Description
上周/本周进度对比视图。**采用双列对比布局，新增进度增量标记（+N%、已完成、NEW），已完成的子事项默认折叠。**

### User Interaction Flow
1. 选择周（type="week" 输入框），显示对应日期范围
2. 查看统计概览（活跃子事项数、本周新完成、进行中、阻塞中）
3. 按主事项分组，每个主事项为一个双列对比卡片
4. 左列：上周子事项状态（徽章、优先级、标题、日期、进度描述）
5. 右列：本周子事项状态 + 进度增量标记（+N% 绿色、已完成 绿色、NEW 琥珀色）
6. 已完成且无变化的子事项默认折叠，可展开查看
7. 点击主事项标题跳转主事项详情

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 主事项标题 | string | 后端 | 可点击链接 |
| 优先级 | string | 后端 | P1/P2/P3 |
| 计划日期 | date range | 后端 | - |
| 子事项数 | number | 后端 | - |
| 整体进度 | number | 后端 | 进度条 + 百分比 |
| 上周子事项 | array | 后端 | 状态、优先级、标题、日期、进度描述 |
| 本周子事项 | array | 后端 | 状态、优先级、标题、日期、进度描述 + 增量标记 |
| 进度增量 | number | 后端计算 | 本周进度 - 上周进度，>0 时显示 +N% |
| 新增标记 | boolean | 后端计算 | 本周新增的子事项标记为 NEW |
| 完成标记 | boolean | 后端计算 | 本周新完成的子事项标记为"已完成" |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 有数据 | 双列对比卡片列表 | 默认 |
| 空数据 | 空状态提示 | 所选周无活跃事项 |
| 折叠态 | 已完成无变化子事项隐藏 | 默认 |
| 展开态 | 显示所有子事项 | 点击展开按钮 |

### Validation Rules
- 不允许选择未来周
- 已完成且本周无变化的子事项默认折叠
- 进度增量仅显示正值（>0），0 或负值不显示标记
- 主事项标题可点击跳转主事项详情

---

## UI Function 9: 周报导出（weekly-report）

### Description
生成并导出 Markdown 格式周报。视觉重做，交互不变。

### User Interaction Flow
1. 选择周
2. 点击"生成预览"
3. 查看预览（等宽字体 Markdown 格式）
4. 点击"导出 Markdown" → 下载 .md 文件

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 周编号 | string | 后端 | - |
| 主事项列表 | array | 后端 | 含标题、完成百分比、负责人 |
| 子事项列表 | array | 后端 | 含标题、状态、成果、卡点 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 初始 | 空预览区域 | 页面加载 |
| 预览中 | Markdown 渲染文本 | 点击生成预览 |
| 导出完成 | Toast 提示 | 导出成功 |

### Validation Rules
- 导出响应时间 <5 秒
- 必须先生成预览才能导出

---

## UI Function 10: 团队管理（team-management）

### Description
团队列表和创建。视觉重做，交互不变。

### User Interaction Flow
1. 查看所有团队列表
2. 点击团队名 → 跳转团队详情
3. 创建团队（弹窗）
4. 添加成员（弹窗）

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 团队名称 | string | 后端 | 可点击链接 |
| 描述 | string | 后端 | - |
| PM | string | 后端 | 含头像 |
| 成员数 | number | 后端 | - |
| 创建日期 | date | 后端 | - |

### States
无新增状态。与原设计一致。

### Validation Rules
无新增规则。

---

## UI Function 11: 团队详情（team-detail）⭐ 结构变更

### Description
单个团队的详情页，含成员管理和团队解散。**从超级管理员页面内嵌只读视图拆分为独立路由页，支持完整成员管理操作。**

### User Interaction Flow
1. 面包屑导航：团队管理 > [团队名]
2. 查看团队信息卡片（名称、PM、成员数、创建日期、描述）
3. 查看成员列表，可搜索和筛选
4. 添加成员（弹窗）
5. 设为 PM（行操作，需确认）
6. 移除成员（行操作，需确认）
7. 解散团队 → 弹窗输入团队名确认

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 团队名称 | string | 后端 | - |
| PM | string | 后端 | 含头像 |
| 成员数 | number | 后端 | - |
| 创建日期 | date | 后端 | - |
| 描述 | string | 后端 | - |
| 成员列表 | array | 后端 | 姓名、角色、加入日期 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 正常 | 完整信息展示 | 默认 |
| 确认解散 | 弹窗，需输入团队名 | 点击解散按钮 |
| 解散完成 | Toast 提示，跳转团队管理 | 确认解散 |

### Validation Rules
- 解散团队：输入框内容必须与团队名完全匹配才启用确认按钮
- PM 行无"设为 PM"和"移除"操作按钮
- 设为 PM 需二次确认

---

## UI Function 12: 用户管理（user-management）⭐ 结构变更

### Description
超级管理员管理所有用户。**从超级管理员页面的 Tab 拆分为独立导航页面。**

### User Interaction Flow
1. 查看所有用户列表
2. 筛选：按用户名/账号搜索，按创建团队权限筛选
3. 创建用户（弹窗）
4. 编辑用户（弹窗，预填当前值）
5. 变更状态（弹窗：启用 → 禁用，或禁用 → 启用）

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 姓名 | string | 后端 | 含头像 |
| 账号 | string | 后端 | 等宽字体 |
| 邮箱 | string | 后端 | - |
| 所属团队 | string[] | 后端 | 徽章列表 |
| 创建团队权限 | boolean | 后端 | 开关控件 |
| 账号状态 | string | 后端 | 已启用/已禁用 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 已启用 | 绿色徽章 | 默认 |
| 已禁用 | 灰色徽章 | 管理员操作后 |

### Validation Rules
- 账号唯一
- 邮箱格式校验
- 禁用用户时需二次确认（提示：禁用后无法登录，但数据保留）

---

## UI Function 13: 侧边栏导航

### Description
所有页面（除登录外）共享的左侧导航栏。

### User Interaction Flow
1. 团队选择器（下拉切换团队）
2. 导航链接（高亮当前页面）
3. 用户信息（头像、姓名、角色）
4. 登出按钮

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| 团队列表 | array | 后端 | 下拉选项 |
| 当前团队 | string | 后端 | 显示 |
| 导航项 | array | 前端固定 | 8 项 |
| 用户名 | string | 后端 | - |
| 用户角色 | string | 后端 | - |

### Navigation Items

| # | 名称 | 路由 |
|---|------|------|
| 1 | 事项清单 | main-items |
| 2 | 每周进展 | weekly-view |
| 3 | 整体进度 | gantt-view |
| 4 | 待办事项 | item-pool |
| 5 | 周报 | weekly-report |
| — | 分隔线 | — |
| 6 | 用户管理 | user-management |
| 7 | 团队管理 | team-management |
