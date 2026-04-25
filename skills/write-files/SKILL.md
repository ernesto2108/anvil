---
name: write-files
description: Crear o modificar archivos con cambios mínimos y enfocados, respetando los permisos. Úsalo al escribir cambios de código para asegurar diffs mínimos, formato preservado, y sin modificaciones no relacionadas.
user-invocable: false
---

Crear o modificar archivos en el repositorio.

Reglas generales:
- modificar el mínimo número de archivos
- cambios pequeños y enfocados
- preservar el formato
- nunca reescribir código no relacionado
- explicar los cambios con claridad

Estrategia de escritura segura:
1. Leer el archivo primero
2. Entender el contexto
3. Aplicar el parche mínimo
4. Mostrar el diff
5. Luego escribir

Formato de salida por cambio:

File: <path>
Reason: <por qué>

Diff:
+ <líneas agregadas>
- <líneas eliminadas>

Nunca:
- reescribir archivos en masa
- cambiar la arquitectura sin diseño
- modificar archivos fuera del scope

## Validación de la Fuente de Diseño

Al modificar código de UI que tiene un archivo de diseño correspondiente (.pen o Figma):

1. **Antes de escribir**: Leer el diseño para entender el espaciado exacto, colores, layout y tipografía
2. **Después de escribir**: Comparar la implementación visualmente contra el diseño
3. **Nunca asumir**: Si no tienes el diseño abierto, abrirlo primero. La memoria de "cómo se veía el diseño" no es confiable
4. **Reportar discrepancias**: Si el diseño y el código actual ya difieren, notificarlo al usuario antes de hacer cambios
