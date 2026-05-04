import js from "@eslint/js";
import tseslint from "typescript-eslint";

export default tseslint.config(
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ["src/**/*.{ts,tsx}"],
    rules: {
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
