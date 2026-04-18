---
created: 2026-04-19
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: Improve UI

## Overview

对 PM Tracker 进行全面 UI 重做。前端从 Ant Design 迁移到 shadcn/ui 模式（Radix UI + Tailwind CSS），一次性替换全部 13 个页面。后端增量调整：扩展 User 模型、新增 Admin CRUD 接口、增强每周进展 API。

## Architecture

### Layer Placement

本次变更横跨前端全部三层，后端仅涉及 Service + Handler 层：

```
Frontend (React SPA)                    Backend (Go/Gin)
┌─────────────────────────┐            ┌──────────────────────┐
│ Pages (13 routes)       │            │ Handler (新增/修改)    │
│ Components (shadcn/ui)  │◄──REST────►│ Service (新增/修改)    │
│ Store + API Client      │            │ Repository (不变)     │
└─────────────────────────┘            │ Model (User 扩展)     │
                                       └──────────────────────┘
```

数据库层仅 User 表新增字段，其余表结构不变。

### Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│ Frontend                                                │
│  ┌──────────┐  ┌───────────┐  ┌───────────────────────┐ │
│  │ Pages    │─►│ Shared    │─►│ Radix UI Primitives   │ │
│  │ (13)     │  │ Components│  │ (Dialog, Dropdown,    │ │
│  │          │  │ (shadcn)  │  │  Popover, Tabs...)    │ │
│  └──────────┘  └───────────┘  └───────────────────────┘ │
│       │              │                                   │
│       ▼              ▼                                   │
│  ┌──────────┐  ┌───────────┐                            │
│  │ Router   │  │ API Client│                            │
│  │ (routes) │  │ (axios)   │                            │
│  └──────────┘  └───────────┘                            │
└─────────────────────────────────────────────────────────┘
         │                        │
         │  REST /api/v1          │
         ▼                        ▼
┌─────────────────────────────────────────────────────────┐
│ Backend                                                 │
│  ┌──────────┐  ┌───────────┐  ┌───────────────────────┐ │
│  │ Handler  │─►│ Service   │─►│ Repository (GORM)     │ │
│  │ (Gin)    │  │           │  │                       │ │
│  └──────────┘  └───────────┘  └───────────────────────┘ │
│                                                          │
│  Modified: AdminHandler, AdminService                    │
│  Modified: ViewHandler, ViewService (weekly)             │
│  Modified: UserModel (email + status)                    │
└─────────────────────────────────────────────────────────┘
```

### Dependencies

**前端移除：**
| Package | Reason |
|---------|--------|
| antd | 替换为 shadcn/ui 模式 |
| sass-embedded | Tailwind CSS 替代 |

**前端新增：**
| Package | Version | Purpose |
|---------|---------|---------|
| tailwindcss | ^4.x | Utility-first CSS |
| @tailwindcss/vite | ^4.x | Vite plugin |
| @radix-ui/react-dialog | latest | Modal/Dialog 原语 |
| @radix-ui/react-dropdown-menu | latest | Dropdown 原语 |
| @radix-ui/react-select | latest | Select 原语 |
| @radix-ui/react-tabs | latest | Tabs 原语 |
| @radix-ui/react-popover | latest | Popover 原语 |
| @radix-ui/react-progress | latest | Progress 原语 |
| @radix-ui/react-toggle | latest | Toggle 原语 |
| @radix-ui/react-tooltip | latest | Tooltip 原语 |
| @radix-ui/react-avatar | latest | Avatar 原语 |
| class-variance-authority | latest | 组件变体管理 |
| clsx | latest | Class 合并 |
| tailwind-merge | latest | Tailwind class 冲突解决 |
| lucide-react | latest | 图标库（与原型 SVG 对应） |

**前端保留：**
| Package | Purpose |
|---------|---------|
| react, react-dom | UI 框架 |
| react-router-dom | 路由 |
| @tanstack/react-query | 数据获取 |
| zustand | 状态管理 |
| axios | HTTP 客户端 |
| dayjs | 日期处理 |
| frappe-gantt | 甘特图（保留，样式用 Tailwind 覆盖） |
| vitest + testing-library + msw | 测试 |

**后端无新增依赖。**

## Interfaces

### 前端路由重构

**旧路由 → 新路由映射：**

| 旧路由 | 新路由 | 变化 |
|--------|--------|------|
| `/login` | `/login` | 不变 |
| `/items` | `/items` | Summary/Detail 视图切换 |
| `/items/:id` | `/items/:id` | 不变 |
| `/items/:id/sub/:subId` | `/items/:id/sub/:subId` | 不变 |
| `/item-pool` | `/item-pool` | 不变 |
| `/table` | `/table` | 独立页面（已有路由） |
| `/weekly` | `/weekly` | 重点变更（对比布局） |
| `/gantt` | `/gantt` | 不变 |
| `/report` | `/report` | 不变 |
| `/teams/:teamId/settings` | `/teams/:teamId` | 路由简化 |
| `/admin` | `/users` + `/teams` | 拆分 |

**新增路由：**

| 路由 | 组件 | Auth |
|------|------|------|
| `/teams` | TeamListPage | Protected |
| `/teams/:teamId` | TeamDetailPage | Protected (team member) |
| `/users` | UserManagementPage | SuperAdmin |

**移除路由：**

| 路由 | 原因 |
|------|------|
| `/admin` | 拆分为 `/users` + `/teams` |

### 前端组件架构

基于 shadcn/ui 模式，组件源码复制到项目中，基于 Radix UI 原语 + Tailwind CSS 构建：

```
src/
  components/
    ui/              ← shadcn/ui 风格基础组件
      button.tsx
      input.tsx
      select.tsx
      dialog.tsx
      badge.tsx
      card.tsx
      table.tsx
      dropdown-menu.tsx
      progress.tsx
      avatar.tsx
      tabs.tsx
      pagination.tsx
      breadcrumb.tsx
      toast.tsx
      toggle.tsx
      tooltip.tsx
      popover.tsx
    layout/          ← 布局组件
      AppLayout.tsx      (侧边栏 + 主内容区)
      Sidebar.tsx
    shared/          ← 业务共享组件
      StatusBadge.tsx
      PriorityBadge.tsx
      ProgressBar.tsx
      UserAvatar.tsx
      ItemFilters.tsx
      ConfirmDialog.tsx
  pages/             ← 13 个页面
  lib/
    utils.ts         ← cn() helper (clsx + tailwind-merge)
```

### 前端 API Client 变更

`src/api/` 目录结构保持不变，新增：

| 模块 | 新增接口 | 说明 |
|------|----------|------|
| admin.ts | `createUser()`, `updateUser()`, `getUser()`, `toggleUserStatus()` | 用户 CRUD |
| views.ts | 增强 `getWeeklyView()` | 返回上周/本周对比数据 |
| client.ts | 无变化 | 拦截器逻辑不变 |

## Data Models

### User 模型扩展

User 表新增两个字段：

```go
type User struct {
    gorm.Model
    Username      string `gorm:"uniqueIndex;size:64"`
    DisplayName   string `gorm:"size:64"`
    Email         string `gorm:"size:100"`           // 新增
    PasswordHash  string `gorm:"size:255" json:"-"`
    IsSuperAdmin  bool
    CanCreateTeam bool
    Status        string `gorm:"size:10;default:'enabled'"` // 新增：enabled/disabled
}
```

**迁移文件**: `migrations/004_user_email_status.sql`

```sql
ALTER TABLE users ADD COLUMN email VARCHAR(100) DEFAULT '';
ALTER TABLE users ADD COLUMN status VARCHAR(10) DEFAULT 'enabled';
```

### 前端 TypeScript 类型变更

```typescript
// types/index.ts 新增/修改

interface User {
  id: number;
  username: string;
  displayName: string;
  email: string;           // 新增
  canCreateTeam: boolean;
  isSuperAdmin: boolean;
  status: 'enabled' | 'disabled';  // 新增
  teams?: TeamSummary[];   // 新增：所属团队列表
}

// 每周进展对比数据
interface MainItemSummary {
  id: number;
  title: string;
  priority: string;
  startDate: string;
  expectedEndDate: string;
  completion: number;
  subItemCount: number;
}

interface WeeklyViewResponse {
  weekStart: string;
  weekEnd: string;
  stats: {
    activeSubItems: number;
    newlyCompleted: number;
    inProgress: number;
    blocked: number;
  };
  groups: WeeklyComparisonGroup[];
}

interface WeeklyComparisonGroup {
  mainItem: MainItemSummary;
  lastWeek: SubItemSnapshot[];
  thisWeek: SubItemSnapshot[];
  completedNoChange: SubItemSnapshot[];  // 默认折叠
}

interface SubItemSnapshot {
  id: number;
  title: string;
  priority: string;
  status: string;
  assigneeName: string;
  expectedEndDate: string;
  completion: number;
  progressDescription: string;
  delta?: number;       // 进度增量（本周-上周），>0 时显示
  isNew?: boolean;      // 本周新增
  justCompleted?: boolean; // 本周刚完成
}
```

## 后端 API 变更

### 需要新增的端点

| Method | Path | 说明 |
|--------|------|------|
| POST | `/admin/users` | 创建用户 |
| GET | `/admin/users/:userId` | 获取单个用户详情 |
| PUT | `/admin/users/:userId` | 编辑用户 |
| PUT | `/admin/users/:userId/status` | 启用/禁用用户 |

### 需要修改的端点

| Method | Path | 变更说明 |
|--------|------|----------|
| GET | `/admin/users` | 响应增加 email、status、teams 字段；增加搜索和筛选参数 |
| GET | `/teams/:teamId/views/weekly` | 响应重构为上周/本周对比格式（详见 API Handbook） |

### 不需要修改的端点

以下端点的请求/响应格式无需变更，前端仅需适配新组件：

- 所有 Team CRUD + Members 端点（团队详情页直接复用）
- 所有 Main Items / Sub Items 端点（事项清单页复用）
- 所有 Item Pool 端点（事项池页复用）
- Table View + Export 端点（全量表格页复用）
- Gantt View 端点（甘特图页复用）
- Report 端点（周报导出页复用）

## Error Handling

### 后端错误传播机制

沿用项目已有的 `AppError` 模式（定义在 `backend/internal/pkg/errors/`）：

```
Service 层                    Handler 层                   HTTP Response
┌──────────────┐            ┌──────────────┐            ┌──────────────┐
│ 返回         │            │ 捕获 error   │            │ {code, msg}  │
│ AppError     │───────────►│ RespondError │───────────►│ JSON 响应    │
│ 或 raw error │            │ (自动映射)   │            │              │
└──────────────┘            └──────────────┘            └──────────────┘
```

- **Service 层**返回 `AppError{Code, Message}` 或普通 `error`
- **Handler 层**调用 `errors.RespondError(c, appErr)` 统一输出 `{code: "ERROR_CODE", message: "..."}`
- 新增错误码注册到 `errors/` 包的预定义错误列表中

新增 AdminService 和 ViewService 的错误映射：

| Service 返回 | Handler 映射 | HTTP Status |
|-------------|-------------|-------------|
| `ErrUserExists` | `USER_EXISTS` | 422 |
| `ErrUserNotFound` | `USER_NOT_FOUND` | 404 |
| `ErrInvalidStatus` | `INVALID_STATUS` | 422 |
| `ErrCannotDisableSelf` | `CANNOT_DISABLE_SELF` | 422 |
| `ErrFutureWeek` | `FUTURE_WEEK_NOT_ALLOWED` | 422 |

### 新增错误码

| Code | HTTP Status | Description |
|------|-------------|-------------|
| USER_EXISTS | 422 | 创建用户时账号已存在 |
| USER_NOT_FOUND | 404 | 用户不存在 |
| USER_DISABLED | 403 | 禁用用户尝试登录 |
| INVALID_STATUS | 422 | 无效的用户状态值 |
| CANNOT_DISABLE_SELF | 422 | 超管不能禁用自己 |
| TEAM_NOT_FOUND | 404 | 团队不存在 |
| FUTURE_WEEK_NOT_ALLOWED | 422 | 不允许选择未来周 |

登录接口增加状态检查：禁用用户登录时返回 `USER_DISABLED`，提示"账号已被禁用"。

### 前端错误处理

axios 拦截器沿用现有机制（401 清除 token 跳转登录，403/404/500 全局 toast）。

新增业务错误码的前端展示策略：

| 错误码 | 展示方式 | 说明 |
|--------|----------|------|
| USER_EXISTS | 表单内联错误 | username 字段下方显示"账号已存在" |
| USER_NOT_FOUND | Toast 错误提示 | - |
| USER_DISABLED | 登录页内联错误 | 替换通用错误提示，显示"账号已被禁用" |
| INVALID_STATUS | Toast 错误提示 | - |
| CANNOT_DISABLE_SELF | Toast 错误提示 | - |
| FUTURE_WEEK_NOT_ALLOWED | Toast 错误提示 | - |
| 422 其他 | 表单内联错误 | 通用校验错误 |
| 400 其他 | Toast 错误提示 | - |

### 前端交互约定

| 场景 | 交互方式 |
|------|----------|
| 创建用户成功 | Dialog 弹窗显示初始密码，提示"请妥善保管，关闭后无法再次查看"，用户确认后关闭 |
| 甘特图 | 保留 frappe-gantt 组件，用 Tailwind 覆盖其默认样式以匹配设计系统 |

## Testing Strategy

### 前端测试

沿用 vitest + testing-library + msw 框架，重写测试适配新组件：

| 层级 | 覆盖内容 | 目标覆盖率 |
|------|----------|-----------|
| 组件测试 | ui/ 基础组件渲染和交互 | 90% |
| 页面测试 | 13 个页面的核心交互流程 | 70% |
| API 客户端测试 | 请求参数和响应处理 | 80% |
| Store 测试 | auth/team 状态管理 | 80% |

### 后端测试

沿用现有 testify + httptest 框架：

| 层级 | 覆盖内容 | 目标覆盖率 |
|------|----------|-----------|
| Handler | 新增 Admin CRUD 端点 | 80% |
| Service | AdminService 用户管理逻辑 | 85% |
| Service | ViewService 每周进展对比计算 | 85% |
| Integration | 登录时禁用用户拦截 | 100% |

## Security Considerations

### 新增风险

| 风险 | 等级 | 缓解措施 |
|------|------|----------|
| 禁用用户仍持有未过期 JWT | 中 | 前端 401 时清除 token 跳转登录；可后续引入 token 黑名单 |
| Admin 创建用户时密码传输 | 低 | HTTPS 传输，bcrypt 存储 |
| 每周进展 API 性能（跨周对比查询） | 低 | 单团队数据量有限，添加复合索引优化 |

### 不变的安全机制

- JWT 认证（HS256, 24h 过期）
- RBAC 中间件链（Auth → TeamScope → RequireRole）
- 团队数据隔离（查询层 teamID 过滤）
- bcrypt 密码哈希
- 登录接口限流（10 req/min/IP）

## Open Questions

（无待解决问题）

### 已解决问题

| 问题 | 决策 | 日期 |
|------|------|------|
| 创建用户时初始密码如何处理？ | 后端随机生成 12 位密码，通过 `initialPassword` 字段返回给前端，仅展示一次 | 2026-04-19 |
| 甘特图组件选型？ | 保留 frappe-gantt，用 Tailwind 覆盖样式以匹配设计系统 | 2026-04-19 |
| 每周进展 API 是否需要缓存？ | 不缓存，直接查询。单团队数据量有限，无需引入缓存复杂度 | 2026-04-19 |

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|----------|------|------|----------------|
| Ant Design + Tailwind 覆盖 | 迁移成本最低 | antd 样式难以完全覆盖，产物体积大（~500KB gzip） | 原型风格与 antd 差异太大 |
| Headless UI + Tailwind | 轻量，Vue/React 都支持 | 组件覆盖不如 Radix 全（无 Dialog、Progress） | Radix 原语更丰富 |
| Tailwind 直接写 | 零依赖 | 大量手写无障碍代码 | 工作量大，质量难保证 |
| **shadcn/ui + Radix + Tailwind** | **可访问性内置、组件可定制、产物小** | **需手动维护组件源码** | **选定方案** |

### References

- 原型文件：`docs/features/improve-ui/ui/prototype/`
- 原技术设计：`docs/features/pm-work-tracker/design/tech-design.md`
- 原 API 手册：`docs/features/pm-work-tracker/design/api-handbook.md`
