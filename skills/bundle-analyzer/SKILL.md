---
name: bundle-analyzer
description: Analiza el impacto en el tamaño del bundle frontend de nuevas dependencias o componentes. Usar cuando el usuario diga "bundle size", "check bundle", "too heavy", "tree shaking", "webpack analyze", o después de agregar una nueva dependencia npm para evaluar su impacto en el tamaño.
---

Monitorea el impacto en el tamaño del bundle final al agregar nuevas dependencias o componentes complejos.

Pasos:
1. Ejecutar análisis del bundle (ej. webpack-bundle-analyzer, vite-bundle-visualizer)
2. Comparar el tamaño actual del bundle contra la línea base
3. Identificar módulos "pesados" y sugerir alternativas

Objetivo: Mantener bajos los tiempos de carga inicial y evitar el bloat.
