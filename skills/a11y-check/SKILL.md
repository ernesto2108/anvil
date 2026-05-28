---
name: a11y-check
description: Audita la UI para verificar cumplimiento de accesibilidad WCAG 2.1 y reporta violaciones. Usar cuando el usuario diga "accessibility check", "a11y audit", "WCAG compliance", "screen reader", "ARIA labels", "keyboard navigation", o al revisar componentes de UI para inclusividad.
user-invocable: true
---

Audita la UI para verificar cumplimiento de accesibilidad (WCAG 2.1).

Pasos:
1. Usar herramientas automatizadas (axe-core, cypress-axe, playwright-axe)
2. Escanear componentes o páginas en busca de violaciones de accesibilidad (ej. falta de texto alternativo, bajo contraste de color, falta de etiquetas ARIA)
3. Reportar los problemas encontrados y proponer correcciones

Salida:
- Reporte de cumplimiento
- Lista de violaciones de accesibilidad (Critical, Moderate, Low)
