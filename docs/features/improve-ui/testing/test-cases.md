---
feature: "improve-ui"
sources:
  - prd/prd-user-stories.md
  - prd/prd-spec.md
  - prd/prd-ui-functions.md
generated: "2026-04-19"
---

# Test Cases: improve-ui

## Summary

| Type | Count |
|------|-------|
| UI   | 52   |
| API  | 4  |
| CLI  | 0  |
| **Total** | **56** |

---

## UI Test Cases

## TC-001: 事项清单 Detail 视图切换
- **Source**: Story 1 / AC-1
- **Type**: UI
- **Pre-conditions**: 用户已登录，当前在事项清单页
- **Steps**:
  1. 确认当前为 Summary 卡片视图
  2. 点击 "Detail" 按钮
- **Expected**: 页面切换为数据表格视图，显示完整字段列（ID、标题、优先级、负责人、进度、状态、日期等），表格底部显示分页控件
- **Priority**: P0

## TC-002: 事项清单 Summary 视图切回
- **Source**: Story 1 / AC-2
- **Type**: UI
- **Pre-conditions**: 用户已登录，当前在事项清单 Detail 表格视图
- **Steps**:
  1. 点击 "Summary" 按钮
- **Expected**: 页面切换回卡片视图，以卡片列表展示主事项
- **Priority**: P0

## TC-003: 事项清单视图切换保留筛选条件
- **Source**: Story 1 / AC-3
- **Type**: UI
- **Pre-conditions**: 用户已登录，当前在事项清单页
- **Steps**:
  1. 在 Summary 视图设置筛选条件（如选择特定状态）
  2. 切换到 Detail 视图
  3. 检查筛选条件是否保留
  4. 切换回 Summary 视图
  5. 检查筛选条件是否保留
- **Expected**: 两种视图切换时，筛选条件保持不变，数据结果一致
- **Priority**: P0

## TC-004: 超管侧边栏进入用户管理页
- **Source**: Story 2 / AC-1
- **Type**: UI
- **Pre-conditions**: 用户为超级管理员，已登录
- **Steps**:
  1. 在侧边栏点击 "用户管理" 导航项
- **Expected**: 进入独立的用户管理页面，展示全部用户列表（含姓名、账号、邮箱、所属团队、权限、状态、操作列）
- **Priority**: P0

## TC-005: 用户管理页 CRUD 操作
- **Source**: Story 2 / AC-2
- **Type**: UI
- **Pre-conditions**: 用户为超级管理员，已在用户管理页
- **Steps**:
  1. 点击 "创建用户" → 弹窗填写姓名、账号、邮箱、团队、权限 → 提交 → 用户出现在列表
  2. 点击某用户行的 "编辑" → 弹窗预填当前值 → 修改并提交 → 列表更新
  3. 点击某用户行的 "变更状态" → 弹窗确认禁用 → 确认 → 状态徽章变为已禁用
  4. 再次变更状态 → 启用 → 状态徽章恢复已启用
- **Expected**: 创建、编辑、启用/禁用操作均可正常执行，列表实时更新
- **Priority**: P0

## TC-006: 团队管理点击团队名进入详情
- **Source**: Story 3 / AC-1
- **Type**: UI
- **Pre-conditions**: 用户已登录，当前在团队管理页，存在至少一个团队
- **Steps**:
  1. 点击某个团队名称
- **Expected**: 进入该团队的详情页，面包屑显示 "团队管理 > [团队名]"
- **Priority**: P0

## TC-007: 团队详情页展示信息和成员列表
- **Source**: Story 3 / AC-2
- **Type**: UI
- **Pre-conditions**: 用户已进入团队详情页
- **Steps**:
  1. 查看团队信息区域，确认展示团队名称、PM（含头像）、成员数、创建日期、描述
  2. 查看成员列表，确认展示姓名（含头像）、角色徽章（PM/成员）、加入日期、操作列
- **Expected**: 团队基本信息完整展示，成员列表正确显示所有成员及其角色
- **Priority**: P0

## TC-008: 团队详情页成员管理操作
- **Source**: Story 3 / AC-3
- **Type**: UI
- **Pre-conditions**: 用户已进入团队详情页
- **Steps**:
  1. 点击 "添加成员" → 弹窗选择用户 → 确认 → 成员出现在列表
  2. 点击某成员行的 "设为 PM" → 确认弹窗 → 确认 → 角色徽章更新为 PM
  3. 点击某成员行的 "移除" → 确认弹窗 → 确认 → 成员从列表消失
- **Expected**: 添加成员、设置 PM、移除成员操作均可正常执行
- **Priority**: P0

## TC-009: 团队详情页解散团队
- **Source**: Story 3 / AC-4
- **Type**: UI
- **Pre-conditions**: 用户已进入团队详情页
- **Steps**:
  1. 点击 "解散团队" 按钮
  2. 弹窗出现，输入框提示输入团队名
  3. 输入错误的团队名 → 确认按钮禁用
  4. 输入完全匹配的团队名 → 确认按钮启用
  5. 点击确认 → 团队解散，跳转回团队管理页
- **Expected**: 解散操作需输入完全匹配的团队名才可确认执行，执行后跳转团队管理页
- **Priority**: P0

## TC-010: 每周进展统计概览展示
- **Source**: Story 4 / AC-1
- **Type**: UI
- **Pre-conditions**: 用户已登录，当前在每周进展页，选择了一个有数据的周次
- **Steps**:
  1. 在周选择器中选择一个周次
  2. 查看页面顶部统计区域
- **Expected**: 顶部显示本周统计概览，包含：活跃子事项数、本周新完成数（绿色）、进行中数（蓝色）、阻塞中数（红色）
- **Priority**: P0

## TC-011: 每周进展双列对比布局
- **Source**: Story 4 / AC-2
- **Type**: UI
- **Pre-conditions**: 用户已登录，当前在每周进展页，选择了一个有数据的周次
- **Steps**:
  1. 查看每个主事项卡片布局
- **Expected**: 每个主事项为一个对比卡片，左列展示上周子事项状态（状态徽章、优先级、标题、日期、进度描述），右列展示本周子事项状态
- **Priority**: P0

## TC-012: 每周进展进度增量徽章
- **Source**: Story 4 / AC-3
- **Type**: UI
- **Pre-conditions**: 存在本周进度有增长的子事项
- **Steps**:
  1. 查看本周有进度增长的子事项
- **Expected**: 该子事项显示绿色增量徽章（如 "+30%"），数值为本周进度减上周进度的差值
- **Priority**: P0

## TC-013: 每周进展新完成标记
- **Source**: Story 4 / AC-4
- **Type**: UI
- **Pre-conditions**: 存在本周新完成的子事项
- **Steps**:
  1. 查看本周新完成的子事项
- **Expected**: 该子事项显示绿色 "已完成" 标记
- **Priority**: P0

## TC-014: 每周进展 NEW 标记
- **Source**: Story 4 / AC-5
- **Type**: UI
- **Pre-conditions**: 存在本周新增的子事项
- **Steps**:
  1. 查看本周新增的子事项
- **Expected**: 该子事项显示琥珀色 "NEW" 标记
- **Priority**: P0

## TC-015: 每周进展已完成无变化折叠
- **Source**: Story 4 / AC-6
- **Type**: UI
- **Pre-conditions**: 存在已完成且本周无变化的子事项
- **Steps**:
  1. 查看包含已完成且无变化子事项的主事项卡片
  2. 确认这些子事项默认不显示在展开区域
  3. 点击展开按钮
- **Expected**: 已完成且本周无变化的子事项默认折叠，点击展开按钮后可查看
- **Priority**: P1

## TC-016: 每周进展主事项标题跳转
- **Source**: Story 4 / AC-7
- **Type**: UI
- **Pre-conditions**: 用户已登录，当前在每周进展页
- **Steps**:
  1. 点击某主事项标题
- **Expected**: 跳转到该主事项的详情页
- **Priority**: P1

## TC-017: 全量表格展示主/子事项
- **Source**: Story 5 / AC-1
- **Type**: UI
- **Pre-conditions**: 用户已登录，当前在全量表格页，团队下有主事项和子事项数据
- **Steps**:
  1. 查看表格内容
- **Expected**: 表格展示所有主事项和子事项，"类型"列通过徽章区分 main/sub
- **Priority**: P0

## TC-018: 全量表格多维筛选
- **Source**: Story 5 / AC-2
- **Type**: UI
- **Pre-conditions**: 用户已登录，当前在全量表格页
- **Steps**:
  1. 在类型筛选中选择 "主事项" → 表格仅显示 main 类型
  2. 在优先级筛选中选择 "P1" → 表格仅显示 P1 事项
  3. 在负责人筛选中选择某个成员 → 表格仅显示该负责人的事项
  4. 在状态筛选中选择多个状态 → 表格按多选状态过滤
- **Expected**: 按类型、优先级、负责人、状态筛选均能正确过滤表格数据
- **Priority**: P0

## TC-019: 全量表格 CSV 导出
- **Source**: Story 5 / AC-3
- **Type**: UI
- **Pre-conditions**: 用户已登录，当前在全量表格页，表格中有数据
- **Steps**:
  1. 设置筛选条件（可选）
  2. 点击 "导出 CSV" 按钮
- **Expected**: 下载一个 CSV 文件，内容为当前筛选结果的全部数据
- **Priority**: P0

## TC-020: 全量表格标题跳转详情
- **Source**: Story 5 / AC-4
- **Type**: UI
- **Pre-conditions**: 用户已登录，当前在全量表格页
- **Steps**:
  1. 点击 main 类型事项的标题
  2. 返回全量表格页
  3. 点击 sub 类型事项的标题
- **Expected**: main 类型标题跳转到主事项详情页，sub 类型标题跳转到子事项详情页
- **Priority**: P0

## TC-021: 按钮统一变体规范
- **Source**: Story 6 / AC-1
- **Type**: UI
- **Pre-conditions**: 用户已登录，浏览各页面
- **Steps**:
  1. 检查所有页面的按钮样式
  2. 确认 primary 按钮风格一致
  3. 确认 secondary 按钮风格一致
  4. 确认 warning / danger / ghost 按钮风格一致
- **Expected**: 所有页面按钮遵循统一的变体规范（颜色、圆角、悬停态、禁用态一致）
- **Priority**: P1

## TC-022: 状态徽章颜色语义一致
- **Source**: Story 6 / AC-2
- **Type**: UI
- **Pre-conditions**: 用户已登录，浏览各页面
- **Steps**:
  1. 检查事项清单页、全量表格页、每周进展页等的状态徽章
  2. 比较相同状态在不同页面的颜色
- **Expected**: 相同状态（如 "进行中"、"已完成"、"阻塞"）在所有页面中徽章颜色一致
- **Priority**: P1

## TC-023: 表单控件风格统一
- **Source**: Story 6 / AC-3
- **Type**: UI
- **Pre-conditions**: 用户已登录，浏览含表单的页面
- **Steps**:
  1. 检查所有页面的输入框样式（边框、聚焦环、内边距）
  2. 检查所有页面的下拉框样式（下拉箭头、选项高亮）
  3. 检查所有页面的日期选择器样式
- **Expected**: 输入框、下拉框、日期选择器在所有页面中风格统一
- **Priority**: P1

## TC-024: 弹窗/卡片/表格样式一致
- **Source**: Story 6 / AC-4
- **Type**: UI
- **Pre-conditions**: 用户已登录，浏览各页面
- **Steps**:
  1. 打开不同页面的弹窗，比较尺寸、遮罩层、关闭按钮
  2. 比较不同页面的卡片样式（间距、圆角、阴影、边框）
  3. 比较不同页面的表格样式（表头背景、行悬停、斑马纹）
- **Expected**: 弹窗、卡片、表格在所有页面中保持一致的样式规范
- **Priority**: P1

## TC-025: 登录页按钮状态切换
- **Source**: UI Function 1 / States
- **Type**: UI
- **Pre-conditions**: 用户在登录页
- **Steps**:
  1. 页面初始加载 → 登录按钮应为禁用态
  2. 仅输入账号 → 登录按钮仍为禁用态
  3. 输入账号和密码 → 登录按钮变为启用态
  4. 点击登录 → 按钮显示加载态
- **Expected**: 登录按钮状态随输入变化正确切换：禁用 → 启用 → 加载
- **Priority**: P0

## TC-026: 登录页错误提示不暴露字段
- **Source**: UI Function 1 / Validation
- **Type**: UI
- **Pre-conditions**: 用户在登录页
- **Steps**:
  1. 输入错误账号和正确密码 → 点击登录
  2. 输入正确账号和错误密码 → 点击登录
- **Expected**: 两种情况均显示相同的错误提示（如 "账号或密码错误"），不暴露具体是账号还是密码错误
- **Priority**: P0

## TC-027: 事项清单默认 Summary 视图
- **Source**: UI Function 2 / States
- **Type**: UI
- **Pre-conditions**: 用户已登录
- **Steps**:
  1. 导航到事项清单页（或刷新页面）
- **Expected**: 默认进入 Summary 卡片视图，以卡片列表展示主事项
- **Priority**: P1

## TC-028: 事项清单 Summary 无限滚动
- **Source**: UI Function 2 / Validation
- **Type**: UI
- **Pre-conditions**: 团队下有超过 5 条主事项
- **Steps**:
  1. 在 Summary 视图滚动到底部
- **Expected**: 自动加载更多主事项（每次 5 条），当无更多数据时显示 "已全部加载" 提示
- **Priority**: P1

## TC-029: 事项清单空状态显示
- **Source**: UI Function 2 / States
- **Type**: UI
- **Pre-conditions**: 团队下无任何事项数据
- **Steps**:
  1. 导航到事项清单页
- **Expected**: 显示空状态提示（"暂无事项"），引导创建
- **Priority**: P1

## TC-030: 事项清单内联状态变更
- **Source**: UI Function 2 / Flow (Step 8)
- **Type**: UI
- **Pre-conditions**: 用户已登录，当前在事项清单页
- **Steps**:
  1. 点击某事项的状态徽章
  2. 从下拉列表中选择新状态
- **Expected**: 状态徽章即时更新为新状态，无需刷新页面
- **Priority**: P1

## TC-031: 子事项追加进度百分比校验
- **Source**: UI Function 4 / Validation
- **Type**: UI
- **Pre-conditions**: 用户在某子事项详情页，该子事项已有进度记录（如 60%）
- **Steps**:
  1. 点击 "追加进度"
  2. 在百分比输入框输入小于当前值（如 50%）
- **Expected**: 输入小于上次记录值时，表单校验不通过，提示百分比需 >= 上次记录值
- **Priority**: P1

## TC-032: 事项池状态颜色区分
- **Source**: UI Function 5 / States
- **Type**: UI
- **Pre-conditions**: 事项池中有待分配、已分配、已拒绝三种状态的事项
- **Steps**:
  1. 查看待分配事项 → 确认蓝色左边框
  2. 查看已分配事项 → 确认灰色半透明样式
  3. 查看已拒绝事项 → 确认显示拒绝原因
- **Expected**: 三种状态事项视觉区分正确：待分配蓝色左边框、已分配灰色半透明、已拒绝显示原因
- **Priority**: P1

## TC-033: 事项池转换子事项需选择父事项
- **Source**: UI Function 5 / Validation
- **Type**: UI
- **Pre-conditions**: 用户在事项池页
- **Steps**:
  1. 点击某待分配事项的 "转换为子事项"
  2. 不选择父事项直接提交
- **Expected**: 表单校验不通过，提示必须选择父事项
- **Priority**: P1

## TC-034: 事项池拒绝原因必填
- **Source**: UI Function 5 / Validation
- **Type**: UI
- **Pre-conditions**: 用户在事项池页
- **Steps**:
  1. 点击某待分配事项的 "拒绝"
  2. 不填写拒绝原因直接提交
- **Expected**: 表单校验不通过，提示拒绝原因必填
- **Priority**: P1

## TC-035: 全量表格逾期日期标红
- **Source**: UI Function 6 / Validation
- **Type**: UI
- **Pre-conditions**: 全量表格中有预期完成日期早于今天且未完成的事项
- **Steps**:
  1. 查看该事项的 "预期完成" 列
- **Expected**: 逾期日期以红色样式显示
- **Priority**: P1

## TC-036: 全量表格分页选择器
- **Source**: Spec 5.4 / 翻页设置
- **Type**: UI
- **Pre-conditions**: 用户已登录，当前在全量表格页
- **Steps**:
  1. 查看分页控件中的页大小选择器
  2. 分别选择 5、10、20、50 条/页
- **Expected**: 页大小选择器支持 5/10/20/50 选项，切换后表格显示对应条数
- **Priority**: P1

## TC-037: 甘特图今日标记线
- **Source**: UI Function 7 / Flow (Step 3)
- **Type**: UI
- **Pre-conditions**: 用户在甘特图页
- **Steps**:
  1. 查看甘特图时间轴
- **Expected**: 当天位置有一条蓝色竖线标记
- **Priority**: P2

## TC-038: 甘特图无日期灰色虚线
- **Source**: UI Function 7 / Validation
- **Type**: UI
- **Pre-conditions**: 存在未设置日期的事项
- **Steps**:
  1. 查看无日期事项在甘特图中的展示
- **Expected**: 无日期事项显示灰色虚线条
- **Priority**: P2

## TC-039: 每周进展不允许选择未来周
- **Source**: UI Function 8 / Validation
- **Type**: UI
- **Pre-conditions**: 用户在每周进展页
- **Steps**:
  1. 尝试选择一个未来的周次
- **Expected**: 未来周不可选（选择器禁用或选择后提示不可用）
- **Priority**: P1

## TC-040: 每周进展进度增量仅显示正值
- **Source**: UI Function 8 / Validation
- **Type**: UI
- **Pre-conditions**: 存在本周进度未增长（0 或负值）的子事项
- **Steps**:
  1. 查看本周进度无增长的子事项
- **Expected**: 该子事项不显示进度增量标记（0 或负值不显示）
- **Priority**: P1

## TC-041: 周报导出需先预览
- **Source**: UI Function 9 / Validation
- **Type**: UI
- **Pre-conditions**: 用户在周报导出页
- **Steps**:
  1. 页面初始加载，不点击 "生成预览"
  2. 直接查看 "导出 Markdown" 按钮状态
- **Expected**: 未生成预览时，"导出 Markdown" 按钮禁用或不可点击
- **Priority**: P2

## TC-042: 团队详情解散需匹配团队名
- **Source**: UI Function 11 / Validation
- **Type**: UI
- **Pre-conditions**: 用户在团队详情页
- **Steps**:
  1. 点击 "解散团队" 按钮 → 弹窗出现
  2. 输入与团队名不完全匹配的文本 → 确认按钮禁用
  3. 输入与团队名完全匹配的文本 → 确认按钮启用
- **Expected**: 仅输入内容与团队名完全匹配时确认按钮才启用
- **Priority**: P0

## TC-043: 团队详情 PM 行无操作按钮
- **Source**: UI Function 11 / Validation
- **Type**: UI
- **Pre-conditions**: 团队有 PM 成员
- **Steps**:
  1. 查看成员列表中 PM 角色的行
- **Expected**: PM 行不显示 "设为 PM" 和 "移除" 操作按钮
- **Priority**: P0

## TC-044: 团队详情设为 PM 二次确认
- **Source**: UI Function 11 / Validation
- **Type**: UI
- **Pre-conditions**: 用户在团队详情页，存在非 PM 成员
- **Steps**:
  1. 点击某成员行的 "设为 PM"
- **Expected**: 弹出确认弹窗，用户确认后才执行操作
- **Priority**: P1

## TC-045: 用户管理账号唯一性校验
- **Source**: UI Function 12 / Validation
- **Type**: UI
- **Pre-conditions**: 超管在用户管理页，系统中已存在账号 "testuser"
- **Steps**:
  1. 点击 "创建用户"
  2. 填写姓名、账号 "testuser"（已存在）、邮箱等信息
  3. 提交
- **Expected**: 表单校验不通过，提示 "账号已存在"
- **Priority**: P0

## TC-046: 用户管理邮箱格式校验
- **Source**: UI Function 12 / Validation
- **Type**: UI
- **Pre-conditions**: 超管在用户管理页
- **Steps**:
  1. 点击 "创建用户"
  2. 填写邮箱字段为非邮箱格式（如 "abc"）
  3. 提交
- **Expected**: 表单校验不通过，提示邮箱格式错误
- **Priority**: P1

## TC-047: 用户管理禁用二次确认
- **Source**: UI Function 12 / Validation
- **Type**: UI
- **Pre-conditions**: 超管在用户管理页
- **Steps**:
  1. 点击某已启用用户的 "变更状态" → 选择禁用
- **Expected**: 弹出确认弹窗，提示 "禁用后无法登录，但数据保留"，用户确认后才执行
- **Priority**: P1

## TC-048: 非超管隐藏用户管理入口
- **Source**: Spec 4.1 / Flow (P1 → 非超管)
- **Type**: UI
- **Pre-conditions**: 用户为普通 PM 或团队成员（非超级管理员）
- **Steps**:
  1. 查看侧边栏导航
- **Expected**: 侧边栏不显示 "用户管理" 导航项
- **Priority**: P0

## TC-049: 侧边栏导航高亮当前页
- **Source**: UI Function 13 / Flow (Step 2)
- **Type**: UI
- **Pre-conditions**: 用户已登录
- **Steps**:
  1. 导航到事项清单页 → 检查侧边栏 "事项清单" 是否高亮
  2. 导航到每周进展页 → 检查侧边栏 "每周进展" 是否高亮
  3. 导航到其他页面 → 检查对应导航项高亮
- **Expected**: 当前页面对应的侧边栏导航项始终高亮
- **Priority**: P2

## TC-050: 侧边栏团队选择器
- **Source**: UI Function 13 / Flow (Step 1)
- **Type**: UI
- **Pre-conditions**: 用户属于多个团队
- **Steps**:
  1. 点击团队选择器下拉
  2. 选择另一个团队
- **Expected**: 下拉显示用户所属团队列表，切换团队后页面数据更新为新团队数据
- **Priority**: P1

## TC-051: 创建主事项截止日期校验
- **Source**: Spec 5.3 / 弹窗操作
- **Type**: UI
- **Pre-conditions**: 用户在事项清单页
- **Steps**:
  1. 点击 "创建主事项"
  2. 设置开始日期为 2026-04-20，截止日期为 2026-04-15（早于开始日期）
  3. 提交
- **Expected**: 表单校验不通过，提示截止日期需 >= 开始日期
- **Priority**: P1

## TC-052: 页面兼容性 ≥1280px
- **Source**: Spec / 兼容性需求
- **Type**: UI
- **Pre-conditions**: 无
- **Steps**:
  1. 将浏览器窗口宽度设为 1280px
  2. 逐一浏览所有 13 个页面
  3. 在 Chrome、Edge、Firefox 最新版分别测试
- **Expected**: 所有页面在 1280px 宽度下正常显示，无布局错乱，关键操作可用
- **Priority**: P2

---

## API Test Cases

## TC-053: 团队详情独立路由成员 CRUD API
- **Source**: Spec 5.7 #1
- **Type**: API
- **Pre-conditions**: 已认证为 PM 或超级管理员
- **Steps**:
  1. GET /api/teams/:id → 返回团队详情和成员列表
  2. POST /api/teams/:id/members → 添加成员，返回 200
  3. PUT /api/teams/:id/members/:uid/role → 设为 PM，返回 200
  4. DELETE /api/teams/:id/members/:uid → 移除成员，返回 200
  5. DELETE /api/teams/:id → 解散团队，返回 200
- **Expected**: 团队详情独立端点支持完整的成员 CRUD 操作和团队解散
- **Priority**: P0

## TC-054: 用户管理全量操作 API
- **Source**: Spec 5.7 #2
- **Type**: API
- **Pre-conditions**: 已认证为超级管理员
- **Steps**:
  1. GET /api/users → 返回全量用户列表
  2. POST /api/users → 创建用户，账号重复返回 409
  3. PUT /api/users/:id → 编辑用户，返回 200
  4. PUT /api/users/:id/status → 变更状态（启用/禁用），返回 200
- **Expected**: 用户管理 API 支持独立页面所需的全部操作
- **Priority**: P0

## TC-055: 事项清单 Detail 分页参数
- **Source**: Spec 5.7 #3
- **Type**: API
- **Pre-conditions**: 已认证，团队下有超过 20 条主事项
- **Steps**:
  1. GET /api/items?page=1&pageSize=20 → 返回第一页 20 条
  2. GET /api/items?page=2&pageSize=20 → 返回第二页数据
  3. 验证响应包含分页元信息（total、page、pageSize）
- **Expected**: API 支持分页参数，返回正确的分页数据和元信息
- **Priority**: P1

## TC-056: 全量表格聚合查询 API
- **Source**: Spec 5.7 #4
- **Type**: API
- **Pre-conditions**: 已认证，团队下有主事项和子事项
- **Steps**:
  1. GET /api/items/all → 返回主/子事项统一列表，每条记录包含 type 字段（main/sub）
  2. GET /api/items/all?type=main → 仅返回主事项
  3. GET /api/items/all?type=sub → 仅返回子事项
  4. GET /api/items/all?type=main&priority=P1&status=in_progress → 多维筛选
- **Expected**: 聚合查询 API 支持跨主/子事项统一列表，type 字段筛选和多维过滤
- **Priority**: P0

---

## Traceability

| TC ID | Source | Type | Priority |
|-------|--------|------|----------|
| TC-001 | Story 1 / AC-1 | UI | P0 |
| TC-002 | Story 1 / AC-2 | UI | P0 |
| TC-003 | Story 1 / AC-3 | UI | P0 |
| TC-004 | Story 2 / AC-1 | UI | P0 |
| TC-005 | Story 2 / AC-2 | UI | P0 |
| TC-006 | Story 3 / AC-1 | UI | P0 |
| TC-007 | Story 3 / AC-2 | UI | P0 |
| TC-008 | Story 3 / AC-3 | UI | P0 |
| TC-009 | Story 3 / AC-4 | UI | P0 |
| TC-010 | Story 4 / AC-1 | UI | P0 |
| TC-011 | Story 4 / AC-2 | UI | P0 |
| TC-012 | Story 4 / AC-3 | UI | P0 |
| TC-013 | Story 4 / AC-4 | UI | P0 |
| TC-014 | Story 4 / AC-5 | UI | P0 |
| TC-015 | Story 4 / AC-6 | UI | P1 |
| TC-016 | Story 4 / AC-7 | UI | P1 |
| TC-017 | Story 5 / AC-1 | UI | P0 |
| TC-018 | Story 5 / AC-2 | UI | P0 |
| TC-019 | Story 5 / AC-3 | UI | P0 |
| TC-020 | Story 5 / AC-4 | UI | P0 |
| TC-021 | Story 6 / AC-1 | UI | P1 |
| TC-022 | Story 6 / AC-2 | UI | P1 |
| TC-023 | Story 6 / AC-3 | UI | P1 |
| TC-024 | Story 6 / AC-4 | UI | P1 |
| TC-025 | UI Function 1 / States | UI | P0 |
| TC-026 | UI Function 1 / Validation | UI | P0 |
| TC-027 | UI Function 2 / States | UI | P1 |
| TC-028 | UI Function 2 / Validation | UI | P1 |
| TC-029 | UI Function 2 / States | UI | P1 |
| TC-030 | UI Function 2 / Flow | UI | P1 |
| TC-031 | UI Function 4 / Validation | UI | P1 |
| TC-032 | UI Function 5 / States | UI | P1 |
| TC-033 | UI Function 5 / Validation | UI | P1 |
| TC-034 | UI Function 5 / Validation | UI | P1 |
| TC-035 | UI Function 6 / Validation | UI | P1 |
| TC-036 | Spec 5.4 / 翻页设置 | UI | P1 |
| TC-037 | UI Function 7 / Flow | UI | P2 |
| TC-038 | UI Function 7 / Validation | UI | P2 |
| TC-039 | UI Function 8 / Validation | UI | P1 |
| TC-040 | UI Function 8 / Validation | UI | P1 |
| TC-041 | UI Function 9 / Validation | UI | P2 |
| TC-042 | UI Function 11 / Validation | UI | P0 |
| TC-043 | UI Function 11 / Validation | UI | P0 |
| TC-044 | UI Function 11 / Validation | UI | P1 |
| TC-045 | UI Function 12 / Validation | UI | P0 |
| TC-046 | UI Function 12 / Validation | UI | P1 |
| TC-047 | UI Function 12 / Validation | UI | P1 |
| TC-048 | Spec 4.1 / Flow | UI | P0 |
| TC-049 | UI Function 13 / Flow | UI | P2 |
| TC-050 | UI Function 13 / Flow | UI | P1 |
| TC-051 | Spec 5.3 / 弹窗操作 | UI | P1 |
| TC-052 | Spec / 兼容性需求 | UI | P2 |
| TC-053 | Spec 5.7 #1 | API | P0 |
| TC-054 | Spec 5.7 #2 | API | P0 |
| TC-055 | Spec 5.7 #3 | API | P1 |
| TC-056 | Spec 5.7 #4 | API | P0 |
