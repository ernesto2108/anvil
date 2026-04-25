---
name: read-files
description: Leer archivos del proyecto de forma segura para recopilar contexto antes de tomar decisiones. Usar cuando se necesite entender el código existente antes de modificarlo, recopilar contexto de múltiples archivos, o resumir codebases grandes.
user-invocable: false
---

Leer solo lo necesario para completar la tarea.

Capacidades:
- leer uno o múltiples archivos
- leer carpetas de forma recursiva
- resumir archivos grandes
- extraer solo las secciones relevantes
- buscar por keyword o símbolo

Reglas de uso:
- evitar cargar todo el repo si no es necesario
- preferir lecturas dirigidas (archivos específicos primero)
- resumir outputs largos
- destacar solo las partes relevantes

Formato de salida:
- path del archivo
- resumen breve
- solo los snippets relevantes

Nunca:
- alucinar contenido de archivos
- asumir código sin haberlo leído
