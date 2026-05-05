---
scope: backend, database
source: feature/item-code-redesign
verified: "2026-05-04"
---

# Item Code Conventions

## BIZ-code-001: Team Code Format

Team code is 2-6 alphabetic characters, globally unique, and **immutable after creation**. Creation is the only time a team code is set; there is no edit/update path.

The backend validates format via regex `^[A-Za-z]{2,6}$` (enforced by Gin `alpha,min=2,max=6` binding tags) and uniqueness via a database unique index (`idx_teams_code`).

**Why**: Team code serves as the human-readable prefix for all item codes under that team. Mutability would cascade changes to thousands of item codes, breaking external references (reports, chat messages, weekly summaries).

**Example**:
- Valid: `FEAT`, `ENG`, `QA`
- Invalid: `F` (too short), `FEATURE` (too long), `F1` (non-alpha)

## BIZ-code-002: Main Item Code Format

Main item code follows `{TEAM_CODE}-NNNNN` — 5-digit zero-padded sequence number prefixed by the team code of the owning team.

The code is **snapshotted** at creation time from the team's current `Code` field. If a team code were ever changed (not currently supported), existing items would retain their original code.

**Why**: 5 digits provide 99,999 unique codes per team, sufficient for internal tooling. Zero-padding ensures fixed-width codes that sort lexicographically.

**Example**: `FEAT-00001`, `FEAT-00042`, `ENG-00100`

## BIZ-code-003: Sub Item Code Format

Sub item code follows `{MAIN_ITEM_CODE}-NN` — 2-digit zero-padded sub-sequence appended to the parent main item's code.

**Why**: 2 digits provide 99 sub-items per main item. The hierarchical format makes parent-child relationships immediately visible.

**Example**: `FEAT-00001-01`, `FEAT-00001-02`, `ENG-00100-99`

## BIZ-code-004: Atomic UPDATE Counter Code Generation

`NextCode()` and `NextSubCode()` use atomic UPDATE counter increment within a database transaction:

- `NextCode(teamBizKey)` increments `item_seq` on the team row (`UPDATE pmw_teams SET item_seq = item_seq + 1 WHERE biz_key = ?`), then reads the team to get the new sequence value. A safety check ensures the generated code is greater than any existing MAX sequence in `pmw_main_items` for that team.
- `NextSubCode(mainItemBizKey)` increments `sub_item_seq` on the main item row (`UPDATE pmw_main_items SET sub_item_seq = sub_item_seq + 1 WHERE biz_key = ?`), then reads the main item to get the new sequence value. A safety check ensures the generated code is greater than any existing MAX sub-sequence in `pmw_sub_items` for that main item.

SQLite note: The real write acquires SQLite write lock, serializing concurrent calls.

**Why**: Atomic UPDATE eliminates race conditions without requiring `SELECT FOR UPDATE`. The counter lives on the parent row (team or main_item), so different teams/parents have zero contention.

**Example**:
```go
// Main item: atomically increment counter, then read
tx.Exec("UPDATE pmw_teams SET item_seq = item_seq + 1 WHERE biz_key = ?", teamBizKey)
tx.Where("biz_key = ?", teamBizKey).First(&team)
code = fmt.Sprintf("%s-%05d", team.Code, team.ItemSeq)

// Sub item: atomically increment counter, then read
tx.Exec("UPDATE pmw_main_items SET sub_item_seq = sub_item_seq + 1 WHERE biz_key = ?", mainItemBizKey)
tx.Where("biz_key = ?", mainItemBizKey).First(&mainItem)
code = fmt.Sprintf("%s-%02d", mainItem.Code, mainItem.SubItemSeq)
```

## BIZ-code-005: Code Snapshot Immutability

Team codes are snapshotted into item codes at creation time. The item code string is stored in the database and never recomputed from the team code. This means:

- A team code change (if ever implemented) would **not** affect existing item codes.
- Item codes are self-contained strings that can be safely used as stable references in reports, chat, and weekly summaries.

**Why**: Prevents cascading updates and broken external references. Items created at different times may reflect different team codes if the team code changed in between — by design.

## BIZ-code-006: SubItem.Code Unique Index Scope

SubItem.Code unique index is **per-main-item** (composite: `main_item_id + code`), not global. Two sub items under different main items may share the same `code` suffix (e.g., both `FEAT-00001-01` and `ENG-00002-01` are valid).

The index name is `idx_sub_items_main_code` on `(main_item_id, code)`.

**Why**: Sub-item codes are only meaningful in the context of their parent main item. A global unique constraint would force artificial coordination between unrelated main items.
