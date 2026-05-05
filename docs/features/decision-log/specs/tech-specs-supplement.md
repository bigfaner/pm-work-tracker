---
feature: decision-log
source: lessons/arch-button-icon-convention.md
generated: "2026-05-04"
status: draft
---

# Technical Specifications Supplement: Button Icon Convention

## UI Component Conventions

### TECH-012: Action Buttons Must Use lucide-react Icon + Text

**Requirement**: All action buttons (add, edit, delete, refresh, retry, etc.) must include a lucide-react icon alongside text. Never use plain text or inline `<svg>`.

**Scope**: [CROSS]

**Source**: lessons/arch-button-icon-convention.md

**Icon mapping**:

| Action | Icon Component | Import |
|--------|---------------|--------|
| 添加 / 新增 | `Plus` | `import { Plus } from "lucide-react"` |
| 编辑 | `Pencil` | `import { Pencil } from "lucide-react"` |
| 刷新 / 重试 | `RefreshCw` | `import { RefreshCw } from "lucide-react"` |
| 删除 | `Trash2` | `import { Trash2 } from "lucide-react"` |
| 查看 / 预览 | `Eye` | `import { Eye } from "lucide-react"` |
| 上传 / 提升 | `ArrowUpCircle` | `import { ArrowUpCircle } from "lucide-react"` |
| 下移 / 降级 | `ArrowDownCircle` | `import { ArrowDownCircle } from "lucide-react"` |
| 驳回 / 取消 | `XCircle` | `import { XCircle } from "lucide-react"` |

**Size conventions**:
- `size={14}` for header/default buttons
- `size={12}` for small/ghost row actions
- `className="w-4 h-4"` as alternative sizing syntax

**Examples**:

```tsx
// ❌ Wrong — plain text
<Button variant="primary" size="sm" onClick={onAdd}>添加决策</Button>

// ❌ Wrong — inline SVG (legacy)
<Button onClick={onEdit}><svg width="16" height="16">...</svg>编辑</Button>

// ✅ Correct — lucide-react icon + text
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
