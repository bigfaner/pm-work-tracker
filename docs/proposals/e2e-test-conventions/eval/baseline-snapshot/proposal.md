---
created: 2026-06-03
author: fanhuifeng
status: Draft
intent: doc
---

# Proposal: E2E Test Conventions — 从实践提炼规范 + Surface 重组

## Problem

两个问题叠加：

1. **规范缺失**：`docs/conventions/testing/` 中的 api/core.md 和 web/core.md 是 forge 自动生成的框架文档，描述理想化策略，缺少从实际代码提炼的实践规范。
2. **Surface 混放**：所有测试（API/Web/CLI）混放在 `tests/e2e/` 下，不符合 Forge surface 模型（每个 surface 应有独立的 build/dev/test 生命周期）。API 测试依赖 Playwright 但完全不使用浏览器能力。

### Evidence

| 维度 | 框架文档现状 | 实际代码中的模式 |
|------|------------|---------------|
| HTTP 客户端 | 未指定具体工具 | 3 种方式混用（curl/fetch/request.newContext） |
| 认证方式 | "配置认证" | 4 种登录方式（localStorage 注入/表单/API token/cache） |
| 状态码断言 | "精确断言 HTTP 状态码" | 混用 `toBe(200)` 和 `=== 200 \|\| === 201` |
| 等待策略 | "事件驱动等待" | 大量 `waitForTimeout(2000)` 硬编码延迟 |
| 测试数据 | "数据隔离" | 无 afterAll 清理，数据累积 |
| Helpers 使用 | 未提及 | 部分文件重复定义 create 函数 |
| 目录结构 | "tests/\<journey\>/" | API/Web/CLI 混放在 tests/e2e/ |
| API 测试框架 | 未指定 | Playwright Test（不使用浏览器能力） |

### Urgency

AI Agent 执行 `/gen-test-scripts` 时依赖 conventions 生成代码。缺少实践规范导致生成的测试与现有代码风格不一致，增加 review 成本。Surface 混放导致无法独立运行单个 surface 的测试。

## Proposed Solution

两部分工作：

1. **Surface 优先重组**：将 `tests/e2e/` 拆分为 `tests/api/`、`tests/web/`、`tests/cli/`，共享 helpers 提取到 `tests/shared/`。API 测试从 Playwright 迁移到 Vitest。
2. **实践规范补充**：在 api/core.md、web/core.md、index.md 中追加实践规范和反模式清单。

### 目标目录结构

```
tests/
├── shared/                  # 跨 surface 共享
│   ├── helpers.ts           # 从 tests/e2e/helpers.ts 提取
│   └── config.yaml          # 从 tests/e2e/config.yaml 移入
├── api/                     # API Functional Tests — Vitest
│   ├── vitest.config.ts
│   ├── package.json
│   ├── items/
│   │   ├── rbac.spec.ts
│   │   ├── status-flow.spec.ts
│   │   └── soft-delete.spec.ts
│   ├── roles/
│   │   └── rbac.spec.ts
│   ├── teams/
│   │   └── rbac.spec.ts
│   ├── users/
│   │   └── user-mgmt.spec.ts
│   ├── item-pool/
│   │   └── pool.spec.ts
│   ├── progress/
│   │   └── progress-auto-status.spec.ts
│   └── smoke/
│       └── deploy.spec.ts
├── web/                     # Web E2E Tests — Playwright
│   ├── playwright.config.ts
│   ├── package.json
│   ├── items/
│   │   ├── item-list.spec.ts
│   │   └── item-ui.spec.ts
│   ├── roles/
│   │   └── role-crud.spec.ts
│   ├── teams/
│   │   └── team-management.spec.ts
│   ├── users/
│   │   └── user-mgmt.spec.ts
│   ├── auth/
│   │   └── login.spec.ts
│   ├── weekly/
│   │   └── weekly-view.spec.ts
│   ├── item-pool/
│   │   └── pool.spec.ts
│   └── smoke/
│       └── full-e2e.spec.ts
└── cli/                     # CLI Functional Tests — Vitest
    ├── vitest.config.ts
    ├── package.json
    └── infra/
        ├── bizkey.spec.ts
        ├── schema.spec.ts
        ├── lint-keywords.spec.ts
        └── config-yaml.spec.ts
```

### API 测试框架迁移

从 Playwright Test 迁移到 Vitest + 原生 fetch：

| 维度 | 迁移前 | 迁移后 |
|------|--------|--------|
| 运行器 | `@playwright/test` | `vitest` |
| 断言库 | Playwright `expect` | Vitest `expect` |
| HTTP 客户端 | 自定义 `curl()` | 保留（已基于原生 fetch） |
| import 路径 | `import { test, expect } from '@playwright/test'` | `import { test, expect } from 'vitest'` |
| 配置文件 | playwright.config.ts | vitest.config.ts |
| 共享 helpers | `../helpers.js` | `../shared/helpers.js` |

迁移改动量：API 测试文件基本只需改 import 和文件名后缀（`-api` 前缀去掉），`curl()` 和 helpers 不变。

### Innovation Highlights

从"规范先行"到"实践先行"的归纳过程。Surface 优先结构使每个测试类型可以用最适合的工具（API → Vitest，Web → Playwright），而不是用浏览器自动化框架跑 HTTP 测试。

### 具体规范清单

#### API Functional Test 实践规范（api/core.md）

测试框架：Vitest (`vitest`) 作为测试运行器和断言库，HTTP 请求走 shared/helpers.ts 的 `curl()`（基于 Node.js 原生 `fetch()`）。

| ID | 规则 | 正确示例 | 错误示例 |
|---|---|---|---|
| **E-100** | HTTP 客户端统一 — 所有 API 调用使用 `curl()` | `const res = await curl('POST', url, { headers, body })` | `fetch(url, ...)`, `playwright.request.newContext()` |
| **E-101** | 响应体解析 — 标准 API 响应用 `parseApiBody()` | `const data = parseApiBody(res.body)` (自动校验 code===0) | `JSON.parse(res.body)` 不检查 code |
| **E-102** | 状态码精确断言 — 一个 `expect` 只断言一个状态码 | `expect(res.status).toBe(200)` | `expect(res.status === 200 \|\| res.status === 201).toBeTruthy()` |
| **E-103** | RBAC 脚手架 — 多用户权限测试用 `setupRbacFixtures()` | `const f = await setupRbacFixtures()` 一站式创建 | 手动分别 createTestTeam + createTestUser + createTestRole + addUserToTeam |
| **E-104** | 测试数据唯一性 — 模块级 `Date.now()` + 前缀命名 | `const runId = Date.now(); createTestTeam(token, 'e2e-rbac-' + runId)` | 硬编码 `'test-team'` |
| **E-105** | 认证模式 — `getApiToken()` 获取 token，`authHeader()` 构造 header | `const token = await getApiToken(apiBaseUrl); curl(..., { headers: authHeader(token) })` | 手动 fetch login + 拼接 `'Bearer ' + token` |
| **E-106** | 测试编号与可追溯性 — TC-XXX 格式 + Traceability 注释 | `// Traceability: TC-001 -> Story 1` + `test('TC-001: 描述', ...)` | `test('should work', ...)` 无编号无注释 |

**API 反模式清单**：

| # | 反模式 | 替代 |
|---|--------|------|
| 1 | 混用 HTTP 客户端（curl + fetch + request.newContext） | 统一用 `curl()` |
| 2 | 松散状态码断言（`\|\|` 接受多个状态码） | 精确断言 + 分开测试每种场景 |
| 3 | 本地重复定义 create 函数 | 用 shared/helpers.ts 共享函数 |
| 4 | 硬编码 `http://127.0.0.1:8080` | 用 `apiBaseUrl` / `apiUrl` 常量 |
| 5 | 仅断言 status code 不检查 body | 至少断言一个响应体字段 |
| 6 | 使用 Playwright 作为 API 测试运行器 | 用 Vitest — 无浏览器开销，启动快 5-10x |

#### Web E2E Test 实践规范（web/core.md）

测试框架：Playwright Test (`@playwright/test`)，浏览器自动化。

| ID | 规则 | 正确示例 | 错误示例 |
|---|---|---|---|
| **E-110** | 登录方式 — `login(page)` localStorage 注入 | `await login(page)` — 注入 Zustand auth state | `browserLogin(page, user, pass)` 表单登录（仅登录专项测试用） |
| **E-111** | 元素定位优先级 — data-testid > role+name > label > text | `[data-testid="item-view-page"]`, `getByRole('button', { name: '新增' })` | `.some-class`, `div > span:nth-child(2)` |
| **E-112** | 串行流程测试 — 有状态依赖的测试用 `.serial` | `test.describe.serial('事项列表 - Full E2E', ...)` | 无 `.serial` 但测试间共享变量 |
| **E-113** | 等待策略 — 显式等待条件，不硬编码延迟 | `await expect(locator).toBeVisible()` | `await page.waitForTimeout(2000)` |
| **E-114** | 测试数据准备 — `beforeAll` 中用 helpers API 创建 | `test.beforeAll(async () => { token = await getAuthToken(); id = await createTestMainItem(...) })` | 每个测试内创建数据 / 依赖手工预置 |
| **E-115** | 页面导航 — `login(page, undefined, targetPath)` 或 `navTo()` | `await login(page, undefined, '/items/' + itemId)` | 手动 `page.click('a')` + `waitForTimeout` |

**Web 反模式清单**：

| # | 反模式 | 替代 |
|---|--------|------|
| 1 | `waitForTimeout()` 硬编码延迟 | `toBeVisible()` / `waitForSelector()` 显式等待 |
| 2 | CSS class 选择器 | `data-testid` 或语义定位器 |
| 3 | 每个测试独立表单登录 | `login(page)` 注入 + `beforeEach` 复用 |
| 4 | 本地重复定义 create 函数 | shared/helpers.ts 共享函数 |
| 5 | `catch(() => {})` 吞掉导航错误 | 显式等待特定条件 + 正确错误传播 |

#### Helpers 工具库规范（index.md）

| ID | 规则 | 说明 |
|---|---|---|
| **E-120** | 配置来源 — `config.yaml` + 环境变量 | 用 `baseUrl`/`apiBaseUrl` 常量，通过 `E2E_CONFIG_PATH` 覆盖，不硬编码 URL |
| **E-121** | Token 缓存 — `getApiToken()` / `getAuthToken()` | 自动缓存 23h TTL，不手动管理 token 生命周期 |
| **E-122** | CLI 执行 — `runCli(cmd, cwd, timeout)` | 统一包装 `execSync`，返回 `{stdout, stderr, exitCode}` |
| **E-123** | 调试截图 — `screenshot(page, tcId)` | 保存到 `results/screenshots/`，不在测试中拼接路径 |

## Requirements Analysis

### Key Scenarios

- AI Agent 通过 `/gen-test-scripts` 生成新测试 → 读取 conventions 生成符合风格的代码
- 开发者写新测试 → 查阅 conventions 快速了解 helpers 用法和项目约定
- Code review → 以 conventions 中的反模式清单为检查标准
- CI 运行 → 按 surface 独立触发（api 只启动后端，web 启动全栈）

### Constraints & Dependencies

- 现有框架文档有 `<!-- auto-generated by forge:test-guide -->` 标记，追加内容不应与此冲突
- 文件 frontmatter 的 `domains` 字段需更新以覆盖新内容
- API 迁移到 Vitest 后，`test`/`expect` API 与 Playwright 几乎一致，但 `test.describe.serial` 无直接等价（Vitest 用 `test.sequential`）

## Alternatives & Industry Benchmarking

### 目录结构方案

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| 维持现状（tests/e2e/ 混放） | 零迁移成本 | Surface 无法独立运行，API 依赖 Playwright | Rejected: 不符合 Forge 模型 |
| Journey 优先（tests/\<journey\>/api/ + web/） | 同功能测试相邻 | 每个 journey 下重复 api/web 子目录，结构深 | Rejected: 嵌套过深 |
| **Surface 优先（tests/api/ + tests/web/ + tests/cli/）** | Surface 完全隔离，各用最适合的工具 | 跨 surface 共享 helpers 需相对路径 | **Selected: 符合 Forge surface 生命周期** |

### API 测试框架方案

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| Playwright Test（现状） | 统一运行器 | 加载浏览器模块，API 测试不使用任何浏览器能力 | Rejected: 大材小用 |
| Playwright request API | 改进 HTTP 客户端 | 仍加载浏览器模块，启动慢 | Rejected: 改进不大 |
| Go httptest | 与后端同语言 | 丢失共享 helpers，不同工具链 | Rejected: 团队以 TS 为主 |
| **Vitest + 原生 fetch** | 项目已有 Vitest，启动快，API 兼容 | 需要迁移 import | **Selected: 最适合 API surface** |

## Feasibility Assessment

### Technical Feasibility

- 目录重组：文件移动 + import 路径调整，机械性操作
- API 迁移：`import { test, expect } from '@playwright/test'` → `import { test, expect } from 'vitest'`，API 几乎一致
- Vitest 的 `test.sequential` 替代 Playwright 的 `test.describe.serial`

### Resource & Timeline

两部分工作量：
1. Surface 重组 + API 框架迁移：约 2-3 小时（~49 个测试文件的路径和 import 调整）
2. 实践规范文档更新：约 1 小时

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| 现有框架文档足够指导测试编写 | Stress Test: Agent 基于现有文档生成测试，对比实际代码 | Confirmed: 生成结果与实际风格差距大 |
| API 测试必须用 Playwright（因为 Web 测试用 Playwright） | Assumption Flip: Surface 分离后各 surface 可独立选择工具 | Overturned: Surface 优先 = 框架自由选择 |
| 测试目录按 journey 组织是唯一方式 | XY Detection: 真正的需求是"能找到相关测试"，不是"按功能分目录" | Refined: Surface 优先更符合 Forge 生命周期模型 |

## Scope

### In Scope

- Surface 优先目录重组：`tests/e2e/` → `tests/api/` + `tests/web/` + `tests/cli/` + `tests/shared/`
- API 测试从 Playwright 迁移到 Vitest
- 更新 `docs/conventions/testing/api/core.md`：追加 API 实践规范（E-100 ~ E-106）+ 反模式清单 + 文件位置更新
- 更新 `docs/conventions/testing/web/core.md`：追加 Web 实践规范（E-110 ~ E-115）+ 反模式清单 + 文件位置更新
- 更新 `docs/conventions/testing/index.md`：追加 helpers 工具库使用规范（E-120 ~ E-123）+ 目录结构更新
- 更新各文件的 `domains` frontmatter

### Out of Scope

- 修复现有测试中的反模式代码（本次只记录规范，不修改测试逻辑）
- 修改 helpers.ts 的实现（仅移动位置）
- 修改 Forge 配置（surface key 定义等）
- CI pipeline 调整（目录变更后的 CI 适配另议）

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 迁移引入 regression（测试因路径/配置变更而失败） | M | H | 逐个 surface 迁移并验证，每步确认测试通过 |
| Vitest 和 Playwright 的 test API 有细微差异（如 .serial） | L | M | 逐个文件检查，`test.describe.serial` → `describe.sequential` |
| 规范与迁移后代码不一致（spec drift） | M | M | 先完成迁移，再从最终代码提炼规范 |
| 框架文档和实践规范内容重叠 | L | L | 实践规范只写具体工具/方法，不重复策略层面 |

## Success Criteria

- [ ] `tests/` 目录按 surface 优先结构组织（api/ web/ cli/ shared/）
- [ ] API 测试全部使用 Vitest，无 Playwright 依赖
- [ ] Web 测试保留 Playwright，配置独立
- [ ] 所有 API 测试迁移后通过
- [ ] 所有 Web 测试迁移后通过
- [ ] api/core.md 包含 7 条 API 实践规范（E-100 ~ E-106），每条有正确/错误代码示例
- [ ] web/core.md 包含 6 条 Web 实践规范（E-110 ~ E-115），每条有正确/错误代码示例
- [ ] index.md 包含 4 条 helpers 使用规范（E-120 ~ E-123）
- [ ] 两个 surface 文件各包含反模式清单（5 条以上），与现有框架文档的反模式不重复
- [ ] 所有更新文件的 `domains` frontmatter 已更新

## Next Steps

- 直接实施（doc + refactor 类型，无需 PRD/tech-design）
- 实施顺序：先完成目录重组 + 框架迁移（验证通过），再更新 conventions 文档
