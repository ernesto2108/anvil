---
name: ui-component-scan
description: Escanea la librería de componentes UI existente para promover la reutilización y mantener la consistencia visual. Úsalo cuando el usuario diga "qué componentes existen", "revisar Storybook", "reutilizar componente", "ya hay un Button", o antes de crear un nuevo componente UI para evitar duplicados.
---

Escanea la librería de componentes UI existente (ej., Storybook, carpeta de componentes) para promover la reutilización y mantener la consistencia visual.

Pasos:
1. Identificar componentes comunes como Button, Input, Modal, etc.
2. Leer su documentación o implementación para entender props y comportamiento
3. Sugerir reutilización en lugar de crear nuevos componentes redundantes

Salida:
- Lista de componentes reutilizables encontrados
- Recomendaciones de uso para la tarea actual
