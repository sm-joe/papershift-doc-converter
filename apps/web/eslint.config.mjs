import eslint from "@eslint/js";
import { globalIgnores } from "eslint/config";
import tseslint from "typescript-eslint";

export default tseslint.config([
  globalIgnores([
    ".next/**",
    "out/**",
    "build/**",
    "next-env.d.ts",
  ]),

  eslint.configs.recommended,

  ...tseslint.configs.recommended,
]);
