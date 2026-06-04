import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    testTimeout: 30_000,
    hookTimeout: 60_000,
    sequence: { sequential: true },
    globalSetup: ['./vitest.global-setup.ts'],
  },
});
