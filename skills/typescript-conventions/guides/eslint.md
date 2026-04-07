# ESLint + TypeScript (2025)

## Flat Config (`eslint.config.ts`)

```typescript
import eslint from "@eslint/js";
import tseslint from "typescript-eslint";

export default tseslint.config(
  eslint.configs.recommended,
  ...tseslint.configs.strictTypeChecked,
  ...tseslint.configs.stylisticTypeChecked,
  {
    languageOptions: {
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
  {
    rules: {
      // Critical — catches unhandled promises
      "@typescript-eslint/no-floating-promises": "error",
      "@typescript-eslint/no-misused-promises": "error",
      "@typescript-eslint/await-thenable": "error",

      // Type safety
      "@typescript-eslint/no-explicit-any": "error",
      "@typescript-eslint/no-unsafe-assignment": "error",
      "@typescript-eslint/no-unsafe-member-access": "error",
      "@typescript-eslint/no-unsafe-call": "error",
      "@typescript-eslint/no-unsafe-return": "error",
      "@typescript-eslint/no-unnecessary-condition": "error",
      "@typescript-eslint/switch-exhaustiveness-check": "error",
      "@typescript-eslint/strict-boolean-expressions": "error",

      // Import hygiene
      "@typescript-eslint/consistent-type-imports": ["error", { prefer: "type-imports" }],
      "@typescript-eslint/consistent-type-exports": "error",
      "@typescript-eslint/no-import-type-side-effects": "error",

      // Modern patterns
      "@typescript-eslint/prefer-nullish-coalescing": "error",
      "@typescript-eslint/prefer-optional-chain": "error",
      "@typescript-eslint/no-unused-vars": ["error", {
        argsIgnorePattern: "^_",
        varsIgnorePattern: "^_",
      }],
    },
  },
  {
    files: ["**/*.test.ts", "**/*.spec.ts"],
    rules: {
      "@typescript-eslint/no-explicit-any": "off",
      "@typescript-eslint/no-unsafe-assignment": "off",
    },
  },
  { ignores: ["dist/", "node_modules/", "*.js"] },
);
```

## Critical Rules Explained

### `no-floating-promises`

```typescript
// WRONG — promise result ignored, error swallowed
async function save() { /* ... */ }
save(); // no await, no .catch, no void

// RIGHT
await save();
void save(); // explicit: "I don't care"
save().catch(handleError);
```

### `no-misused-promises`

```typescript
// WRONG — forEach doesn't await
items.forEach(async (item) => {
  await process(item); // runs concurrently, not sequentially!
});

// RIGHT
for (const item of items) {
  await process(item);
}
// or parallel:
await Promise.all(items.map((item) => process(item)));
```

### `switch-exhaustiveness-check`

```typescript
type Action = "create" | "update" | "delete";
function handle(action: Action) {
  switch (action) {
    case "create": break;
    case "update": break;
    // Error: "delete" not handled
  }
}
```
