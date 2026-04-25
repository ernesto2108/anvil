---
name: orchestrate/plan-approval
description: Protocolo de aprobación de plan Flow 0/A/B, checkpoints de clarificación, enriquecimiento de handoff developer-a-tester e inyección de convenciones para tareas Small. Cargar cuando el orquestador esté por invocar developer para una tarea Medium+.
---

# Aprobación de Plan

**Cargar cuando:** se esté por invocar developer para una tarea Medium+ (decisión Flow 0/A/B, preparación de cadena de handoff).

---

## Checkpoints de clarificación (OBLIGATORIO)

Antes de lanzar ciertos agentes, el orquestador DEBE hacer preguntas al usuario. NO asumir — preguntar primero.

### Antes del Architect (si la tarea toca DB o schema)

Preguntar: (1) "¿Qué tablas existentes están relacionadas?" (2) "¿Extender tabla existente o crear nueva?" (3) "¿Constraints o relaciones a considerar?"

**Por qué:** previene que el Architect diseñe una tabla nueva cuando ALTER TABLE bastaría.

### Antes del Developer

**Para tareas Medium+**, verificar si existe una nota de handoff según la skill `/handoff` (operación Read). Si existe, pasarla inline al developer — es una continuación.

**Si no existe handoff**, preguntar al usuario:
1. "¿Ya tienes avance en esta feature? ¿Qué archivos ya existen?"
2. "¿Hay código parcial o un branch con trabajo previo?"

**Por qué:** El handoff previene que el Developer gaste tokens re-leyendo PRD, diseño y código ya procesado. Si el usuario confirma trabajo previo (y no hay handoff), ser específico: "Solo falta X, Y, Z — no leas el resto."

**Omitir verificación de handoff para tareas Small (1-5 pts).**

---

## Flujos de aprobación de plan (CRÍTICO)

**El USUARIO aprueba planes, no el orquestador.** Regla dura. Tres flujos, elegidos según si un architect ya corrió.

### Flow 0 — Architect corrió, reutilizar checklist §8 (PREFERIDO para tareas Complex)

Cuando el architect acaba de producir `design.md` con un `§8 Checklist de implementación` (o equivalente "pasos secuenciales"), **no re-sintetizar** el plan. El checklist ES el plan a nivel de granularidad de archivos.

1. El orquestador ya tiene `design.md` en contexto desde la llamada al architect
2. Presentar el `checklist §8` **textualmente** al usuario + un resumen de 2 líneas de decisiones clave D1-Dn
3. El usuario aprueba explícitamente (`sí`, `dale`, `apruebo`)
4. Invocar developer con `plan_preapproved=true` y el design.md inlined (ver regla de Design inline abajo)

**Prohibido en Flow 0:** reescribir el checklist, agregar archivos que el architect no listó, eliminar archivos sin aprobación del usuario, pedirle al developer que "genere la lista de archivos". La retrospectiva de DASH-FEAT-008 mostró que la ausencia de Flow 0 causó tokens duplicados y desviación menor.

### Flow A — Plan dictado por el usuario, sin architect (atajo)

Válido SOLO cuando el orquestador ha: leído el contexto él mismo, diseñado un plan concreto con lista de archivos + patrones + decisiones, presentado al usuario en la conversación principal, y recibido aprobación EXPLÍCITA (`sí`, `dale`, `apruebo`) — no un genérico "continúa".

Luego invocar developer una vez con `plan_preapproved=true` + el plan completo inline + instrucciones de crear `.handoff/<TASK-ID>.md` como artefacto de progreso, proceder directamente, actualizar handoff durante el trabajo, llenar `## Handoff para tester` antes de terminar.

**Test de legitimidad de Flow A:** ¿el USUARIO escribió la lista de archivos en el chat, o el ORQUESTADOR la sintetizó y el usuario solo aprobó la estrategia ("ve con la opción B")? Aprobación estratégica = Flow B, no Flow A. Ver `anti-patterns.md` #6.

### Flow B — Developer diseña el plan (sin architect, sin plan del usuario)

1. Invocar developer con: `"PRIMER PASO OBLIGATORIO: Crear .handoff/<TASK-ID>.md con plan de ejecución. Luego DETENERSE y devolver resumen del plan — NO escribir código de producción. NO presentar directamente al usuario."`
2. Developer devuelve resumen del plan
3. **El orquestador muestra el plan al usuario y ESPERA aprobación explícita.** Frases prohibidas: "el plan coincide, apruebo", "sigo adelante" — el usuario decide, no el orquestador
4. Ciclo:
   - `dale` / `ok` / `aprobado` → invocar developer con `plan_preapproved=true` + plan aprobado inline
   - El usuario pide cambios → nuevo plan → mostrar de nuevo
   - El usuario rechaza → reiniciar alcance

**Nunca:** auto-aprobar, interpretar silencio como aprobación, interpretar genérico "sigue" como aprobación, omitir el paso de mostrar.

### Regla de Design inline (Flow 0 + handoffs para tester)

Al invocar el developer después de Flow 0, **incluir el contenido de `design.md` inline en el prompt** en lugar de solo pasar el path, cuando TODO: `design.md` ≤500 líneas, el orquestador ya lo tiene en contexto, el developer necesita el diseño completo.

Esto ahorra una llamada `Read` completa (~3-5k tokens) y garantiza que el developer vea la misma versión que el usuario aprobó. Pasarlo como:

```
Diseño (pre-aprobado por el usuario, de {task_path}/design.md):
<contenido inline>

NO releas el archivo de diseño — tienes el contenido completo arriba.
```

Para archivos >500 líneas: pasar el path + decirle al developer qué secciones son clave.

**Misma regla para handoffs al tester:** incluir inline la sección `## Handoff para tester` si el orquestador la tiene en contexto, en lugar de pedirle al tester que relea el archivo de handoff.

---

## Regla de path de handoff (CRÍTICO)

Handoffs → `.handoff/` en la raíz del proyecto. Docs → `<docs>/` vault. Nunca mezclar. Ver mapa de paths en `vault-setup.md`.

---

## Enriquecimiento de handoff Developer → Tester (OBLIGATORIO)

Antes de que el orquestador invoque al tester, DEBE verificar que el developer llenó la sección `## Handoff para tester` de `.handoff/<TASK-ID>.md`. Esta sección existe precisamente para que el tester no relea archivos de producción.

**Checklist de verificación (antes de invocar tester):**

1. [ ] `.handoff/<TASK-ID>.md` tiene una sección `## Handoff para tester` no vacía
2. [ ] "Interfaces públicas / contratos" tiene las firmas exactas de funciones, tipos y DTOs nuevos/modificados
3. [ ] "Edge cases descubiertos" está lleno (no solo "N/A" — si realmente no hay, el developer debe decir "sin edge cases no triviales")
4. [ ] "Tests requeridos — por stack" tiene tests agrupados por stack (`#### Tests Go`, `#### Tests React/TS`, etc.) — cada grupo con path de archivo, comando de ejecución y lista numerada. **Una lista plana NO es aceptada para tareas cross-stack.**
5. [ ] "Validación ya ejecutada" lista los comandos que el developer ejecutó (go build, go vet, npm run build)
6. [ ] `## Output entregado` tiene tabla llena con resultados de build/lint/test
7. [ ] `## Puente de contratos` está lleno (solo tareas cross-stack) — tanto "Backend expone" como "Frontend/Mobile consume" tienen tipos exactos
8. [ ] `## Dependencias cross-service` está lleno (solo tareas cross-service)

**Si cualquier verificación falla:** re-invocar al developer con el gap específico: "Llena [sección faltante] en `.handoff/<TASK-ID>.md`. NO toques código de producción." Esto es más barato que dejar al tester releer el codebase.

**Después de que QA pasa (antes de archivar):** verificar que el developer llenó `## Retro` → "Qué funcionó" y "Qué no funcionó". El orquestador llena "Métricas" con conteos reales de invocaciones y rebotes de QA.

**Plantilla de prompt para tester (después de que la verificación pasa):**

```
Stack(s): <go|react|flutter|...>. Skill: <convention-skill>.

INPUT PRINCIPAL: Lee `.handoff/<TASK-ID>.md` — específicamente la sección `## Handoff para tester`. Esa sección contiene:
- archivos que el developer tocó (con su rol)
- firmas exactas de interfaces/DTOs nuevos
- patrones aplicados
- edge cases descubiertos
- build tags / constraints
- **tests requeridos — por stack** — tests agrupados por stack (#### Tests Go, #### Tests React/TS, etc.), cada uno con path de archivo, comando de ejecución y lista numerada. Trabaja un stack a la vez.
- validación ya ejecutada (NO repetir verificaciones de build)

Para tareas cross-stack, también revisa `## Puente de contratos` — muestra el contrato exacto entre backend y frontend/mobile. Si tus tests tocan el boundary, verifica que ambos lados coincidan.

NO releas los archivos de producción a menos que el handoff no tenga un detalle específico que necesites. Si el handoff está incompleto, DETENTE y reporta al orquestador.

Tu trabajo: implementar SOLO los tests listados en cada grupo de stack de "Tests requeridos — por stack". NO agregues tests extra más allá de estas listas.
```

Límite del developer: nunca escribe archivos de test. Si el tester encuentra tests escritos por el dev, reportar violación (ver tester.md).

---

## Inyección de convenciones para tareas Small

Para tareas Small (1-5 pts), NO decirle al developer que cargue la skill de convención completa. En su lugar, leer las reglas esenciales e inyectar inline en el prompt:

- **Go:** `go-conventions/rules/coding.md` + `rules/architecture.md`
- **React / Flutter / Astro:** leer reglas esenciales de `<stack>-conventions`, incluir inline

**Regla de inyección de contexto:** si el usuario proporcionó contexto en la conversación (screenshots, archivos, decisiones), pasarlo inline — NO decirle al agente "lee el archivo X". Ver instrucciones globales para el protocolo completo.
