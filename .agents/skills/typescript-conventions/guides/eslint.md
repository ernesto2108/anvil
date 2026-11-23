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
      // Crítico — detecta promises no manejadas
      "@typescript-eslint/no-floating-promises": "error",
      "@typescript-eslint/no-misused-promises": "error",
      "@typescript-eslint/await-thenable": "error",

      // Seguridad de tipos
      "@typescript-eslint/no-explicit-any": "error",
      "@typescript-eslint/no-unsafe-assignment": "error",
      "@typescript-eslint/no-unsafe-member-access": "error",
      "@typescript-eslint/no-unsafe-call": "error",
      "@typescript-eslint/no-unsafe-return": "error",
      "@typescript-eslint/no-unnecessary-condition": "error",
      "@typescript-eslint/switch-exhaustiveness-check": "error",
      "@typescript-eslint/strict-boolean-expressions": "error",

      // Higiene de imports
      "@typescript-eslint/consistent-type-imports": ["error", { prefer: "type-imports" }],
      "@typescript-eslint/consistent-type-exports": "error",
      "@typescript-eslint/no-import-type-side-effects": "error",

      // Patrones modernos
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

## Explicación de Reglas Críticas

### `no-floating-promises`

```typescript
// INCORRECTO — resultado de promise ignorado, error silenciado
async function save() { /* ... */ }
save(); // sin await, sin .catch, sin void

// CORRECTO
await save();
void save(); // explícito: "no me importa el resultado"
save().catch(handleError);
```

### `no-misused-promises`

```typescript
// INCORRECTO — forEach no hace await
items.forEach(async (item) => {
  await process(item); // ¡se ejecuta de forma concurrente, no secuencial!
});

// CORRECTO
for (const item of items) {
  await process(item);
}
// o en paralelo:
await Promise.all(items.map((item) => process(item)));
```

### `switch-exhaustiveness-check`

```typescript
type Action = "create" | "update" | "delete";
function handle(action: Action) {
  switch (action) {
    case "create": break;
    case "update": break;
    // Error: "delete" no está manejado
  }
}
```
