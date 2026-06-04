import js from '@eslint/js'
import tseslint from 'typescript-eslint'
import stylisticTs from '@stylistic/eslint-plugin-ts'

export default tseslint.config(
  {
    ignores: ['**/dist/**', '**/coverage/**', '**/node_modules/**'],
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ['**/*.{ts,tsx}'],
    plugins: {
      '@stylistic/ts': stylisticTs,
    },
    rules: {
      quotes: ['error', 'single', { avoidEscape: true }],
      semi: ['error', 'never'],
      '@stylistic/ts/member-delimiter-style': [
        'error',
        {
          multiline: { delimiter: 'none' },
          singleline: { delimiter: 'comma', requireLast: false },
        },
      ],
      "no-restricted-syntax": [
        "error",
        {
          selector: "Literal[value=/\\b(emerald|red|amber|slate)-\\d/]",
          message:
            "Use theme tokens (success-*, error-*, warning-*, secondary, tertiary) instead of hardcoded colors.",
        },
      ],
    },
  },
);
