import { defineConfig } from '@playwright/test';
import { resolve } from 'node:path';

// Config for running journey web E2E tests from tests/ root.
// Journey specs are in tests/<journey>/ and import from ../web/helpers.ts
export default defineConfig({
  testDir: '.',
  testMatch: /(item-deletion|list-filtering-and-sorting|member-permission-access|sub-item-management|sub-item-move|task-status-transition|team-and-progress-visibility)\/\1\.spec\.ts/,
  timeout: 30_000,
  expect: { timeout: 10_000 },
  retries: 0,
  workers: 1,
  reporter: 'list',
  use: {
    headless: true,
    screenshot: 'only-on-failure',
  },
});
