---
name: mode-gate
description: Protocolo unificado de debate interno (Líder ↔ sub-agentes) y gate de salida (Líder ↔ usuario) que aplica a los cuatro modos del Líder (Explorador, Planeación, Integración, Pruebas). Define cuándo activar el debate interno, cómo orquestarlo hasta consenso o escalamiento, el formato del gate de salida según modo, el protocolo de debate externo con el usuario y las violaciones que marcan el run como failed. Cárgalo al cerrar cualquier modo antes de presentar el output al usuario. Reemplaza las secciones "Debate interno" y "Gate de salida" inline de cada modo del leader.md.
user-invocable: false
---

# mode-gate

Protocolo de control del cierre de cada modo del Líder. Cubre dos responsabilidades acopladas:

1. **Debate interno (Líder ↔ sub-agentes):** detectar y resolver posturas divergentes entre sub-agentes (o entre un sub-agente y la postura técnica del Líder) ANTES de presentar el output al usuario.
2. **Gate de salida (Líder ↔ usuario):** presentar el resultado del modo, abrir potencial debate con el usuario y esperar confirmación explícita del consenso final antes de avanzar.

Estas dos piezas son inseparables — el debate interno alimenta lo que se presenta en el gate; el gate nunca se cierra sin haber pasado por el debate interno cuando hubo divergencia.

## Cuándo se ejecuta

Al cerrar **cualquier** modo (Explorador, Planeación, Integración, Pruebas), inmediatamente antes de construir el "Output al usuario" final. NO se salta — ni siquiera cuando "todos los sub-agentes coincidieron" o "la solución es obvia".

## Debate interno (Líder ↔ sub-agentes) — OBLIGATORIO

### Cuándo activarlo

- Cuando dos sub-agentes producen outputs divergentes o contradictorios sobre el mismo punto.
- Cuando un sub-agente y la postura técnica propia del Líder difieren materialmente.
- Cuando una fuente secundaria consultada por el `explorer` contradice la conclusión principal.

### Cómo resolverlo

El Líder NUNCA resuelve la divergencia por jerarquía ni descarta la postura minoritaria. Aplica el Paso 1 del §Protocolo de debate del `leader.md` en orden:

1. **Consistencia con `.project-context/`** — qué output cuadra con patrones documentados
2. **Alcance del modo** — qué output respeta el scope del modo activo
3. **Menor riesgo reversible** — preferir el más fácil de corregir
4. **Criterio técnico propio** — adoptar la posición más sólida y justificarla

El Líder DEBE re-invocar al sub-agente cuya postura quedó cuestionada pasándole inline la postura contraria con el formato:

> "El otro lado encontró/propuso X. Tu posición fue Y. Gap: [concreto]. Revisa tu posición y devuelve consenso o refutación razonada."

### Cuándo escalar al usuario vs resolver solo

- **Resolver solo:** la divergencia se cierra dentro de los criterios del Paso 1 (consistencia con `.project-context/`, alcance del modo, menor riesgo, criterio técnico) y el sub-agente re-invocado devuelve consenso o refutación razonada que el Líder acepta.
- **Escalar al usuario:** tras una iteración de re-invocación no converge, o la decisión cambia algo previo del usuario, o requiere contexto de negocio que el Líder no tiene → §Protocolo de debate Paso 2 del `leader.md`.

### Formato del bloque `## Debate interno` en el output

El consenso (o el punto medio acordado) DEBE incluirse como subsección `## Debate interno` dentro del bloque del modo (hallazgos en Explorador, plan en Planeación, etc.) listando:

- **Posición A:** [resumen de 1-2 líneas con el sub-agente origen]
- **Posición B:** [resumen de 1-2 líneas con el sub-agente origen]
- **Divergencia exacta:** [el punto concreto donde no coincidían]
- **Resolución acordada:** [qué se aceptó]
- **Justificación:** [referencia a `.project-context/`, ADR, criterio del Paso 1 aplicado]

**Nunca omitir esta subsección cuando hubo debate** — su ausencia equivale a esconder el conflicto y compromete la trazabilidad.

Si NO hubo divergencia detectada, el Líder DEBE registrar una nota explícita "Sin debate interno — todos los sub-agentes coincidieron" en el progress log, pero NO incluir la subsección en el output al usuario.

**Violación:** si el Líder presenta el output sin haber orquestado el debate cuando hubo divergencia, DEBE marcar el run como `failed` con `mcp__anvil__complete_orchestration(run_id, "failed")` y reportar la violación al usuario.

## Gate de salida (Líder ↔ usuario) — OBLIGATORIO

### Principios inviolables

- Ningún modo auto-avanza al siguiente sin confirmación humana explícita.
- Si el usuario solicitó originalmente un pipeline multi-modo ("planea e implementa", "investiga y arregla"), el Líder DEBE igual detenerse al cierre de cada modo — la solicitud inicial NUNCA autoriza saltarse el gate.
- La claridad de la solución NUNCA es razón válida para saltarse el gate. El gate no es un mecanismo para resolver dudas del Líder — es un punto de control humano sobre el avance.
- La urgencia, la trivialidad aparente, el scope claro, la confianza alta en la siguiente fase: ninguna de esas razones justifica saltarse el gate.

### Variantes por modo

| Modo | Qué se presenta en el gate | Confirmación requerida |
|---|---|---|
| Explorador | Hallazgos + recomendación de modo siguiente | "¿Continúo con [modo recomendado], o prefieres otro camino?" |
| Planeación | Plan + PRD + decisiones + tasks + archivos a tocar | "¿Apruebas el plan? ¿Procedo con Integración?" |
| Integración | Diff + tests passing + nota al vault + archivos modificados | "¿Mergeo / cierro la tarea?" |
| Pruebas | Reporte de hallazgos por severidad + estado final | "¿Corrijo / escalo / cierro?" |

### Formato obligatorio del gate de salida

El Líder DEBE cerrar cada modo con un bloque de cierre explícito ubicado al final del output al usuario (después de los hallazgos/plan/diff integrados y antes de cualquier nota suelta). El bloque NO es opcional aun si el sub-agente ya devolvió un resumen extenso — es la señal inequívoca al usuario de que el Líder se detuvo y espera.

**Plantilla genérica:**

```
## [Modo] completado

**[Resultado clave del modo]:** [resumen de 2-4 bullets]

**[Recomendación / próximo paso]:** [una línea explicando por qué]

[Pregunta de confirmación correspondiente al modo]
```

Para los formatos exactos del bloque por modo, ver la skill `leader/output-formats`. Para Explorador en particular, el bloque debe incluir el campo `Modo recomendado:` (con `Pendiente — necesito tu criterio antes de avanzar` si no aplica).

El bloque NO se reemplaza por frases libres tipo "¿avanzamos?", "¿procedo?", "lo arreglo entonces" — el formato es literal.

### Prohibición explícita de acción post-gate sin confirmación (aplica especialmente a Explorador)

Después de presentar el gate, el Líder NO puede spawnear NINGÚN agente de acción hasta recibir confirmación humana explícita. La lista no exhaustiva de agentes prohibidos en este punto incluye: `developer-backend`, `developer-frontend`, `developer-mobile`, `agent-designer`, `dba`, `tester`, `devops`, `designer`, `architect`, `pm`, `reporter` (excepto cuando aplica el cierre estándar de un run que ya modificó archivos antes del modo), y cualquier otro agente que modifique archivos del repo o del sistema de IA.

Solo se permite re-invocar `explorer` (para profundizar) o `context-init` (si emerge `CONTEXT_MISSING` durante el debate del gate).

El Líder NO interpreta el output de los sub-agentes como autorización tácita para actuar — la autorización SOLO viene del usuario, en lenguaje explícito.

### Caso especial — tarea mixta implícita (aplica a Explorador)

Cuando el usuario reporta un problema sin pedir explícitamente exploración (ej. "tengo un problema con X", "algo no funciona en Y", "el feature Z se comporta raro"), el Líder detecta Modo Explorador para investigar primero — pero al terminar la exploración SIEMPRE DEBE hacer el gate antes de pasar a acción, aunque la solución sea obvia y aunque el fraseo del usuario insinúe que quería el arreglo directo.

La interpretación correcta es: "el usuario reportó un problema → primero entiendo, luego confirmo antes de actuar", NUNCA "el usuario reportó un problema → asumo que quiere el fix y procedo". El gate es la barrera explícita que separa las dos interpretaciones.

## Debate externo (Líder ↔ usuario) — OBLIGATORIO

Después de presentar el output del gate, el Líder DEBE tratar la respuesta del usuario como apertura potencial de un debate, no como aprobación automática.

### Principios

- NUNCA asumir que el silencio, la ausencia de contradicción, o un acuse breve ("ok", "vi") equivale a consenso sobre el resultado final.
- Si el usuario expone una visión diferente, dudas, matices o señales de desacuerdo (aunque sean parciales), el Líder DEBE facilitar el debate haciendo las preguntas que considere necesarias — sin límite arbitrario de turnos — hasta entender el desacuerdo en su raíz.
- El Líder DEBE exponer su propia posición en cada turno del debate (usando el campo "Lo que yo pienso" del §Protocolo de debate Paso 2 del `leader.md`), NUNCA limitarse a recoger la opinión del usuario.
- Si la postura del usuario contradice algo de `.project-context/` o un ADR previo, el Líder DEBE señalarlo explícitamente.

### Búsqueda de consenso

- El debate DEBE buscar consenso o punto medio documentado entre la postura del Líder (y sus sub-agentes) y la del usuario.
- Si emerge un consenso distinto al output inicial, el Líder DEBE reformular el output reflejando ese consenso ANTES de cerrar el gate — incluyendo re-invocar sub-agentes si el cambio es material (ej. re-invocar `pm`/`requirements`/`architect` en Planeación).

### Cierre del gate

- El gate de salida NUNCA se cierra hasta que el usuario dé confirmación explícita ("dale", "continúa", "OK", o equivalente literal) sobre el **consenso final** — NO sobre el output inicial del Líder.
- Si el output cambió durante el debate, la confirmación DEBE referirse a la versión final.

### Cierre forzado por el usuario

- Si el usuario pide cerrar sin debate ("está bien así") y el Líder detecta que persiste una divergencia material no resuelta (ej. el plan contradice un ADR vigente y el usuario no lo justifica), el Líder DEBE plantear la divergencia una última vez antes de cerrar.
- Si el usuario insiste en cerrar, registrar la divergencia como nota en el `plan.md` del run y cerrar.

## Violación = run failed

Si el Líder:

- Presenta output sin haber orquestado el debate interno cuando hubo divergencia, o
- Auto-avanza al siguiente modo sin confirmación del usuario, o
- Cierra el gate sin confirmación explícita del consenso final, o
- Spawnea un agente de acción tras el gate de Explorador sin autorización explícita del usuario

→ DEBE marcar el run como `failed` con `mcp__anvil__complete_orchestration(run_id, "failed")` y reportar la violación al usuario en el output final.

## Reglas

- Esta skill aplica a los 4 modos sin excepción — no hay "modo simple" que la saltee.
- El debate interno SIEMPRE precede al gate de salida. Sin debate (cuando hubo divergencia), no se presenta gate.
- El gate de salida SIEMPRE espera confirmación humana explícita. La confirmación implícita (silencio, "ok") no cuenta.
- Si el output cambió durante el debate externo, la confirmación debe referirse al consenso final, no al output inicial.
- Las violaciones de esta skill marcan el run como `failed` — no se silencian ni se "absorben" como notas.
