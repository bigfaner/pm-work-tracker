import { defineConfig } from '@playwright/test';

export default defineConfig({
  timeout: 30_000,
  expect: { timeout: 10_000 },
  retries: 0,
  reporter: 'list',
  use: {
    headless: true,
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'setup',
      testMatch: 'auth-setup.ts',
    },
    {
      name: 'authenticated',
      testDir: 'features',
      dependencies: ['setup'],
      use: {
        storageState: 'results/.auth/state.json',
      },
    },
  ],
});
