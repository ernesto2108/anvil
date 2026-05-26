---
name: agent-teams
description: Convención de Agent Teams del Líder cuando el flag `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` está activo — formato de `team_name`, cuándo asignarlo, casos de uso de `SendMessage` lateral entre miembros del mismo team, restricciones operativas y self-check antes de cada spawn paralelo. Cárgalo cuando el Líder esté por spawnear 2+ sub-agentes en paralelo y necesite decidir si asignar un team compartido. Reemplaza la sección `## Referencia — Agent Teams` del leader.md.
user-invocable: false
---

# agent-teams

Mecánica de Agent Teams del Líder. Depende de que el flag `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` esté activo en `~/.claude/settings.json`. Sin el flag, los campos `team_name` y la tool `SendMessage` no existen y esta skill entera es no-op — los spawns siguen siendo jerárquicos puros (sub-agente → Líder → siguiente sub-agente).

Con el flag activo, el Líder gana **una sola** capacidad nueva: agrupar sub-agentes paralelos en un team compartido para que puedan enviarse mensajes laterales (`SendMessage`) sin pasar por el Líder. La coordinación principal sigue siendo jerárquica — el team es un atajo lateral controlado, no una sustitución del canal Líder → sub-agente.

## Cuándo se ejecuta

Antes de cada spawn paralelo de 2+ sub-agentes (los marcados con `∥` en §Sub-agentes paralelos del `leader.md`). También al registrar el run en `plan.md` cuando hubo teams, para preservar la trazabilidad lateral.

## Cuándo asignar `team_name`

| Situación | `team_name` | Justificación |
|---|---|---|
| Spawn de **2+ sub-agentes en paralelo** (marcados con `∥` en el árbol de §Sub-agentes paralelos) | **Obligatorio** — mismo nombre para todos los del grupo paralelo | Habilita comunicación lateral si la necesitan |
| Spawn **secuencial** (`pm` → `architect`, `developer` → `tester`, etc.) — un sub-agente a la vez | **Omitir** | No hay co-presencia temporal; el handoff va vía Líder |
| Spawn **único** (un solo sub-agente en el modo o en la fase) | **Omitir** | Sin pares con quien comunicarse |
| Sub-agentes paralelos pero **completamente independientes** (no necesitan ni podrían usar `SendMessage` — ej. dos `explorer` consultando fuentes web disjuntas) | Asignar igual el `team_name` | Costo cero si nadie usa `SendMessage`; mantiene la regla simple ("paralelo ⇒ team") |

## Convención de nombres de team

- Formato: `<modo-corto>-<slug-del-objetivo>`. Ejemplos:
  - Planeación con `designer ∥ dba`: `team_name = "plan-checkout-redesign"`
  - Pruebas con `reviewer ∥ security`: `team_name = "test-pr-482"`
  - Explorador con múltiples fuentes paralelas: `team_name = "explore-payments-schema"`
  - Planeación de feature con `architect ∥ agent-designer`: `team_name = "feat-billing-export"`
- Reglas:
  - Minúsculas, kebab-case, sin espacios, máx 40 chars
  - Único por run — si dos grupos paralelos coexisten dentro del mismo modo, sufijar con `-a`, `-b` (`team-x-design-a`, `team-x-design-b`)
  - Si el run tiene un TASK-ID, preferir el TASK-ID como slug (`team_name = "feat-PROJ-123"`)
- Registrar el `team_name` en `plan.md` bajo `## Teams del run` con la lista de miembros y el propósito.

## Cuándo un sub-agente DEBE usar `SendMessage` (en lugar de devolver al Líder)

El mecanismo por defecto sigue siendo: sub-agente termina → output al Líder → Líder pasa al siguiente. `SendMessage` se usa **solo** cuando todas estas condiciones se cumplen:

1. Ambos agentes (emisor y receptor) están en el **mismo `team_name`** durante el spawn paralelo.
2. El emisor produce un artefacto **que el receptor necesita para arrancar su propio trabajo en paralelo**, no al final del pipeline.
3. Esperar al Líder introduce latencia evitable sin ganancia (el Líder no necesita ver el artefacto intermedio para decidir nada).

### Caso canónico — `explorer` → developer del stack en paralelo

(En este ejemplo, `developer` denota el developer del stack relevante: `developer-backend`, `developer-frontend` o `developer-mobile`.)

Escenario: Modo Integración para un bug fix donde la causa raíz no está clara. El Líder spawnea `explorer` y `developer` **en el mismo team** (`team_name = "fix-<slug>"`). El `developer` arranca con el SPEC base pero queda esperando los hallazgos del `explorer` para tocar el archivo correcto.

- Sin teams: `explorer` termina → output al Líder → Líder construye prompt → spawn `developer` (serial). Latencia: explorer + construcción del prompt.
- Con teams: `explorer` y `developer` arrancan juntos en `team="fix-<slug>"`. El `explorer` termina y usa `SendMessage(to="developer", payload=<hallazgos>)`. El `developer` recibe el payload inline mientras ya estaba preparando el entorno y arranca el cambio sin esperar al Líder.

El Líder sigue recibiendo el output final de ambos al cierre del team — el `SendMessage` no salta el reporte hacia arriba, solo evita el cuello de botella lateral.

### Otros casos válidos (lista no exhaustiva)

| Team | Emisor → Receptor (`SendMessage`) | Qué se transfiere lateralmente |
|---|---|---|
| `designer ∥ dba` (Planeación con UI + DB) | `designer` → `dba` | Lista de campos que la UI necesita visible/editable, para que el `dba` valide que el schema los soporta sin pasar por el Líder |
| `reviewer ∥ security` (Pruebas) | `security` → `reviewer` | Hallazgos críticos de auth/crypto encontrados antes, para que el `reviewer` los ancle en su reporte de severidad |
| `architect ∥ agent-designer` (Planeación con artefacto secundario de IA) | `architect` → `agent-designer` | Contrato del nuevo command/skill que el artefacto debe respetar |

## Cuándo NO usar `SendMessage` (vuelta al Líder es obligatoria)

- **Cualquier output final del sub-agente** — el Líder DEBE recibir el resultado completo para aplicar self-critique (#2), registrar en `plan.md` y decidir el siguiente paso. `SendMessage` nunca reemplaza el reporte final hacia arriba.
- **Cualquier decisión que el Líder deba conocer** — cambios de scope, contradicciones con `.project-context/`, bloqueos, preguntas abiertas, contradicciones entre sub-agentes. Todo esto sube al Líder (Regla #8 + §Flujo de escalación).
- **Cuando los sub-agentes están en spawns secuenciales** — no hay team, no hay canal lateral. El handoff va via Líder.
- **Cuando no hay co-presencia temporal** — si el receptor aún no fue spawneado, no se puede enviar mensaje. Esperar y pasar inline al spawnear.

## Restricciones operativas

1. **El Líder NO usa `SendMessage`** — el Líder se comunica con sub-agentes solo vía el spawn (`Agent`). `SendMessage` es lateral entre miembros del mismo team, no vertical.
2. **Auditoría obligatoria** — cuando un sub-agente reporta de vuelta al Líder, debe incluir en su output una sección `## Mensajes laterales emitidos/recibidos` listando los `SendMessage` que envió o recibió, con el destinatario/origen y un resumen de 1 línea del payload. Sin esta sección, el Líder no puede reconstruir el debate interno (#2) ni el `plan.md` correctamente.
3. **El self-critique #2 aplica al output final**, no a los mensajes laterales. Si un `SendMessage` contiene información que contradice `.project-context/`, el Líder lo detectará cuando los outputs finales lleguen — no antes.
4. **Si el flag `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` no está activo**, ignorar toda esta skill. Los spawns paralelos siguen funcionando sin `team_name` y los sub-agentes siempre devuelven al Líder.

## Self-check antes de cada spawn paralelo

Antes de invocar `Agent`, responder:

1. ¿Voy a spawnear **2+ sub-agentes en paralelo** ahora mismo? → Sí: asignar `team_name` compartido. No: omitir el campo.
2. ¿El `team_name` sigue la convención (`<modo>-<slug>`, kebab-case, único por run)?
3. ¿Registré el team en `plan.md` bajo `## Teams del run`?

Si cualquiera falla → corregir antes del spawn.

## Reglas

- No crear teams "por las dudas" — solo cuando hay 2+ sub-agentes paralelos co-presentes.
- No reusar `team_name` entre runs distintos — cada run construye sus propios teams.
- No saltarse la sección `## Teams del run` en `plan.md` — sin ella, los runs con teams quedan sin trazabilidad lateral.
- El Líder NUNCA emite `SendMessage` — solo los miembros del team entre sí.
- Si el flag `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` no está activo, esta skill es no-op: ignorar todos los pasos y spawnear paralelos sin `team_name`.
