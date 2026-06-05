---
name: task-writer
description: Reglas para escribir archivos de task atómicos a partir de un archivo spec. Define los templates de task individual y archivo padre (épica/historia), categorías, inferencia del agente ejecutor y protocolo de escalación. Usada exclusivamente por el agente `task-writer`.
---

# Task Writer

Esta skill define cómo el `task-writer` traduce un archivo spec en un conjunto de archivos `.md` independientes — uno por task — y, cuando aplica, un archivo padre de épica que los agrupa. No cubre actualización del backlog ni transiciones de estado; eso vive en `backlog-management`.

## Tipos de descomposición

El agente pregunta al inicio qué tipo de trabajo es:

- **feature / historia** → genera archivos individuales de task (uno por task).
- **épica** → genera 1 archivo padre + archivos individuales por cada subtask.

La elección determina si se emite un archivo padre y cómo se nombran los archivos.

## Template — archivo de task individual

Cada task es un archivo `.md` independiente con este formato exacto:

```markdown
---
name: "<FEATURE_ID>-<NN>-<slug-corto>"
type: "setup" | "implementation" | "integration" | "validation"
priority: "HIGH" | "MEDIUM" | "LOW"
agent: "developer-backend" | "developer-frontend" | "developer-mobile"
points: 1 | 2 | 3 | 5 | 8
milestone: "<milestone>"
feature_id: "<FEATURE_ID>"
dependencies: []
inputs:
  - "<nombre>: <tipo> (<restricción>)"
outputs:
  - "<nombre>: <tipo>"
validation_rules:
  - "<regla>: <valor>"
---

# {name}

## 🎯 Objetivo
[Comportamiento observable esperado — el QUÉ, no el CÓMO]

## 📋 Contexto Técnico
[Tabla, endpoint, contrato u otra referencia técnica extraída del spec]

## 🔗 Interfaces
- Llamado por: `path/caller`
- Llama a: `path/callee`

## ✅ Criterios de Aceptación
- [ ] [criterio verificable]
```

**Reglas del template:**

- `inputs`, `outputs` y `validation_rules` son opcionales — omitir si no aplican a la task.
- `## 🔗 Interfaces` es opcional — omitir si la task no tiene vecinos claros.
- `dependencies: []` cuando no hay dependencias; si las hay, listar los `name` de las tasks que deben completarse primero.
- Sin sección "Pasos de Ejecución" — el task-writer no sabe implementación.

## Template — archivo padre (épica)

```markdown
---
name: "<FEATURE_ID>-epic-<slug>"
type: "epic"
priority: "HIGH" | "MEDIUM" | "LOW"
milestone: "<milestone>"
feature_id: "<FEATURE_ID>"
subtasks:
  - "<FEATURE_ID>-01-<slug>"
  - "<FEATURE_ID>-02-<slug>"
---

# {name}

## 🎯 Objetivo
[Descripción del valor de negocio de la épica]

## 📋 Subtareas

| ID | Tipo | Agente | Pts | Depende de |
|---|---|---|---|---|
| <FEATURE_ID>-01-<slug> | implementation | developer-backend | 3 | — |
| <FEATURE_ID>-02-<slug> | integration | developer-frontend | 2 | 01 |

**Total pts:** N
```

## Categorías de task

| Categoría | Significado | Ejemplos |
|---|---|---|
| `setup` | Tipos, interfaces, schemas vacíos, sin lógica. Habilita el resto. | Crear interface `EventStore`, crear DTO `CreateEventRequest` |
| `implementation` | Lógica concreta encapsulada en una unidad. | Implementar método `Create` del repositorio, implementar service `BookEvent` |
| `integration` | Conectar dos componentes ya existentes. | Wirear handler con service, registrar handler en router |
| `validation` | Tests, verificación de comportamiento end-to-end. | Tests unit del service, test de contrato del endpoint |

## Reglas de descomposición

1. Una task = un archivo principal (puede tocar 1-2 adicionales de imports/config si son inevitables).
2. Si el developer necesita leer >5 archivos para entender el contexto → la task es demasiado amplia, dividir.
3. Orden topológico obligatorio: setup → implementation → integration → validation.
4. Máx 15 tasks por feature / épica. Si supera, registrar como decisión abierta y entregar las 15 primeras por prioridad.
5. Una preocupación por task — si toca backend Y frontend/mobile, separar en dos tasks.
6. Tests son SIEMPRE una task separada de la implementación.
7. Las dependencias deben ser explícitas — listar los `name` de tasks previas en `dependencies`.
8. Puntos Fibonacci: 1, 2, 3, 5, 8. Si una task se estima en 13+ → dividir.

## Inferencia del agente ejecutor

`agent` es obligatorio en todas las tasks. Valores exactos: `developer-backend`, `developer-frontend`, `developer-mobile`. Inferir en este orden:

1. **Por extensión/path del archivo principal:**
   - `.dart` o path bajo `lib/` → `developer-mobile`
   - `.tsx`, `.jsx`, `.astro`, `.ts` en contexto frontend, o path bajo `src/components/`, `src/pages/`, `src/hooks/` → `developer-frontend`
   - `.go`, `.py`, `.rs`, o path bajo `internal/`, `cmd/`, `pkg/`, `api/` → `developer-backend`
2. **Si el path es ambiguo** → desempatar con el campo `Dominio` del spec (`mobile`, `frontend`, `backend`, `fullstack`).
3. **Si aún no se infiere** → marcar `developer-[?]` y listar en el output de cierre.

## Design reference

Si el spec tiene sección `## Design References` con `Type != none`, agregar campo `design_reference` en el frontmatter de cada task que toque UI (componentes, pantallas, vistas, frontend visible):

```yaml
design_reference: "<valor Location copiado verbatim del spec>"
```

Omitir el campo en tasks sin UI o cuando `Type: none`.

## Protocolo de escalación

Detener y reportar al humano cuando:

| Condición | Mensaje |
|---|---|
| Tasks superan 15 | `Generé >15 tasks — entregué las 15 primeras por prioridad. ¿Partir el feature en sub-features o ampliar el límite?` |
| Dependencia circular detectada | `Ciclo detectado: [A → B → C → A]. Re-invocar spec-writer para resolver el orden.` |
| Una task requiere decisión técnica no presente en el spec | `Task [X] requiere decisión [Y] no resuelta en el spec. Re-invocar spec-writer.` |

## Presupuesto

- Objetivo: 8K tokens | Máximo: 15K tokens
- Máx llamadas a herramientas: 8 (lectura del spec + verificación puntual ≤4 LS/Glob)
- No leer código de producción amplio
