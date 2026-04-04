# Principios de Diseno de Agentes

## Filosofia

Cada agente es un especialista con responsabilidades claras, permisos minimos y outputs definidos. No hay agentes genericos ni "do-it-all".

## Reglas de diseno

### 1. Un rol, un agente

Cada agente hace exactamente una cosa. Si un agente necesita hacer dos cosas distintas, son dos agentes.

- PM escribe PRDs, no disena arquitectura
- Architect disena sistemas, no escribe codigo
- Developer escribe codigo, no hace reviews

### 2. Permisos minimos

Cada agente tiene acceso solo a lo que necesita:

| Permiso | Quien lo tiene |
|---------|---------------|
| Escribir codigo de produccion | Solo developer |
| Escribir tests | Solo tester |
| Escribir migraciones | Solo dba |
| Escribir infra | Solo devops |
| Escribir al vault | Solo el agente que produce ese doc |

### 3. Inputs y outputs explicitos

Cada agente declara:
- **Que lee** antes de trabajar (pre-check)
- **Que produce** (output)
- **Donde lo escribe** (path en vault o repo)

### 4. Pre-check obligatorio

Antes de hacer nada, el agente verifica que sus inputs existen. Si faltan, para y reporta. Nunca adivina.

```
Si prd.md no existe → STOP, reportar al orquestador
Si design.md no existe → STOP, reportar al orquestador
```

### 5. Token-conscious

Los agentes deben ser concisos:
- Frontmatter YAML corto y preciso
- Instrucciones directas, sin texto de relleno
- Skills que se cargan bajo demanda (no embebidas en el agente)

## Estructura de un agente

```markdown
---
name: nombre-del-agente
description: Una linea que el sistema usa para decidir cuando invocarlo
tools: Lista de herramientas permitidas
model: sonnet | opus | haiku
---

# Rol

Que hace y que NO hace.

## Pre-check (OBLIGATORIO)

Que archivos verificar antes de empezar.

## Flujo de trabajo

Pasos concretos.

## Reglas

Restricciones especificas.

## Output

Que produce y donde lo escribe.
```

## Eleccion de modelo

| Modelo | Cuando usarlo | Agentes |
|--------|--------------|---------|
| **opus** | Razonamiento complejo, decisiones de diseno, discovery | pm, architect, designer |
| **sonnet** | Implementacion, analisis, reviews | developer, tester, qa, security, dba, devops |
| **haiku** | Tareas simples, resumen, formato | reporter |

## Anti-patrones

- **Agente god** — un agente que hace todo. Dividir en especialistas.
- **Agente sin pre-check** — empieza a trabajar sin verificar inputs. Puede producir basura.
- **Agente verbose** — instrucciones largas que queman tokens. Mantener conciso.
- **Agente sin permisos claros** — puede escribir donde no debe. Definir permisos explicitos.
- **Agente que adivina** — asume que algo existe sin verificar. Siempre pre-check.
