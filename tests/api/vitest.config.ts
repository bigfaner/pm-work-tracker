import { defineConfig } from 'vitest/config';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  test: {
    include: [
      '**/*.spec.ts',
    ],
    exclude: [
      '**/node_modules/**',
      '**/web/**',
    ],
    testTimeout: 30_000,
    hookTimeout: 60_000,
    sequence: { sequential: true },
    globalSetup: ['./vitest.global-setup.ts'],
  },
  resolve: {
    alias: {
      yaml: resolve(__dirname, 'node_modules/yaml'),
    },
  },
});
