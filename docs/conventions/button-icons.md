---
scope: frontend
source: feature/decision-log TECH-012
---

# Button Icon Convention

All action buttons must use lucide-react icon + text. No plain text, no inline `<svg>`.

## Icon Mapping

| Action | Icon | Import |
|--------|------|--------|
| 添加 / 新增 | Plus | `import { Plus } from "lucide-react"` |
| 编辑 | Pencil | `import { Pencil } from "lucide-react"` |
| 刷新 / 重试 | RefreshCw | `import { RefreshCw } from "lucide-react"` |
| 删除 | Trash2 | `import { Trash2 } from "lucide-react"` |
| 查看 / 预览 | Eye | `import { Eye } from "lucide-react"` |
| 上传 / 提升 | ArrowUpCircle | `import { ArrowUpCircle } from "lucide-react"` |
| 下移 / 降级 | ArrowDownCircle | `import { ArrowDownCircle } from "lucide-react"` |
| 驳回 / 取消 | XCircle | `import { XCircle } from "lucide-react"` |

## Size Conventions

- `size={14}` — header / default buttons
- `size={12}` — small / ghost row actions
- `className="w-4 h-4"` — alternative sizing syntax

## Examples

```tsx
// ❌ Wrong — plain text
<Button onClick={onAdd}>添加决策</Button>

// ❌ Wrong — inline SVG (legacy)
<Button onClick={onEdit}><svg width="16" height="16">...</svg>编辑</Button>

// ✅ Correct
import { Plus, Pencil } from "lucide-react";

<Button variant="primary" size="sm" onClick={onAdd}>
  <Plus size={14} />
  添加决策
</Button>

<Button variant="ghost" size="sm" onClick={() => onEdit(id)}>
  <Pencil size={12} />
  编辑
</Button>
```
