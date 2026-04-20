import js from '@eslint/js';

export default [
  js.configs.recommended,
  {
    files: ['src/**/*.{ts,tsx}'],
    rules: {
      'no-restricted-syntax': [
        'error',
        {
          selector: 'Literal[value=/\\b(emerald|red|amber|slate)-\\d/]',
          message:
            'Use theme tokens (success-*, error-*, warning-*, secondary, tertiary) instead of hardcoded colors.',
        },
      ],
    },
  },
];
