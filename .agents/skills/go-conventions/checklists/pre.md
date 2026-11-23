# Checklist Pre-Implementación

Antes de escribir código Go, verificar:

- [ ] El placement de paquetes sigue la estructura existente del proyecto
- [ ] Las interfaces están definidas por el consumidor, no por el productor
- [ ] El wrapping de errores incluye contexto de la operación
- [ ] El contexto se pasa como primer parámetro
- [ ] Las llamadas externas tienen timeouts
- [ ] El acceso concurrente está protegido
- [ ] Los tests son table-driven con subtests
- [ ] No se introducen importaciones circulares
- [ ] El logging usa pares clave-valor estructurados (sin concatenación de strings)
- [ ] No se loguean datos sensibles (contraseñas, tokens, PII)
- [ ] Las queries SQL usan placeholders parametrizados, no interpolación de strings
