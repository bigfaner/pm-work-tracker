---
feature: "weekly-stats-optimization"
---

# User Stories: 每周进展统计优化

## Story 1: 快速识别风险事项

**As a** PM / 项目负责人
**I want to** 在每周进展页顶部统计栏直接看到未开始、暂停中、逾期中的子事项数量
**So that** 我可以在 1 分钟内判断团队是否存在风险信号，无需逐一展开子事项

**Acceptance Criteria:**
- Given 当前周次有 2 个 pending 子事项、1 个 pausing 子事项、3 个逾期子事项（expectedEndDate < weekEnd 且 status ∉ completed/closed）
- When 我访问每周进展页
- Then 统计栏显示"未开始 2"、"暂停中 1"、"逾期中 3"三个卡片，数字与实际状态一致

---

## Story 2: 理解统计数字含义

**As a** 团队成员
**I want to** 通过 hover 卡片查看每个统计数字的计算规则说明
**So that** 我不需要阅读源码就能理解"活跃"、"逾期中"等数字的口径，消除周会中的疑惑

**Acceptance Criteria:**
- Given 我在桌面端访问每周进展页，统计栏已加载完成
- When 我将鼠标悬停在"本周活跃"卡片上超过 300ms
- Then 显示 tooltip，内容为"本周有进展记录，或计划周期与本周重叠的子事项总数"，与方案定义完全一致

---

## Story 3: 移动端查看统计说明

**As a** 团队成员
**I want to** 在移动设备上点击统计卡片查看规则说明
**So that** 在没有鼠标 hover 的触摸设备上也能获取统计口径信息

**Acceptance Criteria:**
- Given 我在移动端（触摸设备）访问每周进展页
- When 我点击任意统计卡片
- Then tooltip 展开显示该卡片的统计规则说明；再次点击同一卡片时 tooltip 收起

---

## Story 4: 键盘用户访问统计说明

**As a** 使用键盘导航的用户
**I want to** 通过 Tab 键聚焦统计卡片并查看 tooltip
**So that** 无需鼠标也能获取统计规则说明，满足无障碍访问需求

**Acceptance Criteria:**
- Given 我在每周进展页使用 Tab 键导航
- When 我 Tab 到某个统计卡片时
- Then 该卡片获得焦点，tooltip 可见；卡片元素携带 `aria-describedby` 属性，屏幕阅读器可读取 tooltip 内容
