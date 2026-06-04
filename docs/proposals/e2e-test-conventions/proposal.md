---
created: 2026-06-03
author: fanhuifeng
status: Draft
intent: doc
---

# Proposal: E2E Test Conventions — 从实践提炼规范 + Surface 重组

## Problem

两个独立但相互放大的问题：

1. **规范缺失**：`docs/conventions/testing/` 中的 api/core.md 和 web/core.md 是框架文档，缺实践规范。
2. **Surface 混放**：所有测试混放在 `tests/e2e/` 下，API 测试依赖 Playwright 但不使用浏览器。构建检查文件误标为 `-cli`（项目无 CLI surface）。

两者互为前提——重组后才能确定工具选择（API → Vitest），确定工具后才能写准确规范。

### Evidence

| 维度 | 框架文档现状 | 实际代码中的模式 |
|------|------------|---------------|
| HTTP 客户端 | 未指定 | 3 种混用（curl/fetch/request.newContext），3 文件用后者 |
| 认证方式 | "配置认证" | 4 种混用（localStorage/表单/API token/cache） |
| 状态码断言 | "精确断言" | 混用 `toBe(200)` 和 `=== 200 || === 201` |
| 等待策略 | "事件驱动" | 177 处 `waitForTimeout()`（10 文件，item-list.spec.ts 54 处） |
| 测试数据 | "数据隔离" | 无 afterAll 清理 |
| Helpers | 未提及 | 部分文件重复定义 create |
| 目录 | "tests/\<journey\>/" | API/Web/构建检查混放 tests/e2e/ |
| API 框架 | 未指定 | Playwright（不使用浏览器） |

### Urgency

AI Agent 执行 `/gen-test-scripts` 时依赖 conventions 生成代码。缺少规范导致生成代码风格不一致。Surface 混放导致无法独立运行单个 surface。

**Sequencing**：Conventions 排在 Phase 5 因"从实践提炼"需先有正确结构——先重组再提炼，而非凭空编写。

## Proposed Solution

两部分工作：

1. **Surface 优先重组**：将 `tests/e2e/` 拆分为 `tests/api/`、`tests/web/`、`tests/infra/`，共享 helpers 提取到 `tests/shared/`。API 测试从 Playwright 迁移到 Vitest。
2. **实践规范补充**：在 api/core.md、web/core.md、index.md 中追加实践规范和反模式清单。

### 目标目录结构

```
tests/
├── shared/                  # 跨 surface 共享（禁止 import Playwright）
│   ├── helpers.ts           # 从 tests/e2e/helpers.ts 提取，仅含纯 HTTP/CLI 工具函数
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
└── infra/                   # 静态分析 & 构建检查 — Vitest
    ├── vitest.config.ts
    ├── package.json
    ├── bizkey-build.spec.ts
    ├── schema-_mysql.spec.ts
    ├── lint-keywords.spec.ts
    ├── config-yaml-build.spec.ts
    ├── e2e-rebuild.spec.ts
    └── permission-checks-build.spec.ts
```

**`tests/api/vitest.config.ts`**：`testTimeout: 30s`，`hookTimeout: 60s`，`sequence: { sequential: true }`。infra 同配置但 timeout 15s。

> **为什么 `sequence: { sequential: true }`**：API 测试共享 `beforeAll` 创建的测试数据（如 team、role），parallel 会导致数据竞争。当前 Playwright `workers:1` 已是串行，迁移后保持语义不变。未来可考虑 per-file isolation 实现并行。

**无后缀文件分类**：含浏览器 API → `web/`；仅 HTTP → `api/`；`runCli()` 构建检查 → `infra/`；跨 surface 归主导 surface（`web/smoke/`）。

**注意**：项目无 CLI surface。`infra/` 中的 `-cli` 后缀文件实际是静态分析/构建检查（`go build`、`grep`、`lint-staged`），不是 CLI 功能测试。

**Infra 文件迁移映射**（11 spec + 1 helper）：

`tests/api/smoke/`：`deploy-smoke.spec.ts`、`jlc-schema-api.spec.ts`、`schema-alignment-api.spec.ts`、`config-yaml-api.spec.ts`
`tests/web/smoke/`：`schema-alignment-ui.spec.ts`
`tests/infra/`：`config-yaml-build.spec.ts`（原 config-yaml-cli）、`bizkey-build.spec.ts`（原 bizkey-cli）、`schema-mysql.spec.ts`（原 jlc-schema-cli）、`lint-keywords.spec.ts`、`e2e-rebuild.spec.ts`、`permission-checks-build.spec.ts`（原 unify-permission-checks-build）
`tests/shared/`：`config-setup.ts`（非 spec，共享配置）

**Shared helpers 所有权规则**：
- `tests/shared/` 中的代码**禁止** import `@playwright/test` 或任何浏览器 API。只允许纯 Node.js 代码（fetch、execSync 等）。
- Playwright-specific helpers（如 `login(page)`、`screenshot(page, tcId)`）放在 `tests/web/helpers/` 下。
- API-specific helpers（如 `curl()`、`parseApiBody()`）放在 `tests/shared/` 下，因为它们不依赖浏览器。
- Surface-specific 通过相对路径引用 shared，不反向依赖。如 API test `import { curl } from '../shared/helpers'`。

从 Playwright Test 迁移到 Vitest + 原生 fetch：

| 维度 | 迁移前 | 迁移后 |
|------|--------|--------|
| 运行器 | `@playwright/test` | `vitest` |
| 断言库 | Playwright `expect` | Vitest `expect` |
| HTTP 客户端 | 自定义 `curl()` | 保留（已基于原生 fetch） |
| import 路径 | `import { test, expect } from '@playwright/test'` | `import { test, expect } from 'vitest'` |
| 配置文件 | playwright.config.ts | vitest.config.ts |
| 共享 helpers | `../helpers.js` | `../shared/helpers.js` |

**Playwright → Vitest API 兼容性矩阵**：

| Playwright API | Vitest 等价 | 行为差异 |
|----------------|------------|---------|
| `test.describe.serial()` | `describe.sequential()` | **失败传播不同**：Playwright `.serial` 在首个 test 失败后跳过剩余 tests（状态 `skipped`）；Vitest `.sequential` 继续运行后续 tests。依赖"失败即跳过"语义的测试需要额外 guard 逻辑 |
| `test.beforeAll()` | `beforeAll()` | 一致 |
| `test.afterAll()` | `afterAll()` | 一致 |
| `test.beforeEach()` | `beforeEach()` | 一致 |
| `test.skip()` | `test.skip()` | 一致 |
| `expect().toBe()` | `expect().toBe()` | 一致 |
| `expect().toBeTruthy()` | `expect().toBeTruthy()` | 一致 |
| `test.setTimeout(ms)` | `test('name', { timeout: ms }, fn)` | `item-list.spec.ts` 需迁移 |

迁移改动量：多数文件改 import + 去 surface 后缀。3 个用 `request.newContext` 的文件改写为 `curl()`。`.serial` 测试需检查失败跳过语义。infra 中的 `-cli` 后缀改为 `-build` 以反映真实功能。52 spec 文件中 ~41 活跃（`.graduated/` 含 12 个已完成 feature 的归档标记目录，不含 spec 文件），全部迁移到新结构。

### Innovation Highlights

从"规范先行"到"实践先行"。Surface 优先使各测试类型用最适工具（API → Vitest，Web → Playwright）。构建检查独立为 infra/，避免与功能测试混放。

### Developer Observable Behavior

**API Before/After**：`from '@playwright/test'` + 硬编码 URL → `from 'vitest'` + `curl()` + `expect(res.status).toBe(200)` + TC-XXX。
**Web Before/After**：`waitForTimeout(2000)` + CSS → `toBeVisible()` + `getByTestId()` + `login(page)` + `.serial`。

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
| 6 | 使用 Playwright 作为 API 测试运行器 | 用 Vitest — 无浏览器开销，无需启动浏览器进程 |
| 7 | 无测试数据清理 — afterAll/afterEach 中不删除创建的测试资源 | 在 `afterAll` 中删除 `beforeAll` 创建的临时数据，保持测试幂等 |
| 8 | 混合断言风格 — 同一文件中混用 `expect().toBe()` 和 `===` 比较 | 统一使用 `expect()` 断言，不做裸比较 |

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
| 6 | cache-based 登录绕过 — 直接操作 `page.context().addCookies()` 或 storageState 文件绕过认证流程 | 使用 `login(page)` localStorage 注入，保持认证路径可审计 |
| 7 | 无测试数据清理 — Web 测试创建的 UI 数据在 afterAll 中未删除 | `afterAll` 中调用 API 删除测试数据，避免数据累积影响后续运行 |

#### Helpers 工具库规范（index.md）

| ID | 规则 | 说明 |
|---|---|---|
| **E-120** | 配置来源 — `config.yaml` + 环境变量 | 用 `baseUrl`/`apiBaseUrl` 常量，通过 `E2E_CONFIG_PATH` 覆盖，不硬编码 URL |
| **E-121** | Token 缓存 — `getApiToken()` / `getAuthToken()` | 自动缓存 23h TTL，不手动管理 token 生命周期 |
| **E-122** | 子进程执行 — `runCli(cmd, cwd, timeout)` | 统一包装 `execSync`，返回 `{stdout, stderr, exitCode}`。infra 测试使用 |
| **E-123** | 调试截图 — `screenshot(page, tcId)` | 保存到 `results/screenshots/`，不在测试中拼接路径 |

## Requirements Analysis

### Key Scenarios

- AI Agent 通过 `/gen-test-scripts` 生成新测试 → 读取 conventions 生成符合风格的代码
- 开发者写新测试 → 查阅 conventions 快速了解 helpers 用法和项目约定
- Code review → 以 conventions 中的反模式清单为检查标准
- CI 运行 → 按 surface 独立触发（api 启动后端，web 启动全栈），CI 适配在范围内
- 跨 surface 测试（如 `full-e2e.spec.ts`）归 `web/smoke/`

### Constraints & Dependencies

- 框架文档有 `<!-- auto-generated by forge:test-guide -->` 标记，追加内容不应冲突
- 文件 frontmatter `domains` 需更新
- API 迁到 Vitest 后 `test`/`expect` API 几乎一致，但 `.serial` 无直接等价（用 `test.sequential`）
- 3 文件用 `request.newContext`，需改写为 `curl()`
- 20 文件无 surface 后缀，审计时按分类规则判定
- 团队技术栈以 TypeScript 为主，测试工具选型限定 TS 生态

## Alternatives & Industry Benchmarking

### 目录结构方案

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| 维持现状（tests/e2e/ 混放） | 零迁移成本 | Surface 无法独立运行 | Rejected |
| Journey 优先（tests/\<journey\>/api/ + web/） | 同功能测试相邻 | 嵌套过深 | Rejected |
| **Surface 优先** | 运行器隔离 + 共享工具层，各 surface 独立选工具 | shared helpers 通过 `../shared/` 耦合——`shared/helpers.ts` 变更可能级联影响多 surface，但其纯 Node 无状态特性使变更频率低 | **Selected** |

### Industry Solutions

**a) Playwright projects pattern**：`projects[]` 按目标定义独立测试 profile（各自 `testDir`/`timeout`/`retries`）。本提案采纳类似分离但更进一步——用完全独立运行器（API → Vitest，Web → Playwright）而非同一运行器不同配置，彻底消除 API 测试浏览器依赖。

**b) Vitest workspace pattern**：`workspace` 支持根配置下挂载多个子 vitest.config。我们选择独立目录而非 workspace，因 surface 运行时依赖根本不同（浏览器 vs 纯 Node），workspace 会引入不必要耦合。

**c) Django/Next.js test layering**：Django 按 unit/functional/selenium 分层，Next.js 分离 `__tests__/` 与 `e2e/`。两者验证 surface-first 是行业通行做法。

### API 测试框架方案

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| Playwright Test（现状） | 统一运行器 | 加载浏览器模块 | Rejected |
| Playwright request API | 改进 HTTP 客户端 | 仍依赖 Playwright 运行器；已有 curl() | Rejected: 不解决核心问题 |
| Go httptest | 与后端同语言 | 丢失 TS helpers；团队技术栈以 TypeScript 为主（见 Constraints），重写成本高 | Rejected |
| **Vitest + 原生 fetch** | 已有 Vitest，启动快 | 需迁移 import | **Selected** |

## Feasibility Assessment

### Technical Feasibility

- 目录重组：文件移动 + import 路径调整
- API 迁移：`from '@playwright/test'` → `from 'vitest'`，API 几乎一致
- `test.sequential` 替代 `.serial`
- 性能：当前 API 测试套件在 Playwright worker=1 模式下运行约 45s（含浏览器模块加载 ~3s）。迁移 Vitest 后预估 ~20s（消除浏览器加载 + 纯 Node 启动）。正式数据待迁移后 benchmark

### Resource & Timeline

两部分工作量：
1. Surface 重组 + API 迁移：约 3-5 小时
   - **迁移前审计**（~30 min）：识别 fixture/`.serial`/浏览器依赖，产出清单
   - 目录重组 + import 调整：~2-3 小时
   - CI 适配：~1 小时
2. 实践规范文档更新：约 1 小时

**迁移顺序与过渡兼容**：

按 surface 逐个迁移，每个 surface 迁移完成后 CI 仍能运行全部测试：

1. **Phase 1 — shared/**：移动 helpers/config 到 `tests/shared/`，旧路径保留 re-export shim
2. **Phase 2 — api/**：迁移 API 测试到 `tests/api/` + Vitest，更新 CI，删除旧 API 文件
3. **Phase 3 — web/**：移动 Web 测试到 `tests/web/`，更新 CI，删除旧 Web 文件
4. **Phase 4 — infra/**：移动构建/lint 检查到 `tests/infra/` + Vitest，删除 shim 和空的 `tests/e2e/`
5. **Phase 5 — conventions**：从迁移后代码提炼规范

## Assumptions Challenged

| Assumption | Challenge Tool | Finding |
|------------|---------------|---------|
| 现有框架文档足够指导测试编写 | Stress Test: Agent 基于现有文档生成测试，对比实际代码 | Confirmed: 生成结果与实际风格差距大 |
| API 测试必须用 Playwright（因为 Web 测试用 Playwright） | Assumption Flip: Surface 分离后各 surface 可独立选择工具 | Overturned: Surface 优先 = 框架自由选择 |
| 测试目录按 journey 组织是唯一方式 | XY Detection: 真正的需求是"能找到相关测试"，不是"按功能分目录" | Refined: Surface 优先更符合 Forge 生命周期模型 |

## Scope

### In Scope

- Surface 优先目录重组：`tests/e2e/` → `tests/api/` + `tests/web/` + `tests/infra/` + `tests/shared/`
- API 测试从 Playwright 迁移到 Vitest
- 各 surface 的 package.json 创建和依赖拆分（shared deps 通过 workspace 引用）
- 更新 `docs/conventions/testing/api/core.md`：追加 API 实践规范（E-100~E-106）+ 反模式清单
- 更新 `docs/conventions/testing/web/core.md`：追加 Web 实践规范（E-110~E-115）+ 反模式清单
- 更新 `docs/conventions/testing/index.md`：追加 helpers 规范（E-120~E-123）
- 更新各文件 `domains` frontmatter

**Token 成本**：新增 ~2000 字，按需加载，单 surface ~600 字。

### Out of Scope

- 修复现有测试中的反模式代码（仅记录规范，不修改逻辑）
- 修改 helpers.ts 实现（仅移动位置）
- 修改 Forge 配置
- CI 全面重构。CI 适配仅含：(a) 按 surface 分 job，(b) 更新 glob，(c) 移除旧 job

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| 迁移引入 regression | M | H | 逐 surface 迁移并验证，每步确认测试通过 |
| Vitest/Playwright API 差异（如 .serial） | L | M | 参考兼容性矩阵逐文件检查，关注 `.serial` 语义 |
| CI 中断（目录变更致 glob 失效） | H | H | **CI 过渡**：同 PR。(1) 新旧 job 并行；(2) 逐 surface 更新 glob；(3) 完成后移除旧 job |
| Spec drift | M | M | 先迁移再提炼；CI grep 防回退，加 `last-verified` |
| 内容重叠 | L | L | 实践规范只写工具/方法，不重复策略 |
| Phase 失败需回退 | L | H | 每 Phase 独立分支+PR，失败 revert |
| 文件误分类（20 无后缀文件归错 surface） | M | M | 审计标注，PR review 确认 |
| shared helpers API 变更导致多 surface 同时失败 | L | H | shared 只导出纯函数（无状态），接口稳定性通过 TypeScript 类型检查保障 |

## Success Criteria

- [ ] `tests/` 按 surface 优先结构组织（api/ web/ infra/ shared/）
- [ ] 每个 surface 目录有独立 package.json，shared/ 通过 npm workspace 或相对路径引用
- [ ] API 测试全部用 Vitest，无 Playwright 依赖
- [ ] Web 测试保留 Playwright，配置独立
- [ ] 所有迁移后测试通过
- [ ] api/core.md 含 7 条规范（E-100~E-106）+ 8 条反模式
- [ ] web/core.md 含 6 条规范（E-110~E-115）+ 7 条反模式
- [ ] index.md 含 4 条 helpers 规范（E-120~E-123）
- [ ] 所有文件 `domains` frontmatter 已更新
- [ ] `tests/shared/` 创建完成，无 Playwright import（`grep '@playwright/test' tests/shared/` 输出为空）
- [ ] 迁移完成后 re-export shim 和 `tests/e2e/` 已移除
- [ ] Infra 测试已迁至 `tests/infra/`（原 `-cli` 后缀改为 `-build`）
- [ ] 迁移前审计完成，产出文件清单（含 surface 归属判定）
- [ ] CI 有独立 api/web/infra job，旧 job 已移除

## Next Steps

- 直接实施（doc + refactor 类型，无需 PRD/tech-design）
- 实施顺序：迁移前审计 → 按 Phase 1-4 逐 surface 迁移（每步保留向后兼容 shim） → Phase 5 从最终代码提炼 conventions
