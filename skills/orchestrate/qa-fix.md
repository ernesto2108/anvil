---
name: orchestrate/qa-fix
description: Protocolo de re-invocación del loop de corrección QA, plantilla de prompt, modo security-fix y cuándo NO usar este modo. Cargar cuando el orquestador esté por re-invocar developer después de hallazgos bloqueantes de QA o security.
---

# Loop de Corrección QA

**Cargar cuando:** QA o el agente de security devuelve hallazgos bloqueantes en una tarea completada y el orquestador está por re-invocar developer para aplicar correcciones.

---

## Regla: re-invocar al developer en modo `qa-fix`

Cuando QA devuelve hallazgos BLOQUEANTES, el orquestador debe re-invocar al developer para aplicar correcciones. Una invocación NUEVA del developer cuesta ~20-30k tokens (re-cargar skills de convención, PRD, diseño, context.md, archivos de producción, handoff). Eso es puro desperdicio — la invocación anterior del developer ya tenía todo ese contexto y escribió el handoff para probarlo.

El orquestador pasa `Mode: qa-fix` en el prompt del developer. En este modo el developer:
1. Lee `.handoff/<TASK-ID>.md` (el handoff de la primera invocación) como contexto PRINCIPAL
2. **NO** relee PRD, diseño, context.md
3. **NO** recarga la skill de convención completa — el orquestador inyecta SOLO las reglas específicas que aplican a los archivos siendo corregidos (3-5 reglas inline, no el dispatcher completo)
4. Lee **SOLO** los archivos listados en los hallazgos de QA, no el paquete completo ni todo el codebase
5. Aplica correcciones QUIRÚRGICAS — sin refactors, sin limpiezas de "ya que estoy aquí"
6. Re-ejecuta validación solo para los archivos tocados (`go vet -tags <tag> ./internal/<pkg>`, `npm run build` solo si cambió frontend)
7. Actualiza `## Notas` en el handoff con una entrada de una línea por corrección aplicada
8. NO reescribe `## Handoff para tester` a menos que una corrección haya cambiado la firma de una interfaz pública

---

## Plantilla de prompt qa-fix

```
Mode: qa-fix. TASK-ID: <TASK-ID>.

El developer ya completó la implementación inicial para esta tarea. El handoff en `.handoff/<TASK-ID>.md` contiene el contexto completo: archivos tocados, patrones aplicados, decisiones tomadas, validación ejecutada. ESE es tu contexto principal.

La revisión de QA devolvió los siguientes hallazgos BLOQUEANTES (deben corregirse antes de que esta tarea pueda mergearse):

<hallazgos de QA inline — issues exactos con paths de archivos y números de línea cuando estén disponibles, un hallazgo por bullet>

Reglas de ejecución:
1. Lee `.handoff/<TASK-ID>.md` primero. NO releas PRD, diseño ni context.md.
2. Lee SOLO los archivos mencionados en los hallazgos de QA arriba — no el paquete completo, no todo el codebase.
3. Aplica correcciones MÍNIMAS — aborda SOLO los hallazgos. Sin refactors extras, sin limpiezas, sin "ya que estoy aquí".
4. Re-ejecuta comandos de validación con alcance limitado a los archivos que tocaste:
   - Go: `go vet -tags <tag> ./internal/<pkg>` + paquete de unit test relevante
   - Frontend: `npm run build` solo si tocaste .ts/.tsx
5. Actualiza `## Notas` en el handoff — una línea por corrección aplicada.
6. NO modifiques `## Handoff para tester` a menos que una corrección haya cambiado la firma de una interfaz pública.

Reglas de convención que aplican a tu corrección (inyectadas inline — NO cargues la skill completa):
<inline SOLO las 3-5 reglas específicas de la skill de convención que aplican a la corrección — ej., "wrapping de errores: envolver errores SQL con fmt.Errorf('paquete: acción: %w', err)" — NO el dispatcher completo>

Prohibido:
- Cargar la skill de convención completa
- Leer PRD / diseño / context.md
- Tocar archivos fuera de los hallazgos
- Refactorizar código que funciona
```

---

## Cuándo NO usar modo qa-fix

- **Los hallazgos son no-bloqueantes** (puntaje del rubric sigue ≥7) → agregarlos al backlog como tareas de seguimiento, NO re-invocar al developer para nada
- **Los hallazgos requieren cambios arquitecturales** (nuevos patrones, nuevas abstracciones, mover archivos) → re-invocar al developer en modo NORMAL con un nuevo plan, no qa-fix
- **Los hallazgos abarcan > 5 archivos** → las correcciones ya no son quirúrgicas; usar modo normal con un plan enfocado

---

## Modo security-fix

Cuando el agente de security devuelve hallazgos bloqueantes en una tarea completada, usar `Mode: security-fix` con la misma plantilla (intercambiar "QA" por "security" en el prompt). La lógica de ahorro de contexto es idéntica.

---

## Ahorros esperados

En tareas como DASH-FEAT-006, donde una re-invocación nueva del developer para correcciones de a11y costó **22k tokens**, el modo qa-fix debería reducir eso a **~5-8k**. Ahorro por ciclo de QA: **~15-17k tokens**. En 5 tareas con ciclos de QA, eso es el presupuesto de una invocación completa recuperado.
