# E2E 测试脚本生成：必须先读代码再写测试

## Problem

生成的 e2e 测试脚本存在大量基础错误，导致需要多轮修复才能运行：
- 相对路径计算错误（多了一层 `../`）
- API 路径前缀假设错误（`/api/v1/` vs 实际的 `/v1/`）
- 端口号与实际运行服务不符
- 路由名称错误（`/items` vs `/main-items`，`/members/invite` vs `/members`）
- 测试用例引用了不存在的路由
- 测试依赖硬编码的预置数据，但数据库中没有这些数据

## Root Cause

脚本是基于"推断"而非"读取"生成的：

1. **没有读 router.go** → 路由名称、路径前缀全靠猜
2. **没有读 config.yaml** → 端口号、base_path 全靠猜
3. **没有验证相对路径** → `__dirname` 到 backend 的层级算错
4. **没有检查服务端是否实现了对应校验** → 测试断言了未实现的行为（TC-004 的 `bizKey <= 0` 返回 400）
5. **测试数据依赖外部状态** → 硬编码 bizKey 假设数据库中存在特定记录

根本原因：**生成测试脚本时跳过了"读代码"步骤，直接根据需求文档推断实现细节**。

## Solution

生成 e2e 测试脚本前，必须先完成以下核实步骤：

```
1. 读 router.go → 确认所有路由的实际路径和 HTTP 方法
2. 读 config.yaml → 确认端口、base_path
3. 计算并验证相对路径（用 node -e 验证 resolve 结果）
4. 读中间件代码 → 确认哪些校验已实现，哪些没有
5. 检查测试数据依赖 → 需要预置数据的用例必须在测试内自行创建
```

## Key Takeaway

**生成 e2e 测试脚本的正确顺序：先读代码，再写测试。**

- 路由路径：从 `router.go` 读，不要猜
- 端口/前缀：从 `config.yaml` 读，不要猜
- 文件路径：用 `node -e "console.log(resolve(...))"` 验证，不要数 `../`
- 服务端行为：从中间件/handler 代码读，不要假设
- 测试数据：测试内自行创建（setup/teardown），不依赖外部状态

**反模式**：根据 PRD/设计文档推断实现细节 → 测试与实际代码脱节。

---

## 补充发现：UI 测试 testid 未验证导致全量超时（2026-05-14）

### 问题

24 个 UI e2e 测试全部超时（1-2 分钟/个），总耗时 16 分钟。根本原因：测试使用 `[data-testid="map-card"]`，但前端实际写的是 `[data-testid="map-card-${bizKey}"]`。

### 为什么超时而不是立即失败

Playwright 的 `locator()` 使用 auto-wait 机制：元素不存在时不报错，而是持续轮询 DOM 直到 timeout。所以 testid 不匹配时，每个测试都等满 60-120 秒超时才 fail。24 个超时测试 × 平均 60 秒 ÷ 3 个 worker ≈ 8 分钟纯等死。

### 根因

生成 UI 测试脚本时，**没有 grep 前端源码确认实际的 testid 值**。agent 根据测试用例文字描述猜测了 `map-card`，但前端实现用了动态 testid。

### 新增规则

生成 UI 测试脚本时，必须在写 selector 之前：
```bash
# 1. grep 前端源码确认实际 testid
grep -rn "data-testid" frontend/src/ --include="*.tsx"
# 2. 如果是动态 testid（如 map-card-${id}），使用属性前缀选择器
page.locator('[data-testid^="map-card-"]').first()
# 3. 不要假设 testid 命名，从源码中读取
```
