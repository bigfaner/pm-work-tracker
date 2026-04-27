---
feature: "用户管理增强：重置密码 & 删除用户"
---

# 用户管理增强：重置密码 & 删除用户 — UI Functions

> Requirements layer: defines WHAT the UI must do. Not HOW it looks (that's ui-design.md).

## UI Scope

- 用户管理列表页（UserManagementPage）：新增操作列按钮
- 重置密码弹窗（新增）
- 删除确认弹窗（新增）

---

## UI Function 1: 用户列表操作列扩展

### Description
在现有用户列表每行的操作列中，为超级管理员新增「重置密码」和「删除」两个操作入口。

### User Interaction Flow
1. 超级管理员进入用户管理页，列表每行操作列显示：编辑 / 重置密码 / 删除
2. 点击「重置密码」→ 打开重置密码弹窗（UI Function 2）
3. 点击「删除」→ 打开删除确认弹窗（UI Function 3）
4. 若当前行为登录用户自身，「删除」按钮禁用

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| isSuperAdmin | boolean | 当前登录用户信息 | 控制按钮可见性 |
| bizKey | string | 用户列表数据 | 用于调用 API |
| username | string | 用户列表数据 | 用于确认弹窗展示 |
| 当前用户 bizKey | string | auth store | 用于判断是否为自身 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 超级管理员视图 | 显示「重置密码」「删除」按钮 | isSuperAdmin=true |
| 非超级管理员视图 | 不显示「重置密码」「删除」按钮 | isSuperAdmin=false |
| 自身账号行 | 「删除」按钮禁用，Tooltip 提示 | 目标 bizKey === 当前用户 bizKey |

### Validation Rules
- 仅 isSuperAdmin=true 时渲染按钮
- 目标用户为自身时，删除按钮 disabled

---

## UI Function 2: 重置密码弹窗

### Description
模态弹窗，供超级管理员为指定用户设置新密码。

### User Interaction Flow
1. 点击「重置密码」后弹窗打开，标题显示"重置密码 — {displayName}"
2. 输入新密码（密码框，可切换显示/隐藏）
3. 输入确认密码
4. 点击「确认」→ 客户端校验 → 通过则调用 API → 成功关闭弹窗并 Toast 提示
5. 点击「取消」或弹窗外区域 → 关闭弹窗，不提交

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| displayName | string | 用户列表数据 | 弹窗标题展示 |
| newPassword | string | 用户输入 | 必填，≥8位，含字母和数字 |
| confirmPassword | string | 用户输入 | 必须与 newPassword 一致 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 初始 | 两个输入框为空，确认按钮可点击 | 弹窗打开 |
| 校验失败 | 对应字段下方显示红色错误文字，弹窗保持打开 | 点击确认且校验不通过 |
| 提交中 | 确认按钮 loading，输入框禁用 | API 请求进行中 |
| 成功 | 弹窗关闭，Toast "密码已重置" | API 返回成功 |
| 失败 | 弹窗内显示错误提示 | API 返回错误 |

### Validation Rules

| 规则 | 触发时机 | 错误提示 |
|------|----------|----------|
| 新密码不为空 | 提交 | 请输入新密码 |
| 密码 ≥8位，含字母和数字 | 失焦 / 提交 | 密码需至少8位，包含字母和数字 |
| 确认密码与新密码一致 | 失焦 / 提交 | 两次输入的密码不一致 |

---

## UI Function 3: 删除用户确认弹窗

### Description
二次确认弹窗，防止管理员误删用户。弹窗中明确展示目标用户名。

### User Interaction Flow
1. 点击「删除」后弹窗打开，显示"确认删除用户 {username}？此操作不可通过界面撤销。"
2. 点击「确认删除」→ 调用 API → 成功后关闭弹窗，列表移除该行，Toast "用户已删除"
3. 点击「取消」→ 关闭弹窗，不执行任何操作

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| username | string | 用户列表数据 | 弹窗正文展示，明确告知操作对象 |
| bizKey | string | 用户列表数据 | API 调用参数 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 确认中 | 显示用户名，两个按钮可点击 | 弹窗打开 |
| 提交中 | 确认按钮 loading，两个按钮禁用 | API 请求进行中 |
| 成功 | 弹窗关闭，列表刷新，Toast 提示 | API 返回成功 |
| 失败 | 弹窗内显示错误提示 | API 返回错误 |

### Validation Rules
- 无客户端校验，确认即提交

---

## UI Function 4: 创建用户成功结果弹窗 — 复制账号与密码

### Description
创建用户成功后展示的结果弹窗中，新增「复制账号与密码」按钮，让管理员一键将账号和初始密码复制到剪贴板，方便转交给新用户。

### User Interaction Flow
1. 管理员提交创建用户表单，后端返回成功（含 `initialPassword`）
2. 结果弹窗展示用户名和初始密码
3. 管理员点击「复制账号与密码」
4. 系统将以下文本写入剪贴板：
   ```
   账号：{username}
   密码：{initialPassword}
   ```
5. 按钮文字变为"已复制"，约 2 秒后恢复为"复制账号与密码"
6. 若复制失败（浏览器权限拒绝），显示 Toast 错误提示

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| username | string | 创建用户 API 响应 | 复制内容的账号部分 |
| initialPassword | string | 创建用户 API 响应 | 仅在创建成功响应中出现，不持久化展示 |

### States

| State | Display | Trigger |
|-------|---------|---------|
| 默认 | 按钮文字"复制账号与密码" | 弹窗打开 |
| 复制成功 | 按钮文字变为"已复制"，约 2 秒后恢复 | Clipboard API 写入成功 |
| 复制失败 | Toast 错误提示"复制失败，请手动复制" | Clipboard API 抛出异常 |

### Validation Rules
- 无校验，点击即触发复制
- `initialPassword` 为空时隐藏该按钮（防御性处理）
