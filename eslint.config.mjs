// Security-focused lint pass for the wallet effort (Phase 2 of the wallet
// plan) -- not a general style linter for the rest of the repo. Scoped to
// packages/wallet-core and packages/tx-builders since those are the
// packages that touch key material / signing; web/, cli/, faucet/, docs/
// have no lint config here and are unaffected.
import security from "eslint-plugin-security";
import tseslint from "typescript-eslint";

export default tseslint.config(
  {
    ignores: ["**/dist/**", "**/node_modules/**", "**/scripts/**"],
  },
  {
    files: ["packages/wallet-core/src/**/*.ts", "packages/tx-builders/src/**/*.ts"],
    languageOptions: {
      parser: tseslint.parser,
    },
    plugins: { security },
    rules: {
      ...security.configs.recommended.rules,
      // Deliberately not extending tseslint's type-checked rule sets here
      // -- this pass is scoped to security anti-patterns (eval, unsafe
      // regex, child_process, weak crypto), not general style/type lint,
      // which `npm run typecheck` already covers per package.
    },
  },
);
