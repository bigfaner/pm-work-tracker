---
date: "2026-04-28"
doc_dir: "docs/features/permission-granularity/prd/"
iteration: "3"
target_score: ""
evaluator: Claude (automated, adversarial)
---

# PRD Eval — Iteration 3

**Score: 96/100** (target: —)

```
┌─────────────────────────────────────────────────────────────────┐
│                       PRD QUALITY SCORECARD                      │
├──────────────────────────────┬──────────┬──────────┬────────────┤
│ Dimension                    │ Score    │ Max      │ Status     │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 1. Background & Goals        │  19      │  20      │ ⚠️          │
│    Background three elements │   7/7    │          │            │
│    Goals quantified          │   7/7    │          │            │
│    Logical consistency       │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 2. Flow Diagrams             │  20      │  20      │ ✅          │
│    Mermaid diagram exists    │   7/7    │          │            │
│    Main path complete        │   7/7    │          │            │
│    Decision + error branches │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 3. Functional Specs          │  19      │  20      │ ⚠️          │
│    Tables complete           │   7/7    │          │            │
│    Field descriptions clear  │   7/7    │          │            │
│    Validation rules explicit │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 4. User Stories              │  20      │  20      │ ✅          │
│    Coverage per user type    │   7/7    │          │            │
│    Format correct            │   7/7    │          │            │
│    AC per story              │   6/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ 5. Scope Clarity             │  18      │  20      │ ⚠️          │
│    In-scope concrete         │   6/7    │          │            │
│    Out-of-scope explicit     │   7/7    │          │            │
│    Consistent with specs     │   5/6    │          │            │
├──────────────────────────────┼──────────┼──────────┼────────────┤
│ TOTAL                        │  96      │  100     │            │
└──────────────────────────────┴──────────┴──────────┴────────────┘
```

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| prd-spec.md §需求目标 | Goal "新增至少 4 个针对 role:* 权限码的路由中间件测试" has no corresponding Scope In item and no §5.5 entry — the deliverable is promised in the goal but never scoped | -1 pt |
| prd-ui-functions.md §UI Function 3 Validation Rules | No validation rule for `permissionCodes` when `GET /admin/permissions` fails at form-open time — Form Spec notes "若接口返回空，显示'暂无可用权限码'，多选框禁用" but this is in the Form Spec notes, not the Validation Rules section, and covers only empty-array, not request failure | -1 pt |
| prd-ui-functions.md §UI Function 3 States | States table lists `create-visible/hidden`, `edit-visible`, `delete-visible` but omits `edit-hidden` and `delete-hidden` — the table is asymmetric; a developer cannot derive the hidden behavior for edit/delete from the table alone | -1 pt |
| prd-spec.md §Scope In | "member 角色不变" in Scope In conflicts with §5.3 "管理员可按需单独授予 `role:read`" and Flow A `AdminGrant[管理员单独授予 role:read]` — the manual-grant capability is a behavioral change for member but Scope In declares member unchanged | -1 pt |

---

## Attack Points

### Attack 1: Background & Goals — test-coverage goal is unscoped

**Where**: `prd-spec.md §需求目标` — "测试覆盖 | 新增至少 4 个针对 role:* 权限码的路由中间件测试 | 覆盖 read/create/update/delete，验证 403". Neither the Scope In checklist nor §5.5 关联性需求改动 lists this as a deliverable. §5.5 row 4 covers only "VerifyPresetRoleCodes 测试" (updating an existing test), not new middleware tests.

**Why it's weak**: A quantified goal that has no corresponding scope item is unverifiable at delivery. The goal says "新增至少 4 个", implying net-new test functions. But there is no scope item that assigns this work, no file or package named, and no AC in any user story that validates it. A developer reading the spec has no signal that these tests are their responsibility. If the goal is used as a release checklist, this item will be missed.

**What must improve**: Add a Scope In item: "新增至少 4 个路由中间件测试，覆盖 `role:read/create/update/delete`，验证无权限时返回 403（测试文件：`backend/internal/handler/role_handler_test.go`）". Alternatively, add it as row 6 in §5.5 with a concrete test target.

---

### Attack 2: Functional Specs — UI Function 3 States table is asymmetric

**Where**: `prd-ui-functions.md §UI Function 3 States` — the table lists `edit-visible | 显示行内"编辑"按钮 | 用户持有 role:update` and `delete-visible | 显示行内"删除"按钮 | 用户持有 role:delete`, but there are no corresponding `edit-hidden` or `delete-hidden` rows. Compare with UI Function 2 which has both `visible` and `hidden` states.

**Why it's weak**: A frontend engineer implementing the role management table must know what to render when the user lacks `role:update` or `role:delete`. The `create-hidden` state is documented ("不显示'创建角色'按钮"), establishing the pattern. The omission of `edit-hidden` and `delete-hidden` forces the developer to infer behavior — they may render a disabled button, hide it entirely, or show a tooltip, all of which are different UX outcomes. The Validation Rules section adds a special case ("预置角色的'删除'按钮始终禁用") that interacts with the hidden/disabled distinction, making the gap more consequential.

**What must improve**: Add `edit-hidden | 不显示行内"编辑"按钮 | 用户不持有 role:update` and `delete-hidden | 不显示行内"删除"按钮 | 用户不持有 role:delete` to the States table. Clarify whether "hidden" means `display:none` or `disabled` — the Validation Rules section uses "禁用" for the preset-role case, so the distinction matters.

---

### Attack 3: Scope Clarity — "member 角色不变" contradicts Flow A and §5.3

**Where**: `prd-spec.md §Scope In` — "member 角色：默认不含 user/role 权限。管理员可按需单独授予 `role:read`" (§5.3). `prd-spec.md §业务流程图 Flow A` — `AdminGrant[管理员单独授予 role:read] --> CheckPerm`. `prd-spec.md §Scope In checklist` — "migration/rbac.go 预置角色权限码同步：pm 角色新增 user:list/user:read/user:assign_role/role:*，**member 角色不变**".

**Why it's weak**: The Scope In item says member is unchanged, but the feature's primary motivation (Flow A, §需求背景 item 1) is enabling member to access the role dropdown. The resolution — that member gets `role:read` via manual admin grant rather than preset — is a meaningful design decision that is buried in §5.3 prose and never surfaced in Scope. A reader scanning the Scope checklist will conclude member's permissions are untouched and miss that the feature requires an admin action post-deploy to enable the member use case. This also means the feature's success criterion (member can complete the add-member flow) is not automatically satisfied by the migration — it depends on a manual step that is not in any Scope item or AC.

**What must improve**: Update the Scope In item to: "member 角色：预置权限码不变；feature 上线后需管理员手动授予 `role:read` 以启用团队管理页角色下拉功能（非自动迁移）". Add an AC to Story 1 or Story 5 that validates this post-deploy step.

---

## Previous Issues Check

| Issue from Iteration 2 | Status | Evidence |
|------------------------|--------|----------|
| Route count in goal said "10 条" but §5.2 has 14 rows | ✅ Fixed | §需求目标 now reads "14 条受影响路由" |
| Flow A and Flow B had no API error branches | ✅ Fixed | Both flows now have `|失败 5xx/超时|` branches leading to error states |
| `POST /admin/users` labeled "给用户分配角色" with no explanation | ✅ Fixed | §5.2 now has a footnote explaining the endpoint semantics |
| `DELETE /admin/users/:userId` maps to `user:update` with no rationale | ✅ Fixed | §5.2 now has a note: "软删除（status 置为 disabled），属于用户状态变更操作，语义上归属 user:update" |
| `description` overflow had no validation rule | ✅ Fixed | Validation Rules now includes "`description` 超过 200 字符时，输入框下方显示'角色描述不能超过 200 个字符'，提交按钮禁用" |
| Member search empty-state had no validation rule | ✅ Fixed | Validation Rules now includes "成员搜索无结果：搜索框下方显示'未找到匹配成员'，`memberId` 保持未选中，提交按钮禁用" |
| `VerifyPresetRoleCodes` test update absent from Scope In | ❌ Not fixed | §5.5 row 4 covers it but Scope In checklist still does not list it |
| "member 角色不变" vs §5.3 manual-grant ambiguity | ❌ Not fixed | §5.3 still reads "管理员可按需单独授予 role:read" while Scope In says "member 角色不变" — distinction between preset and runtime grant remains unstated in Scope |

---

## Verdict

- **Score**: 96/100
- **Target**: —/100
- **Gap**: —
- **Action**: Strong improvement from iteration 2 (90→96). All 6 iteration-2 attack points addressed. Remaining gaps are narrow: test-coverage goal unscoped (-1), UI Function 3 States table asymmetric (-1), `GET /admin/permissions` failure path in Validation Rules (-1), member/scope ambiguity (-1). Two of the four remaining issues were carried from iteration 2 without resolution (VerifyPresetRoleCodes scope item, member-role ambiguity).
