---
name: generate-diagram
description: Crear diagramas Mermaid.js para arquitectura, flujos, secuencias, ERDs y C4. Usar cuando el usuario diga "dibuja un diagrama", "crea un flowchart", "diagrama de secuencia", "ERD", "diagrama de arquitectura", "visualizar", o cuando necesite documentación visual.
---

Crea o actualiza representaciones visuales de código o arquitectura usando sintaxis Mermaid.js dentro de archivos Markdown.

Capacidades:
- Flowcharts (flujos de proceso)
- Diagramas de secuencia (interacción entre componentes)
- Diagramas de clase/módulo (estructura)
- Diagramas de Relación de Entidades (ERD — esquema de base de datos)
- Diagramas C4 (contexto del sistema y contenedores)

Reglas:
- Usar siempre la sintaxis estándar de Mermaid.js
- Colocar los diagramas dentro de bloques ```mermaid ... ```
- Mantener los diagramas lo suficientemente simples para ser legibles en una pantalla estándar
- Preferir diagramas de secuencia para explicar flujos asíncronos complejos
- Usar C4 para vistas generales arquitectónicas de alto nivel
