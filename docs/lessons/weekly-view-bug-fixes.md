# 每周进展视图 Bug 修复记录

日期: 2026-04-20

## Bug 1: 手动选择周次后下方无数据

### 现象
点击右上角周选择器切换周次后，页面不展示任何数据。

### 根因
前端 `handleWeekChange` 中 ISO 周号转 Monday 日期的计算有误，使用简易算法产生了 Sunday 而非 Monday。后端校验 `weekStart` 必须为 Monday，导致请求被拒绝。

### 修复
改用 ISO 8601 标准的 Jan 4 参考点法计算周一日期：

```typescript
const jan4 = new Date(year, 0, 4);
const dow = jan4.getDay() || 7; // Mon=1...Sun=7
const mondayW1 = new Date(year, 0, 4 - dow + 1);
const target = new Date(mondayW1);
target.setDate(mondayW1.getDate() + (week - 1) * 7);
```

文件: `frontend/src/pages/WeeklyViewPage.tsx`

---

## Bug 2: 所有周次都显示相同数据

### 现象
无论选择哪一周，下方始终展示全量数据，未按周次过滤。

### 根因
后端 `WeeklyComparison` 在判断子事项是否活跃时，仅排除"已完成"状态的子事项，未检查子事项的 `CreatedAt` 和 `ActualEndDate` 是否与所选周次有交集。

### 修复
新增 `isActiveInWeek` 函数，根据周次范围过滤：

- `CreatedAt` 在所选周之后 → 不活跃
- `ActualEndDate` 在所选周之前 → 不活跃
- 其余情况 → 活跃

同时过滤掉无活跃子事项的空主事项分组。

文件: `backend/internal/service/view_service.go`

---

## Bug 3: 后端拒绝当前周为"未来周"

### 现象
选择当前周时 API 返回 `FUTURE_WEEK_NOT_ALLOWED` 错误。

### 根因
Go 的 `time.Parse("2006-01-02", ...)` 返回 UTC 时区，而 `time.Now()` 使用本地时区（UTC+8）。在 UTC+8 的周一，当前周一的 UTC 时间 = 周日 16:00 UTC，`weekStart.After(today)` 为 true，被误判为未来周。

### 修复
`time.Parse` 后立即将日期规范化为本地时区：

```go
loc := time.Now().Location()
weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, loc)
```

文件: `backend/internal/handler/view_handler.go`

---

## Bug 4: E2E 测试 `toISOString` 跨时区日期偏移

### 现象
E2E 测试中用 `monday.toISOString().slice(0, 10)` 构造 `weekStart` 参数，在 UTC+8 时区下得到比本地日期少一天的日期。

### 根因
`toISOString()` 返回 UTC 时间字符串。在 UTC+8 的凌晨/上午，UTC 日期比本地日期早一天。

### 修复
使用本地日期格式化：

```typescript
const weekStart = `${monday.getFullYear()}-${String(monday.getMonth() + 1).padStart(2, '0')}-${String(monday.getDate()).padStart(2, '0')}`;
```

文件: `frontend/__tests__/e2e/weekly-view.spec.ts`

---

## Bug 5: E2E `beforeEach` UTC 周计算与本地时间不一致

### 现象
E2E 测试 3.1 ~ 5.4 在 UTC+8 时区的周一失败——测试数据不可见。

### 根因
`beforeEach` 使用 `getUTCDay()` / `setUTCDate()` 计算当前周一。在 UTC+8 的周一，UTC 时间仍为周日，导致计算出上一周的 Monday，将页面周选择器设为错误的周次。测试数据创建于本周（本地时间），在上一周的 API 响应中不存在。

### 修复
移除 `beforeEach` 中的手动周次设置，让页面使用默认的当前周（前端组件用本地时间计算，始终正确）。

文件: `frontend/__tests__/e2e/weekly-view.spec.ts`

---

## Bug 6: E2E 测试 Playwright strict mode 违规

### 现象
多次运行 E2E 后，`text=E2E周视图测试主事项` 匹配到多个 DOM 元素，Playwright strict mode 报错。

### 根因
`afterAll` 归档测试数据是 best-effort，失败的测试运行会留下未清理的测试数据。多次运行后同名主事项累积，全局 text locator 匹配到多个元素。

### 修复
将所有测试数据的 locator 限定在 `[data-testid="group-card-${testMainItemId}"]` 范围内，避免匹配到其他测试运行的数据。

文件: `frontend/__tests__/e2e/weekly-view.spec.ts`

---

## 教训总结

| 类别 | 要点 |
|------|------|
| 时区 | Go `time.Parse` 产出 UTC；JS `toISOString()` 产出 UTC。与本地时间比较前必须统一时区。 |
| ISO 周计算 | 使用 Jan 4 参考点法，避免简易算法在年末/年初出错。 |
| E2E 周选择器 | `<input type="week">` 的 `fill()` 不可靠；用 `evaluate()` + native value setter。 |
| E2E 数据隔离 | 用 testid 限定 locator 作用域，不依赖全局 text 匹配，避免累积数据的干扰。 |
| 后端过滤 | 时间范围过滤应基于 `CreatedAt`/`ActualEndDate`，不能只看 `status` 字段。 |
