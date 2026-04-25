---
name: visual-diff
disable-model-invocation: true
description: Detecta regresiones visuales comparando capturas de pantalla antes/después de cambios en la UI. Úsalo cuando el usuario diga "visual regression", "comparación de screenshots", "la UI se ve diferente", "el CSS rompió algo", o después de modificar estilos, temas, o componentes compartidos.
---

Detecta regresiones visuales inesperadas al modificar CSS, temas, o componentes principales.

Pasos:
1. Capturar screenshots de las páginas/componentes objetivo antes y después de los cambios
2. Comparar las dos versiones y resaltar diferencias de píxeles (ej., pixelmatch, reg-viz)
3. Revisar las diferencias en busca de regresiones o cambios intencionados

Salida:
- Reporte de diff visual mostrando imágenes "antes" y "después" con resaltado

Nota: Requiere herramientas de screenshot — no disponible sin configuración de automatización de navegador.
