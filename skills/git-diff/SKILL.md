---
name: git-diff
description: Inspeccionar y resumir cambios del repositorio usando git diff. Usar cuando el usuario diga "qué cambió", "muestra el diff", "revisa cambios", "resume modificaciones", o antes de crear un commit o pull request.
user-invocable: true
disable-model-invocation: true
---

Inspecciona solo los cambios realizados en el repositorio usando git diff.

Por qué:
- revisar cambios mínimos
- evitar re-leer archivos completos
- habilitar revisión de código segura
- detectar ediciones no intencionadas
- reducir tokens y ruido

Comandos típicos:
- `git diff` — cambios actuales en el working directory
- `git diff --staged` — solo staged
- `git diff path/to/file.go` — específico de archivo
- `git diff main...feature-branch` — comparación de ramas

Reglas de uso:
- SIEMPRE preferir git diff antes de leer archivos completos
- revisar solo las líneas modificadas
- ignorar cambios de solo espacios en blanco
- resumir diffs grandes
- destacar: cambios de lógica, nuevas condiciones, verificaciones eliminadas, cambios de esquema, riesgos de concurrencia

Formato de output por archivo:

File: <ruta>
Change type: modified | added | deleted
Summary: <una línea>

Diff:
<solo las líneas relevantes>

Buenas prácticas:
- diffs pequeños (< 200 líneas)
- una preocupación por cambio
- evitar refactorizaciones no relacionadas
- dividir cambios grandes

Nunca:
- leer el repositorio completo cuando existe un diff
- aprobar cambios sin diff
- ocultar cambios
- hacer merge automático a ciegas
