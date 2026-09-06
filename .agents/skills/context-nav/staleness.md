# Staleness — Detección de contexto desactualizado

## Cómo detectar staleness al inicio de sesión

```bash
# Fecha del último cambio en código fuente
git log -1 --format=%ci -- internal/ src/ lib/ cmd/ pkg/ 2>/dev/null

# Fecha registrada en NAVIGATOR.md
grep "last_updated:" .project-context/NAVIGATOR.md
```

Calcular la diferencia en días entre ambas fechas.

## Niveles

| Diff | Estado | Acción |
|------|--------|--------|
| 0-3 días | FRESH | Usar sin advertencia |
| 3-7 días | STALE | Mencionar en una línea: "El contexto tiene X días. Puede haber cambios no reflejados." |
| > 7 días | OUTDATED | Recomendar re-scan: "El contexto tiene X días. Sugiero correr `context-init` (modo deep) antes de continuar." |
| No existe | NONE | Informar: "No hay `.project-context/`. Puedes generarlo con `context-init mode: init`." |

## Reglas de comportamiento

- **Nunca bloquear trabajo** por staleness — es informativo, no un gate
- **Mencionar una sola vez** al inicio de sesión, no repetir en cada respuesta
- Si el usuario dice "sé que está desactualizado, continúa" → no volver a mencionarlo
- En modo pipeline, el orquestador verifica staleness en el Paso 0.5 y lo incluye en el brief del architect si aplica

## Staleness granular por dominio

Cada `domains/<name>.md` tiene su propio `last_updated`. Si solo el dominio `memory` está stale pero `pipeline` está fresh, solo advertir sobre `memory` si la tarea toca ese dominio.

## Gap-filling con MCP memory cuando hay staleness

Cuando un dominio está STALE o OUTDATED **y la tarea lo necesita**, antes de advertir al usuario intentar llenar el gap con memoria semántica:

```
mcp__anvil__search_memories(
  query = "<nombre del dominio> cambios decisiones",
  limit = 3
)
```

Si hay hits con `score >= 0.5`:
- Inyectar el contenido de los digests como contexto provisional en el brief del agente
- Etiquetar claramente: `"⚠️ Contexto de domains/<X>.md stale — complementado con digests de memoria (run-N, hace X días)"`
- No actualizar el archivo `.project-context/` en este paso — eso es responsabilidad del reporter al final del run

Si no hay hits útiles → advertir staleness normalmente y continuar.

Esto evita que un dominio stale bloquee el trabajo o fuerce un re-scan completo.

## Coverage tracking

NAVIGATOR.md registra el nivel de cobertura:

| Valor | Significado |
|-------|-------------|
| `none` | No existe o está vacío |
| `bootstrap` | Generado automáticamente por context-init — puede tener gaps |
| `partial` | Bootstrap + algunas actualizaciones manuales |
| `full` | Revisado y completado manualmente por el equipo |

El coverage no determina staleness — un `full` puede estar stale si el código evolucionó sin actualizar `.project-context/`.
