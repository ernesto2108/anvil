---
name: orchestrate/anti-patterns
description: Los 7 anti-patrones del orquestador con razonamiento completo, ejemplos y la sección de disciplina de tokens. Cargar cuando se esté por hacer cualquier operación Read/Write no trivial o cuando surjan dudas de límites.
---

# Anti-Patrones

**Cargar cuando:** se esté por hacer cualquier operación Read/Write no trivial, o cuando surjan dudas de límites.

---

## 1. Handoff manual como bypass de triaje

**Incorrecto:** el usuario pide una feature Medium+, el orquestador escribe el plan en `.handoff/<TASK>.md` directamente y lo ejecuta él mismo (racionaliza: "ya tengo el contexto, los agentes son overhead").

**Por qué está mal:** La skill `/orchestrate` ES el triaje. Saltarlo omite la selección de pipeline, carga de skills y límites de agentes — violando la regla "el developer es el único que escribe código".

**Correcto:** Invocar triaje, seleccionar pipeline, lanzar developer con el plan inline. El handoff lo escribe EL developer, no en lugar de él.

---

## 2. "Esto se ve simple, lo hago yo"

**Incorrecto:** el orquestador decide que una tarea es "básicamente boilerplate" y pasa directo a editar SIN que el usuario haya elegido modo directo.

**Por qué está mal:** El usuario controla el modo de ejecución (ver Regla #0). El orquestador no decide "esto es tan simple que puedo saltarme los agentes." Si el usuario activó el orquestador, usa agentes. Si el usuario dijo "hazlo directo", entonces la ejecución directa es correcta.

**Correcto:** seguir el modo de ejecución elegido por el usuario. Si hay duda, preguntar.

---

## 3. Escribir código sin autorización del usuario

**Incorrecto:** el orquestador escribe `.go`, `.ts`, `.tsx` etc. mientras corre en modo orquestación (agentes activados).

**Por qué está mal:** En modo orquestación, escribir código es trabajo del agente developer. El rol del orquestador es enrutamiento, gates y confirmaciones con el usuario — no implementación.

**Correcto:** En modo orquestación, delegar todo el código al agente developer. En modo directo (elección del usuario), escribir código tú mismo es correcto y esperado.

**Siempre permitido sin importar el modo:**
- `go.mod` vía `go get` / `go mod tidy`
- Archivos de config: `Makefile`, `wails.json`, `.gitignore`, `tsconfig.json`, `package.json`
- Comandos shell: `go build`, `go vet`, `npm run build`
- Docs markdown, archivos de handoff, sprint-current.md

---

## 4. Afirmar que una skill fue cargada sin confirmación

**Incorrecto:** el orquestador nombra una skill en el prompt y asume que el agente la cargó sin verificar.

**Por qué está mal:** el agente puede haberla saltado silenciosamente. El usuario lo detecta después — forzando una re-ejecución.

**Correcto:** después de que el agente complete, verificar que su reporte nombre la skill o referencie sus reglas. Si ninguno lo hace, seguir con: "confirma que cargaste <skill> y revalida los archivos contra sus reglas." Más barato que rehacer.

---

## 5. Construir contexto en Opus en lugar de delegar a un subagente

**Aplica SOLO en modo orquestación (agentes activados).** En modo directo, leer código fuente es esperado y correcto.

**Incorrecto (modo orquestación):** el orquestador lee 10+ archivos de código fuente para "entender la feature" antes de invocar cualquier agente, pagando tarifas Opus por lecturas que el developer haría en Sonnet.

**Por qué está mal:**
1. **Ratio de costo.** Opus es aproximadamente 4x Sonnet por token. Cada archivo que el orquestador lee cuesta 4x la misma lectura dentro de cualquier subagente.
2. **Doble conteo.** Leer archivos PARA pasarlos inline es el modo de falla, no la intención.
3. **Los subagentes leen mejor.** Cargan skills de convención y tienen framing de dominio.

**Correcto (modo orquestación):**
- Para entender el codebase: lanzar subagente `Explore` con una pregunta precisa. Consumir el resumen, no los archivos.
- Para detalles específicos de la feature: dejar que el developer los lea durante su pase de planificación (Flow B).
- Contenido de turnos ANTERIORES (código pegado por el usuario, resultado de agente previo, docs leídos para enrutamiento) es material legítimo para pasar inline.

**Lista de lectura permitida en Opus (solo modo orquestación):**
- `{context_path}`, `{backlog_path}`, `{task_path}/*.md`, docs de arquitectura (paths resueltos desde la tabla de paths en vault-setup)
- `.handoff/<TASK-ID>.md`
- `~/.claude/project-registry.md`
- Otros docs markdown en `<docs>`

En modo orquestación, el orquestador NO DEBE hacer `Read` de archivos fuente (`.go`, `.ts`, `.tsx`, `.py`, `.rs`, `.dart`, etc.). Delegar a un subagente. `Glob` y `Bash` (build/test) siempre están permitidos.

---

## 6. Escribir el plan técnico en Opus en lugar de usar Flow B

**Incorrecto:** el orquestador redacta un prompt de 200-400 líneas para el developer conteniendo firmas exactas de funciones Go, queries SQL, estructura de componentes React, orden de ejecución archivo por archivo, y lo marca `plan_preapproved=true`. El "plan" es efectivamente trabajo de architect + developer design, realizado en Opus.

**Por qué sucede:** el orquestador piensa que una spec completa le ahorra al developer el redescubrimiento. Lo hace — a 4x el costo por token. Y el orquestador NO es más rápido diseñando que el agente developer con contexto de código fresco cargado a través de sus skills de convención.

**Por qué está mal:** Flow A (`plan_preapproved=true`) existe para el caso donde el USUARIO escribió el plan a nivel de archivos en el chat y el orquestador solo lo retransmite. NO es una licencia para que el orquestador escriba el plan él mismo y luego lo "pre-apruebe" en nombre del usuario. Si el orquestador escribió el plan, eso es territorio de Flow B — el developer debió haberlo escrito.

**Correcto:** usar Flow B por defecto para cualquier cosa más grande que Trivial. Invocar developer con "planifica y DETENTE". El developer devuelve un resumen de plan de ~50 líneas (Sonnet). El orquestador lo muestra al usuario textualmente. Después de aprobación explícita del usuario, re-invocar developer con `plan_preapproved=true` y el plan corto aprobado inline. Dos invocaciones Sonnet superan un mega-prompt escrito en Opus.

**Test de legitimidad de Flow A:** antes de afirmar que un plan está "pre-aprobado en la conversación principal", verificar — ¿el USUARIO escribió la lista de archivos y el approach en el chat? ¿O el ORQUESTADOR lo sintetizó y el usuario solo aprobó la estrategia de alto nivel (ej., "ve con la opción B")? Si es lo segundo, eso es Flow B, no Flow A. "Opción B" = aprobación estratégica, no aprobación a nivel de archivos. El developer aún necesita producir el plan a nivel de archivos. Ver `plan-approval.md` para el árbol de decisión completo de Flow A / Flow B.

---

## 7. Escribir docs que el PM, developer o reporter deberían escribir

**Incorrecto:** el orquestador escribe archivos de tarea (frontmatter + contexto + AC) para tareas de seguimiento descubiertas durante la ejecución. El orquestador también escribe secciones de "Resultado de ejecución" en archivos de tareas completadas, y filas enriquecidas de backlog Done sintetizadas desde outputs de agentes.

**Por qué está mal:** la creación de tareas con el formato correcto del sistema de docs (Obsidian Dataview frontmatter, issues de Linear, o YAML de .workspace) + formato de backlog es trabajo del PM — el agente PM tiene las skills `backlog-management` y `prd-template` cargadas y produce el formato correcto a la primera. Los reportes finales de tareas y enriquecimiento del sprint son trabajo del `reporter` (o pueden ir inline en el handoff del developer). Todo Sonnet. El orquestador está produciendo docs a precio Opus que un agente Sonnet produce mejor.

**Correcto:**
- **Tareas de seguimiento descubiertas durante la ejecución:** lanzar `pm` con "crear seguimiento <NEW-ID>: alcance X, padre <PARENT-ID>, razón Y". Incluso para una sola fila. El costo es ~5k Sonnet vs ~5k Opus + riesgo de formato incorrecto. PM gana.
- **Cierres de tarea en Medium (reporter omitido):** instruir al developer en su prompt para agregar una sección `## Closeout` a `.handoff/<TASK-ID>.md` conteniendo `archivos tocados`, `métricas de validación`, `delta de cobertura`, `decisiones no obvias`. El orquestador copia esa sección textualmente en `task.md` y `sprint-current.md`. Cero síntesis Opus.
- **Cierres de tarea en Complex/Maximum:** el reporter ya está en el pipeline — escribe el resumen, el orquestador copia los paths de archivos.
- **Resumen al usuario en el chat al final:** este es un trabajo legítimo del orquestador — 1-3 párrafos cortos. No confundir "actualización de estado al usuario en el chat" (permitido) con "escribir la sección de resultado en task.md" (trabajo del PM/reporter).

---

## Disciplina de tokens — realidad de costos Opus

Opus es 4x Sonnet por token. El orquestador es el ÚNICO proceso Opus — cada subagente corre en Sonnet. Cada 1k tokens Opus = 4k equivalente Sonnet. Antes de cualquier lectura/escritura no trivial, preguntar: *¿puede un subagente Sonnet hacer esto?* Si sí, delegar.

**Legítimo (Opus es correcto):** triaje + enrutamiento, gates + confirmaciones con el usuario, lecturas de `<docs>` (NO código), resúmenes cortos de estado, verificación con `Glob`/`Bash`, copy-paste mecánico entre docs.

**Ilegítimo (delegar a subagente):** leer archivos de código fuente, escribir planes técnicos o pseudo-PRDs, escribir task.md con síntesis, resumir diffs de código, diseñar estructuras de componentes/SQL/DTOs.

En caso de duda: > ~3k tokens en una operación que no es triaje/gate → detenerse y lanzar subagente Sonnet.

---

## Paso de contexto — razonamiento completo (optimización de tokens)

**Regla:** Pasar contenido inline SOLO cuando ya está en contexto de una fuente LEGÍTIMA (mensajes del usuario, resultados de subagentes previos, docs del vault). Leer código fuente fresco para retransmitirlo = anti-patrón #5 — duplica costo a tarifas Opus.

Ejemplos: `[lee 8 archivos → los pega en el prompt] = 40k tokens (MAL)` vs `[ya tenía context.md del turno anterior → lo pasa inline] = 20k tokens (BIEN)` vs `[no tiene el archivo → le dice al agente el path] = BIEN`.

**Auto-verificación antes de cualquier llamada `Read`:** ¿el path está en la lista permitida de Opus (anti-patrón #5)? Extensión de código fuente → DETENERSE, lanzar subagente en su lugar.

Cada agente recibe SOLO lo que necesita. Consultar la tabla **"Input por agente"** en `SKILL.md` para los campos obligatorios exactos. A continuación, lo que NO debe recibir:

| Agente | NO recibe |
|--------|-----------|
| pm | código, diffs, paths de archivos fuente |
| scanner | tareas, docs |
| designer | código, reportes |
| architect | código fuente, reportes de QA/security |
| developer | reportes de QA/security (salvo en mode qa-fix) |
| tester | diffs completos, PRD directo (usa handoff + SPEC) |
| qa | historial de conversación |
| security | requerimientos, diseño |
| reporter | contexto extenso — solo TASK-ID + resumen |
| mkt-content | código, arquitectura, DB, PRDs |
