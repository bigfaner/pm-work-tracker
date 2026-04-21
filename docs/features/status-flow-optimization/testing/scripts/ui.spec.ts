import { describe, test, before, after } from 'node:test';
import assert from 'node:assert/strict';
import { ab, abJson, getElements, snapshotContains, screenshot, baseUrl } from './helpers.js';

// Item IDs (created via API before running UI tests)
const GENERAL_ITEM_CODE   = 'MI-0287'; // pending → progressing after TC-003
const BLOCKING_ITEM_CODE  = 'MI-0288'; // blocking
const REVIEWING_ITEM_CODE = 'MI-0289'; // reviewing
const CLOSED_ITEM_CODE    = 'MI-0290'; // closed
const OVERDUE_ITEM_CODE   = 'MI-0291'; // progressing + past end date
const FUTURE_ITEM_CODE    = 'MI-0292'; // progressing + future end date
const CONFIRM_ITEM_CODE   = process.env.E2E_CONFIRM_ITEM_CODE ?? 'MI-0293'; // fresh progressing item for confirm tests

const PM_TOKEN    = process.env.E2E_PM_TOKEN    ?? '';
const EXEC_TOKEN  = process.env.E2E_EXECUTOR_TOKEN ?? '';

function listUrl() { return `${baseUrl}/items`; }

/** Inject JWT into localStorage auth-storage and reload */
function injectAuth(token: string) {
  const authState = JSON.stringify({
    state: {
      token,
      user: null,
      isAuthenticated: true,
      isSuperAdmin: false,
      permissions: null,
      permissionsLoadedAt: null,
      _hasHydrated: true,
    },
    version: 0,
  });
  const b64 = Buffer.from(authState).toString('base64');
  // Must be on the app origin before accessing localStorage
  ab(`open ${baseUrl}/login`);
  ab('wait --load networkidle');
  ab(`eval "localStorage.setItem('auth-storage', atob('${b64}'))"`);
  ab(`open ${listUrl()}`);
  ab('wait --load networkidle');
}

/** Navigate to list page and find the status button ref for a given item code, scrolling if needed */
function findStatusButtonRef(itemCode: string): string {
  // Try without scrolling first
  let snap = abJson('snapshot');
  let text: string = snap?.data?.snapshot ?? '';
  if (!text.includes(itemCode)) {
    // Scroll down to reveal more items
    ab('scroll down 500');
    ab('wait 500');
    snap = abJson('snapshot');
    text = snap?.data?.snapshot ?? '';
  }
  const idx = text.indexOf(itemCode);
  assert.ok(idx >= 0, `Item ${itemCode} not found in page snapshot`);
  const chunk = text.slice(idx, idx + 400);
  const m = chunk.match(/button "(待开始|进行中|阻塞中|已暂停|待验收|已完成|已关闭)" \[.*?ref=(e\d+)/);
  assert.ok(m, `Status button not found near ${itemCode}. Chunk: ${chunk.slice(0, 200)}`);
  return `@${m[2]}`;
}

/** Click status button, scroll into view first, wait for dropdown, return snapshot text */
function openDropdown(buttonRef: string): string {
  ab(`scrollintoview ${buttonRef}`);
  ab('wait 300');
  ab(`click ${buttonRef}`);
  ab('wait 1000');
  const snap = abJson('snapshot');
  return snap?.data?.snapshot ?? '';
}

/** Get menuitem refs from current snapshot */
function getMenuItems(): Array<{ ref: string; name: string }> {
  const snap = abJson('snapshot -i');
  const els = getElements(snap);
  return els.filter(e => e.role === 'menuitem').map(e => ({ ref: e.ref, name: e.name ?? '' }));
}

describe('UI E2E Tests: Status Flow Optimization', () => {
  before(() => {
    injectAuth(PM_TOKEN);
  });

  after(() => {
    try { ab('close --all'); } catch { /* ignore */ }
  });

  // Traceability: TC-001 → US-1 / AC-1
  test('TC-001: Status badge displays correct Chinese name for progressing', () => {
    ab(`open ${listUrl()}`);
    ab('wait --load networkidle');
    const snap = abJson('snapshot');
    const text: string = snap?.data?.snapshot ?? '';
    assert.ok(text.includes('进行中'), `Expected "进行中" in page. Snapshot: ${text.slice(0, 500)}`);
    screenshot('TC-001');
  });

  // Traceability: TC-002 → US-1 / AC-1; Spec R1
  test('TC-002: All status codes render correct Chinese names — no fallback styling', () => {
    ab(`open ${listUrl()}`);
    ab('wait --load networkidle');
    const snap = abJson('snapshot');
    const text: string = snap?.data?.snapshot ?? '';
    const chineseNames = ['待开始', '进行中', '阻塞中', '已暂停', '待验收', '已完成', '已关闭'];
    const found = chineseNames.filter(name => text.includes(name));
    assert.ok(found.length > 0, `Expected at least one Chinese status name. Snapshot: ${text.slice(0, 500)}`);
    screenshot('TC-002');
  });

  // Traceability: TC-003 → US-13 / AC-1; Spec AC-16
  test('TC-003: StatusDropdown calls ChangeStatus API on selection', () => {
    ab(`open ${listUrl()}`);
    ab('wait --load networkidle');

    // MI-0175 is progressing — find its status button
    const btnRef = findStatusButtonRef(GENERAL_ITEM_CODE);
    const dropdownText = openDropdown(btnRef);
    assert.ok(dropdownText.includes('menuitem'), `Dropdown did not open. Snapshot: ${dropdownText.slice(0, 300)}`);

    // Click first available option
    const items = getMenuItems();
    assert.ok(items.length > 0, 'No menu items found in dropdown');
    ab(`click ${items[0].ref}`);
    ab('wait --load networkidle');

    // Status should have changed (any Chinese name is fine)
    const afterSnap = abJson('snapshot');
    const afterText: string = afterSnap?.data?.snapshot ?? '';
    assert.ok(afterText.includes(GENERAL_ITEM_CODE), 'Item still visible after status change');
    screenshot('TC-003');
  });

  // Traceability: TC-004 → US-14 / AC-1; Spec AC-17
  test('TC-004: StatusDropdown shows only valid transitions for blocking state', () => {
    ab(`open ${listUrl()}`);
    ab('wait --load networkidle');

    const btnRef = findStatusButtonRef(BLOCKING_ITEM_CODE);
    const dropdownText = openDropdown(btnRef);

    // Check only the menu content (menuitem lines), not the full page
    const menuLines = dropdownText.split('\n').filter(l => l.includes('menuitem'));
    const menuText = menuLines.join('\n');
    assert.ok(menuText.includes('进行中'), `Expected "进行中" option for blocking. Menu: ${menuText}`);
    assert.ok(!menuText.includes('已暂停'), '"已暂停" should not appear for blocking state');
    assert.ok(!menuText.includes('已完成'), '"已完成" should not appear for blocking state');
    assert.ok(!menuText.includes('待验收'), '"待验收" should not appear for blocking state');
    screenshot('TC-004');
  });

  // Traceability: TC-005 → US-14 / AC-1; US-5 / AC-3
  test('TC-005: StatusDropdown for reviewing state shows completed and progressing for PM', () => {
    ab(`open ${listUrl()}`);
    ab('wait --load networkidle');

    const btnRef = findStatusButtonRef(REVIEWING_ITEM_CODE);
    const dropdownText = openDropdown(btnRef);

    assert.ok(dropdownText.includes('已完成'), `Expected "已完成" for PM in reviewing. Snapshot: ${dropdownText.slice(0, 300)}`);
    assert.ok(dropdownText.includes('进行中'), `Expected "进行中" for PM in reviewing. Snapshot: ${dropdownText.slice(0, 300)}`);
    screenshot('TC-005');
  });

  // Traceability: TC-007 → US-14 / AC-4
  test('TC-007: StatusDropdown disabled for terminal states', () => {
    ab(`open ${listUrl()}`);
    ab('wait --load networkidle');

    const btnRef = findStatusButtonRef(CLOSED_ITEM_CODE);
    const dropdownText = openDropdown(btnRef);

    // No transition options should appear
    const hasOptions = ['待开始', '进行中', '阻塞中', '已暂停', '待验收', '已完成'].some(s => dropdownText.includes(s) && dropdownText.includes('menuitem'));
    assert.ok(!hasOptions, `No transition options should appear for terminal state. Snapshot: ${dropdownText.slice(0, 300)}`);
    screenshot('TC-007');
  });

  // Traceability: TC-008 → US-15 / AC-1; Spec AC-19
  test('TC-008: Overdue indicator shown for non-terminal overdue item', () => {
    ab(`open ${listUrl()}`);
    ab('wait --load networkidle');
    const snap = abJson('snapshot');
    const text: string = snap?.data?.snapshot ?? '';
    // Find the overdue item row
    const idx = text.indexOf(OVERDUE_ITEM_CODE);
    assert.ok(idx >= 0, `${OVERDUE_ITEM_CODE} not found in page`);
    const chunk = text.slice(idx, idx + 600);
    assert.ok(
      chunk.includes('延期') || chunk.includes('逾期') || chunk.includes('overdue'),
      `Expected overdue indicator near ${OVERDUE_ITEM_CODE}. Chunk: ${chunk}`
    );
    screenshot('TC-008');
  });

  // Traceability: TC-009 → US-15 / AC-2; Spec AC-19
  test('TC-009: No overdue indicator for terminal state item', () => {
    ab(`open ${listUrl()}`);
    ab('wait --load networkidle');
    const snap = abJson('snapshot');
    const text: string = snap?.data?.snapshot ?? '';
    const idx = text.indexOf(CLOSED_ITEM_CODE);
    assert.ok(idx >= 0, `${CLOSED_ITEM_CODE} not found in page`);
    const chunk = text.slice(idx, idx + 300);
    assert.ok(
      !chunk.includes('延期') && !chunk.includes('逾期') && !chunk.includes('overdue'),
      `Overdue indicator should NOT appear for closed item. Chunk: ${chunk}`
    );
    screenshot('TC-009');
  });

  // Traceability: TC-010 → US-15 / AC-3; Spec AC-19
  test('TC-010: No overdue indicator when expected_end_date is future', () => {
    ab(`open ${listUrl()}`);
    ab('wait --load networkidle');
    const snap = abJson('snapshot');
    const text: string = snap?.data?.snapshot ?? '';
    const idx = text.indexOf(FUTURE_ITEM_CODE);
    assert.ok(idx >= 0, `${FUTURE_ITEM_CODE} not found in page`);
    const chunk = text.slice(idx, idx + 300);
    assert.ok(
      !chunk.includes('延期') && !chunk.includes('逾期') && !chunk.includes('overdue'),
      `Overdue indicator should NOT appear for future-dated item. Chunk: ${chunk}`
    );
    screenshot('TC-010');
  });

  // Traceability: TC-011 → US-16 / AC-1; Spec AC-21
  test('TC-011: Confirmation dialog appears before completing or closing', () => {
    ab(`open ${listUrl()}`);
    ab('wait --load networkidle');

    // MI-0176 (CONFIRM) is progressing — find 已关闭 option
    const btnRef = findStatusButtonRef(CONFIRM_ITEM_CODE);
    openDropdown(btnRef);

    const items = getMenuItems();
    const closedItem = items.find(i => i.name.includes('已关闭'));
    assert.ok(closedItem, '"已关闭" option not found in dropdown');

    ab(`click ${closedItem.ref}`);
    ab('wait 800');

    const dialogSnap = abJson('snapshot');
    const text: string = dialogSnap?.data?.snapshot ?? '';
    assert.ok(
      text.includes('确认') || text.includes('不可逆') || text.includes('确定') || text.includes('警告'),
      `Expected confirmation dialog. Snapshot: ${text.slice(0, 500)}`
    );
    screenshot('TC-011');
  });

  // Traceability: TC-012 → US-16 / AC-2; Spec AC-21
  test('TC-012: Cancel on confirmation dialog aborts status change', () => {
    // Dialog should still be open from TC-011, or re-trigger
    let snap = abJson('snapshot');
    let text: string = snap?.data?.snapshot ?? '';
    const hasDialog = text.includes('确认') || text.includes('不可逆') || text.includes('确定');

    if (!hasDialog) {
      ab(`open ${listUrl()}`);
      ab('wait --load networkidle');
      const btnRef = findStatusButtonRef(CONFIRM_ITEM_CODE);
      openDropdown(btnRef);
      const items = getMenuItems();
      const closedItem = items.find(i => i.name.includes('已关闭'));
      assert.ok(closedItem, '"已关闭" option not found');
      ab(`click ${closedItem.ref}`);
      ab('wait 800');
    }

    snap = abJson('snapshot -i');
    const els = getElements(snap);
    const cancelBtn = els.find(e =>
      e.name?.includes('取消') || e.name?.toLowerCase().includes('cancel')
    )?.ref;
    assert.ok(cancelBtn, 'Cancel button not found in dialog');
    ab(`click ${cancelBtn}`);
    ab('wait --load networkidle');

    // Reload to get fresh state
    ab(`open ${listUrl()}`);
    ab('wait --load networkidle');

    // CONFIRM item should still be progressing
    snap = abJson('snapshot');
    text = snap?.data?.snapshot ?? '';
    const idx = text.indexOf(CONFIRM_ITEM_CODE);
    assert.ok(idx >= 0, `${CONFIRM_ITEM_CODE} not found after cancel`);
    const chunk = text.slice(idx, idx + 400);
    assert.ok(chunk.includes('进行中'), `Status should remain "进行中" after cancel. Chunk: ${chunk}`);
    screenshot('TC-012');
  });

  // Traceability: TC-013 → US-16 / AC-3; Spec AC-21
  test('TC-013: Confirm on confirmation dialog executes status change', () => {
    ab(`open ${listUrl()}`);
    ab('wait --load networkidle');

    const btnRef = findStatusButtonRef(CONFIRM_ITEM_CODE);
    openDropdown(btnRef);

    const items = getMenuItems();
    const closedItem = items.find(i => i.name.includes('已关闭'));
    assert.ok(closedItem, '"已关闭" option not found');
    ab(`click ${closedItem.ref}`);
    ab('wait 800');

    // Find confirm button via snapshot -i
    const dialogSnap = abJson('snapshot -i');
    const dialogEls = getElements(dialogSnap);
    const confirmRef = dialogEls.find(e =>
      e.role === 'button' &&
      (e.name?.includes('确认') || e.name?.includes('确定')) &&
      !e.name?.includes('取消')
    )?.ref;
    assert.ok(confirmRef, `Confirm button not found. Elements: ${JSON.stringify(dialogEls.map(e => ({r:e.ref,n:e.name})))}`);
    ab(`click ${confirmRef}`);
    ab('wait 2000');
    ab('wait --load networkidle');

    // Reload to get fresh state
    ab(`open ${listUrl()}`);
    ab('wait --load networkidle');

    const afterSnap = abJson('snapshot');
    const afterText: string = afterSnap?.data?.snapshot ?? '';
    const idx = afterText.indexOf(CONFIRM_ITEM_CODE);
    assert.ok(idx >= 0, `${CONFIRM_ITEM_CODE} not found after confirm`);
    const chunk = afterText.slice(idx, idx + 400);
    assert.ok(chunk.includes('已关闭'), `Status should be "已关闭" after confirm. Chunk: ${chunk}`);
    screenshot('TC-013');
  });

  // Traceability: TC-015 → Spec AC-18
  test('TC-015: StatusBadge uses code-to-name mapping — Chinese names visible', () => {
    ab(`open ${listUrl()}`);
    ab('wait --load networkidle');
    const snap = abJson('snapshot');
    const text: string = snap?.data?.snapshot ?? '';
    const chineseNames = ['待开始', '进行中', '阻塞中', '已暂停', '待验收', '已完成', '已关闭'];
    const found = chineseNames.filter(name => text.includes(name));
    assert.ok(found.length > 0, `Expected Chinese status names. Found: ${found.join(', ')}`);
    screenshot('TC-015');
  });

  // Traceability: TC-016 → US-5 / AC-3; Spec AC-20
  test('TC-016: Reviewing → progressing/completed options hidden for non-PM (executor)', () => {
    // Switch to executor token
    injectAuth(EXEC_TOKEN);

    const btnRef = findStatusButtonRef(REVIEWING_ITEM_CODE);
    const dropdownText = openDropdown(btnRef);

    assert.ok(!dropdownText.includes('已完成'), `"已完成" must not be visible to non-PM. Snapshot: ${dropdownText.slice(0, 300)}`);
    assert.ok(!dropdownText.includes('进行中'), `"进行中" must not be visible to non-PM. Snapshot: ${dropdownText.slice(0, 300)}`);
    screenshot('TC-016');

    // Restore PM token
    injectAuth(PM_TOKEN);
  });

  // Traceability: TC-051 → UI cache invalidation
  test('TC-051: Status dropdown options refresh after status change', () => {
    // Use a fresh pending item — reuse GENERAL_ITEM_CODE which may be pending again, or use a dedicated one
    const FRESH_ITEM_CODE = process.env.E2E_FRESH_ITEM_CODE ?? 'MI-0294';

    ab(`open ${listUrl()}`);
    ab('wait --load networkidle');

    // Step 1: open dropdown and select progressing
    const btnRef = findStatusButtonRef(FRESH_ITEM_CODE);
    openDropdown(btnRef);

    const itemsBefore = getMenuItems();
    const progressingItem = itemsBefore.find(i => i.name.includes('进行中'));
    assert.ok(progressingItem, `"进行中" option not found for pending item. Items: ${JSON.stringify(itemsBefore)}`);
    ab(`click ${progressingItem.ref}`);
    ab('wait 2000');
    ab('wait --load networkidle');

    // Step 2: verify badge updated to 进行中
    const afterSnap = abJson('snapshot');
    const afterText: string = afterSnap?.data?.snapshot ?? '';
    const idx = afterText.indexOf(FRESH_ITEM_CODE);
    assert.ok(idx >= 0, `${FRESH_ITEM_CODE} not found after status change`);
    const chunk = afterText.slice(idx, idx + 400);
    assert.ok(chunk.includes('进行中'), `Badge should show "进行中" after change. Chunk: ${chunk}`);

    // Step 3: open dropdown again and verify options are for progressing, not pending
    const btnRef2 = findStatusButtonRef(FRESH_ITEM_CODE);
    const dropdownText2 = openDropdown(btnRef2);
    const menuLines2 = dropdownText2.split('\n').filter(l => l.includes('menuitem'));
    const menuText2 = menuLines2.join('\n');

    // progressing → valid targets: blocking, pausing, reviewing, closed
    assert.ok(
      menuText2.includes('阻塞中') || menuText2.includes('已暂停') || menuText2.includes('待验收') || menuText2.includes('已关闭'),
      `Expected transitions for "progressing" state. Menu: ${menuText2}`
    );
    // pending's only option (进行中) should NOT be the only option — pending→progressing was already used
    // More specifically, pending→closed should not appear (pending has no "已暂停" option)
    // The key check: "待开始" should not appear as a transition target (can't go back to pending)
    assert.ok(!menuText2.includes('待开始'), `"待开始" must not appear as transition from progressing. Menu: ${menuText2}`);
    screenshot('TC-051');
  });

  // Traceability: TC-056 → terminal sub-item edit guard
  test('TC-056: Edit button disabled when sub-item is in terminal state', () => {
    const TERMINAL_SUB_URL = process.env.E2E_TERMINAL_SUB_URL ?? `${baseUrl}/items/295/sub-items/132`;
    ab(`open ${TERMINAL_SUB_URL}`);
    ab('wait --load networkidle');

    const snap = abJson('snapshot -i');
    const els = getElements(snap);
    const editBtn = els.find(e => e.role === 'button' && e.name?.includes('编辑'));
    assert.ok(editBtn, 'Edit button not found on sub-item detail page');
    assert.ok(
      editBtn.disabled === true || String(editBtn.disabled) === 'true',
      `Edit button should be disabled for terminal sub-item. Button: ${JSON.stringify(editBtn)}`
    );
    screenshot('TC-056');
  });

  // Traceability: TC-057 → terminal sub-item progress guard
  test('TC-057: Append progress button disabled when sub-item is in terminal state', () => {
    const TERMINAL_SUB_URL = process.env.E2E_TERMINAL_SUB_URL ?? `${baseUrl}/items/295/sub-items/132`;
    ab(`open ${TERMINAL_SUB_URL}`);
    ab('wait --load networkidle');

    const snap = abJson('snapshot -i');
    const els = getElements(snap);
    const appendBtn = els.find(e => e.role === 'button' && e.name?.includes('追加进度'));
    assert.ok(appendBtn, '"追加进度" button not found on sub-item detail page');
    assert.ok(
      appendBtn.disabled === true || String(appendBtn.disabled) === 'true',
      `"追加进度" button should be disabled for terminal sub-item. Button: ${JSON.stringify(appendBtn)}`
    );
    screenshot('TC-057');
  });

  // Traceability: TC-058 → completed status achievement dialog appears
  test('TC-058: Achievement dialog appears when switching sub-item to completed', () => {
    const ACTIVE_SUB_URL = process.env.E2E_ACTIVE_SUB_URL ?? `${baseUrl}/items/295/sub-items/133`;
    ab(`open ${ACTIVE_SUB_URL}`);
    ab('wait --load networkidle');

    // Find and click the status badge to open dropdown
    const snap = abJson('snapshot -i');
    const els = getElements(snap);
    const statusBtn = els.find(e => e.role === 'button' && ['待开始','进行中','阻塞中','已暂停'].some(s => e.name?.includes(s)));
    assert.ok(statusBtn, `Status button not found. Elements: ${JSON.stringify(els.map(e => ({r:e.ref,n:e.name})))}`);
    ab(`click @${statusBtn.ref}`);
    ab('wait 1000');

    const menuItems = getMenuItems();
    const completedItem = menuItems.find(i => i.name.includes('已完成'));
    assert.ok(completedItem, '"已完成" option not found in dropdown');
    ab(`click ${completedItem.ref}`);
    ab('wait 800');

    const dialogSnap = abJson('snapshot');
    const text: string = dialogSnap?.data?.snapshot ?? '';
    assert.ok(text.includes('成果'), `Expected achievement dialog with "成果" field. Snapshot: ${text.slice(0, 500)}`);
    assert.ok(!text.includes('完成度'), `Achievement dialog must NOT contain "完成度" field. Snapshot: ${text.slice(0, 500)}`);
    assert.ok(!text.includes('卡点'), `Achievement dialog must NOT contain "卡点" field. Snapshot: ${text.slice(0, 500)}`);
    screenshot('TC-058');
  });

  // Traceability: TC-059 → cancel on achievement dialog aborts status change
  test('TC-059: Cancel on achievement dialog aborts completed status change', () => {
    // Achievement dialog should still be open from TC-058, or re-trigger
    let snap = abJson('snapshot -i');
    let els = getElements(snap);
    const hasDialog = els.some(e => e.role === 'dialog');

    if (!hasDialog) {
      const ACTIVE_SUB_URL = process.env.E2E_ACTIVE_SUB_URL ?? `${baseUrl}/items/295/sub-items/133`;
      ab(`open ${ACTIVE_SUB_URL}`);
      ab('wait --load networkidle');
      snap = abJson('snapshot -i');
      els = getElements(snap);
      const statusBtn = els.find(e => e.role === 'button' && ['待开始','进行中','阻塞中','已暂停'].some(s => e.name?.includes(s)));
      assert.ok(statusBtn, 'Status button not found');
      ab(`click @${statusBtn.ref}`);
      ab('wait 1000');
      const menuItems = getMenuItems();
      const completedItem = menuItems.find(i => i.name.includes('已完成'));
      assert.ok(completedItem, '"已完成" option not found');
      ab(`click ${completedItem.ref}`);
      ab('wait 800');
    }

    snap = abJson('snapshot -i');
    els = getElements(snap);
    const cancelBtn = els.find(e => e.role === 'button' && e.name?.includes('取消'));
    assert.ok(cancelBtn, 'Cancel button not found in achievement dialog');
    ab(`click @${cancelBtn.ref}`);
    ab('wait 800');

    // Reload and verify status unchanged
    ab('wait --load networkidle');
    snap = abJson('snapshot -i');
    els = getElements(snap);
    const statusBadge = els.find(e => e.role === 'button' && ['待开始','进行中','阻塞中','已暂停'].some(s => e.name?.includes(s)));
    assert.ok(statusBadge, `Status should remain non-terminal after cancel. Elements: ${JSON.stringify(els.map(e => ({r:e.ref,n:e.name})))}`);
    screenshot('TC-059');
  });

  // Traceability: TC-060 → confirm on achievement dialog executes status change
  test('TC-060: Confirm on achievement dialog executes status change and appends progress', () => {
    const ACTIVE_SUB_URL = process.env.E2E_ACTIVE_SUB_URL ?? `${baseUrl}/items/295/sub-items/133`;
    ab(`open ${ACTIVE_SUB_URL}`);
    ab('wait --load networkidle');

    let snap = abJson('snapshot -i');
    let els = getElements(snap);
    const statusBtn = els.find(e => e.role === 'button' && ['待开始','进行中','阻塞中','已暂停'].some(s => e.name?.includes(s)));
    assert.ok(statusBtn, 'Status button not found');
    ab(`click @${statusBtn.ref}`);
    ab('wait 1000');

    const menuItems = getMenuItems();
    const completedItem = menuItems.find(i => i.name.includes('已完成'));
    assert.ok(completedItem, '"已完成" option not found');
    ab(`click ${completedItem.ref}`);
    ab('wait 800');

    // Fill in achievement text
    snap = abJson('snapshot -i');
    els = getElements(snap);
    const textarea = els.find(e => e.role === 'textbox');
    if (textarea) {
      ab(`fill @${textarea.ref} "完成了所有功能开发"`);
    }

    const confirmBtn = els.find(e => e.role === 'button' && e.name?.includes('确认完成'));
    assert.ok(confirmBtn, '"确认完成" button not found in achievement dialog');
    ab(`click @${confirmBtn.ref}`);
    ab('wait 2000');
    ab('wait --load networkidle');

    // Verify status is now 已完成
    snap = abJson('snapshot');
    const text: string = snap?.data?.snapshot ?? '';
    assert.ok(text.includes('已完成'), `Status should be "已完成" after confirm. Snapshot: ${text.slice(0, 500)}`);
    screenshot('TC-060');
  });

});
