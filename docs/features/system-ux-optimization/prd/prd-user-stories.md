---
feature: "System UX Optimization Batch"
---

# User Stories: System UX Optimization Batch

## Story 1: 状态流转错误提示

**As a** PM 用户
**I want to** 在状态流转失败时看到具体的错误原因（而非2秒消失的 tooltip）
**So that** 我能理解失败原因并采取正确的后续操作

**Acceptance Criteria:**
- Given 用户点击了状态流转按钮且后端返回不可流转原因
- When 前端接收到错误响应
- Then 在操作区域下方展示行内错误消息（Alert 组件），消息内容为后端返回的具体原因

---

## Story 2: 子事项开始时间编辑

**As a** PM 用户
**I want to** 在子事项编辑弹窗中直接修改开始时间
**So that** 无需绕道详情页即可完成子事项时间调整

**Acceptance Criteria:**
- Given 用户打开子事项编辑弹窗
- When 弹窗渲染完成
- Then 弹窗中包含"开始时间"字段，修改后可成功保存

---

## Story 3: 删除事项

**As a** PM 用户
**I want to** 删除误建的主事项和子事项
**So that** 无需联系管理员即可自行清理错误数据

**Acceptance Criteria:**
- Given PM 用户在主事项界面点击删除按钮
- When 系统弹出确认对话框（提示"将同时删除 N 个子事项"）
- And 用户确认删除
- Then 主事项及其所有子事项被软删除，status_history 记录删除事件

- Given PM 用户在子事项详情界面点击删除按钮
- When 系统弹出确认对话框
- And 用户确认删除
- Then 子事项被软删除

- Given 非 PM 角色用户查看事项
- When 页面渲染完成
- Then 不显示删除按钮

---

## Story 4: 转换表单交互优化

**As a** PM 用户
**I want to** 转换表单有合理的默认值和必填校验
**So that** 减少因遗漏必填项导致的提交失败，以及避免误修改描述字段

**Acceptance Criteria:**
- Given 用户打开待办→子事项转换表单
- When 表单渲染完成
- Then 描述字段为 disabled 灰色样式，不可编辑；开始时间默认为当天

- Given 用户打开待办→主事项或子事项转换表单
- When 未选择负责人或未选择优先级
- Then 提交按钮禁用，两字段标签显示必填标记

- Given 用户关闭任何新增/转换表单（含取消）或提交成功后
- When 再次打开任意新增/转换表单
- Then 所有字段为空（无残留数据），包括之前表单中已填入的内容

---

## Story 5: 子事项列表排序

**As a** PM 用户
**I want to** 主事项详情页的子事项列表按最新创建的排在前面
**So that** 无需手动滚动寻找最新添加的子事项

**Acceptance Criteria:**
- Given 主事项下有多个子事项
- When 用户查看主事项详情页的子事项列表
- Then 子事项按创建时间倒序排列（最新创建的显示在列表最前）

---

## Story 6: Member 角色权限修复

**As a** member 角色用户
**I want to** 登录后看到我权限范围内的菜单和功能
**So that** 能正常使用事项查看和待办提交功能

**Acceptance Criteria:**
- Given member 角色用户成功登录
- When 页面加载完成
- Then 用户能看到至少：待办事项提交(item_pool:submit)、事项查看(main_item:list)

- Given 管理员查看团队成员列表
- When 所有团队成员的角色均已正确分配
- Then 不存在无角色（role_key 为空）的团队成员记录

---

## Story 7: 移动子事项

**As a** PM 用户
**I want to** 将子事项从一个主事项移动到另一个主事项
**So that** 无需手动删除重建即可重组子事项归属

**Acceptance Criteria:**
- Given 用户在子事项详情选择"移动到其他主事项"
- When 选择有效目标主事项并确认
- Then 子事项移动到目标主事项下，编号自动更新，状态和负责人不变

- Given 用户尝试移动子事项到已关闭状态的主事项
- When 提交移动操作
- Then 操作被拒绝并展示提示

- Given 用户尝试移动子事项到同一主事项
- When 提交移动操作
- Then 操作被拒绝

---

## Story 8: 过滤穿透

**As a** PM 用户
**I want to** 通过负责人过滤时看到该负责人负责的子事项所属的主事项
**So that** 不遗漏与该负责人相关的事项

**Acceptance Criteria:**
- Given 用户在事项清单页面选中负责人 A 进行过滤
- When 过滤结果返回
- Then 展示 A 负责的主事项 + 含 A 负责子事项的主事项（连同该子事项一起展示）

- Given 因子事项匹配而展示的主事项
- When 用户查看卡片/行
- Then 卡片/行上有"因子事项匹配"视觉标识

- Given 用户未选择任何状态或负责人过滤器
- When 页面加载
- Then 展示全部事项

---

## Story 9: 终态排序和页面过滤

**As a** PM 用户
**I want to** 事项清单和进度页面中终态主事项排在最后，且整体进度页面默认只看进行中的事项
**So that** 聚焦当前需要关注的事项，减少视觉干扰

**Acceptance Criteria:**
- Given 用户查看事项清单页面
- When 列表渲染完成
- Then 处于终态的主事项排在最后，非终态主事项保持原有顺序

- Given 用户首次打开整体进度页面
- When 页面加载
- Then "进行中"复选框默认选中，仅展示进行中状态的主事项

- Given 用户取消所有状态复选框
- When 过滤条件更新
- Then 展示全部事项

---

## Story 10: 甘特图修复

**As a** PM 用户
**I want to** 甘特图时间范围与实际数据对齐且在 macOS 下可水平滚动
**So that** 甘特图无空白浪费空间且所有平台都能查看完整时间范围

**Acceptance Criteria:**
- Given 甘特图加载可见主事项数据
- When 时间轴渲染完成
- Then 起始日期取最早开始时间的前1天，终止日期取最晚结束时间的后1天，无大片空白

- Given 用户在 macOS 系统下查看甘特图
- When hover 容器或常态下
- Then 底部水平滚动条可见

---

## Story 11: 团队选择器和每周进展过滤

**As a** 所有用户
**I want to** 团队选择器只展示我有权限的团队
**So that** 不会误选无权限团队

**Acceptance Criteria:**
- Given 用户登录后查看团队下拉选择器
- When 选择器渲染完成
- Then 仅展示当前用户有权限的团队

- Given 每周进展页面
- When 页面加载
- Then 不展示处于终态且本周（自然周，周一至周日）和上周都没有活跃事项的主事项；非终态主事项始终展示。**活跃事项定义**：在指定时间范围内（本周或上周），主事项满足以下任一条件即为活跃 — (a) `status_history` 表中存在该主事项的状态变更记录（`created_at` 在时间范围内）；(b) 该主事项下存在 `created_at` 或 `updated_at` 在时间范围内的子事项；(c) 该主事项的 `updated_at` 在时间范围内且变更类型为进度更新
