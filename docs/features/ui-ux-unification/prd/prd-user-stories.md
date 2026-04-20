---
feature: "ui-ux-unification"
---

# User Stories: UI/UX 统一与优化

## Story 1: Safari 用户使用周次选择器

**As a** 使用 Safari 浏览器的 PM 或团队成员
**I want to** 正常选择和切换周次查看每周进展与周报
**So that** 不因浏览器差异而无法使用核心功能

**Acceptance Criteria:**
- Given 用户使用 Safari 浏览器打开每周进展页或周报导出页
- When 页面加载完成
- Then 周次选择器正常显示，格式为 `‹ 2026年第16周  04/13 ~ 04/19  ›`，可正常交互

---

## Story 2: 快速切换上一周/下一周

**As a** PM
**I want to** 点击一个按钮就切换到上一周或下一周
**So that** 快速对比不同周次的进展，无需手动修改输入框

**Acceptance Criteria:**
- Given 用户在每周进展页或周报导出页
- When 用户点击「‹」按钮
- Then 周次切换到上一周，日期范围更新，数据自动刷新
- Given 当前显示的是本周
- When 用户尝试点击「›」按钮
- Then 「›」按钮处于禁用状态，不可点击

---

## Story 3: 通过视觉识别可跳转链接

**As a** 团队成员
**I want to** 一眼看出哪些文字可以点击跳转
**So that** 快速导航到相关事项详情，减少误操作

**Acceptance Criteria:**
- Given 用户浏览任意列表页或详情页
- When 页面中存在可跳转的事项标题或关联链接
- Then 这些文字显示为主色高亮（`text-primary-600`），与普通文字明显区分
- When 用户鼠标悬停在链接上
- Then 显示下划线，进一步确认可点击

---

## Story 4: 在每周进展页快速了解子事项进度

**As a** PM
**I want to** 在每周进展页直接看到每个子事项的完成度百分比
**So that** 无需点击进入详情页即可掌握整体进展

**Acceptance Criteria:**
- Given 用户在每周进展页查看某主事项的子事项列表
- When 子事项行渲染完成
- Then 每个子事项标题旁显示完成度数字（如 `65%`）
- When 子事项完成度为 100%
- Then 百分比数字显示为绿色

---

## Story 5: 在进度记录卡片处追加进度

**As a** 团队成员
**I want to** 在进度记录卡片旁直接点击「追加进度」
**So that** 操作上下文清晰，不需要在页面顶部寻找按钮

**Acceptance Criteria:**
- Given 用户在子事项详情页，拥有 `progress:update` 权限
- When 用户查看进度记录卡片
- Then 「追加进度」按钮显示在进度记录卡片的右上角
- When 用户点击该按钮
- Then 弹出追加进度对话框，行为与原来一致

---

## Story 6: 通过清晰的菜单分组快速导航

**As a** 管理员或 PM
**I want to** 在左侧菜单中快速找到业务功能和管理功能
**So that** 减少寻找菜单项的时间，提升操作效率

**Acceptance Criteria:**
- Given 用户登录系统后查看左侧导航菜单
- When 菜单渲染完成
- Then 业务功能（事项清单、待办事项、每周进展、整体进度、周报导出）在上方分组显示
- Then 管理功能（团队管理、用户管理、角色管理）在下方分组显示，与业务组之间有分隔线
