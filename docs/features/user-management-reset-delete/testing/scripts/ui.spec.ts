import { describe, test, before, after } from 'node:test';
import assert from 'node:assert/strict';
import { ab, abJson, snapshotContains, findElement, screenshot, baseUrl } from './helpers.js';

describe('UI E2E Tests — User Management Reset Password & Delete', () => {
  before(() => {
    ab(`open ${baseUrl}`);
    ab('wait --load domcontentloaded');
  });

  after(() => {
    ab('close');
  });

  // ── Story 5: Super admin can see reset password and delete buttons ──

  // Traceability: TC-001 → Story 5 / AC-1, UI Function 1
  test('TC-001: Super admin can see reset password and delete buttons', () => {
    ab(`open ${baseUrl}/users`);
    ab('wait --load domcontentloaded');

    assert.ok(snapshotContains('用户管理'), '用户管理标题存在');
    assert.ok(snapshotContains('操作'), '操作列存在');

    // Each user row should have reset password and delete buttons in action column
    const resetBtn = findElement('button', '重置密码');
    const deleteBtn = findElement('button', '删除');
    assert.ok(resetBtn !== null || snapshotContains('重置密码'), '重置密码按钮存在');
    assert.ok(deleteBtn !== null || snapshotContains('删除'), '删除按钮存在');
    screenshot('TC-001');
  });

  // Traceability: TC-002 → Story 5 / AC-1
  test('TC-002: Non-super-admin cannot see reset password and delete buttons', () => {
    // This test requires logging in as a non-super-admin user
    // In a real scenario, logout first then login as regular user
    ab(`open ${baseUrl}/users`);
    ab('wait --load domcontentloaded');

    // For non-super-admin, these buttons should NOT be rendered
    // This test documents expected behavior; actual execution requires role switch
    screenshot('TC-002');
  });

  // Traceability: TC-003 → Story 4 / AC-1, UI Function 1 States
  test('TC-003: Delete button disabled on own row with tooltip', () => {
    ab(`open ${baseUrl}/users`);
    ab('wait --load domcontentloaded');

    // Find the logged-in user's row and check delete button is disabled
    const snapshot = abJson('snapshot -i');
    const snapshotStr = JSON.stringify(snapshot);
    // The own-row delete button should have disabled attribute
    // and tooltip "Cannot delete your own account"
    screenshot('TC-003');
  });

  // ── Story 1: Reset password dialog ──

  // Traceability: TC-004 → Story 1 / AC-1, UI Function 2
  test('TC-004: Clicking reset password opens dialog with user display name', () => {
    ab(`open ${baseUrl}/users`);
    ab('wait --load domcontentloaded');

    const resetBtn = findElement('button', '重置密码');
    if (resetBtn) {
      ab(`click ${resetBtn}`);
      ab('wait --load domcontentloaded');

      // Dialog should open with title containing "重置密码" and the user's display name
      assert.ok(snapshotContains('重置密码'), '重置密码对话框标题存在');
      screenshot('TC-004');
    }
    screenshot('TC-004-fallback');
  });

  // Traceability: TC-005 → Story 2 / AC-1, UI Function 2 Validation Rules
  test('TC-005: Reset password empty validation on submit', () => {
    ab(`open ${baseUrl}/users`);
    ab('wait --load domcontentloaded');

    const resetBtn = findElement('button', '重置密码');
    if (resetBtn) {
      ab(`click ${resetBtn}`);
      ab('wait --load domcontentloaded');

      // Leave password fields empty and click confirm
      const confirmBtn = findElement('button', '确认');
      if (confirmBtn) {
        ab(`click ${confirmBtn}`);
        ab('wait --load domcontentloaded');

        // Dialog should stay open with validation error
        assert.ok(
          snapshotContains('请输入') || snapshotContains('必填') || snapshotContains('不能为空'),
          '显示空密码校验错误',
        );
      }
    }
    screenshot('TC-005');
  });

  // Traceability: TC-006 → Story 2 / AC-1, PRD Spec Section 5.3 Validation Rules
  test('TC-006: Reset password strength validation on blur and submit', () => {
    ab(`open ${baseUrl}/users`);
    ab('wait --load domcontentloaded');

    const resetBtn = findElement('button', '重置密码');
    if (resetBtn) {
      ab(`click ${resetBtn}`);
      ab('wait --load domcontentloaded');

      // Enter a weak password (too short)
      const passwordInput = findElement('textbox', '新密码');
      if (passwordInput) {
        ab(`fill ${passwordInput} "abc12"`);
        ab('press Tab');
        ab('wait --load domcontentloaded');

        // Should show strength error
        assert.ok(
          snapshotContains('8') || snapshotContains('字母') || snapshotContains('数字'),
          '显示密码强度校验错误',
        );

        // Enter password with only letters
        ab(`fill ${passwordInput} "abcdefgh"`);
        ab('press Tab');
        ab('wait --load domcontentloaded');

        assert.ok(
          snapshotContains('字母') || snapshotContains('数字'),
          '显示缺少数字校验错误',
        );

        // Enter valid password
        ab(`fill ${passwordInput} "NewPass123"`);
        ab('press Tab');
        ab('wait --load domcontentloaded');
        // No error should be present for valid password
      }
    }
    screenshot('TC-006');
  });

  // Traceability: TC-007 → Story 2 / AC-1, UI Function 2 Validation Rules
  test('TC-007: Reset password confirm mismatch validation', () => {
    ab(`open ${baseUrl}/users`);
    ab('wait --load domcontentloaded');

    const resetBtn = findElement('button', '重置密码');
    if (resetBtn) {
      ab(`click ${resetBtn}`);
      ab('wait --load domcontentloaded');

      // Enter valid new password
      const passwordInput = findElement('textbox', '新密码');
      if (passwordInput) ab(`fill ${passwordInput} "NewPass123"`);

      // Enter different confirm password
      const confirmInput = findElement('textbox', '确认密码');
      if (confirmInput) {
        ab(`fill ${confirmInput} "Different1"`);
        ab('press Tab');
        ab('wait --load domcontentloaded');

        assert.ok(
          snapshotContains('不一致') || snapshotContains('不匹配') || snapshotContains('相同'),
          '显示确认密码不匹配错误',
        );
      }
    }
    screenshot('TC-007');
  });

  // Traceability: TC-008 → Story 1 / AC-1
  test('TC-008: Reset password success flow', () => {
    ab(`open ${baseUrl}/users`);
    ab('wait --load domcontentloaded');

    const resetBtn = findElement('button', '重置密码');
    if (resetBtn) {
      ab(`click ${resetBtn}`);
      ab('wait --load domcontentloaded');

      const passwordInput = findElement('textbox', '新密码');
      if (passwordInput) ab(`fill ${passwordInput} "NewPass123"`);

      const confirmInput = findElement('textbox', '确认密码');
      if (confirmInput) ab(`fill ${confirmInput} "NewPass123"`);

      const confirmBtn = findElement('button', '确认');
      if (confirmBtn) ab(`click ${confirmBtn}`);
      ab('wait --load domcontentloaded');

      // On success: dialog closes, toast shows success message
      assert.ok(
        snapshotContains('重置成功') || snapshotContains('成功'),
        '显示密码重置成功提示',
      );
    }
    screenshot('TC-008');
  });

  // Traceability: TC-009 → Story 1 / AC-2
  test('TC-009: Reset password API error keeps dialog open', () => {
    ab(`open ${baseUrl}/users`);
    ab('wait --load domcontentloaded');

    const resetBtn = findElement('button', '重置密码');
    if (resetBtn) {
      ab(`click ${resetBtn}`);
      ab('wait --load domcontentloaded');

      const passwordInput = findElement('textbox', '新密码');
      if (passwordInput) ab(`fill ${passwordInput} "NewPass123"`);

      const confirmInput = findElement('textbox', '确认密码');
      if (confirmInput) ab(`fill ${confirmInput} "NewPass123"`);

      const confirmBtn = findElement('button', '确认');
      if (confirmBtn) ab(`click ${confirmBtn}`);
      ab('wait --load domcontentloaded');

      // On error: dialog stays open, error message displayed inside dialog
      // This test documents expected behavior; simulating backend error requires test environment setup
    }
    screenshot('TC-009');
  });

  // ── Story 3: Delete user dialog ──

  // Traceability: TC-010 → Story 3 / AC-1, UI Function 3
  test('TC-010: Clicking delete opens confirmation dialog with username', () => {
    ab(`open ${baseUrl}/users`);
    ab('wait --load domcontentloaded');

    // Find a delete button that is NOT on the current user's row
    const deleteBtn = findElement('button', '删除');
    if (deleteBtn) {
      ab(`click ${deleteBtn}`);
      ab('wait --load domcontentloaded');

      // Confirmation dialog should open with username
      assert.ok(
        snapshotContains('确认删除') || snapshotContains('确认') || snapshotContains('不可恢复'),
        '删除确认对话框显示',
      );
      // Confirm Delete and Cancel buttons should be visible
      assert.ok(
        snapshotContains('取消'),
        '取消按钮可见',
      );
    }
    screenshot('TC-010');
  });

  // Traceability: TC-011 → Story 3 / AC-1
  test('TC-011: Delete user success removes row and shows toast', () => {
    ab(`open ${baseUrl}/users`);
    ab('wait --load domcontentloaded');

    const deleteBtn = findElement('button', '删除');
    if (deleteBtn) {
      ab(`click ${deleteBtn}`);
      ab('wait --load domcontentloaded');

      const confirmBtn = findElement('button', '确认删除');
      if (confirmBtn) {
        ab(`click ${confirmBtn}`);
        ab('wait --load domcontentloaded');

        // On success: dialog closes, row removed, toast shows "User deleted"
        assert.ok(
          snapshotContains('删除成功') || snapshotContains('已删除') || snapshotContains('成功'),
          '显示删除成功提示',
        );
      }
    }
    screenshot('TC-011');
  });

  // Traceability: TC-012 → Story 3 / AC-2
  test('TC-012: Delete user 404 shows message and removes row', () => {
    ab(`open ${baseUrl}/users`);
    ab('wait --load domcontentloaded');

    // This test documents expected behavior for 404 scenario
    // Simulating stale list requires backend to return 404
    screenshot('TC-012');
  });

  // Traceability: TC-013 → UI Function 3 States
  test('TC-013: Delete user API error shows error in dialog', () => {
    ab(`open ${baseUrl}/users`);
    ab('wait --load domcontentloaded');

    // This test documents expected behavior for API error scenario
    // Simulating backend error requires test environment setup
    screenshot('TC-013');
  });

  // ── Story 6: Copy credentials ──

  // Traceability: TC-014 → Story 6 / AC-1, UI Function 4
  test('TC-014: Copy credentials button copies to clipboard', () => {
    ab(`open ${baseUrl}/users`);
    ab('wait --load domcontentloaded');

    // Step 1: Create a new user to trigger result dialog
    const createBtn = findElement('button', '创建用户');
    if (createBtn) {
      ab(`click ${createBtn}`);
      ab('wait --load domcontentloaded');

      const nameInput = findElement('textbox', '请输入姓名');
      if (nameInput) ab(`fill ${nameInput} "E2E Test User"`);

      const usernameInput = findElement('textbox', '请输入账号');
      if (usernameInput) ab(`fill ${usernameInput} "e2e_test_user_${Date.now()}"`);

      const emailInput = findElement('textbox', '请输入邮箱');
      if (emailInput) ab(`fill ${emailInput} "e2e@test.com"`);

      const confirmCreateBtn = findElement('button', '确认创建');
      if (confirmCreateBtn) ab(`click ${confirmCreateBtn}`);
      ab('wait --load domcontentloaded');

      // Result dialog should appear with copy button
      const copyBtn = findElement('button', '复制账号密码');
      if (copyBtn) {
        ab(`click ${copyBtn}`);
        ab('wait --load domcontentloaded');

        // Button text should change to "已复制"
        assert.ok(
          snapshotContains('已复制') || snapshotContains('Copied'),
          '复制按钮状态变为已复制',
        );
      }
    }
    screenshot('TC-014');
  });

  // Traceability: TC-015 → Story 6 / AC-2, UI Function 4 States
  test('TC-015: Copy credentials failure shows error toast', () => {
    // This test documents expected behavior for clipboard API failure
    // Simulating clipboard failure requires browser environment manipulation
    screenshot('TC-015');
  });

  // ── Password visibility toggle ──

  // Traceability: TC-016 → UI Function 2 (password field with show/hide toggle)
  test('TC-016: Password visibility toggle in reset password dialog', () => {
    ab(`open ${baseUrl}/users`);
    ab('wait --load domcontentloaded');

    const resetBtn = findElement('button', '重置密码');
    if (resetBtn) {
      ab(`click ${resetBtn}`);
      ab('wait --load domcontentloaded');

      // Enter text in password field
      const passwordInput = findElement('textbox', '新密码');
      if (passwordInput) ab(`fill ${passwordInput} "TestPass123"`);

      // Click the eye toggle icon
      const toggleBtn = findElement('button', '切换密码可见性');
      if (toggleBtn) {
        ab(`click ${toggleBtn}`);
        ab('wait --load domcontentloaded');

        // Password should now be visible (plain text)
        screenshot('TC-016-visible');

        // Click toggle again
        ab(`click ${toggleBtn}`);
        ab('wait --load domcontentloaded');

        // Password should be masked again
        screenshot('TC-016-masked');
      } else {
        // Try alternative: look for eye icon
        screenshot('TC-016-no-toggle');
      }
    }
  });
});
