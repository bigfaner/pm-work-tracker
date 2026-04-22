# Frontend 测试命令

日期: 2026-04-22

## 问题

运行前端测试时出现错误：

```
Error: Expected a single value for option "--run", received [true, true]
```

## 根因

`frontend/package.json` 的 test script 已经包含 `--run`：

```json
"test": "vitest --run"
```

若使用 `npm run test -- --run`，npm 会将 `--run` 追加到 script 末尾，变成 `vitest --run --run`，vitest 报错。

## 正确用法

```bash
# ✅ 正确
npm test
npx vitest run

# ❌ 错误
npm run test -- --run
npm test -- --run
```
